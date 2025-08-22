#!/bin/bash

# PostgreSQL原生架构停止脚本
# 版本: v3.0-PostgreSQL-Native-Revolution

echo "🛑 停止Cube Castle PostgreSQL原生架构"
echo ""

echo "1️⃣ 停止应用服务..."
pkill -f "organization-command-service" 
pkill -f "organization-query-service"
pkill -f "npm run dev"

echo "2️⃣ 停止Docker基础设施..."
docker-compose down

echo "3️⃣ 清理资源..."
docker system prune -f --volumes

echo ""
echo "✅ PostgreSQL原生架构已完全停止"
echo "💾 数据已保留在Docker volumes中"
echo ""
echo "🔄 重新启动命令:"
echo "   ./start-postgresql-native.sh"