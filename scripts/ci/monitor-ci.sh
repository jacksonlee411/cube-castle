#!/usr/bin/env bash
# monitor-ci.sh
# 持续轮询当前提交（或指定 SHA）的 GitHub Actions 检查结果，直到全部完成，输出摘要并落盘到 logs/plan255/ci-summary-<run_id>.txt
# 用法：
#   bash scripts/ci/monitor-ci.sh                # 监控 HEAD 所有检查，自动发现对应 run_id
#   bash scripts/ci/monitor-ci.sh --sha <sha>    # 指定提交哈希
#   bash scripts/ci/monitor-ci.sh --run-id <id>  # 指定运行 id（优先打印该次运行的 jobs，再汇总 check-runs）
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  scripts/ci/monitor-ci.sh [--sha <commit_sha>] [--run-id <actions_run_id>] [--plan <id>] [--owner <owner>] [--repo <repo>] [--interval <sec>] [--timeout <sec>]

Behavior:
  - 加载 GITHUB_TOKEN 顺序：secrets/.env.local -> secrets/.env -> .env.local -> .env -> 环境变量
  - 若未指定 --owner/--repo，则从 git remote origin 自动解析
  - 若未指定 --sha，则使用 HEAD
  - 若未指定 --run-id，将从 /actions/runs?branch=<branch> 中匹配 head_sha 自动选取
  - 每 <interval> 秒轮询一次提交的 check-runs，直到全部 completed 或超时
  - 输出摘要到控制台，并将最终摘要保存到 logs/plan<plan>/ci-summary-<run_id>.txt（默认 plan=255）
EOF
}

# --- args ---
SHA=""
RUN_ID=""
PLAN_ID="${PLAN_ID:-255}"
OWNER=""
REPO=""
INTERVAL=5
TIMEOUT=1200

while [[ $# -gt 0 ]]; do
  case "$1" in
    --sha) SHA="${2:-}"; shift 2;;
    --run-id) RUN_ID="${2:-}"; shift 2;;
    --plan) PLAN_ID="${2:-255}"; shift 2;;
    --owner) OWNER="${2:-}"; shift 2;;
    --repo) REPO="${2:-}"; shift 2;;
    --interval) INTERVAL="${2:-5}"; shift 2;;
    --timeout) TIMEOUT="${2:-1200}"; shift 2;;
    -h|--help) usage; exit 0;;
    *) echo "Unknown arg: $1" >&2; usage; exit 2;;
  esac
done

# --- deps ---
need() { command -v "$1" >/dev/null 2>&1 || { echo "Missing dependency: $1" >&2; exit 3; }; }
need curl; need jq

# --- token ---
load_env_file(){ local f="$1"; [ -f "$f" ] && set -a && . "$f" && set +a || true; }
load_env_file "secrets/.env.local"
load_env_file "secrets/.env"
load_env_file ".env.local"
load_env_file ".env"
TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
AUTH=()
if [ -n "$TOKEN" ]; then AUTH=(-H "Authorization: Bearer ${TOKEN}"); else echo "⚠️  GITHUB_TOKEN 未设置，将以匿名方式访问（可能受限）" >&2; fi

# --- owner/repo ---
if [ -z "$OWNER" ] || [ -z "$REPO" ]; then
  origin_url="$(git remote get-url origin 2>/dev/null || true)"
  # 支持 ssh://git@ssh.github.com:443/owner/repo.git 或 https://github.com/owner/repo.git
  if [[ "$origin_url" =~ github\.com[:/]+([^/]+)/([^/.]+) ]]; then
    OWNER="${OWNER:-${BASH_REMATCH[1]}}"
    REPO="${REPO:-${BASH_REMATCH[2]}}"
  fi
fi
if [ -z "$OWNER" ] || [ -z "$REPO" ]; then
  echo "无法解析 owner/repo；请使用 --owner 与 --repo 指定。" >&2; exit 2
fi

# --- sha/branch ---
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [ -z "$SHA" ]; then SHA="$(git rev-parse HEAD)"; fi
API="https://api.github.com/repos/${OWNER}/${REPO}"
mkdir -p "logs/plan${PLAN_ID}"

# --- resolve run_id if not provided ---
if [ -z "$RUN_ID" ]; then
  runs_json="$(curl -fsSL "${AUTH[@]}" -H "Accept: application/vnd.github+json" \
    "${API}/actions/runs?branch=${BRANCH}&per_page=20")"
  RUN_ID="$(echo "$runs_json" | jq -r ".workflow_runs[] | select(.head_sha==\"${SHA}\") | .id" | head -n1)"
  if [ -z "$RUN_ID" ] || [ "$RUN_ID" = "null" ]; then
    RUN_ID="$(echo "$runs_json" | jq -r '.workflow_runs[0].id // empty')"
  fi
fi
[ -z "$RUN_ID" ] && echo "⚠️ 未发现匹配的 run_id，将仅基于 commit checks 监控。" >&2

echo "🛰️  监控 CI | repo=${OWNER}/${REPO} branch=${BRANCH} sha=${SHA} run_id=${RUN_ID:-unknown}"
start_ts="$(date +%s)"
summary_file="logs/plan${PLAN_ID}/ci-summary-${RUN_ID:-${SHA:0:8}}.txt"
echo "📄 输出摘要：${summary_file}"

print_checks() {
  local out="$1"
  local total; total="$(echo "$out" | jq '.total_count')"
  local completed; completed="$(echo "$out" | jq '[.check_runs[] | select(.status==\"completed\")] | length')"
  echo "⏱️  checks: completed=${completed}/${total}"
  echo "$out" | jq -r '.check_runs[] | [.name,.status, (.conclusion // "-")] | @tsv'
  echo
}

while :; do
  checks_json="$(curl -fsSL "${AUTH[@]}" -H "Accept: application/vnd.github+json" \
    "${API}/commits/${SHA}/check-runs?per_page=100")"
  print_checks "$checks_json" | tee >(sed -e 's/\x1b\[[0-9;]*m//g' >> "$summary_file")
  total="$(echo "$checks_json" | jq '.total_count')"
  completed="$(echo "$checks_json" | jq '[.check_runs[] | select(.status==\"completed\")] | length')"
  if [ "$total" != "0" ] && [ "$completed" = "$total" ]; then
    break
  fi
  now="$(date +%s)"; elapsed=$((now - start_ts))
  if [ "$elapsed" -ge "$TIMEOUT" ]; then
    echo "⏰ 超时 ${TIMEOUT}s，结束监控" | tee -a "$summary_file"
    break
  fi
  sleep "$INTERVAL"
done

echo "--- 失败项（若有） ---" | tee -a "$summary_file"
echo "$checks_json" | jq -r '.check_runs[] | select(.conclusion=="failure") | [.name,.details_url] | @tsv' | tee -a "$summary_file"
echo "✅ 完成：${OWNER}/${REPO}@${SHA:0:8} run_id=${RUN_ID:-unknown}"
