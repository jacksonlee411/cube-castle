# Neo4j集成API文档

**版本**: v3.0.0  
**更新日期**: 2025年7月30日  
**状态**: 已完成

## 📋 概览

Neo4j集成提供了强大的图数据库操作能力，支持复杂关系查询、数据同步、性能监控等企业级功能。本文档详细介绍了所有可用的API接口和使用方法。

## 🏗️ 架构概述

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ 关系型数据库     │───▶│ 数据同步服务      │───▶│ Neo4j图数据库   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                               │                         │
                               ▼                         ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ 监控告警系统     │◀───│ 指标收集器        │◀───│ 图查询接口      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## 🔌 核心组件

### 1. ConnectionManager - 连接管理器

负责Neo4j数据库连接管理、事务处理和健康监控。

#### 1.1 初始化连接

```go
// 创建连接配置
config := neo4j.ConnectionConfig{
    URI:                    "neo4j://localhost:7687",
    Username:               "neo4j",
    Password:               "password",
    MaxConnectionPoolSize:  50,
    ConnectionTimeout:      30 * time.Second,
    MaxTransactionRetryTime: 15 * time.Second,
    Database:               "neo4j",
    EnableEncryption:       false,
    TrustStrategy:          "TRUST_ALL_CERTIFICATES",
}

// 创建连接管理器
connectionManager, err := neo4j.NewConnectionManager(config)
if err != nil {
    log.Fatal("创建连接管理器失败:", err)
}

// 建立连接
ctx := context.Background()
err = connectionManager.Connect(ctx)
if err != nil {
    log.Fatal("连接Neo4j失败:", err)
}
```

#### 1.2 核心方法

##### Connect(ctx context.Context) error
建立并验证数据库连接。

**参数**:
- `ctx`: 上下文对象

**返回值**:
- `error`: 连接失败时返回错误

**示例**:
```go
err := connectionManager.Connect(ctx)
if err != nil {
    log.Printf("连接失败: %v", err)
}
```

##### ExecuteQuery(ctx context.Context, query string, params map[string]interface{}) (*neo4j.EagerResult, error)
执行Cypher查询。

**参数**:
- `ctx`: 上下文对象
- `query`: Cypher查询语句
- `params`: 查询参数

**返回值**:
- `*neo4j.EagerResult`: 查询结果
- `error`: 查询失败时返回错误

**示例**:
```go
query := "MATCH (n:Employee {tenant_id: $tenant_id}) RETURN n.name as name"
params := map[string]interface{}{
    "tenant_id": "tenant-123",
}

result, err := connectionManager.ExecuteQuery(ctx, query, params)
if err != nil {
    log.Printf("查询失败: %v", err)
    return
}

for _, record := range result.Records {
    name, _ := record.Get("name")
    fmt.Printf("员工姓名: %s\n", name)
}
```

##### ExecuteTransaction(ctx context.Context, txFunc func(neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error)
执行读写事务。

**参数**:
- `ctx`: 上下文对象
- `txFunc`: 事务函数

**返回值**:
- `interface{}`: 事务结果
- `error`: 事务失败时返回错误

**示例**:
```go
result, err := connectionManager.ExecuteTransaction(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
    // 在事务中执行多个操作
    query1 := "CREATE (e:Employee {id: $id, name: $name})"
    params1 := map[string]interface{}{
        "id": "emp-001",
        "name": "张三",
    }
    
    _, err := tx.Run(query1, params1)
    if err != nil {
        return nil, err
    }
    
    // 更多操作...
    return "success", nil
})
```

##### CheckHealth(ctx context.Context) error
检查连接健康状态。

**返回值**:
- `error`: 健康检查失败时返回错误

##### GetConnectionInfo() map[string]interface{}
获取连接信息。

**返回值**:
- `map[string]interface{}`: 连接信息对象

### 2. GraphService - 图数据服务

提供高级图数据操作接口，支持节点和关系的CRUD操作。

#### 2.1 初始化服务

```go
graphService := neo4j.NewGraphService(connectionManager)
```

#### 2.2 节点操作

##### CreateEmployeeNode(ctx context.Context, employee *EmployeeNode) error
创建员工节点。

**参数**:
- `ctx`: 上下文对象
- `employee`: 员工节点数据

