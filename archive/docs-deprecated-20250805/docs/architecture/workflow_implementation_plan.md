# 工作流系统实施方案

## 📋 **项目概述**

基于城堡蓝图和元合约v6.0规范，为Cube Castle HR SaaS平台设计并实施事件驱动的工作流系统。

**核心设计原则**：
- 事件驱动架构 (元合约模块7)
- 进程内事务性发件箱 (城堡蓝图3.2节)
- 租户隔离和数据安全 (元合约模块8)
- API优先设计 (城堡蓝图3.1节)
- 切片化开发 (城堡蓝图4.3节)

## 🏗️ **第一阶段：业务流程事件系统**

### **1.1 数据模型设计**

#### **业务流程事件模型**

```go
// go-app/ent/schema/business_process_event.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
    "time"
)

type BusinessProcessEvent struct {
    ent.Schema
}

func (BusinessProcessEvent) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
        field.UUID("tenant_id", uuid.UUID{}).Comment("租户ID - 多租户隔离"),
        field.String("event_type").Comment("事件类型: HR.Employee.Hired, HR.Position.Created等"),
        field.String("entity_type").Comment("实体类型: Employee, Position, OrganizationUnit"),
        field.UUID("entity_id", uuid.UUID{}).Comment("关联的实体ID"),
        field.Time("effective_date").Comment("事件生效日期"),
        field.JSON("event_data", map[string]interface{}{}).Comment("事件负载数据"),
        field.UUID("initiated_by", uuid.UUID{}).Comment("发起人用户ID"),
        field.String("correlation_id").Optional().Comment("关联ID - 用于追踪相关事件"),
        field.Enum("status").Values("PENDING", "PROCESSING", "COMPLETED", "FAILED").Default("PENDING"),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (BusinessProcessEvent) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("tenant_id", "event_type"),
        index.Fields("tenant_id", "entity_type", "entity_id"),
        index.Fields("tenant_id", "effective_date"),
        index.Fields("correlation_id"),
    }
}
```

#### **事务性发件箱模型**

```go
// go-app/ent/schema/outbox_event.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
    "time"
)

type OutboxEvent struct {
    ent.Schema
}

func (OutboxEvent) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
        field.UUID("tenant_id", uuid.UUID{}).Comment("租户ID"),
        field.String("event_type").Comment("事件类型"),
        field.Bytes("payload").Comment("事件负载 - JSON序列化"),
        field.String("destination").Comment("目标系统: neo4j, external_api等"),
        field.Int("retry_count").Default(0).Comment("重试次数"),
        field.Time("next_retry_at").Optional().Comment("下次重试时间"),
        field.Time("processed_at").Optional().Comment("处理完成时间"),
        field.String("error_message").Optional().Comment("错误信息"),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

func (OutboxEvent) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("processed_at"),           // 查询未处理事件
        index.Fields("next_retry_at"),          // 重试队列
        index.Fields("tenant_id", "event_type"), // 租户+事件类型查询
    }
}
```

#### **工作流实例模型**

```go
// go-app/ent/schema/workflow_instance.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
    "time"
)

type WorkflowInstance struct {
    ent.Schema
}

func (WorkflowInstance) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
        field.UUID("tenant_id", uuid.UUID{}).Comment("租户ID"),
        field.String("workflow_type").Comment("工作流类型: EmployeeOnboarding, PositionChange等"),
        field.String("current_state").Comment("当前状态"),
        field.JSON("state_history", []map[string]interface{}{}).Comment("状态历史"),
        field.JSON("context", map[string]interface{}{}).Comment("工作流上下文数据"),
        field.UUID("initiated_by", uuid.UUID{}).Comment("发起人"),
        field.String("correlation_id").Comment("关联ID"),
        field.Time("started_at").Default(time.Now),
        field.Time("completed_at").Optional(),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (WorkflowInstance) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("tenant_id", "workflow_type"),
        index.Fields("tenant_id", "current_state"),
        index.Fields("correlation_id"),
        index.Fields("initiated_by"),
    }
}
```

#### **工作流步骤模型**

