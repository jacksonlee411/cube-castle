#!/bin/bash

# Cube Castle AI Service - 健康检查脚本
# 检查AI服务的健康状态

set -e

# 默认配置
HOST=${AI_HOST:-localhost}
PORT=${AI_PORT:-50051}
TIMEOUT=${HEALTH_TIMEOUT:-10}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --host|-h)
            HOST="$2"
            shift 2
            ;;
        --port|-p)
            PORT="$2"
            shift 2
            ;;
        --timeout|-t)
            TIMEOUT="$2"
            shift 2
            ;;
        --help)
            echo "Cube Castle AI Service 健康检查脚本"
            echo ""
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --host, -h HOST     AI服务主机 (默认: localhost)"
            echo "  --port, -p PORT     AI服务端口 (默认: 50051)"
            echo "  --timeout, -t SEC   检查超时时间 (默认: 10秒)"
            echo "  --help              显示此帮助信息"
            exit 0
            ;;
        *)
            echo "未知选项: $1"
            echo "使用 --help 查看可用选项"
            exit 1
            ;;
    esac
done

echo "🏥 Cube Castle AI Service 健康检查"
echo "=================================="
echo "检查目标: $HOST:$PORT"
echo "超时时间: ${TIMEOUT}秒"
echo ""

# 激活虚拟环境
if [ -d "venv" ]; then
    source venv/bin/activate
fi

# 基础连接检查
echo "🔍 1. 检查端口连接..."
if timeout $TIMEOUT bash -c "</dev/tcp/$HOST/$PORT"; then
    echo "✅ 端口 $PORT 可访问"
else
    echo "❌ 无法连接到 $HOST:$PORT"
    echo "   请确保AI服务正在运行"
    exit 1
fi

# gRPC健康检查
echo ""
echo "🔍 2. 检查gRPC健康状态..."
python - <<EOF
import grpc
from grpc_health.v1 import health_pb2, health_pb2_grpc
import sys
import signal

def timeout_handler(signum, frame):
    print("❌ gRPC健康检查超时")
    sys.exit(1)

try:
    signal.signal(signal.SIGALRM, timeout_handler)
    signal.alarm($TIMEOUT)
    
    channel = grpc.insecure_channel('$HOST:$PORT')
    stub = health_pb2_grpc.HealthStub(channel)
    
    # 检查整体服务健康状态
    request = health_pb2.HealthCheckRequest(service="")
    response = stub.Check(request)
    
    if response.status == health_pb2.HealthCheckResponse.SERVING:
        print("✅ gRPC服务健康状态正常")
    else:
        print(f"⚠️ gRPC服务状态: {response.status}")
        sys.exit(1)
    
    # 检查AI服务健康状态
    request = health_pb2.HealthCheckRequest(service="intelligence")
    response = stub.Check(request)
    
    if response.status == health_pb2.HealthCheckResponse.SERVING:
        print("✅ AI智能服务健康状态正常")
    else:
        print(f"⚠️ AI智能服务状态: {response.status}")
        sys.exit(1)
        
    signal.alarm(0)
    channel.close()
    
except grpc.RpcError as e:
    print(f"❌ gRPC调用失败: {e}")
    sys.exit(1)
except Exception as e:
    print(f"❌ 健康检查异常: {e}")
    sys.exit(1)
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 健康检查完成！AI服务运行正常"
    echo ""
    echo "服务信息:"
    echo "  地址: $HOST:$PORT"
    echo "  状态: 健康"
    echo "  时间: $(date)"
else
    echo ""
    echo "💥 健康检查失败！请检查服务状态"
    exit 1
fi