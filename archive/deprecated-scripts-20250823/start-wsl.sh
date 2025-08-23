#!/bin/bash

# Cube Castle WSL 启动脚本
# 用于在 WSL 环境中快速启动整个项目

set -e

echo "🏰 Cube Castle - WSL 启动脚本"
echo "=============================="

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker 未运行，请先启动 Docker"
    echo "   在 Windows 中启动 Docker Desktop"
    exit 1
fi

# 检查环境变量文件
if [ ! -f ".env" ]; then
    echo "📝 创建环境变量文件..."
    cp env.example .env
    echo "⚠️  请编辑 .env 文件配置您的环境变量"
    echo "   特别是数据库连接和 AI 服务配置"
    read -p "按回车键继续..."
fi

# 启动基础设施
echo "🚀 启动基础设施服务..."
docker-compose up -d postgres neo4j

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 15

# 检查服务状态
echo "📊 检查服务状态..."
if ! docker-compose ps | grep -q "Up"; then
    echo "❌ 服务启动失败，请检查日志："
    docker-compose logs
    exit 1
fi

# 初始化数据库
echo "🗄️ 初始化数据库..."
cd go-app
go run cmd/server/main.go init-db

# 插入种子数据
echo "🌱 插入种子数据..."
go run cmd/server/main.go seed-data
cd ..

# 启动 Python AI 服务
echo "🧙 启动 Python AI 服务..."
cd python-ai
if [ ! -d "venv" ]; then
    echo "📦 创建 Python 虚拟环境..."
    python3 -m venv venv
fi

source venv/bin/activate
pip install -r requirements.txt

echo "🚀 启动 AI 服务 (后台运行)..."
python main.py &
AI_PID=$!
cd ..

# 启动 Go 主服务
echo "🏰 启动 Go 主服务..."
cd go-app
go run cmd/server/main.go &
GO_PID=$!
cd ..

echo ""
echo "✅ Cube Castle 启动完成！"
echo "=========================="
echo "🔗 服务地址："
echo "  - Go 主服务: http://localhost:8080"
echo "  - Python AI 服务: localhost:50051 (gRPC)"
echo "  - PostgreSQL: localhost:5432"
echo "  - Neo4j: http://localhost:7474"
echo ""
echo "📋 健康检查："
echo "  curl http://localhost:8080/health"
echo ""
echo "🛑 停止服务："
echo "  docker-compose down"
echo "  kill $AI_PID $GO_PID"
echo ""

# 等待用户中断
trap "echo '🛑 正在停止服务...'; kill $AI_PID $GO_PID 2>/dev/null; docker-compose down; exit 0" INT
wait 