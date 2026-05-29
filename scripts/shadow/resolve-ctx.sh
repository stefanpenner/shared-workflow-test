#!/usr/bin/env bash
# Resolve the provider PR number and head SHA to shadow-test, writing them as step
# outputs. GHA context values arrive via env so this stays inline-logic-free in YAML.
#   EVENT_NAME, PR_NUMBER, PR_HEAD_SHA (pull_request); INPUT_PR (workflow_dispatch).
# Requires GH_TOKEN for the dispatch lookup.
set -euo pipefail

if [ "$EVENT_NAME" = "pull_request" ]; then
  pr="$PR_NUMBER"
  sha="$PR_HEAD_SHA"
else
  pr="$INPUT_PR"
  sha="$(gh pr view "$INPUT_PR" --json headRefOid --jq .headRefOid)"
fi

{
  echo "pr=$pr"
  echo "sha=$sha"
} >> "$GITHUB_OUTPUT"
