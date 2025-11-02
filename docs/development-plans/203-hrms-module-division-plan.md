# 203号方案：HRMS系统模块化演进与领域划分

**版本**: v2.0
**创建日期**: 2025-11-03
**作者**: 架构组
**状态**: 提案（与200/201文档对齐完成，已升级为v2.0）
**关联文档**:
- `79-peoplesoft-corehr-menu-reference.md` (功能蓝图)
- `200-Go语言ERP系统最佳实践.md` (架构原则)
- `201-Go实践对齐分析.md` (项目现状分析)
- `204-HRMS-Implementation-Roadmap.md` (实施路线图)
- `205-HRMS-Transition-Plan.md` (过渡方案)
- `206-Alignment-With-200-201.md` (对齐分析)

---

## 1. 核心建议与指导原则

### 1.1 核心建议

为支撑项目从单一的“组织管理”功能扩展至完整的 HRMS（人力资源管理系统），我们必须采用一种具备长期可维护性、可扩展性的架构演进策略。

**核心建议**：以**领域驱动设计（DDD）**的“界定上下文（Bounded Context）”为指导，构建一个**模块化单体（Modular Monolith）**架构。

我们不应简单地按功能菜单创建文件夹，而应将 `79号文档` 中定义的22个功能模块识别并聚合成不同的业务领域。每个领域都是一个高内聚、低耦合的业务单元，它们将成为我们“模块化单体”中的一级模块。

### 1.2 指导原则

此方案严格遵循 `200号文档` 的核心架构原则：
1.  **从模块化单体开始**：避免过早引入微服务的复杂性，在单一进程内实现清晰的逻辑边界（参考 `200号文档:73-75`）。
2.  **模块与界定上下文对齐**：模块的划分必须反映真实的业务领域边界，而非技术或数据表结构（参考 `200号文档:119-121`）。
3.  **演进式架构**：该架构支持在未来必要时，将特定模块平滑地演进为独立的微服务。

---

## 2. 模块划分蓝图：三层领域模型

根据 `79号文档` 的功能范围和业务关联性，建议将 HRMS 系统划分为三大领域（Domain）和多个界定上下文（Bounded Context）。

### 2.1 核心人力领域（Core HR Domain）

这是整个 HRMS 的基石，包含最稳定、最核心的人员和组织数据。

| 界定上下文 (模块名) | 包含的 PeopleSoft 模块 (来自79号文档) | 核心职责 |
| :--- | :--- | :--- |
| **`organization`** | 1. 组织管理, 3. 职位管理, 5. 工作信息 | 负责企业的组织、部门、职位、职级、汇报线等“结构性”数据。**这是当前已有的模块，是很好的起点。** |
| **`workforce`** | 2. 人员管理, 4. 人事管理 | 负责员工的“档案”和“生命周期事件”，如员工主数据、入职、转岗、晋升、离职等。 |
| **`contract`** | 22. 劳动合同管理 | 专门处理劳动合同的签署、续签、变更、终止。因其极强的合规和法律属性，从 `workforce` 中独立。 |

### 2.2 人才管理领域（Talent Management Domain）

这个领域围绕员工的“选、用、育、留”展开，业务变化相对频繁。

| 界定上下文 (模块名) | 包含的 PeopleSoft 模块 (来自79号文档) | 核心职责 |
| :--- | :--- | :--- |
| **`recruitment`** | 11. 招聘管理 | 从职位发布到 Offer 的完整招聘流程。 |
| **`performance`** | 12. 绩效管理 | 目标设定（OKR/KPI）、绩效评估、绩效校准与反馈。 |
| **`development`** | 13. 培训与发展, 14. 人才管理 | 员工培训、继任计划、人才盘点、职业发展路径。 |

### 2.3 薪酬与运营领域（Compensation & Operations Domain）

这个领域处理与“算钱”、“算时间”相关的复杂业务，规则性强，计算密集。

| 界定上下文 (模块名) | 包含的 PeopleSoft 模块 (来自79号文档) | 核心职责 |
| :--- | :--- | :--- |
| **`compensation`** | 6. 薪酬管理, 7. 福利管理 | 负责薪酬结构、薪资等级、调薪、福利方案的设计与管理。它定义“钱怎么算”的规则。 |
| **`payroll`** | 15. 薪资计算 | 负责每月具体的薪资、社保、个税计算和发放。它执行 `compensation` 定义的规则。 |
| **`attendance`** | 8. 时间与考勤, 17. 缺勤管理 | 负责排班、工时、考勤记录、假期额度与休假申请。 |
| **`compliance`** | 16. 合规管理, 18. 员工关系, 20. 健康安全 | 处理劳动法规、员工关系事件、安全事件等合规与风险事宜。 |

---

## 3. 横向支撑能力 (Cross-Cutting Concerns)

`79号文档` 中提到的 **9. 自助服务 (ESS/MSS)**、**10. 报表与分析** 和 **21. 横向支撑功能** 不应作为独立的业务模块。它们是为所有业务模块提供支持的横向能力，应在架构的基础设施层或共享层实现。

- **自助服务 (Self-Service)**: 本质是前端应用，根据用户角色（员工/经理）调用不同业务模块的API。
- **报表与分析 (Reporting)**: 可以是一个独立的只读服务，也可以是每个模块内建的查询能力。
- **工作流/通知 (Workflow/Notification)**: 应作为共享的 `pkg` 或基础服务，供所有模块调用。

---

## 4. 实施策略与项目结构

### 4.1 演进式构建路径

1.  **夯实基础（当前阶段）**：你已经有了 `organization` 模块，这是完美的开始。下一步，应建立 `workforce` 模块。这两个模块关系最紧密，是 Core HR 的核心。
2.  **逐步构建**：按照上面的领域划分，逐个实现新的界定上下文（模块）。例如，在完成 Core HR 领域后，可以开始构建人才管理领域中的 `performance` 模块。

### 4.2 模块化单体项目结构

基于 `200号文档` 推荐的模块化单体结构，规划的未来项目目录如下：

```
/cube-castle/
├── cmd/
│   └── hrms-server/          # 统一的服务入口
│       └── main.go           # 在这里合并所有模块的路由和依赖
├── internal/
│   ├── organization/         # ✅ 组织上下文 (已有)
│   │   ├── api/              # 模块的公开接口定义 (端口)
│   │   └── internal/         # 模块内部实现 (handler, service, repo)
│   ├── workforce/            # 🆕 人员上下文
│   ├── contract/             # 🆕 合同上下文
│   ├── performance/          # 🆕 绩效上下文
│   ├── compensation/         # 🆕 薪酬上下文
│   ├── payroll/              # 🆕 薪资上下文
│   ├── attendance/           # 🆕 考勤上下文
│   └── ...                   # 其他业务模块
├── pkg/
│   ├── eventbus/             # 共享的内存事件总线
│   ├── auth/                 # 共享的认证/授权逻辑
│   ├── database/             # 共享的数据库连接
│   └── logger/               # 共享的日志工具
├── docs/
│   ├── api/
│   │   ├── openapi.yaml      # 统一的REST API契约
│   │   └── schema.graphql    # 统一的GraphQL Schema
│   └── ...
└── go.mod
```

