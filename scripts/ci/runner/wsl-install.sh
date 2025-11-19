#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
WSL GitHub Actions Runner installer (Plan 269)

Usage:
  wsl-install.sh [options]

Options:
  --repo <owner/repo>         目标仓库（默认 jacksonlee411/cube-castle）
  --runner-dir <path>         Runner 解压目录（默认 $HOME/actions-runner）
  --runner-version <ver>      actions/runner 版本（默认 2.319.1）
  --labels <csv>              Runner 标签（默认 self-hosted,cubecastle,wsl）
  --name <runner-name>        Runner 名称（默认 wsl-$(hostname)）
  --work <path>               `_work` 目录（默认 <runner-dir>/_work）
  --use-systemd               若 WSL 已启用 systemd，使用 svc.sh 安装守护
  --tmux-session <name>       非 systemd 模式下的 tmux 会话名（默认 cc-runner）
  --env-file <path>           secrets 路径（默认 <repo>/secrets/.env.local）
  --force-reconfigure         已存在配置时强制重新注册（需要 GH_RUNNER_PAT）
  -h, --help                  查看帮助
EOF
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
LOG_DIR="${RUNNER_LOG_DIR:-$REPO_ROOT/logs/wsl-runner}"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/install-$(date +%Y%m%dT%H%M%S).log"
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

RUNNER_REPO="${RUNNER_REPO:-jacksonlee411/cube-castle}"
RUNNER_DIR="${RUNNER_DIR:-$HOME/actions-runner}"
RUNNER_VERSION="${RUNNER_VERSION:-2.330.0}"
RUNNER_NAME="${RUNNER_NAME:-wsl-$(hostname)}"
RUNNER_LABELS="${RUNNER_LABELS:-self-hosted,cubecastle,wsl}"
RUNNER_WORKDIR="${RUNNER_WORKDIR:-$RUNNER_DIR/_work}"
RUNNER_GROUP="${RUNNER_GROUP:-Default}"
RUNNER_TMUX_SESSION="${RUNNER_TMUX_SESSION:-cc-runner}"
ENV_FILE_DEFAULT="$REPO_ROOT/secrets/.env.local"
ENV_FILE="${ENV_FILE:-$ENV_FILE_DEFAULT}"
USE_SYSTEMD=false
FORCE_RECONFIGURE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo) RUNNER_REPO="$2"; shift 2;;
    --runner-dir) RUNNER_DIR="$2"; shift 2;;
    --runner-version) RUNNER_VERSION="$2"; shift 2;;
    --labels) RUNNER_LABELS="$2"; shift 2;;
    --name) RUNNER_NAME="$2"; shift 2;;
    --work) RUNNER_WORKDIR="$2"; shift 2;;
    --use-systemd) USE_SYSTEMD=true; shift;;
    --tmux-session) RUNNER_TMUX_SESSION="$2"; shift 2;;
    --env-file) ENV_FILE="$2"; shift 2;;
    --force-reconfigure|--replace) FORCE_RECONFIGURE=true; shift;;
    -h|--help) usage; exit 0;;
    *) fail "未知参数: $1";;
  esac
done

if [[ -f "$ENV_FILE" ]]; then
  log "加载环境变量: $ENV_FILE"
  # shellcheck disable=SC1090
  set -a && source "$ENV_FILE" && set +a
else
  log "未找到 env 文件（$ENV_FILE），将依赖当前环境变量"
fi

RUNNER_URL="${RUNNER_URL:-https://github.com/${RUNNER_REPO}}"
GITHUB_API_URL="${GITHUB_API_URL:-https://api.github.com}"
DOWNLOAD_URL="https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/actions-runner-linux-x64-${RUNNER_VERSION}.tar.gz"

if [[ -n "${RUNNER_HTTP_PROXY:-}" ]]; then
  export HTTP_PROXY="$RUNNER_HTTP_PROXY"
  export HTTPS_PROXY="${RUNNER_HTTPS_PROXY:-$RUNNER_HTTP_PROXY}"
  export NO_PROXY="${RUNNER_NO_PROXY:-${NO_PROXY:-}}"
fi
if [[ -n "${RUNNER_HTTPS_PROXY:-}" ]]; then
  export HTTPS_PROXY="$RUNNER_HTTPS_PROXY"
fi
if [[ -n "${RUNNER_NO_PROXY:-}" ]]; then
  export NO_PROXY="$RUNNER_NO_PROXY"
fi

require_cmd curl
require_cmd tar
require_cmd python3
require_cmd go
require_cmd node
require_cmd docker
require_cmd tee
if [[ "$USE_SYSTEMD" != true ]]; then
  require_cmd tmux
fi

GO_VERSION="$(go version | awk '{print $3}' | sed 's/go//')"
NODE_VERSION_STR="$(node --version | sed 's/^v//')"
ensure_version "Go" "$GO_VERSION" "1.24.0"
ensure_version "Node.js" "$NODE_VERSION_STR" "18.0.0"

log "检查 Docker CLI / Compose"
docker version >/dev/null
docker compose version >/dev/null
docker info >/dev/null

mkdir -p "$RUNNER_DIR" "$RUNNER_WORKDIR"

