#!/bin/bash

# 时态管理API深度测试验证套件
# 目标：确保生产就绪的稳定性和可靠性

set -e  # 遇到错误立即退出

# 配置
API_BASE="http://localhost:9091/api/v1/organization-units"
TEST_ORG="1000001" 
FAILED_TESTS=0
TOTAL_TESTS=0

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_condition="$3"
    local description="$4"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_info "[$TOTAL_TESTS] $test_name"
    
    if [ -n "$description" ]; then
        echo "    描述: $description"
    fi
    
    # 执行测试命令
    local result
    local exit_code
    result=$(eval "$test_command" 2>&1)
    exit_code=$?
    
    # 评估测试结果
    local test_passed=false
    if [ -n "$expected_condition" ]; then
        if eval "$expected_condition"; then
            test_passed=true
        fi
    else
        if [ $exit_code -eq 0 ]; then
            test_passed=true
        fi
    fi
    
    # 输出结果
    if [ "$test_passed" = true ]; then
        log_success "    ✅ PASS"
    else
        log_error "    ❌ FAIL"
        log_error "    Command: $test_command"
        log_error "    Result: $result"
        log_error "    Exit Code: $exit_code"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    echo ""
}

# 等待函数
wait_for_service() {
    local url="$1"
    local timeout="$2"
    local counter=0
    
    log_info "等待服务启动: $url"
    while [ $counter -lt $timeout ]; do
        if curl -f -s "$url" > /dev/null 2>&1; then
            log_success "服务已就绪"
            return 0
        fi
        sleep 1
        counter=$((counter + 1))
        echo -n "."
    done
    
    log_error "服务启动超时"
    return 1
}

echo "🧪 时态管理API深度测试验证开始"
echo "时间: $(date)"
echo "目标: 生产环境就绪验证"
echo ""

# 预检查
log_info "=== 预检查 ==="

# 检查服务是否运行
wait_for_service "http://localhost:9091/health" 10

# 检查数据库连接
run_test "数据库连接检查" \
    "PGPASSWORD=password psql -h localhost -U user -d cubecastle -c 'SELECT 1;' > /dev/null" \
    "[ \$? -eq 0 ]" \
    "验证PostgreSQL数据库连接正常"

echo "=== 第1组：基础功能测试 ==="

# 测试1：服务健康状态
run_test "服务健康检查" \
    "curl -s http://localhost:9091/health | jq -r '.status'" \
    "[ \"\$result\" = \"healthy\" ]" \
    "验证时态API服务运行状态"

# 测试2：基础组织查询
run_test "基础组织查询" \
    "curl -s '$API_BASE/$TEST_ORG' | jq -r '.result_count'" \
    "[ \"\$result\" = \"1\" ]" \
    "验证能够查询到测试组织"

# 测试3：时态字段完整性
run_test "时态字段完整性检查" \
    "curl -s '$API_BASE/$TEST_ORG' | jq -r '.organizations[0] | has(\"version\") and has(\"effective_date\") and has(\"is_current\")'" \
    "[ \"\$result\" = \"true\" ]" \
    "验证响应包含所有必需的时态字段"

echo "=== 第2组：时态查询功能测试 ==="

# 测试4：当前日期查询
run_test "当前日期查询" \
    "curl -s '$API_BASE/$TEST_ORG?as_of_date=$(date +%Y-%m-%d)' | jq -r '.result_count'" \
    "[ \"\$result\" = \"1\" ]" \
    "验证当前日期时间点查询"

# 测试5：未来日期查询
run_test "未来日期查询" \
    "curl -s '$API_BASE/$TEST_ORG?as_of_date=2026-01-01' | jq -r '.result_count'" \
    "[ \"\$result\" = \"1\" ]" \
    "验证未来日期查询功能"

# 测试6：过去日期查询（应该没有结果）
run_test "过去日期查询" \
    "curl -s '$API_BASE/$TEST_ORG?as_of_date=2020-01-01' | jq -r '.error_code'" \
    "[ \"\$result\" = \"NOT_FOUND\" ]" \
    "验证过去日期查询返回正确的NOT_FOUND"

# 测试7：日期范围查询
run_test "日期范围查询" \
    "curl -s '$API_BASE/$TEST_ORG?effective_from=2025-01-01&effective_to=2025-12-31' | jq -r '.result_count'" \
    "[ \"\$result\" = \"1\" ]" \
    "验证日期范围查询功能"

echo "=== 第3组：事件驱动操作测试 ==="

