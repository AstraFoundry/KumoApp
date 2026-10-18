package sysproxy

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
)

// LinuxService is the one "service" of desktop Linux: the GNOME proxy
// settings, which most desktops and browsers follow.
const LinuxService = "System"

// Linux drives org.gnome.system.proxy with gsettings.
type Linux struct{ Runner Runner }

func (l Linux) Services() ([]string, error)    { return []string{LinuxService}, nil }
func (l Linux) ActiveService() (string, error) { return LinuxService, nil }

func (l Linux) get(schema, key string) string {
	out, err := l.Runner.Run("gsettings", "get", schema, key)
	if err != nil {
		return ""
	}
	return strings.Trim(out, "'")
}

func (l Linux) set(schema, key, value string) error {
	_, err := l.Runner.Run("gsettings", "set", schema, key, value)
	return err
}

// Observe reads the GNOME proxy settings.
func (l Linux) Observe(string) (systemproxy.Snapshot, error) {
	mode := l.get("org.gnome.system.proxy", "mode")
	if mode == "" {
		return systemproxy.Snapshot{}, fmt.Errorf("gsettings is unavailable; Kumo can only set the GNOME proxy settings")
	}
	snap := systemproxy.Snapshot{Service: LinuxService, CapturedAt: time.Now()}
	read := func(kind string) systemproxy.Proxy {
		port, _ := strconv.Atoi(l.get("org.gnome.system.proxy."+kind, "port"))
		return systemproxy.Proxy{Enabled: mode == "manual", Host: l.get("org.gnome.system.proxy."+kind, "host"), Port: port}
	}
	snap.Web, snap.Secure, snap.SOCKS = read("http"), read("https"), read("socks")
	snap.AutoProxy = mode == "auto"
	snap.PACURL = l.get("org.gnome.system.proxy", "autoconfig-url")
	return snap, nil
}

// ApplyManual sets the HTTP, HTTPS and SOCKS proxies.
func (l Linux) ApplyManual(_ string, host string, port int, bypass []string) error {
	for _, kind := range []string{"http", "https", "socks"} {
		schema := "org.gnome.system.proxy." + kind
		if err := l.set(schema, "host", host); err != nil {
			return err
		}
		if err := l.set(schema, "port", strconv.Itoa(port)); err != nil {
			return err
		}
	}
	quoted := make([]string, len(bypass))
	for i, b := range bypass {
		quoted[i] = "'" + strings.ReplaceAll(b, "'", "") + "'"
	}
	if err := l.set("org.gnome.system.proxy", "ignore-hosts", "["+strings.Join(quoted, ", ")+"]"); err != nil {
		return err
	}
	return l.set("org.gnome.system.proxy", "mode", "manual")
}

// ApplyPAC points the system at a PAC URL.
func (l Linux) ApplyPAC(_ string, url string) error {
	if err := l.set("org.gnome.system.proxy", "autoconfig-url", url); err != nil {
		return err
	}
	return l.set("org.gnome.system.proxy", "mode", "auto")
}

// Disable turns the system proxy off.
func (l Linux) Disable(string) error { return l.set("org.gnome.system.proxy", "mode", "none") }
