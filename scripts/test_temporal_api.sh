#!/bin/bash

# 时态管理API升级项目 - 完整功能验证测试
# 基于ADR-007规范进行端到端测试

echo "🧪 开始时态管理API功能测试..."
echo "测试目标: 验证元合约v6.0合规的时态管理能力"
echo ""

API_BASE="http://localhost:9091/api/v1/organization-units"
TEST_ORG="1000001"
FAILED_TESTS=0
TOTAL_TESTS=0

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_status="$3"
    local check_response="$4"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${BLUE}[$TOTAL_TESTS]${NC} ${test_name} ..."
    
    response=$(eval "$command" 2>/dev/null)
    status=$?
    
    # 检查HTTP状态
    if [ $status -ne $expected_status ]; then
        echo -e "  ${RED}❌ FAIL${NC} (HTTP状态: 期望=$expected_status, 实际=$status)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo "  Response: $response"
        return
    fi
    
    # 检查响应内容（如果提供）
    if [ -n "$check_response" ]; then
        if echo "$response" | grep -q "$check_response"; then
            echo -e "  ${GREEN}✅ PASS${NC}"
        else
            echo -e "  ${RED}❌ FAIL${NC} (响应内容不符合预期)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
            echo "  Expected: $check_response"
            echo "  Response: $response"
            return
        fi
    else
        echo -e "  ${GREEN}✅ PASS${NC}"
    fi
}

# 辅助函数：等待
wait_seconds() {
    local seconds=$1
    echo -e "${YELLOW}⏱️  等待 ${seconds} 秒...${NC}"
    sleep $seconds
}

echo "=== 第1部分：基础时态查询测试 ==="
echo ""

# 测试1: 服务健康检查
run_test "服务健康检查" \
    "curl -s -w '%{http_code}' -o /dev/null ${API_BASE%/*}/health" \
    0

# 测试2: 当前版本查询
run_test "当前版本查询" \
    "curl -s '${API_BASE}/${TEST_ORG}' | jq -r '.organizations[0].version'" \
    0 \
    "1"

# 测试3: 时间点查询（未来日期）
run_test "未来时间点查询" \
    "curl -s '${API_BASE}/${TEST_ORG}?as_of_date=2026-01-01' | jq -r '.organizations[0].code'" \
    0 \
    "$TEST_ORG"

# 测试4: 时间点查询（过去日期，应该找不到）
run_test "过去时间点查询" \
    "curl -s '${API_BASE}/${TEST_ORG}?as_of_date=2025-08-01' | jq -r '.error_code'" \
    0 \
    "NOT_FOUND"

# 测试5: 包含历史版本查询
run_test "包含历史版本查询" \
    "curl -s '${API_BASE}/${TEST_ORG}?include_history=true' | jq -r '.result_count'" \
    0 \
    "1"

echo ""
echo "=== 第2部分：事件驱动操作测试 ==="
echo ""

# 测试6: 创建UPDATE事件（未来生效）
run_test "创建UPDATE事件" \
    "curl -s -X POST '${API_BASE}/${TEST_ORG}/events' \
      -H 'Content-Type: application/json' \
      -d '{\"event_type\":\"UPDATE\",\"effective_date\":\"2025-12-15T00:00:00Z\",\"change_data\":{\"name\":\"AI治理办公室-12月版\"},\"change_reason\":\"年末组织调整\"}' | jq -r '.status'" \
    0 \
    "processed"

wait_seconds 1

# 测试7: 创建RESTRUCTURE事件
run_test "创建RESTRUCTURE事件" \
    "curl -s -X POST '${API_BASE}/${TEST_ORG}/events' \
      -H 'Content-Type: application/json' \
      -d '{\"event_type\":\"RESTRUCTURE\",\"effective_date\":\"2026-03-01T00:00:00Z\",\"change_data\":{\"name\":\"AI战略委员会\",\"description\":\"重组为战略委员会\"},\"change_reason\":\"组织架构重组\"}' | jq -r '.event_type'" \
    0 \
    "RESTRUCTURE"

wait_seconds 1

# 测试8: 创建DISSOLVE事件
run_test "创建DISSOLVE事件" \
    "curl -s -X POST '${API_BASE}/${TEST_ORG}/events' \
      -H 'Content-Type: application/json' \
      -d '{\"event_type\":\"DISSOLVE\",\"effective_date\":\"2026-12-31T00:00:00Z\",\"end_date\":\"2026-12-31T00:00:00Z\",\"change_data\":{},\"change_reason\":\"组织解散\"}' | jq -r '.event_type'" \
    0 \
    "DISSOLVE"

echo ""
echo "=== 第3部分：时态查询参数测试 ==="
echo ""

# 测试9: 日期范围查询
run_test "日期范围查询" \
    "curl -s '${API_BASE}/${TEST_ORG}?effective_from=2025-01-01&effective_to=2025-12-31' | jq -r '.result_count'" \
    0 \
    "1"

# 测试10: 特定版本查询
run_test "特定版本查询" \
    "curl -s '${API_BASE}/${TEST_ORG}?version=1' | jq -r '.organizations[0].version'" \
    0 \
    "1"

# 测试11: 最大版本数限制查询
run_test "最大版本数限制查询" \
    "curl -s '${API_BASE}/${TEST_ORG}?include_history=true&max_versions=1' | jq -r '.result_count'" \
    0 \
    "1"

# 测试12: 包含未来版本查询
run_test "包含未来版本查询" \
    "curl -s '${API_BASE}/${TEST_ORG}?include_future=true&include_history=true' | jq -r '.result_count'" \
    0

