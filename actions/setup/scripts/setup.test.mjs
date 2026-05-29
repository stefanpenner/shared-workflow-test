import { test } from "node:test";
import assert from "node:assert/strict";
import { resolveNodeVersion, report, renderOutputs } from "./setup.mjs";

test("resolveNodeVersion trims and returns the requested version", () => {
  assert.equal(resolveNodeVersion(" 20 "), "20");
  assert.equal(resolveNodeVersion("18"), "18");
});

test("resolveNodeVersion throws when empty or undefined", () => {
  assert.throws(() => resolveNodeVersion(""), /required/);
  assert.throws(() => resolveNodeVersion(undefined), /required/);
});

test("report shows the project and node version", () => {
  assert.equal(report("demo", "20"), "▸ Setup\n  project       demo\n  node version  20");
});

test("report falls back when the project name is empty", () => {
  assert.match(report("", "20"), /project +\(unknown project\)/);
  assert.match(report(undefined, "20"), /project +\(unknown project\)/);
});

test("renderOutputs formats key=value lines with a trailing newline", () => {
  assert.equal(renderOutputs({ node_version: "20" }), "node_version=20\n");
  assert.equal(renderOutputs({ a: "1", b: "2" }), "a=1\nb=2\n");
});
