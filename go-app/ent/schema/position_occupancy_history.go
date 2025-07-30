package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// PositionOccupancyHistory holds the schema definition for tracking who occupied positions over time.
// This is separate from PositionAttributeHistory and focuses specifically on the
// employee-position relationship temporal dimension. Supports the dual-history pattern.
type PositionOccupancyHistory struct {
	ent.Schema
}

// Fields of the PositionOccupancyHistory.
func (PositionOccupancyHistory) Fields() []ent.Field {
	return []ent.Field{
		// Core Identity
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Comment("Unique identifier for this occupancy record"),

		field.UUID("tenant_id", uuid.UUID{}).
			Immutable().
			Comment("Multi-tenant isolation foundation"),

		// Primary Relationship
		field.UUID("position_id", uuid.UUID{}).
			Comment("Reference to the position being occupied"),

		field.UUID("employee_id", uuid.UUID{}).
			Comment("Reference to the employee occupying the position"),

		// Occupancy Time Range
		field.Time("start_date").
			Comment("When the employee started occupying this position"),

		field.Time("end_date").
			Optional().
			Nillable().
			Comment("When the employee stopped occupying this position (null = current)"),

		// Occupancy Status
		field.Bool("is_active").
			Default(true).
			Comment("Whether this occupancy relationship is currently active"),

		// Assignment Metadata
		field.Enum("assignment_type").
			Values("REGULAR", "INTERIM", "ACTING", "TEMPORARY", "SECONDMENT").
			Default("REGULAR").
			Comment("Type of position assignment"),

		field.String("assignment_reason").
			Optional().
			Comment("Business reason for this assignment"),

		// Work Arrangement Details
		field.Float("fte_percentage").
			Default(1.0).
			Comment("Full-time equivalent percentage for this assignment"),

		field.Enum("work_arrangement").
			Values("ON_SITE", "REMOTE", "HYBRID").
			Optional().
			Comment("Work location arrangement for this assignment"),

		// Approval and Authorization
		field.UUID("approved_by", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Person who approved this assignment"),

		field.Time("approval_date").
			Optional().
			Nillable().
			Comment("When this assignment was approved"),

		field.String("approval_reference").
			Optional().
			Comment("Reference to approval document or process"),

		// Performance and Compensation Context
		field.JSON("compensation_data", map[string]interface{}{}).
			Optional().
			Comment("Compensation details specific to this assignment"),

		field.String("performance_review_cycle").
			Optional().
			Comment("Performance review cycle for this assignment"),

		// Event Sourcing Integration
		field.UUID("source_event_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Reference to the business event that caused this assignment"),

		// Audit Trail
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("When this occupancy record was created"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("When this occupancy record was last updated"),
	}
}

// Edges of the PositionOccupancyHistory.
func (PositionOccupancyHistory) Edges() []ent.Edge {
	return []ent.Edge{
		// Reference back to the position
		edge.From("position", Position.Type).
			Field("position_id").
			Ref("occupancy_history").
			Unique().
			Required().
			Comment("The position this occupancy record belongs to"),

		// Reference to the employee (when Employee schema is available)
		// edge.From("employee", Employee.Type).
		//     Field("employee_id").
		//     Ref("position_history").
		//     Unique().
		//     Required().
		//     Comment("The employee who occupied the position"),
	}
}

// Indexes of the PositionOccupancyHistory.
func (PositionOccupancyHistory) Indexes() []ent.Index {
	return []ent.Index{
		// Primary occupancy queries
		index.Fields("position_id", "start_date"),

		index.Fields("employee_id", "start_date"),

		// Multi-tenant isolation
		index.Fields("tenant_id"),

		// Date range queries
		index.Fields("start_date", "end_date"),

		// Assignment type analysis
		index.Fields("assignment_type", "start_date"),

		// FTE tracking
		index.Fields("position_id", "fte_percentage", "start_date"),

		// Approval tracking
		index.Fields("approved_by", "approval_date"),

		// Event sourcing integration
		index.Fields("source_event_id"),

		// Overlapping assignments detection
		index.Fields("employee_id", "start_date", "end_date"),

		// Performance review scheduling
		index.Fields("performance_review_cycle", "start_date"),
	}
}
