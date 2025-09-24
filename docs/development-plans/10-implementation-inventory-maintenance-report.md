# 10. 实现清单维护报告

**文档类型**: 评审报告
**创建日期**: 2025-09-21
**负责团队**: 实现清单守护代理 (IIG)
**优先级**: P0（包含过期临时实现）

---

## 执行概览

### 📊 完成状态
- ✅ IIG 最新扫描（2025-09-24）已与当前代码一致：OpenAPI 26、GraphQL 12、Go Handlers 26、Go Services 19、TS 导出 147。
- ✅ 前端表单与时态视图整改落地：统一校验链路、GraphQL codePath/namePath 已投入使用。
- ⚠️ 契约仍缺失 `/organization-units/temporal` 系列端点，`organizationPermissions.ts` 的子组织权限校验仍为临时禁用。

---

## 关键发现

### 🚨 紧急处理项（契约 / 权限风险）
1. **前端调用未入契约的 `/organization-units/temporal` 系列端点**
   - 证据：`frontend/src/features/organizations/components/OrganizationForm/index.tsx:168` 与 `:189` 在创建/更新时使用 `POST /organization-units/temporal`、`PUT /organization-units/{code}/temporal`。
   - 契约缺口：`docs/api/openapi.yaml` 中未声明相关路径，IIG 扫描与 `reports/implementation-inventory.json` 因此遗漏，违反“先契约后实现”。
   - 风险：CI 与监管工具无法检测到端点，部署环境会返回 404；同时破坏 `reports/implementation-inventory.json` 的准确性。
   - 行动：优先补充 OpenAPI 契约与命令服务路由，或在前端回退至 `/api/v1/organization-units/{code}/versions` 与历史事件端点。

2. **`organizationPermissions.ts` 子组织校验继续禁用**
   - 文件：`frontend/src/shared/utils/organizationPermissions.ts:37`。
   - 现状：`TODO-TEMPORARY` 截止 2025-09-20，仍注释 `childCount` 防删逻辑，权限计算缺乏真实数据约束。
   - 行动：重新接入 GraphQL `organizationHierarchy` 或 REST `/organization-units/{code}/refresh-hierarchy` 的子组织计数；若需延期必须更新截止日期与风险说明。

### ✅ 已完成整改与现状确认
1. **`temporalValidation.ts` 清理完成**
   - 现状：统一迁移至 `frontend/src/shared/utils/temporal-validation-adapter.ts`，脚本 `frontend/scripts/migrations/20250921-replace-temporal-validation.ts` 可复核。

2. **组织表单验证链路统一**
   - `frontend/src/features/organizations/components/OrganizationForm/index.tsx:118` 使用 `validateForm`（共享 Schema），`validation.ts` 仅负责数据标准化。
   - 同步接入 `/api/v1/organization-units/validate` 服务器校验，失败时阻断提交。

3. **API 类型临时导出回收**
   - `frontend/src/shared/types/api.ts` 保留纯类型定义且无 `TODO-TEMPORARY`，所有调用迁移至 `frontend/src/shared/api/*`。

4. **`useEnterpriseOrganizations` Hook 规范化**
   - `frontend/src/shared/hooks/useEnterpriseOrganizations.ts` 清除了自删标记与冗余封装，入口统一记录在 `frontend/HOOK_MIGRATION_REPORT.md`。

5. **TemporalMasterDetailView 功能补齐**
   - `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:330` 起，引入 GraphQL `organizationHierarchy`，成功加载 `codePath/namePath` 并回显。
   - 所有时态操作改用统一的 `unifiedRESTClient`/`unifiedGraphQLClient`，测试记录见 `docs/archive/development-plans/16-temporal-master-detail-view-remediation.md`。

### ⏱ 即将到期项
- `docs/reference/04-AUTH-ERROR-CODES-AND-FLOWS.md:56` 保留 `TODO-TEMPORARY`（419 状态码决策），截至 2025-09-30 需完成评审并同步实现。

