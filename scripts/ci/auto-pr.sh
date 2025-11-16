#!/usr/bin/env bash
#
# Auto PR Creator (CI/Local Automation)
# - Push current (or specified) branch to origin
# - Create a GitHub Pull Request via gh CLI or GitHub REST API (curl)
# - Load token from secrets/.env.local → secrets/.env → .env.local → .env → env (GITHUB_TOKEN/GH_TOKEN)
# - Write PR info to logs/plan255/pr-<timestamp>.txt
#
# Usage:
#   bash scripts/ci/auto-pr.sh --title "My PR" --body-file docs/development-plans/255-soft-gate-PR.md \
#     [--base master] [--head <branch>] [--draft]
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
LOG_DIR="${REPO_ROOT}/logs/plan255"
mkdir -p "${LOG_DIR}"
TS="$(date +%Y%m%d_%H%M%S)"
LOG_FILE="${LOG_DIR}/pr-${TS}.txt"

# Defaults
TITLE=""
BODY_FILE=""
BASE_BRANCH="master"
HEAD_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
DRAFT="false"

# Parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
    --title) TITLE="${2:-}"; shift 2 ;;
    --body-file) BODY_FILE="${2:-}"; shift 2 ;;
    --base) BASE_BRANCH="${2:-}"; shift 2 ;;
    --head) HEAD_BRANCH="${2:-}"; shift 2 ;;
    --draft) DRAFT="true"; shift 1 ;;
    *) echo "Unknown arg: $1"; exit 2 ;;
  esac
done

if [[ -z "${TITLE}" || -z "${BODY_FILE}" ]]; then
  echo "Usage: $0 --title \"PR title\" --body-file path/to/body.md [--base master] [--head branch] [--draft]" >&2
  exit 2
fi
if [[ ! -f "${BODY_FILE}" ]]; then
  echo "Body file not found: ${BODY_FILE}" >&2
  exit 2
fi

# Resolve origin owner/repo (supports ssh://git@... or https://...)
ORIGIN_URL="$(git remote get-url origin)"
if [[ -z "${ORIGIN_URL}" ]]; then
  echo "origin remote not found" >&2
  exit 2
fi

