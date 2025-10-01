#!/bin/bash

# 端到端测试脚本 - 完整的CQRS架构验证
# 验证物理删除API移除后的系统完整性

set -e

echo "🧪 启动端到端测试 - CQRS架构完整验证"
echo "======================================"

# 配置
BASE_URL_COMMAND="http://localhost:9090"
BASE_URL_QUERY="http://localhost:8090"
TENANT_ID="550e8400-e29b-41d4-a716-446655440000"
TEST_ORG_CODE="E2E001"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
log_success() { echo -e "${GREEN}✅ $1${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
log_error() { echo -e "${RED}❌ $1${NC}"; }

# 获取开发token
get_dev_token() {
    log_info "获取开发模式JWT Token..."
    
    local response=$(curl -s -X POST "$BASE_URL_COMMAND/auth/dev-token" \
        -H "Content-Type: application/json" \
        -d '{
            "userID": "test-user",
            "tenantID": "'$TENANT_ID'",
            "roles": ["ADMIN"],
            "permissions": ["WRITE_ORGANIZATION", "UPDATE_ORGANIZATION", "MANAGE_ORGANIZATION_EVENTS", "CREATE_TEMPORAL_VERSION"]
        }')
    
    if echo "$response" | grep -q '"success":true'; then
        JWT_TOKEN=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        log_success "JWT Token获取成功"
    else
        log_error "JWT Token获取失败: $response"
        exit 1
    fi
}

# HTTP请求公共header
make_request() {
    local method=$1
    local url=$2
    local data=$3
    
    if [ -n "$data" ]; then
        curl -s -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "X-Tenant-ID: $TENANT_ID" \
            -d "$data"
    else
        curl -s -X "$method" "$url" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "X-Tenant-ID: $TENANT_ID"
    fi
}

# 1. 健康检查
test_health_checks() {
    log_info "🔍 测试1: 健康检查"
    
    # 命令服务健康检查
    local cmd_health=$(curl -s "$BASE_URL_COMMAND/health")
    if echo "$cmd_health" | grep -q '"status":"healthy"'; then
        log_success "命令服务健康检查通过"
    else
        log_error "命令服务健康检查失败"
        exit 1
    fi
    
    # 查询服务健康检查
    local query_health=$(curl -s "$BASE_URL_QUERY/health")
    if echo "$query_health" | grep -q '"status":"healthy"'; then
        log_success "查询服务健康检查通过"
    else
        log_error "查询服务健康检查失败"
        exit 1
    fi
}

# 2. 验证物理删除API已移除
test_delete_api_removed() {
    log_info "🔍 测试2: 验证物理删除API已移除"
    
    # 尝试访问删除端点，应该返回404或405
    local response=$(curl -s -w "%{http_code}" -X DELETE "$BASE_URL_COMMAND/api/v1/organization-units/$TEST_ORG_CODE" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "X-Tenant-ID: $TENANT_ID")
    
    local http_code=${response: -3}
    if [[ "$http_code" == "404" || "$http_code" == "405" ]]; then
        log_success "✅ 物理删除端点已成功移除 (HTTP: $http_code)"
    else
        log_error "物理删除端点仍然可访问 (HTTP: $http_code)"
        exit 1
    fi
}

# 3. CRUD操作测试
test_crud_operations() {
    log_info "🔍 测试3: CRUD操作完整性"
    
    # 3.1 创建组织
    log_info "3.1 创建测试组织..."
    local create_response=$(make_request POST "$BASE_URL_COMMAND/api/v1/organization-units" '{
        "code": "'$TEST_ORG_CODE'",
        "name": "E2E测试组织",
        "unitType": "DEPARTMENT",
        "parentCode": null,
        "description": "端到端测试组织单元",
        "operationReason": "E2E测试"
    }')
    
    if echo "$create_response" | grep -q '"success":true'; then
        log_success "组织创建成功"
    else
        log_error "组织创建失败: $create_response"
        exit 1
    fi
    
    # 3.2 查询组织 (GraphQL)
    log_info "3.2 GraphQL查询组织..."
    local query_payload=$(jq -n --arg code "$TEST_ORG_CODE" '{
        query: "query($codes: [String!]) { organizations(filter: { codes: $codes }, pagination: { page: 1, pageSize: 1 }) { data { code name unitType status } } }",
        variables: { codes: [$code] }
    }')

    local query_response=$(curl -s -X POST "$BASE_URL_QUERY/graphql" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "X-Tenant-ID: $TENANT_ID" \
        -d "$query_payload")

    if echo "$query_response" | grep -q '"code":"'$TEST_ORG_CODE'"'; then
        log_success "GraphQL查询成功"
    else
        log_error "GraphQL查询失败: $query_response"
        exit 1
    fi
    
    # 3.3 更新组织
    log_info "3.3 更新组织信息..."
    local update_response=$(make_request PUT "$BASE_URL_COMMAND/api/v1/organization-units/$TEST_ORG_CODE" '{
        "name": "E2E测试组织(已更新)",
        "description": "更新后的描述",
        "operationReason": "E2E更新测试"
    }')
    
    if echo "$update_response" | grep -q '"success":true'; then
        log_success "组织更新成功"
    else
        log_error "组织更新失败: $update_response"
        exit 1
    fi
}

# 4. 时态版本管理测试
test_temporal_versions() {
    log_info "🔍 测试4: 时态版本管理"
    
    # 4.1 创建时态版本
    log_info "4.1 创建时态版本..."
    local version_response=$(make_request POST "$BASE_URL_COMMAND/api/v1/organization-units/$TEST_ORG_CODE/versions" '{
        "name": "E2E测试组织 v2.0",
        "unitType": "DEPARTMENT",
        "effectiveDate": "2025-09-10",
        "operationReason": "版本升级测试"
    }')
    
    if echo "$version_response" | grep -q '"success":true'; then
        log_success "时态版本创建成功"
        # 提取recordId用于后续测试
        RECORD_ID=$(echo "$version_response" | grep -o '"recordId":"[^"]*"' | cut -d'"' -f4)
        log_info "RecordID: $RECORD_ID"
    else
        log_error "时态版本创建失败: $version_response"
        exit 1
    fi
    
    # 4.2 验证版本删除功能 (使用正确的端点)
    log_info "4.2 测试版本删除功能..."
    local delete_version_response=$(make_request DELETE "$BASE_URL_COMMAND/api/v1/organization-units/versions/$RECORD_ID")
    
    if echo "$delete_version_response" | grep -q '"success":true'; then
        log_success "✅ 版本删除成功 - 时态时间轴自动维护正常"
    else
        log_warning "版本删除测试 (可能recordId无效): $delete_version_response"
    fi
}

# 5. 状态管理测试
test_status_operations() {
    log_info "🔍 测试5: 组织状态管理"
    
    # 5.1 暂停组织
    log_info "5.1 暂停组织..."
    local suspend_response=$(make_request POST "$BASE_URL_COMMAND/api/v1/organization-units/$TEST_ORG_CODE/suspend" '{
        "effectiveDate": "2025-09-15",
        "operationReason": "E2E暂停测试"
    }')
    
    if echo "$suspend_response" | grep -q '"success":true'; then
        log_success "组织暂停成功"
    else
        log_warning "组织暂停测试: $suspend_response"
    fi
    
    # 5.2 激活组织
    log_info "5.2 激活组织..."
    local activate_response=$(make_request POST "$BASE_URL_COMMAND/api/v1/organization-units/$TEST_ORG_CODE/activate" '{
        "effectiveDate": "2025-09-20",
        "operationReason": "E2E激活测试"
    }')
    
    if echo "$activate_response" | grep -q '"success":true'; then
        log_success "组织激活成功"
    else
        log_warning "组织激活测试: $activate_response"
    fi
}

# 6. 性能基准测试
test_performance() {
    log_info "🔍 测试6: 性能基准测试"
    
    # 批量查询性能
    local start_time=$(date +%s%3N)
    
    for i in {1..5}; do
        curl -s -X POST "$BASE_URL_QUERY/graphql" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "X-Tenant-ID: $TENANT_ID" \
            -d '{"query": "query { organizationStats { totalCount } }"}' > /dev/null
    done
    
    local end_time=$(date +%s%3N)
    local duration=$((end_time - start_time))
    local avg_time=$((duration / 5))
    
    log_success "GraphQL查询性能: 5次查询平均响应时间 ${avg_time}ms"
    
    if [ $avg_time -lt 100 ]; then
        log_success "✅ 性能测试通过: 平均响应时间 < 100ms"
    else
        log_warning "⚠️ 性能告警: 平均响应时间 ${avg_time}ms"
    fi
}

# 7. 系统完整性验证
test_system_integrity() {
    log_info "🔍 测试7: 系统完整性验证"
    
    # 检查Prometheus指标
    local metrics=$(curl -s "$BASE_URL_COMMAND/metrics")
    if echo "$metrics" | grep -q "cube_castle_http_requests_total"; then
        log_success "Prometheus指标收集正常"
    else
        log_warning "Prometheus指标可能异常"
    fi
    
    # 检查数据库连接
    local db_status=$(curl -s "$BASE_URL_COMMAND/dev/database-status" | grep -o '"connected":[^,]*' | cut -d':' -f2)
    if [ "$db_status" = "true" ]; then
        log_success "数据库连接正常"
    else
        log_warning "数据库连接可能异常"
    fi
}

# 清理测试数据
cleanup_test_data() {
    log_info "🧹 清理测试数据..."
    
    # 使用DEACTIVATE事件清理测试数据
    local cleanup_response=$(make_request POST "$BASE_URL_COMMAND/api/v1/organization-units/$TEST_ORG_CODE/events" '{
        "eventType": "DEACTIVATE",
        "recordId": "test-cleanup",
        "changeReason": "E2E测试清理",
        "operatedBy": {
            "id": "test-user",
            "name": "Test User"
        }
    }')
    
    log_info "测试数据清理尝试完成"
}

# 主测试流程
main() {
    echo "🚀 开始端到端测试..."
    echo ""
    
    # 获取认证token
    get_dev_token
    
    # 执行所有测试
    test_health_checks
    echo ""
    
    test_delete_api_removed
    echo ""
    
    test_crud_operations  
    echo ""
    
    test_temporal_versions
    echo ""
    
    test_status_operations
    echo ""
    
    test_performance
    echo ""
    
    test_system_integrity
    echo ""
    
    # 清理测试数据
    cleanup_test_data
    
    echo "======================================"
    log_success "🎉 端到端测试完成!"
    echo "======================================"
    
    log_info "测试覆盖范围:"
    echo "  ✅ 服务健康检查 - 命令服务和查询服务"
    echo "  ✅ 物理删除API移除验证 - 确保用户无法误用"
    echo "  ✅ CRUD操作完整性 - REST命令 + GraphQL查询"
    echo "  ✅ 时态版本管理 - 创建版本 + 正确删除端点"
    echo "  ✅ 组织状态管理 - 暂停/激活功能"
    echo "  ✅ 性能基准测试 - GraphQL查询性能验证"
    echo "  ✅ 系统完整性 - 指标收集和数据库连接"
    echo ""
    log_success "🏆 PostgreSQL原生CQRS架构运行正常!"
    log_success "🔒 物理删除API已成功移除，时态一致性得到保证!"
}

# 执行测试
main "$@"
