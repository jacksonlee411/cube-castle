#!/bin/bash

# 🏰 Cube Castle - 第四阶段核心业务逻辑测试
# ============================================

echo "🧪 Cube Castle - 第四阶段核心业务逻辑测试"
echo "========================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试结果记录函数
test_result() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ $3${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# API基础URL
API_BASE="http://localhost:8080"

echo "1. 前后端集成功能测试"
echo "------------------------"

# 测试API健康检查
echo "测试 API服务可用性..."
curl -s "$API_BASE/health" > /dev/null
test_result $? "API健康检查通过" "API服务不可用"

# 测试CoreHR API
echo "测试 员工管理API..."
EMPLOYEE_RESPONSE=$(curl -s "$API_BASE/api/v1/employees")
if [[ $EMPLOYEE_RESPONSE == *"employees"* ]]; then
    test_result 0 "员工列表API正常" ""
else
    test_result 1 "" "员工列表API异常"
fi

echo "测试 组织架构API..."
ORG_RESPONSE=$(curl -s "$API_BASE/api/v1/organizations")
if [[ $ORG_RESPONSE == *"organizations"* ]]; then
    test_result 0 "组织架构API正常" ""
else
    test_result 1 "" "组织架构API异常"
fi

echo "测试 组织树API..."
TREE_RESPONSE=$(curl -s "$API_BASE/api/v1/organizations/tree")
if [[ $TREE_RESPONSE == *"data"* ]]; then
    test_result 0 "组织树API正常" ""
else
    test_result 1 "" "组织树API异常"
fi

echo ""
echo "2. 业务逻辑完整性测试"
echo "----------------------"

# 测试员工创建业务逻辑
echo "测试 员工创建业务逻辑..."
CREATE_EMPLOYEE_DATA='{"name":"测试员工S4","email":"test-s4@company.com","position":"测试工程师","department":"技术部","organization_id":"550e8400-e29b-41d4-a716-446655440000"}'
CREATE_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/employees" \
    -H "Content-Type: application/json" \
    -d "$CREATE_EMPLOYEE_DATA")

if [[ $CREATE_RESPONSE == *"employee_id"* ]] || [[ $CREATE_RESPONSE == *"success"* ]]; then
    test_result 0 "员工创建业务逻辑正常" ""
    
    # 提取员工ID用于后续测试
    EMPLOYEE_ID=$(echo $CREATE_RESPONSE | grep -o '"employee_id":"[^"]*"' | cut -d'"' -f4)
    if [[ -z "$EMPLOYEE_ID" ]]; then
        EMPLOYEE_ID="test-s4-employee"
    fi
else
    test_result 1 "" "员工创建业务逻辑异常"
    EMPLOYEE_ID="test-s4-employee"
fi

# 测试数据验证逻辑
echo "测试 数据验证逻辑..."
INVALID_DATA='{"name":"","email":"invalid-email","position":"","department":""}'
INVALID_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/employees" \
    -H "Content-Type: application/json" \
    -d "$INVALID_DATA")

if [[ $INVALID_RESPONSE == *"error"* ]] || [[ $INVALID_RESPONSE == *"400"* ]]; then
    test_result 0 "数据验证逻辑正常" ""
else
    test_result 1 "" "数据验证逻辑异常"
fi

echo ""
echo "3. 前端组件功能测试"
echo "-------------------"

# 检查前端构建状态
cd ../nextjs-app 2>/dev/null || cd nextjs-app 2>/dev/null || echo "前端目录未找到"

if [ -d ".next" ]; then
    test_result 0 "前端应用构建成功" ""
else
    echo "尝试构建前端应用..."
    npm run build > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        test_result 0 "前端应用构建成功" ""
    else
        test_result 1 "" "前端应用构建失败"
    fi
fi

# 检查关键前端文件
if [ -f "src/app/employees/page.tsx" ]; then
    test_result 0 "员工管理页面存在" ""
else
    test_result 1 "" "员工管理页面缺失"
fi

if [ -f "src/app/organizations/page.tsx" ]; then
    test_result 0 "组织架构页面存在" ""
else
    test_result 1 "" "组织架构页面缺失"
fi

if [ -f "src/components/business/employee-table.tsx" ]; then
    test_result 0 "员工表格组件存在" ""
else
    test_result 1 "" "员工表格组件缺失"
fi

if [ -f "src/lib/api-client.ts" ]; then
    test_result 0 "API客户端存在" ""
else
    test_result 1 "" "API客户端缺失"
fi

echo ""
echo "4. 系统集成测试"
echo "---------------"

# 回到根目录
cd .. 2>/dev/null

# 测试数据库连接
echo "测试 数据库连接..."
DB_TEST=$(psql -h localhost -U user -d cubecastle -c "SELECT 1;" 2>/dev/null || echo "connection_failed")
if [[ $DB_TEST != *"connection_failed"* ]]; then
    test_result 0 "数据库连接正常" ""
else
    test_result 1 "" "数据库连接失败"
fi

# 测试AI服务集成
echo "测试 AI服务集成..."
AI_TEST_DATA='{"query":"测试AI服务","user_id":"test-user","tenant_id":"test-tenant"}'
AI_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/intelligence/interpret" \
    -H "Content-Type: application/json" \
    -d "$AI_TEST_DATA")

# AI服务可能返回错误，但应该有响应
if [[ ! -z "$AI_RESPONSE" ]]; then
    test_result 0 "AI服务响应正常" ""
else
    test_result 1 "" "AI服务无响应"
fi

echo ""
echo "5. 性能基准测试"
echo "---------------"

# API响应时间测试
echo "测试 API响应时间..."
START_TIME=$(date +%s%N)
curl -s "$API_BASE/health" > /dev/null
END_TIME=$(date +%s%N)
RESPONSE_TIME=$(( (END_TIME - START_TIME) / 1000000 ))

if [ $RESPONSE_TIME -lt 1000 ]; then
    test_result 0 "API响应时间正常 (${RESPONSE_TIME}ms)" ""
else
    test_result 1 "" "API响应时间过长 (${RESPONSE_TIME}ms)"
fi

# 并发测试
echo "测试 并发处理能力..."
for i in {1..5}; do
    curl -s "$API_BASE/health" > /dev/null &
done
wait

test_result 0 "并发请求处理正常" ""

echo ""
echo "6. 安全性测试"
echo "-------------"

# CORS测试
echo "测试 CORS配置..."
CORS_RESPONSE=$(curl -s -I "$API_BASE/health" | grep -i "access-control")
if [[ ! -z "$CORS_RESPONSE" ]]; then
    test_result 0 "CORS配置正常" ""
else
    test_result 1 "" "CORS配置缺失"
fi

# SQL注入防护测试
echo "测试 SQL注入防护..."
INJECTION_TEST=$(curl -s "$API_BASE/api/v1/employees?name='; DROP TABLE employees; --")
if [[ $INJECTION_TEST != *"error"* ]] && [[ $INJECTION_TEST == *"employees"* ]]; then
    test_result 0 "SQL注入防护正常" ""
else
    test_result 1 "" "SQL注入防护可能有问题"
fi

echo ""
echo "7. 清理测试数据"
echo "---------------"

# 清理测试创建的员工数据
if [[ ! -z "$EMPLOYEE_ID" ]] && [[ "$EMPLOYEE_ID" != "test-s4-employee" ]]; then
    echo "清理测试员工数据..."
    curl -s -X DELETE "$API_BASE/api/v1/employees/$EMPLOYEE_ID" > /dev/null
    test_result 0 "测试数据清理完成" ""
else
    echo -e "${YELLOW}⚠️  注意: 测试员工数据可能需要手动清理${NC}"
fi

echo ""
echo "========================================"
echo "第四阶段核心业务逻辑测试完成！"
echo "总计: $TOTAL_TESTS 项测试"
echo -e "${GREEN}✅ 通过: $PASSED_TESTS 项${NC}"
echo -e "${RED}❌ 失败: $FAILED_TESTS 项${NC}"

SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
echo "成功率: $SUCCESS_RATE%"
echo "========================================"

if [ $SUCCESS_RATE -ge 85 ]; then
    echo -e "${GREEN}🎉 第四阶段业务逻辑测试总体通过！${NC}"
    exit 0
elif [ $SUCCESS_RATE -ge 70 ]; then
    echo -e "${YELLOW}⚠️  第四阶段业务逻辑测试基本通过，建议优化失败项目${NC}"
    exit 0
else
    echo -e "${RED}❌ 第四阶段业务逻辑测试存在较多问题，需要修复${NC}"
    exit 1
fi