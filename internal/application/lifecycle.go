package application

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/override"
	"github.com/AstraFoundry/KumoApp/internal/domain/profile"
	"github.com/AstraFoundry/KumoApp/internal/domain/runtimeconfig"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/mihomo"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/ports"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/process"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/storage"
)

// Status is the runtime as the surfaces show it.
type Status struct {
	State     core.RunState  `json:"state"`
	PID       int            `json:"pid,omitempty"`
	Owner     core.Owner     `json:"owner,omitempty"`
	Readiness core.Readiness `json:"readiness,omitempty"`
	StartedAt *time.Time     `json:"startedAt,omitempty"`
	Message   string         `json:"message,omitempty"`
	Mode      core.Mode      `json:"mode"`
	// Controller is the external controller's address.
	Controller  string            `json:"controller"`
	MixedPort   int               `json:"mixedPort"`
	CorePath    string            `json:"corePath,omitempty"`
	TUN         bool              `json:"tun"`
	SystemProxy SystemProxyStatus `json:"systemProxy"`
	Profile     profile.Summary   `json:"profile"`
	LastExit    *core.Exit        `json:"lastExit,omitempty"`
}

// Running reports whether the core is up.
func (s Status) Running() bool { return s.State == core.StateRunning }

// StartOptions tune one start.
type StartOptions struct {
	// CorePath overrides the executable for this start.
	CorePath string `json:"corePath,omitempty"`
	// ResolvePortConflicts moves Kumo to free ports instead of failing when
	// another process holds the mixed or controller port.
	ResolvePortConflicts bool `json:"resolvePortConflicts,omitempty"`
}

// Lifecycle starts, stops and watches the core.
type Lifecycle struct {
	deps   Dependencies
	events *Bus
	proxy  *SystemProxyService

	mu sync.Mutex
	// proc is the core this process launched; nil when another surface
	// launched it or it runs no core.
	proc *mihomo.Process
	// stopping marks an exit Kumo asked for.
	stopping bool
	// starting is set while this process starts the core.
	starting bool
	last     Status
}

// Status reconciles the stored runtime with the processes that actually
// run, and describes the result.
func (l *Lifecycle) Status() (Status, error) {
	rt, err := l.reconcile()
	if err != nil {
		return Status{}, err
	}
	return l.describe(rt), nil
}

func (l *Lifecycle) describe(rt core.Runtime) Status {
	s := Status{
		State: rt.State, PID: rt.PID, Owner: rt.Owner, Readiness: rt.Readiness, StartedAt: rt.StartedAt,
		Message: rt.Message, Mode: rt.Mode, Controller: rt.Endpoint.Address(), MixedPort: rt.Settings.MixedPort,
		CorePath: rt.CorePath, TUN: rt.Settings.TUN.Enabled, LastExit: rt.LastExit,
		SystemProxy: l.proxy.describe(rt),
	}
	if id, err := l.deps.Profiles.CurrentID(); err == nil {
		if p, err := l.deps.Profiles.Get(id); err == nil {
			s.Profile = p.Summarize(id)
		}
	}
	return s
}

// publish sends the status when it changed since the last one sent.
func (l *Lifecycle) publish(s Status) {
	l.mu.Lock()
	same := statusKey(l.last) == statusKey(s)
	l.last = s
	l.mu.Unlock()
	if !same {
		l.events.Publish(StatusChanged{Status: s})
	}
}

func statusKey(s Status) string {
	return fmt.Sprint(s.State, s.PID, s.Readiness, s.Mode, s.MixedPort, s.TUN, s.Message, s.Profile.ID, s.Profile.Name,
		s.SystemProxy.Enabled, s.SystemProxy.Active, s.SystemProxy.Ownership, s.SystemProxy.Settings.Mode, s.LastExit != nil)
}

// Refresh reconciles and publishes the status if it changed.
func (l *Lifecycle) Refresh() (Status, error) {
	s, err := l.Status()
	if err == nil {
		l.publish(s)
	}
	return s, err
}

