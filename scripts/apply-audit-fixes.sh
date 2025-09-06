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

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# 顺序：
# 1) 011 修复（回填 + 正确索引列）
# 2) 014 规范化（补齐审计列 + 防止错列）
# 3) 纠正误写（record_id=actor_id -> 使用 resource_id）
# 4) 基于时间窗口回填缺失的 record_id
# 5) 一致性校验报告

run_sql "$ROOT_DIR/database/migrations/011_audit_record_id_fix.sql"
run_sql "$ROOT_DIR/database/migrations/014_normalize_audit_logs.sql"
run_sql "$ROOT_DIR/scripts/fix-audit-recordid-misplaced.sql"
run_sql "$ROOT_DIR/scripts/fix-audit-record-id-backfill.sql"
run_sql "$ROOT_DIR/scripts/validate-audit-recordid-consistency.sql"

echo "==> Audit fixes applied and validated."

