# 组织单元管理API规范

> 重要说明：本文件为“架构与设计说明”，不作为对外契约的唯一权威来源；实际对外 API 契约以 `docs/api/openapi.yaml`（REST 命令）与 `docs/api/schema.graphql`（GraphQL 查询）为准。

**版本**: v4.2  
**架构**: CQRS + PostgreSQL + OAuth 2.0  
**状态**: 生产就绪

## 概述

企业组织架构管理API，基于CQRS架构实现读写分离：
- **查询**: GraphQL (http://localhost:8090/graphql)
- **命令**: REST API (http://localhost:9090/api/v1/organization-units)
- **数据源**: PostgreSQL单一数据源
- **特性**: 时态数据、层级管理、多租户隔离

## 开发环境要求

- Go 1.24 及以上（仓库默认 `toolchain go1.24.9`，请使用 `go version` 自检）。
- Docker Compose 环境提供 PostgreSQL、Redis、Temporal 等依赖，禁止使用宿主机服务。

## CQRS架构

**协议分离**:
```yaml
GraphQL查询端点:
  - organization(code: String!)
  - organizations(filter: OrganizationFilter)
  - organizationStats
  - organizationHierarchy

REST命令端点:
  - POST /api/v1/organization-units (创建)
  - PUT /api/v1/organization-units/{code} (更新/完全替换)
  - POST /api/v1/organization-units/{code}/suspend (停用)
  - POST /api/v1/organization-units/{code}/activate (启用)
  - DELETE /api/v1/organization-units/{code} (删除)
```

## 层级管理

**路径系统**:
- `codePath`: `/1000000/1000001/1000002`
- `namePath`: `/高谷集团/技术部/开发组`
- **深度限制**: 1-17级
- **自动级联**: 父组织变更时自动更新子组织

## 数据模型

### 组织单元核心模型
```json
{
  "code": "1000001",
  "name": "技术部",
  "unitType": "DEPARTMENT",
  "status": "ACTIVE",
  "parentCode": "1000000",
  "level": 2,
  "codePath": "/1000000/1000001",
  "namePath": "/高谷集团/技术部",
  "effectiveDate": "2025-01-01",
  "endDate": null,
  "isCurrent": true,
  "isFuture": false,
  "profile": {
    "description": "负责技术研发",
    "manager": "张三"
  },
  "createdAt": "2025-01-01T00:00:00Z",
  "updatedAt": "2025-01-01T00:00:00Z"
}
```

### 枚举类型

**单元类型**:
- `COMPANY`: 公司
- `DEPARTMENT`: 部门  
- `ORGANIZATION_UNIT`: 组织单位
- `PROJECT_TEAM`: 项目团队

**状态（一维业务状态）**:
- `ACTIVE`: 启用
- `INACTIVE`: 停用（等价于停用/暂停语义）

说明：不再将“PLANNED/历史”作为状态枚举对外暴露；“计划中/历史/当前”语义通过有效期与 asOfDate 计算（见“时态数据”）。

**操作类型**:
- `CREATE`: 创建
- `UPDATE`: 更新
- `SUSPEND`: 停用
- `REACTIVATE`: 重新激活
- `DELETE`: 删除

## API端点详情

### 1. 获取组织列表 (GraphQL)
```graphql
query GetOrganizations($filter: OrganizationFilter) {
  organizations(filter: $filter) {
    code
    name
    unitType
    status
    parentCode
    level
    codePath
  }
}
```

### 2. 创建组织单元 (REST)
```http
POST /api/v1/organization-units
Content-Type: application/json

{
  "code": "1000002",
  "name": "开发组",
  "unitType": "ORGANIZATION_UNIT",
  "parentCode": "1000001",
  "effectiveDate": "2025-01-01",
  "profile": {
    "description": "软件开发团队"
  }
}
```

### 3. 更新组织单元 (REST)
```http
PUT /api/v1/organization-units/1000002
Content-Type: application/json

{
  "name": "高级开发组",
  "unitType": "ORGANIZATION_UNIT",
  "parentCode": "1000001",
  "status": "ACTIVE",
  "profile": {
    "description": "高级软件开发团队"
  }
}
```

说明：`PATCH /{code}` 已移除。请使用 `PUT /{code}` 进行更新；涉及时态或状态变更请使用专用端点（`/{code}/versions`, `/{code}/history/{record_id}`, `/{code}/events`, `/{code}/suspend`, `/{code}/activate`）。

### 4. 停用组织单元 (REST)
```http
POST /api/v1/organization-units/1000002/suspend
Content-Type: application/json

{
  "effectiveDate": "2025-12-31",
  "operationReason": "部门重组"
}
```

## 时态数据

**时态查询支持**:
- `asOfDate`: 指定时间点查询
- `effectiveDate`: 生效日期
- `endDate`: 结束日期
- `isCurrent`: 当前版本标识
- `isFuture`: 未来版本标识

**示例 - 历史时间点查询**:
```graphql
query GetHistoricalOrganization {
  organization(code: "1000001", asOfDate: "2024-12-01") {
    code
    name
    effectiveDate
    endDate
    isCurrent
  }
}
```

## 权限系统

**OAuth 2.0权限**:
- `org:read`: 读取组织信息
- `org:create`: 创建组织
- `org:update`: 更新组织
- `org:delete`: 删除组织
- `org:admin`: 管理员权限

**认证流程**:
```http
POST /oauth/token
Content-Type: application/json

{
  "grant_type": "client_credentials",
  "client_id": "your-client-id", 
  "client_secret": "your-client-secret",
  "scope": "org:read org:create"
}
```

## 响应格式

**成功响应**:
```json
{
  "success": true,
  "data": { },
  "message": "Operation completed successfully",
  "timestamp": "2025-01-01T12:00:00Z",
  "requestId": "uuid"
}
```

**错误响应**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": { }
  },
  "timestamp": "2025-01-01T12:00:00Z",
  "requestId": "uuid"
}
```

## 错误码

**常用错误码**:
- `400 BAD_REQUEST`: 请求参数错误
- `401 UNAUTHORIZED`: 未认证
- `403 FORBIDDEN`: 权限不足
- `404 NOT_FOUND`: 资源不存在
- `409 CONFLICT`: 数据冲突
- `500 INTERNAL_SERVER_ERROR`: 服务器错误

## 性能指标

**响应时间目标**:
- GraphQL查询: < 10ms
- REST命令: < 50ms
- 批量操作: < 200ms

**限制**:
- 单次查询最多返回1000条记录
- 批量操作最多处理100个项目
- API请求频率限制: 1000次/分钟

## 数据库设计（实现约束对齐）

**最小约束**:
- 唯一键: `(tenant_id, code, effective_date)` — 防止同一时点重复版本
- 部分唯一索引: `(tenant_id, code) WHERE is_current=true` — 保证“单当前”并加速点查

**索引**:
- B-tree: `(tenant_id, code, effective_date DESC)` — 相邻版本预检与按日期倒序读取

**实现说明**:
- 边界回填与 is_current 翻转由应用层事务维护（非数据库触发器）
- 自然日边界需要轻量“日切”任务做少量行的 is_current 翻转（可重试、可观测）

## 开发示例

**获取组织列表**:
```bash
curl -X POST http://localhost:8090/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "{ organizations { code name unitType } }"}'
```

**创建组织**:
```bash
curl -X POST http://localhost:9090/api/v1/organization-units \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code":"1000003","name":"测试部","unitType":"DEPARTMENT"}'
```

## 最佳实践

1. **查询优化**: 使用GraphQL字段选择，避免获取不需要的数据
2. **时态数据**: 明确指定effectiveDate，避免隐式时间逻辑
3. **层级管理**: 利用自动级联更新，避免手动维护层级关系
4. **错误处理**: 检查响应的success字段，正确处理错误信息
5. **权限管理**: 按需申请最小权限集合，定期轮换访问令牌

---

*文档版本: v4.2 | 最后更新: 2025-08-23*
