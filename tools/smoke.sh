#!/usr/bin/env bash
# Self-hosted smoke test: time a from-cache build of the action binaries to exercise the local
# Bazel disk/repository cache on a `local-cache` self-hosted runner. Invoked directly (not via
# `bazel run`) so the measurement reflects a real cold-ish `bazel build`, not a nested launch.
#
# This is the sanctioned home for the smoke shell: the no-inline-scripts rule requires a workflow
# `run:` to be a single external invocation (`bash tools/smoke.sh`), with the logic living here.
set -euo pipefail

echo "host: $(hostname)"

# `/usr/bin/time -v` gives peak RSS + wall/CPU; fall back to a plain build where it's unavailable.
if command -v /usr/bin/time >/dev/null 2>&1; then
  /usr/bin/time -v bazel build //actions/... 2>&1 | tail -45
else
  bazel build //actions/... 2>&1 | tail -45
fi
