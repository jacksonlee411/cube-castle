# internal/organization

组织聚合模块的唯一事实来源。目录结构：

- `audit/`: 审计记录写入器及依赖。
- `handler/`: REST/BFF 处理器（命令侧）。
- `middleware/`: 组织模块专用中间件（性能、限流、请求/关联 ID）。
- `repository/`: 命令/查询共享的 PostgreSQL 仓储实现与时间轴管理器。
- `resolver/`: GraphQL Resolver（查询侧入口）。
- `service/`: 领域服务、Job Catalog/Position/Cascade 等（面向业务事务）。
- `scheduler/`: Temporal workflow / monitor / operational scheduler 及 `scheduler.Service` Facade，提供调度与定时任务能力。
- `validator/`: BusinessRuleValidator 及规则定义。
- `dto/`: GraphQL 查询/响应 DTO 与共享类型。
- `utils/`: 处理器/仓储共用的工具函数（响应、验证、metrics、parent code 等）。

## 聚合边界

- Department 是 Organization 聚合内节点，通过 `unitType=DEPARTMENT` 表示。
- Position/JobCatalog/Assignment 共用 PostgreSQL 数据源，命令侧通过 `service/*` 管理，查询侧通过 `resolver` + `repository`。

## 审计规范（219C1）

- 唯一事实来源：`internal/organization/audit/logger.go` 提供的 `AuditLogger`，仅允许通过 `LogEvent`/`LogEventInTransaction` 写入 `audit_logs`。
- 事务化要求：所有命令域服务（OrganizationTemporalService、JobCatalogService、PositionService 等）必须在业务事务中调用 `LogEventInTransaction`；审计写入失败时应回滚业务操作。
- 字段对齐：`AuditEvent` 必须填充 `recordId`、`entityCode`、`actorName`、`requestId`、`correlationId/sourceCorrelation` 等字段；`business_context.payload` 默认保存 `AfterData` 或错误上下文。
- 链路标识：`internal/organization/middleware/request.go` 中间件负责注入 `X-Request-ID`、`X-Correlation-ID` 并写入上下文；服务层从上下文获取并透传给审计。
- 测试入口：`go test ./internal/organization/audit` 验证事务写入/错误记录等关键路径。

## 迁移清单（219A）

| 旧路径 | 新路径 | 说明 |
| --- | --- | --- |
| `cmd/hrms-server/command/internal/handlers/*` | `internal/organization/handler/*` | REST/BFF 入口集中于 handler 包。 |
| `cmd/hrms-server/command/internal/services/*` | `internal/organization/service/*` | 领域服务共享给命令适配层。 |
| `cmd/hrms-server/command/internal/repository/*` | `internal/organization/repository/*` | 命令/查询仓储统一。 |
| `cmd/hrms-server/command/internal/audit/*` | `internal/organization/audit/*` | 审计日志实现。 |
| `cmd/hrms-server/command/internal/validators/*` | `internal/organization/validator/*` | 业务校验统一入口。 |
| `cmd/hrms-server/command/internal/utils/*` | `internal/organization/utils/*` | 公共工具函数。 |
| `cmd/hrms-server/query/internal/graphql/*` | `internal/organization/resolver/*` | GraphQL Resolver 共享。 |
| `cmd/hrms-server/query/internal/repository/*` | `internal/organization/repository/*` | 查询仓储共用组织模块。 |
| `cmd/hrms-server/query/internal/model/*` | `internal/organization/dto/*` | GraphQL DTO 单一来源。 |

## API 适配

- `internal/organization/api.go` 暴露 `CommandModule` 及 `CommandHandlers` 构建函数，命令服务只需依赖该 API。
- 查询服务通过 `internal/organization/resolver` & `repository` 注入 GraphQL 应用。

## 前端入口（Plan 242 对齐）

- 组织详情页面由共享的 `TemporalEntityPage` 驱动：前端路由 `frontend/src/features/temporal/pages/entityRoutes.tsx` 统一处理 `/organizations/:code/temporal` 的参数校验与导航，内容层仍复用 `TemporalMasterDetailView`。  
- 任意前端路线/文档更新需引用 `TemporalEntityPage`，避免回退到 `OrganizationTemporalPage` 旧命名。

