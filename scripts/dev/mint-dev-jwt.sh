#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
用法: scripts/dev/mint-dev-jwt.sh [选项]

选项:
  --user-id <id>          JWT subject / userId（默认: dev）
  --tenant-id <uuid>      租户ID（默认: 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9）
  --roles <r1,r2>         角色列表（默认: ADMIN,USER）
  --duration <ttl>        有效期（默认: 8h，命令服务接受 "1h" 格式）
  --base-url <url>        命令服务基础地址（默认: http://localhost:9090）
  --output <path|->       输出令牌路径，使用 "-" 仅输出到 stdout（默认: .cache/dev.jwt）
  -h, --help              显示本帮助

示例:
  scripts/dev/mint-dev-jwt.sh --user-id alice --roles ADMIN,MANAGER
EOF
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

USER_ID="${USER_ID:-dev}"
TENANT_ID="${TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
ROLES="${ROLES:-ADMIN,USER}"
DURATION="${DURATION:-8h}"
BASE_URL="${DEV_COMMAND_URL:-http://localhost:9090}"
OUTPUT_PATH="${OUTPUT_PATH:-$REPO_ROOT/.cache/dev.jwt}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --user-id)
      USER_ID="$2"
      shift 2
      ;;
    --tenant-id)
      TENANT_ID="$2"
      shift 2
      ;;
    --roles)
      ROLES="$2"
      shift 2
      ;;
    --duration)
      DURATION="$2"
      shift 2
      ;;
    --base-url)
      BASE_URL="$2"
      shift 2
      ;;
    --output)
      OUTPUT_PATH="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "未知参数: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

command -v curl >/dev/null 2>&1 || { echo "❌ 需要 curl" >&2; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "❌ 需要 python3" >&2; exit 1; }

IFS=',' read -r -a ROLE_LIST <<<"$ROLES"
ROLE_JSON="["
for role in "${ROLE_LIST[@]}"; do
  trimmed="${role//\"/}"
  trimmed="${trimmed//[[:space:]]/}"
  if [[ -n "$trimmed" ]]; then
    ROLE_JSON+="\"$trimmed\"," 
  fi
done
ROLE_JSON="${ROLE_JSON%,}"  # 去掉尾部逗号
if [[ "$ROLE_JSON" == "[" ]]; then
  ROLE_JSON="[]"
else
  ROLE_JSON+="]"
fi

BODY=$(printf '{"userId":"%s","tenantId":"%s","roles":%s,"duration":"%s"}' \
  "$USER_ID" "$TENANT_ID" "$ROLE_JSON" "$DURATION")

ENDPOINT="$BASE_URL/auth/dev-token"

if ! RESP=$(curl -sf -X POST "$ENDPOINT" -H 'Content-Type: application/json' -d "$BODY"); then
  echo "❌ 调用 $ENDPOINT 失败，请确认命令服务已启动 (make run-dev)" >&2
  exit 2
fi

TOKEN=$(TOKEN_RESPONSE="$RESP" python3 <<'PY'
import base64
import json
import os
import sys

raw = os.environ.get("TOKEN_RESPONSE", "")
if not raw:
    print("❌ 无法解析命令服务响应: 空响应", file=sys.stderr)
    sys.exit(3)
try:
    data = json.loads(raw)
except json.JSONDecodeError as exc:
    print(f"❌ 无法解析命令服务响应: {exc}", file=sys.stderr)
    sys.exit(3)

if not data.get("success"):
    err = data.get("error") or {}
    message = err.get("message") or data.get("message") or "未知错误"
    print(f"❌ 命令服务返回失败: {message}", file=sys.stderr)
    sys.exit(3)

token = (data.get("data") or {}).get("token")
if not token:
    print("❌ 响应中缺少 data.token", file=sys.stderr)
    sys.exit(3)

header = token.split('.')[0]
padding = '=' * (-len(header) % 4)
decoded = base64.urlsafe_b64decode(header + padding).decode()
meta = json.loads(decoded)
alg = meta.get("alg")
if alg != "RS256":
    print(f"❌ 令牌签名算法不匹配: 期望 RS256, 实际 {alg}", file=sys.stderr)
    sys.exit(3)

print(token)
PY
)

if [[ -z "$TOKEN" ]]; then
  echo "❌ 解析令牌失败" >&2
  exit 3
fi

if [[ "$OUTPUT_PATH" != "-" ]]; then
  mkdir -p "$(dirname "$OUTPUT_PATH")"
  umask 077
  printf '%s' "$TOKEN" >"$OUTPUT_PATH"
  echo "✅ 已生成开发令牌，保存至 $OUTPUT_PATH"
else
  printf '%s\n' "$TOKEN"
fi

echo "ℹ️ 令牌来自 $ENDPOINT (roles: $ROLES, ttl: $DURATION)"
