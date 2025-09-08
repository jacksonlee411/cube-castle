#!/bin/bash

# 生产环境部署脚本
# Cube Castle时态管理API - 生产环境快速部署

set -e

PROJECT_ROOT="/home/shangmeilin/cube-castle"
PROD_ENV_FILE="$PROJECT_ROOT/.env.production"

echo "🚀 === Cube Castle生产环境部署 ==="
echo ""

# 1. 环境检查
echo "📋 1. 生产环境检查..."

# 检查必要的服务
required_services=("postgres:5432" "redis:6379" "neo4j:7474" "kafka:9092")
for service in "${required_services[@]}"; do
    IFS=':' read -r name port <<< "$service"
    if docker ps | grep -q "$name"; then
        echo "✅ $name 容器运行正常"
    else
        echo "❌ $name 容器异常，请检查 docker-compose"
        exit 1
    fi
done

# 2. 创建生产环境配置
echo ""
echo "⚙️  2. 生产环境配置生成..."

cat > "$PROD_ENV_FILE" << EOF
# Cube Castle 生产环境配置
# 生成时间: $(date)

# === 服务端口配置 ===
COMMAND_SERVICE_PORT=9090
QUERY_SERVICE_PORT=8090
FRONTEND_PORT=3000

# === 数据库配置 ===
DATABASE_URL=postgres://user:password@localhost:5432/cubecastle
REDIS_URL=redis://localhost:6379
NEO4J_URI=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=password

# === 时态管理配置 ===
TEMPORAL_MANAGEMENT_ENABLED=true
AUTO_END_DATE_MANAGEMENT=true
TIMELINE_CONSISTENCY_POLICY=NO_GAPS_ALLOWED
DEFAULT_QUERY_MODE=CURRENT_ONLY
SUPPORTS_RETROACTIVITY=true
MAX_RETROACTIVE_DAYS=365

# === 缓存配置 ===
REDIS_CACHE_ENABLED=true
CACHE_DEFAULT_TTL=3600
CACHE_HIT_RATE_TARGET=90

# === 监控配置 ===
PROMETHEUS_ENABLED=true
METRICS_PORT=9999
HEALTH_CHECK_INTERVAL=30s

# === 日志配置 ===
LOG_LEVEL=info
LOG_FORMAT=json

