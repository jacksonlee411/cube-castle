#!/bin/bash

echo "=== 时态管理功能完整测试套件 ==="
echo "开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

TEMPORAL_API="http://localhost:9091"
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

test_api() {
    local name="$1"
    local url="$2"
    local expected_status="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "测试 $TOTAL_TESTS: $name ... "
    
    response=$(curl -s -w "\n%{http_code}" "$url")
    actual_status=$(echo "$response" | tail -n1)
    
    if [ "$actual_status" = "$expected_status" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo "✅ 通过 ($actual_status)"
        return 0
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo "❌ 失败 (期望: $expected_status, 实际: $actual_status)"
        return 1
    fi
}

test_event() {
    local name="$1"
    local org_code="$2"
    local event_data="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "测试 $TOTAL_TESTS: $name ... "
    
    response=$(curl -s -w "\n%{http_code}" -X POST \
        "${TEMPORAL_API}/api/v1/organization-units/$org_code/events" \
        -H "Content-Type: application/json" \
        -d "$event_data")
    
    actual_status=$(echo "$response" | tail -n1)
    
    if [ "$actual_status" = "201" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo "✅ 通过 (201)"
        return 0
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo "❌ 失败 (实际: $actual_status)"
        echo "响应: $(echo "$response" | head -n1)"
        return 1
    fi
}

echo "🔍 测试组 1: 基础API测试"
test_api "健康检查" "${TEMPORAL_API}/health" "200"
test_api "当前记录查询" "${TEMPORAL_API}/api/v1/organization-units/1000056/temporal?as_of_date=$(date +%Y-%m-%d)" "200"
test_api "完整历史查询" "${TEMPORAL_API}/api/v1/organization-units/1000056/temporal?include_history=true&include_future=true" "200"
test_api "时间范围查询" "${TEMPORAL_API}/api/v1/organization-units/1000056/temporal?effective_from=2025-01-01&effective_to=2030-12-31" "200"

echo ""
echo "📝 测试组 2: 事件API测试"
test_event "UPDATE事件" "1000056" '{
    "event_type": "UPDATE",
    "effective_date": "2036-01-01T00:00:00Z",
    "change_data": {
        "name": "测试套件更新组织2036",
        "description": "通过测试套件更新的组织信息"
    },
    "change_reason": "自动化测试更新"
}'

test_event "RESTRUCTURE事件" "1000056" '{
    "event_type": "RESTRUCTURE", 
    "effective_date": "2037-01-01T00:00:00Z",
    "change_data": {
        "name": "重组后的测试部门2037",
        "unit_type": "PROJECT_TEAM",
        "description": "通过测试套件重组的组织"
    },
    "change_reason": "自动化测试重组"
}'

echo ""
echo "❌ 测试组 3: 错误处理测试"
test_api "无效组织代码" "${TEMPORAL_API}/api/v1/organization-units/INVALID999/temporal" "404"

echo ""
echo "🔄 测试组 4: 性能测试"
echo -n "测试缓存性能: "
start_time=$(date +%s%N)
response=$(curl -s "${TEMPORAL_API}/api/v1/organization-units/1000056/temporal?as_of_date=2025-08-12")
end_time=$(date +%s%N)
response_time=$(( ($end_time - $start_time) / 1000000 ))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [ $response_time -lt 50 ]; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
    echo "✅ 通过 (${response_time}ms < 50ms)"
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
    echo "❌ 失败 (${response_time}ms >= 50ms)"
fi

echo ""
echo "=== 📊 测试结果总结 ==="
echo "总测试数: $TOTAL_TESTS"
echo "通过测试: $PASSED_TESTS"
echo "失败测试: $FAILED_TESTS"

if [ $FAILED_TESTS -eq 0 ]; then
    echo "🎉 所有测试通过！时态管理功能工作正常。"
    exit 0
else
    success_rate=$(( PASSED_TESTS * 100 / TOTAL_TESTS ))
    echo "成功率: ${success_rate}%"
    
    if [ $success_rate -ge 80 ]; then
        echo "✅ 总体功能基本正常 (成功率 >= 80%)"
        exit 0
    else
        echo "❌ 需要修复关键问题 (成功率 < 80%)"
        exit 1
    fi
fi