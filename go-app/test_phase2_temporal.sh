#!/bin/bash

echo "🔍 测试第二阶段时态查询功能"
echo "================================"

# API 基本URL
API_BASE="http://localhost:8080"

echo "1. 测试健康检查..."
curl -s "$API_BASE/health" | jq '.'

echo -e "\n2. 测试API基本连通性..."
curl -s "$API_BASE/api/v1/ping" || echo "API ping 端点可能不存在，继续其他测试..."

echo -e "\n3. 插入测试数据..."

# 插入人员数据
echo "插入测试人员..."
PERSON_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/persons" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "email": "zhangsan@company.com",
    "employee_id": "EMP001"
  }' || echo '{"error": "person endpoint not available"}')

echo "人员响应: $PERSON_RESPONSE"

# 插入职位历史数据
echo -e "\n插入职位历史记录..."
POSITION_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/position-history" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_id": "EMP001",
    "organization_id": "ORG001",
    "position_title": "软件工程师",
    "department": "技术部",
    "effective_date": "2024-01-01T00:00:00Z",
    "salary_data": {
      "base_salary": 10000,
      "currency": "CNY"
    },
    "change_reason": "新员工入职"
  }' || echo '{"error": "position-history endpoint not available"}')

echo "职位响应: $POSITION_RESPONSE"

echo -e "\n4. 测试时态查询功能..."

# 测试当前职位查询
echo "查询当前职位..."
curl -s "$API_BASE/api/v1/position-history/current/EMP001" | jq '.' || echo "当前职位查询端点不可用"

# 测试历史职位查询
echo -e "\n查询职位历史..."
curl -s "$API_BASE/api/v1/position-history/timeline/EMP001" | jq '.' || echo "职位历史查询端点不可用"

# 测试特定时间点查询
echo -e "\n查询特定时间点职位..."
curl -s "$API_BASE/api/v1/position-history/as-of/EMP001?date=2024-06-01" | jq '.' || echo "特定时间点查询端点不可用"

echo -e "\n5. 测试第二阶段增强功能..."

# 测试批量时态查询
echo "测试批量查询..."
curl -s -X POST "$API_BASE/api/v1/position-history/batch-query" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_ids": ["EMP001"],
    "query_date": "2024-06-01T00:00:00Z"
  }' | jq '.' || echo "批量查询端点不可用"

# 测试性能指标
echo -e "\n测试性能指标..."
curl -s "$API_BASE/api/v1/metrics/temporal" | jq '.' || echo "性能指标端点不可用"

echo -e "\n6. 验证数据库直接查询..."
echo "数据库中的表："
docker exec cube_castle_postgres psql -U user -d cubecastle -c "\\dt"

echo -e "\n数据库中的数据："
docker exec cube_castle_postgres psql -U user -d cubecastle -c "SELECT COUNT(*) as person_count FROM person;"
docker exec cube_castle_postgres psql -U user -d cubecastle -c "SELECT COUNT(*) as position_history_count FROM position_history;"

echo -e "\n✅ 测试完成！"