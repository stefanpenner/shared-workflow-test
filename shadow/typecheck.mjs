// Isolated type-check: install shadow's deps and run `tsc --noEmit` over the .mts. Node-driven
// (no bash) so the workflow step is a single `node shadow/typecheck.mjs`. The runtime never runs
// tsc — Node 24 strips types on invocation. Installing also brings `yaml` in for the test step.
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { dirname } from "node:path";

const cwd = dirname(fileURLToPath(import.meta.url));

function step(command, args) {
  const result = spawnSync(command, args, { cwd, stdio: "inherit" });
  if (result.status !== 0) process.exit(result.status ?? 1);
}

step("npm", ["ci", "--no-fund", "--no-audit"]); // yaml + typescript + @types/node (lock-pinned)
step("npx", ["--no-install", "tsc", "--noEmit"]);
