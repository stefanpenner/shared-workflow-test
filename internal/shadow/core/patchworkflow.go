package core

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// reusableUses matches a job-level reusable-workflow uses: owner/repo/.github/workflows/<file>@<ref>.
var reusableUses = regexp.MustCompile(`^([^/]+/[^/]+)/(\.github/workflows/[^@]+)@.+$`)

// ReferencesWorkflowsRepo reports whether any job calls workflowsRepo as a reusable workflow — used
// to decide which of a consumer's workflows to mirror-transform.
func ReferencesWorkflowsRepo(yamlText, workflowsRepo string) (bool, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(yamlText), &doc); err != nil {
		return false, err
	}
	jobs := mapGet(docRoot(&doc), "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return false, nil
	}
	for i := 1; i < len(jobs.Content); i += 2 {
		uses := mapGet(jobs.Content[i], "uses")
		if uses == nil || uses.Kind != yaml.ScalarNode {
			continue
		}
		if m := reusableUses.FindStringSubmatch(uses.Value); m != nil && m[1] == workflowsRepo {
			return true, nil
		}
	}
	return false, nil
}

// PatchConsumerWorkflow repoints a consumer's reusable-workflow call at workflowsRef and injects
// with.ref, preserving comments/formatting. Only job-level uses targeting workflowsRepo are touched
// (step-level action uses and other repos' workflows are left alone). Idempotent.
func PatchConsumerWorkflow(yamlText, workflowsRepo, workflowsRef string) (string, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(yamlText), &doc); err != nil {
		return "", err
	}
	jobs := mapGet(docRoot(&doc), "jobs")
	if jobs == nil || jobs.Kind != yaml.MappingNode {
		return marshalNode(&doc)
	}
	for i := 1; i < len(jobs.Content); i += 2 {
		job := jobs.Content[i]
		if job.Kind != yaml.MappingNode {
			continue
		}
		uses := mapGet(job, "uses")
		if uses == nil || uses.Kind != yaml.ScalarNode {
			continue
		}
		m := reusableUses.FindStringSubmatch(uses.Value)
		if m == nil || m[1] != workflowsRepo {
			continue
		}
		path := m[2]
		if strings.HasSuffix(path, ".yml") { // fix the .yml -> .yaml typo
			path = strings.TrimSuffix(path, ".yml") + ".yaml"
		}
		*uses = yaml.Node{Kind: yaml.ScalarNode, Value: workflowsRepo + "/" + path + "@" + workflowsRef}

		if with := mapGet(job, "with"); with != nil && with.Kind == yaml.MappingNode {
			mapSetScalar(with, "ref", workflowsRef)
		} else {
			mapSetNode(job, "with", &yaml.Node{
				Kind:    yaml.MappingNode,
				Content: []*yaml.Node{scalarNode("ref"), scalarNode(workflowsRef)},
			})
		}
	}
	return marshalNode(&doc)
}
