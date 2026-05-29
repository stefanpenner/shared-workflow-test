import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { requireArgs } from '../src/core/args.ts';

describe('requireArgs', () => {
  it('reads named string flags (space and = forms)', () => {
    const a = requireArgs(['workflows-repo', 'consumer-ref'], ['--workflows-repo', 'o/w', '--consumer-ref=main']);
    assert.equal(a['workflows-repo'], 'o/w');
    assert.equal(a['consumer-ref'], 'main');
  });

  it('throws naming a missing flag', () => {
    assert.throws(() => requireArgs(['workflows-repo'], []), /--workflows-repo/);
  });

  it('throws on an empty value', () => {
    assert.throws(() => requireArgs(['x'], ['--x=']), /--x/);
  });
});
