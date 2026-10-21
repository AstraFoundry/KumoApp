package application

import (
	"context"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/runtimeconfig"
	"github.com/AstraFoundry/KumoApp/internal/domain/topology"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/mihomo"
)

// TopologyService draws the configuration as a graph.
type TopologyService struct {
	deps      Dependencies
	lifecycle *Lifecycle
}

// Graph builds the topology of the configuration the core runs (or would
// run), with the live selections when it runs.
func (s *TopologyService) Graph(ctx context.Context) (topology.Graph, error) {
	rt, err := s.deps.Runtime.Load()
	if err != nil {
		return topology.Graph{}, err
	}
	config, err := s.lifecycle.RuntimeConfig(rt)
	if err != nil {
		return topology.Graph{}, err
	}
	root, err := runtimeconfig.Parse(config)
	if err != nil {
		return topology.Graph{}, err
	}
	selections := topology.Selections{}
	if rt.State == core.StateRunning {
		if groups, err := mihomo.NewClient(rt.Endpoint).Groups(ctx); err == nil {
			for _, g := range groups {
				if g.Now != "" {
					selections[g.Name] = g.Now
				}
			}
		}
	}
	return topology.Build(root, string(rt.Mode), selections), nil
}
