package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "time"
)

// PositionHistory holds the schema definition for the PositionHistory entity.
type PositionHistory struct {
    ent.Schema
}

// Fields of the PositionHistory.
func (PositionHistory) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").Unique(),
        field.String("employee_id"),
        field.String("organization_id"),
        field.String("position_title"),
        field.String("department"),
        field.Time("effective_date"),
        field.Time("end_date").Optional().Nillable(),
        field.Bool("is_active").Default(true),
        field.Bool("is_retroactive").Default(false),
        field.JSON("salary_data", map[string]interface{}{}).Optional(),
        field.String("change_reason").Optional().Nillable(),
        field.String("approval_status").Default("approved"),
        field.Time("created_at").Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}