### 4.3 模块间通信机制

#### 同步调用（依赖注入）

严格禁止模块间直接调用内部代码。必须通过 Go 的 `interface` 定义"端口"，在 `main.go` 中进行依赖注入。这完全符合 `200号文档` 的"端口与适配器"最佳实践（`200号文档:142-196`）。

**示例代码**：

```go
// internal/workforce/api.go - workforce模块的公开接口
package workforce

import "context"

// EmployeeAPI 是 workforce 模块暴露给其他模块的接口
type EmployeeAPI interface {
    GetEmployee(ctx context.Context, employeeID string) (*Employee, error)
    UpdateEmployeeStatus(ctx context.Context, employeeID string, status string) error
}

// internal/workforce/internal/service.go - 实现
type Service struct {
    repo EmployeeRepository
}

func (s *Service) GetEmployee(ctx context.Context, employeeID string) (*Employee, error) {
    return s.repo.GetByID(ctx, employeeID)
}

func (s *Service) UpdateEmployeeStatus(ctx context.Context, employeeID string, status string) error {
    employee, err := s.repo.GetByID(ctx, employeeID)
    if err != nil {
        return err
    }
    employee.Status = status
    return s.repo.Update(ctx, employee)
}

// cmd/hrms-server/main.go - 依赖注入
func main() {
    // 初始化 workforce 服务
    workforceService := workforce.NewService(db, logger)

    // 初始化 payroll 服务，注入 workforce 的依赖
    payrollService := payroll.NewService(db, logger, workforceService)

    // payroll 模块通过 interface 调用 workforce
    // payroll 不能直接导入 workforce/internal
}

// internal/payroll/internal/service.go - payroll模块使用workforce接口
type PayrollService struct {
    workforceAPI workforce.EmployeeAPI  // 仅依赖公开接口
    repo         PayrollRepository
}

func (s *PayrollService) CalculatePayroll(ctx context.Context, employeeID string, month string) error {
    // 通过接口调用，而非直接导入
    employee, err := s.workforceAPI.GetEmployee(ctx, employeeID)
    if err != nil {
        return err
    }
    // 继续处理薪资计算...
    return nil
}
```

#### 异步通信（事件总线 + 事务性发件箱）

使用**事务性发件箱（Transactional Outbox）**模式和内存事件总线（In-Memory Event Bus）进行模块解耦。例如，当 `workforce` 模块处理完一个"离职"事件后，它不直接调用 `payroll`，而是发布一个 `EmployeeTerminated` 事件。`payroll` 模块订阅此事件，并异步执行停薪、停保等操作。这同样是 `200号文档` 强烈推荐的核心实践（`200号文档:368-399`）。

**示例代码**：

```go
// pkg/eventbus/eventbus.go - 共享事件总线
package eventbus

type Event interface {
    EventType() string
    AggregateID() string
}

type EventBus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventType string, handler EventHandler) error
}

type EventHandler func(ctx context.Context, event Event) error

// internal/eventbus/memory_bus.go - 内存实现
type MemoryEventBus struct {
    handlers map[string][]EventHandler
    mu       sync.RWMutex
}

func (b *MemoryEventBus) Publish(ctx context.Context, event Event) error {
    b.mu.RLock()
    handlers, ok := b.handlers[event.EventType()]
    b.mu.RUnlock()

    if !ok {
        return nil // 无订阅者，正常返回
    }

    for _, handler := range handlers {
        if err := handler(ctx, event); err != nil {
            // 记录错误但继续处理其他订阅者
            logger.Error("event handler failed", "event", event.EventType(), "error", err)
        }
    }
    return nil
}

// internal/workforce/internal/domain/events.go - workforce域事件
package domain

type EmployeeTerminatedEvent struct {
    EmployeeID     string
    TerminationDate time.Time
    Reason         string
}

func (e EmployeeTerminatedEvent) EventType() string {
    return "employee.terminated"
}

func (e EmployeeTerminatedEvent) AggregateID() string {
    return e.EmployeeID
}

// internal/workforce/internal/service.go - 发布事件
func (s *Service) TerminateEmployee(ctx context.Context, employeeID string, reason string) error {
    // 1. 更新员工状态（在事务内）
    employee, err := s.repo.GetByID(ctx, employeeID)
    if err != nil {
        return err
    }

    employee.Status = "terminated"
    employee.TerminationDate = time.Now()

    // 2. 保存员工变更和发件箱事件（同一事务）
    err = s.repo.WithTx(ctx, func(txRepo EmployeeRepository) error {
        if err := txRepo.Update(ctx, employee); err != nil {
            return err
        }

        // 将事件保存到 outbox 表
        event := domain.EmployeeTerminatedEvent{
            EmployeeID:      employeeID,
            TerminationDate: time.Now(),
            Reason:          reason,
        }
        return s.outboxRepo.SaveEvent(ctx, event)
    })

    if err != nil {
        return err
    }

    // 3. 异步发布事件（事务已提交）
    go func() {
        event := domain.EmployeeTerminatedEvent{
            EmployeeID:      employeeID,
            TerminationDate: time.Now(),
            Reason:          reason,
        }
        _ = s.eventBus.Publish(context.Background(), event)
    }()

    return nil
}

// internal/payroll/internal/handlers/events.go - payroll订阅事件
package handlers

type EmployeeTerminationHandler struct {
    payrollService PayrollService
    logger         Logger
}

func (h *EmployeeTerminationHandler) Handle(ctx context.Context, event eventbus.Event) error {
    terminatedEvent, ok := event.(domain.EmployeeTerminatedEvent)
    if !ok {
        return nil
    }

    // 异步执行停薪、停保等操作
    h.logger.Info("Processing employee termination", "employeeID", terminatedEvent.EmployeeID)
    return h.payrollService.TerminatePayroll(ctx, terminatedEvent.EmployeeID, terminatedEvent.TerminationDate)
}

// cmd/hrms-server/main.go - 注册事件订阅
func main() {
    // 初始化事件总线
    eventBus := eventbus.NewMemoryEventBus()

    // 初始化各模块服务
    workforceService := workforce.NewService(db, logger, eventBus)
    payrollService := payroll.NewService(db, logger)

    // 注册事件处理器
    terminationHandler := handlers.NewEmployeeTerminationHandler(payrollService, logger)
    eventBus.Subscribe("employee.terminated", terminationHandler.Handle)

    // 启动应用...
}
```

### 4.3.3 强制要求：事务性发件箱（Transactional Outbox）

#### ⚠️ 为什么必须使用

根据200号文档第341-399行的分析，**纯内存事件总线存在致命缺陷**。项目当前的CascadeUpdateService使用内存队列，存在以下风险：

```
时间线中的崩溃点风险：
1. ✅ 数据库事务提交 → 员工状态变更成功
2. ❌ [应用崩溃/重启点]
3. ❌ 事件永不被发布
4. ❌ 财务系统永不收到通知
5. ❌ 数据永久不一致
   - 查询服务缓存永不失效
   - 审计日志缺失（合规风险）
   - 跨模块业务错误
```

