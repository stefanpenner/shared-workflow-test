export interface ShadowContext {
  workflowsRepo: string;
  workflowsRef: string;
  consumerRepo: string;
  consumerRef: string;
  workflowsPr: number;
  branch: string;
}

/** Map the shadow context to the receiver workflow_dispatch `inputs` (all values are strings). */
export function buildDispatchInputs(ctx: ShadowContext): Record<string, string> {
  return {
    workflows_repo: ctx.workflowsRepo,
    workflows_ref: ctx.workflowsRef,
    consumer_repo: ctx.consumerRepo,
    consumer_ref: ctx.consumerRef,
    workflows_pr: String(ctx.workflowsPr),
    branch: ctx.branch,
  };
}

/**
 * Read the run id from a `workflow_dispatch` response created with `return_run_details: true`.
 * REST shape: `{ workflow_run_id, run_url, html_url }`.
 */
export type RunState = 'pending' | 'success' | 'failure';

/** Classify a workflow run from its `status`/`conclusion`. Pure: drives quiet polling (log only on
 * change, return on success, throw on failure) instead of the noisy `gh run watch` redraw. */
export function classifyRunState(status: string, conclusion: string | null): RunState {
  if (status !== 'completed') return 'pending';
  return conclusion === 'success' ? 'success' : 'failure';
}

export function extractRunId(response: unknown): number {
  const id = (response as { workflow_run_id?: unknown } | null)?.workflow_run_id;
  if (typeof id !== 'number' || !Number.isFinite(id)) {
    throw new Error(`workflow_dispatch response has no numeric workflow_run_id: ${JSON.stringify(response)}`);
  }
  return id;
}
