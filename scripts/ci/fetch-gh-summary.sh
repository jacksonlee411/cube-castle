#!/usr/bin/env bash
#
# fetch-gh-summary.sh <owner/repo> <run_id>
# 作用：从 GitHub Actions 某次 run 的压缩日志中提取包含 "SUMMARY" 关键字的行
# Token 加载顺序：secrets/.env.local -> secrets/.env -> .env.local -> .env -> 环境变量
# 依赖：curl、unzip（或 bsdtar）、rg(可选)/grep
#
set -euo pipefail

usage() {
  echo "Usage: $0 <owner/repo> <run_id>"
  echo "Example: $0 cube-castle/cube-castle 123456789"
}

if [ $# -lt 2 ]; then
  usage
  exit 2
fi

REPO="$1"
RUN_ID="$2"

# 尝试从本地文件加载 TOKEN
load_token_from_file() {
  local file="$1"
  if [ -f "$file" ]; then
    # shellcheck disable=SC1090
    set -a
    . "$file" || true
    set +a
  fi
}

load_token_from_file "secrets/.env.local"
load_token_from_file "secrets/.env"
load_token_from_file ".env.local"
load_token_from_file ".env"

TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
if [ -z "$TOKEN" ]; then
  echo "⚠️  GITHUB_TOKEN 未设置，将以匿名方式访问（可能受限）" >&2
fi

API="https://api.github.com"
TMP_DIR="$(mktemp -d -t gh-logs-XXXXXX)"
ZIP_FILE="${TMP_DIR}/run-${RUN_ID}-logs.zip"

cleanup() {
  rm -rf "$TMP_DIR" || true
}
trap cleanup EXIT

echo "ℹ️  下载运行日志压缩包: $REPO run_id=$RUN_ID"
curl -fsSL \
  -H "Accept: application/vnd.github+json" \
  ${TOKEN:+-H "Authorization: Bearer ${TOKEN}"} \
  "${API}/repos/${REPO}/actions/runs/${RUN_ID}/logs" \
  -o "$ZIP_FILE"

echo "ℹ️  解压日志..."
EXTRACT_DIR="${TMP_DIR}/unzipped"
mkdir -p "$EXTRACT_DIR"

if command -v unzip >/dev/null 2>&1; then
  unzip -qq "$ZIP_FILE" -d "$EXTRACT_DIR"
elif command -v bsdtar >/dev/null 2>&1; then
  bsdtar -xf "$ZIP_FILE" -C "$EXTRACT_DIR"
else
  echo "❌ 需要 unzip 或 bsdtar 解压日志压缩包" >&2
  exit 3
fi

echo "ℹ️  搜索 SUMMARY 关键字..."
if command -v rg >/dev/null 2>&1; then
  rg -n "SUMMARY" "$EXTRACT_DIR" || true
else
  grep -RIn "SUMMARY" "$EXTRACT_DIR" || true
fi

echo "✅ 完成"

