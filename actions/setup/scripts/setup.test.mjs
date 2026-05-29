import { test } from "node:test";
import assert from "node:assert/strict";
import { resolveNodeVersion, greeting, renderOutputs } from "./setup.mjs";

test("resolveNodeVersion trims and returns the requested version", () => {
  assert.equal(resolveNodeVersion(" 20 "), "20");
  assert.equal(resolveNodeVersion("18"), "18");
});

test("resolveNodeVersion throws when empty or undefined", () => {
  assert.throws(() => resolveNodeVersion(""), /required/);
  assert.throws(() => resolveNodeVersion(undefined), /required/);
});

test("greeting uses the project name", () => {
  assert.equal(greeting("demo"), "Setting up environment for demo...");
});

test("greeting falls back when the project name is empty", () => {
  assert.equal(greeting(""), "Setting up environment for (unknown project)...");
  assert.equal(greeting(undefined), "Setting up environment for (unknown project)...");
});

test("renderOutputs formats key=value lines with a trailing newline", () => {
  assert.equal(renderOutputs({ node_version: "20" }), "node_version=20\n");
  assert.equal(renderOutputs({ a: "1", b: "2" }), "a=1\nb=2\n");
});
