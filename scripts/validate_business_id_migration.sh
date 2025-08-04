#!/bin/bash

# 业务ID系统迁移验证脚本
# 文件: validate_business_id_migration.sh
# 日期: 2025-08-04
# 描述: 全面验证业务ID系统迁移的数据完整性和功能正确性

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_test() {
    echo -e "${PURPLE}[TEST]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 配置变量
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-cubecastle}"
POSTGRES_USER="${POSTGRES_USER:-user}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-password}"

NEO4J_HOST="${NEO4J_HOST:-localhost}"
NEO4J_PORT="${NEO4J_PORT:-7687}"
NEO4J_USER="${NEO4J_USER:-neo4j}"
NEO4J_PASSWORD="${NEO4J_PASSWORD:-password}"

SCRIPT_DIR="$(dirname "$0")"
VALIDATION_LOG="${SCRIPT_DIR}/logs/validation_$(date '+%Y%m%d_%H%M%S').log"

# 创建日志目录
mkdir -p "$(dirname "$VALIDATION_LOG")"

# 全局计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# =============================================
# 测试框架函数
# =============================================

run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_result="${3:-0}"  # 默认期望成功(0)
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_test "运行测试: $test_name"
    
    if eval "$test_command" >> "$VALIDATION_LOG" 2>&1; then
        local result=$?
        if [ "$result" -eq "$expected_result" ]; then
            log_success "✅ 测试通过: $test_name"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            log_error "❌ 测试失败: $test_name (期望结果: $expected_result, 实际结果: $result)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    else
        log_error "❌ 测试执行失败: $test_name"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

run_sql_test() {
    local test_name="$1"
    local sql_query="$2"
    local expected_count="${3:-}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_test "运行SQL测试: $test_name"
    
    local result
    result=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -t -c "$sql_query" 2>>"$VALIDATION_LOG" | xargs)
    
    if [ -n "$expected_count" ]; then
        if [ "$result" = "$expected_count" ]; then
            log_success "✅ SQL测试通过: $test_name (结果: $result)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            log_error "❌ SQL测试失败: $test_name (期望: $expected_count, 实际: $result)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    else
        if [ -n "$result" ] && [ "$result" != "0" ]; then
            log_success "✅ SQL测试通过: $test_name (结果: $result)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            log_error "❌ SQL测试失败: $test_name (结果为空或0)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    fi
}

