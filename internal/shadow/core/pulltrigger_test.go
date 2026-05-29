package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ensurePR(t *testing.T, in string) string {
	t.Helper()
	out, err := EnsurePullRequestTrigger(in)
	require.NoError(t, err)
	return out
}

func TestEnsurePRAddsToMapping(t *testing.T) {
	out := parseYAML(t, ensurePR(t, "on:\n  push:\n    branches: [main]\njobs: {}\n"))
	on := out["on"].(map[string]any)
	_, hasPR := on["pull_request"]
	_, hasPush := on["push"]
	assert.True(t, hasPR)
	assert.True(t, hasPush)
}

func TestEnsurePRNoOpWhenPresentMapping(t *testing.T) {
	in := "on:\n  pull_request:\n  push:\njobs: {}\n"
	once := ensurePR(t, in)
	twice := ensurePR(t, once)
	assert.Equal(t, once, twice)
}

func TestEnsurePRAppendsToSequence(t *testing.T) {
	out := parseYAML(t, ensurePR(t, "on: [push, workflow_dispatch]\njobs: {}\n"))
	on := toStrings(out["on"].([]any))
	assert.Contains(t, on, "pull_request")
	assert.Contains(t, on, "push")
	assert.Contains(t, on, "workflow_dispatch")
}

func TestEnsurePRNoDuplicateInSequence(t *testing.T) {
	out := parseYAML(t, ensurePR(t, "on: [push, pull_request]\njobs: {}\n"))
	on := toStrings(out["on"].([]any))
	count := 0
	for _, e := range on {
		if e == "pull_request" {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

func TestEnsurePRPromotesScalar(t *testing.T) {
	out := parseYAML(t, ensurePR(t, "on: push\njobs: {}\n"))
	on := toStrings(out["on"].([]any))
	assert.Contains(t, on, "push")
	assert.Contains(t, on, "pull_request")
}

func TestEnsurePRPreservesComments(t *testing.T) {
	assert.Contains(t, ensurePR(t, "# triggers\non:\n  push:\njobs: {}\n"), "# triggers")
}

func toStrings(items []any) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i], _ = it.(string)
	}
	return out
}
