#!/bin/bash
# Docker代理配置解决方案 - WSL环境优化版
# 解决Operation Phoenix中Kafka镜像拉取问题

set -e

echo "🔧 Docker代理诊断与修复脚本..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() { echo -e "${GREEN}[INFO]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 1. 检查当前Docker状态
print_status "检查当前Docker配置..."
echo "Docker代理设置:"
docker system info | grep -i proxy || echo "  无系统级代理"

echo -e "\nDocker客户端配置:"
cat ~/.docker/config.json 2>/dev/null || echo "  无客户端配置"

# 2. 测试本地代理
print_status "测试代理连接..."
PROXY_WORKS=false

# 测试本地7890端口
if curl -x http://127.0.0.1:7890 --connect-timeout 5 -s http://www.baidu.com > /dev/null 2>&1; then
    print_status "✅ 本地代理 127.0.0.1:7890 可用"
    PROXY_WORKS=true
    PROXY_URL="http://127.0.0.1:7890"
else
    print_warning "本地7890代理不可用，尝试WSL网关代理..."
    
    # 获取Windows主机IP
    WINDOWS_HOST=$(ip route show | grep default | awk '{print $3}')
    echo "Windows主机IP: $WINDOWS_HOST"
    
    # 测试WSL网关代理
    for port in 7890 1080 8080; do
        if curl -x http://$WINDOWS_HOST:$port --connect-timeout 3 -s http://www.baidu.com > /dev/null 2>&1; then
            print_status "✅ 找到可用代理: $WINDOWS_HOST:$port"
            PROXY_WORKS=true
            PROXY_URL="http://$WINDOWS_HOST:$port"
            break
        fi
    done
fi

# 3. 根据代理状态选择方案
if [ "$PROXY_WORKS" = false ]; then
    print_error "未找到可用代理，使用无代理方案"
    
    # 方案A: 禁用代理，使用国内镜像源
    print_status "配置无代理Docker..."
    mkdir -p ~/.docker
    cat > ~/.docker/config.json << EOF
{
  "credsStore": "desktop.exe"
}
EOF

    # 测试直连Docker Hub
    print_status "测试直连Docker Hub..."
    if docker pull hello-world:latest > /dev/null 2>&1; then
        print_status "✅ 直连成功，开始拉取Kafka镜像"
        docker pull confluentinc/cp-zookeeper:7.4.0 &
        docker pull confluentinc/cp-kafka:7.4.0 &
        wait
        print_status "✅ 无代理方案成功"
    else
        print_warning "直连失败，尝试国内镜像源..."
        print_status "更新docker-compose.yml使用国内镜像"
        
        # 备份原文件
        cp docker-compose.yml docker-compose.yml.backup 2>/dev/null || true
        
        echo "建议手动检查代理软件配置或使用VPN"
    fi
    
else
    print_status "使用代理方案: $PROXY_URL"
    
    # 方案B: 配置代理
    mkdir -p ~/.docker
    cat > ~/.docker/config.json << EOF
{
  "credsStore": "desktop.exe",
  "proxies": {
    "default": {
      "httpProxy": "$PROXY_URL",
      "httpsProxy": "$PROXY_URL",
      "noProxy": "localhost,127.0.0.1,*.local,*.internal,hubproxy.docker.internal"
    }
  }
}
EOF

    print_status "✅ Docker客户端代理配置完成"
    
    # 测试代理效果
    print_status "测试Docker Hub连接..."
    if docker pull hello-world:latest > /dev/null 2>&1; then
        print_status "✅ 代理配置成功，开始拉取Kafka镜像"
        
        print_status "拉取Kafka相关镜像..."
        docker pull confluentinc/cp-zookeeper:7.4.0 &
        ZOOKEEPER_PID=$!
        docker pull confluentinc/cp-kafka:7.4.0 &
        KAFKA_PID=$!
        
        wait $ZOOKEEPER_PID && print_status "✅ Zookeeper镜像拉取完成"
        wait $KAFKA_PID && print_status "✅ Kafka镜像拉取完成"
        
    else
        print_error "代理配置失败"
        exit 1
    fi
fi

# 4. 最终验证和启动Operation Phoenix
print_status "准备启动Operation Phoenix环境..."

# 验证所需镜像
REQUIRED_IMAGES=("confluentinc/cp-zookeeper:7.4.0" "confluentinc/cp-kafka:7.4.0")
MISSING_IMAGES=()

for image in "${REQUIRED_IMAGES[@]}"; do
    if ! docker images --format "table {{.Repository}}:{{.Tag}}" | grep -q "$image"; then
        MISSING_IMAGES+=("$image")
    fi
done

if [ ${#MISSING_IMAGES[@]} -gt 0 ]; then
    print_warning "缺少以下镜像，尝试拉取: ${MISSING_IMAGES[*]}"
    for image in "${MISSING_IMAGES[@]}"; do
        print_status "拉取 $image ..."
        if ! docker pull "$image"; then
            print_error "拉取 $image 失败"
            exit 1
        fi
    done
fi

print_status "✅ 所有必需镜像已准备就绪"

# 提供下一步指令
echo ""
print_status "🚀 Docker代理配置完成！下一步操作："
echo "1. 启动Kafka生态系统: make phoenix-start"
echo "2. 或分步启动: docker-compose up -d zookeeper kafka"
echo "3. 检查服务状态: docker-compose ps"
echo ""
print_status "如有问题，请检查:"
echo "- Windows代理软件是否允许局域网连接"
echo "- Docker Desktop是否需要重启"
echo "- 防火墙设置是否正确"