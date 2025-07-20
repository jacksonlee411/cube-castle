#!/bin/bash

# Cube Castle 1.1.1 验证测试脚本
# 测试所有API功能

echo "🏰 Cube Castle 1.1.1 验证测试开始"
echo "=================================="

BASE_URL="http://localhost:8080"
TEST_EMPLOYEE_ID=""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "${BLUE}测试: $description${NC}"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$endpoint")
    elif [ "$method" = "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -d "$data" "$BASE_URL$endpoint")
    elif [ "$method" = "PUT" ]; then
        response=$(curl -s -w "\n%{http_code}" -X PUT -H "Content-Type: application/json" -d "$data" "$BASE_URL$endpoint")
    elif [ "$method" = "DELETE" ]; then
        response=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL$endpoint")
    fi
    
    # 分离响应体和状态码
    body=$(echo "$response" | head -n -1)
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" -ge 200 ] && [ "$status_code" -lt 300 ]; then
        echo -e "${GREEN}✅ 成功 ($status_code)${NC}"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        echo -e "${RED}❌ 失败 ($status_code)${NC}"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    fi
    echo ""
}

# 1. 基础服务测试
echo -e "${YELLOW}=== 基础服务测试 ===${NC}"

test_endpoint "GET" "/health" "" "健康检查"
test_endpoint "GET" "/health/db" "" "数据库连接"
test_endpoint "GET" "/api/v1/outbox/stats" "" "发件箱统计"

# 2. 员工管理测试
echo -e "${YELLOW}=== 员工管理测试 ===${NC}"

test_endpoint "GET" "/api/v1/corehr/employees" "" "获取员工列表"

# 创建员工
echo -e "${BLUE}测试: 创建员工${NC}"
timestamp=$(date +%s)
employee_data="{\"employee_number\":\"EMP$timestamp\",\"first_name\":\"测试\",\"last_name\":\"用户\",\"email\":\"test$timestamp@example.com\",\"hire_date\":\"2024-01-15\"}"

response=$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -d "$employee_data" "$BASE_URL/api/v1/corehr/employees")
body=$(echo "$response" | head -n -1)
status_code=$(echo "$response" | tail -n 1)

if [ "$status_code" -ge 200 ] && [ "$status_code" -lt 300 ]; then
    echo -e "${GREEN}✅ 创建员工成功 ($status_code)${NC}"
    echo "$body" | jq . 2>/dev/null || echo "$body"
    # 提取员工ID
    TEST_EMPLOYEE_ID=$(echo "$body" | jq -r '.id' 2>/dev/null)
    echo -e "${BLUE}测试员工ID: $TEST_EMPLOYEE_ID${NC}"
else
    echo -e "${RED}❌ 创建员工失败 ($status_code)${NC}"
    echo "$body" | jq . 2>/dev/null || echo "$body"
fi
echo ""

# 如果有员工ID，继续测试其他操作
if [ -n "$TEST_EMPLOYEE_ID" ]; then
    test_endpoint "GET" "/api/v1/corehr/employees/$TEST_EMPLOYEE_ID" "" "获取员工详情"
    
    update_data="{\"first_name\":\"测试更新\",\"email\":\"updated$timestamp@example.com\"}"
    test_endpoint "PUT" "/api/v1/corehr/employees/$TEST_EMPLOYEE_ID" "$update_data" "更新员工"
    
    test_endpoint "DELETE" "/api/v1/corehr/employees/$TEST_EMPLOYEE_ID" "" "删除员工"
else
    echo -e "${RED}⚠️  跳过员工详情、更新、删除测试（创建员工失败）${NC}"
fi

# 3. 发件箱测试
echo -e "${YELLOW}=== 发件箱测试 ===${NC}"

test_endpoint "GET" "/api/v1/outbox/events" "" "查看所有事件"
test_endpoint "GET" "/api/v1/outbox/unprocessed" "" "查看未处理事件"

# 事件重放测试（使用一个存在的aggregate_id）
if [ -n "$TEST_EMPLOYEE_ID" ]; then
    replay_data="{\"aggregate_id\":\"$TEST_EMPLOYEE_ID\"}"
    test_endpoint "POST" "/api/v1/outbox/replay" "$replay_data" "事件重放"
else
    echo -e "${RED}⚠️  跳过事件重放测试（没有员工ID）${NC}"
fi

# 4. 最终验证
echo -e "${YELLOW}=== 最终验证 ===${NC}"

test_endpoint "GET" "/api/v1/corehr/employees" "" "最终员工列表"
test_endpoint "GET" "/api/v1/outbox/stats" "" "最终发件箱统计"

echo -e "${GREEN}🏰 Cube Castle 1.1.1 验证测试完成${NC}"
echo "==================================" 