# **组织与岗位模型统一架构设计方案**

**文档类型**: 架构设计方案  
**创建时间**: 2025-07-29  
**版本**: v1.0  
**状态**: 设计完成，待实施  
**负责模块**: core-hr.keep  
**预计实施周期**: 3-4周  
**优先级**: 🔴 最高优先级

## **第一部分：架构哲学与设计原则**

### **1.1 统一框架核心理念**

基于**城堡蓝图**和**元合约v6.0**，组织与岗位模型设计遵循以下核心原则：

- **设计驱动的多态性** → "核心模型 + 嵌套多态档案"统一实现
- **事件驱动的时态性** → 事件溯源架构，杜绝直接CRUD操作  
- **声明式混合持久化** → 关系型数据库(PostgreSQL) + 图数据库(Neo4j)双重存储
- **安全作为基石** → 多租户隔离 + 行级安全 + OPA策略引擎

### **1.2 现有框架基础分析**

**已有Employee模型**: 
- 基础字段: `id`, `name`, `email`, `position`
- 时间戳: `created_at`, `updated_at`
- **问题**: 缺乏多态性支持、租户隔离、事件驱动机制

**已有PositionHistory模型**:
- 关键字段: `employee_id`, `organization_id`, `position_title`, `department`
- 时态支持: `effective_date`, `end_date`, `is_active`
- **优势**: 已具备基础时态特征

---

## **第二部分：组织模型架构设计**

### **2.1 OrganizationUnit核心实体**

#### **Ent Schema定义**
```go
// ent/schema/organization_unit.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
    "github.com/google/uuid"
    "time"
)

type OrganizationUnit struct {
    ent.Schema
}

func (OrganizationUnit) Fields() []ent.Field {
    return []ent.Field{
        // 元合约v6.0核心身份
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            Immutable(),
            
        field.UUID("tenant_id", uuid.UUID{}).
            Immutable().
            Comment("多租户隔离基石"),
            
        field.Enum("unit_type").
            Values("DEPARTMENT", "COST_CENTER", "COMPANY", "PROJECT_TEAM").
            Comment("多态性鉴别器"),
            
        field.String("name").
            NotEmpty().
            Comment("组织单元人类可读名称"),
            
        field.String("description").
            Optional().
            Nillable(),
            
        field.UUID("parent_unit_id", uuid.UUID{}).
            Optional().
            Nillable().
            Comment("层级结构自引用"),
            
        field.Enum("status").
            Values("ACTIVE", "INACTIVE", "PLANNED").
            Default("ACTIVE"),
            
        // 多态档案插槽
        field.JSON("profile", map[string]interface{}{}).
            Optional().
            Comment("基于unit_type的多态配置"),
            
        // 审计字段
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (OrganizationUnit) Edges() []ent.Edge {
    return []ent.Edge{
        // 层级关系
        edge.To("children", OrganizationUnit.Type).
            From("parent").
            Field("parent_unit_id").
            Unique(),
            
        // 包含关系
        edge.To("positions", Position.Type),
    }
}
```

#### **多态档案定义**

**DepartmentProfile**:
```go
type DepartmentProfile struct {
    HeadOfUnitPersonID *uuid.UUID `json:"head_of_unit_person_id,omitempty"`
    FunctionalArea     string     `json:"functional_area,omitempty"`
    CostCenter         string     `json:"cost_center,omitempty"`
}
```

**CostCenterProfile**:
```go
type CostCenterProfile struct {
    CostCenterCode    string     `json:"cost_center_code"`
    FinancialOwnerID  *uuid.UUID `json:"financial_owner_id"`
    BudgetAllocation  *float64   `json:"budget_allocation,omitempty"`
}
```

### **2.2 元合约规约定义**

```yaml
# core-hr.keep/organization_unit.metacontract.yaml
specification_version: "v6.0"
api_id: "org-unit-api-001"
namespace: "core-hr.keep"
resource_name: "organization_units"
version: "1.0.0"

resource_type: ENTITY
abstraction_level: DATA_ORIENTED
audience: INTERNAL_APP

# 多态性定义
polymorphism:
  discriminator_property: "unit_type"
  mapping:
    DEPARTMENT: DepartmentProfile
    COST_CENTER: CostCenterProfile
    COMPANY: CompanyProfile
    PROJECT_TEAM: ProjectTeamProfile

# 混合持久化配置
persistence_profile:
  primary_store: RELATIONAL
  indexed_in: [GRAPH]
  graph_node_label: "OrgUnit"
  graph_edge_definitions:
    - rel_name: "PART_OF"
      direction: OUTGOING
      target_resource: "organization_units"
    - rel_name: "CONTAINS_POSITION"  
      direction: OUTGOING
      target_resource: "positions"
```

---

## **第三部分：岗位模型架构设计**

### **3.1 Position核心实体**

