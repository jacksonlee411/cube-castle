#!/usr/bin/env bash
set -euo pipefail

# Report usage of deprecated permission names like org:write
# Exit non-zero only if ENFORCE=1 (default: 0)

ENFORCE=${ENFORCE:-0}
ROOT_DIR=$(cd "$(dirname "$0")/../.." && pwd)

echo "[permissions] scanning for deprecated permission tokens..."

# Grep code (exclude docs, CHANGELOGs, node_modules, test reports)
matches=$(grep -RIn --exclude-dir="node_modules" --exclude-dir=".git" \
  --exclude-dir="docs" --exclude-dir="frontend/tests/contract" \
  -E "org:write" "$ROOT_DIR" || true)

if [[ -n "$matches" ]]; then
  echo "[permissions][WARN] found deprecated permission 'org:write' occurrences:" >&2
  echo "$matches" >&2
  if [[ "$ENFORCE" == "1" ]]; then
    echo "[permissions][FAIL] deprecated permissions detected (set ENFORCE=0 to report-only)" >&2
    exit 2
  fi
else
  echo "[permissions] OK — no deprecated permission found"
fi

exit 0

