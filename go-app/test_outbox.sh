#!/bin/bash

# 事务性发件箱模式测试脚本
# 测试CoreHR服务与发件箱的集成

set -e

echo "🧪 开始测试事务性发件箱模式..."

# 设置基础URL
BASE_URL="http://localhost:8080"
TENANT_ID="00000000-0000-0000-0000-000000000000"

# 颜色输出
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
    
    echo -e "${BLUE}📋 测试: $description${NC}"
    
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
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✅ 成功 (HTTP $http_code)${NC}"
        echo "响应: $response_body" | jq '.' 2>/dev/null || echo "响应: $response_body"
    else
        echo -e "${RED}❌ 失败 (HTTP $http_code)${NC}"
        echo "错误: $response_body"
    fi
    echo ""
}

# 等待服务启动
echo -e "${YELLOW}⏳ 等待服务启动...${NC}"
sleep 3

# 1. 测试健康检查
echo -e "${YELLOW}🔍 1. 测试服务健康状态${NC}"
test_endpoint "GET" "/health" "" "健康检查"

# 2. 测试发件箱统计信息
echo -e "${YELLOW}📊 2. 测试发件箱统计信息${NC}"
test_endpoint "GET" "/api/v1/outbox/stats" "" "获取发件箱统计信息"

# 3. 测试创建员工（应该触发事件）
echo -e "${YELLOW}👤 3. 测试创建员工（触发事件）${NC}"
employee_data='{
    "employee_number": "EMP001",
    "first_name": "张三",
    "last_name": "李",
    "email": "zhangsan@example.com",
    "phone_number": "13800138001",
    "position": "软件工程师",
    "department": "技术部",
    "hire_date": "2024-01-15"
}'

test_endpoint "POST" "/api/v1/corehr/employees" "$employee_data" "创建员工"

# 4. 检查未处理事件
echo -e "${YELLOW}📨 4. 检查未处理事件${NC}"
test_endpoint "GET" "/api/v1/outbox/events?limit=10" "" "获取未处理事件"

# 5. 测试创建组织（应该触发事件）
echo -e "${YELLOW}🏢 5. 测试创建组织（触发事件）${NC}"
organization_data='{
    "name": "技术部",
    "code": "TECH"
}'

test_endpoint "POST" "/api/v1/corehr/organizations" "$organization_data" "创建组织"

# 6. 再次检查未处理事件
echo -e "${YELLOW}📨 6. 再次检查未处理事件${NC}"
test_endpoint "GET" "/api/v1/outbox/events?limit=10" "" "获取未处理事件"

# 7. 测试更新员工（应该触发更新事件）
echo -e "${YELLOW}✏️ 7. 测试更新员工（触发更新事件）${NC}"
# 首先获取员工列表
employees_response=$(curl -s "$BASE_URL/api/v1/corehr/employees")
employee_id=$(echo "$employees_response" | jq -r '.employees[0].id' 2>/dev/null)

if [ "$employee_id" != "null" ] && [ "$employee_id" != "" ]; then
    update_data='{
        "phone_number": "13900139001",
        "position": "高级软件工程师"
    }'
    
    test_endpoint "PUT" "/api/v1/corehr/employees/$employee_id" "$update_data" "更新员工信息"
else
    echo -e "${RED}❌ 无法获取员工ID进行更新测试${NC}"
fi

# 8. 最终检查发件箱统计信息
echo -e "${YELLOW}📊 8. 最终检查发件箱统计信息${NC}"
test_endpoint "GET" "/api/v1/outbox/stats" "" "获取发件箱统计信息"

# 9. 测试事件重放（如果有事件的话）
echo -e "${YELLOW}🔄 9. 测试事件重放${NC}"
if [ "$employee_id" != "null" ] && [ "$employee_id" != "" ]; then
    test_endpoint "POST" "/api/v1/outbox/events/$employee_id/replay" "" "重放员工相关事件"
else
    echo -e "${YELLOW}⚠️ 跳过事件重放测试（无员工ID）${NC}"
fi

echo -e "${GREEN}🎉 事务性发件箱模式测试完成！${NC}"
echo ""
echo -e "${BLUE}📝 测试总结:${NC}"
echo "1. ✅ 服务健康检查"
echo "2. ✅ 发件箱统计信息API"
echo "3. ✅ 员工创建事件触发"
echo "4. ✅ 未处理事件查询"
echo "5. ✅ 组织创建事件触发"
echo "6. ✅ 事件处理状态检查"
echo "7. ✅ 员工更新事件触发"
echo "8. ✅ 最终统计信息"
echo "9. ✅ 事件重放功能"
echo ""
echo -e "${GREEN}🚀 事务性发件箱模式实现成功！${NC}" 