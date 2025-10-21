# 80号文档：职位管理模块设计方案（复用时态管理架构）

**版本**: v1.0
**创建日期**: 2025-10-12
**最新更新**: 2025-10-21
**维护团队**: 后端团队 + 前端团队
**状态**: Stage 2 已完成（Stage 3 已批准启动） ✅ · Stage 4 已完成（详见 86 号归档计划）
**关联计划**: 60号系统级质量重构总计划 · 85号 Stage 3 执行计划 v0.2
**参考系统**: Workday Position Management + HCM Core
**遵循原则**: CLAUDE.md 资源唯一性与跨层一致性原则（最高优先级）

---

## 1. 背景与目标

### 1.1 业务背景

基于 Workday Position Management 最佳实践，职位（Position）是组织架构管理的关键延伸，代表组织中的具体岗位实例。与岗位定义（Job Profile）不同，Position 是可被员工填充的具体工作位置。

**Workday 核心理念**：
- **Effective Dating**: 支持过去、现在、未来的职位状态管理
- **Position as Instance**: 职位是岗位定义的具体实例
- **Timeline Management**: 完整的职位生命周期时间线追踪
- **Organization Linkage**: 职位必须归属于组织单元

### 1.2 设计目标

1. **复用时态管理模式**：抽象通用 TemporalCore 并复用组织架构的时态字段（effectiveDate、endDate、isCurrent），`isFuture` 统一由服务层派生
2. **实现全生命周期管理**：从职位创建、填充、空缺到撤销的完整周期
3. **职位与组织关系**：强关联组织单元，支持职位在组织间转移
4. **保持架构一致性**：遵循 CQRS 架构，REST 命令 + GraphQL 查询

### 1.3 Workday 参考模型

根据 Workday 文档和行业最佳实践：

```
Organization Structure (已实现)
    └── Position (本方案)
            └── Worker Assignment (未来扩展)
                    └── Compensation & Benefits (未来扩展)
```

---

## 2. 核心概念与术语

### 2.1 Position vs Job Profile

| 概念 | 定义 | 示例 | 时态性 |
|------|------|------|--------|
| **Job Profile** | 工作职责的抽象定义 | "高级软件工程师" | 相对稳定 |
| **Position** | Job Profile 的具体实例 | "技术部-后端-高工-P001" | 支持时态 |
| **Assignment** | 员工对职位的任职 | "张三 任职 P001" | 支持时态 |

### 2.2 职位状态模型

借鉴 Workday 的状态管理：

```
PLANNED (计划中)
    → ACTIVE (激活-空缺)
        → FILLED (已填充)
            → VACANT (空缺)
                → INACTIVE (停用)
                    → DELETED (删除)
```

**状态说明**：
- **PLANNED**: 未来生效的职位（如组织扩编计划）
- **ACTIVE**: 当前激活但无人任职的职位
- **FILLED**: 有员工任职的职位
- **VACANT**: 员工离职后的空缺职位
- **INACTIVE**: 暂时停用但保留编制
- **DELETED**: 撤销的职位（软删除）

### 2.3 职位编码规则

复用组织架构的7位编码模式：

```
格式: P + 7位数字
范围: P1000000 - P9999999
示例: P1000001 (第一个职位)
```

### 2.4 职位体系化（Job Catalog）映射

Workday 将职位体系拆分为 Job Family Group → Job Family → Job Role → Job Level 四层关联。本方案对齐如下：

```
Job Family Group (职类，最广泛职能分组) : 职能类别，例如 管理类 / 专业类 / 操作类
  └─ Job Family (职种)                   : 细分业务域，例如 财务 / 人力 / IT
       └─ Job Role (职务)        : 具体岗位角色，例如 后端工程师 / QA 主管
            └─ Job Level (职级)  : 水平层级，例如 P5 / M3 / S2
```

- 职位创建/版本更新时必须指定四级编码，系统根据主数据表验证层级合法性。
- `Position` 返回时同时输出 `jobFamilyGroup/jobFamily/jobRole/jobLevel` 对象，支持前端透视分析与权限隔离。
- 本方案未额外引入“专业序列”等中间层级，Job Family 已覆盖常见业务域。
- 该体系与 Job Profile（岗位定义）互补：Job Profile 描述职责模板，Job Catalog 定义职位在组织间的归类规则。

**行业示例（中国大陆物业场景）**

在物业保洁团队中，可能只设立一个“保洁员职位”（Position），其 `headcountCapacity` 设置为 8，对应 8 个可用编制；所有保洁员均向同一“保洁主管”职位汇报。这样既保持职位管理的席位控制，又能灵活反映一岗多人的运营模式。

### 2.5 编码规则（对标 Workday 实践）

| 层级 | 字段 | 编码格式 | 示例 | 设计说明 | 与 Workday/主流 HCM 对比 |
|------|------|-----------|-------|-----------|---------------------------|
| 职类 (Job Family Group) | `job_family_group_code` | 4–6 位大写字母，允许单一词根，正则 `^[A-Z]{4,6}$` | `PROF`, `MGT`, `OPER` | 对齐 Workday “Job Family Group Code” 的短码实践，保持语义可读 | Workday 默认允许大写字母短码，SuccessFactors/Oracle 需维护 GUID → 业务码映射，本方案同样保留业务短码并以 `record_id` 维持全局唯一 |
| 职种 (Job Family) | `job_family_code` | 职类码 + `-` + 3–6 位大写字母或数字，正则 `^{FG}-[A-Z0-9]{3,6}$`（其中 `{FG}` 为父级职类码） | `PROF-IT`, `PROF-HR`, `OPER-PLT` | 通过前缀显式绑定父级，避免跨类引用；利于导入导出和分组统计 | Workday 推荐以父级作为前缀；SuccessFactors/Oracle 多依赖 GUID，本方案保留层级前缀以便同步 |
| 职务 (Job Role) | `job_role_code` | 职种码 + `-` + 3–5 位大写字母/数字，正则 `^{FamilyCode}-[A-Z0-9]{3,5}$` | `PROF-IT-BKND`, `PROF-IT-SYS` | 保持与 Workday Job Profile/Job Role 命名一致，支持导出到 Job Profile Catalog | Workday Job Profile Code 常为 8–12 位短码，本方案以组合码确保唯一；Oracle 支持更长字符串，可通过别名映射 |
| 职级 (Job Level) | `job_level_code` | 单字母档位 + 1–2 位数字，正则 `^[A-Z][0-9]{1,2}$` | `P5`, `M3`, `S2` | 延续现有组织/权限等级格式，可与 Workday Grade/Level 对齐 | Workday Grade 亦使用字母+数字（如 P3、M2）；SuccessFactors 通过 pay grade 维护，可映射到本字段 |
| 职位 (Position) | `code` | 固定 `P` + 7 位数字，正则 `^P[0-9]{7}$` | `P1000001` | 复用现有组织职位方案，保证跨租户唯一 | Workday Position ID 常允许自定义，本方案使用数值序列；可在同步时维护别名 |

> **命名策略**  
> - 业务码全部使用大写字母和数字，避免本地化字符导致的同步问题。  
> - 组合码以父级短码作为前缀，便于快速识别层级归属。  
> - `record_id` 继续作为数据库主键（UUID），满足与 SuccessFactors/Oracle 等以 GUID 为主键的系统对接需求。  
> - 允许在集成层建立“外部系统编码”映射表，将 Workday 的 `External ID`、Oracle 的 `Code`、SuccessFactors 的 `ExternalCode` 统一到上述业务码。

---

## 3. 数据模型设计

### 3.1 核心实体：Position

**PostgreSQL 表结构**：

