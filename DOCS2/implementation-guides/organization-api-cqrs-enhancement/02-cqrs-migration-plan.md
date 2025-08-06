# 组织架构API CQRS迁移详细计划

**项目**: 组织架构API CQRS架构改造专项  
**文档**: 02-CQRS迁移详细计划  
**版本**: v1.0  
**制定日期**: 2025-08-06  
**参考依据**: ADR-004 + CQRS统一实施指南  
**实施状态**: 📋 计划中

---

## 📋 迁移概述

### 迁移使命
严格按照[CQRS统一架构实施指南](../../architecture-foundations/cqrs-unified-implementation-guide.md)的**三阶段城堡化迁移标准**，将组织架构模块从传统REST架构完全迁移到符合城堡架构的CQRS实现。

### 核心原则遵循
基于CQRS架构宪章的**五大宪法原则**：
1. **职责分离原则**：命令端专注写操作与业务逻辑，查询端专注读操作与性能优化
2. **存储分离原则**：PostgreSQL作为命令端事务存储，Neo4j作为查询端图数据库
3. **事件驱动原则**：通过领域事件实现命令端到查询端的数据同步
4. **最终一致性原则**：保证数据最终一致，容忍短暂的数据不一致
5. **独立演进原则**：读写端可独立优化、扩展和演进

---

## 🏗️ 三阶段城堡化迁移计划

### 阶段1: 查询端城堡化 (1-2周) 🟢 低风险

**目标**: 启用CQRS查询功能，保持写操作不变

#### 1.1 Neo4j查询端建立 (Week 1)

**基础设施准备**：
```bash
# Neo4j环境配置
- Docker Neo4j 5.x 集群部署
- 数据库连接池配置 (100连接)
- 图索引策略设计
- 多租户数据隔离设置
```

**数据同步建立**：
```yaml
CDC Pipeline配置:
  - PostgreSQL -> Kafka -> Neo4j
  - 实时数据同步延迟 < 500ms
  - 幂等性消费保障
  - 数据一致性监控
```

#### 1.2 Query Handler城堡化实现 (Week 1-2)

**严格按照城堡标准实现**：

```go
// 城堡查询结构体 - 完全符合指南标准
type GetOrganizationUnitsQuery struct {
    // 租户隔离 - 城堡多租户核心
    TenantID    uuid.UUID `json:"tenant_id" validate:"required"`
    
    // 过滤条件
    Filters     OrganizationFilters   `json:"filters,omitempty"`
    
    // 分页控制 - 城堡性能保障
    Pagination  PaginationParams      `json:"pagination" validate:"required"`
    
    // 排序控制
    SortBy      []SortField           `json:"sort_by,omitempty"`
    
    // 查询优化提示
    QueryHints  QueryHints            `json:"query_hints,omitempty"`
    
    // 审计字段 - 城堡治理要求
    RequestedBy uuid.UUID             `json:"requested_by" validate:"required"`
    RequestID   uuid.UUID             `json:"request_id" validate:"required"`
}

// 城堡查询处理器 - 统一实现标准
type OrganizationQueryHandler struct {
    // 仓储依赖
    repo         repositories.OrganizationQueryRepository
    
    // 城堡缓存基础设施
    cache        cache.Cache
    cacheConfig  CacheConfig
    
    // 城堡基础设施
    logger       *slog.Logger
    metrics      metrics.Registry
    tracer       trace.Tracer
    
    // 业务服务依赖
    authorizer   auth.Authorizer
}
```

#### 1.3 Neo4j仓储城堡标准实现

**完全按照指南第2.3节实现**：

```go
// 城堡Neo4j查询仓储 - 统一实现标准
type Neo4jOrganizationQueryRepository struct {
    driver      neo4j.Driver
    logger      *slog.Logger
    metrics     metrics.Registry
    
    // 城堡降级机制
    fallbackRepo PostgresOrganizationQueryRepository
    circuitBreaker circuit.Breaker
}

// 城堡图查询标准实现
func (r *Neo4jOrganizationQueryRepository) GetWithRelations(ctx context.Context, id, tenantID uuid.UUID) (*OrganizationWithRelations, error) {
    // 熔断器检查
    if !r.circuitBreaker.Allow() {
        r.logger.Warn("Neo4j circuit breaker open, using fallback")
        return r.fallbackRepo.GetWithRelations(ctx, id, tenantID)
    }

    // 城堡标准Cypher查询
    cypher := `
        MATCH (org:OrganizationUnit {code: $code, tenant_id: $tenant_id})
        WHERE org.status <> 'DELETED'
        OPTIONAL MATCH (org)-[:PARENT_OF]->(child:OrganizationUnit)
        WHERE child.tenant_id = $tenant_id AND child.status <> 'DELETED'
        OPTIONAL MATCH (parent:OrganizationUnit)-[:PARENT_OF]->(org)
        WHERE parent.tenant_id = $tenant_id AND parent.status <> 'DELETED'
        RETURN org, 
               collect(DISTINCT child) as children,
               parent as parent_unit
        ORDER BY org.created_at DESC
    `
    // ... 实现细节按指南标准
}
```

