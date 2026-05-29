// Pure logic for the Lint action.
export function lintSummary(paths, config) {
  const resolvedPaths = (paths ?? "").trim() || ".";
  const resolvedConfig = (config ?? "").trim() || ".eslintrc";
  return `Linting ${resolvedPaths}...\nUsing config: ${resolvedConfig}`;
}
