#!/bin/bash

echo "🔧 最终修复脚本"
echo "================"

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "❌ 错误: 请在 go-app 目录下运行此脚本"
    exit 1
fi

echo "✅ 当前目录: $(pwd)"

# 检查 Go 版本
echo "📋 检查 Go 版本..."
go version

# 完全清理环境
echo "🧹 完全清理环境..."
rm -f go.sum
rm -rf vendor/
go clean -modcache
go clean -cache

# 重新初始化模块
echo "🔄 重新初始化 Go 模块..."
go mod download
go mod tidy

# 验证模块
echo "✅ 验证模块..."
go mod verify

# 设置环境变量
export APP_PORT=8080
export INTELLIGENCE_SERVICE_GRPC_TARGET=localhost:50051

echo "📝 环境变量设置:"
echo "  APP_PORT=$APP_PORT"
echo "  INTELLIGENCE_SERVICE_GRPC_TARGET=$INTELLIGENCE_SERVICE_GRPC_TARGET"

# 编译测试
echo "🔨 编译测试..."
go build -o /tmp/test_build cmd/server/main.go
if [ $? -eq 0 ]; then
    echo "✅ 编译成功！"
    rm -f /tmp/test_build
else
    echo "❌ 编译失败，请检查代码"
    exit 1
fi

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