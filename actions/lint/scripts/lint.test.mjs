import { test } from "node:test";
import assert from "node:assert/strict";
import { lintSummary } from "./lint.mjs";

test("lintSummary reports the given paths and config", () => {
  assert.equal(lintSummary("src", "eslint.config.js"), "Linting src...\nUsing config: eslint.config.js");
});

test("lintSummary falls back to defaults when inputs are empty", () => {
  assert.equal(lintSummary("", ""), "Linting ....\nUsing config: .eslintrc");
  assert.equal(lintSummary(undefined, undefined), "Linting ....\nUsing config: .eslintrc");
});
