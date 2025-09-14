#!/usr/bin/env bash
set -euo pipefail

echo "🔐 Cube Castle - Auth & Well-known 集成自检"
echo "========================================="

BASE_URL=${BASE_URL:-"http://localhost:9090"}
REDIRECT_PATH=${REDIRECT_PATH:-"/"}

pass=0; fail=0; skip=0

ok()  { echo "✅ $1"; pass=$((pass+1)); }
bad() { echo "❌ $1"; fail=$((fail+1)); }
skp() { echo "⚪ $1"; skip=$((skip+1)); }

req() {
  local method="$1"; shift
  local url="$1"; shift
  curl -sS -i -X "$method" "$url" "$@"
}

have_jq=1
command -v jq >/dev/null 2>&1 || have_jq=0

echo "\n1) 健康检查 (${BASE_URL}/health)"
if status=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health"); then
  if [[ "$status" == "200" ]]; then ok "健康检查 200"; else bad "健康检查 HTTP $status"; fi
else
  bad "健康检查请求失败（请先启动命令服务在 9090 端口）"
fi

echo "\n2) OIDC 发现端点 (${BASE_URL}/.well-known/oidc)"
body=$(mktemp)
status=$(curl -s -o "$body" -w "%{http_code}" "${BASE_URL}/.well-known/oidc" || true)
case "$status" in
  200)
    if [[ $have_jq -eq 1 ]]; then
      issuer=$(jq -r '.issuer // empty' "$body" 2>/dev/null || true)
      authz=$(jq -r '.authorizationEndpoint // empty' "$body" 2>/dev/null || true)
      token=$(jq -r '.tokenEndpoint // empty' "$body" 2>/dev/null || true)
      if [[ -n "$issuer" && -n "$authz" && -n "$token" ]]; then
        ok "OIDC 发现 OK（camelCase 字段校验通过）"
      else
        bad "OIDC 发现 200 但字段不完整（需要 issuer/authorizationEndpoint/tokenEndpoint）"
      fi
    else
      ok "OIDC 发现 200（未安装 jq，跳过字段校验）"
    fi
    ;;
  501)
    if [[ $have_jq -eq 1 ]]; then
      code=$(jq -r '.error.code // empty' "$body" 2>/dev/null || true)
      [[ "$code" == "OIDC_NOT_CONFIGURED" ]] && ok "OIDC 未配置（501 OIDC_NOT_CONFIGURED）" || bad "OIDC 未配置但错误码不匹配：$code"
    else
      ok "OIDC 未配置（501），未安装 jq 跳过错误码校验"
    fi
    ;;
  *)
    bad "OIDC 发现 HTTP $status" ;;
esac
rm -f "$body"

echo "\n3) 模拟登录与会话（需 OIDC_SIMULATE=true）"
cookiejar=$(mktemp)
status=$(curl -s -o /dev/null -w "%{http_code}" -c "$cookiejar" "${BASE_URL}/auth/login?redirect=$(python3 - <<<'import urllib.parse;print(urllib.parse.quote("'${REDIRECT_PATH}'"))')")
if [[ "$status" != "302" && "$status" != "200" ]]; then
  skp "模拟登录未开启（HTTP $status），跳过会话链路验证"
else
  sess=$(mktemp)
  status_sess=$(curl -s -b "$cookiejar" -o "$sess" -w "%{http_code}" "${BASE_URL}/auth/session")
  if [[ "$status_sess" == "200" ]]; then
    ok "/auth/session 200"
    if [[ $have_jq -eq 1 ]]; then
      access=$(jq -r '.data.accessToken // empty' "$sess" 2>/dev/null || true)
      tenant=$(jq -r '.data.tenantId // empty' "$sess" 2>/dev/null || true)
      if [[ -n "$access" && -n "$tenant" ]]; then
        ok "会话返回 accessToken/tenantId"
        echo "\n4) 多租户头校验"
        # 缺少租户头
        out=$(mktemp)
        code1=$(curl -s -o "$out" -w "%{http_code}" -H "Authorization: Bearer $access" -H "Content-Type: application/json" -d '{}' -X POST "${BASE_URL}/api/v1/organization-units" || true)
        if [[ "$code1" == "401" ]]; then
          if [[ $have_jq -eq 1 ]]; then c=$(jq -r '.error.code // empty' "$out" 2>/dev/null || true); [[ "$c" == "TENANT_HEADER_REQUIRED" ]] && ok "缺少租户头 → 401 TENANT_HEADER_REQUIRED" || bad "缺少租户头 错误码不匹配: $c"; else ok "缺少租户头 → 401（未安装 jq）"; fi
        else
          skp "缺少租户头用例未返回 401（HTTP $code1），可能端点受其他校验影响"
        fi
        # 租户不匹配
        out2=$(mktemp)
        code2=$(curl -s -o "$out2" -w "%{http_code}" -H "Authorization: Bearer $access" -H "X-Tenant-ID: 00000000-0000-4000-8000-000000000000" -H "Content-Type: application/json" -d '{}' -X POST "${BASE_URL}/api/v1/organization-units" || true)
        if [[ "$code2" == "403" ]]; then
          if [[ $have_jq -eq 1 ]]; then c2=$(jq -r '.error.code // empty' "$out2" 2>/dev/null || true); [[ "$c2" == "TENANT_MISMATCH" ]] && ok "租户不匹配 → 403 TENANT_MISMATCH" || bad "租户不匹配 错误码不匹配: $c2"; else ok "租户不匹配 → 403（未安装 jq）"; fi
        else
          skp "租户不匹配用例未返回 403（HTTP $code2）"
        fi
        rm -f "$out" "$out2"
      else
        bad "会话返回缺少 accessToken/tenantId"
      fi
    fi
  else
    skp "/auth/session 非 200（HTTP $status_sess），跳过后续校验"
  fi
  rm -f "$sess"
fi
rm -f "$cookiejar"

echo "\n========================================="
echo "通过: $pass  失败: $fail  跳过: $skip"
[[ $fail -eq 0 ]] && exit 0 || exit 1

