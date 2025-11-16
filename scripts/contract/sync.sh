#!/bin/bash
# 契约同步主脚本
# 用途：从 OpenAPI/GraphQL 契约生成统一中间层与 Go/TS 类型
# 维护：架构组（计划 60 / 61 单人执行）

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

echo "📋 [契约同步] 开始..."
echo "  工作目录: $PROJECT_ROOT"

# 1. 从 OpenAPI 提取契约
echo "  → 提取 OpenAPI 契约..."
node scripts/contract/openapi-to-json.js

# 2. 从 GraphQL 提取契约
echo "  → 提取 GraphQL 契约..."
node scripts/contract/graphql-to-json.js

# 3. 生成 Go 类型
echo "  → 生成 Go 类型..."
node scripts/contract/generate-go-types.js

# 4. 生成 TypeScript 类型
echo "  → 生成 TypeScript 类型..."
node scripts/contract/generate-ts-types.js

echo "✅ [契约同步] 完成"
echo "  输出文件:"
echo "    - shared/contracts/organization.json"
echo "    - internal/types/contract_gen.go"
echo "    - frontend/src/shared/types/contract_gen.ts"