OWNER=""
REPO=""
if [[ "${ORIGIN_URL}" =~ ^ssh://git@.*github.com.*$ ]]; then
  # ssh://git@ssh.github.com:443/owner/repo.git
  OWNER_REPO="${ORIGIN_URL##*/github.com*/}" || true
  # Fallback: strip up to last '/'
  [[ -z "${OWNER_REPO}" ]] && OWNER_REPO="${ORIGIN_URL##*/}"
  OWNER="$(basename "$(dirname "${ORIGIN_URL}")")"
  # Better parse for ssh://git@ssh.github.com:443/owner/repo.git
  # Extract path after last ':' or last '/'
  PATH_PART="${ORIGIN_URL#*github.com:}"
  [[ "${PATH_PART}" == "${ORIGIN_URL}" ]] && PATH_PART="${ORIGIN_URL#*github.com/}"
  OWNER="$(echo "${PATH_PART}" | awk -F'/' '{print $(1)}')"
  REPO="$(echo "${PATH_PART}" | awk -F'/' '{print $(2)}')"
  REPO="${REPO%.git}"
elif [[ "${ORIGIN_URL}" =~ ^git@[^:]+:[^/]+/[^/]+\.git$ ]]; then
  # git@github.com:owner/repo.git
  OWNER="$(echo "${ORIGIN_URL}" | awk -F'[:/]' '{print $(2)}')"
  REPO="$(echo "${ORIGIN_URL}" | awk -F'[:/]' '{print $(3)}')"
  REPO="${REPO%.git}"
elif [[ "${ORIGIN_URL}" =~ ^https://[^/]+/[^/]+/[^/]+(\.git)?$ ]]; then
  OWNER="$(echo "${ORIGIN_URL}" | awk -F'/' '{print $(NF-1)}')"
  REPO="$(basename "${ORIGIN_URL}")"
  REPO="${REPO%.git}"
else
  # Generic fallback
  OWNER="$(basename "$(dirname "${ORIGIN_URL}")")"
  REPO="$(basename "${ORIGIN_URL}")"
  REPO="${REPO%.git}"
fi

if [[ -z "${OWNER}" || -z "${REPO}" ]]; then
  echo "Unable to parse owner/repo from origin URL: ${ORIGIN_URL}" >&2
  exit 2
fi

# Load tokens from known locations (do not fail if missing)
load_env_if_exists() {
  local f="$1"
  if [[ -f "$f" ]]; then
    # shellcheck disable=SC1090
    set -a; source "$f"; set +a
  fi
}
load_env_if_exists "${REPO_ROOT}/secrets/.env.local"
load_env_if_exists "${REPO_ROOT}/secrets/.env"
load_env_if_exists "${REPO_ROOT}/.env.local"
load_env_if_exists "${REPO_ROOT}/.env"

TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"

echo "== Auto PR Creator ==" | tee -a "${LOG_FILE}"
echo "Repo      : ${OWNER}/${REPO}" | tee -a "${LOG_FILE}"
echo "Base/Head : ${BASE_BRANCH} <- ${HEAD_BRANCH}" | tee -a "${LOG_FILE}"
echo "Title     : ${TITLE}" | tee -a "${LOG_FILE}"
echo "Body file : ${BODY_FILE}" | tee -a "${LOG_FILE}"
echo "Draft     : ${DRAFT}" | tee -a "${LOG_FILE}"

# Ensure upstream is set; if not, push with -u
if ! git rev-parse --abbrev-ref --symbolic-full-name "@{u}" >/dev/null 2>&1; then
  echo "No upstream configured for ${HEAD_BRANCH}, pushing to origin..." | tee -a "${LOG_FILE}"
  git push -u origin "${HEAD_BRANCH}" | tee -a "${LOG_FILE}"
else
  echo "Upstream already configured for ${HEAD_BRANCH}" | tee -a "${LOG_FILE}"
fi

# Create PR via gh if available; else fallback to curl
PR_URL=""
if command -v gh >/dev/null 2>&1; then
  echo "Using gh CLI to create PR..." | tee -a "${LOG_FILE}"
  set +e
  if [[ "${DRAFT}" == "true" ]]; then
    gh pr create --base "${BASE_BRANCH}" --head "${HEAD_BRANCH}" --title "${TITLE}" --body-file "${BODY_FILE}" --draft | tee -a "${LOG_FILE}"
  else
    gh pr create --base "${BASE_BRANCH}" --head "${HEAD_BRANCH}" --title "${TITLE}" --body-file "${BODY_FILE}" | tee -a "${LOG_FILE}"
  fi
  RC=$?
  set -e
  if [[ $RC -eq 0 ]]; then
    PR_URL="$(gh pr view --json url -q .url 2>/dev/null || true)"
  fi
else
  if [[ -z "${TOKEN}" ]]; then
    echo "No gh CLI and no GITHUB_TOKEN/GH_TOKEN found. Cannot create PR automatically." | tee -a "${LOG_FILE}"
    echo "You can set GITHUB_TOKEN in environment or secrets/.env.local and re-run." | tee -a "${LOG_FILE}"
    exit 1
  fi
  echo "Using GitHub REST API (curl) to create PR..." | tee -a "${LOG_FILE}"
  # Escape body content
  BODY_ESCAPED="$(sed 's/\"/\\"/g' "${BODY_FILE}")"
  API="https://api.github.com/repos/${OWNER}/${REPO}/pulls"
  RESP="$(curl -sS -X POST -H "Authorization: token ${TOKEN}" -H "Accept: application/vnd.github+json" \
    "${API}" \
    -d "{\"title\":\"${TITLE}\",\"head\":\"${HEAD_BRANCH}\",\"base\":\"${BASE_BRANCH}\",\"body\":\"${BODY_ESCAPED}\",\"draft\":${DRAFT}}")"
  echo "${RESP}" > "${LOG_DIR}/pr-response-${TS}.json"
  PR_URL="$(echo "${RESP}" | sed -n 's/.*\"html_url\"[[:space:]]*:[[:space:]]*\"\\(https:[^"]*\\)\".*/\\1/p' | head -n1)"
fi

if [[ -n "${PR_URL}" ]]; then
  echo "PR URL: ${PR_URL}" | tee -a "${LOG_FILE}"
else
  echo "PR created or attempted. If URL not printed, check GitHub UI or logs under ${LOG_DIR}." | tee -a "${LOG_FILE}"
fi

echo "Done." | tee -a "${LOG_FILE}"
