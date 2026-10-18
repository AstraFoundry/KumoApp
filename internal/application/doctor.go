package application

import (
	"context"
	"runtime"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/mihomo"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/ports"
)

// Check is one item of a doctor report.
type Check struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

// Report is a broad snapshot for diagnosing Kumo.
type Report struct {
	Status      Status                `json:"status"`
	Checks      []Check               `json:"checks"`
	Cores       []mihomo.Candidate    `json:"cores"`
	SystemProxy *systemproxy.Snapshot `json:"observedSystemProxy,omitempty"`
	DataDir     string                `json:"dataDir"`
	Platform    string                `json:"platform"`
}

// Doctor gathers status, the cores Kumo can find, port checks and what the
// system proxy shows.
func (k *Kumo) Doctor(ctx context.Context) (Report, error) {
	status, err := k.Lifecycle.Status()
	if err != nil {
		return Report{}, err
	}
	rt, _ := k.deps.Runtime.Load()
	report := Report{Status: status, DataDir: k.deps.Paths.Root, Platform: runtime.GOOS + "/" + runtime.GOARCH}
	report.Cores = mihomo.Discover(rt.CorePath, k.deps.Paths.ManagedCore(), k.deps.BundledDir)
	if report.Cores == nil {
		report.Cores = []mihomo.Candidate{}
	}
	add := func(name string, ok bool, detail string) {
		report.Checks = append(report.Checks, Check{Name: name, OK: ok, Detail: detail})
	}
	if len(report.Cores) > 0 {
		add("core", true, report.Cores[0].Path)
	} else {
		add("core", false, "no Mihomo executable found; run `kumo core install`")
	}
	if config, err := k.Lifecycle.RuntimeConfig(rt); err != nil {
		add("configuration", false, err.Error())
	} else {
		add("configuration", true, "the current profile and overrides build")
		_ = config
	}
	if status.Running() {
		if v, err := mihomo.NewClient(rt.Endpoint).Version(ctx); err == nil {
			add("controller", true, "Mihomo "+v.Version+" answers on "+rt.Endpoint.Address())
		} else {
			add("controller", false, err.Error())
		}
	} else {
		for _, p := range []struct {
			role string
			host string
			port int
		}{{"mixed port", "127.0.0.1", rt.Settings.MixedPort}, {"controller port", rt.Endpoint.Host, rt.Endpoint.Port}} {
			if ports.Free(p.host, p.port) {
				add(p.role, true, "free")
				continue
			}
			detail := "in use"
			if h, ok := ports.Find(p.port); ok {
				detail = "in use by " + h.Process
			}
			add(p.role, false, detail)
		}
	}
	if service, err := k.SystemProxy.service(rt.SystemProxy.Settings); err == nil {
		if snap, err := k.deps.SystemProxy.Observe(service); err == nil {
			report.SystemProxy = &snap
			switch {
			case rt.SystemProxy.Applied != nil:
				ownership := rt.SystemProxy.Applied.OwnershipOf(snap)
				add("system proxy", ownership == systemproxy.Owned, string(ownership))
			case snap.IsCleared():
				add("system proxy", true, "off")
			default:
				add("system proxy", !rt.SystemProxy.Enabled, "set by another app: "+holder(snap))
			}
		} else {
			add("system proxy", false, err.Error())
		}
	}
	if rt.State == core.StateFailed && rt.Message != "" {
		add("last start", false, rt.Message)
	}
	return report, nil
}
