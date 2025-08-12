#!/bin/bash

# =============================================================================
# Cube Castle 时态管理服务部署脚本
# 支持开发环境和生产环境部署
# =============================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置默认值
DEFAULT_ENV="development"
DEFAULT_PORT_REST="9093"
DEFAULT_PORT_QUERY="8091"
DEFAULT_PORT_SYNC="8092"

# 帮助信息
show_help() {
    echo "Cube Castle 时态管理服务部署脚本"
    echo ""
    echo "用法: $0 [选项] [命令]"
    echo ""
    echo "命令:"
    echo "  start       启动所有时态服务"
    echo "  stop        停止所有时态服务"  
    echo "  restart     重启所有时态服务"
    echo "  status      检查服务状态"
    echo "  build       构建服务"
    echo "  test        执行健康检查测试"
    echo "  logs        查看服务日志"
    echo ""
    echo "选项:"
    echo "  -e, --env ENV        环境 (development|production) [默认: $DEFAULT_ENV]"
    echo "  -h, --help           显示此帮助信息"
    echo "  --rest-port PORT     REST服务端口 [默认: $DEFAULT_PORT_REST]"
    echo "  --query-port PORT    GraphQL服务端口 [默认: $DEFAULT_PORT_QUERY]" 
    echo "  --sync-port PORT     CDC服务端口 [默认: $DEFAULT_PORT_SYNC]"
    echo ""
    echo "示例:"
    echo "  $0 start                    # 启动所有服务 (开发环境)"
    echo "  $0 -e production start      # 生产环境启动"
    echo "  $0 status                   # 检查服务状态"
    echo "  $0 test                     # 执行健康检查"
}

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查系统依赖..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装或不在PATH中"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装或不在PATH中"
        exit 1
    fi
    
    # 检查Go (开发环境需要)
    if [[ "$ENV" == "development" ]] && ! command -v go &> /dev/null; then
        log_warning "Go未安装，将仅支持Docker运行模式"
    fi
    
    log_success "依赖检查完成"
}

# 检查基础设施服务
check_infrastructure() {
    log_info "检查基础设施服务状态..."
    
    local services=("postgres" "neo4j" "redis" "kafka")
    local all_healthy=true
    
    for service in "${services[@]}"; do
        if ! docker-compose ps | grep -q "${service}.*Up.*healthy"; then
            log_error "基础设施服务 $service 未运行或不健康"
            all_healthy=false
        fi
    done
    
    if [[ "$all_healthy" == "false" ]]; then
        log_error "请先启动基础设施服务: docker-compose up -d"
        exit 1
    fi
    
    log_success "基础设施服务检查通过"
}

# 构建服务
build_services() {
    log_info "构建时态管理服务..."
    
    local services=("rest" "query" "sync")
    
    for service in "${services[@]}"; do
        log_info "构建 temporal-$service 服务..."
        
        local service_dir="cmd/organization-temporal-${service}-service"
        if [[ "$service" == "rest" ]]; then
            service_dir="cmd/organization-temporal-rest-service"
        elif [[ "$service" == "query" ]]; then
            service_dir="cmd/organization-temporal-query-service"
        elif [[ "$service" == "sync" ]]; then
            service_dir="cmd/organization-temporal-sync-service"
        fi
        
        if [[ -d "$service_dir" ]]; then
            cd "$service_dir"
            if [[ "$ENV" == "development" ]]; then
                # 开发环境直接构建可执行文件
                go build -o "temporal-$service" main.go
                log_success "构建 temporal-$service 完成"
            else
                # 生产环境构建Docker镜像
                docker build -t "cube-castle-temporal-$service:latest" . || {
                    log_error "构建 temporal-$service Docker镜像失败"
                    return 1
                }
                log_success "构建 temporal-$service Docker镜像完成"
            fi
            cd - > /dev/null
        else
            log_error "服务目录不存在: $service_dir"
            return 1
        fi
    done
    
    log_success "所有服务构建完成"
}

# 启动服务
start_services() {
    log_info "启动时态管理服务..."
    
    # 设置环境变量
    export REST_PORT="$PORT_REST"
    export QUERY_PORT="$PORT_QUERY"
    export SYNC_PORT="$PORT_SYNC"
    
    if [[ "$ENV" == "development" ]]; then
        # 开发环境使用Go直接运行
        start_development_services
    else
        # 生产环境使用Docker Compose
        start_production_services
    fi
    
    log_success "所有服务启动完成"
}

# 开发环境启动
start_development_services() {
    log_info "以开发模式启动服务..."
    
    # 启动REST服务
    cd cmd/organization-temporal-rest-service
    PORT="$PORT_REST" nohup ./temporal-rest > "../../logs/temporal-rest.log" 2>&1 &
    echo $! > "../../logs/temporal-rest.pid"
    cd - > /dev/null
    
    # 启动GraphQL服务
    cd cmd/organization-temporal-query-service  
    PORT="$PORT_QUERY" nohup ./temporal-query > "../../logs/temporal-query.log" 2>&1 &
    echo $! > "../../logs/temporal-query.pid"
    cd - > /dev/null
    
    # 启动CDC服务
    cd cmd/organization-temporal-sync-service
    KAFKA_TOPIC="organization_db.public.organization_units" PORT="$PORT_SYNC" nohup ./temporal-sync > "../../logs/temporal-sync.log" 2>&1 &
    echo $! > "../../logs/temporal-sync.pid"
    cd - > /dev/null
    
    sleep 3
    log_success "开发环境服务启动完成"
}

