#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
WSL GitHub Actions Runner uninstall helper (Plan 269)

Usage:
  wsl-uninstall.sh [options]

Options:
  --repo <owner/repo>   目标仓库（默认 jacksonlee411/cube-castle）
  --runner-dir <path>   Runner 目录（默认 $HOME/actions-runner）
  --env-file <path>     secrets 路径（默认 <repo>/secrets/.env.local）
  --use-systemd         若 Runner 以 systemd 安装，则自动 stop/uninstall
  --tmux-session <name> 自定义 tmux 会话名（默认 cc-runner）
  -h, --help            查看帮助
EOF
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
LOG_DIR="${RUNNER_LOG_DIR:-$REPO_ROOT/logs/wsl-runner}"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/uninstall-$(date +%Y%m%dT%H%M%S).log"
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

RUNNER_REPO="${RUNNER_REPO:-jacksonlee411/cube-castle}"
RUNNER_DIR="${RUNNER_DIR:-$HOME/actions-runner}"
RUNNER_TMUX_SESSION="${RUNNER_TMUX_SESSION:-cc-runner}"
ENV_FILE_DEFAULT="$REPO_ROOT/secrets/.env.local"
ENV_FILE="${ENV_FILE:-$ENV_FILE_DEFAULT}"
USE_SYSTEMD=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo) RUNNER_REPO="$2"; shift 2;;
    --runner-dir) RUNNER_DIR="$2"; shift 2;;
    --env-file) ENV_FILE="$2"; shift 2;;
    --tmux-session) RUNNER_TMUX_SESSION="$2"; shift 2;;
    --use-systemd) USE_SYSTEMD=true; shift;;
    -h|--help) usage; exit 0;;
    *) fail "未知参数: $1";;
  esac
done

if [[ -f "$ENV_FILE" ]]; then
  log "加载环境变量: $ENV_FILE"
  # shellcheck disable=SC1090
  set -a && source "$ENV_FILE" && set +a
else
  log "未找到 env 文件（$ENV_FILE），尝试使用当前环境变量"
fi

RUNNER_URL="${RUNNER_URL:-https://github.com/${RUNNER_REPO}}"
GITHUB_API_URL="${GITHUB_API_URL:-https://api.github.com}"

require_cmd tee
require_cmd curl
require_cmd python3

if [[ ! -d "$RUNNER_DIR" ]]; then
  log "Runner 目录 $RUNNER_DIR 不存在，视为已卸载。"
  exit 0
fi

stop_runner() {
  if [[ "$USE_SYSTEMD" == true && -f "$RUNNER_DIR/svc.sh" ]]; then
    log "停止 systemd Runner 服务"
    sudo "$RUNNER_DIR/svc.sh" stop || true
    sudo "$RUNNER_DIR/svc.sh" uninstall || true
  elif command -v tmux >/dev/null 2>&1; then
    if tmux has-session -t "$RUNNER_TMUX_SESSION" 2>/dev/null; then
      log "终止 tmux 会话 $RUNNER_TMUX_SESSION"
      tmux kill-session -t "$RUNNER_TMUX_SESSION"
    fi
  fi
}

obtain_removal_token() {
  if [[ -n "${GH_RUNNER_REMOVE_TOKEN:-}" ]]; then
    echo "$GH_RUNNER_REMOVE_TOKEN"
    return
  fi
  if [[ -z "${GH_RUNNER_PAT:-}" ]]; then
    fail "缺少 GH_RUNNER_PAT，无法通过 API 获取 removal token"
  fi
  log "向 GitHub API 申请 removal token"
  local response
  response="$(curl -fsSL -X POST \
    -H "Authorization: token $GH_RUNNER_PAT" \
    -H "Accept: application/vnd.github+json" \
    "$GITHUB_API_URL/repos/$RUNNER_REPO/actions/runners/remove-token")"
  token="$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("token",""))' <<<"$response")"
  [[ -n "$token" ]] || fail "无法解析 removal token，响应：$response"
  echo "$token"
}

remove_runner_registration() {
  if [[ -f "$RUNNER_DIR/config.sh" ]]; then
    local removal_token
    removal_token="$(obtain_removal_token)"
    log "执行 ./config.sh remove --token *****"
    (cd "$RUNNER_DIR" && ./config.sh remove --token "$removal_token" || true)
  else
    log "未找到 config.sh，跳过取消注册"
  fi
}

backup_and_cleanup() {
  local backup_dir="${RUNNER_DIR}.bak-$(date +%Y%m%dT%H%M%S)"
  mv "$RUNNER_DIR" "$backup_dir"
  log "Runner 文件已移动至 $backup_dir（可留作审计或后续清理）"
}

log "==== 卸载 WSL Runner (${RUNNER_DIR}) ===="
stop_runner
remove_runner_registration
backup_and_cleanup
log "卸载完成。请在 GitHub → Settings → Actions → Runners 中确认节点已删除，并更新 Plan 265/266/269。"
