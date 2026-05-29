import { readFileSync, appendFileSync } from 'node:fs';
import { requireArgs } from '../core/args.mts';
import { requireEnv } from '../core/requireEnv.mts';
import { parseConsumers } from '../core/parseConsumers.mts';

/** Workflows setup entrypoint: validate shadow-consumers.json and emit it as a matrix for the
 * downstream job (`consumers=<json>` on $GITHUB_OUTPUT). */
function main(): void {
  const { 'consumers-file': consumersFile } = requireArgs(['consumers-file']);
  const consumers = parseConsumers(readFileSync(consumersFile, 'utf8'));
  appendFileSync(requireEnv('GITHUB_OUTPUT'), `consumers=${JSON.stringify(consumers)}\n`);
  console.log(`✅ ${consumers.length} consumer(s): ${consumers.map((c) => `${c.repo}@${c.ref}`).join(', ') || '(none)'}`);
}

main();
