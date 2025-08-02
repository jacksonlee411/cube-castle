#!/bin/bash

# PostgreSQL Organization API Deployment Script
# 确保后端API路由正确配置并连接到PostgreSQL数据库

set -e

echo "🚀 开始部署PostgreSQL组织管理API..."

# 1. 检查Go后端目录
if [ ! -d "/home/shangmeilin/cube-castle/go-app" ]; then
    echo "❌ Go后端目录不存在"
    exit 1
fi

cd /home/shangmeilin/cube-castle/go-app

# 2. 检查必要的文件是否存在
echo "📋 检查必要文件..."
files=(
    "internal/handler/organization_adapter.go"
    "internal/handler/organization_unit_handler.go"
    "internal/routes/organization_routes.go"
    "ent/schema/organization_unit.go"
)

for file in "${files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "❌ 缺少文件: $file"
        exit 1
    fi
    echo "✅ 文件存在: $file"
done

# 3. 编译检查
echo "🔧 编译检查..."
if ! go build -o /tmp/cube-castle-test ./cmd/server; then
    echo "❌ Go编译失败"
    exit 1
fi
echo "✅ Go编译成功"

# 4. 检查数据库连接配置
echo "🗄️ 检查数据库配置..."
if [ ! -f ".env" ] && [ ! -f "config.yaml" ]; then
    echo "⚠️ 警告: 未找到数据库配置文件"
    echo "请确保以下环境变量已设置:"
    echo "  - DATABASE_URL 或"
    echo "  - DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME"
fi

# 5. 检查前端类型定义
echo "🌐 检查前端类型定义..."
cd /home/shangmeilin/cube-castle/nextjs-app

if ! npm run type-check; then
    echo "❌ 前端TypeScript类型检查失败"
    exit 1
fi
echo "✅ 前端类型检查通过"

# 6. 生成API文档摘要
echo "📚 生成API路由摘要..."
cat > /tmp/organization-api-routes.md << 'EOF'
# Organization API Routes Summary

## Backend Routes (PostgreSQL)
- `GET /api/v1/corehr/organizations` - 获取组织列表
- `POST /api/v1/corehr/organizations` - 创建组织
- `GET /api/v1/corehr/organizations/stats` - 获取组织统计
- `GET /api/v1/corehr/organizations/{id}` - 获取组织详情
- `PUT /api/v1/corehr/organizations/{id}` - 更新组织
- `DELETE /api/v1/corehr/organizations/{id}` - 删除组织

## Legacy Routes (兼容性)
- `GET /api/v1/organization-units` - 原始后端API
- `POST /api/v1/organization-units` - 原始后端API
- `GET /api/v1/organization-units/{id}` - 原始后端API
- `PUT /api/v1/organization-units/{id}` - 原始后端API
- `DELETE /api/v1/organization-units/{id}` - 原始后端API

## Data Model Alignment
Frontend现在使用后端模型:
- `unit_type`: DEPARTMENT, COST_CENTER, COMPANY, PROJECT_TEAM
- `status`: ACTIVE, INACTIVE, PLANNED
- `parent_unit_id`: UUID字符串
- `profile`: JSON对象，存储额外配置

## Key Changes
1. Frontend类型定义与后端OrganizationUnit模型完全对齐
2. API适配器提供路由桥接和数据转换
3. 保持向后兼容性，同时支持新的后端枚举值
4. 完整的CRUD操作支持PostgreSQL持久化
EOF

echo "✅ API路由摘要已生成: /tmp/organization-api-routes.md"

echo ""
echo "🎉 PostgreSQL组织管理API部署检查完成!"
echo ""
echo "📋 下一步操作:"
echo "1. 启动Go后端服务: cd /home/shangmeilin/cube-castle/go-app && go run cmd/server/main.go"
echo "2. 启动前端开发服务器: cd /home/shangmeilin/cube-castle/nextjs-app && npm run dev"
echo "3. 访问 http://localhost:3000/organization/chart 测试组织管理功能"
echo ""
echo "🔧 如果遇到问题，请检查:"
echo "- PostgreSQL数据库是否运行 (通常在端口5432)"
echo "- 数据库连接配置是否正确"
echo "- 所有必要的数据库迁移是否已运行"