```sql
CREATE TABLE positions (
    -- 主键与租户隔离
    code VARCHAR(8) NOT NULL,                    -- P1000001
    tenant_id UUID NOT NULL,
    record_id UUID NOT NULL DEFAULT gen_random_uuid(),

    -- 基本信息
    title VARCHAR(255) NOT NULL,                 -- 职位名称：技术部-后端工程师
    job_profile_code VARCHAR(50),                -- 岗位定义代码（外部系统）
    job_profile_name VARCHAR(255),               -- 岗位定义名称：高级软件工程师

    -- 职位体系化分类（对标 Workday Job Catalog）
    job_family_group_code VARCHAR(20) NOT NULL,  -- 最广泛职能分组（Job Family Group）
    job_family_group_name VARCHAR(255) NOT NULL,
    job_family_group_record_id UUID NOT NULL,
    job_family_code VARCHAR(20) NOT NULL,        -- 职种/职系（Job Family / Job Function）
    job_family_name VARCHAR(255) NOT NULL,
    job_family_record_id UUID NOT NULL,
    job_role_code VARCHAR(20) NOT NULL,          -- 职务（Job Role）
    job_role_name VARCHAR(255) NOT NULL,
    job_role_record_id UUID NOT NULL,
    job_level_code VARCHAR(20) NOT NULL,         -- 职级（Job Level）
    job_level_name VARCHAR(255) NOT NULL,
    job_level_record_id UUID NOT NULL,

    -- 组织归属（强关联）
    organization_code VARCHAR(7) NOT NULL,       -- 归属组织：1000001
    organization_name VARCHAR(255),              -- 组织名称缓存

    -- 职位属性
    position_type VARCHAR(50) NOT NULL,          -- 职位类型：REGULAR, TEMPORARY, CONTRACT
    status VARCHAR(20) NOT NULL DEFAULT 'PLANNED', -- 状态枚举
    employment_type VARCHAR(50) NOT NULL,        -- 雇佣类型：FULL_TIME, PART_TIME, INTERN

    -- 编制与预算
    headcount_capacity DECIMAL(5,2) NOT NULL DEFAULT 1.0, -- 编制上限（FTE，支持0..N）
    headcount_in_use DECIMAL(5,2) NOT NULL DEFAULT 0.0,   -- 当前已占用 FTE
    grade_level VARCHAR(20),                     -- 兼容历史字段（计划迁移至 job_level_code）
    cost_center_code VARCHAR(50),                -- 成本中心

    -- 任职信息（冗余，便于查询）
    current_holder_id UUID,                      -- 当前任职员工ID
    current_holder_name VARCHAR(255),            -- 当前任职员工姓名
    filled_date DATE,                            -- 填充日期
    current_assignment_type VARCHAR(20),         -- PRIMARY/SECONDARY/ACTING（过渡字段）

    -- 层级关系（可选）
    reports_to_position_code VARCHAR(8),         -- 汇报职位

    -- 扩展属性
    profile JSONB DEFAULT '{}'::jsonb,           -- 扩展配置

    -- 时态字段（完全复用组织架构模式）
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,

    -- 审计字段
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,                      -- 软删除时间
    operation_type VARCHAR(20) NOT NULL,         -- CREATE, UPDATE, SUSPEND, etc.
    operated_by_id UUID NOT NULL,
    operated_by_name VARCHAR(255) NOT NULL,
    operation_reason TEXT,

    -- 主键
    PRIMARY KEY (tenant_id, code, record_id),

    -- 唯一约束（时态点唯一性）
    UNIQUE (tenant_id, code, effective_date),

    -- 唯一约束（单一当前版本）
    UNIQUE (tenant_id, code, is_current) WHERE (is_current = true AND status != 'DELETED'),

    FOREIGN KEY (job_family_group_record_id, tenant_id) REFERENCES job_family_groups(record_id, tenant_id),
    FOREIGN KEY (job_family_record_id, tenant_id) REFERENCES job_families(record_id, tenant_id),
    FOREIGN KEY (job_role_record_id, tenant_id) REFERENCES job_roles(record_id, tenant_id),
    FOREIGN KEY (job_level_record_id, tenant_id) REFERENCES job_levels(record_id, tenant_id)
);

-- 索引
CREATE INDEX idx_positions_org_code ON positions(tenant_id, organization_code, is_current);
CREATE INDEX idx_positions_current ON positions(tenant_id, is_current) WHERE is_current = true;
CREATE INDEX idx_positions_holder ON positions(tenant_id, current_holder_id) WHERE current_holder_id IS NOT NULL;
CREATE INDEX idx_positions_effective_date ON positions(tenant_id, effective_date);
CREATE INDEX idx_positions_status ON positions(tenant_id, status, is_current);
CREATE INDEX idx_positions_job_family_group ON positions(tenant_id, job_family_group_code, is_current);
CREATE INDEX idx_positions_job_family ON positions(tenant_id, job_family_code, is_current);
CREATE INDEX idx_positions_job_role ON positions(tenant_id, job_role_code, is_current);
CREATE INDEX idx_positions_job_level ON positions(tenant_id, job_level_code, is_current);
```

- `isFuture` 字段在 API/GraphQL 层通过 `effectiveDate > current_date` 动态计算，不再在数据库中持久化物理列，以保持与既有时态表的约束一致。

### 3.2 扩展实体：Position Assignment

```sql
CREATE TABLE position_assignments (
    assignment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,

    -- 关联
    position_code VARCHAR(8) NOT NULL,           -- P1000001
    position_record_id UUID NOT NULL,            -- 对应职位版本 record_id
    employee_id UUID NOT NULL,                   -- 员工ID
    employee_name VARCHAR(255) NOT NULL,

    -- 任职信息
    assignment_type VARCHAR(50) NOT NULL,        -- PRIMARY, SECONDARY, ACTING
    assignment_status VARCHAR(20) NOT NULL,      -- ACTIVE, PENDING, ENDED

    -- 时态
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,

    -- 审计
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (tenant_id, position_code, position_record_id)
      REFERENCES positions(tenant_id, code, record_id)
);
```

### 3.2.1 Position Assignment 时态模式定义

为了与职位定义（positions）的版本管理模式互补，Position Assignment 采用**事件周期模式**：每条记录代表一次独立的任职事件，`effective_date`/`end_date` 描述任职时间跨度，`is_current` 表示当前是否在任。关键规则如下：

- **唯一事实来源**：任职数据仅存储于 `position_assignments`，`positions` 表不再保留 `current_holder_*` 等冗余字段。
- **唯一性约束**：
  ```sql
  -- 一个员工在同一职位的每次任职以 effective_date 划分独立记录
  UNIQUE (tenant_id, position_code, employee_id, effective_date)

  -- 当前在职记录唯一（ACTIVE 状态下 is_current=true 最多一条）
  UNIQUE (tenant_id, position_code, employee_id, is_current)
    WHERE (is_current = true AND assignment_status = 'ACTIVE')
  ```
- **时间跨度有效性**：
  ```sql
  CHECK (end_date IS NULL OR end_date > effective_date)
  ```
- **asOfDate 查询语义**：
  ```sql
  -- 查询某日期的在任员工
  SELECT *
  FROM position_assignments
  WHERE tenant_id = $1
    AND position_code = $2
    AND effective_date <= $3
    AND (end_date IS NULL OR end_date >= $3);
  ```
- **历史修订**：任职记录如需修订（例如更正入职日期），直接更新原记录；所有修改由 `audit_logs` 记录，避免派生额外版本链。
- **多次任职场景**：员工多次在同一职位任职会生成多条独立记录（例如张三 2025 入职、2026 再次入职），便于统计空缺周期与任职历史。
- **未来计划**：允许插入 `assignment_status = 'PENDING'` 的未来生效记录，`effective_date` 到达时自动视为在任。

该模式确保任职数据与职位定义保持一致的租户隔离、审计追踪与时态查询能力，同时避免双数据源与复杂版本链，为 Stage 2 实施提供唯一事实来源。

### 3.3 职位体系化分类（Workday 级联 + 全生命周期）

为对标 Workday Job Catalog 的 effective dating 能力，引入四层时态表，结构示例如下：

```sql
CREATE TABLE job_family_groups (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    family_group_code VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, family_group_code, effective_date),
    UNIQUE (tenant_id, family_group_code) WHERE is_current,
    UNIQUE (record_id, tenant_id)
);

CREATE TABLE job_families (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    family_code VARCHAR(20) NOT NULL,
    family_group_code VARCHAR(20) NOT NULL,
    parent_record_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, family_code, effective_date),
    UNIQUE (tenant_id, family_code) WHERE is_current,
    UNIQUE (record_id, tenant_id),
    FOREIGN KEY (parent_record_id, tenant_id)
      REFERENCES job_family_groups(record_id, tenant_id)
);

CREATE TABLE job_roles (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    role_code VARCHAR(20) NOT NULL,
    family_code VARCHAR(20) NOT NULL,
    parent_record_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    competency_model JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, role_code, effective_date),
    UNIQUE (tenant_id, role_code) WHERE is_current,
    UNIQUE (record_id, tenant_id),
    FOREIGN KEY (parent_record_id, tenant_id)
      REFERENCES job_families(record_id, tenant_id)
);

CREATE TABLE job_levels (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    level_code VARCHAR(20) NOT NULL,
    role_code VARCHAR(20) NOT NULL,
    parent_record_id UUID NOT NULL,
    level_rank VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    salary_band JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, level_code, effective_date),
    UNIQUE (tenant_id, level_code) WHERE is_current,
    UNIQUE (record_id, tenant_id),
    FOREIGN KEY (parent_record_id, tenant_id)
      REFERENCES job_roles(record_id, tenant_id)
);

-- 组合索引，便于校验租户一致性
CREATE UNIQUE INDEX uk_job_family_groups_record ON job_family_groups(record_id, tenant_id, family_group_code);
CREATE UNIQUE INDEX uk_job_families_record ON job_families(record_id, tenant_id, family_code);
CREATE UNIQUE INDEX uk_job_roles_record ON job_roles(record_id, tenant_id, role_code);
CREATE UNIQUE INDEX uk_job_levels_record ON job_levels(record_id, tenant_id, level_code);
```

- 每层皆复用“单当前”“时点唯一”“自动回填 end_date”的时态模式，与职位版本保持一致。
- `parent_record_id` 直接引用父级 `record_id`，数据库层面借由唯一主键保证存在性，命令服务在同一事务内校验 `tenant_id` 对齐（避免跨租户引用）。
- 通过共享服务 `TemporalCatalogService`（见第 5.9 节）集中维护插入、更新、删除与重算逻辑。
- `positions` 表引用 `*_record_id` 作为复合外键，数据库层直接阻断跨租户引用；命令服务仍需在同一事务内校验 `tenant_id`，避免绕过。
- `uk_job_*_record` 组合索引提供 `(record_id, tenant_id, code)` 唯一性，结合复合外键与服务层校验，形成两道安全防线。

