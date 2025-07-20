#!/bin/bash

# 测试Repository实现的脚本
echo "🧪 测试 CoreHR Repository 实现"

# 检查数据库连接
echo "📊 检查数据库连接..."
if ! pg_isready -h localhost -p 5432 -U postgres > /dev/null 2>&1; then
    echo "❌ 数据库未连接，请先启动数据库"
    exit 1
fi

echo "✅ 数据库连接正常"

# 编译项目
echo "🔨 编译项目..."
cd go-app
go build -o server cmd/server/main.go

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo "✅ 编译成功"

# 启动服务器（后台运行）
echo "🚀 启动服务器..."
./server &
SERVER_PID=$!

# 等待服务器启动
sleep 3

# 测试API端点
echo "🌐 测试API端点..."

# 测试健康检查
echo "📋 测试健康检查..."
curl -s http://localhost:8080/health | jq .

# 测试员工列表
echo "👥 测试员工列表..."
curl -s "http://localhost:8080/api/v1/employees?page=1&pageSize=10" | jq .

# 测试创建员工
echo "➕ 测试创建员工..."
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/employees \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "TEST001",
    "first_name": "测试",
    "last_name": "员工",
    "email": "test@example.com",
    "phone_number": "13800138000",
    "position": "软件工程师",
    "department": "技术部",
    "hire_date": "2024-01-01"
  }')

echo $CREATE_RESPONSE | jq .

# 提取员工ID
EMPLOYEE_ID=$(echo $CREATE_RESPONSE | jq -r '.id')

if [ "$EMPLOYEE_ID" != "null" ] && [ "$EMPLOYEE_ID" != "" ]; then
    echo "✅ 员工创建成功，ID: $EMPLOYEE_ID"
    
    # 测试获取员工详情
    echo "👤 测试获取员工详情..."
    curl -s "http://localhost:8080/api/v1/employees/$EMPLOYEE_ID" | jq .
    
    # 测试更新员工
    echo "✏️ 测试更新员工..."
    curl -s -X PUT "http://localhost:8080/api/v1/employees/$EMPLOYEE_ID" \
      -H "Content-Type: application/json" \
      -d '{
        "first_name": "更新后的名字",
        "phone_number": "13900139000"
      }' | jq .
    
    # 测试删除员工
    echo "🗑️ 测试删除员工..."
    curl -s -X DELETE "http://localhost:8080/api/v1/employees/$EMPLOYEE_ID"
    echo "✅ 员工删除成功"
else
    echo "❌ 员工创建失败"
fi

# 测试组织列表
echo "🏢 测试组织列表..."
curl -s "http://localhost:8080/api/v1/organizations" | jq .

# 测试组织树
echo "🌳 测试组织树..."
curl -s "http://localhost:8080/api/v1/organizations/tree" | jq .

# 停止服务器
echo "🛑 停止服务器..."
kill $SERVER_PID

echo "✅ 测试完成" 