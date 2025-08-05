package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"regexp"
	"time"
)

// Employee holds the schema definition for the Employee entity.
// Implements the core-hr.keep namespace for employee management following
// the Castle Model architecture and Meta-Contract v6.0 specifications.
type Employee struct {
	ent.Schema
}

// Fields of the Employee.
func (Employee) Fields() []ent.Field {
	return []ent.Field{
		// Core Identity (Meta-Contract v6.0 compliance)
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Comment("Global unique identifier, immutable primary key"),

		// Business ID (User-friendly identifier)
		field.String("business_id").
			Unique().
			NotEmpty().
			MaxLen(8).
			Match(regexp.MustCompile(`^[1-9][0-9]{0,7}$`)).
			Comment("Business ID - user-friendly identifier (1-99999999)"),

		field.UUID("tenant_id", uuid.UUID{}).
			Immutable().
			Comment("Multi-tenant isolation foundation, enforces data boundary"),

		// Polymorphic Discriminator
		field.Enum("employee_type").
			Values("FULL_TIME", "PART_TIME", "CONTRACTOR", "INTERN").
			Comment("Employee type discriminator for details slot determination"),

		// Core Business Attributes
		field.String("employee_number").
			NotEmpty().
			MaxLen(50).
			Comment("Employee number, unique within enterprise"),

		field.String("first_name").
			NotEmpty().
			MaxLen(100).
			Comment("Employee first name"),

		field.String("last_name").
			NotEmpty().
			MaxLen(100).
			Comment("Employee last name"),

		field.String("email").
			NotEmpty().
			MaxLen(255).
			Comment("Corporate email address"),

		field.String("personal_email").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("Personal email address"),

		field.String("phone_number").
			Optional().
			Nillable().
			MaxLen(20).
			Comment("Contact phone number"),

		// Current Position Relationship (replaces original string position field)
		field.UUID("current_position_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Current primary position reference"),

		// Department Relationship - Direct relationship for improved performance
		field.UUID("department_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Direct department/organization unit reference for efficient queries"),

		// Employment Status
		field.Enum("employment_status").
			Values("ACTIVE", "ON_LEAVE", "TERMINATED", "SUSPENDED", "PENDING_START").
			Default("PENDING_START").
			Comment("Current employment status"),

		// Employment Dates
		field.Time("hire_date").
			Comment("Employment start date"),

		field.Time("termination_date").
			Optional().
			Nillable().
			Comment("Employment end date (if applicable)"),

		// Polymorphic Details Slot
		field.JSON("employee_details", map[string]interface{}{}).
			Optional().
			Comment("Polymorphic configuration based on employee_type discriminator"),

		// Legacy fields (maintain for backward compatibility during migration)
		field.String("name").
			Optional().
			Comment("Legacy name field - will be deprecated"),

		field.String("position").
			Optional().
			Comment("Legacy position field - will be deprecated"),

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

// Edges of the Employee.
func (Employee) Edges() []ent.Edge {
	return []ent.Edge{
		// Current Position Relationship
		edge.From("current_position", Position.Type).
			Field("current_position_id").
			Ref("current_incumbents").
			Unique().
			Comment("Employee current primary position"),

		// Department Relationship - Direct reference for performance
		edge.From("department", OrganizationUnit.Type).
			Field("department_id").
			Ref("employees").
			Unique().
			Comment("Employee current department/organization unit"),

		// Position History Relationship
		edge.To("position_history", PositionOccupancyHistory.Type).
			Comment("Employee position occupancy history records"),

		// Position Assignments Relationship
		edge.To("assignments", PositionAssignment.Type).
			Comment("Employee position assignments"),
	}
}

// Indexes of the Employee.
func (Employee) Indexes() []ent.Index {
	return []ent.Index{
		// Multi-tenant isolation optimization
		index.Fields("tenant_id", "employee_type"),

		// Business ID optimization (primary lookup)
		index.Fields("business_id"),
		index.Fields("tenant_id", "business_id"),

		// Employee number uniqueness (tenant-scoped)
		index.Fields("tenant_id", "employee_number").Unique(),

		// Email uniqueness (tenant-scoped)
		index.Fields("tenant_id", "email").Unique(),

		// Status filtering optimization
		index.Fields("tenant_id", "employment_status"),

		// Current position relationship optimization
		index.Fields("current_position_id"),

		// Department relationship optimization 
		index.Fields("department_id"),
		index.Fields("tenant_id", "department_id"),
		
		// Combined position and department optimization
		index.Fields("current_position_id", "department_id"),

		// Hire date query optimization
		index.Fields("tenant_id", "hire_date"),

		// Composite index for complex queries
		index.Fields("tenant_id", "employment_status", "employee_type"),

		// Legacy field indexes (for migration period)
		index.Fields("email"),
		index.Fields("name"),
	}
}
