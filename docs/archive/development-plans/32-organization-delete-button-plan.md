# 32号文档：组织详情页“删除组织编码”按钮方案

## 背景与单一事实来源
- 命令服务已通过 `POST /api/v1/organization-units/{code}/events` 注册事件处理（见 `cmd/organization-command-service/internal/handlers/organization_routes.go`），并在 `CreateOrganizationEvent` 中对 `eventType === "DELETE_ORGANIZATION"` 执行整组织软删除，要求携带 `If-Match` 与 `effectiveDate`，同时拒绝存在未删子组织的请求（契约详见 `docs/api/openapi.yaml` 中 `DeleteOrganizationEventRequest` 条目）。
- 组织详情页主视图 `TemporalMasterDetailView`（`frontend/src/features/temporal/components/TemporalMasterDetailView.tsx`）依赖 `useTemporalMasterDetail` 管理版本列表；加载逻辑在 `temporalMasterDetailApi.mapOrganizationVersions` 中按 `effectiveDate` 降序排序，因此 `versions.at(-1)` 对应最早生效的记录。
- 表单操作区 `FormActions`（`frontend/src/features/temporal/components/inlineNewVersionForm/FormActions.tsx`）当前在编辑模式下为选中版本渲染“删除此记录”按钮，触发 `DEACTIVATE` 事件（调用链位于 `temporalMasterDetailMutations.ts` → `deactivateOrganizationVersion`），尚未提供整组织删除入口。
- `useTemporalMasterDetail` 将当前版本的 `recordId` 写入 `currentETag`（`frontend/src/features/temporal/components/hooks/useTemporalMasterDetail.ts` 第 73 行附近），可复用为整组织删除请求的 `If-Match` 值，确保跨层一致性。

## 目标与范围
- 在组织详情页为最早一条有效版本提供“删除组织编码”按钮，点击后调用后端 `DELETE_ORGANIZATION` 事件完成整组织软删除。
- 当最早版本被选中时，仅展示“删除组织编码”按钮，隐藏原有“删除此记录”入口；其他版本行为保持不变。
- 完善前端数据刷新与提示，保证删除成功后时间轴与表单状态同步更新。

## 风险与约束
- 整组织删除需校验 `If-Match` 与子组织数量；前端必须传入最新 `currentETag` 并处理 `409 HAS_CHILD_UNITS`/`412 PRECONDITION_FAILED` 等错误，避免误导用户。
- 最早版本可能已被软删或状态已置为 `DELETED`，需要在展示逻辑中排除这类记录，避免在已删除节点上重复渲染按钮。
- “删除组织编码”会使后续请求返回 404；需在操作成功后引导用户返回列表或刷新空态，防止残留编辑状态。

## 实施步骤
1. **封装整组织删除请求 Hook**  
   - 在 `frontend/src/shared/hooks/useOrganizationMutations.ts` 新增 `useDeleteOrganization`，复用 `unifiedRESTClient.request` 调用 `/organization-units/{code}/events`，发送 `{ eventType: "DELETE_ORGANIZATION", effectiveDate, changeReason }`，并在请求头写入 `If-Match`（复用 `formatIfMatchHeader` 对现有 `currentETag` 进行格式化，保持与停用/启用一致）。  
   - `effectiveDate` 直接沿用当前选中版本（即最早有效版本）的 `effectiveDate`，删除确认弹窗中展示该日期供用户确认，不新增日期输入控件。  
   - `changeReason` 使用固定文案（如“通过组织详情页删除组织编码”），确保审计记录可追溯又避免重复采集输入。  
   - 解析响应头中的 `etag`，并处理命令服务返回 `timeline: []` 的情况：不依赖时间线数据，改为触发版本重载与本地状态清空。
2. **扩展时态状态管理**  
   - 在 `useTemporalMasterDetail` 新增派生状态计算：对 `versions` 结果过滤 `status !== "DELETED"`，再按有效日期定位最早一条有效记录。  
   - 新增独立的整组织删除处理器（例如 `createHandleDeleteOrganization`），避免与现有 `createHandleDeleteVersion` 混用；调用新 Hook 后清空 `versions`、刷新缓存，并视结果决定是否触发 `onBack`。
3. **更新 UI 按钮渲染逻辑**  
   - 调整 `TemporalMasterDetailView` 向 `FormActions` 传入新的布尔值（如 `isEarliestVersionSelected`），使其在最早有效版本时仅渲染“删除组织编码”按钮；其余版本仍显示“删除此记录”。  
   - 扩展 `DeactivateConfirmModal` 支持两种模式：沿用既有版本作废流程，新模式下展示“删除组织编码”标题、提示删除后果、显示默认生效日期及子组织限制说明，不新增额外输入。
4. **交互与状态反馈**  
   - 删除成功后调用 `loadVersions()` 刷新；若刷新结果为空（表示组织已删除），调用 `onBack` 返回组织列表并通过 `notifySuccess` 输出“组织编码已删除”等提示，避免用户停留在已删除的详情页。  
   - 针对 `409`/`412`/`404` 等后端响应，复用 `notifyError` 输出命令服务原始文案（如“存在未删除的子组织”“资源已发生变更”），并保持页面状态不变。
5. **验证与清理**  
   - 更新或新增 Vitest 测试覆盖 `useDeleteOrganization` 的请求头与 payload 生成，以及 `FormActions` 在不同版本场景下的按钮可见性。  
   - 执行 `npm run lint` 与相关单测，确保计划内改动不影响既有流程。

## 验收标准
- [ ] 最早有效版本节点仅展示“删除组织编码”按钮，其余版本仍显示“删除此记录”。  
- [ ] 点击“删除组织编码”后正确发送 `DELETE_ORGANIZATION` 事件（包含 `If-Match`、沿用版本 `effectiveDate`、固定 `changeReason`），组织被软删并记录审计。  
- [ ] 当后端返回 `HAS_CHILD_UNITS`、`PRECONDITION_FAILED` 等错误时，页面展示对应错误文案且未误删数据。  
- [ ] 删除完成后时间轴/表单同步更新，若组织被清空则导航回列表并显示成功提示，页面不再保留旧数据。  
- [ ] 质量门禁（`npm run lint`、相关单测）全部通过，并完成计划归档流程。

## 一致性校验说明
- API 契约以 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 为唯一事实来源，新增 Hook 不得引入额外事实。  
- 删除按钮的展示条件、操作对象及错误处理均依据上述源代码路径，开发完成后需再次核对 `cmd/organization-command-service/internal/handlers/organization_events.go` 的约束，确保跨层行为一致。
