package sysproxy

import (
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/storage"
)

// File keeps a pretend system proxy in a JSON file, shared by every
// process that names the same file. Lifecycle simulations use it to play
// the system across the app, the CLI, the guardian and "other apps".
type File struct{ Path string }

func (f File) load() map[string]systemproxy.Snapshot {
	state := map[string]systemproxy.Snapshot{}
	_, _ = storage.ReadJSON(f.Path, &state)
	return state
}

func (f File) update(service string, snap systemproxy.Snapshot) error {
	lock, err := storage.Lock(f.Path + ".lock")
	if err != nil {
		return err
	}
	defer lock.Unlock()
	state := f.load()
	snap.Service = service
	state[service] = snap
	return storage.WriteJSON(f.Path, state)
}

func (f File) Services() ([]string, error)    { return []string{"Wi-Fi"}, nil }
func (f File) ActiveService() (string, error) { return "Wi-Fi", nil }

func (f File) Observe(service string) (systemproxy.Snapshot, error) {
	snap := f.load()[service]
	snap.Service, snap.CapturedAt = service, time.Now()
	return snap, nil
}

func (f File) ApplyManual(service, host string, port int, bypass []string) error {
	p := systemproxy.Proxy{Enabled: true, Host: host, Port: port}
	return f.update(service, systemproxy.Snapshot{Web: p, Secure: p, SOCKS: p, Bypass: bypass})
}

func (f File) ApplyPAC(service, url string) error {
	return f.update(service, systemproxy.Snapshot{AutoProxy: true, PACURL: url})
}

func (f File) Disable(service string) error { return f.update(service, systemproxy.Snapshot{}) }

// Takeover points the service at another proxy, as another app would.
func (f File) Takeover(service, host string, port int) error {
	p := systemproxy.Proxy{Enabled: true, Host: host, Port: port}
	return f.update(service, systemproxy.Snapshot{Web: p, Secure: p, SOCKS: p})
}
