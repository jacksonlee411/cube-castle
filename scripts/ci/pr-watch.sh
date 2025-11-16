#!/usr/bin/env bash
#
# PR Watcher — Poll PR checks and optionally auto-merge when "green"
# - Loads token from: secrets/.env.local → secrets/.env → .env.local → .env → env (GITHUB_TOKEN/GH_TOKEN)
# - Writes rolling logs to logs/ci-monitor/pr-watch-<timestamp>.log
# - "Green" = combined status success; OR (no failures and no in_progress/queued and mergeable=clean)
# - Designed for local automation only; respects AGENTS.md constraints (no privilege escalation)
#
# Usage:
#   bash scripts/ci/pr-watch.sh --prs 6[,7,...] --interval 60 [--merge-if-green]
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
LOG_DIR="${REPO_ROOT}/logs/ci-monitor"
mkdir -p "${LOG_DIR}" "${REPO_ROOT}/.ci"
TS="$(date +%Y%m%d_%H%M%S)"
LOG_FILE="${LOG_DIR}/pr-watch-${TS}.log"

REPO="${REPO:-}"
PRS_ARG=""
INTERVAL=60
AUTO_MERGE=0

# Parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --prs) PRS_ARG="${2:-}"; shift 2 ;;
    --interval) INTERVAL="${2:-60}"; shift 2 ;;
    --merge-if-green) AUTO_MERGE=1; shift 1 ;;
    --repo) REPO="${2:-}"; shift 2 ;;
    *) echo "Unknown arg: $1" >&2; exit 2 ;;
  esac
done

if [[ -z "${PRS_ARG}" ]]; then
  echo "Usage: $0 --prs <number[,number,...]> [--interval 60] [--merge-if-green] [--repo owner/repo]" >&2
  exit 2
fi

# Resolve owner/repo from origin if not provided
if [[ -z "${REPO}" ]]; then
  ORIGIN_URL="$(git -C "${REPO_ROOT}" remote get-url origin 2>/dev/null || true)"
  if [[ -z "${ORIGIN_URL}" ]]; then
    echo "origin remote not found; specify --repo owner/repo" >&2
    exit 2
  fi
  PATH_PART="$(echo "${ORIGIN_URL}" | sed -E 's#^ssh://git@[^/]+/##; s#^git@[^:]+:##; s#^https?://[^/]+/##')"
  OWNER="$(echo "${PATH_PART}" | cut -d'/' -f1)"
  REPO_NAME="$(echo "${PATH_PART}" | cut -d'/' -f2)"
  REPO_NAME="${REPO_NAME%.git}"
  REPO="${OWNER}/${REPO_NAME}"
fi

# Token loading (do not fail if missing; read-only endpoints may still work but rate-limit applies)
load_env_if_exists() {
  local f="$1"
  if [[ -f "$f" ]]; then set -a; # shellcheck disable=SC1090
    source "$f"; set +a; fi
}
load_env_if_exists "${REPO_ROOT}/secrets/.env.local"
load_env_if_exists "${REPO_ROOT}/secrets/.env"
load_env_if_exists "${REPO_ROOT}/.env.local"
load_env_if_exists "${REPO_ROOT}/.env"
TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
AUTH_HEADER=()
if [[ -n "${TOKEN}" ]]; then
  AUTH_HEADER=(-H "Authorization: Bearer ${TOKEN}")
fi

echo "== PR Watcher ==" | tee -a "${LOG_FILE}"
echo "Repo      : ${REPO}" | tee -a "${LOG_FILE}"
echo "PRs       : ${PRS_ARG}" | tee -a "${LOG_FILE}"
echo "Interval  : ${INTERVAL}s" | tee -a "${LOG_FILE}"
echo "Auto-merge: ${AUTO_MERGE}" | tee -a "${LOG_FILE}"
echo "Log file  : ${LOG_FILE}" | tee -a "${LOG_FILE}"
echo "Watcher started. Create '${REPO_ROOT}/.ci/pr-watch.stop' to stop." | tee -a "${LOG_FILE}"

IFS=',' read -r -a PRS <<< "${PRS_ARG}"
API_ROOT="https://api.github.com/repos/${REPO}"

# Helpers (wrapped jq to avoid quoting issues)
json_get() {
  local expr="$1"
  jq -r "${expr}" 2>/dev/null || true
}