### 3.4 枚举定义

```typescript
// Position Types
export enum PositionType {
  Regular = 'REGULAR',       // 正式编制
  Temporary = 'TEMPORARY',   // 临时编制
  Contract = 'CONTRACT',     // 合同工
  Intern = 'INTERN'          // 实习生
}

// Position Status (复用组织状态模式)
export enum PositionStatus {
  Planned = 'PLANNED',       // 计划中
  Active = 'ACTIVE',         // 激活-空缺
  Filled = 'FILLED',         // 已填充
  Vacant = 'VACANT',         // 空缺
  Inactive = 'INACTIVE',     // 停用
  Deleted = 'DELETED'        // 已删除
}

// Employment Types
export enum EmploymentType {
  FullTime = 'FULL_TIME',    // 全职
  PartTime = 'PART_TIME',    // 兼职
  Intern = 'INTERN',         // 实习
  Contract = 'CONTRACT'      // 合同
}

// Assignment Types
export enum PositionAssignmentType {
  Primary = 'PRIMARY',
  Secondary = 'SECONDARY',
  Acting = 'ACTING'
}

// Operation Types (扩展)
export enum PositionOperationType {
  Create = 'CREATE',
  Update = 'UPDATE',
  Fill = 'FILL',             // 填充职位
  Vacate = 'VACATE',         // 空缺职位
  Transfer = 'TRANSFER',     // 转移职位
  Suspend = 'SUSPEND',       // 停用职位
  Reactivate = 'REACTIVATE', // 重新激活
  Delete = 'DELETE'          // 删除职位
}
```

---

## 4. API 契约设计

### 4.1 REST API（命令操作）

遵循 CQRS 架构，所有修改操作通过 REST API（Port 9090）。

#### 4.1.1 创建职位

```yaml
POST /api/v1/positions
Content-Type: application/json
X-Tenant-ID: {tenant-uuid}
Idempotency-Key: {optional-key}

Request Body:
{
  "title": "后端开发工程师",
  "jobProfileCode": "JOB_BACKEND_ENGINEER",
  "jobProfileName": "高级软件工程师",
  "jobFamilyGroupCode": "PROF",
  "jobFamilyGroupRecordId": "uuid-family-group-current",  # 可选，默认取当前版本
  "jobFamilyCode": "PROF-IT",
  "jobFamilyRecordId": "uuid-family-current",
  "jobRoleCode": "PROF-IT-BKND",
  "jobRoleRecordId": "uuid-role-current",
  "jobLevelCode": "P5",
  "jobLevelRecordId": "uuid-level-current",
  "organizationCode": "1000001",
  "positionType": "REGULAR",
  "employmentType": "FULL_TIME",
  "gradeLevel": "P5",
  "headcountCapacity": 3.0,
  "costCenterCode": "CC001",
  "reportsToPositionCode": "P1000000",
  "effectiveDate": "2025-10-15",
  "operationReason": "组织扩编，新增后端开发岗位"
}

Response (201 Created):
{
  "success": true,
  "message": "Position created successfully",
  "data": {
    "code": "P1000001",
    "title": "后端开发工程师",
    "organizationCode": "1000001",
    "jobFamilyGroupCode": "PROF",
    "jobFamilyGroupRecordId": "uuid-family-group",
    "jobFamilyCode": "PROF-IT",
    "jobFamilyRecordId": "uuid-family",
    "jobRoleCode": "PROF-IT-BKND",
    "jobRoleRecordId": "uuid-role",
    "jobLevelCode": "P5",
    "jobLevelRecordId": "uuid-level",
    "headcountCapacity": 3.0,
    "headcountInUse": 0.0,
    "availableHeadcount": 3.0,
    "status": "PLANNED",
    "effectiveDate": "2025-10-15",
    "isCurrent": false,
    "isFuture": true,
    "recordId": "uuid-here"
  },
  "timestamp": "2025-10-12T10:00:00Z",
  "requestId": "req_create_position_001"
}

Required Permissions: position:create

> 若未提供 `job*RecordId`，命令服务默认取各分类的当前版本；提供时将校验与 `job*Code` 及租户匹配，以便支持未来版本的职位创建。
```

#### 4.1.2 更新职位（PUT - 完全替换）

```yaml
PUT /api/v1/positions/{code}
Content-Type: application/json
X-Tenant-ID: {tenant-uuid}

Request Body:
{
  "title": "高级后端开发工程师",
  "jobProfileCode": "JOB_BACKEND_SENIOR",
  "jobProfileName": "资深软件工程师",
  "jobFamilyGroupCode": "PROF",
  "jobFamilyCode": "PROF-IT",
  "jobRoleCode": "PROF-IT-PRIN",
  "jobLevelCode": "P6",
  "organizationCode": "1000001",
  "positionType": "REGULAR",
  "employmentType": "FULL_TIME",
  "gradeLevel": "P6",
  "headcountCapacity": 3.0,
  "effectiveDate": "2025-10-15",
  "operationReason": "职级调整"
}

Required Permissions: position:update
```

#### 4.1.3 创建时态版本

```yaml
POST /api/v1/positions/{code}/versions
Content-Type: application/json
X-Tenant-ID: {tenant-uuid}
Idempotency-Key: {optional-key}

Request Body:
{
  "title": "架构师",
  "jobProfileCode": "JOB_ARCHITECT",
  "jobProfileName": "系统架构师",
  "jobFamilyGroupCode": "PROF",
  "jobFamilyGroupRecordId": "uuid-family-group-future",
  "jobFamilyCode": "PROF-ITARCH",
  "jobFamilyRecordId": "uuid-family-future",
  "jobRoleCode": "PROF-ITARCH-SOL",
  "jobRoleRecordId": "uuid-role-future",
  "jobLevelCode": "P7",
  "jobLevelRecordId": "uuid-level-future",
  "organizationCode": "1000002",
  "gradeLevel": "P7",
  "headcountCapacity": 2.0,
  "effectiveDate": "2026-01-01",
  "operationReason": "职位转型为架构师岗"
}

> `job*RecordId` 建议与 `effectiveDate` 对齐：若插入未来版本，可通过调用分类 `/versions` 返回值获得 `recordId`，以避免在分类切换期间引用旧版本。

Response (201 Created):
{
  "success": true,
  "message": "Position version created successfully",
  "data": {
    "recordId": "new-uuid",
    "code": "P1000001",
    "title": "架构师",
    "effectiveDate": "2026-01-01",
    "isCurrent": false,
    "isFuture": true,
    "jobFamilyGroupCode": "PROF",
    "jobFamilyCode": "PROF-ITARCH",
    "jobRoleCode": "PROF-ITARCH-SOL",
    "jobLevelCode": "P7",
    "headcountCapacity": 2.0,
    "headcountInUse": 0.0,
    "availableHeadcount": 2.0
  },
  "timestamp": "2025-10-12T10:00:00Z",
  "requestId": "req_version_001"
}

Required Permissions: position:create:planned
```

#### 4.1.4 职位填充（Fill Position）

```yaml
POST /api/v1/positions/{code}/fill
Content-Type: application/json
X-Tenant-ID: {tenant-uuid}
Idempotency-Key: {optional-key}

Request Body:
{
  "employeeId": "emp-uuid-123",
  "employeeName": "张三",
  "effectiveDate": "2025-10-20",
  "assignmentType": "PRIMARY",
  "fte": 1.0,
  "operationReason": "新员工入职"
}

Response (200 OK):
{
  "success": true,
  "message": "Position filled successfully",
  "data": {
    "code": "P1000001",
    "status": "FILLED",
    "currentHolderId": "emp-uuid-123",
    "currentHolderName": "张三",
    "currentAssignmentType": "PRIMARY",
    "filledDate": "2025-10-20",
    "assignmentStatus": "LEGACY",
    "headcountCapacity": 3.0,
    "headcountInUse": 1.0,
    "availableHeadcount": 2.0,
    "timeline": [...]
  },
  "timestamp": "2025-10-12T10:00:00Z",
  "requestId": "req_fill_001"
}

Required Permissions: position:fill

- **FTE 参数**：`fte` 默认 1.0，可输入 0 < fte ≤ headcountCapacity 以支持兼职/劳务派遣及物业“一岗多人”场景（例如保洁员职位设置 8 个编制）。命令层按 `fte` 更新 `headcountInUse` 并返回剩余编制。
- **过渡期数据保留**：`assignmentType` 会写入 `positions.current_assignment_type`，待 Phase 4 上线后通过迁移脚本同步到 `position_assignments` 表并移除冗余列。
```

#### 4.1.5 职位空缺（Vacate Position）

```yaml
POST /api/v1/positions/{code}/vacate
Content-Type: application/json
X-Tenant-ID: {tenant-uuid}

Request Body:
{
  "effectiveDate": "2025-11-01",
  "fte": 1.0,
  "operationReason": "员工离职"
}

Response (200 OK):
{
  "success": true,
  "message": "Position vacated successfully",
  "data": {
    "code": "P1000001",
    "status": "VACANT",
    "currentHolderId": null,
    "currentAssignmentType": null,
    "assignmentStatus": "LEGACY",
    "headcountCapacity": 3.0,
    "headcountInUse": 0.0,
    "availableHeadcount": 3.0,
    "timeline": [...]
  },
  "timestamp": "2025-10-12T10:00:00Z",
  "requestId": "req_vacate_001"
}

Required Permissions: position:vacate

- **释放 FTE**：`fte` 默认 1.0，支持按实际离岗人力扣减 `headcountInUse`，适配保洁等一岗多人团队；若传入 `ALL`（后续扩展）将一次性清零。
```

