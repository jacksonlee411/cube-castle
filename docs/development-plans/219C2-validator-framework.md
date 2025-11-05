# Plan 219C2 – Business Validator 框架扩展

**文档编号**: 219C2  
**上级计划**: [219C – Audit & Validator 规则收敛](./219C-audit-validator.md)  
**目标周期**: Week 4 Day 22-24  
**负责人**: 组织后端团队（安全架构组评审）  

---

## 1. 目标

1. 定义可组合的 `BusinessValidator` 接口（统一 `ValidationState` 输入、区分错误/警告），支持链式执行与规则分组。
2. 将现有组织校验逻辑拆分为规则清单，并扩展至职位、Assignment、Job Catalog，覆盖跨域依赖（组织状态、headcount、Job Catalog 连续性）。
3. 在命令服务入口注入校验链，确保命令在事务开始前完成业务验证并返回结构化错误码。
4. 完成单元测试矩阵，确保关键规则（层级循环、headcount 越界、Job Catalog 生效区间冲突、Assignment 状态流转）均有覆盖。

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

### 3.1 验证框架设计
- [ ] 定义 `ValidationRule` 接口（`Name() string` / `Validate(ctx context.Context, state *ValidationState) []ValidationIssue`），在 `internal/organization/validator/state.go` 中定义 `ValidationState` 结构，字段包括 `TenantID`、`ActorID`、`ActorName`、`RequestID`、`Operation`、`Payload`（命令请求体）、预加载的实体快照以及仓储访问器接口。
- [ ] 统一 `ValidationError` / `ValidationWarning` 结构，字段包含 `code`、`message`、`severity`、`field`、`context`；提供 `ValidationIssue` 包装器以区分错误/警告。
- [ ] 提供链式组合器与规则注册器（按命令类型聚合），例如 `validators.ForCreateOrganization(...)`、`validators.ForFillPosition(...)`，并在 `internal/organization/api.go` 中集中构建供 handler/service 注入。

### 3.2 规则实现
- 组织：层级循环、父级状态、时态有效性、code 唯一性、层级深度限制（17 层）。
- 职位：Job Catalog 关联有效、headcount ≥ 1、effectiveDate 与组织状态一致。
- Assignment：一职一人、FTE 负载、状态流转合法、ACTING 自动回退约束。
- Job Catalog：层级依赖存在、版本生效区间无重叠、父级状态。
- 跨域：职位引用的组织/Job Catalog 均需 ACTIVE；任职创建需检查 headcount。
- 执行前需维护完整规则矩阵（示例起点如下，执行阶段需补齐全部 Rule ID，并校验 Severity 与错误码与 `docs/api/openapi.yaml`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 一致）：

| Rule ID | 描述 | 适用命令 | Severity | 错误码/默认消息 | 依赖数据源 |
| --- | --- | --- | --- | --- | --- |
| ORG-DEPTH | 组织层级 ≤ 17 | Create/Update Organization | HIGH | `DEPTH_EXCEEDED` | HierarchyRepository |
| ORG-CIRC | 防止循环引用 | Update Organization | CRITICAL | `CIRCULAR_REFERENCE` | HierarchyRepository |
| ORG-STATUS | 父级必须 ACTIVE | Create/Update Organization | HIGH | `PARENT_INACTIVE` | OrganizationRepository |
| POS-HEADCOUNT | Headcount ≥ 1 且未超额 | Create/Update Position | HIGH | `INVALID_HEADCOUNT` | PositionRepository |
| POS-ORG | 职位引用组织 ACTIVE | Position 命令 | HIGH | `REFERENCE_INACTIVE` | OrganizationRepository |
| ASSIGN-STATE | Assignment 状态流转合法 | Fill/Vacate/Transfer | HIGH | `INVALID_ASSIGNMENT_STATUS` | AssignmentRepository |
| ASSIGN-FTE | 任职 FTE 不超额 | Fill/Update Assignment | HIGH | `INVALID_HEADCOUNT` | AssignmentRepository |
| JC-TEMPORAL | Job Catalog 生效日期无重叠 | Job Catalog create/update | HIGH | `TEMPORAL_CONFLICT` | JobCatalogRepository |
| CROSS-ACTIVE | 跨域引用实体 ACTIVE | Position/Assignment | HIGH | `REFERENCE_INACTIVE` | Org/JobCatalog Repo |
| ... | （执行时补充完整） |  |  |  |  |

### 3.3 注入流程
- [ ] 在各命令 handler/service 开始处构建校验上下文，执行校验链，失败时阻断后续数据库操作。命令接入清单（执行时逐项勾选）：

| 模块 | 命令/操作 | 现有校验 | 调整 |
| --- | --- | --- | --- |
| 组织 | Create/Update Organization、CreateVersion、Suspend/Activate | `utils.Validate*` + 零散判断 | 保留基础字段校验，引入 BusinessValidator 链负责业务规则 |
| 职位 | Create/Replace/Version/Update Position | Service 内手写校验 | 迁移业务规则至 PositionValidator；Service 仅负责持久化与审计 |
| Assignment | Fill/Create/Update/Vacate/Transfer Position | Service 层逻辑 | 引入 AssignmentValidator 链，并在 GraphQL 入口复用 |
| Job Catalog | Create/Update JobFamily/Role/Level（含版本） | Repo/Service 零散校验 | 于 handler/service 调用 Validator，repo 聚焦数据一致性 |
| GraphQL Mutation | `mutation createOrganization`、`mutation updateOrganization`、`mutation fillPosition`、`mutation transferPosition` 等 | 倚赖服务返回错误 | GraphQL 层须显式执行相同校验链（复用命令工厂） |

- [ ] 与审计联动：校验失败时调用 `LogError`（复用 219C1 结果），`OperationReason` 在 `business_context` 中写入 `ruleId`、`severity`、`payload`（例如：`ruleId=POS-HEADCOUNT, severity=HIGH, details=...`）。
- [ ] 在 `internal/organization/api.go` 注册 validator 工厂，命令模块与查询模块共用同一规则定义。

### 3.4 测试矩阵
- [ ] 使用 `table-driven` 测试覆盖正/负案例，每条规则至少包含 1 个失败样例和 1 个通过样例。
- [ ] 为跨域规则准备 stub 仓储（mock `HierarchyRepository`、`OrganizationRepository`、`PositionRepository`、`JobCatalogRepository` 等），避免真实数据库依赖。
- [ ] 运行 `go test -cover ./internal/organization/validator` 并记录覆盖率，关键分支覆盖率 ≥ 80%，同时断言错误码、Severity、Field。
- [ ] 补充 service 层集成测试：示例（CreatePosition 触发 `POS-HEADCOUNT` → 返回 `INVALID_HEADCOUNT`，且未调用仓储写入）。

---

## 4. 交付物

- 新的 validator 框架代码与注入逻辑。
- README 中的规则表与严重级别说明。
- 单元测试与预置伪造数据。
- 错误码/错误消息表（遵循 OpenAPI 契约）。
- 完整的规则矩阵文档（含 Rule ID、Severity、错误码映射）及对照检查记录。
- 命令/操作接入勾选清单（含测试证据与注入位置说明）。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| 规则执行顺序错误导致性能问题 | 中 | 使用链式容器时按“轻量规则 → 重规则 → 跨域规则”排序，并允许短路。 |
| Handler 重复校验造成维护成本 | 中 | 通过统一构建函数（例如 `validators.ForCreateOrganization(...)`）减少重复代码。 |
| 错误码与现有 OpenAPI 不一致 | 高 | 先对照 `docs/api/openapi.yaml`，必要时同步更新契约并发起评审。 |
