import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { renderShadowSummary, renderShadowLog } from '../src/core/summary.mts';

const base = {
  consumerRepo: 'o/consumer',
  consumerRef: 'main',
  workflowsRepo: 'o/workflows',
  workflowsRef: 'abc1234567',
  workflowsPr: 7,
  runUrl: 'https://example.com/run',
  prUrl: 'https://example.com/pr',
} as const;

describe('renderShadowSummary', () => {
  it('renders a passing summary with linked repos, PR, run, and shadow PR', () => {
    const md = renderShadowSummary({ ...base, result: 'passed' });
    assert.match(md, /## ✅ Shadow test passed — \[`o\/consumer`\]\(https:\/\/github\.com\/o\/consumer\)/);
    assert.match(md, /\[`o\/workflows`\]\(https:\/\/github\.com\/o\/workflows\)/);
    assert.match(md, /\[PR #7\]\(https:\/\/github\.com\/o\/workflows\/pull\/7\)/);
    assert.match(md, /`abc1234`/); // short SHA
    assert.match(md, /🏃 Runner run: https:\/\/example\.com\/run/);
    assert.match(md, /🔀 Shadow PR \(consumer CI\): https:\/\/example\.com\/pr/);
    assert.doesNotMatch(md, /Failed/);
  });

  it('renders a failing summary with a pointer to the failure', () => {
    const md = renderShadowSummary({ ...base, result: 'failed' });
    assert.match(md, /## ❌ Shadow test failed/);
    assert.match(md, /❌ \*\*Failed\*\* — open the runner run/);
  });

  it('omits the shadow PR link when none exists', () => {
    const md = renderShadowSummary({ ...base, result: 'passed', prUrl: null });
    assert.doesNotMatch(md, /Shadow PR/);
    assert.match(md, /Runner run/);
  });
});

describe('renderShadowLog', () => {
  it('returns plain-text lines (no markdown) with clickable URLs', () => {
    const lines = renderShadowLog({ ...base, result: 'passed' });
    const text = lines.join('\n');
    assert.match(text, /✅ Shadow test passed: o\/consumer@main/);
    assert.match(text, /runner run: https:\/\/example\.com\/run/);
    assert.match(text, /shadow PR:  https:\/\/example\.com\/pr/);
    assert.doesNotMatch(text, /\]\(|^#{1,6} |^\|/m); // no markdown links/headings/tables
  });

  it('omits the shadow PR line when none exists', () => {
    const lines = renderShadowLog({ ...base, result: 'failed', prUrl: null });
    assert.ok(lines.some((l) => l.includes('❌ Shadow test failed')));
    assert.ok(!lines.some((l) => l.includes('shadow PR')));
  });
});