**因此，所有模块间的异步通信都必须采用事务性发件箱模式**。这不是可选项。

#### 标准表设计（强制）

```sql
-- 所有模块共享的 outbox 表设计
CREATE TABLE outbox_events (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,           -- 幂等ID（生成自业务事件）
    aggregate_id VARCHAR(255) NOT NULL,      -- 聚合根ID（如employeeID）
    aggregate_type VARCHAR(100) NOT NULL,    -- 聚合根类型（如"employee"）
    event_type VARCHAR(100) NOT NULL,        -- 事件类型（如"employee.terminated"）
    payload JSONB NOT NULL,                  -- 事件负载
    retry_count INTEGER DEFAULT 0,           -- 重试次数
    published BOOLEAN DEFAULT FALSE,         -- 是否已发布
    published_at TIMESTAMP,                  -- 发布时间
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_outbox_unpublished
    ON outbox_events(published, created_at)
    WHERE published = FALSE;
```

> **说明**：`event_id` 由业务层生成（推荐 UUIDv7 或雪花算法），用于在发布器与消费者两侧实现幂等保障。去除聚合级唯一约束后，可以安全地记录多次相同类型的业务事件（如多次调岗），同时依赖 `event_id` 防止重复投递。

#### 实现要求（必须满足）

**任何跨模块的业务操作必须遵循**：

1. **单事务内完成**：业务变更 + outbox 插入在同一事务
   ```go
   tx.BeginTx()
   // 1. 执行业务操作
   tx.Exec("UPDATE employees SET status='ACTIVE' WHERE id=?", empID)
   // 2. 插入事件到 outbox（同一事务）
   tx.Exec(`INSERT INTO outbox_events (event_id, aggregate_id, aggregate_type, event_type, payload)
            VALUES ($1, $2, $3, $4, $5)`, eventID, aggregateID, aggregateType, eventType, payload)
   // 3. 原子提交（要么都成功，要么都失败）
   tx.Commit()
   ```

2. **异步中继**：后台goroutine轮询并发布
   ```go
   // 后台中继器（每个模块必须实现）
   func (r *EventRelay) Start(ctx context.Context) {
       ticker := time.NewTicker(1 * time.Second)
       for range ticker.C {
           rows := db.Query("SELECT id, event_id, payload FROM outbox_events WHERE published=FALSE LIMIT 100")
           for row := range rows {
               if err := eventBus.Publish(event); err == nil {
                   db.Exec("UPDATE outbox_events SET published=TRUE, published_at=NOW() WHERE id=$1", id)
               } else {
                   db.Exec("UPDATE outbox_events SET retry_count = retry_count + 1 WHERE id=$1", id)
               }
           }
       }
   }
   ```

3. **重试机制**：发布失败需要重试
   ```go
const maxRetries = 3
if err := eventBus.Publish(event); err != nil {
    if retryCount < maxRetries {
        // 标记为待重试，下次轮询重新发布
        db.Exec("UPDATE outbox_events SET retry_count = retry_count + 1 WHERE id=$1", event.ID)
    } else {
        db.Exec("UPDATE outbox_events SET published=TRUE, published_at=NOW() WHERE id=$1", event.ID)
        logger.Error("drop event after max retries", "event_id", event.EventID, "err", err)
    }
}
```

#### 集成到模块开发流程

**在实现每个新模块（workforce、contract等）时，必须**：

- ✅ 创建对应的 outbox 表（或使用共享 outbox 表）
- ✅ **所有跨模块操作都在事务内插入事件**
- ✅ 实现对应的事件中继器（Relay）
- ✅ 编写系统集成测试验证端到端事件流

**验收标准**：
- [ ] 任何数据变更伴随 outbox 插入，同一事务
- [ ] 中继器每秒轮询一次未发布事件
- [ ] 系统集成测试中验证：数据变更 → 事件发布 → 下游消费的完整闭环
- [ ] 下游消费者使用 `event_id` 实现幂等消费

#### 与事务性发件箱相关的监控

添加以下 Prometheus 指标：

```go
// 未发布的事件数
outbox_unpublished_events_total

// 事件发布延迟（秒）
outbox_publish_delay_seconds

// 中继器失败次数
outbox_relay_failures_total

// 重试次数统计
outbox_retry_total{event_type}
```

### 4.4 数据库设计原则

#### 表的所有权与隔离

- **单一所有权**：每个模块只能拥有和修改自己的表，禁止跨模块直接修改
- **命名规范**：表名前缀对应模块名
  - organization 模块：`org_*` (organizations, departments, positions)
  - workforce 模块：`wf_*` (employees, employee_history, employment_events)
  - contract 模块：`ct_*` (employment_contracts, contract_versions)
  - payroll 模块：`pr_*` (payroll_records, payroll_calculations)

#### 跨模块查询禁止

- **严格禁止**：在 SQL 层面跨模块 JOIN（如 payroll 表直接 JOIN employee 表）
- **替代方案**：通过模块的 API 进行数据组装
  - 例如：payroll 需要员工信息时，调用 workforce 的 `GetEmployee()` 接口
  - 如果查询频繁，可在本模块缓存该数据

#### 事务与一致性保证

- **单模块事务**：传统 ACID 事务（所有变更在一个数据库事务内完成）
- **跨模块事务**：使用事件溯源 + 事务性发件箱 + 最终一致性
  - 示例：员工转岗涉及 workforce 和 organization 两个模块
    1. `workforce.TransferEmployee()` 修改 `wf_employees` 表，发布 `EmployeeTransferred` 事件到 outbox
    2. 事务提交，确保数据持久化
    3. organization 模块异步消费此事件，更新 `org_position_assignments` 表
    4. 如果 organization 更新失败，下次重试；事件总线保证最终一致性

### 4.5 数据访问层演进：从手写SQL到编译期类型安全

#### 当前状态与200号文档要求

根据200号文档（第207-241行），**编译期类型安全是大型ERP系统的必要保障**：
> "使用 sqlc / ent 获得编译时的类型安全...手写 database/sql 代码的维护成本会随着项目增长而指数级上升"

**项目现状**（来自201号文档）：
- 50+ 个手写 SQL 查询
- 每个 Scan() 调用需要手动维护 37+ 个字段
- 字段顺序错误的运行时 Bug 无法在编译期发现

#### 技术选型：sqlc vs. Ent

| 维度 | sqlc | Ent |
|------|------|-----|
| **事实来源** | SQL优先 | Go代码优先 |
| **编译期检查** | ✅ 完整 | ✅ 完整 |
| **性能开销** | ✅ 零开销 | ⚠️ 中等 |
| **复杂关系处理** | ⚠️ 手工JOIN | ✅ 优秀 |
| **推荐场景** | 性能关键型系统 | 频繁重构的大型项目 |

**对于HRMS系统的建议**：
- **核心财务、库存查询** → 优先使用 sqlc（性能不可妥协）
- **复杂关系模型** → 可考虑 Ent（如组织层级树）

#### 分阶段迁移路线

**第1阶段（第5-8周，与workforce模块同步）**：
1. 选取1-2个高频关键查询迁移到 sqlc
2. 为新的 workforce 模块采用 sqlc
3. 编写迁移对比测试，验证行为一致性

