// Package mihomo talks to the Mihomo core: it launches the process and
// calls its external controller.
package mihomo

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/inspect"
	"github.com/AstraFoundry/KumoApp/internal/domain/proxy"
)

// DefaultTestURL is the URL delay tests open when a group names none.
const DefaultTestURL = "https://www.gstatic.com/generate_204"

// Client calls one controller endpoint.
type Client struct {
	endpoint core.Endpoint
	http     *http.Client
	stream   *http.Client
}

// NewClient returns a client for the controller at endpoint. Requests never
// go through a proxy: the controller is local.
func NewClient(endpoint core.Endpoint) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	return &Client{
		endpoint: endpoint,
		http:     &http.Client{Timeout: 10 * time.Second, Transport: transport},
		stream:   &http.Client{Transport: transport},
	}
}

// Error is a non-2xx answer of the controller.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("mihomo: %s (HTTP %d)", e.Message, e.Status)
	}
	return fmt.Sprintf("mihomo: HTTP %d", e.Status)
}

func (c *Client) request(ctx context.Context, method, path string, query url.Values, body any) (*http.Request, error) {
	u := c.endpoint.BaseURL() + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.endpoint.Secret != "" {
		req.Header.Set("Authorization", "Bearer "+c.endpoint.Secret)
	}
	return req, nil
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	req, err := c.request(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var payload struct {
			Message string `json:"message"`
		}
		_ = json.Unmarshal(data, &payload)
		return &Error{Status: resp.StatusCode, Message: payload.Message}
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

func escape(name string) string { return url.PathEscape(name) }

// Version asks the core for its version; it answers once the controller
// listens.
func (c *Client) Version(ctx context.Context) (inspect.Version, error) {
	var v inspect.Version
	return v, c.do(ctx, http.MethodGet, "/version", nil, nil, &v)
}

// Configs is the subset of /configs Kumo reads.
type Configs struct {
	Mode      string `json:"mode"`
	MixedPort int    `json:"mixed-port"`
	LogLevel  string `json:"log-level"`
	AllowLAN  bool   `json:"allow-lan"`
	IPv6      bool   `json:"ipv6"`
	TUN       struct {
		Enable bool `json:"enable"`
	} `json:"tun"`
}

// Configs reads the running configuration's settings.
func (c *Client) Configs(ctx context.Context) (Configs, error) {
	var cfg Configs
	return cfg, c.do(ctx, http.MethodGet, "/configs", nil, nil, &cfg)
}

// Patch changes settings of the running core.
func (c *Client) Patch(ctx context.Context, patch map[string]any) error {
	return c.do(ctx, http.MethodPatch, "/configs", nil, patch, nil)
}

// Reload makes the core read its configuration file again.
func (c *Client) Reload(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodPut, "/configs", url.Values{"force": {"true"}}, map[string]string{"path": path}, nil)
}

// SetMode switches the outbound mode.
func (c *Client) SetMode(ctx context.Context, mode core.Mode) error {
	return c.Patch(ctx, map[string]any{"mode": string(mode)})
}

type rawProxy struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Now     string   `json:"now"`
	All     []string `json:"all"`
	TestURL string   `json:"testUrl"`
	Hidden  bool     `json:"hidden"`
	Icon    string   `json:"icon"`
	Alive   bool     `json:"alive"`
	UDP     bool     `json:"udp"`
	History []struct {
		Delay int `json:"delay"`
	} `json:"history"`
	ExtraByURL map[string]struct {
		History []struct {
			Delay int `json:"delay"`
		} `json:"history"`
	} `json:"extra"`
}

func (p rawProxy) delay(testURL string) int {
	if extra, ok := p.ExtraByURL[testURL]; ok && len(extra.History) > 0 {
		return delayValue(extra.History[len(extra.History)-1].Delay, p.Alive)
	}
	if len(p.History) == 0 {
		return 0
	}
	return delayValue(p.History[len(p.History)-1].Delay, p.Alive)
}

