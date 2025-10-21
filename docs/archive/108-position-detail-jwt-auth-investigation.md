# 108 - 职位详情页面加载失败调查报告（JWT认证问题）

**文档编号**: 108
**创建日期**: 2025-10-21
**调查人员**: Claude (AI Assistant)
**问题级别**: P1 (阻塞核心功能)
**状态**: 已识别根因，待修复

---

## 📋 问题概述

### 问题描述
职位详情页面（例如 P9000004）无法正常加载，前端显示错误信息：
```
加载职位详情失败，请稍后重试。
错误详情：API Error: Query completed with errors
```

### 影响范围
- ✅ **数据层**: 正常（数据库记录完整）
- ✅ **后端逻辑**: 正常（GraphQL resolver 和 repository 实现正确）
- ❌ **认证层**: **异常**（JWT token 签名算法不匹配）
- ❌ **E2E测试**: 失败（无法通过认证进入职位列表和详情页）

### 业务影响
- 用户无法查看任何职位详情
- 职位管理 CRUD 操作全部受阻
- E2E 测试套件无法验证真实后端链路

---

## 🔍 调查过程

### 1. 前端检查

#### GraphQL 查询定义
**文件**: `frontend/src/shared/hooks/useEnterprisePositions.ts:447-579`

```graphql
query PositionDetail($code: PositionCode!, $includeDeleted: Boolean!) {
  position(code: $code) {
    code
    recordId
    title
    # ... 其他字段
  }
  positionTimeline(code: $code) { ... }
  positionAssignments(positionCode: $code, ...) { ... }
  positionTransfers(positionCode: $code, ...) { ... }
  positionVersions(code: $code, includeDeleted: $includeDeleted) { ... }
}
```

**结论**: ✅ 查询定义完整且符合 schema 规范

#### 页面组件
**文件**: `frontend/src/features/positions/PositionTemporalPage.tsx:75-101`

```typescript
const detailQuery = usePositionDetail(isValidCode && !isCreateMode ? code : undefined, {
  enabled: isValidCode && !isCreateMode,
  includeDeleted,
});
```

**结论**: ✅ 组件逻辑正确，正确处理错误状态

### 2. 后端检查

#### GraphQL Schema
**文件**: `docs/api/schema.graphql:126-129`

```graphql
position(
  code: PositionCode!
  asOfDate: Date
): Position
```

**结论**: ✅ Schema 定义正确

#### Resolver 实现
**文件**: `cmd/organization-query-service/internal/graphql/resolver.go:263-274`

```go
func (r *Resolver) Position(ctx context.Context, args struct {
  Code     string
  AsOfDate *string
}) (*model.Position, error) {
  if err := r.permissions.CheckQueryPermission(ctx, "position"); err != nil {
    return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
  }
  r.logger.Printf("[GraphQL] 查询职位详情 code=%s asOfDate=%v", args.Code, args.AsOfDate)

  return r.repo.GetPositionByCode(ctx, sharedconfig.DefaultTenantID, args.Code, args.AsOfDate)
}
```

**结论**: ✅ Resolver 实现正确，包含权限检查和日志

#### Repository 查询
**文件**: `cmd/organization-query-service/internal/repository/postgres_positions.go:235-297`

```go
func (r *PostgreSQLRepository) GetPositionByCode(ctx context.Context, tenantID uuid.UUID, code string, asOfDate *string) (*model.Position, error) {
  // SQL 查询逻辑
  query := fmt.Sprintf(`
    SELECT p.record_id, p.code, p.title, ...
    FROM positions p
    WHERE p.tenant_id = $1 AND p.code = $2
    AND p.is_current = true
    ORDER BY p.effective_date DESC, p.created_at DESC
    LIMIT 1
  `, where)
  // ...
}
```

**结论**: ✅ 数据库查询逻辑正确

### 3. 数据验证

