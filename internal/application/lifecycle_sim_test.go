package application

import (
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/preferences"
	"github.com/AstraFoundry/KumoApp/internal/domain/profile"
	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/mihomo"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/paths"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/ports"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/process"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/sysproxy"
)

// These tests run the real Mihomo. Point KUMO_TEST_MIHOMO at a binary to
// run them: `make test-lifecycle` downloads one.

const simProfile = `proxies:
  - {name: "HK 01", type: ss, server: 127.0.0.2, port: 1, cipher: aes-128-gcm, password: x}
  - {name: "JP 01", type: ss, server: 127.0.0.3, port: 1, cipher: aes-128-gcm, password: x}
proxy-groups:
  - {name: Proxy, type: select, proxies: [DIRECT, "HK 01", "JP 01"]}
rules:
  - DOMAIN-SUFFIX,example.com,Proxy
  - MATCH,DIRECT
`

const otherProfile = `proxies: []
proxy-groups:
  - {name: Other, type: select, proxies: [DIRECT]}
rules:
  - MATCH,Other
`

type sim struct {
	t        *testing.T
	kumo     *Kumo
	platform *sysproxy.Memory
	paths    paths.Paths
	events   []Event
	mu       sync.Mutex
}

func coreBinary(t *testing.T) string {
	path := os.Getenv("KUMO_TEST_MIHOMO")
	if path == "" {
		t.Skip("set KUMO_TEST_MIHOMO to a Mihomo binary to run lifecycle simulations")
	}
	return path
}

