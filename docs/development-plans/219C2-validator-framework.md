# Plan 219C2 – Business Validator 框架扩展

**文档编号**: 219C2  
**上级计划**: [219C – Audit & Validator 规则收敛](./219C-audit-validator.md)  
**目标周期**: Week 4 Day 21-24  
**负责人**: 组织后端团队（安全架构组评审）  

---

## 1. 目标

1. 定义可组合的 `BusinessValidator` 接口（统一 `ValidationState` 输入、区分错误/警告），支持链式执行与规则分组。
2. 将现有组织校验逻辑拆分为规则清单，并扩展至职位、Assignment、Job Catalog，覆盖跨域依赖（组织状态、headcount、Job Catalog 连续性）。
3. 在命令服务入口注入校验链，确保命令在事务开始前完成业务验证并返回结构化错误码。
4. 完成单元测试矩阵，确保关键规则（层级循环、headcount 越界、Job Catalog 生效区间冲突、Assignment 状态流转）均有覆盖。

---

## 1.1 前置条件（P0 必须满足）

- 219C1 审计基础设施已验收，`requestId` / `correlationId` 贯通，`LogError` 事务化写入稳定。
- 规则矩阵与错误码在启动前即冻结：与架构/安全组共同确认 Rule ID、Severity、HTTP Status、错误码映射（详见 §3.2）。
- OpenAPI 契约（`docs/api/openapi.yaml`）中缺失的错误码已补齐或提交契约变更；冻结表格见 §3.2。
- README 已预留 `#validators`、`#validator-test-strategy`、`#validator-implementation-checklist` 小节，用作唯一事实来源。
- 219C 总计划时间轴调整为：Day 21 启动 219C2，Day 24 为缓冲与验收归档日。
- **NEW: 架构 PoC 在 Day 21 早晨完成并通过（详见 §3.0），验证链式执行、工厂模式、性能基准（< 10ms）。**

---

## 2. 范围

| 模块 | 内容 |
|---|---|
| `internal/organization/validator` | 抽象接口、实现链式容器、细分规则集（Organization/Position/Assignment/JobCatalog/CrossDomain）。 |
| `internal/organization/service` | 服务层创建验证上下文对象，执行校验链并将错误转换为领域错误。 |
| `internal/organization/handler` | 在 REST/GraphQL 命令入口，将校验结果转换为标准错误响应，保持 camelCase 字段。 |
| 文档 | `internal/organization/README.md#validators` 增补规则矩阵、严重级别、对应命令；在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 添加 “Business Validator Chains” 条目。 |
| 测试 | 表驱动单元测试，按模块放置到 `internal/organization/validator/*_test.go`。 |

---

## 3. 详细任务

## 3.0 架构 PoC 与假设验证（Day 21 早晨，预算 30 分钟）

在启动正式实施前，需通过 PoC 验证以下关键假设：

1. 链式验证执行器可支持优先级排序 + 短路 + 错误聚合
2. 工厂模式可在现有 DI 框架下集成（复用 `internal/organization/api.go`）
3. 链式执行 10 条规则的耗时 < 10ms（性能可接受）

### PoC 实施范围

创建临时 PoC 代码，验证上述假设（预算 15-20 分钟）：

```go
// internal/organization/validator/validator_chain_poc_test.go
// 实现单一规则（如 ORG-DEPTH）的链式执行
// 验证工厂函数的依赖注入可行性
// 使用 mock 仓储，避免数据库依赖
// 运行 benchmark，输出耗时
```

### PoC 验收标准

- [ ] 链式执行正确识别规则违反（如 depth > 17）
- [ ] 错误应答序列化为 JSON，无 panic
- [ ] 工厂函数成功注入 4 个不同的仓储实例（Hierarchy, Organization, Position, JobCatalog）
- [ ] 性能基准：10 条规则的链式执行耗时 < 10ms

### 失败处理

若 PoC 失败，**中止其他任务**：
1. 与架构组进行 30 分钟的紧急评审
2. 调整设计或修改计划假设
3. 记录"PoC 失败调整计划"到 219C 总计划
4. **不允许带着不确定性进入 A1**

### PoC 输出

- [ ] PoC 代码：`internal/organization/validator/validator_chain_poc_test.go`
- [ ] 性能基准：`logs/219C2/poc-perf.log`
- [ ] 评审纪要：`logs/219C2/poc-review.md`（若失败）

---

### 3.1 子任务拆分（关键节点）

为降低协调成本，本计划以 4 个关键节点推进，每个节点包含“实现 + 测试 + 文档”的完整交付，并在 219C 主计划中记录验收凭证。

