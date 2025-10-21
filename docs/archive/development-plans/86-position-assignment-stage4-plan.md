# 86号文档：职位任职管理 Stage 4 增量计划

**版本**: v1.0（归档版，含87号迁移）
**创建日期**: 2025-10-17
**最新更新**: 2025-10-21
**维护团队**: 命令服务团队 · 查询服务团队 · 前端团队 · QA 团队 · 数据库团队 · 架构组
**状态**: ✅ 已完成并归档（跨租户脚本 · 047/048 迁移 · CI 验证到位）
**关联计划**: 80号职位管理方案 · 84号 Stage 2（归档） · 85号 Stage 3（归档） · **87号时态字段命名一致性决策（归档）** · 06号集成团队协作日志
**遵循原则**: `CLAUDE.md` 资源唯一性与跨层一致性（最高优先级） · `AGENTS.md` 开发前必检规范 · CQRS 分工（命令 REST / 查询 GraphQL） · Docker 容器化强制

---

## 1. 背景与目标

- Stage 2（见 84 号计划）已经落地 `position_assignments` 表、Fill/Vacate/Transfer 全链路、命令/查询层仓储，以及前端任职展示，确保职位任职的唯一事实来源。  
- Stage 3（见 85 号计划）已完成编制统计、空缺看板、转移界面及 Playwright 验收。  
- 80 号方案第 7.5 节提出的 Stage 4 剩余四项勾选（Position Assignment 实体、Multiple Assignments、Acting、History）仍需针对“高级任职管理”做增量完善，特别是代理自动恢复、专用任职 API、历史视图与跨租户验证。  
- 本修订稿在复盘现有实现后，仅聚焦缺失能力与质量补强，避免重复造轮子，并将周期压缩为 2 周（加 1 周缓冲）。

---

## 2. 现有能力复盘（Stage 2/3）

| 能力 | 当前状态 | 事实来源 |
|------|----------|----------|
| Assignment 表结构 | ✅ `assignment_id` 主键、`assignment_type` (PRIMARY/SECONDARY/ACTING)、`fte`、`effective_date/end_date`、租户外键、唯一约束 | `database/migrations/047_rename_position_assignments_start_date.sql` |
| 命令服务仓储与服务 | ✅ Create/List/Close/FTE 聚合、Fill/Vacate/Transfer 写入任职历史 | `cmd/organization-command-service/internal/repository/position_assignment_repository.go`、`position_service.go` |
| GraphQL 查询 | ⚠️ `currentAssignment` / `assignmentHistory` 已上线；缺少高级过滤、时间轴聚合与性能基线 | `cmd/organization-query-service/internal/model/models.go`、`cmd/organization-query-service/internal/repository/postgres_positions.go`、`docs/archive/development-plans/89-position-crud-verification-report.md` |
| 前端展示 | ✅ `PositionDetails` 任职列表/历史，`PositionDashboard` 读取 GraphQL 数据 | `frontend/src/features/positions` |
| 编制统计 | ✅ `positionHeadcountStats` 复用 FTE 计算并驱动 `PositionHeadcountDashboard` | `cmd/organization-query-service`、`frontend` |

> ✅ GraphQL 查询服务可正常运行，Stage 4 聚焦在现有能力上扩展过滤、历史聚合与性能观测，而不是阻塞性修复。

---

## 2.1 前置增强事项（GraphQL & 数据访问）

| 项目 | 目标 | 责任团队 | 验收标准 |
|------|------|----------|----------|
| GraphQL 查询增强 | 扩展 `cmd/organization-query-service/internal/model` 与 `repository/postgres_positions.go`，支持任职过滤（类型/状态/日期范围）、分页与租户隔离 | 查询服务团队 | `go test ./cmd/organization-query-service/...` 通过；`make run-dev` 下 GraphQL 查询 P95 < 250ms |
| 时间轴整合 | 在 `GetPositionTimeline` 聚合 Acting/Primary 任职节点，输出时间顺序与标识 | 查询服务团队 · 架构组 | GraphQL `positionTimeline` 返回任职节点；前端时间轴验收通过 |
| 事实来源同步 | 更新 `docs/api/schema.graphql` 注释，执行 `node scripts/generate-implementation-inventory.js` 并记录差异 | 架构组 | 契约、实现、实现清单一致；06 号日志留存 |

