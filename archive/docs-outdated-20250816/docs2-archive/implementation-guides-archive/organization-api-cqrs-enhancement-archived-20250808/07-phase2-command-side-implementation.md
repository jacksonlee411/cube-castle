# Phase 2 命令端实施计划 - CQRS完整架构

**文档类型**: Phase 2 技术实施计划  
**项目代码**: ORG-API-CQRS-2025  
**版本**: v1.0  
**创建日期**: 2025-08-06  
**实施状态**: 🚀 开始实施

---

## 🎯 Phase 2 目标

### 核心使命
完成CQRS架构的命令端实施，建立完整的事件驱动架构，实现双路径API和适配器模式，达到ADR-004要求的100%架构对齐。

### 技术目标
- ✅ **命令端CQRS**: 创建/更新/删除组织的标准化命令处理
- ✅ **事件驱动**: Kafka集成，组织变更事件发布/消费
- ✅ **CDC管道**: 自动化数据同步，替代手动Python脚本
- ✅ **双路径API**: `/organization-units` + `/corehr/organizations`
- ✅ **适配器模式**: OrganizationAdapter统一接口
- ✅ **性能优化**: P95响应时间 < 300ms

---

## 🏗️ Phase 2 架构设计

### 目标架构图
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端客户端     │────│   适配器层      │────│   命令处理器    │
│                │    │ OrganizationAPI │    │ (Command Side)  │
└─────────────────┘    │      +         │    └─────────────────┘
                       │  CoreHR API    │            │
                       └─────────────────┘            │
                              │                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   查询处理器    │    │   PostgreSQL    │
                       │ (Query Side)    │    │   (命令存储)     │
                       └─────────────────┘    └─────────────────┘
                              │                       │
                       ┌─────────────────┐            │
                       │     Neo4j      │            ▼
                       │   (查询存储)    │    ┌─────────────────┐
                       └─────────────────┘    │  Kafka事件总线   │
                              ▲               │ (组织变更事件)   │
                              │               └─────────────────┘
                       ┌─────────────────┐            │
                       │  CDC Connector  │            │
                       │ (数据同步管道)   │◄───────────┘
                       └─────────────────┘
```

### Phase 2 新增组件
1. **命令处理器** - 处理组织创建/更新/删除命令
2. **事件发布器** - 发布组织变更事件到Kafka
3. **CDC连接器** - 自动同步PostgreSQL变更到Neo4j
4. **适配器层** - 双路径API统一接口
5. **事件消费器** - 处理下游系统集成事件

---

## 📋 实施步骤

### Phase 2.1: 命令端CQRS实现 (1-2天)

#### 命令模型设计
```go
// 组织命令接口
type OrganizationCommand interface {
    GetCommandID() uuid.UUID
    GetTenantID() uuid.UUID
    GetCommandType() string
    Validate() error
}

// 创建组织命令
type CreateOrganizationCommand struct {
    CommandID   uuid.UUID `json:"command_id"`
    TenantID    uuid.UUID `json:"tenant_id"`
    Name        string    `json:"name" validate:"required"`
    ParentCode  *string   `json:"parent_code,omitempty"`
    UnitType    string    `json:"unit_type" validate:"required"`
    Description *string   `json:"description,omitempty"`
    RequestedBy uuid.UUID `json:"requested_by"`
}

// 更新组织命令
type UpdateOrganizationCommand struct {
    CommandID   uuid.UUID `json:"command_id"`
    TenantID    uuid.UUID `json:"tenant_id"`
    Code        string    `json:"code" validate:"required"`
    Name        *string   `json:"name,omitempty"`
    Status      *string   `json:"status,omitempty"`
    Description *string   `json:"description,omitempty"`
    RequestedBy uuid.UUID `json:"requested_by"`
}
```

#### 命令处理器实现
```go
// 命令处理器
type OrganizationCommandHandler struct {
    repo        *PostgresOrganizationRepository
    eventBus    *KafkaEventBus
    logger      *log.Logger
    validator   *validator.Validate
}

