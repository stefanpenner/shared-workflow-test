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
