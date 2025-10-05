#!/usr/bin/env bash
set -euo pipefail

# Plan 18 Phase 1.3 — 本地迁移 + 业务流程 E2E 校验脚本
# 依赖: bash, git, make, curl, python3, npm, psql, docker-compose

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR"

LOG_DIR="$ROOT_DIR/reports/iig-guardian"
mkdir -p "$LOG_DIR"

TIMESTAMP="$(date +%Y%m%dT%H%M%S)"
MIGRATION_LOG="$LOG_DIR/plan18-migration-$TIMESTAMP.log"
E2E_LOG="$LOG_DIR/plan18-business-flow-$TIMESTAMP.log"

DEFAULT_TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
TENANT_ID="${PW_TENANT_ID:-$DEFAULT_TENANT_ID}"

command -v make >/dev/null 2>&1 || { echo "❌ 需要 make"; exit 1; }
command -v curl >/dev/null 2>&1 || { echo "❌ 需要 curl"; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "❌ 需要 python3"; exit 1; }
command -v npm >/dev/null 2>&1 || { echo "❌ 需要 npm"; exit 1; }
command -v psql >/dev/null 2>&1 || { echo "❌ 需要 psql"; exit 1; }

echo "[Plan18] 步骤 1/5: 启动最小依赖 (postgres, redis)"
make docker-up

echo "[Plan18] 步骤 2/5: 执行数据库迁移"
set -o pipefail
make db-migrate-all | tee "$MIGRATION_LOG"
set +o pipefail

echo "[Plan18] 步骤 3/5: 确认命令/查询服务健康"

function wait_health() {
  local url="$1"
  local name="$2"
  for i in {1..20}; do
    if curl -sf "$url" >/dev/null; then
      echo "  ✅ $name 健康"
      return 0
    fi
    sleep 1
  done
  echo "❌ $name 未就绪: $url" >&2
  exit 2
}

wait_health "http://localhost:9090/health" "command-service"
wait_health "http://localhost:8090/health" "query-service"

echo "[Plan18] 步骤 4/5: 调用 /auth/dev-token 生成 RS256 JWT"

BODY=$(python3 - <<'PY'
import json
import os

payload = {
    "userId": os.environ.get("USER_ID", "dev-user"),
    "tenantId": os.environ.get("TENANT_ID", "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"),
    "roles": os.environ.get("ROLES", "ADMIN,USER").split(','),
    "duration": os.environ.get("DURATION", "8h"),
}

print(json.dumps(payload, ensure_ascii=False))
PY
)

RESP=$(curl -sf -X POST http://localhost:9090/auth/dev-token -H 'Content-Type: application/json' -d "$BODY")

TOKEN=$(RESP_JSON="$RESP" python3 - <<'PY'
import base64
import json
import os

resp_raw = os.environ.get("RESP_JSON", "")
if not resp_raw:
    raise SystemExit("生成失败: 响应为空")

data = json.loads(resp_raw)
if not data.get("success"):
    raise SystemExit(data.get('error', {}).get('message') or data.get('message') or '未知错误')

token = (data.get("data") or {}).get("token")
if not token:
    raise SystemExit("生成失败: 响应中缺少 token 字段")

header_b64 = token.split('.')[0]
header_json = base64.urlsafe_b64decode(header_b64 + '=' * (-len(header_b64) % 4)).decode('utf-8')
header = json.loads(header_json)

if header.get("alg") != "RS256":
    raise SystemExit(f"令牌签名算法不匹配: {header.get('alg')}")

print(token)
PY
)

echo "$TOKEN" > .cache/dev.jwt
echo "  ✅ 令牌已保存到 .cache/dev.jwt"

echo "[Plan18] 步骤 5/5: 执行业务流程 E2E (Playwright)"

cd "$ROOT_DIR/frontend"
set -o pipefail
PW_JWT="$TOKEN" PW_TENANT_ID="$TENANT_ID" npm run test:e2e -- tests/e2e/business-flow-e2e.spec.ts | tee "$E2E_LOG"
EXIT_CODE=${PIPESTATUS[0]}
set +o pipefail

echo ""
echo "📄 迁移日志: $MIGRATION_LOG"
echo "📄 E2E 日志: $E2E_LOG"

exit "$EXIT_CODE"
