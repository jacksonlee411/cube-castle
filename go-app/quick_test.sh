#!/bin/bash

echo "🔍 CoreHR API 快速路由测试"
echo "=========================="

BASE_URL="http://localhost:8080"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 测试函数
test_route() {
    local method=$1
    local endpoint=$2
    local description=$3
    
    echo -n "测试 $description... "
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$BASE_URL$endpoint" --max-time 5 2>/dev/null || echo "000")
    fi
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "000" ]; then
        echo -e "${RED}❌ 连接失败${NC}"
        return 1
    elif [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✅ 成功 ($http_code)${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  警告 ($http_code)${NC}"
        if [ -f /tmp/response.json ]; then
            echo "   响应: $(cat /tmp/response.json)"
        fi
        return 0
    fi
}

echo "开始测试..."
echo ""

# 测试基础路由
test_route "GET" "/health" "健康检查"
test_route "GET" "/debug/routes" "调试路由"

echo ""
echo "测试 CoreHR API 路由..."
echo ""

# 测试 CoreHR API 路由
test_route "GET" "/api/v1/corehr/employees" "员工列表"
test_route "GET" "/api/v1/corehr/organizations" "组织列表"
test_route "GET" "/api/v1/corehr/organizations/tree" "组织树"

echo ""
echo "测试静态文件..."
echo ""

# 测试静态文件
test_route "GET" "/test.html" "测试页面"

echo ""
echo "🎉 路由测试完成！"
echo ""
echo "如果看到 ❌ 连接失败，请确保："
echo "1. 服务器正在运行 (go run cmd/server/main.go)"
echo "2. 端口 8080 没有被其他程序占用"
echo "3. 防火墙没有阻止连接"
echo ""
echo "如果看到 ⚠️  警告，请检查："
echo "1. 数据库连接是否正常"
echo "2. 服务是否完全启动" 