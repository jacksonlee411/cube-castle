#!/bin/bash

echo "🧪 测试所有路由..."
echo "=================="

# 测试健康检查
echo "1. 测试健康检查..."
curl -s http://localhost:8080/health
echo -e "\n"

# 测试调试路由
echo "2. 测试调试路由..."
curl -s http://localhost:8080/debug/routes
echo -e "\n"

# 测试员工列表
echo "3. 测试员工列表..."
curl -s http://localhost:8080/api/v1/corehr/employees
echo -e "\n"

# 测试组织列表
echo "4. 测试组织列表..."
curl -s http://localhost:8080/api/v1/corehr/organizations
echo -e "\n"

# 测试组织树
echo "5. 测试组织树..."
curl -s http://localhost:8080/api/v1/corehr/organizations/tree
echo -e "\n"

# 测试静态文件
echo "6. 测试静态文件..."
curl -s http://localhost:8080/test.html
echo -e "\n"

echo "✅ 路由测试完成！" 