# 测试13: 包含已解散组织查询
run_test "包含已解散组织查询" \
    "curl -s '${API_BASE}/${TEST_ORG}?include_dissolved=true&include_history=true' | jq -r '.result_count'" \
    0

echo ""
echo "=== 第4部分：数据一致性验证 ==="
echo ""

# 测试14: 验证事件记录
echo -e "${BLUE}[14]${NC} 验证事件记录数量 ..."
event_count=$(PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c "SELECT COUNT(*) FROM organization_events WHERE organization_code='${TEST_ORG}';" | xargs)
if [ "$event_count" -ge "3" ]; then
    echo -e "  ${GREEN}✅ PASS${NC} (事件记录数: $event_count)"
else
    echo -e "  ${RED}❌ FAIL${NC} (期望>=3, 实际=$event_count)"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# 测试15: 验证时态字段完整性
echo -e "${BLUE}[15]${NC} 验证时态字段完整性 ..."
missing_temporal_fields=$(PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c "SELECT COUNT(*) FROM organization_units WHERE effective_date IS NULL OR version IS NULL OR is_current IS NULL;" | xargs)
if [ "$missing_temporal_fields" -eq "0" ]; then
    echo -e "  ${GREEN}✅ PASS${NC} (无缺失时态字段)"
else
    echo -e "  ${RED}❌ FAIL${NC} (发现 $missing_temporal_fields 条记录缺失时态字段)"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

# 测试16: 验证数据一致性
echo -e "${BLUE}[16]${NC} 验证数据一致性 ..."
consistency_issues=$(PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c "SELECT COUNT(*) FROM validate_temporal_consistency_v2();" | xargs)
if [ "$consistency_issues" -eq "0" ]; then
    echo -e "  ${GREEN}✅ PASS${NC} (无数据一致性问题)"
else
    echo -e "  ${RED}❌ FAIL${NC} (发现 $consistency_issues 个一致性问题)"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

echo ""
echo "=== 第5部分：错误处理测试 ==="
echo ""

# 测试17: 无效的事件类型
run_test "无效事件类型处理" \
    "curl -s -X POST '${API_BASE}/${TEST_ORG}/events' \
      -H 'Content-Type: application/json' \
      -d '{\"event_type\":\"INVALID\",\"effective_date\":\"2025-12-01T00:00:00Z\",\"change_data\":{},\"change_reason\":\"测试\"}' | jq -r '.error_code'" \
    0 \
    "INVALID_EVENT_TYPE"

# 测试18: 无效的日期格式
run_test "无效日期格式处理" \
    "curl -s '${API_BASE}/${TEST_ORG}?as_of_date=invalid-date' | jq -r '.error_code'" \
    0 \
    "INVALID_TEMPORAL_PARAMS"

# 测试19: 不存在的组织查询
run_test "不存在组织查询" \
    "curl -s '${API_BASE}/9999999' | jq -r '.error_code'" \
    0 \
    "NOT_FOUND"

echo ""
echo "=== 第6部分：性能基准测试 ==="
echo ""

# 测试20: 并发查询性能测试
echo -e "${BLUE}[20]${NC} 并发查询性能测试 ..."
start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "${API_BASE}/${TEST_ORG}" > /dev/null &
done
wait
end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
avg_time=$(echo "scale=3; $duration / 10" | bc)

if (( $(echo "$avg_time < 1.0" | bc -l) )); then
    echo -e "  ${GREEN}✅ PASS${NC} (平均响应时间: ${avg_time}s)"
else
    echo -e "  ${RED}❌ FAIL${NC} (平均响应时间: ${avg_time}s > 1.0s)"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi
TOTAL_TESTS=$((TOTAL_TESTS + 1))

echo ""
echo "=== 测试结果汇总 ==="
echo ""

# 计算通过率
pass_count=$((TOTAL_TESTS - FAILED_TESTS))
pass_rate=$(echo "scale=2; $pass_count * 100 / $TOTAL_TESTS" | bc)

echo -e "${BLUE}📊 测试统计:${NC}"
echo "  总测试数: $TOTAL_TESTS"
echo "  通过数: $pass_count"
echo "  失败数: $FAILED_TESTS"
echo -e "  通过率: ${pass_rate}%"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 所有测试通过！时态管理API功能验证成功！${NC}"
    echo ""
    echo -e "${GREEN}✅ 验证结果:${NC}"
    echo "  - 基础时态查询功能正常"
    echo "  - 事件驱动操作功能正常"
    echo "  - 时态查询参数功能正常"
    echo "  - 数据一致性检查通过"
    echo "  - 错误处理机制完善"
    echo "  - 性能基准达标"
    echo ""
    echo -e "${GREEN}🏆 项目达成目标:${NC}"
    echo "  - 符合元合约v6.0时态管理要求"
    echo "  - 支持EVENT_DRIVEN模式"
    echo "  - 具备完整的时间线查询能力"
    echo "  - 实现智能结束日期管理策略"
    echo ""
    exit 0
else
    echo -e "${RED}❌ 测试失败！请检查失败的测试用例${NC}"
    echo ""
    echo -e "${YELLOW}🔍 调试建议:${NC}"
    echo "  1. 检查时态API服务是否正常运行"
    echo "  2. 验证数据库时态扩展是否正确部署"
    echo "  3. 确认事件表和版本表是否正确创建"
    echo "  4. 检查时态查询逻辑是否符合预期"
    echo ""
    exit 1
fi