func delayValue(delay int, alive bool) int {
	if delay == 0 && !alive {
		return -1
	}
	return delay
}

// Groups returns the proxy groups in the order of the GLOBAL group, which
// follows the configuration, with GLOBAL last.
func (c *Client) Groups(ctx context.Context) ([]proxy.Group, error) {
	var payload struct {
		Proxies map[string]rawProxy `json:"proxies"`
	}
	if err := c.do(ctx, http.MethodGet, "/proxies", nil, nil, &payload); err != nil {
		return nil, err
	}
	all := payload.Proxies
	order := map[string]int{}
	if global, ok := all["GLOBAL"]; ok {
		for i, name := range global.All {
			order[name] = i
		}
	}
	var groups []proxy.Group
	for name, p := range all {
		if len(p.All) == 0 || !proxy.IsGroupType(p.Type) {
			continue
		}
		group := proxy.Group{Name: name, Type: p.Type, Now: p.Now, TestURL: p.TestURL, Hidden: p.Hidden, Icon: p.Icon, Live: true}
		for _, member := range p.All {
			m := all[member]
			group.Nodes = append(group.Nodes, proxy.Node{
				Name: member, Type: m.Type, Delay: m.delay(p.TestURL), Alive: m.Alive, UDP: m.UDP,
				Country: proxy.Country(member), Group: len(m.All) > 0,
			})
		}
		groups = append(groups, group)
	}
	sort.SliceStable(groups, func(i, j int) bool {
		a, b := groups[i], groups[j]
		if (a.Name == "GLOBAL") != (b.Name == "GLOBAL") {
			return b.Name == "GLOBAL"
		}
		oa, okA := order[a.Name]
		ob, okB := order[b.Name]
		if okA && okB {
			return oa < ob
		}
		if okA != okB {
			return okA
		}
		return a.Name < b.Name
	})
	return groups, nil
}

// Select makes name the member group uses.
func (c *Client) Select(ctx context.Context, group, name string) error {
	return c.do(ctx, http.MethodPut, "/proxies/"+escape(group), nil, map[string]string{"name": name}, nil)
}

// Delay measures one proxy; a timeout is -1.
func (c *Client) Delay(ctx context.Context, name, testURL string, timeout time.Duration) (int, error) {
	if testURL == "" {
		testURL = DefaultTestURL
	}
	var out struct {
		Delay int `json:"delay"`
	}
	query := url.Values{"url": {testURL}, "timeout": {strconv.Itoa(int(timeout.Milliseconds()))}}
	err := c.do(ctx, http.MethodGet, "/proxies/"+escape(name)+"/delay", query, nil, &out)
	var apiErr *Error
	if errors.As(err, &apiErr) && (apiErr.Status == http.StatusRequestTimeout || apiErr.Status == http.StatusServiceUnavailable || apiErr.Status == http.StatusGatewayTimeout) {
		return -1, nil
	}
	return out.Delay, err
}

// GroupDelay measures every member of a group at once.
func (c *Client) GroupDelay(ctx context.Context, group, testURL string, timeout time.Duration) (map[string]int, error) {
	if testURL == "" {
		testURL = DefaultTestURL
	}
	out := map[string]int{}
	query := url.Values{"url": {testURL}, "timeout": {strconv.Itoa(int(timeout.Milliseconds()))}}
	return out, c.do(ctx, http.MethodGet, "/group/"+escape(group)+"/delay", query, nil, &out)
}

