#!/bin/bash
# 修复WSL mirrored模式下的Docker代理配置

set -e

echo "🔧 修复WSL mirrored模式Docker代理配置..."

# 1. 创建systemd覆盖目录（如果使用systemd）
if command -v systemctl > /dev/null 2>&1; then
    echo "检测到systemd，配置Docker service代理..."
    
    sudo mkdir -p /etc/systemd/system/docker.service.d
    
    # 创建代理配置文件
    sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf > /dev/null << EOF
[Service]
Environment="HTTP_PROXY=http://localhost:7890"
Environment="HTTPS_PROXY=http://localhost:7890"
Environment="NO_PROXY=localhost,127.0.0.1,*.local,*.internal"
EOF

    # 重载systemd并重启docker
    echo "重新加载systemd配置..."
    sudo systemctl daemon-reload
    
    echo "重启Docker服务..."
    sudo systemctl restart docker
    
    # 等待Docker启动
    echo "等待Docker服务启动..."
    sleep 10
    
else
    echo "非systemd环境，尝试其他方法..."
fi

# 2. 配置Docker客户端代理
echo "配置Docker客户端代理..."
mkdir -p ~/.docker

cat > ~/.docker/config.json << EOF
{
  "credsStore": "desktop.exe",
  "proxies": {
    "default": {
      "httpProxy": "http://localhost:7890",
      "httpsProxy": "http://localhost:7890",
      "noProxy": "localhost,127.0.0.1,*.local,*.internal"
    }
  }
}
EOF

# 3. 设置环境变量
echo "设置代理环境变量..."
export HTTP_PROXY=http://localhost:7890
export HTTPS_PROXY=http://localhost:7890
export NO_PROXY=localhost,127.0.0.1,*.local,*.internal

# 4. 测试配置
echo "测试Docker代理配置..."
echo "当前Docker系统信息："
docker system info | grep -i proxy || echo "未找到代理信息"

# 5. 测试镜像拉取
echo "测试Docker镜像拉取..."
if docker pull hello-world > /dev/null 2>&1; then
    echo "✅ Docker代理配置成功！"
    
    echo "🚀 开始拉取Kafka镜像..."
    
    # 拉取原版Kafka镜像
    echo "拉取Zookeeper..."
    docker pull confluentinc/cp-zookeeper:7.4.0
    
    echo "拉取Kafka..."
    docker pull confluentinc/cp-kafka:7.4.0
    
    echo "拉取Debezium Connect..."
    docker pull debezium/connect:2.4
    
    echo "拉取Kafka UI..."
    docker pull provectuslabs/kafka-ui:latest
    
    echo "✅ 所有Kafka镜像拉取完成！"
    
    # 恢复原始docker-compose.yml配置
    echo "恢复原版镜像配置..."
    if [ -f docker-compose.backup.yml ]; then
        # 只恢复Kafka相关的镜像配置
        sed -i 's|registry.cn-hangzhou.aliyuncs.com/zhengqing/zookeeper:3.8.0|confluentinc/cp-zookeeper:7.4.0|g' docker-compose.yml
        sed -i 's|registry.cn-hangzhou.aliyuncs.com/zhengqing/kafka:3.2.0|confluentinc/cp-kafka:7.4.0|g' docker-compose.yml
        sed -i 's|registry.cn-hangzhou.aliyuncs.com/zhengqing/kafka-connect:2.4|debezium/connect:2.4|g' docker-compose.yml
        sed -i 's|registry.cn-hangzhou.aliyuncs.com/zhengqing/kafka-ui:latest|provectuslabs/kafka-ui:latest|g' docker-compose.yml
        
        echo "✅ docker-compose.yml已恢复为原版镜像配置"
    fi
    
else
    echo "❌ Docker代理配置仍有问题"
    echo "请检查Windows端代理软件设置，确保："
    echo "1. 代理软件正在运行"
    echo "2. 监听端口7890"
    echo "3. 允许局域网连接"
    exit 1
fi

echo ""
echo "🎉 Docker代理问题已解决！"
echo "现在可以正常拉取Kafka镜像并启动Operation Phoenix了！"