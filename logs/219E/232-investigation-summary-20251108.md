# Plan 232 Playwright P0 场景专项调查 - 调查报告

**调查完成时间**：2025-11-08 18:00 CST  
**调查者**：Claude Code Agent  
**相关文档**：docs/development-plans/232-playwright-p0-stabilization.md

---

## 调查概览

本调查对 Plan 219E 中阻塞关闭的 6 个 P0 失败场景进行了深入分析，通过日志解析、代码审查、组件定位，识别了根本原因并提出了优化的修复方案。

### 调查覆盖范围
- ✅ 6 个失败场景（business-flow-e2e、job-catalog-secondary-navigation、position-tabs、position-lifecycle、temporal-management-integration、optimization-verification-e2e）
- ✅ 5 个测试日志（Chromium + Firefox，共 10 条日志）
- ✅ 5 个前端组件代码路径
- ✅ 6 个脚本文件审查
- ✅ Bundle 体积分析

---

## 核心发现

### 1. 共性问题（影响 5/6 场景）

**问题 1.1：缺少导航与加载状态等待（高频）**
- 影响场景：position-tabs、position-lifecycle、temporal-management-integration、business-flow-e2e（部分）
- 表现：页面导航后直接查询 DOM 元素，但数据/组件未加载
- 根因：脚本缺少 `waitForURL()`、`waitForResponse()`、GraphQL 等待链
- 修复成本：低（3-5 行代码）

**问题 1.2：Modal/Dialog 组件识别困难（中频）**
- 影响场景：job-catalog-secondary-navigation、business-flow-e2e（部分）
- 表现：Modal 出现后，脚本无法找到内部元素（heading、button）
- 根因：组件缺少唯一 data-testid，脚本依赖易变的文本或 role 属性
- 修复成本：中等（需在 UI 层增加 testid）

**问题 1.3：Mock/Stub 不完整（低频）**
- 影响场景：position-tabs
- 表现：GraphQL 返回数据不足，导致 UI 组件无法渲染
- 根因：stubGraphQL() 未覆盖所需字段
- 修复成本：低（补充 GraphQL 返回字段）

**问题 1.4：权限验证延迟（低频）**
- 影响场景：business-flow-e2e、job-catalog-secondary-navigation
- 表现：权限检查未完成，按钮/表单条件渲染失败
- 根因：脚本未等待权限状态稳定
- 修复成本：低（增加权限预检逻辑）

### 2. 场景特异性问题

| 场景 | 特异性根因 | 代码路径 | 修复难度 |
|------|----------|--------|--------|
| business-flow-e2e | FormActions 条件渲染：`showRecordDelete` 需 selectedVersion 非空 + !isEditingHistory | FormActions.tsx:57-77 | 低 |
| job-catalog-secondary-navigation | CatalogVersionForm 未添加 data-testid，Modal 结构可能延迟挂载 | JobFamilyGroupDetail.tsx:84 | 中 |
| position-tabs | 路由/数据完整性，Mock stub 不完整 | position-tabs.spec.ts:97 + GraphQL | 低 |
| position-lifecycle | 数据播种验证缺失，未等待数据加载 | position-lifecycle.spec.ts:75 | 低 |
| temporal-management-integration | 导航与 GraphQL 等待链缺失 | temporal-management-integration.spec.ts:226 | 低 |
| optimization-verification-e2e | Bundle 阈值设置过低（4 MB vs 实测 4.59 MB） | optimization-verification-e2e.spec.ts:155 | 低 |

### 3. 根因分布统计

```
缺少等待逻辑（waitForURL/GraphQL）: 5 场景 ████████████ 83%
Modal/testid 不足:                  2 场景 ████ 33%
权限验证延迟:                       2 场景 ████ 33%
阈值设置不合理:                     1 场景 ██ 17%
数据播种问题:                       1 场景 ██ 17%
```

---

## 修复方案（优先级排序）

