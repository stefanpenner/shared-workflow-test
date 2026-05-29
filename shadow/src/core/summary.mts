export type ShadowResult = 'passed' | 'failed';

export interface ShadowSummaryInput {
  consumerRepo: string;
  consumerRef: string;
  workflowsRepo: string;
  workflowsRef: string;
  workflowsPr: number;
  result: ShadowResult;
  runUrl: string; // the runner (receiver) run
  prUrl: string | null; // the shadow PR, where the consumer's CI actually runs
}

const repoLink = (repo: string): string => `[\`${repo}\`](https://github.com/${repo})`;

/**
 * Render the shadow result as a GitHub job-summary markdown page — the clickable artifact on the
 * shadow check (PR / repo / run links with emoji, no step-level noise). Pure: no I/O, fully testable.
 */
export function renderShadowSummary(input: ShadowSummaryInput): string {
  const passed = input.result === 'passed';
  const icon = passed ? '✅' : '❌';
  const prHref = `https://github.com/${input.workflowsRepo}/pull/${input.workflowsPr}`;

  const lines = [
    `## ${icon} Shadow test ${passed ? 'passed' : 'failed'} — ${repoLink(input.consumerRepo)}`,
    '',
    `Ran ${repoLink(input.consumerRepo)} \`@${input.consumerRef}\` against ` +
      `${repoLink(input.workflowsRepo)} [PR #${input.workflowsPr}](${prHref}) ` +
      `(\`${input.workflowsRef.slice(0, 7)}\`).`,
    '',
    `- 🏃 Runner run: ${input.runUrl}`,
  ];
  if (input.prUrl) lines.push(`- 🔀 Shadow PR (consumer CI): ${input.prUrl}`);
  if (!passed) {
    lines.push('', '> ❌ **Failed** — open the runner run above and click into the failing job to see why.');
  }
  lines.push('');
  return lines.join('\n');
}

/**
 * Render the result as plain-text log lines (GitHub job logs don't render markdown). Clickable URLs,
 * no markup. The markdown version above is for the job-summary page; this is for the step log.
 */
export function renderShadowLog(input: ShadowSummaryInput): string[] {
  const icon = input.result === 'passed' ? '✅' : '❌';
  const lines = [
    `${icon} Shadow test ${input.result}: ${input.consumerRepo}@${input.consumerRef}`,
    `   vs ${input.workflowsRepo} PR #${input.workflowsPr} (${input.workflowsRef.slice(0, 7)})`,
    `   runner run: ${input.runUrl}`,
  ];
  if (input.prUrl) lines.push(`   shadow PR:  ${input.prUrl}`);
  return lines;
}

export const workflowsPrUrl = (repo: string, pr: number | string): string => `https://github.com/${repo}/pull/${pr}`;
export const commitUrl = (repo: string, sha: string): string => `https://github.com/${repo}/commit/${sha}`;

/** Clean name for the per-consumer custom check, e.g. "Shadow: reusable-workflows-consumer". */
export function checkName(consumerRepo: string): string {
  return `Shadow: ${consumerRepo.split('/').pop()}`;
}

/** One-line check title, e.g. "✅ passed — owner/consumer". */
export function checkTitle(input: { consumerRepo: string; result: ShadowResult }): string {
  const icon = input.result === 'passed' ? '✅' : '❌';
  return `${icon} ${input.result} — ${input.consumerRepo}`;
}
