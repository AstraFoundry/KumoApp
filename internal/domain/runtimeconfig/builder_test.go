package runtimeconfig

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/override"
)

func build(t *testing.T, in Input) map[string]any {
	t.Helper()
	if in.Mode == "" {
		in.Mode = core.ModeRule
	}
	if in.Endpoint.Host == "" {
		in.Endpoint = core.DefaultEndpoint()
	}
	if in.Settings.MixedPort == 0 {
		in.Settings = core.DefaultRuntimeSettings()
	}
	out, err := Build(in)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := yaml.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("output is not YAML: %v\n%s", err, out)
	}
	return m
}

func TestControlledKeysWin(t *testing.T) {
	m := build(t, Input{ProfileYAML: "mixed-port: 1234\nmode: global\nexternal-controller: 0.0.0.0:1\nproxies: []\nrules:\n  - MATCH,DIRECT\n", Mode: core.ModeDirect})
	if m["mixed-port"] != 7890 || m["mode"] != "direct" || m["external-controller"] != "127.0.0.1:9097" {
		t.Fatalf("controlled keys not applied: %v", m)
	}
	if _, ok := m["rules"]; !ok {
		t.Fatal("profile keys dropped")
	}
}

func TestEmptyProfile(t *testing.T) {
	m := build(t, Input{ProfileYAML: ""})
	if m["mode"] != "rule" {
		t.Fatalf("got %v", m)
	}
}

func TestProfileSectionsKeptUntilManaged(t *testing.T) {
	profile := "dns:\n  enable: true\n  nameserver: [1.1.1.1]\n"
	m := build(t, Input{ProfileYAML: profile})
	if m["dns"].(map[string]any)["nameserver"].([]any)[0] != "1.1.1.1" {
		t.Fatal("profile dns replaced while unmanaged")
	}
	settings := core.DefaultRuntimeSettings()
	settings.DNS.Enabled = true
	settings.DNS.Hosts = map[string]core.PolicyValue{"a.lan": {"10.0.0.1"}, "b.lan": {"10.0.0.2", "10.0.0.3"}}
	m = build(t, Input{ProfileYAML: profile, Settings: settings})
	dns := m["dns"].(map[string]any)
	if dns["nameserver"].([]any)[0] != "https://doh.pub/dns-query" {
		t.Fatalf("managed dns not applied: %v", dns)
	}
	hosts := m["hosts"].(map[string]any)
	if hosts["a.lan"] != "10.0.0.1" || len(hosts["b.lan"].([]any)) != 2 {
		t.Fatalf("hosts: %v", hosts)
	}
}

func TestTUNAndSniffer(t *testing.T) {
	settings := core.DefaultRuntimeSettings()
	settings.TUN.Enabled = true
	settings.TUN.MTU = 100000
	settings.Sniffer.Enabled = true
	m := build(t, Input{Settings: settings})
	tun := m["tun"].(map[string]any)
	if tun["enable"] != true || tun["mtu"] != 9000 {
		t.Fatalf("tun: %v", tun)
	}
	sniff := m["sniffer"].(map[string]any)["sniff"].(map[string]any)
	if sniff["TLS"].(map[string]any)["ports"].([]any)[0] != 443 {
		t.Fatalf("sniffer: %v", sniff)
	}
}

func TestTUNStack(t *testing.T) {
	cases := map[string]string{"": "mips", "mips": "mips", "mixed": "mixed", "gvisor": "gvisor", "system": "system", "lwip": "mips"}
	for stack, want := range cases {
		settings := core.DefaultRuntimeSettings()
		settings.TUN.Enabled = true
		settings.TUN.Stack = stack
		m := build(t, Input{Settings: settings})
		if got := m["tun"].(map[string]any)["stack"]; got != want {
			t.Errorf("stack %q built as %v, want %s", stack, got, want)
		}
	}
}

func TestOverridesApplyInOrder(t *testing.T) {
	profile := "proxies:\n  - {name: a, type: direct}\n  - {name: b, type: direct}\nrules:\n  - MATCH,DIRECT\ndns:\n  enable: true\n  ipv6: false\n"
	m := build(t, Input{ProfileYAML: profile, Overrides: []override.Source{
		{Name: "one", Format: override.FormatYAML, Content: "+rules:\n  - DOMAIN,example.com,DIRECT\ndelete-proxies:\n  - a\ndns:\n  ipv6: true\n"},
		{Name: "two", Format: override.FormatYAML, Content: "rules+:\n  - GEOIP,CN,DIRECT\n<custom+>: 1\n"},
		{Name: "three", Format: override.FormatYAML, Content: "dns!:\n  enable: false\n"},
	}})
	rules := m["rules"].([]any)
	if len(rules) != 3 || rules[0] != "DOMAIN,example.com,DIRECT" || rules[2] != "GEOIP,CN,DIRECT" {
		t.Fatalf("rules: %v", rules)
	}
	if proxies := m["proxies"].([]any); len(proxies) != 1 {
		t.Fatalf("proxies: %v", proxies)
	}
	if dns := m["dns"].(map[string]any); len(dns) != 1 || dns["enable"] != false {
		t.Fatalf("dns: %v", dns)
	}
	if m["custom+"] != 1 {
		t.Fatalf("quoted key: %v", m)
	}
}

func TestInvalidProfile(t *testing.T) {
	if _, err := Build(Input{ProfileYAML: "- a\n- b\n", Settings: core.DefaultRuntimeSettings()}); err == nil || !strings.Contains(err.Error(), "mapping") {
		t.Fatalf("err = %v", err)
	}
	if _, err := Build(Input{ProfileYAML: "a: 1", Overrides: []override.Source{{Name: "js", Format: override.FormatJavaScript, Content: "function main(c){return c}"}}}); err == nil {
		t.Fatal("JavaScript without transformer accepted")
	}
}
