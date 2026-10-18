// Package desktop exposes Kumo's use cases to the app's pages: each bound
// service is a thin adapter over the application layer, and Go events
// carry what changes on its own (status, traffic, logs, contention).
package desktop

import (
	"context"
	"encoding/base64"
	"errors"
	"sync"

	"github.com/egoist/mygo"

	"github.com/AstraFoundry/KumoApp/internal/application"
	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/inspect"
	"github.com/AstraFoundry/KumoApp/internal/domain/override"
	"github.com/AstraFoundry/KumoApp/internal/domain/preferences"
	"github.com/AstraFoundry/KumoApp/internal/domain/profile"
	"github.com/AstraFoundry/KumoApp/internal/domain/proxy"
	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
	"github.com/AstraFoundry/KumoApp/internal/domain/topology"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/mihomo"
)

// Runtime starts, stops and describes the core.
type Runtime struct{ k *application.Kumo }

// Status returns the current status.
func (r Runtime) Status() (application.Status, error) { return r.k.Lifecycle.Status() }

// StartOutcome is a start's result. Conflicts lists ports other apps
// hold; the start did not happen, and the page offers to move past them.
type StartOutcome struct {
	Status    application.Status `json:"status"`
	Conflicts core.PortConflicts `json:"conflicts"`
}

// Start starts Mihomo. With resolvePorts, ports held by other apps are
// replaced by free ones instead of reported.
func (r Runtime) Start(ctx context.Context, resolvePorts bool) (StartOutcome, error) {
	status, err := r.k.Lifecycle.Start(context.WithoutCancel(ctx), application.StartOptions{ResolvePortConflicts: resolvePorts})
	var conflicts core.PortConflicts
	if errors.As(err, &conflicts) {
		return StartOutcome{Status: status, Conflicts: conflicts}, nil
	}
	return StartOutcome{Status: status, Conflicts: core.PortConflicts{}}, err
}

// Stop releases the system proxy and stops Mihomo.
func (r Runtime) Stop(ctx context.Context) (application.Status, error) {
	return r.k.Lifecycle.Stop(context.WithoutCancel(ctx))
}

// Restart stops and starts Mihomo.
func (r Runtime) Restart(ctx context.Context) (application.Status, error) {
	return r.k.Lifecycle.Restart(context.WithoutCancel(ctx), application.StartOptions{})
}

// SetMode switches the outbound mode.
func (r Runtime) SetMode(ctx context.Context, mode core.Mode) (application.Status, error) {
	return r.k.Settings.SetMode(ctx, mode)
}

// Doctor collects a diagnosis.
func (r Runtime) Doctor(ctx context.Context) (application.Report, error) { return r.k.Doctor(ctx) }

// SystemProxy points the system at Kumo.
type SystemProxy struct{ k *application.Kumo }

// SetEnabled turns the system proxy on or off.
func (s SystemProxy) SetEnabled(enabled bool) (application.Status, error) {
	return s.k.SystemProxy.SetEnabled(enabled)
}

// UpdateSettings stores new settings and applies them when active.
func (s SystemProxy) UpdateSettings(settings systemproxy.Settings) (application.Status, error) {
	return s.k.SystemProxy.UpdateSettings(settings)
}

// Services lists network services.
func (s SystemProxy) Services() ([]string, error) { return s.k.SystemProxy.Services() }

// Reclaim puts Kumo's proxy back after another app replaced it.
func (s SystemProxy) Reclaim() (application.Status, error) {
	if _, err := s.k.SystemProxy.Apply(); err != nil {
		return application.Status{}, err
	}
	return s.k.Lifecycle.Refresh()
}

// Profiles manages profiles.
type Profiles struct{ k *application.Kumo }

// List returns every profile.
func (p Profiles) List() ([]profile.Summary, error) { return p.k.Profiles.List() }

// Content returns a profile's YAML.
func (p Profiles) Content(id string) (string, error) { return p.k.Profiles.Content(id) }

// Select makes a profile current.
func (p Profiles) Select(ctx context.Context, id string) error {
	return p.k.Profiles.Select(context.WithoutCancel(ctx), id)
}

// Import adds a profile from a URL, a file or text.
func (p Profiles) Import(ctx context.Context, req application.ImportRequest) (profile.Summary, error) {
	return p.k.Profiles.Import(context.WithoutCancel(ctx), req)
}

