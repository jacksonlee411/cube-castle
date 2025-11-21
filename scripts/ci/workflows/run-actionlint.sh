#!/usr/bin/env bash
# shellcheck disable=SC2086
set -euo pipefail

ROOT_DIR="$(git rev-parse --show-toplevel)"
REPORT_DIR="$ROOT_DIR/reports/workflows"
mkdir -p "$REPORT_DIR"

TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
REPORT_FILE="$REPORT_DIR/actionlint-${TIMESTAMP}.txt"
ACTIONLINT_VERSION="${ACTIONLINT_VERSION:-1.7.4}"

ensure_actionlint() {
  if command -v actionlint >/dev/null 2>&1; then
    return 0
  fi

  echo "🛠️ 安装 actionlint v${ACTIONLINT_VERSION}..."
  GO111MODULE=on go install "github.com/rhysd/actionlint/cmd/actionlint@v${ACTIONLINT_VERSION}"
}

run_actionlint() {
  echo "🔍 运行 actionlint（输出保存至 ${REPORT_FILE}）"
  set -o pipefail
  if actionlint "$@" | tee "$REPORT_FILE"; then
    echo "✅ actionlint 通过"
  else
    status=$?
    echo "❌ actionlint 校验失败，详情见 ${REPORT_FILE}"
    exit $status
  fi
  set +o pipefail
}

ensure_actionlint
run_actionlint "$@"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "report_path=${REPORT_FILE}"
    echo "timestamp=${TIMESTAMP}"
  } >> "${GITHUB_OUTPUT}"
fi

echo "📄 actionlint report: ${REPORT_FILE}"
