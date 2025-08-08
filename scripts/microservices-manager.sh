#!/bin/bash

# Cube Castle 微服务管理脚本
# 用于统一管理所有微服务的启动、停止、重启和状态检查

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
PROJECT_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
BIN_DIR="$PROJECT_ROOT/bin"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 服务配置
declare -A SERVICES=(
    ["organization-api-gateway"]="8000:cmd/organization-api-gateway:organization-api-gateway"
    ["organization-api-server"]="8080:cmd/organization-api-server:organization-api-server"
    ["organization-graphql-service"]="8090:cmd/organization-graphql-service:organization-graphql-service"
    ["organization-command-server"]="9090:cmd/organization-command-server:organization-command-server"
    ["employee-server"]="8081:cmd/employee-server:employee-server"
    ["position-server"]="8082:cmd/position-server:position-server"
)

# 检查服务是否运行
check_service() {
    local service_name=$1
    local pid_file="$PROJECT_ROOT/cmd/$service_name/$service_name.pid"
    
    if [[ -f "$pid_file" ]]; then
        local pid=$(cat "$pid_file")
        if ps -p "$pid" > /dev/null 2>&1; then
            return 0  # 运行中
        else
            rm -f "$pid_file"  # 清理过期PID文件
            return 1  # 未运行
        fi
    else
        return 1  # 未运行
    fi
}

# 检查端口是否被占用
check_port() {
    local port=$1
    lsof -i ":$port" > /dev/null 2>&1
}

# 启动单个服务
start_service() {
    local service_name=$1
    local service_info=${SERVICES[$service_name]}
    
    if [[ -z "$service_info" ]]; then
        log_error "未知服务: $service_name"
        return 1
    fi
    
    IFS=':' read -r port dir binary <<< "$service_info"
    local service_dir="$PROJECT_ROOT/$dir"
    local binary_path="$BIN_DIR/$binary"
    local log_dir="$service_dir/logs"
    local pid_file="$service_dir/$service_name.pid"
    
    # 检查服务是否已经运行
    if check_service "$service_name"; then
        log_warn "$service_name 已在运行中"
        return 0
    fi
    
    # 检查二进制文件是否存在
    if [[ ! -f "$binary_path" ]]; then
        log_error "二进制文件不存在: $binary_path"
        log_info "请先编译服务: cd $service_dir && go build -o $binary_path ."
        return 1
    fi
    
    # 创建日志目录
    mkdir -p "$log_dir"
    
    # 检查端口是否被其他进程占用
    if check_port "$port"; then
        log_error "端口 $port 已被其他进程占用，无法启动 $service_name"
        return 1
    fi
    
    # 启动服务
    cd "$service_dir"
    nohup "$binary_path" > "$log_dir/$service_name.log" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    # 等待服务启动
    sleep 2
    
    # 验证服务是否成功启动
    if check_service "$service_name"; then
        log_info "$service_name 启动成功 (PID: $pid, PORT: $port)"
        
        # 检查健康状态
        if check_health "$service_name" "$port"; then
            log_info "$service_name 健康检查通过"
        else
            log_warn "$service_name 健康检查失败，请查看日志"
        fi
        return 0
    else
        log_error "$service_name 启动失败，请查看日志: $log_dir/$service_name.log"
        return 1
    fi
}

# 停止单个服务
stop_service() {
    local service_name=$1
    local service_dir="$PROJECT_ROOT/cmd/$service_name"
    local pid_file="$service_dir/$service_name.pid"
    
    if [[ ! -f "$pid_file" ]]; then
        log_warn "$service_name 未运行或PID文件不存在"
        return 0
    fi
    
    local pid=$(cat "$pid_file")
    
    if ps -p "$pid" > /dev/null 2>&1; then
        log_info "正在停止 $service_name (PID: $pid)..."
        kill "$pid"
        
        # 等待进程优雅关闭
        for i in {1..10}; do
            if ! ps -p "$pid" > /dev/null 2>&1; then
                break
            fi
            sleep 1
        done
        
        # 如果进程仍在运行，强制杀死
        if ps -p "$pid" > /dev/null 2>&1; then
            log_warn "强制停止 $service_name..."
            kill -9 "$pid"
        fi
        
        log_info "$service_name 已停止"
    fi
    
    rm -f "$pid_file"
}

# 检查服务健康状态
check_health() {
    local service_name=$1
    local port=$2
    local health_url="http://localhost:$port/health"
    
    # 最多尝试3次
    for i in {1..3}; do
        if curl -s -f "$health_url" > /dev/null 2>&1; then
            return 0
        fi
        sleep 1
    done
    return 1
}