#### 1.4 前端Query Hook城堡化

**完全替换React Query为CQRS Hooks**：

```typescript
// 城堡查询Hooks - 符合指南标准
export function useOrganizationUnitsQuery(params?: OrganizationQueryParams) {
  const tenantId = useCurrentTenantId();
  
  return useQuery({
    queryKey: ['organization-units-cqrs', tenantId, params],
    queryFn: () => organizationQueryAPI.getOrganizationUnits({
      tenant_id: tenantId,
      filters: params?.filters,
      pagination: params?.pagination || { page: 1, page_size: 50 },
      requested_by: useCurrentUserId(),
      request_id: generateRequestId(),
    }),
    // 城堡缓存策略
    staleTime: 5 * 60 * 1000, // 5分钟
    cacheTime: 10 * 60 * 1000, // 10分钟
  });
}
```

**成功标准**: 查询性能提升≥30%，数据一致性≥99.9%

### 阶段2: 命令端城堡化 (2-3周) 🟡 中等风险

**目标**: 启用CQRS命令功能，实现完整事件驱动

#### 2.1 Command Handler城堡化实现 (Week 3-4)

**严格按照指南第1.2节标准**：

```go
// 城堡命令结构体标准格式 - 所有模块必须遵循
type CreateOrganizationUnitCommand struct {
    // 租户隔离 - 城堡多租户核心
    TenantID  uuid.UUID `json:"tenant_id" validate:"required"`
    
    // 业务字段 - 符合城堡业务模型
    Name        string                 `json:"name" validate:"required,max=100"`
    Description *string                `json:"description,omitempty"`
    ParentCode  *string                `json:"parent_code,omitempty"`
    UnitType    string                 `json:"unit_type" validate:"required,oneof=DEPARTMENT COST_CENTER COMPANY PROJECT_TEAM"`
    Profile     map[string]interface{} `json:"profile,omitempty"`
    
    // 审计字段 - 城堡治理要求
    CreatedBy   uuid.UUID `json:"created_by" validate:"required"`
    RequestID   uuid.UUID `json:"request_id" validate:"required"`
    
    // 元数据
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// 城堡命令处理器 - 所有模块统一标准
type OrganizationCommandHandler struct {
    // 仓储依赖
    repo         repositories.OrganizationCommandRepository
    
    // 事件总线 - 城堡事件驱动核心
    eventBus     events.EventBus
    
    // 城堡基础设施
    logger       *slog.Logger
    metrics      metrics.Registry
    tracer       trace.Tracer
    
    // 业务服务依赖
    validator    validator.Validator
    authorizer   auth.Authorizer
}
```

#### 2.2 PostgreSQL仓储城堡标准实现

**完全按照指南第1.3节**：

```go
// 城堡PostgreSQL命令仓储 - 统一实现标准
type PostgresOrganizationCommandRepository struct {
    db       *sql.DB
    logger   *slog.Logger
    metrics  metrics.Registry
    
    // 城堡Outbox模式支持
    outboxRepo outbox.Repository
}

// 城堡事务包装器 - 确保ACID特性
func (r *PostgresOrganizationCommandRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) (*OrganizationCommandResult, error)) (*OrganizationCommandResult, error) {
    tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // 在上下文中传递事务
    txCtx := context.WithValue(ctx, "tx", tx)
    
    result, err := fn(txCtx)
    if err != nil {
        return nil, err
    }
    
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    return result, nil
}
```

#### 2.3 领域事件城堡标准实现

**完全按照指南第3.1节**：

