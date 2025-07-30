package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// Position holds the schema definition for the Position entity.
// Represents a specific organizational "chair" or role that can be occupied by employees.
// Implements the core-hr.keep namespace for position management following
// the Castle Model architecture and Meta-Contract v6.0 specifications.
type Position struct {
	ent.Schema
}

// Fields of the Position.
func (Position) Fields() []ent.Field {
	return []ent.Field{
		// Core Identity (Meta-Contract v6.0 compliance)
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Comment("Global unique identifier, immutable primary key"),

		field.UUID("tenant_id", uuid.UUID{}).
			Immutable().
			Comment("Multi-tenant isolation foundation, enforces data boundary"),

		// Polymorphic Discriminator
		field.Enum("position_type").
			Values("FULL_TIME", "PART_TIME", "CONTINGENT_WORKER", "INTERN").
			Comment("Position type discriminator for details slot determination"),

		// Job Profile Reference
		field.UUID("job_profile_id", uuid.UUID{}).
			Comment("Reference to JobProfile entity (position template)"),

		// Organization Relationship
		field.UUID("department_id", uuid.UUID{}).
			Comment("Reference to OrganizationUnit this position belongs to"),

		// Reporting Structure
		field.UUID("manager_position_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Self-referencing foreign key for reporting hierarchy"),

		// Position Status
		field.Enum("status").
			Values("OPEN", "FILLED", "FROZEN", "PENDING_ELIMINATION").
			Default("OPEN").
			Comment("Current lifecycle status of the position"),

		// Resource Planning
		field.Float("budgeted_fte").
			Default(1.0).
			Comment("Budgeted Full-Time Equivalent for resource planning"),

		// Polymorphic Details Slot
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("Polymorphic configuration based on position_type discriminator"),

		// Audit Trail (Event Sourcing Support)
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Immutable creation timestamp for audit trail"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last modification timestamp, auto-updated"),
	}
}

// Edges of the Position.
func (Position) Edges() []ent.Edge {
	return []ent.Edge{
		// Reporting Relationships (Management Hierarchy)
		edge.To("direct_reports", Position.Type).
			From("manager").
			Field("manager_position_id").
			Unique().
			Comment("Manager-subordinate reporting hierarchy"),

		// Organization Relationship
		edge.From("department", OrganizationUnit.Type).
			Field("department_id").
			Ref("positions").
			Unique().
			Required().
			Comment("Organization unit that contains this position"),

		// Occupancy History (Temporal Relationships)
		edge.To("occupancy_history", PositionOccupancyHistory.Type).
			Comment("Historical record of employees who occupied this position"),

		// Attribute History (Temporal Changes)
		edge.To("attribute_history", PositionAttributeHistory.Type).
			Comment("Historical record of position attribute changes"),
	}
}

// Indexes of the Position.
func (Position) Indexes() []ent.Index {
	return []ent.Index{
		// Multi-tenant isolation optimization
		index.Fields("tenant_id", "position_type"),

		// Department relationship optimization
		index.Fields("department_id"),

		// Reporting hierarchy optimization
		index.Fields("manager_position_id"),

		// Status filtering optimization
		index.Fields("tenant_id", "status"),

		// Job profile relationship optimization
		index.Fields("job_profile_id"),

		// Resource planning optimization
		index.Fields("tenant_id", "budgeted_fte"),

		// Composite index for complex queries
		index.Fields("tenant_id", "department_id", "status"),
	}
}