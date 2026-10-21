package core

import (
	"encoding/json"
	"slices"
	"strings"
)

// RuntimeSettings are the parts of the generated configuration that Kumo
// owns. They replace whatever the profile says for the same keys.
type RuntimeSettings struct {
	MixedPort       int             `json:"mixedPort"`
	AllowLAN        bool            `json:"allowLAN"`
	LogLevel        string          `json:"logLevel"`
	IPv6            bool            `json:"ipv6"`
	FindProcessMode string          `json:"findProcessMode"`
	GeoData         GeoDataSettings `json:"geoData"`
	TUN             TUNSettings     `json:"tun"`
	DNS             DNSSettings     `json:"dns"`
	Sniffer         SnifferSettings `json:"sniffer"`
}

// DefaultRuntimeSettings mirrors the defaults Kumo has always shipped.
func DefaultRuntimeSettings() RuntimeSettings {
	return RuntimeSettings{
		MixedPort:       7890,
		LogLevel:        "info",
		FindProcessMode: "always",
		GeoData:         DefaultGeoDataSettings(),
		TUN:             DefaultTUNSettings(),
		DNS:             DefaultDNSSettings(),
		Sniffer:         DefaultSnifferSettings(),
	}
}

// UnmarshalJSON fills fields a stored document leaves out with defaults, so
// older state files keep loading.
func (s *RuntimeSettings) UnmarshalJSON(data []byte) error {
	type plain RuntimeSettings
	value := plain(DefaultRuntimeSettings())
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = RuntimeSettings(value)
	return nil
}

// LogLevels are the levels Mihomo accepts.
var LogLevels = []string{"silent", "error", "warning", "info", "debug"}

// FindProcessModes are the process lookup modes Mihomo accepts.
var FindProcessModes = []string{"always", "strict", "off"}

// Normalized trims and clamps every field to values Mihomo accepts.
func (s RuntimeSettings) Normalized() RuntimeSettings {
	if !slices.Contains(LogLevels, s.LogLevel) {
		s.LogLevel = "info"
	}
	if !slices.Contains(FindProcessModes, s.FindProcessMode) {
		s.FindProcessMode = "always"
	}
	s.TUN = s.TUN.Normalized()
	s.DNS = s.DNS.Normalized()
	s.Sniffer = s.Sniffer.Normalized()
	return s
}

// GeoDataSettings choose where Mihomo downloads its GeoIP and GeoSite data.
type GeoDataSettings struct {
	GeoIPURL            string `json:"geoIPURL"`
	GeoSiteURL          string `json:"geoSiteURL"`
	MMDBURL             string `json:"mmdbURL"`
	ASNURL              string `json:"asnURL"`
	AutoUpdate          bool   `json:"autoUpdate"`
	UpdateIntervalHours int    `json:"updateIntervalHours"`
	UsesDatMode         bool   `json:"usesDatMode"`
}

const metaRules = "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/"

// DefaultGeoDataSettings points at MetaCubeX's published data.
func DefaultGeoDataSettings() GeoDataSettings {
	return GeoDataSettings{
		GeoIPURL:            metaRules + "geoip.dat",
		GeoSiteURL:          metaRules + "geosite.dat",
		MMDBURL:             metaRules + "geoip.metadb",
		ASNURL:              metaRules + "GeoLite2-ASN.mmdb",
		UpdateIntervalHours: 24,
	}
}

// TUNSettings configure Mihomo's TUN inbound.
type TUNSettings struct {
	Enabled               bool     `json:"isEnabled"`
	Stack                 string   `json:"stack"`
	AutoRoute             bool     `json:"autoRoute"`
	AutoRedirect          bool     `json:"autoRedirect"`
	AutoDetectInterface   bool     `json:"autoDetectInterface"`
	StrictRoute           bool     `json:"strictRoute"`
	DisableICMPForwarding bool     `json:"disableICMPForwarding"`
	DNSHijack             []string `json:"dnsHijack"`
	RouteExcludeAddress   []string `json:"routeExcludeAddress"`
	MTU                   int      `json:"mtu"`
	Device                string   `json:"device,omitempty"`
}

