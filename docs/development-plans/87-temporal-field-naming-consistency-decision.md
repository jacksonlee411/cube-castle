# 87号文档：时态字段命名一致性决策文档

**版本**: v1.2
**创建日期**: 2025-10-17
**维护团队**: 架构组 + 数据库团队 + 命令服务团队 + 查询服务团队 + 前端团队
**状态**: 已完成（2025-10-21）
**优先级**: 🔴 高（影响架构一致性）
**关联文档**: 80号职位管理方案 · 84号 Stage 2 计划 · 86号 Stage 4 计划评审 · 06号进展日志
**遵循原则**: CLAUDE.md 资源唯一性与跨层一致性原则（最高优先级）

---

## 1. 问题概述

### 1.1 问题发现

在86号计划评审过程中，发现项目存在**时态字段命名不一致**问题：

| 模块 | 表名 | 时态字段命名 | 迁移文件 |
|------|------|------------|---------|
| **组织架构** | `organization_units` | `effective_date` + `end_date` | 008_temporal_management_schema.sql |
| **职位主数据** | `positions` | `effective_date` + `end_date` | 043_create_positions_and_job_catalog.sql |
| **Job Catalog** | `job_family_groups` / `job_families` / `job_roles` / `job_levels` | `effective_date` + `end_date` | 043_create_positions_and_job_catalog.sql |
| **任职记录** | `position_assignments` | **`start_date` + `end_date`** 🔴 | 044_create_position_assignments.sql |

**不一致项**：
- 组织架构、职位主数据、Job Catalog 统一使用 `effective_date`
- 任职记录（position_assignments）单独使用 `start_date`

### 1.2 影响范围

**数据库层**：
- 1个表（`position_assignments`）使用不同命名
- 4个索引包含 `start_date` 字段

**代码层**：
- 仓储层：`position_assignment_repository.go` 字段映射
- 服务层：Fill/Vacate/Transfer 操作
- GraphQL：`positionAssignments` 查询返回字段
- 前端：类型定义（`Assignment` 接口）

**文档层**：
- 80号方案声称"完全复用组织架构模式"但实际未完全对齐
- 84号、86号计划涉及 Assignment 字段的所有描述

---

## 2. 详细调查发现

### 2.1 代码证据

#### 证据1：组织架构使用 effective_date（008迁移）

```sql
-- database/migrations/008_temporal_management_schema.sql:34
CREATE TABLE organization_units (
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT true,
    -- ...
);
```

#### 证据2：职位主数据使用 effective_date（043迁移）

```sql
-- database/migrations/043_create_positions_and_job_catalog.sql:141
CREATE TABLE positions (
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    -- ...
    UNIQUE (tenant_id, code, effective_date)
);
```

**80号方案第184-187行明确承诺**：
```markdown
-- 时态字段（完全复用组织架构模式）
effective_date DATE NOT NULL,
end_date DATE,
is_current BOOLEAN NOT NULL DEFAULT false,
```

#### 证据3：任职记录使用 start_date（044迁移）🔴

```sql
-- database/migrations/044_create_position_assignments.sql:17-18
CREATE TABLE position_assignments (
    start_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    -- ...
    CONSTRAINT chk_position_assignments_dates
        CHECK (end_date IS NULL OR end_date > start_date),
);

-- 索引也基于 start_date
CREATE UNIQUE INDEX uk_position_assignments_start
    ON position_assignments(tenant_id, position_code, employee_id, start_date);

CREATE INDEX idx_position_assignments_position
    ON position_assignments(tenant_id, position_code, start_date DESC);

CREATE INDEX idx_position_assignments_employee
    ON position_assignments(tenant_id, employee_id, start_date DESC);
```

#### 证据4：仓储代码使用 start_date

```go
// cmd/organization-command-service/internal/repository/position_assignment_repository.go:85
func (r *PositionAssignmentRepository) CreateAssignment(...) (*types.PositionAssignment, error) {
    query := `INSERT INTO position_assignments (
        tenant_id, position_code, position_record_id, employee_id, employee_name, employee_number,
        assignment_type, assignment_status, fte, start_date, end_date, is_current, notes
    ) VALUES (
        $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13
    ) RETURNING assignment_id, ...`

    // $10 = entity.StartDate
}
```

### 2.2 可能的设计意图分析

#### 假设1：语义差异论

```yaml
主数据时态语义（Organization/Position）：
  - effective_date: "此版本数据从何时生效"
  - 侧重：数据有效性时间
  - 场景：支持未来版本（如计划中的组织调整、职位设置）
  - 示例：2025-11-01 生效的组织架构调整

关系数据事件语义（Assignment）：
  - start_date: "员工从何时开始任职"
  - 侧重：事件开始时间
  - 场景：记录具体的雇佣关系起始
  - 示例：员工于 2025-10-15 入职某职位
```