// reconcile finds the core's real state: a pid that still runs Mihomo is
// adopted, one that does not any more is cleared, and a session whose core
// vanished is recorded as an unexpected exit.
func (l *Lifecycle) reconcile() (core.Runtime, error) {
	return l.deps.Runtime.Update(func(rt *core.Runtime) error {
		pids := []int{}
		if rt.PID > 0 {
			pids = append(pids, rt.PID)
		}
		if pid := l.readPID(); pid > 0 && pid != rt.PID {
			pids = append(pids, pid)
		}
		for _, pid := range pids {
			if process.Alive(pid) && process.IsCore(pid, rt.CorePath) {
				if rt.State != core.StateRunning || rt.PID != pid {
					rt.Launched(pid, l.deps.Now())
					rt.Message = fmt.Sprintf("Recovered Mihomo running with pid %d.", pid)
				}
				return nil
			}
		}
		if rt.State == core.StateRunning || rt.State == core.StateStopping {
			l.mu.Lock()
			ours := l.proc != nil
			l.mu.Unlock()
			if !ours {
				rt.Exited(core.Exit{At: l.deps.Now(), Code: -1, Message: l.exitMessage("Mihomo is no longer running."), Expected: false})
			}
		}
		if rt.State == core.StateStarting && rt.PID == 0 {
			l.mu.Lock()
			inFlight := l.starting
			l.mu.Unlock()
			stale := rt.StartedAt == nil || l.deps.Now().Sub(*rt.StartedAt) > l.deps.ReadyTimeout+5*time.Second
			if !inFlight && stale {
				// A start that never launched, e.g. its surface died midway.
				rt.Stopped("Mihomo is not running.")
			}
		}
		if len(pids) > 0 {
			_ = os.Remove(l.deps.Paths.CorePID())
		}
		return nil
	})
}

func (l *Lifecycle) readPID() int {
	data, err := os.ReadFile(l.deps.Paths.CorePID())
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return pid
}

// Start launches the core with the current profile and waits until its
// controller answers.
func (l *Lifecycle) Start(ctx context.Context, opts StartOptions) (Status, error) {
	if _, err := l.reconcile(); err != nil {
		return Status{}, err
	}
	rt, err := l.deps.Runtime.Load()
	if err != nil {
		return Status{}, err
	}
	if rt.IsActive() {
		return l.describe(rt), fmt.Errorf("%w (pid %d)", core.ErrAlreadyRunning, rt.PID)
	}
	if err := l.deps.Paths.Prepare(); err != nil {
		return Status{}, err
	}
	corePath, err := l.resolveCore(opts.CorePath, rt.CorePath)
	if err != nil {
		return l.fail(err)
	}
	if err := l.checkPorts(&rt, opts.ResolvePortConflicts); err != nil {
		return l.describe(rt), err
	}
	if err := l.writeConfig(rt); err != nil {
		return l.fail(err)
	}
	l.mu.Lock()
	l.starting = true
	l.mu.Unlock()
	defer func() {
		l.mu.Lock()
		l.starting = false
		l.mu.Unlock()
	}()
	rt, err = l.deps.Runtime.Update(func(rt *core.Runtime) error { return rt.BeginStart(l.deps.Owner, corePath, l.deps.Now()) })
	if err != nil {
		return l.describe(rt), err
	}
	l.publish(l.describe(rt))

	proc, err := mihomo.Start(mihomo.Launch{
		CorePath: corePath, WorkDir: l.deps.Paths.Work(), ConfigPath: l.deps.Paths.RuntimeConfig(),
		LogPath: l.deps.Paths.CoreLog(), Endpoint: rt.Endpoint,
	})
	if err != nil {
		return l.fail(fmt.Errorf("could not launch %s: %w", corePath, err))
	}
	l.mu.Lock()
	l.proc, l.stopping = proc, false
	l.mu.Unlock()
	_ = storage.WriteFile(l.deps.Paths.CorePID(), []byte(strconv.Itoa(proc.PID)+"\n"))
	rt, err = l.deps.Runtime.Update(func(rt *core.Runtime) error {
		rt.Launched(proc.PID, l.deps.Now())
		return nil
	})
	if err != nil {
		return l.describe(rt), err
	}
	go l.watchExit(proc)

	if err := l.waitReady(ctx, proc, rt.Endpoint); err != nil {
		l.mu.Lock()
		l.stopping = true
		l.mu.Unlock()
		process.Terminate(proc.PID, 2*time.Second)
		<-proc.Done()
		_ = os.Remove(l.deps.Paths.CorePID())
		l.mu.Lock()
		l.proc = nil
		l.mu.Unlock()
		return l.fail(err)
	}
	rt, err = l.deps.Runtime.Update(func(rt *core.Runtime) error {
		rt.Ready()
		return nil
	})
	if err != nil {
		return l.describe(rt), err
	}
	if rt.SystemProxy.Enabled {
		if _, err := l.proxy.Apply(); err != nil {
			rt, _ = l.deps.Runtime.Update(func(rt *core.Runtime) error {
				rt.Message = "Mihomo is running, but the system proxy could not be set: " + err.Error()
				return nil
			})
		}
	}
	s, err := l.Status()
	l.publish(s)
	return s, err
}

