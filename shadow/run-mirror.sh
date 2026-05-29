#!/usr/bin/env bash
# Receiver entrypoint: install shadow deps (mirror-and-test needs `yaml`) and mirror the consumer,
# running its CI against the workflows draft. Invoked as a single external script by the runner
# repo's receiver.yaml shim, which has checked out the workflows repo (so shadow/ is on disk here).
set -euo pipefail
cd "$(dirname "$0")"
npm ci
node src/bin/mirror-and-test.ts
