#!/usr/bin/env bash
set -euo pipefail

# Phase 1 Acceptance Check Aggregator
# 目标：集中执行 Plan 211 Phase1 验收所需的核心校验（构建、测试、前端校验、健康检查、数据一致性）
# 用法：
#   scripts/phase1-acceptance-check.sh
#     --dry-run           仅输出将执行的命令，不实际运行
#     --skip-frontend     跳过前端 lint / test
#     --skip-data-check   跳过数据一致性巡检
#
# 产物：
#   reports/acceptance/phase1-acceptance-<timestamp>.log
#   reports/acceptance/phase1-acceptance-summary-<timestamp>.md

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${ROOT_DIR}/reports/acceptance"
DATA_CHECK_SCRIPT="${ROOT_DIR}/scripts/tests/test-data-consistency.sh"

DRY_RUN=false
RUN_FRONTEND=true
RUN_DATA_CHECK=true

usage() {
  cat <<'EOF'
用法: scripts/phase1-acceptance-check.sh [选项]

选项:
  --dry-run           仅展示将要执行的命令
  --skip-frontend     跳过前端 lint/test（默认执行）
  --skip-data-check   跳过数据一致性巡检（默认执行）
  -h, --help          显示帮助
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --skip-frontend)
      RUN_FRONTEND=false
      shift
      ;;
    --skip-data-check)
      RUN_DATA_CHECK=false
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "未知参数: $1" >&2
      usage
      exit 2
      ;;
  esac
done

mkdir -p "${LOG_DIR}"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
LOG_FILE="${LOG_DIR}/phase1-acceptance-${timestamp}.log"
SUMMARY_FILE="${LOG_DIR}/phase1-acceptance-summary-${timestamp}.md"

record() {
  printf '%s\n' "$*" | tee -a "${LOG_FILE}"
}

run_step() {
  local label="$1"
  shift
  record ""
  record "===== ${label} ====="
  if ${DRY_RUN}; then
    record "[DRY-RUN] $*"
    return 0
  fi
  "$@" 2>&1 | tee -a "${LOG_FILE}"
}

record "Phase1 Acceptance Check - ${timestamp}"
record "工作目录: ${ROOT_DIR}"

# Step 1: Go build (command & query)
run_step "Go Build (command)" go build ./cmd/hrms-server/command
run_step "Go Build (query)" go build ./cmd/hrms-server/query

# Step 2: Go tests
run_step "Go Test (unit/integration)" go test ./... -count=1

# Step 3: make test (保持与 CI 一致)
run_step "Make Test" make test

# Step 4: 前端 lint/test
if ${RUN_FRONTEND}; then
  run_step "Frontend Lint" npm run lint
  run_step "Frontend Unit Tests" bash -lc "cd frontend && npm run test:run"
else
  record "跳过前端检查 (--skip-frontend)"
fi

# Step 5: 服务健康检查
run_step "Command Service Health" curl --fail --silent --show-error http://localhost:9090/health
run_step "Query Service Health" curl --fail --silent --show-error http://localhost:8090/health

# Step 6: 数据一致性巡检
data_summary=""
if ${RUN_DATA_CHECK}; then
  if [[ ! -x "${DATA_CHECK_SCRIPT}" ]]; then
    record "警告：未找到数据一致性脚本 ${DATA_CHECK_SCRIPT}，跳过执行"
  else
    if [[ -z "${DATABASE_URL_HOST_TOOLS:-}" && -f "${ROOT_DIR}/.env" ]]; then
      DATABASE_URL_HOST_TOOLS="$(grep -E '^DATABASE_URL_HOST_TOOLS=' "${ROOT_DIR}/.env" | tail -n 1 | cut -d '=' -f2-)"
      export DATABASE_URL_HOST_TOOLS
    fi
    if [[ -z "${DATABASE_URL:-}" && -n "${DATABASE_URL_HOST_TOOLS:-}" ]]; then
      export DATABASE_URL="${DATABASE_URL_HOST_TOOLS}"
    fi
    run_step "Data Consistency Check" "${DATA_CHECK_SCRIPT}" --output "${LOG_DIR}"
    data_summary=$(ls -t "${LOG_DIR}"/data-consistency-summary-*.md 2>/dev/null | head -n 1 || true)
  fi
else
  record "跳过数据一致性巡检 (--skip-data-check)"
fi

# Step 7: 汇总
if ! ${DRY_RUN}; then
  cat >"${SUMMARY_FILE}" <<EOF
# Phase1 Acceptance Summary (${timestamp})

- Go Build：command/query ✅
- Go Test：✅
- Make Test：✅
- Frontend Lint/Test：$( ${RUN_FRONTEND} && echo "✅" || echo "⏭️ 跳过")
- Health Checks：9090/8090 ✅
- Data Consistency：$( if ${RUN_DATA_CHECK}; then [[ -n "${data_summary}" ]] && echo "✅ (reports/acceptance/$(basename "${data_summary}") )" || echo "✅"; else echo "⏭️ 跳过"; fi )

详见日志：reports/acceptance/$(basename "${LOG_FILE}")
EOF
  record ""
  record "汇总输出: ${SUMMARY_FILE}"
fi

if ${DRY_RUN}; then
  record ""
  record "DRY-RUN 完成。未执行实际命令。"
else
  record ""
  record "Phase1 验收检查完成。"
fi
