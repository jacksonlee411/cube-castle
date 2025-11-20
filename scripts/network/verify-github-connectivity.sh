#!/usr/bin/env bash
set -euo pipefail

CMD_TIMEOUT=45
FAIL_FAST=0
SMOKE=0
OUTPUT_FILE=""
RUNNER_COMPOSE="docker-compose.runner.persist.yml"
RUNNER_SERVICE="gh-runner"
SKIP_HOST=0
SKIP_RUNNER=0
LOG_DIR="logs/ci-monitor"
FAILURES=()
GITHUB_URL="https://github.com"
GITHUB_REPO="https://github.com/jacksonlee411/cube-castle"
CURL_USER_AGENT="Mozilla/5.0 (Plan267-D GitHub probe)"
CURL_CMD="curl -sS -o /dev/null -D - --connect-timeout 10 --max-time 60 -H 'User-Agent: ${CURL_USER_AGENT}' ${GITHUB_URL}"
GIT_CMD="GIT_CURL_VERBOSE=1 git ls-remote ${GITHUB_REPO}"
OPENSSL_CMD="openssl s_client -connect github.com:443 -servername github.com </dev/null"

usage() {
  cat <<'EOF'
用法: scripts/network/verify-github-connectivity.sh [选项]

选项:
  --timeout <秒>       单个命令超时时间（默认 20 秒，0 表示不限制）
  --output <文件>      指定输出日志文件（默认 logs/ci-monitor/network-<timestamp>.log）
  --smoke              仅执行 getent/curl 探活，跳过 git/openssl（宿主与 Runner）
  --fail-fast          任一命令失败立即退出
  --skip-host          跳过宿主机检测
  --skip-runner        跳过 Runner 容器检测
  --runner-compose <f> 指定 Runner docker compose 文件（默认 docker-compose.runner.persist.yml）
  --runner-service <s> 指定 Runner 服务名（默认 gh-runner）
  -h, --help           显示本帮助

示例:
  bash scripts/network/verify-github-connectivity.sh --smoke
  bash scripts/network/verify-github-connectivity.sh --timeout 30 --fail-fast
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --timeout)
      CMD_TIMEOUT="$2"
      shift 2
      ;;
    --output)
      OUTPUT_FILE="$2"
      shift 2
      ;;
    --smoke)
      SMOKE=1
      shift
      ;;
    --fail-fast)
      FAIL_FAST=1
      shift
      ;;
    --skip-host)
      SKIP_HOST=1
      shift
      ;;
    --skip-runner)
      SKIP_RUNNER=1
      shift
      ;;
    --runner-compose)
      RUNNER_COMPOSE="$2"
      shift 2
      ;;
    --runner-service)
      RUNNER_SERVICE="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --)
      shift
      break
      ;;
    *)
      echo "[network] 未知参数: $1" >&2
      usage
      exit 1
      ;;
  esac
done

TIMESTAMP=$(date -u +%Y%m%dT%H%M%SZ)
if [[ -z "$OUTPUT_FILE" ]]; then
  OUTPUT_FILE="${LOG_DIR}/network-${TIMESTAMP}.log"
fi
mkdir -p "$(dirname "$OUTPUT_FILE")"

LOG_TS() {
  date -u +"%Y-%m-%dT%H:%M:%SZ"
}

log_line() {
  local msg="$1"
  printf "[%s] %s\n" "$(LOG_TS)" "$msg" | tee -a "$OUTPUT_FILE"
}

run_step() {
  local label="$1"
  local cmd="$2"
  log_line ">>> ($label) $cmd"
  local status
  set +e
  if [[ "$CMD_TIMEOUT" != "0" ]]; then
    timeout --preserve-status "$CMD_TIMEOUT" bash -c "$cmd" 2>&1 | tee -a "$OUTPUT_FILE"
    status=${PIPESTATUS[0]}
  else
    bash -c "$cmd" 2>&1 | tee -a "$OUTPUT_FILE"
    status=${PIPESTATUS[0]}
  fi
  set -e
  if [[ $status -eq 0 ]]; then
    log_line "<<< ($label) OK"
  else
    log_line "<<< ($label) FAIL (exit $status)"
    FAILURES+=("$label:$status")
    if [[ "$FAIL_FAST" -eq 1 ]]; then
      log_line "[fail-fast] 命令失败，提前退出"
      exit "$status"
    fi
  fi
}

