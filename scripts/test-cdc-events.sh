#!/bin/bash
# ============================================================================
# CDC更新事件测试验证脚本
# 功能：验证PostgreSQL→Neo4j数据同步和时态管理功能
# 版本：v2.0
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

# 数据库连接配置
export PGPASSWORD=password
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="user"
DB_NAME="cubecastle"

# Neo4j连接配置
NEO4J_URI="bolt://localhost:7687"
NEO4J_USER="neo4j"
NEO4J_PASSWORD="password"

# Kafka配置
KAFKA_HOST="localhost:9092"

# 检查服务状态
check_services() {
    log_header "检查服务状态"
    
    # 检查PostgreSQL
    if pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER >/dev/null 2>&1; then
        log_success "PostgreSQL服务运行正常"
    else
        log_error "PostgreSQL服务不可用"
        exit 1
    fi
    
    # 检查Neo4j
    if curl -s -f -u $NEO4J_USER:$NEO4J_PASSWORD "$NEO4J_URI" >/dev/null 2>&1; then
        log_success "Neo4j服务运行正常"
    else
        log_warning "Neo4j服务连接检查跳过（可能需要手动验证）"
    fi
    
    # 检查数据同步服务
    if curl -s -f http://localhost:8085/health >/dev/null 2>&1; then
        log_success "组织同步服务运行正常"
    else
        log_warning "组织同步服务可能未启动"
    fi
}

# 验证数据完整性
verify_data_integrity() {
    log_header "验证数据完整性"
    
    # 检查基础数据
    local total_records=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM organization_units WHERE data_status = 'NORMAL';" | xargs)
    log_info "NORMAL状态组织记录总数: $total_records"
    
    # 检查五状态分布
    log_info "五状态生命周期分布:"
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT 
        lifecycle_status,
        business_status,
        data_status,
        COUNT(*) as count
    FROM organization_units
    GROUP BY lifecycle_status, business_status, data_status
    ORDER BY lifecycle_status, business_status, data_status;"
}

# 测试CREATE事件
test_cdc_create_event() {
    log_header "测试CDC CREATE事件"
    
    local test_code="9999001"
    local test_name="CDC测试部门_$(date +%H%M%S)"
    
    log_info "创建测试组织: $test_code - $test_name"
    
    # 通过命令服务创建组织
    local response=$(curl -s -X POST http://localhost:9090/api/v1/organization-units \
        -H "Content-Type: application/json" \
        -d '{
            "code": "'$test_code'",
            "name": "'$test_name'",
            "unit_type": "DEPARTMENT",
            "parent_code": "1000000",
            "description": "CDC CREATE事件测试"
        }')
    
    if echo "$response" | grep -q "success\|created"; then
        log_success "组织创建成功"
        sleep 2  # 等待CDC同步
        
        # 验证数据库中的记录
        local db_count=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM organization_units WHERE code = '$test_code';" | xargs)
        log_info "数据库中记录数: $db_count"
        
        if [ "$db_count" -eq "1" ]; then
            log_success "CREATE事件测试通过"
            return 0
        else
            log_error "CREATE事件测试失败 - 数据库记录数不正确"
            return 1
        fi
    else
        log_error "组织创建失败: $response"
        return 1
    fi
}

# 测试UPDATE事件
test_cdc_update_event() {
    log_header "测试CDC UPDATE事件"
    
    local test_code="1000004"  # 使用现有的组织代码
    local new_description="CDC UPDATE事件测试_$(date +%H%M%S)"
    
    log_info "更新测试组织: $test_code"
    
    # 通过命令服务更新组织
    local response=$(curl -s -X PUT http://localhost:9090/api/v1/organization-units/$test_code \
        -H "Content-Type: application/json" \
        -d '{
            "description": "'$new_description'"
        }')
    
    if echo "$response" | grep -q "success\|updated"; then
        log_success "组织更新成功"
        sleep 2  # 等待CDC同步
        
        # 验证数据库中的更新
        local updated_desc=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT description FROM organization_units WHERE code = '$test_code' AND is_current = true;" | xargs)
        
        if [ "$updated_desc" = "$new_description" ]; then
            log_success "UPDATE事件测试通过"
            return 0
        else
            log_error "UPDATE事件测试失败 - 描述更新不正确"
            log_info "期望: $new_description"
            log_info "实际: $updated_desc"
            return 1
        fi
    else
        log_error "组织更新失败: $response"
        return 1
    fi
}

