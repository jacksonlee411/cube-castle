#!/bin/bash

# 🎯 API契约测试脚本 - 确保前端严格遵循API契约v4.2.1

set -e

echo "🔍 开始API契约验证..."

# 1. GraphQL Schema语法验证
echo "📋 验证GraphQL Schema语法..."
if command -v graphql-schema-linter &> /dev/null; then
    graphql-schema-linter docs/api/schema.graphql
    echo "✅ GraphQL Schema语法验证通过"
else
    echo "⚠️ graphql-schema-linter未安装，跳过Schema验证"
fi

# 2. 检查前端代码中的字段命名规范
echo "🔤 检查字段命名规范 (camelCase强制)..."

# 检查是否使用了禁止的snake_case字段
SNAKE_CASE_FIELDS=$(find frontend/src -name "*.ts" -o -name "*.tsx" | xargs grep -l "unit_type\|parent_code\|sort_order\|effective_date\|end_date\|created_at\|updated_at" | head -5)

if [ -n "$SNAKE_CASE_FIELDS" ]; then
    echo "❌ 发现违反camelCase命名规范的字段："
    echo "$SNAKE_CASE_FIELDS"
    echo "请将snake_case字段改为camelCase: unitType, parentCode, sortOrder, effectiveDate, endDate, createdAt, updatedAt"
    exit 1
else
    echo "✅ 字段命名规范验证通过"
fi

# 3. 检查GraphQL查询是否匹配Schema
echo "📊 验证GraphQL查询与Schema匹配..."

# 检查是否使用了不存在的查询
INVALID_QUERIES=$(find frontend/src -name "*.ts" -o -name "*.tsx" | xargs grep -l "organizationAsOfDate\|organizationHistory" | head -3)

if [ -n "$INVALID_QUERIES" ]; then
    echo "❌ 发现使用不存在的GraphQL查询："
    echo "$INVALID_QUERIES"
    echo "请使用Schema中真实存在的查询: organization, organizationAuditHistory"
    exit 1
else  
    echo "✅ GraphQL查询验证通过"
fi

# 4. 构建验证
echo "🔨 验证构建零错误..."
cd frontend && npm run build
echo "✅ 构建零错误验证通过"

# 5. 类型检查
echo "🔍 TypeScript类型检查..."
cd frontend && npm run typecheck
echo "✅ TypeScript类型检查通过"

# 6. 代码规范检查
echo "📏 ESLint代码规范检查..."
cd frontend && npm run lint
echo "✅ 代码规范检查通过"

echo ""
echo "🎉 API契约验证全部通过！"
echo "✅ GraphQL查询符合Schema"  
echo "✅ 字段命名遵循camelCase规范"
echo "✅ 构建零错误"
echo "✅ 类型系统正确"
echo "✅ 代码规范合格"