package application

import (
	"context"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/process"
)

// Guard watches the app with pid parent. If the app goes away without
// quitting properly (a crash, a force quit, a kill), Guard does what the
// quit would have done: it lets go of the system proxy and stops the core
// the app owned. A clean quit leaves nothing to do, and Guard just ends.
func Guard(ctx context.Context, k *Kumo, parent int, interval time.Duration) ShutdownReport {
	for process.Alive(parent) {
		select {
		case <-ctx.Done():
			return ShutdownReport{}
		case <-time.After(interval):
		}
	}
	rt, err := k.deps.Runtime.Load()
	if err != nil {
		return ShutdownReport{Diagnostics: []string{err.Error()}}
	}
	if rt.IsActive() && rt.Owner == core.OwnerApp {
		return k.Lifecycle.Shutdown(ctx)
	}
	if rt.SystemProxy.Applied != nil && !rt.IsActive() {
		report := ShutdownReport{}
		if err := k.SystemProxy.Release(); err != nil {
			report.Diagnostics = append(report.Diagnostics, err.Error())
		}
		return report
	}
	return ShutdownReport{}
}

// Recover is the first thing the app does: it adopts a core that still
// runs from an earlier session, records one that died, and takes Kumo's
// settings out of the system when no core serves them any more.
func (l *Lifecycle) Recover() (Status, error) {
	rt, err := l.reconcile()
	if err != nil {
		return Status{}, err
	}
	if !rt.IsActive() && rt.SystemProxy.Applied != nil {
		if err := l.proxy.Release(); err != nil {
			return l.describe(rt), err
		}
	}
	return l.Refresh()
}
