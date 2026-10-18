package application

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/inspect"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/mihomo"
)

// InspectService reads what a running core reports.
type InspectService struct {
	deps   Dependencies
	events *Bus

	mu   sync.Mutex
	prev map[string][2]int64
}

func (s *InspectService) client() (*mihomo.Client, error) {
	rt, err := s.deps.Runtime.Load()
	if err != nil {
		return nil, err
	}
	if rt.State != core.StateRunning {
		return nil, core.ErrNotRunning
	}
	return mihomo.NewClient(rt.Endpoint), nil
}

// Connections lists open connections with per-connection speeds, computed
// from the change since the previous call.
func (s *InspectService) Connections(ctx context.Context) (inspect.Connections, error) {
	client, err := s.client()
	if err != nil {
		return inspect.Connections{Connections: []inspect.Connection{}}, err
	}
	conns, err := client.Connections(ctx)
	if err != nil {
		return conns, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	next := map[string][2]int64{}
	for i := range conns.Connections {
		c := &conns.Connections[i]
		if prev, ok := s.prev[c.ID]; ok {
			c.UpSpeed = max(0, c.Upload-prev[0])
			c.DownSpeed = max(0, c.Download-prev[1])
		}
		next[c.ID] = [2]int64{c.Upload, c.Download}
	}
	s.prev = next
	return conns, nil
}

// CloseConnection closes one connection, or every one when id is "".
func (s *InspectService) CloseConnection(ctx context.Context, id string) error {
	client, err := s.client()
	if err != nil {
		return err
	}
	if id == "" {
		return client.CloseConnections(ctx)
	}
	return client.CloseConnection(ctx, id)
}

// Rules lists the running rules.
func (s *InspectService) Rules(ctx context.Context) ([]inspect.Rule, error) {
	client, err := s.client()
	if err != nil {
		return nil, err
	}
	return client.Rules(ctx)
}

// SetRuleEnabled turns a rule off or on until the core restarts.
func (s *InspectService) SetRuleEnabled(ctx context.Context, index int, enabled bool) error {
	client, err := s.client()
	if err != nil {
		return err
	}
	return client.SetRuleEnabled(ctx, index, enabled)
}

// Providers lists proxy and rule providers.
func (s *InspectService) Providers(ctx context.Context) ([]inspect.ProxyProvider, []inspect.RuleProvider, error) {
	client, err := s.client()
	if err != nil {
		return nil, nil, err
	}
	proxies, err := client.ProxyProviders(ctx)
	if err != nil {
		return nil, nil, err
	}
	rules, err := client.RuleProviders(ctx)
	return proxies, rules, err
}

// UpdateProvider refreshes a proxy ("proxy") or rule ("rule") provider.
func (s *InspectService) UpdateProvider(ctx context.Context, kind, name string) error {
	client, err := s.client()
	if err != nil {
		return err
	}
	if kind == "rule" {
		return client.UpdateRuleProvider(ctx, name)
	}
	return client.UpdateProxyProvider(ctx, name)
}

// UpgradeGeoData downloads fresh geo databases.
func (s *InspectService) UpgradeGeoData(ctx context.Context) error {
	client, err := s.client()
	if err != nil {
		return err
	}
	return client.UpgradeGeoData(ctx)
}

// Version is the running core's version.
func (s *InspectService) Version(ctx context.Context) (inspect.Version, error) {
	client, err := s.client()
	if err != nil {
		return inspect.Version{}, err
	}
	return client.Version(ctx)
}

var logLine = regexp.MustCompile(`time="([^"]*)" level=(\w+) msg="((?:[^"\\]|\\.)*)"`)

// RecentLogs reads the last lines of the core's log file.
func (s *InspectService) RecentLogs(limit int) ([]inspect.LogEntry, error) {
	f, err := os.Open(s.deps.Paths.CoreLog())
	if os.IsNotExist(err) {
		return []inspect.LogEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64<<10), 1<<20)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > limit*2 {
			lines = lines[len(lines)-limit:]
		}
	}
	if len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	entries := make([]inspect.LogEntry, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		entry := inspect.LogEntry{Level: "info", Message: line}
		if m := logLine.FindStringSubmatch(line); m != nil {
			entry.Level, entry.Message = normalizeLevel(m[2]), strings.ReplaceAll(m[3], `\"`, `"`)
			entry.Time, _ = time.Parse(time.RFC3339Nano, m[1])
		} else if strings.HasPrefix(line, "---") {
			entry.Level = "kumo"
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func normalizeLevel(level string) string {
	switch strings.ToLower(level) {
	case "warn", "warning":
		return "warning"
	case "fatal", "panic":
		return "error"
	}
	return strings.ToLower(level)
}

// StreamTraffic publishes throughput every second while the core runs and
// ctx lives, reconnecting when the core restarts.
func (s *InspectService) StreamTraffic(ctx context.Context) {
	s.stream(ctx, func(client *mihomo.Client) error {
		return client.Traffic(ctx, func(t inspect.Traffic) { s.events.Publish(TrafficSampled{Traffic: t}) })
	}, func() { s.events.Publish(TrafficSampled{}) })
}

// StreamLogs publishes log lines at level and above while ctx lives.
func (s *InspectService) StreamLogs(ctx context.Context, level string) {
	s.stream(ctx, func(client *mihomo.Client) error {
		return client.Logs(ctx, level, func(e inspect.LogEntry) { s.events.Publish(LogReceived{Entry: e}) })
	}, nil)
}

func (s *InspectService) stream(ctx context.Context, run func(*mihomo.Client) error, idle func()) {
	for ctx.Err() == nil {
		client, err := s.client()
		if err == nil {
			_ = run(client)
		}
		if idle != nil {
			idle()
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
