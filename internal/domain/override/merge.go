package override

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// Merge applies a YAML override to a mapping in place. Its keys follow the
// conventions Clash clients share:
//
//   - "key" merges a mapping into the existing one and replaces anything else;
//   - "key!" replaces the value outright, mappings included;
//   - "+key" and "prepend-key" put a list's items before the existing ones;
//   - "key+" and "append-key" put them after;
//   - "delete-key" removes the items whose value, or "name", is listed;
//   - "<key>" quotes a key that would otherwise look like an operator.
func Merge(target, patch *yaml.Node) {
	if target.Kind != yaml.MappingNode || patch.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(patch.Content); i += 2 {
		rawKey, value := patch.Content[i].Value, patch.Content[i+1]
		if applySequence(target, rawKey, value) {
			continue
		}
		key := unwrap(strings.TrimSuffix(rawKey, "!"))
		existing := Lookup(target, key)
		switch {
		case strings.HasSuffix(rawKey, "!") || existing == nil:
			Set(target, key, value)
		case existing.Kind == yaml.MappingNode && value.Kind == yaml.MappingNode:
			Merge(existing, value)
		default:
			Set(target, key, value)
		}
	}
}

func applySequence(target *yaml.Node, rawKey string, value *yaml.Node) bool {
	if value.Kind != yaml.SequenceNode {
		return false
	}
	switch {
	case strings.HasPrefix(rawKey, "+"):
		prepend(target, unwrap(rawKey[1:]), value.Content)
	case strings.HasSuffix(rawKey, "+"):
		appendItems(target, unwrap(strings.TrimSuffix(rawKey, "+")), value.Content)
	case strings.HasPrefix(rawKey, "prepend-"):
		prepend(target, strings.TrimPrefix(rawKey, "prepend-"), value.Content)
	case strings.HasPrefix(rawKey, "append-"):
		appendItems(target, strings.TrimPrefix(rawKey, "append-"), value.Content)
	case strings.HasPrefix(rawKey, "delete-"):
		remove(target, strings.TrimPrefix(rawKey, "delete-"), value.Content)
	default:
		return false
	}
	return true
}

func unwrap(key string) string {
	if len(key) >= 2 && strings.HasPrefix(key, "<") && strings.HasSuffix(key, ">") {
		return key[1 : len(key)-1]
	}
	return key
}

func sequence(target *yaml.Node, key string) *yaml.Node {
	existing := Lookup(target, key)
	if existing == nil || existing.Kind != yaml.SequenceNode {
		existing = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		Set(target, key, existing)
	}
	return existing
}

func prepend(target *yaml.Node, key string, items []*yaml.Node) {
	seq := sequence(target, key)
	seq.Content = append(append([]*yaml.Node{}, items...), seq.Content...)
}

func appendItems(target *yaml.Node, key string, items []*yaml.Node) {
	seq := sequence(target, key)
	seq.Content = append(seq.Content, items...)
}

func remove(target *yaml.Node, key string, items []*yaml.Node) {
	names := map[string]bool{}
	for _, item := range items {
		if name := itemName(item); name != "" {
			names[name] = true
		}
	}
	existing := Lookup(target, key)
	if len(names) == 0 || existing == nil || existing.Kind != yaml.SequenceNode {
		return
	}
	kept := existing.Content[:0]
	for _, item := range existing.Content {
		if !names[itemName(item)] {
			kept = append(kept, item)
		}
	}
	existing.Content = kept
}

func itemName(node *yaml.Node) string {
	switch node.Kind {
	case yaml.ScalarNode:
		return node.Value
	case yaml.MappingNode:
		if name := Lookup(node, "name"); name != nil && name.Kind == yaml.ScalarNode {
			return name.Value
		}
	}
	return ""
}

// Lookup returns the value of key in a mapping node, or nil.
func Lookup(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// Set replaces or appends key in a mapping node.
func Set(mapping *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1] = value
			return
		}
	}
	mapping.Content = append(mapping.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
}

// Delete removes key from a mapping node.
func Delete(mapping *yaml.Node, key string) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
			return
		}
	}
}
