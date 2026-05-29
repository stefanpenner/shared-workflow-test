import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { requireArgs } from "../core/args.mts";
import { requireEnv } from "../core/requireEnv.mts";
import { mirrorTree } from "../core/mirrorTree.mts";
import { patchWorkflowsInDir } from "../adapters/workflows.mts";
import * as git from "../adapters/git.mts";
import * as github from "../adapters/github.mts";

/**
 * Receiver entrypoint (runs in H on workflow_dispatch). Mirrors the consumer's code onto a shadow
 * branch with its workflows repointed at the workflows PR SHA, opens/refreshes a real PR, and blocks
 * on that PR's checks — so this run's exit status IS the shadow-test result the workflows watches.
 */
async function main(): Promise<void> {
  const args = requireArgs([
    "workflows-repo",
    "workflows-ref",
    "consumer-repo",
    "consumer-ref",
    "workflows-pr",
    "branch",
    "runner-repo",
  ]);
  const workflowsRepo = args["workflows-repo"];
  const workflowsRef = args["workflows-ref"];
  const consumerRepo = args["consumer-repo"];
  const consumerRef = args["consumer-ref"];
  const workflowsPr = args["workflows-pr"];
  const branch = args["branch"];
  const runnerRepo = args["runner-repo"];
  const token = requireEnv("SHADOW_PAT");

  const work = mkdtempSync(join(tmpdir(), "shadow-"));
  const mirrorDir = join(work, "consumer");
  const shadowDir = join(work, "runner");

  // Clone the consumer's files, and a clean copy of the runner to build the branch in (a fresh
  // clone avoids leaking the runner's node_modules into the shadow commit).
  await git.cloneShallow({ repo: consumerRepo, ref: consumerRef, dir: mirrorDir, token });
  await git.cloneShallow({ repo: runnerRepo, ref: "main", dir: shadowDir, token });

  await git.configureBotIdentity(shadowDir);
  await git.resetBranchToEmptyTree(shadowDir, branch);
  mirrorTree(mirrorDir, shadowDir);

  const patched = patchWorkflowsInDir(shadowDir, { workflowsRepo, workflowsRef });
  console.log(`patched workflows: ${patched.join(", ") || "(none — consumer has no workflows?)"}`);

  await git.commitAll(
    shadowDir,
    `shadow: ${consumerRepo}@${consumerRef} vs ${workflowsRepo}@${workflowsRef} (workflows PR #${workflowsPr})`,
  );
  await git.forcePush(shadowDir, branch);
  const sha = await git.headSha(shadowDir);

  const prUrl = await github.ensurePr({
    repo: runnerRepo,
    branch,
    base: "main",
    title: `Shadow: ${consumerRepo} vs ${workflowsRepo}#${workflowsPr}`,
    body: [
      "Automated shadow test — **do not merge**.",
      "",
      `- Consumer: \`${consumerRepo}@${consumerRef}\``,
      `- Workflows draft: \`${workflowsRepo}@${workflowsRef}\``,
      `- Workflows PR: ${workflowsRepo}#${workflowsPr}`,
      "",
      "This PR exists only to run the consumer's CI under a real `pull_request` event.",
    ].join("\n"),
    token,
  });
  console.log(`shadow PR: ${prUrl} (head ${sha})`);

  await github.watchCommitRun({ repo: runnerRepo, sha, token });
}

try {
  await main();
} catch (error) {
  console.error(error);
  process.exit(1);
}
