#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SQL_FILE="${ROOT_DIR}/sql/hierarchy-consistency-check.sql"

if ! command -v psql >/dev/null 2>&1; then
  echo "[hierarchy-guard] ⚠️ 未安装 psql，跳过层级一致性检查。" >&2
  exit 0
fi

if [[ -z "${DATABASE_URL:-}" && -z "${PGHOST:-}" ]]; then
  echo "[hierarchy-guard] ⚠️ 未检测到数据库连接信息 (DATABASE_URL/PGHOST)，跳过层级一致性检查。" >&2
  exit 0
fi

if [[ ! -f "${SQL_FILE}" ]]; then
  echo "[hierarchy-guard] ❌ 未找到 SQL 文件: ${SQL_FILE}" >&2
  exit 1
fi

tmpfile=$(mktemp)
trap 'rm -f "${tmpfile}"' EXIT

psql --set=ON_ERROR_STOP=1 --csv --file="${SQL_FILE}" >"${tmpfile}"

line_count=$(wc -l <"${tmpfile}" | tr -d ' ')
if [[ "${line_count}" -gt 1 ]]; then
  anomaly_count=$((line_count - 1))
else
  anomaly_count=0
fi

if [[ "${anomaly_count}" -gt 0 ]]; then
  echo "[hierarchy-guard] ❌ 检测到 ${anomaly_count} 条层级异常，CI 将失败。" >&2
  head -n 20 "${tmpfile}" >&2
  exit 1
fi

echo "[hierarchy-guard] ✅ 层级一致性检查通过。" >&2
exit 0
