#!/usr/bin/env bash
set -euo pipefail

# Plan 240E – One‑Click Runner
# - Creates unified evidence directories
# - Runs guards and Playwright regression using existing scripts
# - Does not enumerate test lists (SSoT: 232/232T). Only collects outputs.
#
# Logs:
#   - Guards:         logs/plan240/E/*.log
#   - Playwright log: logs/plan240/E/playwright-run-<ts>.log
#   - Traces:         logs/plan240/E/trace/*.zip
# HAR (optional; managed by frontend config):
#   - logs/plan240/B/*.har  or  logs/plan240/BT/*.har

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="${ROOT_DIR}/logs/plan240/E"
TRACE_DIR="${LOG_DIR}/trace"
TS="$(date +%Y%m%d%H%M%S)"

mkdir -p "${LOG_DIR}" "${TRACE_DIR}"

echo "🔧 Plan 240E – Initialize environment variables"
# Tenant can be overridden by caller; use the well-known dev tenant if unset
export PW_TENANT_ID="${PW_TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
# Observability gates (CI recommended)
export VITE_OBS_ENABLED="${VITE_OBS_ENABLED:-true}"
export VITE_ENABLE_MUTATION_LOGS="${VITE_ENABLE_MUTATION_LOGS:-true}"
# HAR capture (managed by frontend/playwright.config.ts)
export E2E_SAVE_HAR="${E2E_SAVE_HAR:-1}"
# Allow Playwright to manage dev server by default
export PW_SKIP_SERVER="${PW_SKIP_SERVER:-0}"

echo "📁 Evidence directories:"
echo "  - ${LOG_DIR}"
echo "  - ${TRACE_DIR}"

echo "🛡  Running guards (outputs will be saved under logs/plan240/E)"
{
  npm run guard:selectors-246 || true
} > "${LOG_DIR}/selector-guard.log" 2>&1
{
  npm run guard:plan245 || true
} > "${LOG_DIR}/guard-plan245.log" 2>&1
{
  node scripts/quality/architecture-validator.js || true
} > "${LOG_DIR}/architecture-validator.log" 2>&1
{
  bash scripts/check-temporary-tags.sh || true
} > "${LOG_DIR}/temporary-tags.log" 2>&1

echo "🎭 Running Playwright regression via frontend script"
pushd "${ROOT_DIR}/frontend" >/dev/null
# Ensure local dev jwt if missing (best‑effort)
if [[ ! -f "${ROOT_DIR}/.cache/dev.jwt" ]]; then
  echo "🔐 Generating dev jwt (best-effort)..."
  make -C "${ROOT_DIR}" jwt-dev-setup >/dev/null 2>&1 || true
  make -C "${ROOT_DIR}" jwt-dev-mint >/dev/null 2>&1 || true
fi

# Run the unified 240E runner (collects stdout to logs/plan240/E and copies traces)
set +e
npm run test:e2e:240e | tee "${LOG_DIR}/playwright-run-${TS}.log"
RC=${PIPESTATUS[0]}
set -e
popd >/dev/null

echo "✅ Completed. Exit code: ${RC}"
echo "   - Guards:         ${LOG_DIR}/*.log"
echo "   - Playwright log: ${LOG_DIR}/playwright-run-${TS}.log"
echo "   - Traces:         ${TRACE_DIR}/*.zip"

exit "${RC}"

