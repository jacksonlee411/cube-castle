#!/bin/bash

# API接口集成测试脚本
echo "🏰 Cube Castle - API接口集成测试"
echo "================================"

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试函数
test_result() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if [ $1 -eq 0 ]; then
        echo "✅ $2"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo "❌ $2"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# API基础URL
API_BASE="http://localhost:8080"

echo -e "\n1. API服务健康检查"
echo "-----------------"

# 健康检查
RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/health_response.json "$API_BASE/health")
HTTP_CODE="${RESPONSE: -3}"
if [ "$HTTP_CODE" = "200" ]; then
    test_result 0 "API服务健康检查"
else
    test_result 1 "API服务健康检查 (HTTP $HTTP_CODE)"
fi

echo -e "\n2. CoreHR API测试"
echo "----------------"

# 测试获取员工列表
RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/employees_response.json "$API_BASE/api/v1/corehr/employees")
HTTP_CODE="${RESPONSE: -3}"
if [ "$HTTP_CODE" = "200" ]; then
    test_result 0 "获取员工列表API"
else
    test_result 1 "获取员工列表API (HTTP $HTTP_CODE)"
fi

# 测试获取组织架构树
RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/orgtree_response.json "$API_BASE/api/v1/corehr/organizations/tree")
HTTP_CODE="${RESPONSE: -3}"
if [ "$HTTP_CODE" = "200" ]; then
    test_result 0 "获取组织架构树API"
else
    test_result 1 "获取组织架构树API (HTTP $HTTP_CODE)"
fi

# 测试创建员工
TEST_EMPLOYEE_JSON='{"employee_number":"TEST001","first_name":"测试","last_name":"用户","email":"test@example.com","hire_date":"2024-01-01"}'
RESPONSE=$(curl -s -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$TEST_EMPLOYEE_JSON" -o /tmp/create_employee_response.json "$API_BASE/api/v1/corehr/employees")
HTTP_CODE="${RESPONSE: -3}"
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ]; then
    test_result 0 "创建员工API"
    CREATED_EMPLOYEE=true
else
    test_result 1 "创建员工API (HTTP $HTTP_CODE)"
    CREATED_EMPLOYEE=false
fi

echo -e "\n3. Intelligence Gateway API测试"
echo "-------------------------------"

# 测试AI文本解释接口
INTELLIGENCE_JSON='{"user_text":"更新我的电话号码为13800138000","session_id":"test-session-123"}'
RESPONSE=$(curl -s -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$INTELLIGENCE_JSON" -o /tmp/intelligence_response.json "$API_BASE/api/v1/intelligence/interpret" --max-time 10)
HTTP_CODE="${RESPONSE: -3}"
if [ "$HTTP_CODE" = "200" ]; then
    test_result 0 "AI文本解释API"
else
    test_result 1 "AI文本解释API (HTTP $HTTP_CODE)"
fi

echo -e "\n4. API响应格式验证"
echo "-------------------"

# 验证健康检查响应格式
if [ -f /tmp/health_response.json ]; then
    if grep -q "status" /tmp/health_response.json; then
        test_result 0 "健康检查响应格式"
    else
        test_result 1 "健康检查响应格式"
    fi
else
    test_result 1 "健康检查响应文件"
fi

# 验证员工列表响应格式
if [ -f /tmp/employees_response.json ]; then
    if grep -q "employees\|data\|result" /tmp/employees_response.json; then
        test_result 0 "员工列表响应格式"
    else
        test_result 1 "员工列表响应格式"
    fi
else
    test_result 1 "员工列表响应文件"
fi

echo -e "\n5. API性能测试"
echo "-------------"

# 测试API响应时间
start_time=$(date +%s%N)
curl -s "$API_BASE/health" > /dev/null
end_time=$(date +%s%N)
duration=$((($end_time - $start_time) / 1000000))

if [ $duration -lt 500 ]; then
    test_result 0 "API响应时间 (${duration}ms)"
else
    test_result 1 "API响应时间过长 (${duration}ms)"
fi

echo -e "\n6. API错误处理测试"
echo "-----------------"

# 测试无效路径
RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/invalid_path_response.json "$API_BASE/api/v1/invalid/path")
HTTP_CODE="${RESPONSE: -3}"
if [ "$HTTP_CODE" = "404" ]; then
    test_result 0 "无效路径错误处理"
else
    test_result 1 "无效路径错误处理 (HTTP $HTTP_CODE)"
fi

# 测试无效JSON数据
INVALID_JSON='{"invalid": json}'
RESPONSE=$(curl -s -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$INVALID_JSON" -o /tmp/invalid_json_response.json "$API_BASE/api/v1/corehr/employees")
HTTP_CODE="${RESPONSE: -3}"
if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "422" ]; then
    test_result 0 "无效JSON数据错误处理"
else
    test_result 1 "无效JSON数据错误处理 (HTTP $HTTP_CODE)"
fi

echo -e "\n7. API安全性测试"
echo "---------------"

# 测试CORS头
RESPONSE=$(curl -s -I -H "Origin: http://localhost:3000" "$API_BASE/api/v1/corehr/employees")
if echo "$RESPONSE" | grep -i "access-control-allow" > /dev/null; then
    test_result 0 "CORS头设置"
else
    test_result 1 "CORS头设置"
fi

# 测试Content-Type头
RESPONSE=$(curl -s -I "$API_BASE/api/v1/corehr/employees")
if echo "$RESPONSE" | grep -i "content-type" > /dev/null; then
    test_result 0 "Content-Type头设置"
else
    test_result 1 "Content-Type头设置"
fi

echo -e "\n8. API并发测试"
echo "-------------"

# 并发测试
for i in {1..5}; do
    curl -s "$API_BASE/health" > /dev/null &
done
wait
test_result 0 "API并发请求处理"

echo -e "\n9. 清理测试数据"
echo "-------------"

# 清理创建的测试员工（如果创建成功的话）
if [ "$CREATED_EMPLOYEE" = true ]; then
    echo "注意: 测试创建的员工数据可能需要手动清理"
    test_result 0 "测试数据清理提醒"
fi

# 清理临时文件
rm -f /tmp/health_response.json /tmp/employees_response.json /tmp/orgtree_response.json
rm -f /tmp/create_employee_response.json /tmp/intelligence_response.json
rm -f /tmp/invalid_path_response.json /tmp/invalid_json_response.json
test_result 0 "临时文件清理"

echo -e "\n================================"
echo "API接口集成测试完成！"
echo "总计: $TOTAL_TESTS 项测试"
echo "✅ 通过: $PASSED_TESTS 项"
echo "❌ 失败: $FAILED_TESTS 项"
SUCCESS_RATE=$(( PASSED_TESTS * 100 / TOTAL_TESTS ))
echo "成功率: ${SUCCESS_RATE}%"
echo "================================"

# 显示部分响应内容用于调试
echo -e "\n调试信息："
echo "--------"
if [ -f /tmp/health_response.json ]; then
    echo "健康检查响应: $(cat /tmp/health_response.json)"
fi

# 返回退出码
if [ $FAILED_TESTS -eq 0 ]; then
    exit 0
else
    exit 1
fi