```go
// go-app/ent/schema/workflow_step.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
    "time"
)

type WorkflowStep struct {
    ent.Schema
}

func (WorkflowStep) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Unique(),
        field.UUID("tenant_id", uuid.UUID{}).Comment("租户ID"),
        field.UUID("workflow_instance_id", uuid.UUID{}).Comment("工作流实例ID"),
        field.String("step_name").Comment("步骤名称"),
        field.String("step_type").Comment("步骤类型: MANUAL, AUTOMATED, APPROVAL"),
        field.Enum("status").Values("PENDING", "IN_PROGRESS", "COMPLETED", "SKIPPED", "FAILED").Default("PENDING"),
        field.UUID("assigned_to", uuid.UUID{}).Optional().Comment("分配给的用户"),
        field.JSON("input_data", map[string]interface{}{}).Optional().Comment("输入数据"),
        field.JSON("output_data", map[string]interface{}{}).Optional().Comment("输出数据"),
        field.Time("due_date").Optional().Comment("截止日期"),
        field.Time("started_at").Optional(),
        field.Time("completed_at").Optional(),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (WorkflowStep) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("workflow_instance", WorkflowInstance.Type).
            Ref("steps").
            Field("workflow_instance_id").
            Required().
            Unique(),
    }
}

func (WorkflowStep) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("tenant_id", "workflow_instance_id"),
        index.Fields("tenant_id", "status"),
        index.Fields("assigned_to", "status"),
    }
}
```

### **1.2 核心服务层设计**

#### **业务流程事件服务**

```go
// go-app/internal/service/business_process_event_service.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/gaogu/cube-castle/go-app/ent"
    "github.com/gaogu/cube-castle/go-app/ent/businessprocessevent"
    "github.com/gaogu/cube-castle/go-app/ent/outboxevent"
    "github.com/google/uuid"
)

type BusinessProcessEventService struct {
    client *ent.Client
}

type CreateEventRequest struct {
    TenantID       uuid.UUID              `json:"tenant_id"`
    EventType      string                 `json:"event_type"`
    EntityType     string                 `json:"entity_type"`
    EntityID       uuid.UUID              `json:"entity_id"`
    EffectiveDate  time.Time              `json:"effective_date"`
    EventData      map[string]interface{} `json:"event_data"`
    InitiatedBy    uuid.UUID              `json:"initiated_by"`
    CorrelationID  string                 `json:"correlation_id,omitempty"`
}

// PublishEvent 发布业务流程事件 (实现元合约BPE-004规约)
func (s *BusinessProcessEventService) PublishEvent(ctx context.Context, req *CreateEventRequest) (*ent.BusinessProcessEvent, error) {
    // 开始数据库事务 - 确保事件和发件箱记录的原子性
    tx, err := s.client.Tx(ctx)
    if err != nil {
        return nil, fmt.Errorf("开始事务失败: %w", err)
    }
    defer tx.Rollback()

    // 1. 创建业务流程事件
    event, err := tx.BusinessProcessEvent.Create().
        SetTenantID(req.TenantID).
        SetEventType(req.EventType).
        SetEntityType(req.EntityType).
        SetEntityID(req.EntityID).
        SetEffectiveDate(req.EffectiveDate).
        SetEventData(req.EventData).
        SetInitiatedBy(req.InitiatedBy).
        SetNillableCorrelationID(&req.CorrelationID).
        Save(ctx)
    if err != nil {
        return nil, fmt.Errorf("创建业务流程事件失败: %w", err)
    }

    // 2. 创建发件箱事件用于异步处理
    payload, err := json.Marshal(event)
    if err != nil {
        return nil, fmt.Errorf("序列化事件负载失败: %w", err)
    }

    _, err = tx.OutboxEvent.Create().
        SetTenantID(req.TenantID).
        SetEventType(req.EventType).
        SetPayload(payload).
        SetDestination("neo4j"). // 目标为图数据库同步
        Save(ctx)
    if err != nil {
        return nil, fmt.Errorf("创建发件箱事件失败: %w", err)
    }

    // 3. 提交事务
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("提交事务失败: %w", err)
    }

    return event, nil
}

// GetEventsByEntity 获取实体相关的所有事件
func (s *BusinessProcessEventService) GetEventsByEntity(ctx context.Context, tenantID, entityID uuid.UUID) ([]*ent.BusinessProcessEvent, error) {
    return s.client.BusinessProcessEvent.Query().
        Where(
            businessprocessevent.TenantID(tenantID),
            businessprocessevent.EntityID(entityID),
        ).
        Order(ent.Desc(businessprocessevent.FieldEffectiveDate)).
        All(ctx)
}
```

