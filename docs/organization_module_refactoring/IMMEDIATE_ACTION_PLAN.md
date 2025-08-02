# 🚀 CQRS+CDC架构实施 - 立即行动计划

## 📅 执行时间线：立即启动
**创建日期**: 2025年8月2日  
**预期完成**: 6-8周  
**项目代号**: Operation Phoenix (凤凰重生)  

---

## 🎯 第一周 - 基础设施革命

### Day 1-2: 环境准备
```bash
# 1. 扩展docker-compose.yml
echo "开始扩展Kafka生态系统..."

# 2. 立即执行的命令
cd /home/shangmeilin/cube-castle
cp docker-compose.yml docker-compose.backup.yml

# 3. PostgreSQL逻辑复制准备
# 编辑docker-compose.yml中的postgres服务
```

### Day 3-5: Kafka生态系统部署
```yaml
立即添加到docker-compose.yml:

  # Kafka生态系统
  zookeeper:
    image: confluentinc/cp-zookeeper:7.4.0
    hostname: zookeeper
    container_name: cube_castle_zookeeper
    ports:
      - "2181:2181"
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    networks:
      - castle-net
      
  kafka:
    image: confluentinc/cp-kafka:7.4.0
    hostname: kafka
    container_name: cube_castle_kafka
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
      - "9101:9101"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: 'zookeeper:2181'
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:29092,PLAINTEXT_HOST://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
      KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
      KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0
    networks:
      - castle-net
      
  kafka-connect:
    image: debezium/connect:2.4
    hostname: kafka-connect
    container_name: cube_castle_kafka_connect
    depends_on:
      - kafka
    ports:
      - 8083:8083
    environment:
      BOOTSTRAP_SERVERS: 'kafka:29092'
      REST_ADVERTISED_HOST_NAME: kafka-connect
      GROUP_ID: compose-connect-group
      CONFIG_STORAGE_TOPIC: docker-connect-configs
      OFFSET_STORAGE_TOPIC: docker-connect-offsets
      STATUS_STORAGE_TOPIC: docker-connect-status
    networks:
      - castle-net
      
  kafka-ui:
    image: provectuslabs/kafka-ui:latest
    container_name: cube_castle_kafka_ui
    depends_on:
      - kafka
    ports:
      - "8080:8080"
    environment:
      KAFKA_CLUSTERS_0_NAME: local
      KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS: kafka:29092
      KAFKA_CLUSTERS_0_KAFKACONNECT_0_NAME: first
      KAFKA_CLUSTERS_0_KAFKACONNECT_0_ADDRESS: http://kafka-connect:8083
    networks:
      - castle-net
```

### Day 6-7: CDC管道验证
```bash
#!/bin/bash
# scripts/setup-cdc-pipeline.sh

echo "🚀 设置CDC数据管道..."

# 1. 启动完整技术栈
docker-compose up -d

# 2. 等待服务启动
echo "等待Kafka Connect启动..."
while ! curl -f http://localhost:8083/; do
  sleep 5
  echo "Kafka Connect还未启动，继续等待..."
done

# 3. 配置PostgreSQL复制用户
docker exec cube_castle_postgres psql -U user -d cubecastle -c "
CREATE USER debezium_user WITH REPLICATION LOGIN PASSWORD 'debezium_pass';
GRANT SELECT ON ALL TABLES IN SCHEMA public TO debezium_user;
CREATE PUBLICATION organization_publication FOR TABLE 
  employees, organization_units, positions;
"

# 4. 创建Debezium连接器
curl -X POST http://localhost:8083/connectors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "organization-postgres-connector",
    "config": {
      "connector.class": "io.debezium.connector.postgresql.PostgreSqlConnector",
      "database.hostname": "postgres",
      "database.port": "5432",
      "database.user": "debezium_user",
      "database.password": "debezium_pass",
      "database.dbname": "cubecastle",
      "database.server.name": "organization_db",
      "table.include.list": "public.employees,public.organization_units,public.positions",
      "publication.name": "organization_publication",
      "plugin.name": "pgoutput",
      "slot.name": "organization_slot"
    }
  }'

echo "✅ CDC管道配置完成！访问 http://localhost:8080 查看Kafka UI"
```

---