## 查询与缓存（219B）

- `AssignmentQueryFacade` 提供统一的任职查询、历史与统计接口，并负责 Redis 缓存键管理（前缀 `org:assignment:stats`）。
- 缓存策略：职位维度统计命中 Redis，TTL 默认 2 分钟，命令侧 Outbox Dispatcher 发布 `assignment.*` 事件后调用 `RefreshPositionCache` 触发失效。
- GraphQL 新增查询：`assignments`、`assignmentHistory`、`assignmentStats` 均通过 Facade 获取数据，保持与 `docs/api/schema.graphql` 契约一致。

后续 219B~219E 将在本 README 中继续补充审计/验证规则、调度说明、测试脚本等章节。

## 测试与验收（219E）
- **组织生命周期冒烟**：`scripts/e2e/org-lifecycle-smoke.sh`（REST 创建→子部门→停用→启用→GraphQL 验证），输出日志 `logs/219E/org-lifecycle-*.log`。
- **REST 性能基准**：`scripts/perf/rest-benchmark.sh`（依赖 `hey`，默认目标 `/api/v1/organization-units`），输出日志 `logs/219E/perf-rest-*.log`。
- **执行前置**：需执行 `make run-dev`（Docker Compose）以启动 command/query 服务；若环境受限，参考 `logs/219E/BLOCKERS-*.md` 处理 Docker 权限。

## Scheduler / Temporal（219D）

- **实现聚合**：`scheduler/` 目录由 `scheduler.Service` 聚合 `TemporalService`、`TemporalMonitor`、`OperationalScheduler`、`OrganizationTemporalService`，入口由 `internal/organization/api.go` 构建，并在 `cmd/hrms-server/command/main.go` `Services.Scheduler.Start/Stop` 中统一托管。
- **配置单一事实来源**：`config/scheduler.yaml` + `internal/config/scheduler.go`。命令服务启动前调用 `config.GetSchedulerConfig()` 解析默认值→YAML→`SCHEDULER_*` 环境变量，元信息（来源/覆盖/校验错误）记录于 `SchedulerConfigMetadata`，验收日志见 `logs/219D2/config-validation.log`、`logs/219D2/ACCEPTANCE-RECORD-2025-11-06.md`。
- **常用覆盖项**：`SCHEDULER_ENABLED`（默认 `false`）、`SCHEDULER_MONITOR_*`、`SCHEDULER_TASK_<NAME>_*`、`SCHEDULER_SCRIPTS_ROOT`。执行 `make run-dev SCHEDULER_ENABLED=false` 可进行关闭演练（日志：`logs/219D2/startup-disabled.log`）。  
  （说明：工作流引擎相关参数 `SCHEDULER_TEMPORAL_ENDPOINT`/`SCHEDULER_NAMESPACE`/`SCHEDULER_TASK_QUEUE` 等已清退。）
- **校验与回滚**：`config.ValidateSchedulerConfig` 执行结构化校验，失败时阻断启动并输出至 `logs/219D2/config-validation.log`；恢复步骤详见 `logs/219D2/failure-test.log` 与 Plan 06 日志。在极端情况下可按 219D1 附录回退至旧 service 目录，但需同时复原 `api.go` 和 `cmd/hrms-server/command/main.go` 的依赖注入。
- **运维与 API**：`OperationalScheduler` 根据配置动态构建任务，`ListTasks`/`RunTask` 与 REST `/api/v1/operational/tasks`、`/api/v1/operational/tasks/{taskName}/trigger`、`/api/v1/operational/cutover`、`/consistency-check` 对齐；操作流程与截图记录在 `logs/219D2/TEST-SUMMARY.txt`。
- **监控与告警（219D3）**：Prometheus/Grafana/Alertmanager 配置集中在 `docs/reference/monitoring/`（`prometheus/`、`grafana/`、`alertmanager/` 子目录），对应端口 `9091/3001/9093` 已写入 `docker-compose*.yml` 与 `.env.example`；验证记录参考 `logs/219D3/VALIDATION-2025-11-06.md`。PromQL 与 Dashboard 操作指南详见 `docs/reference/03-API-AND-TOOLS-GUIDE.md#scheduler-监控栈`。
- **测试与故障演练（219D4）**：`internal/organization/scheduler/operational_scheduler_test.go`、`temporal_monitor_test.go` 覆盖脚本任务与监控告警逻辑；运行命令 `GOCACHE=$(mktemp -d) go test ./internal/organization/scheduler`. 故障注入脚本 `scripts/dev/scheduler-alert-smoke.sh` 用于复现 CRITICAL 告警并验证 Alertmanager 链路，执行记录归档在 `logs/219D4/`（`TEST-SUMMARY.txt`、`FAULT-INJECTION-2025-11-06.md`、`ACCEPTANCE-RECORD-2025-11-06.md`）。

