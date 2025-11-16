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
GRAPHQL_BASE="${GRAPHQL_BASE:-http://localhost:9090}"
JWT_FILE="${ROOT_DIR}/.cache/dev.jwt"

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
  # 7-digit code, first digit non-zero (ORG_CODE rule)
  CODE="$((RANDOM%9000000 + 1000000))"
  CREATE_PAYLOAD=$(cat <<EOF
{"code":"${CODE}","name":"Plan222 验收-${STAMP}","unitType":"DEPARTMENT","parentCode":"1000000","description":"Plan222 acceptance"}
EOF
)
  echo "[Plan222] REST create org ${CODE}"
  curl -fsS -D "${LOG_DIR}/create-headers-${STAMP}.txt" \
    -H "Content-Type: application/json" "${AUTH_HEADER[@]}" "${TENANT_HEADER[@]}" \
    -X POST "${REST_BASE}/api/v1/organization-units" \
    --data "${CREATE_PAYLOAD}" \
    -o "${LOG_DIR}/create-response-${STAMP}.json" || true

  # PUT update (best-effort)
  UPDATE_PAYLOAD=$(cat <<EOF
{"name":"Plan222 验收(已更新)-${STAMP}","sortOrder":2,"description":"Plan222 update"}
EOF
)
  echo "[Plan222] REST update org ${CODE}"
  curl -fsS \
    -H "Content-Type: application/json" "${AUTH_HEADER[@]}" "${TENANT_HEADER[@]}" \
    -X PUT "${REST_BASE}/api/v1/organization-units/${CODE}" \
    --data "${UPDATE_PAYLOAD}" \
    -o "${LOG_DIR}/put-response-${CODE}.json" || true
fi

echo "[Plan222] Done. Outputs in logs/plan222/*-${STAMP}*"
