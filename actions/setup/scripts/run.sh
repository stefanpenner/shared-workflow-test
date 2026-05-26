#!/usr/bin/env bash
set -euo pipefail
echo "Setting up environment for $PROJECT_NAME..."
echo "node_version=$NODE_VERSION" >> "$GITHUB_OUTPUT"