**示例**:
```go
employee := &neo4j.EmployeeNode{
    ID:         "emp-001",
    TenantID:   "tenant-123",
    Name:       "张三",
    Email:      "zhangsan@example.com",
    Position:   "软件工程师",
    Department: "技术部",
    Status:     "ACTIVE",
    CreatedAt:  time.Now(),
    UpdatedAt:  time.Now(),
}

err := graphService.CreateEmployeeNode(ctx, employee)
if err != nil {
    log.Printf("创建员工节点失败: %v", err)
}
```

##### CreatePositionNode(ctx context.Context, position *PositionNode) error
创建岗位节点。

**参数**:
- `ctx`: 上下文对象
- `position`: 岗位节点数据

**示例**:
```go
position := &neo4j.PositionNode{
    ID:          "pos-001",
    TenantID:    "tenant-123",
    Title:       "高级软件工程师",
    Description: "负责系统架构设计",
    Department:  "技术部",
    Level:       5,
    Type:        "FULL_TIME",
    CreatedAt:   time.Now(),
    UpdatedAt:   time.Now(),
}

err := graphService.CreatePositionNode(ctx, position)
```

##### CreateOrganizationUnitNode(ctx context.Context, orgUnit *OrganizationUnitNode) error
创建组织单位节点。

**示例**:
```go
orgUnit := &neo4j.OrganizationUnitNode{
    ID:          "org-001",
    TenantID:    "tenant-123",
    Name:        "技术部",
    Type:        "DEPARTMENT",
    Description: "负责产品技术研发",
    Level:       2,
    CreatedAt:   time.Now(),
    UpdatedAt:   time.Now(),
}

err := graphService.CreateOrganizationUnitNode(ctx, orgUnit)
```

#### 2.3 关系操作

##### CreateEmployeePositionRelationship(ctx context.Context, employeeID, positionID, tenantID string, fromDate, toDate *time.Time, isActive bool) error
创建员工-岗位关系。

**参数**:
- `employeeID`: 员工ID
- `positionID`: 岗位ID
- `tenantID`: 租户ID
- `fromDate`: 开始时间
- `toDate`: 结束时间（可为空）
- `isActive`: 是否活跃

**示例**:
```go
fromDate := time.Now().Add(-30 * 24 * time.Hour)
err := graphService.CreateEmployeePositionRelationship(
    ctx,
    "emp-001",
    "pos-001", 
    "tenant-123",
    &fromDate,
    nil, // 当前职位，结束时间为空
    true,
)
```

#### 2.4 查询操作

##### GetEmployeeCareerPath(ctx context.Context, employeeID, tenantID string) ([]map[string]interface{}, error)
获取员工职业路径。

**返回值**:
- 职业路径记录数组，包含岗位信息、时间范围等

**示例**:
```go
careerPath, err := graphService.GetEmployeeCareerPath(ctx, "emp-001", "tenant-123")
if err != nil {
    log.Printf("获取职业路径失败: %v", err)
    return
}

for _, step := range careerPath {
    fmt.Printf("岗位: %s, 部门: %s, 级别: %v\n", 
        step["position_title"], 
        step["department"], 
        step["level"])
}
```

##### GetOrganizationHierarchy(ctx context.Context, tenantID string, rootID *string) ([]map[string]interface{}, error)
获取组织架构层级。

**参数**:
- `tenantID`: 租户ID
- `rootID`: 根节点ID（可为空，获取全部）

**示例**:
```go
hierarchy, err := graphService.GetOrganizationHierarchy(ctx, "tenant-123", nil)
if err != nil {
    log.Printf("获取组织架构失败: %v", err)
    return
}

for _, org := range hierarchy {
    fmt.Printf("组织: %s, 类型: %s, 级别: %v\n",
        org["name"],
        org["type"], 
        org["level"])
}
```

### 3. SyncService - 数据同步服务

负责关系型数据库到图数据库的数据同步。

#### 3.1 初始化服务

```go
// 注意：需要提供真实的ent.Client实例
syncService := neo4j.NewSyncService(graphService, entClient)
```

#### 3.2 同步操作

##### SyncBusinessProcessEvent(ctx context.Context, eventData map[string]interface{}) (*SyncResult, error)
同步业务流程事件。

**参数**:
- `eventData`: 事件数据，包含事件类型、实体ID等信息

**返回值**:
- `*SyncResult`: 同步结果，包含成功状态、处理时间等

