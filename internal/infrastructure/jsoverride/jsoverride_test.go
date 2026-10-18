package jsoverride

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTransform(t *testing.T) {
	var doc yaml.Node
	_ = yaml.Unmarshal([]byte("rules:\n  - MATCH,DIRECT\nmode: rule\n"), &doc)
	out, err := Runner{}.Transform("t", `function main(c){ c.rules.unshift("DOMAIN,a.com,DIRECT"); c.extra = {n: 1}; return c }`, doc.Content[0])
	if err != nil {
		t.Fatal(err)
	}
	data, _ := yaml.Marshal(out)
	if !strings.Contains(string(data), "DOMAIN,a.com,DIRECT") || !strings.Contains(string(data), `"n": 1`) {
		t.Fatalf("got %s", data)
	}
	Timeout = 200_000_000
	if _, err := (Runner{}).Transform("t", `function main(c){ while(true){} }`, doc.Content[0]); err == nil {
		t.Fatal("infinite loop not stopped")
	}
	if _, err := (Runner{}).Transform("t", `var x = 1`, doc.Content[0]); err == nil {
		t.Fatal("missing main accepted")
	}
}
