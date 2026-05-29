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

- **TDD.** Write the test first; keep logic in small pure functions so it's trivial to test.
- **Simplicity & correctness over cleverness.** ELI5-clean, readable syntax; clean models, not
  hacks. If it needs a comment to explain a trick, prefer the boring version.
- **`try/catch/finally`, not `.catch()`/`.then()`.** Linear control flow reads top-to-bottom.
- Pure functions take every input as an argument and return a value; inject the I/O sink
  (e.g. the `$GITHUB_OUTPUT` path) so tests point it at a temp file.
- **Errors: no silent failures.** Catch only the error you expect; rethrow the rest.
  Attribute failures and chain the original with `new Error('context', { cause: err })`.
- A `.cli.mjs` is the only place that reads env / writes files; keep it tiny.

## Shadow testing (`shadow/`)

All shadow-testing logic lives in `shadow/` in **this** repo (one place to change it). `.mts`
TypeScript run on **Node 24** (type-stripping at runtime — no transpile step). **`yaml` is the ONE
runtime dependency**, brought in via npm and enforced by `shadow/src/bin/check-deps.mts` (no other
deps allowed). Types are checked separately and only by the isolated `tsc` (`node shadow/typecheck.mjs`
→ `tsc --noEmit`); the runtime never runs tsc. devDeps exist solely for that typecheck.

Terminology (used consistently in code, **CLI flags**, and docs):

- **workflows** = this repo (`reusable-workflows`) — its PRs are what we test (`--workflows-repo/-ref/-pr`).
- **runner** = `reusable-workflows-shadow-testing` — the venue where shadow PRs run a consumer's CI; a thin `receiver.yaml` shim that checks out this repo and runs `shadow/`.
- **consumer** = a downstream repo (from `.github/shadow-consumers.json`) we mirror to verify the draft doesn't break it.

Conventions specific to `shadow/`:

- **Config comes from CLI flags, not env** — bins read `--workflows-repo` etc. via `requireArgs`
  (use `--flag=${{ ... }}` in YAML so empty values don't break parsing). Only **secrets**
  (`SHADOW_PAT`) and GHA sinks (`GITHUB_OUTPUT`) stay in env.
- Each shadow test posts a **custom check run** on the PR (`Shadow: <consumer>`) via the Checks
  API — created with the job's `GITHUB_TOKEN` (a PAT can't create check runs), `details_url` →
  the shadow PR, markdown summary as its Details page. It's best-effort (a fork PR's read-only
  token just skips it). The matrix job's own check coexists.
- Logs are **plain text with full URLs** (GitHub logs don't render markdown — that's reserved for
  the check's Details page / job summary).
- **`yaml` is the only runtime dependency** (enforced by `check-deps`). Don't add others.

## Run locally (what CI runs)

```sh
node scripts/lib/guard/check-no-inline-scripts.cli.mjs
node --test --experimental-test-coverage \
  --test-coverage-lines=100 --test-coverage-functions=100 --test-coverage-branches=95 \
  '--test-coverage-include=actions/**/*.mjs' '--test-coverage-include=scripts/**/*.mjs' \
  '--test-coverage-exclude=**/*.test.mjs' '--test-coverage-exclude=**/*.cli.mjs' \
  'actions/**/*.test.mjs' 'scripts/**/*.test.mjs'
```