## Validators

> 状态：预留章节，用作 `BusinessRuleValidator` 链式框架与规则矩阵的唯一事实来源。链路实现将在计划 219C2A 完成后补齐，本段落先给出约束与记录模板，避免信息散落其他文档。

### 设计原则
- 统一入口：命令（REST）、查询（GraphQL Queries）、批处理共用同一校验链工厂，禁止在 handler/service 内写散落规则。
- 规则标识：`Rule ID` 使用 `{域}-{语义}`（如 `ORG-DEPTH`、`POS-HEADCOUNT`），对应错误码需在 OpenAPI 中登记。
- 严重级别：仅允许 `CRITICAL | HIGH | MEDIUM | LOW`，并与审计 `business_context.severity` 保持同步。
- 返回结构：复用 `ValidationResult`/`ValidationError`/`ValidationWarning`，禁止新增并行事实来源。

### 规则登记模版（P0 冻结草案）
| Rule ID | Priority | Severity | Error Code | Triggered At | Handler/Service | Notes |
| ------- | -------- | -------- | ---------- | ------------ | ---------------- | ----- |
| ORG-DEPTH | P0 | HIGH | ORG_DEPTH_LIMIT | REST `POST /api/v1/organization-units` | `internal/organization/validator/core.go`（链式），`handler/organization_helpers.go`（翻译） | 限制最大层级 17，metadata 暴露 `maxDepth` 与 `attemptedDepth`。 |
| ORG-CIRC | P0 | CRITICAL | ORG_CYCLE_DETECTED | REST `PATCH /api/v1/organization-units/{code}` | 同上 | 检测父子循环与自引用；失败阻断事务。 |
| ORG-STATUS | P0 | CRITICAL | ORG_STATUS_GUARD | REST `POST /api/v1/organization-units/{code}/activate` 等状态流转入口 | 同上 | 防止非法激活/停用（含冻结状态）。 |
| POS-ORG | P0 | HIGH | POS_ORG_INACTIVE | REST `POST /api/v1/positions`、`PUT /api/v1/positions/{code}` | `internal/organization/validator/core.go` 链式 + Position handler | 职位引用的组织必须处于激活态。 |
| POS-HEADCOUNT | P0 | HIGH | POS_HEADCOUNT_EXCEEDED | REST `POST /api/v1/positions/{code}/fill` | 同上 | 不允许超过职位核定编制；metadata 暴露 `headcountLimit`、`requested`。 |
| ASSIGN-STATE | P0 | CRITICAL | ASSIGN_INVALID_STATE | REST 任职状态变更入口（`POST /api/v1/positions/{code}/assignments` 等） | 同上 | 状态流转需遵循状态机（ACTIVE→VACATED 等）。 |
| ASSIGN-FTE | P0 | HIGH | ASSIGN_FTE_LIMIT | REST `POST /api/v1/positions/{code}/assignments` | 同上 | FTE 总量 0-1 区间，允许配置容差。 |
| CROSS-ACTIVE | P0 | HIGH | CROSS_ACTIVATION_CONFLICT | 跨域命令（职位↔Job Catalog，Assignment↔Organization 激活） | 同上 | 多聚合联动时要求关联实体均为 ACTIVE。 |

