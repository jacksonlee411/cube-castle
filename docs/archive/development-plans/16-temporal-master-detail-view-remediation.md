# 16 — TemporalMasterDetailView 逾期临时实现整改方案

文档类型：整改计划
创建日期：2025-09-24
责任团队：前端平台组（主责）＋ 实现清单守护代理（监督）
优先级：P0（已过期临时实现）
当前状态：已完成（D1 验收通过，2025-09-26）

---

## 1. 背景
- `docs/development-plans/10-implementation-inventory-maintenance-report.md` 将 TemporalMasterDetailView 的三项 `TODO-TEMPORARY` 标为逾期（原定 2025-09-20 完成），直接影响 IIG 审计与实现清单健康度。
- 该组件是时态管理体验的主入口，未完成的临时实现会导致表单模式未启用、状态映射缺失与历史编辑入口缺口，存在用户操作混乱与数据状态漂移风险。

---

## 2. Rectification Summary
1. **表单模式状态接管** — `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:107` 起新增 `showEditForm` setter 与 `formMode` 统一管理，时间轴点击或编辑入口会触发 `handleEditHistory`，右侧表单与弹窗均可切换至编辑态。
2. **状态映射补全** — `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:32` 新增 `mapLifecycleStatusToApiStatus`，`handleFormSubmit` 与 `handleHistoryEditSubmit` 在请求体内写入 `status`/`lifecycleStatus`，不再依赖后端默认值。
3. **历史编辑入口落实** — 在 `TimelineComponent` 增加“编辑”按钮（`frontend/src/features/temporal/components/TimelineComponent.tsx:365`），并实现 `handleEditHistory`，支持历史版本查看、修改与保存后刷新高亮。
4. **临时标记清零** — 所有与该视图相关的 `TODO-TEMPORARY` 已移除，整改内容同步至实现清单。

---

## 3. 目标
- 恢复 TemporalMasterDetailView 的完整交互：时间轴选择 → 编辑模式 → 保存/取消 → 状态一致。
- 明确并实现时态生命周期到业务状态的映射，确保请求体符合 `docs/api/openapi.yaml` 的 REST 契约。
- 补齐历史版本编辑入口，保证 `InlineNewVersionForm` 的编辑与插入流程可通过主视图触发。
- 更新实现清单/报告，移除逾期 `TODO-TEMPORARY` 告警。

---

## 4. 实施要点
### 4.1 表单模式与弹窗
- `showEditForm`、`formMode` 与 `formInitialData` 统一控制创建/编辑流转；`handleFormClose`、`handleHistoryEditClose` 按需回到新增态。
- 弹窗 `TemporalEditForm` 仅在 `formMode === 'edit'` 时打开，确保与内嵌表单状态一致。

### 4.2 状态映射与请求体
- `handleFormSubmit`、`handleHistoryEditSubmit` 使用 `mapLifecycleStatusToApiStatus` 写入 `status`，历史记录更新后通过 `loadVersions(false, recordId)` 精准回显。
- 创建/插入请求体包含 `lifecycleStatus`，符合 `docs/api/openapi.yaml`（`CreateOrganizationUnitRequest`、`CreateVersionRequest`）。

### 4.3 交互入口与刷新
- 时间轴卡片新增编辑按钮触发 `handleEditHistory`；历史保存/新版本插入后自动刷新时间轴并维持当前选项卡。
- 删除操作继续走 `DEACTIVATE` 事件，刷新后保持当前版本高亮。

### 4.4 文档与守卫
- 运行 `npm --prefix frontend run test -- --run` 验证 Vitest 用例通过。
- 实现清单、IIG 报告将同步更新（见 `docs/development-plans/10-implementation-inventory-maintenance-report.md`）。

---

## 5. 验收结果
- ✅ 代码中所有相关 `TODO-TEMPORARY` 已移除。
- ✅ REST 请求体包含 `status`/`lifecycleStatus`，与 GraphQL 返回值保持一致。
- ✅ `npm --prefix frontend run test -- --run` 通过，未新增 lint/类型告警。
- ✅ IIG 报告将该项转为“已整改”，实现清单与文档同步。
- 🔄 余留事项：路径字段（`path`）仍待 GraphQL 扩展，已在计划中保留提醒。

---

## 6. 里程碑
| 阶段 | 截止日期 | 负责人 | 状态 |
| --- | --- | --- | --- |
| D0：表单/状态逻辑实现 | 2025-09-26 | 前端平台组 | ✅ 完成（2025-09-26） |
| D1：交互验证与文档同步 | 2025-09-27 | IIG ＋ 文档组 | ✅ 完成（本文件更新即为凭证） |
| D2：CI 守卫复核 | 2025-09-28 | DevOps | ⏳ 依赖例行守卫脚本，暂无挂起事项 |

---

## 7. 影响与风险
- **契约漂移风险**：若后端仍假定默认状态，需要提前沟通以避免重复写入；计划在 D0 阶段完成接口确认。
- **UI 行为变更风险**：表单开启/关闭逻辑调整可能影响现有手动流程，需在 D1 阶段安排回归。
- **脚本同步风险**：实现清单与 IIG 报告需即刻同步，否则 CI 仍视作未整改；纳入完成判定。

---

## 8. 下一步
- 持续跟踪 GraphQL `codePath/namePath` 扩展（计划 2025-09-30 前完成）。
- DevOps 例行跑通守卫脚本并在周报中留存记录。
