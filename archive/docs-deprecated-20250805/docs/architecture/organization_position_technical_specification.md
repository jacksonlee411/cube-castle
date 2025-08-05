# **组织与岗位模型技术规范参考**

**文档类型**: 技术规范参考  
**创建时间**: 2025-07-29  
**版本**: v1.1  
**状态**: 参考指南  
**适用对象**: 开发团队、架构师、代码审查员

**相关文档**:
- [Employee Organization Position Analysis](./employee_organization_position_analysis.md) - 当前状态分析
- [Employee Optimization Implementation Plan](./employee_optimization_implementation_plan.md) - 实施方案

---

## **🎯 核心设计模式**

### **1. 多态实体模式 (Polymorphic Entity Pattern)**

#### **实现原理**
```go
// 核心模式：鉴别器 + JSON插槽
type OrganizationUnit struct {
    ID       uuid.UUID `json:"id"`
    UnitType string    `json:"unit_type"` // 鉴别器
    Profile  json.RawMessage `json:"profile"` // 多态插槽
}

// 类型安全的访问方法
func (ou *OrganizationUnit) GetDepartmentProfile() (*DepartmentProfile, error) {
    if ou.UnitType != "DEPARTMENT" {
        return nil, errors.New("not a department")
    }
    var profile DepartmentProfile
    return &profile, json.Unmarshal(ou.Profile, &profile)
}
```

#### **最佳实践**
```go
// ✅ 正确：类型安全的多态处理
func CreateOrganizationUnit(req CreateRequest) error {
    switch req.UnitType {
    case "DEPARTMENT":
        profile := &DepartmentProfile{}
        if err := json.Unmarshal(req.Profile, profile); err != nil {
            return fmt.Errorf("invalid department profile: %w", err)
        }
        return validateDepartmentProfile(profile)
    case "COST_CENTER":
        // 类似处理...
    }
}

// ❌ 错误：缺乏类型验证
func CreateOrganizationUnitWrong(req CreateRequest) error {
    // 直接存储JSON，缺乏验证
    return db.Create(req)
}
```

### **2. 事件溯源模式 (Event Sourcing Pattern)**

#### **事务性发件箱实现**
```go
// 标准模式：业务操作 + 事件发布在同一事务中
func (s *OrganizationService) CreateUnit(ctx context.Context, req *CreateUnitRequest) error {
    return s.db.WithTx(ctx, func(tx *ent.Tx) error {
        // 1. 业务操作
        unit, err := tx.OrganizationUnit.Create().
            SetTenantID(req.TenantID).
            SetUnitType(req.UnitType).
            SetName(req.Name).
            Save(ctx)
        if err != nil {
            return err
        }

        // 2. 事件发布（同一事务）
        event := &OrganizationUnitCreatedEvent{
            UnitID:   unit.ID,
            TenantID: req.TenantID,
            // ... 其他字段
        }
        
        return tx.OutboxEvent.Create().
            SetEventType("organization.unit.created").
            SetEventData(event).
            Save(ctx)
    })
}
```

#### **事件处理器模式**
```go
// 幂等性事件处理器
type EventHandler interface {
    Handle(ctx context.Context, event Event) error
    EventType() string
}

type OrganizationGraphSyncHandler struct {
    neo4j neo4j.Driver
}

func (h *OrganizationGraphSyncHandler) Handle(ctx context.Context, event Event) error {
    // 幂等性检查
    if h.isAlreadyProcessed(event.ID) {
        return nil
    }
    
    // 处理逻辑
    return h.syncToGraph(event)
}
```

### **3. 图关系映射模式 (Graph Relationship Mapping)**

