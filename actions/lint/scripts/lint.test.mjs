import { test } from "node:test";
import assert from "node:assert/strict";
import { report } from "./lint.mjs";

test("report shows the given paths and config", () => {
  assert.equal(report("src", "eslint.config.js"), "▸ Lint\n  paths   src\n  config  eslint.config.js");
});

test("report falls back to defaults when inputs are empty", () => {
  assert.equal(report("", ""), "▸ Lint\n  paths   .\n  config  .eslintrc");
  assert.equal(report(undefined, undefined), "▸ Lint\n  paths   .\n  config  .eslintrc");
});
