#!/usr/bin/env bash
# Cheap gate on the provider draft: download actionlint and lint all workflows.
# Runs from the provider checkout root.
set -euo pipefail

bash <(curl -sSf https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)
./actionlint -color .github/workflows/*.y*ml
