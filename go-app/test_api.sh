#!/bin/bash

# CoreHR API 快速测试脚本
# 用于验证 API 功能是否正常

set -e

echo "🧪 CoreHR API 快速测试"
echo "======================"

# 默认配置
API_BASE="http://localhost:8080/api/v1/corehr"
TIMEOUT=10

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -n "测试 $description... "
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$API_BASE$endpoint" --max-time $TIMEOUT || echo "000")
    elif [ "$method" = "POST" ]; then
        response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X POST -H "Content-Type: application/json" -d "$data" "$API_BASE$endpoint" --max-time $TIMEOUT || echo "000")
    elif [ "$method" = "PUT" ]; then
        response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X PUT -H "Content-Type: application/json" -d "$data" "$API_BASE$endpoint" --max-time $TIMEOUT || echo "000")
    elif [ "$method" = "DELETE" ]; then
        response=$(curl -s -w "%{http_code}" -o /tmp/response.json -X DELETE "$API_BASE$endpoint" --max-time $TIMEOUT || echo "000")
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

# 检查服务器是否运行
echo "🔍 检查服务器状态..."
if ! curl -s --max-time 5 "http://localhost:8080/health" > /dev/null; then
    echo -e "${RED}❌ 服务器未运行，请先启动服务器${NC}"
    echo "   运行: ./start.sh 或 ./start.bat"
    exit 1
fi
echo -e "${GREEN}✅ 服务器正在运行${NC}"

echo ""
echo "开始 API 测试..."
echo ""

# 测试健康检查
test_endpoint "GET" "/health" "" "健康检查"

# 测试员工列表
test_endpoint "GET" "/employees" "" "获取员工列表"

# 测试组织列表
test_endpoint "GET" "/organizations" "" "获取组织列表"

# 测试组织树
test_endpoint "GET" "/organizations/tree" "" "获取组织树"

# 测试创建员工
employee_data='{
  "employee_number": "TEST001",
  "first_name": "测试",
  "last_name": "用户",
  "email": "test@example.com",
  "hire_date": "2024-01-15",
  "phone_number": "13800138000"
}'

test_endpoint "POST" "/employees" "$employee_data" "创建员工"

# 如果创建成功，获取员工ID并测试其他操作
if [ -f /tmp/response.json ]; then
    employee_id=$(cat /tmp/response.json | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    
    if [ ! -z "$employee_id" ]; then
        echo "   获取到员工ID: $employee_id"
        
        # 测试获取员工详情
        test_endpoint "GET" "/employees/$employee_id" "" "获取员工详情"
        
        # 测试更新员工
        update_data='{
          "first_name": "更新",
          "last_name": "姓名",
          "email": "updated@example.com"
        }'
        test_endpoint "PUT" "/employees/$employee_id" "$update_data" "更新员工"
        
        # 测试删除员工
        test_endpoint "DELETE" "/employees/$employee_id" "" "删除员工"
    fi
fi

echo ""
echo "🎉 测试完成！"
echo ""
echo "📋 测试结果说明："
echo "   ✅ 成功: API 端点正常工作"
echo "   ⚠️  警告: API 端点响应但状态码不是 2xx"
echo "   ❌ 失败: 无法连接到 API 端点"
echo ""
echo "🔗 更多测试："
echo "   - 打开浏览器访问: http://localhost:8080/test.html"
echo "   - 查看详细 API 文档: README_CoreHR.md" 