```go
// 城堡领域事件标准实现
type OrganizationUnitCreatedEvent struct {
    // 事件元数据
    EventID       uuid.UUID `json:"event_id"`
    EventType     string    `json:"event_type"`
    Timestamp     time.Time `json:"timestamp"`
    
    // 城堡上下文
    TenantID      uuid.UUID `json:"tenant_id"`
    AggregateID   uuid.UUID `json:"aggregate_id"`
    AggregateType string    `json:"aggregate_type"`
    Version       int64     `json:"version"`
    
    // 城堡治理字段
    CausedBy      uuid.UUID `json:"caused_by"`
    RequestID     uuid.UUID `json:"request_id"`
    CorrelationID uuid.UUID `json:"correlation_id"`
    
    // 业务数据
    Data          OrganizationEventData      `json:"data"`
    Metadata      map[string]interface{}     `json:"metadata"`
}

// 城堡事件数据标准格式
type OrganizationEventData struct {
    // 实体快照 - 事件溯源支持
    EntitySnapshot OrganizationSnapshot `json:"entity_snapshot"`
    
    // 变更详情 - 审计支持
    Changes       []FieldChange `json:"changes,omitempty"`
    
    // 业务上下文
    BusinessContext map[string]interface{} `json:"business_context,omitempty"`
}
```

#### 2.4 Event Consumer城堡标准实现

**严格按照指南第3.3节**：

```go
// 城堡事件消费者标准实现
type OrganizationEventConsumer struct {
    // 仓储依赖
    neo4jRepo    repositories.Neo4jOrganizationRepository
    
    // 城堡基础设施
    logger       *slog.Logger
    metrics      metrics.Registry
    
    // 幂等性保障
    idempotency  idempotency.Service
    
    // 错误处理
    errorHandler ErrorHandler
}

// 城堡事件处理标准流程
func (c *OrganizationEventConsumer) HandleEvent(ctx context.Context, event DomainEvent) error {
    span, ctx := c.tracer.Start(ctx, "OrganizationEventConsumer.HandleEvent")
    defer span.End()
    
    // 第一阶段：幂等性检查
    if processed, err := c.idempotency.IsProcessed(ctx, event.GetID()); err != nil {
        return fmt.Errorf("idempotency check failed: %w", err)
    } else if processed {
        c.logger.Info("Event already processed", "event_id", event.GetID())
        c.metrics.Counter("consumer.duplicate_events").Inc()
        return nil
    }
    
    // 事件类型路由处理...
}
```

**成功标准**: 所有写操作使用CQRS，事件处理延迟<100ms

### 阶段3: 城堡清理与优化 (1周) 🟢 低风险

**目标**: 移除冗余代码，完善城堡监控

#### 3.1 代码城堡化清理

**移除遗留实现**：
- ❌ 删除 `cmd/server/main.go` 中的简化REST实现
- ❌ 删除传统的 `OrganizationHandler` 直接数据库访问
- ❌ 清理前端的传统 React Query 相关代码
- ✅ 保留向后兼容的API路由映射

#### 3.2 双路径API实现

**严格按照ADR-004要求**：

```go
// 业务路径实现 - 适配器模式
r.Route("/api/v1/corehr/organizations", func(r chi.Router) {
    r.Get("/", organizationAdapter.GetOrganizations)      // 通过适配器
    r.Post("/", organizationAdapter.CreateOrganization)   // 通过适配器
    r.Get("/stats", organizationAdapter.GetOrganizationStats) // 业务统计
})

// 技术路径实现 - 直接访问
r.Route("/api/v1/organization-units", func(r chi.Router) {
    r.Get("/", organizationQueryHandler.ListOrganizationUnits)    // 直接查询
    r.Post("/", organizationCommandHandler.CreateOrganizationUnit) // 直接命令
})
```

#### 3.3 城堡监控完善

**按照指南第3节要求**：

```go
// 城堡CQRS监控指标标准定义
var (
    // 命令端指标
    commandDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "castle_cqrs_organization_command_duration_seconds",
            Help: "Duration of organization CQRS command execution",
            Buckets: []float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 10},
        },
        []string{"command_type", "tenant_id", "status"},
    )
    
    // 查询端指标
    queryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "castle_cqrs_organization_query_duration_seconds", 
            Help: "Duration of organization CQRS query execution",
            Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5, 1},
        },
        []string{"query_type", "tenant_id", "cache_hit"},
    )
)
```

**成功标准**: 代码清理度100%，监控覆盖率≥95%

---

## 🎨 GraphQL混合协议实施计划

### GraphQL查询端实现

**按照CQRS指南第2.4节标准**：

