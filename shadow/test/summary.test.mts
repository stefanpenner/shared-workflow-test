import { describe, it } from "node:test";
import assert from "node:assert/strict";
import {
  renderShadowSummary,
  renderShadowLog,
  workflowsPrUrl,
  commitUrl,
} from "../src/core/summary.mts";

const base = {
  consumerRepo: "o/consumer",
  consumerRef: "main",
  workflowsRepo: "o/workflows",
  workflowsRef: "abc1234567",
  workflowsPr: 7,
  runUrl: "https://example.com/run",
  prUrl: "https://example.com/pr",
} as const;

describe("renderShadowSummary", () => {
  it("renders a passing result as a markdown table with links", () => {
    const md = renderShadowSummary({ ...base, result: "passed" });
    assert.match(md, /## ✅ Shadow test passed/);
    assert.match(md, /^\| --- \| --- \|$/m); // it's a table
    assert.match(md, /\| Result \| ✅ passed \|/);
    assert.match(
      md,
      /\| Consumer \| \[`o\/consumer`\]\(https:\/\/github\.com\/o\/consumer\) `@main` \|/,
    );
    assert.match(md, /\[PR #7\]\(https:\/\/github\.com\/o\/workflows\/pull\/7\)/);
    assert.match(md, /\[`abc1234`\]\(https:\/\/github\.com\/o\/workflows\/commit\/abc1234567\)/);
    assert.match(md, /\| Shadow PR \| \[consumer CI\]\(https:\/\/example\.com\/pr\) \|/);
  });

  it("renders a failing result", () => {
    const md = renderShadowSummary({ ...base, result: "failed" });
    assert.match(md, /## ❌ Shadow test failed/);
    assert.match(md, /\| Result \| ❌ failed \|/);
  });

  it("omits the Shadow PR row when none exists", () => {
    const md = renderShadowSummary({ ...base, result: "passed", prUrl: null });
    assert.doesNotMatch(md, /Shadow PR/);
    assert.match(md, /Runner run/);
  });
});

describe("renderShadowLog", () => {
  it("returns plain-text lines (no markdown) with clickable URLs", () => {
    const text = renderShadowLog({ ...base, result: "passed" }).join("\n");
    assert.match(text, /✅ Shadow test passed: o\/consumer@main/);
    assert.match(text, /runner run: https:\/\/example\.com\/run/);
    assert.match(text, /shadow PR:  https:\/\/example\.com\/pr/);
    assert.doesNotMatch(text, /\]\(|^#{1,6} |^\|/m); // no markdown links/headings/tables
  });

  it("omits the shadow PR line when none exists", () => {
    const lines = renderShadowLog({ ...base, result: "failed", prUrl: null });
    assert.ok(lines.some((l) => l.includes("❌ Shadow test failed")));
    assert.ok(!lines.some((l) => l.includes("shadow PR")));
  });
});

describe("url builders", () => {
  it("builds PR and commit URLs", () => {
    assert.equal(workflowsPrUrl("o/w", 7), "https://github.com/o/w/pull/7");
    assert.equal(commitUrl("o/w", "abc"), "https://github.com/o/w/commit/abc");
  });
});
