#!/usr/bin/env bash
# Install the shadow harness and tear down a closed provider PR's shadow PRs/branches.
# Runs with the harness checkout as the working directory; inputs arrive via env.
set -euo pipefail

npm ci
node src/bin/cleanup.ts
