import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { renderShadowList } from '../src/core/summary.mts';

describe('renderShadowList', () => {
  it('lists each consumer with a repo link and a deterministic shadow-PR link', () => {
    const md = renderShadowList({
      consumers: [
        { repo: 'o/consumer-a', ref: 'main' },
        { repo: 'o/consumer-b', ref: 'dev' },
      ],
      workflowsPr: 2,
      runnerRepo: 'o/runner',
    });
    assert.match(md, /## 🛰️ Shadow tests/);
    assert.match(md, /\[`o\/consumer-a`\]\(https:\/\/github\.com\/o\/consumer-a\) `@main`/);
    assert.match(md, /`@dev`/);
    // shadow-PR link is a runner PR search for the deterministic head branch
    assert.match(md, /github\.com\/o\/runner\/pulls\?q=/);
    assert.match(md, /pr-2-o-consumer-a/); // branch slug (unencoded chars survive)
  });

  it('handles an empty consumer list', () => {
    const md = renderShadowList({ consumers: [], workflowsPr: 1, runnerRepo: 'o/runner' });
    assert.match(md, /## 🛰️ Shadow tests/);
    assert.doesNotMatch(md, /github\.com\/o\/runner\/pulls/);
  });
});
