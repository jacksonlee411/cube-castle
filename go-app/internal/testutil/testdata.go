package testdata

// SampleMetaContracts contains sample meta-contract YAML files for testing

const ValidUserContract = `specification_version: "1.0"
api_id: "550e8400-e29b-41d4-a716-446655440000"
namespace: "hr"
resource_name: "user"
version: "1.0.0"

data_structure:
  primary_key: "id"
  data_classification: "CONFIDENTIAL"
  fields:
    - name: "id"
      type: "uuid"
      required: true
      unique: true
      data_classification: "INTERNAL"
    - name: "email"
      type: "string"
      required: true
      unique: true
      data_classification: "CONFIDENTIAL"
      validation_rules:
        - "email_format"
        - "max_length_255"
    - name: "first_name"
      type: "string"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "max_length_100"
    - name: "last_name"
      type: "string"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "max_length_100"
    - name: "phone"
      type: "string"
      required: false
      data_classification: "CONFIDENTIAL"
      validation_rules:
        - "phone_format"
    - name: "date_of_birth"
      type: "time"
      required: false
      data_classification: "CONFIDENTIAL"
    - name: "active"
      type: "bool"
      required: true
      data_classification: "INTERNAL"
    - name: "created_at"
      type: "time"
      required: true
      data_classification: "INTERNAL"
    - name: "updated_at"
      type: "time"
      required: true
      data_classification: "INTERNAL"

security_model:
  tenant_isolation: true
  access_control: "RBAC"
  data_classification: "CONFIDENTIAL"
  compliance_tags:
    - "GDPR"
    - "CCPA"
    - "SOX"

temporal_behavior:
  temporality_paradigm: "EVENT_DRIVEN"
  state_transition_model: "EVENT_DRIVEN"
  history_retention: "7 years"
  event_driven: true

api_behavior:
  rest_enabled: true
  graphql_enabled: true
  events_enabled: true

relationships:
  - name: "user_profile"
    type: "one-to-one"
    target_entity: "user_profile"
    cardinality: "1:1"
    is_optional: false
  - name: "user_roles"
    type: "one-to-many"
    target_entity: "user_role"
    cardinality: "1:N"
    is_optional: true
  - name: "user_permissions"
    type: "many-to-many"
    target_entity: "permission"
    cardinality: "M:N"
    is_optional: true`

const ComplexOrganizationContract = `specification_version: "1.0"
api_id: "660e8400-e29b-41d4-a716-446655440001"
namespace: "hr"
resource_name: "organization"
version: "2.0.0"

data_structure:
  primary_key: "id"
  data_classification: "INTERNAL"
  persistence_profile:
    primary_store: "postgresql"
    indexed_in: ["elasticsearch", "neo4j"]
    graph_node_label: "Organization"
    graph_edge_definitions:
      - "PARENT_OF"
      - "OWNS"
      - "MANAGES"
  polymorphism:
    discriminator: "organization_type"
    concrete_types:
      company: "Company"
      department: "Department"
      team: "Team"
  fields:
    - name: "id"
      type: "uuid"
      required: true
      unique: true
      data_classification: "INTERNAL"
    - name: "name"
      type: "string"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "max_length_200"
        - "unique_within_parent"
    - name: "organization_type"
      type: "enum"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "enum_values: company,department,team"
    - name: "parent_id"
      type: "uuid"
      required: false
      data_classification: "INTERNAL"
    - name: "description"
      type: "string"
      required: false
      data_classification: "INTERNAL"
      validation_rules:
        - "max_length_1000"
    - name: "industry"
      type: "enum"
      required: false
      data_classification: "INTERNAL"
    - name: "size"
      type: "enum"
      required: false
      data_classification: "INTERNAL"
    - name: "headquarters_address"
      type: "json"
      required: false
      data_classification: "INTERNAL"
    - name: "contact_info"
      type: "json"
      required: false
      data_classification: "CONFIDENTIAL"
    - name: "tax_id"
      type: "string"
      required: false
      data_classification: "RESTRICTED"
      validation_rules:
        - "encrypted"
    - name: "legal_structure"
      type: "enum"
      required: false
      data_classification: "INTERNAL"
    - name: "founding_date"
      type: "time"
      required: false
      data_classification: "INTERNAL"
    - name: "status"
      type: "enum"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "enum_values: active,inactive,dissolved"
    - name: "metadata"
      type: "json"
      required: false
      data_classification: "INTERNAL"
    - name: "created_at"
      type: "time"
      required: true
      data_classification: "INTERNAL"
    - name: "updated_at"
      type: "time"
      required: true
      data_classification: "INTERNAL"
    - name: "deleted_at"
      type: "time"
      required: false
      data_classification: "INTERNAL"

security_model:
  tenant_isolation: true
  access_control: "ABAC"
  data_classification: "CONFIDENTIAL"
  compliance_tags:
    - "GDPR"
    - "SOX"
    - "CCPA"

temporal_behavior:
  temporality_paradigm: "HYBRID"
  state_transition_model: "STATE_MACHINE"
  history_retention: "10 years"
  event_driven: true

api_behavior:
  rest_enabled: true
  graphql_enabled: true
  events_enabled: true

relationships:
  - name: "parent_organization"
    type: "one-to-one"
    target_entity: "organization"
    cardinality: "1:1"
    is_optional: true
    graph_edge: "CHILD_OF"
  - name: "child_organizations"
    type: "one-to-many"
    target_entity: "organization"
    cardinality: "1:N"
    is_optional: true
    graph_edge: "PARENT_OF"
  - name: "employees"
    type: "one-to-many"
    target_entity: "employee"
    cardinality: "1:N"
    is_optional: true
    graph_edge: "EMPLOYS"
  - name: "locations"
    type: "one-to-many"
    target_entity: "location"
    cardinality: "1:N"
    is_optional: true
    graph_edge: "HAS_LOCATION"`

