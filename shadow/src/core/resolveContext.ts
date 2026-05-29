/**
 * Decide which workflows PR + head SHA to shadow-test. On `pull_request` both come from the event;
 * on `workflow_dispatch` only the PR number is given, so the head SHA is looked up (injected, so
 * this stays pure/testable). Empty strings count as absent (GHA passes empty flags for the inputs
 * that don't apply to the current event).
 */
export async function resolveContext(opts: {
  eventName: string;
  prNumber?: string;
  headSha?: string;
  inputPr?: string;
  lookupHeadSha: (pr: string) => Promise<string>;
}): Promise<{ pr: string; sha: string }> {
  if (opts.eventName === 'pull_request') {
    if (!opts.prNumber || !opts.headSha) {
      throw new Error('pull_request needs --pr-number and --head-sha');
    }
    return { pr: opts.prNumber, sha: opts.headSha };
  }
  if (!opts.inputPr) throw new Error('workflow_dispatch needs --input-pr');
  return { pr: opts.inputPr, sha: await opts.lookupHeadSha(opts.inputPr) };
}
