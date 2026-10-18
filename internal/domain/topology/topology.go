// Package topology turns a configuration into the graph of how traffic
// flows through it: the outbound mode, the rules that pick a policy, the
// proxy groups that pick a node, and the proxies and providers at the end.
// The canvas draws this graph; the graph itself knows nothing of layout
// beyond which column each node belongs to.
package topology

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/AstraFoundry/KumoApp/internal/domain/override"
	"github.com/AstraFoundry/KumoApp/internal/domain/proxy"
)

// Kind is what a node stands for.
type Kind string

const (
	KindEntry         Kind = "entry"
	KindRuleSet       Kind = "ruleSet"
	KindRules         Kind = "rules"
	KindFinal         Kind = "final"
	KindGroup         Kind = "group"
	KindProxy         Kind = "proxy"
	KindProvider      Kind = "provider"
	KindBuiltin       Kind = "builtin"
	KindMissing       Kind = "missing"
	KindRuleProvider  Kind = "ruleProvider"
	KindGlobalChooser Kind = "global"
)

// Column is where a node sits, left to right in the order traffic flows.
type Column int

const (
	ColumnEntry Column = iota
	ColumnRules
	ColumnGroups
	ColumnNodes
)

// Node is one vertex of the graph.
type Node struct {
	ID     string `json:"id"`
	Kind   Kind   `json:"kind"`
	Column Column `json:"column"`
	Label  string `json:"label"`
	// Type is the group or proxy type, or the rule types of a cluster.
	Type string `json:"type,omitempty"`
	// Count is how many members, rules or proxies the node stands for.
	Count   int    `json:"count,omitempty"`
	Country string `json:"country,omitempty"`
	// Order keeps the configuration's order within a column.
	Order int `json:"order"`
	// Selected is the member a group currently uses, when known.
	Selected string `json:"selected,omitempty"`
	// Payloads are a few examples of what a rule node matches.
	Payloads []string `json:"payloads,omitempty"`
}

// EdgeKind is the relation an edge stands for.
type EdgeKind string

const (
	// EdgeRoute: the mode or a rule sends traffic to a policy.
	EdgeRoute EdgeKind = "route"
	// EdgeMember: a group lists the target as a member.
	EdgeMember EdgeKind = "member"
	// EdgeProvider: a group takes members from a provider.
	EdgeProvider EdgeKind = "provider"
	// EdgeDynamic: a group's filter or include-all brings the target in;
	// the core decides the actual members.
	EdgeDynamic EdgeKind = "dynamic"
)

// Edge is one relation of the graph.
type Edge struct {
	ID     string   `json:"id"`
	Source string   `json:"source"`
	Target string   `json:"target"`
	Kind   EdgeKind `json:"kind"`
	// Active marks the path traffic takes now: the mode's route and each
	// group's selected member.
	Active bool `json:"active,omitempty"`
	// Broken marks a reference to a name the configuration does not define.
	Broken bool `json:"broken,omitempty"`
}

// Issue is a reference the configuration cannot resolve.
type Issue struct {
	From    string `json:"from"`
	Missing string `json:"missing"`
	Message string `json:"message"`
}

// Graph is the whole topology.
type Graph struct {
	Mode   string  `json:"mode"`
	Nodes  []Node  `json:"nodes"`
	Edges  []Edge  `json:"edges"`
	Issues []Issue `json:"issues"`
}

// Selections are the members groups use now, from a running core.
type Selections map[string]string

var builtins = []string{"DIRECT", "REJECT", "REJECT-DROP", "PASS", "COMPATIBLE"}

// ID helpers keep node identities stable across rebuilds, so the canvas
// keeps positions and focus.
func groupID(name string) string    { return "group:" + name }
func proxyID(name string) string    { return "proxy:" + name }
func providerID(name string) string { return "provider:" + name }
func builtinID(name string) string  { return "builtin:" + name }

// Build reads a configuration (a profile with overrides applied) into its
// graph. mode is the outbound mode; selections, when the core runs, mark
// the active path.
func Build(config *yaml.Node, mode string, selections Selections) Graph {
	b := &builder{graph: Graph{Mode: mode, Nodes: []Node{}, Edges: []Edge{}, Issues: []Issue{}}, ids: map[string]bool{}, selections: selections}
	b.collect(config)
	b.entry(mode)
	b.groups()
	b.rules(config, mode)
	b.finish()
	return b.graph
}

