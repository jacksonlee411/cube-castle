#!/bin/bash

echo "🧪 测试 CoreHR API 路由"
echo "======================"

BASE_URL="http://localhost:8080"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local endpoint=$1
    local method=${2:-GET}
    local data=${3:-""}
    
    echo -n "测试 $method $endpoint ... "
    
    if [ "$method" = "POST" ] && [ -n "$data" ]; then
        response=$(curl -s -w "%{http_code}" -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" -o /tmp/response.json)
    else
        response=$(curl -s -w "%{http_code}" -X $method "$BASE_URL$endpoint" -o /tmp/response.json)
    fi
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        echo -e "${GREEN}✅ 成功 ($http_code)${NC}"
        if [ -f /tmp/response.json ]; then
            echo "   响应: $(cat /tmp/response.json | head -c 100)..."
        fi
    else
        echo -e "${RED}❌ 失败 ($http_code)${NC}"
        if [ -f /tmp/response.json ]; then
            echo "   错误: $(cat /tmp/response.json)"
        fi
    fi
}

echo "📍 服务器地址: $BASE_URL"
echo ""

# 测试健康检查
test_endpoint "/health"

# 测试调试路由
test_endpoint "/debug/routes"

# 测试静态文件
test_endpoint "/test.html"

# 测试 CoreHR API
test_endpoint "/api/v1/corehr/employees"

# 测试组织 API
test_endpoint "/api/v1/corehr/organizations"

# 测试组织树 API
test_endpoint "/api/v1/corehr/organizations/tree"

# 测试创建员工（POST 请求）
test_endpoint "/api/v1/corehr/employees" "POST" '{
    "employee_number": "EMP003",
    "first_name": "王",
    "last_name": "五",
    "email": "wangwu@example.com",
    "hire_date": "2023-03-15"
}'

echo ""
echo "🎉 路由测试完成！"
echo ""
echo "📋 如果所有测试都通过，您可以访问："
echo "   🌐 测试页面: $BASE_URL/test.html"
echo "   📊 调试路由: $BASE_URL/debug/routes"
echo "   🏥 健康检查: $BASE_URL/health"
echo "   👥 员工列表: $BASE_URL/api/v1/corehr/employees" 