import { parseDocument, isMap } from "yaml";

export interface PatchOptions {
  /** `owner/repo` of the reusable-workflows workflows whose refs should be repointed. */
  workflowsRepo: string;
  /** The git ref (typically the PR head SHA) to pin the workflows to. */
  workflowsRef: string;
}

/** A job-level reusable-workflow `uses:` — `owner/repo/.github/workflows/<file>@<ref>`. */
const REUSABLE_USES = /^(?<repo>[^/]+\/[^/]+)\/(?<path>\.github\/workflows\/[^@]+)@.+$/;

/** True if any job calls `workflowsRepo` as a reusable workflow. Used to decide which of a consumer's
 * workflows to mirror-transform — so unrelated workflows aren't force-triggered on the shadow PR. */
export function referencesWorkflowsRepo(yaml: string, workflowsRepo: string): boolean {
  const doc = parseDocument(yaml);
  const jobs = doc.get("jobs");
  if (!isMap(jobs)) return false;
  for (const { value: job } of jobs.items) {
    if (!isMap(job)) continue;
    const uses = job.get("uses");
    if (typeof uses !== "string") continue;
    if (REUSABLE_USES.exec(uses)?.groups?.repo === workflowsRepo) return true;
  }
  return false;
}

/**
 * Repoint a consumer's reusable-workflow call at a specific workflows ref so the consumer's CI
 * exercises the workflows's draft state. Pure: takes YAML in, returns YAML out, preserving comments
 * and formatting. Only job-level `uses:` targeting `workflowsRepo` are touched; step-level action
 * `uses:` and other repos' workflows are left alone. Idempotent.
 */
export function patchConsumerWorkflow(yaml: string, opts: PatchOptions): string {
  const doc = parseDocument(yaml);
  const jobs = doc.get("jobs");
  if (!isMap(jobs)) return doc.toString();

  for (const { value: job } of jobs.items) {
    if (!isMap(job)) continue;

    const uses = job.get("uses");
    if (typeof uses !== "string") continue;

    const groups = REUSABLE_USES.exec(uses)?.groups;
    if (!groups || groups.repo !== opts.workflowsRepo || groups.path === undefined) continue;

    const path = groups.path.replace(/\.yml$/, ".yaml");
    job.set("uses", `${opts.workflowsRepo}/${path}@${opts.workflowsRef}`);

    const withBlock = job.get("with", true);
    if (isMap(withBlock)) {
      withBlock.set("ref", opts.workflowsRef);
    } else {
      job.set("with", doc.createNode({ ref: opts.workflowsRef }));
    }
  }

  return doc.toString();
}
