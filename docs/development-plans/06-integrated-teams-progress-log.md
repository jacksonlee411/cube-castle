# 06号文档：集成团队协作进展日志（92号计划实时记录）

> **更新时间**：2025-10-19 22:50
> **负责人**：前端团队 · 职位域
> **关联计划**：92号《职位管理二级导航实施方案》 v2.2

---

## 1. 当前进展快照

- ✅ **Phase 0 POC**：`AppShell` 已替换为 312px 灰底 `SidePanel`，`NavigationItem` 基于 Canvas `Expandable` 实现折叠导航，`AuthContext` 暴露 `hasPermission`/`userPermissions` 并配套单测（`useAuth.test.tsx`）。
- ✅ **Phase 1 导航架构**：Sidebar 与导航配置完成权限过滤、动画与路由联动；`App.tsx` 职位路由迁移为 `<Route path="/positions" element={<Outlet />}>` 嵌套路由。
- ✅ **Phase 2 页面框架**：`frontend/src/features/job-catalog/` 新增职类/职种/职务/职级列表与详情页（含新增表单），共用 `CatalogTable`、`CatalogFilters`、`CatalogForm`、`CatalogVersionForm` 与 `StatusBadge` 组件。
- ✅ **Phase 3 查询与命令集成**：`useJobCatalog.ts` 扩展 GraphQL 字段（含 `recordId`、描述、时态信息），`useJobCatalogMutations.ts` 已覆盖四层新增、更新与版本创建；详情页复用 `CatalogVersionForm` 支持“编辑当前版本”，所有操作均走统一缓存失效逻辑。
- 📗 **测试与校验**：执行 `npm --prefix frontend run test -- --run src/features/job-catalog/__tests__/jobCatalogPages.test.tsx` 与 `npm --prefix frontend run typecheck` 均通过；新增职种/职务/职级详情页更新链路断言，并补充 `src/shared/hooks/__tests__/useJobCatalogMutations.test.tsx` 覆盖 REST 更新入参与缓存失效行为；新增 `src/features/job-catalog/__tests__/jobCatalogPermissions.test.tsx` 验证 `job-catalog:create/update` 权限屏蔽逻辑；`npm --prefix frontend run test:contract` 通过（同步移除 schema 重复字段），确认 GraphQL 契约保持一致。

---

## 2. 页面验证步骤（手动 / 自动混合）

1. **启动环境**：`make docker-up` → `make run-dev` → `make frontend-dev`，确保 `http://localhost:9090/health`、`http://localhost:8090/health` 返回 200。
2. **身份准备**：`make jwt-dev-mint`，前端调试环境载入 `.cache/dev.jwt`，默认具备 `position:read` + `job-catalog:read/write` 权限。
3. **导航验证**：
   - 访问 `http://localhost:5173`，Sidebar 默认展开“职位管理”，确认子项（职位列表/职类/职种/职务/职级）显示。
   - 切换浏览器 localStorage 中 scope（仅保留 `position:read`），刷新后职类/职种/职务/职级菜单自动隐藏。
4. **职类管理**：
   - 路由 `#/positions/catalog/family-groups`，搜索框输入关键字过滤。
   - 点击“新增职类”弹出表单，校验编码格式限制（4-6 位大写）。
   - 提交后观察列表刷新及成功提示（需真实 API 配合）。
5. **职种/职务/职级管理**：依次访问 `#/positions/catalog/families|roles|levels`，通过顶部下拉筛选父级层级，验证下钻逻辑与列表空态文案。
6. **详情页面**：从各列表点击任意记录，确认详情展示 `recordId`、生效区间、描述与“新增版本”按钮；权限不足（移除 `job-catalog:update`）后按钮隐藏。
7. **版本创建**：在具备写权限时，点击“新增版本”→输入名称/生效日期→提交，查看 Mutation 日志（终端）与列表重新加载结果。
8. **回归脚本**：`npm --prefix frontend run test -- --run src/features/job-catalog`（待补充时使用，现阶段关注已存在的布局与导航测试）。

---

## 3. 下一步任务

| 优先级 | 项目 | 负责人 | 截止 | 说明 |
|--------|------|--------|------|------|
| ✅ | 补充 Job Catalog 页面 Vitest（包含更新链路断言） | 前端测试代表 | 2025-10-19 | 新增职类/职种/职务/职级更新断言，确保 CI 回归覆盖 |
| P0 | 前后端联调：REST 更新接口 + 权限校验 | 前端/后端联调小组 | 2025-10-22 | REST 更新 Hook 单测已覆盖入参与缓存；待真实命令服务验证 2xx/412 契约与权限策略 |
| P1 | Playwright 脚本覆盖二级导航与职类 CRUD 正向路径 | QA 团队 | 2025-10-24 | 与 92 号 Phase 4 测试要求对齐 |
| P1 | 文案国际化：将导航配置及表单提示接入现有 i18n 方案 | 前端国际化负责人 | 2025-10-24 | 对齐 92 号文档 2.3.1 国际化前置说明 |
| P1 | 设计评审：确认 Job Catalog 列表/详情在新导航下的视觉稿（含 312px 左侧栏对齐） | 设计团队 | 2025-10-23 | 记录结果并更新 92 号文档 “视觉对齐” 条目 |

---

## 4. 风险与依赖

- 🔶 **后端写接口联调**：目前新增/版本创建调用依赖 REST `/api/v1/job-*` 接口，需确认命令服务环境数据准备与 idempotency header 要求（跟踪 80 号计划接口列表）。
- 🔶 **缓存刷新策略**：`useJobCatalogMutations` 仅做 QueryKey 级别失效；后续若新增分页/搜索参数，需要补充精细化无效化逻辑。
- 🔶 **浏览器上下文依赖**：职级详情页需从列表携带 `roleCode`（`navigate(..., { state })`），直接访问时提示缺少上下文；评估是否改为 URL 查询参数或在页面内重新推导父级信息。
- 🔷 **权限模拟**：`useScopes` 根据本地 token scopes 推导权限；若环境未提供 `job-catalog:*` scope，需在 QA 文档中明确设置方式。

---

## 5. 已知开放事项

- `useJobCatalogMutations` 已补齐更新 Hook，仍缺删除能力；对应 Phase 3 “命令测试” 条目继续跟踪。
- 页面尚未接入国际化与空态插画，需配合设计在 Phase 4 统一收敛。
- Playwright 场景待补充；当前仅依赖 Vitest + 手动验证。

---

**下一次更新触发条件**：
- Hook 更新链路合并 & Job Catalog 页面单测补齐；或
- Playwright 场景落地并通过 CI。