#### **Ent Schema定义**
```go
// ent/schema/position.go
package schema

type Position struct {
    ent.Schema
}

func (Position) Fields() []ent.Field {
    return []ent.Field{
        // 元合约v6.0核心身份
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            Immutable(),
            
        field.UUID("tenant_id", uuid.UUID{}).
            Immutable(),
            
        field.Enum("position_type").
            Values("FULL_TIME", "PART_TIME", "CONTINGENT_WORKER", "INTERN").
            Comment("岗位类型鉴别器"),
            
        field.UUID("job_profile_id", uuid.UUID{}).
            Comment("职位说明书引用"),
            
        field.UUID("department_id", uuid.UUID{}).
            Comment("所属组织单元"),
            
        field.UUID("manager_position_id", uuid.UUID{}).
            Optional().
            Nillable().
            Comment("汇报关系自引用"),
            
        field.Enum("status").
            Values("OPEN", "FILLED", "FROZEN", "PENDING_ELIMINATION").
            Default("OPEN"),
            
        field.Float("budgeted_fte").
            Default(1.0).
            Comment("预算FTE"),
            
        // 多态详情插槽
        field.JSON("details", map[string]interface{}{}).
            Optional().
            Comment("基于position_type的多态配置"),
            
        // 审计字段
        field.Time("created_at").
            Default(time.Now).
            Immutable(),
        field.Time("updated_at").
            Default(time.Now).
            UpdateDefault(time.Now),
    }
}

func (Position) Edges() []ent.Edge {
    return []ent.Edge{
        // 汇报关系
        edge.To("direct_reports", Position.Type).
            From("manager").
            Field("manager_position_id").
            Unique(),
            
        // 组织关系
        edge.From("department", OrganizationUnit.Type).
            Field("department_id").
            Ref("positions").
            Unique(),
            
        // 占据关系
        edge.To("occupancy_history", PositionOccupancyHistory.Type),
    }
}
```

### **3.2 历史记录分离设计**

#### **PositionAttributeHistory** - 岗位属性历史
```go
// ent/schema/position_attribute_history.go
type PositionAttributeHistory struct {
    ent.Schema
}

func (PositionAttributeHistory) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("tenant_id", uuid.UUID{}),
        field.UUID("position_id", uuid.UUID{}),
        
        // 快照属性
        field.String("position_type"),
        field.UUID("job_profile_id", uuid.UUID{}),
        field.UUID("department_id", uuid.UUID{}),
        field.UUID("manager_position_id", uuid.UUID{}).Optional().Nillable(),
        field.String("status"),
        field.Float("budgeted_fte"),
        field.JSON("details", map[string]interface{}{}),
        
        // 时态字段
        field.Time("effective_date"),
        field.Time("end_date").Optional().Nillable(),
        field.String("change_reason").Optional(),
        field.UUID("changed_by", uuid.UUID{}),
        
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}
```

#### **PositionOccupancyHistory** - 岗位占据历史
```go
// ent/schema/position_occupancy_history.go  
type PositionOccupancyHistory struct {
    ent.Schema
}

func (PositionOccupancyHistory) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("tenant_id", uuid.UUID{}),
        field.UUID("position_id", uuid.UUID{}),
        field.UUID("employee_id", uuid.UUID{}),
        
        // 占据时间范围
        field.Time("start_date"),
        field.Time("end_date").Optional().Nillable(),
        field.Bool("is_active").Default(true),
        
        // 任命信息
        field.String("assignment_type").Default("REGULAR"), 
        field.String("assignment_reason").Optional(),
        field.UUID("approved_by", uuid.UUID{}).Optional(),
        
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}
```

---

## **第四部分：事件驱动架构集成**

### **4.1 业务事件定义**

#### **组织单元事件**
```go
// internal/events/organization_events.go
type OrganizationUnitCreatedEvent struct {
    EventID      uuid.UUID    `json:"event_id"`
    TenantID     uuid.UUID    `json:"tenant_id"`
    UnitID       uuid.UUID    `json:"unit_id"`
    UnitType     string       `json:"unit_type"`
    Name         string       `json:"name"`
    ParentUnitID *uuid.UUID   `json:"parent_unit_id,omitempty"`
    Profile      interface{}  `json:"profile"`
    EffectiveDate time.Time   `json:"effective_date"`
    CreatedBy    uuid.UUID    `json:"created_by"`
    CreatedAt    time.Time    `json:"created_at"`
}

type OrganizationUnitRestructuredEvent struct {
    EventID       uuid.UUID    `json:"event_id"`
    TenantID      uuid.UUID    `json:"tenant_id"`
    UnitID        uuid.UUID    `json:"unit_id"`
    OldParentID   *uuid.UUID   `json:"old_parent_id"`
    NewParentID   *uuid.UUID   `json:"new_parent_id"`
    ReasonCode    string       `json:"reason_code"`
    EffectiveDate time.Time    `json:"effective_date"`
    ApprovedBy    uuid.UUID    `json:"approved_by"`
    CreatedAt     time.Time    `json:"created_at"`
}
```

