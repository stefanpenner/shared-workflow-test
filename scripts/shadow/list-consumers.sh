#!/usr/bin/env bash
# Install the shadow harness and emit the consumer matrix. Runs with the harness
# checkout as the working directory (CONSUMERS_FILE points at the provider's list).
set -euo pipefail

npm ci
npx tsx src/bin/list-consumers.ts
