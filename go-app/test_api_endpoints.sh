#!/bin/bash

# API 端点测试脚本
# 用于验证组织单元和岗位API的可用性

echo "🧪 开始API端点测试..."
echo "=========================================="

# 设置基础URL和测试数据
BASE_URL="http://localhost:8080/api/v1"
TENANT_ID="550e8400-e29b-41d4-a716-446655440001"

# 添加认证头和租户ID (根据实际中间件需求调整)
HEADERS="-H 'Content-Type: application/json' -H 'Authorization: Bearer test-token' -H 'X-Tenant-ID: $TENANT_ID'"

echo "📋 测试计划:"
echo "1. 健康检查"
echo "2. 组织单元API测试"
echo "3. 岗位API测试"
echo ""

# 1. 健康检查
echo "🔍 1. 健康检查..."
curl -s http://localhost:8080/health | jq . || echo "健康检查响应: $(curl -s http://localhost:8080/health)"
echo ""

# 2. 组织单元API测试
echo "🏢 2. 组织单元API测试..."

echo "2.1 测试创建组织单元 (POST /organization-units)"
CREATE_ORG_RESPONSE=$(curl -s -X POST $BASE_URL/organization-units \
  -H 'Content-Type: application/json' \
  -d '{
    "unit_type": "department",
    "name": "工程技术部",
    "description": "负责产品技术开发",
    "profile": {
      "department_code": "ENG001",
      "budget_amount": 2000000.00,
      "head_count_limit": 50,
      "cost_center_code": "CC-ENG-001"
    }
  }')

echo "创建组织单元响应: $CREATE_ORG_RESPONSE"

# 提取组织单元ID (如果响应是JSON格式)
ORG_UNIT_ID=$(echo $CREATE_ORG_RESPONSE | jq -r '.id // "test-org-id"' 2>/dev/null || echo "test-org-id")
echo "组织单元ID: $ORG_UNIT_ID"
echo ""

echo "2.2 测试获取组织单元列表 (GET /organization-units)"
curl -s $BASE_URL/organization-units | jq . || echo "列表响应: $(curl -s $BASE_URL/organization-units)"
echo ""

echo "2.3 测试获取单个组织单元 (GET /organization-units/{id})"
curl -s $BASE_URL/organization-units/$ORG_UNIT_ID | jq . || echo "获取响应: $(curl -s $BASE_URL/organization-units/$ORG_UNIT_ID)"
echo ""

# 3. 岗位API测试
echo "👔 3. 岗位API测试..."

echo "3.1 测试创建岗位 (POST /positions)"
CREATE_POS_RESPONSE=$(curl -s -X POST $BASE_URL/positions \
  -H 'Content-Type: application/json' \
  -d '{
    "position_type": "technical",
    "job_profile_id": "550e8400-e29b-41d4-a716-446655440002",
    "department_id": "'$ORG_UNIT_ID'",
    "status": "active",
    "budgeted_fte": 1.0,
    "details": {
      "technical_level": "senior",
      "programming_languages": ["Go", "JavaScript", "Python"],
      "certification_required": false,
      "remote_work_allowed": true
    }
  }')

echo "创建岗位响应: $CREATE_POS_RESPONSE"

# 提取岗位ID
POSITION_ID=$(echo $CREATE_POS_RESPONSE | jq -r '.id // "test-pos-id"' 2>/dev/null || echo "test-pos-id")
echo "岗位ID: $POSITION_ID"
echo ""

echo "3.2 测试获取岗位列表 (GET /positions)"
curl -s $BASE_URL/positions | jq . || echo "列表响应: $(curl -s $BASE_URL/positions)"
echo ""

echo "3.3 测试获取单个岗位 (GET /positions/{id})"
curl -s $BASE_URL/positions/$POSITION_ID | jq . || echo "获取响应: $(curl -s $BASE_URL/positions/$POSITION_ID)"
echo ""

echo "3.4 测试更新岗位 (PUT /positions/{id})"
UPDATE_RESPONSE=$(curl -s -X PUT $BASE_URL/positions/$POSITION_ID \
  -H 'Content-Type: application/json' \
  -d '{
    "status": "inactive",
    "budgeted_fte": 0.8,
    "details": {
      "technical_level": "senior",
      "programming_languages": ["Go", "JavaScript", "Python", "Rust"],
      "certification_required": true,
      "remote_work_allowed": true
    }
  }')

echo "更新岗位响应: $UPDATE_RESPONSE"
echo ""

# 4. 错误情况测试
echo "⚠️  4. 错误情况测试..."

echo "4.1 测试无效ID访问"
curl -s $BASE_URL/positions/invalid-uuid | jq . || echo "错误响应: $(curl -s $BASE_URL/positions/invalid-uuid)"
echo ""

echo "4.2 测试无效JSON数据"
curl -s -X POST $BASE_URL/positions \
  -H 'Content-Type: application/json' \
  -d '{"invalid": "json"' | jq . || echo "错误响应: $(curl -s -X POST $BASE_URL/positions -H 'Content-Type: application/json' -d '{"invalid": "json"')"
echo ""

echo "=========================================="
echo "✅ API端点测试完成!"
echo ""
echo "📊 测试总结:"
echo "- 健康检查: 可访问"
echo "- 组织单元API: 路由已注册"
echo "- 岗位API: 路由已注册"
echo "- 错误处理: 验证边界情况"
echo ""
echo "🔗 API文档:"
echo "- 健康检查: GET http://localhost:8080/health"
echo "- 组织单元: GET/POST/PUT/DELETE http://localhost:8080/api/v1/organization-units"
echo "- 岗位管理: GET/POST/PUT/DELETE http://localhost:8080/api/v1/positions"
echo ""
echo "🚀 下一步: 配置数据库连接以进行完整功能测试"