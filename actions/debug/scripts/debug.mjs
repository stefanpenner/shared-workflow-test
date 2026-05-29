// Pure diagnostics formatting for the Debug action. `exec(file, args)` is injected so
// tests need no real filesystem or git: it returns stdout as a string and throws on
// failure (matching execFileSync). debug.cli.mjs supplies the real implementation.

function tryExec(exec, file, args, fallback) {
  try {
    return exec(file, args);
  } catch {
    return fallback;
  }
}

export function treeReport(exec, env = {}) {
  const home = env.HOME ?? "";
  const workspace = env.GITHUB_WORKSPACE ?? ".";
  return [
    "--- $HOME top-level ---",
    tryExec(exec, "ls", ["-la", home], ""),
    "",
    "--- $HOME/work/ top-level ---",
    tryExec(exec, "ls", ["-la", `${home}/work/`], "(not found)"),
    "",
    "--- project tree ---",
    tryExec(
      exec,
      "find",
      [workspace, "-not", "-path", "*/.git/*", "-not", "-path", "*/node_modules/*"],
      "",
    ),
  ].join("\n");
}

export function gitReport(exec) {
  try {
    exec("git", ["rev-parse", "--git-dir"]);
  } catch {
    return "No git repository in working directory";
  }
  return [
    exec("git", ["status"]),
    "--- unstaged changes ---",
    exec("git", ["diff"]),
    "--- staged changes ---",
    exec("git", ["diff", "--cached"]),
  ].join("\n");
}
