#!/usr/bin/env bash
set -euo pipefail

CMD_BASE="${CMD_BASE:-http://localhost:9090}"
GQL_BASE="${GQL_BASE:-http://localhost:8090}"
TMP_DIR="${TMP_DIR:-.cache}"
COOKIE_JAR="$TMP_DIR/auth_cookies.txt"
mkdir -p "$TMP_DIR"

echo "[1/5] 清理旧Cookie..."
rm -f "$COOKIE_JAR" || true

echo "[2/5] 发起登录 (可能跳转 IdP 或模拟会话)..."
curl -sS -c "$COOKIE_JAR" -b "$COOKIE_JAR" -L \
  "$CMD_BASE/auth/login?redirect=%2F" >/dev/null || true

echo "[3/5] 获取会话 /auth/session ..."
SESSION_JSON=$(curl -sS -c "$COOKIE_JAR" -b "$COOKIE_JAR" "$CMD_BASE/auth/session" || true)
echo "$SESSION_JSON" | sed 's/.\{200\}$/.../'

# 解析 accessToken 与 tenantId
if command -v jq >/dev/null 2>&1; then
  ACCESS_TOKEN=$(echo "$SESSION_JSON" | jq -r '.data.accessToken // .accessToken')
  TENANT_ID=$(echo "$SESSION_JSON" | jq -r '.data.tenantId // .tenantId')
else
  ACCESS_TOKEN=$(echo "$SESSION_JSON" | sed -n 's/.*"accessToken"\s*:\s*"\([^"]*\)".*/\1/p' | head -n1)
  TENANT_ID=$(echo "$SESSION_JSON" | sed -n 's/.*"tenantId"\s*:\s*"\([^"]*\)".*/\1/p' | head -n1)
fi

if [ -z "${ACCESS_TOKEN:-}" ]; then
  echo "❌ 未获取到 accessToken；请确认服务已启动且 OIDC_SIMULATE=true 或已完成登录"
  exit 2
fi

echo "[4/5] GraphQL 请求（携带 Authorization 与 X-Tenant-ID=$TENANT_ID）..."
QUERY='{"query":"query { __typename }"}'
HTTP_CODE=$(curl -sS -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: ${TENANT_ID:-missing}" \
  -d "$QUERY" "$GQL_BASE/graphql")

echo "  → HTTP $HTTP_CODE"
if [ "$HTTP_CODE" != "200" ]; then
  echo "❌ GraphQL 请求失败；请检查 JWT/JWKS 配置与服务日志"
  exit 3
fi

echo "[5/5] 刷新短期 access token ..."
CSRF=$(grep -o 'csrf\s*\t[^\t]*' "$COOKIE_JAR" | awk '{print $2}' | tail -n1 || true)
HTTP_CODE=$(curl -sS -o /dev/null -w "%{http_code}" -X POST \
  -H "X-CSRF-Token: ${CSRF:-}" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  "$CMD_BASE/auth/refresh")
echo "  → HTTP $HTTP_CODE"
if [ "$HTTP_CODE" != "200" ]; then
  echo "❌ 刷新失败；会话可能已过期"
  exit 4
fi

echo "✅ 认证联调流程通过"