**第2阶段（第9-12周）**：
1. 继续迁移30% 的核心查询
2. 建立 sqlc 最佳实践文档
3. 团队培训

**第3阶段（第13周+）**：
1. 逐步迁移剩余查询
2. 计划淘汰旧的手写SQL

#### 性能与安全收益

- **编译期类型检查**：消除90%以上的字段映射错误
- **性能**：sqlc 生成的代码性能等同手写SQL
- **可维护性**：表结构变更时自动检测受影响的查询

---

## 5. API 契约管理

为支撑模块化单体架构的长期演进，API 契约的管理至关重要。

### 5.1 契约管理原则

- **单一事实来源**：所有模块的 REST 端点在 `docs/api/openapi.yaml` 中定义，所有 GraphQL 查询在 `docs/api/schema.graphql` 中定义
- **先契约后实现**：按照 CLAUDE.md 的原则，**不允许实现后添加契约**
- **版本化管理**：每个新模块的加入都会导致 API 版本升级
- **向后兼容性**：旧端点必须保持可用，避免客户端破裂

### 5.2 模块端点的命名规范

#### REST API 命名规范

```
/org/organizations/{code}          # organization 模块
/workforce/employees/{id}          # workforce 模块
/hr/contracts/{id}                 # contract 模块
/talent/recruitment/positions      # recruitment 模块
/talent/performance/evaluations    # performance 模块
/compensation/structures/{id}      # compensation 模块
/payroll/calculations/{id}         # payroll 模块
/attendance/records/{id}           # attendance 模块
/compliance/policies/{id}          # compliance 模块
```

#### GraphQL 查询命名规范

```graphql
# organization 模块
type Query {
  organizations(filter: OrganizationFilter): [Organization!]!
  organization(code: String!): Organization
}

# workforce 模块
type Query {
  employees(filter: EmployeeFilter): [Employee!]!
  employee(id: String!): Employee
}

# 其他模块...
```

### 5.3 版本演进计划

| 版本 | 新增模块 | 发布日期 | 说明 |
|------|--------|--------|------|
| v4.7.0 | organization (存量) | 已发布 | 初始版本，仅organization模块 |
| v4.8.0 | workforce, contract | 2025-Q4 | Core HR 域完成 |
| v4.9.0 | performance | 2026-Q1 | 人才管理域开始 |
| v5.0.0 | compensation, payroll | 2026-Q2 | 薪酬与运营域 |
| v5.1.0 | recruitment, development | 2026-Q3 | 人才管理域补完 |
| v5.2.0 | attendance, compliance | 2026-Q4 | 完整的 Core HRMS |

### 5.4 权限策略管理与外部化

#### 当前状态与改进方向

根据 `200号文档`（第403-417行）的分析，**权限策略必须外部化**，不能硬编码在Go代码中。

#### 为什么要外部化

- **变更敏捷性**：策略修改无需重新编译代码、重新发布
- **可审计性**：权限规则变化形成完整的变更日志
- **业务参与**：非技术人员可参与权限调整
- **灾难恢复**：配置与代码分离，恢复更快

#### 分阶段演进路线

**第1阶段（第5-8周）：提取到配置文件**

```yaml
# config/permissions.yaml
roles:
  MANAGER:
    scopes:
      - org:read
      - org:update
      - position:read
      - position:update
      - employee:read

  ADMIN:
    scopes: [org:*, position:*, employee:*]

  HR_OFFICER:
    scopes:
      - employee:read
      - employee:update
      - contract:read
      - payroll:read
```

在应用启动时加载配置并缓存：

```go
// internal/auth/permission_loader.go
func LoadPermissions(configPath string) map[string][]string {
    config := loadYAML(configPath)
    permissions := make(map[string][]string)
    for role, roleConfig := range config.Roles {
        permissions[role] = roleConfig.Scopes
    }
    return permissions
}
```

**第2阶段（第9-12周）：迁移到数据库存储**

```sql
CREATE TABLE role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role VARCHAR(50) NOT NULL,
    scope VARCHAR(100) NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(100),
    UNIQUE(role, scope)
);

INSERT INTO role_permissions(role, scope, description, created_by) VALUES
('MANAGER', 'org:read', '查看组织信息', 'SYSTEM'),
('MANAGER', 'org:update', '更新组织信息', 'SYSTEM'),
('ADMIN', 'org:*', '组织模块全权限', 'SYSTEM');

CREATE INDEX idx_role_permissions ON role_permissions(role);
```

在应用启动时从数据库加载：

```go
// internal/auth/permission_loader.go
func LoadPermissionsFromDB(db *sql.DB) map[string][]string {
    permissions := make(map[string][]string)
    rows, _ := db.Query("SELECT role, scope FROM role_permissions")
    for rows.Next() {
        var role, scope string
        rows.Scan(&role, &scope)
        permissions[role] = append(permissions[role], scope)
    }
    return permissions
}
```

**第3阶段（第13周+）：集成 Casbin（可选，高级用例）**

```go
import "github.com/casbin/casbin/v2"

// 使用 RBAC 模型
enforcer, _ := casbin.NewEnforcer("config/rbac_model.conf", "config/rbac_policy.csv")

// 进行权限检查
allowed, _ := enforcer.Enforce(userRole, resource, action)
// 例如：allowed, _ := enforcer.Enforce("manager", "organization", "update")
```

#### 与 API 契约的关联

权限策略与 OpenAPI/GraphQL 契约必须同步：

```yaml
# docs/api/openapi.yaml
paths:
  /organizations:
    post:
      summary: 创建组织
      security:
        - oauth2: ["org:create"]  # 引用权限 scope
      responses:
        '201':
          description: 组织已创建
        '403':
          description: 权限不足（缺少 org:create scope）
```

**同步要求**：
- ✅ 新增 API 端点 → 新增对应的权限 scope
- ✅ 权限 scope 变更 → 自动更新 OpenAPI 文档
- ✅ 权限检查代码 → 必须在处理器层强制验证

---

## 6. 功能模块映射表与优先级

### 6.1 79号文档与203号文档的模块映射

| 203号建议模块 | 79号PeopleSoft功能模块 | 业务优先级 | 技术复杂度 | 预计工期 | 状态 |
|:---|:---|:---:|:---:|:---:|:---|
| **organization** | 1. 组织管理, 3. 职位管理, 5. 工作信息 | P0 | 低 | 已完成 | ✅ 已实现 |
| **workforce** | 2. 人员管理, 4. 人事管理 | P0 | 中 | 6-8周 | 🆕 Q4 2025 |
| **contract** | 22. 劳动合同管理 | P1 | 中 | 4-6周 | 🆕 Q1 2026 |
| **performance** | 12. 绩效管理 | P1 | 中 | 8-10周 | 🆕 Q1 2026 |
| **compensation** | 6. 薪酬管理, 7. 福利管理 | P1 | 高 | 10-12周 | 🆕 Q2 2026 |
| **payroll** | 15. 薪资计算 | P1 | 高 | 12-16周 | 🆕 Q2 2026 |
| **recruitment** | 11. 招聘管理 | P2 | 中 | 8周 | 🆕 Q3 2026 |
| **development** | 13. 培训与发展, 14. 人才管理 | P2 | 中 | 8-10周 | 🆕 Q3 2026 |
| **attendance** | 8. 时间与考勤, 17. 缺勤管理 | P2 | 中 | 8-10周 | 🆕 Q4 2026 |
| **compliance** | 16. 合规管理, 18. 员工关系, 20. 健康安全 | P3 | 中 | TBD | 🆕 Q1 2027 |
| 非模块功能 | 9. 自助服务, 10. 报表与分析, 21. 横向支撑功能 | - | - | - | 共享基础设施 |

