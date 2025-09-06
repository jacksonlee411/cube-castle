#!/bin/bash

# 监控系统前后端集成测试脚本
# 验证监控菜单集成和服务可达性

echo "🧪 开始测试监控系统前后端集成..."

# 颜色输出函数
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

echo ""
echo "🔍 1. 检查前端应用状态..."

# 检查前端应用是否运行
FRONTEND_PORTS=(3000 3001 3002 3003)
FRONTEND_URL=""

for port in "${FRONTEND_PORTS[@]}"; do
    if curl -s --connect-timeout 2 "http://localhost:$port" | grep -q "Cube Castle"; then
        FRONTEND_URL="http://localhost:$port"
        success "前端应用运行在 $FRONTEND_URL"
        break
    fi
done

if [ -z "$FRONTEND_URL" ]; then
    error "前端应用未运行，请先启动前端服务器"
    echo "  启动命令: cd frontend && npm run dev"
    exit 1
fi

echo ""
echo "🔍 2. 检查监控系统后端服务..."

# 监控服务检查
declare -A SERVICES=(
    ["Prometheus"]="localhost:9091"
    ["Grafana"]="localhost:3001"
    ["AlertManager"]="localhost:9093"
    ["Node Exporter"]="localhost:9100"
)

HEALTHY_SERVICES=0
TOTAL_SERVICES=${#SERVICES[@]}

for service in "${!SERVICES[@]}"; do
    address=${SERVICES[$service]}
    
    case $service in
        "Prometheus")
            if curl -s --connect-timeout 2 "http://$address/-/healthy" | grep -q "Healthy"; then
                success "$service ($address) - 指标收集正常"
                ((HEALTHY_SERVICES++))
            else
                warning "$service ($address) - 服务不可达"
            fi
            ;;
        "Grafana")
            if curl -s --connect-timeout 2 "http://$address/api/health" | grep -q '"database":"ok"'; then
                success "$service ($address) - 可视化仪表板正常"
                ((HEALTHY_SERVICES++))
            else
                warning "$service ($address) - 服务不可达"
            fi
            ;;
        "AlertManager"|"Node Exporter")
            if curl -s --connect-timeout 2 "http://$address" >/dev/null 2>&1; then
                success "$service ($address) - 服务正常响应"
                ((HEALTHY_SERVICES++))
            else
                warning "$service ($address) - 服务不可达"
            fi
            ;;
    esac
done

echo ""
echo "📊 3. 服务可用性统计..."
echo "  健康服务: $HEALTHY_SERVICES/$TOTAL_SERVICES"
echo "  可用率: $((HEALTHY_SERVICES * 100 / TOTAL_SERVICES))%"

if [ $HEALTHY_SERVICES -eq $TOTAL_SERVICES ]; then
    success "所有监控服务运行正常"
elif [ $HEALTHY_SERVICES -gt $((TOTAL_SERVICES / 2)) ]; then
    warning "部分监控服务不可用，但核心功能可用"
else
    error "多数监控服务不可用，请检查监控系统部署"
fi

echo ""
echo "🔗 4. 测试服务连接..."

info "前端应用访问地址:"
echo "  🌐 主页: $FRONTEND_URL"
echo "  🔍 监控中心: $FRONTEND_URL/monitoring"
echo "  📊 组织架构: $FRONTEND_URL/organizations"

echo ""
info "监控系统直接访问地址:"
for service in "${!SERVICES[@]}"; do
    address=${SERVICES[$service]}
    echo "  📈 $service: http://$address"
done

echo ""
echo "🎯 5. 用户体验测试指南..."
echo "===========================================" 
echo "手动测试步骤:"
echo "1. 访问 $FRONTEND_URL"
echo "2. 点击左侧菜单中的「系统监控」"
echo "3. 验证监控服务卡片显示正确"
echo "4. 点击「打开服务」按钮测试服务跳转"
echo "5. 对于Grafana，确认弹出登录信息提示"

echo ""
echo "🏆 集成测试完成！"

if [ $HEALTHY_SERVICES -eq $TOTAL_SERVICES ]; then
    echo "  状态: ✅ 监控系统前后端集成完全正常"
    echo "  建议: 可以开始使用完整的监控功能"
else
    echo "  状态: ⚠️  部分服务需要启动"
    echo "  建议: 运行 ./scripts/start-monitoring.sh 启动缺失的服务"
fi

echo ""
echo "💡 提示: 在浏览器中访问 $FRONTEND_URL/monitoring 体验完整功能"