func (l *Lifecycle) fail(cause error) (Status, error) {
	rt, err := l.deps.Runtime.Update(func(rt *core.Runtime) error {
		rt.Fail(cause.Error())
		return nil
	})
	if err != nil {
		return Status{}, errors.Join(cause, err)
	}
	s := l.describe(rt)
	l.publish(s)
	return s, cause
}

func (l *Lifecycle) resolveCore(override, stored string) (string, error) {
	if override != "" {
		if _, err := os.Stat(override); err != nil {
			return "", fmt.Errorf("%w at %s", core.ErrCoreNotFound, override)
		}
		return override, nil
	}
	candidates := mihomo.Discover(stored, l.deps.Paths.ManagedCore(), l.deps.BundledDir)
	if len(candidates) == 0 {
		return "", fmt.Errorf("%w: install it from Core, or run `kumo core install`", core.ErrCoreNotFound)
	}
	return candidates[0].Path, nil
}

// checkPorts makes sure the mixed and controller ports are free. A port
// held by another process is a conflict: reported, or moved past when the
// caller allows it.
func (l *Lifecycle) checkPorts(rt *core.Runtime, resolve bool) error {
	var conflicts core.PortConflicts
	check := func(role, host string, port int, avoid int) {
		if ports.Free(host, port) {
			return
		}
		conflict := &core.PortConflict{Role: role, Port: port, Suggested: ports.Suggest(host, port, avoid)}
		if holder, ok := ports.Find(port); ok {
			conflict.PID, conflict.Process = holder.PID, holder.Process
		}
		conflicts = append(conflicts, conflict)
	}
	bind := "127.0.0.1"
	if rt.Settings.AllowLAN {
		bind = "0.0.0.0"
	}
	check("mixed", bind, rt.Settings.MixedPort, rt.Endpoint.Port)
	check("controller", rt.Endpoint.Host, rt.Endpoint.Port, rt.Settings.MixedPort)
	if len(conflicts) == 0 {
		return nil
	}
	if !resolve {
		return conflicts
	}
	updated, err := l.deps.Runtime.Update(func(stored *core.Runtime) error {
		for _, c := range conflicts {
			if c.Suggested == 0 {
				return c
			}
			if c.Role == "mixed" {
				stored.Settings.MixedPort = c.Suggested
			} else {
				stored.Endpoint.Port = c.Suggested
			}
		}
		stored.Message = "Moved to free ports: " + conflicts.Error()
		return nil
	})
	if err != nil {
		return err
	}
	*rt = updated
	return nil
}

// RuntimeConfig builds the configuration the core would run with now.
func (l *Lifecycle) RuntimeConfig(rt core.Runtime) (string, error) {
	id, err := l.deps.Profiles.CurrentID()
	if err != nil {
		return "", err
	}
	yaml, err := l.deps.Profiles.Content(id)
	if err != nil {
		return "", err
	}
	sources, err := l.overrideSources()
	if err != nil {
		return "", err
	}
	return runtimeconfig.Build(runtimeconfig.Input{
		ProfileYAML: yaml, Overrides: sources, Mode: rt.Mode, Endpoint: rt.Endpoint,
		Settings: rt.Settings, Transformer: l.deps.Transformer,
	})
}