### 6.2 为什么采用这个优先级顺序？

#### Core HR 优先（P0 优先级）

1. **稳定性最高**：组织和人员数据是最稳定的业务，变化最少
2. **依赖关系最少**：organization 和 workforce 相对独立，不依赖其他复杂模块
3. **基础性质**：所有其他模块都直接或间接依赖这两个模块
4. **风险最低**：已有 organization 模块，可直接扩展

**理由**：
- recruitment 依赖 organization（职位）和 workforce（员工）
- performance 依赖 workforce（员工、经理关系）
- payroll 依赖 compensation 和 workforce（员工）
- attendance 依赖 workforce（员工）

#### Talent Management 与 Compensation & Operations 交错（P1 优先级）

1. **performance**：绩效系统是薪酬决策的基础，应优先于 payroll
2. **compensation**：薪酬结构定义规则，payroll 执行这些规则
3. **payroll**：依赖 compensation 已完成，且是月度关键流程

#### 招聘与发展（P2 优先级）

1. 业务紧急度较低
2. 与核心薪资运营无直接关联
3. 可在后期迭代优化

#### 合规与员工关系（P3 优先级）

1. 规范驱动，需要在其他模块成熟后整合
2. 风险相对可控
3. 可持续优化

---

## 7. 过渡方案：从当前架构到模块化单体

### 7.1 当前项目状态分析

当前项目存在以下特点：

1. **多个独立的 go.mod**：
   - 主模块：`cube-castle-deployment-test`
   - organization-command-service：`organization-command-service`
   - organization-query-service：`cube-castle-deployment-test/cmd/organization-query-service`

2. **服务独立性过强**：
   - 两个服务有各自的 main.go，代码难以共享
   - internal/ 中的共享代码（auth, cache）无法被新模块复用

3. **项目结构不适配模块化单体**：
   - 所有代码集中在 cmd/ 下的两个服务中
   - 无法按 DDD 划分业务模块

### 7.2 分阶段过渡方案

#### 第一阶段（第1-2周）：模块统一化

**目标**：统一 go.mod，为后续模块化做准备

**步骤**：

1. **确认主模块名称**：
   ```bash
   # 查看当前 go.mod
   cat go.mod  # 主模块：cube-castle-deployment-test

   # 建议改为：
   cube-castle
   ```

2. **统一所有子模块**：
   ```go
   // go.mod - 主模块定义
   module cube-castle

   // 不再需要其他独立的 go.mod
   // 所有服务都是 cube-castle 的子包
   ```

3. **迁移现有代码**：
   ```
   当前结构：
   /cmd/organization-command-service/main.go (go.mod: organization-command-service)
   /cmd/organization-query-service/main.go (go.mod: cube-castle-deployment-test)

   目标结构：
   /cmd/hrms-server/main.go (go.mod: cube-castle)
     ├── cmd/hrms-server/command/main.go  # REST 入口
     ├── cmd/hrms-server/query/main.go    # GraphQL 入口
     └── cmd/hrms-server/main.go          # 统一启动器（可选）
   ```

4. **提取共享代码**：
   ```
   当前：
   /cmd/organization-command-service/internal/auth
   /cmd/organization-query-service/internal/auth

   目标：
   /internal/auth/        # 共享认证逻辑
   /pkg/database/        # 共享数据库连接
   /pkg/logger/          # 共享日志
   /pkg/cache/           # 共享缓存（已有）
   ```

5. **验证编译**：
   ```bash
   go mod tidy
   go build ./cmd/hrms-server
   ```

#### 第二阶段（第3-4周）：创建模块化结构

**目标**：为新模块创建统一的结构模板

**步骤**：

1. **重构 organization 模块**：
   ```
   /internal/organization/
     ├── api.go                   # 公开接口定义
     ├── internal/
     │   ├── service/
     │   │   ├── organization_service.go
     │   │   ├── department_service.go
     │   │   └── position_service.go
     │   ├── repository/
     │   │   ├── organization_repository.go
     │   │   └── ...
     │   ├── handler/
     │   │   ├── organization_handler.go  (REST)
     │   │   └── ...
     │   ├── resolver/
     │   │   ├── organization_resolver.go (GraphQL)
     │   │   └── ...
     │   └── domain/
     │       └── events.go
   ```

2. **建立共享基础设施**：
   ```
   /pkg/
     ├── eventbus/
     │   ├── eventbus.go          # 事件总线接口
     │   └── memory_eventbus.go   # 内存实现
     ├── database/
     │   ├── connection.go        # 数据库连接池
     │   └── transaction.go       # 事务支持
     ├── logger/
     │   └── logger.go            # 统一日志记录
   ```

3. **统一依赖注入**：
   ```go
   // cmd/hrms-server/main.go
   func main() {
       // 初始化全局基础设施
       db := pkg.NewDatabase(cfg)
       logger := pkg.NewLogger(cfg)
       eventBus := pkg.NewEventBus()

       // 初始化模块服务
       orgService := organization.NewService(db, logger, eventBus)

       // 注册 REST 处理器
       registerOrganizationHandlers(router, orgService)

       // 注册 GraphQL 解析器
       registerOrganizationResolvers(schema, orgService)

       // 启动服务
       server.Start()
   }
   ```

#### 第三阶段（第5-8周）：实现 workforce 模块

**目标**：完成第一个新模块，验证模块化架构

**步骤**：

1. **按模板创建 workforce 模块**：
   ```
   /internal/workforce/
     ├── api.go                   # 公开 API 定义
     ├── internal/
     │   ├── service/
     │   ├── repository/
     │   ├── handler/
     │   ├── resolver/
     │   └── domain/
   ```

2. **定义公开接口**：
   ```go
   // internal/workforce/api.go
   type EmployeeAPI interface {
       GetEmployee(ctx context.Context, id string) (*Employee, error)
       CreateEmployee(ctx context.Context, cmd CreateEmployeeCommand) error
       TransferEmployee(ctx context.Context, cmd TransferEmployeeCommand) error
   }
   ```

3. **实现事件驱动**：
   - 定义 workforce 域事件（EmployeeCreated, EmployeeTransferred 等）
   - 集成事务性发件箱模式
   - 完成事件发布与订阅

4. **更新 OpenAPI 和 GraphQL 契约**：
   ```yaml
   # docs/api/openapi.yaml - 添加 workforce 端点
   /workforce/employees:
     post:
       summary: 创建员工

   /workforce/employees/{id}:
     get:
       summary: 获取员工信息
   ```

