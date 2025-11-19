#!/usr/bin/env bash
set -euo pipefail

HOSTS_FILE=${HOSTS_FILE:-/etc/hosts}
MARKER="# Plan 267-D GitHub override"
TIMESTAMP=$(date -u +%Y%m%dT%H%M%SZ)
BACKUP_FILE="${HOSTS_FILE}.plan267.${TIMESTAMP}.bak"
MARKER_LINE="${MARKER} (${TIMESTAMP})"

declare -A HOST_MAP=(
  ["github.com"]="140.82.113.4"
  ["codeload.github.com"]="140.82.116.10"
  ["api.github.com"]="140.82.116.6"
  ["raw.githubusercontent.com"]="185.199.110.133"
  ["github-releases.githubusercontent.com"]="185.199.109.154"
  ["objects.githubusercontent.com"]="185.199.111.133"
  ["objects-origin.githubusercontent.com"]="140.82.114.21"
  ["release-assets.githubusercontent.com"]="185.199.110.133"
)

HOSTS_ORDER=(
  "github.com"
  "codeload.github.com"
  "api.github.com"
  "raw.githubusercontent.com"
  "github-releases.githubusercontent.com"
  "objects.githubusercontent.com"
  "objects-origin.githubusercontent.com"
  "release-assets.githubusercontent.com"
)

require_root() {
  if [[ "$(id -u)" != "0" ]]; then
    echo "[network] 请以 root 身份运行（sudo bash scripts/network/configure-github-hosts.sh）" >&2
    exit 1
  fi
}

ensure_hosts_file() {
  if [[ ! -f "$HOSTS_FILE" ]]; then
    echo "[network] hosts 文件不存在: $HOSTS_FILE" >&2
    exit 1
  fi
}

backup_hosts() {
  cp "$HOSTS_FILE" "$BACKUP_FILE"
  echo "[network] 已备份 hosts => $BACKUP_FILE"
}

line_targets_override() {
  local line=$1

  if [[ "$line" == "$MARKER"* ]]; then
    return 0
  fi

  local host escaped
  for host in "${HOSTS_ORDER[@]}"; do
    escaped=${host//./\\.}
    if printf '%s\n' "$line" | grep -Eq "(^|[[:space:]])${escaped}([[:space:]]|$|#)"; then
      return 0
    fi
  done

  return 1
}

strip_existing_entries() {
  local tmp line
  tmp=$(mktemp)

  while IFS= read -r line || [[ -n "$line" ]]; do
    if line_targets_override "$line"; then
      continue
    fi
    printf '%s\n' "$line" >>"$tmp"
  done <"$HOSTS_FILE"

  cat "$tmp" >"$HOSTS_FILE"
  rm -f "$tmp"
}

append_override_block() {
  {
    printf '\n%s\n' "$MARKER_LINE"
    local host
    for host in "${HOSTS_ORDER[@]}"; do
      printf "%-18s %s\n" "${HOST_MAP[$host]}" "$host"
    done
  } >>"$HOSTS_FILE"
}

require_root
ensure_hosts_file
backup_hosts
strip_existing_entries
append_override_block

echo "[network] GitHub hosts 已更新（$HOSTS_FILE）"
echo "[network] 如需回滚：sudo cp \"$BACKUP_FILE\" \"$HOSTS_FILE\" && wsl.exe --shutdown"
