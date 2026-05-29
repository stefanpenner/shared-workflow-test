import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { requireEnv } from '../core/requireEnv.ts';
import { mirrorTree } from '../core/mirrorTree.ts';
import { patchWorkflowsInDir } from '../adapters/workflows.ts';
import * as git from '../adapters/git.ts';
import * as github from '../adapters/github.ts';

/**
 * Receiver entrypoint (runs in H on workflow_dispatch). Mirrors the consumer's code onto a shadow
 * branch with its workflows repointed at the workflows PR SHA, opens/refreshes a real PR, and blocks
 * on that PR's checks — so this run's exit status IS the shadow-test result the workflows watches.
 */
async function main(): Promise<void> {
  const workflowsRepo = requireEnv('WORKFLOWS_REPO');
  const workflowsRef = requireEnv('WORKFLOWS_REF');
  const consumerRepo = requireEnv('CONSUMER_REPO');
  const consumerRef = requireEnv('CONSUMER_REF');
  const workflowsPr = requireEnv('WORKFLOWS_PR');
  const branch = requireEnv('BRANCH');
  const token = requireEnv('SHADOW_PAT');
  const runnerRepo = requireEnv('GITHUB_REPOSITORY');

  const work = mkdtempSync(join(tmpdir(), 'shadow-'));
  const mirrorDir = join(work, 'consumer');
  const shadowDir = join(work, 'runner');

  // Clone the consumer's files, and a clean copy of the runner to build the branch in (a fresh
  // clone avoids leaking the runner's node_modules into the shadow commit).
  await git.cloneShallow({ repo: consumerRepo, ref: consumerRef, dir: mirrorDir, token });
  await git.cloneShallow({ repo: runnerRepo, ref: 'main', dir: shadowDir, token });

  await git.configureBotIdentity(shadowDir);
  await git.resetBranchToEmptyTree(shadowDir, branch);
  mirrorTree(mirrorDir, shadowDir);

  const patched = patchWorkflowsInDir(shadowDir, { workflowsRepo, workflowsRef });
  console.log(`patched workflows: ${patched.join(', ') || '(none — consumer has no workflows?)'}`);

  await git.commitAll(
    shadowDir,
    `shadow: ${consumerRepo}@${consumerRef} vs ${workflowsRepo}@${workflowsRef} (workflows PR #${workflowsPr})`,
  );
  await git.forcePush(shadowDir, branch);
  const sha = await git.headSha(shadowDir);

  const prUrl = await github.ensurePr({
    repo: runnerRepo,
    branch,
    base: 'main',
    title: `Shadow: ${consumerRepo} vs ${workflowsRepo}#${workflowsPr}`,
    body: [
      'Automated shadow test — **do not merge**.',
      '',
      `- Consumer: \`${consumerRepo}@${consumerRef}\``,
      `- Workflows draft: \`${workflowsRepo}@${workflowsRef}\``,
      `- Workflows PR: ${workflowsRepo}#${workflowsPr}`,
      '',
      "This PR exists only to run the consumer's CI under a real `pull_request` event.",
    ].join('\n'),
    token,
  });
  console.log(`shadow PR: ${prUrl} (head ${sha})`);

  await github.watchCommitRun({ repo: runnerRepo, sha, token });
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