func (h *OrganizationCommandHandler) HandleCreateOrganization(ctx context.Context, cmd CreateOrganizationCommand) (*CreateOrganizationResult, error) {
    // 1. 命令验证
    if err := h.validator.Struct(cmd); err != nil {
        return nil, fmt.Errorf("命令验证失败: %w", err)
    }
    
    // 2. 业务规则验证
    if err := h.validateBusinessRules(ctx, cmd); err != nil {
        return nil, fmt.Errorf("业务规则验证失败: %w", err)
    }
    
    // 3. 数据库事务
    result, err := h.repo.CreateOrganization(ctx, cmd)
    if err != nil {
        return nil, fmt.Errorf("创建组织失败: %w", err)
    }
    
    // 4. 发布事件
    event := OrganizationCreatedEvent{
        EventID:      uuid.New(),
        AggregateID:  result.Code,
        TenantID:     cmd.TenantID,
        Name:         cmd.Name,
        UnitType:     cmd.UnitType,
        CreatedBy:    cmd.RequestedBy,
        CreatedAt:    time.Now(),
    }
    
    if err := h.eventBus.Publish(ctx, "organization.created", event); err != nil {
        h.logger.Printf("事件发布失败 (非致命): %v", err)
        // 注意：事件发布失败不应该回滚业务操作
    }
    
    return result, nil
}
```

### Phase 2.2: Kafka事件总线集成 (1天)

#### 事件模型定义
```go
// 组织事件基础接口
type OrganizationEvent interface {
    GetEventID() uuid.UUID
    GetAggregateID() string
    GetTenantID() uuid.UUID
    GetEventType() string
    GetEventTime() time.Time
}

// 组织创建事件
type OrganizationCreatedEvent struct {
    EventID     uuid.UUID `json:"event_id"`
    AggregateID string    `json:"aggregate_id"` // 组织代码
    TenantID    uuid.UUID `json:"tenant_id"`
    Name        string    `json:"name"`
    UnitType    string    `json:"unit_type"`
    ParentCode  *string   `json:"parent_code,omitempty"`
    CreatedBy   uuid.UUID `json:"created_by"`
    CreatedAt   time.Time `json:"created_at"`
}

// 组织更新事件
type OrganizationUpdatedEvent struct {
    EventID     uuid.UUID              `json:"event_id"`
    AggregateID string                 `json:"aggregate_id"`
    TenantID    uuid.UUID              `json:"tenant_id"`
    Changes     map[string]interface{} `json:"changes"`
    UpdatedBy   uuid.UUID              `json:"updated_by"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

#### Kafka事件总线实现
```go
// Kafka事件总线
type KafkaEventBus struct {
    producer kafka.Producer
    logger   *log.Logger
}

func NewKafkaEventBus(brokers []string, logger *log.Logger) (*KafkaEventBus, error) {
    producer, err := kafka.NewProducer(&kafka.ConfigMap{
        "bootstrap.servers": strings.Join(brokers, ","),
        "client.id":         "organization-command-service",
        "acks":             "all",
        "retries":          "3",
    })
    
    if err != nil {
        return nil, fmt.Errorf("创建Kafka生产者失败: %w", err)
    }
    
    return &KafkaEventBus{
        producer: producer,
        logger:   logger,
    }, nil
}

func (bus *KafkaEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
    eventData, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("事件序列化失败: %w", err)
    }
    
    message := &kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     &topic,
            Partition: kafka.PartitionAny,
        },
        Value: eventData,
        Headers: []kafka.Header{
            {Key: "event-type", Value: []byte(getEventType(event))},
            {Key: "tenant-id", Value: []byte(getTenantID(event).String())},
        },
    }
    
    deliveryChan := make(chan kafka.Event)
    err = bus.producer.Produce(message, deliveryChan)
    if err != nil {
        return fmt.Errorf("事件发布失败: %w", err)
    }
    
    // 等待发布确认
    e := <-deliveryChan
    m := e.(*kafka.Message)
    if m.TopicPartition.Error != nil {
        return fmt.Errorf("事件发布确认失败: %w", m.TopicPartition.Error)
    }
    
    bus.logger.Printf("事件发布成功: topic=%s, partition=%d, offset=%d", 
                      topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
    return nil
}
```

### Phase 2.3: CDC数据同步管道 (1天)

#### Debezium CDC连接器配置
```json
{
  "name": "organization-cdc-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "localhost",
    "database.port": "5432",
    "database.user": "user",
    "database.password": "password",
    "database.dbname": "cubecastle",
    "database.server.name": "cubecastle-org",
    "table.include.list": "public.organization_units",
    "topic.prefix": "cubecastle-org",
    "plugin.name": "pgoutput",
    "slot.name": "debezium_org",
    "publication.name": "dbz_org_publication",
    "transforms": "route",
    "transforms.route.type": "org.apache.kafka.connect.transforms.RegexRouter",
    "transforms.route.regex": "cubecastle-org.public.organization_units",
    "transforms.route.replacement": "organization.changes"
  }
}
```

#### CDC事件消费器
```go
// CDC事件消费器
type OrganizationCDCConsumer struct {
    consumer kafka.Consumer
    syncService *Neo4jSyncService
    logger   *log.Logger
}

