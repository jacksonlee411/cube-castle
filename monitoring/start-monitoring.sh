#!/bin/bash

# Cube Castle 监控服务启动脚本

set -e

echo "🔍 启动 Cube Castle 监控服务..."

# 检查 Docker 网络
if ! docker network ls | grep -q cube_castle_network; then
    echo "📡 创建 Docker 网络..."
    docker network create cube_castle_network
fi

# 启动监控服务
echo "📊 启动 Prometheus 和 Grafana..."
cd /home/shangmeilin/cube-castle/monitoring
docker-compose -f docker-compose.monitoring.yml up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo "✅ 检查监控服务状态..."
docker-compose -f docker-compose.monitoring.yml ps

echo "🎉 监控服务启动完成!"
echo "📊 Prometheus: http://localhost:9091"
echo "📈 Grafana: http://localhost:3333 (admin/admin123)"
echo "🗄️ PostgreSQL Exporter: http://localhost:9187/metrics"

# 提示用户如何配置指标收集
echo ""
echo "⚠️  下一步配置指南:"
echo "1. 在后端服务中添加 /metrics 端点"
echo "2. 在前端应用中集成性能监控"
echo "3. 配置应用级别的业务指标"