#!/bin/bash

# 获取JWT令牌
TOKEN=$(curl -s -X POST http://localhost:9090/auth/dev-token \
  -H "Content-Type: application/json" \
  -d '{"clientId": "dev-client", "userId": "dev-user"}' | jq -r '.data.token')

echo "Using token: ${TOKEN:0:50}..."

# 测试时态错误场景：在2020年创建组织，但选择2025年的上级组织
curl -X POST http://localhost:9090/api/v1/organization-units \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "测试时态错误处理部门",
    "unitType": "DEPARTMENT",
    "parentCode": "1000002",
    "effectiveDate": "2020-01-01",
    "description": "测试时态验证错误"
  }' | jq .