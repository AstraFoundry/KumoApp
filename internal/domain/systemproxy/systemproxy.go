// Package systemproxy models the operating system's proxy settings as Kumo
// sees them: what Kumo wants applied, what it applied, what the system shows
// now, and whether Kumo still owns it when another app may have changed it.
package systemproxy

import (
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Mode is how the system is pointed at the core.
type Mode string

const (
	// ModeManual sets the HTTP, HTTPS and SOCKS proxies to the mixed port.
	ModeManual Mode = "manual"
	// ModePAC serves a proxy auto-config script and points the system at it.
	ModePAC Mode = "pac"
)

// AutomaticService resolves to the network service of the default route.
const AutomaticService = "Automatic"

// Settings are what the user configured for the system proxy.
type Settings struct {
	NetworkService string   `json:"networkService"`
	Host           string   `json:"host"`
	Port           int      `json:"port"`
	Mode           Mode     `json:"mode"`
	BypassList     []string `json:"bypassList"`
	PACScript      string   `json:"pacScript"`
}

// DefaultBypassList keeps local networks off the proxy.
var DefaultBypassList = []string{
	"127.0.0.1/8", "192.168.0.0/16", "10.0.0.0/8", "172.16.0.0/12",
	"localhost", "*.local", "<local>",
}

// DefaultPACScript proxies everything through the mixed port, falling back
// to a direct connection.
const DefaultPACScript = `function FindProxyForURL(url, host) {
  return "PROXY 127.0.0.1:%mixed-port%; SOCKS5 127.0.0.1:%mixed-port%; DIRECT;";
}`

// DefaultSettings follow the default route and use the manual mode.
func DefaultSettings() Settings {
	return Settings{
		NetworkService: AutomaticService,
		Host:           "127.0.0.1",
		Port:           7890,
		Mode:           ModeManual,
		BypassList:     slices.Clone(DefaultBypassList),
		PACScript:      DefaultPACScript,
	}
}

// Validate rejects settings that cannot be applied.
func (s Settings) Validate() error {
	if strings.TrimSpace(s.Host) == "" {
		return fmt.Errorf("system proxy host is empty")
	}
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("system proxy port %d is outside 1–65535", s.Port)
	}
	if s.Mode != ModeManual && s.Mode != ModePAC {
		return fmt.Errorf("unknown system proxy mode %q", s.Mode)
	}
	return nil
}

// RenderPAC fills the mixed port into a PAC script.
func RenderPAC(script string, port int) string {
	return strings.ReplaceAll(script, "%mixed-port%", strconv.Itoa(port))
}

// Proxy is one proxy setting of the system: HTTP, HTTPS or SOCKS.
type Proxy struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

// PointsAt reports whether the proxy is on and targets host:port. Loopback
// addresses are interchangeable.
func (p Proxy) PointsAt(host string, port int) bool {
	return p.Enabled && p.Port == port && sameHost(p.Host, host)
}

func sameHost(a, b string) bool {
	if strings.EqualFold(a, b) {
		return true
	}
	return isLoopback(a) && isLoopback(b)
}

func isLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// Snapshot is what the system's proxy settings are at one moment.
type Snapshot struct {
	Service    string    `json:"service"`
	Web        Proxy     `json:"web"`
	Secure     Proxy     `json:"secure"`
	SOCKS      Proxy     `json:"socks"`
	AutoProxy  bool      `json:"autoProxy"`
	PACURL     string    `json:"pacURL,omitempty"`
	Bypass     []string  `json:"bypass,omitempty"`
	CapturedAt time.Time `json:"capturedAt"`
}

// IsCleared reports whether no proxy of any kind is on.
func (s Snapshot) IsCleared() bool {
	return !s.Web.Enabled && !s.Secure.Enabled && !s.SOCKS.Enabled && !s.AutoProxy
}

// Applied is what Kumo wrote into the system, kept to tell later whether the
// system still shows it.
type Applied struct {
	Service string    `json:"service"`
	Mode    Mode      `json:"mode"`
	Host    string    `json:"host"`
	Port    int       `json:"port"`
	PACURL  string    `json:"pacURL,omitempty"`
	At      time.Time `json:"at"`
}

// Ownership is who controls the system proxy now.
type Ownership string

const (
	// Owned: the system still shows what Kumo applied.
	Owned Ownership = "owned"
	// TakenOver: another app (or the user) pointed the system elsewhere.
	TakenOver Ownership = "takenOver"
	// Cleared: nothing is set any more; someone turned it off.
	Cleared Ownership = "cleared"
)

// OwnershipOf compares what the system shows with what Kumo applied.
func (a Applied) OwnershipOf(observed Snapshot) Ownership {
	if observed.IsCleared() {
		return Cleared
	}
	switch a.Mode {
	case ModePAC:
		if observed.AutoProxy && observed.PACURL == a.PACURL {
			return Owned
		}
	default:
		manual := observed.Web.PointsAt(a.Host, a.Port) &&
			observed.Secure.PointsAt(a.Host, a.Port) &&
			observed.SOCKS.PointsAt(a.Host, a.Port)
		if manual && !observed.AutoProxy {
			return Owned
		}
	}
	return TakenOver
}

// State is the system proxy as Kumo tracks it across launches.
type State struct {
	// Enabled is what the user asked for.
	Enabled  bool     `json:"enabled"`
	Settings Settings `json:"settings"`
	// Applied is set while Kumo's settings are in the system.
	Applied *Applied `json:"applied,omitempty"`
	// Previous is what the system showed before Kumo first applied its own.
	Previous *Snapshot `json:"previous,omitempty"`
	// Ownership is what the last check found while Applied was set.
	Ownership Ownership `json:"ownership,omitempty"`
}

// DefaultState has the proxy off with the default settings.
func DefaultState() State { return State{Settings: DefaultSettings()} }

// Platform is the operating system's side of the system proxy, implemented
// for each OS in infrastructure.
type Platform interface {
	// Services lists the network services proxies can be set on.
	Services() ([]string, error)
	// ActiveService is the service of the default route.
	ActiveService() (string, error)
	// Observe reads a service's current proxy settings.
	Observe(service string) (Snapshot, error)
	// ApplyManual points the HTTP, HTTPS and SOCKS proxies at host:port and
	// turns the auto proxy off.
	ApplyManual(service, host string, port int, bypass []string) error
	// ApplyPAC turns the manual proxies off and the auto proxy on at url.
	ApplyPAC(service, url string) error
	// Disable turns every proxy of the service off.
	Disable(service string) error
}
