# TODO

## Secrets → OIDC only

Eliminate the long-lived `SHADOW_PAT` — per [`shadow/SECURITY.md`](shadow/SECURITY.md) it's the
entire risk surface (a standing cross-repo write token). Move all cross-repo auth to short-lived,
OIDC-minted credentials:

- Exchange the Actions **OIDC token** for an ephemeral, repo-scoped credential (a GitHub App
  installation token, or OIDC→token exchange) granting only `actions:write` on the runner +
  `contents`/`pull-requests` write — time-boxed, per-run.
- Cover every privileged op: receiver `workflow_dispatch`, shadow PR create/close, branch
  push/delete (`internal/shadow/adapters/{git,github}.go`).
- Remove `SHADOW_PAT` from all workflows + repo secrets once OIDC is live; update `shadow/SECURITY.md`.
- **Goal: no standing secret anywhere — every shadow run uses an ephemeral credential.**

## Security / supply chain

- **Least-privilege `permissions:`** on every workflow. `test.yaml` / `ci.yaml` set none → they get
  the broad default token; most need only `contents: read`. (`shadow.yaml` already scopes its token.)
- **SHA-pin all third-party actions** and enable `sha_pinning_required` (currently `false`).
  `checkout`, `setup-go`, `setup-bazel`, `golangci-lint-action`, `yamllint`, `checkout-anywhere`
  are on floating tags.
- **`govulncheck` in CI** — Go vulnerability scanning on every run.
- **Dependabot / Renovate** for `go.mod`, `MODULE.bazel`, `.bazelversion`, and the pinned action SHAs.
- **Validate untrusted receiver inputs** at the Go entrypoint (allow-list provider + consumer repos;
  require a 40-hex `workflows_ref`) — flagged in `receiver.yaml` / `shadow/SECURITY.md`, not enforced.
- **Branch protection on `main`** — require the `test` + `shared / ci` checks and a review.

## Release / versioning

- **Tag the reusable workflow** (a moving `v1` + semver tags) so consumers pin `@v1` instead of
  `@main` — the model `checkout-anywhere` already uses. Everything floats on `@main` today.

## CI cost & DX

- **Kill the cold-Bazel tax** for consumers (~3 min/job): a **remote Bazel cache** (BuildBuddy/GCS)
  and/or **prebuilt released static binaries** so consumer runners *download* an action binary
  instead of building it.
- **Concurrency groups** (`cancel-in-progress`) on `test.yaml` / `ci.yaml` so superseded commits
  don't pile up redundant builds.
- Periodically bump `.bazelversion` + rules_go / gazelle / go-github.

## Polish

- `lint` / `test` actions are scaffolds (echo inputs) — wire real logic into `internal/actions/{lint,test}`.
- Pin the golangci-lint version + enable `depguard` to re-encode the import allowlist.
- Coverage gate is line-only at 90 (Go has no fn/branch coverage) — consider per-package floors.
- **Multi-OS/arch reality check** — only `linux/amd64` is proven; hermetic-cc registers `linux/arm64`
  (unverified) and macOS/Windows runners are untested. Verify or document "linux only."

## Test

- Run the **live shadow flow end-to-end** in CI (label a PR `shadow-test`) — the one path not yet
  exercised against the Go/Bazel binaries.
- Confirm `//shadow/cmd/check-dispatch-auth` + `dispatch-auth-test.yaml` stay green against the runner.
- Add golden-file fixtures for the YAML transforms against real consumer workflows.

## Iterate

- Path-filter auto-run for shadow (changes to `actions/**`, `shared.yaml`, `shadow/**`) alongside the label.

## Docs

- **`CONTRIBUTING.md`** — how to add an action, the TDD loop, the `bazelisk` commands, the no-inline rule.
- Reconcile consumer naming: `shadow-consumers.json` → `stefanpenner-cs/reusable-workflows-consumer`,
  but the local `consumer/` remote is `stefanpenner/shared-workflow-consumer`.

## Cleanup

- Drop now-moot `node_modules` entries from `.bazelignore` / gazelle excludes.
- Sweep docs for any remaining Node-era references.
- Tidy stale branches once the open PRs land.
