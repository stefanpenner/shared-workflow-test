# shadow/ — shadow testing

All shadow-testing logic lives here, in the **workflows** repo (`reusable-workflows`), so it changes
in one place. The **runner** repo (`reusable-workflows-shadow-testing`) is a thin `receiver.yaml`
shim that checks this code out and runs it.

## What it does

On a `shadow-test`-labelled PR, run each real **consumer**'s CI against this PR's draft, under an
authentic `pull_request` event, and report pass/fail back as a check on the PR.

```
workflows PR (labelled)
  └─ .github/workflows/shadow.yaml          (this repo, this checkout)
     ├─ list-consumers        → matrix from .github/shadow-consumers.json
     └─ dispatch-and-watch    → dispatch the runner's receiver, watch it, report a job summary
                                   │ workflow_dispatch (ref: main, workflows_ref = PR head)
                                   ▼
        runner receiver.yaml (shim) → checkout workflows@ref → `bash shadow/run-mirror.sh`
           └─ mirror-and-test  → mirror consumer onto a shadow branch, repoint `uses:` at the
                                  draft, open a shadow PR in the runner; that PR's pull_request
                                  run IS the consumer's real CI (watched, becomes the result)
on PR close → shadow-cleanup.yaml → cleanup (tear down the shadow PRs/branches)
```

## Conventions

- TypeScript on **Node 24** (type-stripping; no `tsx`/transpile). `node:test` + `node:assert`.
- `src/core/*` pure (tested) · `src/adapters/*` I/O · `src/bin/*` entrypoints.
- The provider-invoked bins (`list-consumers`, `dispatch-and-watch`, `cleanup`) are **dependency-free**
  (raw Node 24, no `npm ci`). Only `mirror-and-test` uses `yaml`, and it runs in the runner's
  receiver where `bash run-mirror.sh` does the `npm ci`.

## Develop

```sh
cd shadow
npm ci
bash ci.sh   # typecheck + lint + node --test  (what provider CI runs)
```
