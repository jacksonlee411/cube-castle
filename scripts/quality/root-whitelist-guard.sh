#!/usr/bin/env bash
set -euo pipefail
#
# Guard: prevent drifting files at repository root (depth=1).
# Only allow a small whitelist of root-level tracked files.
#

echo "[guard] Root surface whitelist check..."

# Whitelist patterns (extended globs handled manually)
allowed=(
  ".editorconfig"
  ".env"
  ".env.example"
  ".env.production"
  ".gitignore"
  ".golangci.yml"
  ".spectral.yml"
  ".vscode"          # directory can be tracked if explicitly intended
  "AGENTS.md"
  "CHANGELOG.md"
  "CLAUDE.md"
  "DOCKER_PERMISSION_FIX.md"   # moved but keep for historical PRs; tolerate if present
  "Makefile"
  "README.md"
  "atlas.hcl"
  "docker-compose.dev.yml"
  "docker-compose.e2e.yml"
  "docker-compose.test.yml"
  "docker-compose.yml"
  "go.mod"
  "go.sum"
  "go.work"          # optional
  "go.work.sum"      # optional
  "package.json"
  "package-lock.json"
  "vitest.config.ts"
)

# List tracked files at root (no slash in path)
mapfile -d '' roots < <(git ls-files -z | awk -v RS='\0' -F/ 'NF==1{print $0"\0"}')

fail=0
for f in "${roots[@]}"; do
  # skip empty
  [[ -n "$f" ]] || continue
  ok=0
  for a in "${allowed[@]}"; do
    if [[ "$f" == "$a" ]]; then ok=1; break; fi
  done
  if [[ "$ok" -eq 0 ]]; then
    echo "❌ root file not allowed: $f"
    fail=1
  fi
done

if [[ "$fail" -ne 0 ]]; then
  echo "💥 Guard failed. Please move files into appropriate folders (docs/, reports/, scripts/, logs/, etc.)."
  exit 1
fi

echo "✅ Root surface clean (tracked files)"
