#!/usr/bin/env bash
# Plan 222 Acceptance Runner (health/JWKS/REST/GraphQL)
# - Does not start services; assumes `make docker-up && make run-dev` already running
# - Writes evidence to logs/plan222 without fabricating data
# - Honors AGENTS.md: Docker-only services; no host DB/Redis
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="${ROOT_DIR}/logs/plan222"
mkdir -p "${LOG_DIR}"

ts() { date -u +"%Y%m%d-%H%M%S"; }
STAMP="$(ts)"

REST_BASE="${REST_BASE:-http://localhost:9090}"
# 按 SSoT（CQRS）默认：GraphQL 查询服务运行在 8090；如启用“单体挂载 /graphql(9090)”仅作历史兼容，非默认
GRAPHQL_BASE="${GRAPHQL_BASE:-http://localhost:8090}"
JWT_FILE="${ROOT_DIR}/.cache/dev.jwt"
ORG_PARENT_CODE="${ORG_PARENT_CODE:-1000000}"
# 根组织名称（仅在缺失时用于引导创建，不改变既有数据）
ORG_PARENT_NAME="${ORG_PARENT_NAME:-飞虫与鲜花}"

echo "[Plan222] Collecting health/JWKS..."
set +e
curl -fsS "${REST_BASE}/health" -o "${LOG_DIR}/health-command-${STAMP}.json"
rc_health=$?
curl -fsS "${REST_BASE}/.well-known/jwks.json" -o "${LOG_DIR}/jwks-${STAMP}.json"
rc_jwks=$?
set -e
echo "[Plan222] health=${rc_health} jwks=${rc_jwks}"

# GraphQL smoke probe
echo "[Plan222] GraphQL probe → ${GRAPHQL_BASE}/graphql"
GQL_QUERY='{"query":"{ organizations { data { code name parentCode status } pagination { page pageSize total } } }"}'
set +e
curl -fsS -H "Content-Type: application/json" -X POST \
  --data "${GQL_QUERY}" \
  "${GRAPHQL_BASE}/graphql" \
  -o "${LOG_DIR}/graphql-query-${STAMP}.json"
rc_gql=$?
set -e
echo "[Plan222] GraphQL probe (unauth) rc=${rc_gql} (0=OK, 22/35=connection)"

# REST create + PUT (optional if JWT present)
AUTH_HEADER=()
if [[ -f "${JWT_FILE}" ]]; then
  JWT="$(cat "${JWT_FILE}")"
  AUTH_HEADER=(-H "Authorization: Bearer ${JWT}")
  echo "[Plan222] Using JWT from .cache/dev.jwt"
  # Try to resolve tenant ID for X-Tenant-ID
  TENANT_ID="$(curl -s -H "Authorization: Bearer ${JWT}" "${REST_BASE}/auth/dev-token/info" | jq -r '.data.tenantId // empty')"
  if [[ -n "${TENANT_ID}" ]]; then
    echo "[Plan222] Resolved X-Tenant-ID=${TENANT_ID}"
    TENANT_HEADER=(-H "X-Tenant-ID: ${TENANT_ID}")
  else
    echo "[Plan222] WARN: Could not resolve tenantId from token info; GraphQL/REST may return 401/403 without X-Tenant-ID"
    TENANT_HEADER=()
  fi
  # Authenticated GraphQL probe (best-effort)
  set +e
  curl -fsS -H "Content-Type: application/json" "${AUTH_HEADER[@]}" "${TENANT_HEADER[@]}" \
    -X POST --data "${GQL_QUERY}" \
    "${GRAPHQL_BASE}/graphql" \
    -o "${LOG_DIR}/graphql-query-${STAMP}-auth.json"
  rc_gql_auth=$?
  set -e
  echo "[Plan222] GraphQL probe (auth) rc=${rc_gql_auth}"
else
  echo "[Plan222] JWT not found (.cache/dev.jwt). Skipping write operations. Run: make jwt-dev-setup && make jwt-dev-mint"
fi

# Create
if [[ ${#AUTH_HEADER[@]} -gt 0 ]]; then
  # 确保默认上级组织存在（如缺失则引导创建一个根组织）
  echo "[Plan222] Ensure parent org exists: ${ORG_PARENT_CODE}"
  PARENT_STATUS_CODE=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Accept: application/json" "${AUTH_HEADER[@]}" "${TENANT_HEADER[@]}" \
    "${REST_BASE}/api/v1/organization-units/${ORG_PARENT_CODE}")
  if [[ "${PARENT_STATUS_CODE}" != "200" ]]; then
    echo "[Plan222] Parent not found (status=${PARENT_STATUS_CODE}), bootstrap root ${ORG_PARENT_CODE}"
    ROOT_CREATE_PAYLOAD=$(cat <<EOF
{"code":"${ORG_PARENT_CODE}","name":"${ORG_PARENT_NAME}-${STAMP}","unitType":"DEPARTMENT","parentCode":null,"description":"Plan222 bootstrap root"}
EOF
)
    curl -sS -D "${LOG_DIR}/root-create-headers-${STAMP}.txt" \
      -H "Content-Type: application/json" "${AUTH_HEADER[@]}" "${TENANT_HEADER[@]}" \
      -X POST "${REST_BASE}/api/v1/organization-units" \
      --data "${ROOT_CREATE_PAYLOAD}" \
      -o "${LOG_DIR}/root-create-response-${STAMP}.json" \
      -w "%{http_code}\n" > "${LOG_DIR}/root-create-status-${STAMP}.txt" || true
  else
    echo "[Plan222] Parent exists: ${ORG_PARENT_CODE}"
  fi

  # 7-digit code, first digit non-zero (ORG_CODE rule)
  CODE="$((RANDOM%9000000 + 1000000))"
  CREATE_PAYLOAD=$(cat <<EOF
{"code":"${CODE}","name":"Plan222 验收-${STAMP}","unitType":"DEPARTMENT","parentCode":"1000000","description":"Plan222 acceptance"}
EOF
)
  echo "[Plan222] REST create org ${CODE}"
  curl -sS -D "${LOG_DIR}/create-headers-${STAMP}.txt" \
    -H "Content-Type: application/json" "${AUTH_HEADER[@]}" "${TENANT_HEADER[@]}" \
    -X POST "${REST_BASE}/api/v1/organization-units" \
    --data "${CREATE_PAYLOAD}" \
    -o "${LOG_DIR}/create-response-${STAMP}.json" \
    -w "%{http_code}\n" > "${LOG_DIR}/create-status-${STAMP}.txt" || true

  # PUT update (best-effort)
  UPDATE_PAYLOAD=$(cat <<EOF
{"name":"Plan222 验收(已更新)-${STAMP}","sortOrder":2,"description":"Plan222 update"}
EOF
)
  echo "[Plan222] REST update org ${CODE}"
  curl -sS \
    -H "Content-Type: application/json" "${AUTH_HEADER[@]}" "${TENANT_HEADER[@]}" \
    -X PUT "${REST_BASE}/api/v1/organization-units/${CODE}" \
    --data "${UPDATE_PAYLOAD}" \
    -o "${LOG_DIR}/put-response-${CODE}.json" \
    -w "%{http_code}\n" > "${LOG_DIR}/put-status-${CODE}.txt" || true
fi

echo "[Plan222] Done. Outputs in logs/plan222/*-${STAMP}*"
