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
- **停用**：继续复用 `POST /api/v1/organization-units/{code}/suspend` 与 `POST /api/v1/organization-units/{code}/activate`，在 `docs/api/openapi.yaml` 中明确“停用不影响下级组织”并记录成功示例；请求体强制携带 `effectiveDate`、`If-Match`（ETag），`operationReason` 允许可选用于审计。
- **删除**：不新增独立 `DELETE` 端点，保持契约收敛在现有命令体系。扩展 `POST /api/v1/organization-units/{code}/events`：
  - 新增 `DELETE_ORGANIZATION` 事件枚举，沿用 `POST /events` 实现组织软删除；
  - 请求体要求 `effectiveDate`、`If-Match`，`operationReason` 可选；
  - 响应示例包含时间线重算结果并强调 `status='DELETED'`；
  - 失败示例新增 `HAS_CHILD_UNITS` 冲突；
  - 权限 scope：`org:delete`；
  - 契约更新后执行 `node scripts/generate-implementation-inventory.js` 校验实现登记。

### 4.2 命令服务（Go）
- **停用校验**：维持既有时态管理逻辑，同时保证 suspend/activate 响应统一携带最新 ETag 供前端乐观并发控制。
- **删除防线**：
  - 在 `CreateOrganizationEvent` 内引入 `DELETE_ORGANIZATION` 分支：删除前调用 `CountNonDeletedChildren` 校验 `status <> 'DELETED'` 的子组织，命中时抛出 `HAS_CHILD_UNITS`（409）。
  - 当校验通过时复用时间轴管理器 `DeleteOrganization` 写入 `status='DELETED'` 版本并重算时间线。
  - 所有操作要求 `If-Match` 与当前记录一致后才执行，防止并发误删；失败流程统一记录审计事件。
  - 新增 `organization_children_test.go` 覆盖子组织统计 SQL，`organization_internal_test.go` 验证 `If-Match` 解析。

### 4.3 查询服务（GraphQL）
- `cmd/organization-query-service/main.go`：
  - 维持当前默认行为（返回 status != 'DELETED' 的记录），新增 `onlyActive` 过滤参数供组织选择器使用，避免破坏现有调用；
  - 新增可选字段 `includeDisabledAncestors`（默认 false），当父级被停用时仍可以按 `parentCode` 获取子级；
  - 为 `Organization` 节点增加 `childrenCount` 字段（JOIN 子表统计），契约更新 `docs/api/schema.graphql`。
- `organizationHierarchy`/`organizationSubtree`：保留现有 `childrenCount`，并校验停用父级时仍能返回子节点。

### 4.4 前端
- 统一通过 `ParentOrganizationSelector` 的 GraphQL 查询增加 `childrenCount`，启用 `onlyActive=true` 过滤，确保停用组织不出现在列表中，同时保留 `includeDisabledAncestors` 场景以便子级渲染。
- 恢复 `frontend/src/shared/utils/organizationPermissions.ts:37` 的 `childCount` 校验，改为使用 GraphQL 返回值，并在禁用删除时提供原因提示。
- 在删除按钮交互中追加基于 API 错误码 `HAS_CHILD_UNITS` 的兜底提示。
- 在停用后的组织详情页刷新父级路径缓存，避免 selector 缓存中残留旧状态。

### 4.5 测试与自动化
- 命令服务：`go test ./cmd/organization-command-service/internal/...`（含新增仓储 & If-Match 单测）。
- 查询服务：`go test ./cmd/organization-query-service/...` 验证新过滤条件及 `childrenCount` 聚合。
- 前端：`npm run lint` 通过（命令层与选择器改造待补充 Playwright 场景）。
- 共通：`node scripts/generate-implementation-inventory.js` 更新实现清单，确认新增导出已登记。

## 5. 验收标准
- OpenAPI 与 GraphQL 契约完成更新，含停用/删除行为说明及新字段。
- 停用操作后：
  - API 返回父级 `status=INACTIVE`；
  - 子级状态及查询结果不变；
  - 组织选择器中父级消失但子级可用。
- 删除操作：
  - 有子级 → 返回 409 `HAS_CHILD_UNITS`；
  - 无子级 → 软删除成功，`status` 切换为 `DELETED` 且审计日志写入。
