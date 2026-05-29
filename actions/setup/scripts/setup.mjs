// Pure logic for the Setup action. No side effects on import and no env reads here,
// so setup.test.mjs can import and assert these directly. The GHA-specific I/O
// (reading env, appending to $GITHUB_OUTPUT) lives in setup.cli.mjs.

export function resolveNodeVersion(input) {
  const version = (input ?? "").trim();
  if (!version) throw new Error("node-version is required");
  return version;
}

export function greeting(projectName) {
  const name = (projectName ?? "").trim() || "(unknown project)";
  return `Setting up environment for ${name}...`;
}

// Render a { key: value } map as GHA output lines (key=value), trailing newline.
export function renderOutputs(outputs) {
  return (
    Object.entries(outputs)
      .map(([key, value]) => `${key}=${value}`)
      .join("\n") + "\n"
  );
}