compose_cmd() {
  local inner="$1"
  printf "docker compose -f %q exec -T %q bash -lc %q" "$RUNNER_COMPOSE" "$RUNNER_SERVICE" "$inner"
}

runner_available() {
  if [[ "$SKIP_RUNNER" -eq 1 ]]; then
    echo "skip"
    return
  fi
  if ! command -v docker >/dev/null 2>&1; then
    echo "missing-docker"
    return
  fi
  if [[ ! -f "$RUNNER_COMPOSE" ]]; then
    echo "missing-compose"
    return
  fi
  local container_id
  set +e
  container_id=$(docker compose -f "$RUNNER_COMPOSE" ps -q "$RUNNER_SERVICE" 2>/dev/null | tr -d '\n')
  local rc=$?
  set -e
  if [[ $rc -ne 0 ]]; then
    echo "compose-error"
    return
  fi
  if [[ -z "$container_id" ]]; then
    echo "not-running"
  else
    echo "$container_id"
  fi
}

log_line "# Plan 266/267 GitHub Connectivity Probe"
log_line "# Output: $OUTPUT_FILE"
log_line "# Host: $(uname -a)"
log_line "# Timeout: ${CMD_TIMEOUT:-0}s, smoke=${SMOKE}, fail-fast=${FAIL_FAST}"

HOST_STEPS_FULL=(
  "host:getent|getent hosts github.com"
  "host:curl|${CURL_CMD}"
  "host:git|${GIT_CMD}"
  "host:openssl|${OPENSSL_CMD}"
)

HOST_STEPS_SMOKE=(
  "host:getent|getent hosts github.com"
  "host:curl|${CURL_CMD}"
)

RUNNER_STEPS_FULL=(
  "runner:getent|$(compose_cmd "getent hosts github.com")"
  "runner:curl|$(compose_cmd "${CURL_CMD}")"
  "runner:git|$(compose_cmd "${GIT_CMD}")"
  "runner:openssl|$(compose_cmd "${OPENSSL_CMD}")"
)

RUNNER_STEPS_SMOKE=(
  "runner:getent|$(compose_cmd "getent hosts github.com")"
  "runner:curl|$(compose_cmd "${CURL_CMD}")"
)

if [[ "$SKIP_HOST" -ne 1 ]]; then
  log_line "--- 宿主机检测开始 ---"
  if [[ "$SMOKE" -eq 1 ]]; then
    for step in "${HOST_STEPS_SMOKE[@]}"; do
      IFS="|" read -r label cmd <<<"$step"
      run_step "$label" "$cmd"
    done
  else
    for step in "${HOST_STEPS_FULL[@]}"; do
      IFS="|" read -r label cmd <<<"$step"
      run_step "$label" "$cmd"
    done
  fi
  log_line "--- 宿主机检测结束 ---"
else
  log_line "--- 已根据参数跳过宿主机检测 ---"
fi

RUNNER_STATUS=$(runner_available)
case "$RUNNER_STATUS" in
  skip)
    log_line "--- 已根据参数跳过 Runner 检测 ---"
    ;;
  missing-docker)
    log_line "--- 找不到 docker 命令，跳过 Runner 检测 ---"
    ;;
  missing-compose)
    log_line "--- 未找到 $RUNNER_COMPOSE，跳过 Runner 检测 ---"
    ;;
  compose-error)
    log_line "--- docker compose ps 执行失败，跳过 Runner 检测 ---"
    ;;
  not-running|"")
    log_line "--- Runner 容器未运行，跳过 Runner 检测 ---"
    ;;
  *)
    if [[ "$SMOKE" -eq 1 ]]; then
      steps=("${RUNNER_STEPS_SMOKE[@]}")
    else
      steps=("${RUNNER_STEPS_FULL[@]}")
    fi
    log_line "--- Runner ($RUNNER_SERVICE) 检测开始 ---"
    for step in "${steps[@]}"; do
      IFS="|" read -r label cmd <<<"$step"
      run_step "$label" "$cmd"
    done
    log_line "--- Runner ($RUNNER_SERVICE) 检测结束 ---"
    ;;
esac

if [[ ${#FAILURES[@]} -eq 0 ]]; then
  log_line "# 全部检测通过。"
else
  log_line "# 检测存在失败：${FAILURES[*]}"
  exit 1
fi