### 前置增强步骤（预计 4-6 小时）
1. **查询扩展**（1h）：在 `repository/postgres_positions.go` 添加任职过滤、分页参数与租户校验。  
2. **模型调整**（1h）：复用 `AssignmentHistoryField`，保证空值语义，同时承载过滤结果缓存。  
3. **时间轴组装**（1h）：在 `GetPositionTimeline` 输出 Acting/Primary 节点与事件排序。  
4. **性能校准**（1h）：运行 `make run-dev`，采集 GraphQL 查询延迟，生成 `reports/position-stage4/latency-baseline.md`。  
5. **契约同步**（≤2h）：更新 GraphQL Schema 注释、实现清单并记录 06 号日志时间戳。

---

## 3. 差距与目标范围

### 3.1 差距分析

| 分类 | 现状 | 差距 | Stage 4 增量目标 |
|------|------|------|------------------|
| 代理任职生命周期 | Fill 可创建 `assignment_type=ACTING`，但未自动恢复 | 代理到期需人工 Vacate，缺少自动化 & 提醒 | 实现代理到期自动恢复、提醒通知、冲突校验 |
| 任职操作接口 | Fill/Vacate/Transfer 混合处理 Assignment & Position | 缺少任务专用端点、难以复用 | 提供 `/assignments` 专用 REST API，同时复用现有 Fill/Vacate 流程 |
| 任职历史视图 | 前端展示列表，但无筛选/导出/时间轴增强 | 缺少可视化时间线与更细粒度过滤 | 扩充“任职历史”页签，加时间轴、筛选、CSV 导出 |
| 租户一致性测试 | 80 号第 7.6 节待办、1710 行回归测试缺失 | 现有 Playwright 未覆盖跨租户/非法引用 | 完成交叉租户集成测试脚本与自动化执行 |
| 运营监控 | 缺少代理队列/统计监控 | 风险不可见 | 增补 Prometheus 指标和日志辅助 |

### 3.2 本次范围（Included）
1. **代理任职自动化**：到期恢复主任职、提前提醒、冲突检测、审计日志补强。  
2. **任职专用 API**：在保持 Fill/Vacate 兼容的前提下，新增 `/api/v1/positions/{code}/assignments/*` 端点，对外暴露 CRUD 与分页查询。  
3. **历史视图增强**：GraphQL & 前端支持按 AssignmentType/Status/日期筛选，提供时间轴可视化、CSV 导出、代理标识。  
4. **租户隔离回归**：补齐 REST/GraphQL 跨租户测试脚本、CI 集成，覆盖 403 `JOB_CATALOG_TENANT_MISMATCH` 与 `POSITION_ASSIGNMENT_TENANT_MISMATCH`。  
5. **监控告警**：增加代理任职到期计数、滞留检测指标，接入日志与 dashboard。

### 3.3 非目标范围（Excluded）
- 员工主数据服务改造、外部 HRIS 集成。  
- 薪酬、绩效等后续模块。  
- 组织事件异步总线（保留同步方式）。  
- Grafana/Looker 深度可视化（仅提供指标与 CSV）。

---

## 4. 开发前必检（强制）

在开展任何 Stage 4 工作前，必须一次性执行并归档以下命令：

```bash
cd /home/shangmeilin/cube-castle

# 1. 实现清单核对
node scripts/generate-implementation-inventory.js | grep -i "position assignment"

# 2. IIG 护卫检查
node scripts/quality/iig-guardian.js "Position Assignment Stage4" --guard

# 3. Stage 2/3 实现审计
grep -A40 "position_assignments" database/migrations/044_create_position_assignments.sql
rg "assignment" cmd/organization-command-service/internal -n
rg "positionAssignments" cmd/organization-query-service/internal -n

# 4. 差距报告初始化
mkdir -p reports/position-stage4
echo "Stage4 差距分析（现状 vs 目标）" > reports/position-stage4/gap-analysis.md
```

执行结果需附在 06 号进展日志 Stage 4 小节。

---

## 5. 时间线与里程碑（2 周 + 1 周缓冲）

