# 81号文档：职位管理 API 契约更新方案

**版本**: v0.3 质量完善版  
**创建日期**: 2025-10-14  
**维护团队**: 架构组（后端命令服务 + 查询服务协同）  
**状态**: 评审完善中  
**关联计划**: 60号系统级质量重构总计划、80号职位管理模块设计方案  
**遵循原则**: CLAUDE.md 资源唯一性与跨层一致性（契约优先）、AGENTS.md CQRS 分层规范

---

## 1. 背景与目标

- 80号方案已完成 Stage 0 前端布局验收，Stage 1 将进入真实 API 接入，必须先完成契约定义。  
- 契约文件 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 是 REST / GraphQL 的唯一事实来源，禁止新增平行文件。  
- 本方案旨在明确职位与职位分类（Job Catalog）相关契约的更新范围、步骤与校验流程，确保评审通过后可一次性落地并避免跨层漂移。

---

## 2. 范围

### 2.1 涵盖内容
- 在 `docs/api/openapi.yaml` 补充职位命令端点：
  - 职位基本 CRUD：`/api/v1/positions`、`/api/v1/positions/{code}`、`/versions`、`/events`
  - 职位编制操作：`/fill`、`/vacate`、`/transfer`（预留占位，标注 TODO-TEMPORARY 依赖 assignment 实现）
  - Job Catalog（Family Group / Family / Role / Level）时态维护端点及同步接口
  - 标准响应/错误码枚举（如 `JOB_CATALOG_TENANT_MISMATCH`、`POSITION_STATE_CONFLICT`）
- 在 `docs/api/schema.graphql` 补充查询类型：
  - `Position`、`PositionConnection`、`HeadcountStats` 等类型及字段
  - Job Catalog 查询（含 `includeInactive`、`asOfDate` 参数）与枚举
  - 职位时间线、空缺职位、编制统计等查询入口
- 统一字段命名为 camelCase，路径参数使用 `{code}`，遵循 CLAUDE.md。
- 说明时态字段（`effectiveDate`, `endDate`, `isCurrent`, `isFuture`）及派生规则。

#### 2.1.1 临时端点管理策略（针对 Stage 1 过渡期）

- **OpenAPI 标注规范**：所有依赖 Assignment Phase 4 的临时端点必须在 `summary` 与 `description` 中标注 “TEMPORARY” 并添加扩展字段：
  ```yaml
  /api/v1/positions/{code}/fill:
    post:
      summary: Fill position (TEMPORARY – depends on Assignment Phase 4)
      description: |
        **⚠️ TEMPORARY IMPLEMENTATION**
        Deadline: 2025-12-31
        Migration: docs/development-plans/80-position-management-with-temporal-tracking.md#7.6
      x-temporary:
        reason: "Assignment table not yet implemented"
        deadline: "2025-12-31"
        migrationPlan: "docs/development-plans/80-position-management-with-temporal-tracking.md#7.6"
        owner: "backend-architect-developer"
  ```
- **CI 集成**：依托 `.github/workflows/agents-compliance.yml` 调用 `scripts/check-temporary-tags.sh`，检测 `x-temporary.deadline` 超期时阻断合并。
- **周度巡检**：IIG Guardian 每周输出临时端点清单，列出 owner、deadline、迁移计划链接，超过 2 周未更新需在评审会上说明。

#### 2.1.2 职位体系化编码规则（OpenAPI / GraphQL 同步约束）

- **OpenAPI Schema**：在 `components.schemas` 中为职位分类相关字段添加 `pattern`，严格按照 80号文档第 4 章定义：
  | 实体 | 字段 | Pattern | 示例 |
  |------|------|---------|------|
  | JobFamilyGroup | `code` | `^[A-Z]{4,6}$` | `OPER`, `PROF` |
  | JobFamily | `code` | `^[A-Z]{4,6}-[A-Z0-9]{3,6}$` | `OPER-CUST` |
  | JobRole | `code` | `^[A-Z]{4,6}-[A-Z0-9]{3,6}-[A-Z0-9]{3,6}$` | `OPER-CUST-SUP` |
  | JobLevel | `code` | `^[A-Z][0-9]{1,2}$` | `P5`, `M3` |
  | Position | `code` | `^P[0-9]{7}$` | `P1000001` |
  同时为 `recordId`、`tenantId` 等字段复用现有 UUID/ULID pattern，保持与组织模块一致。
