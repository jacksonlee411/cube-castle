# Cube Castle API cURL 测试示例（开发环境）

## 环境配置

```bash
# 设置基础URL变量
export COMMAND_SERVICE="http://localhost:9090"
export QUERY_SERVICE="http://localhost:8090"
export TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"  # 默认租户（与后端一致）
```
> ⚠️ 端口提示：上述 `localhost` 端点均为 Docker 容器映射结果，宿主机不得安装同名服务占用端口；如检测到冲突，请卸载宿主服务并重新执行 `make docker-up`。

## JWT令牌管理

### 1. 生成开发JWT令牌

```bash
# 生成8小时有效期的JWT令牌（开发模式需开启 DEV_MODE=true）
curl -X POST "${COMMAND_SERVICE}/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "dev-user",
    "tenantId": "'${TENANT_ID}'",
    "roles": ["ADMIN", "USER"],
    "duration": "8h"
  }'

# 提取JWT令牌到环境变量 (需要jq工具)
export JWT_TOKEN=$(curl -s -X POST "${COMMAND_SERVICE}/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "dev-user",
    "tenantId": "'${TENANT_ID}'",
    "roles": ["ADMIN", "USER"],
    "duration": "8h"
  }' | jq -r '.data.token')

echo "JWT令牌: ${JWT_TOKEN}"
```

### 2. 验证令牌信息

```bash
# 获取令牌详细信息
curl -X GET "${COMMAND_SERVICE}/auth/dev-token/info" \
  -H "Authorization: Bearer ${JWT_TOKEN}"
```

### 3. 开发工具状态

```bash
# 检查开发环境状态
curl -X GET "${COMMAND_SERVICE}/dev/status"

# 获取测试端点列表  
curl -X GET "${COMMAND_SERVICE}/dev/test-endpoints"
```

## REST API 命令操作

### 1. 健康检查

```bash
# 命令服务健康检查
curl -X GET "${COMMAND_SERVICE}/health"

# 查询服务健康检查（如可用）
curl -X GET "${QUERY_SERVICE}/health"
```

### 2. 创建组织单元

```bash
# 创建根级部门（unitType: DEPARTMENT|ORGANIZATION_UNIT|PROJECT_TEAM）
curl -X POST "${COMMAND_SERVICE}/api/v1/organization-units" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "name": "技术部",
    "unitType": "DEPARTMENT",
    "parentCode": null,
    "description": "负责技术研发工作",
    "sortOrder": 1,
    "effectiveDate": "2025-08-25",
    "isTemporal": false
  }'

# 创建子部门 (需要先获取父部门code)
export PARENT_CODE="1000000"  # 根据你的环境替换为实际父code（如根节点）
curl -X POST "${COMMAND_SERVICE}/api/v1/organization-units" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "name": "前端开发组",
    "unitType": "DEPARTMENT",
    "parentCode": "'${PARENT_CODE}'",
    "description": "负责前端应用开发",
    "sortOrder": 1,
    "effectiveDate": "2025-08-25", 
    "isTemporal": false
  }'
```

### 3. 更新组织单元

```bash
# 更新组织信息（PUT 语义为完整替换，时态字段不可改）
export ORG_CODE="1000006"  # 替换为实际code（从创建响应中获取）
curl -X PUT "${COMMAND_SERVICE}/api/v1/organization-units/${ORG_CODE}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "name": "技术研发部",
    "description": "负责产品技术研发和架构设计",
    "sortOrder": 2
  }'
```

### 4. 组织状态管理