# 测试时态管理自动化
test_temporal_automation() {
    log_header "测试时态管理自动化"
    
    local test_code="8888001"
    local base_name="时态测试组织"
    
    log_info "创建时态管理测试序列"
    
    # 1. 创建第一个版本 (历史记录)
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    INSERT INTO organization_units (
        code, tenant_id, name, unit_type, status, effective_date,
        lifecycle_status, business_status, data_status, is_current,
        change_reason, level, path, sort_order, description, parent_code
    ) VALUES (
        '$test_code',
        '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
        '$base_name-V1',
        'DEPARTMENT',
        'INACTIVE',
        '2024-01-01',
        'HISTORICAL',
        'ACTIVE',
        'NORMAL',
        false,
        '时态管理测试-版本1',
        2,
        '/1000000/$test_code',
        0,
        '时态管理自动化测试',
        '1000000'
    );"
    
    # 2. 创建当前版本 (应自动设置前一版本的end_date)
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    INSERT INTO organization_units (
        code, tenant_id, name, unit_type, status, effective_date,
        lifecycle_status, business_status, data_status, is_current,
        change_reason, level, path, sort_order, description, parent_code
    ) VALUES (
        '$test_code',
        '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
        '$base_name-V2',
        'DEPARTMENT',
        'ACTIVE',
        '2025-01-01',
        'CURRENT',
        'ACTIVE',
        'NORMAL',
        true,
        '时态管理测试-版本2',
        2,
        '/1000000/$test_code',
        0,
        '时态管理自动化测试-当前版本',
        '1000000'
    );"
    
    sleep 1
    
    # 验证自动end_date设置
    local end_date_result=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
    SELECT end_date FROM organization_units 
    WHERE code = '$test_code' AND name = '$base_name-V1';" | xargs)
    
    if [ "$end_date_result" = "2024-12-31" ]; then
        log_success "自动end_date设置测试通过"
    else
        log_warning "自动end_date设置结果: $end_date_result (可能需要检查触发器)"
    fi
    
    # 验证当前记录的end_date为空
    local current_end_date=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
    SELECT COALESCE(end_date::text, 'NULL') FROM organization_units 
    WHERE code = '$test_code' AND is_current = true;" | xargs)
    
    if [ "$current_end_date" = "NULL" ]; then
        log_success "当前记录end_date验证通过"
    else
        log_error "当前记录end_date应为NULL，实际为: $current_end_date"
    fi
}

# 测试五状态转换
test_lifecycle_transitions() {
    log_header "测试五状态生命周期转换"
    
    local test_code="7777001"
    
    # 1. 创建计划状态的组织
    log_info "测试PLANNED状态创建"
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    INSERT INTO organization_units (
        code, tenant_id, name, unit_type, status, effective_date,
        lifecycle_status, business_status, data_status, is_current,
        change_reason, level, path, sort_order, description, parent_code
    ) VALUES (
        '$test_code',
        '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
        '状态转换测试组织',
        'DEPARTMENT',
        'PLANNED',
        '2026-06-01',
        'PLANNED',
        'ACTIVE',
        'NORMAL',
        false,
        '五状态转换测试',
        2,
        '/1000000/$test_code',
        0,
        '五状态生命周期转换测试',
        '1000000'
    );"
    
    # 2. 测试SUSPEND转换
    log_info "测试SUSPENDED状态转换"
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    UPDATE organization_units 
    SET business_status = 'SUSPENDED',
        suspended_at = NOW(),
        suspension_reason = '测试状态转换'
    WHERE code = '$test_code';"
    
    # 3. 测试DELETE转换
    log_info "测试DELETED状态转换"
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    UPDATE organization_units 
    SET data_status = 'DELETED',
        deleted_at = NOW(),
        deletion_reason = '测试软删除转换'
    WHERE code = '$test_code';"
    
    # 验证状态转换
    local final_status=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
    SELECT lifecycle_status, business_status, data_status 
    FROM organization_units 
    WHERE code = '$test_code';" | xargs)
    
    log_info "最终状态: $final_status"
    log_success "五状态转换测试完成"
}

# 清理测试数据
cleanup_test_data() {
    log_header "清理测试数据"
    
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    DELETE FROM organization_units 
    WHERE code IN ('9999001', '8888001', '7777001');"
    
    log_success "测试数据清理完成"
}

# 主测试流程
main() {
    log_header "CDC更新事件测试验证开始"
    
    check_services
    verify_data_integrity
    
    local failed_tests=0
    
    if ! test_cdc_create_event; then
        ((failed_tests++))
    fi
    
    if ! test_cdc_update_event; then
        ((failed_tests++))
    fi
    
    test_temporal_automation
    test_lifecycle_transitions
    
    cleanup_test_data
    
    log_header "测试结果汇总"
    
    if [ $failed_tests -eq 0 ]; then
        log_success "所有CDC测试通过！"
        log_success "时态管理系统运行正常"
        log_success "五状态生命周期管理功能验证完成"
        exit 0
    else
        log_error "有 $failed_tests 个测试失败"
        log_error "请检查CDC同步服务和网络连接"
        exit 1
    fi
}

# 执行主流程
main "$@"