#!/usr/bin/env bash
set -euo pipefail

die() { echo "❌ $*" >&2; exit 1; }
req() { test -n "${!1:-}" || die "Missing env: $1"; }

req RUNNER_TOKEN
req RUNNER_REPO

RUNNER_NAME="${RUNNER_NAME:-cc-runner-$(hostname)}"
RUNNER_LABELS="${RUNNER_LABELS:-self-hosted,cubecastle,linux,x64,docker}"
RUNNER_WORKDIR="${RUNNER_WORKDIR:-/home/runner/_work}"
EPHEMERAL="${EPHEMERAL:-true}"
DISABLE_AUTO_UPDATE="${DISABLE_AUTO_UPDATE:-true}"

echo "🏁 Configuring runner for repo ${RUNNER_REPO} ..."
./config.sh \
  --url "https://github.com/${RUNNER_REPO}" \
  --token "${RUNNER_TOKEN}" \
  --name "${RUNNER_NAME}" \
  --labels "${RUNNER_LABELS}" \
  --work "${RUNNER_WORKDIR}" \
  --ephemeral "${EPHEMERAL}" \
  --unattended \
  --replace

cleanup() {
  echo "🧹 Removing runner..."
  ./config.sh remove --unattended --token "${RUNNER_TOKEN}" || true
}
trap cleanup EXIT

echo "▶ Running..."
exec ./run.sh

