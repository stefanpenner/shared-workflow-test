package core

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// Helpers for comment-preserving edits over a parsed yaml.Node tree. yaml.v3 keeps comments and
// (with a 2-space encoder) formatting through an unmarshal → mutate → marshal round-trip.

// docRoot returns the top mapping node of a parsed document.
func docRoot(doc *yaml.Node) *yaml.Node {
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0]
	}
	return doc
}

// mapGet returns the value node for key in a mapping node, or nil.
func mapGet(m *yaml.Node, key string) *yaml.Node {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// mapSetScalar sets key=value (scalar) in a mapping, appending the pair if absent.
func mapSetScalar(m *yaml.Node, key, value string) {
	if v := mapGet(m, key); v != nil {
		*v = yaml.Node{Kind: yaml.ScalarNode, Value: value}
		return
	}
	m.Content = append(m.Content, scalarNode(key), scalarNode(value))
}

// mapSetNode sets key to valNode in a mapping, appending the pair if absent.
func mapSetNode(m *yaml.Node, key string, valNode *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = valNode
			return
		}
	}
	m.Content = append(m.Content, scalarNode(key), valNode)
}

func scalarNode(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: value}
}

func emptyMapNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode}
}

// marshalNode serializes a node tree with a 2-space indent (matching the source style).
func marshalNode(doc *yaml.Node) (string, error) {
	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	return b.String(), nil
}
