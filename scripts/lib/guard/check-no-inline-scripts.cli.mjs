// Discover repo action + workflow YAML, run the inline-script guard, exit non-zero on
// any violation. Invoked by .github/workflows/test.yaml.
import { readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";
import { inlineErrors } from "./check-no-inline-scripts.mjs";

function discover() {
  const files = [];
  try {
    for (const name of readdirSync("actions")) {
      const candidate = join("actions", name, "action.yaml");
      try {
        if (statSync(candidate).isFile()) files.push(candidate);
      } catch {
        // action without an action.yaml; skip
      }
    }
  } catch {
    // no actions/ directory
  }
  try {
    for (const name of readdirSync(".github/workflows")) {
      if (name.endsWith(".yaml") || name.endsWith(".yml")) {
        files.push(join(".github/workflows", name));
      }
    }
  } catch {
    // no .github/workflows directory
  }
  return files;
}

const files = discover();
let violations = 0;
for (const file of files) {
  for (const { line, message } of inlineErrors(readFileSync(file, "utf8"))) {
    console.error(`✗ ${file}:${line}  ${message}`);
    violations++;
  }
}

if (violations > 0) {
  console.error(`\n✗ no-inline-scripts: ${violations} violation(s) across ${files.length} file(s)`);
  process.exit(1);
}
console.log(`✓ no-inline-scripts: ${files.length} file(s) clean`);