run_cypher_test() {
    local test_name="$1"
    local cypher_query="$2"
    local expected_count="${3:-}"
    
    if ! command -v cypher-shell >/dev/null 2>&1; then
        log_warning "⚠️ 跳过Cypher测试: $test_name (cypher-shell未安装)"
        return 0
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_test "运行Cypher测试: $test_name"
    
    local result
    result=$(cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" \
        "$cypher_query" 2>>"$VALIDATION_LOG" | tail -n 1 | awk '{print $1}')
    
    if [ -n "$expected_count" ]; then
        if [ "$result" = "$expected_count" ]; then
            log_success "✅ Cypher测试通过: $test_name (结果: $result)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            log_error "❌ Cypher测试失败: $test_name (期望: $expected_count, 实际: $result)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    else
        if [ -n "$result" ] && [ "$result" != "0" ]; then
            log_success "✅ Cypher测试通过: $test_name (结果: $result)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            log_error "❌ Cypher测试失败: $test_name (结果为空或0)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            return 1
        fi
    fi
}

# =============================================
# PostgreSQL验证函数
# =============================================

validate_postgresql_schema() {
    log_info "开始PostgreSQL Schema验证..."
    
    # 测试1: 检查业务ID字段是否存在
    run_sql_test "员工表business_id字段存在" "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='corehr' AND table_name='employees' AND column_name='business_id'" "1"
    
    run_sql_test "组织表business_id字段存在" "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='corehr' AND table_name='organizations' AND column_name='business_id'" "1"
    
    # 测试2: 检查序列是否创建
    run_sql_test "员工业务ID序列存在" "SELECT COUNT(*) FROM information_schema.sequences WHERE sequence_name='employee_business_id_seq'" "1"
    
    run_sql_test "组织业务ID序列存在" "SELECT COUNT(*) FROM information_schema.sequences WHERE sequence_name='org_business_id_seq'" "1"
    
    # 测试3: 检查函数是否创建
    run_sql_test "业务ID生成函数存在" "SELECT COUNT(*) FROM information_schema.routines WHERE routine_name='generate_business_id'" "1"
    
    run_sql_test "业务ID验证函数存在" "SELECT COUNT(*) FROM information_schema.routines WHERE routine_name='validate_business_id'" "1"
    
    # 测试4: 检查约束是否存在
    run_sql_test "员工业务ID唯一约束存在" "SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema='corehr' AND table_name='employees' AND constraint_name='uk_employees_business_id'" "1"
    
    run_sql_test "组织业务ID唯一约束存在" "SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema='corehr' AND table_name='organizations' AND constraint_name='uk_organizations_business_id'" "1"
    
    # 测试5: 检查索引是否存在
    run_sql_test "员工业务ID索引存在" "SELECT COUNT(*) FROM pg_indexes WHERE tablename='employees' AND indexname='idx_employees_business_id'" "1"
    
    run_sql_test "组织业务ID索引存在" "SELECT COUNT(*) FROM pg_indexes WHERE tablename='organizations' AND indexname='idx_organizations_business_id'" "1"
    
    log_success "PostgreSQL Schema验证完成"
}

validate_postgresql_data_integrity() {
    log_info "开始PostgreSQL数据完整性验证..."
    
    # 测试1: 检查所有员工都有业务ID
    run_sql_test "所有员工都有业务ID" "SELECT COUNT(*) FROM corehr.employees WHERE business_id IS NULL" "0"
    
    # 测试2: 检查所有组织都有业务ID
    run_sql_test "所有组织都有业务ID" "SELECT COUNT(*) FROM corehr.organizations WHERE business_id IS NULL" "0"
    
    # 测试3: 检查员工业务ID唯一性
    run_sql_test "员工业务ID无重复" "SELECT COUNT(*) - COUNT(DISTINCT business_id) FROM corehr.employees" "0"
    
    # 测试4: 检查组织业务ID唯一性
    run_sql_test "组织业务ID无重复" "SELECT COUNT(*) - COUNT(DISTINCT business_id) FROM corehr.organizations" "0"
    
    # 测试5: 检查员工业务ID格式
    run_sql_test "员工业务ID格式正确" "SELECT COUNT(*) FROM corehr.employees WHERE business_id !~ '^[1-9][0-9]{0,7}$'" "0"
    
    # 测试6: 检查组织业务ID格式
    run_sql_test "组织业务ID格式正确" "SELECT COUNT(*) FROM corehr.organizations WHERE business_id !~ '^[1-9][0-9]{5}$'" "0"
    
    # 测试7: 检查员工业务ID范围
    run_sql_test "员工业务ID在有效范围内" "SELECT COUNT(*) FROM corehr.employees WHERE business_id::integer < 1 OR business_id::integer > 99999999" "0"
    
    # 测试8: 检查组织业务ID范围
    run_sql_test "组织业务ID在有效范围内" "SELECT COUNT(*) FROM corehr.organizations WHERE business_id::integer < 100000 OR business_id::integer > 999999" "0"
    
    log_success "PostgreSQL数据完整性验证完成"
}

validate_postgresql_functions() {
    log_info "开始PostgreSQL函数验证..."
    
    # 测试1: 测试员工业务ID生成
    run_sql_test "员工业务ID生成功能" "SELECT CASE WHEN generate_business_id('employee') ~ '^[1-9][0-9]{0,7}$' THEN 1 ELSE 0 END" "1"
    
    # 测试2: 测试组织业务ID生成
    run_sql_test "组织业务ID生成功能" "SELECT CASE WHEN generate_business_id('organization') ~ '^[1-9][0-9]{5}$' THEN 1 ELSE 0 END" "1"
    
    # 测试3: 测试员工业务ID验证
    run_sql_test "员工业务ID验证功能" "SELECT CASE WHEN validate_business_id('employee', '12345') = true THEN 1 ELSE 0 END" "1"
    
    # 测试4: 测试组织业务ID验证
    run_sql_test "组织业务ID验证功能" "SELECT CASE WHEN validate_business_id('organization', '123456') = true THEN 1 ELSE 0 END" "1"
    
    # 测试5: 测试无效格式验证
    run_sql_test "无效格式拒绝" "SELECT CASE WHEN validate_business_id('employee', '0123') = false THEN 1 ELSE 0 END" "1"
    
    log_success "PostgreSQL函数验证完成"
}

# =============================================
# Neo4j验证函数
# =============================================

validate_neo4j_data() {
    if ! command -v cypher-shell >/dev/null 2>&1; then
        log_warning "跳过Neo4j验证 (cypher-shell未安装)"
        return 0
    fi
    
    log_info "开始Neo4j数据验证..."
    
    # 测试1: 检查员工节点有业务ID
    run_cypher_test "所有员工节点都有业务ID" "MATCH (e:Employee) WHERE e.business_id IS NULL RETURN count(e)" "0"
    
    # 测试2: 检查组织节点有业务ID
    run_cypher_test "所有组织节点都有业务ID" "MATCH (o:Organization) WHERE o.business_id IS NULL RETURN count(o)" "0"
    
    # 测试3: 检查员工业务ID格式
    run_cypher_test "员工业务ID格式正确" "MATCH (e:Employee) WHERE NOT e.business_id =~ '^[1-9][0-9]{0,7}$' RETURN count(e)" "0"
    
    # 测试4: 检查组织业务ID格式
    run_cypher_test "组织业务ID格式正确" "MATCH (o:Organization) WHERE NOT o.business_id =~ '^[1-9][0-9]{5}$' RETURN count(o)" "0"
    
    # 测试5: 检查索引是否创建
    run_test "员工业务ID索引存在" "cypher-shell -a 'bolt://$NEO4J_HOST:$NEO4J_PORT' -u '$NEO4J_USER' -p '$NEO4J_PASSWORD' 'SHOW INDEXES' | grep -i 'employee_business_id'"
    
    run_test "组织业务ID索引存在" "cypher-shell -a 'bolt://$NEO4J_HOST:$NEO4J_PORT' -u '$NEO4J_USER' -p '$NEO4J_PASSWORD' 'SHOW INDEXES' | grep -i 'organization_business_id'"
    
    log_success "Neo4j数据验证完成"
}

# =============================================
# 性能验证函数
# =============================================

validate_query_performance() {
    log_info "开始查询性能验证..."
    
    # 测试1: 员工业务ID查询性能
    local start_time=$(date +%s%N)
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -c "SELECT * FROM corehr.employees WHERE business_id = '1';" > /dev/null 2>&1
    local end_time=$(date +%s%N)
    local duration=$((($end_time - $start_time) / 1000000))  # 转换为毫秒
    
    if [ "$duration" -lt 100 ]; then  # 小于100ms
        log_success "✅ 员工业务ID查询性能良好: ${duration}ms"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        log_warning "⚠️ 员工业务ID查询性能较慢: ${duration}ms"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # 测试2: 组织业务ID查询性能
    start_time=$(date +%s%N)
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -c "SELECT * FROM corehr.organizations WHERE business_id = '100000';" > /dev/null 2>&1
    end_time=$(date +%s%N)
    duration=$((($end_time - $start_time) / 1000000))
    
    if [ "$duration" -lt 100 ]; then
        log_success "✅ 组织业务ID查询性能良好: ${duration}ms"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        log_warning "⚠️ 组织业务ID查询性能较慢: ${duration}ms"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # Neo4j性能测试
    if command -v cypher-shell >/dev/null 2>&1; then
        start_time=$(date +%s%N)
        cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" \
            "MATCH (e:Employee {business_id: '1'}) RETURN e;" > /dev/null 2>&1
        end_time=$(date +%s%N)
        duration=$((($end_time - $start_time) / 1000000))
        
        if [ "$duration" -lt 200 ]; then  # Neo4j允许稍慢一些
            log_success "✅ Neo4j员工业务ID查询性能良好: ${duration}ms"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            log_warning "⚠️ Neo4j员工业务ID查询性能较慢: ${duration}ms"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    fi
    
    log_success "查询性能验证完成"
}

# =============================================
# 数据一致性验证函数
# =============================================

validate_data_consistency() {
    log_info "开始数据一致性验证..."
    
    # 获取PostgreSQL数据统计
    local pg_emp_count pg_org_count
    pg_emp_count=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -t -c "SELECT COUNT(*) FROM corehr.employees;" | xargs)
    pg_org_count=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -t -c "SELECT COUNT(*) FROM corehr.organizations;" | xargs)
    
    log_info "PostgreSQL统计: 员工 $pg_emp_count, 组织 $pg_org_count"
    
    # 获取Neo4j数据统计 (如果可用)
    if command -v cypher-shell >/dev/null 2>&1; then
        local neo4j_emp_count neo4j_org_count
        neo4j_emp_count=$(cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" \
            "MATCH (e:Employee) RETURN count(e)" 2>/dev/null | tail -n 1 | awk '{print $1}')
        neo4j_org_count=$(cypher-shell -a "bolt://$NEO4J_HOST:$NEO4J_PORT" -u "$NEO4J_USER" -p "$NEO4J_PASSWORD" \
            "MATCH (o:Organization) RETURN count(o)" 2>/dev/null | tail -n 1 | awk '{print $1}')
        
        log_info "Neo4j统计: 员工 $neo4j_emp_count, 组织 $neo4j_org_count"
        
        # 比较数据一致性
        if [ "$pg_emp_count" = "$neo4j_emp_count" ]; then
            log_success "✅ 员工数据一致性验证通过"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            log_error "❌ 员工数据不一致: PostgreSQL($pg_emp_count) vs Neo4j($neo4j_emp_count)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        
        if [ "$pg_org_count" = "$neo4j_org_count" ]; then
            log_success "✅ 组织数据一致性验证通过"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            log_error "❌ 组织数据不一致: PostgreSQL($pg_org_count) vs Neo4j($neo4j_org_count)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    fi
    
    log_success "数据一致性验证完成"
}