| 子编号 | 里程碑日程 | 范围 | 主要输出 | 依赖 |
| --- | --- | --- | --- | --- |
| **219C2A – 框架基座** | Day 21 | ✅ 已完成：链式骨架、handler 集成、文档/日志同步 | `internal/organization/validator/core.go`、`organization_helpers.go`、README `#validators`、`logs/219C2/rule-freeze.md` | 219C1 + §3.0 PoC ✓ |
| **219C2B – 组织域规则** | Day 22 | 将组织规则迁移至链式实现，接入 REST/GraphQL，补齐 P0 单测与错误码对齐 | `validator/organization_*`、命令入口改造、表驱动单测、自测记录 | 219C2A |
| **219C2C – 职位与跨域规则** | Day 23 | 实现职位、Assignment 基础与跨域规则，统一命令入口，补齐 P0/P1 单测 | `validator/position_*`、`validator/assignment_*`、命令改造、测试日志 | 219C2B |
| **219C2D – 扩展与验收** | Day 24 | 完成 Job Catalog 规则、端到端测试、Implementation Inventory 更新与归档 | `validator/job_catalog_*`、端到端测试输出、README 更新、档案记录 | 219C2C |

**节点说明**
- 219C2A 仅在复用现有 `BusinessRuleValidator` / `ValidationResult` 不足时列出最小改造清单（≤3 个方法签名变更），确保不引入第二套结构。
- 219C2B/219C2C 将“规则实现、命令接入、单测”同周期完成，避免跨节点漂移。
- 219C2D 汇总全部文档、Implementation Inventory 更新与归档操作，确保唯一事实来源同步。

### 关键路径与同步

- 串行关键路径：`PoC → 219C2A → 219C2B → 219C2C → 219C2D`。
- 并行机会：219C2D 中端到端测试与文档更新可在 Day 24 PM 并行推进，但须在归档前完成交叉验证。
- 时间对齐（Week 4 Day 21-24）：
  - Day 21：09:00-09:30 PoC；09:30-12:00 完成 219C2A 主体；13:00-17:00 完成 handler 集成与 README 骨架。
  - Day 22：08:30-17:30 完成组织规则迁移、命令接入及单测。
  - Day 23：08:30-17:30 完成职位/Assignment 规则与命令接入，补齐跨域测试。
  - Day 24：AM 执行 Job Catalog 规则与端到端测试，PM 完成文档同步、归档与缓冲（预留 2 小时应急）。

#### 3.1.1 子计划：219C2A – 框架基座（日程 Day 21）

> 详见独立计划：[219C2A – Validator 框架基座](./219C2A-framework-foundation.md)

- 前置条件：§3.0 PoC 四项验收通过；219C1 审计链路可用；OpenAPI 错误码补丁已合入主线。
- 工作项：
  1. 评估现有 `BusinessRuleValidator` / `ValidationResult` 复用范围，必要时列出≤3 项最小化改造清单，溯源至架构评审纪要。
  2. 在 `internal/organization/validator/core.go` 实现链式执行入口与短路控制，保持现有结果结构不变。
  3. 更新 `internal/organization/handler/organization_helpers.go` 以复用统一错误翻译、审计上下文包装。
  4. 在 `internal/organization/README.md#validators` 创建框架说明、规则登记模板，并提交 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 草稿条目。
  5. 与安全/架构组冻结规则矩阵与错误码映射，输出 `logs/219C2/rule-freeze.md` 纪要。
- 交付物：增量代码、README 模板、Implementation Inventory 草稿、冻结纪要。
- 验收：接口改造清单获架构签字；新增 smoke 测试通过（`go test ./internal/organization/validator -run TestValidatorCoreSmoke`）；README 段落合并；rule-freeze 纪要归档。
- 风险：若需要新增 `ValidationState` 等新结构，必须先经架构批准；若链式接口破坏兼容，立即回滚并记录实验。

#### 3.1.2 子计划：219C2B – 组织域规则（日程 Day 22）

> 详见独立计划：[219C2B – 组织域规则迁移](./219C2B-organization-rules.md)

- 前置条件：219C2A 已验收；README 中 P0/P1 列表冻结；命令服务已能加载工厂但未启用新链。
- 工作项：
  1. 将组织域校验拆分为独立规则文件，覆盖所有组织相关 Rule ID（P0/P1）。
  2. 为每条规则编写表驱动单测并使用 stub 仓储，确保无数据库依赖，目标覆盖率≥85%。
  3. 在 REST/GraphQL handler/service 注入统一验证链，移除旧散落校验，保持错误码与 OpenAPI 一致。
 4. 执行 Create/Update Organization 命令自测，验证错误返回与审计日志（business_context.ruleId）。
 5. 更新 `logs/219C2/daily-YYYYMMDD.md` 记录完成度、风险与延迟。
