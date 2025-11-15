#!/usr/bin/env bash
set -euo pipefail

# Trigger Plan 240E Regression GitHub Actions workflow via REST API.
# Requirements:
#   - env GITHUB_TOKEN with 'workflow' + 'repo' (classic) or fine-grained: Actions Read/Write on this repo
#   - Repo settings: Actions workflow permissions -> Read and write
#
# Usage:
#   GITHUB_TOKEN=xxxx scripts/plan240/trigger-240e-ci.sh [owner] [repo] [ref]
# Defaults:
#   owner=jacksonlee411 repo=cube-castle ref=master

OWNER="${1:-jacksonlee411}"
REPO="${2:-cube-castle}"
REF="${3:-master}"

test -n "${GITHUB_TOKEN:-}" || { echo "❌ Missing GITHUB_TOKEN"; exit 2; }

URL="https://api.github.com/repos/${OWNER}/${REPO}/actions/workflows/plan-240e-regression.yml/dispatches"

echo "🚀 Triggering workflow_dispatch: ${OWNER}/${REPO}@${REF}"
CODE=$(curl -sS -o /tmp/resp_240e.txt -w "%{http_code}" \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -X POST "${URL}" \
  -d "{\"ref\":\"${REF}\"}")

echo "HTTP ${CODE}"
if [ "${CODE}" != "204" ]; then
  echo "Response:"
  cat /tmp/resp_240e.txt
  exit 1
fi

echo "✅ Dispatched successfully."

