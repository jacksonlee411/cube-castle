#!/bin/bash

# Temporal测试环境停止脚本
set -e

echo "🛑 停止Cube Castle Temporal测试环境"

# 停止所有服务
docker-compose -f docker-compose.temporal.yml down

echo "🧹 清理容器和网络..."
docker-compose -f docker-compose.temporal.yml down --volumes --remove-orphans

echo "✅ Temporal测试环境已停止"

# 可选：清理所有数据
read -p "是否清理所有测试数据？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🗑️ 清理测试数据..."
    docker-compose -f docker-compose.temporal.yml down --volumes
    docker volume prune -f
    echo "✅ 测试数据已清理"
fi