**但这种区分是否必要？**
- Assignment 本质上也是"有效时间"概念
- "任职从何时生效" = "任职关系的 effective_date"
- 语义差异不足以支撑命名不一致的代价

#### 假设2：Workday 参考模型影响

Workday HCM 系统中：
- Position（职位）使用 `Effective Date`
- Worker Assignment（员工任职）使用 `Start Date` 或 `Hire Date`

**可能是对标 Workday 的命名习惯。**

但：
- Workday 是商业系统，有其历史包袱
- 我们可以设计更一致的模型
- 不应照搬所有细节

#### 假设3：实施疏忽

044迁移可能是：
- 不同开发者实现
- 未充分参考080号方案的架构设计
- 缺少架构评审环节

---

## 3. 不一致性带来的问题

### 3.1 查询复杂度增加

**场景**：查询"2025-10-01 某职位及其任职情况"

```sql
-- 需要JOIN两套不同的时态逻辑
SELECT
    p.code,
    p.title,
    pa.employee_name
FROM positions p
LEFT JOIN position_assignments pa
    ON p.code = pa.position_code
    AND p.tenant_id = pa.tenant_id
WHERE p.tenant_id = 'xxx'
  AND p.effective_date <= '2025-10-01'     -- 注意这里是 effective_date ⚠️
  AND (p.end_date IS NULL OR p.end_date > '2025-10-01')
  AND pa.start_date <= '2025-10-01'        -- 这里却是 start_date ⚠️
  AND (pa.end_date IS NULL OR pa.end_date > '2025-10-01')
  AND pa.is_current = true;
```

**问题**：
- 开发者容易混淆
- SQL 可读性下降
- 查询模板无法复用

### 3.2 API 响应不一致

**GraphQL Schema**：
```graphql
type Position {
  code: String!
  title: String!
  effectiveDate: String!   # 来自 positions.effective_date
  endDate: String
}

type PositionAssignment {
  assignmentId: ID!
  startDate: String!        # 来自 position_assignments.start_date ⚠️
  endDate: String
}
```

**前端类型定义**：
```typescript
// frontend/src/shared/types/positions.ts
interface Position {
  code: string;
  title: string;
  effectiveDate: string;    // 一个命名
  endDate?: string;
}

interface Assignment {
  assignmentId: string;
  startDate: string;        // 另一个命名 ⚠️
  endDate?: string;
}
```

**问题**：
- 前端开发者需要记住两套命名
- 时间轴展示需要特殊处理
- API 文档需要额外说明

### 3.3 代码维护成本

**需要维护两套时态查询逻辑**：

```go
// 职位时态查询
func (r *PositionRepository) GetPositionAsOf(code string, asOfDate time.Time) {
    query := `SELECT * FROM positions
              WHERE code = $1
                AND effective_date <= $2    // effective_date
                AND (end_date IS NULL OR end_date > $2)`
}

// 任职时态查询
func (r *AssignmentRepository) GetAssignmentAsOf(code string, asOfDate time.Time) {
    query := `SELECT * FROM position_assignments
              WHERE position_code = $1
                AND start_date <= $2        // start_date ⚠️
                AND (end_date IS NULL OR end_date > $2)`
}
```

**问题**：
- 无法抽象通用的时态查询工具
- 增加单元测试复杂度
- 新人学习曲线陡峭

### 3.4 违反架构原则

**CLAUDE.md 资源唯一性原则**：
> 所有实现、文档与契约必须保持唯一事实来源与端到端一致性

**80号方案第184行承诺**：
> -- 时态字段（完全复用组织架构模式）

**当前状态**：
- ❌ 未能完全复用
- ❌ 存在两套时态字段命名标准
- ❌ 文档与实现不一致

---

## 4. 唯一方案：统一为 `effective_date`

#### 4.1 方案描述

将 `position_assignments.start_date` 重命名为 `effective_date`，与全系统保持一致。

#### 4.2 实施步骤

**步骤1：创建迁移脚本 047**

