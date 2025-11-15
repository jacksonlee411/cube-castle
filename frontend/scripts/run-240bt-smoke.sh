#!/usr/bin/env bash
set -euo pipefail

# Plan 240BT – 组织详情壳渲染冒烟用例 + HAR 落盘
# - 目录：logs/plan240/BT/
# - 运行：npm run test:e2e:240bt

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${ROOT_DIR}/../logs/plan240/BT"
SPEC="tests/e2e/smoke-org-detail.spec.ts"

mkdir -p "${LOG_DIR}"

timestamp() { date +"%Y-%m-%d %H:%M:%S"; }

echo "[$(timestamp)] 240BT Smoke – 环境健康检查" | tee "${LOG_DIR}/health-checks.log"
for url in \
  "http://localhost:3000/health" \
  "http://localhost:8090/health" \
  "http://localhost:9090/health" \
  "http://localhost:9090/.well-known/jwks.json"; do
  code="$(curl -s -o /dev/null -w "%{http_code}" "$url" || true)"
  echo "[$(timestamp)] ${url} -> ${code}" | tee -a "${LOG_DIR}/health-checks.log"
done

export E2E_PLAN=240BT
export E2E_SAVE_HAR=1
export E2E_STRICT=1

# 若已有本地 dev server，可设置 PW_SKIP_SERVER=1；默认让 Playwright 管控 dev server
export PW_SKIP_SERVER="${PW_SKIP_SERVER:-0}"

echo "[$(timestamp)] 240BT Smoke – 启动 Playwright（SPEC: ${SPEC}）" | tee -a "${LOG_DIR}/smoke-org-detail.log"
npx playwright test "${SPEC}" 2>&1 | tee -a "${LOG_DIR}/smoke-org-detail.log"

echo "[$(timestamp)] 240BT Smoke – 结束。HAR/报告位于：${LOG_DIR}"
