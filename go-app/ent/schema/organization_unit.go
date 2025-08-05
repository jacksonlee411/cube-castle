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

// OrganizationUnit holds the schema definition for the OrganizationUnit entity.
// Implements the core-hr.keep namespace for organizational structure management
// following the Castle Model architecture and Meta-Contract v6.0 specifications.
type OrganizationUnit struct {
	ent.Schema
}

// Fields of the OrganizationUnit.
func (OrganizationUnit) Fields() []ent.Field {
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
			MaxLen(6).
			Match(regexp.MustCompile(`^[1-9][0-9]{5}$`)).
			Comment("Business ID - user-friendly identifier (100000-999999)"),

		field.UUID("tenant_id", uuid.UUID{}).
			Immutable().
			Comment("Multi-tenant isolation foundation, enforces data boundary"),

		// Polymorphic Discriminator
		field.Enum("unit_type").
			Values("DEPARTMENT", "COST_CENTER", "COMPANY", "PROJECT_TEAM").
			Comment("Polymorphic discriminator for profile slot determination"),

		// Core Business Attributes
		field.String("name").
			NotEmpty().
			MaxLen(100).
			Comment("Human-readable organization unit name"),

		field.String("description").
			Optional().
			Nillable().
			MaxLen(500).
			Comment("Detailed description of unit purpose and responsibilities"),

		// Hierarchical Structure
		field.UUID("parent_unit_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Self-referencing foreign key for organizational hierarchy"),

		// Hierarchical Level (computed field for frontend compatibility)
		field.Int("level").
			Default(0).
			Comment("Hierarchy depth level, 0 for root, computed from parent chain"),

		// Operational Status
		field.Enum("status").
			Values("ACTIVE", "INACTIVE", "PLANNED").
			Default("ACTIVE").
			Comment("Current operational status of the organization unit"),

		// Polymorphic Profile Slot
		field.JSON("profile", map[string]interface{}{}).
			Optional().
			Comment("Polymorphic configuration based on unit_type discriminator"),

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

// Edges of the OrganizationUnit.
func (OrganizationUnit) Edges() []ent.Edge {
	return []ent.Edge{
		// Hierarchical Relationships (Tree Structure)
		edge.To("children", OrganizationUnit.Type).
			From("parent").
			Field("parent_unit_id").
			Unique().
			Comment("Parent-child organizational hierarchy"),

		// Containment Relationships - positions belong to this organization unit
		edge.To("positions", Position.Type).
			Comment("Positions contained within this organization unit"),

		// Employee Relationships - employees directly assigned to this organization unit
		edge.To("employees", Employee.Type).
			Comment("Employees directly assigned to this organization unit"),
	}
}

// Indexes of the OrganizationUnit.
func (OrganizationUnit) Indexes() []ent.Index {
	return []ent.Index{
		// Multi-tenant isolation optimization
		index.Fields("tenant_id", "unit_type"),

		// Business ID optimization (primary lookup)
		index.Fields("business_id"),
		index.Fields("tenant_id", "business_id"),

		// Hierarchical query optimization
		index.Fields("parent_unit_id"),

		// Name search optimization (tenant-scoped)
		index.Fields("tenant_id", "name"),

		// Status filtering optimization
		index.Fields("tenant_id", "status"),

		// Composite index for common query patterns
		index.Fields("tenant_id", "unit_type", "status"),
	}
}
