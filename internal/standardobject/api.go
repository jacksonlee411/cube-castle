package standardobject

import (
	"context"
	"time"
)

// ObjectType enumerates the Standard Object families supported by the SOM plan.
type ObjectType string

const (
	// ObjectTypeOrganizationUnit represents organization units.
	ObjectTypeOrganizationUnit ObjectType = "ORGANIZATION_UNIT"
	// ObjectTypePositionRole represents the planned position role object.
	ObjectTypePositionRole ObjectType = "POSITION_ROLE"
)

// LifecycleStatus follows the target state machine defined in Plan 400 §4.2.
type LifecycleStatus string

const (
	StatusDraft     LifecycleStatus = "DRAFT"
	StatusReady     LifecycleStatus = "READY"
	StatusActive    LifecycleStatus = "ACTIVE"
	StatusSuspended LifecycleStatus = "SUSPENDED"
	StatusRetired   LifecycleStatus = "RETIRED"
)

// ObjectKernel stores tenant level metadata that must remain stable across versions.
type ObjectKernel struct {
	ID                 string
	ObjectType         ObjectType
	Code               string
	DisplayName        string
	TenantCode         string
	Status             LifecycleStatus
	Labels             map[string]string
	SchemaVersion      string
	DataClassification string
	RetentionPolicy    string
	CreatedBy          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// TemporalVersion mirrors the payload and audit envelope defined for standard_object_versions.
type TemporalVersion struct {
	VersionID     string
	VersionCode   string
	EffectiveDate time.Time
	EndDate       *time.Time
	IsCurrent     bool
	Payload       map[string]any
	AuditTrail    map[string]any
	Checksum      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Link models the relationship entries that will be persisted in standard_object_links.
type Link struct {
	LinkID     string
	LinkType   string
	SourceCode string
	TargetCode string
	Attributes map[string]any
	TenantCode string
}

// ObjectAggregate bundles the kernel, an active version and optional link metadata.
type ObjectAggregate struct {
	Kernel  ObjectKernel
	Version TemporalVersion
	Links   []Link
}

// ObjectKey is the minimal lookup tuple used by repositories and adapters.
type ObjectKey struct {
	ObjectType ObjectType
	Code       string
	TenantCode string
}

// SchemaFieldBinding captures the DEC/OCL metadata for a single payload field.
type SchemaFieldBinding struct {
	FieldPath   string
	DECID       string
	Description string
	GlossaryURL string
}

// SchemaDefinition references schema-registry.json entries required by Plan 402A.
type SchemaDefinition struct {
	ObjectType    ObjectType
	SchemaVersion string
	SchemaHash    string
	Fields        []SchemaFieldBinding
	OCLGuards     []string
	KnownGaps     []string
}

// RegistryReader exposes the minimal operations the adapter needs to enforce DEC/OCL contracts.
type RegistryReader interface {
	LookupSchema(ctx context.Context, objectType ObjectType) (SchemaDefinition, error)
}

// ObjectRepository describes the storage contract. Implementations are provided in later phases.
type ObjectRepository interface {
	Upsert(ctx context.Context, aggregate ObjectAggregate) error
	Get(ctx context.Context, key ObjectKey) (ObjectAggregate, error)
}

// ObjectService defines the dependency injected into command/query services.
type ObjectService interface {
	Upsert(ctx context.Context, aggregate ObjectAggregate) error
	Get(ctx context.Context, key ObjectKey) (ObjectAggregate, error)
}