#### **双向同步策略**
```go
// PostgreSQL → Neo4j 同步
func (s *GraphSyncService) SyncOrganizationUnit(event *OrganizationUnitCreatedEvent) error {
    session := s.driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close()

    return session.WriteTransaction(func(tx neo4j.Transaction) error {
        // 创建节点
        nodeQuery := `
        MERGE (ou:OrgUnit {id: $id, tenant_id: $tenant_id})
        SET ou.name = $name, ou.unit_type = $unit_type, ou.updated_at = datetime()
        `
        
        _, err := tx.Run(nodeQuery, map[string]interface{}{
            "id":        event.UnitID.String(),
            "tenant_id": event.TenantID.String(),
            "name":      event.Name,
            "unit_type": event.UnitType,
        })
        
        // 创建关系（如果有父级）
        if event.ParentUnitID != nil {
            relationQuery := `
            MATCH (child:OrgUnit {id: $child_id, tenant_id: $tenant_id})
            MATCH (parent:OrgUnit {id: $parent_id, tenant_id: $tenant_id})
            MERGE (child)-[:PART_OF]->(parent)
            `
            _, err = tx.Run(relationQuery, map[string]interface{}{
                "child_id":  event.UnitID.String(),
                "parent_id": event.ParentUnitID.String(),
                "tenant_id": event.TenantID.String(),
            })
        }
        
        return err
    })
}
```

---

## **🔒 安全架构规范**

### **1. 多租户隔离模式**

#### **数据库行级安全(RLS)**
```sql
-- 组织单元表RLS策略
CREATE POLICY tenant_isolation_organization_units ON organization_units
    FOR ALL TO authenticated
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- 岗位表RLS策略  
CREATE POLICY tenant_isolation_positions ON positions
    FOR ALL TO authenticated
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- 历史表RLS策略
CREATE POLICY tenant_isolation_position_history ON position_attribute_history
    FOR ALL TO authenticated
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

#### **应用层租户上下文**
```go
// 中间件：设置租户上下文
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID := extractTenantID(r) // 从JWT或Header提取
        
        // 设置数据库会话变量
        ctx := context.WithValue(r.Context(), "tenant_id", tenantID)
        
        // 传递给下游处理器
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 数据库操作前设置租户ID
func (s *OrganizationService) withTenantScope(ctx context.Context, fn func(*ent.Client) error) error {
    tenantID := ctx.Value("tenant_id").(uuid.UUID)
    
    // 设置PostgreSQL会话变量
    if _, err := s.db.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID); err != nil {
        return err
    }
    
    return fn(s.db)
}
```

### **2. OPA策略集成**

#### **组织权限策略**
```rego
# organization_policies.rego
package organization

# 基础权限检查
default allow = false

# 允许用户访问自己租户的组织单元
allow {
    input.method == "GET"
    input.resource == "organization_units"
    input.user.tenant_id == input.target.tenant_id
}

# 管理员可以创建组织单元
allow {
    input.method == "POST"
    input.resource == "organization_units"
    input.user.roles[_] == "admin"
    input.user.tenant_id == input.data.tenant_id
}

# 部门负责人可以查看下属组织
allow {
    input.method == "GET"
    input.resource == "organization_units"
    manages_department(input.user.id, input.target.id)
}

manages_department(user_id, org_unit_id) {
    # 通过图查询检查管理关系
    data.graph.query[_].user_id == user_id
    data.graph.query[_].manages == org_unit_id
}
```

#### **策略执行代码**
```go
// OPA策略检查
func (s *AuthorizationService) CheckPermission(ctx context.Context, action, resource string, target interface{}) error {
    user := getUserFromContext(ctx)
    
    input := map[string]interface{}{
        "method":   action,
        "resource": resource,
        "user":     user,
        "target":   target,
    }
    
    result, err := s.opa.Query(ctx, "data.organization.allow", input)
    if err != nil {
        return err
    }
    
    if !result.Allowed() {
        return errors.New("access denied")
    }
    
    return nil
}
```

---

## **🎨 API设计规范**

### **1. RESTful API模式**

#### **资源命名约定**
```go
// ✅ 正确：复数名词，层级清晰
GET    /api/v1/organization-units
POST   /api/v1/organization-units  
GET    /api/v1/organization-units/{id}
PUT    /api/v1/organization-units/{id}
DELETE /api/v1/organization-units/{id}

GET    /api/v1/organization-units/{id}/positions
GET    /api/v1/organization-units/{id}/children

GET    /api/v1/positions
POST   /api/v1/positions
GET    /api/v1/positions/{id}
GET    /api/v1/positions/{id}/history
GET    /api/v1/positions/{id}/occupancy-history