#### **事务性发件箱处理器**

```go
// go-app/internal/service/outbox_processor.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/gaogu/cube-castle/go-app/ent"
    "github.com/gaogu/cube-castle/go-app/ent/outboxevent"
    "github.com/gaogu/cube-castle/go-app/internal/logging"
)

type OutboxProcessor struct {
    client        *ent.Client
    neo4jService  *Neo4jService
    logger        *logging.StructuredLogger
    batchSize     int
    pollInterval  time.Duration
}

func NewOutboxProcessor(client *ent.Client, neo4jService *Neo4jService, logger *logging.StructuredLogger) *OutboxProcessor {
    return &OutboxProcessor{
        client:       client,
        neo4jService: neo4jService,
        logger:       logger,
        batchSize:    10,
        pollInterval: 5 * time.Second,
    }
}

// Start 启动发件箱处理器 (符合城堡蓝图3.2节进程内工作线程设计)
func (p *OutboxProcessor) Start(ctx context.Context) {
    ticker := time.NewTicker(p.pollInterval)
    defer ticker.Stop()

    p.logger.Info("OutboxProcessor started", map[string]interface{}{
        "batch_size":    p.batchSize,
        "poll_interval": p.pollInterval,
    })

    for {
        select {
        case <-ctx.Done():
            p.logger.Info("OutboxProcessor停止")
            return
        case <-ticker.C:
            if err := p.processUnprocessedEvents(ctx); err != nil {
                p.logger.Error("处理发件箱事件失败", err, nil)
            }
        }
    }
}

func (p *OutboxProcessor) processUnprocessedEvents(ctx context.Context) error {
    // 查询未处理的事件
    events, err := p.client.OutboxEvent.Query().
        Where(outboxevent.ProcessedAtIsNil()).
        Order(ent.Asc(outboxevent.FieldCreatedAt)).
        Limit(p.batchSize).
        All(ctx)
    if err != nil {
        return fmt.Errorf("查询未处理事件失败: %w", err)
    }

    if len(events) == 0 {
        return nil
    }

    p.logger.Info("处理发件箱事件", map[string]interface{}{
        "event_count": len(events),
    })

    for _, event := range events {
        if err := p.processEvent(ctx, event); err != nil {
            p.logger.Error("处理单个事件失败", err, map[string]interface{}{
                "event_id":   event.ID,
                "event_type": event.EventType,
            })
        }
    }

    return nil
}

func (p *OutboxProcessor) processEvent(ctx context.Context, event *ent.OutboxEvent) error {
    switch event.Destination {
    case "neo4j":
        return p.processNeo4jEvent(ctx, event)
    default:
        return fmt.Errorf("未知的目标系统: %s", event.Destination)
    }
}

func (p *OutboxProcessor) processNeo4jEvent(ctx context.Context, event *ent.OutboxEvent) error {
    // 反序列化事件负载
    var businessEvent map[string]interface{}
    if err := json.Unmarshal(event.Payload, &businessEvent); err != nil {
        return p.markEventFailed(ctx, event, fmt.Errorf("反序列化事件失败: %w", err))
    }

    // 根据事件类型执行相应的Neo4j操作
    if err := p.syncToNeo4j(ctx, businessEvent); err != nil {
        return p.markEventFailed(ctx, event, err)
    }

    // 标记事件为已处理
    return p.markEventProcessed(ctx, event)
}

func (p *OutboxProcessor) syncToNeo4j(ctx context.Context, event map[string]interface{}) error {
    eventType, ok := event["event_type"].(string)
    if !ok {
        return fmt.Errorf("事件类型缺失或无效")
    }

    switch eventType {
    case "HR.Employee.Hired":
        return p.neo4jService.CreateEmployeeNode(ctx, event)
    case "HR.Position.Created":
        return p.neo4jService.CreatePositionNode(ctx, event)
    case "HR.OrganizationUnit.Created":
        return p.neo4jService.CreateOrgUnitNode(ctx, event)
    default:
        p.logger.Warn("未知的事件类型", map[string]interface{}{
            "event_type": eventType,
        })
        return nil // 不处理未知事件类型，但不标记为失败
    }
}

func (p *OutboxProcessor) markEventProcessed(ctx context.Context, event *ent.OutboxEvent) error {
    now := time.Now()
    _, err := p.client.OutboxEvent.UpdateOneID(event.ID).
        SetProcessedAt(now).
        Save(ctx)
    return err
}

func (p *OutboxProcessor) markEventFailed(ctx context.Context, event *ent.OutboxEvent, processingErr error) error {
    retryCount := event.RetryCount + 1
    nextRetryAt := time.Now().Add(time.Duration(retryCount*retryCount) * time.Minute) // 指数退避

    _, err := p.client.OutboxEvent.UpdateOneID(event.ID).
        SetRetryCount(retryCount).
        SetNextRetryAt(nextRetryAt).
        SetErrorMessage(processingErr.Error()).
        Save(ctx)
    return err
}
```

