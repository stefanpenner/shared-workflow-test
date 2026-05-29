import { readFileSync } from 'node:fs';
import { requireArgs } from '../core/args.ts';
import { requireEnv } from '../core/requireEnv.ts';
import { parseConsumers } from '../core/parseConsumers.ts';
import { shadowBranchName } from '../core/shadowBranchName.ts';
import * as github from '../adapters/github.ts';

/** Workflows entrypoint (runs on pull_request: closed). Tears down every consumer's shadow PR +
 * branch in the runner for this workflows PR. */
async function main(): Promise<void> {
  const args = requireArgs(['runner-repo', 'workflows-pr', 'consumers-file']);
  const runnerRepo = args['runner-repo'];
  const workflowsPr = Number(args['workflows-pr']);
  const token = requireEnv('SHADOW_PAT');
  const consumers = parseConsumers(readFileSync(args['consumers-file'], 'utf8'));

  for (const { repo } of consumers) {
    const branch = shadowBranchName({ prNumber: workflowsPr, consumerRepo: repo });
    await github.closePrAndDeleteBranch({ repo: runnerRepo, branch, token });
    console.log(`cleaned up ${runnerRepo}:${branch}`);
  }
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
