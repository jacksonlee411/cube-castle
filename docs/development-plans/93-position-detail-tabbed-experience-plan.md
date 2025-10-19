# 93号文档：职位详情多页签体验方案

**版本**: v0.1  
**创建日期**: 2025-10-20  
**状态**: 草案（待评审）  
**关联计划**: 88号《职位管理前端功能差距分析》、80号《Position Management with Temporal Tracking》、06号《集成团队协作进展日志》  
**唯一事实来源引用**: `CLAUDE.md`、`AGENTS.md`、`docs/api/schema.graphql`、`frontend/src/features/positions/PositionTemporalPage.tsx`、`frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`

## 1. 背景

- 88号文档确认职位模块需与组织架构模块保持交互一致性，目前详情信息混合在单页卡片中，缺乏清晰的信息分区与历史透视。  
- 06号日志显示 P0/P1 版本表单与 GraphQL `positionVersions` 已上线，为多维视图提供数据基础。  
- 组织详情页面已通过 `TemporalMasterDetailView` 将“版本历史 / 审计历史”拆分为页签，形成成熟范式；职位详情需对齐该体验并遵循 CLAUDE.md 关于一致性与可回溯性的原则。

## 2. 约束与引用

- 遵循 `CLAUDE.md`：保持单一事实来源、复用 Docker 化 CQRS 架构、UI/文案使用专业中文。  
- 遵循 `AGENTS.md`：职位命令链路继续走 REST、查询依赖 GraphQL，字段命名保持 camelCase。  
- 契约依据 `docs/api/schema.graphql` 中的 `position`, `positionTimeline`, `positionVersions`, `positionAssignments`, `positionTransfers` 查询；后续如需字段扩展先更新契约。  
- 前端现状与可复用组件来源：`frontend/src/features/positions/PositionTemporalPage.tsx`、`frontend/src/features/positions/components`、`frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`。

## 3. 现状复盘

- `PositionTemporalPage` 在详情模式下以单列堆叠展示基础信息、任职记录、调动记录、时间线和版本工具，页面纵向过长且上下文切换成本高。  
- 编辑、新增版本表单以“展开收起”按钮控制，缺少视图级定位；审计历史未接入，无法支持合规回溯。  
- `PositionVersionList`、`PositionVersionToolbar` 已提供版本过滤/导出能力，但与其他信息杂糅在一个页面卡片中。

## 4. 设计目标

1. 对齐组织详情的“主从视图 + 页签”体验，降低信息噪音。  
2. 为历史维度提供专属页签，确保版本、时间线、审计信息可独立访问。  
3. 保持命令/查询分工：只改前端布局与查询组合，不引入新的 REST/GraphQL 端点。  
4. 明确创建模式、详情模式的差异化布局，避免多余交互步骤。  
5. 为后续增强（CSV 导出、includeDeleted、Future Version）预留操作区。

## 5. 方案概览

### 5.1 布局骨架

- 顶部保持现有返回与操作按钮区，补充当前位置面包屑（可复用 `TemporalMasterDetailHeader` 的结构逻辑）。  
- 主内容区采用左右分栏：
  - 左栏：版本/时间轴导航（复用 `PositionVersionList`，改造成“版本选择器”模式，支持 includeDeleted 过滤）。
  - 右栏：页签容器，容纳不同信息分区；默认选中“概览”。
- 移动端/窄屏 fallback：当宽度不足时左栏折叠为顶部抽屉式列表，页签保持。

### 5.2 模式差异

- 创建模式（/positions/new）：不展示左侧版本导航，右栏只保留“概览”页签，内嵌 `PositionForm` 创建流程。  
- 详情模式：左侧提供版本选择；右栏启用全部页签与表单入口。选择版本后右栏内容按页签刷新，保留已选 tab。

## 6. 页签定义

| 页签 | 定位与主要内容 | 组件/实现建议 | 数据来源 | 关键交互 |
| --- | --- | --- | --- | --- |
| 概览（默认） | 展示职位基本资料、当前状态、编制与当前任职摘要 | 拆分 `PositionDetails` 为基础信息模块；保留状态 pill 与调动按钮 | `position` 基本字段、`currentAssignment` | 切换版本时更新展示；顶部展示版本生效区间 |
| 任职记录 | 当前任职 + 历史列表、支持按状态筛选 | 提炼 `AssignmentItem` 为列表组件；可追加筛选器 | `positionAssignments` 数据集 | 支持搜索/筛选；默认高亮当前任职 |
| 调动记录 | 展示跨组织调动明细 | 复用 `TransferItem`，按时间倒序 | `positionTransfers` 数据集 | 允许按组织过滤（后续迭代可扩展） |
| 时间线 | 按时间轴展示职位状态变更与备注 | 复用 `TimelineItem`，可使用纵向时间线组件 | `positionTimeline` | 提供“仅显示当前/未来”快速筛选 |
| 版本历史 | 列表、过滤、CSV 导出、版本对比入口 | 将 `PositionVersionToolbar` + `PositionVersionList` 嵌入；版本选择驱动左栏同步 | `positionVersions` | 在右栏展示版本元数据；保留 includeDeleted、导出、对比操作按钮 |
| 审计历史 | 展示所选版本的审计日志 | 复用 `AuditHistorySection`（输入 recordId） | `auditHistory(recordId)` 查询 | 未选版本时提示用户选择；支持分页加载 |

