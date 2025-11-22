#!/usr/bin/env bash
set -euo pipefail

# Plan 263 helper:
# Toggle the "performance-impact-analysis" required check inside the "feat" repository ruleset.
#
# Usage:
#   scripts/ci/workflows/toggle-performance-gate.sh --mode enable --reason "stabilized"
#   scripts/ci/workflows/toggle-performance-gate.sh --mode disable --reason "build failures > 2"
#
# Optional flags:
#   --branch <ref>     : Branch name enforced by the ruleset (default: feat/shared-dev)
#   --context <string> : Status check context (default: performance-impact-analysis)
#   --dry-run          : Show planned mutation without calling GitHub API.

MODE=""
BRANCH="feat/shared-dev"
CONTEXT="performance-impact-analysis"
REASON="unspecified"
DRY_RUN=0

usage() {
  cat <<'EOF'
Usage: toggle-performance-gate.sh --mode <enable|disable> [options]
  --mode <enable|disable>   Required. Enable or disable the performance gate.
  --branch <name>           Branch name controlled by the ruleset (default: feat/shared-dev).
  --context <name>          Status check context (default: performance-impact-analysis).
  --reason <text>           Description for audit log (default: unspecified).
  --dry-run                 Print planned payload without applying it.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode)
      MODE="${2:-}"
      shift 2
      ;;
    --branch)
      BRANCH="${2:-}"
      shift 2
      ;;
    --context)
      CONTEXT="${2:-}"
      shift 2
      ;;
    --reason)
      REASON="${2:-}"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "$MODE" ]]; then
  echo "❌ --mode is required." >&2
  usage
  exit 1
fi

if [[ "$MODE" != "enable" && "$MODE" != "disable" ]]; then
  echo "❌ --mode must be either enable or disable." >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "❌ GitHub CLI (gh) is required." >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "❌ jq is required." >&2
  exit 1
fi

REPO="$(gh repo view --json nameWithOwner -q .nameWithOwner)"
LOG_DIR="logs/plan263"
mkdir -p "$LOG_DIR"
TS="$(date +%Y%m%dT%H%M%S)"
LOG_FILE="${LOG_DIR}/plan263-gate-toggle-${TS}.log"

log() {
  local msg="[${TS}] $*"
  echo "$msg" | tee -a "$LOG_FILE"
}

log "Plan263 gate toggle start: mode=${MODE}, branch=${BRANCH}, context=${CONTEXT}, reason='${REASON}'"

RULESETS_JSON="$(gh api "repos/${REPO}/rulesets")"
RULESET_ID="$(echo "$RULESETS_JSON" | jq -r ".[] | select(.name==\"feat\" and .target==\"branch\") | .id")"

if [[ -z "$RULESET_ID" || "$RULESET_ID" == "null" ]]; then
  log "❌ Could not find ruleset named 'feat' targeting branch."
  exit 1
fi

RULESET_DETAIL="$(gh api "repos/${REPO}/rulesets/${RULESET_ID}")"

INTEGRATION_ID="$(echo "$RULESET_DETAIL" | jq -r '.rules[] | select(.type=="required_status_checks") | .parameters.required_status_checks[0].integration_id // empty')"
if [[ -z "$INTEGRATION_ID" ]]; then
  INTEGRATION_ID=15368
  log "ℹ️ Unable to detect integration_id from existing checks. Defaulting to ${INTEGRATION_ID}."
fi

TMP_PAYLOAD="$(mktemp)"
trap 'rm -f "$TMP_PAYLOAD"' EXIT

if [[ "$MODE" == "enable" ]]; then
  JQ_SCRIPT=$(cat <<'JQ'
    .rules = (.rules | map(
      if .type == "required_status_checks" then
        .parameters.required_status_checks =
          (if any(.parameters.required_status_checks[]?; .context == $context) then
            .parameters.required_status_checks
          else
            (.parameters.required_status_checks + [{context: $context, integration_id: $integration_id}])
          end)
      else . end
    ))
JQ
)
else
  JQ_SCRIPT=$(cat <<'JQ'
    .rules = (.rules | map(
      if .type == "required_status_checks" then
        .parameters.required_status_checks =
          (.parameters.required_status_checks | map(select(.context != $context)))
      else . end
    ))
JQ
)
fi

echo "$RULESET_DETAIL" | jq \
  --arg context "$CONTEXT" \
  --argjson integration_id "$INTEGRATION_ID" \
  "$JQ_SCRIPT | {name, target, enforcement, conditions, rules, bypass_actors}" > "$TMP_PAYLOAD"

log "Generated payload stored at ${TMP_PAYLOAD}"

if [[ "$DRY_RUN" -eq 1 ]]; then
  log "Dry-run mode enabled. Payload:"
  cat "$TMP_PAYLOAD" | tee -a "$LOG_FILE"
  log "Skipping GitHub API update."
  exit 0
fi

gh api --method PUT "repos/${REPO}/rulesets/${RULESET_ID}" \
  --input "$TMP_PAYLOAD" >/dev/null

log "✅ Github ruleset updated successfully."
log "Reason: ${REASON}"
log "Plan263 gate toggle completed."
