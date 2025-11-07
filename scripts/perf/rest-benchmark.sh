#!/usr/bin/env bash

set -euo pipefail

# Plan 219E – REST 接口性能基准脚本
# 使用 hey (https://github.com/rakyll/hey) 对指定端点进行并发压测

COMMAND_API="${COMMAND_API:-http://localhost:9090}"
TARGET_PATH="${TARGET_PATH:-/api/v1/organization-units}"
METHOD="${METHOD:-POST}"
CONCURRENCY="${CONCURRENCY:-25}"
DURATION="${DURATION:-15s}"
TENANT_ID="${TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
IDEMPOTENCY_KEY="${IDEMPOTENCY_KEY:-perf-rest-benchmark}"
LOG_DIR="${LOG_DIR:-logs/219E}"
LOAD_DRIVER="${LOAD_DRIVER:-node}" # 支持 node (默认) 或 hey
REQUEST_COUNT="${REQUEST_COUNT:-200}"
THROTTLE_DELAY_MS="${THROTTLE_DELAY_MS:-50}"
REQUEST_TIMEOUT_MS="${REQUEST_TIMEOUT_MS:-8000}"
NAME_PREFIX="${NAME_PREFIX:-Perf Benchmark Dept}"
IDEMPOTENCY_PREFIX="${IDEMPOTENCY_PREFIX:-rest-benchmark}"
NODE_ERROR_SAMPLE="${NODE_ERROR_SAMPLE:-5}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
LOG_FILE="${LOG_DIR}/perf-rest-${TIMESTAMP}.log"

read -r -d '' DEFAULT_REQUEST_BODY <<'JSON' || true
{
  "name": "Perf Benchmark Dept",
  "unitType": "DEPARTMENT",
  "effectiveDate": "2025-11-01"
}
JSON

REQUEST_BODY="${REQUEST_BODY:-$DEFAULT_REQUEST_BODY}"

mkdir -p "${LOG_DIR}" .cache

if ! command -v hey >/dev/null 2>&1; then
  cat <<'EOF'
❌ 未找到 hey 命令。
请先安装：
  go install github.com/rakyll/hey@latest
并将 ~/go/bin 加入 PATH。
EOF
  exit 1
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "❌ 缺少依赖: $1"
    exit 1
  fi
}

require_cmd curl
require_cmd jq

TOKEN="${JWT_TOKEN:-}"

if [[ -z "${TOKEN}" && -f .cache/dev.jwt ]]; then
  TOKEN="$(< .cache/dev.jwt)"
fi

if [[ -z "${TOKEN}" ]]; then
  payload='{"userId":"perf-bot","tenantId":"'"${TENANT_ID}"'","roles":["ADMIN","USER"],"duration":"1h"}'
  response=$(curl -sS -X POST "${COMMAND_API}/auth/dev-token" \
    -H "Content-Type: application/json" \
    -d "${payload}" 2>>"${LOG_FILE}" || true)
  TOKEN="$(echo "${response}" | jq -r '.token // empty')"
fi

if [[ -z "${TOKEN}" ]]; then
  echo "❌ 无法获取 JWT。请通过 make jwt-dev-mint 或设置 JWT_TOKEN 后重试。" | tee -a "${LOG_FILE}"
  exit 1
fi

BASE_PAYLOAD=""
if ! BASE_PAYLOAD="$(printf '%s' "${REQUEST_BODY}" | jq -c '.' 2>/dev/null)"; then
  echo "❌ 请求体不是有效 JSON，请检查 REQUEST_BODY/DEFAULT_REQUEST_BODY。" | tee -a "${LOG_FILE}"
  exit 1
fi

echo "🌐 目标: ${COMMAND_API}${TARGET_PATH}" | tee "${LOG_FILE}"
echo "⚙️  并发: ${CONCURRENCY}  持续/请求: ${DURATION}/${REQUEST_COUNT}  方法: ${METHOD}  驱动: ${LOAD_DRIVER}" | tee -a "${LOG_FILE}"

run_with_hey() {
  AUTH_HEADER="Authorization: Bearer ${TOKEN}"
  TENANT_HEADER="X-Tenant-ID: ${TENANT_ID}"

  HEY_ARGS=(-c "${CONCURRENCY}" -z "${DURATION}" -m "${METHOD}" -H "${AUTH_HEADER}" -H "${TENANT_HEADER}")

  if [[ -n "${IDEMPOTENCY_KEY}" ]]; then
    HEY_ARGS+=(-H "Idempotency-Key: ${IDEMPOTENCY_KEY}")
  fi

  if [[ -n "${REQUEST_BODY}" ]]; then
    HEY_ARGS+=(-T "application/json" -d "${REQUEST_BODY}")
  fi

  hey "${HEY_ARGS[@]}" "${COMMAND_API}${TARGET_PATH}" | tee -a "${LOG_FILE}"
}

