#!/usr/bin/env bash
# Plan 272 - 运行产物归档脚本
# 将 logs/、reports/、test-results/ 归档为 tar.zst，并生成 manifest + 运行日志

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TIMESTAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
PERIOD="${PLAN272_ARCHIVE_PERIOD:-$(date -u +"%Y-%m")}"
ARCHIVE_DIR="$ROOT_DIR/archive/runtime-artifacts/$PERIOD"
LOG_DIR="$ROOT_DIR/logs/plan272/archive"
REPORT_DIR="$ROOT_DIR/reports/plan272"
INCLUDE_DIRS=("logs" "reports" "test-results")

mkdir -p "$ARCHIVE_DIR" "$LOG_DIR" "$REPORT_DIR"

ARCHIVE_BASENAME="run-artifacts-$TIMESTAMP"
if command -v zstd >/dev/null 2>&1; then
  ARCHIVE_EXT="tar.zst"
  TAR_ARGS=(--use-compress-program="zstd -T0 -19")
  COMPRESSOR="zstd"
else
  ARCHIVE_EXT="tar.gz"
  TAR_ARGS=(-z)
  COMPRESSOR="gzip"
fi

ARCHIVE_PATH="$ARCHIVE_DIR/$ARCHIVE_BASENAME.$ARCHIVE_EXT"
MANIFEST_PATH="$ARCHIVE_DIR/$ARCHIVE_BASENAME.manifest.json"
LOG_PATH="$LOG_DIR/archive-run-artifacts-$TIMESTAMP.log"

pushd "$ROOT_DIR" >/dev/null

# Collect directories that still exist
INCLUDE_EXISTING=()
for dir in "${INCLUDE_DIRS[@]}"; do
  if [ -d "$dir" ]; then
    INCLUDE_EXISTING+=("$dir")
  fi
done

if [ ${#INCLUDE_EXISTING[@]} -eq 0 ]; then
  echo "⚪ 未找到可归档目录（logs/reports/test-results 均不存在）" | tee "$LOG_PATH"
  exit 0
fi

echo "🏁 Plan 272 运行产物归档启动 @ $TIMESTAMP" | tee "$LOG_PATH"
echo "📦 归档目录: ${INCLUDE_EXISTING[*]}" | tee -a "$LOG_PATH"
echo "🎯 输出: $ARCHIVE_PATH (compressor: $COMPRESSOR)" | tee -a "$LOG_PATH"

# Create archive
tar "${TAR_ARGS[@]}" -cf "$ARCHIVE_PATH" "${INCLUDE_EXISTING[@]}"
SHA256_SUM="$(sha256sum "$ARCHIVE_PATH" | awk '{print $1}')"
echo "🔐 archive sha256: $SHA256_SUM" | tee -a "$LOG_PATH"

# Generate manifest (relative paths + metadata)
python3 - "$ROOT_DIR" "$MANIFEST_PATH" "$PERIOD" "$TIMESTAMP" "$SHA256_SUM" "$ARCHIVE_PATH" "${INCLUDE_EXISTING[@]}" <<'PY'
import json
import sys
import time
from pathlib import Path

root = Path(sys.argv[1])
manifest_path = Path(sys.argv[2])
period = sys.argv[3]
timestamp = sys.argv[4]
sha256 = sys.argv[5]
archive_path = Path(sys.argv[6])
include_dirs = sys.argv[7:]

artifacts = []
for rel_dir in include_dirs:
    base = root / rel_dir
    if not base.exists():
        continue
    for file in base.rglob('*'):
        if file.is_file():
            stat = file.stat()
            artifacts.append({
                "relativePath": str(file.relative_to(root)).replace('\\', '/'),
                "sizeBytes": stat.st_size,
                "mtime": time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime(stat.st_mtime))
            })

archive_ref = str(manifest_path.parent / (manifest_path.stem + '.%s' % "$ARCHIVE_EXT")).replace('\\', '/')

archive_ref = str(archive_path.relative_to(root)).replace('\\', '/')

manifest = {
    "planId": 272,
    "period": period,
    "generatedAt": timestamp,
    "archive": archive_ref,
    "sha256": sha256,
    "artifacts": artifacts
}

manifest_path.write_text(json.dumps(manifest, indent=2), encoding='utf-8')
PY

echo "🗃️ manifest: $MANIFEST_PATH" | tee -a "$LOG_PATH"
echo "✅ Plan 272 运行产物归档完成" | tee -a "$LOG_PATH"

popd >/dev/null
