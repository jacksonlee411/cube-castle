#!/usr/bin/env bash
set -euo pipefail
#
# start-ghcr-runner.sh
# - Fetch a repository registration token via GitHub API using PAT (repo scope)
# - Start an ephemeral self-hosted runner based on GHCR official image via compose
# - Waits until the runner is "Listening for Jobs" or times out
#
# Usage:
#   bash scripts/ci/runner/start-ghcr-runner.sh
#

REPO_URL="$(git remote get-url origin 2>/dev/null || true)"
if [[ -z "$REPO_URL" ]]; then
  echo "❌ Cannot resolve origin URL" >&2
  exit 2
fi
OWNER="$(echo "$REPO_URL" | sed -E 's#^ssh://git@[^/]+/##; s#^git@[^:]+:##; s#^https?://[^/]+/##' | cut -d'/' -f1)"
REPO="$(echo "$REPO_URL" | sed -E 's#^ssh://git@[^/]+/##; s#^git@[^:]+:##; s#^https?://[^/]+/##' | cut -d'/' -f2 | sed 's/.git$//')"
OWNER_REPO="${OWNER}/${REPO}"

# Load tokens
load_env() { [ -f "$1" ] && set -a && . "$1" && set +a || true; }
load_env "secrets/.env.local"
load_env "secrets/.env"
load_env ".env.local"
load_env ".env"

PAT="${GH_RUNNER_PAT:-${GITHUB_TOKEN:-}}"
if [[ -z "$PAT" ]]; then
  echo "❌ Missing GH_RUNNER_PAT or GITHUB_TOKEN (requires repo scope) in secrets/.env.local" >&2
  exit 3
fi

API="https://api.github.com/repos/${OWNER_REPO}/actions/runners/registration-token"
echo "🔑 Requesting registration token for ${OWNER_REPO}..."
TOKEN_JSON="$(curl -fsSL -X POST -H "Authorization: Bearer ${PAT}" -H "Accept: application/vnd.github+json" "${API}")"
RUNNER_TOKEN="$(echo "$TOKEN_JSON" | jq -r '.token // empty')"
if [[ -z "$RUNNER_TOKEN" ]]; then
  echo "❌ Failed to obtain registration token. Response:" >&2
  echo "$TOKEN_JSON" >&2
  exit 4
fi
echo "✅ Obtained registration token."

echo "🐳 Pull GHCR official actions-runner image (2.329.0)..."
docker pull ghcr.io/actions/actions-runner:2.329.0

echo "🚀 Starting GHCR-based runner via compose..."
RUNNER_TOKEN="$RUNNER_TOKEN" docker compose -f docker-compose.runner.ghcr.yml up -d

echo "⏳ Waiting for runner to become ready (max 120s)..."
for i in {1..60}; do
  sleep 2
  LOG="$(docker logs cubecastle-gh-runner 2>&1 | tail -n 50 || true)"
  if echo "$LOG" | grep -Eqi "Listening for Jobs|Connected to GitHub|Runner reconfigured and ready to work"; then
    echo "✅ Runner registered and ready."
    exit 0
  fi
done

echo "⚠️  Runner did not confirm readiness within timeout. Check logs: docker logs cubecastle-gh-runner"
exit 5