const InvalidContract = `specification_version: "1.0"
api_id: "invalid-uuid"
namespace: ""
resource_name: ""
version: ""

data_structure:
  primary_key: "non_existent_field"
  fields:
    - name: ""
      type: "invalid_type"
      required: true
    - name: "duplicate"
      type: "string"
    - name: "duplicate"
      type: "int"

security_model:
  access_control: "INVALID_ACCESS_CONTROL"
  data_classification: "INVALID_CLASSIFICATION"

temporal_behavior:
  temporality_paradigm: "INVALID_PARADIGM"
  state_transition_model: "INVALID_MODEL"

relationships:
  - name: ""
    type: "invalid-relationship-type"
    target_entity: ""`

const MinimalContract = `specification_version: "1.0"
api_id: "770e8400-e29b-41d4-a716-446655440002"
namespace: "test"
resource_name: "simple"
version: "1.0.0"

data_structure:
  fields:
    - name: "id"
      type: "uuid"
      required: true`

const SecurityFocusedContract = `specification_version: "1.0"
api_id: "880e8400-e29b-41d4-a716-446655440003"
namespace: "security"
resource_name: "secure_document"
version: "1.0.0"

data_structure:
  primary_key: "id"
  data_classification: "RESTRICTED"
  fields:
    - name: "id"
      type: "uuid"
      required: true
      unique: true
      data_classification: "INTERNAL"
    - name: "title"
      type: "string"
      required: true
      data_classification: "CONFIDENTIAL"
      validation_rules:
        - "max_length_500"
        - "sanitize_html"
    - name: "content"
      type: "string"
      required: true
      data_classification: "RESTRICTED"
      validation_rules:
        - "encrypted"
        - "audit_trail"
    - name: "classification_level"
      type: "enum"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "enum_values: public,internal,confidential,restricted,top_secret"
    - name: "owner_id"
      type: "uuid"
      required: true
      data_classification: "INTERNAL"
    - name: "access_control_list"
      type: "json"
      required: true
      data_classification: "CONFIDENTIAL"
      validation_rules:
        - "encrypted"
    - name: "encryption_key_id"
      type: "uuid"
      required: true
      data_classification: "RESTRICTED"
    - name: "digital_signature"
      type: "string"
      required: false
      data_classification: "INTERNAL"
    - name: "access_log"
      type: "json"
      required: false
      data_classification: "INTERNAL"
      validation_rules:
        - "immutable"
    - name: "created_at"
      type: "time"
      required: true
      data_classification: "INTERNAL"
    - name: "created_by"
      type: "uuid"
      required: true
      data_classification: "INTERNAL"

security_model:
  tenant_isolation: true
  access_control: "ABAC"
  data_classification: "RESTRICTED"
  compliance_tags:
    - "GDPR"
    - "SOX"
    - "CCPA"
    - "HIPAA"
    - "ISO27001"

temporal_behavior:
  temporality_paradigm: "EVENT_DRIVEN"
  state_transition_model: "IMMUTABLE"
  history_retention: "forever"
  event_driven: true

api_behavior:
  rest_enabled: true
  graphql_enabled: false
  events_enabled: true`

