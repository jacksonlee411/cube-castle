# Cube Castle 开发者快速参考

版本: v2.1 | 最后更新: 2025-11-17 | 用途: 开发快速查阅手册

> 说明：本文件为开发速查手册，原则与黑名单以仓库根目录 `AGENTS.md` 为唯一事实来源；若存在不一致，请以 `AGENTS.md` 为准并先校正。

---

> 沟通规范：团队协作与提交物默认使用专业、准确、清晰的中文；如需使用其他语言，请在文档或记录中明确说明受众与范围。
> 
> ⚠️ 最高优先级：任何工作先确保资源唯一性与跨层一致性——若发现重复事实来源或契约偏差，必须立即停止交付并修复。

## 🚨 开发前必检清单

### 预先检查：确认 Go 工具链
```bash
go version          # 需输出 go1.24.x，若低于 1.24 请立即升级本地环境
```

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
查询操作 → GraphQL (同一进程挂载 /graphql，端口9090)
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

### 分支与 PR 工作流（主干 + 远程 PR）
- 本地：在 `master` 上直接开发与提交（Trunk-Based）。
- 远程：`master` 为受保护分支，禁止直接推送；仅在需要推送远程时，创建短生命周期分支并通过 PR 合并（默认 squash-merge）。
- 快速命令：
  ```bash
  # 方式一：手工分支 + 推送后在远程发起 PR（目标 base=master）
  git switch -c feat/<scope>
  git push -u origin HEAD

  # 方式二：自动化（需 GITHUB_TOKEN/gh 已配置）
  make pr-255-soft-gate PR_HEAD=feat/<scope> PR_TITLE='feat: <title>'
  ```
- 合并后安全回切（防丢失他人变更）：
  ```bash
  git switch master
  git fetch --prune
  git pull --ff-only
  git branch -d feat/<scope>   # 本地清理已合并分支
  ```
  工作区有未提交修改时，先 `git stash push -m "wip: <desc>"` 或导出补丁，再回切同步。

### 开发环境启动
```bash
make docker-up          # 启动基础设施 (PostgreSQL + Redis)
make run-dev            # 启动统一 hrms-server：REST (9090) + GraphQL (/graphql 挂载在 9090)
make frontend-dev       # 启动前端开发服务器 (端口3000)
make status             # 查看所有服务状态
make db-migrate-all     # 使用 Goose 执行数据库迁移（迁移即真源）
make db-rollback-last   # 使用 Goose 回滚最近一条迁移
```

> **重要**：前端职位管理页面默认使用真实 GraphQL/REST 数据。若环境存在历史配置，请确保 `.env` / `.env.local` 中设置 `VITE_POSITIONS_MOCK_MODE=false`，避免误用 Mock 数据导致验证失真；Mock 模式下界面会显示只读提醒并禁用写操作。