```graphql
# 城堡GraphQL Schema标准格式
type OrganizationUnit {
  # 核心标识
  code: String!
  tenantId: ID!
  
  # 基础信息
  name: String!
  description: String
  unitType: OrganizationUnitType!
  status: OrganizationStatus!
  level: Int!
  
  # 关系数据 - 城堡图查询优势
  parent: OrganizationUnit
  children: [OrganizationUnit!]!
  positions: [Position!]!
  employees: [Employee!]!
  
  # 统计数据
  stats: OrganizationStats
  
  # 元数据
  createdAt: DateTime!
  updatedAt: DateTime!
}

# 城堡复杂关系查询
type Query {
  # 组织树查询 - GraphQL核心优势
  organizationTree(
    tenantId: ID!
    rootCode: String
    includeEmployees: Boolean = false
    includePositions: Boolean = false
    maxDepth: Int = 10
  ): [OrganizationUnit!]!
  
  # 组织统计聚合
  organizationStats(
    tenantId: ID!
    filters: OrganizationFilters
  ): OrganizationStatsAggregation!
}
```

### 前端GraphQL集成

**城堡混合模式Hook**：

```typescript
// 城堡混合模式Hook - 智能降级
export function useOrganizationWithRelations(organizationCode: string) {
  const tenantId = useCurrentTenantId();
  
  // GraphQL查询 - 优先使用
  const { 
    data: graphqlData, 
    loading: graphqlLoading, 
    error: graphqlError 
  } = useQuery(GET_ORGANIZATION_WITH_RELATIONS, {
    variables: { code: organizationCode, tenantId },
    errorPolicy: 'all',
    skip: !organizationCode,
  });

  // REST API降级机制
  const shouldUseFallback = graphqlError || !graphqlData?.organizationUnit;
  const { 
    data: restData, 
    error: restError,
    mutate: restMutate
  } = useQuery(
    shouldUseFallback ? `/api/v1/queries/organization-units/${organizationCode}?tenant_id=${tenantId}` : null,
    fetcher
  );

  // 智能数据合并和状态管理
  return useMemo(() => {
    if (graphqlData?.organizationUnit) {
      return {
        organization: graphqlData.organizationUnit,
        source: 'graphql' as const,
        hasFullRelations: true,
        loading: graphqlLoading,
        error: null,
      };
    }
    
    if (restData) {
      return {
        organization: {
          ...restData,
          _needsChildrenLoad: true,
          _needsPositionsLoad: true,
        },
        source: 'rest' as const,
        hasFullRelations: false,
        loading: false,
        error: restError,
      };
    }
    
    return {
      organization: null,
      source: 'none' as const,
      hasFullRelations: false,
      loading: graphqlLoading,
      error: graphqlError || restError,
    };
  }, [graphqlData, restData, graphqlLoading, graphqlError, restError]);
}
```

---

## 📊 实施时间表

### 详细周计划

| 周次 | 阶段 | 主要任务 | 交付物 | 责任人 |
|------|------|----------|--------|--------|
| **W1** | 阶段1 | Neo4j环境 + CDC Pipeline | 数据同步基础设施 | 数据工程师 |
| **W2** | 阶段1 | Query Handler + Neo4j仓储 | CQRS查询端完整实现 | 后端开发 |
| **W3** | 阶段2 | Command Handler + 领域事件 | CQRS命令端基础架构 | 后端开发 |
| **W4** | 阶段2 | Event Consumer + 前端Hook | 事件驱动完整链路 | 全栈开发 |
| **W5** | 阶段3 | 双路径API + 代码清理 | 完整CQRS架构 | 架构师 |

### 关键里程碑检查点

#### M1: 查询端就绪 (Week 2末)
```bash
验收标准:
✅ Neo4j查询响应时间 < 200ms (P95)
✅ 数据一致性检查 > 99.9%
✅ 查询端缓存命中率 > 80%
✅ 前端Query Hook完全替换React Query
```

#### M2: 命令端就绪 (Week 4末)
```bash
验收标准:
✅ 命令响应时间 < 300ms (P95)
✅ 事件处理延迟 < 100ms (P95)
✅ Outbox模式事务一致性 100%
✅ 前端Command Hook乐观更新正常
```

#### M3: 架构完成 (Week 5末)
```bash
验收标准:
✅ 双路径API功能完整
✅ 遗留代码清理完成
✅ 监控指标覆盖 ≥ 95%
✅ 性能提升达到预期 (40-60%)
```

---

## 🧪 测试策略

### 单元测试 (目标覆盖率 ≥ 90%)

