package core

import "gopkg.in/yaml.v3"

// EnsurePullRequestTrigger guarantees a workflow triggers on `pull_request`, so opening a shadow PR
// actually runs the mirrored consumer CI. Handles `on:` as a mapping, sequence, scalar, or absent.
// Comment-preserving and idempotent.
func EnsurePullRequestTrigger(yamlText string) (string, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(yamlText), &doc); err != nil {
		return "", err
	}
	root := docRoot(&doc)
	on := mapGet(root, "on")

	switch {
	case on == nil:
		mapSetNode(root, "on", &yaml.Node{
			Kind:    yaml.MappingNode,
			Content: []*yaml.Node{scalarNode("pull_request"), emptyMapNode()},
		})
	case on.Kind == yaml.MappingNode:
		if mapGet(on, "pull_request") == nil {
			on.Content = append(on.Content, scalarNode("pull_request"), emptyMapNode())
		}
	case on.Kind == yaml.SequenceNode:
		present := false
		for _, item := range on.Content {
			if item.Kind == yaml.ScalarNode && item.Value == "pull_request" {
				present = true
			}
		}
		if !present {
			on.Content = append(on.Content, scalarNode("pull_request"))
		}
	case on.Kind == yaml.ScalarNode:
		mapSetNode(root, "on", &yaml.Node{
			Kind:    yaml.SequenceNode,
			Content: []*yaml.Node{scalarNode(on.Value), scalarNode("pull_request")},
		})
	}
	return marshalNode(&doc)
}
