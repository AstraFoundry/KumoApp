package application

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/override"
	"github.com/AstraFoundry/KumoApp/internal/domain/proxy"
	"github.com/AstraFoundry/KumoApp/internal/domain/runtimeconfig"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/mihomo"
)

// ProxyService lists proxy groups, selects nodes and measures delays.
type ProxyService struct {
	deps      Dependencies
	lifecycle *Lifecycle
}

func (s *ProxyService) client() (*mihomo.Client, error) {
	rt, err := s.deps.Runtime.Load()
	if err != nil {
		return nil, err
	}
	if rt.State != core.StateRunning {
		return nil, core.ErrNotRunning
	}
	return mihomo.NewClient(rt.Endpoint), nil
}

// Groups returns the live groups of a running core. When the core is off
// it reads them from the configuration instead, marked not live: Mihomo
// decides the actual selection when it starts.
func (s *ProxyService) Groups(ctx context.Context) ([]proxy.Group, error) {
	if client, err := s.client(); err == nil {
		return client.Groups(ctx)
	}
	rt, err := s.deps.Runtime.Load()
	if err != nil {
		return nil, err
	}
	config, err := s.lifecycle.RuntimeConfig(rt)
	if err != nil {
		return nil, err
	}
	return GroupsFromConfig(config)
}

// GroupsFromConfig reads proxy groups out of a configuration.
func GroupsFromConfig(config string) ([]proxy.Group, error) {
	root, err := runtimeconfig.Parse(config)
	if err != nil {
		return nil, err
	}
	groups := []proxy.Group{}
	types := map[string]string{}
	if list := override.Lookup(root, "proxies"); list != nil {
		for _, item := range list.Content {
			if name := override.Lookup(item, "name"); name != nil {
				if t := override.Lookup(item, "type"); t != nil {
					types[name.Value] = t.Value
				}
			}
		}
	}
	list := override.Lookup(root, "proxy-groups")
	if list == nil {
		return groups, nil
	}
	isGroup := map[string]bool{}
	for _, item := range list.Content {
		if name := override.Lookup(item, "name"); name != nil {
			isGroup[name.Value] = true
		}
	}
	for _, item := range list.Content {
		name := override.Lookup(item, "name")
		if name == nil {
			continue
		}
		g := proxy.Group{Name: name.Value, Nodes: []proxy.Node{}}
		if t := override.Lookup(item, "type"); t != nil {
			g.Type = t.Value
		}
		if u := override.Lookup(item, "url"); u != nil {
			g.TestURL = u.Value
		}
		if members := override.Lookup(item, "proxies"); members != nil {
			for _, m := range members.Content {
				g.Nodes = append(g.Nodes, proxy.Node{Name: m.Value, Type: types[m.Value], Country: proxy.Country(m.Value), Group: isGroup[m.Value]})
			}
		}
		groups = append(groups, g)
	}
	return groups, nil
}

// Select picks a group's node and closes connections that used the old
// one, so traffic moves at once.
func (s *ProxyService) Select(ctx context.Context, group, name string) error {
	client, err := s.client()
	if err != nil {
		return err
	}
	groups, err := client.Groups(ctx)
	if err != nil {
		return err
	}
	previous := ""
	for _, g := range groups {
		if g.Name == group {
			previous = g.Now
		}
	}
	if err := client.Select(ctx, group, name); err != nil {
		return err
	}
	if previous == "" || previous == name {
		return nil
	}
	conns, err := client.Connections(ctx)
	if err != nil {
		return nil
	}
	for _, c := range conns.Connections {
		for _, hop := range c.Chain {
			if hop == previous {
				_ = client.CloseConnection(ctx, c.ID)
				break
			}
		}
	}
	return nil
}

// DelayTimeout bounds one delay test.
const DelayTimeout = 5 * time.Second

// TestNode measures one node's delay.
func (s *ProxyService) TestNode(ctx context.Context, name, testURL string) (int, error) {
	client, err := s.client()
	if err != nil {
		return 0, err
	}
	return client.Delay(ctx, name, testURL, DelayTimeout)
}

// TestGroup measures every node of a group; nodes that time out read -1.
func (s *ProxyService) TestGroup(ctx context.Context, group string) (map[string]int, error) {
	client, err := s.client()
	if err != nil {
		return nil, err
	}
	groups, err := client.Groups(ctx)
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		if g.Name != group {
			continue
		}
		results, err := client.GroupDelay(ctx, group, g.TestURL, DelayTimeout)
		if err == nil {
			for _, n := range g.Nodes {
				if _, ok := results[n.Name]; !ok {
					results[n.Name] = -1
				}
			}
			return results, nil
		}
		// Older cores lack the group endpoint: test nodes one by one.
		results = map[string]int{}
		var mu sync.Mutex
		var wg sync.WaitGroup
		limit := make(chan struct{}, 8)
		for _, n := range g.Nodes {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				limit <- struct{}{}
				defer func() { <-limit }()
				delay, err := client.Delay(ctx, name, g.TestURL, DelayTimeout)
				if err != nil {
					delay = -1
				}
				mu.Lock()
				results[name] = delay
				mu.Unlock()
			}(n.Name)
		}
		wg.Wait()
		return results, nil
	}
	return nil, errors.New("no group named " + group)
}
