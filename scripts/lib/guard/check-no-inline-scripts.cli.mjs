// Discover repo action + workflow YAML, run the inline-script guard, exit non-zero on
// any violation. Invoked by .github/workflows/test.yaml.
import { readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";
import { inlineErrors } from "./check-no-inline-scripts.mjs";

// readdir/stat, tolerating only "it isn't there" (ENOENT); anything else propagates.
function listDir(dir) {
  try {
    return readdirSync(dir);
  } catch (err) {
    if (err.code === "ENOENT") return [];
    throw err;
  }
}

function isFile(path) {
  try {
    return statSync(path).isFile();
  } catch (err) {
    if (err.code === "ENOENT") return false;
    throw err;
  }
}

function discover() {
  const files = [];
  for (const name of listDir("actions")) {
    const candidate = join("actions", name, "action.yaml");
    if (isFile(candidate)) files.push(candidate);
  }
  for (const name of listDir(".github/workflows")) {
    if (name.endsWith(".yaml") || name.endsWith(".yml"))
      files.push(join(".github/workflows", name));
  }
  return files;
}

const files = discover();
let violations = 0;
for (const file of files) {
  let content;
  try {
    content = readFileSync(file, "utf8");
  } catch (err) {
    throw new Error(`could not read ${file} for inline-script check`, { cause: err });
  }
  for (const { line, message } of inlineErrors(content)) {
    console.error(`✗ ${file}:${line}  ${message}`);
    violations++;
  }
}

if (violations > 0) {
  console.error(`\n✗ no-inline-scripts: ${violations} violation(s) across ${files.length} file(s)`);
  process.exit(1);
}
console.log(`✓ no-inline-scripts: ${files.length} file(s) clean`);
