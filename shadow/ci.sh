#!/usr/bin/env bash
# CI for the shadow/ subtree: install + typecheck + lint + test. Invoked as a single
# external script from the provider's test workflow (keeps npm out of inline YAML).
set -euo pipefail
cd "$(dirname "$0")"
npm ci
npm run typecheck
npm run lint
npm test
