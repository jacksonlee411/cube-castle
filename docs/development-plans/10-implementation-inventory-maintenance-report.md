# 10. 实现清单维护报告

**文档类型**: 评审报告
**创建日期**: 2025-09-21
**负责团队**: 实现清单守护代理 (IIG)
**优先级**: P0（包含过期临时实现）

---

## 执行概览

### 📊 完成状态
- ✅ 实现清单文档已更新至 v1.9.0
- ✅ 架构合规性验证通过（109 个文件，0 违规）
- ⚠️ 发现 6 项临时实现已过期仍在使用

---

## 关键发现

### 🚨 紧急处理项（截止日期已过）
1. **temporalValidation.ts** — 截止 2025-09-16 已过期
   - 文件：`frontend/src/features/temporal/utils/temporalValidation.ts`
   - 现状：仍被 `TemporalDatePicker` 直接引用为核心校验工具。
   - 行动：制定迁移脚本，将引用统一切换至 `shared/utils/temporal-converter.ts` 后删除；若需保留，请更新 `TODO-TEMPORARY` 截止信息并在清单备案。

2. **OrganizationForm/ValidationRules.ts** — 截止 2025-09-16 已过期
   - 文件：`frontend/src/features/organizations/components/OrganizationForm/ValidationRules.ts`
   - 现状：文件顶部声明 2025-09-16 停用，但仍被表单逻辑导入。
   - 行动：确认表单已支持 `shared/validation/schemas.ts` 后移除该兼容层；或重新评估并调整日期。

3. **API 类型临时导出** — 截止 2025-09-16 已过期
   - 文件：`frontend/src/shared/types/api.ts`
   - 现状：`APIError`、`ValidationError` 及守卫函数仍以临时别名导出，造成类型入口重复。
   - 行动：完成到新错误处理体系的替换，删除临时导出并更新引用路径。

4. **useEnterpriseOrganizations Hook 标记冲突** — 截止 2025-09-16 已过期
   - 文件：`frontend/src/shared/hooks/useEnterpriseOrganizations.ts`
   - 现状：文件底部 `// TODO-TEMPORARY` 标注“将在 2025-09-16 后删除”，但该 Hook 为组织列表页等核心依赖。
   - 行动：立即更新注释与计划，明确真正需要删除的是旧 wrapper `useOrganizations`，或设定新的迁移目标与截止日期，避免误导后续清理。

5. **organizationPermissions.ts 子组织校验禁用** — 截止 2025-09-20 已过期
   - 文件：`frontend/src/shared/utils/organizationPermissions.ts`
   - 现状：`childCount` 防删逻辑仍被注释，权限计算缺乏真实数据约束。
   - 行动：恢复 API 集成或制定延期说明，并在 IIG 清单登记。

6. **TemporalMasterDetailView 时态功能缺口** — 截止 2025-09-20 已过期
   - 文件：`frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`
   - 现状：以下三处 `TODO-TEMPORARY` 均已逾期：
     - 表单模式状态未使用（第 96 行）
     - 状态映射 `mapLifecycleStatusToApiStatus` 未实现（第 431 行）
     - 历史编辑操作 `handleEditHistory` 未补齐（第 531 行）
   - 行动：补齐对应逻辑或正式删除临时占位，并同步更新验证与文档。

### ⏱ 即将到期项
- **TemporalMasterDetailView 路径占位** — 截止 2025-09-30
  - 文件：`frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`（第 372 行）
  - 说明：`path` 字段仍为占位符，需在 v4.3 前完成 GraphQL `codePath/namePath` 对接。

---

## 架构合规成果

### ✅ 已修复问题
- camelCase 命名规范违规 3 项已全部修复。
- 架构验证器检查通过率 100%（109/109 文件）。
- REST/OpenAPI 契约与实现保持一致。

### 📈 统计更新
- REST 端点：17 → 26（新增运维监控 9 个）。
- 认证端点：7 → 8（新增 logout GET）。
- GraphQL 查询：精简至 9 个主要字段，并新增审计相关查询。
- 前端导出组件：162 → 140，重复导出下降 22 个。

---

## 符合项目原则验证
- **单一事实来源**：契约、实现清单与代码一致。
- **唯一性原则**：重复组件持续降低，无新增架构违规。
- **API 优先**：所有 26 个端点均先契约后实现。
- **CQRS 架构**：命令/查询隔离，端口配置标准化。

---

## 行动计划

### P0 — 立即执行
1. 对 `temporalValidation.ts`、`ValidationRules.ts`、`shared/types/api.ts` 完成迁移或延期说明，并更新 `TODO-TEMPORARY`。
2. 纠正 `useEnterpriseOrganizations` 标记，明确真实迁移目标与时间表。
3. 恢复 `organizationPermissions.ts` 子组织校验或提供风险评估。
4. 补齐 `TemporalMasterDetailView` 三个逾期功能点。

### P1 — 短期内（下一个迭代）
5. 完成 `TemporalMasterDetailView` 路径占位的 GraphQL 接入。
6. 对全部 `TODO-TEMPORARY` 执行自动核查，确保日期与责任人同步更新。

### P2 — 持续改进
7. 建立 IIG 周期性审计（日历化提醒 + PR 审查检查项）。
8. 将临时实现列表纳入 CI，结合 `scripts/check-temporary-tags.sh` 自动阻断过期项。

---

## 评审建议
- 重点关注与确认：
  1. 三个核心校验/类型文件的迁移计划与 Owner。
  2. `useEnterpriseOrganizations` 迁移策略及对外 API 稳定性承诺。
  3. `TemporalMasterDetailView` 功能缺口的技术排期。
  4. 权限校验恢复对风险敞口的影响评估。
- 决策需求：
  - [ ] 批准更新后的 P0 处理顺序与负责人。
  - [ ] 确认延期条目（若无法立即修复）。
  - [ ] 设定 IIG 巡检节奏与变更沟通渠道。

---

## 总结
- 实现清单 v1.9.0 已与当前代码同步，架构合规性保持 100%。
- 仍有 6 项已过期临时实现需要立即处理，其中部分为关键依赖，需谨慎迁移。
- 建议在本迭代内完成 P0 清理，并通过自动化手段防止临时实现再次超期。

---

**下一步**：等待相关团队认领任务并执行上述行动计划。
