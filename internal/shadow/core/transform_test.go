package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The real consumer ci.yaml shape (a .yml typo and no pull_request trigger).
const consumerWorkflow = "name: CI\non:\n  push:\n    branches: [main]\n  workflow_dispatch:\njobs:\n  ci:\n    uses: stefanpenner-cs/reusable-workflows/.github/workflows/shared.yml@main\n"

func TestTransformAppliesBoth(t *testing.T) {
	out, err := TransformWorkflowFile(consumerWorkflow, workflowsRepo, sha)
	require.NoError(t, err)
	parsed := parseYAML(t, out)
	assert.Equal(t, workflowsRepo+"/.github/workflows/shared.yaml@"+sha, nav(parsed, "jobs", "ci", "uses"))
	assert.Equal(t, sha, nav(parsed, "jobs", "ci", "with", "ref"))
	on := parsed["on"].(map[string]any)
	_, hasPR := on["pull_request"]
	_, hasPush := on["push"]
	assert.True(t, hasPR)
	assert.True(t, hasPush)
}

func TestTransformIsIdempotent(t *testing.T) {
	once, err := TransformWorkflowFile(consumerWorkflow, workflowsRepo, sha)
	require.NoError(t, err)
	twice, err := TransformWorkflowFile(once, workflowsRepo, sha)
	require.NoError(t, err)
	assert.Equal(t, once, twice)
}
