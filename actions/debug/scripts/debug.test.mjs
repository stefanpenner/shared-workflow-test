import { test } from "node:test";
import assert from "node:assert/strict";
import { treeReport, gitReport } from "./debug.mjs";

// Build a fake exec: returns mapped stdout per "file arg arg" key, or throws for any
// command whose key starts with an entry in throwFor.
function fakeExec(map, throwFor = []) {
  return (file, args) => {
    const key = [file, ...args].join(" ");
    if (throwFor.some((prefix) => key.startsWith(prefix))) throw new Error("command failed");
    return map[key] ?? "";
  };
}

test("treeReport wraps section output in an Environment group", () => {
  const exec = fakeExec({
    "ls -la /home/u": "total 0",
    "ls -la /home/u/work/": "drwxr-xr-x work",
  });
  const out = treeReport(exec, { HOME: "/home/u", GITHUB_WORKSPACE: "/ws" });
  assert.match(out, /^::group::Environment\n/);
  assert.match(out, /\n::endgroup::$/);
  assert.match(out, /\$HOME:/);
  assert.match(out, /total 0/);
  assert.match(out, /project tree:/);
});

test("treeReport falls back to (not found) when the work dir listing fails", () => {
  const exec = fakeExec({}, ["ls -la /home/u/work/"]);
  assert.match(treeReport(exec, { HOME: "/home/u" }), /\(not found\)/);
});

test("treeReport defaults HOME and GITHUB_WORKSPACE when env is absent", () => {
  const calls = [];
  const exec = (file, args) => {
    calls.push([file, ...args].join(" "));
    return "";
  };
  treeReport(exec); // no env arg → exercises default {} and the ?? fallbacks
  assert.ok(calls.includes("ls -la "), "HOME defaults to empty string");
  assert.ok(
    calls.some((c) => c.startsWith("find .")),
    "GITHUB_WORKSPACE defaults to .",
  );
});

test("gitReport groups status and diff sections inside a repo", () => {
  const exec = fakeExec({
    "git rev-parse --git-dir": ".git",
    "git status": "On branch main",
    "git diff": "unstaged-diff",
    "git diff --cached": "staged-diff",
  });
  const out = gitReport(exec);
  assert.match(out, /^::group::Git status\n/);
  assert.match(out, /On branch main/);
  assert.match(out, /unstaged changes:/);
  assert.match(out, /unstaged-diff/);
  assert.match(out, /staged changes:/);
  assert.match(out, /staged-diff/);
});

test("gitReport reports no repository when git rev-parse fails", () => {
  const exec = fakeExec({}, ["git rev-parse"]);
  assert.equal(gitReport(exec), "::group::Git status\nNo git repository in working directory\n::endgroup::");
});
