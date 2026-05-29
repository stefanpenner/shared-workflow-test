// Isolated type-check: install TypeScript into shadow/ and run `tsc --noEmit` over the .mts.
// This is the ONLY place npm appears — the runtime strips types via Node 24 and never needs it.
// Driven by Node (no bash) so the workflow step is a single `node shadow/typecheck.mjs`.
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { dirname } from 'node:path';

const cwd = dirname(fileURLToPath(import.meta.url));

function step(command, args) {
  const result = spawnSync(command, args, { cwd, stdio: 'inherit' });
  if (result.status !== 0) process.exit(result.status ?? 1);
}

step('npm', ['install', '--no-fund', '--no-audit']); // install devDeps (typescript, @types/node)
step('npx', ['--no-install', 'tsc', '--noEmit']);
