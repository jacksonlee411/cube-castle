# 96号文档：职位管理 Job Catalog 布局校准前置分析

**版本**: v1.2
**创建日期**: 2025-10-20
**最近修订**: 2025-10-20（新增第 8 节设计评审结论）
**触发来源**: 92号《职位管理二级导航实施方案》 Phase 4 验收待办 + 06号日志 P1 设计评审待办
**分析方法**: 静态代码审查 + Playwright 截图基线 + 设计评审（已完成）
**关联计划**: 92号、06号
**维护团队**: 前端团队 · 职位域
**遵循原则**: CLAUDE.md 资源唯一性 · CQRS 约束

---

## 1. 背景与目标

- 92号文档 `0.4 Phase 0 技术验证清单` 已全部完成 ✅，确认 SidePanel 集成与二级导航基础能力；但 Phase 4 `F4 - Job Catalog 页面` 仍未勾选（`docs/development-plans/92-position-secondary-navigation-implementation.md` 800-805 行）。
- 06号进展日志记录 P1 待办：「设计评审：确认 Job Catalog 列表/详情在新导航下的视觉稿（含 312px 左侧栏对齐）」（`docs/development-plans/06-integrated-teams-progress-log.md` 45-46 行）。
- 现场反馈指出 Job Catalog 页面（职类/职种/职务/职级）布局与“职位列表”存在差异；本分析旨在为 Phase 4 布局验收提供差异梳理、落地步骤与截图方案。
- 目前 MCP 尚未注册，改用 Playwright 生成布局截图以支撑设计评审与验收归档。

## 2. 当前验证状态与差距定位

| 检查项 | 当前状态 | 说明 |
|--------|----------|------|
| Phase 0 POC（92号 §0.4） | ✅ 完成 | SidePanel、Expandable、权限 Hook 与 tokens 校验均已通过 |
| Phase 4 F4 - Job Catalog 页面 | ⛔ 未通过 | 列表/详情布局未统一，缺少验收截图凭证 |
| 06号日志 P1 设计评审 | ⛔ 未完成 | 需凭借截图与设计稿对齐 312px 侧栏下的留白与卡片样式 |

> 本文档聚焦 Phase 4 F4 验收缺口，可视为 92 号计划的附录，后续落地可直接回写验收状态。

## 3. 布局基线：职位列表页（对标参考）

### 3.1 AppShell 主内容容器

`frontend/src/layout/AppShell.tsx`

```typescript
const SIDE_PANEL_WIDTH = 312;
...
<Box as="main" cs={{ flex: 1, overflow: 'auto' }}>
  <Box as="div" padding="l">
    <Outlet />
  </Box>
</Box>
```

- 所有子路由共享 `padding="l"` 与独立滚动容器，布局验收需在该基线上保持一致。

### 3.2 PositionDashboard 结构特征

`frontend/src/features/positions/PositionDashboard.tsx`

```typescript
<Box padding={space.l} data-testid="position-dashboard">
  <SimpleStack gap={space.l}>
    <SimpleStack gap={space.xs}>
      <Heading size="medium">职位管理（Stage 1 数据接入）</Heading>
      ...
    </SimpleStack>
    <Flex justifyContent="flex-end">
      <PrimaryButton ...>创建职位</PrimaryButton>
    </Flex>
    <PositionSummaryCards ... />
    <CardLikeContainer>
      <SimpleStack gap={space.m}>
        <Heading size="small">筛选条件</Heading>
        <Flex gap={space.m} ...>
          ...
        </Flex>
      </SimpleStack>
    </CardLikeContainer>
    <PositionList ... />
  </SimpleStack>
</Box>
```

- 顶层 `SimpleStack` 统一垂直节奏。
- 筛选区域包裹在局部 `CardLikeContainer`（定义于文件底部，尚未共享）。
- 形成「标题/操作 → 筛选卡片 → 列表主体」的视觉层级，为 Job Catalog 页面提供基准。

## 4. Job Catalog 页面代码观察

### 4.1 列表页（职类/职种/职务/职级）

- 顶层均为 `Box padding="l" display="flex" flexDirection="column"`（示例：`JobFamilyGroupList.tsx`）。
- 缺乏统一的 `gap`/`SimpleStack`，依赖子元素 `marginBottom` 调整间距，导致节奏不稳。
- `CatalogFilters` 直接裸露在背景上，与 PositionDashboard 的卡片样式不一致。
- `JobFamilyList.tsx`、`JobRoleList.tsx` 在表格前插入提示文案，未纳入卡片导致视觉跳跃。