print_status_line() {
  local now iso pr html mergeable combined fails inprog
  pr="$1"; html="$2"; mergeable="$3"; combined="$4"; fails="$5"; inprog="$6"
  iso="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  printf "[%s] PR #%s %s mergeable=%s combined=%s fails=%s in_progress=%s\n" \
    "$iso" "$pr" "$html" "$mergeable" "$combined" "$fails" "$inprog" | tee -a "${LOG_FILE}"
}

should_auto_merge() {
  local combined="$1" fails="$2" inprog="$3" mergeable="$4"
  if [[ "${combined}" == "success" ]]; then
    return 0
  fi
  # docs/ci-only fast path: no failures, no in_progress, mergeable is clean
  if [[ "${fails}" == "0" && "${inprog}" == "0" && "${mergeable}" == "clean" ]]; then
    return 0
  fi
  return 1
}

while true; do
  if [[ -f "${REPO_ROOT}/.ci/pr-watch.stop" ]]; then
    echo "Stop signal detected. Exiting." | tee -a "${LOG_FILE}"
    rm -f "${REPO_ROOT}/.ci/pr-watch.stop" || true
    exit 0
  fi

  for pr in "${PRS[@]}"; do
    # Pull request core info
    PR_RESP="$(curl -sS "${AUTH_HEADER[@]}" -H "Accept: application/vnd.github+json" \
      "${API_ROOT}/pulls/${pr}")"
    HTML_URL="$(echo "${PR_RESP}" | json_get '.html_url // empty')"
    MERGEABLE_STATE="$(echo "${PR_RESP}" | json_get '.mergeable_state // "unknown"')"
    SHA="$(echo "${PR_RESP}" | json_get '.head.sha // empty')"

    # Combined status for the head SHA
    COMBINED_RESP="$(curl -sS "${AUTH_HEADER[@]}" -H "Accept: application/vnd.github+json" \
      "${API_ROOT}/commits/${SHA}/status")"
    COMBINED_STATE="$(echo "${COMBINED_RESP}" | json_get '.state // "pending"')"

    # Check runs (to detect fails/in_progress precisely)
    CHECKS_RESP="$(curl -sS "${AUTH_HEADER[@]}" -H "Accept: application/vnd.github+json" \
      "${API_ROOT}/commits/${SHA}/check-runs?per_page=100")"
    # jq: avoid custom functions; check explicit equality
    FAILS="$(echo "${CHECKS_RESP}" | jq '[.check_runs[]? | select(((.conclusion // "") == "failure") or ((.conclusion // "") == "cancelled") or ((.conclusion // "") == "timed_out") or ((.conclusion // "") == "action_required"))] | length' 2>/dev/null || echo 0)"
    INPROG="$(echo "${CHECKS_RESP}" | jq '[.check_runs[]? | select(((.status // "") == "in_progress") or ((.status // "") == "queued") or ((.status // "") == "requested") or ((.status // "") == "waiting"))] | length' 2>/dev/null || echo 0)"

    print_status_line "${pr}" "${HTML_URL}" "${MERGEABLE_STATE}" "${COMBINED_STATE}" "${FAILS}" "${INPROG}"

    if [[ "${AUTO_MERGE}" -eq 1 ]]; then
      if should_auto_merge "${COMBINED_STATE}" "${FAILS}" "${INPROG}" "${MERGEABLE_STATE}"; then
        echo "→ Conditions satisfied, attempting squash-merge PR #${pr} ..." | tee -a "${LOG_FILE}"
        MERGE_RESP="$(curl -sS -X PUT "${AUTH_HEADER[@]}" -H "Accept: application/vnd.github+json" \
          "${API_ROOT}/pulls/${pr}/merge" \
          -d '{"merge_method":"squash","commit_title":"auto-merge: squash","commit_message":""}' )" || true
        MERGED="$(echo "${MERGE_RESP}" | json_get '.merged // false')"
        if [[ "${MERGED}" == "true" ]]; then
          echo "✅ PR #${pr} merged." | tee -a "${LOG_FILE}"
        else
          MSG="$(echo "${MERGE_RESP}" | json_get '.message // empty')"
          echo "⚠️  Merge attempt did not succeed: ${MSG}" | tee -a "${LOG_FILE}"
        fi
      fi
    fi
  done

  sleep "${INTERVAL}"
done
