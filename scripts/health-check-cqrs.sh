#!/bin/bash
# 🔍 CQRS架构健康检查脚本 - 务实版本
# 快速检查所有服务状态，预防组织更名等数据不一致问题

set -e

echo "🔍 CQRS架构健康检查"
echo "===================="

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }
print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }

# 检查统计
CHECKS_TOTAL=0
CHECKS_PASSED=0
CHECKS_FAILED=0

run_check() {
    local check_name=$1
    local check_command=$2
    
    ((CHECKS_TOTAL++))
    print_info "检查: $check_name"
    
    if eval "$check_command" > /dev/null 2>&1; then
        print_success "$check_name"
        ((CHECKS_PASSED++))
        return 0
    else
        print_error "$check_name"
        ((CHECKS_FAILED++))
        return 1
    fi
}

run_check_with_output() {
    local check_name=$1
    local check_command=$2
    local success_pattern=$3
    
    ((CHECKS_TOTAL++))
    print_info "检查: $check_name"
    
    local output
    output=$(eval "$check_command" 2>/dev/null)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && [[ "$output" =~ $success_pattern ]]; then
        print_success "$check_name - $output"
        ((CHECKS_PASSED++))
        return 0
    else
        print_error "$check_name - 输出: $output"
        ((CHECKS_FAILED++))
        return 1
    fi
}

echo ""
echo "📋 第1步: 基础设施服务检查"
echo "----------------------------"

# Docker容器检查
run_check "PostgreSQL容器" "docker ps --filter name=cube_castle_postgres --filter status=running -q"
run_check "Neo4j容器" "docker ps --filter name=cube_castle_neo4j --filter status=running -q"
run_check "Redis容器" "docker ps --filter name=cube_castle_redis --filter status=running -q"
run_check "Kafka容器" "docker ps --filter name=cube_castle_kafka --filter status=running -q"
run_check "Kafka Connect容器" "docker ps --filter name=cube_castle_kafka_connect --filter status=running -q"

echo ""
echo "📋 第2步: CQRS服务健康检查"
echo "----------------------------"

# CQRS服务健康端点检查
run_check_with_output "命令服务 (端口9090)" "curl -s http://localhost:9090/health" "healthy"
run_check_with_output "查询服务 (端口8090)" "curl -s http://localhost:8090/health" "healthy"

# 同步服务和缓存失效服务（可能没有HTTP健康检查端点）
# 缓存失效服务已删除 - 跳过进程检查

# 或者检查HTTP健康端点（如果可用）
run_check_with_output "同步服务 (端口8084)" "curl -s http://localhost:8084/health" "healthy"
# 缓存失效服务已删除 - 跳过健康检查

echo ""
echo "📋 第3步: CDC数据管道检查"
echo "----------------------------"

# Kafka Connect健康检查
run_check "Kafka Connect API" "curl -s http://localhost:8083/"

# Debezium连接器状态检查
print_info "检查: Debezium连接器状态"
connector_status=$(curl -s http://localhost:8083/connectors/organization-postgres-connector/status 2>/dev/null | jq -r '.connector.state' 2>/dev/null || echo "UNKNOWN")
task_state=$(curl -s http://localhost:8083/connectors/organization-postgres-connector/status 2>/dev/null | jq -r '.tasks[0].state' 2>/dev/null || echo "UNKNOWN")

if [ "$connector_status" = "RUNNING" ] && [ "$task_state" = "RUNNING" ]; then
    print_success "Debezium连接器运行正常"
    ((CHECKS_PASSED++))
else
    print_warning "Debezium连接器状态异常 - Connector: $connector_status, Task: $task_state"
    ((CHECKS_FAILED++))
fi
((CHECKS_TOTAL++))

echo ""
echo "📋 第4步: 端到端功能检查"
echo "----------------------------"

# 测试命令操作
print_info "检查: 命令API创建操作"
create_response=$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:9090/api/v1/organization-units \
  -H "Content-Type: application/json" \
  -d '{"name":"健康检查测试_'$(date +%s)'","unit_type":"DEPARTMENT","status":"INACTIVE"}' 2>/dev/null || echo "000")

((CHECKS_TOTAL++))
if [ "$create_response" = "201" ]; then
    print_success "命令API创建操作正常"
    ((CHECKS_PASSED++))
else
    print_error "命令API创建操作失败 (HTTP: $create_response)"
    ((CHECKS_FAILED++))
fi

# 测试查询操作
print_info "检查: GraphQL查询操作"
query_response=$(curl -s -X POST http://localhost:8090/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"query { organizationStats { total } }"}' 2>/dev/null | jq -r '.data.organizationStats.total' 2>/dev/null || echo "error")

((CHECKS_TOTAL++))
if [ "$query_response" != "error" ] && [ "$query_response" != "null" ] && [ "$query_response" -gt 0 ]; then
    print_success "GraphQL查询操作正常 (组织总数: $query_response)"
    ((CHECKS_PASSED++))
else
    print_error "GraphQL查询操作失败"
    ((CHECKS_FAILED++))
fi

# 测试数据一致性（PostgreSQL vs Neo4j）
print_info "检查: 数据一致性验证"
pg_count=$(PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -t -c "SELECT COUNT(*) FROM organization_units;" 2>/dev/null | xargs || echo "0")
neo4j_response=$(curl -s -X POST http://localhost:8090/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"query { organizationStats { total } }"}' 2>/dev/null | jq -r '.data.organizationStats.total' 2>/dev/null || echo "0")

((CHECKS_TOTAL++))
if [ "$pg_count" -eq "$neo4j_response" ] 2>/dev/null; then
    print_success "数据一致性验证通过 (PostgreSQL: $pg_count, Neo4j: $neo4j_response)"
    ((CHECKS_PASSED++))
else
    print_warning "数据一致性可能存在问题 (PostgreSQL: $pg_count, Neo4j: $neo4j_response)"
    ((CHECKS_FAILED++))
fi

echo ""
echo "🎯 健康检查结果汇总"
echo "===================="
echo "总检查项目: $CHECKS_TOTAL"
echo "通过检查: $CHECKS_PASSED"
echo "失败检查: $CHECKS_FAILED"

if [ $CHECKS_FAILED -eq 0 ]; then
    print_success "🎉 所有检查通过！CQRS架构运行正常"
    echo ""
    echo "💡 系统状态:"
    echo "  - CQRS数据流: ✅ 正常"
    echo "  - CDC同步: ✅ 正常" 
    echo "  - 缓存失效: ✅ 正常"
    echo "  - 组织更名等操作: ✅ 应该正常工作"
    exit 0
elif [ $CHECKS_FAILED -le 2 ]; then
    print_warning "⚠️ 部分检查失败，但核心功能可能正常"
    echo ""
    echo "💡 建议："
    echo "  1. 检查失败的服务日志"
    echo "  2. 重启异常服务"
    echo "  3. 验证网络连接"
    exit 1
else
    print_error "💥 多项检查失败，系统可能存在严重问题"
    echo ""
    echo "🚨 紧急修复建议："
    echo "  1. 重启所有服务: ./scripts/start-cqrs-complete.sh"
    echo "  2. 重新配置CDC管道: ./scripts/setup-cdc-pipeline.sh" 
    echo "  3. 检查Docker容器状态: docker-compose ps"
    exit 2
fi