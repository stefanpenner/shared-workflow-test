package core

import "fmt"

// ShadowContext is the per-consumer context threaded through a shadow run.
type ShadowContext struct {
	WorkflowsRepo string
	WorkflowsRef  string
	ConsumerRepo  string
	ConsumerRef   string
	WorkflowsPR   int
	Branch        string
}

// BuildDispatchInputs maps the context to the receiver workflow_dispatch `inputs` (all strings).
func BuildDispatchInputs(ctx ShadowContext) map[string]string {
	return map[string]string{
		"workflows_repo": ctx.WorkflowsRepo,
		"workflows_ref":  ctx.WorkflowsRef,
		"consumer_repo":  ctx.ConsumerRepo,
		"consumer_ref":   ctx.ConsumerRef,
		"workflows_pr":   fmt.Sprintf("%d", ctx.WorkflowsPR),
		"branch":         ctx.Branch,
	}
}

// RunState is a workflow run's coarse state, driving quiet polling.
type RunState string

const (
	RunPending RunState = "pending"
	RunSuccess RunState = "success"
	RunFailure RunState = "failure"
)

// ClassifyRunState classifies a run from its status/conclusion. An empty conclusion means "null"
// (not yet concluded or no result) — treated as failure once status is "completed".
func ClassifyRunState(status, conclusion string) RunState {
	if status != "completed" {
		return RunPending
	}
	if conclusion == "success" {
		return RunSuccess
	}
	return RunFailure
}

// ExtractRunID reads the numeric workflow_run_id from a `return_run_details:true` dispatch response
// (REST shape `{workflow_run_id, run_url, html_url}`, JSON-decoded into a map).
func ExtractRunID(response any) (int, error) {
	if m, ok := response.(map[string]any); ok {
		if id, ok := m["workflow_run_id"].(float64); ok {
			return int(id), nil
		}
	}
	return 0, fmt.Errorf("workflow_dispatch response has no numeric workflow_run_id: %v", response)
}