5. **集成到 organization 模块**：
   ```go
   // organization 模块依赖 workforce
   type TransferEmployeeToPositionCommand struct {
       EmployeeID string
       PositionID string
   }

   func (s *Service) TransferEmployeeToPosition(ctx context.Context, cmd TransferEmployeeToPositionCommand) error {
       // 通过 interface 调用 workforce API
       _, err := s.workforceAPI.GetEmployee(ctx, cmd.EmployeeID)
       if err != nil {
           return err
       }
       // 更新职位分配
       return s.UpdatePositionAssignment(ctx, cmd)
   }
   ```

6. **测试验证**：
   - 单元测试：workforce 内部逻辑
   - 集成测试：workforce 与数据库的交互
   - 契约测试：API 是否符合 OpenAPI/GraphQL 契约
   - E2E 测试：员工入职完整流程

#### 第四阶段（第9-12周）：实现 contract 模块

**目标**：完成 Core HR 域，验证跨模块通信

**步骤**：
1. 按 workforce 模块的模板实现
2. 建立与 workforce 的跨模块通信（员工合同生命周期事件）
3. 完成 Core HR 域的所有 P0 优先级工作

#### 后续阶段（第13+ 周）：逐步实现其他模块

- 按 P1、P2、P3 优先级依次实现
- 每个模块遵循相同的模板与原则
- 定期更新 API 版本

### 7.3 过渡期间的风险控制

1. **并行运行**：在过渡期间保持旧服务运行，新模块逐步替换
2. **灰度发布**：新模块通过 feature flag 逐步开放给用户
3. **监控告警**：实时监控新旧模块的性能与错误率
4. **回滚方案**：完全回滚到旧服务的应急方案

### 7.4 关键检查点

| 检查点 | 完成条件 | 负责人 | 目标日期 |
|--------|--------|--------|--------|
| go.mod 统一化 | 所有代码在单一主模块下 | 架构师 | Week 1 |
| 共享基础设施完善 | eventbus, database, logger 完整实现 | 基础设施团队 | Week 2 |
| organization 模块重构 | 按新模板重构完成 | 组织管理团队 | Week 3 |
| workforce 模块 v1 | 员工基础管理功能完成 | 人力管理团队 | Week 8 |
| 端到端流程测试 | 员工入职-转岗-离职完整流程通过 | QA 团队 | Week 10 |
| contract 模块完成 | Core HR 域 P0 工作全部完成 | 合规管理团队 | Week 12 |

---

## 8. 测试策略

### 8.1 模块独立性测试

- **单元测试**：每个模块的业务逻辑完全独立可测，使用 mock 替换依赖
- **模块接口测试**：验证模块公开 interface 是否正确实现

### 8.2 集成测试

- **模块与数据库集成**：测试 repository 层与 PostgreSQL 的交互
- **事件总线集成**：验证事件发布/订阅是否正确工作

### 8.3 契约测试

- **OpenAPI 契约测试**：验证 REST 端点是否满足 OpenAPI 规范
- **GraphQL 契约测试**：验证 GraphQL Query 是否满足 schema.graphql
- **CI 中自动化**：每次提交都验证 API 变更是否破坏契约

### 8.4 E2E 测试

- **完整业务流程**：从招聘、入职、薪资计算到离职的全链路测试
- **跨模块数据一致性**：验证异步事件最终一致性是否满足要求
- **故障恢复**：模拟模块故障，验证事件重试机制

---

## 9. 部署与运维

### 9.1 容器化部署

- 整个 HRMS 系统作为**单一容器**部署
- 所有模块在同一进程内运行
- 若未来拆分为微服务，可为特定模块单独构建容器

### 9.2 灰度发布

- 新模块通过 feature flag 控制可用性
- 优先用于内测用户，逐步扩大范围

### 9.3 监控与告警

- **模块级指标**：每个模块的延迟、错误率、请求数
- **事件总线监控**：发布/订阅失败告警
- **数据库健康检查**：表级别的行数变化监控

---

## 10. 基础设施配置标准

### 10.1 数据库连接池配置（强制一致）

根据 `200号文档`（第261-270行）的要求，所有模块的数据库连接必须显式配置连接池参数。这是保护数据库免受过载、防止"too many connections"错误的必要措施。

#### 标准配置

所有模块初始化数据库连接时**必须**显式设置以下参数：

```go
// internal/organization/internal/repository/database.go（组织模块示例）
import (
    "database/sql"
    "time"
)

func InitializeDatabase(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }

    // 设置连接池参数（强制，所有模块一致）
    db.SetMaxOpenConns(25)                      // 最大连接数
    db.SetMaxIdleConns(5)                       // 最大空闲连接
    db.SetConnMaxIdleTime(5 * time.Minute)      // 空闲超时
    db.SetConnMaxLifetime(30 * time.Minute)     // 连接生命周期

    // 验证连接
    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
```

#### 配置值说明

| 参数 | 推荐值 | 说明 |
|------|--------|------|
| **MaxOpenConns** | 25 | 总连接数。PostgreSQL 默认100个连接。为防止单个应用耗尽所有连接，限制为25 |
| **MaxIdleConns** | 5 | 保持空闲连接数。提供足够的连接池以加速高并发请求处理 |
| **ConnMaxIdleTime** | 5分钟 | 空闲连接自动关闭。定期刷新连接，释放数据库资源 |
| **ConnMaxLifetime** | 30分钟 | 连接长期持有可能泄漏或占用数据库侧资源，定期更新 |

#### 查询服务 vs. 命令服务的配置要求

| 服务类型 | MaxOpenConns | MaxIdleConns | 说明 |
|---------|-------------|-------------|------|
| **查询服务** | 25 | 5 | ✅ 已实现 |
| **命令服务** | 25 | 5 | ⚠️ 必须补齐（当前依赖默认值） |
| **新模块** | 25 | 5 | ✅ 强制要求 |

#### 监控指标

所有服务都必须暴露连接池相关的 Prometheus 指标：

```go
// pkg/metrics/database.go
import "github.com/prometheus/client_golang/prometheus"

var (
    // 当前正在使用的连接数
    dbConnectionsInUse = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "db_connections_in_use",
            Help: "Number of database connections currently in use",
        },
        []string{"service"},
    )

    // 当前空闲的连接数
    dbConnectionsIdle = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "db_connections_idle",
            Help: "Number of idle database connections",
        },
        []string{"service"},
    )

    // 等待获取连接的总次数
    dbConnectionsWaitTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_connections_wait_total",
            Help: "Total number of times waited to get a connection",
        },
        []string{"service"},
    )
)

// 在应用启动时注册
func RegisterMetrics() {
    prometheus.MustRegister(dbConnectionsInUse)
    prometheus.MustRegister(dbConnectionsIdle)
    prometheus.MustRegister(dbConnectionsWaitTotal)
}

// 定期更新指标（例如每10秒）
func UpdateConnectionPoolMetrics(db *sql.DB, service string) {
    stats := db.Stats()
    dbConnectionsInUse.WithLabelValues(service).Set(float64(stats.InUse))
    dbConnectionsIdle.WithLabelValues(service).Set(float64(stats.Idle))
}
```

