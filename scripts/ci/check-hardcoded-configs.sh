#!/usr/bin/env bash
set -euo pipefail

# Detect hardcoded CORS origins, ports and X-Tenant-ID strings across services
# Exit non-zero only if ENFORCE=1 (default: 0)
# SCAN_SCOPE: all|frontend|cmd (default: all)

ENFORCE=${ENFORCE:-0}
SCAN_SCOPE=${SCAN_SCOPE:-all}
ROOT_DIR=$(cd "$(dirname "$0")/../.." && pwd)

echo "[configs] scanning scope=$SCAN_SCOPE (CORS, ports, tenant header, JWT inline)..."
issues=0

scan_cmd() {
  # AllowedOrigins hardcoded arrays in Go services
  local cors_hits
  cors_hits=$(grep -RIn --include='*.go' -E "AllowedOrigins:\s*\[\]string\{" "$ROOT_DIR/cmd" || true)
  if [[ -n "$cors_hits" ]]; then
    echo "[configs][WARN] hardcoded CORS AllowedOrigins found in Go services:" >&2
    echo "$cors_hits" >&2
    issues=$((issues+1))
  fi

  # Hardcoded localhost ports inside cmd/*
  local port_hits
  port_hits=$(grep -RIn --include='*.go' -E "localhost:(8090|9090)" "$ROOT_DIR/cmd" || true)
  if [[ -n "$port_hits" ]]; then
    echo "[configs][WARN] hardcoded localhost ports found in cmd/*:" >&2
    echo "$port_hits" >&2
    issues=$((issues+1))
  fi

  # Inline JWT env usage (should use internal/config)
  local jwt_hits
  jwt_hits=$(grep -RIn --include='*.go' -E "os.Getenv\(\"JWT_(SECRET|ISSUER|AUDIENCE|ALG|JWKS_URL|PUBLIC_KEY_PATH|ALLOWED_CLOCK_SKEW)\"\)" "$ROOT_DIR/cmd" || true)
  if [[ -n "$jwt_hits" ]]; then
    echo "[configs][WARN] inline JWT env access in cmd/* (use internal/config/jwt.go):" >&2
    echo "$jwt_hits" >&2
    issues=$((issues+1))
  fi
}

scan_frontend() {
  # Hardcoded localhost ports in frontend/src
  local port_hits
  port_hits=$(grep -RIn --include='*.ts' --include='*.tsx' -E "localhost:(8090|9090)" "$ROOT_DIR/frontend/src" || true)
  if [[ -n "$port_hits" ]]; then
    echo "[configs][WARN] hardcoded localhost ports found in frontend/src:" >&2
    echo "$port_hits" >&2
    issues=$((issues+1))
  fi

  # X-Tenant-ID header (ensure reads from config)
  local tenant_header_hits
  tenant_header_hits=$(grep -RIn --include='*.ts' --include='*.tsx' -E "['\"]X-Tenant-ID['\"]\s*:|\['X-Tenant-ID'\]" "$ROOT_DIR/frontend/src" || true)
  if [[ -n "$tenant_header_hits" ]]; then
    echo "[configs][INFO] X-Tenant-ID header usage in frontend (verify origin=shared/config/tenant):" >&2
    echo "$tenant_header_hits" >&2
  fi
}

scan_all() {
  scan_cmd
  scan_frontend
  # Plus: global ports pattern
  local port_hits
  port_hits=$(grep -RIn --include='*.go' --include='*.js' --include='*.ts' --include='*.tsx' -E "localhost:(8090|9090)" "$ROOT_DIR" || true)
  if [[ -n "$port_hits" ]]; then
    echo "[configs][WARN] hardcoded localhost ports found repo-wide:" >&2
    echo "$port_hits" >&2
    issues=$((issues+1))
  fi
}

case "$SCAN_SCOPE" in
  cmd) scan_cmd ;;
  frontend) scan_frontend ;;
  all) scan_all ;;
  *) echo "[configs][ERROR] unknown SCAN_SCOPE=$SCAN_SCOPE" >&2; exit 1 ;;
esac

if [[ "$issues" -gt 0 && "$ENFORCE" == "1" ]]; then
  echo "[configs][FAIL] hardcoded config issues detected (set ENFORCE=0 to report-only)" >&2
  exit 4
fi

echo "[configs] completed (issues=$issues, enforce=$ENFORCE, scope=$SCAN_SCOPE)"
exit 0
