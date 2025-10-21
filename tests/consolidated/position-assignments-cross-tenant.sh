#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<MSG
Usage: BASE_URL=... TENANT_A_TOKEN=... TENANT_B_TOKEN=... TENANT_A_ID=... TENANT_B_ID=... POSITION_CODE=... \
  $(basename "$0")

Required environment variables:
  BASE_URL           Base URL of the command service (e.g. https://api.example.com)
  TENANT_A_TOKEN     JWT for tenant A (expected to own the target position)
  TENANT_B_TOKEN     JWT for tenant B (should be denied)
  TENANT_A_ID        Tenant UUID associated with TENANT_A_TOKEN
  TENANT_B_ID        Tenant UUID associated with TENANT_B_TOKEN
  POSITION_CODE      Position code to exercise (format Pxxxxxxx)
Optional:
  API_VERSION        Defaults to v1
  ASSIGNMENT_ID      Optional assignmentId for close/update verification
  EMPLOYEE_ID        Override default employeeId (otherwise auto-generated)
  EMPLOYEE_NAME      Override default employeeName (default adds timestamp)
  EFFECTIVE_DATE     Override effectiveDate (default today-1)
  ACTING_UNTIL_DATE  Override actingUntil date (default +7 days)
  END_DATE           Override closing endDate (default equals actingUntil)

Dependencies: curl, jq
MSG
}

if [[ -z "${BASE_URL:-}" || -z "${TENANT_A_TOKEN:-}" || -z "${TENANT_B_TOKEN:-}" || -z "${TENANT_A_ID:-}" || -z "${TENANT_B_ID:-}" || -z "${POSITION_CODE:-}" ]]; then
  usage >&2
  exit 1
fi

API_VERSION="${API_VERSION:-v1}"
ASSIGNMENT_ID="${ASSIGNMENT_ID:-}"

DEFAULT_EMPLOYEE_ID=""
DEFAULT_EMPLOYEE_NAME=""
ACTING_UNTIL_DATE_VALUE=""
CLOSE_END_DATE_VALUE=""
EFFECTIVE_DATE_VALUE=""

generate_uuid() {
  if [[ -r /proc/sys/kernel/random/uuid ]]; then
    tr '[:upper:]' '[:lower:]' < /proc/sys/kernel/random/uuid
    return
  fi

  if command -v uuidgen >/dev/null 2>&1; then
    uuidgen | tr '[:upper:]' '[:lower:]'
    return
  fi

  if command -v python3 >/dev/null 2>&1; then
    python3 - <<'PY'
import uuid
print(uuid.uuid4())
PY
    return
  fi

  echo "无法生成 UUID: 缺少 uuidgen/python3 支持" >&2
  exit 1
}

future_date() {
  local days=$1
  if date -d "${days} days" +%Y-%m-%d >/dev/null 2>&1; then
    date -d "${days} days" +%Y-%m-%d
    return
  fi

  if command -v python3 >/dev/null 2>&1; then
    python3 - "$days" <<'PY'
import sys
from datetime import datetime, timedelta
days = int(sys.argv[1])
print((datetime.now() + timedelta(days=days)).strftime('%Y-%m-%d'))
PY
    return
  fi

  echo "无法计算未来日期: 缺少 date/python3 支持" >&2
  exit 1
}

default_payload() {
  cat <<JSON
{
  "employeeId": "$DEFAULT_EMPLOYEE_ID",
  "employeeName": "$DEFAULT_EMPLOYEE_NAME",
  "assignmentType": "ACTING",
  "fte": 0.1,
  "effectiveDate": "$EFFECTIVE_DATE_VALUE",
  "autoRevert": true,
  "actingUntil": "$ACTING_UNTIL_DATE_VALUE",
  "operationReason": "cross-tenant smoke"
}
JSON
}

target="${BASE_URL}/api/${API_VERSION}/positions/${POSITION_CODE}/assignments"

request() {
  local method=$1
  local url=$2
  local token=$3
  local tenant=$4
  local body=${5:-}

  local tmp
  tmp=$(mktemp)
  local status
  if [[ -n "$body" ]]; then
    status=$(curl -s -o "$tmp" -w '%{http_code}' \
      -X "$method" \
      -H "Authorization: Bearer $token" \
      -H "X-Tenant-ID: $tenant" \
      -H 'Content-Type: application/json' \
      "$url" \
      -d "$body")
  else
    status=$(curl -s -o "$tmp" -w '%{http_code}' \
      -X "$method" \
      -H "Authorization: Bearer $token" \
      -H "X-Tenant-ID: $tenant" \
      "$url")
  fi
  local body_out
  body_out=$(cat "$tmp")
  rm -f "$tmp"
  echo "$status" "$body_out"
}

log_step() {
  echo "[cross-tenant] $1"
}

