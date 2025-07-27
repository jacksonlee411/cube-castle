#!/bin/bash

# Cube Castle AI Service - 启动脚本
# 用于启动Python gRPC AI服务

set -e

# 默认配置
PORT=${AI_PORT:-50051}
HOST=${AI_HOST:-0.0.0.0}
LOG_LEVEL=${LOG_LEVEL:-INFO}
DEV_MODE=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --dev|-d)
            DEV_MODE=true
            shift
            ;;
        --port|-p)
            PORT="$2"
            shift 2
            ;;
        --host|-h)
            HOST="$2"
            shift 2
            ;;
        --log-level|-l)
            LOG_LEVEL="$2"
            shift 2
            ;;
        --help)
            echo "Cube Castle AI Service 启动脚本"
            echo ""
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --dev, -d           开发模式（启用详细日志）"
            echo "  --port, -p PORT     gRPC服务端口 (默认: 50051)"
            echo "  --host, -h HOST     绑定主机 (默认: 0.0.0.0)"
            echo "  --log-level, -l LEVEL 日志级别 (默认: INFO)"
            echo "  --help              显示此帮助信息"
            echo ""
            echo "环境变量:"
            echo "  AI_PORT             gRPC服务端口"
            echo "  AI_HOST             绑定主机"
            echo "  LOG_LEVEL           日志级别"
            echo "  OPENAI_API_KEY      OpenAI API密钥"
            echo "  OPENAI_API_BASE_URL OpenAI API基础URL"
            echo "  REDIS_HOST          Redis主机"
            echo "  REDIS_PORT          Redis端口"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            echo "使用 --help 查看可用选项"
            exit 1
            ;;
    esac
done

echo "🧙 启动 Cube Castle AI Service - The Wizard Tower"
echo "=================================================="

# 检查虚拟环境
if [ ! -d "venv" ]; then
    echo "❌ 虚拟环境不存在，请先运行 ./install.sh"
    exit 1
fi

# 激活虚拟环境
echo "🔄 激活虚拟环境..."
source venv/bin/activate

# 检查环境配置
if [ ! -f ".env" ]; then
    echo "⚠️ .env文件不存在，使用默认配置"
else
    echo "✅ 加载.env配置文件"
    source .env
fi

# 检查Redis连接
echo "🔍 检查Redis连接..."
if timeout 5 bash -c "</dev/tcp/${REDIS_HOST:-localhost}/${REDIS_PORT:-6379}"; then
    echo "✅ Redis连接正常"
else
    echo "⚠️ Redis连接失败，确保Redis服务正在运行"
    echo "   默认地址: ${REDIS_HOST:-localhost}:${REDIS_PORT:-6379}"
fi

# 设置开发模式
if [ "$DEV_MODE" = true ]; then
    echo "🔧 开发模式已启用"
    export LOG_LEVEL=DEBUG
    export PYTHONPATH=".:$PYTHONPATH"
fi

# 显示配置信息
echo ""
echo "配置信息:"
echo "  服务地址: $HOST:$PORT"
echo "  日志级别: $LOG_LEVEL"
echo "  开发模式: $DEV_MODE"
echo "  Redis: ${REDIS_HOST:-localhost}:${REDIS_PORT:-6379}"
echo ""

# 启动服务
echo "🚀 正在启动AI服务..."
echo "   按 Ctrl+C 停止服务"
echo ""

# 设置环境变量
export AI_PORT=$PORT
export AI_HOST=$HOST

# 启动Python gRPC服务
python main.py