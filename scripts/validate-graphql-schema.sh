#!/bin/bash
# GraphQL Schema一致性验证脚本
set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCHEMA_FILE="$PROJECT_ROOT/docs/api/schema.graphql"

echo "🔍 GraphQL Schema单一真源验证"
echo "======================================="

# 检查schema文件是否存在
if [[ ! -f "$SCHEMA_FILE" ]]; then
    echo "❌ Schema文件未找到: $SCHEMA_FILE"
    exit 1
fi
echo "✅ Schema文件存在: $SCHEMA_FILE"

# 检查是否存在硬编码schema（防止回退）
echo "🔍 检查硬编码schema（双源维护检测）..."
HARDCODED_FILES=$(find "$PROJECT_ROOT" -name "*.go" -type f -exec grep -l "var.*schemaString.*=" {} \; || true)

if [[ -n "$HARDCODED_FILES" ]]; then
    echo "❌ 发现硬编码GraphQL Schema，违反单一真源原则！"
    echo "$HARDCODED_FILES"
    exit 1
fi
echo "✅ 无硬编码schema检测通过"

# 计算schema文件hash
SCHEMA_HASH=$(sha256sum "$SCHEMA_FILE" | cut -d' ' -f1)
echo "📊 Schema文件hash: $SCHEMA_HASH"

echo "🎉 GraphQL Schema单一真源验证完成！"
echo "✅ 权威来源：docs/api/schema.graphql"
echo "✅ 无双源维护风险"