#!/usr/bin/env bash
# validate-health.sh - 统一健康检查验证
# 用途：检查 command/query /health 的结构、状态码与关键字段
# 退出码：0=通过，1=不可达，2=结构/状态不符合

set -euo pipefail

COMMAND_URL="${COMMAND_URL:-http://localhost:9090/health}"
QUERY_URL="${QUERY_URL:-http://localhost:8090/health}"
TIMEOUT="${TIMEOUT:-5}"
OUT_DIR="${OUT_DIR:-logs/plan251}"
mkdir -p "${OUT_DIR}"

has_jq() { command -v jq >/dev/null 2>&1; }

check_one() {
  local url="$1"
  local name="$2"
  local ts
  ts="$(date +%Y%m%d-%H%M%S)"
  local outfile="${OUT_DIR}/health-${name}-${ts}.json"

  echo "[*] 检查 ${name} ${url}"
  # -f: fail on non-2xx; we want to capture non-2xx too, so don't use -f here
  # capture status code separately
  http_code=$(curl -sS -m "${TIMEOUT}" -w "%{http_code}" -o "${outfile}" "${url}" || true)
  if [[ -z "${http_code}" ]]; then
    echo "  ✗ 无法访问 ${url}"
    return 1
  fi
  echo "  - HTTP 状态码: ${http_code}"

  if has_jq; then
    status=$(jq -r '.status // empty' < "${outfile}")
    total=$(jq -r '.summary.total // empty' < "${outfile}")
    healthy=$(jq -r '.summary.healthy // empty' < "${outfile}")
  else
    # 简易提取（尽力而为）
    status=$(grep -oE '"status"[^"]*"[a-zA-Z]+"' "${outfile}" | head -n1 | sed -E 's/.*"status"[^"]*"([^"]+)".*/\1/')
    total=$(grep -oE '"total"[^0-9]*[0-9]+' "${outfile}" | head -n1 | grep -oE '[0-9]+')
    healthy=$(grep -oE '"healthy"[^0-9]*[0-9]+' "${outfile}" | head -n1 | grep -oE '[0-9]+')
  fi

  if [[ -z "${status}" || -z "${total}" || -z "${healthy}" ]]; then
    echo "  ✗ 响应缺少必要字段（status/summary）: ${outfile}"
    return 2
  fi
  echo "  - status=${status} summary.total=${total} summary.healthy=${healthy}"

  # 状态码与语义对齐（200/206/503）
  case "${status}" in
    healthy)
      if [[ "${http_code}" != "200" ]]; then
        echo "  ✗ 期望 HTTP 200 实际 ${http_code}"
        return 2
      fi
      ;;
    degraded)
      if [[ "${http_code}" != "206" ]]; then
        echo "  ✗ 期望 HTTP 206 实际 ${http_code}"
        return 2
      fi
      ;;
    unhealthy)
      if [[ "${http_code}" != "503" ]]; then
        echo "  ✗ 期望 HTTP 503 实际 ${http_code}"
        return 2
      fi
      ;;
    *)
      echo "  ✗ 未知 status=${status}"
      return 2
      ;;
  esac

  echo "  ✓ ${name} 健康检查通过（${status}）"
  return 0
}

main() {
  local fail=0
  check_one "${COMMAND_URL}" "command" || fail=1
  check_one "${QUERY_URL}" "query" || fail=1
  if [[ "${fail}" -ne 0 ]]; then
    echo "✗ 健康检查未通过"
    exit 2
  fi
  echo "✓ 健康检查通过"
}

main "$@"

