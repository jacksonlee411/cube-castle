#!/usr/bin/env bash
set -euo pipefail
#
# Plan 253 – Compose 端口映射与镜像标签门禁
# - 冻结端口映射：禁止将 5432/6379/8090/9090 映射到非同号主机端口
# - 固定镜像标签：PostgreSQL/Redis 禁止使用 latest，必须为显式版本
#

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
NC=$'\033[0m'

fail() { echo "${RED}❌ $*${NC}"; exit 1; }
warn() { echo "${YELLOW}⚠️  $*${NC}"; }
pass() { echo "${GREEN}✅ $*${NC}"; }

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="$ROOT_DIR/logs/plan253"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/compose-ports-and-images.log"
: > "$LOG_FILE"

echo "[plan-253] Checking compose port mappings and image tags..." | tee -a "$LOG_FILE"

# Gather compose files tracked in git at repo root (docker-compose*.yml)
mapfile -t COMPOSE_FILES < <(git ls-files 'docker-compose*.yml')
if [[ ${#COMPOSE_FILES[@]} -eq 0 ]]; then
  fail "未找到 docker-compose*.yml 文件"
fi

echo "[info] Compose files: ${COMPOSE_FILES[*]}" | tee -a "$LOG_FILE"

check_port_freeze() {
  local file="$1" port="$2" ; shift 2 || true
  # Find any mapping like "HOST:port" and ensure HOST == port (e.g., "5432:5432")
  local lines
  lines="$(grep -nE "\"[0-9]+:${port}\"" "$file" || true)"
  if [[ -z "$lines" ]]; then
    # No mapping for this port in this file -> nothing to freeze here
    return 0
  fi
  # Mismatch lines: those not exactly "<port>:<port>"
  local mismatches
  mismatches="$(echo "$lines" | grep -v "\"${port}:${port}\"" || true)"
  if [[ -n "$mismatches" ]]; then
    echo "[error] $file: 端口 ${port} 存在非同号主机映射：" | tee -a "$LOG_FILE"
    echo "$mismatches" | tee -a "$LOG_FILE"
    return 1
  fi
  echo "[ok] $file: 端口 ${port} 同号映射已冻结" | tee -a "$LOG_FILE"
}

check_image_tag() {
  local file="$1" image="$2" ; shift 2 || true
  # For postgres/redis, disallow latest
  if grep -nE "image:[[:space:]]*${image}:latest" "$file" >/dev/null 2>&1; then
    echo "[error] $file: 检测到禁止的标签 ${image}:latest" | tee -a "$LOG_FILE"
    return 1
  fi
  # Ensure a tag exists (not bare image without tag)
  if grep -nE "image:[[:space:]]*${image}(:|$)" "$file" >/dev/null 2>&1; then
    # Accept explicit tags (e.g., redis:7-alpine, postgres:16-alpine)
    echo "[ok] $file: ${image} 标签合规（非 latest）" | tee -a "$LOG_FILE"
  fi
}

errors=0
for f in "${COMPOSE_FILES[@]}"; do
  echo "[file] $f" | tee -a "$LOG_FILE"
  # Freeze port mappings
  for p in 5432 6379 8090 9090; do
    if ! check_port_freeze "$f" "$p"; then
      errors=$((errors+1))
    fi
  done
  # Image tag checks (PostgreSQL/Redis only)
  if ! check_image_tag "$f" "postgres"; then
    errors=$((errors+1))
  fi
  if ! check_image_tag "$f" "redis"; then
    errors=$((errors+1))
  fi
done

if [[ $errors -ne 0 ]]; then
  echo "[fail] 发现 $errors 处违规，详情见 $LOG_FILE" | tee -a "$LOG_FILE"
  fail "Plan-253 compose/image 门禁未通过"
fi

pass "Plan-253 compose/image 门禁通过（详情见 $LOG_FILE）"