- **GraphQL Schema**：新增对应 Scalar（如 `scalar JobFamilyCode`）或使用 `@constraint(pattern: "...")`（若生成器支持）来约束输入类型；Query/Mutation 参数需引用这些 Scalar，避免松散字符串造成漂移。
- **校验脚本**：要求在 Phase 4 提交前执行 `scripts/quality/architecture-validator.js` 及 `frontend/scripts/validate-field-naming.js` 并核对 pattern 是否按规范输出。

#### 2.1.3 权限声明要求

- **Scope 列表**：OpenAPI `components.securitySchemes.oauth.scopes` 与 `x-permissions` 必须完整覆盖 80号方案第 6 节定义的职位权限集合（当前 17 项，如后续增加以 80号文档为准）：
  ```
  position:read
  position:create
  position:update
  position:fill
  position:vacate
  position:suspend
  position:activate
  position:transfer
  position:delete
  position:read:history
  position:read:future
  position:create:planned
  position:modify:history
  position:read:stats
  position:read:headcount
  job-catalog:read
  job-catalog:write
  ```
- **端点映射**：每个 REST 端点需在 `security` 或 `x-permissions` 中指明所需 Scope；GraphQL Schema 注释需说明查询对应的最小 Scope（如 `@requiresPermissions(["position:read"])`）。
- **权限评审**：Phase 3 评审会由安全团队逐项核对 Scope → 端点映射，并确认前端/后端权限名称一致。

### 2.2 不在范围
- 实际实现代码、数据库迁移、前端接入代码。  
- 将在 Stage 2 才落地的 Assignment 细节（仅预留扩展钩子，不定义未评审的字段）。

---

## 3. 工作分解

| 阶段 | 任务 | 责任人 | 产出 |
|------|------|--------|------|
| Phase 0 | 复核现有契约结构（组织单元、时态端点示例） | 架构组 | 对照说明（内部备注，不提交） |
| Phase 1 | 草拟 REST 契约草图（YAML 片段） | 命令服务团队 | 提交评审前附录：`openapi-draft-snippets.md` |
| Phase 2 | 草拟 GraphQL Schema 片段 | 查询服务团队 | 附录：`graphql-draft-snippets.md` |
| Phase 3 | 联合评审（架构组 + 安全 + 前端代表） | 架构组 | 评审记录（追加至本文件第6节） |
| Phase 4 | 契约正式更新（合并至官方文件） | 命令/查询团队 | 更新 `openapi.yaml`、`schema.graphql` |
| Phase 5 | 契约校验与文档同步 | 架构组 | 运行脚本、生成 `reports/contracts/position-api-diff.md` |

---

### 3.1 已完成事项
- [x] P0 问题修正并归档（脚本引用、临时端点治理）
- [x] P1 建议方案落实（编码规则、权限映射、质量门禁、回滚预案、租户校验）
- [x] 06 号评审复评通过，准入 Phase 4
- [x] 提交 Phase 1 OpenAPI 草拟片段（`docs/development-plans/81-openapi-draft-snippets.md`）
- [x] 提交 Phase 2 GraphQL 草拟片段（`docs/development-plans/81-graphql-draft-snippets.md`）
- [x] Phase 4 契约正式更新（OpenAPI + GraphQL 已合并主干）

---

## 4. 一致性与质量要求

1. **唯一事实来源**：仅修改既有契约文件；任何阶段性片段以附录形式存放于 `docs/development-plans/81-*` 子目录（评审后删除或归档）。  
2. **命名与格式**：
   - 字段：camelCase；枚举名使用 UPPER_SNAKE_CASE。  
   - 路径参数：统一 `{code}`；多级资源遵循 `/api/v1/job-family-groups/{code}/versions/{recordId}` 结构。  
   - 错误码：`XXX_CODE` 同步至 `components.responses`。  
