# Plan 219C2D – 扩展与验收

**文档编号**: 219C2D  
**上级计划**: [219C2 – Business Validator 框架扩展](./219C2-validator-framework.md)  
**目标周期**: Week 4 Day 24  
**负责人**: 组织后端团队（架构/安全共评估）  

---

## 1. 目标

1. 完成 Job Catalog 规则实现与端到端测试，确保 REST/GraphQL 行为一致。
2. 汇总 README、Implementation Inventory、219C 主计划验收勾选并归档。
3. 出具日终验收纪要与归档材料，预留滚动计划。

---

## 2. 范围

| 模块/文档 | 工作内容 |
|---|---|
| `internal/organization/validator/job_catalog_*.go` | 实现 Job Catalog 规则（JC-TEMPORAL 等）。 |
| 端到端测试 | `tests/e2e/organization-validator/*.spec.ts`（或等效 Playwright 脚本）。 |
| 可观测性 | Prometheus 指标（验证链执行耗时、失败率等）注册与文档记录。 |
| README & Implementation Inventory | 更新规则矩阵、测试策略、实现清单。 |
| `docs/archive/development-plans/` | 创建 219C2 归档文件。 |
| 日志 | `logs/219C2/validation.log`, `logs/219C2/test-Day24.log`, `logs/219C2/daily-YYYYMMDD.md`。 |

---

## 3. 前置条件

- 219C2B/219C2C 验收通过且无阻断缺陷。
- Docker Compose 环境（命令/查询/Redis/PostgreSQL/Temporal）运行正常。
- README 中规则矩阵已更新到最新版本。

---

## 4. 详细任务

### 4.1 Job Catalog 规则实现
- 编写 Job Catalog 规则文件：生效区间冲突、引用存在性、层级依赖。
- 与数据团队确认所需仓储接口，可复用 stub。
- 单元测试覆盖正/反场景，纳入覆盖率统计。

### 4.2 端到端测试
- 按计划执行 9 个用例（3 命令 × 3 场景），验证 REST/GraphQL 错误码一致。
- 输出测试报告 `tests/e2e/organization-validator/report-Day24.json`。
- 若测试依赖初始数据，使用迁移或脚本初始化并记录。

### 4.3 文档与归档
- 更新 README：规则矩阵、执行顺序、测试策略、实现检查清单。
- 更新 Implementation Inventory 条目，标注实际实现与测试路径。
- 在 219C 主计划、Implementation Inventory 中打勾，更新验收进度。
- 归档本计划：复制本文件与执行记录到 `docs/archive/development-plans/219C2-YYYYMMDD.md`。

### 4.4 验证链可观测性
- 注册验证链相关 Prometheus 指标（如执行总数、失败总数、执行耗时），复用现有 metrics 注册框架。
- 在 README/Implementation Inventory 中记录指标名称、含义与监控面板位置。
- 确认指标已在 Prometheus 暴露，并在验收纪要中附上查询示例。

### 4.5 验收会议
- Day 24 下午召开 30 分钟验收会议，参与：后端负责人、架构、安全。
- 输出纪要：完成项、风险余量、滚动任务。
- 若有未完成项，制定滚动计划并抄送 219C 总计划。

---

## 5. 交付物

- Job Catalog 规则代码与单测。
- 端到端测试脚本、报告、运行日志。
- 更新后的 README、Implementation Inventory。
-,219C 主计划验收勾选记录。
- 归档文件：`docs/archive/development-plans/219C2-YYYYMMDD.md`。
- 验收纪要、Day 24 日志。

---

## 6. 验收标准

- [ ] `go test -cover ./internal/organization/validator` ≥ 85%，含 Job Catalog 规则；报告保存。
- [ ] 9 个端到端测试全部通过，报告归档。
- [ ] README 与 Implementation Inventory 更新并合并（含 Prometheus 指标说明）。
- [ ] 验证链 Prometheus 指标注册完成，可在 Prometheus 中查询。
- [ ] 219C 主计划验收勾选完成，归档文件生成。
- [ ] Day 24 验收纪要提交并记录下一步计划（如有）。

---

## 7. 时间安排（Day 24）

| 时间段 | 工作 | 输出 |
|---|---|---|
| 08:30-10:30 | Job Catalog 规则实现 + 单测 | 规则代码、测试报告 |
| 10:30-12:00 | 端到端测试编写与运行 | e2e 脚本与报告 |
| 13:00-14:30 | README/Inventory 更新、219C 勾选 | 文档更新、验收勾选 |
| 14:30-15:30 | 验收会议、纪要 | 纪要、风险余量 |
| 15:30-17:00 | 归档文件、日记、滚动计划 | 归档文件、daily log |

---

## 8. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| 端到端测试依赖数据环境 | 高 | 提前验证 docker-compose 环境；失败时及时使用缓冲并通知负责人。 |
| 文档同步遗漏导致唯一事实来源漂移 | 中 | 使用 checklist：README → Implementation Inventory → 219C 主计划 → 归档；逐项勾选。 |
| Job Catalog 规则需求变更 | 中 | 若发现新需求，立即记录并在验收会议确认是否转移到 219E。 |

---

## 9. 度量与追踪

- `logs/219C2/test-Day24.log`: 单测覆盖率、执行耗时。
- `tests/e2e/organization-validator/report-Day24.json`: e2e 结果。
- `logs/219C2/daily-YYYYMMDD.md`: Day 24 完成情况、风险、缓冲。
- 验收纪要保存于 `logs/219C2/acceptance-Day24.md`。
