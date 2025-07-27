#!/bin/bash

# Temporal测试环境启动脚本
set -e

echo "🚀 启动Cube Castle Temporal测试环境"

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 创建temporal配置目录
mkdir -p temporal-config

# 创建Temporal动态配置文件
cat > temporal-config/development-sql.yaml << EOF
system.forceSearchAttributesCacheRefreshOnRead:
  - value: true
    constraints: {}

frontend.enableClientVersionCheck:
  - value: true
    constraints: {}

history.maxAutoResetPoints:
  - value: 20
    constraints: {}

frontend.keepAlivePermitWithoutStream:
  - value: true
    constraints: {}

frontend.enableTokenNamespaceEnforcement:
  - value: false
    constraints: {}
EOF

echo "📦 启动Docker Compose服务..."
docker-compose -f docker-compose.temporal.yml up -d

echo "⏳ 等待服务启动..."
sleep 30

# 检查Temporal健康状态
echo "🔍 检查Temporal服务状态..."
timeout 60 bash -c 'until docker-compose -f docker-compose.temporal.yml exec -T temporal temporal workflow list --namespace default > /dev/null 2>&1; do sleep 2; done'

if [ $? -eq 0 ]; then
    echo "✅ Temporal服务启动成功"
else
    echo "❌ Temporal服务启动失败"
    docker-compose -f docker-compose.temporal.yml logs temporal
    exit 1
fi

# 创建测试命名空间
echo "🏗️ 创建测试命名空间..."
docker-compose -f docker-compose.temporal.yml exec -T temporal temporal operator namespace create test-namespace || true

# 检查其他服务
echo "🔍 检查服务状态..."
docker-compose -f docker-compose.temporal.yml ps

echo "🎉 Temporal测试环境启动完成！"
echo ""
echo "📊 服务地址:"
echo "  - Temporal gRPC: localhost:7233"
echo "  - Temporal Web UI: http://localhost:8080"
echo "  - PostgreSQL (App): localhost:5432"
echo "  - PostgreSQL (Temporal): localhost:5433" 
echo "  - Redis: localhost:6379"
echo ""
echo "🧪 运行测试:"
echo "  go test -v ./internal/workflow/ -tags integration"
echo ""
echo "🛑 停止环境:"
echo "  ./stop-temporal-test.sh"