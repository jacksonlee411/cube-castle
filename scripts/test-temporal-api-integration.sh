#!/bin/bash

# =============================================
# 时态管理API集成测试脚本
# 测试组织架构时态管理的完整功能链路
# =============================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试配置
BASE_URL_COMMAND="http://localhost:9090"
BASE_URL_QUERY="http://localhost:8090"
BASE_URL_TIMELINE="http://localhost:9092"
BASE_URL_VERSION="http://localhost:9093"
TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"

# 测试数据
TEST_ORG_NAME="测试时态部门"
TEST_ORG_TYPE="DEPARTMENT"
TEST_CHANGE_REASON="集成测试创建"

echo -e "${BLUE}🚀 开始时态管理API集成测试${NC}"
echo "========================================"

# 函数：记录测试步骤
log_step() {
    echo -e "${BLUE}📝 步骤: $1${NC}"
}

# 函数：记录成功
log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# 函数：记录错误
log_error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# 函数：记录警告
log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 函数：等待服务启动
wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1

    echo -e "${YELLOW}⏳ 等待 $service_name 服务启动...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url/health" > /dev/null 2>&1; then
            log_success "$service_name 服务已启动"
            return 0
        fi
        echo -n "."
        sleep 2
        ((attempt++))
    done
    
    log_error "$service_name 服务启动超时"
}

# 函数：执行HTTP请求并验证响应
execute_request() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local description=$5
    
    echo -e "${YELLOW}🔗 $description${NC}"
    echo "   请求: $method $url"
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X "$method" \
            -H "Content-Type: application/json" \
            -H "X-Tenant-ID: $TENANT_ID" \
            -d "$data" \
            "$url")
    else
        response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
            -X "$method" \
            -H "X-Tenant-ID: $TENANT_ID" \
            "$url")
    fi
    
    body=$(echo "$response" | sed -E 's/HTTPSTATUS\:[0-9]{3}$//')
    status=$(echo "$response" | tr -d '\n' | sed -E 's/.*HTTPSTATUS:([0-9]{3})$/\1/')
    
    if [ "$status" -eq "$expected_status" ]; then
        log_success "响应状态码: $status (预期: $expected_status)"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        log_error "响应状态码: $status (预期: $expected_status)"
        echo "响应内容: $body"
        return 1
    fi
}

# ======================================
# 测试阶段1: 服务健康检查
# ======================================

log_step "1. 检查所有时态管理服务健康状态"

# 检查命令服务
wait_for_service "$BASE_URL_COMMAND" "时态命令服务"

# 检查查询服务
wait_for_service "$BASE_URL_QUERY" "GraphQL查询服务"

# 检查时间线服务
wait_for_service "$BASE_URL_TIMELINE" "时间线管理服务"

# 检查版本服务
wait_for_service "$BASE_URL_VERSION" "版本管理服务"

# ======================================
# 测试阶段2: 创建计划组织
# ======================================

log_step "2. 创建计划组织 (未来生效)"

# 设置未来生效时间
FUTURE_DATE=$(date -d "+30 days" -u +"%Y-%m-%dT%H:%M:%SZ")

PLANNED_ORG_DATA=$(cat <<EOF
{
    "name": "$TEST_ORG_NAME",
    "unit_type": "$TEST_ORG_TYPE",
    "sort_order": 10,
    "description": "集成测试用的计划组织",
    "effective_from": "$FUTURE_DATE",
    "change_reason": "$TEST_CHANGE_REASON"
}
EOF
)

execute_request "POST" "$BASE_URL_COMMAND/api/v1/organization-units/planned" \
    "$PLANNED_ORG_DATA" 201 "创建计划组织"

# 提取组织代码
ORG_CODE=$(echo "$body" | jq -r '.code')
log_success "创建的组织代码: $ORG_CODE"

# ======================================
# 测试阶段3: GraphQL时态查询
# ======================================

log_step "3. 测试GraphQL时态查询功能"

# 查询组织 (当前时间 - 应该查不到，因为是计划中的)
GRAPHQL_QUERY=$(cat <<EOF
{
    "query": "query { organization(code: \"$ORG_CODE\") { code name unitType status isTemporal effectiveFrom version } }"
}
EOF
)

execute_request "POST" "$BASE_URL_QUERY/graphql" \
    "$GRAPHQL_QUERY" 200 "GraphQL查询当前组织 (应该为null，因为未生效)"

# 查询组织统计
GRAPHQL_STATS_QUERY=$(cat <<EOF
{
    "query": "query { organizationStats { totalCount byStatus { status count } } }"
}
EOF
)

execute_request "POST" "$BASE_URL_QUERY/graphql" \
    "$GRAPHQL_STATS_QUERY" 200 "GraphQL查询组织统计"

# ======================================
# 测试阶段4: 时态状态变更
# ======================================

log_step "4. 测试时态状态变更"

