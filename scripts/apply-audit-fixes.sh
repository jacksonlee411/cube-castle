#!/usr/bin/env bash
set -euo pipefail

# apply-audit-fixes.sh
# 一键执行：规范化审计表结构 + 回填/纠正 record_id + 一致性校验
#
# 连接方式优先级：
# 1) 使用环境变量 DATABASE_URL （postgres://user:pass@host:port/dbname）
# 2) 使用 PG* 环境变量（PGHOST, PGPORT, PGUSER, PGPASSWORD, PGDATABASE）
# 3) 使用本机 psql 默认配置

echo "==> Applying audit fixes (normalize + backfill + misplaced fix + validate)"

PSQL=(psql -v ON_ERROR_STOP=1)
if [[ -n "${DATABASE_URL:-}" ]]; then
  PSQL+=("${DATABASE_URL}")
fi

run_sql() {
  local file="$1"
  echo "-- Running: $file"
  "${PSQL[@]}" -f "$file"
}

# 宽松执行：用于可能已执行过的迁移脚本，失败时给出提示但不中断流水线
run_sql_lenient() {
  local file="$1"
  echo "-- Running (lenient): $file"
  if ! "${PSQL[@]}" -f "$file"; then
    echo "!! Warning: lenient execution failed for $file (likely already applied). Continuing..." >&2
  fi
}

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# 顺序与模式：
# - 当 APPLY_FIXES=0 时，仅执行校验（适用于 CI 门禁，不对数据做变更）
# - 当 APPLY_FIXES=1（默认）时，先尝试修复/迁移，再进行校验

APPLY_FIXES_DEFAULT=1
APPLY_FIXES_FLAG=${APPLY_FIXES:-$APPLY_FIXES_DEFAULT}

if [[ "$APPLY_FIXES_FLAG" == "0" ]]; then
  echo "==> APPLY_FIXES=0: Skip fixes/migrations; run validations only"
else
  echo "==> APPLY_FIXES=1: Running fixes/migrations before validations"
  # 1) 011 修复（回填 + 正确索引列）
  # 2) 014 规范化（补齐审计列 + 防止错列）
  # 3) 纠正误写（record_id=actor_id -> 使用 resource_id）
  # 4) 基于时间窗口回填缺失的 record_id
  run_sql_lenient "$ROOT_DIR/database/migrations/011_audit_record_id_fix.sql"
  run_sql_lenient "$ROOT_DIR/database/migrations/014_normalize_audit_logs.sql"
  run_sql_lenient "$ROOT_DIR/scripts/fix-audit-recordid-misplaced.sql"
  run_sql_lenient "$ROOT_DIR/scripts/fix-audit-record-id-backfill.sql"
fi

# 报告版校验（始终执行，输出统计与样本）
run_sql "$ROOT_DIR/scripts/validate-audit-recordid-consistency.sql"

# 断言版校验（可选）: 设置 ENFORCE=1 开启强制校验失败（用于CI/发布闸门）
if [[ "${ENFORCE:-}" == "1" ]]; then
  echo "-- ENFORCE=1: running assert checks"
  # 可通过 APP_ASSERT_TRIGGERS_ZERO=0 暂时跳过触发器数为0的断言（例如执行022之前）
  if [[ -n "${APP_ASSERT_TRIGGERS_ZERO:-}" ]]; then
    PSQL+=("-v" "app.assert_triggers_zero=${APP_ASSERT_TRIGGERS_ZERO}")
  fi
  run_sql "$ROOT_DIR/scripts/validate-audit-recordid-consistency-assert.sql"
fi

echo "==> Audit fixes applied and validated."
