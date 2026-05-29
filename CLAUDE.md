# CLAUDE.md

Conventions for this repo (a GitHub reusable-workflow + composite-action provider).
Follow them exactly.

## Non-negotiable rules

1. **No inline scripts in YAML.** Every action/workflow `run:` is a single invocation of
   an external file — `node ${{ github.action_path }}/scripts/<x>.cli.mjs`. No `run: |`
   logic; no one-liners with shell operators (`&&`, `||`, `;`, `|`, `>`, `$(...)`). This
   is enforced by `scripts/lib/guard/check-no-inline-scripts`, run in `test.yaml`. There is
   **no** inline exception: `shared.yaml` bootstraps via `stefanpenner-cs/clone-action`
   (clones this repo to `../_reusable-workflows`, referenced via `uses: ./../_reusable-workflows/...`),
   so even that step is a plain `uses:`. (The guard still supports an injectable allowlist, but
   it's empty.) `actions/github-script` is also banned — it embeds an inline JS `script:` body.
2. **Node + tested; runtime stays dependency-free.** Scripts are ESM `.mjs` on **Node 24**,
   tested with `node:test` + `node:assert/strict`. Action/script **runtime** uses only Node
   built-ins — no third-party runtime deps. The only npm packages are **dev tooling**: the root
   `eslint` + `prettier` (lint this repo's own `.mjs` + YAML) and shadow's isolated `tsc`. CI gates
   coverage (lines/functions/branches) — keep it green.
3. **Lint + format are enforced.** `eslint` checks correctness + the **module allowlist** over
   `.mjs` and `.mts`; `prettier` owns formatting (the two never overlap — `eslint-config-prettier`
   and `yml/prettier` defer all style to Prettier). Both run in `test.yaml` via
   `node scripts/lint.mjs`; config lives in `eslint.config.mjs` + `.prettierrc.json`. Run
   `npm run lint:fix` to auto-resolve. `tsc` still owns `.mts` **types** (`shadow/typecheck.mjs`).
4. **Module allowlist** (ESLint `no-restricted-imports`): source may import only `node:*` ·
   relative · `yaml` · `@actions/*`. `@actions/*` is permitted but unused — narrow it later;
   adopting one means adding it to a `package.json` (`check-deps` allows `yaml` + `@actions/*`).

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
TypeScript run on **Node 24** (type-stripping at runtime — no transpile step). Its runtime
dependency is **`yaml`** (the only one in use); the allowlist also permits `@actions/*` (see the
module allowlist above), and `shadow/src/bin/check-deps.mts` enforces `deps ⊆ {yaml, @actions/*}`.
Types are checked separately by the isolated `tsc` (`node shadow/typecheck.mjs → tsc --noEmit`);
the runtime never runs tsc. devDeps exist solely for that typecheck.

Terminology (used consistently in code, **CLI flags**, and docs):

- **workflows** = this repo (`reusable-workflows`) — its PRs are what we test (`--workflows-repo/-ref/-pr`).
- **runner** = `reusable-workflows-shadow-testing` — the venue where shadow PRs run a consumer's CI; a thin `receiver.yaml` shim that checks out this repo and runs `shadow/`.
- **consumer** = a downstream repo (from `.github/shadow-consumers.json`) we mirror to verify the draft doesn't break it.

Conventions specific to `shadow/`:

- **Config comes from CLI flags, not env** — bins read `--workflows-repo` etc. via `requireArgs`
  (use `--flag=${{ ... }}` in YAML so empty values don't break parsing). Only **secrets**
  (`SHADOW_PAT`) and GHA sinks (`GITHUB_OUTPUT`) stay in env.
- Each shadow test is its own PR check named **`Shadow: <consumer>`** — done by naming the matrix
  job (`name: 'Shadow: ${{ matrix.consumer.repo }}'`), NOT a Checks-API check run. (A check run
  created from inside a workflow can't choose its check-suite, so it nests under the wrong workflow
  and is hard to find — naming the job keeps the check correctly grouped under this workflow.)
- The result is a markdown **table** written to the job summary (`$GITHUB_STEP_SUMMARY`); logs are
  **plain text with full URLs** (GitHub logs don't render markdown).
- **Runtime deps are allowlisted** to `yaml` + `@actions/*` (ESLint import rule + `check-deps`).
  `yaml` is the only one in use; don't add anything outside the allowlist.

## Run locally (what CI runs)

```sh
node scripts/lib/guard/check-no-inline-scripts.cli.mjs   # no inline run: blocks
node scripts/lint.mjs                                     # eslint + prettier (.mjs + YAML)
node shadow/src/bin/check-deps.mts                        # shadow: yaml-only runtime dep
node shadow/typecheck.mjs                                 # isolated tsc --noEmit
node --test --experimental-test-coverage \
  --test-coverage-lines=95 --test-coverage-functions=100 --test-coverage-branches=90 \
  '--test-coverage-include=actions/**/*.mjs' '--test-coverage-include=scripts/**/*.mjs' \
  '--test-coverage-include=shadow/src/**/*.mts' '--test-coverage-exclude=**/*.test.*' \
  '--test-coverage-exclude=**/*.cli.mjs' '--test-coverage-exclude=shadow/src/bin/**' \
  '--test-coverage-exclude=shadow/src/adapters/**' \
  'actions/**/*.test.mjs' 'scripts/**/*.test.mjs' 'shadow/test/*.test.mts'
```