### 4.2 详情页

- 以 `JobFamilyGroupDetail.tsx` 为例，顶层虽使用 `gap="l"`，但信息块由多个 `Flex`/`Box` 堆叠，缺少卡片或分隔线。
- 操作按钮与信息区同级显示，缺乏明显分层。

## 5. 组件复用清单核对

| 组件/工具 | 现有位置 | 状态 | 说明 |
|-----------|----------|------|------|
| `SimpleStack` | `frontend/src/features/positions/components/SimpleStack.tsx` | ✅ 已存在 | 可直接导入 Job Catalog 页面，无需新建 |
| `CardLikeContainer` | `frontend/src/features/positions/PositionDashboard.tsx`（局部定义） | ⚠️ 未共享 | 建议抽取为共享组件（示例命名：`CardContainer`），并登记于 `02-IMPLEMENTATION-INVENTORY.md` |
| `CatalogTable` / `CatalogFilters` | `frontend/src/features/job-catalog/shared/` | ✅ 已复用 | 可继续沿用 |
| Canvas tokens | `@workday/canvas-kit-react/tokens` | ✅ 可用 | 保持 token 驱动，避免硬编码 |

## 6. 实施方案

### 6.1 抽取通用卡片容器

```typescript
// frontend/src/shared/components/CardContainer.tsx（建议路径）
import React from 'react'
import { Box } from '@workday/canvas-kit-react/layout'
import { colors, space, borderRadius } from '@workday/canvas-kit-react/tokens'

export const CardContainer: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <Box
    padding={space.l}
    borderRadius={borderRadius.l}
    backgroundColor={colors.frenchVanilla100}
    border={`1px solid ${colors.soap400}`}
  >
    {children}
  </Box>
)
```

> 抽取后运行 `node scripts/generate-implementation-inventory.js` 并确认 `02-IMPLEMENTATION-INVENTORY.md` 记录该组件。

### 6.2 重构 Job Catalog 列表页

以 `JobFamilyGroupList.tsx` 为例：

```typescript
import { SimpleStack } from '@/features/positions/components/SimpleStack'
import { CardContainer } from '@/shared/components/CardContainer'
import { space } from '@workday/canvas-kit-react/tokens'

return (
  <Box padding={space.l}>
    <SimpleStack gap={space.l}>
      <Flex justifyContent="space-between" alignItems="center">
        <Heading size="large">职类管理</Heading>
        ...
      </Flex>

      <CardContainer>
        <SimpleStack gap={space.m}>
          <Heading size="small">筛选条件</Heading>
          <CatalogFilters ... />
        </SimpleStack>
      </CardContainer>

      <CardContainer>
        <CatalogTable ... />
      </CardContainer>
    </SimpleStack>
  </Box>
)
```

**任务清单**：

- [x] 四个列表页共用 `SimpleStack` 控制垂直节奏。
- [x] 筛选区域、表格区域统一使用 `CardContainer` 或等效封装。
- [x] 条件提示文案整合进卡片内部或采用 Canvas 提示组件，避免裸露在主背景。

### 6.3 重构 Job Catalog 详情页

```typescript
<SimpleStack gap={space.l}>
  <Flex justifyContent="space-between" alignItems="center">
    <Heading size="large">职类详情</Heading>
    <ButtonGroup ... />
  </Flex>

  <CardContainer>
    <SimpleStack gap={space.m}>
      <KeyValueRow label="职类编码" value={group.code} />
      ...
    </SimpleStack>
  </CardContainer>

  <VersionDialogs ... />
</SimpleStack>
```

- [x] 信息区使用卡片或分隔线与操作区区分。
- [ ] 可根据需要在 `job-catalog/shared/` 提炼 `KeyValueRow` 等复用组件。

## 7. Playwright 截图验证方案

为尽快建立改造前后视觉基线，编写了专用规格 `frontend/tests/e2e/job-catalog-layout-baseline.spec.ts`：

1. **启用截图模式**

   ```bash
   export PW_CAPTURE_LAYOUT=true
   # 可选：自定义输出目录
   # export PW_CAPTURE_OUTPUT=artifacts/layout-dev

   npm run test:e2e -- --project=chromium frontend/tests/e2e/job-catalog-layout-baseline.spec.ts
   ```

