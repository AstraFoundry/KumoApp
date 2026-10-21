package topology

import (
	"testing"

	"gopkg.in/yaml.v3"
)

const config = `
proxies:
  - {name: "🇯🇵 Tokyo", type: vless}
  - {name: "HK 01", type: ss}
  - {name: "US 01", type: trojan}
proxy-providers:
  sub: {type: http, url: "https://example.com"}
rule-providers:
  ads: {type: http, behavior: domain}
proxy-groups:
  - {name: Proxy, type: select, proxies: [Auto, "🇯🇵 Tokyo", DIRECT]}
  - {name: Auto, type: url-test, include-all: true, filter: "HK|US"}
  - {name: Broken, type: select, proxies: [Nowhere], use: [gone]}
rules:
  - RULE-SET,ads,REJECT
  - DOMAIN-SUFFIX,a.com,Proxy
  - DOMAIN-KEYWORD,b,Proxy
  - AND,((DOMAIN,c.com),(NETWORK,UDP)),DIRECT
  - RULE-SET,missing,Proxy
  - MATCH,Proxy
`

func load(t *testing.T) *yaml.Node {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(config), &doc); err != nil {
		t.Fatal(err)
	}
	return doc.Content[0]
}

func find(g Graph, id string) *Node {
	for i := range g.Nodes {
		if g.Nodes[i].ID == id {
			return &g.Nodes[i]
		}
	}
	return nil
}

func edge(g Graph, source, target string) *Edge {
	for i := range g.Edges {
		if g.Edges[i].Source == source && g.Edges[i].Target == target {
			return &g.Edges[i]
		}
	}
	return nil
}

func TestRuleModeGraph(t *testing.T) {
	g := Build(load(t), "rule", Selections{"Proxy": "Auto", "Auto": "HK 01"})
	if n := find(g, "rules:1"); n == nil || n.Count != 2 || n.Label != "Proxy" {
		t.Fatalf("clustered rules: %+v", n)
	}
	if n := find(g, "rules:3"); n == nil || n.Label != "DIRECT" || n.Payloads[0] != "((DOMAIN,c.com),(NETWORK,UDP))" {
		t.Fatalf("logical rule: %+v", n)
	}
	if find(g, "final") == nil || edge(g, "final", "group:Proxy") == nil {
		t.Fatal("MATCH missing")
	}
	if e := edge(g, "group:Proxy", "group:Auto"); e == nil || !e.Active {
		t.Fatalf("selection not active: %+v", e)
	}
	if e := edge(g, "group:Auto", "proxy:HK 01"); e == nil || e.Kind != EdgeDynamic || !e.Active {
		t.Fatalf("dynamic member: %+v", e)
	}
	if edge(g, "group:Auto", "proxy:🇯🇵 Tokyo") != nil {
		t.Fatal("filter ignored")
	}
	if find(g, "group:Auto").Count != 2 {
		t.Fatalf("auto count %d", find(g, "group:Auto").Count)
	}
	if len(g.Issues) != 3 {
		t.Fatalf("issues: %+v", g.Issues)
	}
	if e := edge(g, "group:Broken", "missing:Nowhere"); e == nil || !e.Broken {
		t.Fatal("broken reference not marked")
	}
	if find(g, "proxy:🇯🇵 Tokyo").Country != "JP" {
		t.Fatal("country")
	}
	for i := 1; i < len(g.Nodes); i++ {
		if g.Nodes[i].Column < g.Nodes[i-1].Column {
			t.Fatal("nodes not sorted by column")
		}
	}
}

func TestGlobalAndDirectModes(t *testing.T) {
	g := Build(load(t), "global", Selections{"GLOBAL": "Proxy", "Proxy": "Auto"})
	if e := edge(g, "entry", "global"); e == nil || !e.Active {
		t.Fatal("global route")
	}
	if e := edge(g, "global", "group:Proxy"); e == nil || !e.Active {
		t.Fatal("global selection")
	}
	g = Build(load(t), "direct", nil)
	if e := edge(g, "entry", "builtin:DIRECT"); e == nil || !e.Active {
		t.Fatal("direct route")
	}
	if e := edge(g, "group:Proxy", "group:Auto"); e != nil && e.Active {
		t.Fatal("unreached selection active")
	}
}
