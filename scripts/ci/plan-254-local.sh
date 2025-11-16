#!/usr/bin/env bash
set -euo pipefail

# Plan 254 - Local CI-like Gate Runner
# - Compose up E2E stack
# - Health checks
# - Mint dev JWT
# - Architecture Gate (CQRS/Ports/Forbidden)
# - Playwright E2E with E2E_PLAN_ID=254 (evidence to logs/plan254/*)
#
# Usage:
#   bash scripts/ci/plan-254-local.sh
# Optional env:
#   SKIP_INSTALL=1        # do not run npm ci in frontend
#   E2E_SAVE_HAR=1        # save HAR files
#   FRONTEND_BASE=http://localhost:3000

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

echo "==> Plan 254 Local Gate starting..."

command -v docker >/dev/null || { echo "docker not found"; exit 1; }
command -v curl >/dev/null || { echo "curl not found"; exit 1; }
command -v node >/dev/null || { echo "node not found"; exit 1; }
command -v npm >/dev/null || { echo "npm not found"; exit 1; }

FRONTEND_BASE="${FRONTEND_BASE:-http://localhost:3000}"

E2E_STACK_OK=0
if [ -f "docker-compose.e2e.yml" ] && [ -d "cmd/hrms-server/query-unified" ]; then
  echo "==> Compose Up (docker-compose.e2e.yml)"
  docker compose -f docker-compose.e2e.yml build
  docker compose -f docker-compose.e2e.yml up -d
  E2E_STACK_OK=1
else
  echo "==> E2E stack unavailable or unified query not present; falling back to dev stack (make docker-up && make run-dev)"
  make docker-up
  make run-dev
fi

echo "==> Wait for services (9090, 8090, 3000)"
ok=0
for i in {1..60}; do
  ok=0
  curl -fsS http://localhost:9090/health >/dev/null && ok=$((ok+1)) || true
  curl -fsS http://localhost:8090/health >/dev/null && ok=$((ok+1)) || true
  curl -fsS "${FRONTEND_BASE}/" >/dev/null && ok=$((ok+1)) || true
  if [ "$ok" = "3" ]; then echo "services healthy"; break; fi
  sleep 5
done
if [ "$ok" != "3" ]; then
  echo "Services did not become healthy in time"
  if [ "$E2E_STACK_OK" = "1" ]; then
    docker compose -f docker-compose.e2e.yml logs --no-color || true
  else
    echo "(dev stack) see run-dev logs: run-dev*.log"
  fi
  exit 1
fi

echo "==> Mint dev JWT"
make jwt-dev-mint

echo "==> Architecture Gate (CQRS/Ports/Forbidden)"
node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden

echo "==> Frontend E2E (Plan 254)"
pushd frontend >/dev/null
if [ "${SKIP_INSTALL:-0}" != "1" ]; then
  if [ ! -d node_modules ]; then
    npm ci
  else
    npm i --no-fund --no-audit
  fi
fi
# If E2E stack not available, let Playwright start webServer; otherwise use existing frontend
if [ "$E2E_STACK_OK" = "1" ]; then
  E2E_PLAN_ID=254 PW_SKIP_SERVER=1 PW_BASE_URL="${FRONTEND_BASE}" E2E_SAVE_HAR="${E2E_SAVE_HAR:-0}" npm run -s test:e2e:254
else
  E2E_PLAN_ID=254 PW_SKIP_SERVER=0 PW_BASE_URL="${FRONTEND_BASE}" E2E_SAVE_HAR="${E2E_SAVE_HAR:-0}" npm run -s test:e2e:254
fi
popd >/dev/null

echo "==> Evidence"
echo "  logs/plan254/playwright-254-run-*.log"
echo "  logs/plan254/trace/*.zip"
echo "  logs/plan254/report-<timestamp>/"
if [ "${E2E_SAVE_HAR:-0}" = "1" ]; then
  echo "  logs/plan254/har/*.har"
fi
if ls logs/plan254/results-*.json >/dev/null 2>&1; then
  LAST_JSON="$(ls -t logs/plan254/results-*.json | head -1)"
  echo "  ${LAST_JSON}"
  echo "==> Parsing JSON results summary"
  node - "$LAST_JSON" <<'NODE'
const fs = require('fs');
const file = process.argv[1];
try {
  const data = JSON.parse(fs.readFileSync(file, 'utf8'));
  let total = 0, passed = 0, failed = 0, skipped = 0, flaky = 0;
  const walk = (o) => {
    if (!o) return;
    if (Array.isArray(o)) { o.forEach(walk); return; }
    if (typeof o === 'object') {
      if (Object.prototype.hasOwnProperty.call(o, 'outcome')) {
        total++;
        const v = String(o.outcome);
        if (v === 'expected' || v === 'passed') passed++;
        else if (v === 'unexpected' || v === 'failed') failed++;
        else if (v === 'skipped') skipped++;
        else if (v === 'flaky') flaky++;
      }
      for (const k in o) walk(o[k]);
    }
  };
  walk(data);
  console.log(`SUMMARY total=${total} passed=${passed} failed=${failed} skipped=${skipped} flaky=${flaky}`);
  if (failed > 0) process.exit(1);
} catch (e) {
  console.error('Failed to parse results json:', e.message);
  process.exit(1);
}
NODE
  echo "==> Results summary parsed."
fi

echo "==> Plan 254 Local Gate completed successfully."
exit 0
