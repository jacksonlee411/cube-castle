#!/usr/bin/env bash
set -euo pipefail

echo "[IIG] Auth Uniqueness Guard: scanning repository..."

fail=false

# Backend guard: forbid direct jwt.Parse and signing method checks outside authoritative files
echo "- Backend: checking forbidden direct JWT parsing..."
backend_hits=$(rg -n "jwt\.Parse\(|SigningMethodHMAC|SigningMethodRSA" \
  --glob '!vendor/**' --glob '!**/*.md' --hidden \
  -g '!internal/auth/jwt.go' -g '!internal/auth/jwks.go' || true)
if [[ -n "${backend_hits}" ]]; then
  echo "[FAIL] Found direct JWT parsing/signing usage outside internal/auth/jwt.go:"
  echo "${backend_hits}"
  fail=true
fi

# Backend guard: forbid re-introducing legacy validator
if [[ -f internal/auth/validator.go ]]; then
  echo "[FAIL] Duplicate validator detected: internal/auth/validator.go must not exist"
  fail=true
fi

# Frontend guard: forbid ad-hoc token parsing/persistence outside shared/api/auth.ts
echo "- Frontend: checking duplicate token handling..."
fe_forbidden=$(rg -n "(jwtDecode|jwt-decode|parseJwt|atob\(|localStorage\.(getItem|setItem).*(token|jwt))" frontend \
  -g '!frontend/src/shared/api/auth.ts' --hidden || true)
if [[ -n "${fe_forbidden}" ]]; then
  echo "[FAIL] Found token parsing/storage outside frontend/src/shared/api/auth.ts:"
  echo "${fe_forbidden}"
  fail=true
fi

if [[ "${fail}" == true ]]; then
  echo "[IIG] Auth Uniqueness Guard: violations detected. Failing."
  exit 1
fi

echo "[IIG] Auth Uniqueness Guard: no violations found."
exit 0

