#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dev.yml}"
POSTGRES_SERVICE="${JOB_CATALOG_DB_SERVICE:-postgres}"
POSTGRES_USER="${POSTGRES_USER:-user}"
POSTGRES_DB="${POSTGRES_DB:-cubecastle}"
JOB_CATALOG_CODES_RAW="${JOB_CATALOG_CODES:-OPER}"
REQUIRED_LEVELS_RAW="${JOB_CATALOG_LEVELS:-S1,S2,S3}"

if ! command -v docker >/dev/null 2>&1; then
  echo "❌ 未检测到 docker，请先安装并启动 Docker Desktop (check-job-catalog)"
  exit 2
fi

if ! docker compose -f "${COMPOSE_FILE}" ps >/dev/null 2>&1; then
  echo "❌ docker compose -f ${COMPOSE_FILE} ps 执行失败，确认文件存在且已开启 WSL 集成"
  exit 2
fi

container_id="$(docker compose -f "${COMPOSE_FILE}" ps -q "${POSTGRES_SERVICE}" 2>/dev/null || true)"
if [[ -z "${container_id}" ]]; then
  echo "❌ 未找到 ${POSTGRES_SERVICE} 容器，请先运行 make docker-up"
  exit 2
fi

container_status="$(docker inspect -f '{{.State.Status}}' "${container_id}")"
if [[ "${container_status}" != "running" ]]; then
  health="$(docker inspect -f '{{.State.Health.Status}}' "${container_id}" 2>/dev/null || true)"
  echo "❌ ${POSTGRES_SERVICE} 容器状态为 ${container_status}/${health:-unknown}，请运行 make docker-up && make run-dev"
  exit 2
fi

psql_exec() {
  docker compose -f "${COMPOSE_FILE}" exec -T "${POSTGRES_SERVICE}" \
    psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" "$@"
}

IFS=',' read -r -a JOB_CODES <<< "${JOB_CATALOG_CODES_RAW}"
IFS=',' read -r -a REQUIRED_LEVELS <<< "${REQUIRED_LEVELS_RAW}"

missing_any=0

for code_raw in "${JOB_CODES[@]}"; do
  code="$(echo "${code_raw}" | xargs)"
  [[ -z "${code}" ]] && continue

  group_status="$(psql_exec -Atq -v grp="${code}" <<'SQL'
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

  role_count="$(psql_exec -Atq -v grp="${code}" <<'SQL'
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
  for level_raw in "${REQUIRED_LEVELS[@]}"; do
    level_code="$(echo "${level_raw}" | xargs)"
    [[ -z "${level_code}" ]] && continue
    level_status="$(psql_exec -Atq -v grp="${code}" -v lvl="${level_code}" <<'SQL'
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
    echo "✅ Job Catalog ${code} 检查通过 (roles=${role_count}, levels=${REQUIRED_LEVELS_RAW})"
  fi
done

if [[ "${missing_any}" -ne 0 ]]; then
  echo "👉 请参考 docs/development-plans/230-position-crud-job-catalog-restoration.md 运行修复脚本后重试"
  exit 1
fi