## 🏗️ 第二周 - CQRS架构重构

### 新的项目结构
```
go-app/internal/
├── cqrs/
│   ├── commands/
│   │   ├── hire_employee.go
│   │   ├── create_org_unit.go
│   │   ├── update_employee.go
│   │   └── command_bus.go
│   ├── queries/
│   │   ├── get_org_chart.go
│   │   ├── find_employee.go
│   │   ├── get_reporting_hierarchy.go
│   │   └── query_bus.go
│   ├── events/
│   │   ├── employee_events.go
│   │   ├── org_unit_events.go
│   │   └── event_bus.go
│   └── handlers/
│       ├── command_handlers.go
│       └── query_handlers.go
├── repositories/
│   ├── postgres_command_repo.go
│   └── neo4j_query_repo.go
└── routes/
    └── cqrs_routes.go
```

### 立即创建的核心文件

#### 1. 命令定义
```go
// internal/cqrs/commands/employee_commands.go
package commands

import (
    "time"
    "github.com/google/uuid"
)

type HireEmployeeCommand struct {
    TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
    FirstName    string    `json:"first_name" validate:"required,min=1,max=100"`
    LastName     string    `json:"last_name" validate:"required,min=1,max=100"`
    Email        string    `json:"email" validate:"required,email"`
    PositionID   *uuid.UUID `json:"position_id,omitempty"`
    HireDate     time.Time `json:"hire_date" validate:"required"`
    EmployeeType string    `json:"employee_type" validate:"required,oneof=FULL_TIME PART_TIME CONTRACTOR INTERN"`
}

type UpdateEmployeeCommand struct {
    ID        uuid.UUID `json:"id" validate:"required"`
    TenantID  uuid.UUID `json:"tenant_id" validate:"required"`
    FirstName *string   `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
    LastName  *string   `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
    Email     *string   `json:"email,omitempty" validate:"omitempty,email"`
}

type CreateOrganizationUnitCommand struct {
    TenantID     uuid.UUID              `json:"tenant_id" validate:"required"`
    UnitType     string                 `json:"unit_type" validate:"required,oneof=DEPARTMENT COST_CENTER COMPANY PROJECT_TEAM"`
    Name         string                 `json:"name" validate:"required,min=1,max=100"`
    Description  *string                `json:"description,omitempty"`
    ParentUnitID *uuid.UUID             `json:"parent_unit_id,omitempty"`
    Profile      map[string]interface{} `json:"profile,omitempty"`
}
```

#### 2. 查询定义
```go
// internal/cqrs/queries/organization_queries.go
package queries

import (
    "github.com/google/uuid"
)

type GetOrgChartQuery struct {
    TenantID        uuid.UUID  `json:"tenant_id" validate:"required"`
    RootUnitID      *uuid.UUID `json:"root_unit_id,omitempty"`
    MaxDepth        int        `json:"max_depth" validate:"min=1,max=10"`
    IncludeInactive bool       `json:"include_inactive"`
}

type FindEmployeeQuery struct {
    TenantID uuid.UUID `json:"tenant_id" validate:"required"`
    ID       uuid.UUID `json:"id" validate:"required"`
}

type SearchEmployeesQuery struct {
    TenantID   uuid.UUID `json:"tenant_id" validate:"required"`
    Name       *string   `json:"name,omitempty"`
    Email      *string   `json:"email,omitempty"`
    Department *string   `json:"department,omitempty"`
    Limit      int       `json:"limit" validate:"min=1,max=1000"`
    Offset     int       `json:"offset" validate:"min=0"`
}

type GetReportingHierarchyQuery struct {
    TenantID  uuid.UUID `json:"tenant_id" validate:"required"`
    ManagerID uuid.UUID `json:"manager_id" validate:"required"`
    MaxDepth  int       `json:"max_depth" validate:"min=1,max=10"`
}
```