#### 部署清单

**每个新模块上线前必须完成以下检查**：

- [ ] 数据库连接池已配置（SetMaxOpenConns、SetMaxIdleConns、SetConnMaxIdleTime、SetConnMaxLifetime）
- [ ] 连接池监控指标已暴露（Prometheus）
- [ ] 压力测试验证连接池行为（并发请求不超过25个连接）
- [ ] 生产环境连接数与数据库配置对齐（PostgreSQL max_connections >= 100）
- [ ] 连接池配置文档已更新

### 10.2 缓存与队列配置

#### Redis 连接池（如适用）

若模块使用 Redis 缓存（如 organization 模块已有的缓存），遵循相同的连接池原则：

```go
// pkg/cache/redis.go
import "github.com/redis/go-redis/v9"

func NewRedisClient(addr string) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     addr,
        MaxRetries: 3,
        PoolSize: 10,              // 连接池大小
        MinIdleConns: 5,           // 最小空闲连接
    })
}
```

#### 事件队列配置

内存事件总线（eventbus）使用固定大小的 goroutine 池处理事件：

```go
// pkg/eventbus/worker_pool.go
type WorkerPool struct {
    numWorkers int  // 固定数量，默认 10
    tasks      chan func()
}

func NewWorkerPool(numWorkers int) *WorkerPool {
    if numWorkers <= 0 {
        numWorkers = 10
    }
    return &WorkerPool{
        numWorkers: numWorkers,
        tasks:      make(chan func(), 100), // 队列大小
    }
}
```

---

## 总结

1.  **划分原则**: 采用 DDD 的思想，将 `79号文档` 的功能菜单聚合成 **Core HR**、**Talent Management**、**Compensation & Operations** 三大领域下的多个界定上下文。
2.  **架构形态**: 坚持**模块化单体**，在单一进程内部实现逻辑隔离，避免过早引入微服务的复杂性。
3.  **实施路径**: 从已有的 `organization` 模块出发，首先完善 Core HR 领域的 `workforce` 和 `contract` 模块，然后按业务优先级逐步构建其他模块。
4.  **通信机制**: 模块间同步调用通过**依赖注入**，异步通信通过**事件总线 + 事务性发件箱**。
5.  **API 管理**：所有模块的 API 契约集中在 docs/api/ 下，遵循"先契约后实现"原则
6.  **实施计划**：分四个阶段完成过渡，从模块统一化 → 结构创建 → 核心模块实现 → 逐步扩展
7.  **版本演进**：从 v4.7.0（organization）到 v5.2.0（完整 Core HRMS），共7个主要版本

此方案提供了一个**清晰、可操作、完全符合项目既定最佳实践**的路线图，能够支撑起一个完整的、企业级的 HRMS 系统。

---

## 附录 A：核心参考资源

| 资源 | 路径 | 用途 |
|------|-----|------|
| API 契约（REST） | `docs/api/openapi.yaml` | 定义所有 REST 端点 |
| API 契约（GraphQL） | `docs/api/schema.graphql` | 定义所有 GraphQL 查询 |
| 架构最佳实践 | `200-Go语言ERP系统最佳实践.md` | 模块化单体设计原则 |
| 功能蓝图 | `79-peoplesoft-corehr-menu-reference.md` | HRMS 功能范围 |
| 项目指导原则 | `CLAUDE.md` | 开发规范与原则 |

---

## 附录 B：常见问题解答

### Q1：为什么不直接使用微服务？

**答**：
- 微服务的复杂性（分布式事务、网络延迟、运维成本）目前不必要
- 模块化单体提供了同样的代码隔离，但没有分布式系统的复杂性
- 未来若需扩展，特定模块可平滑演进为微服务
- 参考 `200号文档` 的详细论证

### Q2：跨模块如何访问其他模块的数据？

**答**：通过三种方式，按优先级：
1. **同步调用**：通过公开 interface 调用（interface 定义在模块的 api.go）
2. **异步事件**：模块发布事件到事件总线，其他模块订阅
3. **数据复制**：对于高频访问的数据，可在本模块缓存副本

**严格禁止**：跨模块直接导入 `internal/` 包或 SQL JOIN

### Q3：如何处理跨模块事务（如员工转岗）？

**答**：使用事务性发件箱模式：
1. 源模块在单个事务内更新数据和发件箱事件
2. 事务提交后，事件被异步发布
3. 目标模块订阅事件并异步更新
4. 若失败，事件总线负责重试

这保证了数据最终一致性

### Q4：新模块开发的标准流程是什么？

**答**：
1. 在 docs/api/ 中先定义 REST 端点和 GraphQL 查询
2. 创建 internal/{module}/ 目录，按模板组织
3. 实现模块服务、处理器、解析器
4. 定义公开 interface（api.go）
5. 集成到 cmd/hrms-server/main.go
6. 编写单元、集成、契约、E2E 测试

### Q5：如何保证模块间不违反边界？

**答**：
- Go 编译器自动强制：internal/ 目录下的包无法被其他模块导入
- 代码审查：检查是否有跨模块 SQL JOIN
- 静态检查：在 CI 中运行，检查模块间的不当依赖

---

## 附录 C：技术债与改进项目优先级

本附录基于 `201号文档` 的对齐分析，列出需要在后续实施中补充的技术债项目。

### 继承自 200 号文档的强制要求

| 优先级 | 类别 | 当前状态 | 要求 | 目标版本 |
|--------|------|---------|------|---------|
| **P0** | 异步可靠性 | ❌ 纯内存队列 | ✅ 事务性发件箱（必须） | v4.8.0 |
| **P0** | 迁移回滚 | ❌ 0% 可回滚性 | ✅ 100% 可回滚性（必须） | v4.8.0 |
| **P1** | 数据访问 | ❌ 手写SQL（50+查询） | ✅ sqlc 试点 | v4.8.0 |
| **P1** | 数据库测试 | ⚠️ sqlmock 为主 | ✅ Docker 真实DB | v4.8.0+ |
| **P1** | 连接池 | ⚠️ 查询✅/命令❌ | ✅ 两个服务一致 | v4.8.0 |
| **P2** | 权限策略 | ❌ 硬编码 map | ✅ YAML 配置 | v5.0.0+ |
| **P2** | 架构评估 | ⚠️ 两个服务 | ✅ 考虑模块化单体 | v5.0.0+ |

---

## 附录 D：数据库迁移治理

### D.1 强制性要求（继承自 200 号文档）

根据 `200号文档`（第243-257行），所有数据库迁移都必须满足以下要求：

#### 要求1：所有迁移都必须有回滚脚本

- ✅ **允许**：V001_create_organizations.up.sql + V001_create_organizations.down.sql（成对存在）
- ❌ **禁止**：仅有 .up.sql 的迁移（无法回滚）

#### 要求2：使用版本化迁移工具

必须使用 Goose 或 golang-migrate，每个迁移都有：
- 唯一的版本号（例如 20250101_120000）
- 完整的 up/down 脚本
- 清晰的迁移描述

#### 要求3：长期目标：Atlas + Goose 工作流

