#!/bin/bash

# 监控系统验证脚本
# 测试Prometheus、Grafana和监控指标是否正常工作

set -e

echo "🧪 开始验证监控系统..."

# 颜色输出函数
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# 检查服务可用性
echo "🔍 1. 检查监控服务状态..."

# 检查Prometheus
if curl -s http://localhost:9091/-/healthy >/dev/null 2>&1; then
    success "Prometheus (9091) 健康检查通过"
else
    error "Prometheus (9091) 不可访问"
    exit 1
fi

# 检查Grafana
if curl -s http://localhost:3001/api/health >/dev/null 2>&1; then
    success "Grafana (3001) 健康检查通过"
else
    error "Grafana (3001) 不可访问"
    exit 1
fi

# 检查AlertManager
if curl -s http://localhost:9093/-/healthy >/dev/null 2>&1; then
    success "AlertManager (9093) 健康检查通过"
else
    warning "AlertManager (9093) 不可访问或未启动"
fi

echo ""
echo "🎯 2. 验证Prometheus数据采集..."

# 检查Prometheus targets
TARGETS_UP=$(curl -s http://localhost:9091/api/v1/query?query=up | jq -r '.data.result | length')
if [ "$TARGETS_UP" -gt 0 ]; then
    success "发现 $TARGETS_UP 个监控目标"
    
    # 显示具体的targets状态
    curl -s http://localhost:9091/api/v1/targets | jq -r '.data.activeTargets[] | "  • \(.scrapePool): \(.health) (\(.lastScrape))"'
else
    warning "未发现活跃的监控目标，可能需要启动API服务"
fi

# 检查是否有指标数据
METRICS_COUNT=$(curl -s http://localhost:9091/api/v1/label/__name__/values | jq -r '.data | length')
if [ "$METRICS_COUNT" -gt 0 ]; then
    success "采集到 $METRICS_COUNT 个不同的指标"
else
    error "未采集到任何指标数据"
fi

echo ""
echo "📊 3. 验证组织API相关指标..."

# 检查组织API特定指标
API_METRICS=(
    "api_requests_total"
    "activate_requests_total"
    "suspend_requests_total" 
    "deprecated_endpoint_used_total"
    "audit_write_success_total"
)

for metric in "${API_METRICS[@]}"; do
    RESULT=$(curl -s "http://localhost:9091/api/v1/query?query=${metric}" | jq -r '.data.result | length')
    if [ "$RESULT" -gt 0 ]; then
        success "指标 $metric 数据可用"
    else
        warning "指标 $metric 暂无数据（需要API服务运行和请求流量）"
    fi
done

echo ""
echo "📈 4. 验证Grafana仪表板..."

# 检查Grafana数据源
DATASOURCES=$(curl -s -u admin:cube-castle-2025 http://localhost:3001/api/datasources | jq -r '. | length')
if [ "$DATASOURCES" -gt 0 ]; then
    success "Grafana已配置 $DATASOURCES 个数据源"
    
    # 列出数据源详情
    curl -s -u admin:cube-castle-2025 http://localhost:3001/api/datasources | jq -r '.[] | "  • \(.name): \(.type) (\(.url))"'
else
    error "Grafana未配置数据源"
fi

# 检查仪表板
DASHBOARDS=$(curl -s -u admin:cube-castle-2025 http://localhost:3001/api/search | jq -r '. | length')
if [ "$DASHBOARDS" -gt 0 ]; then
    success "发现 $DASHBOARDS 个Grafana仪表板"
    
    # 列出仪表板
    curl -s -u admin:cube-castle-2025 http://localhost:3001/api/search | jq -r '.[] | "  • \(.title) (ID: \(.id))"'
else
    warning "未发现Grafana仪表板，可能需要手动导入"
fi

echo ""
echo "🚨 5. 验证告警规则..."

# 检查Prometheus告警规则
RULES_COUNT=$(curl -s http://localhost:9091/api/v1/rules | jq -r '.data.groups | length')
if [ "$RULES_COUNT" -gt 0 ]; then
    success "加载了 $RULES_COUNT 个告警规则组"
    
    # 显示规则组详情
    curl -s http://localhost:9091/api/v1/rules | jq -r '.data.groups[] | "  • \(.name): \(.rules | length) 条规则"'
else
    warning "未发现告警规则，检查规则文件配置"
fi

# 检查当前告警状态
ALERTS=$(curl -s http://localhost:9091/api/v1/alerts | jq -r '.data.alerts | length')
echo "当前活跃告警: $ALERTS 个"

echo ""
echo "🔧 6. 生成测试指标..."

# 如果API服务运行，生成一些测试流量
if curl -s http://localhost:9090/health >/dev/null 2>&1; then
    success "检测到API服务运行，生成测试请求..."
    
    # 发送几个测试请求
    curl -s http://localhost:9090/health >/dev/null
    curl -s http://localhost:9090/metrics >/dev/null
    
    # 尝试触发弃用端点 (应该返回410)
    curl -s -X POST http://localhost:9090/api/v1/organization-units/TEST001/reactivate \
        -H "Content-Type: application/json" \
        -H "X-Client-ID: monitoring-test" \
        -d '{"operationReason":"monitoring test"}' >/dev/null || true
        
    success "已生成测试指标数据"
    
    # 等待一下让指标被采集
    sleep 5
    
    # 再次检查指标
    echo "📊 验证新生成的指标数据..."
    DEPRECATED_METRIC=$(curl -s "http://localhost:9091/api/v1/query?query=deprecated_endpoint_used_total" | jq -r '.data.result | length')
    if [ "$DEPRECATED_METRIC" -gt 0 ]; then
        success "弃用端点访问指标已记录"
    fi
else
    warning "API服务未运行，跳过测试流量生成"
    echo "  启动API服务: cd go-app && go run cmd/server/main.go"
fi

echo ""
echo "📋 7. 监控系统访问信息..."
echo "===========================================" 
echo "🔗 访问地址:"
echo "  • Prometheus:  http://localhost:9091"
echo "  • Grafana:     http://localhost:3001"
echo "    - 用户名: admin"  
echo "    - 密码: cube-castle-2025"
echo "  • AlertManager: http://localhost:9093"
echo ""

echo "📊 重要链接:"
echo "  • Prometheus Targets: http://localhost:9091/targets"
echo "  • Prometheus Rules: http://localhost:9091/rules"
echo "  • Prometheus Alerts: http://localhost:9091/alerts"
echo "  • API Metrics: http://localhost:9090/metrics"
echo ""

echo "🎯 Grafana快速上手:"
echo "  1. 访问 http://localhost:3001"
echo "  2. 登录 admin/cube-castle-2025"
echo "  3. 查看 '组织启停API - SLO监控仪表板'"
echo "  4. 观察API指标和SLO状态"
echo ""

echo "✨ 监控系统验证完成！"

# 最后检查关键服务状态
echo ""
echo "🏁 系统状态总结:"
echo "===========================================" 

# Docker容器状态
PROMETHEUS_STATUS=$(docker ps --filter "name=cube-castle-prometheus" --format "{{.Status}}" 2>/dev/null || echo "未运行")
GRAFANA_STATUS=$(docker ps --filter "name=cube-castle-grafana" --format "{{.Status}}" 2>/dev/null || echo "未运行")

echo "  • Prometheus: $PROMETHEUS_STATUS"
echo "  • Grafana: $GRAFANA_STATUS"

if [[ "$PROMETHEUS_STATUS" == *"Up"* ]] && [[ "$GRAFANA_STATUS" == *"Up"* ]]; then
    success "监控系统运行正常！"
else
    error "部分监控服务可能存在问题"
    echo "  运行以下命令检查详细状态:"
    echo "  docker compose -f monitoring/docker-compose.monitoring.yml ps"
    echo "  docker compose -f monitoring/docker-compose.monitoring.yml logs"
fi