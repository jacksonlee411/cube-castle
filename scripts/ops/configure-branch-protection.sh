#!/usr/bin/env bash
#
# Configure GitHub repo for "Trunk (local) + Remote PR guard"
# - Enforce protected default branch (no direct push)
# - Allow only squash-merge, delete branch on merge
# - Require key status checks (contexts) before merge
#
# Requirements:
# - gh CLI installed and authenticated (GH_TOKEN/GITHUB_TOKEN or gh auth login)
# - Network access
#
# Note:
# - Contexts are derived from this repo's workflows. Adjust list below if workflow/job names change.
# - This script reads the actual default branch from the repository (master/main).
set -euo pipefail

echo "🔐 Configuring branch protection and merge strategy..."

if ! command -v gh >/dev/null 2>&1; then
  echo "❌ gh CLI not found. Install from https://github.com/cli/cli and ensure authentication."
  exit 1
fi

echo "🔎 Checking gh authentication..."
if ! gh auth status -h github.com >/dev/null 2>&1; then
  echo "❌ gh not authenticated. Set GH_TOKEN/GITHUB_TOKEN or run: gh auth login"
  exit 1
fi

OWNER_REPO="$(gh repo view --json nameWithOwner -q .nameWithOwner)"
DEFAULT_BRANCH="$(gh repo view --json defaultBranchRef -q .defaultBranchRef.name)"
echo "📦 Repository: ${OWNER_REPO}"
echo "🌿 Default branch: ${DEFAULT_BRANCH}"

# Required status checks (contexts). Keep minimal and stable to avoid false blocks.
# Format: "<workflow_name> / <job_name or job display name>"
CONTEXTS=(
  "plan-258-gates / Contract Drift Gate (Plan 258)"
  "plan-253-gates / compose-and-images"
  "plan-250-gates / gates-250"
  "plan-255-gates / gates-255"
  "CI / build-and-test"
)

echo "✅ Will require the following status checks before merge:"
for c in "${CONTEXTS[@]}"; do
  echo "  - ${c}"
done

TMP_JSON="$(mktemp)"
{
  echo '{'
  echo '  "required_status_checks": {'
  echo '    "strict": true,'
  echo '    "contexts": ['
  for i in "${!CONTEXTS[@]}"; do
    SEP=$([ "$i" -lt "$(( ${#CONTEXTS[@]} - 1 ))" ] && echo "," || echo "")
    printf '      "%s"%s\n' "${CONTEXTS[$i]}" "${SEP}"
  done
  echo '    ]'
  echo '  },'
  echo '  "enforce_admins": true,'
  echo '  "required_pull_request_reviews": {'
  echo '    "required_approving_review_count": 0,'
  echo '    "require_code_owner_reviews": false'
  echo '  },'
  echo '  "restrictions": null,'
  echo '  "allow_force_pushes": false,'
  echo '  "allow_deletions": false,'
  echo '  "required_linear_history": true'
  echo '}'
} > "${TMP_JSON}"

echo "🧭 Enforcing merge strategy: squash-only + delete branch on merge..."
gh api -X PATCH "repos/${OWNER_REPO}" \
  -f allow_squash_merge=true \
  -f allow_merge_commit=false \
  -f allow_rebase_merge=false \
  -f delete_branch_on_merge=true >/dev/null

echo "🛡️  Applying protection to ${DEFAULT_BRANCH}..."
gh api --method PUT "repos/${OWNER_REPO}/branches/${DEFAULT_BRANCH}/protection" \
  -H "Accept: application/vnd.github+json" \
  --input "${TMP_JSON}" >/dev/null

rm -f "${TMP_JSON}"
echo "🎉 Done. Default branch is protected; only squash-merge via PR is allowed."

