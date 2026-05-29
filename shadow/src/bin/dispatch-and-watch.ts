import { appendFileSync } from 'node:fs';
import { requireArgs } from '../core/args.ts';
import { requireEnv } from '../core/requireEnv.ts';
import { shadowBranchName } from '../core/shadowBranchName.ts';
import { renderShadowSummary, type ShadowResult } from '../core/summary.ts';
import type { ShadowContext } from '../core/dispatch.ts';
import * as github from '../adapters/github.ts';

/** Append markdown to the GitHub job summary (the shadow check's page); also log it. */
function writeJobSummary(markdown: string): void {
  console.log(markdown);
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

  // One clear, linked annotation up front (PR / repos / run) — no per-step polling noise.
  console.log(
    `::notice title=Shadow test::🛰️ ${consumerRepo} vs ${workflowsRepo}#${workflowsPr} — runner run: ${runUrl}`,
  );

  const summarize = async (result: ShadowResult): Promise<string | null> => {
    const prUrl = await github.findPrUrl({ repo: runnerRepo, branch, token });
    writeJobSummary(
      renderShadowSummary({ consumerRepo, consumerRef, workflowsRepo, workflowsRef, workflowsPr, result, runUrl, prUrl }),
    );
    return prUrl;
  };

  try {
    await github.watchRun({ runnerRepo, runId, token });
  } catch (error) {
    const prUrl = await summarize('failed');
    // Red annotation that links straight to where the failure is visible.
    console.log(`::error title=Shadow test failed::❌ ${consumerRepo} — see ${prUrl ?? runUrl}`);
    throw error;
  }
  await summarize('passed');
  console.log(`::notice title=Shadow test passed::✅ ${consumerRepo}`);
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
