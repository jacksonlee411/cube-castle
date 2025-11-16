#!/usr/bin/env bash
#
# fetch-gh-artifact.sh <owner/repo> <run_id> <artifact_name_pattern> [out_dir]
# 作用：从 GitHub Actions 指定 run 中下载名称匹配的 artifact 并解压到本地
# Token 加载顺序：secrets/.env.local -> secrets/.env -> .env.local -> .env -> 环境变量
# 依赖：curl、jq、unzip（或 bsdtar）
#
set -euo pipefail

usage() {
  echo "用法: $0 <owner/repo> <run_id> <artifact_name_pattern> [out_dir]"
  echo "示例: $0 cube-castle/cube-castle 123456789 plan255-logs logs/plan255/ci-artifacts"
}

if [ $# -lt 3 ]; then
  usage
  exit 2
fi

REPO="$1"
RUN_ID="$2"
PATTERN="$3"
OUT_DIR="${4:-logs/plan${RUN_ID}/ci-artifacts}"

load_token_from_file() {
  local file="$1"
  if [ -f "$file" ]; then
    set -a
    # shellcheck disable=SC1090
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
TMP_DIR="$(mktemp -d -t gh-artifacts-XXXXXX)"
cleanup() { rm -rf "$TMP_DIR" || true; }
trap cleanup EXIT

echo "ℹ️  查询 artifacts: repo=${REPO} run_id=${RUN_ID}"
ARTS_JSON="${TMP_DIR}/artifacts.json"
curl -fsSL \
  -H "Accept: application/vnd.github+json" \
  ${TOKEN:+-H "Authorization: Bearer ${TOKEN}"} \
  "${API}/repos/${REPO}/actions/runs/${RUN_ID}/artifacts?per_page=100" \
  -o "${ARTS_JSON}"

if ! command -v jq >/dev/null 2>&1; then
  echo "❌ 需要 jq 解析 artifacts 列表" >&2
  exit 3
fi

ART_ID=$(jq -r --arg rx "$PATTERN" '.artifacts[] | select(.name|test($rx)) | .id' "${ARTS_JSON}" | head -n1 || true)
ART_NAME=$(jq -r --arg rx "$PATTERN" '.artifacts[] | select(.name|test($rx)) | .name' "${ARTS_JSON}" | head -n1 || true)
if [ -z "$ART_ID" ] || [ "$ART_ID" = "null" ]; then
  echo "❌ 未找到名称匹配 \"$PATTERN\" 的 artifact" >&2
  exit 4
fi

echo "ℹ️  下载 artifact: id=${ART_ID} name=${ART_NAME}"
ZIP_FILE="${TMP_DIR}/${ART_NAME}.zip"
curl -fsSL \
  -H "Accept: application/vnd.github+json" \
  ${TOKEN:+-H "Authorization: Bearer ${TOKEN}"} \
  -L "${API}/repos/${REPO}/actions/artifacts/${ART_ID}/zip" \
  -o "${ZIP_FILE}"

echo "ℹ️  解压到: ${OUT_DIR}"
mkdir -p "${OUT_DIR}"
if command -v unzip >/dev/null 2>&1; then
  unzip -oq "${ZIP_FILE}" -d "${OUT_DIR}"
elif command -v bsdtar >/dev/null 2>&1; then
  bsdtar -xf "${ZIP_FILE}" -C "${OUT_DIR}"
else
  echo "❌ 需要 unzip 或 bsdtar 解压 Artifact ZIP" >&2
  exit 5
fi

echo "✅ 完成：${OUT_DIR}"

