// Install shadow's deps (yaml + the isolated typecheck tooling). npm appears only here and in
// typecheck.mjs; the runtime strips types via Node 24 and imports the installed `yaml`.
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { dirname } from 'node:path';

const cwd = dirname(fileURLToPath(import.meta.url));
const result = spawnSync('npm', ['ci', '--no-fund', '--no-audit'], { cwd, stdio: 'inherit' });
process.exit(result.status ?? 1);