func (l *Lifecycle) overrideSources() ([]override.Source, error) {
	items, err := l.deps.Overrides.List()
	if err != nil {
		return nil, err
	}
	var sources []override.Source
	for _, item := range items {
		if item.Disabled {
			continue
		}
		content, err := l.deps.Overrides.Content(item.ID)
		if err != nil {
			return nil, err
		}
		sources = append(sources, override.Source{Name: item.Name, Format: item.Format, Content: content})
	}
	return sources, nil
}

func (l *Lifecycle) writeConfig(rt core.Runtime) error {
	config, err := l.RuntimeConfig(rt)
	if err != nil {
		return err
	}
	return storage.WriteFile(l.deps.Paths.RuntimeConfig(), []byte(config))
}

func (l *Lifecycle) waitReady(ctx context.Context, proc *mihomo.Process, endpoint core.Endpoint) error {
	client := mihomo.NewClient(endpoint)
	deadline := time.Now().Add(l.deps.ReadyTimeout)
	for {
		// The controller answers before the configuration is applied: the
		// core is ready once its GLOBAL group exists.
		probe, cancel := context.WithTimeout(ctx, time.Second)
		groups, err := client.Groups(probe)
		cancel()
		if err == nil && len(groups) > 0 {
			return nil
		}
		select {
		case <-proc.Done():
			return errors.New(l.exitMessage(fmt.Sprintf("Mihomo exited with code %d while starting.", proc.ExitCode())))
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
		if time.Now().After(deadline) {
			return errors.New(l.exitMessage("Mihomo did not answer on " + endpoint.Address() + " in time."))
		}
	}
}

// watchExit records a core that ends without Kumo stopping it, and lets go
// of the system proxy so traffic is not sent to a dead port.
func (l *Lifecycle) watchExit(proc *mihomo.Process) {
	<-proc.Done()
	l.mu.Lock()
	expected := l.stopping
	if l.proc == proc {
		l.proc = nil
	}
	l.mu.Unlock()
	if expected {
		return
	}
	exit := core.Exit{At: l.deps.Now(), Code: proc.ExitCode(), Expected: false,
		Message: l.exitMessage(fmt.Sprintf("Mihomo exited unexpectedly with code %d.", proc.ExitCode()))}
	_, _ = l.deps.Runtime.Update(func(rt *core.Runtime) error {
		if rt.PID == proc.PID {
			rt.Exited(exit)
		}
		return nil
	})
	_ = os.Remove(l.deps.Paths.CorePID())
	_ = l.proxy.Release()
	l.events.Publish(CoreExited{Exit: exit})
	if s, err := l.Status(); err == nil {
		l.publish(s)
	}
}

// exitMessage adds the last error the core logged to message.
func (l *Lifecycle) exitMessage(message string) string {
	if line := lastErrorLine(l.deps.Paths.CoreLog()); line != "" {
		return message + " " + line
	}
	return message
}

// Stop lets go of the system proxy, then ends the core and waits for its
// ports to free up. The user's wish for the system proxy is kept, so the
// next start sets it again.
func (l *Lifecycle) Stop(ctx context.Context) (Status, error) {
	rt, err := l.reconcile()
	if err != nil {
		return Status{}, err
	}
	if !rt.IsActive() {
		rt, err = l.deps.Runtime.Update(func(rt *core.Runtime) error {
			if rt.State != core.StateFailed {
				rt.Stopped("Mihomo is not running.")
			}
			return nil
		})
		_ = l.proxy.Release()
		s := l.describe(rt)
		l.publish(s)
		return s, err
	}
	releaseErr := l.proxy.Release()
	if err := l.terminate(rt); err != nil {
		return l.fail(err)
	}
	rt, err = l.deps.Runtime.Update(func(rt *core.Runtime) error {
		rt.Stopped("Mihomo stopped.")
		return nil
	})
	if err != nil {
		return Status{}, err
	}
	l.waitPortsFree(ctx, rt)
	s, err := l.Status()
	l.publish(s)
	if releaseErr != nil {
		return s, fmt.Errorf("mihomo stopped, but the system proxy could not be released: %w", releaseErr)
	}
	return s, err
}

func (l *Lifecycle) terminate(rt core.Runtime) error {
	l.mu.Lock()
	l.stopping = true
	proc := l.proc
	l.mu.Unlock()
	pid := rt.PID
	if pid == 0 && proc != nil {
		pid = proc.PID
	}
	if pid > 0 && process.Alive(pid) && !process.IsCore(pid, rt.CorePath) {
		// The pid now belongs to something else; the core is long gone.
		pid = 0
	}
	if pid > 0 && !process.Terminate(pid, 3*time.Second) {
		return fmt.Errorf("mihomo (pid %d) did not exit", pid)
	}
	if proc != nil {
		select {
		case <-proc.Done():
		case <-time.After(3 * time.Second):
		}
	}
	l.mu.Lock()
	if l.proc == proc {
		l.proc = nil
	}
	l.mu.Unlock()
	_ = os.Remove(l.deps.Paths.CorePID())
	return nil
}

func (l *Lifecycle) waitPortsFree(ctx context.Context, rt core.Runtime) {
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && ctx.Err() == nil {
		if ports.Free("127.0.0.1", rt.Settings.MixedPort) && ports.Free(rt.Endpoint.Host, rt.Endpoint.Port) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Restart stops and starts the core again.
func (l *Lifecycle) Restart(ctx context.Context, opts StartOptions) (Status, error) {
	if _, err := l.Stop(ctx); err != nil {
		return Status{}, err
	}
	return l.Start(ctx, opts)
}

// Reload rebuilds the configuration and has the running core read it,
// keeping its process and connections. When the core is off, it only
// writes the file for the next start.
func (l *Lifecycle) Reload(ctx context.Context) error {
	rt, err := l.reconcile()
	if err != nil {
		return err
	}
	previous, _ := os.ReadFile(l.deps.Paths.RuntimeConfig())
	if err := l.writeConfig(rt); err != nil {
		return err
	}
	if rt.State != core.StateRunning {
		return nil
	}
	if err := mihomo.NewClient(rt.Endpoint).Reload(ctx, l.deps.Paths.RuntimeConfig()); err != nil {
		if previous != nil {
			_ = storage.WriteFile(l.deps.Paths.RuntimeConfig(), previous)
		}
		return err
	}
	if rt.SystemProxy.Applied != nil && rt.SystemProxy.Applied.Port != rt.Settings.MixedPort {
		_, err = l.proxy.Apply()
	}
	if s, statusErr := l.Status(); statusErr == nil {
		l.publish(s)
	}
	return err
}

// ShutdownReport says what a shutdown did and what failed along the way.
type ShutdownReport struct {
	Status      Status   `json:"status"`
	Diagnostics []string `json:"diagnostics"`
}

// Shutdown is the quit path: release the system proxy and stop the core,
// whichever surface started it. Every step runs even when an earlier one
// failed, and failures are reported rather than returned, so the app can
// always quit.
func (l *Lifecycle) Shutdown(ctx context.Context) ShutdownReport {
	report := ShutdownReport{Diagnostics: []string{}}
	note := func(stage string, err error) {
		if err != nil {
			report.Diagnostics = append(report.Diagnostics, stage+": "+err.Error())
		}
	}
	note("system proxy", l.proxy.Release())
	l.proxy.pac.Stop()
	rt, err := l.reconcile()
	note("status", err)
	if rt.IsActive() || l.readPID() > 0 {
		note("stop", l.terminate(rt))
		_, err = l.deps.Runtime.Update(func(rt *core.Runtime) error {
			rt.Stopped("Mihomo stopped when Kumo quit.")
			return nil
		})
		note("state", err)
		l.waitPortsFree(ctx, rt)
	}
	report.Status, err = l.Status()
	note("status", err)
	return report
}

// Monitor keeps the status current while ctx lives: it notices cores that
// other surfaces start or stop, and system proxy changes by other apps.
func (l *Lifecycle) Monitor(ctx context.Context) {
	ticker := time.NewTicker(l.deps.WatchInterval)
	defer ticker.Stop()
	for {
		_, _ = l.Refresh()
		l.proxy.Check()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
