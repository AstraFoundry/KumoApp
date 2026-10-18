package application

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/mihomo"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/process"
)

// SettingsService changes how the core runs.
type SettingsService struct {
	deps      Dependencies
	lifecycle *Lifecycle
	proxy     *SystemProxyService
}

// CoreSettings is what the Core and Configure pages edit.
type CoreSettings struct {
	Mode           core.Mode            `json:"mode"`
	Settings       core.RuntimeSettings `json:"settings"`
	ControllerHost string               `json:"controllerHost"`
	ControllerPort int                  `json:"controllerPort"`
	HasSecret      bool                 `json:"hasSecret"`
	CorePath       string               `json:"corePath"`
	Candidates     []mihomo.Candidate   `json:"candidates"`
	TUNPermission  TUNPermission        `json:"tunPermission"`
}

// Get returns the stored settings and the cores Kumo can find.
func (s *SettingsService) Get() (CoreSettings, error) {
	rt, err := s.deps.Runtime.Load()
	if err != nil {
		return CoreSettings{}, err
	}
	candidates := mihomo.Discover(rt.CorePath, s.deps.Paths.ManagedCore(), s.deps.BundledDir)
	if candidates == nil {
		candidates = []mihomo.Candidate{}
	}
	selected := rt.CorePath
	if selected == "" && len(candidates) > 0 {
		selected = candidates[0].Path
	}
	return CoreSettings{
		Mode: rt.Mode, Settings: rt.Settings, ControllerHost: rt.Endpoint.Host, ControllerPort: rt.Endpoint.Port,
		HasSecret: rt.Endpoint.Secret != "", CorePath: selected, Candidates: candidates,
		TUNPermission: tunPermission(selected),
	}, nil
}

// SetMode switches the outbound mode, at once when the core runs, and
// closes connections so they follow the new mode.
func (s *SettingsService) SetMode(ctx context.Context, mode core.Mode) (Status, error) {
	if _, err := core.ParseMode(string(mode)); err != nil {
		return Status{}, err
	}
	rt, err := s.deps.Runtime.Update(func(rt *core.Runtime) error {
		rt.Mode = mode
		return nil
	})
	if err != nil {
		return Status{}, err
	}
	if rt.State == core.StateRunning {
		client := mihomo.NewClient(rt.Endpoint)
		if err := client.SetMode(ctx, mode); err != nil {
			return Status{}, err
		}
		_ = client.CloseConnections(ctx)
	}
	_ = s.lifecycle.writeConfig(rt)
	return s.lifecycle.Refresh()
}

// Update replaces the runtime settings and reloads a running core.
func (s *SettingsService) Update(ctx context.Context, settings core.RuntimeSettings) (Status, error) {
	settings = settings.Normalized()
	if err := core.ValidatePort(settings.MixedPort); err != nil {
		return Status{}, err
	}
	if settings.TUN.Enabled {
		if err := s.requireTUN(); err != nil {
			return Status{}, err
		}
	}
	rt, err := s.deps.Runtime.Update(func(rt *core.Runtime) error {
		if settings.MixedPort == rt.Endpoint.Port {
			return fmt.Errorf("the mixed port and the controller port must differ")
		}
		rt.Settings = settings
		rt.SystemProxy.Settings.Port = settings.MixedPort
		return nil
	})
	if err != nil {
		return Status{}, err
	}
	if err := s.lifecycle.Reload(ctx); err != nil {
		return Status{}, err
	}
	if rt.State == core.StateRunning && settings.TUN.Enabled {
		// TUN needs a fresh process when it was not running before.
		cfg, err := mihomo.NewClient(rt.Endpoint).Configs(ctx)
		if err == nil && !cfg.TUN.Enable {
			return s.lifecycle.Restart(ctx, StartOptions{})
		}
	}
	return s.lifecycle.Refresh()
}

// SetTUN turns TUN on or off.
func (s *SettingsService) SetTUN(ctx context.Context, enabled bool) (Status, error) {
	rt, err := s.deps.Runtime.Load()
	if err != nil {
		return Status{}, err
	}
	settings := rt.Settings
	settings.TUN.Enabled = enabled
	return s.Update(ctx, settings)
}

func (s *SettingsService) requireTUN() error {
	settings, err := s.Get()
	if err != nil {
		return err
	}
	if !settings.TUNPermission.Granted {
		return errors.New("TUN needs permission to create a network interface: grant it in TUN settings first")
	}
	return nil
}

// SetController moves the external controller, which takes a restart.
func (s *SettingsService) SetController(ctx context.Context, host string, port int, secret *string) (Status, error) {
	endpoint := core.Endpoint{Host: host, Port: port}
	if err := endpoint.Validate(); err != nil {
		return Status{}, err
	}
	rt, err := s.deps.Runtime.Update(func(rt *core.Runtime) error {
		if port == rt.Settings.MixedPort {
			return fmt.Errorf("the controller port and the mixed port must differ")
		}
		rt.Endpoint.Host, rt.Endpoint.Port = host, port
		if secret != nil {
			rt.Endpoint.Secret = *secret
		}
		return nil
	})
	if err != nil {
		return Status{}, err
	}
	if rt.State == core.StateRunning {
		return s.lifecycle.Restart(ctx, StartOptions{})
	}
	return s.lifecycle.Refresh()
}

// SetCorePath chooses the Mihomo executable; "" goes back to discovery.
func (s *SettingsService) SetCorePath(path string) error {
	if path != "" {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return fmt.Errorf("%w at %s", core.ErrCoreNotFound, path)
		}
	}
	_, err := s.deps.Runtime.Update(func(rt *core.Runtime) error {
		rt.CorePath = path
		return nil
	})
	return err
}

// InstallCore downloads the latest Mihomo and selects it.
func (s *SettingsService) InstallCore(ctx context.Context) (mihomo.Installed, error) {
	installed, err := mihomo.Install(ctx, s.deps.Paths.ManagedCore())
	if err != nil {
		return installed, err
	}
	return installed, s.SetCorePath(installed.Path)
}

// GrantTUN asks the system for the permission TUN needs.
func (s *SettingsService) GrantTUN(ctx context.Context) (TUNPermission, error) {
	settings, err := s.Get()
	if err != nil {
		return TUNPermission{}, err
	}
	if settings.CorePath == "" {
		return TUNPermission{}, core.ErrCoreNotFound
	}
	rt, _ := s.deps.Runtime.Load()
	if rt.State == core.StateRunning && process.IsCore(rt.PID, settings.CorePath) {
		return TUNPermission{}, errors.New("stop Kumo before granting TUN permission")
	}
	if err := grantTUN(ctx, settings.CorePath); err != nil {
		return tunPermission(settings.CorePath), err
	}
	return tunPermission(settings.CorePath), nil
}
