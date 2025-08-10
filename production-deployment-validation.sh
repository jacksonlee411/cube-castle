#!/bin/bash
# 生产环境部署验证脚本 - 时态管理API项目
# 验证所有关键服务的健康状态和性能指标

set -e

echo "🚀 生产环境部署验证 - 时态管理API"
echo "================================="

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

CHECKS_PASSED=0
CHECKS_FAILED=0

# 1. 服务健康检查
echo ""
echo "📋 第1步: 核心服务健康检查"
echo "------------------------"

check_service() {
    local service_name=$1
    local health_url=$2
    
    if curl -s -f "$health_url" >/dev/null 2>&1; then
        print_success "$service_name 服务运行正常"
        ((CHECKS_PASSED++))
    else
        print_error "$service_name 服务不可用 ($health_url)"
        ((CHECKS_FAILED++))
    fi
}

check_service "时态API服务 (9091)" "http://localhost:9091/health"
check_service "命令服务 (9090)" "http://localhost:9090/health"
check_service "查询服务 (8090)" "http://localhost:8090/health"

# 2. 基础设施检查
echo ""
echo "📋 第2步: 基础设施状态检查"
echo "------------------------"

# PostgreSQL检查
if docker exec cube_castle_postgres pg_isready -U user -d cubecastle >/dev/null 2>&1; then
    print_success "PostgreSQL数据库连接正常"
    ((CHECKS_PASSED++))
else
    print_error "PostgreSQL数据库连接失败"
    ((CHECKS_FAILED++))
fi

# Redis检查
if docker exec cube_castle_redis redis-cli ping | grep -q "PONG" >/dev/null 2>&1; then
    print_success "Redis缓存服务正常"
    ((CHECKS_PASSED++))
else
    print_error "Redis缓存服务异常"
    ((CHECKS_FAILED++))
fi

# Kafka检查
if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "cube_castle_kafka.*Up"; then
    print_success "Kafka消息队列正常"
    ((CHECKS_PASSED++))
else
    print_error "Kafka消息队列异常"
    ((CHECKS_FAILED++))
fi

# 3. API功能验证
echo ""
echo "📋 第3步: API功能验证"
echo "------------------"

# GraphQL查询测试
if curl -s -X POST http://localhost:8090/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"query { organizationStats { totalCount } }"}' \
    | grep -q "totalCount" >/dev/null 2>&1; then
    print_success "GraphQL查询服务功能正常"
    ((CHECKS_PASSED++))
else
    print_error "GraphQL查询服务功能异常"
    ((CHECKS_FAILED++))
fi

# 时态API查询测试
if curl -s http://localhost:9091/api/v1/organization-units/1000001 \
    | grep -q "organizations" >/dev/null 2>&1; then
    print_success "时态API查询功能正常"
    ((CHECKS_PASSED++))
else
    print_error "时态API查询功能异常"
    ((CHECKS_FAILED++))
fi

# 4. 性能基准测试
echo ""
echo "📋 第4步: 性能基准验证"
echo "------------------"

# GraphQL查询响应时间
gql_response_time=$(curl -s -w "%{time_total}" -o /dev/null \
    -X POST http://localhost:8090/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"query { organizationStats { totalCount } }"}')

if (( $(echo "$gql_response_time < 1.0" | bc -l) )); then
    print_success "GraphQL查询响应时间: ${gql_response_time}s (< 1s)"
    ((CHECKS_PASSED++))
else
    print_warning "GraphQL查询响应时间: ${gql_response_time}s (>= 1s)"
    ((CHECKS_FAILED++))
fi

# 时态API响应时间
temporal_response_time=$(curl -s -w "%{time_total}" -o /dev/null \
    http://localhost:9091/api/v1/organization-units/1000001)

if (( $(echo "$temporal_response_time < 1.0" | bc -l) )); then
    print_success "时态API响应时间: ${temporal_response_time}s (< 1s)"
    ((CHECKS_PASSED++))
else
    print_warning "时态API响应时间: ${temporal_response_time}s (>= 1s)"
    ((CHECKS_FAILED++))
fi

# 5. 监控指标验证
echo ""
echo "📋 第5步: 监控指标验证"
echo "------------------"

# 检查命令服务指标
if curl -s http://localhost:9090/metrics | grep -q "http_requests_total" >/dev/null 2>&1; then
    print_success "命令服务Prometheus指标正常"
    ((CHECKS_PASSED++))
else
    print_error "命令服务Prometheus指标缺失"
    ((CHECKS_FAILED++))
fi

# 6. 最终结果
echo ""
echo "📊 验证结果汇总"
echo "============="

total_checks=$((CHECKS_PASSED + CHECKS_FAILED))
success_rate=$(echo "scale=1; $CHECKS_PASSED * 100 / $total_checks" | bc -l)

echo "总检查项: $total_checks"
echo "通过项: $CHECKS_PASSED"
echo "失败项: $CHECKS_FAILED"
echo "成功率: ${success_rate}%"

if [ $CHECKS_FAILED -eq 0 ]; then
    print_success "🎉 生产环境部署验证通过！系统已就绪"
    echo ""
    echo "🚀 部署建议："
    echo "• 系统已具备生产环境运行能力"
    echo "• 建议配置负载均衡和自动重启"
    echo "• 建议配置Prometheus告警规则"
    echo "• 建议进行压力测试验证扩展性"
    exit 0
elif [ $CHECKS_FAILED -le 2 ]; then
    print_warning "⚠️  发现少量问题，建议修复后部署"
    exit 1
else
    print_error "❌ 发现严重问题，不建议部署到生产环境"
    exit 2
fi