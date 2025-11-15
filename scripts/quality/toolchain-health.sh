#!/usr/bin/env bash
set -euo pipefail

# Cube Castle - Toolchain Health Check
# Verifies Node, npm registry, ESLint/TS/GraphQL/Playwright versions for root and frontend
#
# Usage:
#   scripts/quality/toolchain-health.sh [--scope all|root|frontend]
#

SCOPE="${1:-all}"
if [[ "$1" == "--scope" ]]; then
  SCOPE="${2:-all}"
fi

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

green() { printf "\033[32m%s\033[0m\n" "$*"; }
red() { printf "\033[31m%s\033[0m\n" "$*" >&2; }
yellow() { printf "\033[33m%s\033[0m\n" "$*"; }
info() { printf "\033[34m%s\033[0m\n" "$*"; }

failures=0

check_node() {
  info "Node version check (>=18)"
  if ! command -v node >/dev/null 2>&1; then
    red "Node not found"
    failures=$((failures+1))
    return
  fi
  local ver_major
  ver_major=$(node -p "process.versions.node.split('.')[0]")
  if [[ "$ver_major" -lt 18 ]]; then
    red "Node major=$ver_major < 18"
    failures=$((failures+1))
  else
    green "Node OK (v$(node -v))"
  fi
}

check_registry() {
  info "npm registry check (registry.npmjs.org)"
  local reg=""
  if command -v npm >/dev/null 2>&1; then
    reg=$(npm config get registry 2>/dev/null || echo "")
  fi
  if [[ -z "$reg" ]]; then
    if [[ -f "$ROOT_DIR/.npmrc" ]]; then
      reg=$(grep -E '^registry=' "$ROOT_DIR/.npmrc" | head -n1 | cut -d'=' -f2- || true)
    fi
  fi
  if [[ "$reg" == "https://registry.npmjs.org/" ]]; then
    green "Registry OK ($reg)"
  else
    yellow "Registry not normalized (detected: ${reg:-unknown}); See .npmrc"
  fi
}

pkg_ver() {
  local file="$1" key="$2"
  node -e "console.log((require('$file').devDependencies||{})['$key'] || (require('$file').dependencies||{})['$key'] || '')" 2>/dev/null
}

check_root() {
  info "Root toolchain"
  local pkg="$ROOT_DIR/package.json"
  local eslint_ver ts_eslint_ver tsparser_ver ts_ver gql_ver
  eslint_ver="$(pkg_ver "$pkg" "eslint")"
  ts_eslint_ver="$(pkg_ver "$pkg" "@typescript-eslint/eslint-plugin")"
  tsparser_ver="$(pkg_ver "$pkg" "@typescript-eslint/parser")"
  ts_ver="$(pkg_ver "$pkg" "typescript")"
  gql_ver="$(pkg_ver "$pkg" "graphql")"

  echo "  eslint:           $eslint_ver (expect ^9)"
  echo "  @ts-eslint/plugin $ts_eslint_ver (expect ^8)"
  echo "  @ts-eslint/parser $tsparser_ver (expect ^8)"
  echo "  typescript:       $ts_ver (expect ^5.8)"
  echo "  graphql:          $gql_ver (expect ^16)"

  [[ "$eslint_ver" =~ ^\^?9\. ]] || { red "  eslint not ^9"; failures=$((failures+1)); }
  [[ "$ts_eslint_ver" =~ ^\^?8\. ]] || { red "  @typescript-eslint/eslint-plugin not ^8"; failures=$((failures+1)); }
  [[ "$tsparser_ver" =~ ^\^?8\. ]] || { red "  @typescript-eslint/parser not ^8"; failures=$((failures+1)); }
  [[ "$ts_ver" =~ ^\^?5\.8\. ]] || { red "  typescript not ^5.8"; failures=$((failures+1)); }
  [[ "$gql_ver" =~ ^\^?16\. ]] || { red "  graphql not ^16"; failures=$((failures+1)); }
}

check_frontend() {
  info "Frontend toolchain"
  local pkg="$ROOT_DIR/frontend/package.json"
  local pw_ver gql_ver fe_eslint_ver
  pw_ver="$(pkg_ver "$pkg" "@playwright/test")"
  gql_ver="$(pkg_ver "$pkg" "graphql")"
  fe_eslint_ver="$(pkg_ver "$pkg" "eslint")"
  echo "  @playwright/test: $pw_ver (expect ^1.56)"
  echo "  graphql:          $gql_ver (expect ^16)"
  echo "  eslint:           $fe_eslint_ver (expect ^9)"
  [[ "$pw_ver" =~ ^\^?1\.56 ]] || { red "  @playwright/test not ^1.56"; failures=$((failures+1)); }
  [[ "$gql_ver" =~ ^\^?16\. ]] || { red "  graphql not ^16"; failures=$((failures+1)); }
  [[ "$fe_eslint_ver" =~ ^\^?9\. ]] || { red "  frontend eslint not ^9"; failures=$((failures+1)); }
}

check_node
check_registry
case "$SCOPE" in
  root) check_root ;;
  frontend) check_frontend ;;
  all|*) check_root; check_frontend ;;
esac

if [[ "$failures" -gt 0 ]]; then
  red "Toolchain health check failed ($failures issue(s))"
  exit 2
else
  green "Toolchain health OK"
fi
