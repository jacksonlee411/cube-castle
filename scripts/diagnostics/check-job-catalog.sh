#!/usr/bin/env bash
set -euo pipefail

DATABASE_URL="${DATABASE_URL:-postgres://user:password@localhost:5432/cubecastle?sslmode=disable}"
JOB_CATALOG_CODES_RAW="${JOB_CATALOG_CODES:-OPER}"

if ! command -v psql >/dev/null 2>&1; then
  echo "❌ 未找到 psql，请安装 PostgreSQL CLI 后重试 (check-job-catalog)"
  exit 1
fi

IFS=',' read -r -a JOB_CODES <<< "${JOB_CATALOG_CODES_RAW}"

missing_any=0

for code_raw in "${JOB_CODES[@]}"; do
  code="$(echo "${code_raw}" | xargs)"
  [[ -z "${code}" ]] && continue

  group_status="$(psql "${DATABASE_URL}" -Atq -v grp="${code}" <<'SQL'
SELECT status
FROM public.job_family_groups
WHERE family_group_code = :'grp'
  AND is_current = true
LIMIT 1;
SQL
)"
  group_status="$(echo "${group_status}" | tr -d '[:space:]')"

  if [[ -z "${group_status}" ]]; then
    echo "❌ JobFamilyGroup ${code} 缺失，请运行 database/migrations/20251107123000_230_job_catalog_oper_fix.sql"
    missing_any=1
    continue
  fi

  if [[ "${group_status}" != "ACTIVE" ]]; then
    echo "❌ JobFamilyGroup ${code} 状态为 ${group_status}，需激活"
    missing_any=1
  fi

  role_count="$(psql "${DATABASE_URL}" -Atq -v grp="${code}" <<'SQL'
SELECT COUNT(*)::int
FROM public.job_roles
WHERE role_code LIKE (:'grp' || '-%')
  AND status = 'ACTIVE'
  AND is_current = true;
SQL
)"
  role_count="$(echo "${role_count}" | tr -d '[:space:]')"

  if [[ -z "${role_count}" || "${role_count}" == "0" ]]; then
    echo "❌ 未找到以 ${code}- 开头的 ACTIVE JobRole"
    missing_any=1
  fi

  declare -a levels_missing=()
  for level_code in S1 S2 S3; do
    level_status="$(psql "${DATABASE_URL}" -Atq -v grp="${code}" -v lvl="${level_code}" <<'SQL'
SELECT status
FROM public.job_levels
WHERE role_code LIKE (:'grp' || '-%')
  AND level_code = :'lvl'
  AND status = 'ACTIVE'
  AND is_current = true
LIMIT 1;
SQL
)"
    if [[ -z "$(echo "${level_status}" | tr -d '[:space:]')" ]]; then
      levels_missing+=("${level_code}")
    fi
  done

  if [[ "${#levels_missing[@]}" -gt 0 ]]; then
    echo "❌ JobRole ${code}-* 缺少职级: ${levels_missing[*]}"
    missing_any=1
  else
    echo "✅ Job Catalog ${code} 检查通过 (roles=${role_count}, levels=S1/S2/S3)"
  fi
done

if [[ "${missing_any}" -ne 0 ]]; then
  echo "👉 请参考 docs/development-plans/230-position-crud-job-catalog-restoration.md 运行修复脚本后重试"
  exit 1
fi