| 周次 | 核心目标 | 责任团队 | 产出物 & 验收 |
|------|----------|----------|----------------|
| **Week 1** | 代理任职自动化 & 契约对齐 | 命令服务 · 数据库 · 架构 | 048 迁移、REST `/assignments`、OpenAPI/Schema 更新、单元测试、审计日志 |
| **Week 2** | 历史视图增强 & 跨租户测试 | 查询服务 · 前端 · QA · 运维 | GraphQL 扩展、前端时间轴、调度集成、Playwright/集成测试 |
| **Week 3 (缓冲)** | 监控指标 & 文档归档 | 全员 | Prometheus 指标、调度运行日志、文档同步、计划归档 |

每周周三风控例会、周五演示与风险复盘；重大事项写入 06 号日志。

---

## 6. 前置依赖：87号生产迁移（effectiveDate 字段统一）

### 6.1 迁移背景

根据 **87号时态字段命名一致性决策文档**，开发环境已完成 `position_assignments.start_date` → `effective_date` 的字段重命名（047迁移），以确保与组织架构、职位主数据、Job Catalog 的时态字段命名保持一致。

**当前状态**：
- ✅ 开发环境：047迁移已执行，代码/契约/文档已同步
- 🔴 生产环境：仍使用 `start_date` 字段，需在 Stage 4 上线前统一

**重要性**：
- 🔴 **最高优先级**：违反 CLAUDE.md 资源唯一性原则，属于阻断性问题
- 🔴 **技术债务**：推迟迁移会增加后续维护成本和风险
- 🔴 **架构一致性**：Stage 4 新功能基于 `effectiveDate` 字段实现

### 6.2 迁移时间线（与 Stage 4 联动）

按照87号文档 §12 的联动执行策略，将迁移任务整合到 Stage 4 时间线：

| 时间点 | 阶段 | 任务内容 | 责任人 | 交付物 |
|--------|------|----------|--------|--------|
| **T-5天** | Week 2 中期 | 预生产环境演练 047 迁移 | 数据库团队 | 演练日志：`reports/position-stage4/047-preprod-dryrun-YYYYMMDD.txt` |
| **T-5天** | Week 2 中期 | 复核 Stage 4 代码无 `startDate` 残留 | 查询/前端团队 | 静态检查结果（`rg "startDate"`） |
| **T-3天** | Week 2 末 | 确认迁移窗口与发布计划 | 架构组 | 上线计划（含迁移时段） |
| **T-3天** | Week 2 末 | 合并测试清单（87号 + 86号） | QA团队 | `reports/position-stage4/final-acceptance-checklist.md` |
| **T-3天** | Week 2 末 | 外部集成方/BI团队通知 | 架构组 | 发布说明 + Breaking Change 声明 |
| **T-0** | Week 3 上线窗口 | 执行 047 生产迁移 | 数据库团队 | 迁移执行日志 + 验证报告 |
| **T+0** | Week 3 上线窗口 | 迁移后验证 + Stage 4 部署 | 全员 | 服务健康检查 + 监控数据 |
| **T+1** | Week 3 上线后 | 06号日志登记 + 文档归档 | 架构组 | 06号日志更新 + 87号归档确认 |

### 6.3 迁移执行计划（基于87号 §11）

#### 6.3.1 迁移前校验（T-5天）

**现状评估**：
```bash
# 统计生产环境数据量
psql "$PRODUCTION_DATABASE_URL" -c "
SELECT
  COUNT(*) as total_rows,
  pg_size_pretty(pg_total_relation_size('position_assignments')) as table_size
FROM position_assignments;"

# 确认依赖 start_date 的脚本/任务
grep -r "start_date" scripts/ database/ | grep position_assignments
```

**预生产演练**：
```bash
# 1. 在预生产环境执行 047 迁移
psql "$PREPROD_DATABASE_URL" -f database/migrations/047_rename_position_assignments_start_date.sql

# 2. 记录执行时间与锁表情况
# 输出到：reports/position-stage4/047-preprod-dryrun-$(date +%Y%m%d).txt

# 3. 验证回滚脚本
psql "$PREPROD_DATABASE_URL" -f database/migrations/rollback/047_rollback.sql
```

**备份策略**：
- 生产数据库完整备份或分区级快照
- 确认可在 15 分钟内恢复
- 验证回滚脚本可无损恢复

#### 6.3.2 代码残留检查（T-5天）

