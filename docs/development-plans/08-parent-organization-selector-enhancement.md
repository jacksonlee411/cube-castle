# 开发计划 08：上级组织选择器增强方案（测试同步版）

## 变更概览
- 上级组织选择器已回归使用 Canvas Kit `Combobox` + `FormField`，并保持缓存、循环依赖校验、权限门控逻辑。
- 新增 `useComboboxModel` 的受控实现，支持预填、清除、键盘选择等交互，所有 GraphQL 调用继续复用 `UnifiedGraphQLClient`。
- Vitest 用例更新，重点验证候选加载、循环检测与 PBAC 禁用场景；测试桩补齐 `Combobox.Menu.Popper`、`setSelectedIds`、JWKS mock，避免真实 OAuth 依赖。

## 已执行校验
- `npm --prefix frontend run test -- --run --pool=threads ParentOrganizationSelector`
- `npm --prefix frontend run lint`
- ESLint 架构门禁、TypeScript 编译、GraphQL 字段命名校验全部通过。

## 建议的测试范围
1. **功能回归**（Playwright / 手工）
   - `/organizations/new`：搜索并选择上级组织，确认选项列表、回填、循环校验提示正常。
   - 权限不足账号：确认组件禁用且保持“您没有权限查看组织列表”提示。
2. **样式确认**
   - 比对 Canvas Kit 规范，检查组合框对齐、颜色、间距，必要时与设计团队同步。
3. **多入口一致性**
   - `OrganizationDetailForm`、`TemporalEditForm`（若已接入）入口下的表现需一致，避免体验割裂。
4. **缓存与刷新**
   - 在同一生效日期下多次打开组件，验证缓存命中；更换生效日期后需重新触发查询。

## 已知注意事项
- 若后续新增样式微调，请保持 `data-testid` 不变，以免破坏现有用例。
- Playwright 脚本需等待 GraphQL 返回后再断言菜单项，可复用现有 `waitForResponse('/graphql')` 助手。

## 后续动作建议
1. 测试团队按上方范围执行回归，输出报告。
2. 前端根据测试反馈协同调整视觉或交互细节。
3. 如需补充端到端脚本，建议扩展 `tests/e2e/organizations` 套件并上传截图/录像。
