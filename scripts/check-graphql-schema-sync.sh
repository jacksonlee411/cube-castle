#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
SCHEMA_FILE="$ROOT_DIR/docs/api/schema.graphql"
FRONTEND_FILE="$ROOT_DIR/frontend/src/shared/hooks/useEnterprisePositions.ts"

if [[ ! -f "$SCHEMA_FILE" ]]; then
  echo "❌ 未找到 GraphQL 契约文件: $SCHEMA_FILE" >&2
  exit 1
fi

if [[ ! -f "$FRONTEND_FILE" ]]; then
  echo "❌ 未找到前端查询文件: $FRONTEND_FILE" >&2
  exit 1
fi

echo "🔍 校验前端查询与 GraphQL 契约的一致性..."

missing=()

if ! grep -q "type Position" "$SCHEMA_FILE" || ! grep -q "organizationName" "$SCHEMA_FILE"; then
  missing+=('Position.organizationName')
fi

if ! grep -q "type HeadcountStats" "$SCHEMA_FILE" || ! grep -q "byFamily" "$SCHEMA_FILE"; then
  missing+=('HeadcountStats.byFamily')
fi

if ! grep -q "input VacantPositionFilterInput" "$SCHEMA_FILE"; then
  missing+=('VacantPositionFilterInput input type')
fi

if ! grep -q "organizationName" "$FRONTEND_FILE" || ! grep -q "byFamily" "$FRONTEND_FILE"; then
  echo "ℹ️  提示: 前端查询文件中未发现关键字段引用，跳过对齐校验" >&2
fi

if ((${#missing[@]} > 0)); then
  echo "❌ GraphQL 契约缺失以下前端依赖字段:" >&2
  for item in "${missing[@]}"; do
    echo "   - $item" >&2
  done
  echo "👉 请更新 docs/api/schema.graphql 并同步后端实现" >&2
  exit 1
fi

echo "✅ GraphQL 契约与前端查询保持一致"
