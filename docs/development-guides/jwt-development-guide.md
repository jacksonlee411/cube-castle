# JWT开发工具使用指南

> 快速开始（建议）
> 
> 1) 启动后端：`make run-dev`
> 2) 生成令牌：`make jwt-dev-mint`（可选参数：`USER_ID`、`TENANT_ID`、`ROLES`、`DURATION`）
> 3) 导出令牌：`eval $(make jwt-dev-export)`（将 `JWT_TOKEN` 导入当前 shell）
> 4) 调用 API：
>    - REST：`curl -H "Authorization: Bearer $JWT_TOKEN" -H "X-Tenant-ID: <tenantId>" http://localhost:9090/health`
>    - GraphQL：`curl -H "Authorization: Bearer $JWT_TOKEN" -H "X-Tenant-ID: <tenantId>" http://localhost:8090/graphiql`
> 
> 说明：`X-Tenant-ID` 必填，且必须与 JWT 中的 `tenantId/tenant_id` 一致，否则返回 401/403。

> Playwright E2E：
> 
> - 生成令牌并导出：`make jwt-dev-mint && eval $(make jwt-dev-export)`
> - 设置 E2E 认证环境变量：`export PW_JWT=$JWT_TOKEN && export PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9`
> - 运行测试：`npx playwright test`

## 概述

Cube Castle项目提供了完整的JWT开发工具，帮助开发者在开发环境中快速生成和管理JWT令牌，提升开发效率。

## 🔑 JWT开发工具特性

### 核心功能
- **快速令牌生成**: 一键生成具有指定权限的JWT令牌
- **灵活期限设置**: 支持自定义令牌有效期（1h、8h、24h等）
- **角色权限管理**: 支持多角色令牌生成（ADMIN、USER等）
- **令牌信息查询**: 实时查看令牌状态和剩余有效期
- **开发环境集成**: 与开发工具链无缝集成

### 安全特性
- **开发模式限制**: 仅在开发环境(`DEV_MODE=true`)下可用
- **生产环境保护**: 生产环境自动禁用开发工具端点
- **令牌验证**: 完整的JWT签名验证和过期检查
- **权限控制**: 基于角色的API访问控制
- **租户一致性**: 强制 `X-Tenant-ID` 头与令牌声明 `tenantId/tenant_id` 一致

## ⚙️ 配置参考

`.env.example` 已提供推荐配置段，关键变量：

```
AUTH_MODE=dev              # dev|prod
JWT_ALG=HS256              # 开发默认 HS256；生产建议 RS256 + JWKS
JWT_SECRET=...             # HS256 共享密钥
JWT_ISSUER=cube-castle
JWT_AUDIENCE=cube-castle-api
JWT_ALLOWED_CLOCK_SKEW=60  # 秒
# JWT_JWKS_URL=...         # 生产：IdP 的 JWKS 地址
```

## 🚀 快速开始

### 1. 环境准备

确保后端服务运行在开发模式：
```bash
# 检查开发模式状态
curl http://localhost:9090/dev/status

# 响应应该包含 "devMode": true
```

### 2. 生成第一个JWT令牌

#### 使用cURL
```bash
curl -X POST "http://localhost:9090/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "dev-user",
    "tenantId": "dev-tenant",
    "roles": ["ADMIN", "USER"],
    "duration": "8h"
  }'
```

#### 预期响应
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresAt": "2025-08-25T20:00:00Z",
    "userId": "dev-user",
    "tenantId": "dev-tenant", 
    "roles": ["ADMIN", "USER"]
  },
  "message": "Dev token generated successfully",
  "timestamp": "2025-08-25T12:00:00Z",
  "requestId": "req-123456"
}
```

### 3. 验证令牌
```bash
# 使用生成的令牌验证API访问
export JWT_TOKEN="your_generated_token_here"

curl -X GET "http://localhost:9090/auth/dev-token/info" \
  -H "Authorization: Bearer ${JWT_TOKEN}"