func (c *OrganizationCDCConsumer) StartConsuming(ctx context.Context) error {
    err := c.consumer.SubscribeTopics([]string{"organization.changes"}, nil)
    if err != nil {
        return fmt.Errorf("订阅主题失败: %w", err)
    }
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            msg, err := c.consumer.ReadMessage(time.Second)
            if err != nil {
                if err.(kafka.Error).Code() == kafka.ErrTimedOut {
                    continue
                }
                c.logger.Printf("读取消息失败: %v", err)
                continue
            }
            
            if err := c.processChangeEvent(ctx, msg); err != nil {
                c.logger.Printf("处理变更事件失败: %v", err)
                // 根据错误类型决定是否重试
            }
        }
    }
}

func (c *OrganizationCDCConsumer) processChangeEvent(ctx context.Context, msg *kafka.Message) error {
    var changeEvent DebeziumChangeEvent
    if err := json.Unmarshal(msg.Value, &changeEvent); err != nil {
        return fmt.Errorf("反序列化变更事件失败: %w", err)
    }
    
    switch changeEvent.Payload.Op {
    case "c": // CREATE
        return c.syncService.HandleCreate(ctx, changeEvent.Payload.After)
    case "u": // UPDATE  
        return c.syncService.HandleUpdate(ctx, changeEvent.Payload.After)
    case "d": // DELETE
        return c.syncService.HandleDelete(ctx, changeEvent.Payload.Before)
    default:
        c.logger.Printf("未知的变更操作: %s", changeEvent.Payload.Op)
        return nil
    }
}
```

### Phase 2.4: 双路径API实现 (1天)

#### 适配器模式设计
```go
// 组织适配器接口
type OrganizationAdapter interface {
    // 查询操作
    GetOrganizations(ctx context.Context, req GetOrganizationsRequest) (*OrganizationsResponse, error)
    GetOrganization(ctx context.Context, code string) (*OrganizationResponse, error)
    GetOrganizationStats(ctx context.Context, req StatsRequest) (*StatsResponse, error)
    
    // 命令操作
    CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*CreateOrganizationResponse, error)
    UpdateOrganization(ctx context.Context, code string, req UpdateOrganizationRequest) (*UpdateOrganizationResponse, error)
    DeleteOrganization(ctx context.Context, code string) (*DeleteOrganizationResponse, error)
}

// 标准组织API适配器
type StandardOrganizationAdapter struct {
    queryHandler   *OrganizationQueryHandler
    commandHandler *OrganizationCommandHandler
    logger         *log.Logger
}

// CoreHR组织API适配器
type CoreHROrganizationAdapter struct {
    adapter *StandardOrganizationAdapter
    mapper  *CoreHRMapper
    logger  *log.Logger
}

