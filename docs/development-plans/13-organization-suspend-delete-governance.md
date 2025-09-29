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
- **停用**：继续复用 `POST /api/v1/organization-units/{code}/suspend` 与 `POST /api/v1/organization-units/{code}/activate`，在 `docs/api/openapi.yaml` 中明确“停用不影响下级组织”并记录成功示例；请求体强制携带 `effectiveDate`、`If-Match`（ETag），`operationReason` 允许可选用于审计。契约更新完成后必须与查询/前端联动说明同步，避免出现第二事实来源。
- **删除**：不新增独立 `DELETE` 端点，保持契约收敛在现有命令体系。扩展 `POST /api/v1/organization-units/{code}/events`：
  - 新增 `DELETE_ORGANIZATION` 事件枚举，沿用 `POST /events` 实现组织软删除；
  - 请求体要求 `effectiveDate`、`If-Match`，`operationReason` 可选；
  - 响应示例包含时间线重算结果并强调 `status='DELETED'`；
  - **重复校验**：`HAS_CHILD_UNITS` 错误示例已存在，无需再次编写，重点在于合同文本解释并关联现有示例；
  - 权限 scope：`org:delete`；
  - 契约更新后执行 `node scripts/generate-implementation-inventory.js` 校验实现登记，并在变更记录中引用本计划。

### 4.2 命令服务（Go）
- **停用校验**：维持既有时态管理逻辑，同时保证 suspend/activate 响应统一携带最新 ETag 供前端乐观并发控制。
- **删除防线**：
  - 在 `CreateOrganizationEvent` 内引入 `DELETE_ORGANIZATION` 分支：删除前调用 `CountNonDeletedChildren` 校验 `status <> 'DELETED'` 的子组织，命中时抛出 `HAS_CHILD_UNITS`（409）。
  - 当校验通过时复用时间轴管理器 `DeleteOrganization` 写入 `status='DELETED'` 版本并重算时间线。
  - 所有操作要求 `If-Match` 与当前记录一致后才执行，防止并发误删；失败流程统一记录审计事件。
  - **测试复用**：仓储层已有 `organization_children_test.go`，无需重复创建，仅追加覆盖空集/错误分支或移动到契约态测试套件；`organization_internal_test.go` 中可扩展现有 ETag 解析测试，无需新文件。

### 4.3 查询服务（GraphQL）
- `cmd/organization-query-service/main.go`：
  - 维持当前默认行为（返回 status != 'DELETED' 的记录），**优先复用现有 `status` 过滤能力** —— 组织选择器改为通过 `status: ACTIVE` 获得同效行为，无需新增 `onlyActive` 字段；
  - 增加可选字段 `includeDisabledAncestors`（默认 false），当父级被停用时仍可以按 `parentCode` 获取子级；
  - 为 `Organization` 节点增加 `childrenCount` 字段（JOIN 子表统计），契约更新 `docs/api/schema.graphql`。
- `organizationHierarchy`/`organizationSubtree`：保留现有 `childrenCount`，并校验停用父级时仍能返回子节点。

### 4.4 前端
- 统一通过 `ParentOrganizationSelector` 的 GraphQL 查询增加 `childrenCount` 并明确使用 `status: ACTIVE` 过滤，避免与后端新增参数重复；保留 `excludeDescendantsOf` 等现有能力以兼容旧逻辑。
- 恢复 `frontend/src/shared/utils/organizationPermissions.ts` 中的子组织校验逻辑，直接基于 GraphQL 的 `childrenCount` 判断，不再依赖注释代码或本地计算。
- 在删除按钮交互中追加基于 API 错误码 `HAS_CHILD_UNITS` 的兜底提示。
- 在停用后的组织详情页刷新父级路径缓存，避免 selector 缓存中残留旧状态。

