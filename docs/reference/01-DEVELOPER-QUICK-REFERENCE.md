# Cube Castle 开发者快速参考

版本: v1.0  
最后更新: 2025-09-09  
用途: 开发过程中的快速查阅手册

---

## 📋 目录
- [开发前必检清单](#开发前必检清单)
- [常用命令速查](#常用命令速查)
- [端口配置参考](#端口配置参考)
- [API端点速查](#api端点速查)
- [前端组件速查](#前端组件速查)
- [错误排查指南](#错误排查指南)
- [代码规范速查](#代码规范速查)

---

## 🚨 开发前必检清单

### 第一步: 检查实现清单 (强制)
```bash
# 运行实现清单生成器，查看现有功能
node scripts/generate-implementation-inventory.js

# 分析输出，确认新功能是否已存在
# 优先使用现有API/函数/组件，避免重复造轮子
```

### 第二步: 检查API契约
```bash
# 查看REST API规范
cat docs/api/openapi.yaml

# 查看GraphQL Schema
cat docs/api/schema.graphql

# 确保新功能与现有契约一致
```

### 第三步: 确认CQRS使用
```yaml
查询操作 → GraphQL (端口8090)
命令操作 → REST API (端口9090)
严禁混用！
```

---

## ⚡ 常用命令速查

### 开发环境启动
```bash
# 启动基础设施 (PostgreSQL + Redis)
make docker-up

# 启动后端服务 (命令9090 + 查询8090)
make run-dev

# 启动前端开发服务器 (端口3000)
make frontend-dev

# 查看所有服务状态
make status
```

### JWT认证管理
```bash
# 生成开发JWT令牌
make jwt-dev-mint USER_ID=dev TENANT_ID=default ROLES=ADMIN,USER DURATION=8h

# 导出令牌到环境变量
eval $(make jwt-dev-export)

# 查看令牌信息
make jwt-dev-info

# 设置JWT开发环境
make jwt-dev-setup
```

### 质量检查命令
```bash
# 运行重复代码检测
npm run quality:duplicates

# 运行架构一致性验证
npm run quality:architecture

# 运行契约测试
npm test:contract

# 检查文档同步状态
npm run quality:docs
```

### 数据库操作
```bash
# 连接数据库
PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle

# 查看组织表结构
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "\\d organization_units;"

# 查看组织数据
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "SELECT code, name, status FROM organization_units WHERE is_current = true LIMIT 10;"
```

---

## 🔗 端口配置参考

### 核心服务端口
```yaml
前端应用: http://localhost:3000
REST命令API: http://localhost:9090
GraphQL查询API: http://localhost:8090
GraphiQL调试: http://localhost:8090/graphiql
PostgreSQL: localhost:5432
Redis: localhost:6379
```

### 监控服务端口
```yaml
Prometheus: http://localhost:9091
Grafana: http://localhost:3001 (admin/cube-castle-2025)
AlertManager: http://localhost:9093
Node Exporter: http://localhost:9100
```

### ⚠️ 端口配置权威来源
```typescript
// 端口配置统一管理位置
frontend/src/shared/config/ports.ts

// 绝对禁止硬编码端口！
// 违者严重后果自负 - 见CLAUDE.md第16条
```

---

## 🔄 API端点速查

### REST命令API (端口9090)
```bash
# 健康检查
curl http://localhost:9090/health

# 组织CRUD
POST   /api/v1/organization-units           # 创建组织
PUT    /api/v1/organization-units/{code}    # 更新组织
POST   /api/v1/organization-units/{code}/suspend    # 暂停
POST   /api/v1/organization-units/{code}/activate   # 激活

# 时态版本管理
POST   /api/v1/organization-units/{code}/versions   # 创建版本
POST   /api/v1/organization-units/{code}/events     # 时态事件

# 层级管理  
POST   /api/v1/organization-units/{code}/refresh-hierarchy
POST   /api/v1/organization-units/batch-refresh-hierarchy

# 开发工具 (仅DEV模式)
POST   /auth/dev-token         # 生成令牌
GET    /auth/dev-token/info    # 令牌信息
GET    /dev/status             # 开发状态
```

### GraphQL查询API (端口8090)
```graphql
# 基础查询
organizations(filter, pagination): OrganizationConnection!
organization(code, asOfDate): Organization
organizationStats(asOfDate, includeHistorical): OrganizationStats!
organizationHierarchy(code, tenantId): OrganizationHierarchy

# 时态查询示例
query {
  organization(code: "DEPT001", asOfDate: "2025-01-01") {
    name status effectiveDate endDate
  }
}
```

### 认证头部模板
```bash
Authorization: Bearer <JWT_TOKEN>
X-Tenant-ID: <TENANT_ID>
Content-Type: application/json
```

---

## 🎨 前端组件速查

### 核心Hook使用
```typescript
// 查询数据 (GraphQL)
import { useOrganizations, useOrganization } from '@/shared/hooks/useOrganizations';
import { useEnterpriseOrganizations } from '@/shared/hooks/useEnterpriseOrganizations';

// 修改数据 (REST)
import { 
  useCreateOrganization, 
  useUpdateOrganization,
  useSuspendOrganization,
  useActivateOrganization 
} from '@/shared/hooks/useOrganizationMutations';

// 时态数据
import { 
  useTemporalHealth,
  useTemporalAsOfDateQuery,
  useTemporalQueryStats 
} from '@/shared/hooks/useTemporalAPI';

// 消息处理
import { useMessages } from '@/shared/hooks/useMessages';
```

### API客户端
```typescript
// 统一客户端 (自动处理认证、租户、错误)
import { 
  unifiedGraphQLClient,
  unifiedRESTClient 
} from '@/shared/api/unified-client';

// GraphQL查询
const data = await unifiedGraphQLClient.query(QUERY, variables);

// REST命令
const result = await unifiedRESTClient.post('/organization-units', data);
```

### 类型验证和转换
```typescript
// 类型守卫
import { 
  validateOrganizationUnit,
  isAPIError,
  isGraphQLError 
} from '@/shared/api/type-guards';

// 类型转换
import { 
  convertGraphQLToOrganizationUnit,
  convertCreateInputToREST 
} from '@/shared/types/converters';

// 错误处理
import { 
  UserFriendlyError,
  withErrorHandling,
  withOAuthRetry 
} from '@/shared/api/error-handling';
```

### 配置和工具
```typescript
// 端口配置
import { SERVICE_PORTS, CQRS_ENDPOINTS } from '@/shared/config/ports';

// 租户管理
import { tenantManager, getCurrentTenantId } from '@/shared/config/tenant';

// 组织工具
import { 
  normalizeParentCode,
  isRootOrganization,
  getOrganizationLevelText 
} from '@/shared/utils/organization-helpers';

// 时态工具
import { TemporalConverter, TemporalUtils } from '@/shared/utils/temporal-converter';
```

---

## 🔧 错误排查指南

### 常见错误类型
```yaml
401 UNAUTHORIZED:
  原因: JWT令牌无效或过期
  解决: 重新生成令牌 make jwt-dev-mint

403 FORBIDDEN:
  原因: 权限不足或租户ID不匹配
  解决: 检查X-Tenant-ID头部和用户权限

404 NOT_FOUND:
  原因: 组织不存在或URL路径错误
  解决: 检查组织编码和API路径

409 CONFLICT:
  原因: 组织编码重复或版本冲突
  解决: 检查唯一性约束和并发更新

500 INTERNAL_SERVER_ERROR:
  原因: 服务器内部错误
  解决: 查看服务日志和数据库连接
```

### 调试工具
```bash
# 服务健康检查
curl http://localhost:9090/health
curl http://localhost:8090/health

# GraphiQL调试界面
open http://localhost:8090/graphiql

# 查看开发工具端点
curl http://localhost:9090/dev/test-endpoints

# 数据库连接测试
curl http://localhost:9090/dev/database-status

# 性能指标查看
curl http://localhost:9090/dev/performance-metrics
```

### 前端调试
```typescript
// 启用调试模式
localStorage.setItem('debug', 'cube-castle:*');

// 查看API请求日志
// 打开浏览器开发者工具，Network标签页

// 检查Redux DevTools (如果使用)
// 安装Redux DevTools浏览器扩展

// 验证类型安全
console.log('Type validation:', validateOrganizationUnit(data));
```

---

## 📏 代码规范速查

### API命名规范
```yaml
字段命名: 统一使用camelCase
  ✅ parentCode, unitType, isDeleted, createdAt
  ❌ parent_code, unit_type, is_deleted, created_at

路径参数: 统一使用{code}
  ✅ /api/v1/organization-units/{code}
  ❌ /api/v1/organization-units/{id}

协议选择:
  ✅ 查询用GraphQL，命令用REST
  ❌ 混用协议
```

### 前端开发规范
```typescript
// ✅ 使用现有Hook
const { data, loading } = useOrganizations();

// ✅ 使用类型守卫
if (validateOrganizationUnit(response)) {
  // response 现在是类型安全的
}

// ✅ 使用统一错误处理
const result = await withErrorHandling(() => apiCall());

// ❌ 避免硬编码
const API_URL = 'http://localhost:9090'; // 应该用CQRS_ENDPOINTS

// ❌ 避免重复实现
const customOrgHook = () => { ... }; // 应该用useOrganizations
```

### 后端开发规范
```go
// ✅ 统一响应格式
response := &types.APIResponse{
    Success:   true,
    Data:      data,
    Message:   "Operation successful",
    Timestamp: time.Now(),
    RequestID: requestID,
}

// ✅ 错误处理
if err != nil {
    return &types.APIResponse{
        Success: false,
        Error: &types.APIError{
            Code:    "VALIDATION_ERROR",
            Message: "Invalid input",
            Details: details,
        },
        Timestamp: time.Now(),
        RequestID: requestID,
    }
}
```

### Git提交规范
```bash
# ✅ 规范的提交消息
git commit -m "feat: 添加组织暂停功能

- 新增组织暂停API端点
- 添加前端暂停按钮和确认对话框
- 更新API文档和测试用例

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>"

# ❌ 不规范的提交
git commit -m "fix"
git commit -m "更新文件"
```

---

## 🔄 开发工作流速查

### 新功能开发流程
```yaml
1. 运行实现清单检查:
   node scripts/generate-implementation-inventory.js

2. 检查API契约:
   查阅 docs/api/openapi.yaml 和 schema.graphql

3. 优先使用现有资源:
   搜索现有API、Hook、组件

4. 开发实现:
   遵循CQRS架构和命名规范

5. 测试验证:
   运行契约测试和质量检查

6. 更新文档:
   重新运行实现清单生成器
```

### 质量检查流程
```bash
# 提交前检查
npm run quality:all

# 手动检查重点
npm run quality:duplicates      # 重复代码检测
npm run quality:architecture    # 架构一致性
npm test:contract              # 契约测试
npm run build                  # 构建检查
```

### 部署准备检查
```bash
# 服务健康检查
make status

# 数据库连接检查
curl http://localhost:9090/dev/database-status

# API端点检查
curl http://localhost:9090/dev/test-endpoints

# 前端构建检查
npm run build
npm run typecheck
npm run lint
```

---

## 🎯 重点提醒

### 🚨 绝对禁止事项
- ❌ 跳过实现清单检查就开始开发
- ❌ 重复创建已有的API/函数/组件
- ❌ 混用CQRS协议 (GraphQL命令/REST查询)
- ❌ 硬编码端口配置
- ❌ 忽视API契约文件
- ❌ 使用snake_case字段命名

### ✅ 必须遵守
- ✅ 开发前运行 `node scripts/generate-implementation-inventory.js`
- ✅ 优先使用现有资源，避免重复造轮子
- ✅ 查询用GraphQL (8090)，命令用REST (9090)
- ✅ 统一使用camelCase字段命名
- ✅ 所有API调用包含认证头和租户ID
- ✅ 使用类型守卫确保类型安全

---

## 📚 更多资源

### 核心文档
- [实现清单](./IMPLEMENTATION-INVENTORY.md) - 查看所有现有功能
- [API使用指南](./API-USAGE-GUIDE.md) - 详细API使用说明
- [项目指导原则](../../CLAUDE.md) - 开发规范和原则

### API规范
- [REST API规范](../api/openapi.yaml) - OpenAPI 3.0规范
- [GraphQL Schema](../api/schema.graphql) - 查询Schema定义

### 开发计划
- [开发计划目录](../development-plans/) - 项目规划和架构设计

---

*保持这份文档在手边，开发效率提升100%！*

*最后更新: 2025-09-09 | 版本: v1.0*