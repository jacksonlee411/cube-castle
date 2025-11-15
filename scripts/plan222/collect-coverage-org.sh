#!/usr/bin/env bash
# Collect coverage for internal/organization and write artifacts to logs/plan222
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="${ROOT_DIR}/logs/plan222"
mkdir -p "${LOG_DIR}"
STAMP="$(date -u +%Y%m%d-%H%M%S)"
OUT="${LOG_DIR}/coverage-org-${STAMP}.out"
TXT="${LOG_DIR}/coverage-org-${STAMP}.txt"
HTML="${LOG_DIR}/coverage-org-${STAMP}.html"

echo "[Plan222] Running coverage for internal/organization/..."
go test -coverprofile="${OUT}" ./internal/organization/... >/dev/null
echo "[Plan222] Coverprofile: ${OUT}"

echo "[Plan222] Generating text summary..."
go tool cover -func="${OUT}" > "${TXT}"
echo "[Plan222] Text summary: ${TXT}"

echo "[Plan222] Generating HTML report..."
go tool cover -html="${OUT}" -o "${HTML}"
echo "[Plan222] HTML report: ${HTML}"

echo "[Plan222] Done."

