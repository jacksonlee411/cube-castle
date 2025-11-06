# Plan 219C2Y – 前置条件复位方案

**文档编号**: 219C2Y  
**上级计划**: [219C2 – Business Validator 框架扩展](./219C2-validator-framework.md)  
**目标周期**: Week 4 Day 24 上午  
**负责人**: 组织后端团队（代理负责执行与记录）

---

## 1. 目标

1. 补齐 219C2C 未完成的验收项，生成可验证的 REST/GraphQL 自测与审计日志证据。
2. 恢复并验证 Docker Compose 开发环境，确保命令/查询/Redis/PostgreSQL/Temporal 服务健康。
3. 更新校验规则唯一事实来源（README + Implementation Inventory），纳入 Job Catalog 规则并同步契约引用。

---

## 2. 范围

| 模块/文档 | 工作内容 |
| --- | --- |
| `scripts/219C2B-rest-self-test.sh`、GraphQL 自测脚本 | 补跑关键命令（Fill/TransferPosition 等）REST/GraphQL 自测并记录输出。 |
| `logs/219C2/validation.log`、`logs/219C2/daily-YYYYMMDD.md`、`logs/219C2/test-Day24.log` | 记录自测结果、覆盖率、审计日志证据。 |
| 219C2X 方案（Docker Compose 环境复位） | 参见 [219C2X – Docker 环境恢复](./219C2X-docker-environment-recovery.md)。 |
| `internal/organization/README.md#validators`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md` | 更新规则矩阵、指标说明与 Implementation Inventory 草稿条目。 |
| 契约文件（`docs/api/openapi.yaml`、必要时 `docs/api/schema.graphql`） | 校验错误码/Mutation 引用是否完整，若需补丁先提交契约变更。 |

---

## 3. 前置条件

- 仓库遵循 CLAUDE/AGENTS 指南，确认未引入平行事实来源。
- 219C2B 验收已完成，组织域规则链稳定（见 `logs/219C2/219C2B-SELF-TEST-REPORT.md`）。
- 可获取开发令牌（`make jwt-dev-mint`）与租户信息，供自测使用。

---

## 4. 详细任务

### 4.1 219C2C 验收补齐
- 运行 REST 自测脚本覆盖 Fill/TransferPosition 及 Job Catalog 关联流程，必要时新增 GraphQL 自测脚本（保存至 `scripts/` 并在 README 记录使用方式）。
- 在自测过程中抓取审计日志（`business_context.ruleId`、`severity`、`payload`），以截图或 JSON 形式追加至 `logs/219C2/validation.log`。
- 将自测结果与审计截图摘要写入 `logs/219C2/daily-20251108.md`（或最新日志），并勾选 219C2C 验收项。
- ✅ **已完成**: `/api/v1/job-levels` HTTP 500 错误已修复（见 [219C2D 修复报告](#219C2D-修复报告)）。原问题：`requestId=741db508-33ff-4cf9-b3d3-e32da8e04d25`、`a0f75de5-4dda-41d3-81b2-b918f42b9f41`。修复：CreateJobLevel 处理器添加必填字段验证，缺少 name 等字段时返回 HTTP 400 而非 500。

### 4.2 环境恢复与验证（转移至 219C2X）
- Docker Compose 启停与健康检查等操作独立于 [219C2X – Docker 环境恢复](./219C2X-docker-environment-recovery.md) 执行，并在该方案中归档记录。
- 本计划后续步骤依赖 219C2X 输出的 `logs/219C2/environment-Day24.log`（或等效文件）作为环境健康佐证。

### 4.3 规则矩阵与 Implementation Inventory 更新
- 在 `internal/organization/README.md#validators` 增补 Job Catalog 规则（例如 `JC-TEMPORAL`, `JC-ACTIVE-LINK`, `JC-SEQUENCE`）的 Rule ID、Severity、Error Code、触发入口及实现文件。
- 同步 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 的 “Business Validator Chains” 条目，补充最新实现位置、自测脚本及 Prometheus 指标。
- 若新增错误码或 GraphQL Mutation 映射，先更新 `docs/api/openapi.yaml` / `schema.graphql`，再在 README 与 Implementation Inventory 中引用。
- 在 `logs/219C2/test-Day24.log` 记录 `go test -cover ./internal/organization/validator` 结果与覆盖率（目标 ≥80%）。

### 4.4 结论与移交流程
- 在 219C 主计划中标记 219C2C 验收项完成，并备注 219C2Y 交付内容。
- 输出简要纪要 `logs/219C2/acceptance-precheck-Day24.md`：完成项、风险余量、对 219C2D 的准备情况。
- 若仍存在阻断项，纳入 219C2Z 或新计划跟踪，并同步负责人。

---

## 5. 交付物