```bash
# 停用组织单元（业务操作，插入一个INACTIVE版本）
curl -X POST "${COMMAND_SERVICE}/api/v1/organization-units/${ORG_CODE}/suspend" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "operationReason": "部门重组-计划暂停", # 可选字段
    "effectiveDate": "2025-12-01"
  }'

# 重新激活组织单元（业务操作，插入一个ACTIVE版本）
curl -X POST "${COMMAND_SERVICE}/api/v1/organization-units/${ORG_CODE}/activate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "operationReason": "恢复业务运营", # 可选字段
    "effectiveDate": "2026-01-10"
  }'

## 时态版本管理（推荐）

### 1) 新增版本（插入中间版本并自动桥接前后边界）

```bash
curl -X POST "${COMMAND_SERVICE}/api/v1/organization-units/${ORG_CODE}/versions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "name": "技术部-历史更名",
    "unitType": "DEPARTMENT",
    "effectiveDate": "2025-09-01",
    "operationReason": "历史更名-数据补正" # 可选字段
  }'
```

### 2) 作废版本（推荐删除姿势，单事务“软删+全链重算”）

```bash
# 先查询要作废版本的recordId（可通过API响应或数据库查询获得）
export RECORD_ID="<uuid>"

curl -X POST "${COMMAND_SERVICE}/api/v1/organization-units/${ORG_CODE}/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "eventType": "DEACTIVATE",
    "recordId": "'${RECORD_ID}'",
    "changeReason": "版本作废-数据纠正"
  }'
```
```

## GraphQL 查询操作

### 1. 组织统计查询

```bash
# 获取组织统计信息
curl -X POST "${QUERY_SERVICE}/graphql" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "query": "query OrganizationStats { organizationStats { totalCount temporalStats { totalVersions averageVersionsPerOrg oldestEffectiveDate newestEffectiveDate } byType { unitType count percentage } } }"
  }'
```

### 2. 组织列表查询

```bash
# 分页查询活跃组织
curl -X POST "${QUERY_SERVICE}/graphql" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "query": "query Organizations($filter: OrganizationFilter, $pagination: PaginationInput) { organizations(filter: $filter, pagination: $pagination) { nodes { code name unitType status level effectiveDate isCurrent } pagination { total hasNext hasPrevious } } }",
    "variables": {
      "filter": {
        "status": "ACTIVE"
      },
      "pagination": {
        "limit": 10,
        "offset": 0
      }
    }
  }'

# 查询特定类型组织
curl -X POST "${QUERY_SERVICE}/graphql" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "query": "query Organizations($filter: OrganizationFilter) { organizations(filter: $filter) { nodes { code name unitType level parentCode } } }",
    "variables": {
      "filter": {
        "unitType": "DEPARTMENT",
        "status": "ACTIVE"
      }
    }
  }'
```

### 3. 单个组织查询

```bash
# 查询指定组织详细信息
export ORG_CODE="TECH001"  # 替换为实际code
curl -X POST "${QUERY_SERVICE}/graphql" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "query": "query Organization($code: String!) { organization(code: $code) { code name unitType status level path parentCode effectiveDate endDate isCurrent description createdAt updatedAt } }",
    "variables": {
      "code": "'${ORG_CODE}'"
    }
  }'
```

## 完整测试流程

