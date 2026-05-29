import { appendFileSync } from 'node:fs';
import { parseArgs } from 'node:util';
import { requireEnv } from '../core/requireEnv.mts';
import { resolveContext } from '../core/resolveContext.mts';
import { capture } from '../adapters/exec.mts';

/**
 * Workflows setup entrypoint: resolve the PR number + head SHA to shadow-test and emit them on
 * $GITHUB_OUTPUT. Inputs vary by event, so flags are optional (= form, so empties are tolerated);
 * GH_TOKEN (env) authorizes the dispatch-time `gh pr view` lookup.
 */
async function main(): Promise<void> {
  const { values } = parseArgs({
    options: {
      'event-name': { type: 'string' },
      'pr-number': { type: 'string' },
      'head-sha': { type: 'string' },
      'input-pr': { type: 'string' },
      'workflows-repo': { type: 'string' },
    },
  });

  const { pr, sha } = await resolveContext({
    eventName: values['event-name'] ?? '',
    prNumber: values['pr-number'],
    headSha: values['head-sha'],
    inputPr: values['input-pr'],
    lookupHeadSha: async (prNumber) => {
      const repo = values['workflows-repo'];
      if (!repo) throw new Error('missing required --workflows-repo for the dispatch lookup');
      const out = await capture('gh', ['pr', 'view', prNumber, '-R', repo, '--json', 'headRefOid', '--jq', '.headRefOid']);
      return out.trim();
    },
  });

  appendFileSync(requireEnv('GITHUB_OUTPUT'), `pr=${pr}\nsha=${sha}\n`);
  console.log(`✅ resolved PR #${pr} @ ${sha.slice(0, 7)}`);
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
