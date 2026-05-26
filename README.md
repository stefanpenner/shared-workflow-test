# shared-workflow-test

A reusable GitHub Actions workflow that ships its own shell scripts.

## TL;DR

Reusable workflows (`workflow_call`) are interpreted **server-side** by GitHub — the provider repo's files are never cloned to the runner. So if your shared workflow needs to run a script, it must check out its own repo first:

```yaml
steps:
  - uses: actions/checkout@v4
    with:
      repository: stefanpenner/shared-workflow-test
      ref: ${{ github.job_workflow_sha }}
      path: _self
      sparse-checkout: scripts
      sparse-checkout-cone-mode: true
  - run: _self/scripts/hello.sh
```

`github.job_workflow_sha` ensures the checkout matches the exact ref the consumer pinned (e.g. `@main`, `@v1`, `@abc123`).

## Usage

From any other repo:

```yaml
jobs:
  example:
    uses: stefanpenner/shared-workflow-test/.github/workflows/shared.yml@main
```

## See also

- [shared-workflow-consumer](https://github.com/stefanpenner/shared-workflow-consumer) — example consumer repo
