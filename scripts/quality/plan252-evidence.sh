#!/usr/bin/env bash
set -euo pipefail

OUT_DIR="${1:-reports/permissions}"
LOG_DIR="logs/plan252"
TS="$(date +%Y%m%d_%H%M%S)"
mkdir -p "$OUT_DIR" "$LOG_DIR"

echo "[plan252] 生成权限契约校验报告..."
node scripts/quality/auth-permission-contract-validator.js \
  --openapi docs/api/openapi.yaml \
  --graphql docs/api/schema.graphql \
  --resolver-dirs internal/organization/resolver,cmd/hrms-server/query/internal/auth \
  --out "$OUT_DIR" \
  --fail-on unregistered-scope,mapping-missing,resolver-bypass || true

SUMMARY="$OUT_DIR/summary.txt"
DEST_SUMMARY="$LOG_DIR/validator-summary-$TS.txt"
if [ -f "$SUMMARY" ]; then
  cp -f "$SUMMARY" "$DEST_SUMMARY"
  echo "[plan252] summary -> $DEST_SUMMARY"
else
  echo "[plan252] 未找到 summary.txt"
fi

echo "[plan252] 复制明细报告到日志目录（快照）..."
SNAP_DIR="$LOG_DIR/reports-$TS"
mkdir -p "$SNAP_DIR"
cp -f "$OUT_DIR/openapi-scope-usage.json" "$SNAP_DIR/" 2>/dev/null || true
cp -f "$OUT_DIR/openapi-scope-registry.json" "$SNAP_DIR/" 2>/dev/null || true
cp -f "$OUT_DIR/graphql-query-permissions.json" "$SNAP_DIR/" 2>/dev/null || true
cp -f "$OUT_DIR/resolver-permission-calls.json" "$SNAP_DIR/" 2>/dev/null || true
echo "[plan252] 完成：$SNAP_DIR"