// ChooseFile asks for a YAML file and imports it; "" when canceled.
func (p Profiles) ChooseFile(ctx context.Context, activate bool) (*profile.Summary, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent:  mygo.CallerWindow(ctx),
		Filters: []mygo.FileFilter{{Name: "Mihomo profile", Extensions: []string{"yaml", "yml"}}},
	})
	if err != nil || len(paths) == 0 {
		return nil, err
	}
	summary, err := p.k.Profiles.Import(context.WithoutCancel(ctx), application.ImportRequest{Path: paths[0], Activate: activate})
	return &summary, err
}

// Refresh downloads a remote profile again.
func (p Profiles) Refresh(ctx context.Context, id string) (profile.Summary, error) {
	return p.k.Profiles.Refresh(context.WithoutCancel(ctx), id)
}

// Update edits a profile.
func (p Profiles) Update(ctx context.Context, req application.UpdateRequest) (profile.Summary, error) {
	return p.k.Profiles.Update(context.WithoutCancel(ctx), req)
}

// Delete removes a profile.
func (p Profiles) Delete(ctx context.Context, id string) error {
	return p.k.Profiles.Delete(context.WithoutCancel(ctx), id)
}

// Overrides manages overrides.
type Overrides struct{ k *application.Kumo }

// List returns overrides in order.
func (o Overrides) List() ([]override.Override, error) { return o.k.Overrides.List() }

// Get returns an override with its content.
func (o Overrides) Get(id string) (application.OverrideDetail, error) { return o.k.Overrides.Get(id) }

// Save creates or edits an override.
func (o Overrides) Save(ctx context.Context, req application.SaveRequest) (override.Override, error) {
	return o.k.Overrides.Save(context.WithoutCancel(ctx), req)
}

// SetDisabled turns an override off or on.
func (o Overrides) SetDisabled(ctx context.Context, id string, disabled bool) error {
	return o.k.Overrides.SetDisabled(context.WithoutCancel(ctx), id, disabled)
}

// Delete removes an override.
func (o Overrides) Delete(ctx context.Context, id string) error {
	return o.k.Overrides.Delete(context.WithoutCancel(ctx), id)
}

// Reorder sets the order overrides apply in.
func (o Overrides) Reorder(ctx context.Context, ids []string) error {
	return o.k.Overrides.Reorder(context.WithoutCancel(ctx), ids)
}

// Proxies lists groups and picks nodes.
type Proxies struct{ k *application.Kumo }

// Groups returns proxy groups; read from the profile while the core is off.
func (p Proxies) Groups(ctx context.Context) ([]proxy.Group, error) { return p.k.Proxies.Groups(ctx) }

// Select uses a node in a group.
func (p Proxies) Select(ctx context.Context, group, name string) error {
	return p.k.Proxies.Select(ctx, group, name)
}

// TestGroup measures a group's nodes.
func (p Proxies) TestGroup(ctx context.Context, group string) (map[string]int, error) {
	return p.k.Proxies.TestGroup(ctx, group)
}

// TestNode measures one node.
func (p Proxies) TestNode(ctx context.Context, name, testURL string) (int, error) {
	return p.k.Proxies.TestNode(ctx, name, testURL)
}

// Inspect reads what the core reports.
type Inspect struct {
	k     *application.Kumo
	mu    sync.Mutex
	icons map[string]string
}

// Connections lists open connections.
func (i *Inspect) Connections(ctx context.Context) (inspect.Connections, error) {
	return i.k.Inspect.Connections(ctx)
}

// CloseConnection closes one connection, or all when id is "".
func (i *Inspect) CloseConnection(ctx context.Context, id string) error {
	return i.k.Inspect.CloseConnection(ctx, id)
}

// Rules lists the running rules.
func (i *Inspect) Rules(ctx context.Context) ([]inspect.Rule, error) { return i.k.Inspect.Rules(ctx) }

// SetRuleEnabled turns a rule off or on until the core restarts.
func (i *Inspect) SetRuleEnabled(ctx context.Context, index int, enabled bool) error {
	return i.k.Inspect.SetRuleEnabled(ctx, index, enabled)
}

// ProviderList is every provider of the running core.
type ProviderList struct {
	Proxy []inspect.ProxyProvider `json:"proxy"`
	Rule  []inspect.RuleProvider  `json:"rule"`
}