```bash
#!/bin/bash
# 完整API测试流程脚本

set -e  # 遇到错误立即退出

echo "🚀 开始Cube Castle API测试流程"

# 1. 环境设置
export COMMAND_SERVICE="http://localhost:9090" 
export QUERY_SERVICE="http://localhost:8090"
export TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"

# 2. 生成JWT令牌
echo "📝 生成JWT令牌..."
export JWT_TOKEN=$(curl -s -X POST "${COMMAND_SERVICE}/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "dev-user",
    "tenantId": "'${TENANT_ID}'", 
    "roles": ["ADMIN", "USER"],
    "duration": "8h"
  }' | jq -r '.data.token')

if [ "$JWT_TOKEN" = "null" ] || [ -z "$JWT_TOKEN" ]; then
  echo "❌ JWT令牌生成失败"
  exit 1
fi

echo "✅ JWT令牌生成成功: ${JWT_TOKEN:0:20}..."

# 3. 检查服务健康状态
echo "🔍 检查服务健康状态..."
curl -s "${COMMAND_SERVICE}/health" | jq '.'
curl -s "${QUERY_SERVICE}/health" | jq '.'

# 4. 查询组织统计
echo "📊 查询组织统计信息..."
curl -s -X POST "${QUERY_SERVICE}/graphql" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "query": "query { organizationStats { totalCount } }"
  }' | jq '.'

# 5. 创建测试组织
echo "🏢 创建测试组织..."
ORG_RESPONSE=$(curl -s -X POST "${COMMAND_SERVICE}/api/v1/organization-units" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "name": "API测试部门",
    "unitType": "DEPARTMENT",
    "description": "用于API测试的部门",
    "sortOrder": 999,
    "effectiveDate": "2025-08-25",
    "isTemporal": false
  }')

export TEST_ORG_CODE=$(echo "$ORG_RESPONSE" | jq -r '.data.code')
echo "✅ 测试组织创建成功: ${TEST_ORG_CODE}"

# 6. 查询创建的组织
echo "🔎 查询创建的组织..."
curl -s -X POST "${QUERY_SERVICE}/graphql" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "query": "query Organization($code: String!) { organization(code: $code) { code name status } }",
    "variables": {"code": "'${TEST_ORG_CODE}'"}
  }' | jq '.'

# 7. 更新组织
echo "📝 更新组织信息..."
curl -s -X PUT "${COMMAND_SERVICE}/api/v1/organization-units/${TEST_ORG_CODE}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{
    "name": "API测试部门(已更新)",
    "description": "更新后的API测试部门"
  }' | jq '.'

# 8. 清理测试数据
echo "🧹 清理测试数据..."
curl -s -X DELETE "${COMMAND_SERVICE}/api/v1/organization-units/${TEST_ORG_CODE}" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" | jq '.'

echo "✅ API测试流程完成!"
```

## 常见问题处理

### JWT令牌相关

```bash
# 检查令牌是否过期
curl -X GET "${COMMAND_SERVICE}/auth/dev-token/info" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  | jq '.data.valid'

# 令牌过期时重新生成
if [ "$(curl -s -X GET "${COMMAND_SERVICE}/auth/dev-token/info" -H "Authorization: Bearer ${JWT_TOKEN}" | jq -r '.data.valid')" = "false" ]; then
  echo "令牌已过期，重新生成..."
  export JWT_TOKEN=$(curl -s -X POST "${COMMAND_SERVICE}/auth/dev-token" \
    -H "Content-Type: application/json" \
    -d '{"userId":"dev-user","tenantId":"dev-tenant","roles":["ADMIN","USER"],"duration":"8h"}' \
    | jq -r '.data.token')
fi
```

### 错误处理

```bash
# 带错误检查的API调用示例
api_call() {
  local response=$(curl -s -w "\n%{http_code}" "$@")
  local body=$(echo "$response" | head -n -1)
  local code=$(echo "$response" | tail -n 1)
  
  if [ "$code" -ge 400 ]; then
    echo "❌ API调用失败 (HTTP $code):"
    echo "$body" | jq '.'
    return 1
  else
    echo "✅ API调用成功 (HTTP $code):"
    echo "$body" | jq '.'
    return 0
  fi
}

# 使用示例
api_call -X GET "${COMMAND_SERVICE}/health"
```

## 性能测试

```bash
# 使用curl进行简单性能测试
echo "⏱️  性能测试 - 组织统计查询"
time curl -s -X POST "${QUERY_SERVICE}/graphql" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "X-Tenant-ID: ${TENANT_ID}" \
  -d '{"query": "query { organizationStats { totalCount } }"}' \
  > /dev/null

# 连续10次请求的平均响应时间
echo "📈 连续10次查询性能测试"
for i in {1..10}; do
  time curl -s -X POST "${QUERY_SERVICE}/graphql" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${JWT_TOKEN}" \
    -H "X-Tenant-ID: ${TENANT_ID}" \
    -d '{"query": "query { organizationStats { totalCount } }"}' \
    > /dev/null
done
```