- `organizationPermissions` 再次启用子组织校验，无 TODO 过期项。
- 全量测试（`make test`, `make test-integration`, `npm run test`, `npm run lint`, Playwright 场景）通过。

## 6. 风险与缓解
| 风险 | 影响 | 缓解措施 |
| --- | --- | --- |
| 契约更新影响现有调用方 | 前后端需同步切换 | 发布前发布 Contract 变更公告，提供多环境验证窗口 |
| 停用后缓存未刷新 | 选择器仍显示旧数据 | 停用/删除成功后触发层级缓存刷新 + 前端失效本地缓存 |
| 子组织检测漏算历史数据 | 误删存在未来版本的组织 | SQL 检测条件包含 `status <> 'DELETED'`，并对 `effective_date` 不做限制确保捕获未来记录 |

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
4. `curl -X POST http://localhost:9090/api/v1/organization-units/{code}/events -H 'If-Match: <etag>' -d '{"eventType":"DELETE_ORGANIZATION","effectiveDate":"2025-09-30","operationReason":"合规清理"}'` → 有子级返回 409 `HAS_CHILD_UNITS`，无子级成功并返回 `status='DELETED'`（示例带可选 `operationReason` 以便审计记录）。
5. `make test-integration` / Playwright `organization-create.spec.ts` → 验证端到端流程。

## 9. 开放问题建议
- **删除请求需要的字段**：沿用命令端现有模式，强制请求体携带 `effectiveDate`，可选提交 `operationReason`，并通过 `If-Match` 传递最新版本的 ETag，保持幂等与并发控制。
- **复用既有逻辑**：新增的 `SoftDeleteOrganization` 应封装在 service 层，内部复用现有 `timelineManager.DeleteVersion` 与审计记录逻辑，避免在 handler 中散落 SQL，符合 Go“组合胜于继承”的可维护理念。
- **客户端改造**：统一 REST 客户端新增 `deleteOrganization` 方法时沿用 `POST /events`，复用现有中间件传递 `X-Tenant-ID`、`If-Match`、`Idempotency-Key`，确保无重复事实来源。

---

**完成后**：归档至 `docs/archive/development-plans/`，并在 `06-integrated-teams-progress-log.md` 记录验收结论与测试证据。

## 10. 最新进展（2025-09-27）
- **契约落地**：`docs/api/openapi.yaml`、`docs/api/schema.graphql` 已完成增量更新，替换示例与新增 `OrganizationEventRequest` 结构、`childrenCount` 字段、`onlyActive` / `includeDisabledAncestors` 过滤，并执行 `node scripts/generate-implementation-inventory.js` 验证清单同步。
- **命令服务**：`CreateOrganizationEvent` 引入 `DELETE_ORGANIZATION` 分支、`CountNonDeletedChildren` 检查与 `If-Match` 并发防线；新增 `DeleteOrganization` 时间轴流程与审计记录复用；单测覆盖子组织统计与 ETag 解析。
- **查询服务**：`GetOrganizations` 聚合 `childrenCount`、默认排除停用父节点并在开启 `includeDisabledAncestors` 时放宽；GraphQL schema 映射 `childrenCount` 字段；`go test ./cmd/organization-query-service/...` 通过。
- **前端联动**：Parent Selector 启用 `onlyActive` + `includeDisabledAncestors` 查询，恢复删除权限对子组织的校验，新增 `useDeleteOrganization` 钩子并对 `HAS_CHILD_UNITS` 错误给出提示；`npm run lint` 已通过。
- **验证命令**：
  - `go test ./cmd/organization-command-service/internal/...`
  - `go test ./cmd/organization-query-service/...`
  - `npm run lint`
  - `node scripts/generate-implementation-inventory.js`

## 11. 后续工作
- **测试补齐**：补充命令服务集成测试与查询服务 SQL golden test，新增 Parent Selector / 删除交互的 Vitest/Playwright 场景。
- **缓存刷新**：评估停用/删除后层级缓存刷新策略（目前依赖前端刷新），补充自动化验证。
- **文档归档准备**：待前述测试完备后整理验收报告并归档至 `docs/archive/development-plans/`，同步更新进展日志 `06-integrated-teams-progress-log.md`。
