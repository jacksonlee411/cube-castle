#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(git rev-parse --show-toplevel)"
export GOCACHE="$(mktemp -d)"

cleanup() {
  rm -rf "${GOCACHE}"
}
trap cleanup EXIT

echo "[scheduler-alert-smoke] running Temporal monitor alert injection test"
cd "${ROOT_DIR}"
go test ./internal/organization/scheduler -run TestTemporalMonitorCheckAlertsCritical -v