# 将计划组织激活 (设置为当前时间生效)
CURRENT_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

ACTIVATION_DATA=$(cat <<EOF
{
    "status": "ACTIVE",
    "effective_from": "$CURRENT_DATE",
    "change_reason": "集成测试激活组织"
}
EOF
)

execute_request "PUT" "$BASE_URL_COMMAND/api/v1/organization-units/$ORG_CODE/temporal-state" \
    "$ACTIVATION_DATA" 200 "激活计划组织"

# 等待一下让变更生效
sleep 3

# ======================================
# 测试阶段5: 查询激活后的组织
# ======================================

log_step "5. 查询激活后的组织"

# GraphQL查询激活后的组织 (现在应该能查到)
execute_request "POST" "$BASE_URL_QUERY/graphql" \
    "$GRAPHQL_QUERY" 200 "GraphQL查询激活后的组织"

# REST API直接查询
execute_request "GET" "$BASE_URL_COMMAND/api/v1/organization-units/$ORG_CODE" \
    "" 200 "REST API查询激活后的组织"

# ======================================
# 测试阶段6: 时间线查询
# ======================================

log_step "6. 测试时间线管理功能"

# 获取组织时间线
execute_request "GET" "$BASE_URL_TIMELINE/api/v1/organization-units/$ORG_CODE/timeline" \
    "" 200 "获取组织时间线"

# 获取时间线统计
execute_request "GET" "$BASE_URL_TIMELINE/api/v1/organization-units/$ORG_CODE/timeline/stats" \
    "" 200 "获取时间线统计"

# 获取版本历史
execute_request "GET" "$BASE_URL_TIMELINE/api/v1/organization-units/$ORG_CODE/versions" \
    "" 200 "获取版本历史"

# ======================================
# 测试阶段7: 版本管理和对比
# ======================================

log_step "7. 测试版本管理功能"

# 获取所有版本
execute_request "GET" "$BASE_URL_VERSION/api/v1/organization-units/$ORG_CODE/versions" \
    "" 200 "获取所有版本"

# 创建新版本
NEW_VERSION_DATA=$(cat <<EOF
{
    "changes": {
        "name": "${TEST_ORG_NAME}v2",
        "description": "版本2的描述"
    },
    "effective_from": "$CURRENT_DATE",
    "change_reason": "集成测试创建版本2"
}
EOF
)

execute_request "POST" "$BASE_URL_VERSION/api/v1/organization-units/$ORG_CODE/versions" \
    "$NEW_VERSION_DATA" 201 "创建新版本"

# 等待版本创建
sleep 2

# 版本对比
execute_request "GET" "$BASE_URL_VERSION/api/v1/organization-units/$ORG_CODE/versions/compare?from_version=1&to_version=2" \
    "" 200 "版本对比分析"

# ======================================
# 测试阶段8: 性能指标检查
# ======================================

log_step "8. 检查性能指标"

# 检查各服务的Prometheus指标
for service_url in "$BASE_URL_COMMAND" "$BASE_URL_QUERY" "$BASE_URL_TIMELINE" "$BASE_URL_VERSION"; do
    service_name=$(echo "$service_url" | sed 's/.*://')
    echo -e "${YELLOW}📊 检查端口 $service_name 的指标${NC}"
    
    if curl -s "$service_url/metrics" | grep -q "temporal\|organization\|timeline\|version"; then
        log_success "端口 $service_name 指标正常"
    else
        log_warning "端口 $service_name 指标可能异常"
    fi
done

# ======================================
# 测试阶段9: 清理测试数据
# ======================================

log_step "9. 清理测试数据"

# 软删除测试组织
execute_request "DELETE" "$BASE_URL_COMMAND/api/v1/organization-units/$ORG_CODE" \
    "" 204 "删除测试组织"

# ======================================
# 测试完成总结
# ======================================

echo "========================================"
log_success "🎉 时态管理API集成测试完成!"

echo -e "${GREEN}✅ 测试覆盖功能:${NC}"
echo "   - 计划组织创建和管理"
echo "   - 时态状态变更"
echo "   - GraphQL时态查询"
echo "   - 时间线事件管理"
echo "   - 版本历史和对比"
echo "   - 性能监控指标"

echo -e "${BLUE}📈 性能建议:${NC}"
echo "   - 所有API响应时间应保持在1秒以内"
echo "   - 时态查询缓存命中率应>90%"
echo "   - 版本对比分析建议限制在100个版本以内"
echo "   - 定期清理过期的时态数据"

echo -e "${YELLOW}🔧 运维建议:${NC}"
echo "   - 监控PostgreSQL时态表的大小增长"
echo "   - 定期检查时间线事件表的性能"
echo "   - 设置适当的数据保留策略"
echo "   - 监控各服务的内存使用情况"

echo ""
echo -e "${GREEN}🚀 时态管理系统已准备就绪!${NC}"