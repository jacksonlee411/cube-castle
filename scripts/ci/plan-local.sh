#!/usr/bin/env bash
#
# 本地 CI-like（与 PR 等效）脚本：
# compose → 迁移/启动 → 统一门禁（前端/后端）→ E2E（DevServer）→ SUMMARY
#
# 依赖：docker, make, node, npm, (可选) jq/rg, go, golangci-lint(脚本会安装固定版本)
#
set -euo pipefail

# 计划号：默认与 PR 等效（255）。如需自定义，外部传入 E2E_PLAN_ID 覆盖
PLAN_ID="${E2E_PLAN_ID:-255}"
TS="$(date +%Y%m%d_%H%M%S)"
echo "🏁 Local CI-like start (plan=${PLAN_ID})"

# 准备日志目录（与 CI 工件路径一致）
mkdir -p "logs/plan${PLAN_ID}/trace"

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

echo "🔐 Prepare dev JWT keys and token"
make jwt-dev-setup
make jwt-dev-mint

echo "🛡  Frontend Architecture gate (cqrs,ports,forbidden)"
node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden

echo "📝 ESLint Architecture guard (non-blocking; AST-level hints)"
npx eslint -c eslint.config.architecture.mjs "frontend/src/**/*.{ts,tsx}" \
  2>&1 | tee "logs/plan${PLAN_ID}/eslint-architecture-${TS}.log" || true

echo "🔎 Audit root (ports/forbidden; non-blocking)"
node scripts/quality/architecture-validator.js --scope root --rule ports,forbidden \
  2>&1 | tee "logs/plan${PLAN_ID}/audit-root-${TS}.log" || true

echo "🧰 Install golangci-lint (pinned v1.59.1)"
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
"$(go env GOPATH)"/bin/golangci-lint version

echo "🛡  Backend gate (golangci-lint)"
"$(go env GOPATH)"/bin/golangci-lint run 2>&1 | tee "logs/plan${PLAN_ID}/golangci-lint-${TS}.log"
test ${PIPESTATUS[0]} -eq 0

echo "🎭 Playwright E2E (DevServer; plan ${PLAN_ID})"
pushd frontend >/dev/null
if [ "${SKIP_INSTALL:-0}" != "1" ]; then
  npm ci
  npx playwright install --with-deps
fi
# 复刻 PR：允许通过环境变量注入 JWT 与 TENANT；PW_JWT 若未设置由配置自动读取 .cache/dev.jwt
export PW_TENANT_ID="${PW_TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
export PW_JWT="${PW_JWT:-$(cat ../.cache/dev.jwt 2>/dev/null || true)}"
E2E_PLAN_ID="${PLAN_ID}" PW_SKIP_SERVER="${PW_SKIP_SERVER:-0}" npm run -s test:e2e:plan
popd >/dev/null

echo "🧾 Print JSON SUMMARY (if any)"
node scripts/ci/print-e2e-summary.js "${PLAN_ID}" || true

echo "✅ Done. Artifacts in logs/plan${PLAN_ID}/*"
