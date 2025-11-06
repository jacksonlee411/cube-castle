# Plan 219C2W – Validator Error Reconciliation

**文档编号**: 219C2W  
**上级计划**: [219C2 – Business Validator 框架扩展](./219C2-validator-framework.md)  
**目标周期**: Week 4 Day 25  
**负责人**: 组织后端团队（命令/审计协作）  

---

## 1. 目标

1. 将 Job Catalog 版本创建中的数据库冲突（duplicate timestamp）转换为 JC-* 验证错误，统一 `ruleId`/`severity` 展示与审计。
2. 校正 Position Fill/Assignment 失败时的错误映射，确保 POS-HEADCOUNT、ASSIGN-STATE 等规则在 REST/GraphQL 输出一致。
3. 完成自测脚本与 GraphQL 查询输出，生成可归档的 REST/GraphQL 双通道证据与 `tests/e2e/organization-validator/report-Day24.json`。

---

## 2. 范围

| 模块/文档 | 工作内容 |
|---|---|
| `internal/organization/service/job_catalog_service.go` | 捕获 PostgreSQL `23505` 等错误并调用 `ValidationFailedError`，补足 JC-* 错误翻译。 |
| `internal/organization/service/position_service.go`、`handler/position_handler.go` | 调整 Fill/Assignment 失败路径，确保 `validator.ValidationFailedError` 与 handler 的 `writeValidationFailure` 输出一致。 |
| `scripts/219C2D-validator-self-test.sh` | 更新测试逻辑，在 Job Catalog 与 Position 场景中校验 `ruleId`、错误码、审计日志。 |
| `docs/api/openapi.yaml` / `docs/reference/02-IMPLEMENTATION-INVENTORY.md` | 如错误码新增或描述变动，对应更新唯一事实来源。 |
| `tests/e2e/organization-validator/report-Day24.json` | 生成包含 REST/GraphQL 场景的正式报告。 |

---

## 3. 前置条件

- 219C2D Job Catalog 规则与单测已合并（参考 `logs/219C2/test-Day24.log` 覆盖率 85.3%）。
- 命令服务与 GraphQL 服务以最新代码运行，可通过 `make run-dev` 或 `docker compose -f docker-compose.dev.yml up` 获取日志。
- 审计日志表可写入，验证 `LogError` / `LogEvent` 调用无异常。

---

## 4. 详细任务

### 4.1 Job Catalog 错误翻译
- 在 `JobCatalogService` 内捕获 `InsertFamilyGroupVersion` 等 repository 返回的 `pq.Error{Code=23505}` 或 `invalid effective date` 错误，调用 `validator.NewValidationFailedError`。
- 增补 `validator` fallback 逻辑（若无 validator 实例，构造最小 `ValidationResult`），保证 `ruleId=JC-TEMPORAL` 与 `JOB_CATALOG_TEMPORAL_CONFLICT` 对齐。
- 添加单元测试 `job_catalog_service_test.go`，模拟重复插入并断言 `ValidationFailedError`。

### 4.2 Position / Assignment 错误映射
- 检查 `PositionService.createAssignment`，对 `ErrInvalidHeadcount`、`ErrInvalidAssignmentState` 等返回路径转为 `ValidationFailedError`，并确保 handler 能透传 `ruleId`。
- 更新/新增单测覆盖 `FillPosition` 超额、关闭任职等场景，验证回传错误码与 `ValidationResult` 内容。

### 4.3 自测脚本与报告
- 调整 `scripts/219C2D-validator-self-test.sh`：
  - 校验 REST 响应中的 `error.code`、`details.ruleId`、`severity`。
  - GraphQL 查询若返回 schema 错误（变量类型不符），转换为内部查询或忽略。
  - 生成 `tests/e2e/organization-validator/report-Day24.json`（使用 `jq` 汇总）。
- 运行脚本并将结果写入 `logs/219C2/validation.log`、`report-Day24.json`。

### 4.4 文档与计划同步
- 更新 README 与 Implementation Inventory 对应条目，记录错误翻译逻辑。
- 若 OpenAPI 错误示例变更，提交契约更新。
- 在 219C2 主计划与 219C2D 文档中标记补齐时间。

---

## 5. 交付物

- 修正后的命令服务（job catalog 与 position）代码、单测。
- 更新后的自测脚本与报告 `tests/e2e/organization-validator/report-Day24.json`。
- `logs/219C2/validation.log` 内的 REST/GraphQL 审计证据。
- 文档更新（README、Implementation Inventory、OpenAPI 如适用）。

---

## 6. 验收标准

- [ ] Job Catalog 版本重复插入返回 `JOB_CATALOG_TEMPORAL_CONFLICT` / `JC-TEMPORAL`，请求 HTTP 400。
- [ ] Position Fill 超编返回 `POS_HEADCOUNT_EXCEEDED` / `POS-HEADCOUNT`，HTTP 400；任职关闭状态异常返回 `ASSIGN_INVALID_STATE`。
- [ ] 自测脚本成功执行并输出报告；日志中含 REST/GraphQL 双通道证据与审计 `ruleId`。
- [ ] README、Implementation Inventory 记录更新，OpenAPI（如修改）合并。

---

## 7. 时间安排（Day 25）

| 时间段 | 工作 | 输出 |
|---|---|---|
| 09:00-10:30 | Job Catalog 错误翻译实现与单测 | service/validator 代码、测试报告 |
| 10:30-12:00 | Position/Assignment 错误映射修复 | 相关服务代码、测试覆盖 |
| 13:00-14:00 | 自测脚本运行与报告生成 | `report-Day24.json`、日志证据 |
| 14:00-15:00 | 文档更新、计划对齐 | README、Implementation Inventory、计划勾选 |
| 15:00-16:30 | 验收检查与风险评估 | 日终纪要、风险项、滚动任务 |

---

## 8. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| PostgreSQL 错误码条件遗漏，仍返回 500 | 中 | 逐一捕获 `pq.Error` 情况并写单测，必要时在 repository 层增加错误类型。 |
| Position 服务修正影响既有逻辑 | 高 | 保留现有错误映射回退路径，单测验证旧场景。 |
| GraphQL Schema 限制导致自测困难 | 中 | 改用 REST + `auditHistory` 查询验证 `ruleId`；对 GraphQL 仅保留读取校验。 |
| 时间不足导致文档更新延迟 | 低 | 使用 checklist（README→Inventory→OpenAPI）并在计划日志记录状态。 |

---

## 9. 度量与追踪

- `logs/219C2/validation.log`：REST/GraphQL 自测输出与审计信息。
- `tests/e2e/organization-validator/report-Day24.json`：正式报告。
- 文档更新记录：README、Implementation Inventory 的 Git diff。
- CI 覆盖率：`go test -cover ./internal/organization/validator` ≥85%，`./service` 层单测通过。

---

*附注：本计划为 219C2D “端到端测试与审计校验” 子任务的专项补充，完成后需同步 219C 主计划勾选状态。*
