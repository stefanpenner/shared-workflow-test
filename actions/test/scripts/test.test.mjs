import { test } from "node:test";
import assert from "node:assert/strict";
import { testSummary } from "./test.mjs";

test("testSummary reports the given suite and coverage", () => {
  assert.equal(testSummary("integration", "false"), "Running tests...\nSuite: integration\nCoverage: false");
});

test("testSummary falls back to defaults when inputs are empty", () => {
  assert.equal(testSummary("", ""), "Running tests...\nSuite: unit\nCoverage: true");
  assert.equal(testSummary(undefined, undefined), "Running tests...\nSuite: unit\nCoverage: true");
});
