// Package runtimeconfig builds the configuration Mihomo runs with: the
// current profile, the user's overrides in order, and the settings Kumo
// controls, which always win.
package runtimeconfig

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/override"
)

// controlledKeys are top-level keys Kumo always writes itself.
var controlledKeys = []string{
	"external-controller", "secret", "port", "socks-port", "redir-port", "tproxy-port",
	"mixed-port", "mode", "allow-lan", "log-level", "ipv6", "find-process-mode",
	"geodata-mode", "geo-auto-update", "geo-update-interval", "geox-url",
}

// Transformer runs an override written in a scripting language. The builder
// hands it the configuration so far and takes back the result.
type Transformer interface {
	Transform(name, script string, config *yaml.Node) (*yaml.Node, error)
}

// Input is everything a build needs.
type Input struct {
	ProfileYAML string
	Overrides   []override.Source
	Mode        core.Mode
	Endpoint    core.Endpoint
	Settings    core.RuntimeSettings
	// Transformer runs JavaScript overrides; without one they are an error.
	Transformer Transformer
}

// Build returns the YAML Mihomo runs with.
func Build(in Input) (string, error) {
	root, err := parseMapping(in.ProfileYAML, "profile")
	if err != nil {
		return "", err
	}
	for _, source := range in.Overrides {
		switch source.Format {
		case override.FormatJavaScript:
			if in.Transformer == nil {
				return "", fmt.Errorf("override %q: JavaScript overrides are not supported here", source.Name)
			}
			transformed, err := in.Transformer.Transform(source.Name, source.Content, root)
			if err != nil {
				return "", fmt.Errorf("override %q: %w", source.Name, err)
			}
			root = transformed
		default:
			patch, err := parseMapping(source.Content, "override "+strconv.Quote(source.Name))
			if err != nil {
				return "", err
			}
			override.Merge(root, patch)
		}
	}

	settings := in.Settings.Normalized()
	controlled := Controlled(in.Mode, in.Endpoint, settings)
	for _, key := range controlledKeys {
		override.Delete(root, key)
	}
	for i := 0; i+1 < len(controlled.Content); i += 2 {
		override.Delete(root, controlled.Content[i].Value)
	}

	var out bytes.Buffer
	if len(root.Content) > 0 {
		if err := encode(&out, root); err != nil {
			return "", err
		}
		out.WriteString("\n")
	}
	out.WriteString("# Kumo controlled runtime settings\n")
	if err := encode(&out, controlled); err != nil {
		return "", err
	}
	return out.String(), nil
}

// Controlled is the mapping of every key Kumo writes for these settings.
// TUN, DNS, sniffer and hosts appear only when Kumo manages them; otherwise
// the profile's own sections apply.
func Controlled(mode core.Mode, endpoint core.Endpoint, s core.RuntimeSettings) *yaml.Node {
	m := mapping()
	put(m, "external-controller", str(endpoint.Address()))
	put(m, "secret", quoted(endpoint.Secret))
	put(m, "mixed-port", integer(s.MixedPort))
	put(m, "mode", str(string(mode)))
	put(m, "allow-lan", boolean(s.AllowLAN))
	put(m, "log-level", str(s.LogLevel))
	put(m, "ipv6", boolean(s.IPv6))
	put(m, "find-process-mode", str(s.FindProcessMode))
	put(m, "geodata-mode", boolean(s.GeoData.UsesDatMode))
	put(m, "geo-auto-update", boolean(s.GeoData.AutoUpdate))
	put(m, "geo-update-interval", integer(s.GeoData.UpdateIntervalHours))
	geox := mapping()
	put(geox, "geoip", quoted(s.GeoData.GeoIPURL))
	put(geox, "geosite", quoted(s.GeoData.GeoSiteURL))
	put(geox, "mmdb", quoted(s.GeoData.MMDBURL))
	put(geox, "asn", quoted(s.GeoData.ASNURL))
	put(m, "geox-url", geox)
	if s.TUN.Enabled {
		put(m, "tun", tun(s.TUN))
	}
	if s.DNS.Enabled {
		put(m, "dns", dns(s.DNS))
		if len(s.DNS.Hosts) > 0 {
			put(m, "hosts", policies(s.DNS.Hosts))
		}
	}
	if s.Sniffer.Enabled {
		put(m, "sniffer", sniffer(s.Sniffer))
	}
	return m
}

func tun(t core.TUNSettings) *yaml.Node {
	m := mapping()
	put(m, "enable", boolean(true))
	put(m, "stack", str(t.Stack))
	put(m, "auto-route", boolean(t.AutoRoute))
	put(m, "auto-redirect", boolean(t.AutoRedirect))
	put(m, "auto-detect-interface", boolean(t.AutoDetectInterface))
	put(m, "strict-route", boolean(t.StrictRoute))
	put(m, "disable-icmp-forwarding", boolean(t.DisableICMPForwarding))
	put(m, "dns-hijack", strings_(t.DNSHijack))
	if len(t.RouteExcludeAddress) > 0 {
		put(m, "route-exclude-address", strings_(t.RouteExcludeAddress))
	}
	put(m, "mtu", integer(t.MTU))
	if t.Device != "" {
		put(m, "device", str(t.Device))
	}
	return m
}

