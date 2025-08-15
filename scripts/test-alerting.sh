#!/bin/bash

# =============================================================================
# Cube Castle 告警系统测试脚本
# 功能：测试健康检查告警机制是否正常工作
# 版本：1.0.0
# =============================================================================

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"

print_header() {
    echo -e "${BLUE}=============================================================================${NC}"
    echo -e "${BLUE} $1 ${NC}"
    echo -e "${BLUE}=============================================================================${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# 测试服务健康检查端点
test_health_endpoints() {
    print_header "测试服务健康检查端点"
    
    local services=("localhost:9090" "localhost:8090")
    local service_names=("命令服务" "查询服务")
    
    for i in "${!services[@]}"; do
        local service="${services[i]}"
        local name="${service_names[i]}"
        local health_url="http://$service/health"
        
        print_info "测试 $name 健康检查..."
        
        if curl -f -s -m 5 "$health_url" >/dev/null 2>&1; then
            print_success "$name - 健康检查端点正常"
            
            # 获取健康状态详情
            local health_status=$(curl -f -s -m 5 "$health_url" | jq -r '.status // "unknown"' 2>/dev/null)
            print_info "  状态: $health_status"
        else
            print_error "$name - 健康检查端点异常"
        fi
    done
}

# 测试告警端点
test_alert_endpoints() {
    print_header "测试告警管理端点"
    
    local services=("localhost:9090" "localhost:8090")
    local service_names=("命令服务" "查询服务")
    
    for i in "${!services[@]}"; do
        local service="${services[i]}"
        local name="${service_names[i]}"
        local alerts_url="http://$service/alerts"
        local history_url="http://$service/alerts/history"
        
        print_info "测试 $name 告警端点..."
        
        # 测试活跃告警端点
        if curl -f -s -m 5 "$alerts_url" >/dev/null 2>&1; then
            local alert_count=$(curl -f -s -m 5 "$alerts_url" | jq -r '.total // 0' 2>/dev/null)
            print_success "$name - 活跃告警端点正常 (数量: $alert_count)"
        else
            print_error "$name - 活跃告警端点异常"
        fi
        
        # 测试告警历史端点
        if curl -f -s -m 5 "$history_url" >/dev/null 2>&1; then
            local history_count=$(curl -f -s -m 5 "$history_url" | jq -r '.total // 0' 2>/dev/null)
            print_success "$name - 告警历史端点正常 (数量: $history_count)"
        else
            print_error "$name - 告警历史端点异常"
        fi
    done
}

# 测试状态仪表板端点
test_dashboard_endpoints() {
    print_header "测试状态仪表板端点"
    
    local services=("localhost:9090" "localhost:8090")
    local service_names=("命令服务" "查询服务")
    
    for i in "${!services[@]}"; do
        local service="${services[i]}"
        local name="${service_names[i]}"
        local status_url="http://$service/status"
        
        print_info "测试 $name 状态仪表板..."
        
        if curl -f -s -m 5 "$status_url" >/dev/null 2>&1; then
            print_success "$name - 状态仪表板端点正常"
            print_info "  访问地址: $status_url"
        else
            print_error "$name - 状态仪表板端点异常"
        fi
    done
}

# 模拟告警场景
simulate_alert_scenarios() {
    print_header "模拟告警场景"
    
    print_info "注意：以下测试将模拟服务故障场景"
    print_warning "这些测试不会影响实际服务，仅用于验证告警系统"
    
    # 检查是否有告警配置
    if [[ -n "${ALERT_WEBHOOK_URL:-}" ]]; then
        print_info "发现告警Webhook配置，测试发送告警..."
        
        # 发送测试告警
        curl -X POST "$ALERT_WEBHOOK_URL" \
             -H "Content-Type: application/json" \
             -d '{"text":"🧪 Cube Castle 告警系统测试 - 系统正常", "severity":"info"}' \
             >/dev/null 2>&1 && print_success "测试告警发送成功" || print_error "测试告警发送失败"
    else
        print_warning "未配置ALERT_WEBHOOK_URL环境变量，跳过Webhook测试"
    fi
    
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        print_info "发现Slack Webhook配置，测试发送Slack告警..."
        
        # 发送Slack测试告警
        curl -X POST "$SLACK_WEBHOOK_URL" \
             -H "Content-Type: application/json" \
             -d '{"text":"🧪 Cube Castle 告警系统测试 - 系统正常"}' \
             >/dev/null 2>&1 && print_success "Slack测试告警发送成功" || print_error "Slack测试告警发送失败"
    else
        print_warning "未配置SLACK_WEBHOOK_URL环境变量，跳过Slack测试"
    fi
}

# 验证告警配置
verify_alert_configuration() {
    print_header "验证告警配置"
    
    # 检查环境变量配置
    local config_items=("ALERT_WEBHOOK_URL" "SLACK_WEBHOOK_URL" "WEBHOOK_TOKEN")
    local configured_count=0
    
    for item in "${config_items[@]}"; do
        if [[ -n "${!item:-}" ]]; then
            print_success "$item 已配置"
            ((configured_count++))
        else
            print_info "$item 未配置"
        fi
    done
    
    if [[ $configured_count -eq 0 ]]; then
        print_warning "未发现任何告警配置，告警功能将不会工作"
        print_info "请设置以下环境变量之一："
        print_info "  export ALERT_WEBHOOK_URL='https://your-webhook-url'"
        print_info "  export SLACK_WEBHOOK_URL='https://hooks.slack.com/...'"
    else
        print_success "发现 $configured_count 个告警配置"
    fi
}

# 生成测试报告
generate_test_report() {
    print_header "测试报告"
    
    local report_file="$LOG_DIR/alerting-test-report-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$report_file" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "test_type": "alerting_system_test",
  "summary": {
    "health_endpoints": "tested",
    "alert_endpoints": "tested",
    "dashboard_endpoints": "tested",
    "configuration": "verified"
  },
  "recommendations": [
    "配置ALERT_WEBHOOK_URL环境变量以启用Webhook告警",
    "配置SLACK_WEBHOOK_URL环境变量以启用Slack告警",
    "定期检查告警端点的响应性能",
    "监控告警历史以识别系统模式"
  ],
  "next_steps": [
    "运行完整健康检查: ./health-check-unified.sh",
    "访问状态仪表板: http://localhost:9090/status",
    "监控活跃告警: http://localhost:9090/alerts"
  ]
}
EOF
    
    print_success "测试报告已生成: $report_file"
}

