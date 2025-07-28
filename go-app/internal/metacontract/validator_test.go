package metacontract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_NewValidator(t *testing.T) {
	validator := NewValidator()
	assert.NotNil(t, validator)
}

func TestValidator_Validate_Success(t *testing.T) {
	validator := NewValidator()
	contract := createValidTestContract()
	
	err := validator.Validate(contract)
	assert.NoError(t, err)
}

func TestValidator_validateBasicStructure(t *testing.T) {
	validator := NewValidator()
	
	testCases := []struct {
		name          string
		modifyFunc    func(*MetaContract)
		expectError   bool
		errorContains string
	}{
		{
			name: "valid basic structure",
			modifyFunc: func(c *MetaContract) {
				// No modifications - should be valid
			},
			expectError: false,
		},
		{
			name: "missing specification version",
			modifyFunc: func(c *MetaContract) {
				c.SpecificationVersion = ""
			},
			expectError:   true,
			errorContains: "specification_version is required",
		},
		{
			name: "missing resource name",
			modifyFunc: func(c *MetaContract) {
				c.ResourceName = ""
			},
			expectError:   true,
			errorContains: "resource_name is required",
		},
		{
			name: "missing namespace",
			modifyFunc: func(c *MetaContract) {
				c.Namespace = ""
			},
			expectError:   true,
			errorContains: "namespace is required",
		},
		{
			name: "missing version",
			modifyFunc: func(c *MetaContract) {
				c.Version = ""
			},
			expectError:   true,
			errorContains: "version is required",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			contract := createValidTestContract()
			tc.modifyFunc(contract)
			
			err := validator.validateBasicStructure(contract)
			
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_validateDataStructure(t *testing.T) {
	validator := NewValidator()
	
	testCases := []struct {
		name          string
		modifyFunc    func(*MetaContract)
		expectError   bool
		errorContains string
	}{
		{
			name: "valid data structure",
			modifyFunc: func(c *MetaContract) {
				// No modifications - should be valid
			},
			expectError: false,
		},
		{
			name: "empty fields array",
			modifyFunc: func(c *MetaContract) {
				c.DataStructure.Fields = []FieldDefinition{}
			},
			expectError:   true,
			errorContains: "at least one field must be defined",
		},
		{
			name: "duplicate field names",
			modifyFunc: func(c *MetaContract) {
				c.DataStructure.Fields = append(c.DataStructure.Fields, FieldDefinition{
					Name: "id", // Duplicate of existing field
					Type: "string",
				})
			},
			expectError:   true,
			errorContains: "duplicate field name: id",
		},
		{
			name: "primary key not found in fields",
			modifyFunc: func(c *MetaContract) {
				c.DataStructure.PrimaryKey = "nonexistent_field"
			},
			expectError:   true,
			errorContains: "primary_key field 'nonexistent_field' not found",
		},
		{
			name: "empty primary key is valid",
			modifyFunc: func(c *MetaContract) {
				c.DataStructure.PrimaryKey = ""
			},
			expectError: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			contract := createValidTestContract()
			tc.modifyFunc(contract)
			
			err := validator.validateDataStructure(contract)
			
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_validateField(t *testing.T) {
	validator := NewValidator()
	
	testCases := []struct {
		name          string
		field         FieldDefinition
		expectError   bool
		errorContains string
	}{
		{
			name: "valid field",
			field: FieldDefinition{
				Name:               "test_field",
				Type:               "string",
				DataClassification: "INTERNAL",
			},
			expectError: false,
		},
		{
			name: "empty field name",
			field: FieldDefinition{
				Name: "",
				Type: "string",
			},
			expectError:   true,
			errorContains: "field name is required",
		},
		{
			name: "invalid field name format",
			field: FieldDefinition{
				Name: "123invalid", // Starts with number
				Type: "string",
			},
			expectError:   true,
			errorContains: "invalid field name format",
		},
		{
			name: "invalid field type",
			field: FieldDefinition{
				Name: "test_field",
				Type: "invalid_type",
			},
			expectError:   true,
			errorContains: "invalid field type: invalid_type",
		},
		{
			name: "invalid data classification",
			field: FieldDefinition{
				Name:               "test_field",
				Type:               "string",
				DataClassification: "INVALID_CLASS",
			},
			expectError:   true,
			errorContains: "invalid data classification: INVALID_CLASS",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.validateField(tc.field)
			
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_validateSecurityModel(t *testing.T) {
	validator := NewValidator()
	
	testCases := []struct {
		name          string
		modifyFunc    func(*MetaContract)
		expectError   bool
		errorContains string
	}{
		{
			name: "valid security model",
			modifyFunc: func(c *MetaContract) {
				c.SecurityModel.AccessControl = "RBAC"
				c.SecurityModel.DataClassification = "CONFIDENTIAL"
			},
			expectError: false,
		},
		{
			name: "invalid access control",
			modifyFunc: func(c *MetaContract) {
				c.SecurityModel.AccessControl = "INVALID_AC"
			},
			expectError:   true,
			errorContains: "invalid access control model: INVALID_AC",
		},
		{
			name: "invalid data classification",
			modifyFunc: func(c *MetaContract) {
				c.SecurityModel.DataClassification = "INVALID_CLASS"
			},
			expectError:   true,
			errorContains: "invalid data classification: INVALID_CLASS",
		},
		{
			name: "empty access control is valid",
			modifyFunc: func(c *MetaContract) {
				c.SecurityModel.AccessControl = ""
			},
			expectError: false,
		},
		{
			name: "all valid access control types",
			modifyFunc: func(c *MetaContract) {
				// Test will be run multiple times with different values
			},
			expectError: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			contract := createValidTestContract()
			tc.modifyFunc(contract)
			
			err := validator.validateSecurityModel(contract)
			
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
	
	// Test all valid access control types
	validAccessControls := []string{"RBAC", "ABAC", "DAC", "MAC"}
	for _, ac := range validAccessControls {
		t.Run("valid_access_control_"+ac, func(t *testing.T) {
			contract := createValidTestContract()
			contract.SecurityModel.AccessControl = ac
			
			err := validator.validateSecurityModel(contract)
			assert.NoError(t, err)
		})
	}
}

func TestValidator_validateTemporalBehavior(t *testing.T) {
	validator := NewValidator()
	
	testCases := []struct {
		name          string
		modifyFunc    func(*MetaContract)
		expectError   bool
		errorContains string
	}{
		{
			name: "valid temporal behavior",
			modifyFunc: func(c *MetaContract) {
				c.TemporalBehavior.TemporalityParadigm = "EVENT_DRIVEN"
				c.TemporalBehavior.StateTransitionModel = "EVENT_DRIVEN"
			},
			expectError: false,
		},
		{
			name: "invalid temporality paradigm",
			modifyFunc: func(c *MetaContract) {
				c.TemporalBehavior.TemporalityParadigm = "INVALID_PARADIGM"
			},
			expectError:   true,
			errorContains: "invalid temporality paradigm: INVALID_PARADIGM",
		},
		{
			name: "invalid state transition model",
			modifyFunc: func(c *MetaContract) {
				c.TemporalBehavior.StateTransitionModel = "INVALID_MODEL"
			},
			expectError:   true,
			errorContains: "invalid state transition model: INVALID_MODEL",
		},
		{
			name: "empty paradigm is valid",
			modifyFunc: func(c *MetaContract) {
				c.TemporalBehavior.TemporalityParadigm = ""
			},
			expectError: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			contract := createValidTestContract()
			tc.modifyFunc(contract)
			
			err := validator.validateTemporalBehavior(contract)
			
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
	
	// Test all valid paradigms
	validParadigms := []string{"EVENT_DRIVEN", "SNAPSHOT", "HYBRID"}
	for _, paradigm := range validParadigms {
		t.Run("valid_paradigm_"+paradigm, func(t *testing.T) {
			contract := createValidTestContract()
			contract.TemporalBehavior.TemporalityParadigm = paradigm
			
			err := validator.validateTemporalBehavior(contract)
			assert.NoError(t, err)
		})
	}
	
	// Test all valid state models
	validStateModels := []string{"EVENT_DRIVEN", "STATE_MACHINE", "IMMUTABLE"}
	for _, model := range validStateModels {
		t.Run("valid_state_model_"+model, func(t *testing.T) {
			contract := createValidTestContract()
			contract.TemporalBehavior.StateTransitionModel = model
			
			err := validator.validateTemporalBehavior(contract)
			assert.NoError(t, err)
		})
	}
}

func TestValidator_validateRelationships(t *testing.T) {
	validator := NewValidator()
	
	testCases := []struct {
		name          string
		relationships []RelationshipDef
		expectError   bool
		errorContains string
	}{
		{
			name: "valid relationships",
			relationships: []RelationshipDef{
				{
					Name:         "test_relation",
					Type:         "one-to-one",
					TargetEntity: "target",
				},
			},
			expectError: false,
		},
		{
			name: "empty relationship name",
			relationships: []RelationshipDef{
				{
					Name:         "",
					Type:         "one-to-one",
					TargetEntity: "target",
				},
			},
			expectError:   true,
			errorContains: "relationship name is required",
		},
		{
			name: "invalid relationship type",
			relationships: []RelationshipDef{
				{
					Name:         "test_relation",
					Type:         "invalid-type",
					TargetEntity: "target",
				},
			},
			expectError:   true,
			errorContains: "invalid relationship type: invalid-type",
		},
		{
			name: "empty target entity",
			relationships: []RelationshipDef{
				{
					Name:         "test_relation",
					Type:         "one-to-one",
					TargetEntity: "",
				},
			},
			expectError:   true,
			errorContains: "target_entity is required for relationship: test_relation",
		},
		{
			name:          "empty relationships is valid",
			relationships: []RelationshipDef{},
			expectError:   false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			contract := createValidTestContract()
			contract.Relationships = tc.relationships
			
			err := validator.validateRelationships(contract)
			
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
	
	// Test all valid relationship types
	validRelationshipTypes := []string{"one-to-one", "one-to-many", "many-to-many"}
	for _, relType := range validRelationshipTypes {
		t.Run("valid_relationship_type_"+relType, func(t *testing.T) {
			contract := createValidTestContract()
			contract.Relationships = []RelationshipDef{
				{
					Name:         "test_relation",
					Type:         relType,
					TargetEntity: "target",
				},
			}
			
			err := validator.validateRelationships(contract)
			assert.NoError(t, err)
		})
	}
}

// Test helper functions
func TestValidator_isValidFieldName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid_simple", "test", true},
		{"valid_underscore", "test_field", true},
		{"valid_mixed_case", "TestField", true},
		{"valid_single_char", "a", true},
		{"valid_starts_underscore", "_test", true},
		{"valid_with_numbers", "test123", true},
		{"invalid_empty", "", false},
		{"invalid_starts_number", "123test", false},
		{"invalid_special_chars", "test-field", false},
		{"invalid_special_chars_space", "test field", false},
		{"invalid_special_chars_dot", "test.field", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidFieldName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidator_isValidFieldType(t *testing.T) {
	validTypes := []string{"string", "int", "int64", "float64", "bool", "time", "uuid", "enum", "json"}
	
	for _, validType := range validTypes {
		t.Run("valid_"+validType, func(t *testing.T) {
			result := isValidFieldType(validType)
			assert.True(t, result)
		})
	}
	
	invalidTypes := []string{"invalid", "number", "text", "datetime", ""}
	for _, invalidType := range invalidTypes {
		t.Run("invalid_"+invalidType, func(t *testing.T) {
			result := isValidFieldType(invalidType)
			assert.False(t, result)
		})
	}
}

func TestValidator_isValidDataClassification(t *testing.T) {
	validClassifications := []string{"PUBLIC", "INTERNAL", "CONFIDENTIAL", "RESTRICTED"}
	
	for _, valid := range validClassifications {
		t.Run("valid_"+valid, func(t *testing.T) {
			result := isValidDataClassification(valid)
			assert.True(t, result)
		})
		
		// Test case insensitive
		t.Run("valid_lowercase_"+valid, func(t *testing.T) {
			result := isValidDataClassification(strings.ToLower(valid))
			assert.True(t, result)
		})
	}
	
	invalidClassifications := []string{"INVALID", "SECRET", "PRIVATE", ""}
	for _, invalid := range invalidClassifications {
		t.Run("invalid_"+invalid, func(t *testing.T) {
			result := isValidDataClassification(invalid)
			assert.False(t, result)
		})
	}
}

// Benchmark tests
func BenchmarkValidator_Validate(b *testing.B) {
	validator := NewValidator()
	contract := createValidTestContract()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.Validate(contract)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidator_validateField(b *testing.B) {
	validator := NewValidator()
	field := FieldDefinition{
		Name:               "test_field",
		Type:               "string",
		Required:           true,
		DataClassification: "INTERNAL",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validator.validateField(field)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Integration test with complex contract
func TestValidator_Validate_ComplexContract(t *testing.T) {
	validator := NewValidator()
	
	contract := &MetaContract{
		SpecificationVersion: "2.0",
		Namespace:           "complex_test",
		ResourceName:        "complex_entity",
		Version:            "2.1.0",
		DataStructure: DataStructure{
			PrimaryKey:         "id",
			DataClassification: "CONFIDENTIAL",
			Fields: []FieldDefinition{
				{Name: "id", Type: "uuid", Required: true, Unique: true, DataClassification: "INTERNAL"},
				{Name: "name", Type: "string", Required: true, DataClassification: "INTERNAL"},
				{Name: "email", Type: "string", Required: true, Unique: true, DataClassification: "CONFIDENTIAL"},
				{Name: "age", Type: "int", Required: false, DataClassification: "INTERNAL"},
				{Name: "score", Type: "float64", Required: false, DataClassification: "INTERNAL"},
				{Name: "active", Type: "bool", Required: true, DataClassification: "INTERNAL"},
				{Name: "created_at", Type: "time", Required: true, DataClassification: "INTERNAL"},
				{Name: "metadata", Type: "json", Required: false, DataClassification: "INTERNAL"},
				{Name: "status", Type: "enum", Required: true, DataClassification: "INTERNAL"},
			},
		},
		SecurityModel: SecurityModel{
			TenantIsolation:    true,
			AccessControl:      "RBAC",
			DataClassification: "CONFIDENTIAL",
			ComplianceTags:     []string{"GDPR", "CCPA", "SOX"},
		},
		TemporalBehavior: TemporalBehaviorModel{
			TemporalityParadigm:  "HYBRID",
			StateTransitionModel: "STATE_MACHINE",
			HistoryRetention:     "10 years",
			EventDriven:          true,
		},
		APIBehavior: APIBehaviorModel{
			RESTEnabled:    true,
			GraphQLEnabled: true,
			EventsEnabled:  true,
		},
		Relationships: []RelationshipDef{
			{Name: "parent", Type: "one-to-one", TargetEntity: "parent_entity", Cardinality: "1:1", IsOptional: true},
			{Name: "children", Type: "one-to-many", TargetEntity: "child_entity", Cardinality: "1:N", IsOptional: true},
			{Name: "siblings", Type: "many-to-many", TargetEntity: "sibling_entity", Cardinality: "M:N", IsOptional: true},
		},
	}
	
	err := validator.Validate(contract)
	assert.NoError(t, err)
}