```bash
PGPASSWORD=password psql -h localhost -U user -d cubecastle \
  -c "SELECT code, title, status FROM positions WHERE code = 'P9000004';"
```

**查询结果**:
```
   code   |   title    | status
----------+------------+--------
 P9000004 | 产品设计师 | ACTIVE
```

**结论**: ✅ 数据完整且状态正常

### 4. 日志分析

#### 后端 GraphQL 服务日志
```bash
docker logs cubecastle-graphql 2>&1 | grep -i "P9000004\|error"
```

**关键发现**:
```
[PG-GraphQL] 2025/10/21 09:43:06 Dev mode: JWT validation failed:
  token parsing failed: token is malformed:
  token contains an invalid number of segments

[PG-GraphQL] 2025/10/21 09:51:41 Dev mode: JWT validation failed:
  token parsing failed: token is unverifiable:
  error while executing keyfunc: invalid signing method: HS256
```

**结论**: ❌ **认证失败是根本原因**

---

## 🎯 根本原因分析

### 问题核心
**JWT Token 签名算法不匹配**

### 详细说明

#### 后端配置（`.env` 文件）
```bash
JWT_ALG=RS256
JWT_PRIVATE_KEY_PATH=./secrets/dev-jwt-private.pem
JWT_PUBLIC_KEY_PATH=./secrets/dev-jwt-public.pem
JWT_KEY_ID=bff-key-1
```

**后端期望**: 使用 **RS256**（RSA 非对称加密）算法签名的 JWT token

#### JWT 生成工具
**文件**: `./generate-dev-jwt`

**实际输出**:
```
Valid JWT Token:
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

解码 header 部分：
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

**二进制产物现状**: 仓库根目录的 `./generate-dev-jwt` 可执行文件仍停留在 HS256 版本；执行 `go run ./scripts/cmd/generate-dev-jwt` 可得到 RS256 令牌，进一步确认问题源于本地产物未与源码同步。

**工具生成**: 使用 **HS256**（HMAC 对称加密）算法签名的 token

#### 不匹配的后果
1. 后端 JWT 中间件拒绝所有请求（401 Unauthorized）
2. GraphQL 查询在认证层被拦截，无法到达 resolver
3. 前端收到错误响应，显示为"Query completed with errors"

### 影响链路
```
前端发起请求
  ↓ (Authorization: Bearer HS256_TOKEN)
后端 JWT 中间件
  ↓ (验证失败: 期望 RS256，收到 HS256)
❌ 401 Unauthorized
  ↓
前端显示: "加载职位详情失败"
```

---

## 🛠️ 解决方案

### 方案一：修复 JWT 生成工具（推荐）

#### 目标
确保 `generate-dev-jwt` 生成 RS256 签名的 token

#### 实施步骤
```bash
# 1. 检查私钥文件是否存在
ls -la ./secrets/dev-jwt-private.pem

# 2. 检查 generate-dev-jwt 工具源码
# 确认使用正确的签名算法和密钥路径

# 3. 重新编译 CLI 可执行文件，确保产物最新
go build -o ./generate-dev-jwt ./scripts/cmd/generate-dev-jwt

# 4. 生成新的 token
./generate-dev-jwt

# 5. 验证 token header
# 应该显示 "alg": "RS256"
```

#### 验证方法
```bash
# 解码 token header
TOKEN=$(./generate-dev-jwt | tail -n1)
echo "$TOKEN" | cut -d'.' -f1 | base64 -d | jq .

# 期望输出:
# {
#   "alg": "RS256",
#   "typ": "JWT",
#   "kid": "bff-key-1"
# }
```

### ⚠️ 废弃方案：临时调整后端配置

`internal/config/jwt.go` 已强制 `JWT_ALG=RS256`，在 `.env` 中改为 HS256 会触发启动 panic，且违背 CLAUDE.md 的 RS256-only 原则，故此方案判定为不可行并禁止执行。

### 方案三：前端 Auth Manager 检查

#### 检查点
**文件**: `frontend/src/shared/api/auth.ts`

```typescript
// 确认 token 获取逻辑
async getAccessToken(): Promise<string | null> {
  // 检查 localStorage 中的 token
  // 确认 token 格式完整（三段式）
  // 验证 token 未过期
}
```

#### 验证步骤
```bash
# 在浏览器控制台检查
localStorage.getItem('cubeCastleOauthToken')

