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

### 第四步: 建立/更新开发计划 (强制)
```md
在开始实现前，在 `docs/development-plans/` 建立或更新对应计划条目：
- 填写目标/范围/依赖/验收标准/权限契约（基于 docs/api/）
- 执行完成后将计划文档移动到 `docs/archive/development-plans/`
- 入口: docs/development-plans/00-README.md
```

---

## ⚡ 常用命令速查

### 开发环境启动
```bash
make docker-up          # 启动基础设施 (PostgreSQL + Redis)
make run-dev            # 启动后端服务 (命令9090 + 查询8090)
make frontend-dev       # 启动前端开发服务器 (端口3000)
make status             # 查看所有服务状态
make db-migrate-all     # 一键执行数据库迁移（迁移即真源）
```

### 最小依赖与启动顺序（现行 PostgreSQL 原生架构）
- 依赖：PostgreSQL 16+，Redis 7.x
- 顺序：
  1) `make docker-up`（基础设施）
  2) `make run-dev`（命令 9090 + 查询 8090）
  3) `make frontend-dev`（可选）

前端 UI/组件规范详见项目指导原则文档 `CLAUDE.md`（Canvas Kit v13 图标与用法规范）。

### 数据库初始化（迁移优先）
- 规范：严禁使用过时的初始建表脚本；仅通过 `database/migrations/` 按序迁移来初始化/升级数据库。
- 一键迁移：
```bash
# 如未设置，将使用默认: postgres://user:password@localhost:5432/cubecastle?sslmode=disable
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
make db-migrate-all
```
- 说明：审计历史依赖迁移后的 `audit_logs` 列（before_data/after_data/modified_fields/changes/business_context/record_id）。
- 注意：`sql/init/01-schema.sql` 已归档为历史快照，禁止用于初始化；参阅 `docs/archive/deprecated-setup/01-schema.sql`。

### JWT认证管理
```bash
make jwt-dev-mint USER_ID=dev TENANT_ID=default ROLES=ADMIN,USER DURATION=8h
eval $(make jwt-dev-export)     # 导出令牌到环境变量
make jwt-dev-info               # 查看令牌信息
export TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9  # 若未设置，使用默认租户
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
4. 建立/更新计划文档: 在 docs/development-plans/ 添加/更新本次工作计划（完成后归档至 archived/）
5. 开发实现: 遵循CQRS架构和命名规范
6. 测试验证: 运行契约测试和质量检查
7. 更新文档: 重新运行实现清单生成器
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

### 权威链接与治理
- 项目原则与黑名单（长期稳定）：`../../CLAUDE.md`
- 代理/实现强制规范：`../../AGENTS.md`
- API 契约（唯一事实来源）：`../api/openapi.yaml`、`../api/schema.graphql`
- 文档治理与目录边界：`../DOCUMENT-MANAGEMENT-GUIDELINES.md`、`../README.md`

- [实现清单](./02-IMPLEMENTATION-INVENTORY.md) - 查看所有现有功能
- [API与质量工具指南](./03-API-AND-TOOLS-GUIDE.md) - API使用与质量工具指导
- [项目指导原则](../../CLAUDE.md) - 开发规范和原则
- [REST API规范](../api/openapi.yaml) - OpenAPI 3.0规范
- [GraphQL Schema](../api/schema.graphql) - 查询Schema定义
- [开发计划目录使用指南](../development-plans/00-README.md) - 建立/更新计划与归档流程

---

*保持这份文档在手边，开发效率提升100%！*
### GraphQL 示例（新契约，分页包装）
```bash
curl -X POST http://localhost:8090/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"query":"query($p:Int,$s:Int){ organizations(pagination:{page:$p,pageSize:$s}) { data { code name unitType status } pagination { total page pageSize hasNext } } }","variables":{"p":1,"s":10}}'
```

### E2E（Playwright）全局认证
在运行 Playwright E2E 测试前，设置以下环境变量以为所有请求注入认证头：
```bash
export PW_TENANT_ID=$TENANT_ID
export PW_JWT=$JWT_TOKEN
npx playwright test
```

### 组织名称验证说明
- 前端与后端统一验证：组织名称需非空、≤100字符；允许常见字符（中文/英文/数字/空格/连字符/括号等）。
- 建议在回归测试中覆盖含括号名称的创建/更新用例。
