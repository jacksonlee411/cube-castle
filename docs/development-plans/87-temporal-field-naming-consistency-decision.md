# 87号文档：时态字段命名一致性决策文档

**版本**: v1.0
**创建日期**: 2025-10-17
**维护团队**: 架构组 + 数据库团队 + 命令服务团队 + 查询服务团队 + 前端团队
**状态**: 待决策
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

## 4. 决策方案

### 方案A：统一为 `effective_date`（推荐）⭐⭐⭐

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

---

### 方案B：保持 `start_date`（不推荐）❌

#### 4.6 方案描述

维持现状，接受两套命名体系并存。

#### 4.7 理由

- 避免迁移风险
- 保持"雇佣合同"语义
- 已有代码无需修改

#### 4.8 代价

❌ **永久背负两套命名体系**
❌ **查询逻辑复杂，无法复用**
❌ **违反 CLAUDE.md 一致性原则**
❌ **违反 80号方案承诺**
❌ **新人学习成本高**
❌ **长期维护成本持续累积**

#### 4.9 工作量

- 无需立即工作
- 但每次涉及 Assignment 的开发都会付出额外成本
- 长期累积成本 > 方案A的一次性成本

---

### 方案C：API层映射统一（折衷）⚠️

#### 4.10 方案描述

数据库层保持不变，在 API 层（GraphQL/REST）统一对外暴露为 `effectiveDate`。

```yaml
数据库层（内部）：
  - positions.effective_date
  - position_assignments.start_date

API 层（对外）：
  - Position.effectiveDate → effective_date
  - Assignment.effectiveDate → start_date (映射)

Resolver 层实现映射：
  effectiveDate: (parent) => parent.start_date
```

#### 4.11 优点

✅ 对外 API 一致性
✅ 避免数据库迁移
✅ 前端无需感知差异

#### 4.12 缺点

❌ 数据库层仍然不一致
❌ 增加映射逻辑复杂度
❌ SQL 查询仍然复杂
❌ 仓储层仍需维护两套命名
❌ 治标不治本

#### 4.13 工作量

| 任务 | 工作量 |
|------|--------|
| GraphQL Resolver 映射逻辑 | 2小时 |
| 文档说明 | 1小时 |
| **总计** | **3小时** |

**但长期维护成本仍然较高。**

---

## 5. 决策建议

### 5.1 架构组推荐：方案A（统一为 effective_date）⭐

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

1. **架构组复核本文档**（1天）
2. **召开技术评审会议**（命令/查询/前端/数据库/QA 参与）（半天）
3. **投票决策**（采用方案A/B/C）
4. **执行实施**（如选择方案A，预计2个工作日）
5. **验收与归档**（更新文档，本文档归档）

### 5.3 决策矩阵

| 维度 | 方案A（统一） | 方案B（维持） | 方案C（映射） |
|------|------------|------------|------------|
| 架构一致性 | ⭐⭐⭐⭐⭐ | ❌ | ⭐⭐⭐ |
| 查询复杂度 | ⭐⭐⭐⭐⭐ | ❌ | ⭐⭐ |
| 代码可维护性 | ⭐⭐⭐⭐⭐ | ❌ | ⭐⭐ |
| 实施风险 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| 一次性成本 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| 长期成本 | ⭐⭐⭐⭐⭐ | ❌ | ⭐⭐ |
| **综合评分** | **⭐⭐⭐⭐⭐** | **❌ 不推荐** | **⭐⭐⭐** |

---

## 6. 回滚预案

### 6.1 方案A回滚脚本

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

### 7.1 如果选择方案A

**交付物**：
- [ ] 047 迁移脚本（含回滚脚本）
- [ ] 更新后的仓储层代码（Go）
- [ ] 更新后的 GraphQL Schema 与 Resolver
- [ ] 更新后的前端类型定义
- [ ] 更新后的单元测试与集成测试
- [ ] 更新后的 80/84/86 号文档
- [ ] 在 06号日志中记录决策与执行结果
- [ ] 本文档归档至 `docs/archive/development-plans/`

**验收标准**：
- ✅ 全系统时态字段统一为 `effective_date`
- ✅ 所有测试通过
- ✅ 文档与代码同步
- ✅ 架构组验收签字

### 7.2 如果选择方案B

**交付物**：
- [ ] 在本文档中记录"决策保持现状"及理由
- [ ] 在 CLAUDE.md 或 AGENTS.md 中补充"例外说明"
- [ ] 更新 80号方案，说明"部分复用"而非"完全复用"
- [ ] 本文档归档

**后果**：
- ⚠️ 长期维护成本持续累积
- ⚠️ 违反架构一致性原则

### 7.3 如果选择方案C

**交付物**：
- [ ] GraphQL Resolver 映射逻辑
- [ ] API 文档补充说明（effectiveDate 实际映射到 start_date）
- [ ] 前端类型定义更新
- [ ] 本文档归档

**限制**：
- ⚠️ 仅解决对外 API 一致性
- ⚠️ 内部仍不一致

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

### 9.1 决策会议

- **计划时间**：待定（架构组确定）
- **参与方**：架构组 + 数据库团队 + 命令服务 + 查询服务 + 前端 + QA
- **决策方式**：技术评审 + 投票
- **决策人**：架构组长

### 9.2 决策结果

- [ ] **方案A**：统一为 `effective_date`（推荐）
- [ ] **方案B**：保持 `start_date`（不推荐）
- [ ] **方案C**：API层映射统一（折衷）

**决策日期**：_________
**决策人签字**：_________
**执行负责人**：_________

---

## 10. 变更记录

| 版本 | 日期 | 说明 | 作者 |
|------|------|------|------|
| v1.0 | 2025-10-17 | 初始版本，提交决策 | 架构组 Claude Code 助手 |

---

**文档状态**：⏳ 待决策
**下一步行动**：架构组召集决策会议
**预期完成日期**：待定
