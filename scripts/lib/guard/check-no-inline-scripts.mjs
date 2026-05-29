// Enforce "no inline scripts": every `run:` in an action or workflow must be a single
// external-script invocation, never a block of embedded shell logic. Pure + exported
// for testing; file discovery and process exit live in check-no-inline-scripts.cli.mjs.

// Step names allowed to keep inline `run:` logic, each with a justification.
export const ALLOW_NAMES = new Set([
  // Pre-checkout bootstrap in the reusable workflow: it runs before any checkout, so
  // nothing is on disk yet and it cannot call a repo-local script.
  // See .github/workflows/shared.yaml.
  "Set up shared actions (exclude from git)",
]);

// Shell metacharacters that indicate embedded logic rather than a single invocation.
const SHELL_OPS = /&&|\|\||[;|`<>]|\$\(/;

function unquote(text) {
  const t = text.trim();
  if ((t.startsWith('"') && t.endsWith('"')) || (t.startsWith("'") && t.endsWith("'"))) {
    return t.slice(1, -1);
  }
  return t;
}

// True when `value` is a single external-script invocation with no embedded logic.
export function isSingleInvocation(value) {
  // Drop GHA expressions (e.g. ${{ github.action_path }}) before inspecting.
  const v = value.replace(/\$\{\{[^}]*\}\}/g, "X").trim();
  if (!v) return false;
  if (SHELL_OPS.test(v)) return false;
  // Inline eval defeats the rule even without shell operators.
  if (/^(node|deno|bun)\s+(-e|--eval|-p|--print)\b/.test(v)) return false;
  // Allowed: run through an interpreter, or a bare path to a script file.
  if (/^(node|bash|sh)\s+\S/.test(v)) return true;
  if (/^\S+\.(mjs|cjs|js|sh)$/.test(v)) return true;
  return false;
}

// Scan one YAML document, returning [{ line, message }] for each violation.
export function inlineErrors(yamlText) {
  const lines = yamlText.split("\n");
  const errors = [];
  let lastName = "";
  for (let i = 0; i < lines.length; i++) {
    const nameMatch = lines[i].match(/^\s*-?\s*name:\s*(.+?)\s*$/);
    if (nameMatch) lastName = unquote(nameMatch[1]);

    const runMatch = lines[i].match(/^\s*-?\s*run:\s*(.*)$/);
    if (!runMatch) continue;

    const allowed = ALLOW_NAMES.has(lastName);
    // A name applies to a single step; don't let a later unnamed run inherit it.
    lastName = "";
    if (allowed) continue;

    const raw = runMatch[1].trim();
    if (raw === "" || /^[|>][+-]?\d*$/.test(raw)) {
      errors.push({ line: i + 1, message: "block scalar run: — move logic into an external script" });
      continue;
    }
    const value = unquote(raw);
    if (!isSingleInvocation(value)) {
      errors.push({ line: i + 1, message: `inline logic in run: "${value}" — call an external script instead` });
    }
  }
  return errors;
}
