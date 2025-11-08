# Plan 219C3A – REST 自测问题修复与回归

**文档编号**: 219C3A  
**上级计划**: [219C3 – 文档、测试与验收对齐](./219C3-docs-and-tests.md)  
**创建日期**: 2025-11-06  
**责任团队**: 组织命令域后端组 + 测试验证组  

---

## 1. 背景

219C3 要求通过 `scripts/219C3-rest-self-test.sh` 验证关键命令（Create/Fill/Close、Job Level 版本）并提供审计/日志凭证。然而在 2025-11-06 的全栈自测中，脚本产生多项失败：

- `logs/219C3/validation.log` & `logs/219C3/report.json` 显示多个用例返回 `500 INTERNAL_ERROR` 或错误码不符。
- Docker `cubecastle-rest` 日志记录 `unhandled position service error`（例如 requestId `4cc75fa5-21ab-4429-815b-0013f43255b2`）。
- 审计查询结果缺少预期的 `ruleId`/`severity`，无法满足验收证据要求。

为确保 219C3 可顺利收口，需要单独立项梳理并修复这些阻断问题。

---

## 2. 现状与问题列表

| 序号 | 场景 | 期望 | 实际 | 证据 |
|----|------|------|------|------|
| P1 | `POST /api/v1/positions/{code}/fill` 首次填充 | HTTP 200，返回 assignmentId 并记录审计 | HTTP 500，日志 `unhandled position service error` | `logs/219C3/validation.log` 中 requestId `4cc75fa5-21ab-4429-815b-0013f43255b2`；`docker logs cubecastle-rest` |
| P2 | `fill` headcount 超限 | HTTP 400，错误码 `POS_HEADCOUNT_EXCEEDED`，ruleId `POS-HEADCOUNT` | HTTP 500 + `INTERNAL_ERROR` | requestId `c6303e37-75ce-433f-af9f-426387bbdbbb` |
| P3 | `POST /api/v1/positions/{code}/assignments/{id}/close`（成功/重复关闭） | 200 与 400，错误码 `ASSIGN_INVALID_STATE`、ruleId `ASSIGN-STATE` | 400 + `INVALID_ASSIGNMENT_ID`（未能获取 assignmentId） | requestId `eb020cee-0dca-417e-9dfb-f3c1a922771f` |
| P4 | `POST /api/v1/job-levels/{code}/versions`（成功与时间冲突） | 201 / 400（`JOB_CATALOG_TEMPORAL_CONFLICT`，ruleId `JC-TEMPORAL`） | 均返回 500 `INTERNAL_ERROR` | requestId `79521ac4-bf70-49e7-863d-06d09c5120e4`、`1071bf3d-46db-4a59-a441-c9d65d172e4e` |

衍生影响：上述失败导致审计验证无法完成，也阻塞了 219C3 的验收勾选。

---

## 3. 目标

1. 修复职位填充与任职关闭流程，使脚本能够获得 assignmentId 并返回正确错误码/审计数据。
2. 修复 Job Level 版本写入逻辑及错误映射，重新对齐 `JC-TEMPORAL` 规则与错误码。
3. 重新运行 `scripts/219C3-rest-self-test.sh`，生成通过的 `logs/219C3/validation.log` 与 `report.json`，并在 219C3 文档中更新验收记录。

---

## 4. 范围

- **代码修复**：命令服务职位/任职流程、Job Catalog 版本仓储与错误映射。
- **验证与日志**：重新执行 REST 自测脚本 + 必要的单元/集成测试，确认审计日志写入。
- **文档更新**：同步 `internal/organization/README.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md` 和 219C3 计划验收章节。

不包含：新增 GraphQL Mutation、前端改动或非 219C3 范围外的需求。

---

## 5. 详细任务