#### **岗位事件**
```go
type PositionCreatedEvent struct {
    EventID        uuid.UUID    `json:"event_id"`
    TenantID       uuid.UUID    `json:"tenant_id"`
    PositionID     uuid.UUID    `json:"position_id"`
    PositionType   string       `json:"position_type"`
    JobProfileID   uuid.UUID    `json:"job_profile_id"`
    DepartmentID   uuid.UUID    `json:"department_id"`
    ManagerID      *uuid.UUID   `json:"manager_position_id,omitempty"`
    BudgetedFTE    float64      `json:"budgeted_fte"`
    Details        interface{}  `json:"details"`
    EffectiveDate  time.Time    `json:"effective_date"`
    CreatedBy      uuid.UUID    `json:"created_by"`
    CreatedAt      time.Time    `json:"created_at"`
}

type PositionAssignmentEvent struct {
    EventID      uuid.UUID  `json:"event_id"`
    TenantID     uuid.UUID  `json:"tenant_id"`
    PositionID   uuid.UUID  `json:"position_id"`
    EmployeeID   uuid.UUID  `json:"employee_id"`
    StartDate    time.Time  `json:"start_date"`
    EndDate      *time.Time `json:"end_date,omitempty"`
    AssignmentType string   `json:"assignment_type"`
    ApprovedBy   uuid.UUID  `json:"approved_by"`
    CreatedAt    time.Time  `json:"created_at"`
}
```

### **4.2 事务性发件箱实现**

```go
// internal/service/organization_event_service.go
type OrganizationEventService struct {
    db     *ent.Client
    logger *logger.Logger
}

func (s *OrganizationEventService) CreateOrganizationUnit(
    ctx context.Context, 
    req *CreateOrganizationUnitRequest,
) (*OrganizationUnit, error) {
    return s.db.WithTx(ctx, func(tx *ent.Tx) (*OrganizationUnit, error) {
        // 1. 创建组织单元记录
        orgUnit, err := tx.OrganizationUnit.
            Create().
            SetTenantID(req.TenantID).
            SetUnitType(organizationunit.UnitType(req.UnitType)).
            SetName(req.Name).
            SetNillableParentUnitID(req.ParentUnitID).
            SetProfile(req.Profile).
            SetStatus(organizationunit.StatusACTIVE).
            Save(ctx)
        if err != nil {
            return nil, err
        }

        // 2. 创建业务事件（事务性发件箱）
        event := &OrganizationUnitCreatedEvent{
            EventID:       uuid.New(),
            TenantID:      req.TenantID,
            UnitID:        orgUnit.ID,
            UnitType:      req.UnitType,
            Name:          req.Name,
            ParentUnitID:  req.ParentUnitID,
            Profile:       req.Profile,
            EffectiveDate: time.Now(),
            CreatedBy:     req.CreatedBy,
            CreatedAt:     time.Now(),
        }

        // 3. 持久化事件到发件箱表
        _, err = tx.OutboxEvent.
            Create().
            SetEventType("organization.unit.created").
            SetAggregateID(orgUnit.ID.String()).
            SetEventData(event).
            SetTenantID(req.TenantID).
            Save(ctx)
        if err != nil {
            return nil, err
        }

        return orgUnit, nil
    })
}
```

---

## **第五部分：图数据库集成策略**

### **5.1 Neo4j节点与关系映射**

#### **节点定义**
```cypher
-- 组织单元节点
CREATE CONSTRAINT org_unit_id IF NOT EXISTS 
FOR (ou:OrgUnit) REQUIRE ou.id IS UNIQUE;

-- 岗位节点  
CREATE CONSTRAINT position_id IF NOT EXISTS
FOR (p:Position) REQUIRE p.id IS UNIQUE;

-- 员工节点
CREATE CONSTRAINT employee_id IF NOT EXISTS  
FOR (e:Employee) REQUIRE e.id IS UNIQUE;
```

#### **关系定义**
```cypher
-- 组织层级关系
(child:OrgUnit)-[:PART_OF]->(parent:OrgUnit)

-- 岗位包含关系
(dept:OrgUnit)-[:CONTAINS_POSITION]->(pos:Position)

-- 汇报关系
(subordinate:Position)-[:REPORTS_TO]->(manager:Position)

-- 岗位占据关系（带时间属性）
(emp:Employee)-[:OCCUPIES {start_date, end_date}]->(pos:Position)
```

### **5.2 同步服务实现**