## 7. 交互流程细节

1. 版本选择：左栏点击版本或在“版本历史”页签里点击某行，统一触发 `setActiveVersion(recordId)`；更新右栏所有页签内容。  
2. 表单入口：右上角按钮不变；提交成功后保持当前页签，刷新数据并回到“概览”。  
3. includeDeleted：控制左栏与“版本历史”页签的数据显示；与 GraphQL 查询参数对齐。  
4. 审计页签：若所选版本缺少 `recordId`，显示提示并禁用请求。记录选择后自动触发列表刷新。  
5. 错误处理：页签内容请求失败时显示局部错误卡片，不影响其他页签。

## 8. 数据与契约要求

- GraphQL 查询：继续复用 `usePositionDetail` 聚合请求，确保 `positionVersions` 返回 `recordId` 字段；若缺失需在 `schema.graphql` 与查询层补齐。  
- 审计日志：`AuditHistorySection` 依赖 `auditHistory(recordId)`，职位版本入参为 `PositionRecord.recordId`。查询服务需确认职位版本已写入审计记录。  
- REST 命令：编辑/新建版本仍使用现有 `PositionForm` + `usePositionMutations`，无需新增端点。  
- 状态同步：`react-query` key 保持 `positionDetailQueryKey(code)`；切换版本后通过本地状态驱动，不触发额外查询，避免破坏缓存。

## 9. 实施步骤与验收标准

1. UI 拆分：重构 `PositionTemporalPage`，引入 `PositionDetailTabs` 容器组件；将现有卡片内容拆分为独立子组件。  
2. 版本导航：改造 `PositionVersionList` 支持“侧栏导航”模式，添加选中态与回调。  
3. 审计页签：引入 `AuditHistorySection`，确保 recordId 传递正确并添加空状态文案。  
4. 状态管理：在页面级维护 `activeTab`、`selectedVersionRecordId`；预留 URL 查询参数占位（便于分享）。  
5. 验收标准：
   - 切换页签或版本不会刷新整个页面，只更新对应内容。  
   - includeDeleted 选项对左栏和“版本历史”页签同步生效。  
   - 审计页签加载成功显示日志表格；无数据时提示“暂无审计记录”。  
   - `npm --prefix frontend run test -- PositionTemporalPage` 及新增单测全部通过。

## 10. 风险与依赖

- recordId 完整性：部分历史版本可能缺失 recordId（历史迁移遗留），需要在查询层补齐或提供降级提示。  
- 审计数据量：大批量记录加载可能影响性能，需启用 `AuditHistorySection` 内置的分页/limit 参数。  
- 左栏复用：`PositionVersionList` 当前为表格样式，改造成侧栏列表时需配合设计评审，确认 Canvas Kit 组件选型。  
- 回归范围：重构涉及 Position 详情多处组件，需配合 Vitest 与 Playwright 规格更新（按 88号计划后续待办）。

## 11. 布局示意图

```mermaid
flowchart LR
  subgraph Container[PositionTemporalPage]
    direction LR
    subgraph LeftPane[左侧：版本导航]
      VList[PositionVersionList \n· includeDeleted过滤 \n· 选中版本高亮]
    end
    subgraph RightPane[右侧：页签容器]
      Tabs[TabNavigation]
      subgraph TabStacks[页签内容]
        direction TB
        Overview[概览 \n· 基础信息/当前任职]
        Assignments[任职记录 \n· AssignmentList]
        Transfers[调动记录 \n· TransferList]
        Timeline[时间线 \n· TimelineView]
        Versions[版本历史 \n· VersionToolbar + VersionTable]
        Audit[审计历史 \n· AuditHistorySection(recordId)]
      end
    end
  end

  VList --> Tabs
  Tabs --> Overview & Assignments & Transfers & Timeline & Versions & Audit
```

> 示意图说明：左侧 `PositionVersionList` 控制当前版本，右侧通过 `TabNavigation` 切换，各页签分别渲染相应组件并共享选中版本上下文。创建模式时隐藏 `LeftPane` 并仅保留“概览”页签。

## 12. 后续工作

- 设计评审：与 UX/业务确认页签命名与排序，确保与组织模块一致。  
- 技术实现：创建 `feature/position-tabbed-detail` 分支，分阶段提交并同步 06号日志。  
- 测试计划：补充 Vitest 覆盖页签状态切换、审计加载；更新 Playwright 脚本覆盖“职位详情 → 审计历史”。  
- 文档同步：实施完成后更新 88号文档第 7 节状态，并在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 登记。

> 草案提交人：前端团队 · 架构组（代）