> 说明：表格作为 P0 规则冻结草案，后续 219C2A 执行时若需调整，必须先更新本节并取得架构/安全组确认，再同步 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 与 OpenAPI。

#### P1 扩展规则（219C2C）
| Rule ID | Priority | Severity | Error Code | Triggered At | Handler/Service | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| POS-JC-LINK | P1 | MEDIUM | JOB_CATALOG_NOT_FOUND | REST Position Create & Replace | `validator/position_assignment_validation.go` 链式 | Job Catalog Family/Role/Level 必须存在且处于 ACTIVE 状态，错误以 `JOB_CATALOG_NOT_FOUND` 暴露。 |
| JC-ACTIVE-LINK | P1 | MEDIUM | JOB_CATALOG_NOT_FOUND | REST Job Catalog 关联入口（职位、任职链） | `validator/position_assignment_validation.go:newPosJobCatalogRule` | 复用 POS-JC-LINK，强调 Job Catalog 节点需处于 ACTIVE；219C2Y 自测脚本记录审计上下文。 |
| JC-TEMPORAL | P1 | HIGH | JOB_CATALOG_TEMPORAL_CONFLICT / INVALID_EFFECTIVE_DATE / JOB_CATALOG_TIMELINE_UNAVAILABLE | REST `POST /api/v1/job-*/{code}/versions` | `internal/organization/validator/job_catalog_validation.go:newJobCatalogTemporalRule` | 检测 Job Catalog 版本时间冲突及时间线加载失败，要求新版本生效日严格大于最新版本并记录 `latestEffective`、`attemptedEffective`。 |
| JC-SEQUENCE | P1 | MEDIUM | JOB_CATALOG_SEQUENCE_MISMATCH / JOB_CATALOG_SEQUENCE_MISSING_PARENT / JOB_CATALOG_SEQUENCE_MISSING_BASE | REST Job Catalog 版本补充 | `internal/organization/validator/job_catalog_validation.go:newJobCatalogSequenceRule` | 保证版本父子链连续性：要求 `parentRecordId` 匹配最新版本 `recordId`，若时间线为空直接阻断并提示补齐基线。 |

### 错误码与契约对齐
- 错误码来源：`docs/api/openapi.yaml` `components.responses.BadRequest.examples` 以及具体端点的 `4xx/5xx` 响应。
- 新增错误码流程：提交契约补丁（OpenAPI + GraphQL Schema，如适用）→ 获架构/安全组确认 → 更新本节表格 → 更新实现。
- 审计联动：验证失败需调用 `audit.LogError`，并将 `ruleId`、`severity`、`payload` 写入 `business_context`。

### 实现清单引用
- 计划执行完毕后，在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 添加 “Business Validator Chains” 草稿条目并指向本节，确保唯一事实来源链路完整。

### 链式执行骨架
- `internal/organization/validator/core.go` 提供 `ValidationChain` 与 `Rule` 注册机制，支持优先级排序、短路控制与 `RuleOutcome` 聚合。
- `Rule.Severity` 仅允许 `CRITICAL/HIGH/MEDIUM/LOW`，缺省自动归约为 `HIGH`；`SeverityToHTTPStatus` 用于 REST/GraphQL 统一状态码映射。
- `TestValidatorCoreSmoke`（同目录）验证链路装配、短路行为与执行顺序，作为最小回归脚本；执行记录保存在 `logs/219C2/test-Day21.log`。
- 链式执行上下文统一写入 `ValidationResult.Context.executedRules`，方便 handler/审计层追踪执行路径。

### 错误翻译与审计联动
- `internal/organization/handler/organization_helpers.go:writeValidationErrors` 负责提取首条违规、汇总所有错误/警告，并将 `ruleId`、`severity`、`metadata` 映射为统一响应结构。
- 严重级别通过 `validator.SeverityToHTTPStatus` 转换为 HTTP 状态码；当规则上下文提供 `ruleId` 时会写入响应 `details` 及审计 payload。
- 审计链路调用 `audit.LogError`，并将 `ruleId`、`severity`、`payload` 注入 `business_context`，满足 219C2A 对验证失败的追踪与唯一事实来源要求。