#### 3. 新的路由结构
```go
// internal/routes/cqrs_routes.go
package routes

import (
    "github.com/go-chi/chi/v5"
    "github.com/gaogu/cube-castle/go-app/internal/cqrs/handlers"
)

func SetupCQRSRoutes(r chi.Router, cmdHandler *handlers.CommandHandler, queryHandler *handlers.QueryHandler) {
    // 命令端点 - 所有写操作
    r.Route("/commands", func(r chi.Router) {
        // 员工管理命令
        r.Post("/hire-employee", cmdHandler.HireEmployee)
        r.Put("/update-employee", cmdHandler.UpdateEmployee)
        r.Post("/terminate-employee", cmdHandler.TerminateEmployee)
        
        // 组织单元管理命令
        r.Post("/create-organization-unit", cmdHandler.CreateOrganizationUnit)
        r.Put("/update-organization-unit", cmdHandler.UpdateOrganizationUnit)
        r.Delete("/delete-organization-unit", cmdHandler.DeleteOrganizationUnit)
        
        // 职位管理命令
        r.Post("/assign-employee-position", cmdHandler.AssignEmployeePosition)
        r.Post("/create-position", cmdHandler.CreatePosition)
    })
    
    // 查询端点 - 所有读操作  
    r.Route("/queries", func(r chi.Router) {
        // 员工查询
        r.Get("/employees/{id}", queryHandler.GetEmployee)
        r.Get("/employees", queryHandler.SearchEmployees)
        
        // 组织结构查询
        r.Get("/organization-chart", queryHandler.GetOrgChart)
        r.Get("/organization-units/{id}", queryHandler.GetOrganizationUnit)
        r.Get("/organization-units", queryHandler.ListOrganizationUnits)
        
        // 层级关系查询
        r.Get("/reporting-hierarchy/{manager_id}", queryHandler.GetReportingHierarchy)
        r.Get("/employee-path/{from_id}/{to_id}", queryHandler.FindEmployeePath)
        
        // 高级查询
        r.Get("/department-structure/{dept_id}", queryHandler.GetDepartmentStructure)
        r.Get("/common-manager", queryHandler.FindCommonManager)
    })
}
```

---

## 🎯 第三周 - 事件驱动架构

### 事件定义
```go
// internal/cqrs/events/employee_events.go
package events

import (
    "time"
    "github.com/google/uuid"
)

type EmployeeHired struct {
    EventID    uuid.UUID `json:"event_id"`
    EmployeeID uuid.UUID `json:"employee_id"`
    TenantID   uuid.UUID `json:"tenant_id"`
    FirstName  string    `json:"first_name"`
    LastName   string    `json:"last_name"`
    Email      string    `json:"email"`
    HireDate   time.Time `json:"hire_date"`
    Timestamp  time.Time `json:"timestamp"`
}

type EmployeeUpdated struct {
    EventID    uuid.UUID              `json:"event_id"`
    EmployeeID uuid.UUID              `json:"employee_id"`
    TenantID   uuid.UUID              `json:"tenant_id"`
    Changes    map[string]interface{} `json:"changes"`
    Timestamp  time.Time              `json:"timestamp"`
}

type OrganizationUnitCreated struct {
    EventID  uuid.UUID `json:"event_id"`
    UnitID   uuid.UUID `json:"unit_id"`
    TenantID uuid.UUID `json:"tenant_id"`
    UnitType string    `json:"unit_type"`
    Name     string    `json:"name"`
    Timestamp time.Time `json:"timestamp"`
}
```

---

## 📊 第四周 - 监控与元合约

