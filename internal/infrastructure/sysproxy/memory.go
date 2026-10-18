package sysproxy

import (
	"slices"
	"sync"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
)

// Memory is an in-memory system proxy. Tests and lifecycle simulations use
// it, and act as another app with Takeover.
type Memory struct {
	mu    sync.Mutex
	state map[string]systemproxy.Snapshot
	// Fail makes every change fail, like a denied permission.
	Fail error
}

// NewMemory has one service, "Wi-Fi", with nothing set.
func NewMemory() *Memory {
	return &Memory{state: map[string]systemproxy.Snapshot{"Wi-Fi": {Service: "Wi-Fi"}}}
}

func (m *Memory) Services() ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var names []string
	for name := range m.state {
		names = append(names, name)
	}
	slices.Sort(names)
	return names, nil
}

func (m *Memory) ActiveService() (string, error) { return "Wi-Fi", nil }

func (m *Memory) Observe(service string) (systemproxy.Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	snap := m.state[service]
	snap.Service = service
	snap.CapturedAt = time.Now()
	return snap, nil
}

func (m *Memory) ApplyManual(service, host string, port int, bypass []string) error {
	if m.Fail != nil {
		return m.Fail
	}
	p := systemproxy.Proxy{Enabled: true, Host: host, Port: port}
	m.set(service, systemproxy.Snapshot{Web: p, Secure: p, SOCKS: p, Bypass: bypass})
	return nil
}

func (m *Memory) ApplyPAC(service, url string) error {
	if m.Fail != nil {
		return m.Fail
	}
	m.set(service, systemproxy.Snapshot{AutoProxy: true, PACURL: url})
	return nil
}

func (m *Memory) Disable(service string) error {
	if m.Fail != nil {
		return m.Fail
	}
	m.set(service, systemproxy.Snapshot{})
	return nil
}

// Takeover points the service at another proxy, as another app would.
func (m *Memory) Takeover(service, host string, port int) {
	p := systemproxy.Proxy{Enabled: true, Host: host, Port: port}
	m.set(service, systemproxy.Snapshot{Web: p, Secure: p, SOCKS: p})
}

func (m *Memory) set(service string, snap systemproxy.Snapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	snap.Service = service
	m.state[service] = snap
}