// TUNStacks are the network stacks Mihomo implements.
var TUNStacks = []string{"mixed", "gvisor", "system"}

// DefaultTUNSettings leave TUN off with Mihomo's recommended routing.
func DefaultTUNSettings() TUNSettings {
	return TUNSettings{
		Stack:               "mixed",
		AutoRoute:           true,
		AutoDetectInterface: true,
		DNSHijack:           []string{"any:53"},
		RouteExcludeAddress: []string{},
		MTU:                 1500,
	}
}

func (s *TUNSettings) UnmarshalJSON(data []byte) error {
	type plain TUNSettings
	value := plain(DefaultTUNSettings())
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = TUNSettings(value)
	return nil
}

// Normalized clamps the MTU and drops empty list entries.
func (s TUNSettings) Normalized() TUNSettings {
	s.MTU = max(576, min(9000, s.MTU))
	if !slices.Contains(TUNStacks, s.Stack) {
		s.Stack = "mixed"
	}
	s.DNSHijack = NormalizeList(s.DNSHijack)
	s.RouteExcludeAddress = NormalizeList(s.RouteExcludeAddress)
	s.Device = strings.TrimSpace(s.Device)
	return s
}

// DNSSettings configure Mihomo's DNS server. Enabled means Kumo writes its
// own dns section, replacing the profile's.
type DNSSettings struct {
	Enabled                      bool                   `json:"isEnabled"`
	Listen                       string                 `json:"listen"`
	IPv6                         bool                   `json:"ipv6"`
	IPv6Timeout                  int                    `json:"ipv6Timeout"`
	PreferH3                     bool                   `json:"preferH3"`
	EnhancedMode                 string                 `json:"enhancedMode"`
	FakeIPRange                  string                 `json:"fakeIPRange"`
	FakeIPRange6                 string                 `json:"fakeIPRange6"`
	FakeIPFilter                 []string               `json:"fakeIPFilter"`
	FakeIPFilterMode             string                 `json:"fakeIPFilterMode"`
	UseHosts                     bool                   `json:"useHosts"`
	UseSystemHosts               bool                   `json:"useSystemHosts"`
	RespectRules                 bool                   `json:"respectRules"`
	DefaultNameserver            []string               `json:"defaultNameserver"`
	Nameserver                   []string               `json:"nameserver"`
	Fallback                     []string               `json:"fallback"`
	FallbackFilter               FallbackFilter         `json:"fallbackFilter"`
	ProxyServerNameserver        []string               `json:"proxyServerNameserver"`
	DirectNameserver             []string               `json:"directNameserver"`
	DirectNameserverFollowPolicy bool                   `json:"directNameserverFollowPolicy"`
	NameserverPolicy             map[string]PolicyValue `json:"nameserverPolicy"`
	ProxyServerNameserverPolicy  map[string]PolicyValue `json:"proxyServerNameserverPolicy"`
	CacheAlgorithm               string                 `json:"cacheAlgorithm"`
	Hosts                        map[string]PolicyValue `json:"hosts"`
}

// EnhancedModes are the DNS modes Mihomo implements.
var EnhancedModes = []string{"fake-ip", "redir-host", "normal"}

