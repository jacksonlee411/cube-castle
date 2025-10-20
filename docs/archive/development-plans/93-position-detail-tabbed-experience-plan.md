# 93号文档：职位详情多页签体验方案

**版本**: v0.2
**创建日期**: 2025-10-20
**状态**: 已完成
**实际完成日期**: 2025-10-19（实现并通过验收）
**关联计划**: 88号《职位管理前端功能差距分析》、80号《Position Management with Temporal Tracking》、06号《集成团队协作进展日志》
**唯一事实来源引用**: `CLAUDE.md`、`AGENTS.md`、`docs/api/schema.graphql`、`frontend/src/features/positions/PositionTemporalPage.tsx`、`frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`

> **文档说明**：本文档为事后补充的实施记录。实际功能已于 2025-10-19 完成并通过验收（详见 [93号验收报告](./93-position-detail-tabbed-experience-acceptance.md)），本方案文档于 2025-10-20 补充编写以完整记录设计决策与实现细节。

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
- 左栏：版本/时间轴导航。优先复用组织模块的 `TimelineComponent`，仅依赖 `recordId`、`status`、`isCurrent` 等基础字段；当前 GraphQL 仅返回 `status ∈ {ACTIVE, INACTIVE}` 与 `isCurrent`，`PLANNED/DELETED`、`dataStatus` 等值仍未从后端提供——任何扩展必须先更新 `docs/api/` 契约。若需临时保留 `PositionVersionList`，必须完成视觉对齐并记录降级说明。
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
2. 版本导航：改造 `PositionVersionList` 支持“侧栏导航”模式，添加选中态与回调；复用 `TimelineComponent` 时仅传递 `recordId`/`status`/`isCurrent` 等基础字段，保持二态现实（ACTIVE/INACTIVE）。若未来要展示 `PLANNED/DELETED`，必须先扩展 API 契约并同步 95 号报告。
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
- 左栏复用：`PositionVersionList` 当前为表格样式，改造成侧栏列表时需配合设计评审，确认 Canvas Kit 组件选型，并遵循 95 号文档的状态字段约束，不扩展额外状态标签。  
- 回归范围：重构涉及 Position 详情多处组件，需配合 Vitest 与 Playwright 规格更新（按 88号计划后续待办）。

## 11. 组件适配与前置校验

### 11.1 TimelineComponent 复用策略
- 新增 `PositionTimelineAdapter`（或等价包装函数），在查询结果映射阶段生成 `TimelineVersion` 所需字段：  
  - `unitType` 固定写入 `POSITION`，`level/sortOrder` 按 `effectiveDate` 派生，缺失时回退为 0。  
  - `codePath/namePath` 暂为空字符串，确保组件宽度计算正常；如后续需要组织链路再行扩展契约。  
  - `businessStatus/lifecycleStatus/dataStatus` 仅依据现有字段派生，不引入假设值，明确记录降级逻辑。  
- 若 `TimelineComponent` 内存在组织专有样式（例如层级色块），通过 `showStatusBadge=false` 或样式覆盖关闭；如无法关闭则拆出轻量版 `TimelineList` 并记录在案。

### 11.2 审计链路验证
- 在进入开发前，执行以下检查并记录于 06 号日志：  
  1. 通过 GraphQL `positionVersions` 抽样确认所有版本均携带 `recordId`。  
  2. 调用 `auditHistory(recordId)` 验证至少一条职位历史能返回非空结果；如为空需与命令服务核对审计写入流程。  
  3. 若发现缺口，优先修复命令服务审计逻辑（参考 `PositionService.logPositionEvent`），然后再落地审计页签。

### 11.3 设计评审里程碑
- 负责人：前端团队 @职位体验小组；协同：UX 设计、产品。  
- 计划时间：2025-10-25 前完成评审会议，形成《职位页签命名/排序确认稿》。  
- 交付物：页签命名、排序、移动端交互说明；评审结论需同步至 06 号日志与本计划附录。

### 11.4 响应式策略
- ≥1280px：采用左右分栏布局，左栏固定 320-360px，右栏自适应。  
- 960-1279px：左栏缩至 260px，并启用 `PositionVersionList` 折叠按钮，默认展开。  
- <960px：隐藏固定左栏，使用顶部 `Drawer` 展示版本列表；保留 Tabs，页签内模块按顺序垂直排布。  
- 需在实现阶段补充 Storybook/测试覆盖窄屏切换，并在 README/设计稿中标注降级策略。

## 12. 布局示意图

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

## 13. 后续工作

- 设计评审：按照 11.3 节排期完成评审会议并输出结论。  
- 时间轴适配：实现 11.1 节所述包装层或组件降级，提交复用验证记录。  
- 审计校验：执行 11.2 节三项检查，若需修复需附带脚本或补丁说明。  
- 技术实现：创建 `feature/position-tabbed-detail` 分支，分阶段提交并同步 06 号日志。  
- 测试计划：补充 Vitest 覆盖页签状态切换、审计加载；更新 Playwright 脚本覆盖“职位详情 → 审计历史”。  
- 文档同步：实施完成后更新 88 号文档第 7 节状态，并在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 登记。

> 草案提交人：前端团队 · 架构组（代）

## 14. 前置校验记录（2025-10-20）