```

## 🛠️ API端点详解

### 1. 生成开发令牌 `POST /auth/dev-token`

**功能**: 生成用于开发和测试的JWT令牌

**请求参数**:
```typescript
interface TestTokenRequest {
  userId?: string;      // 用户ID，默认: "dev-user"
  tenantId?: string;    // 租户ID，默认: "dev-tenant"
  roles?: string[];     // 用户角色，默认: ["ADMIN", "USER"]
  duration?: string;    // 有效期，默认: "24h"
}
```

**支持的duration格式**:
- `"1h"` - 1小时
- `"8h"` - 8小时 (推荐开发使用)
- `"24h"` - 24小时
- `"168h"` - 7天 (长期开发)

**使用示例**:
```bash
# 生成管理员权限令牌
curl -X POST "http://localhost:9090/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "admin-user",
    "tenantId": "dev-tenant",
    "roles": ["ADMIN"],
    "duration": "8h"
  }'

# 生成普通用户令牌
curl -X POST "http://localhost:9090/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "normal-user", 
    "roles": ["USER"],
    "duration": "1h"
  }'
```

### 2. 获取令牌信息 `GET /auth/dev-token/info`

**功能**: 查看当前JWT令牌的详细信息和有效性

**请求头**: 
```
Authorization: Bearer <your_jwt_token>
```

**响应示例**:
```json
{
  "success": true,
  "data": {
    "userId": "dev-user",
    "tenantId": "dev-tenant", 
    "roles": ["ADMIN", "USER"],
    "expiresAt": "2025-08-25T20:00:00Z",
    "valid": true
  },
  "message": "Token information retrieved",
  "timestamp": "2025-08-25T12:30:00Z",
  "requestId": "req-789012"
}
```

**使用场景**:
- 检查令牌是否即将过期
- 验证当前用户权限
- 调试认证问题

### 3. 开发环境状态 `GET /dev/status`

**功能**: 获取开发环境配置信息和功能状态

**响应示例**:
```json
{
  "success": true,
  "data": {
    "devMode": true,
    "timestamp": "2025-08-25T12:00:00Z",
    "service": "organization-command-service",
    "environment": "development",
    "features": {
      "jwtDevTools": true,
      "testEndpoints": true,
      "debugEndpoints": true,
      "mockData": true
    }
  },
  "message": "Development status retrieved",
  "requestId": "req-345678"
}
```

### 4. 测试端点列表 `GET /dev/test-endpoints`

**功能**: 获取所有可用的API端点列表，用于快速查看API结构

**响应示例**:
```json
{
  "success": true,
  "data": {
    "devTools": [
      {"method": "POST", "path": "/auth/dev-token", "description": "Generate development JWT token"},
      {"method": "GET", "path": "/auth/dev-token/info", "description": "Get token information"}
    ],
    "api": [
      {"method": "POST", "path": "/api/v1/organization-units", "description": "Create organization unit"},
      {"method": "PUT", "path": "/api/v1/organization-units/{code}", "description": "Update organization unit"}
    ]
  },
  "message": "Test endpoints listed",
  "requestId": "req-456789"
}
```

## 🔧 开发工具集成

### IDE集成（VSCode）

创建VSCode任务配置 `.vscode/tasks.json`:
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Generate JWT Token",
      "type": "shell",
      "command": "curl",
      "args": [
        "-X", "POST",
        "http://localhost:9090/auth/dev-token",
        "-H", "Content-Type: application/json",
        "-d", "{\"userId\":\"dev-user\",\"duration\":\"8h\"}"
      ],
      "group": "build",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "shared"
      }
    }
  ]
}
```

### 环境变量管理

创建开发环境配置文件 `.env.dev`:
```bash
# Cube Castle 开发环境配置
COMMAND_SERVICE_URL=http://localhost:9090
QUERY_SERVICE_URL=http://localhost:8090
TENANT_ID=dev-tenant

# JWT配置  
JWT_USER_ID=dev-user
JWT_ROLES=ADMIN,USER
JWT_DURATION=8h
```

