# reusable-workflows

A reusable GitHub Actions workflow that ships its own composite actions, written in **Go** and built
on the runner with **Bazel** (via [bazelisk](https://github.com/bazelbuild/bazelisk), pinned by
`.bazelversion`).

## Why — the GitHub gaps

Reusable workflows (`workflow_call`) are interpreted **server-side**: the provider repo's files —
its scripts and composite actions — **never reach the runner**. GitHub offers no native "use the
calling workflow's own repo files," so a workflow that wants to run its own scripts must put them on
the runner itself. Two constraints shape how:

- **`uses:` can't be an expression** — and the provider repo isn't on disk to host a local
  bootstrap action — so the bootstrap must be an _external_ action
  ([`checkout-anywhere`](https://github.com/stefanpenner/checkout-anywhere)).
- **A local `uses:` must start with `./`**, even when it traverses out with `..`; a bare leading
  `..` is rejected by the workflow parser.

## How

The workflow checks itself out onto the runner **outside the workspace**, then references its
actions by path:

```yaml
steps:
  - name: Set up shared actions
    uses: stefanpenner/checkout-anywhere@v1
    with:
      repository: stefanpenner-cs/reusable-workflows
      ref: ${{ inputs.ref || 'main' }} # branch / tag / SHA
      path: ../_reusable-workflows # outside $GITHUB_WORKSPACE

  - uses: bazel-contrib/setup-bazel@0.15.0 # the actions are Go, built on demand by Bazel

  - uses: ./../_reusable-workflows/actions/setup
  - uses: ./../_reusable-workflows/actions/lint
  - uses: ./../_reusable-workflows/actions/test
```

Checking out to `../_reusable-workflows` (outside `$GITHUB_WORKSPACE`) keeps the fetched files out of
the consumer's `git status` — no `.git/info/exclude` needed. Each action's `run:` is a single
`bazelisk run //actions/<x>` (with `working-directory` set to the checkout, so Bazel finds the
provider's `MODULE.bazel`/`.bazelversion`); `setup-bazel` installs + caches bazelisk so the first
action pays a warm build.

### How it fits together

Four repos — three under `stefanpenner-cs`, plus `stefanpenner/checkout-anywhere`:

```
reusable-workflows                  this repo — the reusable workflow + its Go composite actions + shadow/
stefanpenner/checkout-anywhere      checks out a repo@ref into any path (the bootstrap)
reusable-workflows-shadow-testing   "runner" — isolated venue where shadow PRs run a consumer's CI
reusable-workflows-consumer         an example consumer

Consume ─ any repo calls the reusable workflow:
  consumer/ci.yaml ──uses──▶ reusable-workflows/.github/workflows/shared.yaml@<ref>
      └ shared.yaml ──uses checkout-anywhere@v1──▶ ../_reusable-workflows   (this repo @ref, OUTSIDE the workspace)
                    └ uses ./../_reusable-workflows/actions/{setup,lint,test,debug}

Shadow-test ─ label this repo's PR `shadow-test`, run each consumer against the draft:
  shadow.yaml ──dispatch──▶ runner: receiver.yaml ──▶ opens a shadow PR per consumer
      └ each shadow PR runs the consumer's real CI vs the PR draft  →  PR check "Shadow: <consumer>"
```

## What

Composite actions, each self-contained (`action.yaml` + a Go `go_binary` under `actions/<x>`):

| action  | does                                       |
| ------- | ------------------------------------------ |
| `setup` | set up the project environment             |
| `lint`  | run linting checks                         |
| `test`  | run the test suite                         |
| `debug` | print file tree + git status (diagnostics) |

The action bodies are intentionally **scaffolds** — they echo their inputs rather than do real
work. This project's focus is the **glue** (getting a provider's own code onto the runner, above)
and the **testing story**: the pure logic in `internal/` is unit-tested (testify) under a
line-coverage gate, and `shadow/` integration-tests changes against real consumers before merge. To
make an action do real work, put the logic in its `internal/actions/<x>` package.

Use it from any repo:

```yaml
jobs:
  ci:
    uses: stefanpenner-cs/reusable-workflows/.github/workflows/shared.yaml@main
    with:
      ref: main # required: which version of the actions to fetch
      project-name: my-app # optional
```

`shadow/` pre-merge-tests this repo's own changes against real consumers — see
[`shadow/README.md`](shadow/README.md). Repo conventions (no inline scripts, Go + Bazel, lint,
CLI args, TDD) live in [`CLAUDE.md`](CLAUDE.md).

## Caveats

- **Pass an explicit `ref`.** `github.job_workflow_sha` is empty in some contexts (e.g.
  `workflow_dispatch` self-tests), so the caller pins the version via the `ref` input.
- **Keep the `./` prefix** on action paths — without it GHA reads the path as `org/repo@ref`.
- **Private provider repo:** give `checkout-anywhere` a token with `contents: read`
  (`token: ${{ secrets.PROVIDER_REPO_TOKEN }}`), or enable org-wide Actions access so the caller's
  `GITHUB_TOKEN` works.
- The checked-out actions live **outside the workspace**, so they never show up in the consumer's
  working tree.

## See also

- [reusable-workflows-consumer](https://github.com/stefanpenner-cs/reusable-workflows-consumer) — example consumer
