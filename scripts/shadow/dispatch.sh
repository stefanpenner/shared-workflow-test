#!/usr/bin/env bash
# Install the shadow harness and dispatch one consumer's shadow run, watching it to
# completion. Runs with the harness checkout as the working directory; all inputs
# (PROVIDER_*/CONSUMER_*/SHADOW_PAT/HARNESS_REPO) arrive via env.
set -euo pipefail

npm ci
node src/bin/dispatch-and-watch.ts
