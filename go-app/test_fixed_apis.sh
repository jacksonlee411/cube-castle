#!/bin/bash

# 测试修复后的API
# 使用方法: ./test_fixed_apis.sh

set -e

echo "🧪 测试修复后的API"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

# 函数：测试API
test_api() {
    local name="$1"
    local method="$2"
    local url="$3"
    local data="$4"
    
    echo -e "${BLUE}🔍 测试: $name${NC}"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL$url")
    elif [ "$method" = "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST -H "Content-Type: application/json" -d "$data" "$BASE_URL$url")
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

# 等待服务启动
echo -e "${YELLOW}⏳ 等待服务启动...${NC}"
sleep 3

# 测试健康检查
test_api "健康检查" "GET" "/health"

# 测试数据库连接
test_api "数据库连接" "GET" "/health/db"

# 测试发件箱统计
test_api "发件箱统计" "GET" "/api/v1/outbox/stats"

# 测试发件箱事件
test_api "发件箱事件" "GET" "/api/v1/outbox/events"

# 测试未处理事件
test_api "未处理事件" "GET" "/api/v1/outbox/unprocessed"

# 测试员工列表
test_api "员工列表" "GET" "/api/v1/corehr/employees"

# 测试创建员工
EMPLOYEE_DATA='{"employee_number":"EMP002","first_name":"Jane","last_name":"Smith","email":"jane.smith@example.com","phone_number":"+1234567891","hire_date":"2024-02-15"}'
test_api "创建员工" "POST" "/api/v1/corehr/employees" "$EMPLOYEE_DATA"

# 获取员工ID用于后续测试
echo -e "${BLUE}🔍 获取员工ID用于测试...${NC}"
employees_response=$(curl -s "$BASE_URL/api/v1/corehr/employees")
employee_id=$(echo "$employees_response" | jq -r '.employees[0].id' 2>/dev/null)

if [ "$employee_id" != "null" ] && [ "$employee_id" != "" ]; then
    echo -e "${GREEN}✅ 找到员工ID: $employee_id${NC}"
    
    # 测试获取员工详情
    test_api "获取员工详情" "GET" "/api/v1/corehr/employees/$employee_id"
    
    # 测试更新员工
    UPDATE_DATA='{"first_name":"Jane Updated","phone_number":"+1234567899"}'
    test_api "更新员工" "POST" "/api/v1/corehr/employees/$employee_id" "$UPDATE_DATA"
    
    # 测试事件重放
    REPLAY_DATA="{\"aggregate_id\":\"$employee_id\"}"
    test_api "事件重放" "POST" "/api/v1/outbox/replay" "$REPLAY_DATA"
    
else
    echo -e "${YELLOW}⚠️  没有找到员工ID，跳过相关测试${NC}"
fi

echo -e "${GREEN}🎉 API测试完成${NC}" 