# === 安全配置 ===
CORS_ALLOWED_ORIGINS=${FRONTEND_BASE_URL:-http://localhost:3000}
API_RATE_LIMIT=1000
SESSION_TIMEOUT=3600

# === 性能配置 ===
DATABASE_MAX_CONNECTIONS=20
QUERY_TIMEOUT=30s
BATCH_SIZE=100
EOF

echo "✅ 生产环境配置文件已生成: $PROD_ENV_FILE"

# 3. 生产环境服务启动
echo ""
echo "🔧 3. 启动生产环境服务..."

# 创建生产环境启动脚本
cat > "$PROJECT_ROOT/start-production.sh" << 'EOF'
#!/bin/bash

echo "🚀 启动Cube Castle生产环境服务..."

PROJECT_ROOT="/home/shangmeilin/cube-castle"
cd "$PROJECT_ROOT"

# 加载环境变量
if [ -f .env.production ]; then
    set -a
    source .env.production
    set +a
    echo "✅ 已加载生产环境配置"
fi

# 启动核心服务
echo "🔧 启动核心服务..."

# 1. 启动命令服务
echo "启动命令服务 (端口 9090)..."
cd cmd/organization-command-service
go run main.go > /tmp/command-service.log 2>&1 &
COMMAND_PID=$!
echo "✅ 命令服务已启动 (PID: $COMMAND_PID)"

# 2. 启动查询服务
cd ../organization-query-service-unified
echo "启动查询服务 (端口 8090)..."
go run main.go > /tmp/query-service.log 2>&1 &
QUERY_PID=$!
echo "✅ 查询服务已启动 (PID: $QUERY_PID)"

# 3. 启动同步服务
echo "启动数据同步服务..."
go run main.go > /tmp/sync-service.log 2>&1 &
SYNC_PID=$!
echo "✅ 同步服务已启动 (PID: $SYNC_PID)"

# 4. 启动缓存失效服务
# 缓存失效服务已删除 - 不再启动
echo "ℹ️  缓存失效服务已移除（架构简化）"
# CACHE_PID已不存在

# 保存PID文件
echo "$COMMAND_PID" > /tmp/cube-castle-command.pid
echo "$QUERY_PID" > /tmp/cube-castle-query.pid
echo "$SYNC_PID" > /tmp/cube-castle-sync.pid
# 缓存失效服务PID文件已不需要

# 5. 启动前端 (可选)
cd ../../frontend
echo "启动前端应用 (端口 3000)..."
npm run dev > /tmp/frontend.log 2>&1 &
FRONTEND_PID=$!
echo "$FRONTEND_PID" > /tmp/cube-castle-frontend.pid
echo "✅ 前端应用已启动 (PID: $FRONTEND_PID)"

# 等待服务启动
echo ""
echo "⏳ 等待服务启动..."
sleep 5

# 健康检查
echo ""
echo "🔍 执行健康检查..."
services_healthy=true

# 检查命令服务
if curl -f -s "${COMMAND_API_URL:-http://localhost:9090}/health" > /dev/null; then
    echo "✅ 命令服务健康检查通过"
else
    echo "❌ 命令服务健康检查失败"
    services_healthy=false
fi

# 检查查询服务
if curl -f -s "${GRAPHQL_API_URL:-http://localhost:8090}/health" > /dev/null; then
    echo "✅ 查询服务健康检查通过"
else
    echo "❌ 查询服务健康检查失败"
    services_healthy=false
fi

if [ "$services_healthy" = true ]; then
    echo ""
    echo "🎉 === Cube Castle生产环境启动成功！ ==="
    echo ""
    echo "📊 服务访问地址:"
    echo "   • 命令API: ${COMMAND_API_URL:-http://localhost:9090}"
    echo "   • 查询API (GraphQL): ${GRAPHQL_API_URL:-http://localhost:8090}/graphql"
    echo "   • 前端应用: ${FRONTEND_BASE_URL:-http://localhost:3000}"
    echo ""
    echo "🔧 管理命令:"
    echo "   • 停止服务: ./stop-production.sh"
    echo "   • 查看日志: tail -f /tmp/cube-castle-*.log"
    echo "   • 健康检查: ./health-check.sh"
else
    echo ""
    echo "❌ 部分服务启动失败，请查看日志文件:"
    echo "   • /tmp/command-service.log"
    echo "   • /tmp/query-service.log"
    echo "   • /tmp/sync-service.log"
    echo "   • 缓存失效服务已移除"
    exit 1
fi
EOF

chmod +x "$PROJECT_ROOT/start-production.sh"

# 4. 创建停止脚本
cat > "$PROJECT_ROOT/stop-production.sh" << 'EOF'
#!/bin/bash

echo "🛑 停止Cube Castle生产环境服务..."

# 停止所有服务
if [ -f /tmp/cube-castle-command.pid ]; then
    kill $(cat /tmp/cube-castle-command.pid) 2>/dev/null && echo "✅ 命令服务已停止"
    rm -f /tmp/cube-castle-command.pid
fi

if [ -f /tmp/cube-castle-query.pid ]; then
    kill $(cat /tmp/cube-castle-query.pid) 2>/dev/null && echo "✅ 查询服务已停止"
    rm -f /tmp/cube-castle-query.pid
fi

if [ -f /tmp/cube-castle-sync.pid ]; then
    kill $(cat /tmp/cube-castle-sync.pid) 2>/dev/null && echo "✅ 同步服务已停止"
    rm -f /tmp/cube-castle-sync.pid
fi

if [ -f /tmp/cube-castle-cache.pid ]; then
    kill $(cat /tmp/cube-castle-cache.pid) 2>/dev/null && echo "✅ 缓存失效服务已停止"
    rm -f /tmp/cube-castle-cache.pid
fi

if [ -f /tmp/cube-castle-frontend.pid ]; then
    kill $(cat /tmp/cube-castle-frontend.pid) 2>/dev/null && echo "✅ 前端应用已停止"
    rm -f /tmp/cube-castle-frontend.pid
fi

echo "✅ 所有服务已停止"
EOF

chmod +x "$PROJECT_ROOT/stop-production.sh"

# 5. 创建健康检查脚本
cat > "$PROJECT_ROOT/health-check.sh" << 'EOF'
#!/bin/bash

echo "🔍 === Cube Castle服务健康检查 ==="
echo ""

services_ok=0
total_services=4

# 检查命令服务
if curl -f -s "${COMMAND_API_URL:-http://localhost:9090}/health" > /dev/null; then
    echo "✅ 命令服务 (9090) - 健康"
    services_ok=$((services_ok + 1))
else
    echo "❌ 命令服务 (9090) - 异常"
fi

# 检查查询服务
if curl -f -s "${GRAPHQL_API_URL:-http://localhost:8090}/health" > /dev/null; then
    echo "✅ 查询服务 (8090) - 健康"
    services_ok=$((services_ok + 1))
else
    echo "❌ 查询服务 (8090) - 异常"
fi

# 检查前端服务
if curl -f -s "${FRONTEND_BASE_URL:-http://localhost:3000}" > /dev/null; then
    echo "✅ 前端应用 (3000) - 健康"
    services_ok=$((services_ok + 1))
else
    echo "⚠️  前端应用 (3000) - 异常或未启动"
fi

# 检查数据库连接
if PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ PostgreSQL数据库 - 连接正常"
    services_ok=$((services_ok + 1))
else
    echo "❌ PostgreSQL数据库 - 连接异常"
fi

echo ""
echo "📊 服务健康状态: $services_ok/$total_services"

if [ $services_ok -eq $total_services ]; then
    echo "🎉 所有核心服务运行正常！"
    exit 0
else
    echo "⚠️  部分服务异常，请检查日志"
    exit 1
fi
EOF

chmod +x "$PROJECT_ROOT/health-check.sh"

echo ""
echo "📋 6. 生产环境部署脚本已生成:"
echo "   • 启动: ./start-production.sh"
echo "   • 停止: ./stop-production.sh" 
echo "   • 健康检查: ./health-check.sh"

echo ""
echo "✅ 生产环境部署准备完成！"
echo ""
echo "🚀 执行以下命令开始部署:"
echo "   cd $PROJECT_ROOT"
echo "   ./start-production.sh"