#### **工作流引擎服务**

```go
// go-app/internal/service/workflow_engine.go
package service

import (
    "context"
    "fmt"
    "time"

    "github.com/gaogu/cube-castle/go-app/ent"
    "github.com/gaogu/cube-castle/go-app/ent/workflowinstance"
    "github.com/gaogu/cube-castle/go-app/internal/logging"
    "github.com/google/uuid"
)

type WorkflowEngine struct {
    client               *ent.Client
    eventService         *BusinessProcessEventService
    logger               *logging.StructuredLogger
    workflowDefinitions  map[string]*WorkflowDefinition
}

type WorkflowDefinition struct {
    Name        string                    `json:"name"`
    States      []StateDefinition         `json:"states"`
    Transitions map[string][]Transition   `json:"transitions"`
}

type StateDefinition struct {
    Name        string                 `json:"name"`
    Type        string                 `json:"type"` // MANUAL, AUTOMATED, APPROVAL
    Handler     string                 `json:"handler,omitempty"`
    Timeout     time.Duration          `json:"timeout,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type Transition struct {
    ToState   string                 `json:"to_state"`
    Condition string                 `json:"condition,omitempty"`
    Action    string                 `json:"action,omitempty"`
}

func NewWorkflowEngine(client *ent.Client, eventService *BusinessProcessEventService, logger *logging.StructuredLogger) *WorkflowEngine {
    engine := &WorkflowEngine{
        client:              client,
        eventService:        eventService,
        logger:              logger,
        workflowDefinitions: make(map[string]*WorkflowDefinition),
    }
    
    // 注册默认工作流定义
    engine.registerDefaultWorkflows()
    return engine
}

