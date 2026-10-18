package sysproxy

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
)

const networksetup = "/usr/sbin/networksetup"

// Darwin drives macOS network services with networksetup.
type Darwin struct{ Runner Runner }

// Services lists network services, without disabled markers.
func (d Darwin) Services() ([]string, error) {
	out, err := d.Runner.Run(networksetup, "-listallnetworkservices")
	if err != nil {
		return nil, err
	}
	var services []string
	for line := range strings.SplitSeq(out, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "*"))
		if line == "" || strings.HasPrefix(line, "An asterisk") {
			continue
		}
		services = append(services, line)
	}
	return services, nil
}

// ActiveService maps the default route's interface to its service.
func (d Darwin) ActiveService() (string, error) {
	route, err := d.Runner.Run("/sbin/route", "-n", "get", "default")
	if err != nil {
		return "", err
	}
	device := ""
	for line := range strings.SplitSeq(route, "\n") {
		if value, ok := strings.CutPrefix(strings.TrimSpace(line), "interface:"); ok {
			device = strings.TrimSpace(value)
		}
	}
	if device == "" {
		return "", fmt.Errorf("no default route")
	}
	order, err := d.Runner.Run(networksetup, "-listnetworkserviceorder")
	if err != nil {
		return "", err
	}
	return ServiceForDevice(order, device)
}

// ServiceForDevice reads networksetup -listnetworkserviceorder output.
func ServiceForDevice(order, device string) (string, error) {
	for block := range strings.SplitSeq(order, "\n\n") {
		if !strings.Contains(block, "Device: "+device+")") && !strings.Contains(block, "Device: "+device+",") {
			continue
		}
		for line := range strings.SplitSeq(block, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "(") && !strings.HasPrefix(line, "(Hardware") {
				if _, name, ok := strings.Cut(line, ") "); ok {
					return strings.TrimSpace(name), nil
				}
			}
		}
	}
	return "", fmt.Errorf("no network service uses %s", device)
}

func parseProxy(out string) systemproxy.Proxy {
	var p systemproxy.Proxy
	for line := range strings.SplitSeq(out, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "Enabled":
			p.Enabled = value == "Yes"
		case "Server":
			p.Host = value
		case "Port":
			p.Port, _ = strconv.Atoi(value)
		}
	}
	return p
}

// Observe reads every proxy setting of a service.
func (d Darwin) Observe(service string) (systemproxy.Snapshot, error) {
	snap := systemproxy.Snapshot{Service: service, CapturedAt: time.Now()}
	for _, item := range []struct {
		flag   string
		target *systemproxy.Proxy
	}{{"-getwebproxy", &snap.Web}, {"-getsecurewebproxy", &snap.Secure}, {"-getsocksfirewallproxy", &snap.SOCKS}} {
		out, err := d.Runner.Run(networksetup, item.flag, service)
		if err != nil {
			return snap, err
		}
		*item.target = parseProxy(out)
	}
	auto, err := d.Runner.Run(networksetup, "-getautoproxyurl", service)
	if err != nil {
		return snap, err
	}
	for line := range strings.SplitSeq(auto, "\n") {
		key, value, _ := strings.Cut(line, ":")
		switch strings.TrimSpace(key) {
		case "URL":
			snap.PACURL = strings.TrimSpace(value)
			if snap.PACURL == "(null)" {
				snap.PACURL = ""
			}
		case "Enabled":
			snap.AutoProxy = strings.TrimSpace(value) == "Yes"
		}
	}
	if bypass, err := d.Runner.Run(networksetup, "-getproxybypassdomains", service); err == nil && !strings.Contains(bypass, "There aren't any") {
		snap.Bypass = strings.Fields(bypass)
	}
	return snap, nil
}

// ApplyManual sets the web, secure web and SOCKS proxies and turns PAC off.
func (d Darwin) ApplyManual(service, host string, port int, bypass []string) error {
	p := strconv.Itoa(port)
	commands := [][]string{
		{"-setwebproxy", service, host, p},
		{"-setsecurewebproxy", service, host, p},
		{"-setsocksfirewallproxy", service, host, p},
		{"-setautoproxystate", service, "off"},
	}
	if len(bypass) > 0 {
		commands = append(commands, append([]string{"-setproxybypassdomains", service}, bypass...))
	}
	return d.run(commands)
}

// ApplyPAC turns the manual proxies off and points the auto proxy at url.
func (d Darwin) ApplyPAC(service, url string) error {
	return d.run([][]string{
		{"-setwebproxystate", service, "off"},
		{"-setsecurewebproxystate", service, "off"},
		{"-setsocksfirewallproxystate", service, "off"},
		{"-setautoproxyurl", service, url},
		{"-setautoproxystate", service, "on"},
	})
}

// Disable turns every proxy of the service off.
func (d Darwin) Disable(service string) error {
	return d.run([][]string{
		{"-setwebproxystate", service, "off"},
		{"-setsecurewebproxystate", service, "off"},
		{"-setsocksfirewallproxystate", service, "off"},
		{"-setautoproxystate", service, "off"},
	})
}

func (d Darwin) run(commands [][]string) error {
	for _, args := range commands {
		if _, err := d.Runner.Run(networksetup, args...); err != nil {
			return err
		}
	}
	return nil
}
