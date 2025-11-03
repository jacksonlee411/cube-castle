#!/usr/bin/env bash
set -euo pipefail

# 数据一致性巡检脚本
# 依赖 SQL：scripts/data-consistency-check.sql
# 用法：
#   DATABASE_URL=postgres://user:pass@localhost:5432/cubecastle \
#     scripts/tests/test-data-consistency.sh
#   scripts/tests/test-data-consistency.sh --dry-run

usage() {
  cat <<'EOF'
用法: test-data-consistency.sh [选项]

选项:
  --output DIR   指定报告输出目录（默认: reports/consistency）
  --dry-run      演练模式，仅展示即将执行的步骤
  -h, --help     显示本帮助

所需环境变量:
  DATABASE_URL 或 DATABASE_URL_HOST
  （或通过 PGHOST/PGUSER/PGPASSWORD/PGDATABASE 组合连接）
EOF
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SQL_FILE="${ROOT_DIR}/scripts/data-consistency-check.sql"
ENV_FILE="${ROOT_DIR}/.env"
DEFAULT_OUTPUT_DIR="${ROOT_DIR}/reports/consistency"

OUTPUT_DIR="${DEFAULT_OUTPUT_DIR}"
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output)
      if [[ -z "${2:-}" ]]; then
        echo "[data-consistency] --output 需要目录参数" >&2
        usage
        exit 2
      fi
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "[data-consistency] 未知参数: $1" >&2
      usage
      exit 2
      ;;
  esac
done

if [[ ! -f "${SQL_FILE}" ]]; then
  echo "[data-consistency] 未找到 SQL 文件: ${SQL_FILE}" >&2
  exit 1
fi

PSQL_BIN="${PSQL_BIN:-psql}"
if ! command -v "${PSQL_BIN}" >/dev/null 2>&1; then
  echo "[data-consistency] 未找到 psql，请确认已安装 PostgreSQL 客户端或设置 PSQL_BIN" >&2
  exit 1
fi

# 加载 .env（若存在且未显式提供连接信息）
if [[ -z "${DATABASE_URL:-}" && -z "${PGHOST:-}" && -f "${ENV_FILE}" ]]; then
  echo "[data-consistency] 检测到 .env，尝试加载数据库连接参数…" >&2
  set -a
  # shellcheck disable=SC1090
  source <(sed $'s/\r$//' "${ENV_FILE}")
  set +a
fi

if [[ -z "${DATABASE_URL:-}" && -n "${DATABASE_URL_HOST:-}" ]]; then
  export DATABASE_URL="${DATABASE_URL_HOST}"
fi

if [[ "${DRY_RUN}" == "true" ]]; then
  cat <<EOF
[data-consistency] Dry-run 模式：
  - SQL 文件：${SQL_FILE}
  - 输出目录：${OUTPUT_DIR}
  - 使用命令：${PSQL_BIN} (需要 DATABASE_URL / PGHOST 等连接信息)
  - 产物：原始 CSV + Markdown 摘要
EOF
  exit 0
fi

if [[ -z "${DATABASE_URL:-}" && -z "${PGHOST:-}" ]]; then
  echo "[data-consistency] 未检测到数据库连接信息 (DATABASE_URL 或 PGHOST)。" >&2
  echo "请在执行前设置相关环境变量，例如:" >&2
  echo "  DATABASE_URL=postgres://user:pass@localhost:5432/cubecastle" >&2
  exit 1
fi

mkdir -p "${OUTPUT_DIR}"

timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
raw_output="${OUTPUT_DIR}/data-consistency-${timestamp}.csv"
summary_output="${OUTPUT_DIR}/data-consistency-summary-${timestamp}.md"

echo "[data-consistency] 开始执行数据一致性巡检…" >&2

PSQL_ARGS=(-v ON_ERROR_STOP=1 -qAt -F',' -f "${SQL_FILE}")
if [[ -n "${DATABASE_URL:-}" ]]; then
  "${PSQL_BIN}" "${DATABASE_URL}" "${PSQL_ARGS[@]}" >"${raw_output}"
else
  "${PSQL_BIN}" "${PSQL_ARGS[@]}" >"${raw_output}"
fi

multiple_current=0
temporal_overlap=0
invalid_parent=0
deleted_but_current=0
audit_recent=0

while IFS=',' read -r issue col2 col3 col4 col5; do
  [[ -z "${issue}" ]] && continue
  case "${issue}" in
    MULTIPLE_CURRENT)
      multiple_current=$((multiple_current + 1))
      ;;
    TEMPORAL_OVERLAP)
      temporal_overlap=$((temporal_overlap + 1))
      ;;
    INVALID_PARENT)
      invalid_parent=$((invalid_parent + 1))
      ;;
    DELETED_BUT_CURRENT)
      deleted_but_current=$((deleted_but_current + 1))
      ;;
    AUDIT_RECENT)
      if [[ -n "${col2:-}" ]]; then
        audit_recent="${col2}"
      fi
      ;;
    *)
      ;;
  esac
done <"${raw_output}"

issue_total=$((multiple_current + temporal_overlap + invalid_parent + deleted_but_current))
if [[ "${issue_total}" -eq 0 ]]; then
  summary_status="✅ PASS"
  exit_code=0
else
  summary_status="❌ FAIL"
  exit_code=2
fi

cat >"${summary_output}" <<EOF
# 数据一致性巡检报告 (${timestamp})

| 检查项 | 异常数量 |
|--------|----------|
| 多个 is_current 版本冲突 | ${multiple_current} |
| 时态区间重叠 | ${temporal_overlap} |
| 无效父节点 | ${invalid_parent} |
| 软删除仍标记为当前 | ${deleted_but_current} |
| 最近 7 天审计日志数 | ${audit_recent} |
| 结论 | ${summary_status} |

- SQL 来源：\`scripts/data-consistency-check.sql\`
- 原始输出：\`$(basename "${raw_output}")\`
- 生成时间（UTC）：${timestamp}

如发现异常，请参考 \`docs/architecture/temporal-consistency-implementation-report.md\` 制定修复计划，并在修复后重新执行本脚本。
EOF

cat <<EOF
[data-consistency] 巡检完成：
  - 多个 is_current 版本冲突：${multiple_current}
  - 时态区间重叠：${temporal_overlap}
  - 无效父节点：${invalid_parent}
  - 软删除仍标记为当前：${deleted_but_current}
  - 最近 7 天审计日志数：${audit_recent}
  - 摘要报告：${summary_output}
  - 原始输出：${raw_output}
  - 判定：${summary_status}
EOF

exit "${exit_code}"
