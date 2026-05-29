import { shadowBranchName } from "./shadowBranchName.mts";
import type { Consumer } from "./parseConsumers.mts";

export type ShadowResult = "passed" | "failed";

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

export const workflowsPrUrl = (repo: string, pr: number | string): string =>
  `https://github.com/${repo}/pull/${pr}`;
export const commitUrl = (repo: string, sha: string): string =>
  `https://github.com/${repo}/commit/${sha}`;
const repoLink = (repo: string): string => `[\`${repo}\`](https://github.com/${repo})`;
const link = (label: string, url: string): string => `[${label}](${url})`;

/**
 * Render the shadow result as a clean markdown **table** for the job-summary page (the shadow
 * check's artifact). Pure: no I/O, fully testable.
 */
export function renderShadowSummary(input: ShadowSummaryInput): string {
  const passed = input.result === "passed";
  const icon = passed ? "✅" : "❌";

  const rows: Array<[string, string]> = [
    ["Result", `${icon} ${passed ? "passed" : "failed"}`],
    ["Consumer", `${repoLink(input.consumerRepo)} \`@${input.consumerRef}\``],
    [
      "Draft",
      `${repoLink(input.workflowsRepo)} · ${link(`PR #${input.workflowsPr}`, workflowsPrUrl(input.workflowsRepo, input.workflowsPr))} · ${link(`\`${input.workflowsRef.slice(0, 7)}\``, commitUrl(input.workflowsRepo, input.workflowsRef))}`,
    ],
    ["Runner run", link("logs", input.runUrl)],
  ];
  if (input.prUrl) rows.push(["Shadow PR", link("consumer CI", input.prUrl)]);

  return [
    `## ${icon} Shadow test ${passed ? "passed" : "failed"}`,
    "",
    "| | |",
    "| --- | --- |",
    ...rows.map(([k, v]) => `| ${k} | ${v} |`),
    "",
  ].join("\n");
}

/**
 * Render the result as plain-text log lines (GitHub job logs don't render markdown). Clickable URLs,
 * no markup. The table version above is for the job-summary page; this is for the step log.
 */
export function renderShadowLog(input: ShadowSummaryInput): string[] {
  const icon = input.result === "passed" ? "✅" : "❌";
  const lines = [
    `${icon} Shadow test ${input.result}: ${input.consumerRepo}@${input.consumerRef}`,
    `   vs ${input.workflowsRepo} PR #${input.workflowsPr} (${input.workflowsRef.slice(0, 7)})`,
    `   runner run: ${input.runUrl}`,
  ];
  if (input.prUrl) lines.push(`   shadow PR:  ${input.prUrl}`);
  return lines;
}

/**
 * Render the up-front index of all shadow tests for the `prepare` summary: one row per consumer with
 * its repo and a link to its shadow PR in the runner (which holds the consumer's CI run). Built from
 * the deterministic branch name, so it's knowable before any consumer dispatches. Pass/fail lives on
 * the per-consumer checks, not here. Pure.
 */
export function renderShadowList(input: {
  consumers: Consumer[];
  workflowsPr: number;
  runnerRepo: string;
}): string {
  const rows = input.consumers.map(({ repo, ref }) => {
    const branch = shadowBranchName({ prNumber: input.workflowsPr, consumerRepo: repo });
    const shadowPr = `https://github.com/${input.runnerRepo}/pulls?q=${encodeURIComponent(`is:pr head:${branch}`)}`;
    return `| ${repoLink(repo)} \`@${ref}\` | ${link("shadow PR + run", shadowPr)} |`;
  });
  return [
    "## 🛰️ Shadow tests",
    "",
    "_Pass/fail is on the `Shadow: …` checks above. Each shadow PR runs the consumer’s CI._",
    "",
    "| Consumer | Runner |",
    "| --- | --- |",
    ...rows,
    "",
  ].join("\n");
}