# 生产环境启动
start_production_services() {
    log_info "以生产模式启动服务..."
    
    # 使用Docker Compose启动时态服务
    docker-compose -f docker-compose.temporal.yml up -d
    
    log_success "生产环境服务启动完成"
}

# 停止服务
stop_services() {
    log_info "停止时态管理服务..."
    
    if [[ "$ENV" == "development" ]]; then
        # 开发环境：通过PID文件停止进程
        for service in rest query sync; do
            local pid_file="logs/temporal-$service.pid"
            if [[ -f "$pid_file" ]]; then
                local pid=$(cat "$pid_file")
                if kill "$pid" 2>/dev/null; then
                    log_success "停止 temporal-$service 服务 (PID: $pid)"
                    rm "$pid_file"
                else
                    log_warning "temporal-$service 进程 (PID: $pid) 可能已停止"
                    rm -f "$pid_file"
                fi
            fi
        done
        
        # 额外确认：通过进程名停止
        pkill -f "temporal-rest" 2>/dev/null || true
        pkill -f "temporal-query" 2>/dev/null || true
        pkill -f "temporal-sync" 2>/dev/null || true
    else
        # 生产环境：使用Docker Compose停止
        docker-compose -f docker-compose.temporal.yml down
    fi
    
    log_success "所有服务已停止"
}

# 检查服务状态
check_status() {
    log_info "检查时态管理服务状态..."
    
    local services=(
        "REST服务:$PORT_REST:/health"
        "GraphQL服务:$PORT_QUERY:/health" 
        "CDC服务:$PORT_SYNC:/health"
    )
    
    for service_info in "${services[@]}"; do
        IFS=':' read -r name port path <<< "$service_info"
        
        if curl -s "http://localhost:$port$path" > /dev/null 2>&1; then
            log_success "$name (端口 $port): 健康"
        else
            log_error "$name (端口 $port): 不可用"
        fi
    done
}

# 执行健康检查测试
run_tests() {
    log_info "执行时态管理服务健康检查测试..."
    
    # 等待服务启动
    sleep 5
    
    # 测试REST服务
    log_info "测试REST时态命令服务..."
    if response=$(curl -s "http://localhost:$PORT_REST/health"); then
        log_success "REST服务健康检查: ✓"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        log_error "REST服务健康检查: ✗"
    fi
    
    # 测试GraphQL服务  
    log_info "测试GraphQL时态查询服务..."
    if response=$(curl -s "http://localhost:$PORT_QUERY/health"); then
        log_success "GraphQL服务健康检查: ✓"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        log_error "GraphQL服务健康检查: ✗"
    fi
    
    # 测试CDC服务
    log_info "测试CDC时态同步服务..."
    if response=$(curl -s "http://localhost:$PORT_SYNC/health"); then
        log_success "CDC服务健康检查: ✓"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
    else
        log_error "CDC服务健康检查: ✗"
    fi
    
    # 综合功能测试
    log_info "执行时态功能端到端测试..."
    if curl -s "http://localhost:$PORT_REST/api/v1/organization-units/1000001/temporal" > /dev/null; then
        log_success "时态查询功能测试: ✓"
    else
        log_warning "时态查询功能测试: ⚠️ (可能需要数据初始化)"
    fi
}

# 查看日志
show_logs() {
    local service="${1:-all}"
    
    if [[ "$ENV" == "development" ]]; then
        case "$service" in
            "rest")
                tail -f logs/temporal-rest.log
                ;;
            "query")
                tail -f logs/temporal-query.log
                ;;
            "sync")
                tail -f logs/temporal-sync.log
                ;;
            "all"|*)
                log_info "显示所有服务日志 (Ctrl+C 退出)..."
                tail -f logs/temporal-*.log
                ;;
        esac
    else
        docker-compose -f docker-compose.temporal.yml logs -f "$service"
    fi
}

# 初始化目录结构
init_directories() {
    # 创建日志目录
    mkdir -p logs
    
    # 创建配置目录
    mkdir -p config
}

# 参数解析
ENV="$DEFAULT_ENV"
PORT_REST="$DEFAULT_PORT_REST"
PORT_QUERY="$DEFAULT_PORT_QUERY"
PORT_SYNC="$DEFAULT_PORT_SYNC"

while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            ENV="$2"
            shift 2
            ;;
        --rest-port)
            PORT_REST="$2"
            shift 2
            ;;
        --query-port)
            PORT_QUERY="$2"
            shift 2
            ;;
        --sync-port)
            PORT_SYNC="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        start|stop|restart|status|build|test|logs)
            COMMAND="$1"
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 主程序逻辑
main() {
    # 显示配置信息
    log_info "=== Cube Castle 时态管理服务部署 ==="
    log_info "环境: $ENV"
    log_info "端口配置: REST=$PORT_REST, GraphQL=$PORT_QUERY, CDC=$PORT_SYNC"
    echo ""
    
    # 初始化
    init_directories
    
    # 执行命令
    case "${COMMAND:-start}" in
        "build")
            check_dependencies
            build_services
            ;;
        "start")
            check_dependencies
            check_infrastructure
            build_services
            start_services
            check_status
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            stop_services
            sleep 2
            start_services
            check_status
            ;;
        "status")
            check_status
            ;;
        "test")
            check_dependencies
            check_infrastructure
            run_tests
            ;;
        "logs")
            show_logs "$1"
            ;;
        *)
            log_error "未知命令: ${COMMAND:-start}"
            show_help
            exit 1
            ;;
    esac
}

# 错误处理
trap 'log_error "脚本执行失败，退出码: $?"' ERR

# 执行主程序
main "$@"