### 元合约实现
```yaml
# contracts/organization_module_contract.yaml
metadata:
  name: "OrganizationModuleCQRS"
  version: "v1.0.0"
  owner: "cube-castle-team"
  created_at: "2025-08-02"

parties:
  provider: "OrganizationModule"
  consumer: "ClientApplications"

commandModelContract:
  assumptions:
    - "认证用户提供有效的JWT token"
    - "租户隔离边界已通过中间件建立"
    - "命令负载通过JSON Schema验证"
  
  guarantees:
    - "所有命令在PostgreSQL事务中处理"
    - "成功的命令产生相应的领域事件"
    - "命令处理延迟 P99 < 100ms"
    - "数据完整性通过数据库约束保证"
  
  commands:
    - name: "HireEmployee"
      endpoint: "POST /commands/hire-employee"
      schema: "./schemas/hire_employee_command.json"
      postconditions:
        - "员工记录在PostgreSQL中创建"
        - "EmployeeHired事件发布到事件总线"
        - "员工状态设置为PENDING_START"
    
    - name: "CreateOrganizationUnit"
      endpoint: "POST /commands/create-organization-unit"
      schema: "./schemas/create_org_unit_command.json"
      postconditions:
        - "组织单元记录在PostgreSQL中创建"
        - "OrganizationUnitCreated事件发布"
        - "层级关系正确建立"

queryModelContract:
  assumptions:
    - "查询的资源在Neo4j中存在"
    - "租户隔离通过查询参数控制"
  
  guarantees:
    - "查询从Neo4j只读副本执行"
    - "查询响应延迟 P99 < 500ms"
    - "数据最终一致性延迟 P99 < 1000ms"
    - "查询结果缓存5分钟"
  
  queries:
    - name: "GetOrgChart"
      endpoint: "GET /queries/organization-chart"
      schema: "./schemas/org_chart_response.json"
      slo:
        response_time_p99: "500ms"
        availability: "99.9%"
    
    - name: "FindEmployee"
      endpoint: "GET /queries/employees/{id}"
      schema: "./schemas/employee_response.json"
      slo:
        response_time_p99: "200ms"
        cache_hit_rate: ">80%"

dataConsistencyContract:
  cdc_pipeline:
    source: "PostgreSQL WAL"
    sink: "Neo4j Graph Database"
    latency_slo: "P99 < 1000ms"
    reliability: "99.9% uptime"
  
  monitoring:
    metrics:
      - "command_processing_latency"
      - "query_response_time"
      - "cdc_end_to_end_latency"
      - "event_bus_throughput"
    
    alerts:
      - condition: "cdc_latency > 5000ms"
        severity: "critical"
      - condition: "command_error_rate > 1%"
        severity: "warning"
```

---

## 🚀 立即执行清单

### 今天就开始 (Day 1)
- [ ] 备份当前docker-compose.yml
- [ ] 扩展Docker Compose配置（添加Kafka生态系统）
- [ ] 更新PostgreSQL配置支持逻辑复制
- [ ] 启动新的技术栈验证

### 本周完成 (Week 1)
- [ ] CDC数据管道完全工作
- [ ] Kafka UI可以查看数据流
- [ ] Neo4j接收PostgreSQL变更数据
- [ ] 基础监控指标收集

### 两周内完成 (Week 2)
- [ ] 完整CQRS项目结构
- [ ] 命令和查询处理器实现
- [ ] 新API路由全面运行
- [ ] 事件发布机制工作

### 月内完成 (Week 4)
- [ ] 元合约体系实施
- [ ] 性能指标达到目标
- [ ] 完整的文档和运维手册
- [ ] 团队技能达到预期水平

---

## 🎯 成功验证标准

### 技术指标
```yaml
Week 1 目标:
  - PostgreSQL → Kafka → Neo4j 数据流畅通
  - CDC延迟 < 1秒 (开发环境)
  - 零数据丢失

Week 2 目标:
  - 命令处理延迟 < 100ms
  - 查询响应延迟 < 200ms
  - CQRS架构完全分离

Week 4 目标:
  - P99查询延迟 < 500ms
  - CDC端到端延迟 < 1秒
  - 测试覆盖率 > 95%
```

### 团队能力指标
```yaml
Week 2:
  - 100%团队成员理解CQRS基础概念
  - 80%成员能独立开发命令处理器

Week 4:
  - 60%成员掌握Kafka基础运维
  - 40%成员具备Neo4j优化能力
```

---

## 🛡️ 风险应对预案

### 技术风险
- **CDC配置问题**: 准备Manual Sync作为备选
- **Kafka学习曲线**: 外部专家支持 + 结对编程
- **性能不达标**: 分阶段优化 + 缓存策略

### 时间风险
- **里程碑延期**: 每周评估 + 灵活调整
- **学习进度慢**: 增加培训时间 + 简化MVP

---

## 🎉 项目口号

**"Phoenix Rising - 在开发初期的黄金窗口，一次性建立正确的架构！"**

**现在就开始 - 这是千载难逢的机会！** 🚀

---

**行动计划版本**: v1.0  
**创建日期**: 2025年8月2日  
**责任人**: 全体开发团队  
**口号**: "心动不如行动，立即开始Phoenix项目！"