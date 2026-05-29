# reusable-workflows

A reusable GitHub Actions workflow that ships its own composite actions and scripts.

## TL;DR

Reusable workflows (`workflow_call`) are interpreted **server-side** by GitHub — the provider repo's files are never cloned to the runner. To use scripts or composite actions that live in the same repo, the workflow must check itself out first:

```yaml
steps:
  - name: Set up shared actions (exclude from git)
    run: mkdir -p .git/info && echo '.github/_shared-workflow/' >> .git/info/exclude

  - name: Set up shared actions (checkout)
    uses: actions/checkout@v4
    with:
      repository: stefanpenner-cs/reusable-workflows
      ref: ${{ inputs.ref || 'main' }}
      path: .github/_shared-workflow

  - uses: ./.github/_shared-workflow/actions/setup
  - uses: ./.github/_shared-workflow/actions/lint
  - uses: ./.github/_shared-workflow/actions/test
```

- The workflow takes an explicit **`ref` input** so the caller controls which version of the actions is fetched. `github.job_workflow_sha` is empty in some contexts (e.g. `workflow_dispatch` self-tests), so an explicit input is more reliable.
- Checking out into `.github/_shared-workflow` and adding it to `.git/info/exclude` keeps the fetched actions out of the consumer's working tree.
- The `./` prefix on action paths is **required** — without it GHA interprets the path as `org/repo@ref`.

## Structure

```
actions/
  setup/          # Set up the project environment
  lint/           # Run linting checks
  test/           # Run the test suite
  debug/          # Print file tree + git status (diagnostics)
```

Each action is self-contained: `action.yaml` defines inputs/outputs and delegates to `scripts/run.sh` via `${{ github.action_path }}` (the `debug` action is inline).

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

The default `GITHUB_TOKEN` is scoped to the caller repo. For private provider repos, pass a token with `contents: read` access:

```yaml
- uses: actions/checkout@v4
  with:
    repository: stefanpenner-cs/reusable-workflows
    ref: ${{ inputs.ref || 'main' }}
    token: ${{ secrets.PROVIDER_REPO_TOKEN }}
    path: .github/_shared-workflow
```

Alternatively, if both repos are in the same org, enable "Accessible from repositories in the organization" in the provider repo's Actions settings — then the caller's `GITHUB_TOKEN` works.

## See also

- [reusable-workflows-consumer](https://github.com/stefanpenner-cs/reusable-workflows-consumer) — example consumer repo
