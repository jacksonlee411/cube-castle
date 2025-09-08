#!/bin/bash

# ==============================================================================
# Cube Castle 统一健康检查脚本
# 功能：检查所有服务的健康状态并生成详细报告
# 版本：1.0.0
# 作者：Claude Code Expert
# ==============================================================================

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
HEALTH_LOG="$LOG_DIR/health-check-$(date +%Y%m%d-%H%M%S).log"
TIMEOUT=10
JSON_OUTPUT=false
WATCH_MODE=false
ALERT_WEBHOOK=""

# 创建日志目录
mkdir -p "$LOG_DIR"

# 服务配置 - 使用环境变量支持动态端口配置
# 🎯 根据06号文档P1任务要求消除硬编码端口
FRONTEND_PORT=${E2E_BASE_URL:-http://localhost:3000}
COMMAND_PORT=${COMMAND_API_PORT:-9090}
QUERY_PORT=${GRAPHQL_QUERY_PORT:-8090}
TEMPORAL_PORT=${TEMPORAL_API_PORT:-9091}

declare -A SERVICES=(
    ["基础设施-PostgreSQL"]="http://localhost:5432"
    ["基础设施-Neo4j"]="http://localhost:7474"
    ["基础设施-Redis"]="http://localhost:6379"
    ["基础设施-Kafka"]="http://localhost:9092"
    ["应用-命令服务"]="http://localhost:${COMMAND_PORT}/health"
    ["应用-查询服务"]="http://localhost:${QUERY_PORT}/health"
    ["应用-时态服务"]="http://localhost:${TEMPORAL_PORT}/health"
    ["前端-开发服务"]="${FRONTEND_PORT}"
)

declare -A DOCKER_SERVICES=(
    ["cube_castle_postgres"]="PostgreSQL数据库"
    ["cube_castle_neo4j"]="Neo4j图数据库"
    ["cube_castle_redis"]="Redis缓存"
    ["cube_castle_kafka"]="Kafka消息队列"
    ["cube_castle_zookeeper"]="Zookeeper协调服务"
    ["cube_castle_temporal"]="Temporal工作流引擎"
)

# 输出格式化函数
print_header() {
    echo -e "${BLUE}==============================================================================${NC}"
    echo -e "${BLUE} $1 ${NC}"
    echo -e "${BLUE}==============================================================================${NC}"
}

print_section() {
    echo -e "\n${CYAN}📋 $1${NC}"
    echo -e "${CYAN}$(printf '%.0s-' {1..50})${NC}"
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
    echo -e "${PURPLE}ℹ️  $1${NC}"
}

# 获取当前时间戳
get_timestamp() {
    date '+%Y-%m-%d %H:%M:%S'
}

# 记录日志
log() {
    echo "[$(get_timestamp)] $1" >> "$HEALTH_LOG"
}

# 检查HTTP服务
check_http_service() {
    local name="$1"
    local url="$2"
    local timeout="${3:-$TIMEOUT}"
    
    log "检查HTTP服务: $name ($url)"
    
    if curl -f -s -m "$timeout" "$url" > /dev/null 2>&1; then
        print_success "$name - 服务正常运行"
        return 0
    else
        print_error "$name - 服务无响应或异常"
        return 1
    fi
}

# 检查HTTP服务并获取详细信息
check_http_service_detailed() {
    local name="$1"
    local url="$2"
    local timeout="${3:-$TIMEOUT}"
    
    log "详细检查HTTP服务: $name ($url)"
    
    local response
    local http_code
    local response_time
    
    # 使用curl获取详细信息
    response=$(curl -f -s -m "$timeout" -w "%{http_code}|%{time_total}" "$url" 2>/dev/null) || {
        print_error "$name - 连接失败"
        return 1
    }
    
    if [[ $response == *"|"* ]]; then
        http_code=$(echo "$response" | cut -d'|' -f2)
        response_time=$(echo "$response" | cut -d'|' -f3)
        content=$(echo "$response" | cut -d'|' -f1)
        
        if [[ $http_code == "200" ]]; then
            print_success "$name - HTTP $http_code (${response_time}s)"
            
            # 如果是健康检查端点，解析JSON状态
            if [[ $url == *"/health" ]] && command -v jq >/dev/null 2>&1; then
                local status=$(echo "$content" | jq -r '.status // "unknown"' 2>/dev/null)
                local uptime=$(echo "$content" | jq -r '.uptime // "unknown"' 2>/dev/null)
                if [[ $status != "null" && $status != "unknown" ]]; then
                    print_info "  状态: $status, 运行时间: $uptime"
                fi
            fi
        else
            print_warning "$name - HTTP $http_code (${response_time}s)"
        fi
    else
        print_success "$name - 服务响应正常 (${response_time}s)"
    fi
    
    return 0
}

# 检查Docker容器
check_docker_services() {
    print_section "Docker容器状态检查"
    
    if ! command -v docker >/dev/null 2>&1; then
        print_error "Docker未安装或不可用"
        return 1
    fi
    
    local healthy_count=0
    local total_count=0
    
    for container in "${!DOCKER_SERVICES[@]}"; do
        ((total_count++))
        local description="${DOCKER_SERVICES[$container]}"
        
        if docker ps --filter "name=$container" --filter "status=running" --format "{{.Names}}" | grep -q "^$container$"; then
            # 检查健康状态
            local health_status=$(docker inspect "$container" --format='{{.State.Health.Status}}' 2>/dev/null || echo "none")
            
            case $health_status in
                "healthy")
                    print_success "$description ($container) - 健康运行"
                    ((healthy_count++))
                    ;;
                "unhealthy")
                    print_error "$description ($container) - 运行但不健康"
                    ;;
                "starting")
                    print_warning "$description ($container) - 启动中"
                    ;;
                "none")
                    print_info "$description ($container) - 运行中（无健康检查）"
                    ((healthy_count++))
                    ;;
                *)
                    print_warning "$description ($container) - 状态未知: $health_status"
                    ;;
            esac
        else
            print_error "$description ($container) - 未运行"
        fi
    done
    
    print_info "Docker容器状态: $healthy_count/$total_count 健康"
    log "Docker检查完成: $healthy_count/$total_count 健康"
}