// Rules lists the running rules.
func (c *Client) Rules(ctx context.Context) ([]inspect.Rule, error) {
	var payload struct {
		Rules []struct {
			Index   *int   `json:"index"`
			Type    string `json:"type"`
			Payload string `json:"payload"`
			Proxy   string `json:"proxy"`
			Size    int    `json:"size"`
			Extra   *struct {
				Disabled bool   `json:"disabled"`
				HitCount int64  `json:"hitCount"`
				HitAt    string `json:"hitAt"`
			} `json:"extra"`
		} `json:"rules"`
	}
	if err := c.do(ctx, http.MethodGet, "/rules", nil, nil, &payload); err != nil {
		return nil, err
	}
	rules := make([]inspect.Rule, 0, len(payload.Rules))
	for i, r := range payload.Rules {
		rule := inspect.Rule{Index: i, Type: r.Type, Payload: r.Payload, Proxy: r.Proxy, Size: r.Size, Enabled: true}
		if r.Index != nil {
			rule.Index = *r.Index
		}
		if r.Extra != nil {
			rule.Enabled = !r.Extra.Disabled
			rule.HitCount = r.Extra.HitCount
			rule.HitAt = r.Extra.HitAt
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// SetRuleEnabled turns one rule on or off without editing the profile.
func (c *Client) SetRuleEnabled(ctx context.Context, index int, enabled bool) error {
	return c.do(ctx, http.MethodPatch, "/rules/disable", nil, map[string]bool{strconv.Itoa(index): !enabled}, nil)
}

type rawConnection struct {
	ID       string `json:"id"`
	Metadata struct {
		Network         string `json:"network"`
		Host            string `json:"host"`
		DestinationIP   string `json:"destinationIP"`
		DestinationPort string `json:"destinationPort"`
		Process         string `json:"process"`
		ProcessPath     string `json:"processPath"`
		SniffHost       string `json:"sniffHost"`
	} `json:"metadata"`
	Upload      int64     `json:"upload"`
	Download    int64     `json:"download"`
	Start       time.Time `json:"start"`
	Chains      []string  `json:"chains"`
	Rule        string    `json:"rule"`
	RulePayload string    `json:"rulePayload"`
}

// Connections lists open connections.
func (c *Client) Connections(ctx context.Context) (inspect.Connections, error) {
	var payload struct {
		DownloadTotal int64           `json:"downloadTotal"`
		UploadTotal   int64           `json:"uploadTotal"`
		Connections   []rawConnection `json:"connections"`
	}
	if err := c.do(ctx, http.MethodGet, "/connections", nil, nil, &payload); err != nil {
		return inspect.Connections{}, err
	}
	out := inspect.Connections{UploadTotal: payload.UploadTotal, DownloadTotal: payload.DownloadTotal, Connections: []inspect.Connection{}}
	for _, raw := range payload.Connections {
		host := raw.Metadata.Host
		if host == "" {
			host = raw.Metadata.SniffHost
		}
		if host == "" {
			host = raw.Metadata.DestinationIP
		}
		chain := raw.Chains
		if chain == nil {
			chain = []string{}
		}
		out.Connections = append(out.Connections, inspect.Connection{
			ID: raw.ID, Host: host, Network: raw.Metadata.Network,
			Destination: raw.Metadata.DestinationIP + ":" + raw.Metadata.DestinationPort,
			Process:     raw.Metadata.Process, ProcessPath: raw.Metadata.ProcessPath,
			Rule: raw.Rule, RulePayload: raw.RulePayload, Chain: chain,
			Upload: raw.Upload, Download: raw.Download, Start: raw.Start,
		})
	}
	return out, nil
}

// CloseConnection closes one connection.
func (c *Client) CloseConnection(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/connections/"+escape(id), nil, nil, nil)
}

// CloseConnections closes every connection.
func (c *Client) CloseConnections(ctx context.Context) error {
	return c.do(ctx, http.MethodDelete, "/connections", nil, nil, nil)
}

type rawProvider struct {
	Name             string    `json:"name"`
	VehicleType      string    `json:"vehicleType"`
	Behavior         string    `json:"behavior"`
	Format           string    `json:"format"`
	UpdatedAt        time.Time `json:"updatedAt"`
	RuleCount        int       `json:"ruleCount"`
	Proxies          []any     `json:"proxies"`
	SubscriptionInfo *struct {
		Upload   int64 `json:"Upload"`
		Download int64 `json:"Download"`
		Total    int64 `json:"Total"`
		Expire   int64 `json:"Expire"`
	} `json:"subscriptionInfo"`
}

// ProxyProviders lists proxy providers, without the built-in "default".
func (c *Client) ProxyProviders(ctx context.Context) ([]inspect.ProxyProvider, error) {
	var payload struct {
		Providers map[string]rawProvider `json:"providers"`
	}
	if err := c.do(ctx, http.MethodGet, "/providers/proxies", nil, nil, &payload); err != nil {
		return nil, err
	}
	out := []inspect.ProxyProvider{}
	for name, p := range payload.Providers {
		if p.VehicleType == "Compatible" || name == "default" {
			continue
		}
		provider := inspect.ProxyProvider{Name: name, VehicleType: p.VehicleType, UpdatedAt: p.UpdatedAt, Count: len(p.Proxies)}
		if info := p.SubscriptionInfo; info != nil {
			provider.Usage = &inspect.Usage{Upload: info.Upload, Download: info.Download, Total: info.Total, Expire: info.Expire}
		}
		out = append(out, provider)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// UpdateProxyProvider refreshes a proxy provider.
func (c *Client) UpdateProxyProvider(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodPut, "/providers/proxies/"+escape(name), nil, nil, nil)
}

// RuleProviders lists rule providers.
func (c *Client) RuleProviders(ctx context.Context) ([]inspect.RuleProvider, error) {
	var payload struct {
		Providers map[string]rawProvider `json:"providers"`
	}
	if err := c.do(ctx, http.MethodGet, "/providers/rules", nil, nil, &payload); err != nil {
		return nil, err
	}
	out := []inspect.RuleProvider{}
	for name, p := range payload.Providers {
		out = append(out, inspect.RuleProvider{Name: name, VehicleType: p.VehicleType, Behavior: p.Behavior, Format: p.Format, UpdatedAt: p.UpdatedAt, Count: p.RuleCount})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// UpdateRuleProvider refreshes a rule provider.
func (c *Client) UpdateRuleProvider(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodPut, "/providers/rules/"+escape(name), nil, nil, nil)
}

// UpgradeGeoData downloads fresh GeoIP and GeoSite data.
func (c *Client) UpgradeGeoData(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/configs/geo", nil, map[string]string{}, nil)
}

// streamLines reads the newline-delimited JSON the controller streams
// for /traffic, /memory and /logs, until ctx ends or the core goes away.
func (c *Client) streamLines(ctx context.Context, path string, query url.Values, line func([]byte) error) error {
	req, err := c.request(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return err
	}
	resp, err := c.stream.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &Error{Status: resp.StatusCode}
	}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	for scanner.Scan() {
		data := bytes.TrimSpace(scanner.Bytes())
		if len(data) == 0 {
			continue
		}
		if err := line(data); err != nil {
			return err
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return scanner.Err()
}

// Traffic streams throughput once a second.
func (c *Client) Traffic(ctx context.Context, fn func(inspect.Traffic)) error {
	return c.streamLines(ctx, "/traffic", nil, func(data []byte) error {
		var t inspect.Traffic
		if err := json.Unmarshal(data, &t); err != nil {
			return err
		}
		fn(t)
		return nil
	})
}

// Memory streams memory use.
func (c *Client) Memory(ctx context.Context, fn func(inspect.Memory)) error {
	return c.streamLines(ctx, "/memory", nil, func(data []byte) error {
		var m inspect.Memory
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		fn(m)
		return nil
	})
}

// Logs streams log lines at level and above.
func (c *Client) Logs(ctx context.Context, level string, fn func(inspect.LogEntry)) error {
	return c.streamLines(ctx, "/logs", url.Values{"level": {level}}, func(data []byte) error {
		var entry struct {
			Type    string `json:"type"`
			Payload string `json:"payload"`
		}
		if err := json.Unmarshal(data, &entry); err != nil {
			return err
		}
		fn(inspect.LogEntry{Level: strings.ToLower(entry.Type), Message: entry.Payload, Time: time.Now()})
		return nil
	})
}
