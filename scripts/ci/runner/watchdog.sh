#!/usr/bin/env bash
set -euo pipefail
#
# Runner Watchdog (持久化 Online 保活)
# - 每 60s 检查 cubecastle-gh-runner 是否存在且处于运行状态
# - 若不存在或退出，调用 start-ghcr-runner-persistent.sh 重启
# - 日志输出到 logs/ci-monitor/runner-watchdog-*.log
#

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
LOG_DIR="${ROOT}/logs/ci-monitor"
mkdir -p "${LOG_DIR}" "${ROOT}/.ci"
TS="$(date +%Y%m%d_%H%M%S)"
LOG_FILE="${LOG_DIR}/runner-watchdog-${TS}.log"

echo "== Runner Watchdog ==" | tee -a "${LOG_FILE}"
echo "start: $(date -u +"%Y-%m-%dT%H:%M:%SZ")" | tee -a "${LOG_FILE}"
echo "log: ${LOG_FILE}" | tee -a "${LOG_FILE}"

interval="${1:-60}"

while true; do
  if [[ -f "${ROOT}/.ci/runner-watchdog.stop" ]]; then
    echo "stop signal detected. exiting." | tee -a "${LOG_FILE}"
    rm -f "${ROOT}/.ci/runner-watchdog.stop" || true
    exit 0
  fi
  now="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
  status="$(docker ps --format '{{.Names}} {{.Status}}' | grep -E '^cubecastle-gh-runner ' || true)"
  if [[ -z "$status" ]]; then
    echo "[$now] runner not running; restarting..." | tee -a "${LOG_FILE}"
    bash "${ROOT}/scripts/ci/runner/start-ghcr-runner-persistent.sh" >> "${LOG_FILE}" 2>&1 || echo "[$now] restart attempt failed" | tee -a "${LOG_FILE}"
  else
    echo "[$now] ok: ${status}" | tee -a "${LOG_FILE}"
  fi
  sleep "${interval}"
done

