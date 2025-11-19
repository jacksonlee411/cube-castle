#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 <service> [timeout_in_seconds]" >&2
  exit 1
fi

SERVICE="$1"
TIMEOUT="${2:-180}"
INTERVAL=5
COMPOSE_FILE="${COMPOSE_FILE_OVERRIDE:-docker-compose.dev.yml}"

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "[check-health] compose file not found: $COMPOSE_FILE" >&2
  exit 1
fi

end_time=$((SECONDS + TIMEOUT))

while (( SECONDS < end_time )); do
  CONTAINER_ID=$(docker compose -f "$COMPOSE_FILE" ps -q "$SERVICE" 2>/dev/null || true)
  if [[ -z "$CONTAINER_ID" ]]; then
    sleep "$INTERVAL"
    continue
  fi

  HEALTH=$(docker inspect --format '{{.State.Health.Status}}' "$CONTAINER_ID" 2>/dev/null || true)
  STATUS=$(docker inspect --format '{{.State.Status}}' "$CONTAINER_ID" 2>/dev/null || true)

  if [[ "$HEALTH" == "healthy" || ( -z "$HEALTH" && "$STATUS" == "running" ) ]]; then
    echo "[check-health] $SERVICE healthy (container=$CONTAINER_ID, health=${HEALTH:-n/a}, status=${STATUS:-unknown})"
    exit 0
  fi

  if [[ "$STATUS" == "exited" || "$STATUS" == "dead" ]]; then
    echo "[check-health] $SERVICE container exited (status=$STATUS)." >&2
    docker logs "$CONTAINER_ID" >&2 || true
    exit 1
  fi

  echo "[check-health] waiting for $SERVICE (health=${HEALTH:-none}, status=${STATUS:-unknown})"
  sleep "$INTERVAL"
done

echo "[check-health] timeout waiting for $SERVICE after ${TIMEOUT}s" >&2
exit 1
