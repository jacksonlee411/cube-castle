#!/usr/bin/env bash
#
# Plan 259 — 切换 business GET 门禁为硬阻断（阈值=0）并触发 plan-258-gates 工作流
# 说明：
# - 读取 secrets/.env.local → secrets/.env → .env.local → .env 中的 GITHUB_TOKEN 或 GH_TOKEN
# - 将仓库变量 PLAN259_BUSINESS_GET_THRESHOLD 设为 0
# - 触发 .github/workflows/plan-258-gates.yml 工作流（ref=main）
# - 输出调度摘要（含 run_id 与 run_url，如可用）
# 注意：
# - 需要可访问 GitHub API 的网络环境与具备 repo+workflow 权限的 token
# - 不在 CI 中自动调用；请在本地或自托管 Runner 上手动运行
set -euo pipefail

OWNER="jacksonlee411"
REPO="cube-castle"
WF_FILE="plan-258-gates.yml"
API="https://api.github.com"

# Load token
TOKEN=""
for f in secrets/.env.local secrets/.env .env.local .env; do
  if [[ -z "${TOKEN}" && -f "$f" ]]; then
    # shellcheck disable=SC1090
    source "$f"
    TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
  fi
done
if [[ -z "${TOKEN}" ]]; then
  echo "ERROR: GITHUB_TOKEN/GH_TOKEN not found in secrets/.env* or .env*" >&2
  exit 1
fi

echo "[plan259] Set repo variable PLAN259_BUSINESS_GET_THRESHOLD=0 ..."
code=$(curl -s -o /tmp/resp1.json -w '%{http_code}' -X PATCH \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  "${API}/repos/${OWNER}/${REPO}/actions/variables/PLAN259_BUSINESS_GET_THRESHOLD" \
  -d '{"name":"PLAN259_BUSINESS_GET_THRESHOLD","value":"0"}')
if [[ "${code}" != "204" ]]; then
  echo "WARN: PATCH=${code}; try PUT(create) ..."
  code=$(curl -s -o /tmp/resp1.json -w '%{http_code}' -X PUT \
    -H "Authorization: Bearer ${TOKEN}" \
    -H "Accept: application/vnd.github+json" \
    "${API}/repos/${OWNER}/${REPO}/actions/variables" \
    -d '{"name":"PLAN259_BUSINESS_GET_THRESHOLD","value":"0"}')
  if [[ "${code}" != "201" && "${code}" != "204" ]]; then
    echo "ERROR: set variable failed, http=${code} body=$(cat /tmp/resp1.json)" >&2
    exit 2
  fi
fi

echo "[plan259] Dispatch workflow ${WF_FILE} (ref=main) ..."
code=$(curl -s -o /tmp/resp2.json -w '%{http_code}' -X POST \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  "${API}/repos/${OWNER}/${REPO}/actions/workflows/${WF_FILE}/dispatches" \
  -d '{"ref":"main"}')
if [[ "${code}" != "204" ]]; then
  echo "ERROR: dispatch failed, http=${code} body=$(cat /tmp/resp2.json)" >&2
  exit 3
fi

echo "[plan259] Poll latest workflow run ..."
summary=/tmp/plan258_dispatch_summary.json
run_id=""
run_url=""
status=""
conclusion=""
for i in {1..10}; do
  sleep 3
  code=$(curl -s -o /tmp/resp3.json -w '%{http_code}' \
    -H "Authorization: Bearer ${TOKEN}" -H "Accept: application/vnd.github+json" \
    "${API}/repos/${OWNER}/${REPO}/actions/workflows/${WF_FILE}/runs?event=workflow_dispatch&per_page=1")
  if [[ "${code}" != "200" ]]; then
    echo "WARN: list runs http=${code}"
    continue
  fi
  run_id=$(jq -r '.workflow_runs[0].id // empty' /tmp/resp3.json)
  run_url=$(jq -r '.workflow_runs[0].html_url // empty' /tmp/resp3.json)
  status=$(jq -r '.workflow_runs[0].status // empty' /tmp/resp3.json)
  conclusion=$(jq -r '.workflow_runs[0].conclusion // empty' /tmp/resp3.json)
  if [[ -n "${run_id}" ]]; then
    printf '{"run_id":%s,"run_url":"%s","status":"%s","conclusion":"%s"}\n' "${run_id}" "${run_url}" "${status}" "${conclusion}" | tee "${summary}"
    echo "[plan259] Latest run: ${run_url} (status=${status}, conclusion=${conclusion})"
    exit 0
  fi
done
echo "WARN: No run id obtained yet; it may appear shortly on GitHub UI."
exit 0