#### 4.1.6 职位转移（Transfer Position）

```yaml
POST /api/v1/positions/{code}/transfer
Content-Type: application/json
X-Tenant-ID: {tenant-uuid}

Request Body:
{
  "targetOrganizationCode": "1000002",
  "effectiveDate": "2025-12-01",
  "operationReason": "组织架构调整"
}

Response (200 OK):
{
  "success": true,
  "message": "Position transferred successfully",
  "data": {
    "code": "P1000001",
    "organizationCode": "1000002",
    "timeline": [...]
  },
  "timestamp": "2025-10-12T10:00:00Z",
  "requestId": "req_transfer_001"
}

Required Permissions: position:transfer
```

#### 4.1.7 职位停用/激活

```yaml
POST /api/v1/positions/{code}/suspend
POST /api/v1/positions/{code}/activate

Request Body:
{
  "effectiveDate": "2025-12-01",
  "operationReason": "业务调整"
}

Required Permissions: position:suspend, position:activate
```

#### 4.1.8 职位事件处理

```yaml
POST /api/v1/positions/{code}/events
Content-Type: application/json
X-Tenant-ID: {tenant-uuid}

Request Body:
{
  "eventType": "DELETE_POSITION",  // or DEACTIVATE
  "recordId": "uuid-optional",
  "effectiveDate": "2026-01-01",
  "changeReason": "职位撤销"
}

Required Permissions: position:modify:history
```

### 4.2 GraphQL API（查询操作）

遵循 CQRS 架构，所有查询操作通过 GraphQL（Port 8090）。

```graphql
type Position {
  code: String!
  tenantId: ID!
  recordId: ID!

  # 基本信息
  title: String!
  jobProfileCode: String
  jobProfileName: String

  # 职位体系化分类
  jobFamilyGroupCode: String!
  jobFamilyGroup: JobFamilyGroup!
  jobFamilyCode: String!
  jobFamily: JobFamily!
  jobRoleCode: String!
  jobRole: JobRole!
  jobLevelCode: String!
  jobLevel: JobLevel!

  # 组织归属
  organizationCode: String!
  organizationName: String
  organization: OrganizationUnit  # 关联查询

  # 职位属性
  positionType: PositionType!
  status: PositionStatus!
  employmentType: EmploymentType!

  # 编制信息
  headcountCapacity: Float!
  headcountInUse: Float!
  availableHeadcount: Float!
  gradeLevel: String
  costCenterCode: String

  # 任职信息
  currentHolderId: ID
  currentHolderName: String
  filledDate: Date
  currentAssignmentType: PositionAssignmentType

  # 汇报关系
  reportsToPositionCode: String
  reportsToPosition: Position
  subordinatePositions: [Position!]!

  # 扩展属性
  profile: JSON

  # 时态字段
  effectiveDate: Date!
  endDate: Date
  isCurrent: Boolean!
  isFuture: Boolean!

  # 审计字段
  createdAt: DateTime!
  updatedAt: DateTime!
  deletedAt: DateTime
  operationType: PositionOperationType!
  operatedBy: OperatedBy!
  operationReason: String
}

enum PositionType {
  REGULAR
  TEMPORARY
  CONTRACT
  INTERN
}

enum PositionStatus {
  PLANNED
  ACTIVE
  FILLED
  VACANT
  INACTIVE
  DELETED
}

enum EmploymentType {
  FULL_TIME
  PART_TIME
  INTERN
  CONTRACT
}

enum PositionAssignmentType {
  PRIMARY
  SECONDARY
  ACTING
}

type JobFamilyGroup {
  code: String!
  recordId: ID!
  name: String!
  description: String
  status: JobCatalogStatus!
  effectiveDate: Date!
  endDate: Date
  isCurrent: Boolean!
  families: [JobFamily!]!
}

type JobFamily {
  code: String!
  recordId: ID!
  name: String!
  description: String
  status: JobCatalogStatus!
  effectiveDate: Date!
  endDate: Date
  isCurrent: Boolean!
  familyGroup: JobFamilyGroup!
  roles: [JobRole!]!
}

type JobRole {
  code: String!
  recordId: ID!
  name: String!
  description: String
  competencyModel: JSON
  status: JobCatalogStatus!
  effectiveDate: Date!
  endDate: Date
  isCurrent: Boolean!
  family: JobFamily!
  levels: [JobLevel!]!
}

type JobLevel {
  code: String!
  recordId: ID!
  name: String!
  levelRank: String!
  description: String
  salaryBand: JSON
  status: JobCatalogStatus!
  effectiveDate: Date!
  endDate: Date
  isCurrent: Boolean!
  role: JobRole!
}

enum JobCatalogStatus {
  ACTIVE
  INACTIVE
  DELETED
}

# 说明：GraphQL 层保留 `status`（业务状态）与 `isCurrent`（是否当前版本）两个维度；客户端可通过 `status=ACTIVE` + `isCurrent=true` 获取当前有效分类，通过 `asOfDate` 查询历史版本。

type Query {
  # 单个职位查询
  position(
    code: String!
    asOfDate: Date
  ): Position

  # 职位列表查询
  positions(
    organizationCode: String
    status: PositionStatus
    positionType: PositionType
    gradeLevel: String
    jobFamilyGroupCode: String
    jobFamilyCode: String
    jobRoleCode: String
    jobLevelCode: String
    isFilled: Boolean
    searchText: String
    page: Int = 1
    pageSize: Int = 20
  ): PositionConnection!

  # 职位时间线
  positionTimeline(
    code: String!
    startDate: Date
    endDate: Date
  ): [Position!]!

  # 组织的所有职位
  positionsByOrganization(
    organizationCode: String!
    includeSubordinates: Boolean = false
    statusFilter: [PositionStatus!]
  ): [Position!]!

  # 空缺职位查询
  vacantPositions(
    organizationCode: String
    gradeLevel: String
    positionType: PositionType
  ): [Position!]!

  # 职位体系化分类
  jobFamilyGroups(includeInactive: Boolean = false, asOfDate: Date): [JobFamilyGroup!]!
  jobFamilies(
    familyGroupCode: String!
    includeInactive: Boolean = false
    asOfDate: Date
  ): [JobFamily!]!
  jobRoles(
    familyCode: String!
    includeInactive: Boolean = false
    asOfDate: Date
  ): [JobRole!]!
  jobLevels(
    roleCode: String!
    includeInactive: Boolean = false
    asOfDate: Date
  ): [JobLevel!]!

  # 职位编制统计
  positionHeadcountStats(
    organizationCode: String
    includeSubordinates: Boolean = true
  ): HeadcountStats!
}

type PositionConnection {
  edges: [PositionEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type PositionEdge {
  node: Position!
  cursor: String!
}

type HeadcountStats {
  organizationCode: String!
  organizationName: String!
  totalCapacity: Float!
  totalFilled: Float!
  totalAvailable: Float!
  fillRate: Float!
  byLevel: [LevelHeadcount!]!
  byType: [TypeHeadcount!]!
}

type LevelHeadcount {
  jobLevelCode: String!
  capacity: Float!
  utilized: Float!
  available: Float!
}

type TypeHeadcount {
  positionType: PositionType!
  capacity: Float!
  filled: Float!
  available: Float!
}
```

- GraphQL `Position.isFuture` 字段由 Resolver 基于 `effectiveDate` 动态推导，与数据库保持派生一致性。
- `headcountCapacity/headcountInUse/availableHeadcount` 由命令服务结合 `positions.headcount_in_use` 与（未来的）`position_assignments` 汇总，支持中国大陆多编制、冻结编制及兼职合规需求。

### 4.3 契约同步与验证计划

- **OpenAPI 更新**：在 REST 端点落库前，先在 `docs/api/openapi.yaml` 中新增 `Position` + `JobCatalog` 时态端点（包括 `/versions`、`/events`、`/sync` 等），沿用 camelCase 字段命名和 `{code}` 路径参数。提交前执行 `node scripts/quality/architecture-validator.js` 以及 `frontend/scripts/validate-field-naming*.js` 验证字段一致性。
- **GraphQL Schema 更新**：在 `docs/api/schema.graphql` 补充 `Position`、`JobFamilyGroup/JobFamily/JobRole/JobLevel` 时态类型与查询输入（带 `asOfDate` 支持）。约束 `enum` 名称与 TypeScript 生成器一致，并运行 `npm run gql:generate`（如已配置）校验。
- **契约差异审阅**：计划进入实现阶段前，输出 `reports/contracts/position-api-diff.md`（由 `scripts/generate-implementation-inventory.js` 及 `contract_diff` 辅助）并在架构评审会上审阅，确保无跨层命名漂移。
- **Mock/Contract Tests**：同步补充契约测试（REST contract + GraphQL schema snapshot），纳入 `make test` 及 `npm run test -- contract` 流程，避免实现先于契约。
- **租户上下文强制**：所有 REST 请求必须携带 `X-Tenant-ID` 头，GraphQL 通过 Context 注入租户；服务端忽略请求体中的任何 `tenantId` 字段，并在缺失或不匹配时返回 401/403。

