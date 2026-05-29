# CLAUDE.md

Conventions for this repo (a GitHub reusable-workflow + composite-action provider).
Follow them exactly.

## Non-negotiable rules

1. **No inline scripts in YAML.** Every action/workflow `run:` is a single invocation of
   an external file — `node ${{ github.action_path }}/scripts/<x>.cli.mjs`. No `run: |`
   logic; no one-liners with shell operators (`&&`, `||`, `;`, `|`, `>`, `$(...)`). This
   is enforced by `scripts/lib/guard/check-no-inline-scripts`, run in `test.yaml`. The
   only whitelisted inline step is the pre-checkout bootstrap in `shared.yaml` (nothing is
   on disk yet to call) — whitelisted by exact step name.
2. **Everything is Node + tested, zero `node_modules`.** Scripts are ESM `.mjs`, run on
   **Node 24**, tested with `node:test` + `node:assert/strict`. No third-party deps.
   CI gates coverage (lines/functions/branches) — keep it green.

## Layout (per action)

- `actions/<name>/scripts/<name>.mjs` — **pure** logic: no side effects on import, no
  `process.env` reads. Imported by the test.
- `actions/<name>/scripts/<name>.cli.mjs` — thin entry the action invokes: reads env, does
  the real I/O, calls the pure module.
- `actions/<name>/scripts/<name>.test.mjs` — `node:test` over the pure module.
- Shared tooling lives in `scripts/lib/**` (guard, formatters), each with a sibling test.

## Style

- Pure functions take every input as an argument and return a value; inject the I/O sink
  (e.g. the `$GITHUB_OUTPUT` path) so tests point it at a temp file.
- **Errors: no silent failures.** Catch only the error you expect; rethrow the rest.
  Attribute failures and chain the original with `new Error('context', { cause: err })`.
- A `.cli.mjs` is the only place that reads env / writes files; keep it tiny.

## Shadow testing (`shadow/`)

All shadow-testing logic lives in `shadow/` in **this** repo (one place to change it). It's a
small TypeScript project run on raw Node 24; it is the **one npm+`node_modules` island** (it
depends on `yaml` for comment-preserving workflow edits), kept entirely inside `shadow/` so the
actions above stay zero-dep. Its CI (typecheck + lint + `node:test`) runs via `bash shadow/ci.sh`.

Terminology (used consistently in code, env vars, and docs):

- **workflows** = this repo (`reusable-workflows`) — its PRs are what we test (`WORKFLOWS_REPO`/`_REF`/`_PR`).
- **runner** = `reusable-workflows-shadow-testing` — the venue where shadow PRs run a consumer's CI; it's a thin `receiver.yaml` shim that checks out this repo and runs `shadow/`.
- **consumer** = a downstream repo (from `.github/shadow-consumers.json`) we mirror to verify the draft doesn't break it.

Provider-invoked bins (`list-consumers`, `dispatch-and-watch`, `cleanup`) are dependency-free and
run on raw Node 24; only `mirror-and-test` (in the runner's receiver) uses `yaml`. Don't pull a dep
into a dep-free bin's import graph. Result is reported as a **job summary** on the shadow check (no
PR comments); the dispatch job watches **silently** — no per-step polling logs, just links + a clear
pass/fail (failures emit a red `::error::` annotation linking to where to look).

## Run locally (what CI runs)

```sh
node scripts/lib/guard/check-no-inline-scripts.cli.mjs
node --test --experimental-test-coverage \
  --test-coverage-lines=100 --test-coverage-functions=100 --test-coverage-branches=95 \
  '--test-coverage-include=actions/**/*.mjs' '--test-coverage-include=scripts/**/*.mjs' \
  '--test-coverage-exclude=**/*.test.mjs' '--test-coverage-exclude=**/*.cli.mjs' \
  'actions/**/*.test.mjs' 'scripts/**/*.test.mjs'
```
