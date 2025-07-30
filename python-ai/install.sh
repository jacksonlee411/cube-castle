#!/bin/bash

# Cube Castle AI Service - 安装脚本
# 用于设置Python AI服务的开发和生产环境

set -e

echo "🏗️ 开始安装 Cube Castle AI Service..."

# 检查Python版本
python_version=$(python3 --version 2>&1)
echo "✅ Python版本: $python_version"

# 检查并创建虚拟环境
if [ ! -d "venv" ]; then
    echo "📦 创建Python虚拟环境..."
    python3 -m venv venv
    echo "✅ 虚拟环境创建完成"
else
    echo "✅ 虚拟环境已存在"
fi

# 激活虚拟环境
echo "🔄 激活虚拟环境..."
source venv/bin/activate

# 升级pip
echo "⬆️ 升级pip..."
pip install --upgrade pip

# 安装生产依赖
echo "📥 安装生产环境依赖..."
pip install -r requirements.txt

# 检查是否需要安装开发依赖
if [ "$1" = "--dev" ] || [ "$1" = "-d" ]; then
    echo "🔧 安装开发环境依赖..."
    pip install -r requirements-dev.txt
fi

# 验证关键导入
echo "🧪 验证关键依赖包..."
python -c "
import grpc
from grpc_health.v1 import health, health_pb2, health_pb2_grpc
import openai
import redis
from dotenv import load_dotenv
print('✅ 所有关键依赖导入成功!')
print('gRPC版本:', grpc.__version__)
print('OpenAI版本:', openai.__version__)
print('Redis版本:', redis.__version__)
"

# 检查.env文件
if [ ! -f ".env" ]; then
    echo "⚠️ 未找到.env文件，请创建并配置以下环境变量:"
    echo "   OPENAI_API_KEY=your_openai_api_key"
    echo "   OPENAI_API_BASE_URL=https://api.openai.com/v1"
    echo "   REDIS_HOST=localhost"
    echo "   REDIS_PORT=6379"
else
    echo "✅ .env配置文件已存在"
fi

echo "🎉 Cube Castle AI Service 安装完成！"
echo ""
echo "使用方法："
echo "  生产环境: ./start.sh"
echo "  开发模式: ./start.sh --dev"
echo "  健康检查: ./health-check.sh"
echo ""