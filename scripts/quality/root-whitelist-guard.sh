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
  ".dockerignore"
  ".env"
  ".env.example"
  ".env.production"
  ".gitignore"
  ".golangci.yml"
  ".spectral.yml"
  ".vscode"          # directory can be tracked if explicitly intended
  ".eslintrc.architecture.js"
  ".eslintrc.interface-freeze.json"
  ".jscpdrc.json"
  ".npmrc"
  ".nvmrc"
  "AGENTS.md"
  "CHANGELOG.md"
  "CLAUDE.md"
  "DOCKER_PERMISSION_FIX.md"   # moved but keep for historical PRs; tolerate if present
  "Makefile"
  "README.md"
  "atlas.hcl"
  "docker-compose.dev.yml"
  "docker-compose.e2e.yml"
  "docker-compose.runner.yml"
  "docker-compose.runner.ghcr.yml"
  "docker-compose.runner.docker.yml"
  "docker-compose.runner.persist.yml"
  "docker-compose.test.yml"
  "docker-compose.yml"
  "go.mod"
  "go.sum"
  "go.work"          # optional
  "go.work.sum"      # optional
  "goose.yaml"
  "package.json"
  "package-lock.json"
  "eslint.config.architecture.mjs"
  "eslint.config.js"
  "vitest.config.ts"
)

# List all tracked files, we will select those at depth=1 (no slash)
mapfile -d '' roots < <(git ls-files -z)

fail=0
for f in "${roots[@]}"; do
  # consider only root-level files
  [[ -n "$f" ]] || continue
  [[ "$f" == */* ]] && continue
  ok=0
  for a in "${allowed[@]}"; do
    if [[ "$f" == "$a" ]]; then ok=1; break; fi
  done
  if [[ "$ok" -eq 0 ]]; then
    printf '❌ root file not allowed: %s\n' "$f"
    fail=1
  fi
done

if [[ "$fail" -ne 0 ]]; then
  echo "💥 Guard failed. Please move files into appropriate folders (docs/, reports/, scripts/, logs/, etc.)."
  exit 1
fi

echo "✅ Root surface clean (tracked files)"
