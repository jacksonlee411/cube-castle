#!/bin/bash
# 基础架构启动脚本 - PostgreSQL + Redis

set -e

echo "🚀 启动Cube Castle基础架构..."

# 检查Docker
if ! command -v docker >/dev/null 2>&1; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

# 启动基础服务
echo "🔧 启动基础架构服务..."
docker-compose -f docker-compose.dev.yml up -d postgres redis

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 15

# 检查服务状态
echo "🔧 检查服务状态..."
docker-compose -f docker-compose.dev.yml ps

echo "✅ 基础架构启动完成!"
echo "📊 PostgreSQL: localhost:5432"
echo "📊 Redis: localhost:6379"

exit 0
