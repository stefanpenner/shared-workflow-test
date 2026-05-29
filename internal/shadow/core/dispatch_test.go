package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildDispatchInputs(t *testing.T) {
	got := BuildDispatchInputs(ShadowContext{
		WorkflowsRepo: "stefanpenner-cs/reusable-workflows",
		WorkflowsRef:  "deadbeef",
		ConsumerRepo:  "o/r",
		ConsumerRef:   "main",
		WorkflowsPR:   7,
		Branch:        "shadow/pr-7-o-r",
	})
	assert.Equal(t, map[string]string{
		"workflows_repo": "stefanpenner-cs/reusable-workflows",
		"workflows_ref":  "deadbeef",
		"consumer_repo":  "o/r",
		"consumer_ref":   "main",
		"workflows_pr":   "7",
		"branch":         "shadow/pr-7-o-r",
	}, got)
}

func TestExtractRunID(t *testing.T) {
	id, err := ExtractRunID(map[string]any{"workflow_run_id": float64(1234567890)})
	require.NoError(t, err)
	assert.Equal(t, 1234567890, id)

	_, err = ExtractRunID(map[string]any{})
	assert.Error(t, err)
	_, err = ExtractRunID(map[string]any{"workflow_run_id": "nope"})
	assert.Error(t, err)
}

func TestClassifyRunState(t *testing.T) {
	assert.Equal(t, RunPending, ClassifyRunState("queued", ""))
	assert.Equal(t, RunPending, ClassifyRunState("in_progress", ""))
	assert.Equal(t, RunSuccess, ClassifyRunState("completed", "success"))
	assert.Equal(t, RunFailure, ClassifyRunState("completed", "failure"))
	assert.Equal(t, RunFailure, ClassifyRunState("completed", "cancelled"))
	assert.Equal(t, RunFailure, ClassifyRunState("completed", ""))
}