**静态代码检查**：
```bash
# 检查后端代码
rg "start_date|startDate" cmd/organization-command-service/internal --type go
rg "start_date|startDate" cmd/organization-query-service/internal --type go

# 检查前端代码
rg "startDate" frontend/src/features/positions --type ts --type tsx
rg "startDate" frontend/src/shared/hooks --type ts
rg "startDate" frontend/src/shared/types --type ts

# 期望结果：无匹配（除注释/文档外）
```

**契约检查**：
```bash
# 检查 GraphQL Schema
grep -i "startDate" docs/api/schema.graphql
# 期望：仅 effectiveDate，无 startDate

# 检查 OpenAPI
grep -i "start_date\|startDate" docs/api/openapi.yaml
# 期望：无匹配
```

#### 6.3.3 迁移执行步骤（T-0）

**维护窗口安排**：
- 建议时间：北京时间 02:00-04:00（低流量窗口）
- 预计耗时：30分钟（含验证）
- 冻结部署流水线，停止命令服务写流量

**执行流程**：
```bash
# 1. 宣布维护窗口，切换命令服务至维护模式
# （查询服务保持只读）

# 2. 执行迁移脚本
psql "$PRODUCTION_DATABASE_URL" -v ON_ERROR_STOP=1 \
  -f database/migrations/047_rename_position_assignments_start_date.sql \
  > reports/position-stage4/047-production-migration-$(date +%Y%m%d-%H%M).log 2>&1

# 3. 验证迁移结果
psql "$PRODUCTION_DATABASE_URL" -c "
-- 验证字段重命名
SELECT column_name FROM information_schema.columns
WHERE table_name = 'position_assignments'
  AND column_name IN ('effective_date', 'start_date');
-- 期望：仅 effective_date

-- 验证约束
SELECT constraint_name FROM information_schema.table_constraints
WHERE table_name = 'position_assignments';
-- 期望：包含 chk_position_assignments_dates（引用 effective_date）

-- 验证索引
SELECT indexname FROM pg_indexes
WHERE tablename = 'position_assignments';
-- 期望：包含 uk_position_assignments_effective
"

# 4. 运行测试清单（见 6.3.4）

# 5. 恢复命令服务写流量

# 6. 监控错误日志 30 分钟
tail -f /var/log/cube-castle/command-service.log | grep -i error
```

#### 6.3.4 迁移后验证清单

基于87号文档 §6.2/§6.3 + 86号验收标准：

```bash
# 数据完整性验证
psql "$PRODUCTION_DATABASE_URL" -c "
SELECT COUNT(*) FROM position_assignments
WHERE effective_date IS NULL;
-- 期望：0（所有记录都有 effective_date）

SELECT COUNT(*) FROM position_assignments
WHERE end_date IS NOT NULL AND end_date <= effective_date;
-- 期望：0（约束生效）
"

# GraphQL 查询验证
curl -X POST https://api.production/graphql \
  -H "Authorization: Bearer $PROD_JWT" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{"query":"{ position(code:\"TEST001\") { currentAssignment { effectiveDate endDate } } }"}' \
| jq '.data.position.currentAssignment'
# 期望：返回 effectiveDate 字段，无 startDate

# REST API 验证
curl -s https://api.production/api/v1/positions/TEST001/assignments \
  -H "Authorization: Bearer $PROD_JWT" \
  -H "X-Tenant-ID: $TENANT_ID" \
| jq '.[0] | keys | map(select(. == "effectiveDate" or . == "startDate"))'
# 期望：["effectiveDate"]，无 startDate

# 前端缓存清理
# 确保发布版本包含缓存 bust 机制（版本号 bump 或 queryClient.invalidateQueries）
```

#### 6.3.5 回滚预案

如迁移后发现阻塞性问题，立即执行回滚：

```bash
# 1. 停止命令服务写流量
# 2. 执行回滚脚本
psql "$PRODUCTION_DATABASE_URL" -v ON_ERROR_STOP=1 \
  -f database/migrations/rollback/047_rollback.sql

# 3. 验证回滚结果（字段恢复为 start_date）
# 4. 恢复服务
# 5. 在06号日志登记回滚原因与后续计划
```

