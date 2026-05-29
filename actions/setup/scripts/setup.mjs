// Pure logic for the Setup action. No side effects on import and no env reads here,
// so setup.test.mjs can import and assert these directly. The GHA-specific I/O
// (reading env, appending to $GITHUB_OUTPUT) lives in setup.cli.mjs.
import { section } from "../../../scripts/lib/log/format.mjs";

export function resolveNodeVersion(input) {
  const version = (input ?? "").trim();
  if (!version) throw new Error("node-version is required");
  return version;
}

export function report(projectName, nodeVersion) {
  const name = (projectName ?? "").trim() || "(unknown project)";
  return section("Setup", { project: name, "node version": nodeVersion });
}

// Render a { key: value } map as GHA output lines (key=value), trailing newline.
export function renderOutputs(outputs) {
  return (
    Object.entries(outputs)
      .map(([key, value]) => `${key}=${value}`)
      .join("\n") + "\n"
  );
}
