# Plan 219C2A – Validator 框架基座

**文档编号**: 219C2A  
**上级计划**: [219C2 – Business Validator 框架扩展](./219C2-validator-framework.md)  
**目标周期**: Week 4 Day 21  
**负责人**: 组织后端团队（架构/安全组评审）  

---

## 1. 目标

1. 在 PoC 通过后，巩固可复用的 `BusinessRuleValidator` / `ValidationResult` 结构，明确最小改造清单。
2. 提供链式执行骨架、短路控制以及统一错误转换入口，确保 REST/GraphQL/批处理可共享。
3. 建立规则登记与错误码冻结机制，落实唯一事实来源。

---

## 2. 范围

| 模块/文档 | 工作内容 |
|---|---|
| `internal/organization/validator/core.go` | 增量实现链式执行入口、短路策略、Severity 映射，保持现有结果结构。 |
| `internal/organization/handler/organization_helpers.go` | 补充统一错误翻译、审计上下文封装、日志字段。 |
| `internal/organization/README.md#validators` | 新增框架说明、规则登记模板、命名约束。 |
| `docs/reference/02-IMPLEMENTATION-INVENTORY.md` | 添加 “Business Validator Chains” 草稿条目并引用 README。 |
| `logs/219C2/rule-freeze.md` | 记录规则矩阵与错误码冻结纪要。 |

---

## 3. 前置条件

- §3.0 PoC 验收四项均通过：链式执行、工厂注入、性能基准、JSON 序列化。
- 219C1 审计基础设施已验收，`LogError`/`LogEventInTransaction` 可用。
- OpenAPI 契约缺失的错误码补丁已合并，或已有待审 PR。

---

## 4. 详细任务

### 4.1 框架复用评估
- 核对 `internal/organization/validator/business.go` 与现有接口，列出≤3 项必须的签名调整；若无需调整，记录“无新增签名”。
- 若需要扩展（例如新增上下文类型），先提交评审纪要并确保与架构组达成一致。

### 4.2 链式执行骨架
- 在 `core.go` 实现：
  - 验证链装配与优先级排序；
  - 错误聚合、短路控制；
  - 复用现有 `ValidationResult`/`ValidationError` 结构。
- 新增 smoke 测试 `TestValidatorCoreSmoke` 验证最小链路。

### 4.3 命令入口辅助
- 在 handler helper 中实现统一错误翻译（错误码、HTTP 状态、字段映射）。
- 集成审计上下文写入：`business_context.ruleId`、`severity`、`payload`。

### 4.4 文档与错误码冻结
- 在 README 新增章节：
  - 框架说明、注册流程、规则命名规范；
  - 规则登记模板（Rule ID / 优先级 / 错误码 / 所属命令）。
- 在 Implementation Inventory 新增条目，指向 README 章节。
- 与安全/架构组会签规则与错误码冻结，记录至 `logs/219C2/rule-freeze.md`。

### 4.5 日志记录
- 在 `logs/219C2/daily-YYYYMMDD.md` 记录当日完成度、风险、阻塞项。
- 若发现结构性风险（例如需新增类型），立即在日记中抄送架构组。

---

## 5. 交付物

- 更新后的 `core.go` 与 handler helper 代码。
- `TestValidatorCoreSmoke` 测试及报告。
- README 与 Implementation Inventory 草稿段落。
- `logs/219C2/rule-freeze.md` 纪要。
- 日志：`logs/219C2/daily-YYYYMMDD.md`（Week 4 Day 21）。

---

## 6. 验收标准

- [x] `go test ./internal/organization/validator -run TestValidatorCoreSmoke` 通过（详见 `logs/219C2/test-Day21.log` / `logs/219C2/test-Day22.log`）。
- [x] README 新章节过架构/安全组评审并合并（`internal/organization/README.md#validators`）。
- [x] 规则/错误码冻结纪要签字完成（`logs/219C2/rule-freeze.md`）。
- [x] 最小改造清单归档 —— 219C2A 复用既有接口，无新增签名。
- [x] 日志条目填写并提交到 219C 总计划（参见 `logs/219C2/daily-20251105.md`、`logs/219C2/daily-20251106.md`）。

---

## 7. 时间安排（Day 21）

| 时间段 | 工作 | 输出 |
|---|---|---|
| 09:00-09:30 | PoC 验收回顾、确定改造范围 | PoC 结论、改造清单草稿 |
| 09:30-12:00 | 链式执行骨架 + helper 增量开发 | `core.go`、helper 代码、smoke 测试 |
| 13:00-15:00 | README/Inventory 草稿、错误码对齐 | README 增量、Inventory 条目 |
| 15:00-16:30 | 冻结会议 & 纪要 | `logs/219C2/rule-freeze.md` |
| 16:30-17:30 | 自测、验收打勾、日记更新 | `go test` 报告、daily log |

---

## 8. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| 需要新增 `ValidationState` 等新结构 | 中 | 先经架构组评审，签字后执行；否则禁止合并。 |
| PoC 与实际落地不一致 | 中 | 若出现差异，记录于 `logs/219C2/poc-review.md` 并立即调整计划。 |
| 错误码与 OpenAPI 不对齐 | 高 | 开发前确认补丁状态；若缺失立即提交契约 PR。 |

---

## 9. 度量与追踪

- `go test` 结果附于 `logs/219C2/test-Day21.log`。
- `daily-YYYYMMDD.md` 记录完成项/风险/缓冲占用。
- 冻结纪要在 219C 总计划会议中复核。
