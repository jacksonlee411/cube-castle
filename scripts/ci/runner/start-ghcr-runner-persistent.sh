#!/usr/bin/env bash
set -euo pipefail
#
# start-ghcr-runner-persistent.sh
# 持久化（非 Ephemeral）Runner：注册后常驻接单，改为由 docker compose 管控
# 依赖：secrets/.env.local 中提供 GH_RUNNER_PAT 或 GITHUB_TOKEN（scope: repo）
#

REPO_URL="$(git remote get-url origin 2>/dev/null || true)"
if [[ -z "$REPO_URL" ]]; then
  echo "❌ 无法解析 origin URL" >&2; exit 2
fi
OWNER="$(echo "$REPO_URL" | sed -E 's#^ssh://git@[^/]+/##; s#^git@[^:]+:##; s#^https?://[^/]+/##' | cut -d'/' -f1)"
REPO="$(echo "$REPO_URL" | sed -E 's#^ssh://git@[^/]+/##; s#^git@[^:]+:##; s#^https?://[^/]+/##' | cut -d'/' -f2 | sed 's/.git$//')"
OWNER_REPO="${OWNER}/${REPO}"

load_env(){ [ -f "$1" ] && set -a && . "$1" && set +a || true; }
load_env "secrets/.env.local"
load_env "secrets/.env"
load_env ".env.local"
load_env ".env"

PAT="${GH_RUNNER_PAT:-${GITHUB_TOKEN:-}}"
if [[ -z "$PAT" ]]; then
  echo "❌ 缺少 GH_RUNNER_PAT/GITHUB_TOKEN（需要 repo scope）" >&2; exit 3
fi

echo "🔑 申请仓库注册令牌: ${OWNER_REPO}"
TOKEN_JSON="$(curl -fsSL -X POST -H "Authorization: Bearer ${PAT}" -H "Accept: application/vnd.github+json" "https://api.github.com/repos/${OWNER_REPO}/actions/runners/registration-token")"
RUNNER_TOKEN="$(echo "$TOKEN_JSON" | jq -r '.token // empty')"
if [[ -z "$RUNNER_TOKEN" ]]; then
  echo "❌ 获取注册令牌失败：$TOKEN_JSON" >&2; exit 4
fi
echo "✅ 已获取注册令牌"

echo "🐳 启动持久化 Runner（compose 管控，非 Ephemeral）..."
RUNNER_TOKEN="$RUNNER_TOKEN" \
GH_RUNNER_PAT="$PAT" \
RUNNER_REPO="$OWNER_REPO" \
RUNNER_NAME="${RUNNER_NAME:-cc-runner-${HOSTNAME}}" \
RUNNER_LABELS="${RUNNER_LABELS:-self-hosted,cubecastle,linux,x64,docker}" \
RUNNER_WORKDIR="${RUNNER_WORKDIR:-/home/runner/_work}" \
docker compose -f docker-compose.runner.persist.yml up -d

echo "⏳ 等待 Runner 就绪（最长 90s）..."
for i in {1..60}; do
  sleep 1
  if docker logs cubecastle-gh-runner 2>&1 | grep -Eq "Listening for Jobs|Connected to GitHub|Runner reconfigured and ready to work"; then
    echo "✅ Runner 在线，持久化接单中"
    docker ps --format '{{.Names}} {{.Status}}' | grep cubecastle-gh-runner || true
    exit 0
  fi
done

echo "⚠️ Runner 未在预期时间内确认就绪，请查看日志：docker logs -f cubecastle-gh-runner"
exit 5