# =============================================
# 业务逻辑验证函数
# =============================================

validate_business_logic() {
    log_info "开始业务逻辑验证..."
    
    # 测试1: 创建新员工时自动生成业务ID
    log_test "测试员工创建时业务ID自动生成"
    local test_email="test_$(date +%s)@example.com"
    
    if psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -c "INSERT INTO corehr.employees (tenant_id, first_name, last_name, email, hire_date) VALUES ('00000000-0000-0000-0000-000000000000', 'Test', 'User', '$test_email', '2025-01-01'); 
            SELECT business_id FROM corehr.employees WHERE email = '$test_email';" \
        | grep -E '^[1-9][0-9]{0,7}$' > /dev/null 2>&1; then
        log_success "✅ 员工创建时业务ID自动生成测试通过"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        log_error "❌ 员工创建时业务ID自动生成测试失败"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # 清理测试数据
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -c "DELETE FROM corehr.employees WHERE email = '$test_email';" > /dev/null 2>&1
    
    log_success "业务逻辑验证完成"
}

# =============================================
# 向后兼容性验证函数
# =============================================

validate_backward_compatibility() {
    log_info "开始向后兼容性验证..."
    
    # 测试1: UUID字段仍然存在且可查询
    run_sql_test "UUID字段仍然存在" "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema='corehr' AND table_name='employees' AND column_name='id'" "1"
    
    # 测试2: UUID查询仍然工作
    local test_uuid
    test_uuid=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
        -t -c "SELECT id FROM corehr.employees LIMIT 1;" | xargs)
    
    if [ -n "$test_uuid" ]; then
        run_sql_test "UUID查询功能正常" "SELECT COUNT(*) FROM corehr.employees WHERE id = '$test_uuid'" "1"
    else
        log_warning "⚠️ 无法找到测试UUID，跳过UUID查询测试"
    fi
    
    log_success "向后兼容性验证完成"
}