**Command Handler测试**：
```go
func TestOrganizationCommandHandler_HandleCreateOrganizationUnit(t *testing.T) {
    // Given - 城堡测试环境准备
    mockRepo := &mocks.OrganizationCommandRepository{}
    mockEventBus := &mocks.EventBus{}
    mockValidator := &mocks.Validator{}
    mockAuthorizer := &mocks.Authorizer{}
    
    handler := NewOrganizationCommandHandler(mockRepo, mockEventBus, mockValidator, mockAuthorizer)
    
    cmd := CreateOrganizationUnitCommand{
        TenantID:    testTenantID,
        Name:        "Test Organization",
        UnitType:    "DEPARTMENT",
        CreatedBy:   testUserID,
        RequestID:   testRequestID,
    }

    // 城堡依赖Mock配置
    mockAuthorizer.On("Authorize", mock.Anything, cmd.CreatedBy, "create", "organization_unit").Return(nil)
    mockValidator.On("Validate", cmd).Return(nil)
    mockRepo.On("WithTransaction", mock.Anything, mock.AnythingOfType("func")).Return(&OrganizationCommandResult{}, nil)

    // When - 执行命令
    result, err := handler.HandleCreateOrganizationUnit(context.Background(), cmd)

    // Then - 城堡断言验证
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.NotEmpty(t, result.Code) // 7位编码
    
    // 城堡Mock验证
    mockRepo.AssertExpectations(t)
    mockEventBus.AssertExpectations(t)
    mockValidator.AssertExpectations(t)
    mockAuthorizer.AssertExpectations(t)
}
```

### 集成测试

**CQRS完整链路测试**：
```go
func TestOrganizationCQRSIntegration(t *testing.T) {
    // 城堡测试环境初始化
    testEnv := setupCastleCQRSTestEnvironment(t)
    defer testEnv.Cleanup()

    // 第一阶段：通过Command创建数据
    cmd := CreateOrganizationUnitCommand{
        TenantID:    testTenantID,
        Name:        "Integration Test Org",
        UnitType:    "DEPARTMENT",
        CreatedBy:   testUserID,
        RequestID:   uuid.New(),
    }

    result, err := testEnv.CommandHandler.HandleCreateOrganizationUnit(context.Background(), cmd)
    require.NoError(t, err)
    require.NotNil(t, result)

    // 第二阶段：验证PostgreSQL写入
    pgEntity, err := testEnv.PostgresRepo.GetByCode(context.Background(), result.Code, testTenantID)
    require.NoError(t, err)
    assert.Equal(t, cmd.Name, pgEntity.Name)

    // 第三阶段：等待事件处理完成
    testEnv.WaitForEventProcessing(result.Code, 5*time.Second)

    // 第四阶段：验证Neo4j查询
    neo4jEntity, err := testEnv.Neo4jRepo.GetByCode(context.Background(), result.Code, testTenantID)
    require.NoError(t, err)
    assert.Equal(t, cmd.Name, neo4jEntity.Name)

    // 第五阶段：验证前端Query Hook
    query := GetOrganizationUnitsQuery{
        TenantID:    testTenantID,
        Filters:     OrganizationFilters{Codes: []string{result.Code}},
        RequestedBy: testUserID,
        RequestID:   uuid.New(),
    }

    view, err := testEnv.QueryHandler.HandleGetOrganizationUnits(context.Background(), query)
    require.NoError(t, err)
    assert.Len(t, view.Organizations, 1)
    assert.Equal(t, cmd.Name, view.Organizations[0].Name)
}
```

### 性能测试

**基准测试标准**：
```bash
# 命令端性能基准
go test -bench=BenchmarkCreateOrganizationCommand -benchmem -count=5
# 目标: < 300ms P95, > 1000 QPS

# 查询端性能基准  
go test -bench=BenchmarkGetOrganizationQuery -benchmem -count=5
# 目标: < 200ms P95, > 5000 QPS

# Neo4j图查询基准
go test -bench=BenchmarkOrganizationTreeQuery -benchmem -count=5
# 目标: < 500ms P95 for 5-level deep trees
```

---

## ⚠️ 风险管理

### 高风险项识别

#### 风险1: 数据一致性风险 🔴
**风险描述**: PostgreSQL与Neo4j数据不同步
**影响程度**: 严重 - 可能导致查询结果不准确
**缓解措施**:
- 实时一致性监控 (每分钟检查)
- 自动数据修复机制
- 熔断器降级到PostgreSQL查询

#### 风险2: 性能下降风险 🟡  
**风险描述**: CQRS引入的复杂性可能短期降低性能
**影响程度**: 中等 - 影响用户体验
**缓解措施**:
- 渐进式切换，保持旧接口作为fallback
- 性能监控告警阈值设置
- 缓存预热策略

