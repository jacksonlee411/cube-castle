// internal/types/metacontract.go
package types

import (
	"github.com/google/uuid"
)

// MetaContract represents the complete meta-contract specification
type MetaContract struct {
	SpecificationVersion string                 `yaml:"specification_version"`
	APIID               uuid.UUID              `yaml:"api_id"`
	Namespace           string                 `yaml:"namespace"`
	ResourceName        string                 `yaml:"resource_name"`
	Version             string                 `yaml:"version"`
	DataStructure       DataStructure          `yaml:"data_structure"`
	SecurityModel       SecurityModel          `yaml:"security_model"`
	TemporalBehavior    TemporalBehaviorModel  `yaml:"temporal_behavior"`
	APIBehavior         APIBehaviorModel       `yaml:"api_behavior"`
	Relationships       []RelationshipDef      `yaml:"relationships"`
}

// DataStructure defines the data model structure
type DataStructure struct {
	Fields               []FieldDefinition      `yaml:"fields"`
	PrimaryKey          string                 `yaml:"primary_key"`
	DataClassification  string                 `yaml:"data_classification"`
	PersistenceProfile  *PersistenceProfile    `yaml:"persistence_profile,omitempty"`
	Polymorphism        *PolymorphismDef       `yaml:"polymorphism,omitempty"`
}

// FieldDefinition defines a single field in the data structure
type FieldDefinition struct {
	Name                string   `yaml:"name"`
	Type                string   `yaml:"type"`
	Required            bool     `yaml:"required"`
	Unique              bool     `yaml:"unique"`
	DataClassification  string   `yaml:"data_classification"`
	ValidationRules     []string `yaml:"validation_rules,omitempty"`
}

// PersistenceProfile defines where and how data is stored
type PersistenceProfile struct {
	PrimaryStore         string   `yaml:"primary_store"`
	IndexedIn           []string  `yaml:"indexed_in,omitempty"`
	GraphNodeLabel      string    `yaml:"graph_node_label,omitempty"`
	GraphEdgeDefinitions []string `yaml:"graph_edge_definitions,omitempty"`
}

// PolymorphismDef defines polymorphic behavior
type PolymorphismDef struct {
	Discriminator       string              `yaml:"discriminator"`
	ConcreteTypes       map[string]string   `yaml:"concrete_types"`
}

// SecurityModel defines security constraints
type SecurityModel struct {
	TenantIsolation     bool              `yaml:"tenant_isolation"`
	AccessControl       string            `yaml:"access_control"`
	DataClassification  string            `yaml:"data_classification"`
	ComplianceTags      []string          `yaml:"compliance_tags"`
}

// TemporalBehaviorModel defines time-based behavior
type TemporalBehaviorModel struct {
	TemporalityParadigm    string    `yaml:"temporality_paradigm"`
	StateTransitionModel   string    `yaml:"state_transition_model"`
	HistoryRetention       string    `yaml:"history_retention"`
	EventDriven           bool      `yaml:"event_driven"`
}

// APIBehaviorModel defines API generation behavior
type APIBehaviorModel struct {
	RESTEnabled    bool `yaml:"rest_enabled"`
	GraphQLEnabled bool `yaml:"graphql_enabled"`
	EventsEnabled  bool `yaml:"events_enabled"`
}

// RelationshipDef defines relationships between entities
type RelationshipDef struct {
	Name           string `yaml:"name"`
	Type           string `yaml:"type"` // "one-to-one", "one-to-many", "many-to-many"
	TargetEntity   string `yaml:"target_entity"`
	Cardinality    string `yaml:"cardinality"`
	IsOptional     bool   `yaml:"is_optional"`
	GraphEdge      string `yaml:"graph_edge,omitempty"`
}

// CompilerInterface defines the meta-contract compiler interface
type CompilerInterface interface {
	// ParseMetaContract parses a YAML meta-contract file
	ParseMetaContract(yamlPath string) (*MetaContract, error)
	
	// GenerateEntSchemas generates Ent schema files
	GenerateEntSchemas(contract *MetaContract, outputDir string) error
	
	// GenerateBusinessLogic generates business logic skeleton
	GenerateBusinessLogic(contract *MetaContract, outputDir string) error
	
	// GenerateAPIRoutes generates API route definitions
	GenerateAPIRoutes(contract *MetaContract, outputDir string) error
}