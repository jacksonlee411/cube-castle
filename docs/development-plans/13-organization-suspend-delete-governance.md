# 13. 组织停用与删除一致性治理计划

**文档类型**: 合规整改 / 能力增强  
**创建日期**: 2025-09-26  
**优先级**: P0（违反唯一事实来源 + 权限防线风险）  
**负责团队**: 命令服务团队（Owner） / 查询服务团队（Co-owner） / 前端组织域团队（协作）  
**关联文档**: `CLAUDE.md`、`AGENTS.md`、`docs/development-plans/10-implementation-inventory-maintenance-report.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/api/openapi.yaml`

---

## 1. 背景与触发
- IIG 报告指出 `/organization-units/temporal` 契约缺失与 `organizationPermissions.ts` 子组织校验被临时禁用（`docs/development-plans/10-implementation-inventory-maintenance-report.md`）。
- 业务提出“停用仅影响本级、停用期间不可选、子级保持有效”及“有下级即禁止删除”新约束，若未治理将导致跨层事实不一致与误删风险。
- 现有命令服务仅实现单点停用逻辑，但缺乏显式回归验证；删除场景缺少系统化的子组织拦截与前端提示。

## 2. 问题定义
1. **停用语义缺口**：需要确保 `POST /api/v1/organization-units/{code}/suspend`（OpenAPI v4.2.1）仅写入目标组织的 `INACTIVE` 版本，并提供回归验证证明未级联影响子组织。
2. **组织选择器缺陷**：停用后组织仍可被选择；缺少基于 GraphQL 的状态过滤能力。
3. **删除防线缺失**：命令层未强制阻断“存在子组织（含停用）”的删除操作，前端也因 `childCount` 校验被注释而放行。

## 3. 目标与范围
- 保持唯一事实来源：契约 → 命令实现 → 查询/前端一致对齐。
- 停用操作仅影响本级组织状态，子级状态原样保留。
- GraphQL/前端在停用期间屏蔽被停用组织在选择器中的出现。
- 删除前强制检测所有下级组织（ACTIVE/INACTIVE 均算），存在则返回 `HAS_CHILD_UNITS` 错误并阻断。

## 4. 方案设计

### 4.1 契约与事实来源治理
- **停用**：继续复用 `POST /api/v1/organization-units/{code}/suspend` 和 `POST /api/v1/organization-units/{code}/activate`，在 `docs/api/openapi.yaml` 中补充“停用不影响下级组织”的行为描述与成功示例，明确请求体需包含 `effectiveDate`、`operationReason`、`If-Match`（ETag）。
- **删除**：不新增独立 `DELETE` 端点，保持契约收敛在现有命令体系。扩展 `POST /api/v1/organization-units/{code}/events`：
  - 新增 `operation` 枚举值 `DELETE_ORGANIZATION`（或沿用 `DEACTIVATE` 并在文档说明“对当前版本执行业务删除”）；
  - 请求体包含 `effectiveDate`、`operationReason`、`If-Match`；
  - 成功时返回时间线重算结果，并在响应示例中强调 `isDeleted=true`；
  - 失败示例增加 `HAS_CHILD_UNITS`；
  - 权限 scope：`org:delete`；
  - 契约更新后运行 `node scripts/generate-implementation-inventory.js` 并刷新 IIG。

### 4.2 命令服务（Go）
- **停用校验**：
  - 在 `cmd/organization-command-service/internal/services/organization_temporal_service.go` 增加单元测试，调用 `timelineManager.SuspendOrganization` 后读取子组织状态，确认未被更改。
  - 新增集成测试：构造父子层级，停用父级后断言 `SELECT status FROM organization_units WHERE code = child` 仍为 `ACTIVE`。
  - 在 `SuspendOrganization` handler 增加“停用后重新刷新父级层级缓存”调用，避免 selector 缓存脏数据。
- **删除防线**：
  - 在 `CreateOrganizationEvent` 流程内新增 `DELETE_ORGANIZATION` 分支：
    1. 在 service 层先执行 `SELECT 1 FROM organization_units WHERE tenant_id=$tenant AND parent_code=$code AND status <> 'DELETED' AND deleted_at IS NULL LIMIT 1`，若存在则返回 `HAS_CHILD_UNITS`（409）。
    2. 无子组织时调用封装方法 `SoftDeleteOrganization(ctx, tenantID, code, effectiveDate, operationReason)`，内部复用 `timelineManager.DeleteVersion`/写入 `status='DELETED'` 并触发时间线重算，保持单事务。
    3. 通过已有 `LogOrganizationDelete` 写入审计；必要时扩展以支持 `DELETE_ORGANIZATION` 事件类型。
  - 增补事务级回滚保障、命名遵循 camelCase；新增 Go 单测/集成测试覆盖“存在子组织 → 409(HAS_CHILD_UNITS)”与“无子组织 → 成功软删 + 审计记录”。

