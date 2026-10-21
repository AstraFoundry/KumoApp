// Package override models the ordered patches Kumo applies on top of the
// current profile before handing the configuration to Mihomo.
package override

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Kind is where an override's content comes from.
type Kind string

const (
	KindLocal  Kind = "local"
	KindRemote Kind = "remote"
)

// Format is the language an override is written in.
type Format string

const (
	FormatYAML       Format = "yaml"
	FormatJavaScript Format = "javascript"
)

// Override is one patch, applied in the order of the list.
type Override struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Kind        Kind      `json:"kind"`
	Format      Format    `json:"format"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Global      bool      `json:"isGlobal"`
	RemoteURL   string    `json:"remoteURL,omitempty"`
	Fingerprint string    `json:"fingerprint,omitempty"`
	// Disabled overrides stay in the list but are not applied.
	Disabled bool `json:"disabled,omitempty"`
}

// Source is an override's content ready to apply.
type Source struct {
	Name    string
	Format  Format
	Content string
}

// Validate checks that content parses in its format.
func Validate(format Format, content string) error {
	switch format {
	case FormatYAML:
		var node yaml.Node
		if err := yaml.Unmarshal([]byte(content), &node); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}
		if doc := documentRoot(&node); doc != nil && doc.Kind != yaml.MappingNode {
			return fmt.Errorf("an override must be a YAML mapping")
		}
		return nil
	case FormatJavaScript:
		if !strings.Contains(content, "main") {
			return fmt.Errorf("a JavaScript override must define function main(config)")
		}
		return nil
	}
	return fmt.Errorf("unknown override format %q", format)
}

func documentRoot(node *yaml.Node) *yaml.Node {
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return nil
		}
		return node.Content[0]
	}
	return node
}

// Repository keeps overrides in their order.
type Repository interface {
	List() ([]Override, error)
	Content(id string) (string, error)
	// Save adds or replaces an override; content nil keeps the stored one.
	Save(o Override, content *string) error
	Delete(id string) error
	Reorder(ids []string) error
}
