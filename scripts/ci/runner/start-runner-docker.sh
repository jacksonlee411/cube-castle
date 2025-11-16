#!/usr/bin/env bash
set -euo pipefail
#
# start-runner-docker.sh
# - 构建带 docker/compose 的自定义 Runner 镜像（基于 ghcr actions-runner）
# - 申请注册 token 并以持久化方式启动 Runner 容器
#

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
REPO_URL="$(git -C "$ROOT" remote get-url origin 2>/dev/null || true)"
if [[ -z "$REPO_URL" ]]; then
  echo "❌ 无法解析 origin URL" >&2; exit 2
fi
OWNER="$(echo "$REPO_URL" | sed -E 's#^ssh://git@[^/]+/##; s#^git@[^:]+:##; s#^https?://[^/]+/##' | cut -d'/' -f1)"
REPO="$(echo "$REPO_URL" | sed -E 's#^ssh://git@[^/]+/##; s#^git@[^:]+:##; s#^https?://[^/]+/##' | cut -d'/' -f2 | sed 's/.git$//')"
OWNER_REPO="${OWNER}/${REPO}"

load_env(){ [ -f "$1" ] && set -a && . "$1" && set +a || true; }
load_env "${ROOT}/secrets/.env.local"
load_env "${ROOT}/secrets/.env"
load_env "${ROOT}/.env.local"
load_env "${ROOT}/.env"

PAT="${GH_RUNNER_PAT:-${GITHUB_TOKEN:-}}"
if [[ -z "$PAT" ]]; then
  echo "❌ 缺少 GH_RUNNER_PAT/GITHUB_TOKEN（需要 repo scope）" >&2; exit 3
fi

echo "🐳 构建自定义 Runner 镜像（含 docker/compose）..."
docker build -t cc-actions-runner-docker:2.315.0 -f "${ROOT}/runner/Dockerfile.docker" "${ROOT}/runner"

echo "🔑 申请注册令牌..."
TOKEN_JSON="$(curl -fsSL -X POST -H "Authorization: Bearer ${PAT}" -H "Accept: application/vnd.github+json" "https://api.github.com/repos/${OWNER_REPO}/actions/runners/registration-token")"
RUNNER_TOKEN="$(echo "$TOKEN_JSON" | jq -r '.token // empty')"
if [[ -z "$RUNNER_TOKEN" ]]; then
  echo "❌ 获取注册令牌失败：$TOKEN_JSON" >&2; exit 4
fi

echo "🚀 启动持久化 Runner（自定义镜像）..."
docker rm -f cubecastle-gh-runner >/dev/null 2>&1 || true
RUNNER_TOKEN="$RUNNER_TOKEN" docker compose -f "${ROOT}/docker-compose.runner.docker.yml" up -d

echo "⏳ 等待 Runner 就绪（最长 90s）..."
for i in {1..90}; do
  sleep 1
  if docker logs cubecastle-gh-runner 2>&1 | grep -Eq "Listening for Jobs|Connected to GitHub|Runner reconfigured and ready to work"; then
    echo "✅ Runner 在线（docker/compose 可用）"
    docker exec cubecastle-gh-runner docker version >/dev/null 2>&1 && docker exec cubecastle-gh-runner docker compose version || true
    exit 0
  fi
done

echo "⚠️ Runner 未确认就绪，请查看日志：docker logs -f cubecastle-gh-runner"
exit 5
