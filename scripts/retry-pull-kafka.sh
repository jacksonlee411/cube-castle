#!/bin/bash
# 持续重试拉取Kafka镜像脚本
# 作者: DevOps专家

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_status() { echo -e "${GREEN}[INFO]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 需要拉取的镜像列表
IMAGES=(
    "confluentinc/cp-zookeeper:7.4.0"
    "confluentinc/cp-kafka:7.4.0"
    "debezium/connect:2.4"
    "provectuslabs/kafka-ui:latest"
)

print_status "开始持续重试拉取Kafka相关镜像..."

# 检查已存在的镜像
check_image_exists() {
    local image=$1
    docker images --format "{{.Repository}}:{{.Tag}}" | grep -q "^${image}$"
}

# 拉取单个镜像的函数
pull_image_with_retry() {
    local image=$1
    local max_attempts=5
    local attempt=1
    
    print_status "拉取镜像: $image"
    
    # 如果镜像已存在，跳过
    if check_image_exists "$image"; then
        print_status "✅ 镜像 $image 已存在，跳过拉取"
        return 0
    fi
    
    while [ $attempt -le $max_attempts ]; do
        print_status "尝试 $attempt/$max_attempts: 拉取 $image"
        
        if timeout 600 docker pull "$image"; then
            print_status "✅ 成功拉取: $image"
            return 0
        else
            print_warning "❌ 第 $attempt 次尝试失败: $image"
            attempt=$((attempt + 1))
            
            if [ $attempt -le $max_attempts ]; then
                print_status "等待 30 秒后重试..."
                sleep 30
                
                # 清理可能的部分下载
                docker system prune -f > /dev/null 2>&1 || true
            fi
        fi
    done
    
    print_error "镜像 $image 在 $max_attempts 次尝试后仍然失败"
    return 1
}

# 主拉取循环
failed_images=()
for image in "${IMAGES[@]}"; do
    if ! pull_image_with_retry "$image"; then
        failed_images+=("$image")
    fi
done

# 结果汇总
echo ""
if [ ${#failed_images[@]} -eq 0 ]; then
    print_status "🎉 所有镜像拉取成功！"
    
    print_status "验证拉取的镜像:"
    for image in "${IMAGES[@]}"; do
        if check_image_exists "$image"; then
            echo "  ✅ $image"
        else
            echo "  ❌ $image"
        fi
    done
    
    print_status "准备启动Operation Phoenix..."
    exit 0
else
    print_error "以下镜像拉取失败:"
    for image in "${failed_images[@]}"; do
        echo "  ❌ $image"
    done
    
    print_warning "建议使用轻量级替代镜像:"
    echo "  - bitnami/zookeeper:3.8 替代 confluentinc/cp-zookeeper:7.4.0"
    echo "  - bitnami/kafka:3.4 替代 confluentinc/cp-kafka:7.4.0"
    
    exit 1
fi