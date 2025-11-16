#!/usr/bin/env bash
#
# 通用本地 CI-like 脚本：compose -> 迁移 -> 统一门禁 -> E2E -> 打印 SUMMARY
# 依赖：docker, make, node, npm, (可选) jq/rg
#
# 用法：
#   E2E_PLAN_ID=254 bash scripts/ci/plan-local.sh
# 变量：
#   SKIP_INSTALL=1     跳过 npm ci
#   PW_SKIP_SERVER=1   跳过 Playwright 启动 webServer（本地 dev 已运行）
#   E2E_SAVE_HAR=1     生成 HAR（依赖前端配置）
#
set -euo pipefail

PLAN_ID="${E2E_PLAN_ID:-254}"
echo "🏁 Local CI-like start (plan=${PLAN_ID})"

echo "🐳 Compose up minimal deps + run services (includes migrations)"
make docker-up
make run-dev >/dev/null 2>&1 &  # 后台启动，内部有健康检查与日志
RUN_DEV_PID=$!
trap 'kill ${RUN_DEV_PID} 2>/dev/null || true' EXIT

echo "⏳ Wait for backends..."
for i in {1..60}; do
  ok=0
  curl -fsS http://localhost:9090/health >/dev/null && ok=$((ok+1)) || true
  curl -fsS http://localhost:8090/health >/dev/null && ok=$((ok+1)) || true
  if [ "$ok" = "2" ]; then echo "✅ backends healthy"; break; fi
  sleep 2
done
[ "${ok:-0}" = "2" ] || { echo "❌ backends not healthy"; exit 1; }

echo "🛡  Architecture gate (frontend: cqrs,ports,forbidden)"
node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden

echo "🎭 Playwright E2E (plan ${PLAN_ID})"
pushd frontend >/dev/null
if [ "${SKIP_INSTALL:-0}" != "1" ]; then
  npm ci
  npx playwright install --with-deps
fi
E2E_PLAN_ID="${PLAN_ID}" PW_SKIP_SERVER="${PW_SKIP_SERVER:-0}" npm run -s test:e2e:plan
popd >/dev/null

echo "🧾 Print JSON SUMMARY (if any)"
node scripts/ci/print-e2e-summary.js "${PLAN_ID}" || true

echo "✅ Done. Artifacts in logs/plan${PLAN_ID}/*"