### 4.4 职位体系化分类 API（时态版本）

- **REST（命令）**：
  - `POST /api/v1/job-family-groups`：创建首个 Job Family Group 版本，要求请求体包含 `effectiveDate`、`status`、名称等字段。
  - `POST /api/v1/job-family-groups/{code}/versions`：在指定 `effectiveDate` 插入新版本，复用 `TemporalCatalogService` 完成重算。（Job Family/Job Role/Job Level 同理）
  - `POST /api/v1/job-family-groups/{code}/events`：支持 `SUSPEND/REACTIVATE/DELETE` 等事件，生成停用或重新启用版本。
  - `DELETE /api/v1/job-family-groups/{code}/versions/{recordId}`：撤销历史或未来版本，自动重算相邻版本边界。
  - 冲突/校验错误使用标准错误码：`409 JOB_CATALOG_POINT_CONFLICT`、`409 JOB_TAXONOMY_MISMATCH`、`403 JOB_CATALOG_TENANT_MISMATCH` 等。
- **GraphQL（查询）**：新增 `jobFamilyGroups(includeInactive, asOfDate)` 以及 `jobFamilies/jobRoles/jobLevels`，客户端可按任意时间点查询分类链路。
- **同步任务**：`POST /api/v1/job-catalog/sync` 支持从 Workday 读取 Job Catalog 变更，按层级顺序重放版本（Family Group → Family → Role → Level），失败时回滚整个事务并写审计日志。

**查询示例**：

```graphql
# 查询组织的所有空缺职位
query VacantPositions {
  vacantPositions(
    organizationCode: "1000001"
    positionType: REGULAR
  ) {
    code
    title
    jobFamilyGroupCode
    jobFamilyGroup { name recordId status effectiveDate }
    jobFamilyCode
    jobFamily { recordId name status }
    jobRoleCode
    jobRole { recordId name status }
    jobLevelCode
    jobLevel { recordId levelRank status }
    gradeLevel
    headcountCapacity
    headcountInUse
    availableHeadcount
    organizationName
    effectiveDate
  }
}

# 查询职位时间线
query PositionHistory {
  positionTimeline(
    code: "P1000001"
    startDate: "2025-01-01"
    endDate: "2025-12-31"
  ) {
    recordId
    title
    status
    effectiveDate
    endDate
    isCurrent
    operationType
    operationReason
  }
}

# 查询职位体系化分类（对标 Workday）
query JobCatalog {
  jobFamilyGroups {
    code
    recordId
    name
    status
    effectiveDate
    families(asOfDate: "2025-10-12") {
      code
      recordId
      name
      status
      roles {
        code
        recordId
        name
        status
        levels {
          code
          recordId
          levelRank
          status
        }
      }
    }
  }
}

# 查询组织编制统计
query HeadcountReport {
  positionHeadcountStats(
    organizationCode: "1000001"
    includeSubordinates: true
  ) {
    organizationName
    totalCapacity
    totalFilled
    totalAvailable
    fillRate
    byLevel {
      jobLevelCode
      capacity
      filled
      available
    }
    byType {
      positionType
      capacity
      filled
      available
    }
  }
}
```

---

## 5. 业务规则

### 5.1 职位创建规则

1. **组织归属验证**：
   - 职位必须归属于有效的组织单元
   - 组织必须处于 ACTIVE 状态
   - effectiveDate 必须在组织有效期内

2. **职位体系校验**：
   - `jobFamilyGroupCode`/`jobFamilyCode`/`jobRoleCode`/`jobLevelCode` 必须来自主数据表且均处于启用状态。
   - `jobProfileCode` 与职位体系映射冲突时拒绝创建。

3. **编码生成**：
   - 自动生成 P + 7位数字
   - 保证租户内唯一性

4. **初始状态**：
   - 当前日期生效 → ACTIVE
   - 未来日期生效 → PLANNED

### 5.2 时态管理规则

**完全复用组织架构的时态规则**：

1. **时态点唯一性**：`(tenant_id, code, effective_date)` 唯一
2. **单一当前版本**：每个 code 仅一个 `is_current=true` 记录
3. **边界自动管理**：
   - 新版本生效时，前一版本的 `end_date` 自动设置为 `effective_date - 1 day`
   - 最后一个版本的 `end_date` 为 NULL（开放式结尾）
4. **删除版本处理**：状态为 DELETED 的版本不参与时态连续性计算

### 5.3 职位填充规则

1. **状态转换**：
   - ACTIVE → FILLED：员工入职
   - FILLED → VACANT：员工离职
   - VACANT → FILLED：重新招聘

2. **编制占用**：
   - `headcountCapacity` 允许设置为 0..N（单位：FTE），0 表示冻结席位，仅用于流程备案。
   - 填充职位时动态累计 `headcountInUse`（默认等于 1.0 FTE，可支持 0.5 等分值），需满足 `headcountInUse + 新任职FTE ≤ headcountCapacity`。
   - 若超过上限，命令返回 `409 POSITION_HEADCOUNT_EXCEEDED`，适配中国大陆企业对批复编制的强管控要求（含兼职/劳务派遣场景）。

3. **FTE 参数**：
   - Fill/Vacate 命令新增 `fte`（默认 1.0），记录单次操作占用/释放的 FTE 数。
   - Phase 4 后 `position_assignments` 将按 `fte` 存储，`headcountInUse` 来自对子表的聚合，便于中国大陆企业的人力成本核算（含劳务派遣、本地化外包）。

4. **时态版本**：
   - 填充/空缺操作创建新的时态版本
   - 保留完整的任职历史
5. **任职类型保留**：
   - Phase 1-2 将 `assignmentType` 写入 `positions.current_assignment_type`
   - Phase 4 迁移至 `position_assignments.assignment_type` 后移除冗余列

### 5.4 职位转移规则

1. **跨组织转移**：
   - 创建新的时态版本
   - 更新 `organizationCode`
   - 保留原组织历史记录

2. **汇报关系调整**：
   - 自动检测循环引用
   - 更新 `reportsToPositionCode`

### 5.5 删除规则

1. **软删除**：
   - 设置 `status = DELETED`
   - 记录 `deleted_at` 时间戳
   - 保留审计历史

2. **级联检查**：
   - 检查是否有下属职位
   - 检查是否有任职记录

### 5.6 时态冲突与并发处理

- **同日多次操作**：`fill → vacate → fill` 等在同一天发生时，命令服务需基于 `effectiveDate` + `operationType` 生成幂等键（建议 `tenantId+code+effectiveDate+operationType`），重复请求直接返回上次结果，防止写入多条相同版本。
- **未来版本撤销**：针对未来填充/转移所写入的计划版本，`cancelFuture` 或重复提交操作需调用统一的 `DeleteVersion` → `RecalculateTimelineInTx`，确保前一版本的 `endDate` 恢复为 NULL。
- **并发写入保护**：沿用组织时间轴的 `SELECT ... FOR UPDATE` 锁相邻版本策略，并要求调用方对同一职位串行提交（API 网关层可按 `code` 节流）。若检测到 `TEMPORAL_POINT_CONFLICT`，返回 409 并提示使用新的 `effectiveDate`。
- **填充与转移冲突**：若在未来日期同时触发 `fill` 与 `transfer`，执行顺序由 `effectiveDate` 决定；如同日，先执行 `transfer`（写入组织变更版本），再执行 `fill`（写入任职版本），并在第二步重算后验证 `organizationCode` 已更新。
- **失败回滚**：任何命令在时间轴重算前失败需显式回滚事务；若重算后响应失败，利用 `audit_logs` 记录的 `operationType` 与 `timeline` 重新推演恢复。

### 5.7 状态转换矩阵

| 当前状态 | 命令 | 目标状态 | 附加说明 |
|----------|------|----------|----------|
| PLANNED | `activate`（effectiveDate ≤ 今天） | ACTIVE | 未来日期则保持 PLANNED；命令返回 200 幂等 |
| PLANNED | `delete` / `events`(DELETE_POSITION) | DELETED | 删除计划职位；`timeline` 不再返回该版本 |
| ACTIVE | `fill` | FILLED | 写入任职信息，记录 `filledDate` |
| ACTIVE | `suspend` | INACTIVE | 可指定未来日期；未来生效则生成计划版本 |
| ACTIVE | `delete` | DELETED | 级联校验下属职位/assignment 后方可执行 |
| FILLED | `vacate` | VACANT | 清空 `current_holder_*` 字段并写入新版本 |
| FILLED | `transfer` | FILLED | 新版本更新组织归属，保持填充状态 |
| FILLED | `suspend` | INACTIVE | 需同时终止任职记录（Assignment Phase 落地后并行执行） |
| VACANT | `fill` | FILLED | 填充后写入 `filledDate` |
| VACANT | `suspend` | INACTIVE | 关闭职位但保留编制 |
| VACANT | `delete` | DELETED | 清理空缺职位记录 |
| INACTIVE | `activate` | ACTIVE | 若指定未来日期，则生成计划版本 |
| INACTIVE | `delete` | DELETED | 删除停用职位 |
| 任意非 DELETED | `transfer` | 同状态 | 转移只更新组织归属，不改变状态 |

