#!/usr/bin/env bash
#
# Configure branch protection and required checks using gh CLI.
# Requirements:
#  - gh CLI installed and authenticated (`gh auth status`)
#  - Maintainer/admin permission on the target repository
#
# This script sets:
#  - Branch: master (default)
#  - Enforce admins: true
#  - Require status checks: strict, contexts = gates-250, Compose/Image Gates (Blocking), gates-255, Plan 254 Gate – ubuntu
#  - Require PR review: dismiss stale reviews, require conversation resolution
#  - Disallow force pushes/deletions
#
set -euo pipefail

BRANCH="${BRANCH:-master}"

if ! command -v gh >/dev/null 2>&1; then
  echo "❌ gh CLI not found. Install: https://cli.github.com/" >&2
  exit 2
fi

if ! gh auth status >/dev/null 2>&1; then
  echo "❌ gh not authenticated. Run: gh auth login" >&2
  exit 2
fi

REPO_URL="$(git remote get-url origin 2>/dev/null || true)"
if [[ -z "$REPO_URL" ]]; then
  echo "❌ Cannot determine origin remote" >&2
  exit 2
fi
OWNER="$(echo "$REPO_URL" | sed -E 's#^ssh://git@[^/]+/##; s#^git@[^:]+:##; s#^https?://[^/]+/##' | cut -d/ -f1)"
NAME="$(basename "$REPO_URL" .git)"
REPO="${OWNER}/${NAME}"

echo "🔒 Configuring branch protection for ${REPO}@${BRANCH}"
echo "ℹ️  Required checks to set:"
checks=("gates-250" "Compose/Image Gates (Blocking)" "gates-255" "Plan 254 Gate – ubuntu")
for c in "${checks[@]}"; do echo " - $c"; done

TMP="$(mktemp)"
cat >"$TMP" <<JSON
{
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "gates-250",
      "Compose/Image Gates (Blocking)",
      "gates-255",
      "Plan 254 Gate – ubuntu",
      "PR Body Policy – required"
    ]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false
  },
  "restrictions": null,
  "required_linear_history": true,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_conversation_resolution": true
}
JSON

set +e
OUT=$(gh api -X PUT -H "Accept: application/vnd.github+json" \
  "repos/${REPO}/branches/${BRANCH}/protection" \
  --input "$TMP" 2>&1)
RC=$?
set -e
rm -f "$TMP"

if [ $RC -ne 0 ]; then
  echo "⚠️  Failed to set branch protection via API."
  echo "$OUT"
  echo ""
  echo "👉 请手动到 GitHub Settings → Branches，为 ${BRANCH} 启用保护并添加 Required checks："
  for c in "${checks[@]}"; do echo "   - $c"; done
  exit 1
fi

echo "✅ Branch protection configured for ${REPO}@${BRANCH}"
