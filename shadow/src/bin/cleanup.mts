import { readFileSync } from "node:fs";
import { requireArgs } from "../core/args.mts";
import { requireEnv } from "../core/requireEnv.mts";
import { parseConsumers } from "../core/parseConsumers.mts";
import { shadowBranchName } from "../core/shadowBranchName.mts";
import * as github from "../adapters/github.mts";

/** Workflows entrypoint (runs on pull_request: closed). Tears down every consumer's shadow PR +
 * branch in the runner for this workflows PR. */
async function main(): Promise<void> {
  const args = requireArgs(["runner-repo", "workflows-pr", "consumers-file"]);
  const runnerRepo = args["runner-repo"];
  const workflowsPr = Number(args["workflows-pr"]);
  const token = requireEnv("SHADOW_PAT");
  const consumers = parseConsumers(readFileSync(args["consumers-file"], "utf8"));

  for (const { repo } of consumers) {
    const branch = shadowBranchName({ prNumber: workflowsPr, consumerRepo: repo });
    await github.closePrAndDeleteBranch({ repo: runnerRepo, branch, token });
    console.log(`cleaned up ${runnerRepo}:${branch}`);
  }
}

try {
  await main();
} catch (error) {
  console.error(error);
  process.exit(1);
}
