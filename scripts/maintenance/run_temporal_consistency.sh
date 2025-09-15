#!/usr/bin/env bash
set -euo pipefail

# Runs temporal consistency inspection and optional fix.
# Usage:
#   export DATABASE_URL_HOST="postgresql://user:password@localhost:5432/cubecastle?sslmode=disable"
#   ./scripts/maintenance/run_temporal_consistency.sh check
#   ./scripts/maintenance/run_temporal_consistency.sh fix
#   ./scripts/maintenance/run_temporal_consistency.sh check-and-fix
#
# The script prefers DATABASE_URL_HOST, then DATABASE_URL, else exits.

SQL_CHECK="sql/inspection/check_temporal_consistency.sql"
SQL_FIX="sql/maintenance/fix_is_temporal_alignment.sql"
SQL_FIX_TIMELINE="sql/maintenance/fix_temporal_timeline_continuity.sql"

DB_URL="${DATABASE_URL_HOST:-${DATABASE_URL:-}}"
if [[ -z "${DB_URL}" ]]; then
  echo "ERROR: DATABASE_URL_HOST or DATABASE_URL must be set." >&2
  exit 1
fi

cmd="${1:-check}"

run_sql() {
  local file="$1"
  echo "Running: ${file}"
  PGPASSWORD="" psql "${DB_URL}" -v ON_ERROR_STOP=1 -f "${file}"
}

case "${cmd}" in
  check)
    run_sql "${SQL_CHECK}"
    ;;
  fix)
    # Run is_temporal alignment only if column still exists (backward-compat)
    if psql "${DB_URL}" -tAc "SELECT 1 FROM information_schema.columns WHERE table_name='organization_units' AND column_name='is_temporal'" | grep -q 1; then
      run_sql "${SQL_FIX}"
    else
      echo "is_temporal column not present; skipping alignment fix"
    fi
    ;;
  fix-timeline)
    run_sql "${SQL_FIX_TIMELINE}"
    ;;
  fix-all)
    run_sql "${SQL_FIX_TIMELINE}"
    if psql "${DB_URL}" -tAc "SELECT 1 FROM information_schema.columns WHERE table_name='organization_units' AND column_name='is_temporal'" | grep -q 1; then
      run_sql "${SQL_FIX}"
    else
      echo "is_temporal column not present; skipping alignment fix"
    fi
    ;;
  check-and-fix)
    run_sql "${SQL_CHECK}"
    run_sql "${SQL_FIX_TIMELINE}"
    if psql "${DB_URL}" -tAc "SELECT 1 FROM information_schema.columns WHERE table_name='organization_units' AND column_name='is_temporal'" | grep -q 1; then
      run_sql "${SQL_FIX}"
    else
      echo "is_temporal column not present; skipping alignment fix"
    fi
    run_sql "${SQL_CHECK}"
    ;;
  *)
    echo "Unknown command: ${cmd}" >&2
    echo "Valid: check | fix | fix-timeline | fix-all | check-and-fix" >&2
    exit 2
    ;;
esac

echo "Done."
