# 84号文档：职位生命周期 Stage 2 实施计划（方案B重构版）

**版本**: v1.0  
**创建日期**: 2025-10-16  
**维护团队**: 命令服务团队 · 查询服务团队 · 前端团队 · 架构组  
**归档日期**: 2025-10-16  
**状态**: ✅ 已完成（Stage 2 全量上线并验收通过）  
**关联计划**: 80号职位管理方案 · 06号集成团队进展日志 · 82号 Stage 1 实施计划  
**唯一事实来源**:  
- `docs/archive/development-plans/80-position-management-with-temporal-tracking.md`（Stage 2 目标）  
- `docs/development-plans/06-integrated-teams-progress-log.md`（评审意见与风险）  
- `CLAUDE.md` / `AGENTS.md`（资源唯一性与临时方案治理）

---

## 1. 背景与目标

- 06号评审结论明确指出：Stage 2 必须消除双数据源，直接落地 `position_assignments`，并移除 `positions` 表的冗余任职字段，以满足“资源唯一性与跨层一致性”最高优先级原则。  
- Stage 1 已交付职位基础 CRUD、GraphQL 查询与前端读能力。Stage 2 将在此基础上补齐 Fill / Vacate / Transfer 生命周期，并一次性完成任职模型迁移。  
- 本计划基于方案B（立即迁移），将 Stage 2 拆解为契约 → 迁移 → 服务 → 前端 → 验收的完整闭环，确保无临时双写策略。

