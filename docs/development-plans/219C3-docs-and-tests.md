# Plan 219C3 – 文档、测试与验收对齐

**文档编号**: 219C3  
**上级计划**: [219C – Audit & Validator 规则收敛](./219C-audit-validator.md)  
**目标周期**: Week 4 Day 24-25  
**负责人**: 后端团队 + 技术写作组  

---

## 1. 目标

1. 汇总 219C1/219C2 的实现结果，更新 README、参考文档、计划索引，确保唯一事实来源一致。
2. 完成必要的单元/集成测试，覆盖审计与 validator 关键路径，计划完成后输出验收记录。
3. 在 `docs/development-plans/` 与 `docs/reference/` 之间更新引用关系，归档执行证据，满足审计可追溯性。

---

## 2. 范围

| 模块 | 内容 |
|---|---|
| 文档 | `internal/organization/README.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`docs/development-plans/219C-audit-validator.md` 同步更新并互相引用。 |
| 测试 | `go test ./internal/organization/audit ./internal/organization/validator`, 以及必要的 service 层回归测试。 |
| 计划/归档 | 将阶段性结果追加到 `docs/development-plans/219C-audit-validator.md`，并准备归档（完成后移动至 `docs/archive/development-plans/`）。 |
| 验收表 | 输出勾选清单：审计操作全覆盖、validator 规则矩阵落地、README 更新完成。 |

---

## 3. 详细任务

### 3.1 文档同步
- [ ] 更新 `internal/organization/README.md`：新增 `#audit`、`#validators` 子章节内容，列出字段、操作清单、规则矩阵。
- [ ] `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 增加“审计/Validator 检查”条目，提示相关命令与脚本。
- [ ] `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 核对并更新审计、校验模块条目，确保与 README 描述、代码实现一致，并记录版本标签。
- [ ] `docs/development-plans/219C-audit-validator.md` 引用 219C1/219C2/219C3 的最新结果，在计划末尾加入验收记录与执行凭证链接。

### 3.2 测试与验证
- [ ] 运行 `go test ./internal/organization/audit ./internal/organization/validator`，保存执行结果（供验收引用）。
- [ ] 按 219C2 规则矩阵梳理 service 层集成测试缺口：凡触发 HIGH/CRITICAL 级别规则的命令都需至少 1 条集成回归用例，补测完成后在计划中勾选命令清单。
- [ ] 验证审计记录字段（`tenant`,`entityType`,`requestId` 等）实际落盘：使用 `make run-dev` 环境或等效集成测试，执行组织创建流程后通过 `psql -c "SELECT tenant_id, resource_type, request_id, business_context->>'correlationId' FROM audit_logs WHERE request_id='<当前测试 requestId>'"` 或对应测试断言确认字段与 README 描述一致，并将查询结果粘贴至计划附录或 `logs/219C3/` 下的文本文件。
- [ ] 将上述测试与查询的 CLI 输出统一保存到 `logs/219C3/validation.log`，在验收记录中引用。

### 3.3 验收记录
- [ ] 在计划文档中补充勾选项，标记完成状态与测试证据（含日志路径、执行日期、责任人）。
- [ ] 完成所有勾选后，按照仓库规范立即将计划复制至 `docs/archive/development-plans/`，记录归档时间与版本标签。

---

## 4. 交付物

- 更新后的 README/参考文档。
- 完整的测试结果（CLI 输出或截图引用）。
- 219C 主计划的验收标记与回溯信息。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| 文档与实现不一致 | 高 | 以代码为准同步更新 README，必要时拉 Reviewer 联合检查。 |
| 测试覆盖不足 | 中 | 在 219C2 完成后统计未覆盖路径，纳入本阶段补测。 |
| 计划归档遗漏 | 低 | 在 PR 模板中添加检查项，确保归档操作可追踪。 |
