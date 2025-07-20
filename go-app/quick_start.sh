#!/bin/bash

echo "⚡ 快速启动脚本"
echo "=============="

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "❌ 错误: 请在 go-app 目录下运行此脚本"
    exit 1
fi

echo "✅ 当前目录: $(pwd)"

# 快速清理
echo "🧹 快速清理..."
rm -f go.sum

# 设置环境变量
export APP_PORT=8080
export INTELLIGENCE_SERVICE_GRPC_TARGET=localhost:50051

echo "📝 环境变量设置:"
echo "  APP_PORT=$APP_PORT"
echo "  INTELLIGENCE_SERVICE_GRPC_TARGET=$INTELLIGENCE_SERVICE_GRPC_TARGET"

# 启动服务器
echo ""
echo "🚀 启动 CoreHR API 服务器..."
echo "📍 服务地址: http://localhost:$APP_PORT"
echo "📋 API 文档: http://localhost:$APP_PORT/test.html"
echo "🏥 健康检查: http://localhost:$APP_PORT/health"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""

# 启动服务器
go run cmd/server/main.go 