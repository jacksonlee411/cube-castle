#!/bin/bash

echo "🔧 测试 CoreHR API 404 错误修复"
echo "================================"

BASE_URL="http://localhost:8080"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "开始测试..."
echo ""

# 测试基础路由
echo "1. 测试健康检查..."
if curl -s "$BASE_URL/health" > /dev/null; then
    echo -e "${GREEN}✅ 健康检查正常${NC}"
else
    echo -e "${RED}❌ 健康检查失败${NC}"
fi

echo ""
echo "2. 测试调试路由..."
if curl -s "$BASE_URL/debug/routes" > /dev/null; then
    echo -e "${GREEN}✅ 调试路由正常${NC}"
else
    echo -e "${RED}❌ 调试路由失败${NC}"
fi

echo ""
echo "3. 测试 CoreHR API 路由..."
if curl -s "$BASE_URL/api/v1/corehr/employees" > /dev/null; then
    echo -e "${GREEN}✅ 员工列表 API 正常${NC}"
else
    echo -e "${RED}❌ 员工列表 API 失败${NC}"
fi

echo ""
echo "4. 测试组织 API..."
if curl -s "$BASE_URL/api/v1/corehr/organizations" > /dev/null; then
    echo -e "${GREEN}✅ 组织列表 API 正常${NC}"
else
    echo -e "${RED}❌ 组织列表 API 失败${NC}"
fi

echo ""
echo "5. 测试组织树 API..."
if curl -s "$BASE_URL/api/v1/corehr/organizations/tree" > /dev/null; then
    echo -e "${GREEN}✅ 组织树 API 正常${NC}"
else
    echo -e "${RED}❌ 组织树 API 失败${NC}"
fi

echo ""
echo "6. 测试静态文件..."
if curl -s "$BASE_URL/test.html" > /dev/null; then
    echo -e "${GREEN}✅ 测试页面正常${NC}"
else
    echo -e "${RED}❌ 测试页面失败${NC}"
fi

echo ""
echo "🎉 测试完成！"
echo ""
echo "如果所有测试都通过，说明 404 错误已修复。"
echo "现在您可以："
echo "1. 访问 http://localhost:8080/test.html 进行完整测试"
echo "2. 使用 API 端点进行开发"
echo "3. 查看 http://localhost:8080/debug/routes 了解所有可用路由" 