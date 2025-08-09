#!/bin/bash
# 🚀 CQRS完整架构启动脚本 - 务实版本
# 确保所有必需服务正确启动，避免组织更名等问题

set -e

echo "🏰 启动 Cube Castle CQRS 完整架构"
echo "===================================="

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }

# 存储服务PID用于清理
declare -a SERVICE_PIDS=()

# 清理函数
cleanup() {
    echo ""
    print_warning "正在停止所有服务..."
    for pid in "${SERVICE_PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null || true
        fi
    done
    wait 2>/dev/null || true
    print_success "服务已停止"
    exit 0
}

# 设置信号处理
trap cleanup INT TERM

# 启动并验证服务的函数
start_and_verify_service() {
    local service_path=$1
    local service_name=$2
    local health_endpoint=$3
    local max_wait=${4:-30}
    
    echo "🚀 启动 $service_name..."
    
    # 检查服务目录是否存在
    if [ ! -d "$service_path" ]; then
        print_error "服务目录不存在: $service_path"
        return 1
    fi
    
    # 启动服务
    cd "$service_path"
    go run main.go &
    local service_pid=$!
    SERVICE_PIDS+=($service_pid)
    cd - > /dev/null
    
    # 等待服务启动
    echo "⏳ 等待 $service_name 启动..."
    local count=0
    while [ $count -lt $max_wait ]; do
        if curl -sf "$health_endpoint" > /dev/null 2>&1; then
            print_success "$service_name 启动成功 (PID: $service_pid)"
            return 0
        fi
        sleep 1
        ((count++))
    done
    
    print_error "$service_name 启动超时"
    return 1
}

# 检查CDC管道的函数
check_cdc_pipeline() {
    echo "🔍 检查CDC管道状态..."
    
    # 检查Debezium连接器状态
    local connector_status=$(curl -s http://localhost:8083/connectors/organization-postgres-connector/status 2>/dev/null | jq -r '.connector.state' 2>/dev/null || echo "UNKNOWN")
    
    if [ "$connector_status" = "RUNNING" ]; then
        local task_status=$(curl -s http://localhost:8083/connectors/organization-postgres-connector/status 2>/dev/null | jq -r '.tasks[0].state' 2>/dev/null || echo "UNKNOWN")
        if [ "$task_status" = "RUNNING" ]; then
            print_success "CDC管道运行正常"
            return 0
        fi
    fi
    
    print_warning "CDC连接器状态: $connector_status"
    print_warning "尝试重新配置Debezium连接器..."
    
    # 重新配置连接器（使用正确的网络配置）
    curl -X DELETE http://localhost:8083/connectors/organization-postgres-connector 2>/dev/null || true
    sleep 2
    
    curl -X POST http://localhost:8083/connectors \
      -H "Content-Type: application/json" \
      -d '{
        "name": "organization-postgres-connector",
        "config": {
          "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
          "database.hostname": "postgres",
          "database.port": "5432",
          "database.user": "user",
          "database.password": "password",
          "database.dbname": "cubecastle",
          "database.server.name": "organization_db",
          "table.include.list": "public.organization_units",
          "plugin.name": "pgoutput",
          "slot.name": "debezium_org_slot",
          "publication.name": "debezium_org_publication",
          "topic.prefix": "organization_db"
        }
      }' > /dev/null 2>&1
    
    sleep 5
    connector_status=$(curl -s http://localhost:8083/connectors/organization-postgres-connector/status 2>/dev/null | jq -r '.connector.state' 2>/dev/null || echo "FAILED")
    
    if [ "$connector_status" = "RUNNING" ]; then
        print_success "CDC连接器重新配置成功"
        return 0
    else
        print_error "CDC连接器配置失败，但继续启动服务"
        return 1
    fi
}

echo "📋 第1步: 检查基础设施服务"
echo "--------------------------------"

# 检查Docker容器状态
if ! docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "(postgres|neo4j|redis|kafka)" | grep -q "Up"; then
    print_error "基础设施服务未运行，请先执行: docker-compose up -d"
    exit 1
fi

print_success "基础设施服务运行正常"

echo ""
echo "📋 第2步: 启动CQRS核心服务"  
echo "--------------------------------"

# 启动4个必需的服务（顺序很重要）
start_and_verify_service "cmd/organization-command-service" "命令服务 (端口9090)" "http://localhost:9090/health" || exit 1
start_and_verify_service "cmd/organization-query-service-unified" "查询服务 (端口8090)" "http://localhost:8090/health" || exit 1
start_and_verify_service "cmd/organization-sync-service" "同步服务" "http://localhost:8084/health" || exit 1
start_and_verify_service "cmd/organization-cache-invalidator" "缓存失效服务" "http://localhost:8086/health" || { 
    print_warning "缓存失效服务健康检查失败，但服务可能正在运行"
}

echo ""
echo "📋 第3步: 验证CDC数据管道"
echo "--------------------------------"
check_cdc_pipeline

echo ""
echo "📋 第4步: 系统整体健康检查"
echo "--------------------------------"

# 综合健康检查
echo "🔍 测试完整的CQRS数据流..."

# 测试命令操作
echo "测试命令服务..."
test_response=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:9090/api/v1/organization-units \
  -H "Content-Type: application/json" \
  -d '{"name":"健康检查测试部门","unit_type":"DEPARTMENT","status":"INACTIVE"}' 2>/dev/null || echo "000")

if [ "$test_response" = "201" ]; then
    print_success "命令服务测试通过"
else
    print_warning "命令服务测试失败 (HTTP: $test_response)"
fi

# 测试查询操作
echo "测试查询服务..."
query_response=$(curl -s -X POST http://localhost:8090/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"query { organizationStats { total } }"}' 2>/dev/null | jq -r '.data.organizationStats.total' 2>/dev/null || echo "error")

if [ "$query_response" != "error" ] && [ "$query_response" != "null" ]; then
    print_success "查询服务测试通过 (组织总数: $query_response)"
else
    print_warning "查询服务测试失败"
fi

echo ""
echo "🎉 CQRS架构启动完成！"
echo "===================================="
echo ""
echo "📊 服务状态总览:"
echo "  🔧 命令服务: http://localhost:9090/health"
echo "  📊 查询服务: http://localhost:8090/health"  
echo "  🔄 同步服务: http://localhost:8084/health"
echo "  🗑️  缓存失效: http://localhost:8086/health"
echo ""
echo "🌐 访问地址:"
echo "  📱 前端应用: http://localhost:3000/ (需单独启动: cd frontend && npm run dev)"
echo "  🔧 GraphiQL: http://localhost:8090/graphiql"
echo "  📊 Kafka UI: http://localhost:8081"
echo ""
echo "🛑 停止所有服务: Ctrl+C"
echo ""

# 保持脚本运行，等待用户中断
print_success "所有服务正在运行，按 Ctrl+C 停止..."
while true; do
    sleep 10
    # 简单的服务健康检查
    for pid in "${SERVICE_PIDS[@]}"; do
        if ! kill -0 "$pid" 2>/dev/null; then
            print_error "检测到服务异常退出 (PID: $pid)"
            cleanup
        fi
    done
done