**示例**:
```go
eventData := map[string]interface{}{
    "event_type":  "HR.Employee.Hired",
    "entity_id":   "emp-001",
    "entity_type": "Employee",
    "tenant_id":   "tenant-123",
}

result, err := syncService.SyncBusinessProcessEvent(ctx, eventData)
if err != nil {
    log.Printf("同步失败: %v", err)
    return
}

fmt.Printf("同步结果: 成功=%v, 耗时=%v\n", 
    result.Success, 
    result.ProcessingTime)
```

##### FullSync(ctx context.Context, tenantID string) (*SyncStats, error)
执行全量数据同步。

**返回值**:
- `*SyncStats`: 同步统计信息

**示例**:
```go
stats, err := syncService.FullSync(ctx, "tenant-123")
if err != nil {
    log.Printf("全量同步失败: %v", err)
    return
}

fmt.Printf("同步统计: 总计=%d, 成功=%d, 失败=%d, 成功率=%.2f%%\n",
    stats.TotalEvents,
    stats.SuccessCount, 
    stats.FailureCount,
    stats.SuccessRate)
```

### 4. GraphQueryInterface - 图查询接口

提供高级图查询和分析功能。

#### 4.1 初始化接口

```go
queryInterface := neo4j.NewGraphQueryInterface(graphService)
```

#### 4.2 分析查询

##### GetCareerPathAnalysis(ctx context.Context, employeeID, tenantID string) (*CareerPathAnalysis, error)
获取职业路径分析。

**返回值**:
- `*CareerPathAnalysis`: 详细的职业路径分析结果

**示例**:
```go
analysis, err := queryInterface.GetCareerPathAnalysis(ctx, "emp-001", "tenant-123")
if err != nil {
    log.Printf("获取职业路径分析失败: %v", err)
    return
}

fmt.Printf("员工: %s, 当前职位: %s\n", 
    analysis.EmployeeName, 
    analysis.CurrentPosition)

fmt.Printf("职业指标: 总岗位数=%d, 晋升次数=%d, 平均任职时间=%v\n",
    analysis.CareerMetrics.TotalPositions,
    analysis.CareerMetrics.PromotionCount,
    analysis.CareerMetrics.AverageStayPeriod)

for _, step := range analysis.CareerProgression {
    fmt.Printf("  - %s (%s) [%s - %s]\n",
        step.Position,
        step.Department, 
        step.StartDate.Format("2006-01-02"),
        func() string {
            if step.EndDate != nil {
                return step.EndDate.Format("2006-01-02")
            }
            return "至今"
        }())
}
```

##### GetOrganizationInsight(ctx context.Context, tenantID string, rootID *string) (*OrganizationInsight, error)
获取组织洞察分析。

**示例**:
```go
insight, err := queryInterface.GetOrganizationInsight(ctx, "tenant-123", nil)
if err != nil {
    log.Printf("获取组织洞察失败: %v", err)
    return
}

fmt.Printf("组织指标: 总节点数=%d, 最大深度=%d, 平均子节点数=%.1f\n",
    insight.Metrics.TotalNodes,
    insight.Metrics.MaxDepth,
    insight.Metrics.AvgChildrenCount)

for nodeType, count := range insight.PositionDistribution {
    fmt.Printf("岗位类型 %s: %d个\n", nodeType, count)
}
```

#### 4.3 通用查询

##### ExecuteQuery(ctx context.Context, req *QueryRequest) (*QueryResult, error)
执行通用查询。

**参数**:
- `req`: 查询请求对象

**查询类型**:
- `career_path`: 职业路径查询
- `organization_hierarchy`: 组织架构查询
- `workflow_dependencies`: 工作流依赖查询
- `relationship_analysis`: 关系分析查询
- `custom_cypher`: 自定义Cypher查询

**示例**:
```go
// 组织架构查询
req := &neo4j.QueryRequest{
    TenantID:  "tenant-123",
    QueryType: "organization_hierarchy",
    Parameters: map[string]interface{}{
        "root_id": "org-001",
    },
    Limit: 100,
}

result, err := queryInterface.ExecuteQuery(ctx, req)
if err != nil {
    log.Printf("查询失败: %v", err)
    return
}

fmt.Printf("查询结果: 类型=%s, 总数=%d, 耗时=%v\n",
    result.QueryType,
    result.Total, 
    result.ExecutionTime)

// 自定义Cypher查询
customReq := &neo4j.QueryRequest{
    TenantID:  "tenant-123",
    QueryType: "custom_cypher",
    Parameters: map[string]interface{}{
        "query": `
            MATCH (e:Employee {tenant_id: $tenant_id})-[r:HOLDS_POSITION]->(p:Position)
            RETURN e.name as employee_name, p.title as position_title
            ORDER BY e.name
        `,
    },
}

customResult, err := queryInterface.ExecuteQuery(ctx, customReq)
```