```sql
-- 047_rename_assignment_start_date_to_effective_date.sql
BEGIN;

-- 1. 重命名字段
ALTER TABLE position_assignments
RENAME COLUMN start_date TO effective_date;

-- 2. 更新约束（引用了字段名）
ALTER TABLE position_assignments
DROP CONSTRAINT chk_position_assignments_dates;

ALTER TABLE position_assignments
ADD CONSTRAINT chk_position_assignments_dates
    CHECK (end_date IS NULL OR end_date > effective_date);

-- 3. 重建索引
DROP INDEX IF EXISTS uk_position_assignments_start;
CREATE UNIQUE INDEX uk_position_assignments_effective
    ON position_assignments(tenant_id, position_code, employee_id, effective_date);

DROP INDEX IF EXISTS idx_position_assignments_position;
CREATE INDEX idx_position_assignments_position
    ON position_assignments(tenant_id, position_code, effective_date DESC);

DROP INDEX IF EXISTS idx_position_assignments_employee;
CREATE INDEX idx_position_assignments_employee
    ON position_assignments(tenant_id, employee_id, effective_date DESC);

COMMIT;
```

**步骤2：更新仓储层**

```go
// cmd/organization-command-service/internal/types/positions.go
type PositionAssignment struct {
    AssignmentID     uuid.UUID      `db:"assignment_id"`
    TenantID         uuid.UUID      `db:"tenant_id"`
    PositionCode     string         `db:"position_code"`
    EffectiveDate    time.Time      `db:"effective_date"`  // 改名
    EndDate          sql.NullTime   `db:"end_date"`
    // ...
}

// cmd/organization-command-service/internal/repository/position_assignment_repository.go
func (r *PositionAssignmentRepository) CreateAssignment(...) {
    query := `INSERT INTO position_assignments (
        tenant_id, position_code, position_record_id, employee_id, employee_name,
        assignment_type, assignment_status, fte, effective_date, end_date, is_current, notes
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`

    // 参数顺序调整
}
```

**步骤3：更新 GraphQL Schema**

```graphql
# docs/api/schema.graphql
type PositionAssignment {
  assignmentId: ID!
  tenantId: ID!
  positionCode: String!
  employeeId: ID!
  employeeName: String!
  assignmentType: AssignmentType!
  assignmentStatus: AssignmentStatus!
  fte: Float!
  effectiveDate: String!     # 统一命名
  endDate: String
  isCurrent: Boolean!
  notes: String
  createdAt: String!
  updatedAt: String!
}
```

**步骤4：更新前端类型**

```typescript
// frontend/src/shared/types/positions.ts
export interface PositionAssignment {
  assignmentId: string;
  tenantId: string;
  positionCode: string;
  employeeId: string;
  employeeName: string;
  assignmentType: 'PRIMARY' | 'SECONDARY' | 'ACTING';
  assignmentStatus: 'PENDING' | 'ACTIVE' | 'ENDED';
  fte: number;
  effectiveDate: string;     // 统一命名
  endDate?: string;
  isCurrent: boolean;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}
```

**步骤5：更新文档**

- 更新 80号方案（确认与实现一致）
- 更新 84号计划（归档版本，补充说明）
- 更新 86号计划（如果继续，需同步字段名）
- 在 06号日志中记录此架构决策

#### 4.3 优点

✅ **架构一致性**：全系统统一使用 `effective_date`
✅ **查询简化**：可复用时态查询逻辑
✅ **代码可维护性**：单一命名标准
✅ **符合80号承诺**："完全复用组织架构模式"
✅ **长期收益**：降低新人学习成本

#### 4.4 风险与缓解

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| 迁移失败导致数据损坏 | 中 | 执行前完整备份；先在测试环境验证；提供回滚脚本 |
| 现有代码未完全更新 | 中 | 编译期类型检查；单元测试全面覆盖；代码审查 |
| 前端字段名不匹配 | 低 | TypeScript 类型系统保证；契约测试验证 |
| 文档同步遗漏 | 低 | 使用文档同步检查脚本；架构组审核 |

#### 4.5 工作量评估

| 任务 | 工作量 | 责任人 |
|------|--------|--------|
| 047 迁移脚本编写与测试 | 2小时 | 数据库团队 |
| 仓储层代码更新 | 3小时 | 命令服务团队 |
| GraphQL Schema 与 Resolver 更新 | 2小时 | 查询服务团队 |
| 前端类型与组件更新 | 3小时 | 前端团队 |
| 单元测试与集成测试更新 | 4小时 | QA + 各团队 |
| 文档同步 | 2小时 | 架构组 |
| **总计** | **16小时（2个工作日）** | 全团队 |

> **备注**：如在执行中发现额外测试或通知需求，可在此基础上按 20% 缓冲动态调整，无需单独变更计划。

#### 4.6 测试与验证清单

- **数据库层**：
  - `go test ./cmd/organization-command-service/...`（覆盖 Fill/Vacate/Transfer/Acting 场景）。
  - `go test ./cmd/organization-query-service/...`（验证 GraphQL `currentAssignment`、`assignmentHistory` 查询）。
  - 如有集成脚本，执行 `make test-integration` 并检查迁移前后数据一致性。
- **前端层**：
  - `npm --prefix frontend run test -- --run src/features/positions`。
  - `npm --prefix frontend run test -- --run src/shared/hooks`（覆盖 Assignment 相关 Hook）。
  - Playwright：`npm --prefix frontend run test:e2e -- position-lifecycle.spec.ts`（验证时间轴与任职流程）。
- **契约校验**：
  - `npm --prefix frontend run test:contract`（GraphQL/REST 字段同步）。
  - `node scripts/quality/architecture-validator.js`（确保命名一致性检查通过）。
- **上线后验证**：
  - 调用 REST `/api/v1/positions/{code}/assignments` 与 GraphQL `position { currentAssignment assignmentHistory }` 核对字段名。
  - 手动验证前端缓存刷新（清理 service worker / local storage 或 bump 版本号）。

#### 4.7 下游影响与通知

- **对外契约**：OpenAPI 与 GraphQL 字段名将改为 `effectiveDate`。需提前在发布说明中标注 breaking change，并向使用 REST/GraphQL 的集成方至少提前 5 个工作日通知。
- **内部依赖**：
  - BI 报表/ETL 若直接读取 `position_assignments`，需同步更新 SQL 脚本。
  - 监控告警如引用 `start_date`，需调整指标标签或查询。
- **缓存策略**：前端需在发布版本中增加缓存 bust 机制（如增加版本号或调用 `queryClient.invalidateQueries('positions')`）。
- **日志与审计**：更新审计事件 payload 字段名，确保后续排障时的日志字段一致。

---

## 5. 决策建议

### 5.1 架构组推荐：统一为 `effective_date`

**理由**：
1. ✅ 符合 CLAUDE.md 最高优先级原则（资源唯一性与一致性）
2. ✅ 兑现 80号方案承诺（"完全复用组织架构模式"）
3. ✅ 长期收益显著（可维护性、可扩展性）
4. ✅ 一次性成本可控（2个工作日）
5. ✅ 为未来扩展（如员工主数据、薪酬模块）奠定一致基础

**时机**：
- ✅ 当前 Stage 3 刚完成，Stage 4 尚未启动
- ✅ 现有代码量较小，改动范围可控
- ✅ 越晚处理，累积成本越高

### 5.2 决策流程

1. **架构组复核**：确认唯一方案内容完整，补充必要细节后在团队频道发布。
2. **跨团队异步确认**：命令/查询/前端/数据库/QA 在 24 小时内通过评论或 ✅ 反应确认无异议。
3. **指派执行人**：由架构组在任务看板上指派实施负责人并排定时间窗口。
4. **执行与自验**：按照第4节步骤完成迁移与测试，并在共享文档中记录结果。
5. **验收归档**：执行人汇总验证结果，架构组复核后更新 06 号日志并归档本文档。

---

## 6. 回滚预案

### 6.1 回滚脚本（统一命名方案）

如果047迁移执行后发现问题，可立即回滚：

```sql
-- 047_rollback.sql
BEGIN;

-- 1. 重命名回 start_date
ALTER TABLE position_assignments
RENAME COLUMN effective_date TO start_date;

-- 2. 恢复约束
ALTER TABLE position_assignments
DROP CONSTRAINT chk_position_assignments_dates;

ALTER TABLE position_assignments
ADD CONSTRAINT chk_position_assignments_dates
    CHECK (end_date IS NULL OR end_date > start_date);

-- 3. 恢复索引
DROP INDEX IF EXISTS uk_position_assignments_effective;
CREATE UNIQUE INDEX uk_position_assignments_start
    ON position_assignments(tenant_id, position_code, employee_id, start_date);

DROP INDEX IF EXISTS idx_position_assignments_position;
CREATE INDEX idx_position_assignments_position
    ON position_assignments(tenant_id, position_code, start_date DESC);

DROP INDEX IF EXISTS idx_position_assignments_employee;
CREATE INDEX idx_position_assignments_employee
    ON position_assignments(tenant_id, employee_id, start_date DESC);

COMMIT;
```

### 6.2 验证清单

- [ ] 数据完整性：行数一致，无数据丢失
- [ ] 约束有效：CHECK 约束正常工作
- [ ] 索引性能：查询计划无退化
- [ ] 单元测试：全部通过
- [ ] 集成测试：Assignment CRUD 正常
- [ ] E2E测试：Position 相关流程通过

---

## 7. 预期输出

**交付物**：
- [x] 047 迁移脚本（含回滚脚本）
- [x] 更新后的仓储层代码（Go）
- [x] 更新后的 GraphQL Schema 与 Resolver
- [x] 更新后的前端类型定义
- [x] 更新后的单元测试与集成测试
- [x] 更新后的 80/84/86 号文档
- [x] 在 06号日志中记录决策与执行结果
- [ ] 本文档归档至 `docs/archive/development-plans/`

**验收标准**：
- ✅ 全系统时态字段统一为 `effective_date`
- ✅ 所有测试通过
- ✅ 文档与代码同步
- ✅ 架构组验收签字

---

## 8. 关联文档

- `docs/development-plans/80-position-management-with-temporal-tracking.md` - 职位管理总方案（承诺"完全复用"）
- `docs/development-plans/86-position-assignment-stage4-plan.md` - Stage 4 计划（触发此次调查）
- `docs/development-plans/06-integrated-teams-progress-log.md` - 进展日志（记录86号评审）
- `database/migrations/008_temporal_management_schema.sql` - 组织架构时态模式
- `database/migrations/043_create_positions_and_job_catalog.sql` - 职位主数据时态模式
- `database/migrations/044_create_position_assignments.sql` - 任职记录时态模式（使用 start_date）
- `CLAUDE.md` - 项目核心原则（资源唯一性与一致性）

---

## 9. 决策记录

### 9.1 决策确认流程

- **时间**：由架构组在共享频道发起，24 小时内收敛反馈
- **参与方**：架构组（牵头）· 数据库团队 · 命令服务 · 查询服务 · 前端 · QA
- **确认方式**：在文档或任务评论中以 ✅ 回复表示同意，如有异议须同步提供修订建议
- **决策人**：架构组长（对外发布最终确认）

### 9.2 决策结果

- [x] **统一为 `effective_date`**（唯一建议，待签字）

**决策日期**：2025-10-21
**决策人签字**：架构组（异步确认）
**执行负责人**：数据库 + 全栈团队

---

## 10. 变更记录

| 版本 | 日期 | 说明 | 作者 |
|------|------|------|------|
| v1.2 | 2025-10-21 | 完成迁移与全栈改造，更新测试结果与决议状态 | Claude Code 助手 |
| v1.1 | 2025-10-21 | 明确唯一方案、补充测试/迁移计划、调整决策流程 | Claude Code 助手 |
| v1.0 | 2025-10-17 | 初始版本，提交决策 | 架构组 Claude Code 助手 |

---

## 11. 生产迁移计划

### 11.1 现状评估

- 统计生产环境 `position_assignments` 行数与索引大小：`SELECT COUNT(*), pg_size_pretty(pg_total_relation_size('position_assignments'));`
- 确认是否存在跨租户写入高峰，选择低流量窗口（建议北京时间 02:00-04:00）。
- 盘点依赖 `start_date` 的定时任务/报表脚本，准备同步更新时间点。

### 11.2 迁移前校验

1. 在预生产环境执行 047 迁移脚本，记录执行时间与锁表情况。
2. 导出生产数据库备份或创建分区级快照，确认可在 15 分钟内恢复。
3. 将回滚脚本（第6节）预置到运维工具，验证可无损恢复。

### 11.3 执行步骤

1. 宣布维护窗口并冻结相关部署流水线。
2. 停止命令服务写流量（或切换至维护模式），保留查询服务只读。
3. 执行 047 迁移脚本，记录开始/结束时间与日志。
4. 运行第4.6节列出的测试清单（至少覆盖数据库单测、关键 GraphQL 查询与一条 Playwright 脚本）。
5. 恢复命令服务写流量，并监控错误日志 30 分钟。

### 11.4 回滚策略

- 若出现阻塞或验证失败，立即执行第6节回滚脚本，恢复服务写流量。
- 回滚后需通知所有团队保持 `start_date` 字段命名，并在 06 号日志登记原因与后续动作。

### 11.5 发布与通知

- 在发布说明中声明 REST/OpenAPI 与 GraphQL 的 breaking change，附上新旧字段对照表。
- 向外部集成方及 BI 团队发送邮件/Slack 通知，包含迁移窗口、风险提示与测试结果链接。
- 迁移完成后于 06 号日志补充生产验证截图或日志片段。

---

**文档状态**：✅ 已完成（开发环境改造交付）
**下一步行动**：按第 11 节流程在生产窗口执行迁移并归档文档
**预期完成日期**：2025-10-21（开发侧）