自动化令牌生成脚本 `scripts/get-jwt-token.sh`:
```bash
#!/bin/bash
# JWT令牌获取脚本

set -e
source .env.dev

echo "🔑 获取JWT开发令牌..."

JWT_TOKEN=$(curl -s -X POST "${COMMAND_SERVICE_URL}/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d "{
    \"userId\": \"${JWT_USER_ID}\",
    \"tenantId\": \"${TENANT_ID}\",
    \"roles\": [\"$(echo ${JWT_ROLES} | sed 's/,/","/g')\"],
    \"duration\": \"${JWT_DURATION}\"
  }" | jq -r '.data.token')

if [ "$JWT_TOKEN" != "null" ] && [ -n "$JWT_TOKEN" ]; then
  export JWT_TOKEN
  echo "✅ JWT令牌获取成功: ${JWT_TOKEN:0:20}..."
  echo "export JWT_TOKEN='${JWT_TOKEN}'" > .jwt-token
  echo "💡 令牌已保存到 .jwt-token 文件，使用 'source .jwt-token' 加载"
else
  echo "❌ JWT令牌获取失败"
  exit 1
fi
```

## 🧪 测试与调试

### 自动化测试脚本

JWT功能测试脚本 `tests/jwt-test.sh`:
```bash
#!/bin/bash
# JWT开发工具功能测试

set -e

BASE_URL="http://localhost:9090"
TESTS_PASSED=0
TESTS_TOTAL=0

# 测试函数
run_test() {
  local test_name="$1"
  local command="$2"
  local expected_status="$3"
  
  echo "🧪 测试: $test_name"
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  
  HTTP_CODE=$(curl -s -o /tmp/test_response -w "%{http_code}" $command)
  
  if [ "$HTTP_CODE" -eq "$expected_status" ]; then
    echo "✅ 通过: HTTP $HTTP_CODE"
    TESTS_PASSED=$((TESTS_PASSED + 1))
  else
    echo "❌ 失败: 期望HTTP $expected_status, 实际HTTP $HTTP_CODE"
    cat /tmp/test_response
  fi
  echo ""
}

echo "🚀 开始JWT开发工具测试"

# 测试1: 开发状态检查
run_test "开发状态检查" \
  "-X GET $BASE_URL/dev/status" \
  200

# 测试2: 生成JWT令牌
run_test "生成JWT令牌" \
  "-X POST $BASE_URL/auth/dev-token -H 'Content-Type: application/json' -d '{\"duration\":\"1h\"}'" \
  200

# 获取生成的令牌
JWT_TOKEN=$(curl -s -X POST "$BASE_URL/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{"duration":"1h"}' | jq -r '.data.token')

# 测试3: 令牌信息查询
run_test "令牌信息查询" \
  "-X GET $BASE_URL/auth/dev-token/info -H 'Authorization: Bearer $JWT_TOKEN'" \
  200

# 测试4: 无效令牌处理
run_test "无效令牌处理" \
  "-X GET $BASE_URL/auth/dev-token/info -H 'Authorization: Bearer invalid_token'" \
  401

# 测试结果汇总
echo "📊 测试结果: $TESTS_PASSED/$TESTS_TOTAL 通过"
if [ "$TESTS_PASSED" -eq "$TESTS_TOTAL" ]; then
  echo "🎉 所有测试通过!"
  exit 0
else  
  echo "⚠️  有测试失败，请检查!"
  exit 1
fi
```

### 性能测试

令牌生成性能测试:
```bash
#!/bin/bash
# JWT令牌生成性能测试

echo "⏱️  JWT令牌生成性能测试"
echo "测试1000次令牌生成请求..."

start_time=$(date +%s)

for i in {1..1000}; do
  curl -s -X POST "http://localhost:9090/auth/dev-token" \
    -H "Content-Type: application/json" \
    -d '{"duration":"1h"}' > /dev/null
done

end_time=$(date +%s)
duration=$((end_time - start_time))

echo "✅ 1000次令牌生成完成"
echo "⏱️  总耗时: ${duration}秒"
echo "📊 平均响应时间: $((duration * 1000 / 1000))毫秒/请求"
echo "🚀 QPS: $((1000 / duration)) 请求/秒"
```