# 显示服务状态
show_status() {
    echo
    log_info "=== Cube Castle 微服务状态 ==="
    printf "%-30s %-8s %-8s %-10s\n" "服务名" "状态" "端口" "健康状态"
    echo "------------------------------------------------------------"
    
    for service_name in "${!SERVICES[@]}"; do
        local service_info=${SERVICES[$service_name]}
        IFS=':' read -r port dir binary <<< "$service_info"
        
        local status="❌ 停止"
        local health="N/A"
        
        if check_service "$service_name"; then
            status="✅ 运行"
            if check_health "$service_name" "$port"; then
                health="✅ 健康"
            else
                health="❌ 异常"
            fi
        fi
        
        printf "%-30s %-8s %-8s %-10s\n" "$service_name" "$status" "$port" "$health"
    done
    echo
}

# 启动所有服务
start_all() {
    log_info "启动所有微服务..."
    
    # 按依赖顺序启动服务
    local services_order=(
        "organization-command-server"
        "organization-api-server" 
        "organization-graphql-service"
        "organization-api-gateway"
        "employee-server"
        "position-server"
    )
    
    local failed_services=()
    
    for service_name in "${services_order[@]}"; do
        if ! start_service "$service_name"; then
            failed_services+=("$service_name")
        fi
        sleep 1  # 服务间启动间隔
    done
    
    if [[ ${#failed_services[@]} -eq 0 ]]; then
        log_info "所有服务启动成功！"
    else
        log_error "以下服务启动失败: ${failed_services[*]}"
        return 1
    fi
}

# 停止所有服务
stop_all() {
    log_info "停止所有微服务..."
    
    # 逆序停止服务
    local services_order=(
        "position-server"
        "employee-server"
        "organization-api-gateway"
        "organization-graphql-service"
        "organization-api-server"
        "organization-command-server"
    )
    
    for service_name in "${services_order[@]}"; do
        stop_service "$service_name"
    done
    
    log_info "所有服务已停止"
}

# 重启所有服务
restart_all() {
    log_info "重启所有微服务..."
    stop_all
    sleep 2
    start_all
}

# 清理过期PID文件
cleanup() {
    log_info "清理过期PID文件..."
    find "$PROJECT_ROOT/cmd" -name "*.pid" -exec rm -f {} \;
    log_info "清理完成"
}

# 编译所有服务
build_all() {
    log_info "编译所有微服务..."
    
    local failed_builds=()
    
    for service_name in "${!SERVICES[@]}"; do
        local service_info=${SERVICES[$service_name]}
        IFS=':' read -r port dir binary <<< "$service_info"
        local service_dir="$PROJECT_ROOT/$dir"
        local binary_path="$BIN_DIR/$binary"
        
        log_info "编译 $service_name..."
        
        if cd "$service_dir" && go build -o "$binary_path" . ; then
            chmod +x "$binary_path"
            log_info "$service_name 编译成功"
        else
            log_error "$service_name 编译失败"
            failed_builds+=("$service_name")
        fi
    done
    
    if [[ ${#failed_builds[@]} -eq 0 ]]; then
        log_info "所有服务编译成功！"
    else
        log_error "以下服务编译失败: ${failed_builds[*]}"
        return 1
    fi
}

# 显示帮助信息
show_help() {
    echo "Cube Castle 微服务管理脚本"
    echo
    echo "用法: $0 [命令] [服务名]"
    echo
    echo "命令:"
    echo "  start [service]     启动指定服务或所有服务"
    echo "  stop [service]      停止指定服务或所有服务"
    echo "  restart [service]   重启指定服务或所有服务"
    echo "  status              显示所有服务状态"
    echo "  build               编译所有服务"
    echo "  cleanup             清理过期PID文件"
    echo "  help                显示此帮助信息"
    echo
    echo "可用服务:"
    for service_name in "${!SERVICES[@]}"; do
        local service_info=${SERVICES[$service_name]}
        IFS=':' read -r port dir binary <<< "$service_info"
        echo "  $service_name (端口: $port)"
    done
}

# 主函数
main() {
    case "${1:-}" in
        start)
            if [[ -n "${2:-}" ]]; then
                start_service "$2"
            else
                start_all
            fi
            ;;
        stop)
            if [[ -n "${2:-}" ]]; then
                stop_service "$2"
            else
                stop_all
            fi
            ;;
        restart)
            if [[ -n "${2:-}" ]]; then
                stop_service "$2"
                sleep 1
                start_service "$2"
            else
                restart_all
            fi
            ;;
        status)
            show_status
            ;;
        build)
            build_all
            ;;
        cleanup)
            cleanup
            ;;
        help|--help|-h)
            show_help
            ;;
        "")
            show_status
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"