#### 风险3: 开发复杂度风险 🟡
**风险描述**: 团队对CQRS架构的学习曲线
**影响程度**: 中等 - 可能延期交付
**缓解措施**:
- CQRS架构培训 (每周五下午)
- 结对编程和Code Review
- 参考员工和职位模块的成功实践

### 应急预案

#### 预案A: 查询性能不达标
```bash
触发条件: 查询P95响应时间 > 500ms
应急措施:
1. 立即启用PostgreSQL查询降级
2. 优化Neo4j查询和索引
3. 调整缓存策略
4. 必要时回滚到阶段1
```

#### 预案B: 数据不一致超阈值
```bash
触发条件: 数据一致性 < 99%
应急措施:  
1. 暂停新的写操作
2. 执行数据修复脚本
3. 重建Neo4j数据
4. 必要时从PostgreSQL完全重同步
```

#### 预案C: 完整回滚
```bash
触发条件: 严重系统故障或多个高风险同时触发
回滚策略:
1. 立即切回传统REST API
2. 停用CQRS命令和查询端点  
3. 恢复原始代码部署
4. 数据修复和一致性验证
回滚时间: < 30分钟
```

---

## 📈 成功度量标准

### 技术指标

```yaml
性能提升:
  查询响应时间: 目标提升 60% (当前 100ms → 40ms P95)
  命令响应时间: 目标保持 < 300ms P95
  系统吞吐量: 目标提升 50% (查询 > 5000 QPS)

架构质量:
  代码覆盖率: ≥ 90%
  架构合规度: 100% (完全符合ADR-004和CQRS指南)
  技术债务消除: 100% (完全重构)

运维质量:
  系统可用性: > 99.9%
  数据一致性: > 99.9% 
  错误率: < 0.1%
  部署成功率: 100%
```

### 业务指标

```yaml
用户体验:
  API响应时间感知: 用户满意度 > 95%
  功能完整性: 100% 向后兼容
  界面响应流畅度: 前端交互延迟 < 100ms

开发效率:
  新功能开发速度: 提升 30% (统一CQRS模式)
  Bug修复时间: 缩短 40% (架构清晰)
  代码维护成本: 降低 50% (消除技术债务)
```

---

## 📚 参考文档对齐

### 严格遵循的架构文档
- ✅ **[ADR-004: 组织单元管理架构决策](../../architecture-decisions/ADR-004-organization-units-architecture.md)** - 适配器模式和双路径API设计
- ✅ **[CQRS统一架构实施指南](../../architecture-foundations/cqrs-unified-implementation-guide.md)** - 三阶段迁移标准和城堡组件规范  
- ✅ **[城堡蓝图](../../architecture-foundations/castle-blueprint.md)** - 整体架构原则
- ✅ **[元合约v6.0规范](../../architecture-foundations/metacontract-v6.0-specification.md)** - 开发标准

### 成功案例参考
- 📋 **[员工管理CQRS实施](../employees-8digit-optimization-guide.md)** - 命令端实施模式
- 📋 **[职位管理CQRS实施](../positions-radical-optimization-guide.md)** - 查询端优化策略

### 技术标准遵循
- ✅ **[组织单元API规范](../../api-specifications/organization-units-api-specification.md)** - 接口设计标准
- ✅ **[开发测试修复标准](../../standards/development-testing-fixing-standards.md)** - 质量保证流程
- ✅ **[标识符命名策略](../../architecture-decisions/ADR-006-identifier-naming-strategy.md)** - 7位编码标准

---

## 🎯 下一步行动

### 立即行动项 (本周)
- [ ] **架构审查**: 项目架构委员会审批本迁移计划
- [ ] **团队组建**: 分配专门的CQRS实施团队
- [ ] **环境准备**: 建立开发和测试环境
- [ ] **培训安排**: 安排团队CQRS架构培训

### Week 1 启动准备
- [ ] **Neo4j环境**: Docker集群部署和配置
- [ ] **CDC Pipeline**: Kafka数据同步管道建立  
- [ ] **基础设施**: 监控、日志、追踪系统配置
- [ ] **代码仓库**: 分支管理和CI/CD准备

**负责人**: 系统架构师  
**审核人**: 项目架构委员会  
**下次更新**: 2025-08-13

---

*本计划严格遵循城堡架构宪章，确保100%架构合规，零偏离实施。*