// Providers lists proxy and rule providers.
func (i *Inspect) Providers(ctx context.Context) (ProviderList, error) {
	proxies, rules, err := i.k.Inspect.Providers(ctx)
	if proxies == nil {
		proxies = []inspect.ProxyProvider{}
	}
	if rules == nil {
		rules = []inspect.RuleProvider{}
	}
	return ProviderList{Proxy: proxies, Rule: rules}, err
}

// UpdateProvider refreshes a "proxy" or "rule" provider.
func (i *Inspect) UpdateProvider(ctx context.Context, kind, name string) error {
	return i.k.Inspect.UpdateProvider(ctx, kind, name)
}

// UpgradeGeoData downloads fresh geo databases.
func (i *Inspect) UpgradeGeoData(ctx context.Context) error { return i.k.Inspect.UpgradeGeoData(ctx) }

// RecentLogs reads the end of the core's log.
func (i *Inspect) RecentLogs(limit int) ([]inspect.LogEntry, error) {
	return i.k.Inspect.RecentLogs(limit)
}

// ProcessIcon returns a process's icon as a PNG data URL, or "".
func (i *Inspect) ProcessIcon(path string) string {
	if path == "" {
		return ""
	}
	i.mu.Lock()
	if icon, ok := i.icons[path]; ok {
		i.mu.Unlock()
		return icon
	}
	i.mu.Unlock()
	png, err := mygo.Shell.FileIcon(appBundle(path), 64)
	icon := ""
	if err == nil && len(png) > 0 {
		icon = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	}
	i.mu.Lock()
	if i.icons == nil {
		i.icons = map[string]string{}
	}
	i.icons[path] = icon
	i.mu.Unlock()
	return icon
}

// Settings changes how the core runs.
type Settings struct{ k *application.Kumo }

// Get returns the stored settings.
func (s Settings) Get() (application.CoreSettings, error) { return s.k.Settings.Get() }

// Update replaces the runtime settings.
func (s Settings) Update(ctx context.Context, settings core.RuntimeSettings) (application.Status, error) {
	return s.k.Settings.Update(context.WithoutCancel(ctx), settings)
}

// SetTUN turns TUN on or off.
func (s Settings) SetTUN(ctx context.Context, enabled bool) (application.Status, error) {
	return s.k.Settings.SetTUN(context.WithoutCancel(ctx), enabled)
}

// GrantTUN asks for the permission TUN needs.
func (s Settings) GrantTUN(ctx context.Context) (application.TUNPermission, error) {
	return s.k.Settings.GrantTUN(ctx)
}

// SetController moves the external controller.
func (s Settings) SetController(ctx context.Context, host string, port int, secret *string) (application.Status, error) {
	return s.k.Settings.SetController(context.WithoutCancel(ctx), host, port, secret)
}

// SetCorePath picks the Mihomo executable; "" finds one automatically.
func (s Settings) SetCorePath(path string) error { return s.k.Settings.SetCorePath(path) }

// ChooseCore asks for an executable and selects it.
func (s Settings) ChooseCore(ctx context.Context) (string, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: mygo.CallerWindow(ctx)})
	if err != nil || len(paths) == 0 {
		return "", err
	}
	return paths[0], s.k.Settings.SetCorePath(paths[0])
}

// InstallCore downloads the latest Mihomo.
func (s Settings) InstallCore(ctx context.Context) (mihomo.Installed, error) {
	return s.k.Settings.InstallCore(context.WithoutCancel(ctx))
}

// RuntimeConfig is the YAML the core runs with, or would run with.
func (s Settings) RuntimeConfig() (string, error) {
	rt, err := s.k.Runtime()
	if err != nil {
		return "", err
	}
	return s.k.Lifecycle.RuntimeConfig(rt)
}

// Topology draws the configuration as a graph.
type Topology struct{ k *application.Kumo }

// Graph returns the topology of the current configuration.
func (t Topology) Graph(ctx context.Context) (topology.Graph, error) { return t.k.Topology.Graph(ctx) }

// Preferences keeps app-level preferences.
type Preferences struct {
	k      *application.Kumo
	onSave func(preferences.Preferences)
}

// Get returns the preferences.
func (p Preferences) Get() (preferences.Preferences, error) { return p.k.Preferences.Get() }

// Save stores the preferences and applies those the app acts on.
func (p Preferences) Save(prefs preferences.Preferences) error {
	if err := p.k.Preferences.Save(prefs); err != nil {
		return err
	}
	if p.onSave != nil {
		p.onSave(prefs)
	}
	return nil
}