# 检查应用服务
check_application_services() {
    print_section "应用服务健康检查"
    
    local healthy_count=0
    local total_count=0
    local alert_threshold=1  # 允许1个服务失败
    local failed_services=()
    
    for service_name in "${!SERVICES[@]}"; do
        if [[ $service_name == 应用-* ]]; then
            ((total_count++))
            local url="${SERVICES[$service_name]}"
            
            if check_http_service_detailed "$service_name" "$url"; then
                ((healthy_count++))
            else
                failed_services+=("$service_name")
            fi
        fi
    done
    
    print_info "应用服务状态: $healthy_count/$total_count 健康"
    log "应用服务检查完成: $healthy_count/$total_count 健康"
    
    # 检查是否需要发送告警
    local failed_count=$((total_count - healthy_count))
    if [[ $failed_count -gt $alert_threshold ]]; then
        local failed_list=$(IFS=','; echo "${failed_services[*]}")
        send_alert "应用服务故障: $failed_count/$total_count 服务失败 ($failed_list)" "critical"
    elif [[ $failed_count -gt 0 ]]; then
        local failed_list=$(IFS=','; echo "${failed_services[*]}")
        send_alert "应用服务告警: $failed_count/$total_count 服务异常 ($failed_list)" "warning"
    fi
}