# 期望格式:
# {
#   "accessToken": "eyJ...",  # 完整的三段式 token
#   "tokenType": "Bearer",
#   "expiresIn": 28800,
#   "issuedAt": 1729504289
# }
```

---

## ✅ 验证步骤

### 1. Token 格式验证
```bash
# 检查 token 段数（应该是 3）
./generate-dev-jwt | head -1 | awk -F'.' '{print NF}'

# 期望输出: 3
```

### 2. Token 签名算法验证
```bash
# 解码并检查算法
TOKEN=$(./generate-dev-jwt | grep "Valid JWT Token:" | cut -d' ' -f4)
echo $TOKEN | cut -d'.' -f1 | base64 -d

# 期望包含: "alg":"RS256"
```

### 3. 手动 GraphQL 查询测试
```bash
# 使用正确的 token 测试查询
TOKEN=$(./generate-dev-jwt | grep "Valid JWT Token:" | cut -d' ' -f4)

curl -X POST http://localhost:8090/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "query": "query { position(code: \"P9000004\") { code title status organizationCode } }"
  }' | jq .

# 期望输出:
# {
#   "data": {
#     "position": {
#       "code": "P9000004",
#       "title": "产品设计师",
#       "status": "ACTIVE",
#       "organizationCode": "1000021"
#     }
#   }
# }
```

### 4. E2E 测试验证
```bash
# 运行真实后端链路测试
cd /home/shangmeilin/cube-castle
export PW_REQUIRE_LIVE_BACKEND=1
npx playwright test frontend/tests/e2e/position-crud-live.spec.ts --reporter=line

# 期望结果: 所有测试通过
```

### 5. 前端页面手动验证
1. 打开浏览器访问 `http://localhost:5173/positions`
2. 等待认证完成
3. 点击任意职位进入详情页
4. 确认详情页正常加载，显示完整信息

**期望结果**:
- ✅ 职位列表正常显示
- ✅ 详情页正常加载
- ✅ 所有标签页（概览、任职记录、调动记录、时间线、版本历史、审计历史）可切换
- ✅ 无错误提示

---

## 📊 相关文件清单

### 前端
- `frontend/src/shared/hooks/useEnterprisePositions.ts:447-579` - GraphQL 查询定义
- `frontend/src/features/positions/PositionTemporalPage.tsx:75-101` - 页面组件
- `frontend/src/shared/api/auth.ts` - 认证管理器
- `frontend/src/shared/api/unified-client.ts` - GraphQL 客户端
- `frontend/tests/e2e/position-crud-live.spec.ts` - E2E 测试

### 后端
- `docs/api/schema.graphql:126-129` - GraphQL Schema
- `cmd/organization-query-service/internal/graphql/resolver.go:263-274` - Resolver
- `cmd/organization-query-service/internal/repository/postgres_positions.go:235-297` - Repository
- `cmd/organization-query-service/internal/auth/graphql_middleware.go` - JWT 中间件

### 配置
- `.env` - 环境变量配置（JWT_ALG=RS256）
- `./generate-dev-jwt` - JWT 生成工具
- `./secrets/dev-jwt-private.pem` - RSA 私钥
- `./secrets/dev-jwt-public.pem` - RSA 公钥

---

## 🎓 经验总结

### 调查方法论
1. **分层排查**: 从前端 → 后端 → 数据库逐层验证
2. **日志先行**: 后端日志往往包含最直接的错误信息
3. **配置一致性**: 检查配置文件与实际使用是否匹配
4. **手动验证**: 使用 curl 等工具绕过前端直接测试后端