// StartWorkflow 启动新工作流实例
func (w *WorkflowEngine) StartWorkflow(ctx context.Context, req *StartWorkflowRequest) (*ent.WorkflowInstance, error) {
    definition, exists := w.workflowDefinitions[req.WorkflowType]
    if !exists {
        return nil, fmt.Errorf("未知的工作流类型: %s", req.WorkflowType)
    }

    // 获取初始状态
    if len(definition.States) == 0 {
        return nil, fmt.Errorf("工作流定义缺少状态: %s", req.WorkflowType)
    }
    initialState := definition.States[0].Name

    // 创建工作流实例
    instance, err := w.client.WorkflowInstance.Create().
        SetTenantID(req.TenantID).
        SetWorkflowType(req.WorkflowType).
        SetCurrentState(initialState).
        SetContext(req.Context).
        SetInitiatedBy(req.InitiatedBy).
        SetCorrelationID(req.CorrelationID).
        Save(ctx)
    if err != nil {
        return nil, fmt.Errorf("创建工作流实例失败: %w", err)
    }

    // 创建初始步骤
    if err := w.createWorkflowSteps(ctx, instance, definition); err != nil {
        return nil, fmt.Errorf("创建工作流步骤失败: %w", err)
    }

    // 发布工作流启动事件
    _, err = w.eventService.PublishEvent(ctx, &CreateEventRequest{
        TenantID:      req.TenantID,
        EventType:     fmt.Sprintf("Workflow.%s.Started", req.WorkflowType),
        EntityType:    "WorkflowInstance",
        EntityID:      instance.ID,
        EffectiveDate: time.Now(),
        EventData: map[string]interface{}{
            "workflow_type":  req.WorkflowType,
            "initial_state":  initialState,
            "context":        req.Context,
        },
        InitiatedBy:   req.InitiatedBy,
        CorrelationID: req.CorrelationID,
    })
    if err != nil {
        w.logger.Error("发布工作流启动事件失败", err, map[string]interface{}{
            "workflow_id": instance.ID,
        })
    }

    return instance, nil
}

type StartWorkflowRequest struct {
    TenantID      uuid.UUID              `json:"tenant_id"`
    WorkflowType  string                 `json:"workflow_type"`
    Context       map[string]interface{} `json:"context"`
    InitiatedBy   uuid.UUID              `json:"initiated_by"`
    CorrelationID string                 `json:"correlation_id"`
}

// registerDefaultWorkflows 注册默认工作流定义
func (w *WorkflowEngine) registerDefaultWorkflows() {
    // 员工入职工作流
    w.workflowDefinitions["EmployeeOnboarding"] = &WorkflowDefinition{
        Name: "EmployeeOnboarding",
        States: []StateDefinition{
            {Name: "INITIATED", Type: "AUTOMATED", Handler: "initiate_onboarding"},
            {Name: "BACKGROUND_CHECK", Type: "MANUAL", Timeout: 72 * time.Hour},
            {Name: "DOCUMENTATION", Type: "MANUAL", Timeout: 48 * time.Hour},
            {Name: "SYSTEM_SETUP", Type: "AUTOMATED", Handler: "setup_systems"},
            {Name: "COMPLETED", Type: "AUTOMATED", Handler: "complete_onboarding"},
        },
        Transitions: map[string][]Transition{
            "INITIATED": {{ToState: "BACKGROUND_CHECK"}},
            "BACKGROUND_CHECK": {{ToState: "DOCUMENTATION", Condition: "background_check_passed"}},
            "DOCUMENTATION": {{ToState: "SYSTEM_SETUP", Condition: "documentation_complete"}},
            "SYSTEM_SETUP": {{ToState: "COMPLETED"}},
        },
    }

    // 岗位变更工作流
    w.workflowDefinitions["PositionChange"] = &WorkflowDefinition{
        Name: "PositionChange",
        States: []StateDefinition{
            {Name: "REQUESTED", Type: "AUTOMATED", Handler: "process_request"},
            {Name: "MANAGER_APPROVAL", Type: "APPROVAL", Timeout: 48 * time.Hour},
            {Name: "HR_REVIEW", Type: "APPROVAL", Timeout: 24 * time.Hour},
            {Name: "EFFECTIVE", Type: "AUTOMATED", Handler: "apply_position_change"},
            {Name: "COMPLETED", Type: "AUTOMATED", Handler: "complete_position_change"},
        },
        Transitions: map[string][]Transition{
            "REQUESTED": {{ToState: "MANAGER_APPROVAL"}},
            "MANAGER_APPROVAL": {{ToState: "HR_REVIEW", Condition: "manager_approved"}},
            "HR_REVIEW": {{ToState: "EFFECTIVE", Condition: "hr_approved"}},
            "EFFECTIVE": {{ToState: "COMPLETED"}},
        },
    }
}