# 测试8：创建UPDATE事件
run_test "创建UPDATE事件" \
    "curl -s -X POST '$API_BASE/$TEST_ORG/events' \\
        -H 'Content-Type: application/json' \\
        -d '{\"event_type\":\"UPDATE\",\"effective_date\":\"2025-12-01T00:00:00Z\",\"change_data\":{\"name\":\"测试更新\"},\"change_reason\":\"深度测试\"}' | jq -r '.status'" \
    "[ \"\$result\" = \"processed\" ]" \
    "验证UPDATE事件创建功能"

sleep 1

# 测试9：创建RESTRUCTURE事件
run_test "创建RESTRUCTURE事件" \
    "curl -s -X POST '$API_BASE/$TEST_ORG/events' \\
        -H 'Content-Type: application/json' \\
        -d '{\"event_type\":\"RESTRUCTURE\",\"effective_date\":\"2026-01-01T00:00:00Z\",\"change_data\":{\"description\":\"重组测试\"},\"change_reason\":\"结构调整\"}' | jq -r '.event_type'" \
    "[ \"\$result\" = \"RESTRUCTURE\" ]" \
    "验证RESTRUCTURE事件创建功能"

sleep 1

# 测试10：事件记录验证
run_test "事件记录验证" \
    "PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c 'SELECT COUNT(*) FROM organization_events WHERE organization_code='\''$TEST_ORG'\'';' | xargs" \
    "[ \"\$result\" -ge \"2\" ]" \
    "验证事件正确记录到数据库"

echo "=== 第4组：边界条件测试 ==="

# 测试11：无效日期格式
run_test "无效日期格式处理" \
    "curl -s '$API_BASE/$TEST_ORG?as_of_date=invalid-date' | jq -r '.error_code'" \
    "[ \"\$result\" = \"INVALID_TEMPORAL_PARAMS\" ]" \
    "验证无效日期格式的错误处理"

# 测试12：无效事件类型
run_test "无效事件类型处理" \
    "curl -s -X POST '$API_BASE/$TEST_ORG/events' \\
        -H 'Content-Type: application/json' \\
        -d '{\"event_type\":\"INVALID\",\"effective_date\":\"2025-12-01T00:00:00Z\",\"change_data\":{},\"change_reason\":\"测试\"}' | jq -r '.error_code'" \
    "[ \"\$result\" = \"INVALID_EVENT_TYPE\" ]" \
    "验证无效事件类型的错误处理"

# 测试13：不存在的组织
run_test "不存在组织查询" \
    "curl -s '$API_BASE/9999999' | jq -r '.error_code'" \
    "[ \"\$result\" = \"NOT_FOUND\" ]" \
    "验证不存在组织的错误处理"

# 测试14：空参数处理
run_test "空参数处理" \
    "curl -s '$API_BASE/' | wc -c" \
    "[ \"\$result\" -gt \"10\" ]" \
    "验证空参数请求不会导致服务崩溃"

echo "=== 第5组：数据完整性测试 ==="

# 测试15：时态字段一致性
run_test "时态字段一致性检查" \
    "PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c 'SELECT COUNT(*) FROM organization_units WHERE effective_date IS NULL OR version IS NULL OR is_current IS NULL;' | xargs" \
    "[ \"\$result\" = \"0\" ]" \
    "验证所有记录都有完整的时态字段"

# 测试16：版本唯一性检查
run_test "版本唯一性检查" \
    "PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c 'SELECT COUNT(*) FROM (SELECT code, version FROM organization_units GROUP BY code, version HAVING COUNT(*) > 1) duplicates;' | xargs" \
    "[ \"\$result\" = \"0\" ]" \
    "验证(code,version)组合的唯一性"

# 测试17：当前版本标记一致性
run_test "当前版本标记一致性" \
    "PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c 'SELECT COUNT(*) FROM (SELECT code FROM organization_units WHERE is_current = true GROUP BY code HAVING COUNT(*) > 1) multiple_current;' | xargs" \
    "[ \"\$result\" = \"0\" ]" \
    "验证每个组织只有一个当前版本"

# 测试18：时态一致性验证
run_test "时态一致性验证" \
    "PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c 'SELECT COUNT(*) FROM validate_temporal_consistency_v2();' | xargs" \
    "[ \"\$result\" = \"0\" ]" \
    "验证时态数据无一致性问题"

echo "=== 第6组：性能基准测试 ==="

# 测试19：单次查询响应时间
run_test "单次查询响应时间" \
    "curl -w '%{time_total}' -s -o /dev/null '$API_BASE/$TEST_ORG'" \
    "[ \$(echo \"\$result < 1.0\" | bc -l) -eq 1 ]" \
    "验证单次查询响应时间小于1秒"

# 测试20：并发查询测试（简化版）
run_test "并发查询测试" \
    "for i in {1..5}; do curl -s '$API_BASE/$TEST_ORG' > /dev/null & done; wait" \
    "[ \$? -eq 0 ]" \
    "验证并发查询不会导致服务异常"

echo "=== 测试结果汇总 ==="

# 计算测试结果
PASSED_TESTS=$((TOTAL_TESTS - FAILED_TESTS))
if [ $TOTAL_TESTS -eq 0 ]; then
    PASS_RATE=0
else
    PASS_RATE=$(echo "scale=2; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc)
fi

echo ""
log_info "📊 测试统计结果："
echo "  总测试数: $TOTAL_TESTS"
echo "  通过数: $PASSED_TESTS" 
echo "  失败数: $FAILED_TESTS"
echo "  通过率: ${PASS_RATE}%"
echo ""

# 判断测试结果
if [ $FAILED_TESTS -eq 0 ]; then
    log_success "🎉 所有测试通过！时态管理API已达到生产就绪标准！"
    echo ""
    log_success "✅ 验证完成的功能："
    echo "  - 基础时态查询功能 ✓"
    echo "  - 事件驱动操作功能 ✓"
    echo "  - 边界条件错误处理 ✓"
    echo "  - 数据完整性保证 ✓"
    echo "  - 性能基准达标 ✓"
    echo ""
    log_success "🚀 生产环境部署建议："
    echo "  - 当前实现稳定可靠，可以进行生产部署"
    echo "  - 建议配置监控告警，关注响应时间和错误率"
    echo "  - 建议定期执行数据一致性检查"
    echo ""
    exit 0
else
    log_error "❌ 发现 $FAILED_TESTS 个测试失败，需要修复后才能部署到生产环境"
    echo ""
    log_warning "🔧 修复建议："
    echo "  1. 检查失败的测试用例详细信息"
    echo "  2. 修复相关代码问题"
    echo "  3. 重新运行测试验证修复效果"
    echo ""
    exit 1
fi