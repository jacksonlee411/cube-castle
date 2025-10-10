# 60号文档：系统级质量整合与重构计划

**版本**: v1.1
**创建日期**: 2025-10-10
**最后更新**: 2025-10-10
**维护团队**: 架构组 + 后端团队 + 前端团队 + 平台/DevOps团队
**状态**: 规划中
**遵循原则**: CLAUDE.md 资源唯一性与跨层一致性原则（最高优先级）

## 背景与唯一事实来源
- 本计划汇总 50~59 号质量分析文档（命令服务处理器、服务层、中间件、共享模块、GraphQL 适配、前端配置/API/Hooks/工具）的调查结论，全部内容直接来源于这些已发布分析与对应源码，实现资源唯一性。
- 对照 `docs/api/openapi.yaml`、`docs/api/schema.graphql`、设计系统 Canvas Tokens 及项目运行手册，确认当前发现与契约、设计规范保持一致；未引入第二事实来源。
- 启动各阶段前、验收后需执行 `node scripts/generate-implementation-inventory.js` 并与 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 对照，确保未新增重复实现；如存在差异需先补齐实现或更新清单。
- 若重构影响开发/运行手册，须同步修订 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`，并在本计划验收记录中注明更新位置。

### 关联质量分析文档清单
本计划基于以下质量分析文档的发现，确保所有改进措施有据可查：
- **[50号]** 命令服务处理器质量复盘 (`docs/development-plans/50-organization-command-handlers-quality-review.md`)
- **[51号]** 服务层质量分析 (`docs/development-plans/51-services-quality-analysis.md`)
- **[52号]** 中间件质量分析 (`docs/development-plans/52-middleware-quality-analysis.md`)
- **[53号]** 共享模块质量分析 (`docs/development-plans/53-shared-modules-quality-analysis.md`)
- **[54号]** 查询服务 GraphQL 中间件质量分析 (`docs/development-plans/54-query-service-graphql-middleware-quality-analysis.md`)
- **[55号]** 共享中间件质量复盘 (`docs/development-plans/55-shared-middleware-quality-review.md`)
- **[56号]** 前端配置质量分析 (`docs/development-plans/56-frontend-config-quality-analysis.md`)
- **[57号]** 前端 API 与类型质量分析 (`docs/development-plans/57-api-and-types-quality-analysis.md`)
- **[58号]** 前端 Hooks 质量分析 (`docs/development-plans/58-hooks-quality-analysis.md`)
- **[59号]** 工具与验证模块质量分析 (`docs/development-plans/59-tools-and-validation-quality-analysis.md`)

## 契约与事实来源映射
- **REST 命令契约**：`docs/api/openapi.yaml` 为唯一真源，当前由 `cmd/organization-command-service/internal/types/models.go`、`cmd/organization-command-service/internal/types/responses.go` 和 `frontend/src/shared/types/organization.ts`、`frontend/src/shared/validation/schemas.ts` 手动对齐，配合 `scripts/check-api-naming.sh`、`scripts/generate-implementation-inventory.js` 做一致性巡检。
- **GraphQL 查询契约**：`docs/api/schema.graphql` 为唯一真源，对应实现集中在 `cmd/organization-query-service/internal/graphql/resolver.go` 与 `frontend/src/shared/api/graphql-enterprise-adapter.ts`；当前采用人工校对，计划在阶段一补齐自动校验脚本。
- **设计系统令牌**：`frontend/src/design-system/tokens/brand.ts` 与 `@workday/canvas-kit-react/tokens` 为官方定义，前端复用点集中在 `frontend/src/shared/utils/statusUtils.ts` 与 `frontend/src/shared/utils/colorTokens.ts`。
- **业务枚举与约束**：以 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 为真源，具体枚举落地于 `cmd/organization-command-service/internal/types/models.go` 与 `frontend/src/shared/types/organization.ts`，并通过 `frontend/src/shared/validation/schemas.ts` 校验输入输出。
- **审计字段规范**：以 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 和 `cmd/organization-command-service/internal/repository/audit_writer.go` 为真源。

> 阶段一将补齐契约校验脚本并纳入 CI（计划命名为 `contract-sync`）。在脚本就绪前，契约变更需先更新上述权威文件，再通过人工校对和 `scripts/check-api-naming.sh` 复核。

## 系统性问题总览
- **契约漂移**：组织层级上限、UnitType/Status 枚举、请求 ID、错误码等在后端校验、前端常量、Hook/工具之间多次不一致，合法请求常被拒或响应无法解析。
- **重复实现与“企业化”空壳**：响应封装、错误处理、时间轴转换、限流、端口探测等能力在多个文件夹重复；Dev/Operational handler、GraphQL 企业信封等标称企业化的模块实际仅保留 TODO 或占位逻辑。
- **类型与可观察性脆弱**：广泛使用 `map[string]interface{}`、字符串匹配判断错误、审计记录缺字段、请求 ID 在中间件链路丢失，导致运行时异常与追踪断裂。
- **生命周期治理不足**：调度与级联服务、限流清理、Hook 状态定时器缺乏重启/取消机制，日志与监控信号不可靠，影响运维。
- **环境与工具混乱**：前端配置访问 `process.env`/`localStorage` 无校验，测试/脚本重复造轮子；DevTools/Operational API 未设置白名单与真正执行逻辑，扩大攻击面。

## 重构目标
1. **契约真源统一**：以 OpenAPI/GraphQL 规范与设计系统为唯一事实来源，驱动后端校验、前端常量、枚举/类型生成及测试快照，消除跨层漂移。
2. **能力收敛与模块精简**：统一响应与错误处理栈、时间轴/时态服务、限流与性能监控实现，删除或合并重复工具和“空壳”企业能力。
3. **提升类型安全与可观察性**：引入结构化 DTO/类型、强化请求 ID/审计字段链路，标准化日志与监控输出，确保关键链路可追踪。
4. **完善生命周期管理**：为后台任务、限流清理、Hook 定时器等提供显式启动/停止、上下文取消与资源清理，避免泄漏与竞态。
5. **稳固环境与安全界面**：提供浏览器/SSR/Node 安全访问层，强化 Dev/Operational 工具白名单与速率控制，缩小攻击面。

## 分阶段改进计划

### 第一阶段：契约与类型统一（优先级高，预估 2 周）
- **范围与影响**
  - 涉及文件：计划新增的 `scripts/contract/*`、现有的 `cmd/organization-command-service/internal/utils/validation.go`、`cmd/.../validators/business.go`、`frontend/src/shared/config/constants.ts`、`frontend/src/shared/types/`、`frontend/src/features/organizations/constants/*` 等。
  - 输出工件：计划生成统一的 `shared/contracts/organization.json`、Go/TS 代码与 Vitest/Go 快照测试，作为后续一致性校验依据。
- **前置条件**
  - OpenAPI/GraphQL 契约在 `docs/api/` 中的变更已获架构负责人审批并提交主干。
  - 相关计划文档（如 53、56 号）无待解决阻塞项，执行成员完成契约同步工具的培训演示。
- **关键任务**
  1. 建立契约同步脚本（脚本命名：`scripts/contract/sync.sh`），将 OpenAPI/GraphQL/Canvas Token 转译成 JSON 中间层，再生成 Go/TS 常量与类型。
  2. 调整现有校验逻辑，删除本地硬编码枚举，统一读取生成工件；修正组织层级上限为 17，并对 `docs/reference` 中相关表格做同步更新。
  3. 为时间轴响应、审计记录补齐契约字段（如 `requestId`、`actorType`），并在 `tests/contract` 目录新增跨层快照测试。
  4. 阶段首尾执行 `node scripts/generate-implementation-inventory.js`，比对实现清单并记录结果。
- **验证方式**
  - 新增 CI Job：`contract-sync`（校验生成文件与仓库一致）、`contract-snapshot`（比较 Go/TS 快照）。
  - 手工验收：命令服务创建/更新组织流程回归（REST + GraphQL），并确认 `docs/reference` 更新已提交。
  - 阶段结束前运行 `make test`、`make lint`、`npm run lint`，确保契约调整未破坏现有门禁。

### 第二阶段：后端服务与中间件收敛（预估 3 周）
- **范围与影响**
  - 核心模块：`TemporalService`、`OrganizationTemporalService`、`TemporalTimelineManager`、`DevToolsHandler`、`OperationalHandler`、中间件链路（`internal/middleware`）。
  - 关联测试：`cmd/organization-command-service/internal/services/*_test.go`、`tests/integration/temporal/*`。
- **前置条件**
  - 阶段一交付物已合并主干并在 CI 稳定运行 ≥ 48 小时。
  - Temporal 相关数据库迁移与备份已完成（可通过 `make backup` 或运维预留的备份脚本执行），确保回滚时有可用快照。
- **关键任务与迁移策略**
  1. 先抽取共享事务与审计封装（`internal/services/temporal_transaction.go`），在旧服务中引入，再逐个方法切换至新封装，实现“双写+比对”日志。
  2. 定义统一响应/错误结构体，替换 REST/GraphQL 层散落的写法；为 Dev/Operational Endpoint 制定白名单与角色校验（参考 OpenAPI scopes）。
  3. 引入 Prometheus/Otel 中间件与 `golang.org/x/time/rate` 限流器，并记录请求 ID，提供回滚开关（环境变量 `TEMPORAL_REFACTOR_ENABLED`）。
  4. 阶段验收必须更新相关运行手册片段（若有影响），并执行 `make test`、`make lint`、`make test-integration`。
- **验证方式**
  - 阶段性灰度：在 `make run-dev` 中启用双写，比较新旧时间轴数据日志。
  - 新增监控：Prometheus 指标 `temporal_transaction_duration_seconds`、`devtools_requests_total`。
  - 阶段收尾再次运行 `node scripts/generate-implementation-inventory.js` 校报表，并确认 `docs/reference` 更新已合入。

### 第三阶段：前端 API/Hooks/配置整治（预估 2-3 周）
- **范围与影响**
  - 模块：`frontend/src/shared/api/*`、`frontend/src/shared/hooks/*`、`frontend/src/shared/config/*`、`frontend/tests/`。
  - 向后兼容：保留旧 API 客户端作为适配层，逐模块切换。
- **前置条件**
  - 第二阶段发布的统一错误结构与审计字段已在 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 更新说明。
  - JWKS/认证模拟服务 (`make run-auth-rs256-sim`) 与前端本地开发脚手架在 CI `make frontend-dev` 任务中通过健康检查。
- **关键任务**
  1. 统一 React Query 客户端（计划新增 `shared/api/queryClient.ts`），建立标准错误包装（含请求 ID 与错误码），通过 Feature Flag `QUERY_REFACTOR_ENABLED` 控制生效。
  2. 调整 Hooks：先迁移查询（`useOrganizationsQuery` 等），再迁移写操作；如需兼容旧实现，补充临时桥接层（命名建议 `legacyOrganizationApi`），并在文档中标记废弃计划。
  3. 重写端口/环境助手，新增 SSR/Node 安全访问封装，运行时代码中的报告函数迁移到脚本目录（计划新增 `scripts/report/*`）离线执行。
  4. 更新前端相关参考文档（如配置说明、错误处理策略），并在阶段起止执行 `npm run lint`、`npm run test`、`frontend/scripts/validate-field-naming*.js`、`node scripts/quality/architecture-validator.js`。
- **验证方式**
  - 前端测试：Vitest 覆盖率保持 ≥ 75%，Playwright 冒烟场景通过。
  - QA 验收：组织管理关键路径手工巡检，确认错误提示、重试机制正常。
  - 阶段结束前对照 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`，确保前端命令/脚本说明同步。

### 第四阶段：工具与验证体系巩固（预估 1-2 周）
- **范围与影响**
  - 工具链：Temporal/Validation helper、审计写入、设计令牌同步脚本、CI 任务。
- **前置条件**
  - 前三阶段在 staging 环境稳定运行 ≥ 7 天且无高优告警。
  - `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 已更新至最新模块清单。
- **关键任务**
  1. 将 Temporal/Validation 工具折叠至单一实现，提供 `deprecated` 别名，在两个迭代后移除；新增 `docs/development-plans/xx-temporal-tool-migration.md` 记录迁移。
  2. 审计字段统一采用结构化 DTO，回补 `OldValue/NewValue` 等字段，增加集成测试验证。
  3. 构建长期守护：CI 新增 `lint-contract`, `lint-audit`, `doc-archive-check`，防止新增漂移，并更新参考文档的相关章节。
- **验证方式**
  - 运行 `make security`, `npm run lint`, `node scripts/quality/architecture-validator.js` 全绿。
  - 审计链路在日志与数据库记录中均可查到完整字段。
  - 阶段闭环再执行 `node scripts/generate-implementation-inventory.js` 并记录确认结果。

## 关键里程碑与成果指标
- **M1 契约统一完成**
  - `contract-sync`、`contract-snapshot` CI 必须稳定通过。
  - `tests/contract/*.snap` 更新记录签字，组织层级 17、状态枚举一致。
  - 命令服务创建组织接口（REST/GraphQL）回归通过，错误码与文档一致。
- **M2 服务与中间件整合**
  - 双写期间新旧时间轴数据 diff 为 0，Prometheus 仪表板展示延迟 < 200ms。
  - Dev/Operational API 引入白名单、速率限制，安全测试（渗透脚本）通过。
  - 日志中请求 ID 覆盖率达到 100%，审计表字段完整。
- **M3 前端客户端与 Hook 重构**
  - React Query 查询命中率日志 ≥ 90%，本地与 CI JWKS 自动刷新成功。
  - Playwright 冒烟套餐（组织创建/更新/时态查询）全绿。
  - 运行时代码包体积下降 ≥ 5%，报告型函数移至脚本目录。
- **M4 工具栈巩固**
  - `lint-contract`、`lint-audit`、`doc-archive-check` 纳入 CI。
  - 审计记录含 `oldValue/newValue/requestId` 等字段，集成测试覆盖。
  - Temporal/Validation 旧别名标记废弃并在 2 个迭代后删除。

## 验收标准
- [ ] 契约同步脚本落地，Go/TS 校验与常量引用统一，快照测试覆盖 `UnitType`、状态、组织层级等关键枚举。
- [ ] 后端响应/错误封装统一，Dev/Operational/GraphQL 中间件具备真实执行与安全限制，请求 ID 与日志链路完整。
- [ ] 前端 API/Hooks 重构后编译通过、React Query 缓存逻辑合理、错误处理产出结构化信息；JWKS/环境适配支持本地与 CI。
- [ ] Temporal/Validation 工具仅保留一套核心实现，审计/仓储字段完整；颜色/状态引用基于 Canvas Token。
- [ ] CI 增设契约/类型一致性检查，实施过程中定期执行 `node scripts/generate-implementation-inventory.js`、`make test`、`make lint`、`npm run lint`、`frontend/scripts/validate-field-naming*.js` 等门禁命令；相关文档（本计划及 50~59 号）在完成阶段性目标后归档至 `docs/archive/development-plans/` 并更新验收记录。

## 风险与回滚策略
- **阶段门禁**
  - 每个阶段完结后需在 staging 环境观察 ≥ 48 小时（阶段四延长至 7 天）且无高优告警，再推进下一阶段。
  - 回滚路径：阶段一恢复上一个契约快照；阶段二通过计划新增的环境变量 `TEMPORAL_REFACTOR_ENABLED=false` 切换旧实现；阶段三通过 Feature Flag `QUERY_REFACTOR_ENABLED=false` 与计划中的 `legacyOrganizationApi` 适配层恢复旧客户端；阶段四保留 `deprecated` 别名与脚本备份。
- **灰度双写与监控**
  - Temporal/审计模块迁移期间计划新增双写日志（建议命名 `logs/temporal-doublewrite.log`）及比对脚本（计划路径 `scripts/quality/check-doublewrite.js`）；在工具就绪前先人工比对数据库记录，一旦差异 ≠ 0 即暂停发布。
  - 新中间件上线前先在 dev 环境运行 24 小时，监控 Prometheus 指标 `temporal_transaction_duration_seconds`、`devtools_requests_total`；出现异常峰值即回滚。
- **安全与配置**
  - 新增限流器、白名单、Feature Flag 必须记录在 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 并由平台团队复核。
  - 前端环境工具重构前保留现有适配层（必要时新增 `legacyEnvironmentAdapter` 并同步说明），Playwright/Vitest 未全部通过前不得移除；所有配置切换需记录在现有文档或新增的 `ENV_CHANGELOG.md`。

## 安全与合规检查清单
- **鉴权**：Dev/Operational Handler 必须校验 JWT、角色/Scope；执行命令需二次确认或审计记录。
- **速率限制**：REST/GraphQL/DevTools 所有入口绑定统一限流器（突发 + 均值），默认阈值写入配置文件。
- **审计**：每个敏感操作写入 `audit_entries`，包含 `tenantId`、`resourceId`、`requestId`、`actorType`。
- **日志**：强制携带 `requestId`、`tenantId`、`spanId`，禁止输出敏感字段（clientSecret 等）。
- **运行环境**：前端配置访问 `localStorage`/`process.env` 前先判空，SSR/Node 环境 fallback；脚本访问密钥统一使用 `secrets/`。
- **文档治理**：阶段完成后同步更新 `docs/archive/development-plans/60-system-wide-quality-refactor-plan.md`，并在 `CHANGELOG.md` 记录关键变更。

## 一致性与变更治理
- 所有改动需同步检查 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 与设计令牌，保持跨层一致；若契约需调整，先更新规范再落地实现。
- 每个阶段结束前更新对应计划文档并归档旧版本，维持唯一事实来源链路；提交信息按编号引用（如 “ref: plan-60 M1”）。
- 变更期间执行 `node scripts/generate-implementation-inventory.js` 与 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 对照，确保无轮子重复引入，必要时同步更新参考文档并在验收记录中写明版本号。

## 后续行动
- 指派跨团队小组按阶段推进（后端、前端、平台/DevOps），每阶段明确负责人、验收人以及回滚责任人。
- 将本计划纳入迭代 backlog，发布迭代节奏（例如双周同步会），并在每次阶段启动前开展 30 分钟培训或演练。
- 整理《系统级重构 FAQ》（归档于 `docs/development-plans`），记录常见问题、临时方案与回滚路径，并在执行完毕后更新运行手册。
- 计划执行过程中若出现与其他计划冲突，需优先以 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`CLAUDE.md` 指南为准，必要时更新本计划并通知相关团队。