## 📊 监控和指标

### 5. MetricsCollector - 指标收集器

提供工作流性能指标收集功能。

#### 5.1 初始化收集器

```go
// 需要提供ent.Client实例
metricsCollector := service.NewMetricsCollector(entClient)

// 启动指标收集（5秒间隔）
ctx := context.Background()
go metricsCollector.Start(ctx, 5*time.Second)
```

#### 5.2 获取指标

##### GetWorkflowMetrics(tenantID, workflowType string) (*WorkflowMetrics, bool)
获取工作流指标。

**示例**:
```go
metrics, exists := metricsCollector.GetWorkflowMetrics("tenant-123", "EmployeeOnboarding")
if exists {
    fmt.Printf("工作流指标:\n")
    fmt.Printf("  总实例数: %d\n", metrics.TotalInstances)
    fmt.Printf("  活跃实例: %d\n", metrics.ActiveInstances)
    fmt.Printf("  成功率: %.2f%%\n", metrics.SuccessRate)
    fmt.Printf("  平均持续时间: %v\n", metrics.AverageDuration)
    fmt.Printf("  每小时吞吐量: %.2f\n", metrics.ThroughputPerHour)
}
```

##### GetPerformanceSnapshot(ctx context.Context, tenantID string) (*PerformanceSnapshot, error)
获取性能快照。

**示例**:
```go
snapshot, err := metricsCollector.GetPerformanceSnapshot(ctx, "tenant-123")
if err != nil {
    log.Printf("获取性能快照失败: %v", err)
    return
}

fmt.Printf("性能快照 [%s]:\n", snapshot.Timestamp.Format("2006-01-02 15:04:05"))
fmt.Printf("工作流指标数量: %d\n", len(snapshot.WorkflowMetrics))
fmt.Printf("步骤指标数量: %d\n", len(snapshot.StepMetrics))
fmt.Printf("告警数量: %d\n", len(snapshot.Alerts))

// 遍历告警
for _, alert := range snapshot.Alerts {
    fmt.Printf("  [%s] %s: %s\n", alert.Severity, alert.Type, alert.Message)
}
```

### 6. MonitoringService - 监控服务

提供系统健康监控和告警功能。

#### 6.1 初始化监控服务

```go
monitoringService := service.NewMonitoringService(metricsCollector)

// 启动监控服务（30秒间隔）
go monitoringService.Start(ctx, 30*time.Second)
```

#### 6.2 健康检查

##### GetSystemHealth(ctx context.Context) *SystemHealth
获取系统健康状态。

**示例**:
```go
health := monitoringService.GetSystemHealth(ctx)

fmt.Printf("系统健康状态: %s [%s]\n", 
    health.Status, 
    health.Timestamp.Format("2006-01-02 15:04:05"))

fmt.Printf("组件统计: 健康=%d, 不健康=%d, 降级=%d\n",
    health.Summary.Healthy,
    health.Summary.Unhealthy, 
    health.Summary.Degraded)

// 遍历组件状态
for name, component := range health.Components {
    fmt.Printf("  %s: %s - %s\n", 
        name, 
        component.Status, 
        component.Message)
}
```

#### 6.3 告警管理

##### GetActiveAlerts() []*ActiveAlert
获取活跃告警。

**示例**:
```go
alerts := monitoringService.GetActiveAlerts()

fmt.Printf("活跃告警数量: %d\n", len(alerts))

for _, alert := range alerts {
    fmt.Printf("[%s] %s: %s\n", 
        alert.Severity, 
        alert.RuleName, 
        alert.Message)
    fmt.Printf("  触发时间: %s\n", 
        alert.TriggerTime.Format("2006-01-02 15:04:05"))
    fmt.Printf("  当前值: %v\n", alert.Value)
}
```

##### AddAlertRule(rule AlertRule)
添加告警规则。

