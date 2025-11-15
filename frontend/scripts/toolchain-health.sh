#!/usr/bin/env bash
set -euo pipefail

# Frontend Toolchain Health Check (wrapper)
# Delegates to repository root checker with frontend scope

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

bash "$PROJECT_ROOT/scripts/quality/toolchain-health.sh" --scope frontend
