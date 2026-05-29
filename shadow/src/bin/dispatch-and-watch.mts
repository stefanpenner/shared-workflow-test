import { appendFileSync } from 'node:fs';
import { requireArgs } from '../core/args.mts';
import { requireEnv } from '../core/requireEnv.mts';
import { shadowBranchName } from '../core/shadowBranchName.mts';
import { renderShadowSummary, renderShadowLog, workflowsPrUrl, type ShadowResult } from '../core/summary.mts';
import type { ShadowContext } from '../core/dispatch.mts';
import * as github from '../adapters/github.mts';

/** Append markdown to the job-summary page only (GitHub step logs don't render markdown). */
function appendSummary(markdown: string): void {
  const file = process.env.GITHUB_STEP_SUMMARY;
  if (file) appendFileSync(file, markdown);
}

/**
 * Workflows entrypoint (runs in P on a labeled pull_request, one invocation per consumer). Dispatches
 * the runner receiver, captures its run id natively, and watches it to completion. The job's exit
 * status is the PR's shadow check; the result + links are rendered into the job summary (the check's
 * markdown page) rather than a PR comment.
 */
async function main(): Promise<void> {
  const args = requireArgs(['runner-repo', 'workflows-repo', 'workflows-ref', 'workflows-pr', 'consumer-repo', 'consumer-ref']);
  const runnerRepo = args['runner-repo'];
  const workflowsRepo = args['workflows-repo'];
  const workflowsRef = args['workflows-ref'];
  const workflowsPr = Number(args['workflows-pr']);
  const consumerRepo = args['consumer-repo'];
  const consumerRef = args['consumer-ref'];
  const token = requireEnv('SHADOW_PAT');

  const branch = shadowBranchName({ prNumber: workflowsPr, consumerRepo });
  const ctx: ShadowContext = { workflowsRepo, workflowsRef, consumerRepo, consumerRef, workflowsPr, branch };

  const runId = await github.dispatchReceiver({ runnerRepo, ctx, token });
  const runUrl = `https://github.com/${runnerRepo}/actions/runs/${runId}`;

  // Up front: clean plain-text lines with full (clickable) URLs — logs don't render markdown.
  console.log(`🛰️  Shadow test: ${consumerRepo}@${consumerRef}`);
  console.log(`    vs ${workflowsPrUrl(workflowsRepo, workflowsPr)} — runner run: ${runUrl}`);
  console.log(`::notice title=Shadow test::🛰️ ${consumerRepo} — runner run: ${runUrl}`);

  const finish = async (result: ShadowResult): Promise<string | null> => {
    const prUrl = await github.findPrUrl({ repo: runnerRepo, branch, token });
    const fields = { consumerRepo, consumerRef, workflowsRepo, workflowsRef, workflowsPr, result, runUrl, prUrl };
    appendSummary(renderShadowSummary(fields)); // table → this job's summary (the check's artifact)
    for (const line of renderShadowLog(fields)) console.log(line); // plain text → the step log
    return prUrl;
  };

  try {
    await github.watchRun({ runnerRepo, runId, token });
  } catch (error) {
    const prUrl = await finish('failed');
    console.log(`::error title=Shadow test failed::❌ ${consumerRepo} — open ${prUrl ?? runUrl} to see the failing job`);
    throw error;
  }
  await finish('passed');
}

try {
  await main();
} catch (error) {
  console.error(error);
  process.exit(1);
}