### 4.3 查询服务（GraphQL）
- `cmd/organization-query-service/main.go`：
  - 维持当前默认行为（返回 ACTIVE + INACTIVE，排除 `isDeleted=true`），新增 `onlyActive` 过滤参数供组织选择器使用，避免破坏现有调用；
  - 新增可选字段 `includeDisabledAncestors`（默认 false），当父级被停用时仍可以按 `parentCode` 获取子级；
  - 为 `Organization` 节点增加 `childrenCount` 字段（JOIN 子表统计），契约更新 `docs/api/schema.graphql`。
- `organizationHierarchy`/`organizationSubtree`：保留现有 `childrenCount`，并校验停用父级时仍能返回子节点。

### 4.4 前端
- 统一通过 `ParentOrganizationSelector` 的 GraphQL 查询增加 `childrenCount`，启用 `onlyActive=true` 过滤，确保停用组织不出现在列表中，同时保留 `includeDisabledAncestors` 场景以便子级渲染。
- 恢复 `frontend/src/shared/utils/organizationPermissions.ts:37` 的 `childCount` 校验，改为使用 GraphQL 返回值，并在禁用删除时提供原因提示。
- 在删除按钮交互中追加基于 API 错误码 `HAS_CHILD_UNITS` 的兜底提示。
- 在停用后的组织详情页刷新父级路径缓存，避免 selector 缓存中残留旧状态。

### 4.5 测试与自动化
- 命令服务：Go 单测 + 集成测试覆盖停用/删除路径及幂等行为。
- 查询服务：新增针对停用父级/子级查询的 SQL golden test。
- 前端：Vitest 单测覆盖 `organizationPermissions` 与 selector 过滤逻辑；E2E（Playwright）增加“停用父级后仍可选子级”场景。
- CI 强制执行 `node scripts/quality/architecture-validator.js`、`node scripts/generate-implementation-inventory.js`，确保契约与实现一致。

## 5. 验收标准
- OpenAPI 与 GraphQL 契约完成更新，含停用/删除行为说明及新字段。
- 停用操作后：
  - API 返回父级 `status=INACTIVE`；
  - 子级状态及查询结果不变；
  - 组织选择器中父级消失但子级可用。
- 删除操作：
  - 有子级 → 返回 409 `HAS_CHILD_UNITS`；
  - 无子级 → 软删除成功，`isDeleted=true` 且审计日志写入。
- `organizationPermissions` 再次启用子组织校验，无 TODO 过期项。
- 全量测试（`make test`, `make test-integration`, `npm run test`, `npm run lint`, Playwright 场景）通过。

## 6. 风险与缓解
| 风险 | 影响 | 缓解措施 |
| --- | --- | --- |
| 契约更新影响现有调用方 | 前后端需同步切换 | 发布前发布 Contract 变更公告，提供多环境验证窗口 |
| 停用后缓存未刷新 | 选择器仍显示旧数据 | 停用/删除成功后触发层级缓存刷新 + 前端失效本地缓存 |
| 子组织检测漏算历史数据 | 误删存在未来版本的组织 | SQL 检测条件包含 `status <> 'DELETED' AND deleted_at IS NULL`，并对 `effective_date` 不做限制确保捕获未来记录 |

## 7. 里程碑
1. **契约更新**（+1 天）：更新 OpenAPI/GraphQL 并产出审阅记录。
2. **命令服务实现**（+3 天）：停用测试补强、删除拦截与审计。
3. **查询服务增强**（+2 天）：`childrenCount` 与状态过滤优化。
4. **前端联调**（+2 天）：权限校验恢复、选择器行为验证。
5. **联合测试与归档**（+1 天）：跑通测试套件、更新 IIG、归档计划。

## 8. 验证步骤
1. `node scripts/generate-implementation-inventory.js` → 确认停用/删除端点一致。
2. `curl -X POST http://localhost:9090/api/v1/organization-units/{code}/suspend` → 子组织 `status` 不变。
3. `npm --prefix frontend run test -- ParentOrganizationSelector` → 确认过滤逻辑。
4. `curl -X POST http://localhost:9090/api/v1/organization-units/{code}/events -H 'If-Match: <etag>' -d '{"operation":"DELETE_ORGANIZATION","effectiveDate":"2025-09-30","operationReason":"合规清理"}'` → 有子级返回 409 `HAS_CHILD_UNITS`，无子级成功并返回 `isDeleted=true`。
5. `make test-integration` / Playwright `organization-create.spec.ts` → 验证端到端流程。

## 9. 开放问题建议
- **删除请求需要的字段**：沿用命令端现有模式，强制请求体携带 `effectiveDate`、`operationReason`，并通过 `If-Match` 传递最新版本的 ETag，保持幂等与并发控制。
- **复用既有逻辑**：新增的 `SoftDeleteOrganization` 应封装在 service 层，内部复用现有 `timelineManager.DeleteVersion` 与审计记录逻辑，避免在 handler 中散落 SQL，符合 Go“组合胜于继承”的可维护理念。
- **客户端改造**：统一 REST 客户端新增 `deleteOrganization` 方法时沿用 `POST /events`，复用现有中间件传递 `X-Tenant-ID`、`If-Match`、`Idempotency-Key`，确保无重复事实来源。

---

**完成后**：归档至 `docs/archive/development-plans/`，并在 `06-integrated-teams-progress-log.md` 记录验收结论与测试证据。
