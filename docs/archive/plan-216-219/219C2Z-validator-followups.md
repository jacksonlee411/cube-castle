# Plan 219C2Z – 验证链问题跟踪

**文档编号**: 219C2Z  
**上级计划**: [219C2 – Business Validator 框架扩展](./219C2-validator-framework.md)  
**目标周期**: Week 4 Day 22-23（滚动）  
**负责人**: Team 06 + 架构组联合排查

---

## 1. 目标

1. 修复 REST 自测报告中暴露的循环检测错误码问题，确保自引用返回 `ORG_CYCLE_DETECTED`。  
2. 诊断并解决组织更新 400、暂停 500 的异常，明确错误码与处理逻辑。  
3. 形成复盘总结，若需新增错误码或文档调整，回填到主计划与相关文档。

---

## 2. 问题列表

| 编号 | 问题描述 | 当前表现 | 预期行为 | 责任人 | 计划修复时间 | 状态 |
|------|----------|----------|----------|--------|----------------|------|
| Z-01 | 自引用循环返回 `INVALID_PARENT` | **已复测**：HTTP 400 / error.code=`ORG_CYCLE_DETECTED`，details.ruleId=`ORG-CIRC` | 返回 `ORG_CYCLE_DETECTED`，并在响应 details/审计中记录该 Rule | Team 06 | 2025-11-07 AM | ✅ 已完成 |
| Z-02 | 普通更新请求返回 400 | **已复测**：HTTP 200，支持中文括号 | 成功返回 200，或提供明确的验证错误码 | Team 06 + 仓储组 | 2025-11-07 PM | ✅ 已完成 |
| Z-03 | `SuspendOrganization` 返回 500 | **已复测**：HTTP 200，时间轴返回 `INACTIVE` 版本 | 返回业务错误码或明确冲突原因；不应产生 500 | 架构组 | 2025-11-08 | ✅ 已完成 |

---

## 3. 详细任务

### 3.1 Z-01 循环检测错误码
- 自引用在基础校验阶段直接返回 `ORG_CYCLE_DETECTED`，`details.ruleId=ORG-CIRC`。  
- 更新 `organization_rules_test.go`：新增 `TestOrganizationCreateSelfReferentialParent`（验证自引用）。  
- REST 回归验证见 `logs/219C2/z-followups/2025-11-05-validator-retest.md`。  
- OpenAPI 已包含 `ORG_CYCLE_DETECTED`，无需额外调整。

### 3.2 Z-02 组织更新 400
- 扩展名称正则，允许中文/英文括号，错误提示同步更新。  
- 单元测试：`TestOrganizationCreateNameAllowsLocalizedParentheses`、`TestOrganizationUpdateAllowsLocalizedParentheses`。  
- REST 回归显示 `PUT /organization-units/{code}` 返回 200，名称成功写入。  
- 诊断结果与修复步骤记录于 `logs/219C2/z-followups/2025-11-05-validator-retest.md`。

### 3.3 Z-03 暂停 500
- `TemporalTimelineManager` 插入新版本时补齐 `created_at`/`updated_at` 字段，避免列对齐导致的 SQL 错误。  
- `POST /organization-units/{code}/suspend` 回归返回 200，时间轴附带 `status=INACTIVE`。  
- 相关请求与响应记录见 `logs/219C2/z-followups/2025-11-05-validator-retest.md`。

---

## 4. 交付物

- 修复后的代码、单元测试、集成/自测结果。  
- 更新后的自测报告或附录，说明问题诊断与结论。  
- 如有错误码或文档改动，提交对应的 README / Implementation Inventory 更新。  
- 在 219C 主计划中记录完成状态。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解 |
|------|------|------|
| 问题定位耗时 | 中 | 优先使用自测报告与日志定位，必要时报架构组协助。 |
| OpenAPI 变更审批 | 中 | 若需新增错误码，提前与安全/架构沟通，使用 219C2D 时间窗口同步提交。 |
| 复现难度高 | 中 | 编写复现脚本，使用自测脚本自动化执行并记录。 |

---

## 6. 追踪日志

- `logs/219C2/219C2B-SELF-TEST-REPORT.md` – 原始自测报告。  
- `logs/219C2/z-followups/2025-11-05-validator-retest.md` – 回归验证记录。  
- 后续若有新增日志/截图，请存放在 `logs/219C2/z-followups/` 下。

> 文档更新完成后，请在主计划与 Team 06 进展日志中同步状态。
