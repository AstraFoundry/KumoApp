package desktop

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"

	"github.com/AstraFoundry/KumoApp/internal/application"
	"github.com/AstraFoundry/KumoApp/internal/domain/preferences"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/process"
	"github.com/AstraFoundry/KumoApp/locales"
)

// GuardianEnv names the parent a guardian process watches.
const GuardianEnv = "KUMO_GUARDIAN_PARENT"

func bg() context.Context { return context.Background() }

// Shell is the app around the pages: its window, tray, background work and
// the quit path that always leaves the system as it found it.
type Shell struct {
	k *application.Kumo

	mu     sync.Mutex
	window *mygo.Window
	tray   *Tray
	route  string
	// locale is the language preference last applied, nil before any.
	locale *string

	cancel   context.CancelFunc
	shutdown atomic.Bool
	quitting atomic.Bool
}

// Run runs the app until it quits.
func Run(k *application.Kumo) error {
	if !mygo.App.RequestSingleInstanceLock() {
		return nil
	}
	s := &Shell{k: k}
	Bind(k, s.applyPreferences)

	mygo.App.OnSecondInstance(func([]string, string) { s.Open("") })
	mygo.App.OnOpenURL(s.openURL)
	mygo.App.OnActivate(func(hasVisibleWindows bool) {
		if !hasVisibleWindows {
			s.Open("")
		}
	})
	mygo.App.OnWindowAllClosed(func() {
		prefs, _ := k.Preferences.Get()
		if prefs.QuitOnLastWindowClose || s.tray == nil || s.tray.tray == nil {
			mygo.App.Quit()
		}
	})
	mygo.App.OnBeforeQuit(s.beforeQuit)
	mygo.App.WhenReady(s.ready)
	return mygo.App.Run()
}

func (s *Shell) ready() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	if _, err := s.k.Lifecycle.Recover(); err != nil {
		log.Println("recover:", err)
	}
	spawnGuardian()
	s.startDriver()
	prefs, _ := s.k.Preferences.Get()
	if runtime.GOOS == "darwin" && prefs.HideMenuBarIcon {
		// Without the menu bar item, the Dock icon is how to reach Kumo.
	} else {
		s.tray = NewTray(s.k, s.Open)
	}
	s.k.Events.Subscribe(s.onEvent)
	if !(prefs.LaunchAtLogin && mygo.App.WasOpenedAtLogin()) || !prefs.CompletedSetup {
		s.Open("")
	}
	go s.k.Lifecycle.Monitor(ctx)
	go s.k.Inspect.StreamTraffic(ctx)
	go s.refreshProfiles(ctx)
	if prefs.AutoCheckUpdates && mygo.Updater.Enabled() {
		go s.checkUpdate(ctx)
	}
	if prefs.StartCoreOnLaunch && prefs.CompletedSetup {
		go func() {
			status, _ := s.k.Lifecycle.Status()
			if !status.Running() {
				_, _ = s.k.Lifecycle.Start(ctx, application.StartOptions{})
			}
		}()
	}
}

// Open shows the main window at route ("" keeps the page where it was).
func (s *Shell) Open(route string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.window != nil {
		s.window.Show()
		s.window.Focus()
		if route != "" {
			_ = NavigateEvent.Emit(s.window, route)
		}
		return
	}
	prefs, _ := s.k.Preferences.Get()
	start := "/"
	if !prefs.CompletedSetup {
		start = "/setup"
	} else if route != "" {
		start = route
	}
	options := mygo.WindowOptions{
		Title:           "Kumo",
		URL:             start,
		Width:           1180,
		Height:          780,
		MinWidth:        940,
		MinHeight:       620,
		BackgroundColor: "light-dark(#f7f7f8, #161618)",
		StateKey:        "main",
	}
	// On macOS the content runs under the title bar, with the traffic
	// lights over the sidebar. Elsewhere the system's title bar keeps its
	// own controls, so page headers never have to make room for them.
	if runtime.GOOS == "darwin" {
		options.TitleBarStyle = mygo.TitleBarHidden
		options.TrafficLightPosition = &mygo.Point{X: 18, Y: 20}
	}
	s.window = mygo.NewWindow(options)
	s.window.OnClosed(func() {
		s.mu.Lock()
		s.window = nil
		s.mu.Unlock()
	})
}