// ❌ 错误：动词形式，嵌套过深
POST   /api/v1/createOrganization
GET    /api/v1/organizations/{id}/departments/{dept_id}/positions/{pos_id}/history
```

#### **请求/响应格式**
```go
// 标准请求格式
type CreateOrganizationUnitRequest struct {
    TenantID     uuid.UUID   `json:"tenant_id" validate:"required"`
    UnitType     string      `json:"unit_type" validate:"required,oneof=DEPARTMENT COST_CENTER COMPANY PROJECT_TEAM"`
    Name         string      `json:"name" validate:"required,min=1,max=100"`
    Description  *string     `json:"description,omitempty"`
    ParentUnitID *uuid.UUID  `json:"parent_unit_id,omitempty"`
    Profile      interface{} `json:"profile,omitempty"`
}

// 标准响应格式
type OrganizationUnitResponse struct {
    ID           uuid.UUID   `json:"id"`
    TenantID     uuid.UUID   `json:"tenant_id"`
    UnitType     string      `json:"unit_type"`
    Name         string      `json:"name"`
    Description  *string     `json:"description,omitempty"`
    ParentUnitID *uuid.UUID  `json:"parent_unit_id,omitempty"`
    Status       string      `json:"status"`
    Profile      interface{} `json:"profile,omitempty"`
    CreatedAt    time.Time   `json:"created_at"`
    UpdatedAt    time.Time   `json:"updated_at"`
}