### 4.5 一致性校验与回归策略
- **跨层一致性**：每个阶段完成后执行 `node scripts/generate-implementation-inventory.js`，并对照 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 确认导出与契约更新保持一致。
- **数据验证**：通过数据库快照比对，确认停用仅改写目标组织的最新有效版本，删除将 `status` 置为 `DELETED` 且不影响历史版本。
- **安全回滚**：保留迁移前后数据库备份，若发现契约/实现漂移立即回滚代码并恢复数据库，再重新评估方案。
- **审计追踪**：确保所有停用、恢复、删除操作写入审计日志并附带 `operationReason`，供后续合规稽核。

### 4.6 测试与自动化
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
1. `node scripts/generate-implementation-inventory.js` → 确认停用/删除端点一致且已登记。
2. `curl -X POST http://localhost:9090/api/v1/organization-units/{code}/suspend` → 子组织 `status` 不变，响应返回最新 ETag。
3. `npm --prefix frontend run test -- ParentOrganizationSelector` → 确认 `onlyActive` 过滤逻辑与 `includeDisabledAncestors` 场景。
4. `curl -X POST http://localhost:9090/api/v1/organization-units/{code}/events -H 'If-Match: <etag>' -d '{"eventType":"DELETE_ORGANIZATION","effectiveDate":"2025-09-30","operationReason":"合规清理"}'` → 有子级返回 409 `HAS_CHILD_UNITS` 且审计记录仍写入失败原因，无子级成功并返回 `status='DELETED'` 与时间轴版本。
5. `make test-integration` / Playwright `organization-create.spec.ts` / 新增删除用例 → 验证端到端流程。

## 9. 开放问题建议
- **删除请求需要的字段**：沿用命令端现有模式，强制请求体携带 `effectiveDate`，可选提交 `operationReason`，并通过 `If-Match` 传递最新版本的 ETag，保持幂等与并发控制。
- **复用既有逻辑**：新增的 `SoftDeleteOrganization` 应封装在 service 层，内部复用现有 `timelineManager.DeleteVersion` 与审计记录逻辑，避免在 handler 中散落 SQL，符合 Go“组合胜于继承”的可维护理念。
- **客户端改造**：统一 REST 客户端新增 `deleteOrganization` 方法时沿用 `POST /events`，复用现有中间件传递 `X-Tenant-ID`、`If-Match`、`Idempotency-Key`，确保无重复事实来源。

---

**完成后**：归档至 `docs/archive/development-plans/`，并在 `06-integrated-teams-progress-log.md` 记录验收结论与测试证据。

- **契约**：`docs/api/openapi.yaml`、`docs/api/schema.graphql` 尚未更新，本计划提出的变更待评审审批后执行。
- **命令服务**：`DELETE_ORGANIZATION`、子组织计数、`If-Match` 逻辑尚未落地；现有代码仍停留在单点停用流程。
- **查询服务**：暂未提供 `childrenCount` 聚合或 `includeDisabledAncestors` 参数，仍沿用原有查询；`organizations` 查询已支持 `status` 过滤，可直接复用。
- **前端**：`ParentOrganizationSelector` 通过 `status: ACTIVE` 查询仍返回停用组织（缺少 childrenCount 支撑及联动刷新），`organizationPermissions.ts` 的子组织校验被注释，缺乏删除失败提示。
- **测试与脚本**：目前仅完成可行性评估；自动化用例与脚本尚未改造。

## 11. 后续工作
- **阶段一：契约评审** → 在本计划基础上提交 OpenAPI/GraphQL 变更草案，完成评审会议并锁定唯一事实来源。
- **阶段二：命令服务实现** → 引入 `DELETE_ORGANIZATION` 事件、子节点校验、时间轴与审计补强，同步补写单测。
- **阶段三：查询服务增强** → 实现 `childrenCount`、`onlyActive`、`includeDisabledAncestors`，完善 SQL 与缓存策略。
- **阶段四：前端改造** → 更新 Selector、恢复权限校验、实现删除反馈与 Playwright/Vitest 场景。
- **阶段五：回归与归档** → 完成端到端测试、质量脚本、生成实现清单，并将计划归档至 `docs/archive/development-plans/`，更新 `06-integrated-teams-progress-log.md`。
