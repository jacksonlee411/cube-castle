#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'MSG'
Usage: TENANT_A_TOKEN=... TENANT_B_TOKEN=... TENANT_A_ID=... TENANT_B_ID=... POSITION_CODE=... \
       [GRAPHQL_URL=...] $(basename "$0")

Required environment variables:
  TENANT_A_TOKEN   JWT for tenant A (expected owner of the position)
  TENANT_B_TOKEN   JWT for tenant B (should not see tenant A assignments)
  TENANT_A_ID      Tenant UUID associated with TENANT_A_TOKEN
  TENANT_B_ID      Tenant UUID associated with TENANT_B_TOKEN
  POSITION_CODE    Target position code (e.g. P9000003)

Optional:
  GRAPHQL_URL      GraphQL endpoint (default: http://localhost:8090/graphql)
  API_TIMEOUT      curl timeout in seconds (default: 15)

Dependencies: curl, jq
MSG
}

missing_env() {
  echo "✖ Missing required environment variables" >&2
  usage >&2
  exit 1
}

GRAPHQL_URL=${GRAPHQL_URL:-http://localhost:8090/graphql}
API_TIMEOUT=${API_TIMEOUT:-15}

if [[ -z "${TENANT_A_TOKEN:-}" || -z "${TENANT_B_TOKEN:-}" || -z "${TENANT_A_ID:-}" || -z "${TENANT_B_ID:-}" || -z "${POSITION_CODE:-}" ]]; then
  missing_env
fi

log_step() {
  printf '[graphql-cross-tenant] %s\n' "$1"
}

graphql_payload_assignments() {
  local code=$1
  local query
  read -r -d '' query <<'GQL'
query($code: PositionCode!) {
  positionAssignments(positionCode: $code) {
    totalCount
    data {
      assignmentId
      assignmentStatus
      actingUntil
      autoRevert
      effectiveDate
    }
  }
}
GQL
  jq -n --arg query "$query" --arg code "$code" '{query: $query, variables: {code: $code}}'
}

graphql_ping_payload() {
  jq -n '{query: "{ __typename }"}'
}

graphql_request() {
  local token=$1
  local tenant=$2
  local payload=$3

  local tmp
  tmp=$(mktemp)
  local status
  status=$(curl -sS -o "$tmp" -w '%{http_code}' \
    -X POST "$GRAPHQL_URL" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $token" \
    -H "X-Tenant-ID: $tenant" \
    --data "$payload" \
    --max-time "$API_TIMEOUT")

  local body
  body=$(cat "$tmp")
  rm -f "$tmp"
  printf '%s %s' "$status" "$body"
}

expect_status() {
  local expected=$1
  local actual=$2
  local context=$3
  if [[ "$expected" != "$actual" ]]; then
    printf '✖ %s (expected %s, got %s)\n' "$context" "$expected" "$actual" >&2
    exit 1
  fi
}

log_step "检查租户-Token 不匹配时的 403 Tenant Mismatch"
ping_payload=$(graphql_ping_payload)
response=$(graphql_request "$TENANT_A_TOKEN" "$TENANT_B_ID" "$ping_payload")
status=${response%% *}
body=${response#* }
expect_status 403 "$status" "Tenant mismatch should return 403"
if [[ $(echo "$body" | jq -r '.error.code' 2>/dev/null) != "TENANT_MISMATCH" ]]; then
  echo "✖ Unexpected error body for tenant mismatch" >&2
  echo "$body" >&2
  exit 1
fi
echo "✔ Tenant mismatch 返回 403"

log_step "租户 B 查询 positionAssignments（预期 totalCount=0）"
assign_payload=$(graphql_payload_assignments "$POSITION_CODE")
response=$(graphql_request "$TENANT_B_TOKEN" "$TENANT_B_ID" "$assign_payload")
status=${response%% *}
body=${response#* }
expect_status 200 "$status" "Tenant B query should succeed with 200"
total_count=$(echo "$body" | jq -r '.data.positionAssignments.totalCount')
if [[ "$total_count" != "0" ]]; then
  echo "✖ Tenant B 应得 totalCount=0，实际 $total_count" >&2
  echo "$body" >&2
  exit 1
fi
echo "✔ Tenant B 查询未暴露他租户数据"

log_step "租户 A 查询 positionAssignments（预期 totalCount>0）"
response=$(graphql_request "$TENANT_A_TOKEN" "$TENANT_A_ID" "$assign_payload")
status=${response%% *}
body=${response#* }
expect_status 200 "$status" "Tenant A query should succeed with 200"
total_count=$(echo "$body" | jq -r '.data.positionAssignments.totalCount')
if ! [[ "$total_count" =~ ^[0-9]+$ ]] || [[ "$total_count" -le 0 ]]; then
  echo "✖ Tenant A totalCount 应大于 0，实际 $total_count" >&2
  echo "$body" >&2
  exit 1
fi

first_entry=$(echo "$body" | jq '.data.positionAssignments.data[0]')
if [[ "$first_entry" == "null" ]]; then
  echo "✖ Tenant A 返回数据为空" >&2
  echo "$body" >&2
  exit 1
fi

if [[ $(echo "$first_entry" | jq -r '.effectiveDate // empty') == "" ]]; then
  echo "✖ 响应缺少 effectiveDate 字段" >&2
  echo "$body" >&2
  exit 1
fi
echo "✔ 租户 A 查询成功返回任职数据 (totalCount=$total_count)"

log_step "GraphQL 跨租户校验完成"
echo "所有 GraphQL 跨租户校验通过"