- 备忘（Day 22）：已在 OpenAPI 中登记 ORG-TEMPORAL 规则与 `INVALID_PARENT`/`ORG_TEMPORAL_PARENT_INACTIVE` 错误码，REST 命令入口全面依赖验证链输出统一错误结构；GraphQL Mutation 接入与字段级错误码枚举扩展纳入 219C2C 后续跟进。REST 自测与审计凭证已由 Team 06 提交，详见 `logs/219C2/219C2B-SELF-TEST-REPORT.md`。
- 交付物：组织域规则实现与单测、命令接入改造、自测日志、日同步记录。
- 验收：`go test -cover ./internal/organization/validator` ≥85%；REST/GraphQL 错误码一致；审计日志写入 ruleId；日同步提交。
- 风险：旧逻辑残留导致重复校验；若覆盖率未达标需在当日 17:00 前上报并申请 Day 23 上午补测。

#### 3.1.3 子计划：219C2C – 职位与跨域规则（日程 Day 23）

> 详见独立计划：[219C2C – 职位与跨域规则落地](./219C2C-position-crossdomain.md)

- 前置条件：组织链稳定；Position/Assignment 服务接口可被验证链装配；跨域依赖 stub 已就位。
- 工作项：
  1. 实现 Position 与 Assignment 基础规则及跨域检测 helper，覆盖 P0 规则并补足 P1 单测。
  2. 将 Position/Assignment REST 与 GraphQL 命令接入统一验证链，保留事务回滚。
  3. 运行 Fill/TransferPosition 等关键命令自测，确认错误码、审计上下文一致。
  4. 梳理跨域依赖清单（Hierarchy、JobCatalog 等）并确认仓储接口满足注入需求。
  5. 若发现新增业务规则需求，提交 219C2D 或 219E 变更申请。
- 交付物：Position/Assignment 规则代码与单测、命令入口改造、自测日志、跨域依赖清单。
- 验收：`go test ./internal/organization/validator -run TestPosition -run TestAssignment` 全部通过；关键命令自测截图归档；依赖清单获数据团队确认。
- 风险：跨域仓储缺失导致注入失败；Assignment 状态流转复杂需及时与业务对齐。
- ✅ 219C2C 验收完成（2025-11-08）：`logs/219C2/test-Day24.log` 覆盖率 83.7%，`logs/219C2/acceptance-precheck-Day24.md` 记录 219C2Y 交付与风险，REST 自测补齐计划挂靠 219C2D。

#### 3.1.4 子计划：219C2D – 扩展与验收（日程 Day 24）

> 详见独立计划：[219C2D – 扩展与验收](./219C2D-extension-acceptance.md)

- 前置条件：219C2B/219C2C 验收完成；P0 单测覆盖率目标已达成；端到端环境（Docker compose）可用。
- 工作项：
  1. 实现 Job Catalog 规则链及 P1 单测，确认引用一致性与时态冲突处理。
  2. 编写 9 个端到端测试（REST/GraphQL × 3 命令 × 3 场景），输出报告至 `tests/e2e/organization-validator`。
  3. 汇总 README、Implementation Inventory、219C 主计划的验收勾选并提交归档。
  4. 召开 Day 24 日终验收会议，记录余量与滚动计划（若有）。
  5. 归档材料：`docs/archive/development-plans/219C2-YYYYMMDD.md`、`logs/219C2/validation.log`、端到端报告。
- 交付物：Job Catalog 规则代码与单测、端到端测试脚本与报告、文档更新、归档文件、验收纪要。
- 验收：端到端 9/9 通过；README 与 Implementation Inventory 提交合并；归档文件入库并在 219C 主计划登记；验收纪要签署。
- 风险：端到端测试依赖 Docker 服务，若阻塞>4 小时需动用缓冲并通知 219C 负责人。

### 3.2 规则实现矩阵（唯一事实来源）

唯一事实来源：`internal/organization/README.md#validators`。本计划不再维护独立矩阵，仅引用 README 中冻结的规则表，并在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 中登记“Business Validator Chains”条目。

- **P0 必须规则**（组织层级、循环、防激活、职位引用、Headcount、Assignment 状态与 FTE、跨域激活）——详见 README 表格中 `Priority=P0` 行，并同步错误码与 OpenAPI 状态。
- **P1 一致性规则**（时态父级、Job Catalog 链接、Job Catalog 时态冲突）——详见 README 对应行，必要时在 Day 23 PM 后补充实现。
- **P2 扩展规则**（如 ASSIGN-ACTING-AUTO）——交由后续 219E 计划执行，保留引用以便跟踪。