func (w *WorkflowEngine) createWorkflowSteps(ctx context.Context, instance *ent.WorkflowInstance, definition *WorkflowDefinition) error {
    // 创建工作流步骤的实现
    // 这里需要根据工作流定义创建相应的步骤
    return nil
}
```

## 🎯 **实施路线图**

### **第一阶段 - 核心基础设施 (第1-2周)**

#### **里程碑1.1：数据模型建立**
- [ ] 创建Ent Schema文件
- [ ] 生成数据库迁移
- [ ] 运行数据库迁移
- [ ] 验证表结构

**命令序列**：
```bash
# 1. 创建新的schema文件
# 已在上面提供完整代码

# 2. 生成Ent代码
cd go-app
go generate ./ent

# 3. 创建和运行迁移
go run cmd/migrate/main.go
```

#### **里程碑1.2：核心服务实现**
- [ ] 实现BusinessProcessEventService
- [ ] 实现OutboxProcessor
- [ ] 实现WorkflowEngine基础框架
- [ ] 编写单元测试

#### **里程碑1.3：集成测试**
- [ ] 端到端事件流测试
- [ ] 事务性发件箱测试
- [ ] 租户隔离验证

### **第二阶段 - 工作流引擎完善 (第3-4周)**

#### **里程碑2.1：工作流定义系统**
- [ ] 工作流定义存储和管理
- [ ] 状态机验证逻辑
- [ ] 工作流步骤自动创建

#### **里程碑2.2：第一个完整工作流**
- [ ] 员工入职工作流完整实现
- [ ] 工作流状态转换逻辑
- [ ] 超时和错误处理

#### **里程碑2.3：API层开发**
- [ ] 工作流管理API
- [ ] 工作流实例查询API
- [ ] 事件查询API

### **第三阶段 - Neo4j集成和可观测性 (第5-6周)**

#### **里程碑3.1：图数据库集成**
- [ ] Neo4j服务增强
- [ ] 复杂关系建模
- [ ] 图查询接口

#### **里程碑3.2：监控和可观测性**
- [ ] 工作流性能指标
- [ ] 事件处理监控
- [ ] 错误和告警机制

#### **里程碑3.3：压力测试和优化**
- [ ] 工作流性能测试
- [ ] 数据库查询优化
- [ ] 并发处理优化

## 📊 **验证标准**

### **功能验证**
- ✅ 所有业务事件必须通过事务性发件箱
- ✅ 工作流状态转换符合定义
- ✅ 租户数据完全隔离
- ✅ 事件处理具有幂等性

### **性能标准**
- 🎯 事件发布延迟 < 100ms
- 🎯 发件箱处理延迟 < 5s
- 🎯 工作流状态转换 < 200ms
- 🎯 支持并发处理 > 100 TPS

### **质量标准**
- 📋 单元测试覆盖率 > 80%
- 📋 集成测试覆盖所有关键路径
- 📋 错误处理和重试机制完备
- 📋 日志记录结构化和可搜索

## 🔧 **开发工具和依赖**

### **必需依赖**
```go
// go.mod 新增依赖
github.com/google/uuid v1.3.0
entgo.io/ent v0.12.0
github.com/neo4j/neo4j-go-driver/v5 v5.0.0
```

### **开发工具**
- Ent CLI：`go install entgo.io/ent/cmd/ent@latest`
- 数据库迁移工具
- Neo4j Desktop (开发环境)

### **测试工具**
- testcontainers-go (集成测试)
- httptest (API测试)
- 性能测试框架

## 📚 **相关文档**

- [城堡蓝图](/docs/architecture/castle_blueprint.md)
- [元合约v6.0规范](/docs/architecture/metacontract_v6.0_specification.md)
- [现有API文档](/docs/api/)
- [数据库设计文档](/docs/architecture/database_design.md)

## 🚀 **下一步行动**

1. **立即开始**：创建数据模型Schema文件
2. **并行开发**：实现核心服务类
3. **持续集成**：建立自动化测试流水线
4. **文档更新**：保持技术文档同步更新

---

**最后更新**：2025-07-29  
**版本**：v1.0  
**负责人**：架构师团队  
**审核状态**：待审核