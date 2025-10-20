# Job Catalog 二级导航使用指南

**版本**: v1.0（2025-10-20）  
**适用对象**: 职位管理模块业务操作与支持团队  
**关联文档**: [92号职位管理二级导航实施方案](../development-plans/92-position-secondary-navigation-implementation.md)、[06号设计评审任务确认报告](../development-plans/06-design-review-task-assessment.md)

> 依据 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 定义的唯一事实来源编写；如契约更新，请先同步上述文件后再维护本指南。

---

## 1. 导航结构与视觉基线

- 职位管理入口位于侧栏「职位管理」，采用 Canvas Kit `SidePanel` + `Expandable` 组合。默认展开后展示以下二级菜单：
  1. 职位列表 (`#/positions`)
  2. 职类 (`#/positions/catalog/family-groups`)
  3. 职种 (`#/positions/catalog/families`)
  4. 职务 (`#/positions/catalog/roles`)
  5. 职级 (`#/positions/catalog/levels`)
- 布局与卡片分层应与基线截图保持一致，可在 `frontend/artifacts/layout/` 查看：
  - `positions-list.png`
  - `job-family-groups-list.png`
  - `job-family-group-detail.png`
- 若检测到布局偏差，请参照 96 号文档《职位管理 Job Catalog 布局校准前置分析》执行校准。

## 2. 操作流程

1. **进入模块**：以具备 `job-catalog:read` scope 的身份访问任一菜单，系统会从 GraphQL 查询 `jobFamilyGroups/jobFamilies/jobRoles/jobLevels` 获取列表数据。
2. **筛选与列表**：顶部筛选区复用 `CatalogFilters` 组件，可按父级层级、状态、时间范围过滤；表格由 `CatalogTable` 渲染，默认按编码排序。
3. **查看详情**：在列表中点击任意行进入详情页（独立路由），详情页展示 `recordId`、生效区间、描述等字段，并提供「新增版本」等操作按钮。
4. **新增/更新**：
   - 点击「新增职类/职种/职务/职级」按钮调起 `CatalogForm`，REST 命令接口分别为 `/api/v1/job-family-groups|job-families|job-roles|job-levels`。
   - 在详情页触发「新增版本」调用 `/api/v1/job-*/{code}/versions`；表单由 `CatalogVersionForm` 提供校验与提交。
5. **回归验证**：提交成功后列表会自动失效缓存并刷新；如需进一步验证，可使用 GraphQL 查询对比最新数据或检查命令服务审计日志。

## 3. 权限配置

- Scope 对应关系（详见 `docs/api/openapi.yaml` → `components.securitySchemes.PBAC.scopes`）：
  - `job-catalog:read` — 允许访问二级菜单、查看列表与详情。
  - `job-catalog:create` — 允许创建新的职类/职种/职务/职级记录。
  - `job-catalog:update` — 允许编辑现有记录、创建版本与执行时态管理操作。
- 若用户缺失 `job-catalog:read`，前端不会渲染二级菜单；缺少写权限时，「新增」「新增版本」按钮将隐藏或禁用。后端命令服务同时会拒绝未经授权的操作。
- 测试或演示环境可通过 `make jwt-dev-mint ROLES=ADMIN,JOB_CATALOG_EDITOR` 生成包含上述 scopes 的令牌（详见开发者快速参考“JWT认证管理”）。

## 4. 验收与回归脚本

- 单元测试：
  - `npm --prefix frontend run test -- --run src/features/job-catalog/__tests__/jobCatalogPages.test.tsx`
  - `npm --prefix frontend run test -- --run src/features/job-catalog/__tests__/jobCatalogPermissions.test.tsx`
- E2E 测试：
  - `PW_CAPTURE_LAYOUT=true PW_JWT=<token> PW_TENANT_ID=<tenant> npm --prefix frontend run test:e2e -- tests/e2e/job-catalog-secondary-navigation.spec.ts`
- 执行前确保 Docker 环境中的命令、查询服务已通过 `make run-dev` 启动，并完成最新数据库迁移。

## 5. 关联资源

- 导航配置：`frontend/src/layout/navigationConfig.ts`
- 二级菜单实现：`frontend/src/layout/NavigationItem.tsx`
- Job Catalog 共用组件：`frontend/src/features/job-catalog/shared/`
- Hook：`frontend/src/shared/hooks/useJobCatalog.ts`、`frontend/src/shared/hooks/useJobCatalogMutations.ts`
- 设计评审日志：`docs/development-plans/06-integrated-teams-progress-log.md`

---

> 如需扩展多语言、删除能力或缓存策略，请参考 92 号文档“后续扩展”章节并在更新后同步维护本指南。