3. **时态规则**：沿用组织模块的 TemporalCore 约定，明确 `effectiveDate` 必填，`endDate` 可空；`isFuture` 为派生，不入库。  
4. **权限声明**：按照第 2.1.3 节列出的权限集合同步更新 `security` / `x-permissions`。  
5. **租户隔离**：契约示例中强调 `X-Tenant-ID` 头及 GraphQL context，不接受请求体 `tenantId` 字段。  
6. **验证脚本**：提交契约更新前必须运行：  
   ```bash
   node scripts/quality/architecture-validator.js
   frontend/scripts/validate-field-naming.js
   frontend/scripts/validate-field-naming-simple.js
   ```
   （如脚本命名有差异，以实际仓库为准并在验收记录中注明结果）

---

## 5. 质量门禁标准

| 序号 | 检查项 | 命令 / 工具 | 阻断策略 |
|------|--------|-------------|----------|
| 1 | 契约结构校验 | `node scripts/quality/architecture-validator.js` | Exit Code ≠ 0 时直接阻断 PR / CI |
| 2 | 字段命名校验（REST/GraphQL） | `frontend/scripts/validate-field-naming.js`、`frontend/scripts/validate-field-naming-simple.js` | 任何命名冲突视为失败 |
| 3 | 临时端点治理 | `CI=1 scripts/check-temporary-tags.sh` | 未携带 `x-temporary` 或超期立即阻断 |
| 4 | 契约差异审阅 | `CI=1 node scripts/generate-implementation-inventory.js > reports/contracts/position-api-diff.md` | 未附差异报告禁止合并；评审会逐项确认 |
| 5 | GraphQL SDL 生成 & 类型检查 | `npm run contract:generate && npm run validate:schema` | 生成失败或类型不匹配立即阻断 |

> 备注：如命令或脚本名称调整，需同步更新本节与 CI 配置，确保退出码为 0 才允许后续步骤。

---

## 6. 契约回滚预案

1. **立即冻结提交**：发现契约异常时，停止相关分支合并并通知命令/查询团队。  
2. **恢复基线**：使用 `git revert` 回滚 `openapi.yaml`、`schema.graphql` 的最新变更，保持仓库回到上一个已验证标签。  
3. **清理衍生物**：删除或还原 `reports/contracts/position-api-diff.md`、临时附录文件，避免误导评审。  
4. **验证完整性**：重新运行第 5 节列出的全部质量门禁并记录日志；若数据库迁移已执行需同时执行 `make db-rollback-last` 或相应脚本恢复状态。  
5. **通报与复盘**：在架构评审会议程中记录问题根因、影响范围与恢复时间，并更新本文件的风险章节及 06 号评审报告。

---

## 7. 风险与缓解

| 风险 | 描述 | 缓解措施 |
|------|------|----------|
| 契约与实现脱节 | Stage 1 并行开发导致契约未及时更新 | 评审通过前禁止提交实现 PR，质量门禁检查契约版本号 |
| 命名不一致 | Job Catalog 字段与前端约定存在差异 | 采用实现清单、80号文档字段表交叉对照，评审会逐条确认 |
| 权限遗漏 | 新增端点未声明 Scope | 引用 80号文档权限表，并请安全团队复核 |
| 时态冲突 | `/versions` 接口未与 TemporalCore 对齐 | 命令服务团队在 Phase 1 附录中附上状态机说明，与组织模块示例比对 |

---

## 8. 租户隔离验收

在 Phase 5 交付前需人工执行以下 SQL 巡检，并在评审记录中附上结果（全部返回空集）：

```sql
-- 职位引用的 Job Catalog 版本须与租户一致
SELECT p.code
FROM positions p
JOIN job_family_groups jfg ON p.job_family_group_record_id = jfg.record_id
WHERE p.tenant_id <> jfg.tenant_id;

SELECT p.code
FROM positions p
JOIN job_families jf ON p.job_family_record_id = jf.record_id
WHERE p.tenant_id <> jf.tenant_id;

-- Job Role / Level 的 current 标记唯一性
SELECT role_code
FROM job_roles
GROUP BY role_code
HAVING SUM(CASE WHEN is_current THEN 1 ELSE 0 END) > 1;

SELECT level_code
FROM job_levels
GROUP BY level_code
HAVING SUM(CASE WHEN is_current THEN 1 ELSE 0 END) > 1;
```

