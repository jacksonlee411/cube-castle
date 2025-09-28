# 15. 组织详情时间轴导航行为复核

**文档类型**: 缺陷调查 / 前端交互一致性

**创建日期**: 2025-09-19

**负责团队**: 前端团队（Owner） / CQRS 体验联席小组（协作）

**唯一事实来源**:
- `frontend/src/features/temporal/components/TimelineComponent.tsx`
- `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`
- `frontend/src/features/temporal/components/InlineNewVersionForm.tsx`
- `frontend/src/features/temporal/components/TemporalEditForm.tsx`

---

## 1. 背景与触发
- 业务方反馈：在组织详情页点击时间轴导航卡片时会弹出“编辑时态版本”模态窗，违背“点击节点即切换版本详情”的预期导航行为。
- 同一页面右侧“版本历史”区域已提供“修改记录”入口，时间轴上的“编辑”按钮疑似重复功能。
- 本调查旨在确认交互分工是否出现“重复事实来源”或“跨层不一致”违规，并给出整改建议。

## 2. 调查发现

### 2.1 时间轴卡片点击被绑定到编辑流程
- `TimelineComponent.tsx` 中卡片容器的 `onClick` 委托至 `onVersionSelect`。
- `TemporalMasterDetailView.tsx` 将 `onVersionSelect` 直接绑定到 `handleVersionSelect → handleEditHistory → setShowEditForm(true)`，每次点击都会拉起 `TemporalEditForm` 模态窗。
- 导航意图被编辑流程接管，导致“选择版本”与“编辑版本”没有分离。

### 2.2 时间轴内嵌“编辑”按钮与卡片点击逻辑完全一致
- `TimelineComponent.tsx` 在状态徽章旁渲染 `TertiaryButton`（文案“编辑”），点击同样调用 `onEditVersion`。
- `TemporalMasterDetailView.tsx` 将 `onEditVersion` 指向 `handleEditHistory`，与卡片点击触发链路相同，没有额外参数或权限控制。
- 结果是“点击卡片任意区域”与“点击编辑按钮”效果一致，形成重复入口。

### 2.3 右侧“版本历史”页签已提供完整编辑工作流
- `InlineNewVersionForm.tsx` 的“修改记录”按钮通过 `handleEditHistoryToggle` 进入内联编辑模式。
- 提交时调用 `handleEditHistorySubmit`（REST `PUT /organization-units/{code}/history/{recordId}`），并在同组件内维护校验、作废、插入新版本等操作。
- 该工作流已与 `TemporalMasterDetailView` 的状态、提示和权限耦合，是当前唯一的受控修改入口。

## 3. 原则评估
- **资源唯一性**: 同一页面存在三个并行“编辑入口”（时间轴卡片点击、时间轴按钮、版本历史按钮），指向同一事实但散落在不同 UI 单元，违背“唯一事实来源”与“入口一致”约束。
- **跨层一致性**: 时间轴组件的职责应聚焦“导航”，但当前实现强制触发“命令侧编辑”，与页面模块划分不符，破坏 CQRS 前端分层设计。

## 4. 整改建议与验收标准

| # | 事项 | 负责人 | 验收标准 | 状态 |
| - | --- | --- | --- | --- |
| 1 | 调整 `TimelineComponent`：卡片点击仅切换选中版本，不再触发编辑模态；移除或隐藏卡片内“编辑”按钮 | 前端团队 | 点击时间轴节点只更新右侧详情；控制台不再打印 `setShowEditForm(true)` | ✅ 已完成 |
| 2 | 保留“版本历史”页签作为唯一编辑入口，并在 UI 文案中注明“编辑需从版本历史区进入” | 前端团队 | 版本历史页签仍可正常创建/修改/删除；时间轴区域无编辑控件 | ✅ 已完成 |
| 3 | 更新前端交互规范（参考 `docs/reference/` 目录）并在回归前执行 `frontend/scripts/validate-field-naming*.js`、`npm run lint` | 文档与 QA 协作组 | 发布更新说明，CI 校验通过，避免新增一致性违规 | ⚠️ 阻塞中 |

> 阻塞说明：`node scripts/validate-field-naming.js` 当前输出大量历史命名违规（847 条 camelCase、454 条 snake_case 等），需另行治理；本次改动已记录执行结果，后续需在命名治理专项中统一回收。

---

## 5. 过度设计评估结论
- **评估日期**: 2025-09-19
- **评审结论**: 时间轴组件承担导航与编辑双重职责，导致交互重复、状态管理膨胀，被认定为过度设计。
- **整改共识**: 已确认采用第 4 节列出的方案，回收时间轴上的编辑能力，将编辑流程统一收敛至“版本历史”页签，并同步更新交互规范说明。

---

## 6. 最新进展与测试要求
- **实现进展**
  - `frontend/src/features/temporal/components/TimelineComponent.tsx` 已移除“编辑”按钮与相关回调，改为专职导航并提示“编辑请前往版本历史”。
  - `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx` 已删除 `TemporalEditForm` 模态触发逻辑，仅在页签区域保持单一编辑通路。
  - `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 新增时间轴职责约束，提醒开发与测试遵循唯一事实来源。

- **必备测试清单**
  - `npm run lint` ✅ 已通过。
  - `node scripts/validate-field-naming.js` ⚠️ 阻塞：脚本输出大量历史命名违规（847 条 camelCase、454 条 snake_case 等），本次变更未引入新增违例，待命名治理专项统一处理。
  - 手工验证场景（测试团队需覆盖）：
    - 在组织详情页点击时间轴卡片，仅更新选中版本，页面不再弹出编辑模态。
    - 确认时间轴区域不再出现“编辑”按钮或其他命令入口。
    - 于“版本历史”页签执行新增、修改、作废操作，确保流程仍可正常完成。
    - 切换到“审计历史”页签时，仍可加载选中版本的审计记录。

- **验收标准**
  - 时间轴导航卡片仅负责导航，无任何命令/编辑入口。
  - 编辑、作废、插入新版本等命令操作仅在“版本历史”页签内部提供。
  - lint 通过，字段命名校验存在的历史遗留已在阻塞说明中备案，不影响本次上线。
