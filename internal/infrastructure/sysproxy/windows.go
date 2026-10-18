package sysproxy

import (
	"strconv"
	"strings"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
)

// WindowsService is the one "service" of Windows: the user's Internet
// Settings, which WinINet and most apps follow.
const WindowsService = "Internet Settings"

const internetSettings = `HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`

// Windows drives the Internet Settings registry key with reg.exe.
type Windows struct{ Runner Runner }

func (w Windows) Services() ([]string, error)    { return []string{WindowsService}, nil }
func (w Windows) ActiveService() (string, error) { return WindowsService, nil }

func (w Windows) query(name string) string {
	out, err := w.Runner.Run("reg", "query", internetSettings, "/v", name)
	if err != nil {
		return ""
	}
	for line := range strings.SplitSeq(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == name {
			return strings.Join(fields[2:], " ")
		}
	}
	return ""
}

func (w Windows) add(name, kind, value string) error {
	_, err := w.Runner.Run("reg", "add", internetSettings, "/v", name, "/t", kind, "/d", value, "/f")
	return err
}

// Observe reads ProxyEnable, ProxyServer and AutoConfigURL.
func (w Windows) Observe(string) (systemproxy.Snapshot, error) {
	snap := systemproxy.Snapshot{Service: WindowsService, CapturedAt: time.Now()}
	enabled := w.query("ProxyEnable") == "0x1"
	server := w.query("ProxyServer")
	if i := strings.Index(server, "="); i >= 0 {
		server = strings.Split(server[i+1:], ";")[0]
	}
	host, portText, _ := strings.Cut(server, ":")
	port, _ := strconv.Atoi(portText)
	p := systemproxy.Proxy{Enabled: enabled, Host: host, Port: port}
	snap.Web, snap.Secure, snap.SOCKS = p, p, p
	snap.PACURL = w.query("AutoConfigURL")
	snap.AutoProxy = snap.PACURL != ""
	return snap, nil
}

// ApplyManual sets one proxy for every protocol.
func (w Windows) ApplyManual(_ string, host string, port int, bypass []string) error {
	if err := w.add("ProxyServer", "REG_SZ", host+":"+strconv.Itoa(port)); err != nil {
		return err
	}
	if err := w.add("ProxyOverride", "REG_SZ", strings.Join(bypass, ";")); err != nil {
		return err
	}
	_, _ = w.Runner.Run("reg", "delete", internetSettings, "/v", "AutoConfigURL", "/f")
	return w.add("ProxyEnable", "REG_DWORD", "1")
}

// ApplyPAC sets the auto-config URL.
func (w Windows) ApplyPAC(_ string, url string) error {
	if err := w.add("ProxyEnable", "REG_DWORD", "0"); err != nil {
		return err
	}
	return w.add("AutoConfigURL", "REG_SZ", url)
}

// Disable turns the proxy and the auto-config off.
func (w Windows) Disable(string) error {
	_, _ = w.Runner.Run("reg", "delete", internetSettings, "/v", "AutoConfigURL", "/f")
	return w.add("ProxyEnable", "REG_DWORD", "0")
}