func (a *CoreHROrganizationAdapter) GetOrganizations(ctx context.Context, req GetOrganizationsRequest) (*OrganizationsResponse, error) {
    // 1. 请求格式转换
    stdReq := a.mapper.ToStandardGetRequest(req)
    
    // 2. 调用标准适配器
    stdResp, err := a.adapter.GetOrganizations(ctx, stdReq)
    if err != nil {
        return nil, err
    }
    
    // 3. 响应格式转换
    return a.mapper.ToCoreHRResponse(stdResp), nil
}
```

#### 双路径路由配置
```go
// 双路径API路由
func setupRoutes(r chi.Router, standardAdapter, coreHRAdapter OrganizationAdapter) {
    // 标准组织API路径
    r.Route("/api/v1/organization-units", func(r chi.Router) {
        // 查询端点
        r.Get("/", wrapAdapter(standardAdapter.GetOrganizations))
        r.Get("/{code}", wrapAdapter(standardAdapter.GetOrganization))
        r.Get("/stats", wrapAdapter(standardAdapter.GetOrganizationStats))
        
        // 命令端点
        r.Post("/", wrapAdapter(standardAdapter.CreateOrganization))
        r.Put("/{code}", wrapAdapter(standardAdapter.UpdateOrganization))
        r.Delete("/{code}", wrapAdapter(standardAdapter.DeleteOrganization))
    })
    
    // CoreHR组织API路径
    r.Route("/api/v1/corehr/organizations", func(r chi.Router) {
        // 查询端点
        r.Get("/", wrapAdapter(coreHRAdapter.GetOrganizations))
        r.Get("/{code}", wrapAdapter(coreHRAdapter.GetOrganization))
        r.Get("/stats", wrapAdapter(coreHRAdapter.GetOrganizationStats))
        
        // 命令端点  
        r.Post("/", wrapAdapter(coreHRAdapter.CreateOrganization))
        r.Put("/{code}", wrapAdapter(coreHRAdapter.UpdateOrganization))
        r.Delete("/{code}", wrapAdapter(coreHRAdapter.DeleteOrganization))
    })
}
```

---

## 🎯 Phase 2 验收标准

### 功能验收
- [ ] **命令处理**: 创建/更新/删除组织成功率 > 99.9%
- [ ] **事件发布**: Kafka事件发布成功率 > 99.9%  
- [ ] **数据同步**: CDC管道同步延迟 < 5秒
- [ ] **双路径API**: 两个API路径功能完全对等
- [ ] **适配器模式**: 统一接口，不同数据格式支持

### 性能验收
```yaml
命令端性能:
  - 创建组织: P95 < 200ms
  - 更新组织: P95 < 250ms  
  - 删除组织: P95 < 150ms

事件处理:
  - 事件发布延迟: P95 < 50ms
  - 事件消费延迟: P95 < 100ms
  - CDC同步延迟: P95 < 5000ms

API响应:
  - 双路径API一致性: 100%
  - 错误处理: 完善
  - 并发处理: 支持100+ QPS
```

### 质量验收
- [ ] **数据一致性**: 最终一致性保证
- [ ] **事务完整性**: 命令失败时正确回滚
- [ ] **事件可靠性**: 事件不重复、不丢失
- [ ] **监控完整**: 关键指标全覆盖
- [ ] **文档完善**: API文档和运维手册

---

## 🚀 实施时间表

### Phase 2.1: 命令端 (Day 1-2)
- Day 1上午: 命令模型和处理器实现
- Day 1下午: PostgreSQL仓储层实现  
- Day 2上午: 业务规则和验证逻辑
- Day 2下午: 单元测试和集成测试

### Phase 2.2: 事件驱动 (Day 3)  
- 上午: Kafka集成和事件总线实现
- 下午: 事件模型和发布机制测试

### Phase 2.3: CDC管道 (Day 4)
- 上午: Debezium连接器配置和部署
- 下午: CDC消费器实现和测试

### Phase 2.4: 双路径API (Day 5)
- 上午: 适配器模式实现
- 下午: 路由配置和E2E测试

### Phase 2.5: 验证优化 (Day 6)
- 上午: 性能测试和调优
- 下午: 文档更新和交付

---

**预计完成时间**: 6个工作日  
**当前状态**: 🚀 Ready to Start  
**下一步**: 开始命令端CQRS实现

---

*Phase 2将建立完整的CQRS事件驱动架构，实现ADR-004的全部技术要求*