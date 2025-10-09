# 33号文档：组织停用/重新启用一致性修复计划

## 背景与单一事实来源
- 停用/启用接口由 `cmd/organization-command-service/internal/handlers/organization_update.go` 中的 `changeOrganizationStatusWithTimeline` 驱动，进一步调用 `TemporalTimelineManager.changeOrganizationStatus`（见 `internal/repository/temporal_timeline_status.go`）。该逻辑始终插入一条新的时态版本，并依赖唯一索引 `tenant_id + code + effective_date` 保障时间点唯一。
- 当前 GraphQL/REST 返回的状态字段取自命令服务写入的最新版本；当停用与启用在同一天连续执行时，后一操作会因为生效日期冲突抛出 `TEMPORAL_POINT_CONFLICT`，在 `/tmp/run-dev.log` 中可复现。
- 前端 `SuspendActivateButtons`（`frontend/src/features/temporal/components/SuspendActivateButtons.tsx`）在初次渲染时用 `useMemo` 固定 `effectiveDate = today`，按钮点击时无法改动；请求底层由 `useSuspendOrganization` / `useActivateOrganization`（`frontend/src/shared/hooks/useOrganizationMutations.ts`）发送，仅传入当天日期，无用户确认与理由输入。
- 当组织被 `DELETE_ORGANIZATION` 软删后（`status === 'DELETED'`），按钮仍判定为“重新启用”，但命令服务已不再返回当前版本，导致客户端看到 404/500 错误。

## 问题描述
1. **生效日期冲突**：命令服务默认插入新版本而不是更新现有记录，若停用后立刻启用且日期相同，则触发唯一索引冲突，用户无法完成“立即恢复”。
2. **UI 缺少输入与确认**：前端按钮固定使用今日日期、自动填充文案“自动生成停用/重新启用”，没有确认弹窗或操作原因输入；用户无法知晓请求失败原因，也难以选择未来生效时间。
3. **状态异常场景未处理**：当组织已被软删或不存在当前版本时，按钮仍展示“重新启用”，实际请求返回 404；缺少空态和提示处理。
4. **子组织停用策略空缺**：`CascadeUpdateService.processStatusUpdate` 仅打印日志，不会级联停用子组织，业务期望尚未满足。

## 目标与范围
- 支持“同日停用后立即恢复”场景，确保命令服务与前端默认操作相容，不再出现 `TEMPORAL_POINT_CONFLICT`。
- 为停用/启用操作提供生效日期、操作原因输入及确认提示，提升用户可控性。
- 在前端隐藏对 `DELETED` 或缺少当前版本记录的启用按钮，提供明确的导航/提示。
- 梳理父级停用对子组织的策略，至少在 UI 或后端给出明确反馈，避免状态不一致。

## 风险评估
- 修改命令服务的时态写入逻辑需谨慎，直接改写当日版本时必须同步写入审计记录、保持版本唯一性，确保不会破坏 `organization_units` 的单一事实来源。
- 增加前端输入（生效日期、理由）会改变交互流程，需要文案、设计确认及 i18n 处理。
- 对 `DELETED` 状态隐藏按钮会影响现有测试/回退流程，需更新 e2e 场景。
- 若引入子组织级联停用，需评估性能与事务边界，防止大规模组织树更新导致锁表。

## 实施步骤建议
1. **命令服务同日恢复策略（优先级 P0）**
   - 当 `changeOrganizationStatusWithTimeline` 针对同一组织、同一天再次收到停用/启用请求时，直接更新当日最新版本的 `status`、`change_reason` 与审计信息，而非新增记录或顺延日期；保持 `effective_date` 不变，同时记录“撤销停用”类审计事件，确保误操作可当日撤回且时间轴连续。
2. **前端交互增强（优先级 P0）**
   - 在 `SuspendActivateButtons` 引入确认对话框，允许选择生效日期、填写操作原因；默认值为当天但用户可调整。提交时传入 `operationReason` 与 `effectiveDate`。  
   - `useSuspendOrganization` / `useActivateOrganization` 增加错误码映射（如 `TEMPORAL_POINT_CONFLICT`、`HAS_CHILD_UNITS`），在通知中反馈具体原因并建议下一步操作。
3. **异常状态处理（优先级 P0）**
   - 在 `TemporalMasterDetailView` 的状态派生中识别 `status === 'DELETED'` 或缺少当前版本的场景，隐藏启用按钮并展示“组织已删除”提示及返回入口。
4. **子组织策略（优先级 P1）**
   - 明确业务约束：停用组织仅会阻止在停用期间选择该组织，也不会影响下级组织状态。需在前端文案中提示用户，避免误解为整棵树停用；如子组织需要额外处理，应在模块中给出操作引导。
5. **验证与文档（优先级 P0）**
   - 新增单元测试 / 合同测试 / e2e 场景，覆盖同日停启、未来日期生效、软删组织、子组织存在等关键路径。
   - 更新 `docs/api/openapi.yaml` 与前端操作指南，说明停用/启用所需输入及典型错误码。

## 验收标准
- [ ] 停用后立即启用（默认相同生效日期）会直接更新当日版本，操作成功且不再出现 `TEMPORAL_POINT_CONFLICT`。
- [ ] 前端在停用/启用时提供生效日期与操作原因输入，并在提交前显示确认提醒。
- [ ] 组织处于 `DELETED` 或无当前版本时不再显示启用按钮，页面提供明确导航或状态说明。
- [ ] 若启用操作因子组织存在被拒绝，前端能读取错误码（如 `HAS_CHILD_UNITS`）并展示对应提示。
- [ ] Lint、单测、关键 e2e 用例（停用/启用流程）全部通过，并在文档中更新操作指南。

## 一致性校验说明
- API 与错误码以 `docs/api/openapi.yaml`、`cmd/organization-command-service/internal/handlers/organization_update.go` 为唯一事实来源；任何交互改动需与契约同步。
- 生效日期与时态逻辑以 `TemporalTimelineManager` 实现为准；修改后需确保版本时间轴与审计记录保持一致。

## 现状记录
- 2025-10-09：实现同日停用/启用的冲突容错逻辑，并更新前端交互弹窗（新增生效日期与原因输入）。
- 2025-10-09：补充 `SuspendActivateButtons` 单元测试并通过 `npm run test` 全量校验。
