#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SQL_FILE="${ROOT_DIR}/sql/hierarchy-consistency-check.sql"
OUTPUT_DIR="${ROOT_DIR}/reports/hierarchy-consistency"

if ! command -v psql >/dev/null 2>&1; then
  echo "[hierarchy-check] 未找到 psql 命令，请确认已安装 PostgreSQL 客户端。" >&2
  exit 1
fi

if [[ ! -f "${SQL_FILE}" ]]; then
  echo "[hierarchy-check] 未找到 SQL 文件: ${SQL_FILE}" >&2
  exit 1
fi

if [[ -z "${DATABASE_URL:-}" && -z "${PGHOST:-}" ]]; then
  echo "[hierarchy-check] 未检测到数据库连接信息 (DATABASE_URL 或 PGHOST)。" >&2
  echo "请在执行前设置相关环境变量，例如: DATABASE_URL=postgres://user:pass@host:5432/db" >&2
  exit 1
fi

mkdir -p "${OUTPUT_DIR}"

timestamp=$(date -u +%Y%m%dT%H%M%SZ)
output_csv="${OUTPUT_DIR}/hierarchy_anomalies_${timestamp}.csv"

echo "[hierarchy-check] 开始执行层级一致性巡检，结果输出至 ${output_csv}" >&2

psql --set=ON_ERROR_STOP=1 --csv --file="${SQL_FILE}" >"${output_csv}"

# 统计异常数量（减去表头）
line_count=$(wc -l <"${output_csv}" | tr -d ' ')
if [[ "${line_count}" -gt 1 ]]; then
  anomaly_count=$((line_count - 1))
else
  anomaly_count=0
fi

if [[ "${anomaly_count}" -gt 0 ]]; then
  echo "[hierarchy-check] 检测到 ${anomaly_count} 条层级异常，详情见 ${output_csv}" >&2
  exit 2
else
  echo "[hierarchy-check] 未检测到层级异常。" >&2
  # 清理空结果，避免生成空文件
  rm -f "${output_csv}"
fi

exit 0