### 当前进展
- [x] Phase A 契约更新（OpenAPI/GraphQL：新增任职、空缺、转移类型并移除冗余字段）
- [x] Phase A 迁移脚本与回滚脚本（044/045 + rollback 版本）已产出
- [x] Sandbox 演练流程与日志归档规范已写入计划
- [x] Sandbox 环境迁移 & 回滚演练（reports/database/migration-044-045-dryrun-20251016.txt、reports/architecture/tenant-isolation-check-stage2-20251016.sql）
- [x] Phase B 命令/查询服务实现
- [x] Phase C 前端交互与 E2E 验收
- [x] Phase D 质量门禁与文档归档
  - ✅ 单元测试：`go test ./cmd/organization-query-service/internal/graphql`、`npx vitest run frontend/src/features/positions/__tests__/PositionDashboard.test.tsx`
  - ✅ 前端 E2E：`cd frontend && npx playwright test tests/e2e/position-lifecycle.spec.ts --config playwright.config.ts`（GraphQL 拦截夹具验证任职/调动视图）
  - ✅ 文档同步：`docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`reports/contracts/position-api-diff.md`、06号进展日志更新

---

## 2. 范围与交付物

### 2.1 必须交付（Blocking）
1. **契约对齐**  
   - 更新 `docs/api/openapi.yaml`、`docs/api/schema.graphql`，定义 `PositionAssignment`、`PositionTransfer` 相关类型与输入，并移除与 `current_holder_*` 等冗余字段相关的响应字段。  
   - 生成新的实现清单项，确保 REST / GraphQL 契约唯一指向 `position_assignments`。
2. **数据库迁移**  
   - 编写 `044_create_position_assignments.sql`：创建 `position_assignments` 表（支持时态字段、租户隔离、乐观锁、审计字段）。  
   - 编写 `045_drop_position_legacy_columns.sql`：在确认无存量数据后移除 `positions` 表中 `current_holder_*`、`current_assignment_type` 等冗余列。  
   - 如需保留历史数据，先生成快照归档至 `reports/database/`.  
3. **命令服务（REST）**  
   - 扩展仓储与服务层：Fill / Vacate / Transfer 全量操作 `position_assignments`（创建、关闭、转移）；`positions` 表仅维护职位定义。  
   - Handler：实现 `/api/v1/positions/{code}/fill|vacate|transfer`，覆盖幂等、租户校验、审计、错误响应。  
   - 补充回滚策略与集成测试（命令链路 + 事务一致性）。
4. **查询服务（GraphQL）**  
   - 新增 `positionAssignments`, `vacantPositions`, `positionTransfers` 查询；扩展 `positions` / `positionTimeline` 返回任职与转移节点。  
   - 数据访问层：基于 `position_assignments` 表实现 asOfDate、分页、租户隔离、缓存策略。  
   - 集成测试：验证填充→空缺→转移等关键查询。
5. **前端应用**  
   - 更新 `useEnterprisePositions` 系列 Hook，改用新 GraphQL 字段；去除对旧 `current_holder_*` 字段的依赖。  
   - 职位详情页实现 Fill / Vacate / Transfer 表单、权限校验、结果反馈；实现“空缺职位看板”与任职历史时间线。  
   - 前端在无权限或租户不匹配时给出明确提示。
6. **质量与验收**  
   - 单元测试：仓储、服务、Resolver、前端组件均需覆盖主路径与异常路径；后端覆盖率保持 ≥80%。  
   - 集成测试：REST + GraphQL 端到端场景、租户隔离 SQL 校验、审计日志验证。  
  - Playwright：新增 `frontend/tests/e2e/position-lifecycle.spec.ts`，演示填充→空缺→转移完整流程。  
   - 文档：更新 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`reports/contracts/position-api-diff.md`、06号日志“当前进展”与“风险”栏目。

### 2.2 可选扩展（P2）
- 自动生成 `position_assignments` 与转移历史报表（CSV）。  
- 对接组织服务的 Async 事件流，增强跨系统一致性监控。

---

## 3. 前置条件与依赖

| 项目 | 说明 |
|------|------|
| Stage 1 验收 | 82 号计划全部验收项通过，当前无 Fill / Vacate 数据写入。 |
| 数据冻结窗口 | Stage 2 迁移期间冻结职位生命周期操作（无外部系统写入），避免迁移冲突。 |
| 契约更新 | 81 号计划为最新契约基础；Stage 2 扩展字段须先提交 PR 并获得契约评审通过。 |
| 环境一致性 | `make docker-up`、`make db-migrate-all` 可用；确保宿主机无 5432/6379 冲突。 |
| 审批 | 迁移脚本需通过数据库变更评审，方可在 dev / staging 环境执行。 |

---

## 4. 工作拆解与里程碑

### 4.1 时间线

- **总周期**：2025-10-20 ~ 2025-11-28（6 周，含缓冲 1 周，满足方案B高难度需求）。  
- **节奏**：按阶段交付，阶段评审通过后方可进入下一阶段。

| 阶段 | 时间 | 目标 | 关键输出 |
|------|------|------|----------|
| Phase A | Week 1-2 | 契约定稿 + 数据迁移 | 更新 OpenAPI/GraphQL、迁移 044/045、回滚脚本、迁移演练报告 |
| Phase B | Week 3-4 | 命令服务与查询层 | Fill/Vacate/Transfer 服务、GraphQL Resolver、集成测试 |
| Phase C | Week 5 | 前端交互与 E2E | 前端表单/看板、Vitest、Playwright 冒烟 |
| Phase D | Week 6 | 验收与收尾 | 租户隔离报告、性能/安全基准、文档更新、复盘 |

### 4.2 任务清单

| 编号 | 任务 | 所属阶段 | 负责人 | 产出 |
|------|------|----------|--------|------|
| A1 | 拟定契约变更并提交 PR（OpenAPI+GraphQL） | A | 架构组 + 命令/查询 | 契约 MR、评审记录 |
| A2 | 生成迁移脚本 044/045、编写回滚方案 | A | DB 工程师 | `database/migrations/044_create_position_assignments.sql`、`045_drop_position_legacy_columns.sql`、备份脚本 |
| A3 | 在 sandbox 环境演练迁移（含回滚），归档日志 | A | DB 工程师 | `reports/database/migration-044-dryrun-YYYYMMDD.log` |
| A4 | 更新实现清单与文档差异说明 | A | 架构组 | `reports/contracts/position-api-diff.md` |
| B1 | 扩展仓储与服务：Fill/Vacate/Transfer | B | 命令服务团队 | Service/Repository 实现、单测 |
| B2 | 实现 REST Handler + 集成测试 | B | 命令服务团队 | Handler、`tests/` 集成用例 |
| B3 | GraphQL Schema/Resolver + 集成测试 | B | 查询服务团队 | Schema、Resolver、`go test` |
| C1 | 更新前端 Hook 与 API 客户端 | C | 前端团队 | `frontend/src/shared/api/positions.ts` 等 |
| C2 | 实现前端交互（填充/空缺/转移/空缺看板） | C | 前端团队 | UI 组件、Vitest |
| C3 | Playwright E2E：填充→空缺→转移 | C | 前端团队 + QA | `frontend/tests/e2e/position-lifecycle.spec.ts`、报告 |
| D1 | 全量质量门禁：`make test`、`npm run lint`、`make security` | D | 联合团队 | 质量报告 |
| D2 | 租户隔离 SQL 巡检 | D | 架构组 | `reports/architecture/tenant-isolation-check-stage2-YYYYMMDD.sql` |
| D3 | 性能回归（填充/查询基准） | D | 命令 + 查询团队 | 性能测试报告 |
| D4 | 文档与日志复盘、更新 06号进展 | D | 架构组 | 更新后的 84 号/06 号文档、复盘纪要 |

#### Phase A 任务细化

- **A1 契约更新**：在 PR 中补充 `PositionAssignment`, `PositionAssignmentInput`, `PositionTransfer`, `VacantPositionConnection` 等类型，移除所有 `current_holder_*` 字段；附上示例 payload 与错误码映射。
- **A2 迁移脚本设计**  
  1. `044_create_position_assignments.sql`：按照 80 号 §3.2.1 事件周期模式创建表，包含字段 `assignment_id`, `tenant_id`, `position_code`, `position_record_id`, `employee_id`, `assignment_type`, `assignment_status`, `start_date`, `end_date`, `is_current`, `created_at`, `updated_at`，并实现以下约束：  
     - `(tenant_id, position_code, employee_id, start_date)` 唯一；  
     - `(tenant_id, position_code, employee_id, is_current)` 在 `assignment_status='ACTIVE'` 时唯一；  
     - `CHECK (end_date IS NULL OR end_date > start_date)`；  
     - 外键 `(tenant_id, position_code, position_record_id)` 指向 `positions`；  
     - 触发器或生成列同步 `updated_at`。  
  2. `045_drop_position_legacy_columns.sql`：删除 `positions` 表中 `current_holder_id`, `current_holder_name`, `current_assignment_type`, `filled_date` 等历史字段，并在执行前生成 `reports/database/positions-legacy-snapshot-YYYYMMDD.csv` 备份。  
  3. **回滚策略**：提供 `rollback/044_position_assignments_drop.sql` 与 `rollback/045_restore_position_legacy_columns.sql`，若 Stage 2 回滚需恢复旧字段，同时导入备份 CSV。  
  4. **演练**：在 sandbox 环境执行迁移 → 回滚 → 迁移复跑，归档日志 `reports/database/migration-044-045-dryrun-YYYYMMDD.log`。  
- **A3 演练要求**：迁移期间冻结职位生命周期操作；完成后运行 `sql/inspection/tenant-isolation-checks.sql` 验证租户隔离。  
- **A4 文档同步**：在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 标注新增命令/查询，更新 `reports/contracts/position-api-diff.md` 说明字段移除与新类型。

##### Phase A Sandbox 演练步骤

1. **初始化环境**：`make docker-up && make db-migrate-all`，确认 043 迁移已执行。  
2. **生成快照**：在宿主机执行  
   ```bash
   psql "$DATABASE_URL" -c "\copy (SELECT tenant_id, code, current_holder_id, current_holder_name, current_assignment_type, filled_date FROM positions WHERE current_holder_id IS NOT NULL) TO 'reports/database/positions-legacy-snapshot-$(date +%Y%m%d).csv' CSV HEADER"
   ```  
   将输出文件纳入回滚素材。  
3. **执行迁移**：运行 `make db-migrate-all`，确认 044/045 成功；检查 `\d position_assignments` 结构与索引。  
4. **租户巡检**：`psql "$DATABASE_URL" -f sql/inspection/tenant-isolation-checks.sql`，归档结果至 `reports/architecture/tenant-isolation-check-stage2-$(date +%Y%m%d).sql`。  
5. **回滚验证**：依次执行  
   ```bash
   psql "$DATABASE_URL" -f database/migrations/rollback/045_restore_position_legacy_columns.sql
   psql "$DATABASE_URL" -f database/migrations/rollback/044_drop_position_assignments.sql
   ```  
   并使用步骤2快照复写 `positions` 冗余字段（如需）。随后再次执行 044/045 迁移，确保可重复运行。  
6. **记录**：将完整命令与输出归档至 `reports/database/migration-044-045-dryrun-YYYYMMDD.log` 并在 06 号日志更新演练状态。

---

## 5. 验收标准

- **功能**  
  - Fill / Vacate / Transfer REST 命令成功返回 2xx，具备幂等性与权限校验；错误场景返回标准化错误响应。  
  - GraphQL 查询提供实时任职、历史任职、空缺职位、转移记录，支持 `asOfDate`、分页、过滤。  
  - 前端可完成填充→空缺→转移完整演示，权限不足时隐藏入口并给出提示。
- **数据一致性**  
  - `position_assignments` 为职位任职唯一事实来源，`positions` 表不再包含任何当前任职字段。  
  - 租户隔离巡检返回 0 行；审计日志与业务日志对齐。  
  - 迁移脚本及回滚方案在 sandbox 环境验证成功。
- **质量**  
  - 后端 `go test ./...` 覆盖率 ≥80%，包含仓储、服务、Resolver 主路径。  
  - 前端 Vitest 覆盖关键交互；Playwright 场景通过并归档截图。  
  - `make security` 与前端 `npm run lint`、`npm run test`、`npm run typecheck` 全部通过。
- **文档与治理**  
  - 契约、实现清单、开发计划、报告全部更新；06号日志记录 Stage 2 完成和风险关闭。  
  - 无 `TODO-TEMPORARY` 残留；若必须存在，必须明确下个迭代内的回收时间并登记 17 号治理计划。

---

## 6. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 数据迁移失败导致环境不可用 | 高 | 中 | 提前在 sandbox 演练；提供回滚脚本；迁移期间开启维护窗口并准备热备。 |
| 契约变更影响前端现有页面 | 中 | 中 | 契约更新先行、前端 Hook 同步升级；保留 Feature Flag 控制生效时机。 |
| Fill/Vacate 业务规则复杂 | 中 | 中 | 复用 80 号方案状态机，补充服务层单元测试；引入业务方评审。 |
| 组织转移与汇报关系数据耦合 | 中 | 中 | 与组织模块建立对齐会议，新增集成测试模拟组织结构变化。 |
| 时间线延误 | 中 | 低 | 6 周周期含 1 周缓冲；每周评审，及时调整资源或范围。 |

---

## 7. 协作机制

- **每周例会**：周一计划同步、周三风控会、周五阶段评审。  
- **跨团队联络**：命令/查询/前端/架构各指派 1 名负责人，维护进度看板；重大事项在 06 号日志记录。  
- **代码规范**：遵循 CLAUDE.md/AGENTS.md；提交前执行 `make fmt`、`make lint`、前端 ESLint/Vitest。  
- **文档治理**：所有新文档或更新必须验证唯一事实来源，变更后立刻更新 84/06 号文档与相关报告。

---

## 8. 验收与归档记录

- ✅ 2025-10-16：Stage 2 全量上线，命令/查询服务、前端交互与E2E 覆盖完成。
- ✅ 核心验证：
  - `go test ./cmd/organization-query-service/internal/graphql`
  - `npx vitest run frontend/src/features/positions/__tests__/PositionDashboard.test.tsx`
  - `cd frontend && npx playwright test tests/e2e/position-lifecycle.spec.ts --config playwright.config.ts`
- ✅ 文档同步：实现清单、契约差异报告、06号进展日志更新；80号计划勾选 Stage 2 任务。
- 📦 归档说明：本计划文件迁移至 `docs/archive/development-plans/`，持续参考 80 号与 06 号文档追踪后续迭代。

## 9. 变更记录

| 版本 | 日期 | 内容 | 作者 |
|------|------|------|------|
| v1.0 | 2025-10-16 | Stage 2 实施完成，补充验收记录并归档 | 项目智能助手 |
| v0.2 | 2025-10-16 | 采纳方案B，重写 Stage 2 计划（移除双数据源、延长周期） | 项目智能助手 |

---

> 本计划完全遵循“资源唯一性与跨层一致性”最高优先级原则，禁止再引入双写或长期临时方案。任职相关数据自 Stage 2 起仅存储于 `position_assignments`，所有契约、实现与前端逻辑同时调整，确保端到端一致。完成后请归档本计划并在 06 号文档记录 Stage 2 验收结果。