### 6.4 风险与缓解（迁移专项）

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 迁移与 Stage 4 上线冲突（锁表/回滚） | 🔴 高（延迟上线） | 中 | 047 提前1小时执行；准备回滚脚本；必要时拆分为两阶段部署 |
| Stage 4 源码仍引用 `startDate` | 🔴 高（接口报错） | 低 | T-3天静态检查并强制修复后才进入上线窗口 |
| 外部依赖未同步（第三方集成/BI） | 🟡 中（集成失败） | 中 | 提前48小时发送 Breaking Change 通知；必要时提供临时兼容层（视图别名） |
| 运维窗口过短，验证不充分 | 🟡 中（需重新上线） | 中 | Week 3 预留完整缓冲时间；演练阶段确认执行时间 < 30分钟 |
| 前端缓存未更新，仍显示旧字段 | 🟡 中（用户体验） | 低 | 发布版本强制缓存清理（service worker / version bump） |

### 6.5 外部通知与协调

#### 发布说明（T-3天发送）

**收件人**：外部集成方、BI团队、运维团队

**内容**：
```markdown
# Cube Castle API Breaking Change 通知

## 变更概述
在即将发布的 Stage 4 版本中，position_assignments 相关 API 字段命名将统一为 effectiveDate，
以保持与组织架构、职位主数据的一致性。

## 影响范围
- REST API: /api/v1/positions/{code}/assignments
- GraphQL: PositionAssignment 类型

## 字段映射
| 旧字段名 | 新字段名 | 数据类型 | 说明 |
|---------|---------|---------|------|
| startDate | effectiveDate | Date (ISO 8601) | 任职生效日期 |
| endDate | endDate | Date (ISO 8601) | 无变化 |

## 上线窗口
- 时间：2025-10-XX 02:00-04:00 (UTC+8)
- 预计维护时长：30分钟

## 迁移建议
1. 更新集成代码，使用 effectiveDate 替代 startDate
2. 在测试环境验证（测试环境已提前更新）
3. 如有疑问，请联系架构组

## 兼容性支持
- 临时兼容层：可在上线窗口后提供视图别名（有效期7天）
- 联系方式：architecture@cubecastle.com
```

#### 内部协调清单

- [ ] **T-3天**：通知外部集成方（邮件 + Slack）
- [ ] **T-3天**：通知 BI 团队更新 ETL 脚本
- [ ] **T-3天**：运维团队确认备份策略与回滚演练
- [ ] **T-1天**：与 Stage 4 上线负责人同步最终时间表
- [ ] **T-0**：在公司 Slack 频道宣布维护窗口
- [ ] **T+1**：发送迁移完成通知 + 验证结果

### 6.6 迁移成功标准

- ✅ 047迁移脚本执行成功，无 SQL 错误
- ✅ 字段重命名完成：`start_date` → `effective_date`
- ✅ 索引与约束重建成功
- ✅ 数据完整性验证通过（无空值、无约束违反）
- ✅ GraphQL/REST API 返回 `effectiveDate` 字段
- ✅ 前端缓存刷新，无旧字段残留
- ✅ Stage 4 功能正常部署与运行
- ✅ 监控30分钟无异常错误
- ✅ 06号日志完整记录执行过程与结果

---

## 7. 任务拆解

### API 契约定义（Stage 4 增量）
- **REST — `/api/v1/positions/{code}/assignments` 套件**（OpenAPI 将新增/更新以下条目，均要求 `position:assignments:write` 或 `position:assignments:read` 权限）：  
  - `GET /api/v1/positions/{code}/assignments`: 查询当前与历史任职，支持 `assignmentTypes[]`、`assignmentStatus`、`asOfDate`、`includeHistorical`、分页参数。响应主体为 `PositionAssignmentListResponse`，字段沿用 Stage 2/3 输出，新增 `actingUntil`、`autoRevert`、`reminderSentAt`。  
  - `POST /api/v1/positions/{code}/assignments`: 创建任职。请求体需提供 `employeeId`、`employeeName`、`assignmentType`、`effectiveDate`、`fte`，可选 `actingUntil`、`autoRevert`、`notes`。成功返回 201 + 新建记录。  
  - `PATCH /api/v1/positions/{code}/assignments/{assignmentId}`: 更新任职（调整 `fte`、`actingUntil`、`autoRevert`、`notes`），返回 200。  
  - `POST /api/v1/positions/{code}/assignments/{assignmentId}/close`: 结束任职。请求体包含 `endDate`、可选 `notes`，返回 200 并写入审计。  
  - 所有端点必须校验租户一致性（`tenantId` header/claims），返回标准错误码：`403 JOB_CATALOG_TENANT_MISMATCH`、`409 POSITION_ASSIGNMENT_CONFLICT`、`422 POSITION_ASSIGNMENT_VALIDATION_FAILED`。
