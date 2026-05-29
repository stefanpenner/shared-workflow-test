import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { resolveContext } from '../src/core/resolveContext.mts';

const noLookup = async (): Promise<string> => {
  throw new Error('should not look up');
};

describe('resolveContext', () => {
  it('uses the event PR + head SHA on pull_request', async () => {
    const ctx = await resolveContext({ eventName: 'pull_request', prNumber: '7', headSha: 'abc', lookupHeadSha: noLookup });
    assert.deepEqual(ctx, { pr: '7', sha: 'abc' });
  });

  it('looks up the head SHA on workflow_dispatch', async () => {
    const ctx = await resolveContext({ eventName: 'workflow_dispatch', inputPr: '9', lookupHeadSha: async (pr) => `sha-of-${pr}` });
    assert.deepEqual(ctx, { pr: '9', sha: 'sha-of-9' });
  });

  it('throws when a pull_request is missing its SHA', async () => {
    await assert.rejects(() => resolveContext({ eventName: 'pull_request', prNumber: '7', lookupHeadSha: noLookup }), /head-sha/);
  });

  it('throws when a dispatch is missing its PR', async () => {
    await assert.rejects(() => resolveContext({ eventName: 'workflow_dispatch', inputPr: '', lookupHeadSha: noLookup }), /input-pr/);
  });
});
