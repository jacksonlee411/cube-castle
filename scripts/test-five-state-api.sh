#!/bin/bash
# ============================================================================
# 五状态生命周期管理系统 API 集成测试脚本
# 功能：全面测试后端API的五状态生命周期管理功能
# 版本：v2.1
# 创建时间：2025-08-18
# ============================================================================

set -e  # 遇到错误立即退出

# 颜色输出函数
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

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

log_header() {
    echo -e "${PURPLE}========================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}========================================${NC}"
}

# API端点配置
COMMAND_API="http://localhost:9090/api/v1/organization-units"
QUERY_API="http://localhost:8090/graphql"
TEMPORAL_API="http://localhost:9091/api/v1/organization-units"

# 数据库连接配置
export PGPASSWORD=password
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="user"
DB_NAME="cubecastle"

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试辅助函数
run_test() {
    local test_name="$1"
    local test_function="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_info "执行测试: $test_name"
    
    if $test_function; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        log_success "✅ $test_name"
        return 0
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        log_error "❌ $test_name"
        return 1
    fi
}

# 数据库查询辅助函数
query_db() {
    local query="$1"
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "$query" | xargs
}

# API请求辅助函数
post_json() {
    local url="$1"
    local data="$2"
    curl -s -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$data"
}

get_json() {
    local url="$1"
    curl -s -X GET "$url" \
        -H "Content-Type: application/json"
}

# 测试1: 验证五状态数据完整性
test_five_state_data_integrity() {
    log_info "验证五状态生命周期管理的数据完整性"
    
    # 检查当前记录数量
    local current_count=$(query_db "SELECT COUNT(*) FROM organization_units WHERE lifecycle_status = 'CURRENT' AND data_status = 'NORMAL';")
    log_info "当前记录数量: $current_count"
    
    # 检查历史记录数量
    local historical_count=$(query_db "SELECT COUNT(*) FROM organization_units WHERE lifecycle_status = 'HISTORICAL' AND data_status = 'NORMAL';")
    log_info "历史记录数量: $historical_count"
    
    # 检查计划记录数量
    local planned_count=$(query_db "SELECT COUNT(*) FROM organization_units WHERE lifecycle_status = 'PLANNED' AND data_status = 'NORMAL';")
    log_info "计划记录数量: $planned_count"
    
    # 检查停用记录数量
    local suspended_count=$(query_db "SELECT COUNT(*) FROM organization_units WHERE business_status = 'SUSPENDED' AND data_status = 'NORMAL';")
    log_info "停用记录数量: $suspended_count"
    
    # 检查删除记录数量
    local deleted_count=$(query_db "SELECT COUNT(*) FROM organization_units WHERE data_status = 'DELETED';")
    log_info "删除记录数量: $deleted_count"
    
    # 验证约束：每个组织代码应该最多只有一个当前记录
    local duplicate_current=$(query_db "SELECT COUNT(*) FROM (SELECT code, COUNT(*) as cnt FROM organization_units WHERE lifecycle_status = 'CURRENT' AND data_status = 'NORMAL' GROUP BY code HAVING COUNT(*) > 1) as duplicates;")
    
    if [ "$duplicate_current" -eq "0" ]; then
        log_success "约束验证通过：每个组织代码最多只有一个当前记录"
        return 0
    else
        log_error "约束验证失败：发现 $duplicate_current 个组织有多个当前记录"
        return 1
    fi
}

# 测试2: 验证自动结束日期管理
test_auto_end_date_management() {
    log_info "验证自动结束日期管理功能"
    
    local test_code="TEST9001"
    local test_name="自动结束日期测试组织"
    
    # 创建第一个版本
    local response1=$(post_json "$COMMAND_API" '{
        "code": "'$test_code'",
        "name": "'$test_name'-V1",
        "unit_type": "DEPARTMENT",
        "status": "ACTIVE",
        "effective_date": "2024-01-01",
        "parent_code": "1000000"
    }')
    
    sleep 1
    
    # 创建第二个版本（应该自动设置第一个版本的结束日期）
    local response2=$(post_json "$COMMAND_API" '{
        "code": "'$test_code'",
        "name": "'$test_name'-V2",
        "unit_type": "DEPARTMENT", 
        "status": "ACTIVE",
        "effective_date": "2025-01-01",
        "parent_code": "1000000"
    }')
    
    sleep 1
    
    # 验证第一个版本的结束日期是否自动设置
    local end_date=$(query_db "SELECT end_date FROM organization_units WHERE code = '$test_code' AND effective_date = '2024-01-01';")
    
    # 清理测试数据
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM organization_units WHERE code = '$test_code';" > /dev/null
    
    if [ "$end_date" = "2024-12-31" ]; then
        log_success "自动结束日期设置正确: $end_date"
        return 0
    else
        log_error "自动结束日期设置失败，期望: 2024-12-31, 实际: $end_date"
        return 1
    fi
}

