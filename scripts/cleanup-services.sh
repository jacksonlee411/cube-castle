#!/bin/bash

echo "🧹 清理旧服务进程..."

# 清理所有Go临时进程
pkill -f "/tmp/go-build.*exe/main" 2>/dev/null || true
pkill -f "go run main" 2>/dev/null || true

# 等待进程完全退出
sleep 2

echo "🔍 检查端口占用..."
PORTS=(9090 9091 8087 8090)
for port in "${PORTS[@]}"; do
    if netstat -tlnp 2>/dev/null | grep -q ":$port "; then
        echo "⚠️  端口 $port 仍被占用"
        # 强制清理占用端口的进程
        fuser -k $port/tcp 2>/dev/null || true
    else
        echo "✅ 端口 $port 可用"
    fi
done

echo "✨ 清理完成，可以安全启动服务"