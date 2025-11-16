#!/usr/bin/env bash
set -euo pipefail
#
# Fetch GitHub Actions SUMMARY lines for a workflow run (machine-readable)
# Usage:
#   GITHUB_TOKEN=xxxx scripts/ci/fetch-gh-summary.sh <owner/repo> <run_id>
#
# Notes:
# - Requires: curl, unzip (or bsdtar), node
# - Does not print token; reads it from env var $GITHUB_TOKEN
#

REPO="${1:-}"
RUN_ID="${2:-}"

if [ -z "${GITHUB_TOKEN:-}" ]; then
  echo "GITHUB_TOKEN is required (export GITHUB_TOKEN=...)"
  exit 1
fi
if [ -z "$REPO" ] || [ -z "$RUN_ID" ]; then
  echo "Usage: GITHUB_TOKEN=xxxx $0 <owner/repo> <run_id>"
  exit 1
fi

API="https://api.github.com/repos/${REPO}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "==> Listing jobs for run ${RUN_ID}"
curl -fsSL -H "Authorization: token ${GITHUB_TOKEN}" \
  "${API}/actions/runs/${RUN_ID}/jobs?per_page=100" > "${TMP_DIR}/jobs.json"

JOB_IDS=$(node -e "const f=require('fs');const j=JSON.parse(f.readFileSync('${TMP_DIR}/jobs.json','utf8'));console.log((j.jobs||[]).map(x=>x.id).join(' '));")
if [ -z "$JOB_IDS" ]; then
  echo "No jobs found for run ${RUN_ID}"
  exit 1
fi

FOUND=0
for id in $JOB_IDS; do
  echo "==> Downloading logs for job ${id}"
  OUT="${TMP_DIR}/job-${id}.zip"
  curl -fsSL -H "Authorization: token ${GITHUB_TOKEN}" \
    "${API}/actions/jobs/${id}/logs" -o "${OUT}" || continue

  if command -v unzip >/dev/null 2>&1; then
    unzip -p "${OUT}" | grep -E '^SUMMARY ' && FOUND=1 || true
  elif command -v bsdtar >/dev/null 2>&1; then
    bsdtar -xOf "${OUT}" | grep -E '^SUMMARY ' && FOUND=1 || true
  else
    echo "WARN: unzip/bsdtar not found; attempting strings fallback"
    strings "${OUT}" | grep -E '^SUMMARY ' && FOUND=1 || true
  fi
done

if [ "$FOUND" = "0" ]; then
  echo "No SUMMARY lines found in job logs."
  exit 1
fi

exit 0

