// Pure diagnostics formatting for the Debug action. `exec(file, args)` is injected so
// tests need no real filesystem or git: it returns stdout as a string and throws on
// failure (matching execFileSync). debug.cli.mjs supplies the real implementation.
// Output is wrapped in collapsible GitHub Actions log groups to keep runs tidy.
import { group } from "../../../scripts/lib/log/format.mjs";

function tryExec(exec, file, args, fallback) {
  try {
    return exec(file, args);
  } catch (err) {
    if (err instanceof TypeError) throw err; // a bug in our code, not a failed probe
    return fallback; // command missing or exited non-zero: expected for a probe
  }
}

export function treeReport(exec, env = {}) {
  const home = env.HOME ?? "";
  const workspace = env.GITHUB_WORKSPACE ?? ".";
  const body = [
    "$HOME:",
    tryExec(exec, "ls", ["-la", home], "(unavailable)"),
    "",
    "$HOME/work/:",
    tryExec(exec, "ls", ["-la", `${home}/work/`], "(not found)"),
    "",
    "project tree:",
    tryExec(
      exec,
      "find",
      [workspace, "-not", "-path", "*/.git/*", "-not", "-path", "*/node_modules/*"],
      "(unavailable)",
    ),
  ].join("\n");
  return group("Environment", body);
}

export function gitReport(exec) {
  try {
    exec("git", ["rev-parse", "--git-dir"]);
  } catch (err) {
    if (err instanceof TypeError) throw err; // a bug in our code, not a missing repo
    return group("Git status", "No git repository in working directory");
  }
  const body = [
    exec("git", ["status"]),
    "",
    "unstaged changes:",
    exec("git", ["diff"]),
    "",
    "staged changes:",
    exec("git", ["diff", "--cached"]),
  ].join("\n");
  return group("Git status", body);
}
