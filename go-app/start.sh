#!/bin/bash

# CoreHR API 启动脚本
# 用于快速启动和测试 CoreHR API

set -e

echo "🏰 Cube Castle CoreHR API 启动脚本"
echo "=================================="

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "❌ 错误: 未找到 Go 命令，请先安装 Go"
    exit 1
fi

echo "✅ Go 已安装: $(go version)"

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "❌ 错误: 请在 go-app 目录下运行此脚本"
    exit 1
fi

echo "✅ 当前目录: $(pwd)"

# 检查依赖
echo "📦 检查依赖..."
go mod tidy

# 编译项目
echo "🔨 编译项目..."
if go build -o server cmd/server/main.go; then
    echo "✅ 编译成功"
else
    echo "❌ 编译失败"
    exit 1
fi

# 检查环境变量
if [ -z "$APP_PORT" ]; then
    export APP_PORT=8080
    echo "📝 设置默认端口: $APP_PORT"
fi

if [ -z "$INTELLIGENCE_SERVICE_GRPC_TARGET" ]; then
    export INTELLIGENCE_SERVICE_GRPC_TARGET=localhost:50051
    echo "📝 设置默认 gRPC 目标: $INTELLIGENCE_SERVICE_GRPC_TARGET"
fi

# 检查数据库连接
echo "🗄️  检查数据库连接..."
if [ -f ".env" ]; then
    echo "📝 加载环境变量文件"
    source .env
fi

# 启动服务器
echo "🚀 启动 CoreHR API 服务器..."
echo "📍 服务地址: http://localhost:$APP_PORT"
echo "📋 API 文档: http://localhost:$APP_PORT/test.html"
echo "🏥 健康检查: http://localhost:$APP_PORT/health"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""

# 启动服务器
./server 