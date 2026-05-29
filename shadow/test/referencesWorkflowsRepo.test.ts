import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { referencesWorkflowsRepo } from '../src/core/patchConsumerWorkflow.ts';

const WORKFLOWS = 'stefanpenner-cs/reusable-workflows';

describe('referencesWorkflowsRepo', () => {
  it('is true when a job calls the workflows as a reusable workflow', () => {
    const yaml = [
      'jobs:',
      '  ci:',
      '    uses: stefanpenner-cs/reusable-workflows/.github/workflows/shared.yaml@main',
    ].join('\n');
    assert.equal(referencesWorkflowsRepo(yaml, WORKFLOWS), true);
  });

  it('is false when no job calls the workflows', () => {
    const yaml = [
      'jobs:',
      '  build:',
      '    runs-on: ubuntu-latest',
      '    steps:',
      '      - uses: actions/checkout@v4',
    ].join('\n');
    assert.equal(referencesWorkflowsRepo(yaml, WORKFLOWS), false);
  });

  it('is false for a different workflows', () => {
    const yaml = ['jobs:', '  ci:', '    uses: someorg/other/.github/workflows/x.yaml@main'].join('\n');
    assert.equal(referencesWorkflowsRepo(yaml, WORKFLOWS), false);
  });
});
