// Package proxy models the proxy groups and nodes of a configuration, as the
// running core reports them or as a profile declares them.
package proxy

import "slices"

// Node is one proxy, or a group nested in another group.
type Node struct {
	Name string `json:"name"`
	Type string `json:"type"`
	// Delay is the last measured latency in milliseconds; 0 is unknown and
	// a negative value a timeout.
	Delay   int    `json:"delay"`
	Alive   bool   `json:"alive"`
	UDP     bool   `json:"udp"`
	Country string `json:"country,omitempty"`
	// Group is set when the node is itself a proxy group.
	Group bool `json:"group,omitempty"`
}

// Group is a proxy group: select, url-test, fallback, load-balance or relay.
type Group struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Now     string `json:"now,omitempty"`
	TestURL string `json:"testUrl,omitempty"`
	Hidden  bool   `json:"hidden,omitempty"`
	Icon    string `json:"icon,omitempty"`
	Nodes   []Node `json:"nodes"`
	// Live is false for groups read from a profile while the core is off:
	// Mihomo picks the actual selection when it starts.
	Live bool `json:"live"`
}

// Selectable reports whether the user picks the group's node by hand.
func (g Group) Selectable() bool { return g.Type == "Selector" || g.Type == "select" }

// GroupTypes are the types Mihomo reports for groups.
var GroupTypes = []string{"Selector", "URLTest", "Fallback", "LoadBalance", "Relay", "select", "url-test", "fallback", "load-balance", "relay"}

// IsGroupType reports whether a proxy type is a group.
func IsGroupType(t string) bool { return slices.Contains(GroupTypes, t) }

// DelayGrade buckets a delay for presentation.
type DelayGrade string

const (
	DelayUnknown DelayGrade = "unknown"
	DelayFast    DelayGrade = "fast"
	DelaySlow    DelayGrade = "slow"
	DelayTimeout DelayGrade = "timeout"
)

// Grade buckets a delay: under 300 ms is fast.
func Grade(delay int) DelayGrade {
	switch {
	case delay < 0:
		return DelayTimeout
	case delay == 0:
		return DelayUnknown
	case delay < 300:
		return DelayFast
	}
	return DelaySlow
}