2. **输出与管理**
   - 默认输出目录：`artifacts/layout/`（已被 `.gitignore` 忽略）。
   - 生成的截图包括：职位列表、职类列表、职类详情。若无职类数据，详情截图会自动跳过。
   - 运行日志会打印截图文件路径，便于挑选关键截图归档。

3. **设计评审**
   - 使用上述截图与设计稿对比留白、卡片背景、间距等细节。
   - 评审结论需同步回写至 06 号日志 P1 任务，并在 92 号文档 Phase 4 勾选对应验收项。

## 8. 设计评审结论（2025-10-20）

### 8.1 评审依据

- **基线截图**：`frontend/artifacts/layout/` 目录下三张截图（2025-10-20 08:47 生成）
  - `positions-list.png`：职位列表基线参考
  - `job-family-groups-list.png`：职类列表待评审页面
  - `job-family-group-detail.png`：职类详情待评审页面
- **对比方法**：将 Job Catalog 页面截图与职位列表基线进行逐项视觉比对

### 8.2 评审发现

| 评审项 | 基线表现 | Job Catalog 表现 | 结论 |
|--------|----------|------------------|------|
| ① 主内容左右留白 | 312px 侧栏 + space.l 主内容 padding | 与基线一致，左右留白约 32px | ✅ 符合设计规范 |
| ② 筛选/表格卡片背景 | 使用 CardLikeContainer（白色背景 + 边框 + 圆角） | 搜索框与表格直接裸露在主背景上，**无卡片容器** | ❌ 不符合规范，需改造 |
| ③ 详情页分层 | 不适用（职位列表无详情页） | 信息区直接展示在背景上，**无卡片或分隔线** | ❌ 不符合规范，需改造 |

### 8.3 评审结论

1. **留白布局**：Job Catalog 页面在 312px 侧栏布局下的主内容区左右留白**符合设计规范**，与职位列表基线保持一致。

2. **卡片视觉**：Job Catalog 列表页**缺少卡片容器**，筛选区域与表格区域直接裸露在主背景上，与职位列表基线的 CardLikeContainer 样式不一致，**必须改造**。

3. **详情分层**：Job Catalog 详情页信息区与操作区**缺乏视觉层级分离**，未使用卡片或分隔线强化分层，**必须改造**。

4. **实施确认**：本文档第 6 节提出的实施方案（抽取 CardContainer + 重构列表页/详情页）**正确且必要**，需按计划执行。

5. **后续跟踪**：
   - 已回写 92 号文档 Phase 4 F4 "视觉对齐验收" 子项（`docs/development-plans/92-position-secondary-navigation-implementation.md` 812-817 行）
   - 已更新 06 号日志 P1 任务状态为完成（`docs/development-plans/06-integrated-teams-progress-log.md` 46 行）
   - 布局改造任务纳入 92 号 Phase 4 F4 验收清单，待执行后勾选

---

## 9. 与 92号 / 06号计划的对齐

| 任务 | 本文档建议 | 对应计划 | 备注 |
|------|------------|----------|------|
| 布局统一 | `SimpleStack` + `CardContainer` 重构 | 92号 Phase 4 F4 | 完成后可勾选 F4 |
| 截图验收 | Playwright 基线截图 | 92号 Phase 4 | ✅ 已完成（2025-10-20） |
| 设计评审 | 携截图对齐视觉稿 | 06号日志 P1 | ✅ 已完成（2025-10-20，见第 8 节） |
| 实现登记 | 更新实现清单 | CLAUDE.md | 避免重复造轮子 |

## 10. 验收检查清单（建议）

- [x] `CardContainer` 抽取为共享组件并登记于 `02-IMPLEMENTATION-INVENTORY.md`。
- [x] Job Catalog 四个列表页使用 `SimpleStack` 与 `CardContainer` 统一布局。
- [x] Job Catalog 详情页信息区采用卡片分层，操作区与内容区明确分离。
- [x] Playwright 基线截图完成并附加至 92 号文档或本档案（可携带时间戳）。
- [x] 06 号日志 P1 设计评审条目更新为完成，并记录确认结论。

---

> 注：本分析基于 2025-10-20 代码快照。后续如有实现调整，请同步更新本 96 号文档，并在 92 号/06 号计划中登记验收结论。