// DefaultDNSSettings mirror the fake-ip setup Kumo has always shipped. They
// start disabled: until the user turns them on, the profile's own dns
// section applies.
func DefaultDNSSettings() DNSSettings {
	return DNSSettings{
		IPv6Timeout:                 100,
		EnhancedMode:                "fake-ip",
		FakeIPRange:                 "198.18.0.1/16",
		FakeIPRange6:                "fc00::/18",
		FakeIPFilter:                []string{"*", "+.lan", "+.local", "time.*.com", "ntp.*.com", "+.market.xiaomi.com"},
		FakeIPFilterMode:            "blacklist",
		DefaultNameserver:           []string{"tls://223.5.5.5"},
		Nameserver:                  []string{"https://doh.pub/dns-query", "https://dns.alidns.com/dns-query"},
		Fallback:                    []string{},
		ProxyServerNameserver:       []string{},
		DirectNameserver:            []string{},
		NameserverPolicy:            map[string]PolicyValue{},
		ProxyServerNameserverPolicy: map[string]PolicyValue{},
		Hosts:                       map[string]PolicyValue{},
	}
}

func (s *DNSSettings) UnmarshalJSON(data []byte) error {
	type plain DNSSettings
	value := plain(DefaultDNSSettings())
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = DNSSettings(value)
	return nil
}

// Normalized trims every value and falls back to fake-ip for an unknown mode.
func (s DNSSettings) Normalized() DNSSettings {
	if !slices.Contains(EnhancedModes, s.EnhancedMode) {
		s.EnhancedMode = "fake-ip"
	}
	s.Listen = strings.TrimSpace(s.Listen)
	s.FakeIPRange = strings.TrimSpace(s.FakeIPRange)
	s.FakeIPRange6 = strings.TrimSpace(s.FakeIPRange6)
	s.FakeIPFilter = NormalizeList(s.FakeIPFilter)
	s.FakeIPFilterMode = strings.TrimSpace(s.FakeIPFilterMode)
	s.DefaultNameserver = NormalizeList(s.DefaultNameserver)
	s.Nameserver = NormalizeList(s.Nameserver)
	s.Fallback = NormalizeList(s.Fallback)
	s.ProxyServerNameserver = NormalizeList(s.ProxyServerNameserver)
	s.DirectNameserver = NormalizeList(s.DirectNameserver)
	s.FallbackFilter = s.FallbackFilter.Normalized()
	s.NameserverPolicy = normalizePolicies(s.NameserverPolicy)
	s.ProxyServerNameserverPolicy = normalizePolicies(s.ProxyServerNameserverPolicy)
	s.Hosts = normalizePolicies(s.Hosts)
	s.CacheAlgorithm = strings.TrimSpace(s.CacheAlgorithm)
	return s
}

// PolicyValue is one or more values of a DNS policy or hosts entry. Mihomo
// accepts a single string or a list; one value renders as a string.
type PolicyValue []string

// UnmarshalJSON accepts a string as well as a list, as older state files
// stored single values as strings.
func (p *PolicyValue) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*p = PolicyValue{single}
		return nil
	}
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*p = list
	return nil
}

func normalizePolicies(values map[string]PolicyValue) map[string]PolicyValue {
	result := make(map[string]PolicyValue, len(values))
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		result[key] = NormalizeList(value)
	}
	return result
}

// FallbackFilter decides when Mihomo prefers the fallback nameservers.
type FallbackFilter struct {
	GeoIP     bool     `json:"geoip"`
	GeoIPCode string   `json:"geoipCode"`
	GeoSite   []string `json:"geosite"`
	IPCIDR    []string `json:"ipcidr"`
	Domain    []string `json:"domain"`
}

// IsZero reports whether the filter sets nothing, so it is left out.
func (f FallbackFilter) IsZero() bool {
	return !f.GeoIP && f.GeoIPCode == "" && len(f.GeoSite) == 0 && len(f.IPCIDR) == 0 && len(f.Domain) == 0
}

// Normalized trims the filter's values.
func (f FallbackFilter) Normalized() FallbackFilter {
	f.GeoIPCode = strings.TrimSpace(f.GeoIPCode)
	f.GeoSite = NormalizeList(f.GeoSite)
	f.IPCIDR = NormalizeList(f.IPCIDR)
	f.Domain = NormalizeList(f.Domain)
	return f
}