run_with_node_driver() {
  require_cmd node

  SUMMARY_JSON=$(
    API_URL="${COMMAND_API}${TARGET_PATH}" \
    AUTH_HEADER="Bearer ${TOKEN}" \
    TENANT_ID="${TENANT_ID}" \
    TOTAL_REQUESTS="${REQUEST_COUNT}" \
    CONCURRENCY="${CONCURRENCY}" \
    THROTTLE_MS="${THROTTLE_DELAY_MS}" \
    TIMEOUT_MS="${REQUEST_TIMEOUT_MS}" \
    METHOD="${METHOD}" \
    BASE_PAYLOAD="${BASE_PAYLOAD}" \
    NAME_PREFIX="${NAME_PREFIX}" \
    IDEMPOTENCY_PREFIX="${IDEMPOTENCY_PREFIX}" \
    LOG_LIMIT="${NODE_ERROR_SAMPLE}" \
    node <<'NODE'
    const { performance } = require('node:perf_hooks');

    (async () => {
      const total = Number(process.env.TOTAL_REQUESTS || 100);
      const concurrency = Math.max(1, Number(process.env.CONCURRENCY || 5));
      const throttle = Math.max(0, Number(process.env.THROTTLE_MS || 0));
      const timeoutMs = Math.max(100, Number(process.env.TIMEOUT_MS || 5000));
      const method = process.env.METHOD || 'POST';
      const apiUrl = process.env.API_URL;
      const authHeader = process.env.AUTH_HEADER;
      const tenantId = process.env.TENANT_ID;
      const namePrefix = process.env.NAME_PREFIX || 'Perf Benchmark';
      const idPrefix = process.env.IDEMPOTENCY_PREFIX || 'rest-bench';
      const logLimit = Math.max(0, Number(process.env.LOG_LIMIT || 5));

      if (!apiUrl) {
        console.error(JSON.stringify({ error: 'API_URL missing' }));
        process.exit(1);
      }

      let basePayload = {};
      try {
        basePayload = JSON.parse(process.env.BASE_PAYLOAD || '{}');
      } catch (error) {
        console.error(JSON.stringify({ error: 'Invalid BASE_PAYLOAD', details: String(error) }));
        process.exit(1);
      }

      const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
      const randomToken = () => Math.random().toString(36).slice(2, 10);

      const results = [];
      const failures = [];
      const generatedCodes = new Set();

      const nextCode = () => {
        let code;
        let attempts = 0;
        do {
          const random = Math.floor(Math.random() * 9_000_000) + 1_000_000;
          code = String(random).padStart(7, '0');
          attempts += 1;
        } while (generatedCodes.has(code) && attempts < 5);
        generatedCodes.add(code);
        return code;
      };

      async function fire(seq) {
        const payload = { ...basePayload };
        payload.name = `${namePrefix}-${seq}-${randomToken()}`;
        payload.code = payload.code ?? nextCode();
        if (!payload.effectiveDate) {
          payload.effectiveDate = new Date().toISOString().slice(0, 10);
        }

        const headers = {
          'Authorization': authHeader,
          'X-Tenant-ID': tenantId,
          'Content-Type': 'application/json',
          'X-Idempotency-Key': `${idPrefix}-${Date.now()}-${seq}-${randomToken()}`,
        };

        const controller = new AbortController();
        const timeout = setTimeout(() => controller.abort(), timeoutMs);
        const started = performance.now();
        let status = 0;
        let ok = false;
        let errorMessage = '';
        try {
          const response = await fetch(apiUrl, {
            method,
            headers,
            body: JSON.stringify(payload),
            signal: controller.signal,
          });
          status = response.status;
          ok = response.ok;
          if (!ok) {
            errorMessage = (await response.text()).slice(0, 512);
          }
        } catch (error) {
          errorMessage = String(error && error.message ? error.message : error);
        } finally {
          clearTimeout(timeout);
        }

        const duration = Number((performance.now() - started).toFixed(2));
        const entry = { seq, status, ok, duration };
        if (errorMessage) {
          entry.error = errorMessage;
        }
        results.push(entry);
        if (!ok && failures.length < logLimit) {
          failures.push(entry);
        }
      }

      let nextIndex = 0;
      async function worker() {
        while (true) {
          const current = nextIndex;
          nextIndex += 1;
          if (current >= total) break;
          await fire(current);
          if (throttle > 0) {
            await sleep(throttle);
          }
        }
      }

      const workerCount = Math.min(concurrency, total);
      const workers = Array.from({ length: workerCount }, () => worker());

      await Promise.all(workers);
      const statusCounts = results.reduce((acc, { status }) => {
        const key = String(status);
        acc[key] = (acc[key] || 0) + 1;
        return acc;
      }, {});
      const successDurations = results
        .filter(({ status }) => status >= 200 && status < 400)
        .map(({ duration }) => duration)
        .sort((a, b) => a - b);

      const percentile = (arr, percent) => {
        if (!arr.length) return null;
        const idx = Math.min(arr.length - 1, Math.round((percent / 100) * (arr.length - 1)));
        return Number(arr[idx].toFixed(2));
      };

      const successCount = successDurations.length;
      const summary = {
        mode: 'node-driver',
        requested: total,
        completed: results.length,
        successCount,
        successRate: results.length ? Number((successCount / results.length).toFixed(4)) : 0,
        statusCounts,
        latency: {
          p50: percentile(successDurations, 50),
          p95: percentile(successDurations, 95),
          p99: percentile(successDurations, 99),
          min: successDurations.length ? Number(Math.min(...successDurations).toFixed(2)) : null,
          max: successDurations.length ? Number(Math.max(...successDurations).toFixed(2)) : null,
        },
        failures,
      };
      console.log(JSON.stringify(summary, null, 2));
    })();
NODE
  )

  if [[ -z "${SUMMARY_JSON}" ]]; then
    echo "❌ Node 驱动执行失败，请检查上方日志。" | tee -a "${LOG_FILE}"
    exit 1
  fi

  echo "📊 JSON Summary:" | tee -a "${LOG_FILE}"
  echo "${SUMMARY_JSON}" | tee -a "${LOG_FILE}"
}

if [[ "${LOAD_DRIVER}" == "hey" ]]; then
  run_with_hey
else
  run_with_node_driver
fi

echo ""
echo "📄 结果日志: ${LOG_FILE}"