type builder struct {
	graph      Graph
	ids        map[string]bool
	edgeIDs    map[string]bool
	selections Selections

	proxyNames    []string
	groupNames    []string
	groupNodes    map[string]*yaml.Node
	providerNames []string
	ruleProviders map[string]string
}

func (b *builder) addNode(n Node) {
	if b.ids[n.ID] {
		return
	}
	b.ids[n.ID] = true
	n.Order = len(b.graph.Nodes)
	b.graph.Nodes = append(b.graph.Nodes, n)
}

func (b *builder) addEdge(e Edge) {
	if b.edgeIDs == nil {
		b.edgeIDs = map[string]bool{}
	}
	e.ID = string(e.Kind) + ":" + e.Source + "->" + e.Target
	if b.edgeIDs[e.ID] {
		return
	}
	b.edgeIDs[e.ID] = true
	b.graph.Edges = append(b.graph.Edges, e)
}

func (b *builder) collect(config *yaml.Node) {
	b.groupNodes = map[string]*yaml.Node{}
	b.ruleProviders = map[string]string{}
	if proxies := override.Lookup(config, "proxies"); proxies != nil {
		for _, item := range proxies.Content {
			name := scalar(override.Lookup(item, "name"))
			if name == "" {
				continue
			}
			b.proxyNames = append(b.proxyNames, name)
			b.addNode(Node{ID: proxyID(name), Kind: KindProxy, Column: ColumnNodes, Label: name,
				Type: scalar(override.Lookup(item, "type")), Country: proxy.Country(name)})
		}
	}
	if providers := override.Lookup(config, "proxy-providers"); providers != nil && providers.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(providers.Content); i += 2 {
			name := providers.Content[i].Value
			b.providerNames = append(b.providerNames, name)
			b.addNode(Node{ID: providerID(name), Kind: KindProvider, Column: ColumnNodes, Label: name,
				Type: scalar(override.Lookup(providers.Content[i+1], "type"))})
		}
	}
	if providers := override.Lookup(config, "rule-providers"); providers != nil && providers.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(providers.Content); i += 2 {
			b.ruleProviders[providers.Content[i].Value] = scalar(override.Lookup(providers.Content[i+1], "behavior"))
		}
	}
	if groups := override.Lookup(config, "proxy-groups"); groups != nil {
		for _, item := range groups.Content {
			name := scalar(override.Lookup(item, "name"))
			if name == "" {
				continue
			}
			b.groupNames = append(b.groupNames, name)
			b.groupNodes[name] = item
		}
	}
}

func (b *builder) entry(mode string) {
	b.addNode(Node{ID: "entry", Kind: KindEntry, Column: ColumnEntry, Label: mode})
}

// policy returns the node ID a rule or group refers to by name, adding a
// built-in or a missing node as needed.
func (b *builder) policy(name, from string) (string, bool) {
	switch {
	case b.groupNodes[name] != nil:
		return groupID(name), true
	case slices.Contains(b.proxyNames, name):
		return proxyID(name), true
	case slices.Contains(builtins, strings.ToUpper(name)):
		id := builtinID(strings.ToUpper(name))
		b.addNode(Node{ID: id, Kind: KindBuiltin, Column: ColumnNodes, Label: strings.ToUpper(name)})
		return id, true
	}
	id := "missing:" + name
	b.addNode(Node{ID: id, Kind: KindMissing, Column: ColumnNodes, Label: name})
	b.graph.Issues = append(b.graph.Issues, Issue{From: from, Missing: name,
		Message: fmt.Sprintf("%s refers to %q, which the configuration does not define", from, name)})
	return id, false
}

