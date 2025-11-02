# Cube Castle 开发者快速参考

版本: v2.1 | 最后更新: 2025-11-05 | 用途: 开发快速查阅手册

---

> 沟通规范：团队协作与提交物默认使用专业、准确、清晰的中文；如需使用其他语言，请在文档或记录中明确说明受众与范围。
> 
> ⚠️ 最高优先级：任何工作先确保资源唯一性与跨层一致性——若发现重复事实来源或契约偏差，必须立即停止交付并修复。

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
make run-dev            # 启动统一 hrms-server：REST (9090) + GraphQL (8090)
make frontend-dev       # 启动前端开发服务器 (端口3000)
make status             # 查看所有服务状态
make db-migrate-all     # 一键执行数据库迁移（迁移即真源）
```

> **重要**：前端职位管理页面默认使用真实 GraphQL/REST 数据。若环境存在历史配置，请确保 `.env` / `.env.local` 中设置 `VITE_POSITIONS_MOCK_MODE=false`，避免误用 Mock 数据导致验证失真；Mock 模式下界面会显示只读提醒并禁用写操作。

### 最小依赖与启动顺序（现行 PostgreSQL 原生架构）
- 依赖：PostgreSQL 16+，Redis 7.x
- 顺序：
  1) `make docker-up`（基础设施）
  2) `make run-dev`（模块化单体 hrms-server，统一注入所有模块）
  3) `make frontend-dev`（可选）

前端 UI/组件规范详见项目指导原则文档 `CLAUDE.md`（Canvas Kit v13 图标与用法规范）。

### 模块化单体结构导航
- 统一入口：`cmd/hrms-server/`（命令/查询共享配置，通过依赖注入注册各模块）
- 核心业务模块：`internal/organization`（已投产），`internal/workforce`, `internal/contract`（按 203 号计划逐步落地）
- 共享基础设施：`pkg/database`（连接池 + 事务 + outbox）、`pkg/eventbus`、`pkg/logger`、`internal/auth`
- 迁移与 Schema 管理：`database/migrations/`（Goose up/down + Atlas diff），配置文件位于 `atlas.hcl`、`goose.toml`

### 数据库初始化（迁移优先）
- 规范：严禁使用过时的初始建表脚本；仅通过 `database/migrations/` 按序迁移来初始化/升级数据库。
- 一键迁移：
```bash
# 如未设置，将使用默认: postgres://user:password@localhost:5432/cubecastle?sslmode=disable
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
make db-migrate-all
```
- 适用场景：
  - 首次在本地或新环境初始化数据库。
  - 拉取上游变更后，发现 `database/migrations/` 存在新增或修改。
  - 需要验证、评审或回归新的迁移脚本时。
  - 部署/CI 环节中，确保数据库模式与当前代码一致。
- 说明：审计历史依赖迁移后的 `audit_logs` 列（before_data/after_data/modified_fields/changes/business_context/record_id）。
- 注意：`sql/init/01-schema.sql` 已归档为历史快照，禁止用于初始化；参阅 `docs/archive/deprecated-setup/01-schema.sql`。

### JWT认证管理
```bash
make jwt-dev-setup              # 首次运行时生成 RS256 密钥对 (secrets/dev-jwt-*.pem)
scripts/dev/mint-dev-jwt.sh --user-id dev --roles ADMIN,USER   # 直接调用脚本（写入 .cache/dev.jwt）
make jwt-dev-mint USER_ID=dev TENANT_ID=default ROLES=ADMIN,USER DURATION=8h  # 包装脚本，支持 make 变量
eval $(make jwt-dev-export)     # 导出令牌到环境变量
make jwt-dev-info               # 查看令牌信息
export TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9  # 若未设置，使用默认租户
```

#### RS256 首选流程（建议）
- 统一链路：命令服务以 RS256 铸造访问令牌并暴露 JWKS，查询服务用 JWKS 验签。
- 获取令牌（BFF 会话）：
  - 登录建立会话并获取 RS256 短期访问令牌（无需本地存储私钥）：
  - 示例：
    ```bash
    # 建立会话（DEV 或 OIDC_SIMULATE 环境下可用）
    curl -s -c ./.cache/bff.cookies -L "http://localhost:9090/auth/login?redirect=/" >/dev/null
    # 拉取会话，获取 RS256 访问令牌
    curl -s -b ./.cache/bff.cookies http://localhost:9090/auth/session | jq .
    # 使用 accessToken 调用 GraphQL（务必携带 X-Tenant-ID）
    ACCESS_TOKEN="..."; TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
    curl -sS -X POST http://localhost:8090/graphql \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "X-Tenant-ID: $TENANT_ID" \
      -H "Content-Type: application/json" \
      -d '{"query":"query($page:Int,$pageSize:Int){ organizations(pagination:{page:$page,pageSize:$pageSize}) { pagination { total page pageSize hasNext } } }","variables":{"page":1,"pageSize":1}}'
    ```
- JWKS 预览：`curl http://localhost:9090/.well-known/jwks.json`（应返回 RSA 公钥，kid 一般为 `bff-key-1`）。

#### 关于 dev-token（开发专用）
- `scripts/dev/mint-dev-jwt.sh` / `make jwt-dev-mint` 通过 `/auth/dev-token` 生成开发令牌，签名算法固定为 RS256。
- 缺少私钥或 JWKS 配置时，命令/查询服务会拒绝启动；请执行 `make jwt-dev-setup` 或使用运维提供的正式密钥。
- `.well-known/jwks.json` 为唯一公钥来源，前端与自动化测试会检测该端点以确认 RS256 已启用。

### 质量检查命令
```bash
# 代码质量门禁（需要 golangci-lint v1.61.0+ 支持 Go 1.23）
make lint                      # Go 代码质量检查
make security                  # Go 安全扫描 (gosec)
make sqlc-generate             # 生成并验证类型安全查询（CI 会执行并要求无 diff）
make db-migrate-verify         # Goose up/down 预演 + Atlas diff 校验
make test-db                   # Docker 化 PostgreSQL 集成测试（含 outbox 验证）

# 前端质量检查
npm run quality:duplicates      # 运行重复代码检测
npm run quality:architecture    # 运行架构一致性验证
npm test:contract              # 运行契约测试
npm run quality:docs           # 检查文档同步状态
```

### 质量门禁工具要求
```bash
# 确认工具版本（必需）
golangci-lint --version       # 要求 v1.61.0+ (支持 Go 1.23)
gosec --version              # 要求 v2.22.8+
which golangci-lint          # 应在 PATH 中可访问
which gosec                  # 应在 PATH 中可访问

# 工具安装参考
# 详见: docs/development-plans/06-integrated-teams-progress-log.md
```

### E2E 快速入口（本地 / CI 对齐）
```bash
# 1. 启动依赖 + RS256 联调
make docker-up
make run-auth-rs256-sim

# 2. 生成 RS256 开发令牌
make jwt-dev-mint
PW_JWT=$(cat .cache/dev.jwt)
PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9

# 3. 前端目录执行 Playwright
cd frontend
PW_JWT=$PW_JWT PW_TENANT_ID=$PW_TENANT_ID npm run test:e2e

# 4. 指定套件
PW_JWT=$PW_JWT PW_TENANT_ID=$PW_TENANT_ID \
  npm run test:e2e -- tests/e2e/regression-e2e.spec.ts

# 5. 查看报告 / Trace
npx playwright show-report
```
- 规范、调试技巧详见 `docs/development-tools/e2e-testing-guide.md`（Plan 18）。
- CI 门禁：`.github/workflows/e2e-tests.yml` 在 PR 上运行完整 Playwright 套件；失败将阻止合并并上传报告。

---

## 🗂️ Job Catalog 模块速查

### 导航入口
- 侧栏“职位管理”使用 Canvas Kit `SidePanel` + `Expandable` 组合；二级菜单包含“职位列表”“职类”“职种”“职务”“职级”五个子项。
- 导航结构配置位于 `frontend/src/layout/navigationConfig.ts`，二级菜单逻辑集中在 `frontend/src/layout/NavigationItem.tsx`。
- 布局基线截图存放于 `frontend/artifacts/layout/{positions-list,job-family-groups-list,job-family-group-detail}.png`，用于验证 312px 侧栏与卡片分层。

### 权限与后端依赖
- 读取菜单需 `job-catalog:read` scope；写操作分别依赖 `job-catalog:create`、`job-catalog:update`，后端 PBAC 映射详见 `docs/api/openapi.yaml`。
- 前端复用 GraphQL 查询 `jobFamilyGroups/jobFamilies/jobRoles/jobLevels` 与 REST 命令 `/api/v1/job-*` 系列，请确保命令、查询服务均由 Docker 环境提供。

### 验证脚本
- 单元测试：`npm --prefix frontend run test -- --run src/features/job-catalog/__tests__/jobCatalogPages.test.tsx`
- 权限断言：`npm --prefix frontend run test -- --run src/features/job-catalog/__tests__/jobCatalogPermissions.test.tsx`
- E2E 场景：`PW_CAPTURE_LAYOUT=true PW_JWT=... PW_TENANT_ID=... npm --prefix frontend run test:e2e -- tests/e2e/job-catalog-secondary-navigation.spec.ts`

---

## 🔗 端口配置参考

### 核心服务端口
```yaml
> ⚠️ **端口声明**：以下 `localhost` 端点均由 `docker-compose.dev.yml` 暴露的容器服务提供。禁止在宿主机安装 PostgreSQL / Redis / Go 服务占用这些端口；如发现冲突，请优先卸载宿主服务而非修改容器映射。

前端应用: http://localhost:3000（宿主机 Vite，依赖容器服务）
REST命令API: http://localhost:9090（容器 `rest-service` 映射）
GraphQL查询API: http://localhost:8090（容器 `graphql-service` 映射）
GraphiQL调试: http://localhost:8090/graphiql（同上）
PostgreSQL: localhost:5432（容器 `postgres` 映射）
Redis: localhost:6379（容器 `redis` 映射）
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
POST   /api/v1/workforce/employees          # 创建员工（Core HR：workforce v1，按203号计划上线）
PATCH  /api/v1/workforce/employees/{id}     # 更新员工状态/岗位（203号计划）
POST   /api/v1/contracts                    # 创建劳动合同（Core HR：contract v1，203号计划）
POST   /auth/dev-token         # 生成令牌 (仅DEV模式)
```

### GraphQL查询API (端口8090)
```graphql
organizations(filter, pagination): OrganizationConnection!
organization(code, asOfDate): Organization
organizationStats(asOfDate, includeHistorical): OrganizationStats!
organizationHierarchy(code, tenantId): OrganizationHierarchy
employees(filter, pagination): WorkforceEmployeeConnection!        # Core HR（203号计划）
employee(id): WorkforceEmployee                                     # Core HR（203号计划）
contracts(filter, pagination): ContractConnection!                  # Core HR（203号计划）
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
  ✅ parentCode, unitType, status, createdAt
  ❌ parent_code, unit_type, is_deleted, created_at

路径参数: 统一使用{code}
  ✅ /api/v1/organization-units/{code}
  ❌ /api/v1/organization-units/{id}

协议选择:
  ✅ 查询用GraphQL，命令用REST
  ❌ 混用协议
```

### 日志输出规范
- **客户端统一日志**：使用 `@/shared/utils/logger`，禁止直接调用 `console.*`
- **受控桥接**：`logger` 在开发环境输出 `debug/info`；`warn/error` 全环境保留
- **Mutation 调试日志**：使用 `logger.mutation('[Mutation] ...')`，可通过 `VITE_ENABLE_MUTATION_LOGS` 在生产启用
- **例外注释**：`eslint-disable-next-line camelcase` 必须追加 `-- 原因` 说明，CI 会校验执行理由
- **基准示例**：
  ```ts
  import { logger } from '@/shared/utils/logger';

  logger.info('Refreshing hierarchy', { code });
  logger.warn('本地缓存缺失，已触发回源');
  logger.error('命令执行失败', error);
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
- ❌ 绕过 sqlc/Goose/Atlas 流程提交 SQL 变更或事件 outbox 改动

### ✅ 必须遵守
- ✅ 开发前运行 `node scripts/generate-implementation-inventory.js`
- ✅ 优先使用现有资源，避免重复造轮子
- ✅ 查询用GraphQL (8090)，命令用REST (9090)
- ✅ 统一使用camelCase字段命名
- ✅ 所有API调用包含认证头和租户ID
- ✅ 软删除判定仅依赖 `status='DELETED'`；`deletedAt` 仅做审计输出
- ✅ 组织详情页时间轴仅承担导航职责；编辑请在“版本历史”页签内完成
- ✅ 数据库迁移附带 `-- +goose Down` 脚本，并通过 `make db-migrate-verify` 验证
- ✅ 事件发布走 `pkg/database/outbox`（event_id + retry_count + relay），CI 中以 `make test-db` 回归

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
