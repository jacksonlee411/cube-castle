# 组织启停API迁移清单 ✅

## 标准调用方式

### ✅ 启用组织
```http
POST /api/v1/organization-units/{code}/activate
Authorization: Bearer {token}
X-Scopes: org:activate

{
  "operationReason": "恢复组织运营", // 可选字段，省略时将记录为空
  "effectiveDate": "2025-09-06"
}
```

**响应格式**:
```json
{
  "success": true,
  "data": {
    "code": "ORG001",
    "operationType": "REACTIVATE",
    "status": "ACTIVE",
    "businessStatus": "ACTIVE",
    "operatedBy": {"id": "uuid", "name": "User Name"},
    "updatedAt": "2025-09-06T10:30:00Z"
  },
  "timestamp": "2025-09-06T10:30:00Z",
  "requestId": "req-uuid"
}
```

### ✅ 停用组织
```http
POST /api/v1/organization-units/{code}/suspend
Authorization: Bearer {token}
X-Scopes: org:suspend

{
  "operationReason": "部门重组", // 可选字段，省略时将记录为空
  "effectiveDate": "2025-09-06"
}
```

### ✅ 权限要求
- **启用权限**: `org:activate` (替代过时的 `org:reactivate`)
- **停用权限**: `org:suspend`

## ❌ 常见误用案例 (禁止使用)

### ❌ 弃用端点 
```http
# 错误 - 返回 410 Gone
POST /api/v1/organization-units/{code}/reactivate
```
**正确做法**: 使用 `POST /activate`

### ❌ 直接修改状态（PATCH 已移除）
```http
# 错误 - 端点已移除且违反唯一实现原则
PATCH /api/v1/organization-units/{code}
{
  "status": "ACTIVE"
}
```
**正确做法**: 使用 `POST /activate` 或 `POST /suspend`。当前契约已不提供 PATCH，客户端若仍调用将收到 404/405。

### ❌ 过时权限
```yaml
# 错误 - 已废弃的权限
security:
  - oauth2: ["org:reactivate"]
```
**正确做法**: 使用 `org:activate`

### ❌ 前端别名方法
```typescript
// 错误 - 已移除的方法
organizationAPI.reactivate(code, reason);
```
**正确做法**: 使用 `organizationAPI.activate(code, reason)`

## 审计字段说明

### 操作类型与端点分离
- **端点路径**: `/activate` (用户友好，简洁明了)
- **operationType**: `REACTIVATE` (审计精确，语义完整)

这是**领域概念与HTTP路径的合理分离**，两者职责不同：
- HTTP路径：面向用户的简洁接口
- 操作类型：面向审计的精确语义

### 状态字段统一
- **status**: `ACTIVE` | `INACTIVE` (一维业务状态)
- **businessStatus**: 同status值 (向后兼容别名，将逐步废弃)
- **时态语义**: 通过 `isCurrent/isFuture` 计算字段表达

## 410弃用处理

访问 `/reactivate` 端点将收到：

```http
HTTP/1.1 410 Gone
Deprecation: true
Link: </api/v1/organization-units/{code}/activate>; rel="successor-version"
Sunset: 2026-01-01T00:00:00Z

{
  "success": false,
  "error": {
    "code": "ENDPOINT_DEPRECATED", 
    "message": "Use /activate instead of /reactivate"
  },
  "timestamp": "2025-09-06T10:30:00Z",
  "requestId": "req-uuid"
}
```

同时记录审计事件 `DEPRECATED_ENDPOINT_USED`，包含完整访问信息。

## 检查清单

### 开发者检查清单
- [ ] 仅使用 `/activate` 和 `/suspend` 端点
- [ ] 使用正确权限: `org:activate` 和 `org:suspend`
- [ ] 请求体使用 `operationReason` 字段
- [ ] 解析响应中的 `operationType` 和 `businessStatus`
- [ ] 处理标准错误响应格式

### CI/CD检查清单
- [ ] Pre-commit hook阻断禁用模式
- [ ] GitHub Actions验证API合规性
- [ ] Spectral规则校验OpenAPI规范
- [ ] ESLint规则检查TypeScript代码

### 生产环境检查清单
- [ ] 监控410响应频率 (目标: 0)
- [ ] 审计DEPRECATED_ENDPOINT_USED事件 (目标: 0) 
- [ ] 验证activate/suspend成功率 ≥ 99.9%
- [ ] 确认P95延迟 ≤ 150ms

## 支持资源

- **技术规范**: `docs/api/openapi.yaml` - REST API完整规范
- **查询Schema**: `docs/api/schema.graphql` - GraphQL查询规范  
- **架构文档**: `docs/architecture/` - 系统架构设计和决策记录
- **测试用例**: `tests/e2e/activate-suspend-workflow.test.ts`

---

✅ **迁移完成标志**: 生产环境7天内410响应=0，DEPRECATED_ENDPOINT_USED事件=0
