#!/bin/bash

# 优化后的组织架构服务启动脚本
# 6个服务合并为2个核心服务

set -e

echo "🚀 启动优化后的组织架构服务 (2个核心服务)"
echo "==============================================="

# 设置环境变量
export NEO4J_URI="bolt://localhost:7687"
export NEO4J_USER="neo4j"
export NEO4J_PASSWORD="password"
export REDIS_ADDR="localhost:6379"
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"

# 检查必要的服务
echo "📋 检查基础服务状态..."

# 检查PostgreSQL
if ! pg_isready -h localhost -p 5432 -U user -d cubecastle > /dev/null 2>&1; then
    echo "❌ PostgreSQL未启动，请先启动数据库"
    exit 1
fi
echo "✅ PostgreSQL连接正常"

# 检查Neo4j
if ! curl -s http://localhost:7474/db/data/ > /dev/null 2>&1; then
    echo "❌ Neo4j未启动，请先启动图数据库"
    exit 1
fi
echo "✅ Neo4j连接正常"

# 检查Redis (可选)
if ! redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then
    echo "⚠️  Redis未启动，缓存功能将被禁用"
else
    echo "✅ Redis连接正常"
fi

echo ""
echo "🛠️  编译和启动服务..."

# 停止旧服务
echo "🛑 停止旧的6个服务实例..."
pkill -f "organization-command-server" 2>/dev/null || true
pkill -f "organization-graphql-service" 2>/dev/null || true
pkill -f "organization-api-gateway" 2>/dev/null || true
pkill -f "organization-api-server" 2>/dev/null || true
pkill -f "organization-query" 2>/dev/null || true

sleep 2

# 创建日志目录
mkdir -p logs

# 启动核心服务1: 简化命令服务 (9090端口)
echo "🚀 启动简化命令服务 (端口9090)..."
cd cmd/organization-command-service-simplified

# 编译
go mod tidy
go build -o organization-command-service-simplified main.go

# 后台启动
nohup ./organization-command-service-simplified > ../../logs/command-service.log 2>&1 &
COMMAND_PID=$!

cd ../..

# 等待命令服务启动
sleep 3
if kill -0 $COMMAND_PID 2>/dev/null; then
    echo "✅ 简化命令服务启动成功 (PID: $COMMAND_PID)"
else
    echo "❌ 简化命令服务启动失败"
    exit 1
fi

# 启动核心服务2: 统一查询服务 (8090端口)
echo "🚀 启动统一查询服务 (端口8090)..."
cd cmd/organization-query-service

# 编译
go mod tidy
go build -o organization-query-service main.go

# 后台启动
nohup ./organization-query-service > ../../logs/query-service.log 2>&1 &
QUERY_PID=$!

cd ../..

# 等待查询服务启动
sleep 5
if kill -0 $QUERY_PID 2>/dev/null; then
    echo "✅ 统一查询服务启动成功 (PID: $QUERY_PID)"
else
    echo "❌ 统一查询服务启动失败"
    exit 1
fi

echo ""
echo "🎯 服务优化完成!"
echo "================================="
echo "原架构: 6个服务"
echo "新架构: 2个核心服务 (减少67%)"
echo ""
echo "📍 核心服务1: 简化命令服务"
echo "   端口: 9090"
echo "   功能: 所有写操作 + 统一业务验证"
echo "   数据库: PostgreSQL"
echo "   API: http://localhost:9090/api/v1/organization-units"
echo ""
echo "📍 核心服务2: 统一查询服务" 
echo "   端口: 8090"
echo "   功能: GraphQL + REST查询 + 缓存 + 监控"
echo "   数据库: Neo4j + Redis缓存"
echo "   GraphQL: http://localhost:8090/graphql"
echo "   GraphiQL: http://localhost:8090/graphiql"
echo "   REST API: http://localhost:8090/api/v1/organization-units"
echo ""
echo "📊 监控端点:"
echo "   命令服务指标: http://localhost:9090/metrics"
echo "   查询服务指标: http://localhost:8090/metrics"
echo ""
echo "📋 健康检查:"
echo "   命令服务: http://localhost:9090/health"
echo "   查询服务: http://localhost:8090/health"
echo ""
echo "📝 日志文件:"
echo "   命令服务: logs/command-service.log"
echo "   查询服务: logs/query-service.log"
echo ""

# 保存PID
echo $COMMAND_PID > logs/command-service.pid
echo $QUERY_PID > logs/query-service.pid

echo "✅ 所有服务启动完成!"
echo ""
echo "🔧 快速测试命令:"
echo "curl http://localhost:9090/health  # 命令服务健康检查"
echo "curl http://localhost:8090/health  # 查询服务健康检查" 
echo ""
echo "🛑 停止服务: ./stop_optimized_services.sh"