执行顺序、Severity、HTTP 状态与依赖仓储以 README 为准；任何变更必须先更新 README、提交架构评审，再同步此计划与 Implementation Inventory。若发现 OpenAPI 缺失错误码，应在进入 219C2B 之前完成契约补丁。

### 3.3 注入流程
- [ ] 在 REST/GraphQL handler/service 开始处构建校验上下文，调用统一验证链工厂，失败时阻断后续数据库操作。命令接入清单（执行时逐项勾选）：

| 模块 | 命令/操作 | 现有校验 | 调整 |
| --- | --- | --- | --- |
| 组织 | Create/Update Organization、CreateVersion、Suspend/Activate | `utils.Validate*` + 零散判断 | 复用 `validators.ForCreate/UpdateOrganization` 工厂，REST 与 GraphQL 共用 |
| 职位 | Create/Replace/Version/Update Position | Service 内手写校验 | 复用 `validators.ForCreate/ReplacePosition`，Service 仅持久化与审计 |
| Assignment | Fill/Create/Update/Vacate/Transfer Position | Service 层逻辑 | 引入 `validators.ForAssignment*` 链，并在 GraphQL 入口复用 |
| Job Catalog | Create/Update JobFamily/Role/Level（含版本） | Repo/Service 零散校验 | Handler/service 调用 `validators.ForJobCatalog*`，Repo 聚焦数据一致性 |
| GraphQL Mutation | `mutation createOrganization`、`mutation updateOrganization`、`mutation fillPosition`、`mutation transferPosition` 等 | 倚赖服务返回错误 | 调用与 REST 相同工厂，统一错误码与结构 |

- [ ] 与审计联动：校验失败时调用 `LogError`（复用 219C1 结果），`business_context` 中写入 `ruleId`、`severity`、`payload`（例如：`ruleId=POS-HEADCOUNT, severity=HIGH, details=...`）。
- [ ] 在 `internal/organization/api.go` 中注册验证链工厂，并确保 REST、GraphQL、批处理流程共用同一实例。

### 3.4 测试矩阵（优先级化）

**目标**：确保 P0 规则的可靠性，P1/P2 规则的基本覆盖。

#### 第 1 层：单元测试（必须）

**P0 规则**（8 条）：ORG-DEPTH, ORG-CIRC, ORG-STATUS, POS-ORG, POS-HEADCOUNT, ASSIGN-STATE, ASSIGN-FTE, CROSS-ACTIVE
- 每条规则：正向 1 个 + 反向 2-3 个样例（覆盖分支情况）
- 使用 stub 仓储，避免数据库依赖
- 表驱动测试，覆盖关键分支
- **目标覆盖率**：≥ 85%
- **时间**：Day 22-23 约 2-3 天
- **验收**：`go test -cover ./internal/organization/validator` ≥ 85%

**P1 规则**（3 条）：ORG-TEMPORAL, POS-JC-LINK, JC-TEMPORAL
- 每条规则：正向 1 个 + 反向 1 个样例
- **目标覆盖率**：≥ 70%
- **时间**：若 P0 提前完成，Day 23 PM 补充；否则延迟至 219E
- **验收**：单测覆盖正反场景

**P2 规则**（1 条）：ASSIGN-ACTING-AUTO
- 基本覆盖，非关键路径
- **时间**：可延迟至 219E
- **验收**：至少 1 个通过 + 1 个失败样例

#### 第 2 层：集成测试（REST/GraphQL 一致性，P0 优先）

仅为 P0 规则中的**关键命令**编写端到端测试，验证 REST 与 GraphQL 返回一致错误码：

- **CreateOrganization 场景**
  - 触发 ORG-DEPTH：REST 与 GraphQL 均返回 `DEPTH_EXCEEDED` / 400
  - 触发 ORG-CIRC：REST 与 GraphQL 均返回 `CIRCULAR_REFERENCE` / 400
  - 触发 ORG-STATUS：REST 与 GraphQL 均返回 `REFERENCE_INACTIVE` / 400

- **CreatePosition 场景**
  - 触发 POS-ORG：REST 与 GraphQL 均返回 `REFERENCE_INACTIVE` / 400
  - 触发 POS-HEADCOUNT：REST 与 GraphQL 均返回 `INVALID_HEADCOUNT` / 400

