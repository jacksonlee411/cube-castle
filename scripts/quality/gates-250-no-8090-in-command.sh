#!/usr/bin/env bash
set -euo pipefail
# Purpose: ensure command server does not default/bind to 8090
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CMD_MAIN="$ROOT_DIR/cmd/hrms-server/command/main.go"
if rg -n '":?8090"' "$CMD_MAIN" >/dev/null 2>&1; then
  echo "[gate-250][FAIL] found 8090 literal in command main: $CMD_MAIN"
  exit 1
fi
echo "[gate-250][OK] no 8090 literal in command main."