log_step "Tenant B 尝试读取 (预期 403/404)"
read_resp=$(request GET "$target" "$TENANT_B_TOKEN" "$TENANT_B_ID")
status_b=${read_resp%% *}
if [[ "$status_b" != "403" && "$status_b" != "404" ]]; then
  echo "✖ Tenant B 读取未被拒绝 (状态 $status_b)"
  exit 1
fi
echo "✔ 读取被拒绝 (状态 $status_b)"

log_step "Tenant A 读取 (预期 200)"
read_resp=$(request GET "$target" "$TENANT_A_TOKEN" "$TENANT_A_ID")
status_a=${read_resp%% *}
if [[ "$status_a" != "200" ]]; then
  echo "✖ Tenant A 读取失败 (状态 $status_a)"
  exit 1
fi

if [[ -n "${EMPLOYEE_ID:-}" ]]; then
  DEFAULT_EMPLOYEE_ID="$EMPLOYEE_ID"
else
  DEFAULT_EMPLOYEE_ID="$(generate_uuid)"
fi

if [[ -n "${EMPLOYEE_NAME:-}" ]]; then
  DEFAULT_EMPLOYEE_NAME="$EMPLOYEE_NAME"
else
  DEFAULT_EMPLOYEE_NAME="跨租户验收 $(date +%Y%m%dT%H%M%S)"
fi

if [[ -n "${EFFECTIVE_DATE:-}" ]]; then
  EFFECTIVE_DATE_VALUE="$EFFECTIVE_DATE"
else
  EFFECTIVE_DATE_VALUE="$(future_date -1)"
fi

if [[ -n "${ACTING_UNTIL_DATE:-}" ]]; then
  ACTING_UNTIL_DATE_VALUE="$ACTING_UNTIL_DATE"
else
  ACTING_UNTIL_DATE_VALUE="$(future_date 6)"
fi

if [[ -n "${END_DATE:-}" ]]; then
  CLOSE_END_DATE_VALUE="$END_DATE"
else
  CLOSE_END_DATE_VALUE="$ACTING_UNTIL_DATE_VALUE"
fi

default_body=$(default_payload)
log_step "使用租户 A 员工 ${DEFAULT_EMPLOYEE_ID} (${DEFAULT_EMPLOYEE_NAME})"

log_step "Tenant B 尝试创建 (预期 403/404)"
create_resp=$(request POST "$target" "$TENANT_B_TOKEN" "$TENANT_B_ID" "$default_body")
status_create_b=${create_resp%% *}
if [[ "$status_create_b" != "403" && "$status_create_b" != "404" ]]; then
  echo "✖ Tenant B 创建未被拒绝 (状态 $status_create_b)"
  exit 1
fi
echo "✔ 创建被拒绝 (状态 $status_create_b)"

log_step "Tenant A 创建任职 (预期 201)"
create_resp=$(request POST "$target" "$TENANT_A_TOKEN" "$TENANT_A_ID" "$default_body")
status_create_a=${create_resp%% *}
body_create_a=${create_resp#* }
if [[ "$status_create_a" != "201" ]]; then
  echo "✖ Tenant A 创建失败 (状态 $status_create_a)"
  echo "$body_create_a"
  exit 1
fi
created_assignment=$(echo "$body_create_a" | jq -r '.data.assignmentId // empty')
ASSIGNMENT_ID=${ASSIGNMENT_ID:-$created_assignment}
if [[ -z "$ASSIGNMENT_ID" ]]; then
  echo "✖ 无法解析创建后的 assignmentId"
  exit 1
fi
echo "✔ 创建成功 assignmentId=$ASSIGNMENT_ID"

close_url="$target/${ASSIGNMENT_ID}/close"
close_body="{\"endDate\":\"${CLOSE_END_DATE_VALUE}\",\"operationReason\":\"cross-tenant smoke\"}"

log_step "Tenant B 尝试关闭 (预期 403/404)"
close_resp=$(request POST "$close_url" "$TENANT_B_TOKEN" "$TENANT_B_ID" "$close_body")
status_close_b=${close_resp%% *}
if [[ "$status_close_b" != "403" && "$status_close_b" != "404" ]]; then
  echo "✖ Tenant B 关闭未被拒绝 (状态 $status_close_b)"
  exit 1
fi
echo "✔ 关闭被拒绝 (状态 $status_close_b)"

log_step "Tenant A 关闭 (预期 200)"
close_resp=$(request POST "$close_url" "$TENANT_A_TOKEN" "$TENANT_A_ID" "$close_body")
status_close_a=${close_resp%% *}
body_close_a=${close_resp#* }
if [[ "$status_close_a" != "200" ]]; then
  echo "✖ Tenant A 关闭失败 (状态 $status_close_a)"
  echo "$body_close_a"
  exit 1
fi
echo "✔ Tenant A 成功关闭"

log_step "验收完成"
echo "所有跨租户校验通过"