### 5.1 职位填充/任职关闭修复（Owner: Position Service）
- [x] 分析 `unhandled position service error` 根因（调试日志见 `run-dev-command-new.log`），确认因 handler 未捕捉 validator 结果导致 500。
- [x] 修复 `PositionService.Fill` / `CloseAssignmentRecord` handler 的错误处理，确保：
  - 首次填充返回 200 并包含 assignmentId（`logs/219C3/validation.log` 内 `position.fill/success`）。
  - headcount 超限场景返回 `POS_HEADCOUNT_EXCEEDED` + ruleId `POS-HEADCOUNT`。
  - handler 在业务规则失败时写入审计 (`audit_logs` 查询：`requestId=7961a07c-9d93-4c5d-a219-3fd98b668a9b`，ruleId=`ASSIGN-STATE`，severity=`CRITICAL`)。
- [x] 调整自测脚本 assignmentId 获取逻辑并附加 endDate，统一写回日志与报告。
- [x] 增补 Position handler 审计记录逻辑与单元测试（`go test ./internal/organization/handler/...`）。

### 5.2 Job Level 版本修复（Owner: Job Catalog Service）
- [x] 校正 `JobCatalogRepository.InsertJobLevelVersion` 列映射，复用父版本 `level_rank`/`role_code`/`salary_band` 并默认新版本 `is_current=false`。
- [x] 更新错误翻译，使 temporal 冲突返回 `JOB_CATALOG_TEMPORAL_CONFLICT` + ruleId `JC-TEMPORAL`。
- [x] 新增 SQLMock 单元测试覆盖成功与 parent mismatch 场景（`internal/organization/repository/job_catalog_repository_test.go`）。

### 5.3 回归验证与归档
- [x] 使用修复后的命令执行 `./scripts/219C3-rest-self-test.sh`（命令服务本地日志 `run-dev-command-new.log`；结果写入 `logs/219C3/validation.log`）。
- [x] 生成并保留最新 `logs/219C3/validation.log`、`logs/219C3/report.json`，报告涵盖 Headcount/Assignment/JobLevel 版本场景。
- [x] 更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 自测脚本说明（新增 REST 自测指引与审计查询提示）。
- [x] 在计划验收章节记录执行时间与日志路径。

---

## 6. 交付物

- 修复后的代码与测试（Position / Job Catalog）。
- 通过的自测日志与报告 (`logs/219C3/validation.log`、`report.json`)。
- 更新的 README / Reference / 219C3 文档条目。

---

## 7. 时间线与依赖

| 任务 | 截止 | 依赖 |
|-----|------|------|
| 职位填充 bug 修复 + 测试 | D+1 | 现有 validator 链 |
| Job Level 版本修复 + 测试 | D+2 | Goose 迁移 & validator 实现 |
| 自测回归与文档更新 | D+3 | 上述两个修复完成 |

---

## 8. 风险与缓解

| 风险 | 描述 | 缓解 |
|------|------|------|
| 填充流程依赖外部 worker | 若当前实现仍需其他服务配合，可能导致测试无法独立通过 | 在命令服务内降级处理或提供 stub，必要时补充集成测试容器 |
| SQL 映射修改影响历史数据 | Job Level 表字段调整需确认迁移兼容性 | 在临时库跑 `goose up/down` + `make test-db`，确认不会破坏现有数据 |
| 自测脚本参数变化 | 脚本更新可能影响文档引用 | 同步更新 README/Reference，并在 219C3 验收记录中注明脚本版本 |

---

## 9. 验收标准

- [x] `scripts/219C3-rest-self-test.sh` 全部步骤通过，`logs/219C3/report.json` 标记三大场景均为 `passed`。
- [x] `logs/219C3/validation.log` 记录成功/失败场景以及 ruleId 信息；`audit_logs` 查询确认 `ASSIGN-STATE` 等规则写入审计。
- [x] 单元/集成测试覆盖新增路径，`go test ./...` 全绿。
- [x] 219C3 主计划“REST 命令补测”项更新（参见本文档验收节）。

---

## 10. 参考资料

- `logs/219C3/validation.log`（2025-11-06 17:16 生成版本）
- `docker logs cubecastle-rest`（requestId 详见上表）
- `internal/organization/service/job_catalog_service.go`
- `internal/organization/repository/job_catalog_repository.go`
- `scripts/219C3-rest-self-test.sh`