const PerformanceOptimizedContract = `specification_version: "1.0"
api_id: "990e8400-e29b-41d4-a716-446655440004"
namespace: "performance"
resource_name: "high_volume_transaction"
version: "1.0.0"

data_structure:
  primary_key: "id"
  data_classification: "INTERNAL"
  persistence_profile:
    primary_store: "postgresql"
    indexed_in: ["redis", "elasticsearch"]
    graph_node_label: "Transaction"
  fields:
    - name: "id"
      type: "uuid"
      required: true
      unique: true
      data_classification: "INTERNAL"
    - name: "transaction_id"
      type: "string"
      required: true
      unique: true
      data_classification: "INTERNAL"
      validation_rules:
        - "indexed"
        - "max_length_50"
    - name: "user_id"
      type: "uuid"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "indexed"
    - name: "amount"
      type: "float64"
      required: true
      data_classification: "CONFIDENTIAL"
      validation_rules:
        - "positive"
        - "max_decimal_places_2"
    - name: "currency"
      type: "enum"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "enum_values: USD,EUR,GBP,JPY"
        - "indexed"
    - name: "transaction_type"
      type: "enum"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "enum_values: debit,credit,transfer"
        - "indexed"
    - name: "status"
      type: "enum"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "enum_values: pending,completed,failed,cancelled"
        - "indexed"
    - name: "merchant_id"
      type: "uuid"
      required: false
      data_classification: "INTERNAL"
      validation_rules:
        - "indexed"
    - name: "reference_number"
      type: "string"
      required: false
      data_classification: "INTERNAL"
      validation_rules:
        - "indexed"
        - "max_length_100"
    - name: "description"
      type: "string"
      required: false
      data_classification: "INTERNAL"
      validation_rules:
        - "max_length_255"
    - name: "metadata"
      type: "json"
      required: false
      data_classification: "INTERNAL"
    - name: "processed_at"
      type: "time"
      required: false
      data_classification: "INTERNAL"
      validation_rules:
        - "indexed"
    - name: "created_at"
      type: "time"
      required: true
      data_classification: "INTERNAL"
      validation_rules:
        - "indexed"
        - "partitioned_by_month"

security_model:
  tenant_isolation: true
  access_control: "RBAC"
  data_classification: "CONFIDENTIAL"
  compliance_tags:
    - "PCI_DSS"
    - "SOX"

temporal_behavior:
  temporality_paradigm: "EVENT_DRIVEN"
  state_transition_model: "EVENT_DRIVEN"
  history_retention: "7 years"
  event_driven: true

api_behavior:
  rest_enabled: true
  graphql_enabled: true
  events_enabled: true

relationships:
  - name: "user"
    type: "one-to-one"
    target_entity: "user"
    cardinality: "1:1"
    is_optional: false
  - name: "merchant"
    type: "one-to-one"
    target_entity: "merchant"
    cardinality: "1:1"
    is_optional: true`

// Sample natural language inputs for NLP testing
const (
	CreateUserEntityNL        = "Create a user entity with id, name, email, and phone fields"
	AddFieldsNL              = "Add address and date_of_birth fields to the user entity"
	CreateRelationshipNL     = "User has many posts and belongs to one organization"
	ModifyFieldNL            = "Make email field required and unique"
	CreateComplexEntityNL    = "Create an order entity with id, user_id, total_amount, status, items array, and created_at timestamp"
	SecurityRequirementNL    = "Add encryption to credit card field and make it restricted access"
	PerformanceOptimizationNL = "Add indexes on frequently queried fields like email and status"
	ValidationRulesNL        = "Add email format validation and length limits to name fields"
	EnumFieldNL              = "Create status field with values: draft, published, archived"
	JSONFieldNL              = "Add metadata field that can store any JSON data"
)

// Expected AI responses for testing
const (
	ExpectedUserEntityYAML = `name: user
fields:
  - name: id
    type: uuid
    required: true
    unique: true
  - name: name
    type: string
    required: true
  - name: email
    type: string
    required: true
    unique: true
    validation_rules:
      - email_format
  - name: phone
    type: string
    required: false`

	ExpectedOrderEntityYAML = `name: order
fields:
  - name: id
    type: uuid
    required: true
    unique: true
  - name: user_id
    type: uuid
    required: true
  - name: total_amount
    type: float64
    required: true
  - name: status
    type: enum
    required: true
  - name: items
    type: json
    required: true
  - name: created_at
    type: time
    required: true`
)

// Error scenarios for testing
const (
	MalformedYAML = `invalid_yaml: [
  unclosed: bracket`

	EmptyContract = `# Empty contract with minimal required fields missing`

	ContractWithSyntaxError = `specification_version: "1.0"
namespace: "test"
resource_name: "invalid"
version: "1.0.0"
data_structure:
  fields:
    - name: "field1"
      type: "string"
    - name: "field1"  # Duplicate field name
      type: "int"
  primary_key: "non_existent_field"  # References non-existent field`
)