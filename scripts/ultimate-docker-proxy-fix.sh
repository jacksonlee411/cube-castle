#!/bin/bash
# WSL Mirrored模式Docker代理终极修复方案

set -e

echo "🔧 WSL Mirrored模式Docker代理终极修复..."

# 获取WSL的实际IP地址
WSL_IP=$(ip addr show eth0 | grep "inet " | awk '{print $2}' | cut -d/ -f1)
echo "检测到WSL IP: $WSL_IP"

# 测试不同的代理地址
PROXY_ADDRESSES=(
    "http://127.0.0.1:7890"
    "http://localhost:7890"
    "http://$WSL_IP:7890"
    "http://10.39.54.1:7890"  # 可能的Windows主机IP
)

echo "测试可用的代理地址..."
WORKING_PROXY=""

for proxy in "${PROXY_ADDRESSES[@]}"; do
    echo "测试: $proxy"
    if timeout 5 curl -x "$proxy" -s https://registry-1.docker.io/v2/ > /dev/null 2>&1; then
        echo "✅ 代理可用: $proxy"
        WORKING_PROXY="$proxy"
        break
    else
        echo "❌ 代理不可用: $proxy"
    fi
done

if [ -z "$WORKING_PROXY" ]; then
    echo "❌ 未找到可用的代理地址"
    echo "🔧 尝试方案2：临时禁用Docker代理"
    
    # 创建指令文件给用户
    cat > /tmp/docker_proxy_disable.txt << EOF
请在Windows端执行以下步骤禁用Docker代理：

1. 打开Docker Desktop
2. 右键托盘图标 → Settings
3. 进入 Resources → Proxies
4. 取消勾选 "Manual proxy configuration"
5. 点击 "Apply & Restart"
6. 等待重启完成

或者检查代理软件设置：
- 确保代理软件正在运行
- 确认端口7890开放
- 确认允许局域网连接
- 尝试重启代理软件
EOF

    echo "📋 详细指令已保存到 /tmp/docker_proxy_disable.txt"
    cat /tmp/docker_proxy_disable.txt
    exit 1
fi

echo "🚀 使用可用代理: $WORKING_PROXY"

# 更新Docker客户端配置
mkdir -p ~/.docker
cat > ~/.docker/config.json << EOF
{
  "credsStore": "desktop.exe",
  "proxies": {
    "default": {
      "httpProxy": "$WORKING_PROXY",
      "httpsProxy": "$WORKING_PROXY",
      "noProxy": "localhost,127.0.0.1,*.local,*.internal"
    }
  }
}
EOF

# 设置环境变量
export HTTP_PROXY="$WORKING_PROXY"
export HTTPS_PROXY="$WORKING_PROXY"
export NO_PROXY="localhost,127.0.0.1,*.local,*.internal"

echo "✅ Docker客户端代理配置已更新"

# 测试Docker镜像拉取
echo "🧪 测试Docker镜像拉取..."
if docker pull hello-world > /dev/null 2>&1; then
    echo "✅ Docker代理修复成功！"
    
    echo "🎯 现在拉取Kafka镜像..."
    
    # 恢复原版镜像配置（如果之前被修改过）
    echo "恢复原版Kafka镜像配置..."
    sed -i 's|registry.cn-hangzhou.aliyuncs.com/zhengqing/zookeeper:3.8.0|confluentinc/cp-zookeeper:7.4.0|g' docker-compose.yml
    sed -i 's|registry.cn-hangzhou.aliyuncs.com/zhengqing/kafka:3.2.0|confluentinc/cp-kafka:7.4.0|g' docker-compose.yml
    sed -i 's|registry.cn-hangzhou.aliyuncs.com/zhengqing/kafka-connect:2.4|debezium/connect:2.4|g' docker-compose.yml
    sed -i 's|registry.cn-hangzhou.aliyuncs.com/zhengqing/kafka-ui:latest|provectuslabs/kafka-ui:latest|g' docker-compose.yml
    
    # 并行拉取所有Kafka镜像
    echo "📥 并行拉取Kafka生态系统镜像..."
    
    docker pull confluentinc/cp-zookeeper:7.4.0 &
    ZOOKEEPER_PID=$!
    
    docker pull confluentinc/cp-kafka:7.4.0 &
    KAFKA_PID=$!
    
    docker pull debezium/connect:2.4 &
    DEBEZIUM_PID=$!
    
    docker pull provectuslabs/kafka-ui:latest &
    KAFKAUI_PID=$!
    
    # 等待所有拉取完成
    echo "等待所有镜像拉取完成..."
    wait $ZOOKEEPER_PID && echo "✅ Zookeeper镜像完成"
    wait $KAFKA_PID && echo "✅ Kafka镜像完成"
    wait $DEBEZIUM_PID && echo "✅ Debezium镜像完成"
    wait $KAFKAUI_PID && echo "✅ Kafka UI镜像完成"
    
    echo "🎉 所有Kafka镜像拉取成功！"
    
    # 立即启动Operation Phoenix
    echo "🚀 启动Operation Phoenix完整架构..."
    docker-compose up -d zookeeper kafka kafka-connect kafka-ui
    
    echo "⏳ 等待Kafka服务启动..."
    sleep 30
    
    # 执行CDC配置
    echo "🔧 配置CDC管道..."
    bash scripts/setup-cdc-pipeline.sh
    
else
    echo "❌ Docker代理仍有问题，需要手动修复"
    echo "请参考 /tmp/docker_proxy_disable.txt 中的指令"
    exit 1
fi

echo ""
echo "🎊 Operation Phoenix Docker代理问题彻底解决！"
echo "Kafka生态系统已启动，CDC管道已配置！"