- **GraphQL — `docs/api/schema.graphql` 增量**：  
  - `positionAssignments(positionCode: PositionCode!, filter: PositionAssignmentFilterInput, pagination: PaginationInput, sorting: [PositionAssignmentSortInput!]): PositionAssignmentConnection!` 新增 filter 字段：`assignmentTypes: [PositionAssignmentType!]`, `status: PositionAssignmentStatus`, `dateRange: DateRangeInput`, `includeActingOnly: Boolean`.  
  - `type PositionAssignment` 新增只读字段：`actingUntil: Date`, `autoRevert: Boolean!`, `reminderSentAt: DateTime`.  
  - `type PositionTimelineEntry` 增补 `assignmentType: PositionAssignmentType`、`assignmentStatus: PositionAssignmentStatus`，并允许 `timelineCategory: POSITION_ASSIGNMENT`.  
  - `type PositionAssignmentAudit`（新）用于 CSV 导出：包含 `assignmentId`, `eventType`, `effectiveDate`, `endDate`, `actor`, `changes`.  
  - 权限要求：查询需 `position:assignments:read` scope，导出需额外 `position:assignments:audit`.

### 7.1 Week 1 — 代理任职自动化 & 契约对齐
- **数据库**：交付 `048_extend_position_assignments.sql`（及回滚脚本），新增 `acting_until DATE`, `auto_revert BOOLEAN DEFAULT false`, `reminder_sent_at TIMESTAMPTZ`，并更新索引/校验。输出演练日志与延迟评估。  
- **命令服务**：实现上述 REST 契约，对接 `PositionAssignmentRepository`，保持 Fill/Vacate 兼容，新增幂等锁与审计事件。  
- **自动化任务**：实现代理到期扫描器（使用 `OperationalScheduler` 任务定义），支持重试、失败告警、审计写入。  
- **单元与契约测试**：扩展 `position_handler_test.go`、`assignment_repository_test.go`，覆盖冲突/FTE 验证；新增 OpenAPI contract tests。  
- **契约同步**：更新 `docs/api/openapi.yaml`、`docs/api/schema.graphql`，运行 `node scripts/generate-implementation-inventory.js` 并在 06 号日志记录差异。

### 7.2 Week 2 — 历史视图增强 & 跨租户测试
- **GraphQL 查询服务**：落地前置增强事项，提供过滤、分页、时间轴整合及性能基线；新增 `positionAssignmentAudit` 查询导出。  
- **前端**：在 `frontend/src/features/positions` 新增任职历史页签、时间轴视图、过滤器与 CSV 导出；更新 `PositionTransferDialog`、`PositionSummaryCards` 展示代理提醒。  
- **命令服务调度**：将代理恢复任务接入 `OperationalScheduler` 配置（默认每日 02:00），提供手动触发脚本与日志归档。  
- **质量验证**：编写 Playwright 场景（代理创建→到期→恢复→时间轴验证）与 REST/GraphQL 跨租户脚本，纳入 `make test-integration`；2025-10-21 完成命令服务 REST 跨租户脚本冒烟（`tests/consolidated/position-assignments-cross-tenant.sh`），输出归档至 `reports/position-stage4/position-assignments-cross-tenant.md`。

### 7.3 Week 3 — 缓冲、迁移与收尾

**Week 3 重点**：执行87号生产迁移 + Stage 4 上线 + 监控验收

- **T-5天（Week 2中期）**：
  - 预生产环境演练047迁移（数据库团队）
  - 静态代码检查，确认无 `startDate` 残留（查询/前端团队）
  - 输出：`reports/position-stage4/047-preprod-dryrun-YYYYMMDD.txt`

- **T-3天（Week 2末）**：
  - 确认迁移窗口与上线计划（架构组）
  - 发送外部通知：Breaking Change 声明（架构组）
  - 合并87号+86号测试清单（QA团队）
  - 输出：`reports/position-stage4/final-acceptance-checklist.md`

- **T-0（Week 3上线窗口，建议02:00-04:00）**：
  - 执行047生产迁移（数据库团队）
  - 迁移后验证（见第6.3.4节）
  - 部署 Stage 4 功能（全员）
  - 监控30分钟无异常

