#!/usr/bin/env bash
set -euo pipefail

# Detect REST reads bypassing GraphQL in frontend code
# Exit non-zero only if ENFORCE=1 (default: 0)

ENFORCE=${ENFORCE:-0}
ROOT_DIR=$(cd "$(dirname "$0")/../.." && pwd)
FRONTEND="$ROOT_DIR/frontend/src"

echo "[cqrs] scanning for REST read queries in frontend..."

issues=0

# 1) direct fetch to /api/v1 in frontend src
fetch_hits=$(grep -RIn --include='*.ts' --include='*.tsx' -E "fetch\(.*['\"]/api/v1" "$FRONTEND" || true)
if [[ -n "$fetch_hits" ]]; then
  echo "[cqrs][WARN] direct REST fetch calls in frontend detected:" >&2
  echo "$fetch_hits" >&2
  issues=$((issues+1))
fi

# 2) usage of REST organizations client for queries
org_client_imports=$(grep -RIn --include='*.ts' --include='*.tsx' \
  -E "from\s+['\"][.]{0,2}/shared/api/organizations['\"]" "$FRONTEND" || true)
if [[ -n "$org_client_imports" ]]; then
  echo "[cqrs][WARN] imports from shared/api/organizations (REST) found — ensure used only for commands:" >&2
  echo "$org_client_imports" >&2
  issues=$((issues+1))
fi

if [[ "$issues" -gt 0 && "$ENFORCE" == "1" ]]; then
  echo "[cqrs][FAIL] REST read path usage detected (set ENFORCE=0 to report-only)" >&2
  exit 3
fi

echo "[cqrs] report-only completed (issues=$issues)"
exit 0