# 测试3: 验证状态转换约束
test_state_transition_constraints() {
    log_info "验证五状态转换约束"
    
    local test_code="TEST9002"
    
    # 创建测试组织
    local response=$(post_json "$COMMAND_API" '{
        "code": "'$test_code'",
        "name": "状态转换测试组织",
        "unit_type": "DEPARTMENT",
        "status": "ACTIVE",
        "effective_date": "2025-01-01",
        "parent_code": "1000000"
    }')
    
    sleep 1
    
    # 测试ACTIVE -> SUSPENDED转换
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    UPDATE organization_units 
    SET business_status = 'SUSPENDED', 
        suspended_at = NOW(),
        suspension_reason = '测试停用转换'
    WHERE code = '$test_code';" > /dev/null
    
    # 验证转换结果
    local suspended_status=$(query_db "SELECT business_status FROM organization_units WHERE code = '$test_code' AND is_current = true;")
    
    # 测试SUSPENDED -> ACTIVE转换  
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    UPDATE organization_units 
    SET business_status = 'ACTIVE',
        suspended_at = NULL,
        suspension_reason = NULL
    WHERE code = '$test_code';" > /dev/null
    
    local restored_status=$(query_db "SELECT business_status FROM organization_units WHERE code = '$test_code' AND is_current = true;")
    
    # 清理测试数据
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM organization_units WHERE code = '$test_code';" > /dev/null
    
    if [ "$suspended_status" = "SUSPENDED" ] && [ "$restored_status" = "ACTIVE" ]; then
        log_success "状态转换约束验证通过"
        return 0
    else
        log_error "状态转换约束验证失败，停用状态: $suspended_status, 恢复状态: $restored_status"
        return 1
    fi
}

# 测试4: 验证软删除功能
test_soft_delete_functionality() {
    log_info "验证软删除功能"
    
    local test_code="TEST9003"
    
    # 创建测试组织
    post_json "$COMMAND_API" '{
        "code": "'$test_code'",
        "name": "软删除测试组织",
        "unit_type": "DEPARTMENT",
        "status": "ACTIVE",
        "effective_date": "2025-01-01",
        "parent_code": "1000000"
    }' > /dev/null
    
    sleep 1
    
    # 执行软删除
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    UPDATE organization_units 
    SET data_status = 'DELETED',
        deleted_at = NOW(),
        deletion_reason = '测试软删除功能'
    WHERE code = '$test_code';" > /dev/null
    
    # 验证记录仍存在但标记为删除
    local data_status=$(query_db "SELECT data_status FROM organization_units WHERE code = '$test_code';")
    local deleted_at=$(query_db "SELECT deleted_at FROM organization_units WHERE code = '$test_code';")
    
    # 验证正常查询中不包含已删除记录
    local normal_count=$(query_db "SELECT COUNT(*) FROM organization_units WHERE code = '$test_code' AND data_status = 'NORMAL';")
    
    # 清理测试数据
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM organization_units WHERE code = '$test_code';" > /dev/null
    
    if [ "$data_status" = "DELETED" ] && [ "$deleted_at" != "" ] && [ "$normal_count" -eq "0" ]; then
        log_success "软删除功能验证通过"
        return 0
    else
        log_error "软删除功能验证失败"
        return 1
    fi
}

# 测试5: 验证GraphQL查询支持五状态
test_graphql_five_state_support() {
    log_info "验证GraphQL查询五状态支持"
    
    local graphql_query='{
        "query": "query { organizations(first: 5) { code name lifecycle_status business_status data_status effective_date } }"
    }'
    
    local response=$(post_json "$QUERY_API" "$graphql_query")
    
    # 检查响应是否包含五状态字段
    if echo "$response" | grep -q "lifecycle_status" && echo "$response" | grep -q "business_status" && echo "$response" | grep -q "data_status"; then
        log_success "GraphQL五状态查询支持验证通过"
        return 0
    else
        log_error "GraphQL五状态查询支持验证失败"
        log_error "响应内容: $response"
        return 1
    fi
}

