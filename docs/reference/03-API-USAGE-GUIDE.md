# Cube Castle API使用指南

版本: v1.0  
维护人: 架构组  
适用对象: 前端开发者、后端集成开发者、API使用者

---

## 📖 目录
- [快速开始](#快速开始)
- [CQRS架构说明](#cqrs架构说明)
- [认证与授权](#认证与授权)
- [REST命令API使用](#rest命令api使用)
- [GraphQL查询API使用](#graphql查询api使用)
- [前端集成指南](#前端集成指南)
- [错误处理](#错误处理)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

---

## 🚀 快速开始

### 环境配置
```bash
# 启动开发环境
make docker-up          # 启动PostgreSQL + Redis
make run-dev            # 启动后端服务 (端口9090 + 8090)
make frontend-dev       # 启动前端开发服务器

# 生成开发JWT令牌
make jwt-dev-mint USER_ID=dev TENANT_ID=default ROLES=ADMIN,USER DURATION=8h
```

### 服务端点
- **REST命令服务**: http://localhost:9090/api/v1
- **GraphQL查询服务**: http://localhost:8090/graphql  
- **GraphiQL调试界面**: http://localhost:8090/graphiql
- **前端应用**: http://localhost:3000

---

## 🏗️ CQRS架构说明

### 架构原则
Cube Castle采用严格的CQRS（Command Query Responsibility Segregation）架构：

```yaml
查询操作 (Query):
  协议: GraphQL
  端口: 8090
  用途: 数据查询、统计、报表
  特点: 只读操作，支持复杂查询和聚合

命令操作 (Command):
  协议: REST API
  端口: 9090  
  用途: 数据写入、更新、删除
  特点: 写操作，遵循REST语义
```

### ⚠️ **严格禁止**
- ❌ 使用REST API进行查询操作
- ❌ 使用GraphQL进行数据修改
- ❌ 混用两种协议

### 命名与路径一致性（来自 CLAUDE.md）
- JSON 字段一律使用 camelCase（如 `parentCode`, `effectiveDate`, `recordId`）。
- 组织单元路径参数统一为 `{code}`（禁止 `{id}`）。
- API 契约唯一来源：`docs/api/openapi.yaml`（REST）与 `docs/api/schema.graphql`（GraphQL）。

---

## 🔐 认证与授权

### JWT认证
```bash
# 生成开发令牌
curl -X POST http://localhost:9090/auth/dev-token \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "dev",
    "tenantId": "3b240b62-ea54-4d73-b6c5-1db2a8b4c9e4",
    "roles": ["ADMIN", "USER"],
    "duration": "8h"
  }'
```

### 请求头设置
```bash
# 所有API请求必须包含以下头部
Authorization: Bearer <JWT_TOKEN>
X-Tenant-ID: <TENANT_ID>
Content-Type: application/json
```

### 前端认证使用
```typescript
// 使用统一的认证管理器
import { authManager } from '@/shared/api/auth';

// 自动处理令牌和租户头
const client = new UnifiedRESTClient();
await client.post('/organization-units', data);
```

---

## 🔄 REST命令API使用

### 核心业务操作

#### 创建组织
```bash
curl -X POST http://localhost:9090/api/v1/organization-units \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "研发部门",
    "unitType": "DEPARTMENT", 
    "parentCode": "CORP001",
    "description": "负责产品研发",
    "effectiveDate": "2025-01-01"
  }'
```

#### 更新组织
```bash
curl -X PUT http://localhost:9090/api/v1/organization-units/DEPT001 \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "技术研发部",
    "description": "更新后的描述"
  }'
```

#### 暂停组织
```bash
curl -X POST http://localhost:9090/api/v1/organization-units/DEPT001/suspend \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "部门重组",
    "effectiveDate": "2025-03-01"
  }'
```

### 时态版本管理
```bash
# 创建新版本
curl -X POST http://localhost:9090/api/v1/organization-units/DEPT001/versions \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "AI研发部",
    "effectiveDate": "2025-06-01",
    "description": "转型为AI研发部门"
  }'
```

### 前端REST客户端使用
```typescript
import { unifiedRESTClient } from '@/shared/api/unified-client';

// 创建组织
const createOrganization = async (data: CreateOrganizationInput) => {
  const response = await unifiedRESTClient.post('/organization-units', data);
  return response.data;
};

// 使用专用Hook
import { useCreateOrganization } from '@/shared/hooks/useOrganizationMutations';

const { mutate: createOrg, isLoading, error } = useCreateOrganization();

createOrg({
  name: "新部门",
  unitType: "DEPARTMENT",
  parentCode: "CORP001"
});
```

---

## 📊 GraphQL查询API使用

### 基本查询

#### 组织列表查询
```graphql
query GetOrganizations($filter: OrganizationFilter, $pagination: PaginationInput) {
  organizations(filter: $filter, pagination: $pagination) {
    edges {
      node {
        code
        name
        unitType
        status
        effectiveDate
        endDate
        isCurrent
        parentCode
        level
        codePath
      }
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
      totalCount
    }
  }
}
```

#### 单个组织查询
```graphql
query GetOrganization($code: String!, $asOfDate: String) {
  organization(code: $code, asOfDate: $asOfDate) {
    code
    name
    unitType
    status
    description
    effectiveDate
    endDate
    isCurrent
    parentCode
    level
    codePath
    namePath
  }
}
```

#### 时态查询示例
```graphql
# 查询2025年1月1日时的组织状态
query GetHistoricalOrganization {
  organization(code: "DEPT001", asOfDate: "2025-01-01") {
    name
    status
    effectiveDate
    endDate
  }
}
```

### 统计查询
```graphql
query GetOrganizationStats($asOfDate: String, $includeHistorical: Boolean) {
  organizationStats(asOfDate: $asOfDate, includeHistorical: $includeHistorical) {
    totalCount
    temporalStats {
      totalVersions
      averageVersionsPerOrg
      oldestEffectiveDate
      newestEffectiveDate
    }
    byType {
      unitType
      count
    }
  }
}
```

### 层级查询
```graphql
query GetOrganizationHierarchy($code: String!, $tenantId: String!) {
  organizationHierarchy(code: $code, tenantId: $tenantId) {
    code
    name
    level
    children {
      code
      name
      unitType
    }
    ancestors {
      code
      name
      level
    }
  }
}
```

### 前端GraphQL客户端使用
```typescript
import { unifiedGraphQLClient } from '@/shared/api/unified-client';

// 直接查询
const getOrganizations = async (variables: OrganizationQueryVariables) => {
  const query = `
    query GetOrganizations($filter: OrganizationFilter, $pagination: PaginationInput) {
      organizations(filter: $filter, pagination: $pagination) {
        edges {
          node {
            code
            name
            unitType
            status
          }
        }
      }
    }
  `;
  
  return await unifiedGraphQLClient.query(query, variables);
};

// 使用专用Hook
import { useOrganizations } from '@/shared/hooks/useOrganizations';

const { data, loading, error } = useOrganizations({
  filter: { unitType: 'DEPARTMENT' },
  pagination: { first: 10 }
});
```

---

## 🎨 前端集成指南

### 推荐Hook使用
```typescript
// 1. 数据查询 - 使用GraphQL Hook
import { 
  useOrganizations, 
  useOrganization,
  useEnterpriseOrganizations 
} from '@/shared/hooks';

// 2. 数据修改 - 使用REST Hook  
import {
  useCreateOrganization,
  useUpdateOrganization, 
  useSuspendOrganization,
  useActivateOrganization
} from '@/shared/hooks/useOrganizationMutations';

// 3. 时态查询
import {
  useTemporalHealth,
  useTemporalAsOfDateQuery,
  useTemporalQueryStats
} from '@/shared/hooks/useTemporalAPI';
```

### 组件集成示例
```typescript
import React from 'react';
import { useOrganizations, useCreateOrganization } from '@/shared/hooks';

const OrganizationManager: React.FC = () => {
  // 查询数据
  const { data: organizations, loading, error } = useOrganizations({
    filter: { status: 'ACTIVE' },
    pagination: { first: 20 }
  });

  // 修改操作
  const { mutate: createOrg, isLoading: creating } = useCreateOrganization();

  const handleCreate = (formData: CreateOrganizationInput) => {
    createOrg(formData, {
      onSuccess: () => {
        // 成功后刷新列表
        refetch();
      }
    });
  };

  return (
    <div>
      {/* 组织列表 */}
      {loading ? <Spinner /> : (
        <OrganizationTable data={organizations} />
      )}
      
      {/* 创建表单 */}
      <OrganizationForm onSubmit={handleCreate} loading={creating} />
    </div>
  );
};
```

### 类型安全使用
```typescript
// 使用类型守卫确保安全
import { 
  validateOrganizationUnit,
  isAPIError,
  UserFriendlyError 
} from '@/shared/api/type-guards';

const handleApiResponse = (response: unknown) => {
  if (isAPIError(response)) {
    throw new UserFriendlyError('操作失败', response.message);
  }
  
  if (validateOrganizationUnit(response)) {
    // response 现在是类型安全的 OrganizationUnit
    return response;
  }
  
  throw new Error('Invalid response format');
};
```

---

## ⚠️ 错误处理

### 统一错误格式
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": {
      "field": "name",
      "reason": "Required field missing"
    }
  },
  "timestamp": "2025-09-09T10:30:00Z",
  "requestId": "req-123456"
}
```

### 常见错误码
```yaml
认证错误:
  UNAUTHORIZED: 401 - 未提供有效的JWT令牌
  FORBIDDEN: 403 - 权限不足
  
验证错误:
  VALIDATION_ERROR: 400 - 输入数据验证失败
  BUSINESS_RULE_VIOLATION: 422 - 违反业务规则
  
资源错误:
  NOT_FOUND: 404 - 组织不存在
  CONFLICT: 409 - 组织编码冲突
  
系统错误:
  INTERNAL_SERVER_ERROR: 500 - 服务器内部错误
  SERVICE_UNAVAILABLE: 503 - 服务暂时不可用
```

### 前端错误处理
```typescript
import { 
  ErrorHandler, 
  UserFriendlyError,
  withErrorHandling 
} from '@/shared/api/error-handling';

// 使用错误处理装饰器
const safeApiCall = withErrorHandling(async () => {
  return await unifiedRESTClient.post('/organization-units', data);
});

// 手动错误处理
try {
  const result = await createOrganization(data);
} catch (error) {
  if (error instanceof UserFriendlyError) {
    showToast(error.userMessage);
  } else {
    showToast('操作失败，请稍后重试');
  }
}
```

---

## 💡 最佳实践

### 1. CQRS使用最佳实践
```typescript
// ✅ 正确：查询使用GraphQL
const organizations = await useOrganizations({ status: 'ACTIVE' });

// ✅ 正确：命令使用REST
await useCreateOrganization().mutate(newOrgData);

// ❌ 错误：混用协议
// const organizations = await fetch('/api/v1/organization-units'); // 应该用GraphQL
```

### 2. 时态数据查询
```typescript
// ✅ 当前数据查询
const currentOrg = await useOrganization({ code: 'DEPT001' });

// ✅ 历史数据查询
const historicalOrg = await useOrganization({ 
  code: 'DEPT001', 
  asOfDate: '2025-01-01' 
});

// ✅ 时态统计查询
const stats = await useTemporalQueryStats({ includeHistorical: true });
```

### 3. 错误处理最佳实践
```typescript
// ✅ 使用专用Hook的错误处理
const { mutate, error, isError } = useCreateOrganization();

if (isError && error) {
  // 自动处理用户友好错误消息
  console.error('Creation failed:', error.userMessage);
}

// ✅ 使用统一错误处理
const result = await withOAuthAwareErrorHandling(() => {
  return apiCall();
});
```

### 4. 性能优化
```typescript
// ✅ 使用分页查询
const { data } = useOrganizations({
  pagination: { first: 20, after: cursor }
});

// ✅ 使用查询过滤
const { data } = useOrganizations({
  filter: { 
    unitType: 'DEPARTMENT',
    status: 'ACTIVE' 
  }
});

// ✅ 合理使用时态查询
const { data } = useTemporalAsOfDateQuery({
  asOfDate: selectedDate,
  enabled: !!selectedDate // 只在需要时查询
});
```

---

## ❓ 常见问题

### Q1: 为什么查询和命令要分开？
**A**: CQRS架构提供以下优势：
- **性能优化**: 查询和命令可以独立优化
- **扩展性**: 查询服务可以独立缓存和扩展
- **职责清晰**: 读写操作职责分离，降低复杂度
- **技术选型**: GraphQL适合复杂查询，REST适合标准CRUD

### Q2: 如何处理并发更新？
**A**: 使用乐观锁和版本控制：
```typescript
// 更新时传入版本号
await updateOrganization({
  code: 'DEPT001',
  version: currentVersion,
  data: updatedData
});
```

### Q3: 时态数据如何理解？
**A**: 时态数据支持历史版本查询：
```typescript
// 查询组织当前状态
const current = await useOrganization({ code: 'DEPT001' });

// 查询组织历史状态
const historical = await useOrganization({ 
  code: 'DEPT001', 
  asOfDate: '2024-12-31' 
});
```

### Q4: 如何调试API问题？
**A**: 使用开发工具：
```bash
# 检查服务健康状态
curl http://localhost:9090/health
curl http://localhost:8090/health

# 使用GraphiQL调试GraphQL查询
open http://localhost:8090/graphiql

# 查看API测试工具
curl http://localhost:9090/dev/test-endpoints
```

### Q5: 前端组件如何选择？
**A**: 参考实现清单选择现有组件：
```typescript
// ✅ 使用现有Hook
import { useEnterpriseOrganizations } from '@/shared/hooks';

// ✅ 使用现有工具函数
import { normalizeParentCode, isRootOrganization } from '@/shared/utils/organization-helpers';

// ❌ 避免重复实现
// const customOrganizationHook = () => { ... } // 已有useOrganizations
```

---

## 📚 相关文档
- [实现清单](./IMPLEMENTATION-INVENTORY.md) - 查看所有可用API和组件
- [OpenAPI规范](../api/openapi.yaml) - REST API详细规范
- [GraphQL Schema](../api/schema.graphql) - GraphQL查询规范
- [开发计划文档](../development-plans/) - 项目架构和规范

---

*最后更新: 2025-09-09*  
*版本: v1.0*
