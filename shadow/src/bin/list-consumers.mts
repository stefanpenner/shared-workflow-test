import { readFileSync, appendFileSync } from 'node:fs';
import { requireArgs } from '../core/args.mts';
import { requireEnv } from '../core/requireEnv.mts';
import { parseConsumers } from '../core/parseConsumers.mts';
import { renderShadowList } from '../core/summary.mts';

/** Workflows setup entrypoint: validate shadow-consumers.json, emit it as a matrix for the
 * downstream job (`consumers=<json>` on $GITHUB_OUTPUT), and write the index of all shadow tests
 * (consumer + shadow-PR links) to the job summary. */
function main(): void {
  const args = requireArgs(['consumers-file', 'workflows-pr', 'runner-repo']);
  const consumers = parseConsumers(readFileSync(args['consumers-file'], 'utf8'));

  appendFileSync(requireEnv('GITHUB_OUTPUT'), `consumers=${JSON.stringify(consumers)}\n`);

  const summaryFile = process.env.GITHUB_STEP_SUMMARY;
  if (summaryFile) {
    const list = renderShadowList({ consumers, workflowsPr: Number(args['workflows-pr']), runnerRepo: args['runner-repo'] });
    appendFileSync(summaryFile, list);
  }

  console.log(`✅ ${consumers.length} consumer(s): ${consumers.map((c) => `${c.repo}@${c.ref}`).join(', ') || '(none)'}`);
}

main();