```go
// internal/service/graph_sync_service.go
type GraphSyncService struct {
    neo4jDriver neo4j.Driver
    logger      *logger.Logger
}

func (s *GraphSyncService) ProcessOrganizationUnitCreatedEvent(
    ctx context.Context,
    event *OrganizationUnitCreatedEvent,
) error {
    session := s.neo4jDriver.NewSession(neo4j.SessionConfig{
        AccessMode: neo4j.AccessModeWrite,
    })
    defer session.Close()

    _, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
        // 创建组织单元节点
        createQuery := `
        CREATE (ou:OrgUnit {
            id: $id,
            tenant_id: $tenant_id,
            unit_type: $unit_type,
            name: $name,
            status: 'ACTIVE',
            created_at: $created_at
        })
        `
        
        _, err := tx.Run(createQuery, map[string]interface{}{
            "id":         event.UnitID.String(),
            "tenant_id":  event.TenantID.String(),
            "unit_type":  event.UnitType,
            "name":       event.Name,
            "created_at": event.CreatedAt.Format(time.RFC3339),
        })
        if err != nil {
            return nil, err
        }

        // 如果有父级，创建层级关系
        if event.ParentUnitID != nil {
            relationQuery := `
            MATCH (child:OrgUnit {id: $child_id})
            MATCH (parent:OrgUnit {id: $parent_id})
            WHERE child.tenant_id = $tenant_id AND parent.tenant_id = $tenant_id
            CREATE (child)-[:PART_OF]->(parent)
            `
            
            _, err = tx.Run(relationQuery, map[string]interface{}{
                "child_id":  event.UnitID.String(),
                "parent_id": event.ParentUnitID.String(),
                "tenant_id": event.TenantID.String(),
            })
        }

        return nil, err
    })

    return err
}
```

---

## **第六部分：架构符合性检查单**

### **6.1 元合约v6.0符合性审计**

| 元合约v6.0规约 | OrganizationUnit符合性 | Position符合性 |
|--------------|---------------------|----------------|
| **核心身份** | ✅ UUID主键 + 租户隔离 | ✅ UUID主键 + 租户隔离 |
| **多态性** | ✅ unit_type鉴别器 + profile插槽 | ✅ position_type鉴别器 + details插槽 |
| **时态性** | ✅ 事件驱动 + 历史表 | ✅ 双重历史分离设计 |
| **持久化** | ✅ PostgreSQL + Neo4j混合 | ✅ PostgreSQL + Neo4j混合 |
| **安全性** | ✅ RLS + OPA + 租户边界 | ✅ RLS + OPA + 租户边界 |

### **6.2 关键技术决策**

- **数据模型统一性**: 完全对齐现有Employee模型架构模式
- **事件驱动一致性**: 与Temporal工作流无缝集成  
- **图数据库战略**: Neo4j作为"洞察系统"的权威实现
- **多租户安全**: 数据库RLS + 应用层OPA双重保障

---

## **第七部分：风险评估与缓解策略**

### **7.1 主要技术风险**

| 风险类别 | 风险描述 | 影响程度 | 缓解策略 |
|---------|---------|---------|---------|
| **数据迁移** | 现有PositionHistory与新设计不兼容 | 🔴 高 | 渐进式迁移 + 数据校验脚本 |
| **性能影响** | 图数据库同步延迟 | 🟡 中 | 异步处理 + 监控告警 |
| **架构复杂性** | 事件驱动增加调试难度 | 🟡 中 | 完善日志 + 调试工具 |
| **多态验证** | JSON字段类型安全性 | 🟢 低 | Schema验证 + 单元测试 |

### **7.2 实施前置条件**

✅ **技术前置**:
- Neo4j实例部署完成
- Ent框架升级到最新版本
- 事务性发件箱基础设施就绪

✅ **团队前置**:
- 图数据库操作培训
- 事件驱动架构理解
- 多态设计模式掌握

---

## **第八部分：后续文档规划**

### **8.1 待创建文档**

1. **实施路线图TODO清单** (`docs/architecture/organization_position_implementation_roadmap.md`)
2. **API接口规范** (`docs/api/organization_position_api_spec.md`)
3. **数据迁移指南** (`docs/deployment/data_migration_guide.md`)
4. **Neo4j集成手册** (`docs/troubleshooting/neo4j_integration_guide.md`)

### **8.2 相关文档引用**

- 📋 [城堡蓝图](castle_blueprint.md) - 架构哲学基础
- 📋 [元合约v6.0](../api/metacontract_v6.0_specification.md) - 技术规约依据
- 📋 [员工模型设计规划](employee_model_design_development_plan.md) - 统一框架参考
- 📋 [TODO实现计划](todo_implementation_plan.md) - 当前开发状态

---

**预期成果**: 高度内聚的组织与岗位管理体系，为HR SaaS平台提供坚实的核心域基础，支持未来向微服务架构的平滑演进。

**下一步行动**: 创建详细的实施路线图TODO清单，开始Phase 1基础架构开发。