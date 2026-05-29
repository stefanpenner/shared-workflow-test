# reusable-workflows

A reusable GitHub Actions workflow that ships its own composite actions and scripts.

## TL;DR

Reusable workflows (`workflow_call`) are interpreted **server-side** by GitHub — the provider repo's files are never cloned to the runner. To use scripts or composite actions that live in the same repo, the workflow must get them onto the runner first. We clone them **outside the workspace** and reference them via traversal `uses:`:

```yaml
steps:
  - name: Set up shared actions
    uses: stefanpenner-cs/clone-action@v1
    with:
      repository: stefanpenner-cs/reusable-workflows
      ref: ${{ inputs.ref || 'main' }}   # branch / tag / SHA
      path: ../_reusable-workflows          # outside $GITHUB_WORKSPACE

  - uses: ./../_reusable-workflows/actions/setup
  - uses: ./../_reusable-workflows/actions/lint
  - uses: ./../_reusable-workflows/actions/test
```

- The workflow takes an explicit **`ref` input** so the caller controls which version of the actions is fetched. `github.job_workflow_sha` is empty in some contexts (e.g. `workflow_dispatch` self-tests), so an explicit input is more reliable.
- [`clone-action`](https://github.com/stefanpenner-cs/clone-action) clones to **`../_reusable-workflows`** (outside the workspace), so the fetched actions never appear in the consumer's `git status` — no `.git/info/exclude` needed. (It's a separate action because a `uses:` ref can't be an expression, and the reusable workflow's own repo isn't on disk to host a local one.)
- `uses: ./../_reusable-workflows/...` works — a local `uses:` may traverse out of the workspace with `..`, **as long as it starts with `./`** (a bare leading `..` is rejected by the workflow parser).

## How it fits together

Four repos under `stefanpenner-cs`:

```
reusable-workflows                  this repo — the reusable workflow + its composite actions + shadow/
clone-action                        clones a repo@ref into any path (the bootstrap)
reusable-workflows-shadow-testing   "runner" — isolated venue where shadow PRs run a consumer's CI
reusable-workflows-consumer         an example consumer

Consume ─ any repo calls the reusable workflow:
  consumer/ci.yaml ──uses──▶ reusable-workflows/.github/workflows/shared.yaml@<ref>
      └ shared.yaml ──uses clone-action@v1──▶ ../_reusable-workflows   (this repo @ref, OUTSIDE the workspace)
                    └ uses ./../_reusable-workflows/actions/{setup,lint,test,debug}

Shadow-test ─ label this repo's PR `shadow-test`, run each consumer against the draft:
  shadow.yaml ──dispatch──▶ runner: receiver.yaml ──▶ opens a shadow PR per consumer
      └ each shadow PR runs the consumer's real CI vs the PR draft  →  PR check "Shadow: <consumer>"
```

## Structure

```
actions/
  setup/   # Set up the project environment
  lint/    # Run linting checks
  test/    # Run the test suite
  debug/   # Print file tree + git status (diagnostics)
scripts/
  lib/guard/   # check-no-inline-scripts: enforces the "no inline scripts" rule
  lib/log/     # shared output formatting
shadow/        # pre-merge testing against real consumers — see shadow/README.md
```

Each action is self-contained: `action.yaml` defines inputs/outputs and invokes an
external Node script via `node ${{ github.action_path }}/scripts/<name>.cli.mjs`.

## Scripts &amp; testing conventions

All executable logic lives in **external scripts**, never in inline `run:` blocks
(see the guard below). Each script follows a three-file pattern:

- `<name>.mjs` — pure logic, no side effects on import, no `process.env` reads. Imported by the test.
- `<name>.cli.mjs` — thin entry the action invokes; reads env and does the real I/O.
- `<name>.test.mjs` — [`node:test`](https://nodejs.org/api/test.html) + `node:assert`, **zero `node_modules`**.

One harness covers the whole repo — the actions and shared scripts (`.mjs`) plus
[`shadow/`](shadow/) (`.mts`, run natively on Node 24). Reproduce CI
(`.github/workflows/test.yaml`) locally, no install needed:

```sh
node scripts/lib/guard/check-no-inline-scripts.cli.mjs   # no inline run: blocks
node shadow/src/bin/check-deps.mts                       # shadow's only dep is `yaml`
node shadow/typecheck.mjs                                # isolated tsc --noEmit
node --test 'actions/**/*.test.mjs' 'scripts/**/*.test.mjs' 'shadow/test/*.test.mts'
```

CI runs these on every push and PR and adds a coverage gate (thresholds in `test.yaml`).
There is **no** inline `run:` exception: `shared.yaml` bootstraps with
[`stefanpenner-cs/clone-action`](https://github.com/stefanpenner-cs/clone-action), which clones
this repo to `../_reusable-workflows` (outside the workspace, so nothing leaks into the consumer's
`git status`); the actions are then referenced via `uses: ./../_reusable-workflows/...`.

## Usage

From any other repo:

```yaml
jobs:
  ci:
    uses: stefanpenner-cs/reusable-workflows/.github/workflows/shared.yaml@main
    with:
      ref: main            # required: which version of the shared actions to fetch
      project-name: my-app # optional
```

To self-test within this repo, `ci.yaml` calls the workflow with `ref: ${{ github.sha }}`.

## Private repos

The default `GITHUB_TOKEN` is scoped to the caller repo. For private provider repos, pass a token with `contents: read` access to the clone action:

```yaml
- uses: stefanpenner-cs/clone-action@v1
  with:
    repository: stefanpenner-cs/reusable-workflows
    ref: ${{ inputs.ref || 'main' }}
    path: ../_reusable-workflows
    token: ${{ secrets.PROVIDER_REPO_TOKEN }}
```

Alternatively, if both repos are in the same org, enable "Accessible from repositories in the organization" in the provider repo's Actions settings — then the caller's `GITHUB_TOKEN` works.

## See also

- [reusable-workflows-consumer](https://github.com/stefanpenner-cs/reusable-workflows-consumer) — example consumer repo
