#!/usr/bin/env bash

set -euo pipefail

# Plan 219E – REST 接口性能基准脚本
# 使用 hey (https://github.com/rakyll/hey) 对指定端点进行并发压测

COMMAND_API="${COMMAND_API:-http://localhost:9090}"
TARGET_PATH="${TARGET_PATH:-/api/v1/organization-units}"
REQUEST_BODY="${REQUEST_BODY:-}"
METHOD="${METHOD:-GET}"
CONCURRENCY="${CONCURRENCY:-25}"
DURATION="${DURATION:-15s}"
TENANT_ID="${TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
LOG_DIR="${LOG_DIR:-logs/219E}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
LOG_FILE="${LOG_DIR}/perf-rest-${TIMESTAMP}.log"

mkdir -p "${LOG_DIR}" .cache

if ! command -v hey >/dev/null 2>&1; then
  cat <<'EOF'
❌ 未找到 hey 命令。
请先安装：
  go install github.com/rakyll/hey@latest
并将 ~/go/bin 加入 PATH。
EOF
  exit 1
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "❌ 缺少依赖: $1"
    exit 1
  fi
}

require_cmd curl
require_cmd jq

TOKEN="${JWT_TOKEN:-}"

if [[ -z "${TOKEN}" && -f .cache/dev.jwt ]]; then
  TOKEN="$(< .cache/dev.jwt)"
fi

if [[ -z "${TOKEN}" ]]; then
  payload='{"userId":"perf-bot","tenantId":"'"${TENANT_ID}"'","roles":["ADMIN","USER"],"duration":"1h"}'
  response=$(curl -sS -X POST "${COMMAND_API}/auth/dev-token" \
    -H "Content-Type: application/json" \
    -d "${payload}" 2>>"${LOG_FILE}" || true)
  TOKEN="$(echo "${response}" | jq -r '.token // empty')"
fi

if [[ -z "${TOKEN}" ]]; then
  echo "❌ 无法获取 JWT。请通过 make jwt-dev-mint 或设置 JWT_TOKEN 后重试。" | tee -a "${LOG_FILE}"
  exit 1
fi

echo "🌐 目标: ${COMMAND_API}${TARGET_PATH}" | tee "${LOG_FILE}"
echo "⚙️  并发: ${CONCURRENCY}  持续: ${DURATION}  方法: ${METHOD}" | tee -a "${LOG_FILE}"

AUTH_HEADER="Authorization: Bearer ${TOKEN}"
TENANT_HEADER="X-Tenant-ID: ${TENANT_ID}"

HEY_ARGS=(-c "${CONCURRENCY}" -z "${DURATION}" -m "${METHOD}" -H "${AUTH_HEADER}" -H "${TENANT_HEADER}")

if [[ -n "${REQUEST_BODY}" ]]; then
  HEY_ARGS+=(-T "application/json" -d "${REQUEST_BODY}")
fi

hey "${HEY_ARGS[@]}" "${COMMAND_API}${TARGET_PATH}" | tee -a "${LOG_FILE}"

echo ""
echo "📄 结果日志: ${LOG_FILE}"
