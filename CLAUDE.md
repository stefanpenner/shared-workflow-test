# CLAUDE.md

Conventions for this repo (a GitHub reusable-workflow + composite-action provider, written in **Go**
and built with **Bazel**). Follow them exactly.

## Non-negotiable rules

1. **No inline scripts in YAML.** Every action/workflow `run:` is a single external invocation —
   `bazelisk run //…`, `go …`, or a bare script — never embedded shell logic (`&&`, `||`, `;`, `|`,
   `>`, `$(...)`, `run: |` blocks). `actions/github-script` is banned (it embeds inline JS). Enforced
   by `internal/noinlinescripts`, run in CI as `bazelisk run //tools/no-inline-scripts`.
2. **Go, pure, built with Bazel.** All code is Go (`CGO_ENABLED=0`, no cgo — see `.bazelrc`), built
   and tested with **Bazel** via **bazelisk** (version pinned in `.bazelversion`). `rules_go` +
   `gazelle` + `hermetic_cc_toolchain`; the Go SDK is downloaded by Bazel. Third-party deps: cobra
   (CLIs), testify (tests), go-github (shadow API), yaml.v3 (workflow edits) — declared in `go.mod`,
   wired into `MODULE.bazel` via `bazel mod tidy`. **BUILD.bazel files are gazelle-generated** — run
   `bazelisk run //:gazelle`, don't hand-edit.
3. **Runtime model: `bazelisk run` on the consumer runner.** Each composite action's `run:` is
   `bazelisk run //actions/<x> -- --flags` with `working-directory: ${{ github.action_path }}` (so
   Bazel finds `MODULE.bazel`/`.bazelversion` at the checked-out provider root). `shared.yaml`
   installs bazelisk via `bazel-contrib/setup-bazel` before the actions.
4. **CLI args, not env vars, for parameters.** Parameters are named cobra flags
   (`--flag=value`), validated non-empty via `ghactions.RequireFlags`. Env is reserved for global
   state/sinks/secrets: `GITHUB_OUTPUT`, `GITHUB_STEP_SUMMARY`, `GITHUB_WORKSPACE`, `HOME`,
   `SHADOW_PAT`, `GH_TOKEN`. In YAML pass values as `--flag=${{ inputs.x }}`.
5. **Lint + coverage are enforced.** `golangci-lint` (Go) + `yamllint` (YAML) + `gofmt`; a **line**
   coverage gate (`go run ./tools/covergate -min 90`) over the pure layers (mains/bins/adapters are
   excluded — Go has no function/branch coverage). All run in `test.yaml`.

## Layout

- `internal/**` — **pure** logic, all unit-tested (testify): `ghactions` (log formatting, output
  sink, flag validation), `actions/<x>` (per-action logic), `guard`, `shadow/core`. This is the
  coverage-gated set.
- `internal/shadow/adapters` — **I/O** (process exec, git, go-github, workflow patching). Excluded
  from the coverage gate; exercised by the live shadow flow + httptest.
- `actions/<x>/main.go`, `shadow/cmd/<x>/main.go` — thin cobra entrypoints (`go_binary`). The only
  place that parses argv / reads env sinks / writes files. Targets: `//actions/<x>`,
  `//shadow/cmd/<x>`.
- `tools/no-inline-scripts`, `tools/covergate` — CI tools.

## Style

- **TDD — red → green → refactor.** Write a failing test first, the minimum code to pass, then clean
  up green. Refactors change structure, not behavior. Keep logic in small pure functions.
- **Simplicity & correctness over cleverness.** Clean models, not hacks; prefer the boring version.
- **Errors: no silent failures.** Wrap with context (`fmt.Errorf("…: %w", err)`); only swallow the
  error you expect (e.g. `debug.ErrProbe`).
- A `main.go` stays tiny — flags + sinks + one call into `internal/`.

## Shadow testing (`shadow/cmd` + `internal/shadow`)

Pre-merge-tests this repo's changes against real consumers (`.github/shadow-consumers.json`) under a
real `pull_request` event. **workflows** = this repo; **runner** =
`reusable-workflows-shadow-testing` (a `receiver.yaml` shim that checks this repo out and runs
`bazelisk run //shadow/cmd/mirror-and-test`); **consumer** = a downstream repo we mirror. Each shadow
test is its own PR check named `Shadow: <consumer>` (the matrix job name); results render as a
markdown table to `$GITHUB_STEP_SUMMARY`, logs as plain text. GitHub ops go through **go-github**
(token from `SHADOW_PAT`/`GH_TOKEN`); commits are reproducible (fixed dates) so re-runs are no-ops.
The trust model (and who may dispatch the runner) is documented in `shadow/SECURITY.md`, with a
runtime regression test at `//shadow/cmd/check-dispatch-auth` (workflow: `dispatch-auth-test.yaml`).

## Run locally (what CI runs)

```sh
bazelisk run //tools:lint           # no-inline-scripts + golangci-lint + yamllint (hermetic)
bazelisk run //tools:fix            # apply the auto-fixable subset (golangci-lint --fix)
bazelisk test //...                 # every *_test.go
go run ./tools/covergate -min 90    # line-coverage gate over the pure layers
```

### Hermetic dev workbench (no host installs)

All CLI tools are Bazel-pinned, so local runs match CI byte-for-byte: golangci-lint (via
`rules_multitool`, `tools/multitool.lock.json`), yamllint (via `rules_python`, `tools/requirements_lock.txt`),
and the Go SDK. Two ways to use them:

```sh
bazelisk run //tools:lint           # aggregate; or //tools:lint-go, //tools:lint-yaml individually
```

Or put them on `PATH` with [`bazel_env`](https://github.com/buildbuddy-io/bazel_env.bzl) + direnv —
a `.envrc` is committed; run `direnv allow` once (after enabling the direnv shell hook), then plain
`yamllint .`, `golangci-lint run`, and `go` resolve to the pinned versions:

```sh
bazelisk run //tools:bazel_env      # (re)generate the PATH bin dir; direnv picks it up
```

Tools that operate on the source tree run through `//tools/wsrun`, which cd's into
`BUILD_WORKSPACE_DIRECTORY` first (the non-Go analogue of the `os.Chdir` our Go tools do). golangci-lint
config is `.golangci.yml`; covergate still uses the host Go toolchain (`go run`).
