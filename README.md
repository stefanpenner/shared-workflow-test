# shared-workflow-test

A reusable GitHub Actions workflow that ships its own composite actions and scripts.

## TL;DR

Reusable workflows (`workflow_call`) are interpreted **server-side** by GitHub — the provider repo's files are never cloned to the runner. To use scripts or composite actions that live in the same repo, the workflow must check itself out first:

```yaml
steps:
  - uses: actions/checkout@v4
    with:
      repository: stefanpenner/shared-workflow-test
      ref: ${{ github.job_workflow_sha }}
      path: _self
      # Optional: only fetch the actions/ directory
      sparse-checkout: actions
      sparse-checkout-cone-mode: true
  - uses: ./_self/actions/setup
  - uses: ./_self/actions/lint
  - uses: ./_self/actions/test
```

- `github.job_workflow_sha` ensures the checkout matches the exact ref the consumer pinned.
- The `./` prefix on action paths is **required** — without it GHA interprets the path as `org/repo@ref`.
- Sparse checkout is optional but keeps the clone minimal.

## Structure

```
actions/
  setup/          # Set up the project environment
    action.yaml
    scripts/run.sh
  lint/           # Run linting checks
    action.yaml
    scripts/run.sh
  test/           # Run the test suite
    action.yaml
    scripts/run.sh
```

Each action is self-contained: `action.yaml` defines inputs/outputs and delegates to `scripts/run.sh` via `${{ github.action_path }}`.

## Usage

From any other repo:

```yaml
jobs:
  ci:
    uses: stefanpenner/shared-workflow-test/.github/workflows/shared.yaml@main
    with:
      project-name: my-app
```

## Private repos

The default `GITHUB_TOKEN` is scoped to the caller repo. For private provider repos, pass a token with `contents: read` access:

```yaml
- uses: actions/checkout@v4
  with:
    repository: stefanpenner/shared-workflow-test
    ref: ${{ github.job_workflow_sha }}
    token: ${{ secrets.PROVIDER_REPO_TOKEN }}
    path: _self
```

Alternatively, if both repos are in the same org, enable "Accessible from repositories in the organization" in the provider repo's Actions settings — then the caller's `GITHUB_TOKEN` works.

## See also

- [shared-workflow-consumer](https://github.com/stefanpenner/shared-workflow-consumer) — example consumer repo