// UnmarshalJSON also reads the dictionary older state files stored, keyed by
// Mihomo's own names.
func (f *FallbackFilter) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	list := func(keys ...string) []string {
		for _, key := range keys {
			if value, ok := raw[key]; ok {
				var p PolicyValue
				if json.Unmarshal(value, &p) == nil {
					return p
				}
			}
		}
		return nil
	}
	*f = FallbackFilter{GeoSite: list("geosite"), IPCIDR: list("ipcidr"), Domain: list("domain")}
	if value, ok := raw["geoip"]; ok {
		_ = json.Unmarshal(value, &f.GeoIP)
	}
	for _, key := range []string{"geoipCode", "geoip-code"} {
		if value, ok := raw[key]; ok {
			_ = json.Unmarshal(value, &f.GeoIPCode)
		}
	}
	return nil
}

// SnifferSettings configure Mihomo's domain sniffer. Enabled means Kumo
// writes its own sniffer section, replacing the profile's.
type SnifferSettings struct {
	Enabled                 bool     `json:"isEnabled"`
	ParsePureIP             bool     `json:"parsePureIP"`
	ForceDNSMapping         bool     `json:"forceDNSMapping"`
	OverrideDestination     bool     `json:"overrideDestination"`
	HTTPOverrideDestination bool     `json:"httpOverrideDestination"`
	HTTPPorts               []int    `json:"httpPorts"`
	TLSPorts                []int    `json:"tlsPorts"`
	QUICPorts               []int    `json:"quicPorts"`
	SkipDomain              []string `json:"skipDomain"`
	ForceDomain             []string `json:"forceDomain"`
	SkipDstAddress          []string `json:"skipDstAddress"`
	SkipSrcAddress          []string `json:"skipSrcAddress"`
}

// DefaultSnifferSettings sniff HTTP and TLS and skip Apple push and Telegram.
// Like DNS, they start disabled so the profile's sniffer section applies.
func DefaultSnifferSettings() SnifferSettings {
	return SnifferSettings{
		ParsePureIP:     true,
		ForceDNSMapping: true,
		HTTPPorts:       []int{80, 443},
		TLSPorts:        []int{443},
		QUICPorts:       []int{},
		SkipDomain:      []string{"+.push.apple.com"},
		ForceDomain:     []string{},
		SkipDstAddress: []string{
			"91.105.192.0/23", "91.108.4.0/22", "91.108.8.0/21", "91.108.16.0/21",
			"91.108.56.0/22", "95.161.64.0/20", "149.154.160.0/20", "185.76.151.0/24",
			"2001:67c:4e8::/48", "2001:b28:f23c::/47", "2001:b28:f23f::/48", "2a0a:f280:203::/48",
		},
		SkipSrcAddress: []string{},
	}
}

func (s *SnifferSettings) UnmarshalJSON(data []byte) error {
	type plain SnifferSettings
	value := plain(DefaultSnifferSettings())
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = SnifferSettings(value)
	return nil
}

// Normalized drops invalid ports and empty entries.
func (s SnifferSettings) Normalized() SnifferSettings {
	s.HTTPPorts = validPorts(s.HTTPPorts)
	s.TLSPorts = validPorts(s.TLSPorts)
	s.QUICPorts = validPorts(s.QUICPorts)
	s.SkipDomain = NormalizeList(s.SkipDomain)
	s.ForceDomain = NormalizeList(s.ForceDomain)
	s.SkipDstAddress = NormalizeList(s.SkipDstAddress)
	s.SkipSrcAddress = NormalizeList(s.SkipSrcAddress)
	return s
}

func validPorts(ports []int) []int {
	result := []int{}
	for _, port := range ports {
		if ValidatePort(port) == nil {
			result = append(result, port)
		}
	}
	return result
}

// NormalizeList trims entries and drops empty ones. It never returns nil,
// so lists encode as [] rather than null.
func NormalizeList(values []string) []string {
	result := []string{}
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