# 测试6: 验证时态API支持
test_temporal_api_support() {
    log_info "验证时态API对五状态的支持"
    
    # 查询组织1000004的时态历史
    local temporal_response=$(get_json "$TEMPORAL_API/1000004/temporal?include_history=true")
    
    # 检查响应是否包含五状态信息
    if echo "$temporal_response" | grep -q "lifecycle_status\|business_status\|data_status"; then
        log_success "时态API五状态支持验证通过"
        return 0
    else
        log_error "时态API五状态支持验证失败"
        log_error "响应内容: $temporal_response"
        return 1
    fi
}

# 测试7: 验证数据库约束
test_database_constraints() {
    log_info "验证数据库五状态约束"
    
    local test_code="TEST9004"
    
    # 尝试插入违反约束的数据 - SUSPENDED状态但没有suspended_at
    local constraint_test=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    INSERT INTO organization_units (
        code, tenant_id, name, unit_type, status, effective_date,
        lifecycle_status, business_status, data_status, is_current,
        change_reason, level, path, sort_order, parent_code
    ) VALUES (
        '$test_code', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
        '约束测试组织', 'DEPARTMENT', 'ACTIVE', '2025-01-01',
        'CURRENT', 'SUSPENDED', 'NORMAL', true,
        '约束测试', 2, '/1000000/$test_code', 0, '1000000'
    );" 2>&1)
    
    # 清理可能的测试数据
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM organization_units WHERE code = '$test_code';" > /dev/null 2>&1
    
    # 检查是否因为约束失败
    if echo "$constraint_test" | grep -q "check_suspended_metadata"; then
        log_success "数据库约束验证通过 - 正确拒绝违反约束的数据"
        return 0
    else
        log_error "数据库约束验证失败 - 应该拒绝违反约束的数据"
        return 1
    fi
}

# 测试8: 验证触发器功能
test_trigger_functionality() {
    log_info "验证五状态生命周期管理触发器"
    
    local test_code="TEST9005"
    
    # 创建计划状态的组织（未来日期）
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    INSERT INTO organization_units (
        code, tenant_id, name, unit_type, status, effective_date,
        lifecycle_status, business_status, data_status, is_current,
        change_reason, level, path, sort_order, parent_code
    ) VALUES (
        '$test_code', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
        '触发器测试组织', 'DEPARTMENT', 'PLANNED', '2026-06-01',
        'PLANNED', 'ACTIVE', 'NORMAL', false,
        '触发器测试', 2, '/1000000/$test_code', 0, '1000000'
    );" > /dev/null
    
    # 检查触发器是否正确设置了lifecycle_status
    local lifecycle_status=$(query_db "SELECT lifecycle_status FROM organization_units WHERE code = '$test_code';")
    local is_current=$(query_db "SELECT is_current FROM organization_units WHERE code = '$test_code';")
    
    # 清理测试数据
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "DELETE FROM organization_units WHERE code = '$test_code';" > /dev/null
    
    if [ "$lifecycle_status" = "PLANNED" ] && [ "$is_current" = "f" ]; then
        log_success "触发器功能验证通过"
        return 0
    else
        log_error "触发器功能验证失败，lifecycle_status: $lifecycle_status, is_current: $is_current"
        return 1
    fi
}

# 主测试执行函数
main() {
    log_header "五状态生命周期管理系统 API 集成测试开始"
    
    # 执行所有测试
    run_test "五状态数据完整性验证" test_five_state_data_integrity
    run_test "自动结束日期管理功能" test_auto_end_date_management
    run_test "状态转换约束验证" test_state_transition_constraints
    run_test "软删除功能验证" test_soft_delete_functionality
    run_test "GraphQL五状态支持验证" test_graphql_five_state_support
    run_test "时态API支持验证" test_temporal_api_support
    run_test "数据库约束验证" test_database_constraints
    run_test "触发器功能验证" test_trigger_functionality
    
    # 输出测试结果汇总
    log_header "测试结果汇总"
    log_info "总测试数: $TOTAL_TESTS"
    log_success "通过测试: $PASSED_TESTS"
    log_error "失败测试: $FAILED_TESTS"
    
    local success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    log_info "成功率: ${success_rate}%"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        log_header "🎉 所有测试通过！五状态生命周期管理系统运行正常"
        exit 0
    else
        log_header "⚠️ 有测试失败，请检查系统配置"
        exit 1
    fi
}

# 执行主函数
main "$@"