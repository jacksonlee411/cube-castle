# 16 — TemporalMasterDetailView 逾期临时实现整改方案

文档类型：整改计划
创建日期：2025-09-24
责任团队：前端平台组（主责）＋ 实现清单守护代理（监督）
优先级：P0（已过期临时实现）
当前状态：进行中（D0 准备中）

---

## 1. 背景
- `docs/development-plans/10-implementation-inventory-maintenance-report.md` 将 TemporalMasterDetailView 的三项 `TODO-TEMPORARY` 标为逾期（原定 2025-09-20 完成），直接影响 IIG 审计与实现清单健康度。
- 该组件是时态管理体验的主入口，未完成的临时实现会导致表单模式未启用、状态映射缺失与历史编辑入口缺口，存在用户操作混乱与数据状态漂移风险。

---

## 2. 问题现状
1. **表单模式状态未落地** — `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:97` 保留 `formMode` 状态但未驱动 UI，`showEditForm` 亦缺 setter，导致选中时间轴版本时右侧表单无法切换至编辑模式。
2. **状态映射缺失** — `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:433` 未实现 `mapLifecycleStatusToApiStatus`，新建组织/版本仍依赖后端默认状态，违背“先契约后实现”，也使 `lifecycleStatus` 与 REST `status` 不一致。
3. **历史编辑入口缺口** — `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:533` 的 `handleEditHistory` 留白，时间轴仅能触发选择/作废，无法把选中版本推入编辑流程，弱化 `InlineNewVersionForm` 中的历史编辑体验。

---

## 3. 目标
- 恢复 TemporalMasterDetailView 的完整交互：时间轴选择 → 编辑模式 → 保存/取消 → 状态一致。
- 明确并实现时态生命周期到业务状态的映射，确保请求体符合 `docs/api/openapi.yaml` 的 REST 契约。
- 补齐历史版本编辑入口，保证 `InlineNewVersionForm` 的编辑与插入流程可通过主视图触发。
- 更新实现清单/报告，移除逾期 `TODO-TEMPORARY` 告警。

---

## 4. 解决方案
### 4.1 表单模式与可见性接管
- 将 `const [showEditForm] = useState(...)` 改为带 setter 的 `useState`；移除冗余的 `editMode` 常量，仅保留 `formMode`。
- 在 `handleVersionSelect` 中 `setFormMode('edit')`、`setShowEditForm(true)`，同时预填 `formInitialData` 并保持 `selectedVersion` 与 `activeTab` 同步。
- 在 `handleFormClose`、新建成功、作废成功后恢复为 `setFormMode('create')`、`setShowEditForm(false)`、`setFormInitialData(null)`，确保回到新增态。
- 将 `InlineNewVersionForm` 的 `mode` 属性与 `TemporalEditForm` 渲染条件改为依据 `formMode`，保证 UI 与状态一致。

### 4.2 状态映射与请求体修正
- 实现 `mapLifecycleStatusToApiStatus(lifecycleStatus: TemporalEditFormData['lifecycleStatus'])`，返回 `OrganizationStatus`（`ACTIVE`/`INACTIVE`/`PLANNED`），逻辑参考 `frontend/src/shared/utils/statusUtils.ts`。
  - 映射建议：`CURRENT→ACTIVE`、`PLANNED→PLANNED`、`HISTORICAL/INACTIVE/DELETED→INACTIVE`。
- 在 `handleFormSubmit` 创建/插入路径中补充 `status: mapLifecycleStatusToApiStatus(formData.lifecycleStatus)`，并同步传入 `InlineNewVersionForm` 的历史编辑提交路径（`handleHistoryEditSubmit`）。
- 对照 `docs/api/openapi.yaml` 的 `CreateOrganizationUnitRequest` 与 `CreateVersionRequest`，必要时补充 `lifecycleStatus` 字段至 GraphQL 请求或与后端确认契约，避免重复默认值逻辑。
- 变更后执行 `npm --prefix frontend run test`、`npm --prefix frontend run test:contract`，验证前后端契约未破。

### 4.3 历史编辑入口与交互
- 实现 `handleEditHistory`：接收 `TimelineVersion`，设置 `selectedVersion`、`formMode('edit')`、`setShowEditForm(true)`、`setActiveTab('edit-history')`，并将版本数据传给 `formInitialData`。
- 在 `TimelineComponent` 或右侧操作区新增「编辑此版本」入口（可透传回调），调用 `handleEditHistory` 进入编辑态。
- 调整 `InlineNewVersionForm`：若存在 `selectedVersion` 且 `formMode === 'edit'`，默认进入只读视图，点击「修改记录」时从 `TemporalMasterDetailView` 取回的版本数据无缝衔接；插入新版本后，通过 `loadVersions()`+`setFormMode('create')` 刷新。

### 4.4 文档与守卫
- 运行 `node scripts/generate-implementation-inventory.js`、更新 `reports/implementation-inventory.json` 与 `reports/iig-guardian/iig-guardian-report.json`，确保临时实现列表已清零。
- 执行 `bash scripts/check-temporary-tags.sh`，确认文件内无残留 `TODO-TEMPORARY`。
- 在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 的时态管理章节追加状态映射说明，保持“单一事实来源”。

---

## 5. 验收标准
- `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx` 无 `TODO-TEMPORARY` 注释；手动测试：
  1. 创建模式下成功提交组织 → 自动跳转详情。
  2. 时间轴选择历史版本 → 右侧进入编辑只读视图，点击「修改记录」可保存并刷新列表。
  3. 插入新版本后，时间轴与路径面包屑即时更新。
- REST 请求体包含正确的 `status`，后端日志无默认值警告；GraphQL 查询仍可回显正确状态。
- 前端测试与契约检查均通过，无新增 lint/类型告警。
- IIG 报告与实现清单不再标红 TemporalMasterDetailView。

---

## 6. 里程碑
| 阶段 | 截止日期 | 负责人 | 验收要点 |
| --- | --- | --- | --- |
| D0：表单/状态逻辑实现 | 2025-09-26 | 前端平台组 | 代码合并、单元测试通过，`TODO` 清零 |
| D1：交互验证与文档同步 | 2025-09-27 | IIG ＋ 文档组 | 手工验证流程、报告与 Reference 更新完毕 |
| D2：CI 守卫复核 | 2025-09-28 | DevOps | `scripts/check-temporary-tags.sh` 与契约/质量脚本全部绿灯 |

---

## 7. 影响与风险
- **契约漂移风险**：若后端仍假定默认状态，需要提前沟通以避免重复写入；计划在 D0 阶段完成接口确认。
- **UI 行为变更风险**：表单开启/关闭逻辑调整可能影响现有手动流程，需在 D1 阶段安排回归。
- **脚本同步风险**：实现清单与 IIG 报告需即刻同步，否则 CI 仍视作未整改；纳入完成判定。

---

## 8. 下一步
- 指派具体开发者认领 D0 任务，并在 MR 中引用本整改计划。（进行中，已获管理批准）
- 开发完成后附带测试截图/日志与脚本运行输出，便于 IIG 验收与归档。