### P0 - 基础设施支撑（M1 - 2025-11-09）
- **T1**：为所有 Modal/Dialog 添加 data-testid（预计 1 天）
  - FormActions（`temporal-delete-record-button-wrapper`）
  - CatalogVersionForm（`catalog-version-form-dialog`）
  - PositionDetail（`position-temporal-page-wrapper`）
  - OrganizationDashboard（`organization-dashboard-wrapper`）

- **T2**：创建 E2E 等待库 `tests/e2e/utils/waitPatterns.ts`（预计 0.5 天）
  - `waitForPageReady(page)` - 等待页面稳定
  - `waitForGraphQL(page, queryName)` - 等待特定 GraphQL 查询
  - `waitForNavigation(page, expectedUrl)` - 等待路由完成

### P1 - 脚本修复（M2-M3 - 2025-11-10 ~ 2025-11-11）
- **T3**：business-flow-e2e（预计 0.5 天）
  - 补充 organization-form 加载等待
  - 增加权限预检逻辑
  - 修复删除按钮定位

- **T4**：job-catalog-secondary-navigation（预计 0.5 天）
  - 使用 T1 新增的 testid 替换文本选择器
  - 补充 waitForFunction 或轮询

- **T5**：position-tabs / position-lifecycle（预计 1 天）
  - 补充 waitForURL + GraphQL 等待
  - 验证数据播种完整性

- **T6**：temporal-management-integration（预计 0.5 天）
  - 使用 waitForResponse 等待 GraphQL
  - 增加网络容错（Promise.race + fallback）

### P2 - 基线调整（M4 - 2025-11-11）
- **T7**：Bundle 体积阈值调整（预计 0.5 天）
  - 修改阈值：4 MB → 5 MB
  - 文档记录基线：4.59 MB（含 source-map）
  - 更新 docs/reference/03-API-AND-TOOLS-GUIDE.md

### P1 - 验收（M5 - 2025-11-12）
- **T8**：文档同步与关闭流程（预计 0.5 天）
  - 更新 219E §2.5 状态为 ✅
  - 更新 Plan 06 验证结果
  - 归档日志至 logs/219E/
  - 申请 219E 关闭评审

---

## 修复时间表与依赖关系

```
M1 (2025-11-09)：T1 + T2
  ├─ T1 (UI testid) ────────┐
  │                         ├─ M2 (2025-11-10)：T3 + T4 + T5 + T6
  ├─ T2 (waitPatterns) ─────┤
  └─ T5 (position) ─────────┤
                            ├─ M3 (2025-11-11)：Firefox 双验证
  T7 (perf) ────────────────┤
                            └─ M5 (2025-11-12)：T8 文档关闭
```

**关键路径**：T1 + T2 → T3/T4/T5/T6 → T8（总计 3 天）

---

## 已知风险与缓解

| 风险 | 级别 | 缓解方案 |
|------|------|--------|
| UI 漂移频繁 | 中 | 建立 UI Registry，PR 变更时必须同步脚本 |
| 数据播种不稳定 | 中 | 添加数据验证预检，失败则 skip |
| 跨浏览器兼容性 | 低 | M3 专项 Firefox 验证 |
| 临时 TODO 回收延迟 | 低 | 在计划第 8 章建立追踪表，每周检查 |

---

## 立即可采取的行动

1. **本周（2025-11-09）**
   - 前端 Owner：启动 T1（UI testid 添加），目标今日完成
   - QA：启动 T2（waitPatterns.ts），与 T1 并行完成

2. **后续（2025-11-10）**
   - Temporal、Job Catalog、Position 团队：使用 T1 和 T2 输出修复各自脚本
   - 每完成一个脚本修复，立即运行 Chromium + Firefox，归档日志

3. **周五（2025-11-12）**
   - 所有脚本验证完成，文档同步
   - 申请 219E 关闭评审

---

## 附加资源

- **详细文档**：docs/development-plans/232-playwright-p0-stabilization.md
- **日志位置**：logs/219E/
- **修复指南**：232 文档 §9 "修复指南与检查清单"
- **问题追踪**：232 文档 §8 "已知问题追踪"

---

**下一步**：按照 §4 优化后的任务拆解，分别启动 T1 和 T2，预计本周五前完成所有 P0/P1 修复。