### 验证链可观测性（219C2D）
- 指标注册位置：`internal/organization/validator/metrics.go`，在第一次执行验证链时通过 `prometheus.MustRegister` 完成注册。
- 采集项：
  - `validator_rule_duration_seconds{rule_id}` — 单条规则执行耗时直方图。
  - `validator_rule_outcome_total{rule_id, outcome}` — 规则执行结果计数（`success|warning|failed|error`）。
  - `validator_chain_duration_seconds{operation}` — 链路整体耗时直方图。
  - `validator_chain_outcome_total{operation, outcome}` — 链路执行结果计数（`success|failed|cancelled`）。
- `WithOperationLabel`/`WithBaseContext` 负责注入 `operation` 标签；Position/Assignment/Job Catalog/Organization 各入口均已设置，查询示例：
  ```bash
  curl -s http://localhost:9090/metrics | rg validator_chain_outcome_total
  ```
  PromQL 示例：`sum by (rule_id, outcome) (rate(validator_rule_outcome_total[5m]))`。

### 组织域规则实现
- `internal/organization/validator/organization_rules.go` 封装 ORG-DEPTH / ORG-CIRC / ORG-STATUS / ORG-TEMPORAL 规则，链路按照优先级（10/20/30/25）执行，支持短路与上下文聚合。
- ORG-TEMPORAL 规则在父组织指定时会检查时态有效性，父组织在指定生效日缺失返回 `INVALID_PARENT`，存在但非激活态返回 `ORG_TEMPORAL_PARENT_INACTIVE`。
- `ValidateOrganizationCreation` / `ValidateOrganizationUpdate` 通过链式执行替换旧的散落校验，并保留代码唯一性、业务逻辑等增量验证。
- Handler 层（`organization_create.go`, `organization_update.go`, `organization_history.go`）在持久化前调用验证链，失败时统一触发审计与结构化错误响应。

### 职位与任职规则实现（219C2C）
- `internal/organization/validator/position_assignment_validation.go` 提供 `NewPositionAssignmentValidationService`，在命令模块中为职位/任职命令注入统一链式校验（见 `internal/organization/api.go`）。
- 职位规则：`POS-ORG` 校验引用组织 ACTIVE；`POS-HEADCOUNT` 在填充/更新任职前验证编制；`POS-JC-LINK` 校验 Job Catalog 链路，违反时返回 `JOB_CATALOG_NOT_FOUND`。
- 任职与跨域规则：`ASSIGN-FTE`、`ASSIGN-STATE`、`CROSS-ACTIVE` 在创建/更新/关闭任职时执行，阻断非法状态与跨域激活冲突。
- 单元测试位于 `internal/organization/validator/position_assignment_validation_test.go`，覆盖职位/任职正反场景并记录于 `logs/219C2/test-Day23.log`，最新补充的 helper/stub 测试（见 Day24 更新）将包覆盖率提升至 **83.7%**。
- 命令服务默认依赖链式验证：`PositionService` 在写库事务前执行链路，REST handler 通过 `ValidationFailedError` 捕获返回结构化错误并同步审计上下文。
- Job Catalog 规则（`JC-TEMPORAL` / `JC-SEQUENCE`）由 `internal/organization/validator/job_catalog_validation.go` 提供，命令层在 `JobCatalogService` 中调用；配套单元测试位于 `job_catalog_validation_test.go`，Day24 运行记录见 `logs/219C2/test-Day24.log`，`internal/organization/validator` 包覆盖率提升至 **85.3%**，满足 219C2D 覆盖率基线。
- E2E 自测完成（2025-11-06 21:53）：脚本 `scripts/219C2D-validator-self-test.sh` 执行通过，覆盖 Job Catalog/Position/Assignment 关键验证场景，生成双通道测试报告 `tests/e2e/organization-validator/report-Day24.json`，日志见 `logs/219C2/validation.log`。
- REST 命令补测（219C3）：脚本 `scripts/219C3-rest-self-test.sh` 聚焦 `createPosition` / `fillPosition` / `closeAssignment` 等命令与审计验证，输出 `logs/219C3/validation.log` 供计划验收与归档。