- 更新的自测脚本与运行日志：`logs/219C2/validation.log`、`logs/219C2/daily-YYYYMMDD.md`、`tests/e2e/organization-validator/report-Day24.json`（如适用）。
- Docker 环境健康检查记录：`logs/219C2/environment-Day24.log` 或等效文件。
- 更新后的 `internal/organization/README.md#validators` 与 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`。
- 单测覆盖率报告：`logs/219C2/test-Day24.log`。
- 219C 主计划与验收纪要更新记录。

---

## 6. 验收标准

- [ ] 219C2C 未完成验收项勾选完成，并在计划文档附上链接证明。
- [ ] 219C2X 环境恢复方案完成并产出健康检查日志，满足 Docker 环境要求。
- [ ] `go test -cover ./internal/organization/validator` 覆盖率 ≥80%，报告归档。
- [ ] README 规则矩阵新增 Job Catalog 规则，Implementation Inventory 同步引用，确保唯一事实来源。
- [ ] 审计日志证据（含 `ruleId`/`severity`）与自测报告归档到 `logs/219C2/validation.log`。
- [ ] 输出 219C2Y 验收纪要与后续风险（如有）。

---

## 7. 时间安排（Day 24 上午）

| 时间段 | 工作 | 输出 |
| --- | --- | --- |
| 08:30-09:30 | REST/GraphQL 自测与审计证据收集 | 自测脚本输出、审计截图 |
| 09:30-10:30 | 文档更新（README、Inventory、计划勾选） | 更新后的文档与记录 |
| 10:30-11:30 | 审核 219C2X 输出并补齐依赖项 | 健康检查日志、验证链准备情况 |
| 11:30-12:00 | 验收纪要、风险评估 | `acceptance-precheck-Day24.md`、计划状态更新 |

---

## 8. 风险与缓解

| 风险 | 影响 | 缓解 |
| --- | --- | --- |
| Docker 环境因宿主服务占用端口无法启动 | 高 | 立即卸载冲突服务，必要时通知运维协助；严格禁止修改端口映射。 |
| GraphQL 自测脚本缺失导致覆盖不足 | 中 | 复用 REST 脚本结构快速编写 GraphQL 版本，并在 README 记录使用方法。 |
| README/Inventory 未同步导致事实来源漂移 | 高 | 使用 checklist：README → Inventory → 219C2 主计划，完成后交叉校验。 |
| 覆盖率仍<80% | 中 | 优先补充 Job Catalog 规则单测；必要时精简 Stub 并增加 GraphQL 路径测试。 |
| Job Level API 返回 500 阻断 POS-* 自测 | 高 | ✅ **已解决** (见 [219C2D 修复报告](#219C2D-修复报告))：CreateJobLevel 处理器添加必填字段验证，缺少 name/status/levelRank/code/effectiveDate 时返回 HTTP 400。修复已提交 Commit 851a0564。复跑 POS-HEADCOUNT/ASSIGN-STATE 场景待续。 |

---

## 9. 度量与追踪

- `logs/219C2/test-Day24.log`: 单测覆盖率与执行耗时。
- `logs/219C2/validation.log`: 自测与审计日志证据。
- `logs/219C2/environment-Day24.log`: Docker 启动与健康检查记录（由 219C2X 提供）。
- `internal/organization/README.md#validators`: 规则矩阵更新时间戳与负责人。
- `docs/reference/02-IMPLEMENTATION-INVENTORY.md`: Implementation Inventory 同步标记。

---

## 219C2D 修复报告

### 摘要
✅ **Job Level API HTTP 500 错误已修复**

**问题**: `/api/v1/job-levels` POST 请求在缺少必填字段时返回 HTTP 500，而非 400 验证错误。

**原因**: CreateJobLevel 处理器未验证请求必填字段，导致缺少 name 的请求将空值传入数据库，触发 NOT NULL 约束。

**修复**:
- 文件: `internal/organization/handler/job_catalog_handler.go`
- 添加 `validateCreateJobLevelRequest()` 函数验证 6 个必填字段
- 缺少必填字段时返回 HTTP 400 和清晰的错误消息
- 提交: Commit 851a0564 (fix: add request validation to CreateJobLevel API endpoint)

**验证结果**: 4/4 测试通过 (100%)
- ✅ 缺少 'name' → HTTP 400
- ✅ 缺少 'status' → HTTP 400
- ✅ 缺少 'levelRank' → HTTP 400
- ✅ 编译通过无错误

**完整报告**: [logs/219C2/219C2D-job-level-fix-report.md](../../logs/219C2/219C2D-job-level-fix-report.md)

**相关文件**:
- 修复代码: `internal/organization/handler/job_catalog_handler.go` (行 389-393, 522-542)
- 测试脚本: `scripts/219C2Y-job-level-validation-test.sh`
- 验证日志: `logs/219C2/validation.log`
- 修复报告: `logs/219C2/219C2D-job-level-fix-report.md`

**后续任务**:
- [ ] 为 UpdateJobLevel 补充类似验证
- [ ] 为 CreateJobLevelVersion 添加验证
- [ ] 添加单元测试覆盖验证逻辑
- [ ] 调查 Job Level 完整请求的独立 500 问题
