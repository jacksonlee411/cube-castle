#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd -- "$SCRIPT_DIR/../../.." && pwd)
LOG_DIR="$REPO_ROOT/logs/ci-monitor"
COMPOSE_FILE_DEFAULT="$REPO_ROOT/docker-compose.dev.yml"

usage() {
  echo "Usage: $0 <workflow-id> [--teardown]" >&2
  exit 1
}

if [[ $# -lt 1 ]]; then
  usage
fi

WORKFLOW_ID="$1"
MODE="prepare"
if [[ $# -ge 2 ]]; then
  case "$2" in
    --teardown) MODE="teardown" ;;
    *) usage ;;
  esac
fi

LOG_FILE="$LOG_DIR/${WORKFLOW_ID}-prepare.log"
mkdir -p "$LOG_DIR"
mkdir -p "$REPO_ROOT/.cache"

log() {
  echo "[$(date -u +'%Y-%m-%dT%H:%M:%SZ')] $*"
}

# Tee all output to log file for artifact/debug
exec > >(tee -a "$LOG_FILE") 2>&1

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    log "missing required command: $1"
    exit 1
  fi
}

check_versions() {
  log "docker version: $(docker --version)"
  log "docker compose version: $(docker compose version)"
  log "go version: $(go version)"
  log "node version: $(node --version)"
}

compose_file="${COMPOSE_FILE_OVERRIDE:-$COMPOSE_FILE_DEFAULT}"
if [[ ! -f "$compose_file" ]]; then
  log "compose file not found: $compose_file"
  exit 1
fi

require_env() {
  require_cmd docker
  require_cmd go
  require_cmd node
  require_cmd make
}

install_goose_if_missing() {
  if command -v goose >/dev/null 2>&1; then
    log "goose already installed: $(goose -version || true)"
    return
  fi
  log "goose not found, installing v3.26.0"
  curl -sSL https://github.com/pressly/goose/releases/download/v3.26.0/goose_linux_x86_64.tar.gz \
    | sudo tar -xz -f - -C /usr/local/bin goose
  log "goose version: $(goose -version)"
}

bring_up_services() {
  log "Compose up services (postgres, redis) with $compose_file"
  docker compose -f "$compose_file" config -q
  docker compose -f "$compose_file" up -d postgres redis
  COMPOSE_FILE_OVERRIDE="$compose_file" bash "$REPO_ROOT/scripts/ci/docker/check-health.sh" postgres 180
  COMPOSE_FILE_OVERRIDE="$compose_file" bash "$REPO_ROOT/scripts/ci/docker/check-health.sh" redis 180
}

tear_down_services() {
  log "Teardown compose services via $compose_file"
  docker compose -f "$compose_file" down --remove-orphans || true
  # Best-effort prune labeled volumes if present
  docker volume prune --filter label=cubecastle-ci --force || true
}

run_prepare() {
  require_env
  install_goose_if_missing
  check_versions
  bring_up_services
  if [[ "${CI_PREPARE_RUN_MIGRATIONS:-0}" == "1" ]]; then
    log "CI_PREPARE_RUN_MIGRATIONS=1, running make db-migrate-all"
    (cd "$REPO_ROOT" && make db-migrate-all)
  fi
  log "prepare completed for $WORKFLOW_ID"
}

run_teardown() {
  log "teardown requested for $WORKFLOW_ID"
  tear_down_services
  log "teardown completed for $WORKFLOW_ID"
}

case "$MODE" in
  prepare) run_prepare ;;
  teardown) run_teardown ;;
  *) usage ;;
esac