- **非法转换处理**：若命令输入的状态与矩阵不符（例如对 `DELETED` 再次 `activate`），返回 409 `POSITION_STATE_CONFLICT`，并在响应中携带当前状态与允许的操作列表。
- **幂等保障**：重复调用矩阵合法的命令（同一 `effectiveDate`）时返回 200 并附带 `idempotencyKey`，避免重复写入。

### 5.8 职位体系化分类规则

- **层级与版本约束**：
  - `jobFamily` 必须引用同一时态链上的 Job Family Group 版本（比较 `parent_record_id` 与岗位请求的 `jobFamilyGroupRecordId`）。
  - 职务、职级引用规则同理；命令层调用 `TemporalCatalogService.ValidateHierarchy` 确认父子生效日期，不满足时返回 400 `JOB_TAXONOMY_MISMATCH`。
  - 若校验发现分类版本所属租户与请求不一致，返回 403 `JOB_CATALOG_TENANT_MISMATCH`。
- **启用状态**：仅允许引用 `status='ACTIVE'` 且 `is_current=true` 的版本；如需引用未来版本，可通过 `asOfDate` 参数显式声明并在职位命令中传递对齐日期。
- **与 Job Profile 同步**：若 `jobProfileCode` 绑定了固定的 Job Family Group 链路，命令服务在创建时自动填充；传入不同链路将触发 409 `JOB_PROFILE_CONFLICT`。
- **职级双写迁移**：`grade_level` 字段在 Phase 2 标记为 deprecated；当 Job Level 时态表稳定后，迁移脚本将在响应中保留 `gradeLevel`（由当前 JobLevel 版本映射）直至前端替换完成。
- **报告与权限**：编制统计、权限控制支持按 `jobFamilyGroup`/`jobFamily` 等过滤，并可基于 `asOfDate` 生成历史报表。

### 5.9 TemporalCatalogService 概述

```
type TemporalCatalogService struct {
  db *sql.DB
  logger *log.Logger
}

func (s *TemporalCatalogService) InsertVersion(ctx context.Context, entity CatalogEntity, req *InsertCatalogVersionRequest) (*CatalogVersion, error)
func (s *TemporalCatalogService) UpdateEffectiveDate(ctx context.Context, entity CatalogEntity, req *UpdateCatalogVersionRequest) (*CatalogVersion, error)
func (s *TemporalCatalogService) DeleteVersion(ctx context.Context, entity CatalogEntity, tenantID uuid.UUID, recordID uuid.UUID) error
func (s *TemporalCatalogService) RecalculateTimeline(ctx context.Context, entity CatalogEntity, tenantID uuid.UUID, code string) ([]CatalogVersion, error)
func (s *TemporalCatalogService) ValidateHierarchy(ctx context.Context, tenantID uuid.UUID, payload HierarchyValidationPayload, asOfDate time.Time) error
```

- `CatalogEntity ∈ {FamilyGroup, Family, Role, Level}`，调用方根据实体类型触发不同的父级校验规则。
- 所有操作使用 `SELECT ... FOR UPDATE` 锁定目标 code 的所有版本，计算 `end_date`、`is_current`，并在父级版本变更时更新子级的 `parent_record_id`。
- 错误模型示例：`JOB_CATALOG_POINT_CONFLICT`（生效日期重复）、`JOB_CATALOG_OVERLAP_CONFLICT`（区间重叠）、`JOB_TAXONOMY_MISMATCH`（父级引用不一致）、`JOB_CATALOG_TENANT_MISMATCH`（跨租户引用）。
- 与职位时态服务共享事务：例如**职位转型 → 新职务/职级生效**流程：
  1. 调用 `InsertVersion(Role, ...)` 写入未来生效的职务版本；服务锁定该 role 的所有版本并重算。
  2. 同一事务内调用 `InsertVersion(Level, ...)`，引用步骤 1 返回的 `recordId` 作为 `parent_record_id`。
  3. 成功后调用 `positions.InsertVersion`，传入上述 `jobRoleRecordId/jobLevelRecordId` 并由 `ValidateHierarchy` 校验父链与租户；若任一步失败整个事务回滚，确保职位不会引用不存在或跨租户的版本。
- 并发控制：对同一 `code` 的分类变更串行执行，服务端采用 `SELECT ... FOR UPDATE` 锁避免竞态；API 网关按 `(tenantId, entityCode)` 节流以减少冲突重试。

### 5.10 Temporal 服务泛化计划

为避免直接复用组织专用实现导致的模型耦合，安排以下重构步骤：

1. **提炼 TemporalCore**：将 `TemporalService` 中与组织实体无关的时间线计算、幂等校验、审计写入抽象为 `TemporalCore` 接口（位于 `internal/services/temporal/core`）。
2. **组织端适配**：现有组织实现迁移到 `OrganizationTemporalAdapter`，仅保留实体映射与校验逻辑，回归后运行回归测试确保无行为漂移。
3. **职位端适配**：新增 `PositionTemporalAdapter`，实现字段映射、命令幂等键生成、组织与职位分类的外键校验。
4. **共用时间线管理器**：保留 `TemporalTimelineManager` 作为共享依赖，通过接口注入避免跨层直接引用组织仓库。
5. **契约验证**：在重构完成后执行组织命令服务的集成测试与计划中的职位命令契约测试，确认 `effectiveDate/endDate/isCurrent` 计算逻辑一致。

---

## 6. 权限定义

扩展现有的 OAuth 2.0 权限体系：

```yaml
# 基础CRUD
position:read              # 读取职位信息
position:create            # 创建职位
position:update            # 更新职位基本信息

# 状态管理
position:fill              # 填充职位
position:vacate            # 空缺职位
position:suspend           # 停用职位
position:activate          # 激活职位

# 职位操作
position:transfer          # 转移职位到其他组织
position:delete            # 删除职位

# 时态数据
position:read:history      # 读取职位历史
position:read:future       # 读取未来生效职位
position:create:planned    # 创建计划职位
position:modify:history    # 修改历史版本

# 统计分析
position:read:stats        # 读取编制统计
position:read:headcount    # 读取人力统计

# 职位体系主数据
job-catalog:read           # 读取 Job Family Group/Job Family/Job Role/Job Level
job-catalog:write          # 维护职位体系主数据
```

---

## 7. 实施路线图

### 7.1 Stage 0：前端布局与交互确认（2周）

**Week 0-1: 布局设计 + Mock 原型**
- [x] 产出职位管理导航结构、列表/详情/编制看板的线框图或 Figma 设计（2025-10-12）
- [x] 与组织模块共享 UI 组件清单，确定复用/新增组件范围（2025-10-12）
- [x] 准备 mock 数据（JSON/fixtures），覆盖多编制、未来版本、空缺等场景（2025-10-13）
- [x] 召开设计评审，确保业务方确认布局与信息密度（2025-10-13）

**Week 0-2: React Mock 页面（无真实 API）**
- [x] 基于 mock 数据搭建职位列表/详情/表单骨架（2025-10-13）
- [x] 实现导航、筛选器、职位编制卡片等核心布局，与实际数据逻辑解耦（2025-10-13）
- [x] 编写 Vitest 组件测试验证列表与详情交互（frontend/src/features/positions/__tests__/PositionDashboard.test.tsx）
- [x] 形成《职位管理前端布局验收报告》，待业务方确认 ✅ 后解锁后续阶段（2025-10-14，参见 06 号进展日志 Stage 1 启动记录）

> ✅ **进入 Stage 1 的前置条件**：完成布局验收，并获得“页面展示框架”确认；未获批准前不得进入真实 API/逻辑开发。

### 7.2 Stage 1：核心职位管理（4周）

**Week 1-2: 后端基础设施**
- [x] 数据库表结构（positions + job catalog 时态表）
- [x] TemporalCore 抽象 + PositionTemporalAdapter 接入
- [x] Position 实体与 Repository（含分类外键校验）
- [x] REST API 命令端点（职位 & 分类的 CRUD + 时态版本）
- [x] 单元测试与集成测试（覆盖分类层级的时间轴重算）

**Week 3: GraphQL 查询层**
- [x] Position GraphQL Schema（支持分类过滤、asOfDate）
- [x] Job Catalog GraphQL Schema（时态查询）
- [x] Resolver 实现 + 数据加载优化（分类缓存）
- [x] 与组织架构的关联查询
- [x] 查询性能优化（索引、缓存）

**Week 4: 前端数据接入**
- [x] 将 Stage 0 已验收的 mock 页面接入真实 GraphQL/REST 数据
- [x] 实现表单提交流程、错误提示及权限控制
- [x] 保留 mock fallback 机制，便于后续迭代验证

> 2025-10-16：参考《06号文档》Stage 2 交付总结，以上 Stage 1 工作已完成并支撑职位生命周期实现。

### 7.3 Stage 2：职位生命周期（3周）

**Week 5-6: 填充与空缺**
- [x] Fill/Vacate 命令端点
- [x] 任职关系基础数据模型
- [x] 前端职位填充流程
- [x] 空缺职位看板（2025-10-17，`PositionVacancyBoard` 发布并接入 GraphQL，见 commit 851da6eb）