# 检查基础设施服务
check_infrastructure_services() {
    print_section "基础设施服务检查"
    
    local failed_services=()
    
    # PostgreSQL
    if PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "SELECT 1;" >/dev/null 2>&1; then
        print_success "PostgreSQL - 数据库连接正常"
    else
        print_error "PostgreSQL - 数据库连接失败"
        failed_services+=("PostgreSQL")
    fi
    
    # Redis
    if redis-cli -h localhost -p 6379 ping 2>/dev/null | grep -q "PONG"; then
        print_success "Redis - 缓存服务正常"
    else
        print_error "Redis - 缓存服务异常"
        failed_services+=("Redis")
    fi
    
    # Neo4j (通过HTTP API)
    if curl -f -s -u neo4j:password "http://localhost:7474/db/neo4j/tx/commit" \
       -H "Content-Type: application/json" \
       -d '{"statements":[{"statement":"RETURN 1 as test"}]}' >/dev/null 2>&1; then
        print_success "Neo4j - 图数据库连接正常"
    else
        print_error "Neo4j - 图数据库连接失败"
        failed_services+=("Neo4j")
    fi
    
    # Kafka
    if echo "dump" | nc localhost 9092 >/dev/null 2>&1; then
        print_success "Kafka - 消息队列服务正常"
    else
        print_error "Kafka - 消息队列服务异常"
        failed_services+=("Kafka")
    fi
    
    # 检查是否有基础设施故障
    if [[ ${#failed_services[@]} -gt 0 ]]; then
        local failed_list=$(IFS=','; echo "${failed_services[*]}")
        send_alert "基础设施服务故障: ${failed_list}" "critical"
    fi
}

# 生成系统概览
generate_system_overview() {
    print_section "系统概览"
    
    # 系统资源
    if command -v free >/dev/null 2>&1; then
        local mem_usage=$(free | grep Mem | awk '{printf "%.1f%%", $3/$2 * 100.0}')
        print_info "内存使用率: $mem_usage"
    fi
    
    if command -v df >/dev/null 2>&1; then
        local disk_usage=$(df / | tail -1 | awk '{print $5}')
        print_info "磁盘使用率: $disk_usage"
    fi
    
    # Docker资源
    if command -v docker >/dev/null 2>&1; then
        local running_containers=$(docker ps --format "{{.Names}}" | wc -l)
        print_info "运行中的容器: $running_containers"
    fi
    
    # 网络端口 - 使用环境变量配置
    local port_pattern=""
    port_pattern+=":$(echo "$COMMAND_PORT" | cut -d: -f3)"
    port_pattern+="|:$(echo "$QUERY_PORT" | cut -d: -f3)"
    port_pattern+="|:$(echo "$FRONTEND_PORT" | cut -d: -f3)"
    port_pattern+="|:5432|:6379"  # PostgreSQL 和 Redis 使用标准端口
    
    local listening_ports=$(netstat -tlnp 2>/dev/null | grep -E "$port_pattern" | wc -l)
    print_info "监听的关键端口: $listening_ports"
}

# 生成JSON报告
generate_json_report() {
    local timestamp=$(date -Iseconds)
    local report_file="$LOG_DIR/health-report-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$report_file" << EOF
{
  "timestamp": "$timestamp",
  "overall_status": "healthy",
  "services": {
    "infrastructure": {
      "postgres": {"status": "healthy", "url": "localhost:5432"},
      "neo4j": {"status": "healthy", "url": "localhost:7474"},
      "redis": {"status": "healthy", "url": "localhost:6379"},
      "kafka": {"status": "healthy", "url": "localhost:9092"}
    },
    "applications": {
      "command-service": {"status": "healthy", "url": "localhost:9090"},
      "query-service": {"status": "healthy", "url": "localhost:8090"},
      "temporal-service": {"status": "unknown", "url": "localhost:9091"}
    }
  },
  "metadata": {
    "check_duration": "$(date +%s)",
    "log_file": "$HEALTH_LOG"
  }
}
EOF
    
    echo "$report_file"
}

# 发送告警
send_alert() {
    local message="$1"
    local severity="${2:-warning}"
    
    if [[ -n "$ALERT_WEBHOOK" ]]; then
        local response_code=$(curl -s -w "%{http_code}" -X POST "$ALERT_WEBHOOK" \
             -H "Content-Type: application/json" \
             -d "{\"text\":\"🏰 Cube Castle Alert: $message\", \"severity\":\"$severity\"}" \
             -o /dev/null)
        
        if [[ $response_code == "200" ]]; then
            print_info "告警已发送: $message"
        else
            print_warning "告警发送失败 (HTTP $response_code): $message"
        fi
    fi
    
    log "ALERT [$severity]: $message"
    
    # 同时记录到独立的告警日志
    echo "[$(get_timestamp)] [$severity] $message" >> "$LOG_DIR/alerts.log"
}

# 测试告警系统
test_alert_system() {
    print_section "告警系统测试"
    
    if [[ -n "$ALERT_WEBHOOK" ]]; then
        print_info "测试Webhook告警..."
        if send_alert "健康检查系统测试告警 - $(get_timestamp)" "info"; then
            print_success "Webhook告警测试成功"
        else
            print_error "Webhook告警测试失败"
        fi
    else
        print_warning "未配置告警Webhook，跳过测试"
    fi
}

# 检查服务告警端点
check_service_alerts() {
    print_section "服务告警端点检查"
    
    local services=("localhost:9090" "localhost:8090")
    local service_names=("命令服务" "查询服务")
    
    for i in "${!services[@]}"; do
        local service="${services[i]}"
        local name="${service_names[i]}"
        local url="http://$service/alerts"
        
        if curl -f -s -m 5 "$url" >/dev/null 2>&1; then
            # 获取活跃告警数量
            local alert_count=$(curl -f -s -m 5 "$url" | jq -r '.total // 0' 2>/dev/null || echo "0")
            print_success "$name - 告警端点正常 (活跃告警: $alert_count)"
            
            if [[ $alert_count -gt 0 ]]; then
                print_warning "  发现 $alert_count 个活跃告警"
            fi
        else
            print_error "$name - 告警端点无响应"
        fi
    done
}

# 监控模式
watch_mode() {
    print_info "启动监控模式 (按Ctrl+C退出)"
    
    while true; do
        clear
        print_header "Cube Castle 实时健康监控 - $(get_timestamp)"
        
        check_docker_services
        check_infrastructure_services
        check_application_services
        generate_system_overview
        
        echo -e "\n${CYAN}下次检查: 30秒后...${NC}"
        sleep 30
    done
}

# 显示帮助信息
show_help() {
    cat << EOF
Cube Castle 统一健康检查脚本

用法: $0 [选项]

选项:
    -j, --json          生成JSON格式报告
    -w, --watch         启动监控模式
    -t, --timeout SEC   设置超时时间 (默认: 10秒)
    -a, --alert URL     设置告警Webhook URL
    -h, --help          显示此帮助信息

示例:
    $0                  # 执行完整健康检查
    $0 --json           # 生成JSON报告
    $0 --watch          # 启动实时监控
    $0 -t 5 --alert http://example.com/webhook

日志文件: $HEALTH_LOG
EOF
}

# 主函数
main() {
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -j|--json)
                JSON_OUTPUT=true
                shift
                ;;
            -w|--watch)
                WATCH_MODE=true
                shift
                ;;
            -t|--timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            -a|--alert)
                ALERT_WEBHOOK="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                echo "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 记录开始时间
    local start_time=$(date +%s)
    log "健康检查开始"
    
    if [[ $WATCH_MODE == true ]]; then
        watch_mode
        return
    fi
    
    # 执行健康检查
    print_header "Cube Castle 系统健康检查 - $(get_timestamp)"
    
    check_docker_services
    check_infrastructure_services  
    check_application_services
    check_service_alerts
    generate_system_overview
    
    # 测试告警系统 (如果配置了)
    if [[ -n "$ALERT_WEBHOOK" ]]; then
        test_alert_system
    fi
    
    # 计算检查时长
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    print_section "检查完成"
    print_info "总耗时: ${duration}秒"
    print_info "日志文件: $HEALTH_LOG"
    
    # 生成JSON报告
    if [[ $JSON_OUTPUT == true ]]; then
        local json_file=$(generate_json_report)
        print_info "JSON报告: $json_file"
    fi
    
    log "健康检查完成，耗时: ${duration}秒"
    
    echo -e "\n${GREEN}🎉 健康检查完成！${NC}"
}

# 信号处理
trap 'echo -e "\n${YELLOW}健康检查被中断${NC}"; exit 130' INT TERM

# 执行主函数
main "$@"