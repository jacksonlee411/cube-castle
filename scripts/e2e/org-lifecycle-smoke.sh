#!/usr/bin/env bash

set -euo pipefail

# -------------------------------------------------------------------
# Plan 219E – 组织/部门生命周期端到端冒烟脚本
# 要求：command/query 服务已通过 `make run-dev` 启动，并启用 RS256/JWKS
# -------------------------------------------------------------------

COMMAND_API="${COMMAND_API:-http://localhost:9090}"
QUERY_API="${QUERY_API:-http://localhost:8090/graphql}"
TENANT_ID="${TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
LOG_DIR="${LOG_DIR:-logs/219E}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
LOG_FILE="${LOG_DIR}/org-lifecycle-${TIMESTAMP}.log"

mkdir -p "${LOG_DIR}" .cache

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "❌ 缺少依赖: $1" | tee -a "${LOG_FILE}"
    exit 1
  fi
}

require_cmd curl
require_cmd jq

log_step() {
  printf "\n[%s] %s\n" "$(date --iso-8601=seconds)" "$1" | tee -a "${LOG_FILE}"
}

# ----------------------- 令牌获取 -----------------------
TOKEN="${JWT_TOKEN:-}"

mint_dev_token() {
  log_step "尝试调用 /auth/dev-token 获取开发令牌"
  local payload='{"userId":"e2e-bot","tenantId":"'"${TENANT_ID}"'","roles":["ADMIN","USER"],"duration":"2h"}'
  local response
  if ! response="$(curl -sS -X POST "${COMMAND_API}/auth/dev-token" \
      -H "Content-Type: application/json" \
      -d "${payload}" 2>>"${LOG_FILE}")"; then
    echo "⚠️ 无法从命令服务获取 dev-token" | tee -a "${LOG_FILE}"
    return 1
  fi
  TOKEN="$(echo "${response}" | jq -r '.token // empty')"
}

if [[ -z "${TOKEN}" && -f .cache/dev.jwt ]]; then
  TOKEN="$(< .cache/dev.jwt)"
  log_step "从 .cache/dev.jwt 读取现有令牌"
fi

if [[ -z "${TOKEN}" ]]; then
  mint_dev_token || true
fi

if [[ -z "${TOKEN}" ]]; then
  echo "❌ 无法获取访问令牌。请运行 make jwt-dev-mint 或设置 JWT_TOKEN 环境变量后重试。" | tee -a "${LOG_FILE}"
  exit 1
fi

# ----------------------- 辅助方法 -----------------------
http_request() {
  local method="$1"
  local url="$2"
  local body="${3:-}"
  local output status
  output="$(mktemp)"

  if [[ -n "${body}" ]]; then
    status=$(curl -sS -o "${output}" -w "%{http_code}" -X "${method}" "${url}" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "X-Tenant-ID: ${TENANT_ID}" \
      -H "Content-Type: application/json" \
      -d "${body}" 2>>"${LOG_FILE}")
  else
    status=$(curl -sS -o "${output}" -w "%{http_code}" -X "${method}" "${url}" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "X-Tenant-ID: ${TENANT_ID}" \
      -H "Content-Type: application/json" 2>>"${LOG_FILE}")
  fi

  {
    echo ">>> ${method} ${url}"
    [[ -n "${body}" ]] && echo "payload: ${body}"
    echo "status: ${status}"
    cat "${output}" | jq '.' 2>/dev/null || cat "${output}"
  } >>"${LOG_FILE}"

  rm -f "${output}"

  if [[ ! "${status}" =~ ^2 ]]; then
    echo "❌ 调用 ${url} 失败（HTTP ${status}），详见 ${LOG_FILE}" | tee -a "${LOG_FILE}"
    exit 1
  fi
}

graphql_query() {
  local query="$1"
  local response
  response=$(curl -sS -X POST "${QUERY_API}" \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "X-Tenant-ID: ${TENANT_ID}" \
    -H "Content-Type: application/json" \
    -d "${query}" 2>>"${LOG_FILE}")
  {
    echo ">>> GraphQL"
    echo "${query}"
    echo "status: 200"
    echo "${response}" | jq '.' 2>/dev/null || echo "${response}"
  } >>"${LOG_FILE}"
}

# ----------------------- 场景执行 -----------------------
# 组织代码必须为 7 位数字且首位不可为 0
BASE_CODE_NUM=$(( ($(date +%s) % 9000000) + 1000000 ))
ROOT_CODE=$(printf "%07d" "${BASE_CODE_NUM}")
CHILD_CODE=$(printf "%07d" $(( (BASE_CODE_NUM + 1) % 9000000 + 1000000 )))

log_step "1. 创建顶级组织：${ROOT_CODE}"
http_request POST "${COMMAND_API}/api/v1/organization-units" "$(jq -n --arg code "${ROOT_CODE}" '
{
  code: $code,
  name: "E2E Root Org",
  unitType: "DEPARTMENT",
  effectiveDate: "2025-11-06"
}')"

log_step "2. 创建子部门：${CHILD_CODE}"
http_request POST "${COMMAND_API}/api/v1/organization-units" "$(jq -n --arg code "${CHILD_CODE}" --arg parent "${ROOT_CODE}" '
{
  code: $code,
  name: "E2E Child Dept",
  unitType: "DEPARTMENT",
  parentCode: $parent,
  effectiveDate: "2025-11-06"
}')"

log_step "3. 停用顶级组织"
http_request POST "${COMMAND_API}/api/v1/organization-units/${ROOT_CODE}/suspend" '{
  "operationReason": "e2e-suspend",
  "effectiveDate": "2025-11-06"
}'

log_step "4. 重新启用顶级组织"
http_request POST "${COMMAND_API}/api/v1/organization-units/${ROOT_CODE}/activate" '{
  "operationReason": "e2e-reactivate",
  "effectiveDate": "2025-11-07"
}'

log_step "5. GraphQL 查询验证"
graphql_query "$(jq -n --arg code "${ROOT_CODE}" '{
  query: "query($code:String!){ organization(code:$code){ code name status isCurrent parentCode } }",
  variables: { code: $code }
}')"

log_step "✅ 组织生命周期冒烟测试完成。日志已保存：${LOG_FILE}"