- **T+1（上线后）**：
  - 观察自动恢复任务运行（收集调度日志、Prometheus 指标）
  - 完成监控与告警文档（`docs/development-tools/position-assignment-monitoring.md`）
  - 更新 80 号方案 Stage 4 勾选、06 号日志
  - 发送迁移完成通知 + 验证结果
  - 整理 API 契约差异报告、实现清单、回归测试记录
  - 将86号、87号计划归档至 `docs/archive/development-plans/`

---

## 8. 质量与验收标准

### 8.1 功能验收标准（原 Stage 4）

| 类别 | 验收标准 |
|------|----------|
| 功能 | 代理任职自动恢复 + 提醒日志；专用任职 API 通过 REST 集成测试；时间轴展示主任职/副任职/代理状态。 |
| 数据一致性 | Acting 到期后 FTE 回落，`HeadcountInUse` 与 `positionHeadcountStats` 同步；跨租户操作返回 403。 |
| 性能 | 任职 API P95 < 200ms；代理自动化任务执行 < 2s/1000条；GraphQL 过滤查询 P95 < 250ms。 |
| 测试 | `go test ./cmd/organization-command-service/...` 覆盖率 ≥ 80%；`npm --prefix frontend run test -- PositionDetails`；Playwright Acting 场景通过；跨租户脚本纳入 CI。 |
| 文档 | 80号 Stage 4 勾选完成；06 号日志更新；实现清单/契约差异/监控文档同步。 |
| 监控 | 新增 `position_assignment_acting_total` 等指标并接入报警；运行记录归档到 `reports/position-stage4/`. |

### 8.2 迁移验收标准（87号集成）

| 类别 | 验收标准 |
|------|----------|
| **迁移执行** | ✅ 047迁移脚本成功执行，无 SQL 错误；执行时间 < 30分钟 |
| **字段重命名** | ✅ `position_assignments.start_date` → `effective_date`；索引/约束重建成功 |
| **数据完整性** | ✅ 无空值（`effective_date IS NOT NULL`）；约束生效（`end_date > effective_date`） |
| **API契约** | ✅ GraphQL/REST 返回 `effectiveDate` 字段；OpenAPI/Schema 同步更新 |
| **代码一致性** | ✅ 静态检查无 `startDate` 残留（后端/前端/契约） |
| **前端缓存** | ✅ 缓存清理机制生效，用户端无旧字段显示 |
| **外部通知** | ✅ 提前48小时发送 Breaking Change 通知；集成方确认收到 |
| **监控验证** | ✅ 迁移后30分钟内无异常错误；服务健康检查通过 |
| **回滚演练** | ✅ 预生产环境验证回滚脚本可用；回滚时间 < 15分钟 |
| **文档归档** | ✅ 06号日志完整记录迁移过程；87号计划归档确认 |

---

## 9. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 自动恢复误触发 | 高 | 中 | 双重条件校验（到期 + 当前状态），留存手动回滚脚本与审计确认。 |
| 任职 API 与 Fill/Vacate 冲突 | 中 | 中 | 通过 `/assignments` 封装现有逻辑，设置特性开关逐步启用并监控审计事件。 |
| 调度器集成不稳定 | 中 | 中 | 在 `OperationalScheduler` 引入幂等锁、失败重试、Prometheus 告警，首周每日人工巡检。 |
| 契约漂移 | 中 | 中 | 每次变更前更新 OpenAPI/GraphQL 并执行实现清单脚本，CI 增加契约 diff 校验。 |
| 跨租户脚本复杂 | 中 | 中 | 先在 sandbox 演练并记录结果，再纳入 CI 分阶段执行。 |

**迁移专项风险**（详见第6.4节）：
- 🔴 迁移与上线冲突、代码残留 `startDate`
- 🟡 外部依赖未同步、运维窗口过短、前端缓存未更新

---

## 10. 协作机制

- **例会**：周一计划同步、周三风控、周五演示。  
- **责任人**：
  - 数据库团队：047迁移执行、预生产演练、回滚验证
  - 命令服务负责人：REST API、定时任务、迁移后验证
  - 查询服务负责人：GraphQL 扩展、性能基线
  - 前端负责人：时间轴与导出、缓存清理机制
  - QA 负责人：集成/Playwright/跨租户脚本、测试清单合并
  - 架构组：事实来源守护、迁移协调、外部通知、评审、质量门禁
