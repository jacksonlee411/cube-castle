#!/usr/bin/env bash
set -euo pipefail
# Purpose: Block legacy dual service in CI
if [[ "${ENABLE_LEGACY_DUAL_SERVICE:-}" == "true" ]]; then
  echo "[gate-250][FAIL] ENABLE_LEGACY_DUAL_SERVICE=true is forbidden in CI."
  exit 1
fi
echo "[gate-250][OK] legacy env gate passed."