### 权限契约校验（Plan 252）
```bash
# 生成与校验（阻断未注册引用/映射缺失/授权绕过）
make validate-permissions

# 生成证据快照（logs/plan252/*）便于归档/PR 附件
make plan252-evidence
```
产物路径：
- reports/permissions/*（openapi-scope-usage.json、openapi-scope-registry.json、graphql-query-permissions.json、resolver-permission-calls.json、summary.txt）
- cmd/hrms-server/query/internal/auth/generated/graphql-permissions.json（PBAC 运行时映射，构建期由脚本同步）
- logs/plan252/*（summary + 报告快照）

生产守卫：
- DEV_MODE 默认 false（查询服务）；CI 检查脚本：scripts/quality/validate-devmode-default.sh

### 最小依赖与启动顺序（现行 PostgreSQL 原生架构）
- 依赖：PostgreSQL 16+，Redis 7.x
- 顺序：
  1) `make docker-up`（基础设施）
  2) `make run-dev`（模块化单体 hrms-server，统一注入所有模块）
  3) `make frontend-dev`（可选）

前端 UI/组件规范请参考 `docs/reference/temporal-entity-experience-guide.md`；通用约束以 `AGENTS.md` 为准。

### Temporal Entity 命名与文档入口
- 统一规范文档：`docs/reference/temporal-entity-experience-guide.md`（多页签详情的页面架构/交互/A11y/命名）
- 测试与选择器：E2E 统一使用 `temporalEntity-*` 前缀的 `data-testid`
- 代码入口参考：`frontend/src/features/temporal/*`（页面路由 `pages/entityRoutes.tsx`、适配器 `entity/*`）

### 模块化单体结构导航
- 统一入口：`cmd/hrms-server/`（命令/查询共享配置，通过依赖注入注册各模块）
- 核心业务模块：`internal/organization`（已投产），`internal/workforce`, `internal/contract`（按 203 号计划逐步落地）
- 共享基础设施：`pkg/database`（连接池 + 事务 + outbox）、`pkg/eventbus`、`pkg/logger`、`internal/auth`
- 迁移与 Schema 管理：`database/migrations/`（Goose up/down + Atlas diff），配置文件位于 `atlas.hcl`、`goose.yaml`

### 命令服务启动依赖
- 数据库：命令服务通过 `pkg/database.NewDatabaseWithConfig` 创建连接池，默认 DSN `postgres://user:password@localhost:5432/cubecastle?sslmode=disable`，ServiceName 请设置为 `command-service` 方便指标区分。  
  ```go
  dbClient, _ := database.NewDatabaseWithConfig(database.ConnectionConfig{
      DSN:         os.Getenv("DATABASE_URL"),
      ServiceName: "command-service",
  })
  sqlDB := dbClient.GetDB()
  database.RegisterMetrics(prometheus.DefaultRegisterer)
  outboxRepo := database.NewOutboxRepository(dbClient)
  ```
  > 所有 repository/service/audit 组件均复用同一 `*sql.DB`；Plan 217B 的 outbox dispatcher 将直接注入 `outboxRepo`。
- 事件总线：启动时创建单例 `eventbus.NewMemoryEventBus(logger, metrics)`，并注入需要的模块。Plan 217B 会复用该实例消费 outbox 事件。
- 日志：默认使用 `pkg/logger.NewLogger` + `WithFields` 嵌入 `service=command` 等上下文字段；Plan 218 已全面移除 `log.*` 直接调用。
- 优雅停机：命令服务捕获 SIGINT/SIGTERM，需确保未来的 outbox dispatcher 在 goroutine 中启动，支持 context 取消并在 shutdown 阶段调用 `Stop()`。
- Outbox Dispatcher 配置：通过环境变量 `OUTBOX_DISPATCH_INTERVAL`、`OUTBOX_DISPATCH_BATCH_SIZE`、`OUTBOX_DISPATCH_MAX_RETRY`、`OUTBOX_DISPATCH_BACKOFF_BASE`、`OUTBOX_DISPATCH_METRIC_PREFIX` 调整行为，默认值分别为 `5s`、`50`、`10`、`5s`、`outbox_dispatch`。
- 集成测试：执行 `make test-db-up` 后运行 `go test -tags=integration ./cmd/hrms-server/command/internal/outbox`，验证成功/重试/停机场景；完成后 `make test-db-down` 清理环境。

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

### 审计执行检查（命令侧）
- 审计写入唯一入口：`internal/organization/audit.AuditLogger`，命令域服务在事务中调用 `LogEventInTransaction`，无事务场景才可回退到 `LogEvent`。
- 字段要求：`recordId`、`entityCode`、`actorName`、`requestId`、`correlationId/sourceCorrelation` 必须填充，`business_context.payload` 默认使用 `AfterData` 或错误请求体。
- 链路 ID：确认 `internal/organization/middleware/request.go` / `internal/middleware/request_id.go` 已生效（响应头携带 `X-Request-ID`、`X-Correlation-ID`），服务层通过上下文读取并透传给审计。
- 快速回归：`go test ./internal/organization/audit` 验证事务审计、错误事件、payload 兜底逻辑。

### Scheduler 配置与调试（219D2）
- **唯一事实来源**：`config/scheduler.yaml` + `internal/config/scheduler.go`；启动命令服务时通过 `config.GetSchedulerConfig()` 解析默认值→YAML→`SCHEDULER_*` 环境变量，校验失败会写入 `logs/219D2/config-validation.log` 并阻断启动。完整执行记录参见 `logs/219D2/ACCEPTANCE-RECORD-2025-11-06.md`，由 `docs/development-plans/06-integrated-teams-progress-log.md` 驱动。
- **环境变量覆盖**：统一使用 `SCHEDULER_` 前缀（详见 `.env.example`），常用项包括：
  - `SCHEDULER_ENABLED`（默认 `false`）：可通过 `make run-dev SCHEDULER_ENABLED=true` 临时启用运维任务调度器。
  - （工作流引擎已清退）不再提供 `SCHEDULER_TEMPORAL_ENDPOINT`/`SCHEDULER_NAMESPACE`/`SCHEDULER_TASK_QUEUE` 等配置。
  - `SCHEDULER_MONITOR_ENABLED` / `SCHEDULER_MONITOR_CHECK_INTERVAL`：监控开关与巡检间隔。
  - `SCHEDULER_MONITOR_ENABLED` / `SCHEDULER_MONITOR_CHECK_INTERVAL`：监控开关与巡检间隔（219D3 计划会扩展指标）。
  - `SCHEDULER_TASK_<NAME>_*`：逐任务覆盖 Cron、脚本、初始延迟、启用状态；`<NAME>` 采用任务标识（例如 `DAILY_CUTOVER`）。
  - `SCHEDULER_SCRIPTS_ROOT`：脚本根目录，默认 `./scripts`，路径会做安全校验。
- **运维入口**：`/api/v1/operational/tasks` 返回实时任务状态（含 `NextRun/LastRun/Running`），`/api/v1/operational/tasks/{taskName}/trigger` 支持手动触发；`/api/v1/operational/cutover`、`/consistency-check` 复用相同入口。重放验收流程可参考 `logs/219D2/TEST-SUMMARY.txt`。
- **回滚策略**：若配置出现异常，执行 `make run-dev SCHEDULER_ENABLED=false` 或恢复 `.env`、YAML 默认值即可；必要时按 219D1 附录回退旧目录（详见 `logs/219D2/failure-test.log`）。
- **监控准备**：219D3 将在 `docs/reference/monitoring/` 目录落地 Prometheus/Grafana/Alertmanager 配置，Compose 新增服务端口（Prometheus 9091、Grafana 3001、Alertmanager 9093）；届时请同步检查该目录并更新部署脚本。

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
- 使用 `scripts/dev/mint-dev-jwt.sh` 或 `make jwt-dev-mint` 生成开发令牌（RS256），令牌保存在 `.cache/dev.jwt`。
- 缺少私钥或 JWKS 配置时，命令/查询服务会拒绝启动；请执行 `make jwt-dev-setup` 或使用运维提供的正式密钥。
- `.well-known/jwks.json` 为唯一公钥来源，前端与自动化测试会检测该端点以确认 RS256 已启用。

### 质量检查命令
```bash
# 代码质量门禁（需要 golangci-lint v1.61.0+ 支持 Go 1.24）
make lint                      # Go 代码质量检查
make security                  # Go 安全扫描 (gosec)
make sqlc-generate             # 生成并验证类型安全查询（CI 会执行并要求无 diff）
# 迁移验证建议：本地使用 Goose up/down 预演（make db-migrate-all / make db-rollback-last）
make test-db                   # Docker 化 PostgreSQL 集成测试（含 outbox 验证）

# 前端质量检查
npm run quality:duplicates      # 运行重复代码检测
npm run quality:architecture    # 运行架构一致性验证
npm test:contract              # 运行契约测试
npm run quality:docs           # 检查文档同步状态
```

### REST 命令自测（219C3）
```bash
# 前置：make docker-up && make run-dev（确保命令服务 9090 就绪）
./scripts/219C3-rest-self-test.sh \
  BASE_URL_COMMAND=http://localhost:9090 \
  TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9

# 输出：
#   logs/219C3/validation.log   # 含请求/响应、ruleId、severity
#   logs/219C3/report.json      # 统一结果摘要（passed/failed）
#
# 如需自动验证审计，可在执行前导出 DATABASE_URL 并安装 psql：
#   export DATABASE_URL=postgres://user:password@localhost:5432/cubecastle?sslmode=disable
```
> 场景覆盖：职位填充 + Headcount 超限、任职关闭 + 状态校验、Job Level 版本创建与冲突。所有校验失败都会写入 `audit_logs.business_context.ruleId/severity`。

### 质量门禁工具要求
```bash
# 确认工具版本（必需）
golangci-lint --version       # 要求 v1.61.0+ (支持 Go 1.24)
gosec --version              # 要求 v2.22.8+
which golangci-lint          # 应在 PATH 中可访问
which gosec                  # 应在 PATH 中可访问

# 工具安装参考
# 详见: docs/development-plans/06-integrated-teams-progress-log.md
```

## 🧾 结构化日志（Plan 218 最终规范）

- **统一接口**：生产代码必须通过 `pkg/logger.Logger` 输出日志，禁止使用 `log.Printf` / `log.Println` / `*log.Logger`。运行 `rg "log\\.Print"` 应仅命中文档或第三方生成文件（如 `tools/atlas/...`）。
- **字段模板**：
  - 服务入口：`{"service":"query","component":"query-app","operation":"startup"}`。
  - 仓储/Resolver：追加 `tenantId`、`code`、`operation`、`duration_ms`、`result_count` 等业务字段。
  - 监控告警：`AlertManager.WithLogger(...)` + 渠道 `SetLogger(...)` 输出 `channel`、`alertId`、`level`、`service`、`component`。
- **错误级别**：数据库/外部依赖失败使用 `Errorf`；访问拒绝/输入异常使用 `Warnf`；调试信息使用 `Debugf`，默认级别为 `INFO`。
- **测试 Logger**：使用各模块提供的 `newTestLogger()`（缓存、查询 Resolver、审计等），避免再构造标准库 logger。
- **验收门禁**：提交前执行 `go test ./...` 并确认 `rg "log\\.Print"` 仅匹配 README / vendor；若新增模块依赖日志，应在开发计划与文档中引用本节作为唯一事实来源。

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

# Standard Object REST Port（Plan 402D 启用，只写 SOM 三表）
POST   /api/v1/standard-objects/{objectType}                # 创建 SOM Kernel + 初始版本
POST   /api/v1/standard-objects/{objectType}/{code}/versions # 追加版本（append-only）
PATCH  /api/v1/standard-objects/{objectType}/{code}/status   # 生命周期切换（READY/ACTIVE/SUSPENDED/RETIRED）
# 调用须携带 X-Tenant-ID、Authorization，与 docs/api/openapi.yaml 一致；日志写入 logs/plan402/cleanup/*（锁切换）与 logs/plan402/migration/*（validator）。
# objectType 仅接受 ORGANIZATION_UNIT（职位 ROLE 在 402E 解禁）；code 路径参数沿用标准对象 code。
# 失败时返回 JSON：{"success":false,"error":{"code":...,"message":...},"requestId":...}，可据此排查 Runbook。
```

### Standard Object 双写守则
- 组织命令（创建/更新/版本/停用/激活/作废/删除）复用 `internal/organization/handler/standard_object_adapter.go` 中的 `OrganizationHandler.upsertStandardObject`/`syncOrganizationTimeline`；职位 `Create/Replace/CreateVersion` 则使用 `internal/organization/service/position_standard_object_adapter.go` 的 `PositionService.syncPositionStandardObject`。两侧都在主事务内调用 `standardobject.ObjectService`，失败即回滚。
- Outbox 事件由 `OrganizationHandler.emitStandardObjectEvent` 与 `PositionService.emitStandardObjectEvent` 统一写入 `standard_object.created/updated/versioned/status_changed/retired`，日志与巡检样本见 `logs/plan402/eventbus/*.log`、`logs/plan402/capability/*.log`。
- `standardobject.MakeVersionCode(code, effectiveDate, updatedAt, recordId)` 提供幂等版本号，配合 `pkg/temporal/clock` 生成 `transaction_range`，确保同日纠偏不产生冲突。
- 组织层级 Link (`ORG_HIERARCHY`) 与职位归属 Link (`POSITION_BELONGS_TO_ORG`) 由上述 handler/service 自动维护；若目标对象尚未迁移到 SOM，API 会返回 500，需要先执行 migrator/validator。
- 运行日志：`logs/plan402/migration/*.log`（migrator/validator）、`logs/plan402/capability/*.log`（federate 证据）、`logs/plan400/manifest/*.log` & `logs/plan402/ui/*.log`（Manifest/Slot + OBS/Playwright 证据）。详见 `docs/reference/standard-object-evidence-guide.md`。
- 开发自检：执行 `node scripts/quality/architecture-validator.js --rule capabilityContracts`、`make test`、`npm run quality:preflight`，确保 CQRS、Manifest/Slot、命名守卫均通过。

### Manifest & 快照脚本（402C/C4）
- 表单 manifest：`npm run manifest:forms`（读取 `docs/api/openapi.yaml`，输出 `logs/plan400/manifest/<ts>-forms.log`）
- 列 manifest：`npm run manifest:columns`（读取 `docs/api/schema.graphql`，输出 `logs/plan402/ui/<ts>-columns.log`）
- 快照刷新：`go run -tags legacy ./cmd/tools/standardobject-snapshot-refresh -dsn "$DATABASE_URL" -tenant <TENANT_ID>`，默认日志写入 `logs/plan402/snapshots|metrics`
- Outbox/Redis 巡检：`PGPASSWORD=password psql ...` + `docker exec cubecastle-redis redis-cli info keyspace`，输出到 `logs/plan402/eventbus/<ts>-dispatcher-scan.log`

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
  ℹ️ GraphQL 层禁止写入；命令验证与验收脚本统一使用 REST（参见 `scripts/219C3-rest-self-test.sh`）。

命令自测脚本:
  ```bash
  ./scripts/219C3-rest-self-test.sh   # 产出 logs/219C3/validation.log，供 219C3 验收引用
  ```
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

### 🚨 绝对禁止事项（摘录，权威以 AGENTS.md 为准）
- ❌ 跳过实现清单检查就开始开发
- ❌ 重复创建已有的API/函数/组件
- ❌ 混用CQRS协议
- ❌ 硬编码端口配置
- ❌ 使用snake_case字段命名
- ❌ 绕过 sqlc/Goose/Atlas 流程提交 SQL 变更或事件 outbox 改动

### ✅ 必须遵守（摘录，权威以 AGENTS.md 为准）
- ✅ 开发前运行 `node scripts/generate-implementation-inventory.js`
- ✅ 优先使用现有资源，避免重复造轮子
- ✅ 查询用GraphQL (8090)，命令用REST (9090)
- ✅ 统一使用camelCase字段命名
- ✅ 所有API调用包含认证头和租户ID
- ✅ 软删除判定仅依赖 `status='DELETED'`；`deletedAt` 仅做审计输出
- ✅ 组织详情页时间轴仅承担导航职责；编辑请在“版本历史”页签内完成
- ✅ 数据库迁移附带 `-- +goose Down` 脚本，并通过 Goose up/down 本地验证（`make db-migrate-all` / `make db-rollback-last`）
- ✅ 事件发布走 `pkg/database/outbox`（event_id + retry_count + relay），CI 中以 `make test-db` 回归

---

## 📚 更多资源

### 质量相关（门禁与工具链）
```bash
# 前端统一门禁（阻断）
node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden

# 后端 golangci-lint（与 CI 对齐：v1.59.1，固定调用路径避免误用 PATH 旧版本）
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
$(go env GOPATH)/bin/golangci-lint version
$(go env GOPATH)/bin/golangci-lint run
```
> 说明：golangci-lint 固定版本与路径是 PR 等效门禁的组成部分；若 PATH 中存在其他版本，请勿直接使用 `golangci-lint run`。

### 运行产物治理（Plan 272）
- `logs/`、`reports/`、`test-results/` 均已添加 README，说明保留份数、归档方式与允许文件类型；完成归档守卫（W4）前，可暂以本 README + manifest 手动校验，后续将由 `npm run guard:plan272` 自动执行（阈值：未压缩日志 <2 MB）。
- 超出保留份数后需执行 `make archive-run-artifacts`（Plan 272 W3 将提供脚本）将产物压缩到 `archive/runtime-artifacts/<yyyy-mm>/`，并按 `templates/plan272-manifest.example.json` 模版生成 `manifest.json`（含 `sha256`、Plan ID、引用路径）。
- Runbook/计划只能引用 README/manifest 或 `archive/runtime-artifacts` 中的压缩包；调取历史截图/日志前须根据 manifest 校验 `sha256`。
- Stage 2 需输出 vendored GitHub Actions/第三方依赖的治理结论，并在 `reports/plan272/vendor-audit-*.md` + `docs/development-plans/272-runtime-artifact-cleanup.md` 中追踪。

### 权威链接与治理
- 项目原则与黑名单（唯一事实来源）：`../../AGENTS.md`
- API 契约（唯一事实来源）：`../api/openapi.yaml`、`../api/schema.graphql`
- 文档治理与目录边界：`../DOCUMENT-MANAGEMENT-GUIDELINES.md`、`../README.md`

- [实现清单](./02-IMPLEMENTATION-INVENTORY.md) - 查看所有现有功能
- [API与质量工具指南](./03-API-AND-TOOLS-GUIDE.md) - API使用与质量工具指导
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
