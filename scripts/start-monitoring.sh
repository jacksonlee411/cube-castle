#!/bin/bash

# Cube Castle 监控系统一键启动脚本
# 自动部署 Prometheus + Grafana + AlertManager

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MONITORING_DIR="$PROJECT_ROOT/monitoring"

echo "🚀 启动 Cube Castle 监控系统..."

# 检查Docker是否运行
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 检查Docker Compose是否可用
if ! docker compose version >/dev/null 2>&1; then
    echo "❌ Docker Compose不可用，请安装Docker Compose"
    exit 1
fi

# 切换到监控目录
cd "$MONITORING_DIR"

echo "📂 当前工作目录: $MONITORING_DIR"

# 检查配置文件
echo "🔍 检查配置文件..."
REQUIRED_FILES=(
    "docker-compose.monitoring.yml"
    "prometheus.yml"
    "prometheus-rules.yml"
    "grafana/provisioning/datasources/prometheus.yml"
    "grafana/provisioning/dashboards/dashboard-config.yml"
    "grafana/dashboards/slo-dashboard.json"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [[ ! -f "$file" ]]; then
        echo "❌ 缺少配置文件: $file"
        exit 1
    fi
done

echo "✅ 所有配置文件检查完毕"

# 创建必要的目录
echo "📁 创建数据目录..."
mkdir -p data/{prometheus,grafana,alertmanager}

# 设置Grafana权限
echo "🔐 设置Grafana目录权限..."
sudo chown -R 472:472 data/grafana 2>/dev/null || true

# 启动监控系统
echo "🎬 启动监控服务..."
docker compose -f docker-compose.monitoring.yml up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo "🔍 检查服务状态..."
SERVICES=("prometheus" "grafana" "alertmanager" "node-exporter")

for service in "${SERVICES[@]}"; do
    container_name="cube-castle-$service"
    if docker ps --format "table {{.Names}}" | grep -q "$container_name"; then
        echo "✅ $service 运行中"
    else
        echo "❌ $service 启动失败"
        docker logs "$container_name" --tail 20
    fi
done

# 显示访问信息
echo ""
echo "🎉 监控系统启动完成！"
echo ""
echo "📊 访问地址:"
echo "  • Prometheus:  http://localhost:9091"
echo "  • Grafana:     http://localhost:3001 (admin/cube-castle-2025)"
echo "  • AlertManager: http://localhost:9093"
echo "  • Node Exporter: http://localhost:9100"
echo ""

# 检查API服务连接
echo "🔗 检查API服务连接..."
if curl -s http://localhost:9090/health >/dev/null 2>&1; then
    echo "✅ 组织API服务 (9090) 连接正常"
else
    echo "⚠️  组织API服务 (9090) 未运行或不可访问"
    echo "   请确保后端API服务已启动"
fi

if curl -s http://localhost:8090/health >/dev/null 2>&1; then
    echo "✅ GraphQL服务 (8090) 连接正常"
else
    echo "⚠️  GraphQL服务 (8090) 未运行或不可访问"
    echo "   请确保GraphQL查询服务已启动"
fi

echo ""
echo "📖 使用指南:"
echo "  1. 访问Grafana: http://localhost:3001"
echo "  2. 使用账号: admin / cube-castle-2025"
echo "  3. 查看'组织启停API - SLO监控仪表板'"
echo "  4. 监控指标将在API服务运行时自动采集"
echo ""

# 显示有用的命令
echo "🛠️  常用命令:"
echo "  查看日志: docker compose -f docker-compose.monitoring.yml logs -f [service]"
echo "  停止监控: docker compose -f docker-compose.monitoring.yml down"
echo "  重启监控: docker compose -f docker-compose.monitoring.yml restart"
echo ""

echo "✨ 监控系统部署完成，开始监控您的组织API服务！"