## 🛡️ 安全最佳实践

### 令牌管理
1. **有效期设置**: 开发期间使用8小时有效期，避免频繁刷新
2. **权限最小化**: 根据测试需要设置最小必要权限
3. **定期轮换**: 长期开发项目定期更换令牌

### 环境隔离
1. **开发环境限制**: 确保JWT开发工具仅在开发环境启用
2. **生产环境检查**: 部署前确认生产环境禁用开发工具
3. **配置验证**: 使用`/dev/status`端点验证环境配置

### 数据保护
1. **令牌存储**: 避免将JWT令牌提交到版本控制系统
2. **日志过滤**: 确保日志系统不记录完整的JWT令牌
3. **网络安全**: 开发环境使用HTTPS（如果可能）

## 🔍 故障排除

### 常见问题及解决方案

#### 1. 令牌生成失败
**现象**: 
```json
{
  "success": false,
  "error": {
    "code": "DEV_MODE_DISABLED",
    "message": "Development tools are disabled"
  }
}
```

**解决方案**:
```bash
# 检查开发模式配置
curl http://localhost:9090/dev/status

# 确认环境变量设置
echo $DEV_MODE  # 应该是 "true"

# 重启服务并确认开发模式
DEV_MODE=true go run cmd/organization-command-service/main.go
```

#### 2. 令牌验证失败
**现象**:
```json
{
  "success": false,
  "error": {
    "code": "DEV_INVALID_TOKEN",
    "message": "Invalid token format"
  }
}
```

**解决方案**:
```bash
# 检查令牌格式
echo $JWT_TOKEN | cut -d'.' -f1 | base64 -d | jq '.'

# 重新生成令牌
JWT_TOKEN=$(curl -s -X POST "http://localhost:9090/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{"duration":"8h"}' | jq -r '.data.token')
```

#### 3. 权限不足错误
**现象**: API调用返回403错误

**解决方案**:
```bash
# 检查当前令牌权限
curl -X GET "http://localhost:9090/auth/dev-token/info" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq '.data.roles'

# 生成具有管理员权限的令牌
curl -X POST "http://localhost:9090/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{"roles":["ADMIN"],"duration":"8h"}'
```

### 调试工具

令牌调试脚本 `debug-jwt.sh`:
```bash
#!/bin/bash
# JWT令牌调试工具

if [ -z "$1" ]; then
  echo "使用方法: $0 <jwt_token>"
  exit 1
fi

JWT_TOKEN="$1"

echo "🔍 JWT令牌调试信息"
echo "===================="

# 解析令牌头部
echo "📋 令牌头部:"
echo "$JWT_TOKEN" | cut -d'.' -f1 | base64 -d | jq '.'

# 解析令牌载荷  
echo "📋 令牌载荷:"
echo "$JWT_TOKEN" | cut -d'.' -f2 | base64 -d | jq '.'

# 检查令牌有效期
EXPIRY=$(echo "$JWT_TOKEN" | cut -d'.' -f2 | base64 -d | jq -r '.exp')
CURRENT=$(date +%s)

if [ "$EXPIRY" -gt "$CURRENT" ]; then
  REMAINING=$((EXPIRY - CURRENT))
  echo "✅ 令牌有效，剩余时间: $((REMAINING / 3600))小时$((REMAINING % 3600 / 60))分钟"
else
  echo "❌ 令牌已过期"
fi

# 验证令牌（调用API）
echo "🧪 API验证测试:"
curl -s -X GET "http://localhost:9090/auth/dev-token/info" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq '.'
```

## 📚 相关资源

- [API规范文档](../architecture/01-organization-units-api-specification.md)
- [API测试工具集](../development-tools/README.md)
- [开发者快速参考](../reference/01-DEVELOPER-QUICK-REFERENCE.md)
- [项目安全规范](../../CLAUDE.md#安全最佳实践)

---

*本指南随JWT开发工具的更新而持续维护*