func freePort(t *testing.T) int {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

// newSim sets up a Kumo with its own data directory, free ports and an
// in-memory system proxy.
func newSim(t *testing.T) *sim {
	t.Helper()
	binary := coreBinary(t)
	p := paths.Paths{Root: t.TempDir()}
	platform := sysproxy.NewMemory()
	s := &sim{t: t, platform: platform, paths: p}
	s.kumo = s.instance(core.OwnerApp)
	if err := p.Prepare(); err != nil {
		t.Fatal(err)
	}
	yaml := simProfile
	if err := s.kumo.deps.Profiles.Save(profile.Profile{ID: "sim", Name: "Sim", Kind: profile.KindLocal}, &yaml); err != nil {
		t.Fatal(err)
	}
	if err := s.kumo.deps.Profiles.SetCurrent("sim"); err != nil {
		t.Fatal(err)
	}
	mixed, controller := freePort(t), freePort(t)
	if _, err := s.kumo.deps.Runtime.Update(func(rt *core.Runtime) error {
		rt.CorePath = binary
		rt.Settings.MixedPort = mixed
		rt.Endpoint.Port = controller
		rt.SystemProxy.Settings.NetworkService = "Wi-Fi"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		report := s.kumo.Lifecycle.Shutdown(context.Background())
		if rt, _ := s.kumo.deps.Runtime.Load(); rt.PID > 0 && process.Alive(rt.PID) {
			process.Terminate(rt.PID, time.Second)
		}
		_ = report
	})
	return s
}

// instance is another Kumo on the same data, as another process (the CLI,
// or the app after a relaunch) would be.
func (s *sim) instance(owner core.Owner) *Kumo {
	deps := DefaultDependencies(s.paths, owner)
	deps.SystemProxy = s.platform
	deps.WatchInterval = 50 * time.Millisecond
	deps.ReadyTimeout = 10 * time.Second
	k := New(deps)
	k.Events.Subscribe(func(e Event) {
		s.mu.Lock()
		s.events = append(s.events, e)
		s.mu.Unlock()
	})
	return k
}

func (s *sim) start(opts StartOptions) Status {
	s.t.Helper()
	status, err := s.kumo.Lifecycle.Start(context.Background(), opts)
	if err != nil {
		s.t.Fatalf("start: %v (log: %s)", err, tail(s.paths.CoreLog(), 2000))
	}
	if !status.Running() || status.Readiness != core.ReadinessControllerReady {
		s.t.Fatalf("not ready after start: %+v", status)
	}
	return status
}

func (s *sim) observe() systemproxy.Snapshot {
	snap, _ := s.platform.Observe("Wi-Fi")
	return snap
}

func (s *sim) eventsOf(match func(Event) bool) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, e := range s.events {
		if match(e) {
			n++
		}
	}
	return n
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestSimStartSelectStop(t *testing.T) {
	s := newSim(t)
	status := s.start(StartOptions{})
	ctx := context.Background()

	groups, err := s.kumo.Proxies.Groups(ctx)
	if err != nil || len(groups) == 0 {
		rt, _ := s.kumo.deps.Runtime.Load()
		cfg, _ := os.ReadFile(s.paths.RuntimeConfig())
		t.Fatalf("groups: %v %v state=%s\n%s", groups, err, rt.State, cfg)
	}
	if err := s.kumo.Proxies.Select(ctx, "Proxy", "JP 01"); err != nil {
		t.Fatal(err)
	}
	groups, _ = s.kumo.Proxies.Groups(ctx)
	if groups[0].Name != "Proxy" || groups[0].Now != "JP 01" {
		t.Fatalf("selection not applied: %+v", groups[0])
	}
	if _, err := s.kumo.Settings.SetMode(ctx, core.ModeGlobal); err != nil {
		t.Fatal(err)
	}
	cfg, err := mihomo.NewClient(core.Endpoint{Host: "127.0.0.1", Port: portOf(status.Controller)}).Configs(ctx)
	if err != nil || cfg.Mode != "global" {
		t.Fatalf("mode not applied: %+v %v", cfg, err)
	}

	stopped, err := s.kumo.Lifecycle.Stop(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if stopped.State != core.StateStopped || process.Alive(status.PID) {
		t.Fatalf("still running after stop: %+v", stopped)
	}
	if !ports.Free("127.0.0.1", status.MixedPort) {
		t.Fatal("mixed port still held after stop")
	}
	if _, err := os.Stat(s.paths.CorePID()); !os.IsNotExist(err) {
		t.Fatal("pid file left behind")
	}
}

func portOf(address string) int {
	_, port, _ := net.SplitHostPort(address)
	n, _ := strconv.Atoi(port)
	return n
}

func TestSimSystemProxyFollowsCore(t *testing.T) {
	s := newSim(t)
	ctx := context.Background()
	if _, err := s.kumo.SystemProxy.SetEnabled(true); err != nil {
		t.Fatalf("enabling intent while stopped: %v", err)
	}
	if !s.observe().IsCleared() {
		t.Fatal("system proxy set before the core runs")
	}
	status := s.start(StartOptions{})
	snap := s.observe()
	if !snap.Web.PointsAt("127.0.0.1", status.MixedPort) || !status.SystemProxy.Active || status.SystemProxy.Ownership != systemproxy.Owned {
		t.Fatalf("system proxy not applied on start: %+v %+v", snap, status.SystemProxy)
	}
	stopped, err := s.kumo.Lifecycle.Stop(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !s.observe().IsCleared() {
		t.Fatal("system proxy left pointing at a stopped core")
	}
	if !stopped.SystemProxy.Enabled || stopped.SystemProxy.Active {
		t.Fatalf("intent lost or still active: %+v", stopped.SystemProxy)
	}
	s.start(StartOptions{})
	if s.observe().IsCleared() {
		t.Fatal("system proxy not restored on the next start")
	}
}

func TestSimSystemProxyContention(t *testing.T) {
	s := newSim(t)
	ctx := context.Background()
	_, _ = s.kumo.SystemProxy.SetEnabled(true)
	s.start(StartOptions{})

	// Another client points the system at itself.
	other, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	otherPort := other.Addr().(*net.TCPAddr).Port
	s.platform.Takeover("Wi-Fi", "127.0.0.1", otherPort)
	s.kumo.SystemProxy.Check()
	status, _ := s.kumo.Lifecycle.Status()
	if status.SystemProxy.Ownership != systemproxy.TakenOver {
		t.Fatalf("takeover not detected: %+v", status.SystemProxy)
	}
	if s.eventsOf(func(e Event) bool {
		c, ok := e.(SystemProxyContended)
		return ok && c.Ownership == systemproxy.TakenOver && c.Holder != ""
	}) != 1 {
		t.Fatal("no contention event naming the holder")
	}
	// A second check does not report it again.
	s.kumo.SystemProxy.Check()
	if s.eventsOf(func(e Event) bool { _, ok := e.(SystemProxyContended); return ok }) != 1 {
		t.Fatal("contention reported twice")
	}
	// Stopping Kumo must not clear the other client's proxy.
	if _, err := s.kumo.Lifecycle.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if !s.observe().Web.PointsAt("127.0.0.1", otherPort) {
		t.Fatal("Kumo cleared a proxy it did not own")
	}
}

func TestSimSystemProxyReclaim(t *testing.T) {
	s := newSim(t)
	_, _ = s.kumo.Preferences.Update(func(p *preferences.Preferences) { p.ReclaimSystemProxy = true })
	_, _ = s.kumo.SystemProxy.SetEnabled(true)
	status := s.start(StartOptions{})
	s.platform.Takeover("Wi-Fi", "127.0.0.1", 1)
	s.kumo.SystemProxy.Check()
	if !s.observe().Web.PointsAt("127.0.0.1", status.MixedPort) {
		t.Fatal("proxy not reclaimed")
	}
	if s.eventsOf(func(e Event) bool { c, ok := e.(SystemProxyContended); return ok && c.Reclaimed }) != 1 {
		t.Fatal("reclaim not reported")
	}
}

func TestSimSystemProxyTurnedOffElsewhere(t *testing.T) {
	s := newSim(t)
	_, _ = s.kumo.SystemProxy.SetEnabled(true)
	s.start(StartOptions{})
	_ = s.platform.Disable("Wi-Fi")
	s.kumo.SystemProxy.Check()
	status, _ := s.kumo.Lifecycle.Status()
	if status.SystemProxy.Enabled || status.SystemProxy.Active {
		t.Fatalf("a proxy the user turned off is still wanted: %+v", status.SystemProxy)
	}
}

func TestSimPortConflict(t *testing.T) {
	s := newSim(t)
	rt, _ := s.kumo.deps.Runtime.Load()
	squatter, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(rt.Settings.MixedPort))
	if err != nil {
		t.Fatal(err)
	}
	defer squatter.Close()

	_, err = s.kumo.Lifecycle.Start(context.Background(), StartOptions{})
	var conflicts core.PortConflicts
	if !errors.As(err, &conflicts) || len(conflicts) != 1 || conflicts[0].Role != "mixed" {
		t.Fatalf("expected a mixed port conflict, got %v", err)
	}
	if conflicts[0].PID != os.Getpid() || conflicts[0].Suggested == 0 {
		t.Fatalf("conflict does not name this process or a free port: %+v", conflicts[0])
	}
	after, _ := s.kumo.deps.Runtime.Load()
	if after.IsActive() {
		t.Fatal("a refused start left the runtime active")
	}

	status := s.start(StartOptions{ResolvePortConflicts: true})
	if status.MixedPort != conflicts[0].Suggested {
		t.Fatalf("did not move to the suggested port: %d", status.MixedPort)
	}
}

func TestSimCoreCrash(t *testing.T) {
	s := newSim(t)
	_, _ = s.kumo.SystemProxy.SetEnabled(true)
	status := s.start(StartOptions{})
	if err := syscall.Kill(status.PID, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the crash to be recorded", func() bool {
		return s.eventsOf(func(e Event) bool { _, ok := e.(CoreExited); return ok }) == 1
	})
	after, _ := s.kumo.Lifecycle.Status()
	if after.State != core.StateFailed || after.LastExit == nil || after.LastExit.Expected {
		t.Fatalf("crash not recorded: %+v", after)
	}
	if !s.observe().IsCleared() {
		t.Fatal("system proxy left pointing at a dead core")
	}
	// The next start works and sets the proxy again.
	s.start(StartOptions{})
	if s.observe().IsCleared() {
		t.Fatal("system proxy not restored after restarting a crashed core")
	}
}

func TestSimAdoptAndStopFromAnotherInstance(t *testing.T) {
	s := newSim(t)
	// The CLI starts the core and exits.
	cli := s.instance(core.OwnerCLI)
	status, err := cli.Lifecycle.Start(context.Background(), StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// The app finds it on launch.
	recovered, err := s.kumo.Lifecycle.Recover()
	if err != nil || recovered.PID != status.PID || !recovered.Running() || recovered.Owner != core.OwnerCLI {
		t.Fatalf("not adopted: %+v %v", recovered, err)
	}
	// A second start from the app is refused, not a second core.
	if _, err := s.kumo.Lifecycle.Start(context.Background(), StartOptions{}); !errors.Is(err, core.ErrAlreadyRunning) {
		t.Fatalf("second start: %v", err)
	}
	// Quitting the app stops the core it adopted.
	report := s.kumo.Lifecycle.Shutdown(context.Background())
	if len(report.Diagnostics) > 0 || process.Alive(status.PID) {
		t.Fatalf("shutdown left the core: %+v", report)
	}
}

func TestSimStalePIDIsNotKilled(t *testing.T) {
	s := newSim(t)
	sleeper := exec.Command("sleep", "30")
	if err := sleeper.Start(); err != nil {
		t.Fatal(err)
	}
	defer sleeper.Process.Kill()
	// A crash left a pid that now belongs to an unrelated process.
	_, _ = s.kumo.deps.Runtime.Update(func(rt *core.Runtime) error {
		rt.State, rt.PID = core.StateRunning, sleeper.Process.Pid
		return nil
	})
	status, err := s.kumo.Lifecycle.Status()
	if err != nil || status.State != core.StateFailed {
		t.Fatalf("stale pid adopted: %+v %v", status, err)
	}
	if _, err := s.kumo.Lifecycle.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !process.Alive(sleeper.Process.Pid) {
		t.Fatal("Kumo killed an unrelated process")
	}
}

func TestSimBrokenProfileFailsCleanly(t *testing.T) {
	s := newSim(t)
	broken := "proxy-groups:\n  - {name: G, type: select, proxies: [Nowhere]}\nrules:\n  - MATCH,G\n"
	_ = s.kumo.deps.Profiles.Save(profile.Profile{ID: "sim", Name: "Sim", Kind: profile.KindLocal}, &broken)
	status, err := s.kumo.Lifecycle.Start(context.Background(), StartOptions{})
	if err == nil {
		t.Fatal("broken profile started")
	}
	if status.State != core.StateFailed || status.Message == "" {
		t.Fatalf("failure not recorded: %+v", status)
	}
	rt, _ := s.kumo.deps.Runtime.Load()
	if rt.PID != 0 {
		t.Fatal("failed start left a pid")
	}
	t.Logf("failure message: %s", status.Message)
}

func TestSimSwitchProfileWhileRunning(t *testing.T) {
	s := newSim(t)
	ctx := context.Background()
	status := s.start(StartOptions{})
	yaml := otherProfile
	_ = s.kumo.deps.Profiles.Save(profile.Profile{ID: "other", Name: "Other", Kind: profile.KindLocal}, &yaml)
	if err := s.kumo.Profiles.Select(ctx, "other"); err != nil {
		t.Fatal(err)
	}
	groups, err := s.kumo.Proxies.Groups(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, g := range groups {
		found = found || g.Name == "Other"
	}
	if !found {
		t.Fatalf("running core did not switch profiles: %+v", groups)
	}
	after, _ := s.kumo.Lifecycle.Status()
	if after.PID != status.PID {
		t.Fatal("switching profiles restarted the core")
	}
	// A profile the core rejects leaves the running one in place.
	bad := "proxy-groups:\n  - {name: G, type: select, proxies: [Nowhere]}\nrules:\n  - MATCH,G\n"
	_ = s.kumo.deps.Profiles.Save(profile.Profile{ID: "bad", Name: "Bad", Kind: profile.KindLocal}, &bad)
	if err := s.kumo.Profiles.Select(ctx, "bad"); err == nil {
		t.Fatal("a rejected profile reported success")
	}
	if alive, _ := s.kumo.Lifecycle.Status(); !alive.Running() {
		t.Fatal("a rejected profile stopped the core")
	}
}

func TestSimConcurrentStarts(t *testing.T) {
	s := newSim(t)
	other := s.instance(core.OwnerCLI)
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, k := range []*Kumo{s.kumo, other} {
		wg.Add(1)
		go func(i int, k *Kumo) {
			defer wg.Done()
			_, errs[i] = k.Lifecycle.Start(context.Background(), StartOptions{})
		}(i, k)
	}
	wg.Wait()
	successes := 0
	for _, err := range errs {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one start to win, got %v", errs)
	}
	out, _ := exec.Command("pgrep", "-f", s.paths.Work()).Output()
	if n := len(splitNonEmpty(string(out))); n != 1 {
		t.Fatalf("%d cores run for one data directory", n)
	}
}

func splitNonEmpty(text string) []string {
	var out []string
	for _, line := range splitLines(text) {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func splitLines(text string) []string {
	var lines []string
	start := 0
	for i := range len(text) {
		if text[i] == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	return append(lines, text[start:])
}

func TestSimShutdownReleasesEverything(t *testing.T) {
	s := newSim(t)
	_, _ = s.kumo.SystemProxy.SetEnabled(true)
	status := s.start(StartOptions{})
	report := s.kumo.Lifecycle.Shutdown(context.Background())
	if len(report.Diagnostics) > 0 {
		t.Fatalf("shutdown diagnostics: %v", report.Diagnostics)
	}
	if process.Alive(status.PID) || !ports.Free("127.0.0.1", status.MixedPort) || !s.observe().IsCleared() {
		t.Fatal("shutdown left the core, its port or the system proxy behind")
	}
	if report.Status.State != core.StateStopped {
		t.Fatalf("state after shutdown: %+v", report.Status)
	}
}