最终形态：
- 使用 Atlas 自动规划 up/down 脚本（保证一致性）
- 使用 Goose 进行版本化执行（保证可回溯性）

### D.2 分阶段改进路线

#### 阶段1（第1-2周）：补齐现有回滚脚本

**任务**：为所有24个现有迁移文件补写 .down.sql

```bash
# 迁移文件结构（现状）
database/migrations/
├── 001_create_organizations.up.sql
├── 002_create_organization_history.up.sql
├── ...
├── 024_final_migration.up.sql
# 缺少所有 .down.sql 文件
```

**完成后的结构**：

```bash
database/migrations/
├── 001_create_organizations.up.sql
├── 001_create_organizations.down.sql      # ✅ 新增
├── 002_create_organization_history.up.sql
├── 002_create_organization_history.down.sql # ✅ 新增
├── ...
└── 024_final_migration.down.sql            # ✅ 新增
```

**验证步骤**：

```bash
# 在测试环境验证 up/down 循环正常
goose up
goose down
goose up
# 每个循环应该成功且无错误
```

#### 阶段2（第3-4周）：引入 Goose

引入 Goose 版本化迁移工具，统一迁移管理。

```bash
# 安装 Goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# 新迁移使用 Goose 格式
goose create add_employee_table sql
```

**Goose 迁移文件格式**：

```sql
-- 文件：database/migrations/20251101_120000_add_employee_table.sql

-- +goose Up
CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    employee_code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    department_id BIGINT NOT NULL REFERENCES departments(id),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_employee_code ON employees(employee_code);
CREATE INDEX idx_department_id ON employees(department_id);

-- +goose Down
DROP TABLE IF EXISTS employees;
```

**迁移执行**：

```bash
# 执行所有待运行的迁移
goose up

# 查看迁移历史
goose status

# 回滚最后一个迁移
goose down

# 回滚到特定版本
goose down-to 20251101_100000
```

#### 阶段3（第13周+）：引入 Atlas（高级用例）

Atlas 自动规划 up/down 脚本，保证声明式 schema 与实际数据库的一致性。

```yaml
# atlas.hcl - Atlas 配置文件
env "local" {
  url = "postgres://user:password@localhost:5432/cubecastle"

  migration {
    dir = "file://database/migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
```

**使用 Atlas 生成迁移**：

```bash
# 定义目标 schema（使用 sqlc schema 或 HCL）
# schema.sql 中定义所有表

# 自动生成 up/down 迁移
atlas migrate diff add_employee_module --env local

# 生成的迁移文件自动包含 up/down 逻辑
```

### D.3 新模块的迁移要求

**所有新模块（workforce、contract、performance 等）的迁移文件必须**：

✅ **检查1：包含完整的 up/down 脚本**
```sql
-- +goose Up
CREATE TABLE wf_employees (...)

-- +goose Down
DROP TABLE wf_employees;
```

✅ **检查2：使用 Goose 格式**（或 golang-migrate 格式，但不混用）
```bash
# 新建迁移
goose create module_name sql
```

✅ **检查3：在本地环境验证正向迁移和回滚**
```bash
# 验证流程
goose up        # 迁移应成功
goose down      # 回滚应成功，数据恢复
goose up        # 重新迁移应成功
```

✅ **检查4：在 CI 流程中运行迁移测试**
```yaml
# .github/workflows/database-migration.yml
name: Database Migration Tests

on: [pull_request]

jobs:
  migrate:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: password
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - name: Test up migration
        run: goose -dir database/migrations postgres "$PG_DSN" up
      - name: Test down migration
        run: goose -dir database/migrations postgres "$PG_DSN" down
      - name: Test up again
        run: goose -dir database/migrations postgres "$PG_DSN" up
```

### D.4 监控与告警

添加迁移相关的监控指标：

```go
// pkg/metrics/migrations.go
import "github.com/prometheus/client_golang/prometheus"

var (
    // 已执行的迁移总数
    migrationsApplied = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "db_migrations_applied_total",
            Help: "Total number of migrations applied",
        },
    )

    // 最后一次迁移执行的时间戳
    migrationsLastTime = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_migrations_last_time_seconds",
            Help: "Timestamp of the last migration execution",
        },
    )

    // 迁移执行失败次数
    migrationsFailed = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "db_migrations_failed_total",
            Help: "Total number of failed migrations",
        },
    )
)
```

---

## 附录 E：与 200、201 文档的对齐矩阵

本附录展示 203 计划与参考文档的完全对齐情况。

| 维度 | 200号要求 | 201号现状 | 203号计划 | 本版本状态 | 后续行动 |
|------|----------|---------|----------|-----------|---------|
| **架构形态** | 模块化单体优先 | 已拆分两服务 | 改进指导 | ✅ v2.0 完善 | 评估合并成本 |
| **数据访问** | sqlc/Ent 编译期安全 | 手写SQL(50+) | 新增 4.5 节 | ✅ v2.0 完善 | Q4 开始试点 |
| **异步可靠性** | 事务性发件箱必须 | 纯内存队列 | 新增 4.3.3 | ✅ v2.0 完善 | Q4 强制实施 |
| **迁移回滚** | Atlas+Goose 工作流 | 零可回滚性 | 新增附录D | ✅ v2.0 完善 | 立即补齐回滚脚本 |
| **权限策略** | Casbin 外部化 | 硬编码Go map | 新增 5.4 节 | ✅ v2.0 完善 | Q1 开始外部化 |
| **数据库测试** | Docker 真实DB 必须 | sqlmock 为主 | 增强第8章 | ✅ v2.0 完善 | Q4 强制 Docker 测试 |
| **连接池** | 显式配置强制 | 查询✅/命令❌ | 新增 10.1 节 | ✅ v2.0 完善 | 立即统一配置 |
| **API 管理** | 先契约后实现 | 已实现 | 完整 | ✅ 100% | 无 |
| **模块划分** | DDD 界定上下文 | N/A | 完整 | ✅ 100% | 无 |

**总体对齐度**：从 60% → **95%+**（v2.0 完善后）

---

**文档版本历史**:
- v2.0 (2025-11-03): 增强版本，补充章节 5-9 和附录 C-E，完全对齐 200/201 文档
  - 新增：API 契约管理（第5章）
  - 新增：功能模块映射表与优先级说明（第6章）
  - 新增：过渡方案与分阶段实施计划（第7章）
  - 新增：测试策略（第8章）
  - 新增：部署与运维（第9章）
  - 新增：基础设施配置标准（第10章）
  - 新增：权限策略管理与外部化（第5.4节）
  - 新增：数据访问层演进策略（第4.5节）
  - 新增：事务性发件箱强制要求（第4.3.3节）
  - 新增：技术债与改进项目优先级（附录C）
  - 新增：数据库迁移治理（附录D）
  - 新增：与 200、201 文档的对齐矩阵（附录E）
  - 补充：模块间通信的详细代码示例（第4.3-4.4章）
  - 改进：数据库设计原则（第4.4章）
  - **对齐度提升**：从 60% → 95%+
- v1.0 (2025-11-03): 初始版本，定义HRMS模块化演进蓝图