func (b *builder) groups() {
	for _, name := range b.groupNames {
		item := b.groupNodes[name]
		members := strings_(override.Lookup(item, "proxies"))
		uses := strings_(override.Lookup(item, "use"))
		count := len(members)
		node := Node{ID: groupID(name), Kind: KindGroup, Column: ColumnGroups, Label: name,
			Type: scalar(override.Lookup(item, "type")), Selected: b.selections[name]}
		b.addNode(node)
		for _, member := range members {
			target, ok := b.policy(member, "group "+strconv.Quote(name))
			b.addEdge(Edge{Source: node.ID, Target: target, Kind: EdgeMember, Broken: !ok, Active: b.selections[name] == member})
		}
		for _, use := range uses {
			if !slices.Contains(b.providerNames, use) {
				id := "missing:" + use
				b.addNode(Node{ID: id, Kind: KindMissing, Column: ColumnNodes, Label: use})
				b.graph.Issues = append(b.graph.Issues, Issue{From: "group " + strconv.Quote(name), Missing: use,
					Message: fmt.Sprintf("group %q uses provider %q, which the configuration does not define", name, use)})
				b.addEdge(Edge{Source: node.ID, Target: id, Kind: EdgeProvider, Broken: true})
				continue
			}
			b.addEdge(Edge{Source: node.ID, Target: providerID(use), Kind: EdgeProvider})
		}
		// include-all brings in every proxy (and provider); a filter narrows
		// them. The core resolves these at run time, so they are dynamic.
		includeAll := truthy(override.Lookup(item, "include-all"))
		includeProxies := includeAll || truthy(override.Lookup(item, "include-all-proxies"))
		includeProviders := includeAll || truthy(override.Lookup(item, "include-all-providers"))
		filter := compileFilter(scalar(override.Lookup(item, "filter")))
		exclude := compileFilter(scalar(override.Lookup(item, "exclude-filter")))
		if includeProxies {
			for _, p := range b.proxyNames {
				if slices.Contains(members, p) || !matches(filter, exclude, p) {
					continue
				}
				count++
				b.addEdge(Edge{Source: node.ID, Target: proxyID(p), Kind: EdgeDynamic, Active: b.selections[name] == p})
			}
		}
		if includeProviders {
			for _, p := range b.providerNames {
				if !slices.Contains(uses, p) {
					b.addEdge(Edge{Source: node.ID, Target: providerID(p), Kind: EdgeDynamic})
				}
			}
		}
		b.setCount(node.ID, count+len(uses))
	}
}

func (b *builder) setCount(id string, count int) {
	for i := range b.graph.Nodes {
		if b.graph.Nodes[i].ID == id {
			b.graph.Nodes[i].Count = count
			return
		}
	}
}

