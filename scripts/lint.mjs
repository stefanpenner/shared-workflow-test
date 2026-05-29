// Lint gate for the repo's own source: install the dev tooling, then run ESLint (correctness
// over .mjs + YAML) and Prettier (formatting). npm appears only here and under shadow/. Keeping
// this a single `node scripts/lint.mjs` lets the workflow step satisfy the no-inline-scripts rule.
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { dirname } from "node:path";

const root = dirname(dirname(fileURLToPath(import.meta.url))); // scripts/.. -> repo root

function step(command, args) {
  const result = spawnSync(command, args, { cwd: root, stdio: "inherit" });
  if (result.status !== 0) process.exit(result.status ?? 1);
}

step("npm", ["ci", "--no-fund", "--no-audit"]);
step("npx", ["--no-install", "eslint", "."]);
step("npx", ["--no-install", "prettier", "--check", "**/*.{mjs,yaml,yml}"]);