1. **GraphQL recordId 验证**  
   - `docs/api/schema.graphql` 的 `positionVersions` 查询返回 `recordId`（第 146-183 行）。  
   - 查询服务 `Resolver.PositionVersions` 将结果映射到仓储 `GetPositionVersions`（`cmd/organization-query-service/internal/graphql/resolver.go:320-346`），仓储扫描函数 `scanPosition` 同步填充 `RecordIDField`（`cmd/organization-query-service/internal/repository/postgres_positions.go:1293-1364`）。  
   - 前端 `POSITION_DETAIL_QUERY_DOCUMENT` 请求 `recordId`，`transformPositionNode` 直接赋值至 `PositionRecord.recordId`（`frontend/src/shared/hooks/useEnterprisePositions.ts:299-361`、`:693-722`）。  
   - 结论：契约与实现均支持 `recordId`，后续仅需对历史缺失情况进行数据抽样确认。

2. **审计链路静态检查**  
   - 命令服务在创建、版本新增、填充、清空、转移、事件等操作中均调用 `logPositionEvent`，写入 `audit_logs`（`cmd/organization-command-service/internal/services/position_service.go:107-689`）。  
   - `logPositionEvent` 通过 `auditLogger.LogEvent` 写库，对应仓储 `audit_writer.go` 使用最新迁移字段写入 `audit_logs`（`cmd/organization-command-service/internal/repository/audit_writer.go:84-164`）。  
   - 结论：审计写入路径完备，后续需在实际环境执行抽样查询验证数据存在。

3. **设计评审排期**  
   - 评审负责人：前端 @职位体验小组，协同 UX、产品。  
   - 会议目标：确认页签命名/排序与响应式折叠策略，形成《职位页签命名/排序确认稿》。  
   - 计划时间：2025-10-25 前完成；评审结论将在 06 号日志与本计划中记录。

---

## 附录 A. 审计抽样脚本与操作指南

### A.1 GraphQL 抽样查询（验证 `recordId`）
```bash
curl -sS -X POST http://localhost:8090/graphql \
  -H "Authorization: Bearer $DEV_JWT" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query($code: PositionCode!) { positionVersions(code: $code, includeDeleted: false) { recordId code status effectiveDate isCurrent } }",
    "variables": { "code": "P1000001" }
  }' | jq .
```
- 期望：列表中所有版本均返回非空 `recordId`。  
- 若发现缺失，记录 `code` 与版本信息，通知后端核查数据迁移。

### A.2 GraphQL 审计历史抽样
```bash
curl -sS -X POST http://localhost:8090/graphql \
  -H "Authorization: Bearer $DEV_JWT" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query($recordId: String!) { auditHistory(recordId: $recordId, pageSize: 5) { data { auditId operationType operatedAt operatedBy { name } modifiedFields } } }",
    "variables": { "recordId": "REPLACE_WITH_RECORD_ID" }
  }' | jq .
```
- 期望：返回 `auditId`、`operationType`、`modifiedFields` 等字段。  
- 若返回空数组，确认对应版本是否曾触发命令；如确定应有记录则需排查命令服务审计写入。

### A.3 SQL 审计抽样（可选）
```sql
SELECT
  record_id,
  operation_type,
  operated_by_name,
  operated_at,
  target_type,
  target_id
FROM audit_logs
WHERE tenant_id = :TENANT_ID
  AND target_type = 'POSITION'
  AND target_id = :POSITION_RECORD_ID
ORDER BY operated_at DESC
LIMIT 5;
```
- 通过 `psql` 或 `pgcli` 执行，核对记录数量与 GraphQL 返回保持一致。

### A.4 抽样记录模板
| 抽样时间 | 职位编码 | 版本 `recordId` | 审计记录条数 | 结果 |
|----------|----------|-----------------|--------------|------|
| 2025-10-XX | P1000001 | 9da3-... | 3 | ✅ 正常 |

> 建议抽样覆盖至少 1 条当前版本、1 条历史版本；结果回填至 06 号日志第 10 节。

---

## 附录 B. 设计评审纪要模板

```
会议名称：职位详情多页签设计评审
会议时间：2025-10-25 14:00-15:00
与会人员：前端（姓名）、UX（姓名）、产品（姓名）、架构（姓名）

议题：
1. 页签命名与排序（概览 / 任职记录 / 调动记录 / 时间线 / 版本历史 / 审计历史）
2. 左侧版本导航样式（宽度、折叠、状态徽章）
3. 移动端/窄屏降级方案
4. Mock 模式下的只读指示

结论：
- 页签排序：概览 → 任职记录 → 调动记录 → 时间线 → 版本历史 → 审计历史
- 左侧导航：桌面宽度 320px，窄屏折叠为 Drawer
- 响应式策略：<960px 时默认折叠版本列表，通过按钮展开
- Mock 模式：顶部展示只读提示，禁用所有命令按钮

后续行动：
- 前端：根据结论更新组件拆分与样式实现（负责人 / 截止时间）
- 设计：提供最终视觉稿（负责人 / 截止时间）
- 文档：在 06 号日志与 93 号计划更新评审结果（负责人 / 截止时间）
```

> 建议使用该模板记录实际会议纪要，必要时归档至 `docs/archive/meetings/`。

---

## 15. 实施进展记录（2025-10-20）

- 多页签布局正式落地，涵盖「概览 / 任职记录 / 调动记录 / 时间线 / 版本历史 / 审计历史」六个内容区域，并保持桌面 + 窄屏自适应。  
- 左侧版本导航复用 `TimelineComponent`，支持版本选择同步 tabs；版本列表支持点击高亮与回到概览。  
- 审计页签接入 `AuditHistorySection`，选中版本缺少 `recordId` 时展示提示信息。  
- 前端单测：`npm --prefix frontend run test -- PositionTemporalPage` 通过（Vitest）。