func compileFilter(pattern string) *regexp.Regexp {
	if pattern == "" {
		return nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re
}

func matches(filter, exclude *regexp.Regexp, name string) bool {
	if filter != nil && !filter.MatchString(name) {
		return false
	}
	if exclude != nil && exclude.MatchString(name) {
		return false
	}
	return true
}

// rules clusters the rule list: each RULE-SET becomes a node of its own,
// consecutive inline rules sending traffic to the same policy become one,
// and MATCH ends the chain. In global mode traffic skips the rules and
// takes the GLOBAL group; in direct mode it goes straight out.
func (b *builder) rules(config *yaml.Node, mode string) {
	switch mode {
	case "global":
		target := builtinID("DIRECT")
		if b.groupNodes["GLOBAL"] != nil {
			target = groupID("GLOBAL")
		} else {
			b.addNode(Node{ID: "global", Kind: KindGlobalChooser, Column: ColumnGroups, Label: "GLOBAL", Selected: b.selections["GLOBAL"]})
			target = "global"
			for _, name := range append(slices.Clone(b.groupNames), b.proxyNames...) {
				id, _ := b.policy(name, "GLOBAL")
				b.addEdge(Edge{Source: "global", Target: id, Kind: EdgeMember, Active: b.selections["GLOBAL"] == name})
			}
		}
		b.addEdge(Edge{Source: "entry", Target: target, Kind: EdgeRoute, Active: true})
	case "direct":
		id, _ := b.policy("DIRECT", "mode")
		b.addEdge(Edge{Source: "entry", Target: id, Kind: EdgeRoute, Active: true})
	}

	list := override.Lookup(config, "rules")
	if list == nil {
		return
	}
	type cluster struct {
		id       string
		target   string
		types    []string
		payloads []string
		count    int
	}
	var current *cluster
	flush := func() {
		if current == nil {
			return
		}
		b.addNode(Node{ID: current.id, Kind: KindRules, Column: ColumnRules, Label: current.target,
			Type: strings.Join(current.types, ", "), Count: current.count, Payloads: current.payloads})
		b.linkRule(current.id, current.target, mode)
		current = nil
	}
	for index, item := range list.Content {
		parts := splitRule(item.Value)
		if len(parts) < 2 {
			continue
		}
		kind := strings.ToUpper(parts[0])
		switch kind {
		case "MATCH", "FINAL":
			flush()
			id := "final"
			b.addNode(Node{ID: id, Kind: KindFinal, Column: ColumnRules, Label: parts[1], Type: "MATCH"})
			b.linkRule(id, parts[1], mode)
			continue
		case "RULE-SET":
			flush()
			if len(parts) < 3 {
				continue
			}
			id := "ruleset:" + parts[1] + "@" + strconv.Itoa(index)
			behavior := b.ruleProviders[parts[1]]
			b.addNode(Node{ID: id, Kind: KindRuleSet, Column: ColumnRules, Label: parts[1], Type: behavior, Payloads: []string{parts[2]}})
			if _, ok := b.ruleProviders[parts[1]]; !ok {
				b.graph.Issues = append(b.graph.Issues, Issue{From: "rules", Missing: parts[1],
					Message: fmt.Sprintf("RULE-SET %q is not defined in rule-providers", parts[1])})
			}
			b.linkRule(id, parts[2], mode)
			continue
		}
		if len(parts) < 3 {
			continue
		}
		target := parts[2]
		if current == nil || current.target != target {
			flush()
			current = &cluster{id: "rules:" + strconv.Itoa(index), target: target}
		}
		current.count++
		if !slices.Contains(current.types, kind) {
			current.types = append(current.types, kind)
		}
		if len(current.payloads) < 3 {
			current.payloads = append(current.payloads, parts[1])
		}
	}
	flush()
}

func (b *builder) linkRule(id, target, mode string) {
	policy, ok := b.policy(target, "rule "+strconv.Quote(target))
	b.addEdge(Edge{Source: "entry", Target: id, Kind: EdgeRoute})
	b.addEdge(Edge{Source: id, Target: policy, Kind: EdgeRoute, Broken: !ok})
}

// finish keeps a group's selection active only when traffic can reach the
// group: through the rules in rule mode, or the mode's route otherwise.
func (b *builder) finish() {
	reached := map[string]bool{"entry": true}
	for changed := true; changed; {
		changed = false
		for i := range b.graph.Edges {
			e := &b.graph.Edges[i]
			follows := e.Active || (e.Kind == EdgeRoute && b.graph.Mode == "rule")
			if follows && reached[e.Source] && !reached[e.Target] {
				reached[e.Target] = true
				changed = true
			}
		}
	}
	active := reached
	for i := range b.graph.Edges {
		e := &b.graph.Edges[i]
		if e.Kind != EdgeRoute && !active[e.Source] {
			e.Active = false
		}
	}
	sort.SliceStable(b.graph.Nodes, func(i, j int) bool {
		if b.graph.Nodes[i].Column != b.graph.Nodes[j].Column {
			return b.graph.Nodes[i].Column < b.graph.Nodes[j].Column
		}
		return b.graph.Nodes[i].Order < b.graph.Nodes[j].Order
	})
	for i := range b.graph.Nodes {
		b.graph.Nodes[i].Order = i
	}
}

// splitRule splits "DOMAIN-SUFFIX,example.com,Proxy,no-resolve" on commas
// outside parentheses, so logical rules such as AND,((A,B),(C,D)),Proxy
// keep their payload whole.
func splitRule(rule string) []string {
	var parts []string
	depth, start := 0, 0
	for i, r := range rule {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(rule[start:i]))
				start = i + 1
			}
		}
	}
	return append(parts, strings.TrimSpace(rule[start:]))
}

func scalar(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}
	return node.Value
}

func truthy(node *yaml.Node) bool { return scalar(node) == "true" }

func strings_(node *yaml.Node) []string {
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}
	var values []string
	for _, item := range node.Content {
		if item.Kind == yaml.ScalarNode {
			values = append(values, item.Value)
		}
	}
	return values
}
