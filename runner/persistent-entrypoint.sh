#!/usr/bin/env bash
set -euo pipefail

die() { echo "❌ $*" >&2; exit 1; }
maybe_require() {
  local var="$1"
  [[ -n "${!var:-}" ]] || die "缺少必要环境变量: ${var}"
}

cd /home/runner

RUNNER_REPO="${RUNNER_REPO:-jacksonlee411/cube-castle}"
RUNNER_NAME="${RUNNER_NAME:-cc-runner-$(hostname)}"
RUNNER_LABELS="${RUNNER_LABELS:-self-hosted,cubecastle,linux,x64,docker}"
RUNNER_WORKDIR="${RUNNER_WORKDIR:-/home/runner/_work}"
DISABLE_AUTO_UPDATE="${DISABLE_AUTO_UPDATE:-true}"
FORCE_RECONFIGURE="${FORCE_RECONFIGURE:-false}"
CONFIG_SENTINEL_PRIMARY=".runner/.credentials"
CONFIG_SENTINEL_FALLBACK=".credentials"

if [[ "${FORCE_RECONFIGURE}" == "true" ]]; then
  echo "⚠️ FORCE_RECONFIGURE=true, 清理既有 runner 状态"
  rm -rf .runner .credentials .credentials_migrated || true
fi

needs_config="true"
if [[ -f "${CONFIG_SENTINEL_PRIMARY}" || -f "${CONFIG_SENTINEL_FALLBACK}" ]]; then
  needs_config="false"
fi

if [[ "${needs_config}" == "true" ]]; then
  RUNNER_TOKEN="${RUNNER_TOKEN:-${GH_RUNNER_REG_TOKEN:-${GH_RUNNER_PAT:-}}}"
  maybe_require RUNNER_TOKEN
  echo "🔧 首次初始化 Runner（${RUNNER_NAME} → ${RUNNER_REPO}）"
  ./config.sh \
    --url "https://github.com/${RUNNER_REPO}" \
    --token "${RUNNER_TOKEN}" \
    --name "${RUNNER_NAME}" \
    --labels "${RUNNER_LABELS}" \
    --work "${RUNNER_WORKDIR}" \
    --unattended \
    --disableupdate
else
  echo "ℹ️ 检测到现有 .runner 配置，跳过 config.sh"
fi

echo "▶ 启动 run.sh（persistent 模式，不自动 remove）"
exec ./run.sh