// 错误响应格式
type ErrorResponse struct {
    Error   string                 `json:"error"`
    Code    string                 `json:"code"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

### **2. GraphQL查询优化**

#### **组织架构查询**
```graphql
# 高效的层级查询
query GetOrganizationChart($tenantId: UUID!, $maxDepth: Int = 5) {
  organizationUnits(tenantId: $tenantId, rootOnly: true) {
    id
    name
    unitType
    children(maxDepth: $maxDepth) {
      id
      name
      unitType
      positions {
        id
        status
        occupant {
          id
          name
        }
      }
      children {
        # 递归结构
      }
    }
  }
}
```

#### **解析器优化**
```go
// DataLoader模式避免N+1查询
func (r *OrganizationUnitResolver) Children(ctx context.Context, obj *OrganizationUnit, maxDepth *int) ([]*OrganizationUnit, error) {
    // 使用DataLoader批量加载
    loader := dataloader.GetOrganizationUnitLoader(ctx)
    
    children, err := loader.LoadMany(ctx, []uuid.UUID{obj.ID})
    if err != nil {
        return nil, err
    }
    
    // 递归深度控制
    if maxDepth != nil && *maxDepth <= 1 {
        return children, nil
    }
    
    // 继续加载下级
    return r.loadChildrenRecursive(ctx, children, maxDepth)
}
```

---

## **📊 性能优化规范**

### **1. 数据库查询优化**

#### **索引策略**
```sql
-- 租户隔离查询优化
CREATE INDEX CONCURRENTLY idx_org_units_tenant_type ON organization_units(tenant_id, unit_type);
CREATE INDEX CONCURRENTLY idx_positions_tenant_dept ON positions(tenant_id, department_id);

-- 层级查询优化
CREATE INDEX CONCURRENTLY idx_org_units_parent ON organization_units(parent_unit_id) WHERE parent_unit_id IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_positions_manager ON positions(manager_position_id) WHERE manager_position_id IS NOT NULL;

-- 历史查询优化
CREATE INDEX CONCURRENTLY idx_position_history_effective ON position_attribute_history(position_id, effective_date DESC);
CREATE INDEX CONCURRENTLY idx_occupancy_history_active ON position_occupancy_history(position_id, is_active) WHERE is_active = true;
```

#### **查询模式**
```go
// ✅ 高效：批量预加载
func (s *OrganizationService) GetUnitsWithPositions(ctx context.Context, unitIDs []uuid.UUID) ([]*OrganizationUnit, error) {
    return s.db.OrganizationUnit.
        Query().
        Where(organizationunit.IDIn(unitIDs...)).
        WithPositions(func(q *ent.PositionQuery) {
            q.Where(position.StatusEQ(position.StatusFILLED)).
                WithCurrentOccupant()
        }).
        All(ctx)
}

// ❌ 低效：N+1查询
func (s *OrganizationService) GetUnitsWithPositionsSlow(ctx context.Context, unitIDs []uuid.UUID) ([]*OrganizationUnit, error) {
    units, err := s.db.OrganizationUnit.Query().Where(organizationunit.IDIn(unitIDs...)).All(ctx)
    if err != nil {
        return nil, err
    }
    
    for _, unit := range units {
        // 每个单元单独查询位置 - N+1问题
        positions, _ := s.db.Position.Query().Where(position.DepartmentIDEQ(unit.ID)).All(ctx)
        unit.Positions = positions
    }
    
    return units, nil
}
```

### **2. 图数据库优化**

#### **Cypher查询优化**
```cypher
-- ✅ 高效：使用索引和限制范围
MATCH (emp:Employee {tenant_id: $tenant_id})-[:OCCUPIES]->(pos:Position)
WHERE pos.tenant_id = $tenant_id
WITH pos
MATCH path = (pos)-[:REPORTS_TO*0..5]->(manager:Position)
WHERE manager.tenant_id = $tenant_id
RETURN path
LIMIT 100

-- ❌ 低效：无索引，无深度限制
MATCH path = (emp:Employee)-[:OCCUPIES]->(pos:Position)-[:REPORTS_TO*]->(manager:Position)
RETURN path
```

#### **连接池配置**
```go
// Neo4j连接池优化
func NewNeo4jDriver(uri, username, password string) (neo4j.Driver, error) {
    return neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""), func(config *neo4j.Config) {
        config.MaxConnectionPoolSize = 100
        config.MaxTransactionRetryTime = 15 * time.Second
        config.MaxConnectionLifetime = 5 * time.Minute
        config.ConnectionAcquisitionTimeout = 2 * time.Minute
    })
}
```

---

## **🧪 测试策略规范**

### **1. 单元测试模式**

#### **模型测试**
```go
func TestOrganizationUnit_Validation(t *testing.T) {
    tests := []struct {
        name    string
        unit    *OrganizationUnit
        wantErr bool
    }{
        {
            name: "valid department",
            unit: &OrganizationUnit{
                UnitType: "DEPARTMENT",
                Name:     "Engineering",
                Profile:  json.RawMessage(`{"head_of_unit_person_id": "123e4567-e89b-12d3-a456-426614174000"}`),
            },
            wantErr: false,
        },
        {
            name: "invalid profile for type",
            unit: &OrganizationUnit{
                UnitType: "DEPARTMENT",
                Profile:  json.RawMessage(`{"invalid_field": "value"}`),
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.unit.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### **服务测试**
```go
func TestOrganizationService_CreateUnit_TransactionalOutbox(t *testing.T) {
    db := setupTestDB(t)
    service := NewOrganizationService(db)
    
    ctx := context.Background()
    req := &CreateOrganizationUnitRequest{
        TenantID: uuid.New(),
        UnitType: "DEPARTMENT",
        Name:     "Test Department",
    }
    
    // 执行创建
    unit, err := service.CreateOrganizationUnit(ctx, req)
    require.NoError(t, err)
    require.NotNil(t, unit)
    
    // 验证组织单元已创建
    found, err := db.OrganizationUnit.Get(ctx, unit.ID)
    require.NoError(t, err)
    assert.Equal(t, req.Name, found.Name)
    
    // 验证事件已发布到发件箱
    events, err := db.OutboxEvent.Query().
        Where(outboxevent.AggregateIDEQ(unit.ID.String())).
        All(ctx)
    require.NoError(t, err)
    assert.Len(t, events, 1)
    assert.Equal(t, "organization.unit.created", events[0].EventType)
}
```

### **2. 集成测试模式**

#### **API集成测试**
```go
func TestOrganizationAPI_CRUD_Flow(t *testing.T) {
    server := setupTestServer(t)
    client := server.Client()
    
    // 创建组织单元
    createReq := CreateOrganizationUnitRequest{
        UnitType: "DEPARTMENT",
        Name:     "Test Department",
        Profile: map[string]interface{}{
            "functional_area": "Engineering",
        },
    }
    
    createResp, err := client.R().
        SetBody(createReq).
        Post("/api/v1/organization-units")
    
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, createResp.StatusCode())
    
    var unit OrganizationUnitResponse
    err = json.Unmarshal(createResp.Body(), &unit)
    require.NoError(t, err)
    
    // 获取组织单元
    getResp, err := client.R().
        Get(fmt.Sprintf("/api/v1/organization-units/%s", unit.ID))
    
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, getResp.StatusCode())
    
    // 验证返回数据
    var fetchedUnit OrganizationUnitResponse
    err = json.Unmarshal(getResp.Body(), &fetchedUnit)
    require.NoError(t, err)
    assert.Equal(t, unit.ID, fetchedUnit.ID)
    assert.Equal(t, unit.Name, fetchedUnit.Name)
}
```

### **3. 图数据库测试**

#### **同步测试**
```go
func TestGraphSyncService_OrganizationUnitCreated(t *testing.T) {
    neo4jContainer := setupNeo4jContainer(t)
    driver := neo4jContainer.Driver()
    service := NewGraphSyncService(driver)
    
    ctx := context.Background()
    event := &OrganizationUnitCreatedEvent{
        UnitID:   uuid.New(),
        TenantID: uuid.New(),
        UnitType: "DEPARTMENT",
        Name:     "Test Department",
    }
    
    // 执行同步
    err := service.ProcessOrganizationUnitCreatedEvent(ctx, event)
    require.NoError(t, err)
    
    // 验证节点已创建
    session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
    defer session.Close()
    
    result, err := session.Run(
        "MATCH (ou:OrgUnit {id: $id, tenant_id: $tenant_id}) RETURN ou.name as name",
        map[string]interface{}{
            "id":        event.UnitID.String(),
            "tenant_id": event.TenantID.String(),
        },
    )
    require.NoError(t, err)
    
    record, err := result.Single()
    require.NoError(t, err)
    
    name, found := record.Get("name")
    require.True(t, found)
    assert.Equal(t, event.Name, name)
}
```

---

## **📈 监控与可观测性**

### **1. 指标收集**

#### **业务指标**
```go
// 业务操作计数器
var (
    orgUnitsCreated = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "organization_units_created_total",
            Help: "Total number of organization units created",
        },
        []string{"tenant_id", "unit_type"},
    )
    
    positionsAssigned = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "positions_assigned_total", 
            Help: "Total number of position assignments",
        },
        []string{"tenant_id", "position_type"},
    )
    
    graphSyncLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "graph_sync_duration_seconds",
            Help:    "Time spent syncing data to graph database",
            Buckets: prometheus.DefBuckets,
        },
        []string{"event_type", "success"},
    )
)

// 在服务中使用
func (s *OrganizationService) CreateOrganizationUnit(ctx context.Context, req *CreateOrganizationUnitRequest) (*OrganizationUnit, error) {
    start := time.Now()
    defer func() {
        orgUnitsCreated.WithLabelValues(req.TenantID.String(), req.UnitType).Inc()
    }()
    
    // ... 业务逻辑
}
```

#### **健康检查**
```go
// 健康检查端点
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
    health := struct {
        Status     string            `json:"status"`
        Components map[string]string `json:"components"`
        Timestamp  time.Time         `json:"timestamp"`
    }{
        Status:     "healthy",
        Components: make(map[string]string),
        Timestamp:  time.Now(),
    }
    
    // 检查PostgreSQL
    if err := s.db.Ping(); err != nil {
        health.Status = "unhealthy"
        health.Components["postgresql"] = "down"
    } else {
        health.Components["postgresql"] = "up"
    }
    
    // 检查Neo4j
    if err := s.neo4j.VerifyConnectivity(); err != nil {
        health.Status = "unhealthy"
        health.Components["neo4j"] = "down"
    } else {
        health.Components["neo4j"] = "up"
    }
    
    statusCode := http.StatusOK
    if health.Status == "unhealthy" {
        statusCode = http.StatusServiceUnavailable
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(health)
}
```

### **2. 日志规范**

#### **结构化日志**
```go
// 使用结构化日志记录关键操作
func (s *OrganizationService) CreateOrganizationUnit(ctx context.Context, req *CreateOrganizationUnitRequest) (*OrganizationUnit, error) {
    logger := s.logger.WithFields(logrus.Fields{
        "operation":  "create_organization_unit",
        "tenant_id":  req.TenantID,
        "unit_type":  req.UnitType,
        "request_id": getRequestID(ctx),
    })
    
    logger.Info("Creating organization unit")
    
    unit, err := s.createUnit(ctx, req)
    if err != nil {
        logger.WithError(err).Error("Failed to create organization unit")
        return nil, err
    }
    
    logger.WithField("unit_id", unit.ID).Info("Organization unit created successfully")
    return unit, nil
}

// 事件处理日志
func (h *GraphSyncHandler) Handle(ctx context.Context, event Event) error {
    logger := h.logger.WithFields(logrus.Fields{
        "handler":    "graph_sync",
        "event_type": event.Type,
        "event_id":   event.ID,
        "tenant_id":  event.TenantID,
    })
    
    logger.Debug("Processing event")
    
    if err := h.processEvent(ctx, event); err != nil {
        logger.WithError(err).Error("Event processing failed")
        return err
    }
    
    logger.Info("Event processed successfully")
    return nil
}
```

---

## **🔧 开发工具与自动化**

### **1. 代码生成工具**

#### **Ent代码生成**
```bash
# 生成Ent代码
go generate ./ent

# 创建迁移
go run -mod=mod entgo.io/ent/cmd/ent migrate diff --dir file://ent/migrate/migrations --to ent://ent/schema --dev-url "postgres://localhost/dev?sslmode=disable"

# 应用迁移  
go run -mod=mod entgo.io/ent/cmd/ent migrate apply --dir file://ent/migrate/migrations --url "postgres://localhost/cube_castle?sslmode=disable"
```

#### **代码质量检查**
```bash
# 静态分析
golangci-lint run

# 测试覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 安全扫描
gosec ./...
```

### **2. CI/CD流水线**

#### **GitHub Actions配置**
```yaml
name: Organization & Position Model CI

on:
  push:
    branches: [ main, develop ]
    paths: 
      - 'go-app/ent/schema/organization_unit.go'
      - 'go-app/ent/schema/position*.go'
      - 'go-app/internal/service/organization_*.go'
      - 'go-app/internal/service/position_*.go'

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: password
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      
      neo4j:
        image: neo4j:5.0
        env:
          NEO4J_AUTH: neo4j/password
        options: >-
          --health-cmd "cypher-shell -u neo4j -p password 'RETURN 1'"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: |
        cd go-app
        go test -v -race -coverprofile=coverage.out ./internal/service/organization_*
        go test -v -race -coverprofile=coverage.out ./internal/service/position_*
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./go-app/coverage.out
```

### **3. 开发环境设置**

#### **Docker Compose开发环境**
```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: cube_castle_dev
      POSTGRES_USER: developer
      POSTGRES_PASSWORD: dev_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_dev_data:/var/lib/postgresql/data

  neo4j:
    image: neo4j:5.0
    environment:
      NEO4J_AUTH: neo4j/dev_password
      NEO4J_PLUGINS: '["apoc"]'
    ports:
      - "7474:7474"
      - "7687:7687"
    volumes:
      - neo4j_dev_data:/data

volumes:
  postgres_dev_data:
  neo4j_dev_data:
```

---

## **📝 总结与最佳实践检查单**

### **设计原则遵循**
- [ ] ✅ 多态性通过鉴别器+JSON插槽实现
- [ ] ✅ 事件驱动架构杜绝直接CRUD
- [ ] ✅ 双重存储策略(PostgreSQL+Neo4j)
- [ ] ✅ 多租户隔离(RLS+应用层)

### **代码质量标准**
- [ ] ✅ 单元测试覆盖率 ≥85%
- [ ] ✅ 静态分析工具通过
- [ ] ✅ 安全扫描无高危漏洞
- [ ] ✅ 性能基准达标

### **架构合规性**
- [ ] ✅ 符合元合约v6.0规范
- [ ] ✅ 城堡模型边界清晰
- [ ] ✅ API设计RESTful标准
- [ ] ✅ 事件溯源模式正确

### **运维就绪性**
- [ ] ✅ 监控指标完整
- [ ] ✅ 健康检查端点
- [ ] ✅ 结构化日志规范
- [ ] ✅ CI/CD流水线配置

---

**下一步**: 将此技术规范作为代码审查和架构评审的标准依据，确保实施过程中严格遵循规范要求。