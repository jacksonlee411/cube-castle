#!/usr/bin/env bash
set -euo pipefail
#
# List GitHub repository Actions secrets (metadata only; values are not retrievable).
# - Reads token from secrets/.env.local (GITHUB_TOKEN=...)
# - Requires: repo admin permissions for listing secrets metadata
# - Output saved to logs/plan258/github-secrets-<ts>.json
#

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

if [[ ! -f "secrets/.env.local" ]]; then
  echo "✗ secrets/.env.local not found; please provide GITHUB_TOKEN there."
  exit 2
fi

# shellcheck disable=SC1091
source secrets/.env.local
if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "✗ GITHUB_TOKEN not set in secrets/.env.local"
  exit 3
fi

remote_url="$(git remote get-url origin 2>/dev/null || true)"
if [[ -z "$remote_url" ]]; then
  echo "✗ git remote 'origin' not found"
  exit 4
fi

# Extract owner/repo from ssh or https remote
owner=""
repo=""
if [[ "$remote_url" =~ github\.com[:/]+([^/]+)/([^/.]+) ]]; then
  owner="${BASH_REMATCH[1]}"
  repo="${BASH_REMATCH[2]}"
else
  # ssh over 443 form: ssh://git@ssh.github.com:443/owner/repo.git
  if [[ "$remote_url" =~ ssh\.github\.com.*[:/]+([^/]+)/([^/.]+) ]]; then
    owner="${BASH_REMATCH[1]}"
    repo="${BASH_REMATCH[2]}"
  fi
fi

if [[ -z "$owner" || -z "$repo" ]]; then
  echo "✗ Unable to parse owner/repo from: $remote_url"
  exit 5
fi

mkdir -p logs/plan258
ts="$(date +%Y%m%d_%H%M%S)"
out="logs/plan258/github-secrets-${ts}.json"

echo "🔐 Listing Actions repository secrets for $owner/$repo (metadata only)..."
curl -sSf -H "Authorization: Bearer ${GITHUB_TOKEN}" \
     -H "Accept: application/vnd.github+json" \
     "https://api.github.com/repos/${owner}/${repo}/actions/secrets" \
     > "$out"

echo "✅ Saved to $out"
echo "ℹ️  This endpoint returns names/metadata only. Secret values are never retrievable via API."

