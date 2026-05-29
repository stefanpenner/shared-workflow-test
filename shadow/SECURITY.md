# Shadow testing — security model

Shadow testing runs **un-merged draft code in a repo that holds a powerful secret**
(`SHADOW_PAT`). That is the whole risk surface. This document states the trust model, who is and
isn't allowed to set it in motion, and the runtime test that proves the boundary holds.

## The trust boundary

```
workflows repo (reusable-workflows)                runner repo (…-shadow-testing)
  shadow.yaml ──(SHADOW_PAT)── workflow_dispatch ──▶ receiver.yaml
   gated by 'shadow-test' label or manual dispatch     checks out workflows_repo@workflows_ref
                                                       and runs ITS Go with SHADOW_PAT in scope
```

The receiver (`harness/.github/workflows/receiver.yaml`) checks out `inputs.workflows_repo` at
`inputs.workflows_ref` and runs that checkout's Go (`bazelisk run //shadow/cmd/mirror-and-test`)
with `secrets.SHADOW_PAT` and `contents`/`pull-requests` write. **Whoever dispatches the receiver
fully controls what code executes with the secret.** So the dispatcher is a *fully trusted
principal*, not an arbitrary caller — applying the `shadow-test` label (or manually dispatching)
means "I have reviewed this draft and trust it to run with secrets."

The workflow *definition* always runs from the dispatched `ref` (the flow passes `main`), so
`receiver.yaml` itself can't be swapped by a PR. But `inputs` are otherwise **untrusted strings**;
they must be validated by the Go entrypoint (allow-list the provider + consumer repos, require a
40-hex SHA for `workflows_ref`) — do not rely on the shim to bound them.

## Who can dispatch the receiver

`workflow_dispatch` requires `actions:write`. Therefore **only**:

- principals with write/maintain/admin on the runner repo (Actions UI, `gh workflow run`, REST), and
- any holder of a token with `actions:write` there — in practice `SHADOW_PAT`, which the shadow flow
  uses to dispatch programmatically.

**Cannot** dispatch: fork contributors, read-only collaborators, the public, and another repo's
default `GITHUB_TOKEN`. GitHub withholds `actions:write` from all of these.

Consequence: `SHADOW_PAT`'s scope and secrecy are the entire ballgame. Keep it a fine-grained,
minimum-scope token limited to the runner repo, and prefer delivering it via an environment-scoped
secret the shadow PR's own CI can't read.

## Runtime verification

`//shadow/cmd/check-dispatch-auth` proves the boundary at runtime: it POSTs the receiver's
`workflow_dispatch` endpoint with **no credential** and with an **invalid credential** and asserts
GitHub refuses both (HTTP 401/403/404). It deliberately never sends a valid token, so it cannot
trigger a real run — safe to run anywhere, including inside Actions (it ignores `GITHUB_TOKEN`).

Run it locally:

```sh
bazelisk run //shadow/cmd/check-dispatch-auth -- --runner-repo=stefanpenner-cs/reusable-workflows-shadow-testing
```

It runs weekly + on demand via `.github/workflows/dispatch-auth-test.yaml`. Last manual run (against
the public runner repo) returned `HTTP 401` for both the unauthenticated and invalid-token probes.

### What the automated probe does and does not cover

| Caller | Covered by probe? | Result / basis |
| --- | --- | --- |
| Unauthenticated / public | ✅ exercised | HTTP 401 (refused) |
| Invalid / garbage token | ✅ exercised | HTTP 401 (refused) |
| Valid token, read-only collaborator | ⚠️ not exercised | GitHub returns 403/404 per its documented permission model |
| Fork contributor | ⚠️ not exercised | no `actions:write`; cannot dispatch per GitHub's model |

The two ⚠️ rows rest on GitHub's documented behavior, not a live assertion — exercising them needs a
second, lower-privileged identity. The classifier (`core.DispatchRejected`) already treats 403/404
as refusals, so the probe will assert them too if such a credential is ever supplied.
