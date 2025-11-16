#!/usr/bin/env bash
set -euo pipefail

# Lint only changed TS/TSX files under frontend/src, and fail on any no-restricted-syntax warnings.
# This enforces "no hard-coded data-testid" as errors for changed files (suggestion #1).
#
# ENV:
#   BASE_REF: git base ref to diff against (default: origin/master)
#   HEAD_REF: git head ref (default: HEAD)
#
# Usage:
#   npm run lint:changed:testid
#

BASE_REF="${BASE_REF:-origin/master}"
HEAD_REF="${HEAD_REF:-HEAD}"

echo "[lint-changed] Diff range: ${BASE_REF}...${HEAD_REF}"

mapfile -t FILES < <(git diff --name-only "${BASE_REF}...${HEAD_REF}" -- 'frontend/src/**/*.ts' 'frontend/src/**/*.tsx' | tr -d '\r' | sed '/^\s*$/d')

if [[ ${#FILES[@]} -eq 0 ]]; then
  echo "[lint-changed] No changed TS/TSX files under frontend/src"
  exit 0
fi

echo "[lint-changed] Changed files:"
for f in "${FILES[@]}"; do
  echo " - $f"
done

# Run ESLint from the frontend workspace, escalate no-restricted-syntax to error and fail on warnings.
pushd frontend >/dev/null
  # Convert absolute/relative paths to paths relative to frontend/
  REL_FILES=()
  for f in "${FILES[@]}"; do
    # strip leading 'frontend/' if present
    REL="${f#frontend/}"
    REL_FILES+=("$REL")
  done
  echo "[lint-changed] Running ESLint with no-restricted-syntax=error on changed files..."
  npx eslint --max-warnings 0 --rule 'no-restricted-syntax:error' "${REL_FILES[@]}"
popd >/dev/null

echo "[lint-changed] Completed: all changed files passed strict selector rules."
