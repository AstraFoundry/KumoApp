package application

import (
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/override"
	"github.com/AstraFoundry/KumoApp/internal/domain/preferences"
	"github.com/AstraFoundry/KumoApp/internal/domain/profile"
	"github.com/AstraFoundry/KumoApp/internal/domain/runtimeconfig"
	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/fetch"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/pac"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/paths"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/storage"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/sysproxy"
)

// Dependencies are the adapters the services run on. Tests and
// simulations swap the system proxy for an in-memory one.
type Dependencies struct {
	Paths       paths.Paths
	Runtime     core.Store
	Profiles    profile.Repository
	Overrides   override.Repository
	Preferences preferences.Store
	SystemProxy systemproxy.Platform
	Fetcher     interface {
		profile.Fetcher
		Get(url string, proxyPort int) ([]byte, map[string][]string, error)
	}
	Transformer runtimeconfig.Transformer
	// Owner is the surface these services run in.
	Owner core.Owner
	// BundledDir holds a Mihomo shipped with the app, if any.
	BundledDir string
	// ReadyTimeout bounds how long a start waits for the controller.
	ReadyTimeout time.Duration
	// WatchInterval is how often the system proxy is checked for
	// contention while Kumo holds it.
	WatchInterval time.Duration
	Now           func() time.Time
}

// DefaultDependencies wires the real adapters under p.
func DefaultDependencies(p paths.Paths, owner core.Owner) Dependencies {
	return Dependencies{
		Paths:         p,
		Runtime:       storage.NewRuntimeStore(p),
		Profiles:      storage.NewProfileStore(p),
		Overrides:     storage.NewOverrideStore(p),
		Preferences:   storage.NewPreferenceStore(p),
		SystemProxy:   sysproxy.ForOS(),
		Fetcher:       fetchAdapter{fetch.New()},
		Owner:         owner,
		ReadyTimeout:  15 * time.Second,
		WatchInterval: 3 * time.Second,
		Now:           time.Now,
	}
}

type fetchAdapter struct{ *fetch.Client }

func (f fetchAdapter) Get(url string, proxyPort int) ([]byte, map[string][]string, error) {
	body, header, err := f.Client.Get(url, proxyPort)
	return body, header, err
}

// Kumo is every use case, sharing one set of adapters and one event bus.
type Kumo struct {
	deps   Dependencies
	Events *Bus

	Lifecycle   *Lifecycle
	SystemProxy *SystemProxyService
	Profiles    *ProfileService
	Overrides   *OverrideService
	Proxies     *ProxyService
	Inspect     *InspectService
	Settings    *SettingsService
	Topology    *TopologyService
	Preferences *PreferenceService
}

// New wires the services.
func New(deps Dependencies) *Kumo {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.ReadyTimeout == 0 {
		deps.ReadyTimeout = 15 * time.Second
	}
	if deps.WatchInterval == 0 {
		deps.WatchInterval = 3 * time.Second
	}
	k := &Kumo{deps: deps, Events: NewBus()}
	k.SystemProxy = &SystemProxyService{deps: deps, events: k.Events, pac: &pac.Server{}}
	k.Lifecycle = &Lifecycle{deps: deps, events: k.Events, proxy: k.SystemProxy}
	k.SystemProxy.status = k.Lifecycle.Status
	k.Profiles = &ProfileService{deps: deps, events: k.Events, lifecycle: k.Lifecycle}
	k.Overrides = &OverrideService{deps: deps, lifecycle: k.Lifecycle}
	k.Proxies = &ProxyService{deps: deps, lifecycle: k.Lifecycle}
	k.Inspect = &InspectService{deps: deps, events: k.Events}
	k.Settings = &SettingsService{deps: deps, lifecycle: k.Lifecycle, proxy: k.SystemProxy}
	k.Topology = &TopologyService{deps: deps, lifecycle: k.Lifecycle}
	k.Preferences = &PreferenceService{deps: deps}
	return k
}

// Paths are where this Kumo keeps its files.
func (k *Kumo) Paths() paths.Paths { return k.deps.Paths }