download_runner() {
  if [[ -x "$RUNNER_DIR/run.sh" ]] && [[ "$FORCE_RECONFIGURE" != true ]]; then
    log "检测到现有 Runner 可执行文件，跳过下载（使用 --force-reconfigure 重新下载）"
    return
  fi
  log "下载 GitHub Actions Runner ${RUNNER_VERSION}"
  local tmp_dir
  tmp_dir="$(mktemp -d)"
  curl -fSL "$DOWNLOAD_URL" -o "$tmp_dir/actions-runner.tar.gz"
  tar -xzf "$tmp_dir/actions-runner.tar.gz" -C "$RUNNER_DIR"
  rm -rf "$tmp_dir"
}

obtain_registration_token() {
  if [[ -n "${RUNNER_TOKEN:-}" ]]; then
    echo "$RUNNER_TOKEN"
    return
  fi
  if [[ -n "${GH_RUNNER_REG_TOKEN:-}" ]]; then
    echo "$GH_RUNNER_REG_TOKEN"
    return
  fi
  if [[ -z "${GH_RUNNER_PAT:-}" ]]; then
    fail "缺少 GH_RUNNER_PAT 或 RUNNER_TOKEN，无法申请 registration token"
  fi
  log "通过 GitHub API 获取 registration token"
  local response
  response="$(curl -fsSL -X POST \
    -H "Authorization: token $GH_RUNNER_PAT" \
    -H "Accept: application/vnd.github+json" \
    "$GITHUB_API_URL/repos/$RUNNER_REPO/actions/runners/registration-token")"
  token="$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("token",""))' <<<"$response")"
  [[ -n "$token" ]] || fail "无法解析 registration token，响应：$response"
  echo "$token"
}

obtain_removal_token() {
  if [[ -n "${GH_RUNNER_REMOVE_TOKEN:-}" ]]; then
    echo "$GH_RUNNER_REMOVE_TOKEN"
    return
  fi
  if [[ -z "${GH_RUNNER_PAT:-}" ]]; then
    fail "缺少 GH_RUNNER_PAT，无法获取 removal token（--force-reconfigure 需要 PAT）"
  fi
  log "通过 GitHub API 获取 removal token"
  local response
  response="$(curl -fsSL -X POST \
    -H "Authorization: token $GH_RUNNER_PAT" \
    -H "Accept: application/vnd.github+json" \
    "$GITHUB_API_URL/repos/$RUNNER_REPO/actions/runners/remove-token")"
  token="$(python3 -c 'import json,sys; print(json.load(sys.stdin).get("token",""))' <<<"$response")"
  [[ -n "$token" ]] || fail "无法解析 removal token，响应：$response"
  echo "$token"
}

stop_runner_service() {
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

maybe_remove_previous_config() {
  if [[ "$FORCE_RECONFIGURE" == true && -f "$RUNNER_DIR/.runner" ]]; then
    stop_runner_service
    local removal_token
    removal_token="$(obtain_removal_token)"
    log "执行 ./config.sh remove --token *****"
    (cd "$RUNNER_DIR" && ./config.sh remove --token "$removal_token")
  fi
}

stop_tmux_session() {
  if command -v tmux >/dev/null 2>&1; then
    if tmux has-session -t "$RUNNER_TMUX_SESSION" 2>/dev/null; then
      log "终止旧 tmux 会话 $RUNNER_TMUX_SESSION"
      tmux kill-session -t "$RUNNER_TMUX_SESSION"
    fi
  fi
}

configure_runner() {
  if [[ -f "$RUNNER_DIR/.runner" && "$FORCE_RECONFIGURE" != true ]]; then
    log "检测到现有 Runner 配置，跳过 config（使用 --force-reconfigure 重新配置）"
    return
  fi
  local registration_token
  registration_token="$(obtain_registration_token)"
  log "执行 ./config.sh --unattended --url $RUNNER_URL"
  (cd "$RUNNER_DIR" && ./config.sh --unattended \
    --replace \
    --url "$RUNNER_URL" \
    --token "$registration_token" \
    --name "$RUNNER_NAME" \
    --labels "$RUNNER_LABELS" \
    --runnergroup "$RUNNER_GROUP" \
    --work "$RUNNER_WORKDIR" \
    --disableupdate)
}

start_runner() {
  if [[ "$USE_SYSTEMD" == true ]]; then
    [[ -f "$RUNNER_DIR/svc.sh" ]] || fail "未找到 svc.sh，无法安装 systemd 服务"
    log "使用 systemd 安装/启动 Runner 服务"
    sudo "$RUNNER_DIR/svc.sh" install || true
    sudo "$RUNNER_DIR/svc.sh" start
    systemctl status "actions.runner.${RUNNER_REPO//\//-}.${RUNNER_NAME}.service" || true
  else
    require_cmd tmux
    local run_log="$LOG_DIR/run-$(date +%Y%m%dT%H%M%S).log"
    stop_tmux_session
    log "启动 tmux 会话 $RUNNER_TMUX_SESSION，日志输出至 $run_log"
    tmux new -d -s "$RUNNER_TMUX_SESSION" "cd \"$RUNNER_DIR\" && ./run.sh >>\"$run_log\" 2>&1"
    tmux ls
  fi
}

log "==== Plan 269 WSL Runner 安装 ===="
log "目标仓库: $RUNNER_REPO"
log "Runner 目录: $RUNNER_DIR"
log "Runner 标签: $RUNNER_LABELS"
download_runner
maybe_remove_previous_config
configure_runner
start_runner

log "安装完成。请在 GitHub → Settings → Actions → Runners 确认 '${RUNNER_NAME}' (labels: ${RUNNER_LABELS}) 在线，并在 Plan 265/266/269 记录 Run ID + 日志。"
