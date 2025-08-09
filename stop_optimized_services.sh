#!/bin/bash

# 停止优化后的组织架构服务脚本

echo "🛑 停止优化后的组织架构服务..."
echo "================================"

# 读取PID文件并停止服务
if [ -f logs/command-service.pid ]; then
    COMMAND_PID=$(cat logs/command-service.pid)
    if kill -0 $COMMAND_PID 2>/dev/null; then
        echo "🛑 停止简化命令服务 (PID: $COMMAND_PID)..."
        kill $COMMAND_PID
        sleep 2
        if kill -0 $COMMAND_PID 2>/dev/null; then
            echo "⚠️  强制停止命令服务..."
            kill -9 $COMMAND_PID
        fi
        echo "✅ 简化命令服务已停止"
    else
        echo "ℹ️  简化命令服务未运行"
    fi
    rm -f logs/command-service.pid
else
    echo "ℹ️  未找到命令服务PID文件"
fi

if [ -f logs/query-service.pid ]; then
    QUERY_PID=$(cat logs/query-service.pid)
    if kill -0 $QUERY_PID 2>/dev/null; then
        echo "🛑 停止统一查询服务 (PID: $QUERY_PID)..."
        kill $QUERY_PID
        sleep 2
        if kill -0 $QUERY_PID 2>/dev/null; then
            echo "⚠️  强制停止查询服务..."
            kill -9 $QUERY_PID
        fi
        echo "✅ 统一查询服务已停止"
    else
        echo "ℹ️  统一查询服务未运行"
    fi
    rm -f logs/query-service.pid
else
    echo "ℹ️  未找到查询服务PID文件"
fi

# 停止任何遗留进程
echo "🧹 清理遗留进程..."
pkill -f "organization-command-service-simplified" 2>/dev/null || true
pkill -f "organization-query-service" 2>/dev/null || true

echo ""
echo "✅ 所有优化后的服务已停止"
echo ""
echo "💡 重新启动服务: ./start_optimized_services.sh"