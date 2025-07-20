#!/bin/bash

echo "🔧 修复 Go 模块锁定问题"
echo "========================"

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "❌ 错误: 请在 go-app 目录下运行此脚本"
    exit 1
fi

echo "✅ 当前目录: $(pwd)"

# 清理 Go 模块缓存
echo "🧹 清理 Go 模块缓存..."
go clean -modcache

# 删除可能损坏的文件
echo "🗑️  删除可能损坏的文件..."
rm -f go.sum
rm -rf vendor/

# 重新初始化模块
echo "🔄 重新初始化 Go 模块..."
go mod tidy

# 验证模块
echo "✅ 验证模块..."
go mod verify

echo ""
echo "🎉 修复完成！"
echo ""
echo "现在可以尝试启动服务器："
echo "go run cmd/server/main.go" 