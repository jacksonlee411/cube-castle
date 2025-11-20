#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
WSL Runner preflight & verify (Plan 269)

Usage:
  wsl-verify.sh [options]

Options:
  --runner-dir <path>   Runner 目录（默认 $HOME/actions-runner）
  --repo <owner/repo>   目标仓库（默认 jacksonlee411/cube-castle）
  --env-file <path>     secrets 路径（默认 <repo>/secrets/.env.local）
  --tmux-session <name> tmux 会话名称（默认 cc-runner）
  --skip-network        跳过网络探测脚本
  -h, --help            查看帮助
EOF
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
LOG_DIR="${RUNNER_LOG_DIR:-$REPO_ROOT/logs/wsl-runner}"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/verify-$(date +%Y%m%dT%H%M%S).log"
exec > >(tee -a "$LOG_FILE") 2>&1

log() {
  printf '[%s] %s\n' "$(date --iso-8601=seconds)" "$*" >&2
}

fail() {
  log "ERROR: $*"
  exit 1
}

require_cmd() {
  local cmd="$1"
  command -v "$cmd" >/dev/null 2>&1 || fail "缺少依赖命令: $cmd"
}

compare_versions() {
  python3 - "$1" "$2" <<'PY'
import sys
from itertools import zip_longest
a = [int(x) for x in sys.argv[1].split('.')]
b = [int(x) for x in sys.argv[2].split('.')]
for left, right in zip_longest(a, b, fillvalue=0):
    if left > right:
        print(1)
        break
    if left < right:
        print(0)
        break
else:
    print(1)
PY
}

ensure_version() {
  local name="$1" current="$2" minimum="$3"
  if [[ "$(compare_versions "$current" "$minimum")" != "1" ]]; then
    fail "$name 版本过低：当前 $current，至少需要 $minimum"
  fi
  log "$name $current ✓"
}

RUNNER_DIR="${RUNNER_DIR:-$HOME/actions-runner}"
RUNNER_REPO="${RUNNER_REPO:-jacksonlee411/cube-castle}"
RUNNER_NAME="${RUNNER_NAME:-wsl-$(hostname)}"
RUNNER_TMUX_SESSION="${RUNNER_TMUX_SESSION:-cc-runner}"
ENV_FILE_DEFAULT="$REPO_ROOT/secrets/.env.local"
ENV_FILE="${ENV_FILE:-$ENV_FILE_DEFAULT}"
SKIP_NETWORK=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --runner-dir) RUNNER_DIR="$2"; shift 2;;
    --repo) RUNNER_REPO="$2"; shift 2;;
    --env-file) ENV_FILE="$2"; shift 2;;
    --tmux-session) RUNNER_TMUX_SESSION="$2"; shift 2;;
    --skip-network) SKIP_NETWORK=true; shift;;
    -h|--help) usage; exit 0;;
    *) fail "未知参数: $1";;
  esac
done

if [[ -f "$ENV_FILE" ]]; then
  log "加载 env 文件: $ENV_FILE"
  # shellcheck disable=SC1090
  set -a && source "$ENV_FILE" && set +a
fi

GITHUB_API_URL="${GITHUB_API_URL:-https://api.github.com}"

log "==== WSL Runner 预检 (${RUNNER_DIR}) ===="
require_cmd tee
require_cmd uname
require_cmd go
require_cmd node
require_cmd docker
require_cmd python3

log "系统: $(uname -a)"
if command -v lsb_release >/dev/null 2>&1; then
  log "发行版: $(lsb_release -sd)"
fi

GO_VERSION="$(go version | awk '{print $3}' | sed 's/go//')"
NODE_VERSION_STR="$(node --version | sed 's/^v//')"
ensure_version "Go" "$GO_VERSION" "1.24.0"
ensure_version "Node.js" "$NODE_VERSION_STR" "18.0.0"

log "Docker version: $(docker version --format '{{.Server.Version}}' 2>/dev/null || docker version --format '{{.Client.Version}}')"
docker compose version >/dev/null
docker info >/dev/null
if docker context show >/dev/null 2>&1; then
  log "Docker context: $(docker context show)"
fi

if [[ "$SKIP_NETWORK" != true ]] && [[ -x "$REPO_ROOT/scripts/network/verify-github-connectivity.sh" ]]; then
  log "执行网络探测（smoke）"
  "$REPO_ROOT/scripts/network/verify-github-connectivity.sh" --smoke --output "$LOG_DIR/network-smoke-$(date +%Y%m%dT%H%M%S).log" || fail "网络探测失败"
else
  log "跳过网络探测 (--skip-network 或脚本缺失)"
fi

if [[ -d "$RUNNER_DIR" ]]; then
  log "Runner 目录存在：$RUNNER_DIR"
  ls -al "$RUNNER_DIR" | head -20
  if [[ -d "$RUNNER_DIR/_diag" ]]; then
    log "最近的 _diag 日志："
    ls -1t "$RUNNER_DIR/_diag" | head -5
  fi
  if [[ -f "$RUNNER_DIR/.runner" ]]; then
    log ".runner 配置存在"
  else
    log "未发现 .runner，可能尚未执行 config.sh"
  fi
else
  fail "Runner 目录不存在：$RUNNER_DIR"
fi

if command -v tmux >/dev/null 2>&1; then
  if tmux has-session -t "$RUNNER_TMUX_SESSION" 2>/dev/null; then
    log "tmux 会话 $RUNNER_TMUX_SESSION 已运行"
  else
    log "未发现 tmux 会话 $RUNNER_TMUX_SESSION"
  fi
fi

response=""
if command -v gh >/dev/null 2>&1; then
  log "查询 GitHub Runner 状态（gh api）"
  response="$(gh api --method GET -H "Accept: application/vnd.github+json" "/repos/$RUNNER_REPO/actions/runners?per_page=100")"
elif [[ -n "${GH_RUNNER_PAT:-}" ]]; then
  log "查询 GitHub Runner 状态（PAT）"
  if ! response="$(curl -fsSL \
    -H "Authorization: token $GH_RUNNER_PAT" \
    -H "Accept: application/vnd.github+json" \
    -H "User-Agent: wsl-runner-verify" \
    "$GITHUB_API_URL/repos/$RUNNER_REPO/actions/runners?per_page=100")"; then
    log "curl 查询失败，跳过在线状态检查"
    response=""
  fi
fi
if [[ -n "$response" ]]; then
  printf '%s' "$response" | python3 - "$RUNNER_NAME" <<'PY'
import json, sys
target = sys.argv[1]
data = json.loads(sys.stdin.read())
for runner in data.get("runners", []):
    if runner.get("name") == target:
        print(f"Runner {target}: status={runner.get('status')} busy={runner.get('busy')} labels={[l.get('name') for l in runner.get('labels', [])]}")
        break
else:
    print(f"Runner {target} 未在仓库中注册")
PY
else
  log "Runner 列表响应为空或凭据缺失，跳过在线状态检查"
fi

log "验证完成。请将本日志上传到 Plan 265/266/269 的运行记录中。"