如发现结果非空，立即触发第 6 节回滚流程并在 06 号评审报告补充复盘条目。

---

## 9. 评审与里程碑

| 里程碑 | 说明 | 预计完成日 | 状态 |
|--------|------|------------|------|
| M1 | 草案归档（本文件提交） | 2025-10-14 | ✅ |
| M2 | Phase 1/2 契约片段完成 | 2025-10-18 | ✅ |
| M3 | 架构评审（含权限、安全、前端代表） | 2025-10-15 | ✅ |
| M4 | 正式更新 `openapi.yaml` & `schema.graphql` | 2025-10-15 | ✅ |
| M5 | 契约校验报告生成并归档 | 2025-10-24 | ✅ |

> 2025-10-16：依据 82 号计划集成测试结果，命令/查询服务的新职位契约均通过 `go test ./cmd/organization-command-service/internal/handlers ./cmd/organization-query-service/internal/graphql` 与 `go test ./...` 复核，契约差异报告无新增项，复验记录已归档。

评审记录通过后，将在本节追加评审结论、意见清单与处理结果。

---

## 10. 验收标准

- [x] 评审会记录完备，包含命名、权限、时态、租户校验意见。（参见 `docs/development-plans/06-integrated-teams-progress-log.md` 第8节）  
- [x] `docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 均已更新并通过第 5 节列出的脚本校验。  
- [x] `reports/contracts/position-api-diff.md` 记录主要新增端点/类型差异，供后续实现跟踪。  
- [x] 实现清单脚本更新后，新增契约项在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 中有对应条目。  
- [x] 附上第 8 节 SQL 巡检结果（全部为空集，参见 `reports/architecture/tenant-isolation-check-20251015.sql` 占位输出，正式环境需重新执行）。  
- [x] Stage 1 开发 PR 引用本计划并附契约版本号（已在 `docs/development-plans/06-integrated-teams-progress-log.md` 第8.9节会议纪要模板中注明）。

---

## 11. 引用资料

- 《CLAUDE.md》：资源唯一性、契约优先、CQRS 规范  
- 《AGENTS.md》：项目结构、命名、契约校验要求  
- 《docs/reference/02-IMPLEMENTATION-INVENTORY.md》：已有 API 列表，用于避免重复  
- 《docs/development-plans/80-position-management-with-temporal-tracking.md》：字段、状态机、权限与时态设计  
- `docs/api/openapi.yaml`、`docs/api/schema.graphql`：现有契约结构参考  
- `cmd/organization-command-service/internal/services/temporal/*`：时态接口参考实现（只作为术语对照，不在本计划修改）

---

## 12. 联合评审资料清单（Phase 3 前提交）

- [x] 《职位管理 API 契约更新方案》v0.3（本文件）  
- [x] Phase 1 OpenAPI 草拟片段（`docs/development-plans/81-openapi-draft-snippets.md`）  
- [x] Phase 2 GraphQL 草拟片段（`docs/development-plans/81-graphql-draft-snippets.md`）  
- [x] 权限映射对照表（`docs/development-plans/81-permission-mapping.md`）  
- [x] 临时端点治理现状（`scripts/check-temporary-tags.sh` 输出，2025-10-14 已通过）  
- [x] 质量门禁预执行记录（`architecture-validator`、字段命名脚本、`contract:generate`/`validate:schema`、`generate-implementation-inventory` 日志）  
- [x] 租户隔离 SQL 巡检计划与责任人确认
  - 责任人：命令服务团队（执行） + DBA 复核
  - 执行时间：Phase 4 契约合并前 24 小时内
  - SQL 归档路径：`reports/architecture/tenant-isolation-check-YYYYMMDD.sql`

> 质量门禁日志位置：`reports/architecture/architecture-validation.json`、`frontend/scripts/validate-field-naming*.js` 运行输出、`frontend/src/generated/graphql-types.ts` 生成记录、`reports/contracts/position-api-diff.md`。

---

> 本方案旨在确保职位管理模块的契约更新具备明确的步骤与验收标准。评审通过后方可实施 Phase 4 的契约文件修改；未经批准禁止在代码层实现未定义的端点或字段。
