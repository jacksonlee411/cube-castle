#!/usr/bin/env bash
set -euo pipefail

# Temporal regression script for Organization Units
# Covers: create version, update effectiveDate, deactivate middle version, suspend/activate

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

BASE=${BASE:-http://localhost:9090}
TENANT=${TENANT:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}
CODE=${CODE:-1000002}

psql_exec() {
  source .env >/dev/null 2>&1 || true
  local URL=${DATABASE_URL_HOST:-postgresql://user:password@localhost:5432/cubecastle?sslmode=disable}
  URL=$(printf "%s" "$URL" | tr -d '\r')
  PGPASSWORD=$(printf "%s" "$URL" | sed -n 's/.*:\/\/(.*):\(.*\)@.*/\2/p' 2>/dev/null || true) psql "$URL" -v ON_ERROR_STOP=1 "$@"
}

ensure_token() {
  if [[ -f ./.cache/dev.jwt ]]; then
    JWT_TOKEN=$(cat ./.cache/dev.jwt)
  else
    echo "Minting dev token..."
    make jwt-dev-mint >/dev/null
    JWT_TOKEN=$(cat ./.cache/dev.jwt)
  fi
}

http_json() {
  local method=$1 path=$2 body=${3:-}
  if [[ -n "$body" ]]; then
    curl -s -X "$method" "$BASE$path" \
      -H "Authorization: Bearer $JWT_TOKEN" \
      -H 'Content-Type: application/json' \
      -H "X-Tenant-ID: $TENANT" \
      -d "$body"
  else
    curl -s -X "$method" "$BASE$path" \
      -H "Authorization: Bearer $JWT_TOKEN" \
      -H "X-Tenant-ID: $TENANT"
  fi
}

get_rid_by_date() {
  local date="$1"
  psql_exec -t -c "SELECT record_id FROM organization_units WHERE code='${CODE}' AND effective_date='${date}'::date AND status<>'DELETED' ORDER BY created_at DESC LIMIT 1;" | tr -d ' \n'
}

echo "== Temporal Regression for code=${CODE} =="
ensure_token

echo "-- Create version @ 2025-09-10"
http_json POST "/api/v1/organization-units/${CODE}/versions" '{
  "name":"产品部","unitType":"DEPARTMENT","parentCode":"1000000",
  "description":"回归-版本创建","effectiveDate":"2025-09-10","operationReason":"回归测试-创建版本"
}' | jq .

sleep 0.3

echo "-- Update effectiveDate: 2025-09-06 -> 2025-09-04 (if exists)"
RID=$(get_rid_by_date 2025-09-06 || true)
if [[ -n "$RID" ]]; then
  http_json PUT "/api/v1/organization-units/${CODE}/history/${RID}" '{
    "effectiveDate":"2025-09-04","changeReason":"回归测试-调整生效日"}' | jq .
else
  echo "skip: rid at 2025-09-06 not found"
fi

sleep 0.2

echo "-- Deactivate version @ 2025-09-08 (if exists)"
RID_DEL=$(get_rid_by_date 2025-09-08 || true)
if [[ -n "$RID_DEL" ]]; then
  http_json POST "/api/v1/organization-units/${CODE}/events" \
    "{\"eventType\":\"DEACTIVATE\",\"recordId\":\"$RID_DEL\",\"effectiveDate\":\"2025-09-10\",\"changeReason\":\"回归测试-作废中间版本\"}" | jq .
else
  echo "skip: rid at 2025-09-08 not found"
fi

echo "-- Suspend @ 2025-09-11"
http_json POST "/api/v1/organization-units/${CODE}/suspend" '{
  "effectiveDate":"2025-09-11","operationReason":"回归测试-暂停"}' | jq .

echo "-- Activate @ 2025-09-13"
http_json POST "/api/v1/organization-units/${CODE}/activate" '{
  "effectiveDate":"2025-09-13","operationReason":"回归测试-激活"}' | jq .

echo "-- Timeline (non-deleted)"
psql_exec -c "SELECT code,effective_date,end_date,is_current,status FROM organization_units WHERE code='${CODE}' AND status<>'DELETED' ORDER BY effective_date;"

echo "== Done =="

