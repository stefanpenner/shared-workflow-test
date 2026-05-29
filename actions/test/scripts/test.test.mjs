import { test } from "node:test";
import assert from "node:assert/strict";
import { report } from "./test.mjs";

test("report shows the suite and coverage state", () => {
  assert.equal(report("integration", "true"), "▸ Test\n  suite     integration\n  coverage  enabled");
});

test("report treats coverage=false as disabled", () => {
  assert.match(report("unit", "false"), /coverage +disabled/);
});

test("report falls back to defaults when inputs are empty", () => {
  assert.equal(report("", ""), "▸ Test\n  suite     unit\n  coverage  enabled");
  assert.equal(report(undefined, undefined), "▸ Test\n  suite     unit\n  coverage  enabled");
});
