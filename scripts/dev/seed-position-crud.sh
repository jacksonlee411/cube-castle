#!/usr/bin/env bash
set -euo pipefail

COMMAND_BASE_URL="${COMMAND_BASE_URL:-http://localhost:9090}"
TENANT_ID="${TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
ORGANIZATION_CODE="${ORGANIZATION_CODE:-1000000}"
JOB_FAMILY_GROUP="${JOB_FAMILY_GROUP:-OPER}"
JOB_FAMILY="${JOB_FAMILY:-OPER-OPS}"
JOB_ROLE="${JOB_ROLE:-OPER-OPS-MGR}"
JOB_LEVEL="${JOB_LEVEL:-S1}"
EFFECTIVE_DATE="${EFFECTIVE_DATE:-2025-01-01}"
ASSIGNMENT_EFFECTIVE_DATE="${ASSIGNMENT_EFFECTIVE_DATE:-2025-02-01}"
VACATE_EFFECTIVE_DATE="${VACATE_EFFECTIVE_DATE:-2025-03-01}"
TOKEN_FILE="${TOKEN_FILE:-.cache/dev.jwt}"
LOG_DIR="${LOG_DIR:-logs/230}"
RUN_ID="$(date +%Y%m%dT%H%M%S)"
LOG_PATH="${LOG_DIR}/position-seed-${RUN_ID}.log"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "❌ 需要命令 $1，请先安装" >&2
    exit 1
  fi
}

require_cmd curl
require_cmd jq

if [[ ! -f "${TOKEN_FILE}" ]]; then
  echo "❌ 未找到 ${TOKEN_FILE}，请运行 'make jwt-dev-setup && make jwt-dev-mint' 后重试" >&2
  exit 1
fi

JWT="$(tr -d '\n' < "${TOKEN_FILE}")"

mkdir -p "${LOG_DIR}"

log() {
  printf '%s\n' "$*" | tee -a "${LOG_PATH}" >&2
}

call_api() {
  local method="$1"; shift
  local path="$1"; shift
  local payload="${1:-}"
  local tmp_body
  tmp_body="$(mktemp)"

  log ""
  log "=== ${method} ${path} ==="
  if [[ -n "${payload}" ]]; then
    log "Payload: ${payload}"
  fi

  http_status="$(curl -sS -o "${tmp_body}" -w "%{http_code}" \
    -X "${method}" \
    -H "Authorization: Bearer ${JWT}" \
    -H "X-Tenant-ID: ${TENANT_ID}" \
    -H "Content-Type: application/json" \
    -H "X-Idempotency-Key: position-seed-${RUN_ID}-${method}" \
    --data "${payload}" \
    "${COMMAND_BASE_URL}${path}")"

  body="$(cat "${tmp_body}")"
  rm -f "${tmp_body}"

  log "HTTP ${http_status}"
  if [[ -n "${body}" ]]; then
    echo "${body}" | jq '.' >>"${LOG_PATH}" 2>/dev/null || echo "${body}" >>"${LOG_PATH}"
  fi

  if [[ "${http_status}" -ge 400 ]]; then
    echo "❌ 请求 ${method} ${path} 失败 (HTTP ${http_status})，详见 ${LOG_PATH}" >&2
    exit 1
  fi

  printf '%s\n' "${body}"
}

uniq_suffix="$(date +%s)"
position_title="230-seed-position-${uniq_suffix}"
create_payload="$(jq -n \
  --arg title "${position_title}" \
  --arg jfg "${JOB_FAMILY_GROUP}" \
  --arg jf "${JOB_FAMILY}" \
  --arg jr "${JOB_ROLE}" \
  --arg jl "${JOB_LEVEL}" \
  --arg org "${ORGANIZATION_CODE}" \
  --arg eff "${EFFECTIVE_DATE}" \
  --arg reason "230-position-seed-${RUN_ID}" \
  '{
    title:$title,
    jobFamilyGroupCode:$jfg,
    jobFamilyCode:$jf,
    jobRoleCode:$jr,
    jobLevelCode:$jl,
    organizationCode:$org,
    positionType:"REGULAR",
    employmentType:"FULL_TIME",
    headcountCapacity:1.0,
    effectiveDate:$eff,
    operationReason:$reason
  }')"

create_resp="$(call_api POST "/api/v1/positions" "${create_payload}")"
position_code="$(echo "${create_resp}" | jq -r '.data.code // empty')"

if [[ -z "${position_code}" ]]; then
  echo "❌ 无法解析创建职位响应中的 code，详见 ${LOG_PATH}" >&2
  exit 1
fi

log "✅ 职位创建成功，code=${position_code}"

fill_payload="$(jq -n \
  --arg eff "${ASSIGNMENT_EFFECTIVE_DATE}" \
  '{
    employeeId:"00000000-0000-0000-0000-00000000a001",
    employeeName:"Position Seed Bot",
    assignmentType:"PRIMARY",
    effectiveDate:$eff,
    operationReason:"230-position-crud-seed",
    fte:1.0
  }')"

fill_resp="$(call_api POST "/api/v1/positions/${position_code}/fill" "${fill_payload}")"
assignment_id="$(echo "${fill_resp}" | jq -r '.data.currentAssignment.assignmentId // .data.assignmentHistory[0].assignmentId // empty')"

if [[ -z "${assignment_id}" ]]; then
  echo "❌ 无法解析 fill 响应中的 assignmentId，详见 ${LOG_PATH}" >&2
  exit 1
fi

log "✅ 职位填充成功 assignmentId=${assignment_id}"

vacate_payload="$(jq -n \
  --arg assign "${assignment_id}" \
  --arg eff "${VACATE_EFFECTIVE_DATE}" \
  --arg reason "230-position-crud-seed" \
  '{assignmentId:$assign,effectiveDate:$eff,operationReason:$reason}')"
call_api POST "/api/v1/positions/${position_code}/vacate" "${vacate_payload}" >/dev/null
log "✅ 职位 vacate 完成，保持 timeline 可重复使用"

log ""
log "🎉 Position CRUD 数据播种完成，日志: ${LOG_PATH}"
