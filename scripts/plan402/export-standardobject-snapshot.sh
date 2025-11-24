#!/usr/bin/env bash
# Plan 402D · Standard Object 快照导出脚本
# 生成 organization/position 三表的 CSV，并封装为 tar.{zst|gz}，用于回滚或对账。
# 输出（均存放在 logs/plan402/rollback/）：
#   - <ts>-standardobject-snapshot.tar.{zst|gz}
#   - <ts>-standardobject-snapshot.manifest.json
#   - <ts>-standardobject-snapshot.log （脚本执行日志）

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOG_DIR="$ROOT_DIR/logs/plan402/rollback"
mkdir -p "$LOG_DIR"

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "❌ DATABASE_URL 未设置，无法导出 SOM 快照。" >&2
  exit 1
fi

TIMESTAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
LOG_PATH="$LOG_DIR/${TIMESTAMP}-standardobject-snapshot.log"
ARCHIVE_BASENAME="${TIMESTAMP}-standardobject-snapshot"

exec > >(tee -a "$LOG_PATH") 2>&1

echo "🏁 Plan 402D · Standard Object 快照导出启动 @ $TIMESTAMP"

if command -v zstd >/dev/null 2>&1; then
  ARCHIVE_EXT="tar.zst"
  TAR_ARGS=(--use-compress-program="zstd -T0 -19")
  COMPRESSOR="zstd"
else
  ARCHIVE_EXT="tar.gz"
  TAR_ARGS=(-z)
  COMPRESSOR="gzip"
fi

WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/plan402-snapshot.XXXXXX")"
trap 'rm -rf "$WORK_DIR"' EXIT

KERNEL_PATH="$WORK_DIR/standard_objects.csv"
VERSIONS_PATH="$WORK_DIR/standard_object_versions.csv"
LINKS_PATH="$WORK_DIR/standard_object_links.csv"

echo "📥 导出 standard_objects ..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "\copy (
    SELECT id, object_type, code, display_name, tenant_code, status, labels,
           schema_version, data_classification, retention_policy,
           created_by, created_at, updated_at
    FROM standard_objects
    ORDER BY object_type, code
  ) TO '${KERNEL_PATH}' WITH (FORMAT csv, HEADER true)"

echo "📥 导出 standard_object_versions ..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "\copy (
    SELECT object_id, version_code, effective_date, end_date,
           validity_range, transaction_range, is_current,
           payload, audit, checksum, created_at, updated_at
    FROM standard_object_versions
    ORDER BY object_id, effective_date
  ) TO '${VERSIONS_PATH}' WITH (FORMAT csv, HEADER true)"

echo "📥 导出 standard_object_links ..."
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "\copy (
    SELECT link_type, source_object_id, target_object_id, tenant_code,
           validity_range, transaction_range, attributes, created_by, created_at
    FROM standard_object_links
    ORDER BY link_type, source_object_id
  ) TO '${LINKS_PATH}' WITH (FORMAT csv, HEADER true)"

ARCHIVE_PATH="$LOG_DIR/${ARCHIVE_BASENAME}.${ARCHIVE_EXT}"
MANIFEST_PATH="$LOG_DIR/${ARCHIVE_BASENAME}.manifest.json"
ARCHIVE_RELATIVE="$(python3 - <<'PY' "$ROOT_DIR" "$ARCHIVE_PATH"
import os
import sys
root, path = sys.argv[1], sys.argv[2]
print(os.path.relpath(path, root).replace('\\', '/'))
PY
)"

echo "📦 封装归档 (${COMPRESSOR}) ..."
tar "${TAR_ARGS[@]}" -cf "$ARCHIVE_PATH" -C "$WORK_DIR" .
if command -v sha256sum >/dev/null 2>&1; then
  SHA256_SUM="$(sha256sum "$ARCHIVE_PATH" | awk '{print $1}')"
else
  SHA256_SUM="$(shasum -a 256 "$ARCHIVE_PATH" | awk '{print $1}')"
fi
echo "🔐 archive sha256: $SHA256_SUM"

echo "📝 生成 manifest ..."
kernel_count=$(psql "$DATABASE_URL" -At -c "SELECT COUNT(*) FROM standard_objects")
version_count=$(psql "$DATABASE_URL" -At -c "SELECT COUNT(*) FROM standard_object_versions")
link_count=$(psql "$DATABASE_URL" -At -c "SELECT COUNT(*) FROM standard_object_links")

cat >"$MANIFEST_PATH" <<JSON
{
  "planId": 402,
  "generatedAt": "$TIMESTAMP",
  "archive": "$ARCHIVE_RELATIVE",
  "sha256": "$SHA256_SUM",
  "counts": {
    "standard_objects": $kernel_count,
    "standard_object_versions": $version_count,
    "standard_object_links": $link_count
  },
  "files": [
    "standard_objects.csv",
    "standard_object_versions.csv",
    "standard_object_links.csv"
  ]
}
JSON

echo "✅ 导出完成："
echo "   • Archive : $ARCHIVE_PATH"
echo "   • Manifest: $MANIFEST_PATH"
echo "   • Log     : $LOG_PATH"
echo "📚 请将上述文件引用到 docs/development-plans/402D-cutover-and-recovery.md 的 rollback 证据中。"