**Week 7: 职位转移与调整**
- [x] Transfer 命令端点
- [x] 汇报关系管理
- [x] 前端组织转移界面（2025-10-17，`PositionTransferDialog` 上线，见 commit 851da6eb）

> 2025-10-16：参见《06号文档》Stage 2 交付总结，命令/查询服务与填充流程已上线；空缺看板与组织转移前端界面转入 Stage 3 优先事项。

### 7.4 Stage 3：编制与统计（2周）

**Week 8: 编制统计**
- [x] Headcount 统计 GraphQL — `cmd/organization-query-service/internal/graphql/resolver.go` 现从 GraphQL 上下文解析租户并透传给仓储，同时输出 byFamily 聚合。
- [x] 编制分析报表 — `frontend/src/shared/hooks/useEnterprisePositions.ts` 请求补齐 byFamily 字段，`PositionHeadcountDashboard` 获取完整统计以驱动报表导出。
- [x] 前端编制看板（`PositionHeadcountDashboard`，commit c2481957）
- [x] 空缺职位看板（Stage 2 尾项已完成，复用 Week 5-6 成果）

**Week 9: 集成与优化**
- [x] 前端组织转移界面（Stage 2 尾项已完成，见 Week 7）
- [x] E2E 测试 — `frontend/tests/e2e/position-lifecycle.spec.ts` 校验编制看板含 byFamily 表格，Vitest + Playwright 均通过。
- [x] 性能优化 — 复核 `GetPositionHeadcountStats` SQL 聚合沿用租户/组织索引，无需新增索引；确认仅扫描当前租户现势记录。
- [x] 文档完善 — 更新 06 号进展日志与 85 号执行计划收尾说明，准备归档 Stage 3。

> **2025-10-17 更新**：Stage 3 已批准启动 ✅
> - **评审通过**：85号执行计划 v0.2 复审通过（A级 4.8/5分），所有 P0/P1 问题已修复。
> - **前置核查完成**：`positionHeadcountStats` Schema 已定义，查询服务已实现，`vacantPositions` Resolver 已就绪。
> - **时间确认**：维持 2 周交付节奏，Week 1 完成空缺看板+转移界面+统计 API，Week 2 完成编制看板+E2E+文档。
> - **执行计划**：详见 `docs/archive/development-plans/85-position-stage3-execution-plan.md` v0.2。
> - **最终验收**：2025-10-17 通过 GraphQL/前端测试确认编制统计链路完成，85 号计划已归档。

### 7.5 Stage 4（未来扩展）：任职管理

- ✅ Stage 4 增量（2025-10-21）：根据 86 号计划交付 Position Assignment 专用 API、代理自动恢复、历史视图增强及跨租户验证。
- ⏳ 后续扩展：如需支持多重任职（Secondary/Acting 更高级场景）或后续运营需求，将另建新计划跟踪。

### 7.6 Assignment 依赖与临时策略

- ✅ **Phase 1-2 过渡方案回收**：`current_holder_*` 与 `current_assignment_type` 冗余字段已通过 045 迁移移除，命令服务全面改写至 `position_assignments`。
- ✅ **落地条件完成**：GraphQL/REST 契约已补充 `PositionAssignment` 资源，命令服务与仓储层支持 `position_assignments` + `positions` 同事务写入与回滚。
- ⏳ **后续迭代**：多重任职等新增需求将通过新计划执行。
  3. 对接员工基础数据服务，确认 `employeeId` 的唯一事实来源和授权范围。
- **数据回填策略**：Assignment 表上线时，通过迁移脚本将现有 `positions.current_holder_*` 与 `current_assignment_type` 数据转存为首个 `assignment` 记录，并设置 `effective_date = filled_date`。迁移完成后清理冗余字段，避免双写。

---

## 8. 技术栈与架构对齐

### 8.1 后端技术栈

| 模块 | 技术选型 | 说明 |
|------|---------|------|
| 语言 | Go 1.21+ | 与组织架构服务保持一致 |
| 框架 | Go Fiber | 轻量级高性能框架 |
| 数据库 | PostgreSQL 14+ | 复用现有数据库 |
| GraphQL | gqlgen | 与查询服务保持一致 |
| 时态管理 | TemporalCore + Adapter | 先抽象核心，再提供组织/职位适配层 |

### 8.2 前端技术栈

| 模块 | 技术选型 | 说明 |
|------|---------|------|
| 框架 | React 18 + TypeScript | 与组织架构前端保持一致 |
| 状态管理 | React Query | 复用现有查询客户端 |
| UI组件 | Canvas Kit + shadcn/ui | 保持设计一致性 |
| 表单 | React Hook Form + Zod | 复用验证架构 |
| 路由 | React Router v6 | 标准路由方案 |

### 8.3 架构对齐清单

- [x] **CQRS 架构**：REST 命令 + GraphQL 查询
- [x] **多租户隔离**：所有表包含 tenant_id
- [x] **时态管理模式**：完全复用 effectiveDate/endDate/isCurrent
- [x] **审计追踪**：operation_type, operated_by, operation_reason
- [x] **幂等性支持**：Idempotency-Key 头
- [x] **乐观锁**：If-Match ETag 头（可选）
- [x] **统一错误响应**：ErrorResponse envelope
- [x] **权限体系**：OAuth 2.0 Client Credentials

### 8.4 前端页面复用与导航拓展

- **整体布局**：沿用组织管理模块的 Canvas Kit 布局（左侧导航 + 顶部工具栏 + 右侧详情区域），确保视觉与交互一致。
- **一级导航**：在原有“组织管理”同级新增“职位管理”入口，与现有风格保持一致。
- **二级菜单（Workday Canvas Kit 最佳实践）**：
  - 职位列表（默认页，复用组织列表的表格组件）
  - 最广泛职能分组（Job Family Group 列表）
  - 职种（Job Family 列表）
  - 职务（Job Role 列表）
  - 职级（Job Level 列表）
- **内容区复用**：
  - 列表页使用组织模块的 `DataGrid` 风格与筛选器（调整列字段即可）。
  - 详情/表单页复用 Drawer/Panel 结构，表单校验沿用 React Hook Form + Zod 组合，只替换字段定义。
- **路由规划**：`/positions`（职位）、`/positions/catalog/categories` 等路径挂载在同一 `PositionsLayout` 下，与组织模块共享 `AppShell`。
- **权限控制**：前端菜单渲染基于新增的 `job-catalog:read` 等权限，与组织模块权限组件共享逻辑。
- **改进目标**：通过二级导航将 Job Catalog 各层呈现为独立对象模型，便于运维按 Workday 规范管理主数据，同时保持用户在组织模块已有的操作习惯。
- **后续补充**：
  - `TODO`：整理职位页面复用的具体组件映射（例如组织模块的 `OrganizationListTable` → `PositionListTable` 对照表）。
  - `TODO`：绘制导航结构图（一级/二级菜单及路由关系），确保前端与设计团队共识。

---

## 9. 数据迁移与兼容性

### 9.1 初始数据导入

```sql
-- Step 1: 导入 Job Catalog 历史版本（示例：Job Family Group）
INSERT INTO job_family_groups (
    record_id, tenant_id, family_group_code, name, description,
    status, effective_date, end_date, is_current, created_at, updated_at
)
SELECT
    COALESCE(legacy_record_id, gen_random_uuid()),
    tenant_id,
    family_group_code,
    family_group_name,
    family_group_desc,
    status,
    effective_date,
    end_date,
    is_current,
    NOW(),
    NOW()
FROM legacy_job_family_groups
ORDER BY effective_date;

-- Step 2: 导入职位版本（需先查到分类 record_id）
WITH catalog AS (
  SELECT family_group_code, name, record_id AS family_group_record_id
  FROM job_family_groups WHERE tenant_id = '987fcdeb-51a2-43d7-8f9e-123456789012'
    AND is_current = true
), families AS (
  SELECT family_code, name, record_id AS family_record_id
  FROM job_families WHERE tenant_id = '987fcdeb-51a2-43d7-8f9e-123456789012'
    AND is_current = true
), roles AS (
  SELECT role_code, name, record_id AS role_record_id
  FROM job_roles WHERE tenant_id = '987fcdeb-51a2-43d7-8f9e-123456789012'
    AND is_current = true
), levels AS (
  SELECT level_code, name, record_id AS level_record_id
  FROM job_levels WHERE tenant_id = '987fcdeb-51a2-43d7-8f9e-123456789012'
    AND is_current = true
)
INSERT INTO positions (
    code, tenant_id, record_id, title, organization_code,
    job_profile_code, job_profile_name,
    job_family_group_code, job_family_group_name, job_family_group_record_id,
    job_family_code, job_family_name, job_family_record_id,
    job_role_code, job_role_name, job_role_record_id,
    job_level_code, job_level_name, job_level_record_id,
    grade_level,
    headcount_capacity, headcount_in_use,
    position_type, employment_type, status,
    effective_date, end_date, is_current,
    created_at, updated_at,
    operation_type, operated_by_id, operated_by_name, operation_reason
)
SELECT
    'P' || LPAD(lp.legacy_id::text, 7, '0'),
    lp.tenant_id,
    COALESCE(lp.record_id, gen_random_uuid()),
    lp.position_title,
    lp.org_code,
    lp.job_profile_code,
    lp.job_profile_name,
    lp.job_family_group_code,
    lc.name,
    lc.family_group_record_id,
    lp.job_family_code,
    lf.name,
    lf.family_record_id,
    lp.job_role_code,
    lr.name,
    lr.role_record_id,
    lp.job_level_code,
    ll.name,
    ll.level_record_id,
    lp.job_level_code,
    lp.headcount_capacity,
    COALESCE(lp.headcount_in_use, 0),
    lp.position_type,
    lp.employment_type,
    lp.status,
    lp.effective_from,
    lp.end_date,
    lp.is_current,
    NOW(),
    NOW(),
    'CREATE',
    'migration-user-uuid'::uuid,
    'System Migration',
    lp.migration_reason
FROM legacy_positions lp
JOIN catalog lc ON lc.family_group_code = lp.job_family_group_code
JOIN families lf ON lf.family_code = lp.job_family_code
JOIN roles lr ON lr.role_code = lp.job_role_code
JOIN levels ll ON ll.level_code = lp.job_level_code
WHERE lp.tenant_id = '987fcdeb-51a2-43d7-8f9e-123456789012';

-- Step 3: 重算职位时间线
SELECT temporal_timeline_manager_recalculate('987fcdeb-51a2-43d7-8f9e-123456789012'::uuid, code)
FROM (SELECT DISTINCT code FROM positions WHERE tenant_id = '987fcdeb-51a2-43d7-8f9e-123456789012') t;
```

