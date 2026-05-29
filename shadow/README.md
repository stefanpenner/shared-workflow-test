# shadow/ — shadow testing

**Goal:** before merging a PR to this reusable-workflows repo, prove the change doesn't break the
real repos that use it.

**How:** for each consumer, copy its repo onto a throwaway branch, repoint its `uses:` at this PR's
commit, open a PR with it, and let that PR's CI run for real. Green = safe. The result shows up as
a check on your PR.

```
your PR (label it `shadow-test`)
        │
        ▼
shadow.yaml ──► for each consumer ──► tell the runner repo to:
                                        copy consumer → repoint at your PR → open a PR → run its CI
        ▲                                                                                   │
        └──────────────── ✅/❌ shows as a check on your PR ◄───────────────────────────────┘
```

Three repos: **workflows** (this repo, under test) · **runner**
(`reusable-workflows-shadow-testing`, where the throwaway PRs run) · **consumer** (a real user repo,
listed in `.github/shadow-consumers.json`).

## Why it's this convoluted (GitHub gaps)

Every awkward part here exists to work around something GitHub doesn't offer:

- **No "test my unreleased workflow against a consumer."** Consumers pin `@main`/a tag; there's no
  native way to say "run consumer X using *this* draft." → we mirror the consumer and rewrite its
  `uses:` to this PR's SHA.
- **CI only runs on real events.** You can't ask GitHub to "run repo X's CI as if a PR existed." →
  we actually open a PR (in the runner repo) so a genuine `pull_request` run happens.
- **It needs its own repo (the runner).** Those throwaway PRs would be noise in the workflows or
  consumer repos, and a workflow can't open PRs against arbitrary repos cleanly. → an isolated venue.
- **`workflow_dispatch` won't take a SHA** (branch/tag only) — so the runner shim is dispatched on
  `main` and the SHA is passed as an input.
- **Cross-repo writes need a PAT** (`SHADOW_PAT`), and pushing workflow files needs that token to
  have the **`workflow`** scope, or GitHub rejects the push.

## Develop

Everything runs on raw Node 24 — no install:

```sh
node --test 'shadow/test/*.test.mts'   # from the repo root; part of the one repo-wide harness
```

`src/core/*` is pure (tested); `src/adapters/*` does I/O; `src/bin/*` are entrypoints (read CLI
flags, no logic). `vendor/yaml/` is the one vendored dependency.
