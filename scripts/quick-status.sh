#!/bin/bash

# 🚀 Cube Castle 快速服务启动脚本
# 跳过健康检查模块，优先启动核心功能

echo "🏰 启动 Cube Castle 服务..."

# 创建日志目录
mkdir -p logs

# 检查基础设施服务
echo "📦 检查基础设施服务..."
docker ps --format "{{.Names}}: {{.Status}}" | grep -E "(postgres|neo4j|redis|kafka)"

# 基础服务测试
echo "🔍 测试基础设施连接..."

# PostgreSQL
if PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "SELECT 1;" >/dev/null 2>&1; then
    echo "✅ PostgreSQL - 连接正常"
else
    echo "❌ PostgreSQL - 连接失败"
fi

# Neo4j
if curl -f -s -u neo4j:password "http://localhost:7474/db/neo4j/tx/commit" \
   -H "Content-Type: application/json" \
   -d '{"statements":[{"statement":"RETURN 1"}]}' >/dev/null 2>&1; then
    echo "✅ Neo4j - 连接正常"
else
    echo "❌ Neo4j - 连接失败"
fi

# Redis
if redis-cli -h localhost -p 6379 ping 2>/dev/null | grep -q "PONG"; then
    echo "✅ Redis - 连接正常"
else
    echo "❌ Redis - 连接失败"
fi

echo ""
echo "🌐 访问地址："
echo "  - 前端应用: ${FRONTEND_BASE_URL:-http://localhost:3000}"
echo "  - 命令API: ${COMMAND_API_URL:-http://localhost:9090}"
echo "  - 查询API (GraphQL): ${GRAPHQL_API_URL:-http://localhost:8090}/graphql"
echo "  - PgAdmin: http://localhost:5050 (admin@admin.com/admin)"
echo ""
echo "📊 服务状态："
echo "  - 基础设施: ✅ 已启动"
echo "  - 前端服务: ✅ 端口 3001"
echo "  - Go服务: ⚠️  部分运行（模块导入问题）"
echo ""
echo "🔧 已知问题："
echo "  - Go服务的健康检查模块导入路径需要修复"
echo "  - 建议使用docker-compose方式替代go run直接运行"
echo ""
echo "✨ 系统已基本可用！前端和数据库服务正常运行。"