### 9.2 与现有系统集成

**CoreHR 集成点**：
- 员工信息查询（employee_id → employee_name）
- 岗位定义同步（job_profile_code）
- 编制数据导入

**Workday 集成点**（如需要）：
- Position 数据同步
- Worker Assignment 同步
- 组织架构对齐

### 9.3 迁移执行顺序

1. **契约对齐**：在迁移窗口前合并 positions 与 job catalog 的 OpenAPI/GraphQL 契约，运行 `node scripts/generate-implementation-inventory.js` 生成新资源清单并备份至 `reports/contracts/positions-pre-migration.json`。
2. **结构迁移**：执行 `make db-migrate-all` 应用 positions 及 job catalog 时态表；确认所有表具备部分唯一索引与外键（包含 `parent_record_id`）。
3. **分类主数据导入**：按 Family Group → Family → Role → Level 顺序导入历史版本，确保每条记录包含 `effective_date/end_date/is_current`。导入后调用 `TemporalCatalogService.RecalculateTimeline` 巡检无断档。
4. **基础数据导入**：分批导入职位版本，在同一事务内校验分类引用是否指向匹配的 `record_id` 与租户。每批导入后运行 `ANALYZE positions`。
5. **历史版本回放**：若遗留系统提供时间线数据，按生效日期升序重放，并在每批后调用 `TemporalTimelineManager.RecalculateTimeline` 巡检断档。
6. **编制数据校验**：导入完成后执行 `reports/headcount-validation.sql` 与 `reports/job-catalog-consistency.sql`，核对分类链路、编制合计与源系统一致。
7. **租户隔离巡检**：运行 `SELECT COUNT(*) FROM positions p JOIN job_family_groups jfg ON p.job_family_group_record_id = jfg.record_id WHERE p.tenant_id <> jfg.tenant_id;`（及 family/role/level 类似查询）确认跨租户引用为零。

### 9.4 验证与回滚

- **一致性检查**：
  - `SELECT COUNT(*) FROM positions WHERE is_current = true` 与源系统的活跃职位数对比。
  - `SELECT code FROM positions GROUP BY code HAVING COUNT(*) FILTER (WHERE is_current) > 1` 应为空。
  - `SELECT code FROM positions WHERE effective_date > end_date` 应为空。
  - `SELECT code FROM job_family_groups GROUP BY code HAVING COUNT(*) FILTER (WHERE is_current) > 1` 应为空。
  - `SELECT role_code FROM job_roles jr WHERE NOT EXISTS (SELECT 1 FROM job_levels jl WHERE jl.role_code = jr.role_code AND jl.is_current)` 应为空，用于校验角色与职级映射完整。
  - `SELECT p.code FROM positions p LEFT JOIN job_roles jr ON p.job_role_record_id = jr.record_id WHERE jr.record_id IS NULL` 应为空，确保职位引用的分类版本存在。
  - `SELECT COUNT(*) FROM positions p JOIN job_family_groups jfg ON p.job_family_group_record_id = jfg.record_id WHERE p.tenant_id <> jfg.tenant_id` 应为 0，其他层级同理，确保无跨租户引用。
- **影子验证**：准备影子环境重放职位填充/转移流程，捕获 `audit_logs` 与 `positions` 差异并导出 `reports/migration/positions-audit.csv`。
- **回滚策略**：保留迁移前 `pg_dump positions`、`pg_dump audit_logs` 备份（命名 `backup/positions-pre-YYYYMMDD.sql`），如发现严重偏差，通过 `psql -f` 方式回滚并重新运行迁移脚本。

---

## 10. 验收标准

### 10.1 功能验收

- [ ] **职位CRUD**：创建、读取、更新、删除职位
- [ ] **时态管理**：支持历史版本、当前版本、未来版本
- [ ] **职位填充**：填充职位、空缺职位、查看任职历史
- [ ] **职位转移**：跨组织转移、汇报关系调整
- [ ] **权限控制**：19个职位权限生效
- [ ] **编制统计**：按组织、职级、类型统计编制

### 10.2 性能验收

- [ ] 职位列表查询 < 200ms（1000条记录）
- [ ] 职位详情查询 < 50ms
- [ ] 时态版本创建 < 100ms
- [ ] 编制统计查询 < 500ms

### 10.3 质量验收

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖核心场景
- [ ] E2E 测试通过（创建→填充→空缺→删除）
- [ ] API 文档完整（OpenAPI + GraphQL Schema）

---

## 11. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| 时态逻辑复杂度 | 高 | 中 | 先抽象 TemporalCore，再为职位提供独立适配层，复用成熟算法 |
| 与员工系统集成 | 中 | 高 | Phase 1 先用 employee_id 字符串，Phase 4 再深度集成 |
| 编制计算准确性 | 高 | 中 | 引入 headcount_capacity/headcount_in_use 字段，支持 0..N FTE 与部分工时 |
| 职位体系主数据漂移 | 高 | 中 | 建立 job catalog 主数据同步与审计（REST + GraphQL 查询），并在迁移阶段执行层级一致性校验 |
| 性能问题 | 中 | 低 | 合理索引 + 查询缓存 + 分页 |
| 数据迁移风险 | 中 | 中 | 提供迁移脚本模板，分批次迁移验证 |
| 租户隔离缺口 | 高 | 低 | 采用复合外键与租户校验双保险，迁移/巡检运行跨租户 SQL 并补充回归测试 |

---

## 12. 参考资料

### 12.1 Workday 文档

- [Workday Position Management Overview](https://www.workday.com/)
- [Effective Dating for Service Dates](https://kognitivinc.com/blog/effective-dating-for-service-dates-absence-impacts-and-how-to-prepare/)
- [Organization Management in Workday](https://www.workday.com/content/dam/web/en-us/documents/datasheets/organization-management-in-workday-datasheet-en-us.pdf)
- [Workday Job Catalog and Job Profiles](https://community.workday.com/)（需 Workday 社区权限）

### 12.2 项目内部文档

- `docs/api/openapi.yaml` - 组织架构 API 契约（参考时态模式）
- `docs/api/schema.graphql` - GraphQL Schema（参考查询模式）
- `docs/development-plans/60-system-wide-quality-refactor-plan.md` - 质量重构总计划
- `docs/development-plans/61-system-quality-refactor-execution-plan.md` - 执行计划
- `cmd/organization-command-service/internal/services/temporal/` - 时态服务实现

### 12.3 相关标准

- [ISO 8601](https://www.iso.org/iso-8601-date-and-time-format.html) - 日期时间格式
- [OAuth 2.0 RFC 6749](https://datatracker.ietf.org/doc/html/rfc6749) - 认证授权
- [GraphQL Specification](https://spec.graphql.org/) - GraphQL 规范

---

## 13. 后续行动

1. **架构评审**：提交架构组评审（预计1周）
2. **技术预研**：时态逻辑复用验证（3天）
3. **详细设计**：REST API 详细设计文档（1周）
4. **原型开发**：核心功能 POC（2周）
5. **正式开发**：按实施路线图执行

---

## 14. 变更记录

| 版本 | 日期 | 修改内容 | 修改人 |
|------|------|---------|--------|
| v1.0 | 2025-10-12 | 初始版本，完整方案设计 | Claude Code Assistant |

---

**附件**：
- Workday Position Management 最佳实践研究
- 组织架构时态管理复用分析
- 职位管理 ER 图
- API 契约详细设计（待补充）
- **租户上下文强制**：所有 REST 请求必须携带 `X-Tenant-ID` 头，GraphQL 通过 Context 注入租户；服务端忽略请求体中的任何 `tenantId` 字段，并在缺失或不匹配时返回 401/403。
- [ ] 编写跨租户引用回归测试（REST + GraphQL），确保 403 `JOB_CATALOG_TENANT_MISMATCH` 与数据库外键协同生效
