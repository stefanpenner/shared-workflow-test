import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { disallowedDeps } from '../core/deps.mts';

/** Enforce the policy: `yaml` is the only runtime dependency (devDeps = the isolated typecheck only). */
function main(): void {
  const packageJsonPath = join(dirname(fileURLToPath(import.meta.url)), '..', '..', 'package.json');
  const pkg = JSON.parse(readFileSync(packageJsonPath, 'utf8'));

  const extra = disallowedDeps(pkg, { deps: ['yaml'], devDeps: ['typescript', '@types/node'] });
  if (extra.length > 0) {
    console.error(`❌ disallowed dependencies: ${extra.join(', ')} — only "yaml" (plus the typecheck tooling) is allowed`);
    process.exit(1);
  }
  console.log('✅ dependencies OK — yaml only');
}

main();