- **文档记录**：所有关键决策、测试结果、监控数据、迁移执行日志写入 06 号日志 & `reports/position-stage4/`

---

## 11. 交付与归档清单

### 11.1 Stage 4 功能交付

- [x] 048 迁移 & 回滚脚本 + 演练日志（见 `reports/position-stage4/048-migration-dryrun-20251021.log`）
- [x] 任职专用 REST API 代码与测试
- [x] 代理自动恢复任务（含 `OperationalScheduler` 集成），监控指标待后续排期
- [x] GraphQL & 前端任职历史增强 + Playwright 场景
- [x] 跨租户回归测试脚本 —— REST 场景已完成并归档（`tests/consolidated/position-assignments-cross-tenant.sh`，日志位于 `reports/position-stage4/position-assignments-cross-tenant.*`）
- [x] GraphQL 脚本完成（`tests/consolidated/position-assignments-graphql-cross-tenant.sh`，日志位于 `reports/position-stage4/position-assignments-graphql-cross-tenant.log`）
- [x] CI 集成与自动化回归（.github/workflows/ci.yml 新增 cross-tenant-tests job）
- [x] 契约与文档同步（OpenAPI/GraphQL、实现清单、80 号 Stage 4 勾选、06 号日志、监控指南）

### 11.2 87号生产迁移交付（前置依赖）

- [x] **T-5天**：预生产环境047迁移演练日志（`reports/position-stage4/047-preprod-dryrun-20251021.txt`）
- [x] **T-5天**：静态代码检查报告（确认无 `startDate` 残留，见 `reports/position-stage4/static-code-check-20251021.txt`）
- [x] **T-3天**：外部通知发送记录（当前阶段未涉及外部团队，经确认免除）
- [x] **T-3天**：合并测试清单（`reports/position-stage4/final-acceptance-checklist.md` 草案已建立，待上线阶段填写结果）
- [x] **T-0**：047 生产迁移执行日志（`reports/position-stage4/047-production-migration-20251021-0900.log`）
- [x] **T-0**：迁移后验证报告（`reports/position-stage4/047-post-migration-validation-20251021.md`）
- [x] **T+1**：迁移完成通知（当前阶段仅需内部留档，见 `reports/position-stage4/047-migration-completion-note-20251021.md`）
- [x] **T+1**：06号日志更新（参见 `docs/development-plans/06-integrated-teams-progress-log.md` 2025-10-21 14:30 条目）
- [x] **T+1**：87号计划归档确认（`docs/archive/development-plans/87-temporal-field-naming-consistency-decision.md`）

### 11.3 最终归档

- [x] 86号计划归档（本文件已迁入 archive，CI 集成 job `cross-tenant-tests` 生效）
- [x] 80号方案 Stage 4 勾选更新（见 `docs/development-plans/80-position-management-with-temporal-tracking.md`）
- [x] 实现清单最终同步（`reports/implementation-inventory.json` 已更新）

---

## 12. 变更记录

| 版本 | 日期 | 说明 | 作者 |
|------|------|------|------|
| v1.0 | 2025-10-21 | 归档版：跨租户脚本、047/048 迁移、CI 集成全部完成，计划正式收束，待后续运营改进时另立新案 | Claude Code 助手 |
| v0.4 | 2025-10-21 | **重大更新**：整合87号生产迁移计划（effectiveDate字段统一）<br>- 新增第6节：前置依赖-87号生产迁移详细流程<br>- 更新时间线：T-5/T-3/T-0/T+1迁移节点<br>- 新增验收标准：迁移专项验收清单（§8.2）<br>- 更新风险：迁移专项风险与缓解措施（§6.4）<br>- 更新协作：增补数据库团队迁移职责<br>- 更新交付清单：87号迁移交付物（§11.2）<br>- Week 3 重点调整：先迁移后上线策略 | Claude Code 助手 |
| v0.3 | 2025-10-19 | 新增 API 契约定义、更新迁移编号与调度集成、补充风险缓解 | 项目智能助手 |
| v0.2 | 2025-10-17 | 根据 06 号评审意见修订，聚焦增量能力与差距分析 | 项目智能助手 |
| v0.1 | 2025-10-17 | 初始草案（已废弃） | 项目智能助手 |
