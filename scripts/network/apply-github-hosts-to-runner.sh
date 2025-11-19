#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
CONFIG_SCRIPT="scripts/network/configure-github-hosts.sh"
COMPOSE_FILE=${COMPOSE_FILE:-docker-compose.runner.persist.yml}
RUNNER_SERVICE=${RUNNER_SERVICE:-gh-runner}
TEMP_PATH="/tmp/configure-github-hosts.sh"

if [[ ! -f "$REPO_ROOT/$CONFIG_SCRIPT" ]]; then
  echo "[network] 找不到 $CONFIG_SCRIPT" >&2
  exit 1
fi

echo "[network] 先更新宿主机 /etc/hosts ..."
sudo bash "$REPO_ROOT/$CONFIG_SCRIPT"

if ! command -v docker >/dev/null 2>&1; then
  echo "[network] docker 命令不存在，跳过 Runner 同步" >&2
  exit 0
fi

if [[ ! -f "$REPO_ROOT/$COMPOSE_FILE" ]]; then
  echo "[network] 未找到 compose 文件: $COMPOSE_FILE，跳过 Runner 同步" >&2
  exit 0
fi

if ! docker compose -f "$REPO_ROOT/$COMPOSE_FILE" ps "$RUNNER_SERVICE" >/dev/null 2>&1; then
  echo "[network] compose service $RUNNER_SERVICE 不存在，跳过 Runner 同步" >&2
  exit 0
fi

echo "[network] 拷贝脚本到 Runner 容器..."
docker compose -f "$REPO_ROOT/$COMPOSE_FILE" cp "$REPO_ROOT/$CONFIG_SCRIPT" "$RUNNER_SERVICE:$TEMP_PATH"

echo "[network] 在 Runner 容器内写入 hosts ..."
docker compose -f "$REPO_ROOT/$COMPOSE_FILE" exec "$RUNNER_SERVICE" bash -lc "sudo bash $TEMP_PATH && sudo rm -f $TEMP_PATH"

echo "[network] GitHub hosts 已同步至宿主与 Runner"
