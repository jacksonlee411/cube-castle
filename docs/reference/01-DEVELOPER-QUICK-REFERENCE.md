# Cube Castle 开发者快速参考

版本: v2.0 | 最后更新: 2025-09-09 | 用途: 开发快速查阅手册

---

## 🚨 开发前必检清单

### 第一步: 检查实现清单 (强制)
```bash
# 运行实现清单生成器，查看现有功能
node scripts/generate-implementation-inventory.js
# 优先使用现有API/函数/组件，避免重复造轮子
```

### 第二步: 检查API契约
```bash
# 查看REST API规范和GraphQL Schema
cat docs/api/openapi.yaml
cat docs/api/schema.graphql
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
make docker-up          # 启动基础设施 (PostgreSQL + Redis)
make run-dev            # 启动后端服务 (命令9090 + 查询8090)
make frontend-dev       # 启动前端开发服务器 (端口3000)
make status             # 查看所有服务状态
```

### JWT认证管理
```bash
make jwt-dev-mint USER_ID=dev TENANT_ID=default ROLES=ADMIN,USER DURATION=8h
eval $(make jwt-dev-export)     # 导出令牌到环境变量
make jwt-dev-info               # 查看令牌信息
```

### 质量检查命令
```bash
npm run quality:duplicates      # 运行重复代码检测
npm run quality:architecture    # 运行架构一致性验证
npm test:contract              # 运行契约测试
npm run quality:docs           # 检查文档同步状态
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

### ⚠️ 端口配置权威来源
```typescript
// 端口配置统一管理位置
frontend/src/shared/config/ports.ts
// 绝对禁止硬编码端口！违者严重后果自负
```

---

## 🔄 API端点速查

### REST命令API (端口9090)
```bash
POST   /api/v1/organization-units           # 创建组织
PUT    /api/v1/organization-units/{code}    # 更新组织
POST   /api/v1/organization-units/{code}/suspend    # 暂停
POST   /api/v1/organization-units/{code}/activate   # 激活
POST   /api/v1/organization-units/{code}/versions   # 创建版本
POST   /auth/dev-token         # 生成令牌 (仅DEV模式)
```

### GraphQL查询API (端口8090)
```graphql
organizations(filter, pagination): OrganizationConnection!
organization(code, asOfDate): Organization
organizationStats(asOfDate, includeHistorical): OrganizationStats!
organizationHierarchy(code, tenantId): OrganizationHierarchy
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

// 修改数据 (REST)
import { 
  useCreateOrganization, 
  useUpdateOrganization,
  useSuspendOrganization 
} from '@/shared/hooks/useOrganizationMutations';

// 统一客户端
import { unifiedGraphQLClient, unifiedRESTClient } from '@/shared/api/unified-client';
```

---

## 🔧 错误排查指南

### 常见错误类型
```yaml
401 UNAUTHORIZED: JWT令牌无效，重新生成令牌 make jwt-dev-mint
403 FORBIDDEN: 权限不足，检查X-Tenant-ID头部和用户权限
404 NOT_FOUND: 组织不存在，检查组织编码和API路径
409 CONFLICT: 组织编码重复，检查唯一性约束
500 INTERNAL_SERVER_ERROR: 服务器内部错误，查看服务日志
```

### 调试工具
```bash
curl http://localhost:9090/health       # 服务健康检查
curl http://localhost:8090/health
open http://localhost:8090/graphiql     # GraphiQL调试界面
curl http://localhost:9090/dev/database-status  # 数据库连接测试
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

---

## 🔄 开发工作流速查

### 新功能开发流程
```yaml
1. 运行实现清单检查: node scripts/generate-implementation-inventory.js
2. 检查API契约: 查阅 docs/api/openapi.yaml 和 schema.graphql
3. 优先使用现有资源: 搜索现有API、Hook、组件
4. 开发实现: 遵循CQRS架构和命名规范
5. 测试验证: 运行契约测试和质量检查
6. 更新文档: 重新运行实现清单生成器
```

---

## 🎯 重点提醒

### 🚨 绝对禁止事项
- ❌ 跳过实现清单检查就开始开发
- ❌ 重复创建已有的API/函数/组件
- ❌ 混用CQRS协议
- ❌ 硬编码端口配置
- ❌ 使用snake_case字段命名

### ✅ 必须遵守
- ✅ 开发前运行 `node scripts/generate-implementation-inventory.js`
- ✅ 优先使用现有资源，避免重复造轮子
- ✅ 查询用GraphQL (8090)，命令用REST (9090)
- ✅ 统一使用camelCase字段命名
- ✅ 所有API调用包含认证头和租户ID

---

## 📚 更多资源

- [实现清单](./IMPLEMENTATION-INVENTORY.md) - 查看所有现有功能
- [API使用指南](./API-USAGE-GUIDE.md) - 详细API使用说明
- [项目指导原则](../../CLAUDE.md) - 开发规范和原则
- [REST API规范](../api/openapi.yaml) - OpenAPI 3.0规范
- [GraphQL Schema](../api/schema.graphql) - 查询Schema定义

---

*保持这份文档在手边，开发效率提升100%！*