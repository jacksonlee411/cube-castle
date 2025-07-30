package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// PositionAttributeHistory holds the schema definition for position attribute changes over time.
// This implements the temporal dimension for position attributes, separate from occupancy history.
// Part of the event-driven architecture where position changes are captured as immutable events.
type PositionAttributeHistory struct {
	ent.Schema
}

// Fields of the PositionAttributeHistory.
func (PositionAttributeHistory) Fields() []ent.Field {
	return []ent.Field{
		// Core Identity
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Comment("Unique identifier for this history record"),

		field.UUID("tenant_id", uuid.UUID{}).
			Immutable().
			Comment("Multi-tenant isolation foundation"),

		// Position Reference
		field.UUID("position_id", uuid.UUID{}).
			Comment("Reference to the position this history record belongs to"),

		// Snapshot of Position Attributes at Point in Time
		field.String("position_type").
			Comment("Position type at this point in time"),

		field.UUID("job_profile_id", uuid.UUID{}).
			Comment("Job profile reference at this point in time"),

		field.UUID("department_id", uuid.UUID{}).
			Comment("Department reference at this point in time"),

		field.UUID("manager_position_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Manager position reference at this point in time"),

		field.String("status").
			Comment("Position status at this point in time"),

		field.Float("budgeted_fte").
			Comment("Budgeted FTE at this point in time"),

		// Polymorphic Details Snapshot
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("Position details snapshot at this point in time"),

		// Temporal Boundaries
		field.Time("effective_date").
			Comment("When this attribute configuration became effective"),

		field.Time("end_date").
			Optional().
			Nillable().
			Comment("When this attribute configuration ended (null = current)"),

		// Change Metadata
		field.String("change_reason").
			Optional().
			Comment("Business reason for this attribute change"),

		field.UUID("changed_by", uuid.UUID{}).
			Comment("Person who initiated this change"),

		field.String("change_type").
			Optional().
			Comment("Type of change (e.g., CREATION, MODIFICATION, CLOSURE)"),

		// Event Sourcing Integration
		field.UUID("source_event_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Reference to the business event that caused this change"),

		// Audit Trail
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("When this history record was created"),
	}
}

// Edges of the PositionAttributeHistory.
func (PositionAttributeHistory) Edges() []ent.Edge {
	return []ent.Edge{
		// Reference back to the position
		edge.From("position", Position.Type).
			Field("position_id").
			Ref("attribute_history").
			Unique().
			Required().
			Comment("The position this history record belongs to"),
	}
}

// Indexes of the PositionAttributeHistory.
func (PositionAttributeHistory) Indexes() []ent.Index {
	return []ent.Index{
		// Primary temporal query pattern
		index.Fields("position_id", "effective_date"),

		// Multi-tenant isolation
		index.Fields("tenant_id"),

		// Date range queries
		index.Fields("effective_date", "end_date"),

		// Change tracking queries
		index.Fields("changed_by", "created_at"),

		// Event sourcing integration
		index.Fields("source_event_id"),

		// Department change tracking
		index.Fields("department_id", "effective_date"),

		// Status change tracking
		index.Fields("position_id", "status", "effective_date"),
	}
}
