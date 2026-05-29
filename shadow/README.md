# shadow/ — shadow testing

**Goal:** before merging a PR to this reusable-workflows repo, prove the change doesn't break the
real repos that use it.

**How:** for each consumer, copy its repo onto a throwaway branch, repoint its `uses:` at this PR's
commit, open a PR with it, and let that PR's CI run for real. Green = safe. The result shows up as
a `Shadow: <consumer>` check on your PR.

```
your PR (label it `shadow-test`)
        │
        ▼
shadow.yaml ──► for each consumer ──► dispatch the runner's receiver to:
                                        copy consumer → repoint at your PR → open a PR → run its CI
        ▲                                                                                   │
        └──────────────── ✅/❌ shows as a check on your PR ◄───────────────────────────────┘
```

Three repos: **workflows** (this repo, under test) · **runner**
(`reusable-workflows-shadow-testing`, where the throwaway PRs run) · **consumer** (a real user repo,
listed in `.github/shadow-consumers.json`).

## Code

All shadow logic is Go, built with Bazel (same toolchain as the rest of the repo — see the top-level
[`CLAUDE.md`](../CLAUDE.md)):

- `shadow/cmd/*` — thin cobra entrypoints (`//shadow/cmd/<name>`), invoked from the workflows as
  `bazelisk run //shadow/cmd/<name> -- --flags`: `resolve-ctx`, `list-consumers`,
  `dispatch-and-watch` (workflows side) and `mirror-and-test` (runner side) + `cleanup`.
- `internal/shadow/core` — pure, unit-tested logic: branch naming, consumer parsing, dispatch
  classification, summary rendering, and the comment-preserving workflow transforms (`yaml.v3`).
- `internal/shadow/adapters` — the I/O: process exec, git (reproducible commits), the GitHub API
  (**google/go-github**, token from `SHADOW_PAT`/`GH_TOKEN`), and on-disk workflow patching.

## Why it's this convoluted (GitHub gaps)

Every awkward part works around something GitHub doesn't offer:

- **No "test my unreleased workflow against a consumer."** → we mirror the consumer and rewrite its
  `uses:` to this PR's SHA.
- **CI only runs on real events.** → we actually open a PR (in the runner repo) so a genuine
  `pull_request` run happens.
- **It needs its own repo (the runner).** Those throwaway PRs would be noise here, and a workflow
  can't open PRs against arbitrary repos cleanly. → an isolated venue.
- **`workflow_dispatch` won't take a SHA** (branch/tag only) — so the runner shim is dispatched on
  `main` and the SHA is passed as an input.
- **Cross-repo writes need a PAT** (`SHADOW_PAT`) with the `workflow` scope to push workflow files.

## Develop

```sh
bazelisk test //internal/shadow/... //shadow/...   # from the repo root
```

`internal/shadow/core/*` is pure (tested); `internal/shadow/adapters/*` does I/O (excluded from the
coverage gate); `shadow/cmd/*` are entrypoints (flags + env sinks, no logic).