### 常见陷阱
- ❌ 仅看前端错误信息（"Query completed with errors"过于笼统）
- ❌ 忽略认证层（认证失败会导致所有查询失败）
- ❌ 假设配置正确（环境变量与工具实现可能不匹配）
- ✅ **优先检查日志** - 后端日志包含准确的错误原因

### 最佳实践
1. **JWT 配置标准化**: 确保所有工具使用相同的签名算法
2. **E2E 测试覆盖**: 真实后端链路测试能及早发现认证问题
3. **日志可观测性**: 关键错误应有清晰的日志输出
4. **配置文档化**: `.env` 配置应有注释说明用途

---

## 📌 后续行动

### 立即修复（P0）
- [x] 重新编译根目录下的 `generate-dev-jwt`（`go build -o ./generate-dev-jwt ./scripts/cmd/generate-dev-jwt`），确保产物使用 RS256
- [x] 通过 `./generate-dev-jwt | tail -n1` 解码 header 并调用 GraphQL，确认 token 头部 `alg=RS256` 且后端验证通过
- [x] 重新运行相关 E2E 测试确认链路恢复

### 改进建议（P1）
- [x] 在 `generate-dev-jwt` 工具中添加签名算法验证，避免旧产物再次出现 HS256
- [ ] 更新 `docs/development-guides/jwt-development-guide.md`，移除 HS256 作为开发默认值并强调 RS256-only 策略
- [ ] 在 E2E 测试中添加 token 格式验证

### 长期优化（P2）
- [ ] 统一开发环境 JWT 配置管理
- [ ] 添加自动化检查脚本验证 JWT 配置一致性
- [ ] 在 CI 流程中集成 JWT 配置验证

## 执行进展（2025-10-21 继续）
- ✅ 执行 `go build -o ./generate-dev-jwt ./scripts/cmd/generate-dev-jwt` 重新编译开发令牌工具，产物位于仓库根目录。
- ✅ 使用 `./generate-dev-jwt` 生成令牌并通过 `base64` 解码 header，确认 `alg=RS256` 与 `kid=bff-key-1`，符合 CLAUDE.md Docker/JWT 约束。
- ✅ 在 CLI 新增 `ensureTokenAlgorithm` 运行时校验，若签名算法非 RS256 将立即失败，防止旧二进制或错误配置。
- ✅ 重建并滚动更新 `cubecastle-graphql`/`cubecastle-rest` 容器，校验启动日志确保 JWT 初始化为 RS256，无错误告警。
- ✅ 修复 `GetPositionTimeline` SQL（为 `p.operation_reason` 补齐 `AS change_reason` 别名），同步重建 GraphQL 服务容器后，`PositionDetail` 查询返回完整数据（含时间线、版本、任职记录）。
- ✅ 通过 `curl` + RS256 令牌验证 GraphQL `PositionDetail` 查询，确认响应成功。
- ✅ 顺序执行真实链路 Playwright 覆盖：`PW_REQUIRE_LIVE_BACKEND=1 npx playwright test tests/e2e/position-crud-live.spec.ts --project=chromium --workers=1` 与 `--project=firefox --workers=1`，全部通过；并记录并发执行存在偶发超时，后续需在测试计划中单独跟踪。

---

## 🔗 相关文档
- [CLAUDE.md](../../CLAUDE.md) - 项目指导原则
- [JWT开发工具指南](../development-guides/jwt-development-guide.md)
- [开发者快速参考](../reference/01-DEVELOPER-QUICK-REFERENCE.md)
- [107 - 职位收口缺口报告](./107-position-closeout-gap-report.md)

---

**文档状态**: 已完成
**下一步**: 等待开发人员修复 `generate-dev-jwt` 工具
**预计影响**: 修复后，所有职位相关功能将恢复正常