---

## 架构合规成果

### ✅ 已修复问题
- camelCase 命名规范违规 3 项已全部修复。
- 架构验证器检查通过率 100%（109/109 文件）。
- REST/OpenAPI 契约与实现基本一致，唯独缺少 `/organization-units/temporal` 声明需补齐。

### 📈 统计更新
- REST 端点：26（新增 `/api/v1/organization-units/validate`、`/api/v1/organization-units/{code}/refresh-hierarchy`、`/batch-refresh-hierarchy`、`/api/v1/corehr/organizations` 已纳入契约，唯独 `/organization-units/temporal` 待补录）。
- GraphQL 查询：12 个主要字段，`organizationHierarchy` 已对接 `codePath/namePath`。
- Go 组件：Handlers 26、Services 19（匹配生成脚本输出）。
- 前端导出：147（较 2025-09-15 再压缩 15 项，聚合入口更集中）。

---

## 符合项目原则验证
- **单一事实来源**：除 `/organization-units/temporal` 缺少 OpenAPI 声明外，其余契约、实现清单与代码一致。
- **唯一性原则**：重复组件持续降低，无新增架构违规。
- **API 优先**：大部分端点遵循先契约后实现，需修复 `/organization-units/temporal` 违例。
- **CQRS 架构**：命令/查询隔离，端口配置标准化。

---

## 行动计划

### P0 — 立即执行
1. 补齐 `/organization-units/temporal` 系列契约与命令服务实现，或调整前端改用现有 `/api/v1/organization-units/{code}/versions` / `events` 路径；完成后更新 IIG 报告。
2. 恢复 `organizationPermissions.ts` 子组织校验逻辑，使用实时子组织计数；若判定延期，必须在文件中更新 `TODO-TEMPORARY` 截止与责任人，并补登记风险。
3. 将 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 与 `reports/implementation-inventory.json` 提升至 v1.9.x 数据（同步 147 TS 导出统计）。

### P1 — 下一个迭代
4. 扩展 `scripts/check-temporary-tags.sh`，在 CI 中强制校验截止日期与契约缺口，避免出现未登记的端点。
5. 对接 GraphQL `organizationSubtree` / REST `/batch-refresh-hierarchy` 的监控指标，纳入运营面板（参考 09 号计划）。

### P2 — 持续改进
6. 建立 IIG 守护例行巡检（日历化 + PR 模板勾选），并与 `docs/archive` 档案同步。
7. 将 `/organization-units/validate` 的调用结果纳入表单埋点，统计后端校验命中率，验证新流程效果。

---

## 评审建议
- 重点关注：
  1. `/organization-units/temporal` 契约缺口的补救方案（补契约 vs. 回退前端）。
  2. 权限校验恢复后的风险评估与测试覆盖度（含 Playwright/E2E）。
  3. `docs/reference/02` 与脚本统计版本漂移的解决时间表。
  4. `TODO-TEMPORARY` 自动治理范围是否扩展到文档（如 419 决策）。
- 决策需求：
  - [ ] 审批 P0 处理顺序与 Owner（建议：命令服务团队 + 前端权限组）。
  - [ ] 确认 `/organization-units/temporal` 端点的最终归属与上线窗口。
  - [ ] 确定 IIG 周期巡检节奏与通知渠道（建议：每周一 + Slack #cqrs-guardian）。

---

## 总结
- IIG 最新扫描覆盖率达 100%，但契约未覆盖的 `/organization-units/temporal` 仍是最大风险点。
- 表单与时态相关遗留问题已完成整改，统一验证链路落地，Temporal 详情页可展示完整路径。
- 当前仅剩 1 项代码层临时实现超期（权限校验），另有契约漂移与文档 TODO 需在下个迭代前关闭。
- 建议本迭代内完成 P0 任务，并将契约校验、TODO 治理纳入 CI，避免再次出现超期项。

---

**下一步**：等待相关团队认领任务并执行上述行动计划。
