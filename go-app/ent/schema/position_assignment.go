package schema

import (
    "time"
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
)

// PositionAssignment 职位分配 - 核心关系简化
type PositionAssignment struct {
    ent.Schema
}

func (PositionAssignment) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("tenant_id", uuid.UUID{}),
        field.UUID("position_id", uuid.UUID{}),
        field.UUID("employee_id", uuid.UUID{}),
        
        // 核心时间信息
        field.Time("start_date"),
        field.Time("end_date").Optional().Nillable(),
        field.Bool("is_current").Default(false),
        
        // 简化的分配信息
        field.Float("fte").Default(1.0),
        field.Enum("assignment_type").Values("PRIMARY", "SECONDARY", "ACTING").Default("PRIMARY"),
        
        // 审计信息
        field.Time("created_at").Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (PositionAssignment) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("position", Position.Type).
            Ref("assignments").
            Field("position_id").
            Required().
            Unique(),
            
        edge.From("employee", Employee.Type).
            Ref("assignments").
            Field("employee_id").
            Required().
            Unique(),

        // Assignment Details
        edge.To("details", AssignmentDetails.Type).
            Comment("Detailed assignment information"),

        // Assignment History  
        edge.To("history", AssignmentHistory.Type).
            Comment("Assignment change history"),
    }
}

func (PositionAssignment) Indexes() []ent.Index {
    return []ent.Index{
        // 核心查询索引
        index.Fields("tenant_id", "employee_id", "is_current"),
        index.Fields("tenant_id", "position_id", "is_current"), 
        
        // 唯一性约束 - 每个员工同时只能有一个当前主要职位
        index.Fields("tenant_id", "employee_id", "is_current", "assignment_type").
            Unique(),
    }
}

// AssignmentDetails 分配详情 - 复杂业务信息分离
type AssignmentDetails struct {
    ent.Schema
}

func (AssignmentDetails) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("assignment_id", uuid.UUID{}), // 关联到PositionAssignment
        field.UUID("tenant_id", uuid.UUID{}),
        
        // 复杂业务信息
        field.UUID("pay_grade_id", uuid.UUID{}).Optional(),
        field.UUID("compensation_plan_id", uuid.UUID{}).Optional(),
        field.String("work_location").MaxLen(200).Optional(),
        field.Enum("work_arrangement").Values("ON_SITE", "REMOTE", "HYBRID").Optional(),
        
        // 变更信息
        field.Enum("assignment_reason").Values(
            "NEW_HIRE", "PROMOTION", "TRANSFER", "DEMOTION", 
            "LATERAL_MOVE", "TEMPORARY_ASSIGNMENT", "RETURN_FROM_LEAVE",
        ).Optional(),
        field.String("change_reason").MaxLen(500).Optional(),
        field.Text("notes").Optional(),
        
        // 审批信息
        field.UUID("approved_by", uuid.UUID{}).Optional(),
        field.Time("approved_at").Optional(),
        field.Enum("approval_status").Values("PENDING", "APPROVED", "REJECTED").Default("APPROVED"),
        
        // 扩展字段
        field.JSON("custom_fields", map[string]interface{}{}).Optional(),
        
        field.Time("created_at").Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (AssignmentDetails) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("assignment", PositionAssignment.Type).
            Ref("details").
            Field("assignment_id").
            Required().
            Unique(),
    }
}

// AssignmentHistory 分配历史事件 - 审计追踪分离
type AssignmentHistory struct {
    ent.Schema
}

func (AssignmentHistory) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.UUID("assignment_id", uuid.UUID{}),
        field.UUID("tenant_id", uuid.UUID{}),
        
        // 历史事件信息
        field.Enum("event_type").Values("CREATED", "UPDATED", "ENDED", "EXTENDED"),
        field.Time("event_date"),
        field.JSON("previous_values", map[string]interface{}{}).Optional(),
        field.JSON("new_values", map[string]interface{}{}).Optional(),
        field.String("reason").MaxLen(500).Optional(),
        field.UUID("changed_by", uuid.UUID{}).Optional(),
        
        field.Time("created_at").Default(time.Now),
    }
}

func (AssignmentHistory) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("assignment", PositionAssignment.Type).
            Ref("history").
            Field("assignment_id").
            Required().
            Unique(),
    }
}