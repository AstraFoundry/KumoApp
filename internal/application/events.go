// Package application holds Kumo's use cases. The desktop app, the CLI and
// the guardian process all drive the same services, so every surface
// starts, stops and configures the core the same way.
package application

import (
	"sync"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/inspect"
	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
)

// Event is something the surfaces may want to show.
type Event interface{ eventName() string }

// StatusChanged carries the new status after any change of the runtime.
type StatusChanged struct{ Status Status }

// CoreExited reports a core that ended without Kumo stopping it.
type CoreExited struct{ Exit core.Exit }

// SystemProxyContended reports that another app changed the system proxy
// while Kumo had it.
type SystemProxyContended struct {
	Ownership systemproxy.Ownership `json:"ownership"`
	Observed  systemproxy.Snapshot  `json:"observed"`
	// Holder is the process listening where the system now points, when
	// that is on this machine.
	Holder string `json:"holder,omitempty"`
	// Reclaimed is true when Kumo put its proxy back.
	Reclaimed bool `json:"reclaimed"`
}

// TrafficSampled carries one second of throughput.
type TrafficSampled struct{ Traffic inspect.Traffic }

// LogReceived carries one line of the core's log.
type LogReceived struct{ Entry inspect.LogEntry }

// ProfilesChanged reports that profiles were added, changed or removed.
type ProfilesChanged struct{}

func (StatusChanged) eventName() string        { return "status" }
func (CoreExited) eventName() string           { return "coreExited" }
func (SystemProxyContended) eventName() string { return "systemProxyContended" }
func (TrafficSampled) eventName() string       { return "traffic" }
func (LogReceived) eventName() string          { return "log" }
func (ProfilesChanged) eventName() string      { return "profilesChanged" }

// Bus fans events out to subscribers. Publishing never blocks on a slow
// subscriber's work: subscribers return quickly or hand off.
type Bus struct {
	mu   sync.RWMutex
	subs map[int]func(Event)
	next int
}

// NewBus returns an empty bus.
func NewBus() *Bus { return &Bus{subs: map[int]func(Event){}} }

// Subscribe registers fn and returns a function that removes it.
func (b *Bus) Subscribe(fn func(Event)) func() {
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.next
	b.next++
	b.subs[id] = fn
	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.subs, id)
	}
}

// Publish delivers e to every subscriber.
func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	subs := make([]func(Event), 0, len(b.subs))
	for _, fn := range b.subs {
		subs = append(subs, fn)
	}
	b.mu.RUnlock()
	for _, fn := range subs {
		fn(e)
	}
}