func dns(d core.DNSSettings) *yaml.Node {
	m := mapping()
	put(m, "enable", boolean(true))
	if d.Listen != "" {
		put(m, "listen", quoted(d.Listen))
	}
	put(m, "ipv6", boolean(d.IPv6))
	put(m, "ipv6-timeout", integer(d.IPv6Timeout))
	put(m, "prefer-h3", boolean(d.PreferH3))
	put(m, "enhanced-mode", str(d.EnhancedMode))
	put(m, "fake-ip-range", str(d.FakeIPRange))
	if d.FakeIPRange6 != "" {
		put(m, "fake-ip-range6", str(d.FakeIPRange6))
	}
	optionalList(m, "fake-ip-filter", d.FakeIPFilter)
	if d.FakeIPFilterMode != "" {
		put(m, "fake-ip-filter-mode", str(d.FakeIPFilterMode))
	}
	put(m, "use-hosts", boolean(d.UseHosts))
	put(m, "use-system-hosts", boolean(d.UseSystemHosts))
	put(m, "respect-rules", boolean(d.RespectRules))
	optionalList(m, "default-nameserver", d.DefaultNameserver)
	optionalList(m, "nameserver", d.Nameserver)
	optionalList(m, "fallback", d.Fallback)
	if !d.FallbackFilter.IsZero() {
		f := mapping()
		put(f, "geoip", boolean(d.FallbackFilter.GeoIP))
		if d.FallbackFilter.GeoIPCode != "" {
			put(f, "geoip-code", str(d.FallbackFilter.GeoIPCode))
		}
		optionalList(f, "geosite", d.FallbackFilter.GeoSite)
		optionalList(f, "ipcidr", d.FallbackFilter.IPCIDR)
		optionalList(f, "domain", d.FallbackFilter.Domain)
		put(m, "fallback-filter", f)
	}
	optionalList(m, "proxy-server-nameserver", d.ProxyServerNameserver)
	optionalList(m, "direct-nameserver", d.DirectNameserver)
	put(m, "direct-nameserver-follow-policy", boolean(d.DirectNameserverFollowPolicy))
	if len(d.NameserverPolicy) > 0 {
		put(m, "nameserver-policy", policies(d.NameserverPolicy))
	}
	if len(d.ProxyServerNameserverPolicy) > 0 {
		put(m, "proxy-server-nameserver-policy", policies(d.ProxyServerNameserverPolicy))
	}
	if d.CacheAlgorithm != "" {
		put(m, "cache-algorithm", str(d.CacheAlgorithm))
	}
	return m
}

func sniffer(s core.SnifferSettings) *yaml.Node {
	m := mapping()
	put(m, "enable", boolean(true))
	put(m, "parse-pure-ip", boolean(s.ParsePureIP))
	put(m, "force-dns-mapping", boolean(s.ForceDNSMapping))
	put(m, "override-destination", boolean(s.OverrideDestination))
	sniff := mapping()
	if len(s.HTTPPorts) > 0 || s.HTTPOverrideDestination {
		http := mapping()
		if len(s.HTTPPorts) > 0 {
			put(http, "ports", ints(s.HTTPPorts))
		}
		if s.HTTPOverrideDestination {
			put(http, "override-destination", boolean(true))
		}
		put(sniff, "HTTP", http)
	}
	if len(s.TLSPorts) > 0 {
		tls := mapping()
		put(tls, "ports", ints(s.TLSPorts))
		put(sniff, "TLS", tls)
	}
	if len(s.QUICPorts) > 0 {
		quic := mapping()
		put(quic, "ports", ints(s.QUICPorts))
		put(sniff, "QUIC", quic)
	}
	if len(sniff.Content) > 0 {
		put(m, "sniff", sniff)
	}
	optionalList(m, "skip-domain", s.SkipDomain)
	optionalList(m, "force-domain", s.ForceDomain)
	optionalList(m, "skip-dst-address", s.SkipDstAddress)
	optionalList(m, "skip-src-address", s.SkipSrcAddress)
	return m
}

func policies(values map[string]core.PolicyValue) *yaml.Node {
	m := mapping()
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		value := values[key]
		if len(value) == 1 {
			put(m, key, quoted(value[0]))
		} else {
			put(m, key, quotedList(value))
		}
	}
	return m
}

func parseMapping(source, what string) (*yaml.Node, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(source), &doc); err != nil {
		return nil, fmt.Errorf("%s is not valid YAML: %w", what, err)
	}
	if doc.Kind == 0 || len(doc.Content) == 0 {
		return mapping(), nil
	}
	root := doc.Content[0]
	if root.Kind == yaml.ScalarNode && root.Tag == "!!null" {
		return mapping(), nil
	}
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s must be a YAML mapping", what)
	}
	return root, nil
}

func encode(out *bytes.Buffer, node *yaml.Node) error {
	enc := yaml.NewEncoder(out)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return err
	}
	return enc.Close()
}

func mapping() *yaml.Node { return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"} }

func put(m *yaml.Node, key string, value *yaml.Node) {
	m.Content = append(m.Content, str(key), value)
}

func str(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}

func quoted(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value, Style: yaml.DoubleQuotedStyle}
}

func integer(value int) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.Itoa(value)}
}

func boolean(value bool) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(value)}
}

func strings_(values []string) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, value := range values {
		seq.Content = append(seq.Content, str(value))
	}
	return seq
}

func quotedList(values []string) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, value := range values {
		seq.Content = append(seq.Content, quoted(value))
	}
	return seq
}

func ints(values []int) *yaml.Node {
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Style: yaml.FlowStyle}
	for _, value := range values {
		seq.Content = append(seq.Content, integer(value))
	}
	return seq
}

func optionalList(m *yaml.Node, key string, values []string) {
	if len(values) > 0 {
		put(m, key, strings_(values))
	}
}

// Parse reads YAML into a mapping node; an empty document is an empty
// mapping. Other packages use it to inspect configurations.
func Parse(source string) (*yaml.Node, error) { return parseMapping(source, "configuration") }

// ValidateProfile checks that a profile is a YAML mapping.
func ValidateProfile(source string) error {
	if strings.TrimSpace(source) == "" {
		return fmt.Errorf("the profile is empty")
	}
	_, err := parseMapping(source, "profile")
	return err
}
