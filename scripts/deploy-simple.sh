#!/bin/bash

# 简化版生产部署 - 适用于开发环境测试

echo "🚀 开始简化生产部署..."
echo "==============================="

# 检查当前服务
if pgrep -f "./bin/server" > /dev/null; then
    echo "🛑 停止现有服务..."
    pkill -f "./bin/server"
    sleep 2
fi

# 构建生产版本
echo "🔨 构建生产版本..."
go build -ldflags="-w -s" -o ./bin/server-production ./cmd/server/main.go
echo "✅ 构建完成"

# 创建生产配置
echo "⚙️ 创建生产配置..."
cat > ./production.env <<EOF
API_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_NAME=cubecastle
DB_USER=user
DB_PASSWORD=password
LOG_LEVEL=info
EOF

# 创建日志目录
mkdir -p ./logs

# 启动生产服务
echo "🚀 启动生产服务..."
nohup ./bin/server-production > ./logs/production.log 2>&1 &
echo $! > ./production.pid

sleep 3

# 健康检查
echo "🩺 健康检查..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ 生产服务启动成功！"
    echo ""
    echo "📊 服务信息:"
    echo "  PID: $(cat ./production.pid)"
    echo "  端口: 8080"
    echo "  健康检查: http://localhost:8080/health"
    echo "  API: http://localhost:8080/api/v1/organization-units"
    echo "  日志: ./logs/production.log"
    echo ""
    echo "管理命令:"
    echo "  停止: kill \$(cat ./production.pid)"
    echo "  日志: tail -f ./logs/production.log"
else
    echo "❌ 服务启动失败"
    exit 1
fi