# =============================================
# 生成验证报告
# =============================================

generate_validation_report() {
    local report_file="${SCRIPT_DIR}/logs/validation_report_$(date '+%Y%m%d_%H%M%S').html"
    
    cat > "$report_file" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>业务ID系统迁移验证报告</title>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .success { color: #28a745; }
        .error { color: #dc3545; }
        .warning { color: #ffc107; }
        .test-section { margin: 20px 0; padding: 15px; border-left: 4px solid #007bff; }
        .summary { background-color: #e9ecef; padding: 15px; border-radius: 5px; margin: 20px 0; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>业务ID系统迁移验证报告</h1>
        <p>生成时间: $(date)</p>
        <p>验证目标: PostgreSQL ($POSTGRES_HOST:$POSTGRES_PORT) & Neo4j ($NEO4J_HOST:$NEO4J_PORT)</p>
    </div>
    
    <div class="summary">
        <h2>验证总结</h2>
        <table>
            <tr><th>指标</th><th>结果</th></tr>
            <tr><td>总测试数</td><td>$TOTAL_TESTS</td></tr>
            <tr><td><span class="success">通过测试</span></td><td class="success">$PASSED_TESTS</td></tr>
            <tr><td><span class="error">失败测试</span></td><td class="error">$FAILED_TESTS</td></tr>
            <tr><td>成功率</td><td>$(( PASSED_TESTS * 100 / TOTAL_TESTS ))%</td></tr>
        </table>
    </div>
    
    <div class="test-section">
        <h3>详细验证日志</h3>
        <pre>$(cat "$VALIDATION_LOG")</pre>
    </div>
    
    <div class="test-section">
        <h3>建议和后续步骤</h3>
        <ul>
EOF

    if [ "$FAILED_TESTS" -eq 0 ]; then
        echo "<li class='success'>✅ 所有测试通过，迁移成功！</li>" >> "$report_file"
        echo "<li>可以开始部署API代码更新</li>" >> "$report_file"
        echo "<li>建议进行用户验收测试</li>" >> "$report_file"
    else
        echo "<li class='error'>❌ 存在 $FAILED_TESTS 个失败的测试，需要修复</li>" >> "$report_file"
        echo "<li>请检查验证日志中的具体错误信息</li>" >> "$report_file"
        echo "<li>修复问题后重新运行验证</li>" >> "$report_file"
    fi

    cat >> "$report_file" << EOF
            <li>定期运行此验证脚本以确保数据一致性</li>
            <li>监控查询性能并优化索引</li>
        </ul>
    </div>
</body>
</html>
EOF

    log_info "验证报告已生成: $report_file"
}

# =============================================
# 主函数
# =============================================

main() {
    local validation_type="${1:-all}"
    
    log_info "开始业务ID系统迁移验证..."
    log_info "验证日志保存在: $VALIDATION_LOG"
    
    # 设置PostgreSQL密码环境变量
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    case "$validation_type" in
        "all")
            validate_postgresql_schema
            validate_postgresql_data_integrity
            validate_postgresql_functions
            validate_neo4j_data
            validate_query_performance
            validate_data_consistency
            validate_business_logic
            validate_backward_compatibility
            ;;
        "schema")
            validate_postgresql_schema
            ;;
        "data")
            validate_postgresql_data_integrity
            validate_neo4j_data
            validate_data_consistency
            ;;
        "performance")
            validate_query_performance
            ;;
        "business")
            validate_business_logic
            ;;
        "compatibility")
            validate_backward_compatibility
            ;;
        *)
            echo "用法: $0 {all|schema|data|performance|business|compatibility}"
            echo ""
            echo "验证类型说明:"
            echo "  all           - 运行所有验证测试 (默认)"
            echo "  schema        - 验证数据库Schema更改"
            echo "  data          - 验证数据完整性和一致性"
            echo "  performance   - 验证查询性能"
            echo "  business      - 验证业务逻辑"
            echo "  compatibility - 验证向后兼容性"
            exit 1
            ;;
    esac
    
    # 生成验证报告
    generate_validation_report
    
    # 显示最终结果
    echo ""
    echo "==========================================="
    echo "验证完成"
    echo "==========================================="
    echo "总测试数: $TOTAL_TESTS"
    echo -e "通过测试: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失败测试: ${RED}$FAILED_TESTS${NC}"
    echo "成功率: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
    echo ""
    
    if [ "$FAILED_TESTS" -eq 0 ]; then
        echo -e "${GREEN}🎉 验证完全通过！业务ID系统迁移成功！${NC}"
        exit 0
    else
        echo -e "${RED}⚠️ 存在 $FAILED_TESTS 个失败的测试，请检查并修复${NC}"
        echo "详细信息请查看: $VALIDATION_LOG"
        exit 1
    fi
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi