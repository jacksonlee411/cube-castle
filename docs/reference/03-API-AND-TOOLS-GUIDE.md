# Cube Castle API与质量工具使用指南

版本: v2.0 | 最后更新: 2025-09-13 | 用途: API使用与质量工具统一指南

---

## 🚀 快速开始

### 环境启动
```bash
make docker-up          # 启动基础设施 (PostgreSQL + Redis)
make run-dev            # 启动后端服务 (命令9090 + 查询8090)
make frontend-dev       # 启动前端开发服务器 (3000)
```
> ℹ️ **开发代理说明**：前端 Vite Dev Server 通过 `frontend/src/shared/config/ports.ts` 代理命令/查询服务。默认按浏览器当前协议自动选择；若无法检测则回退为 HTTP，以避免 `.well-known/jwks.json` 出现 `EPROTO`。如需强制使用 HTTPS，请为后端配置有效证书并显式设置 `VITE_SERVICE_PROTOCOL=https`。

### JWT认证设置（全环境统一 RS256）
```bash
# 第一次启动或密钥丢失时生成 RS256 密钥对（secrets/dev-jwt-*.pem）
make jwt-dev-setup

# 启动后端服务（命令9090/查询8090），内部自动加载 RS256 配置并暴露 /.well-known/jwks.json
make run-dev

# 生成 RS256 开发令牌（命令服务 /auth/dev-token）
make jwt-dev-mint USER_ID=dev TENANT_ID=default ROLES=ADMIN,USER DURATION=8h
eval $(make jwt-dev-export)     # 导出令牌到环境变量
```
> ⚠️ **禁止使用 HS256**：命令/查询/前端已经移除 HS256 兜底，若缺少 RS256 私钥或 JWKS 配置，服务将直接失败启动。请务必保证 `.well-known/jwks.json` 可访问，否则前端与测试用例会提示“未启用 RS256”。

### 服务端点
- **REST命令API**: http://localhost:9090/api/v1
- **GraphQL查询API**: http://localhost:8090/graphql
- **GraphiQL调试界面**: http://localhost:8090/graphiql
- **前端应用**: http://localhost:3000

---

## 🏗️ CQRS架构使用

### 核心原则
```yaml
查询操作 (Query):
  协议: GraphQL (端口8090)
  用途: 数据查询、统计、报表

命令操作 (Command):
  协议: REST API (端口9090)
  用途: 创建、更新、删除

严格禁止:
  ❌ REST API进行查询
  ❌ GraphQL进行数据修改
```

### API认证头部
```bash
Authorization: Bearer <JWT_TOKEN>
X-Tenant-ID: <TENANT_ID>
Content-Type: application/json
```

---

## 🔄 REST命令API

### 核心操作
```bash
# 创建组织
curl -X POST http://localhost:9090/api/v1/organization-units \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "name": "研发部门",
    "unitType": "DEPARTMENT",
    "parentCode": "CORP001",
    "effectiveDate": "2025-01-01"
  }'

# 更新组织
curl -X PUT http://localhost:9090/api/v1/organization-units/DEPT001 \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"name": "技术研发部"}'

# 暂停/激活组织
curl -X POST http://localhost:9090/api/v1/organization-units/DEPT001/suspend
curl -X POST http://localhost:9090/api/v1/organization-units/DEPT001/activate
```

### 前端REST使用
```typescript
import { unifiedRESTClient } from '@/shared/api/unified-client';
import { useCreateOrganization } from '@/shared/hooks/useOrganizationMutations';

// Hook方式 (推荐)
const { mutate: createOrg, isLoading } = useCreateOrganization();
createOrg({ name: "新部门", unitType: "DEPARTMENT" });

// 直接调用
const response = await unifiedRESTClient.post('/organization-units', data);
```

### 软删除语义（Status-only）
- **唯一事实来源**: `status = 'DELETED'` 即代表记录已软删除；禁止再使用 `deleted_at` 条件过滤。
- **审计字段**: `deletedAt` 仅用于审计/追踪，可为空，不参与业务判定。
- **客户端逻辑**: 判断是否展示/过滤时统一使用 `status`，前端/脚本不得维护备用布尔字段。

---

## 📊 GraphQL查询API

### 基本查询
```graphql
# 组织列表
query GetOrganizations($filter: OrganizationFilter) {
  organizations(filter: $filter) {
    edges {
      node {
        code
        name
        unitType
        status
        effectiveDate
        isCurrent
        parentCode
      }
    }
    pageInfo {
      totalCount
      hasNextPage
    }
  }
}

# 单个组织
query GetOrganization($code: String!, $asOfDate: String) {
  organization(code: $code, asOfDate: $asOfDate) {
    code
    name
    unitType
    description
    effectiveDate
    endDate
  }
}

# 组织统计
query GetStats {
  organizationStats {
    totalCount
    temporalStats {
      totalVersions
      averageVersionsPerOrg
    }
    byType {
      unitType
      count
    }
  }
}
```

### 前端GraphQL使用
```typescript
import { useOrganizations, useOrganization } from '@/shared/hooks';

// Hook方式 (推荐)
const { data, loading, error } = useOrganizations({
  filter: { status: 'ACTIVE' },
  pagination: { first: 20 }
});

// 时态查询
const { data: historical } = useOrganization({
  code: 'DEPT001',
  asOfDate: '2024-12-31'
});
```

### 软删除语义（GraphQL）
- 查询结果中的 `status` 可能返回 `DELETED`，该状态应视为软删除且默认从业务列表中过滤。
- `deletedAt`、`deletedBy`、`deletionReason` 仅作为审计补充字段返回，可能为空。
- `includeDeleted` 过滤参数（若存在）等价于允许返回 `status = 'DELETED'` 的记录，不再依赖 `deleted_at`。

---

## 🛡️ 质量工具使用

### 开发前必检
```bash
# 检查现有实现 (强制)
node scripts/generate-implementation-inventory.js

# IIG护卫检查 (防重复开发)
node scripts/quality/iig-guardian.js "新功能描述" --guard

# Go代码质量门禁 (需要 golangci-lint v1.61.0+ 支持 Go 1.23)
make lint                                       # Go 代码质量检查
make security                                   # Go 安全扫描 (gosec v2.22.8+)

# P3质量检查套件
bash scripts/quality/duplicate-detection.sh      # 重复代码检测
node scripts/quality/architecture-validator.js   # 架构一致性
node scripts/quality/document-sync.js           # 文档同步
```

### 质量门禁工具配置
```bash
# 确认工具版本和路径
golangci-lint --version    # 要求 v1.61.0+ (支持 Go 1.23 新语法)
gosec --version           # 要求 v2.22.8+
which golangci-lint       # 应在 PATH 中
which gosec              # 应在 PATH 中

# 工具安装说明
# 参考: docs/development-plans/06-integrated-teams-progress-log.md
# golangci-lint v1.55.2 → v1.61.0 解决 Go 1.23 兼容性问题
```

### 质量指标监控
```bash
# 当前质量状态
重复代码率: 2.11% (目标 <5%) ✅
架构违规数: 25个已识别 ⚠️
文档同步率: 20% (目标 >80%) ⚠️
```

### CI/CD集成
- **自动触发**: push到分支，PR合并
- **质量门禁**: 重复代码>5%阻止合并
- **报告位置**: `reports/` 目录下各子系统报告

### E2E冒烟与门禁（新增）
- 本地运行：
```bash
docker compose -f docker-compose.e2e.yml up -d --build   # 拉起完整栈
npm --prefix frontend ci && npm --prefix frontend run -s test:contract
./simplified-e2e-test.sh                                  # 简化E2E（curl）
cat reports/QUALITY_GATE_TEST_REPORT.md                   # 汇总报告
```
- CI 工作流：`.github/workflows/e2e-smoke.yml`
  - 步骤：Compose Up → 健康等待 → 前端契约测试 → 简化E2E → 上传产物
  - 产出：`e2e-smoke-outputs`（包含 E2E 输出与 reports/* 摘要）

### 前端浏览器版 E2E（Playwright）
- CI 工作流：`.github/workflows/frontend-e2e.yml`
- JWT 注入：使用 `PW_JWT` 与 `PW_TENANT_ID` 作为全局认证环境变量
- 执行命令：`npm --prefix frontend run test:e2e`

### 审计一致性门禁（新增）
- 目标：保障“空UPDATE=0 / recordId载荷一致 / 目标触发器不存在（022生效）”。
- 脚本：
  - 报告版 SQL：`scripts/validate-audit-recordid-consistency.sql`
  - 断言版 SQL：`scripts/validate-audit-recordid-consistency-assert.sql`
  - 一键执行：`scripts/apply-audit-fixes.sh`
- CI 工作流：
  - `.github/workflows/audit-consistency.yml`
  - `.github/workflows/consistency-guard.yml`（job: audit）
- 本地等效（仅校验，不改数据）：
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
ENFORCE=1 APPLY_FIXES=0 bash scripts/apply-audit-fixes.sh
```
- 本地修复后校验（建议先执行 021→022）：
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f database/migrations/021_audit_and_temporal_sane_updates.sql
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f database/migrations/022_remove_db_triggers_and_functions.sql
ENFORCE=1 APPLY_FIXES=1 bash scripts/apply-audit-fixes.sh
```

---

## 📊 运行监控（Prometheus）

### 指标端点
```bash
# 命令服务指标端点（无需认证）
curl http://localhost:9090/metrics
```

### 可用指标

#### 1. HTTP 请求计数器（立即可见）
- **名称**: `http_requests_total{method, route, status}`
- **说明**: 由性能中间件自动记录所有 HTTP 请求
- **标签**:
  - `method`: HTTP 方法（GET、POST、PUT、DELETE）
  - `route`: 路由模式（如 `/api/v1/organization-units`）
  - `status`: HTTP 状态码（200、400、500 等）
- **示例查询**:
```bash
curl -s http://localhost:9090/metrics | grep http_requests_total
# 输出: http_requests_total{method="POST",route="/api/v1/organization-units",status="201"} 15
```

#### 2. 时态操作计数器（业务触发）
- **名称**: `temporal_operations_total{operation, status}`
- **说明**: 记录时态版本管理操作的执行情况
- **标签**:
  - `operation`: 操作类型（create、update、delete、suspend、reactivate）
  - `status`: 操作结果（success、error）
- **触发操作**:
  - 创建版本: `POST /api/v1/organization-units/{code}/versions`
  - 更新生效日期: `PUT /api/v1/organization-units/{code}/versions/{versionId}/effective-date`
  - 删除版本: `DELETE /api/v1/organization-units/{code}/versions/{versionId}`
  - 暂停组织: `POST /api/v1/organization-units/{code}/suspend`
  - 激活组织: `POST /api/v1/organization-units/{code}/activate`

#### 3. 审计日志写入计数器（业务触发）
- **名称**: `audit_writes_total{status}`
- **说明**: 记录审计日志写入操作的成功/失败情况
- **标签**:
  - `status`: 写入结果（success、error）
- **触发**: 所有命令操作都会自动触发审计日志写入

### 指标验证

#### 自动化验证脚本
```bash
# 运行指标验证脚本
./scripts/quality/validate-metrics.sh

# 自定义 metrics 端点
METRICS_URL=http://localhost:9090 ./scripts/quality/validate-metrics.sh
```

脚本会验证：
- ✅ 服务可达性
- ✅ `/metrics` 端点响应
- ✅ 关键指标定义存在（`http_requests_total`）
- ⚠️ 业务触发指标状态（`temporal_operations_total`、`audit_writes_total`）

#### 手动验证步骤
```bash
# 1. 检查 metrics 端点可访问性
curl -s http://localhost:9090/metrics | head -5

# 2. 验证 HTTP 请求计数器（应立即可见）
curl -s http://localhost:9090/metrics | grep http_requests_total

# 3. 触发业务操作以生成指标数据
curl -X POST http://localhost:9090/api/v1/organization-units \
  -H "Authorization: Bearer $(cat /tmp/jwt.txt)" \
  -H "X-Tenant-ID: 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9" \
  -H "Content-Type: application/json" \
  -d '{"name":"测试部门","unitType":"DEPARTMENT","parentCode":"0","effectiveDate":"2025-10-10"}'

# 4. 再次检查业务指标（应显示数据点）
curl -s http://localhost:9090/metrics | grep -E "temporal_operations_total|audit_writes_total"
```

### 技术说明

**Prometheus Counter 行为**:
- Counter 指标只有在至少被记录一次（调用 `.Inc()`）后才会出现在 `/metrics` 输出中
- `http_requests_total` 由中间件自动触发，因此启动后立即可见
- `temporal_operations_total` 和 `audit_writes_total` 需要实际业务操作触发
- 这是 Prometheus 的标准行为，不代表指标未正确集成

**代码位置**:
- 指标定义: `cmd/organization-command-service/internal/utils/metrics.go`
- 端点暴露: `cmd/organization-command-service/main.go:202-207`
- 时态操作插桩: `internal/services/organization_temporal_service.go`
- 审计插桩: `internal/audit/logger.go`、`internal/repository/audit_writer.go`

---

## ⚠️ 错误处理

---

## 🔗 进一步阅读与治理
- 项目原则与单一事实来源索引：`../../CLAUDE.md`
- 代理/实现强制规范：`../../AGENTS.md`
- API 契约（唯一事实来源）：`../api/openapi.yaml`、`../api/schema.graphql`
- 文档治理与目录边界：`../DOCUMENT-MANAGEMENT-GUIDELINES.md`、`../README.md`

### 常见错误码
```yaml
401 UNAUTHORIZED: JWT令牌无效 → make jwt-dev-mint
403 FORBIDDEN: 权限不足 → 检查X-Tenant-ID和角色
404 NOT_FOUND: 资源不存在 → 检查组织编码
409 CONFLICT: 编码冲突 → 使用唯一编码
412 PRECONDITION_FAILED: If-Match ETag 不匹配 → 重新获取最新数据并重试
500 INTERNAL_SERVER_ERROR: 服务错误 → 查看日志
```

### 乐观锁（If-Match）
- `/api/v1/organization-units/{code}/suspend` 与 `/activate` 响应头会返回最新 `ETag`。
- 二次提交前端需携带 `If-Match: <ETag>`，以避免覆盖其他用户的最新修改。
- 若返回 412，说明服务器版本已更新：提示用户刷新详情页以获取新的 `ETag` 后再重试。
- 和 `Idempotency-Key` 配合使用，可同时解决重复提交与并发覆盖问题。

### 调试工具
```bash
curl http://localhost:9090/health       # 服务健康检查
curl http://localhost:8090/health
open http://localhost:8090/graphiql     # GraphiQL调试界面
```

---

## 💡 最佳实践

### CQRS使用规范
```typescript
// ✅ 正确：查询使用GraphQL
const orgs = await useOrganizations({ status: 'ACTIVE' });

// ✅ 正确：命令使用REST
await useCreateOrganization().mutate(data);

// ❌ 错误：混用协议
// const orgs = await fetch('/api/v1/organization-units'); // 应用GraphQL
```

### 开发工作流
```yaml
1. 运行实现清单检查: node scripts/generate-implementation-inventory.js
2. 检查API契约: docs/api/openapi.yaml 和 schema.graphql
3. IIG护卫检查: node scripts/quality/iig-guardian.js "功能" --guard
4. 优先使用现有API/Hook/组件
5. 开发实现: 遵循CQRS和camelCase命名
6. 质量检查: 运行P3检测套件
7. 提交代码: Pre-commit Hook自动验证
```

### 时态数据处理
```typescript
// 当前数据
const current = await useOrganization({ code: 'DEPT001' });

// 历史数据
const historical = await useOrganization({
  code: 'DEPT001',
  asOfDate: '2025-01-01'
});

// 版本管理
POST /api/v1/organization-units/DEPT001/versions
{
  "name": "新名称",
  "effectiveDate": "2025-06-01"
}
```

---

## 🔧 故障排除

### 质量工具问题
```bash
# jscpd未找到
npm install -g jscpd

# 脚本权限问题
chmod +x scripts/quality/*.sh

# Pre-commit Hook未安装
cp scripts/git-hooks/pre-commit-architecture.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# GitHub Actions失败
查看Actions页面 → 点击失败workflow → 查看详细日志
```

### API调试
```bash
# JWT令牌问题
make jwt-dev-mint
eval $(make jwt-dev-export)

# 服务连接问题
curl http://localhost:9090/health
curl http://localhost:8090/health

# 数据库连接
curl http://localhost:9090/dev/database-status
```

---

## 📚 相关资源

- [实现清单](./02-IMPLEMENTATION-INVENTORY.md) - 查看现有API和组件
- [开发者快速参考](./01-DEVELOPER-QUICK-REFERENCE.md) - 核心命令速查
- [OpenAPI规范](../api/openapi.yaml) - REST API详细定义
- [GraphQL Schema](../api/schema.graphql) - 查询Schema定义
- [项目指导原则](../../CLAUDE.md) - 开发规范和原则

---

## 🎯 核心提醒

### 绝对禁止
- ❌ 跳过实现清单检查就开发
- ❌ 重复创建已有功能
- ❌ 混用CQRS协议
- ❌ 硬编码端口配置
- ❌ 使用snake_case字段命名

### 必须遵守
- ✅ 开发前运行IIG护卫检查
- ✅ 优先复用现有资源
- ✅ 查询用GraphQL (8090)，命令用REST (9090)
- ✅ 统一使用camelCase字段命名
- ✅ 所有API调用包含认证头和租户ID

---

*Cube Castle API与质量工具统一指南 - 一站式开发参考*
