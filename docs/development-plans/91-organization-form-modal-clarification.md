# 91号文档：组织架构表单 Modal 误用澄清与优化建议

**版本**: v0.1  
**创建日期**: 2025-10-18  
**问题级别**: 🟡 中风险（容易造成误解）  
**维护团队**: 前端团队 · 架构组  
**关联文档**: 88号《职位管理前端功能差距分析》、06号《集成团队协作进展日志》

---

## 1. 背景

- 组织架构模块当前的创建/编辑流程采用**独立详情页**（`/organizations/new`、`/organizations/:code/temporal`）完成。
- `OrganizationDashboard` 文件仍保留 `OrganizationForm`（Canvas Kit Modal）的调用，但 `isFormOpen` 常量始终为 `false`，导致 Modal 从未实际渲染。
- 在代码阅读或评审中，团队成员容易误以为“组织模块通过 Modal 完成创建”，与现状不符，进而影响职位模块的交互设计讨论。

---

## 2. 问题描述

| 维度 | 现状 | 影响 |
|------|------|------|
| 代码结构 | `OrganizationDashboard` 导入 `OrganizationForm`，但 Modal 永远不打开 | 代码与实际行为不一致，增加理解成本 |
| 文档说明 | 历史文档仍提到“组织表单使用 Modal” | 对比职位模块时产生错误结论 |
| 交互一致性 | 组织与职位都已采用“页面式表单” | Modal 残留代码成为干扰项 |

---

## 3. 证据

```typescript
// frontend/src/features/organizations/OrganizationDashboard.tsx:118-138
const isFormOpen = false;
const selectedOrg: OrganizationUnit | undefined = undefined;
const handleFormClose = () => {};
const handleFormSubmit = () => {};

// ...
{!isHistorical && (
  <OrganizationForm
    organization={selectedOrg}
    isOpen={isFormOpen}          // 始终 false
    onClose={handleFormClose}
    onSubmit={handleFormSubmit}
  />
)}
```

Playwright 场景 `frontend/tests/e2e/organization-create.spec.ts` 也验证了点击“新增组织”后 URL 跳转到 `/organizations/new`，而非出现 Modal。

---

## 4. 建议方案

### 4.1 方案一：移除未使用的 Modal 代码（推荐）
- 删除 `OrganizationDashboard` 中 `OrganizationForm` 相关代码与导入。
- 将表单逻辑完全交由 `OrganizationTemporalPage` 维护，减少重复入口。
- 优点：结构清晰，消除误解；缺点：若未来需要 Modal，需要重新实现。

### 4.2 方案二：恢复 Modal 功能为“快速创建”入口
- 保留页面式表单为主流程，同时实现 Modal 作为轻量入口（例如创建草稿）。
- 优点：提供多种创建方式；缺点：需要额外维护与测试成本。

### 4.3 方案三：保留代码但加入显著注释与文档说明
- 在 `OrganizationDashboard` 中添加注释，指明 Modal 暂未启用。
- 在 `README` 或模块说明文档中补充“组织创建在独立页面进行”说明。
- 优点：实施成本最低；缺点：仍会增加长期维护负担。

---

## 5. 推荐决策

| 方案 | 评分 | 说明 |
|------|------|------|
| 方案一 | ⭐⭐⭐⭐⭐ | 与当前交互保持一致，最小化认知负担 |
| 方案二 | ⭐⭐ | 需要额外设计与开发资源，目前无明确需求 |
| 方案三 | ⭐⭐⭐ | 可做临时补救，但不建议长期保留冗余代码 |

> 建议采用**方案一（移除 Modal 代码）**，并在 06 号进展日志记录此调整，确保组织与职位模块在交互层面的描述保持一致。

---

## 6. 后续行动

- [x] 在 `OrganizationDashboard` 中移除未使用的 Modal 代码，并运行 `npm run test -- OrganizationDashboard` 验证。
- [x] 在 `docs/development-plans/06-integrated-teams-progress-log.md` 记录此澄清与修改。
- [ ] 若未来需要 Modal 快捷入口，再重新评估需求和设计。

---

## 7. 状态更新（2025-10-19）

- ✅ 已彻底删除 `frontend/src/features/organizations/components/OrganizationForm/` 目录及其测试，消除冗余 Modal 实现。
- ✅ 单一创建入口保留在 `/organizations/new`（`OrganizationTemporalPage` + `TemporalMasterDetailView`），E2E 场景 `organization-create` 仍通过。
- ✅ `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 已移除对 `OrganizationForm` 相关常量的引用，保持事实来源一致。

---

**创建人**：架构组 Claude Code 助手  
**下次复核**：完成代码清理后由前端负责人确认
