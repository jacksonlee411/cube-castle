#!/bin/bash

# 生产环境部署脚本
# 组织单元API v2.0 - 7位编码版本

set -e

echo "🚀 开始生产环境部署..."
echo "==============================="

# 配置变量
API_PORT=${API_PORT:-8080}
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432} 
DB_NAME=${DB_NAME:-cubecastle}
DB_USER=${DB_USER:-user}
DB_PASSWORD=${DB_PASSWORD:-password}
SERVICE_NAME="organization-units-api"
LOG_DIR="/var/log/${SERVICE_NAME}"
PID_FILE="/var/run/${SERVICE_NAME}.pid"

# 创建必要目录
echo "📁 创建目录结构..."
sudo mkdir -p $LOG_DIR
sudo mkdir -p /var/run
sudo mkdir -p /opt/${SERVICE_NAME}

# 检查数据库连接
echo "🔍 检查数据库连接..."
if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ 数据库连接失败"
    exit 1
fi
echo "✅ 数据库连接正常"

# 构建应用
echo "🔨 构建生产版本..."
if [ ! -f "go.mod" ]; then
    go mod init cube-castle-production
    go mod tidy
fi

CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ./bin/server-production ./cmd/server/main.go
echo "✅ 应用构建完成"

# 复制文件到生产目录
echo "📦 部署文件..."
sudo cp ./bin/server-production /opt/${SERVICE_NAME}/server
sudo chmod +x /opt/${SERVICE_NAME}/server

# 创建配置文件
echo "⚙️ 创建配置文件..."
sudo tee /opt/${SERVICE_NAME}/config.env > /dev/null <<EOF
API_PORT=$API_PORT
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_NAME=$DB_NAME
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
LOG_LEVEL=info
GIN_MODE=release
EOF

# 创建systemd服务文件
echo "🔧 创建系统服务..."
sudo tee /etc/systemd/system/${SERVICE_NAME}.service > /dev/null <<EOF
[Unit]
Description=Organization Units API v2.0
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=cubecastle
Group=cubecastle
WorkingDirectory=/opt/${SERVICE_NAME}
EnvironmentFile=/opt/${SERVICE_NAME}/config.env
ExecStart=/opt/${SERVICE_NAME}/server
ExecReload=/bin/kill -HUP \$MAINPID
Restart=always
RestartSec=5
StandardOutput=append:${LOG_DIR}/access.log
StandardError=append:${LOG_DIR}/error.log

# 安全配置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=${LOG_DIR}

[Install]
WantedBy=multi-user.target
EOF

# 创建用户
echo "👤 创建服务用户..."
if ! id "cubecastle" &>/dev/null; then
    sudo useradd -r -s /bin/false cubecastle
fi
sudo chown -R cubecastle:cubecastle /opt/${SERVICE_NAME}
sudo chown -R cubecastle:cubecastle $LOG_DIR

# 重载systemd并启动服务
echo "🔄 启动服务..."
sudo systemctl daemon-reload
sudo systemctl enable ${SERVICE_NAME}
sudo systemctl start ${SERVICE_NAME}

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 5

# 检查服务状态
if sudo systemctl is-active --quiet ${SERVICE_NAME}; then
    echo "✅ 服务启动成功"
    
    # 健康检查
    echo "🩺 执行健康检查..."
    if curl -s http://localhost:$API_PORT/health > /dev/null; then
        echo "✅ 健康检查通过"
        
        # 显示服务信息
        echo ""
        echo "🎉 生产环境部署完成！"
        echo "==============================="
        echo "服务名称: $SERVICE_NAME"
        echo "监听端口: $API_PORT"
        echo "健康检查: http://localhost:$API_PORT/health"
        echo "API端点: http://localhost:$API_PORT/api/v1/organization-units"
        echo "日志目录: $LOG_DIR"
        echo ""
        echo "管理命令:"
        echo "  启动: sudo systemctl start $SERVICE_NAME"
        echo "  停止: sudo systemctl stop $SERVICE_NAME"
        echo "  重启: sudo systemctl restart $SERVICE_NAME"
        echo "  状态: sudo systemctl status $SERVICE_NAME"
        echo "  日志: sudo journalctl -u $SERVICE_NAME -f"
        
    else
        echo "❌ 健康检查失败"
        sudo systemctl status ${SERVICE_NAME}
        exit 1
    fi
else
    echo "❌ 服务启动失败"
    sudo systemctl status ${SERVICE_NAME}
    exit 1
fi

echo ""
echo "🔥 生产环境API服务器已就绪！"