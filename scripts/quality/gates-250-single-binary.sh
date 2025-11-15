#!/usr/bin/env bash
set -euo pipefail
# Purpose: Ensure only one non-legacy main binary is present under ./cmd
# Strategy: count main.go without //go:build legacy
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

count=$(
  rg -l '^package main\b' cmd 2>/dev/null \
    | while read -r f; do
        if ! rg -q '^//go:build\s+legacy' "$f"; then
          echo "$f"
        fi
      done | wc -l | tr -d ' '
)
echo "[gate-250] non-legacy main count: $count"
if [[ "$count" -ne 1 ]]; then
  echo "[gate-250][FAIL] expected exactly 1 non-legacy main under ./cmd (monolith), found: $count"
  exit 1
fi
echo "[gate-250][OK] single binary gate passed."

