package application

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/pac"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/ports"
)

// SystemProxyStatus is the system proxy as the surfaces show it.
type SystemProxyStatus struct {
	// Enabled is what the user asked for; it survives restarts.
	Enabled bool `json:"enabled"`
	// Active is true while Kumo's settings are in the system.
	Active    bool                  `json:"active"`
	Ownership systemproxy.Ownership `json:"ownership,omitempty"`
	Settings  systemproxy.Settings  `json:"settings"`
	Applied   *systemproxy.Applied  `json:"applied,omitempty"`
}

// SystemProxyService points the operating system at the core, and notices
// when another app points it elsewhere.
type SystemProxyService struct {
	deps   Dependencies
	events *Bus
	pac    *pac.Server
	status func() (Status, error)
	mu     sync.Mutex
}

func (s *SystemProxyService) describe(rt core.Runtime) SystemProxyStatus {
	return SystemProxyStatus{
		Enabled: rt.SystemProxy.Enabled, Active: rt.SystemProxy.Applied != nil,
		Ownership: rt.SystemProxy.Ownership, Settings: rt.SystemProxy.Settings, Applied: rt.SystemProxy.Applied,
	}
}

// Services lists the network services the proxy can be set on.
func (s *SystemProxyService) Services() ([]string, error) { return s.deps.SystemProxy.Services() }

func (s *SystemProxyService) service(settings systemproxy.Settings) (string, error) {
	service := strings.TrimSpace(settings.NetworkService)
	if service == "" || service == systemproxy.AutomaticService {
		return s.deps.SystemProxy.ActiveService()
	}
	return service, nil
}

// SetEnabled records the user's wish and acts on it: with the core
// running, Kumo's settings go into the system; turned off, they come out.
func (s *SystemProxyService) SetEnabled(enabled bool) (Status, error) {
	rt, err := s.deps.Runtime.Update(func(rt *core.Runtime) error {
		rt.SystemProxy.Enabled = enabled
		return nil
	})
	if err != nil {
		return Status{}, err
	}
	if enabled && rt.State == core.StateRunning {
		_, err = s.Apply()
	} else if !enabled {
		err = s.Release()
	}
	status, statusErr := s.status()
	return status, errors.Join(err, statusErr)
}

// Apply writes Kumo's settings into the system and checks that they took.
func (s *SystemProxyService) Apply() (SystemProxyStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rt, err := s.deps.Runtime.Load()
	if err != nil {
		return SystemProxyStatus{}, err
	}
	settings := rt.SystemProxy.Settings
	settings.Port = rt.Settings.MixedPort
	if err := settings.Validate(); err != nil {
		return s.describe(rt), err
	}
	if settings.Mode == systemproxy.ModePAC && s.deps.Owner == core.OwnerCLI {
		return s.describe(rt), errors.New("PAC mode is served by the Kumo app: open Kumo, or switch the system proxy to manual")
	}
	if !ports.Accepting(settings.Host, settings.Port) {
		return s.describe(rt), fmt.Errorf("nothing accepts connections on %s:%d; start Kumo first", settings.Host, settings.Port)
	}
	service, err := s.service(settings)
	if err != nil {
		return s.describe(rt), err
	}
	before, observeErr := s.deps.SystemProxy.Observe(service)
	applied := systemproxy.Applied{Service: service, Mode: settings.Mode, Host: settings.Host, Port: settings.Port, At: s.deps.Now()}
	switch settings.Mode {
	case systemproxy.ModePAC:
		url, err := s.pac.Serve(systemproxy.RenderPAC(settings.PACScript, settings.Port))
		if err != nil {
			return s.describe(rt), err
		}
		applied.PACURL = url
		err = s.deps.SystemProxy.ApplyPAC(service, url)
		if err != nil {
			return s.describe(rt), err
		}
	default:
		s.pac.Stop()
		if err := s.deps.SystemProxy.ApplyManual(service, settings.Host, settings.Port, settings.BypassList); err != nil {
			return s.describe(rt), err
		}
	}
	after, err := s.deps.SystemProxy.Observe(service)
	if err != nil {
		return s.describe(rt), err
	}
	if ownership := applied.OwnershipOf(after); ownership != systemproxy.Owned {
		return s.describe(rt), fmt.Errorf("the system did not take Kumo's proxy settings on %s", service)
	}
	rt, err = s.deps.Runtime.Update(func(rt *core.Runtime) error {
		if observeErr == nil && (rt.SystemProxy.Applied == nil || rt.SystemProxy.Applied.OwnershipOf(before) != systemproxy.Owned) && !before.IsCleared() {
			rt.SystemProxy.Previous = &before
		}
		rt.SystemProxy.Applied = &applied
		rt.SystemProxy.Ownership = systemproxy.Owned
		rt.SystemProxy.Settings.Port = settings.Port
		return nil
	})
	return s.describe(rt), err
}