- **FillPosition 场景**
  - 触发 ASSIGN-STATE：REST 与 GraphQL 均返回 `INVALID_ASSIGNMENT_STATE` / 409
  - 触发 ASSIGN-FTE：REST 与 GraphQL 均返回 `INVALID_HEADCOUNT` / 400
  - 触发 CROSS-ACTIVE：REST 与 GraphQL 均返回 `REFERENCE_INACTIVE` / 400

**共 3 个关键命令 × 3 个错误场景 = 9 个端到端测试用例**
**时间**：Day 24 早晨 2-4 小时
**验收**：9 个测试全部通过，错误码与响应结构一致

#### 第 3 层：服务层回归（可选）

验证以下关键场景（若 Day 24 还有时间）：
- [ ] CreatePosition 触发 `POS-HEADCOUNT` → 返回 `INVALID_HEADCOUNT`，且未调用仓储写入
- [ ] 审计日志记录规则触发（`business_context.ruleId`, `severity`, `payload` 字段准确）
- [ ] 规则失败时事务正确回滚（使用事务化 mock）

**时间**：Day 24 PM（2-3 小时，非关键路径）
**验收**：测试通过，日志格式符合 §219C1 审计规范

#### 测试总工作量

- 单元测试：8×3 + 3×2 + 1×1 = 30 个用例 ≈ 16-20 小时
- 端到端测试：3×3 = 9 个用例 ≈ 4-6 小时
- 服务层回归（可选）：2-3 个用例 ≈ 2-3 小时

**优化建议**：
- 使用测试框架的 sub-test，减少重复代码
- 预置 stub 仓储工厂，加快测试编写速度
- 若时间紧张，P1/P2 规则延迟至 219E 可接受

---

## 4. 交付物

- 新的 validator 框架代码与注入逻辑（含链式工厂与依赖注入）。
- `internal/organization/README.md`（唯一事实来源）中更新的规则矩阵、执行顺序、测试策略与实现检查清单。
- 单元测试与预置伪造数据，`logs/219C2/validation.log` 存档关键场景输出。
- 错误码/错误消息对照表（与 `docs/api/openapi.yaml` 对齐），若有新增需附契约变更记录。
- 命令/操作接入勾选清单（REST/GraphQL 均勾选），并在 219C 主计划记录验收凭证。
- 更新后的 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 条目及文档归档记录。

---

## 4.1 验收标准（修订）

### 文档与契约
- [ ] 规则矩阵、测试策略、实现清单均更新于 README 对应小节，并指向唯一事实来源。
- [ ] 错误码映射表与 OpenAPI 契约对齐，缺失项已合并至 master 或附带变更申请。
- [ ] 219C 主计划与 219C2 子计划记录了验收凭证、规则冻结时间戳、执行日志路径。

### 实现与测试
- [ ] `go test -cover ./internal/organization/validator` 覆盖率 ≥ 80%，关键规则覆盖正/反场景。
- [ ] REST 与 GraphQL 端对端测试（P0 规则）返回一致的错误码与响应结构。
- [ ] 关键路径命令的审计日志 `business_context.ruleId`、`severity`、`payload` 字段准确记录规则触发。
- [ ] `logs/219C2/validation.log` 存档 REST/GraphQL/Service 测试输出。

### 归档与同步
- [ ] `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 新增 “Business Validator Chains” 条目。
- [ ] 完成后将本计划及执行记录归档至 `docs/archive/development-plans/`，并更新 219C 总计划进度。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| **PoC 失败验证不了假设** | 高 | Day 21 早晨立即评审，允许调整计划或推迟启动，不允许带着不确定性继续 |
| 规则执行顺序错误导致性能/体验问题 | 中 | 按 README `#validators` 中冻结的优先级配置链式执行，结合性能基准 < 10ms，必要时提供短路与计时指标 |
| Handler/Resolver 重复校验造成漂移 | 中 | 通过统一验证链工厂（§3.3），REST/GraphQL 共用实例并编写端到端回归 |
| 错误码与 OpenAPI 契约不一致 | 高 | 启动前完成错误码映射表（依 README 冻结矩阵），缺失项先提交契约补丁后再编码 |
| 规则范围变更触发超期 | 高 | 以 README `#validators` 为唯一来源冻结规则，新增规则需单独立项并更新计划 |
| 任务拆分过细导致协调成本 | 中 | 已合并为 4 个关键节点（§3.1），同步点 < 5 次/天 |
| 关键路径延迟级联 | 高 | 若 A1 延迟 > 4 小时，立即调整 B-D 计划；若 B-C 累计延迟 > 1 天，考虑合并或延迟规则 |
