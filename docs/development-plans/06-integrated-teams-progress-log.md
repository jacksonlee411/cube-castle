# 06号文档：集成团队协作进展日志（92号计划实时记录）

> **更新时间**：2025-10-20 19:45
> **负责人**：前端团队 · 职位域
> **关联计划**：92号《职位管理二级导航实施方案》 v2.2

---

## 1. 当前进展快照

- ✅ **Phase 0 POC**：`AppShell` 已替换为 312px 灰底 `SidePanel`，`NavigationItem` 基于 Canvas `Expandable` 实现折叠导航，`AuthContext` 暴露 `hasPermission`/`userPermissions` 并配套单测（`useAuth.test.tsx`）。
- ✅ **Phase 1 导航架构**：Sidebar 与导航配置完成权限过滤、动画与路由联动；`App.tsx` 职位路由迁移为 `<Route path="/positions" element={<Outlet />}>` 嵌套路由。
- ✅ **Phase 2 页面框架**：`frontend/src/features/job-catalog/` 新增职类/职种/职务/职级列表与详情页（含新增表单），共用 `CatalogTable`、`CatalogFilters`、`CatalogForm`、`CatalogVersionForm` 与 `StatusBadge` 组件。
- ✅ **Phase 3 查询与命令集成**：`useJobCatalog.ts` 扩展 GraphQL 字段（含 `recordId`、描述、时态信息），`useJobCatalogMutations.ts` 已覆盖四层新增、更新与版本创建；详情页复用 `CatalogVersionForm` 支持“编辑当前版本”，所有操作均走统一缓存失效逻辑。
- ✅ **93号职位详情多页签验收**：2025-10-19 完成六个页签布局、时间轴侧栏与审计页签接入，详见 `docs/archive/development-plans/93-position-detail-tabbed-experience-acceptance.md`，对应实现已迁移至归档方案文档。
- ✅ **Job Catalog 数据补充**：通过迁移 `047_seed_additional_job_catalog_data.sql` 新增“数据智能”域 10 组职类、10 条职种、10 条职务与 10 条职级样板数据，确保每类实体均具备真实链路（如 DATA → DATA-CORE → DATA-CORE-ENG → IC5-DATA）。
- 📗 **测试与校验**：执行 `npm --prefix frontend run test -- --run src/features/job-catalog/__tests__/jobCatalogPages.test.tsx` 与 `npm --prefix frontend run typecheck` 均通过；新增职种/职务/职级详情页更新链路断言，并补充 `src/shared/hooks/__tests__/useJobCatalogMutations.test.tsx` 覆盖 REST 更新入参与缓存失效行为；新增 `src/features/job-catalog/__tests__/jobCatalogPermissions.test.tsx` 验证 `job-catalog:create/update` 权限屏蔽逻辑；Playwright 规格 `frontend/tests/e2e/job-catalog-secondary-navigation.spec.ts` 已接入真实命令/查询服务，验证管理员更新成功、普通用户 403 拒绝及 If-Match 412 并发保护；`npm --prefix frontend run test:contract` 通过（同步移除 schema 重复字段），确认 GraphQL 契约保持一致。2025-10-20 11:22~11:23 再次执行 `NavigationItem.test.tsx`、`Sidebar.test.tsx`、`jobCatalogPermissions.test.tsx`、`jobCatalogPages.test.tsx`、`useJobCatalogMutations.test.tsx`（全部通过），并运行 `go test ./cmd/organization-command-service/...` 确认审计链路。
- ✅ **文档对齐**：`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 新增 Job Catalog 模块速查段落，并发布 `docs/reference/job-catalog-navigation-guide.md`（导航说明、截图索引、权限映射），92号文档 D1/D2 验收项已勾选。
- 🚧 **计划 97 Phase 0**：已在 `fix/typescript-errors-remediation` 分支记录 `npm run build`（输出存档于 `docs/development-plans/97-build-errors-baseline.txt`）与 `npm run test -- --run`（输出存档于 `docs/development-plans/97-test-baseline.txt`）的 TypeScript 错误基线，将按阶段推进枚举修复与 Canvas Kit 迁移。
- 🚧 **计划 97 Phase 2**：完成 Job Catalog/Position 表单 Canvas Kit 迁移，`CatalogForm` 改用 Modal 模型事件、`Select` 全面替换为原生封装，职类/职种/职务/职级表单与 Position 表单统一使用 FormField 输入封装；`npm run test -- --run src/features/job-catalog` 与 `npm run test -- --run src/features/positions` 均通过，当前 `npm run build` 剩余 14 项（集中在 Temporal/GraphQL 枚举与类型）。
- 🚧 **计划 97 Phase 2**：完成 Job Catalog/Position 表单 Canvas Kit 迁移，`CatalogForm` 改用 Modal 模型事件、`Select` 全面替换为原生封装，职类/职种/职务/职级表单与 Position 表单统一使用 FormField 输入封装；`npm run test -- --run src/features/job-catalog` 与 `npm run test -- --run src/features/positions` 均通过，`npm run build` 错误降至 14 项（集中在 Temporal/GraphQL 枚举与类型）。
- 🚧 **计划 97 Phase 3**：校准 Temporal `lifecycleStatus` 映射并以 `normalizeLifecycleStatus` 约束，修复职位 GraphQL 变量 `JsonValue` 建模与 `logger.mutation` 日志参数；`npm run build` 仅剩 Storybook 类型 2 项（按 Phase 4 处理）。
- 🚧 **计划 97 Phase 3**：校准 Temporal `lifecycleStatus` 映射并以 `normalizeLifecycleStatus` 约束，修复职位 GraphQL 变量 `JsonValue` 建模与 `logger.mutation` 日志参数；`npm run build` 仅剩 Storybook 类型 2 项（已在 Phase 4 处理）。
- ✅ **计划 97 Phase 4**：通过 `tsconfig.app.json` 排除 Storybook 目录完成收尾，`npm run build`、`npm run test -- --run src/features/job-catalog`、`npm run test -- --run src/features/positions` 均通过；文档及进度表已更新。
- ✅ **Position 多页签 Playwright 验证**：2025-10-20 新增 `frontend/tests/e2e/position-tabs.spec.ts`，覆盖六个页签切换与审计空态提示；`position-crud-live.spec.ts` 补充 Mock 守护断言。
- ✅ **Mock 模式策略决议**：确认长期保留开关（选项B），发布前由前端团队验证 `VITE_POSITIONS_MOCK_MODE=false` 并在 QA checklist 记录；FE-1181/1182/1183 将分别跟踪 Playwright 覆盖、PositionForm 抽象与组件层次化收尾。
- ✅ **Playwright 运行指引更新**：`frontend/tests/e2e/README.md` 新增真实链路与 Mock 守护执行步骤，统一 PW_REQUIRE_LIVE_BACKEND / PW_REQUIRE_MOCK_CHECK 环境变量说明，便于 QA 快速切换验证模式。
- ✅ **FE-1182 代码落地**：抽离 `usePositionCatalogOptions` 到共享 Hook，统一 PositionForm payload/validation，并新增 Storybook 错误态示例与 README（待评审合并）。
- ✅ **FE-1183 目录整理推进**：components/ 重组为 dashboard/details/list/layout/transfer 子目录，新增聚合导出与 README，移除遗留 `PositionVersionList.tsx`。
- ✅ **DS-147 设计规范 v0.1**：新增 `docs/reference/positions-tabbed-experience-guide.md`，定义多页签布局、Mock 提示与响应式规则，并在 `frontend/artifacts/layout/README.md` 记录截图要求。
- ✅ **101~104 计划归档**：职位 Playwright、PositionForm、组件目录与 DS-147 设计计划均已生成正式文档并标记完成。

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
| ✅ | 前后端联调：REST 更新接口 + 权限校验 | 前端/后端联调小组 | 2025-10-22 | 命令服务补齐 Job Catalog `PUT` 接口＋PBAC `job-catalog:write` 映射，支持 If-Match 412 并运行 `go test ./cmd/organization-command-service/...` 通过，前端 hooks 联动已可直连容器服务 |
| ✅ | Playwright 脚本覆盖二级导航与职类 CRUD 正向路径 | QA 团队 | 2025-10-24 | `frontend/tests/e2e/job-catalog-secondary-navigation.spec.ts` 接入真实后端，覆盖管理员成功编辑、普通用户 403 及 If-Match 412 场景 |
| P1 | 文案国际化：将导航配置及表单提示接入现有 i18n 方案 | 前端国际化负责人 | 2025-10-24 | 对齐 92 号文档 2.3.1 国际化前置说明 |
| ✅ | 设计评审：确认 Job Catalog 列表/详情在新导航下的视觉稿（含 312px 侧栏） | 设计团队 | 2025-10-20 | **评审已完成**。基线截图 `artifacts/layout/{positions-list,job-family-groups-list,job-family-group-detail}.png` 用于留白对比；根据结论已完成 CardContainer 抽取、列表筛选/表格卡片化与详情页分层，并同步更新 92 号 Phase 4 F4 子项与 96 号文档第 8 节。 |
| P0 | 101 · FE-1181 Position Playwright hardening | QA 团队（陈慧） | 2025-10-28 | ✅ 计划已归档，Mock 守护用例与 README 指南已更新。 |
| P0 | 102 · FE-1182 PositionForm 数据层抽象深化 | 前端组件组（李程） | 2025-10-31 | ✅ 计划已归档，共享 Hook 与 Storybook/README 完成。 |
| P1 | 103 · FE-1183 Position 组件目录收尾 | 职位域前端组（赵琳） | 2025-11-08 | ✅ 计划已归档，目录分层与聚合导出就绪。 |
| P1 | 104 · DS-147 Positions Tabbed Experience | 设计团队（刘冉） | 2025-10-29 | ✅ 计划已归档，规范发布并约定截图路径。 |

> **优先级执行顺序**：P0 任务需优先于 P1；FE-1183 依赖 FE-1182 完成后再开展，DS-147 可与 FE-1182 并行但需在 FE-1183 启动前冻结规范。

---

## 4. 风险与依赖

- 🔶 **后端写接口联调**：目前新增/版本创建调用依赖 REST `/api/v1/job-*` 接口，需确认命令服务环境数据准备与 idempotency header 要求（跟踪 80 号计划接口列表）。
- 🔶 **缓存刷新策略**：`useJobCatalogMutations` 仅做 QueryKey 级别失效；暂缓深度优化，保持监控并在新增分页/搜索参数时补充精细化逻辑。
- 🔸 **懒加载优化**：性能指标当前满足 92 号计划 T3 要求，作为监控项保留，若首屏或路由时延回退再恢复实施。
- 🔶 **浏览器上下文依赖**：职级详情页需从列表携带 `roleCode`（`navigate(..., { state })`），直接访问时提示缺少上下文；评估是否改为 URL 查询参数或在页面内重新推导父级信息。
- 🔷 **权限模拟**：`useScopes` 根据本地 token scopes 推导权限；若环境未提供 `job-catalog:*` scope，需在 QA 文档中明确设置方式。

---

## 5. 已知开放事项

- `useJobCatalogMutations` 已补齐更新 Hook，仍缺删除能力；对应 Phase 3 “命令测试” 条目继续跟踪。
- 页面尚未接入国际化与空态插画，需配合设计在 Phase 4 统一收敛。
- 性能优化（懒加载/缓存策略）暂缓执行，纳入监控清单，性能指标回退时优先恢复。
- 后续若扩展 Playwright 场景，请同步调优真实/Mock job 配置，并继续在 README 中登记运行方式。

---

**下一次更新触发条件**：
- Hook 更新链路合并 & Job Catalog 页面单测补齐；或
- Playwright 场景落地并通过 CI。
