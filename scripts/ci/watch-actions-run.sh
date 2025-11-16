#!/usr/bin/env bash
set -euo pipefail
#
# watch-actions-run.sh
# - Watch a GitHub Actions run until it completes (success/failure/cancelled)
# - Option A: provide --run-id <id>
# - Option B: provide --workflow <name> (e.g. plan-258-gates) and optional --branch <name> to watch latest run
# - Reads token from secrets/.env.local (GITHUB_TOKEN) or env var GITHUB_TOKEN/GH_TOKEN
# - Prints status every --interval seconds (default: 10), with --timeout seconds (default: 1800)
# - On completion prints run URL, conclusion, and lists artifacts (names)
#
# Usage examples:
#   bash scripts/ci/watch-actions-run.sh --workflow plan-258-gates --branch master
#   bash scripts/ci/watch-actions-run.sh --run-id 19403384404
#

INTERVAL=10
TIMEOUT=1800
RUN_ID=""
WORKFLOW_NAME=""
BRANCH="master"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --run-id) RUN_ID="${2:-}"; shift 2 ;;
    --workflow) WORKFLOW_NAME="${2:-}"; shift 2 ;;
    --branch) BRANCH="${2:-}"; shift 2 ;;
    --interval) INTERVAL="${2:-}"; shift 2 ;;
    --timeout) TIMEOUT="${2:-}"; shift 2 ;;
    -h|--help)
      grep '^#' "$0" | sed -E 's/^# ?//'
      exit 0
      ;;
    *) echo "Unknown arg: $1"; exit 2 ;;
  esac
done

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

# Load token
if [[ -f "secrets/.env.local" ]]; then
  # shellcheck disable=SC1091
  source secrets/.env.local
fi
TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
if [[ -z "${TOKEN}" ]]; then
  echo "✗ Missing GITHUB_TOKEN (set env or secrets/.env.local)"
  exit 1
fi

# Resolve owner/repo from git remote
REMOTE_URL="$(git remote get-url origin 2>/dev/null || true)"
if [[ -z "${REMOTE_URL}" ]]; then
  echo "✗ git remote 'origin' not found"
  exit 2
fi
# Normalize to "<owner>/<repo>.git" path regardless of scheme/host/port
OWNER=""
REPO=""
PATH_PART="$(echo "${REMOTE_URL}" | sed -E 's#^ssh://git@[^/]+(:[0-9]+)?/##; s#^git@[^:]+:##; s#^https?://[^/]+/##')"
if [[ -n "${PATH_PART}" ]]; then
  OWNER="$(echo "${PATH_PART}" | cut -d'/' -f1)"
  REPO="$(echo "${PATH_PART}" | cut -d'/' -f2)"
fi
REPO="${REPO%.git}"
if [[ -z "${OWNER}" || -z "${REPO}" ]]; then
  echo "✗ Unable to parse owner/repo from origin URL: ${REMOTE_URL}"
  exit 3
fi

api() {
  local url="$1"
  curl -sSf -H "Authorization: Bearer ${TOKEN}" -H "Accept: application/vnd.github+json" "${url}"
}

get_workflow_id_by_name() {
  local name="$1"
  local url="https://api.github.com/repos/${OWNER}/${REPO}/actions/workflows"
  api "${url}" | jq -r --arg n "${name}" '.workflows[] | select(.name == $n) | .id' | head -n1
}

get_latest_run_id_for_workflow() {
  local wid="$1"
  local url="https://api.github.com/repos/${OWNER}/${REPO}/actions/workflows/${wid}/runs?branch=${BRANCH}&per_page=1"
  api "${url}" | jq -r '.workflow_runs[0].id'
}

get_run_status() {
  local rid="$1"
  local url="https://api.github.com/repos/${OWNER}/${REPO}/actions/runs/${rid}"
  api "${url}"
}

list_artifacts() {
  local rid="$1"
  local url="https://api.github.com/repos/${OWNER}/${REPO}/actions/runs/${rid}/artifacts"
  api "${url}" | jq -r '.artifacts[]?.name'
}

if [[ -z "${RUN_ID}" ]]; then
  if [[ -z "${WORKFLOW_NAME}" ]]; then
    echo "✗ Provide --run-id or --workflow"
    exit 2
  fi
  WID="$(get_workflow_id_by_name "${WORKFLOW_NAME}")"
  if [[ -z "${WID}" || "${WID}" == "null" ]]; then
    echo "✗ Workflow not found by name: ${WORKFLOW_NAME}"
    exit 4
  fi
  RUN_ID="$(get_latest_run_id_for_workflow "${WID}")"
  if [[ -z "${RUN_ID}" || "${RUN_ID}" == "null" ]]; then
    echo "✗ No runs found for workflow ${WORKFLOW_NAME} on branch ${BRANCH}"
    exit 5
  fi
fi

echo "🔎 Watching run: https://github.com/${OWNER}/${REPO}/actions/runs/${RUN_ID}"
deadline=$(( $(date +%s) + TIMEOUT ))
last_status=""
while true; do
  RUN_JSON="$(get_run_status "${RUN_ID}")" || { echo "✗ Failed to fetch run status"; exit 6; }
  status="$(echo "${RUN_JSON}" | jq -r '.status')"
  conclusion="$(echo "${RUN_JSON}" | jq -r '.conclusion // empty')"
  name="$(echo "${RUN_JSON}" | jq -r '.name')"
  head_branch="$(echo "${RUN_JSON}" | jq -r '.head_branch')"
  echo "⏱  ${name} on ${head_branch}: status=${status} conclusion=${conclusion:-<pending>}"
  if [[ "${status}" == "completed" ]]; then
    echo "✅ Completed with conclusion: ${conclusion}"
    echo "📦 Artifacts:"
    list_artifacts "${RUN_ID}" || true
    if [[ "${conclusion}" != "success" ]]; then
      exit 7
    fi
    exit 0
  fi
  now="$(date +%s)"
  if (( now >= deadline )); then
    echo "⏰ Timeout after ${TIMEOUT}s waiting for run ${RUN_ID}"
    exit 8
  fi
  sleep "${INTERVAL}"
done