// Release takes Kumo's settings out of the system, if they are still
// there. Settings another app put in since are left alone: Kumo never
// clears a proxy it does not own.
func (s *SystemProxyService) Release() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rt, err := s.deps.Runtime.Load()
	if err != nil {
		return err
	}
	applied := rt.SystemProxy.Applied
	if applied == nil {
		return nil
	}
	var releaseErr error
	observed, err := s.deps.SystemProxy.Observe(applied.Service)
	if err != nil || applied.OwnershipOf(observed) == systemproxy.Owned {
		releaseErr = s.deps.SystemProxy.Disable(applied.Service)
	}
	if applied.Mode == systemproxy.ModePAC {
		s.pac.Stop()
	}
	if releaseErr != nil {
		return releaseErr
	}
	_, err = s.deps.Runtime.Update(func(rt *core.Runtime) error {
		rt.SystemProxy.Applied = nil
		rt.SystemProxy.Ownership = ""
		return nil
	})
	return err
}

// Check looks whether the system still shows Kumo's settings. Another
// app's proxy is reported (and replaced, when the user asked Kumo to
// reclaim it); a proxy someone turned off stays off.
func (s *SystemProxyService) Check() {
	rt, err := s.deps.Runtime.Load()
	if err != nil || rt.SystemProxy.Applied == nil {
		return
	}
	applied := *rt.SystemProxy.Applied
	observed, err := s.deps.SystemProxy.Observe(applied.Service)
	if err != nil {
		return
	}
	ownership := applied.OwnershipOf(observed)
	if ownership == rt.SystemProxy.Ownership {
		return
	}
	event := SystemProxyContended{Ownership: ownership, Observed: observed}
	switch ownership {
	case systemproxy.TakenOver:
		event.Holder = holder(observed)
		prefs, _ := s.deps.Preferences.Load()
		if prefs.ReclaimSystemProxy && rt.State == core.StateRunning {
			if _, err := s.Apply(); err == nil {
				event.Reclaimed = true
				ownership = systemproxy.Owned
			}
		}
	case systemproxy.Cleared:
		// Someone turned the proxy off, in System Settings or another app.
		_, _ = s.deps.Runtime.Update(func(rt *core.Runtime) error {
			rt.SystemProxy.Enabled = false
			rt.SystemProxy.Applied = nil
			return nil
		})
	}
	_, _ = s.deps.Runtime.Update(func(rt *core.Runtime) error {
		if rt.SystemProxy.Applied != nil {
			rt.SystemProxy.Ownership = ownership
		}
		return nil
	})
	if ownership != systemproxy.Owned || event.Reclaimed {
		s.events.Publish(event)
	}
	if status, err := s.status(); err == nil {
		s.events.Publish(StatusChanged{Status: status})
	}
}

// holder names the local process the system proxy now points at.
func holder(observed systemproxy.Snapshot) string {
	for _, p := range []systemproxy.Proxy{observed.Web, observed.Secure, observed.SOCKS} {
		if p.Enabled && p.Port > 0 && slices.Contains([]string{"127.0.0.1", "localhost", "::1"}, p.Host) {
			if h, ok := ports.Find(p.Port); ok {
				if h.Process != "" {
					return fmt.Sprintf("%s (pid %d) on port %d", h.Process, h.PID, p.Port)
				}
				return fmt.Sprintf("pid %d on port %d", h.PID, p.Port)
			}
			return fmt.Sprintf("%s:%d", p.Host, p.Port)
		}
	}
	if observed.AutoProxy {
		return observed.PACURL
	}
	return ""
}

// UpdateSettings stores new settings and applies them when active.
func (s *SystemProxyService) UpdateSettings(settings systemproxy.Settings) (Status, error) {
	settings.BypassList = core.NormalizeList(settings.BypassList)
	if settings.Port == 0 {
		settings.Port = 7890
	}
	if err := settings.Validate(); err != nil {
		return Status{}, err
	}
	rt, err := s.deps.Runtime.Update(func(rt *core.Runtime) error {
		settings.Port = rt.Settings.MixedPort
		rt.SystemProxy.Settings = settings
		return nil
	})
	if err != nil {
		return Status{}, err
	}
	if rt.SystemProxy.Applied != nil {
		_, err = s.Apply()
	}
	status, statusErr := s.status()
	return status, errors.Join(err, statusErr)
}