func (s *Shell) onEvent(e application.Event) {
	switch e := e.(type) {
	case application.StatusChanged:
		if s.tray != nil {
			s.tray.Update(e.Status)
		}
	case application.CoreExited:
		notify(locales.NotifyCoreExitedTitle(), e.Exit.Message)
	case application.SystemProxyContended:
		switch {
		case e.Reclaimed:
			notify(locales.NotifyReclaimedTitle(), "")
		case e.Ownership == "takenOver":
			holder := e.Holder
			if holder == "" {
				holder = "?"
			}
			notify(locales.NotifyContentionTitle(), locales.NotifyContentionBody(holder))
		}
	}
}

func notify(title, body string) {
	if !mygo.NotificationsSupported() {
		return
	}
	go func() { _ = mygo.NewNotification(mygo.NotificationOptions{Title: title, Body: body}).Show() }()
}

// checkUpdate tells about a new version once, shortly after launch.
func (s *Shell) checkUpdate(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
	}
	up, err := mygo.Updater.Check(ctx)
	if err != nil || up == nil {
		return
	}
	notify(locales.NotifyUpdateTitle(up.Version), "")
}

// refreshProfiles refreshes subscriptions whose interval passed.
func (s *Shell) refreshProfiles(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute):
		case <-ticker.C:
		}
		if _, err := s.k.Profiles.RefreshDue(ctx); err != nil {
			log.Println("profile refresh:", err)
		}
	}
}

func (s *Shell) applyPreferences(p preferences.Preferences) {
	if p.LaunchAtLogin != mygo.App.OpenAtLogin() {
		_ = mygo.App.SetOpenAtLogin(p.LaunchAtLogin)
	}
	// "" follows the system's language.
	if s.locale == nil || *s.locale != p.Language {
		language := p.Language
		s.locale = &language
		mygo.App.SetLocale(language)
	}
	if status, err := s.k.Lifecycle.Status(); err == nil && s.tray != nil {
		s.tray.mu.Lock()
		s.tray.last = ""
		s.tray.mu.Unlock()
		s.tray.Update(status)
	}
}

// openURL imports a profile from clash://install-config?url=… or the same
// on kumo://.
func (s *Shell) openURL(raw string) {
	u, err := url.Parse(raw)
	if err != nil || u.Host != "install-config" {
		s.Open("")
		return
	}
	target := u.Query().Get("url")
	if target == "" {
		s.Open("")
		return
	}
	s.Open("/profiles?import=" + url.QueryEscape(target))
}

// beforeQuit holds the quit until Kumo has released the system proxy and
// stopped the core, then lets it finish. Every quit takes this path: the
// menu, the tray, Cmd+Q and SIGINT/SIGTERM. A crash or SIGKILL does not;
// the guardian covers those.
func (s *Shell) beforeQuit(e *mygo.QuitEvent) {
	if s.shutdown.Load() {
		return
	}
	e.PreventDefault()
	if !s.quitting.CompareAndSwap(false, true) {
		return
	}
	QuittingEvent.Broadcast(true)
	go func() {
		if s.cancel != nil {
			s.cancel()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		report := s.k.Lifecycle.Shutdown(ctx)
		for _, d := range report.Diagnostics {
			log.Println("shutdown:", d)
		}
		s.shutdown.Store(true)
		mygo.App.Quit()
	}()
}

// spawnGuardian starts a copy of this executable that waits for the app to
// exit and cleans up if it did not quit properly.
func spawnGuardian() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(exe)
	cmd.Env = append(os.Environ(), GuardianEnv+"="+strconv.Itoa(os.Getpid()))
	cmd.SysProcAttr = process.Detach()
	if err := cmd.Start(); err != nil {
		log.Println("guardian:", err)
		return
	}
	go cmd.Wait()
}
