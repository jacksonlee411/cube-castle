#!/usr/bin/env bash
set -euo pipefail
#
# Move noisy root-level logs into logs/root-archive-<timestamp>/
# Safe: only moves files if they exist; no git index changes.
#
ts="$(date +%Y%m%d_%H%M%S)"
dest="logs/root-archive-$ts"
mkdir -p "$dest"

moved=0
patterns=(
  "run-dev*.log"
  "run-frontend*.log"
  "run-query*.log"
  "run-auth-*.log"
  "frontend-dev.log"
  "frontend_dev.log"
  "orphaned-processes.log"
  "all-services-started.log"
  "backend-started.log"
  "baseline-ports.log"
  "baseline-processes.log"
)

for p in "${patterns[@]}"; do
  for f in $p; do
    [[ -e "$f" ]] || continue
    echo "Move: $f -> $dest/"
    mv -f "$f" "$dest/"
    moved=1
  done
done

if [[ "$moved" = "0" ]]; then
  echo "Nothing to move."
else
  echo "Done. Logs archived under: $dest"
fi