**示例**:
```go
rule := service.AlertRule{
    Name:        "高错误率告警",
    Description: "当步骤错误率超过5%时触发告警",
    MetricType:  "step_error_rate",
    Condition: service.AlertCondition{
        Operator:  "gt",
        Threshold: 5.0,
        Duration:  5 * time.Minute,
        Function:  "avg",
    },
    Severity: service.SeverityWarning,
    Enabled:  true,
    Cooldown: 10 * time.Minute,
    Actions: []service.AlertAction{
        {
            Type:   "log",
            Target: "monitoring.log",
        },
    },
    Tags: map[string]string{
        "category": "performance",
        "priority": "high",
    },
}

monitoringService.AddAlertRule(rule)
```

## 🔧 配置参数

### ConnectionConfig - 连接配置

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| URI | string | neo4j://localhost:7687 | Neo4j连接URI |
| Username | string | neo4j | 用户名 |
| Password | string | password | 密码 |
| MaxConnectionPoolSize | int | 50 | 最大连接池大小 |
| ConnectionTimeout | time.Duration | 30s | 连接超时时间 |
| MaxTransactionRetryTime | time.Duration | 15s | 事务重试时间 |
| Database | string | neo4j | 数据库名称 |
| EnableEncryption | bool | false | 是否启用加密 |
| TrustStrategy | string | TRUST_ALL_CERTIFICATES | 证书信任策略 |

### AlertCondition - 告警条件

| 操作符 | 说明 | 示例 |
|--------|------|------|
| gt | 大于 | value > threshold |
| gte | 大于等于 | value >= threshold |
| lt | 小于 | value < threshold |
| lte | 小于等于 | value <= threshold |
| eq | 等于 | value == threshold |
| ne | 不等于 | value != threshold |

## 🚨 错误处理

### 常见错误类型

1. **连接错误**
```go
// 连接失败
if err != nil {
    if strings.Contains(err.Error(), "connection refused") {
        log.Printf("Neo4j服务不可用: %v", err)
        // 实施降级策略
    }
}
```

2. **查询错误**
```go
// Cypher语法错误
if err != nil {
    if strings.Contains(err.Error(), "SyntaxError") {
        log.Printf("Cypher语法错误: %v", err)
        // 记录查询语句用于调试
    }
}
```

3. **事务错误**
```go
// 事务冲突
if err != nil {
    if strings.Contains(err.Error(), "DeadlockDetected") {
        log.Printf("检测到死锁，将重试: %v", err)
        // 自动重试逻辑
    }
}
```

### 重试策略

系统内置指数退避重试机制：
- 初始延迟: 1秒
- 最大重试次数: 3次
- 退避因子: 2
- 最大延迟: 30秒

## 📈 性能优化建议

### 1. 连接池优化
```go
config.MaxConnectionPoolSize = 100  // 高并发场景
config.ConnectionTimeout = 10 * time.Second  // 快速失败
```

### 2. 查询优化
```go
// 1. 使用索引
// 确保关键字段有索引：id, tenant_id, email等

// 2. 限制结果集
query := "MATCH (n:Employee) RETURN n LIMIT 100"

// 3. 避免笛卡尔积
query := "MATCH (e:Employee)-[r:HOLDS_POSITION]->(p:Position) RETURN e, p"
```

### 3. 批量操作
```go
// 批量创建节点
tx, err := connectionManager.ExecuteTransaction(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
    for _, employee := range employees {
        query := "CREATE (e:Employee {id: $id, name: $name})"
        params := map[string]interface{}{
            "id": employee.ID,
            "name": employee.Name,
        }
        tx.Run(query, params)
    }
    return nil, nil
})
```

## 🧪 测试指南

### 单元测试
```bash
# 运行Neo4j集成测试
go test -v ./internal/neo4j/...

# 运行基准测试
go test -bench=. ./internal/neo4j/...
```

### 集成测试
```bash
# 需要先启动Neo4j服务
docker run -p 7687:7687 -p 7474:7474 -e NEO4J_AUTH=neo4j/testpassword neo4j:latest

# 运行集成测试
go test -v -tags=integration ./internal/neo4j/...
```

## 📞 支持和反馈

- **技术支持**: 架构师团队
- **Bug报告**: 通过项目Issue系统
- **功能请求**: 产品需求管理流程
- **文档更新**: 技术文档团队

---

**文档版本**: v3.0.0  
**最后更新**: 2025年7月30日  
**维护团队**: Neo4j集成开发团队