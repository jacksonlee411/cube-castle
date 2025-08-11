#!/bin/bash
# 时态API功能测试脚本
# 文件: tests/api/test_temporal_api_functionality.sh

set -e

echo "🧪 开始执行时态API功能测试"
echo "测试目标: 验证删除organization_versions表后时态API的完整性"

# 测试配置
TEMPORAL_API_BASE="http://localhost:9091/api/v1/organization-units"
TEST_ORG_CODE="1000056"
TEST_RESULTS=()

# 辅助函数：记录测试结果
log_test_result() {
    local test_name="$1"
    local status="$2"
    local details="$3"
    
    if [[ "$status" == "PASS" ]]; then
        echo "✅ PASSED: $test_name"
        [[ -n "$details" ]] && echo "   详情: $details"
    else
        echo "❌ FAILED: $test_name"
        [[ -n "$details" ]] && echo "   错误: $details"
    fi
    
    TEST_RESULTS+=("$status:$test_name")
}

# 辅助函数：API请求
api_request() {
    local endpoint="$1"
    local expected_status="${2:-200}"
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" "$endpoint" 2>/dev/null || echo "HTTPSTATUS:000")
    http_status=$(echo "$response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d':' -f2)
    response_body=$(echo "$response" | sed 's/HTTPSTATUS:[0-9]*$//')
    
    if [[ "$http_status" == "$expected_status" ]]; then
        echo "$response_body"
        return 0
    else
        echo "HTTP_ERROR:$http_status:$response_body"
        return 1
    fi
}

echo ""
echo "📡 测试1: 基础时态API端点可用性"
response=$(api_request "${TEMPORAL_API_BASE}/${TEST_ORG_CODE}/temporal")
if [[ "$response" != "HTTP_ERROR"* ]]; then
    org_count=$(echo "$response" | jq -r '.result_count // 0' 2>/dev/null || echo "0")
    if [[ "$org_count" -gt 0 ]]; then
        log_test_result "基础时态API端点可用性" "PASS" "返回 $org_count 条记录"
    else
        log_test_result "基础时态API端点可用性" "FAIL" "API返回数据为空"
    fi
else
    log_test_result "基础时态API端点可用性" "FAIL" "$response"
fi

echo ""
echo "📡 测试2: 当前有效记录查询"
response=$(api_request "${TEMPORAL_API_BASE}/${TEST_ORG_CODE}/temporal")
if [[ "$response" != "HTTP_ERROR"* ]]; then
    org_name=$(echo "$response" | jq -r '.organizations[0].name // "null"' 2>/dev/null || echo "null")
    is_current=$(echo "$response" | jq -r '.organizations[0].is_current // false' 2>/dev/null || echo "false")
    
    if [[ "$org_name" != "null" && "$is_current" == "true" ]]; then
        log_test_result "当前有效记录查询" "PASS" "组织: $org_name, 当前有效: $is_current"
    else
        log_test_result "当前有效记录查询" "FAIL" "数据字段异常: name=$org_name, is_current=$is_current"
    fi
else
    log_test_result "当前有效记录查询" "FAIL" "$response"
fi

echo ""
echo "📡 测试3: 时间点查询(as_of_date)"
test_date="2025-08-01"
response=$(api_request "${TEMPORAL_API_BASE}/${TEST_ORG_CODE}/temporal?as_of_date=${test_date}")
if [[ "$response" != "HTTP_ERROR"* ]]; then
    queried_at=$(echo "$response" | jq -r '.queried_at // "null"' 2>/dev/null || echo "null")
    if [[ "$queried_at" != "null" ]]; then
        log_test_result "时间点查询功能" "PASS" "成功查询 $test_date 的记录"
    else
        log_test_result "时间点查询功能" "FAIL" "时间点查询响应格式异常"
    fi
else
    log_test_result "时间点查询功能" "FAIL" "$response"
fi

echo ""
echo "📡 测试4: 时间范围查询"
from_date="2024-01-01"
to_date="2025-12-31"
response=$(api_request "${TEMPORAL_API_BASE}/${TEST_ORG_CODE}/temporal?effective_from=${from_date}&effective_to=${to_date}")
if [[ "$response" != "HTTP_ERROR"* ]]; then
    result_count=$(echo "$response" | jq -r '.result_count // 0' 2>/dev/null || echo "0")
    if [[ "$result_count" -gt 0 ]]; then
        log_test_result "时间范围查询功能" "PASS" "查询到 $result_count 条记录"
    else
        log_test_result "时间范围查询功能" "FAIL" "时间范围查询无结果"
    fi
else
    log_test_result "时间范围查询功能" "FAIL" "$response"
fi

echo ""
echo "📡 测试5: 时态字段完整性验证"
response=$(api_request "${TEMPORAL_API_BASE}/${TEST_ORG_CODE}/temporal")
if [[ "$response" != "HTTP_ERROR"* ]]; then
    effective_date=$(echo "$response" | jq -r '.organizations[0].effective_date // "null"' 2>/dev/null)
    end_date=$(echo "$response" | jq -r '.organizations[0].end_date // "null"' 2>/dev/null)
    change_reason=$(echo "$response" | jq -r '.organizations[0].change_reason // "null"' 2>/dev/null)
    is_current=$(echo "$response" | jq -r '.organizations[0].is_current // "null"' 2>/dev/null)
    
    missing_fields=()
    [[ "$effective_date" == "null" ]] && missing_fields+=("effective_date")
    [[ "$change_reason" == "null" ]] && missing_fields+=("change_reason")  
    [[ "$is_current" == "null" ]] && missing_fields+=("is_current")
    
    if [[ ${#missing_fields[@]} -eq 0 ]]; then
        log_test_result "时态字段完整性" "PASS" "所有时态字段正常返回"
    else
        log_test_result "时态字段完整性" "FAIL" "缺失字段: ${missing_fields[*]}"
    fi
else
    log_test_result "时态字段完整性" "FAIL" "$response"
fi

echo ""
echo "📡 测试6: 错误处理验证"
# 测试不存在的组织代码
response=$(api_request "${TEMPORAL_API_BASE}/9999999/temporal" "404")
if [[ "$response" != "HTTP_ERROR"* ]]; then
    error_code=$(echo "$response" | jq -r '.error_code // "null"' 2>/dev/null || echo "null")
    if [[ "$error_code" == "NOT_FOUND" ]]; then
        log_test_result "404错误处理" "PASS" "正确返回NOT_FOUND错误"
    else
        log_test_result "404错误处理" "FAIL" "错误码异常: $error_code"
    fi
else
    log_test_result "404错误处理" "FAIL" "$response"
fi

echo ""
echo "📡 测试7: 性能基准验证"
start_time=$(date +%s.%N)
for i in {1..5}; do
    api_request "${TEMPORAL_API_BASE}/${TEST_ORG_CODE}/temporal" >/dev/null
done
end_time=$(date +%s.%N)
avg_time=$(echo "scale=3; ($end_time - $start_time) / 5" | bc)

if (( $(echo "$avg_time < 1.0" | bc -l) )); then
    log_test_result "API性能基准" "PASS" "平均响应时间: ${avg_time}秒"
else
    log_test_result "API性能基准" "FAIL" "平均响应时间超标: ${avg_time}秒"
fi

echo ""
echo "📡 测试8: 健康检查端点"
health_response=$(api_request "http://localhost:9091/health")
if [[ "$health_response" != "HTTP_ERROR"* ]]; then
    service_status=$(echo "$health_response" | jq -r '.status // "null"' 2>/dev/null || echo "null")
    if [[ "$service_status" == "healthy" ]]; then
        log_test_result "健康检查端点" "PASS" "服务状态正常"
    else
        log_test_result "健康检查端点" "FAIL" "服务状态异常: $service_status"
    fi
else
    log_test_result "健康检查端点" "FAIL" "$health_response"
fi

# 汇总测试结果
echo ""
echo "🎉 时态API功能测试完成"
echo "📊 测试结果汇总:"

pass_count=0
fail_count=0

for result in "${TEST_RESULTS[@]}"; do
    status=$(echo "$result" | cut -d':' -f1)
    if [[ "$status" == "PASS" ]]; then
        ((pass_count++))
    else
        ((fail_count++))
    fi
done

total_count=$((pass_count + fail_count))
success_rate=$(( (pass_count * 100) / total_count ))

echo "  ✅ 通过: $pass_count/$total_count"
echo "  ❌ 失败: $fail_count/$total_count" 
echo "  📈 成功率: $success_rate%"

if [[ $fail_count -eq 0 ]]; then
    echo ""
    echo "🏆 所有API功能测试通过！时态管理系统运行正常。"
    exit 0
else
    echo ""
    echo "⚠️  发现 $fail_count 个测试失败，请检查相关功能。"
    exit 1
fi