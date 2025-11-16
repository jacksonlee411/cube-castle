#!/usr/bin/env bash
set -euo pipefail
#
# Plan 253 – 冷启动计时（记录型，不阻断）
# - 预拉取镜像后测量冷启动（compose up）与数据库健康就绪时间
# - 结果写入 logs/plan253/coldstart-*.log，并由工作流上传为构件
#

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
NC=$'\033[0m'

warn() { echo "${YELLOW}⚠️  $*${NC}"; }
pass() { echo "${GREEN}✅ $*${NC}"; }

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="$ROOT_DIR/logs/plan253"
mkdir -p "$LOG_DIR"
ts="$(date +%Y%m%d%H%M%S)"
LOG_FILE="$LOG_DIR/coldstart-$ts.log"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dev.yml}"
DB_SERVICE="${DB_SERVICE:-postgres}"
CACHE_SERVICE="${CACHE_SERVICE:-redis}"
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-60}"   # seconds

echo "[plan-253] Coldstart metrics (${COMPOSE_FILE})" | tee "$LOG_FILE"
echo "timestamp: $ts" | tee -a "$LOG_FILE"
echo "runner: ${GITHUB_RUN_ID:-local}" | tee -a "$LOG_FILE"

if ! command -v docker >/dev/null 2>&1; then
  echo "❌ 未找到 docker，跳过冷启动测量" | tee -a "$LOG_FILE"
  exit 0
fi

docker version >/dev/null 2>&1 || { warn "docker 不可用，跳过"; exit 0; }

if [[ "${SKIP_PULL:-0}" == "1" ]]; then
  echo "[step] docker compose pull (infra) - skipped by SKIP_PULL=1" | tee -a "$LOG_FILE"
else
  echo "[step] docker compose pull (infra)" | tee -a "$LOG_FILE"
  docker compose -f "$COMPOSE_FILE" pull "$DB_SERVICE" "$CACHE_SERVICE" >>"$LOG_FILE" 2>&1 || true
fi

echo "[step] docker compose down (cleanup)" | tee -a "$LOG_FILE"
docker compose -f "$COMPOSE_FILE" down -v >>"$LOG_FILE" 2>&1 || true

start_up_ms() { date +%s%3N; }
diff_ms() { echo $(( $2 - $1 )); }

echo "[step] compose up (infra)" | tee -a "$LOG_FILE"
t0="$(start_up_ms)"
docker compose -f "$COMPOSE_FILE" up -d "$DB_SERVICE" "$CACHE_SERVICE" >>"$LOG_FILE" 2>&1
t1="$(start_up_ms)"
up_ms="$(diff_ms "$t0" "$t1")"
echo "compose_up_ms: $up_ms" | tee -a "$LOG_FILE"

echo "[step] wait for database healthy (timeout=${HEALTH_TIMEOUT}s)" | tee -a "$LOG_FILE"
t2_start="$(date +%s)"
healthy=0
for i in $(seq 1 "$HEALTH_TIMEOUT"); do
  if docker compose -f "$COMPOSE_FILE" ps | grep -E "${DB_SERVICE}.*(healthy)" >/dev/null 2>&1; then
    healthy=1
    break
  fi
  sleep 1
done
t2_end="$(date +%s)"
db_ready_s=$(( t2_end - t2_start ))
echo "db_ready_seconds: $db_ready_s" | tee -a "$LOG_FILE"

if [[ $healthy -ne 1 ]]; then
  warn "数据库健康探针未在 ${HEALTH_TIMEOUT}s 内通过（仅记录）"
else
  pass "数据库健康就绪: ${db_ready_s}s"
fi

echo "[done] metrics written to: $LOG_FILE"
