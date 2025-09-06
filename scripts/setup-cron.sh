#!/usr/bin/env bash
set -euo pipefail

# scripts/setup-cron.sh
# 说明：为时态维护与一致性检查安装/移除 cron 任务
# 需求：已安装 psql；提供数据库连接信息（DATABASE_URL 或 PG* 环境变量）

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="${SCRIPT_DIR%/scripts}/logs"
mkdir -p "$LOG_DIR"

PSQL_CMD=${PSQL_CMD:-psql}

usage() {
  cat <<EOF
用法: $0 [--install | --remove | --dry-run]

选项:
  --install   安装每日任务：
              - 02:00 执行 scripts/daily-cutover.sql
              - 03:00 执行 scripts/data-consistency-check.sql
  --remove    移除上述 cron 任务
  --dry-run   仅输出将要安装的 cron 配置，不写入 crontab

环境变量:
  PSQL_CMD       psql 命令（默认: psql）
  DATABASE_URL   PostgreSQL 连接串 (优先级最高)
  PGHOST/PGUSER/PGPASSWORD/PGDATABASE 用于构造连接
EOF
}

build_psql_prefix() {
  if [[ -n "${DATABASE_URL:-}" ]]; then
    echo "${PSQL_CMD} \"${DATABASE_URL}\""
  else
    echo "PGHOST=${PGHOST:-localhost} PGUSER=${PGUSER:-user} PGPASSWORD=${PGPASSWORD:-password} PGDATABASE=${PGDATABASE:-cubecastle} ${PSQL_CMD}"
  fi
}

ensure_psql() {
  if ! command -v ${PSQL_CMD%% *} >/dev/null 2>&1; then
    echo "错误: 未找到 psql，请安装 PostgreSQL 客户端或设置 PSQL_CMD" >&2
    exit 1
  fi
}

install_cron() {
  ensure_psql
  local PSQL_PREFIX
  PSQL_PREFIX="$(build_psql_prefix) -v ON_ERROR_STOP=1"

  local CUTOVER_SQL="$SCRIPT_DIR/daily-cutover.sql"
  local CHECK_SQL="$SCRIPT_DIR/data-consistency-check.sql"
  local CUTOVER_LOG="$LOG_DIR/cron-daily-cutover.log"
  local CHECK_LOG="$LOG_DIR/cron-consistency.log"

  if [[ ! -f "$CUTOVER_SQL" || ! -f "$CHECK_SQL" ]]; then
    echo "错误: 找不到 SQL 脚本：$CUTOVER_SQL 或 $CHECK_SQL" >&2
    exit 1
  fi

  # 生成新 crontab 内容（保留其他任务）
  local TMP_CRON
  TMP_CRON="$(mktemp)"
  crontab -l 2>/dev/null | grep -v "# CubeCastle Temporal Jobs" | grep -v "cron-daily-cutover" | grep -v "cron-consistency" > "$TMP_CRON" || true

  {
    echo "# CubeCastle Temporal Jobs"
    echo "0 2 * * * cd $SCRIPT_DIR && $PSQL_PREFIX -f $CUTOVER_SQL >> $CUTOVER_LOG 2>&1"
    echo "0 3 * * * cd $SCRIPT_DIR && $PSQL_PREFIX -f $CHECK_SQL >> $CHECK_LOG 2>&1"
  } >> "$TMP_CRON"

  if [[ "${1:-}" == "--dry-run" ]]; then
    echo "将安装如下 cron 配置："
    cat "$TMP_CRON"
  else
    crontab "$TMP_CRON"
    echo "✅ 已安装 cron 任务：02:00 daily-cutover，03:00 consistency-check"
  fi

  rm -f "$TMP_CRON"
}

remove_cron() {
  local TMP_CRON
  TMP_CRON="$(mktemp)"
  crontab -l 2>/dev/null | grep -v "# CubeCastle Temporal Jobs" | grep -v "cron-daily-cutover" | grep -v "cron-consistency" > "$TMP_CRON" || true
  crontab "$TMP_CRON" || true
  rm -f "$TMP_CRON"
  echo "✅ 已移除 CubeCastle Temporal cron 任务"
}

main() {
  case "${1:-}" in
    --install)
      install_cron
      ;;
    --dry-run)
      install_cron --dry-run
      ;;
    --remove)
      remove_cron
      ;;
    -h|--help|*)
      usage
      ;;
  esac
}

main "$@"