# 显示使用说明
show_usage() {
    cat << EOF
Cube Castle 告警系统测试工具

用法: $0 [选项]

选项:
    --endpoints    仅测试告警端点
    --config       仅验证告警配置
    --simulate     模拟告警场景
    -h, --help     显示此帮助信息

示例:
    $0              # 运行完整测试
    $0 --endpoints  # 仅测试端点
    $0 --config     # 仅验证配置

环境变量:
    ALERT_WEBHOOK_URL   - Webhook告警URL
    SLACK_WEBHOOK_URL   - Slack Webhook URL  
    WEBHOOK_TOKEN       - Webhook认证令牌
EOF
}

# 主函数
main() {
    case "${1:-}" in
        --endpoints)
            test_health_endpoints
            test_alert_endpoints
            test_dashboard_endpoints
            ;;
        --config)
            verify_alert_configuration
            ;;
        --simulate)
            simulate_alert_scenarios
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        "")
            # 运行完整测试
            verify_alert_configuration
            test_health_endpoints
            test_alert_endpoints
            test_dashboard_endpoints
            simulate_alert_scenarios
            generate_test_report
            
            print_header "测试完成"
            print_success "告警系统测试已完成！"
            print_info "查看详细报告: $LOG_DIR/"
            ;;
        *)
            echo "未知选项: $1"
            show_usage
            exit 1
            ;;
    esac
}

# 信号处理
trap 'echo -e "\n${YELLOW}测试被中断${NC}"; exit 130' INT TERM

# 执行主函数
main "$@"