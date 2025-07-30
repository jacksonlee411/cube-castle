package metacontract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_NewParser(t *testing.T) {
	parser := NewParser()
	assert.NotNil(t, parser)
}

func TestParser_ParseMetaContract_Success(t *testing.T) {
	parser := NewParser()

	// Use the valid test file
	testFile := filepath.Join("testdata", "valid_person.yaml")

	contract, err := parser.ParseMetaContract(testFile)

	require.NoError(t, err)
	require.NotNil(t, contract)

	// Verify basic structure
	assert.Equal(t, "person", contract.ResourceName)
	assert.Equal(t, "hr.employees", contract.Namespace)
	assert.Equal(t, "1.0.0", contract.Version)
	assert.NotEmpty(t, contract.Description)

	// Verify data structure fields exist
	assert.NotNil(t, contract.DataStructure)
	assert.True(t, len(contract.DataStructure.Fields) > 0)

	// Find and verify ID field
	var idField *Field
	for _, field := range contract.DataStructure.Fields {
		if field.Name == "id" {
			idField = field
			break
		}
	}
	require.NotNil(t, idField, "ID field should exist")
	assert.Equal(t, "UUID", idField.Type)

	// Find and verify email field
	var emailField *Field
	for _, field := range contract.DataStructure.Fields {
		if field.Name == "email" {
			emailField = field
			break
		}
	}
	require.NotNil(t, emailField, "Email field should exist")
	assert.Equal(t, "String", emailField.Type)

	// Verify relationships
	assert.True(t, len(contract.Relationships) > 0)

	// Verify security model
	assert.Equal(t, "rbac", contract.SecurityModel.AccessControl)
	assert.Equal(t, "internal", contract.SecurityModel.DataClassification)

	// Verify temporal behavior
	assert.Equal(t, "event_driven", contract.TemporalBehavior.TemporalityParadigm)
	assert.Equal(t, "discrete", contract.TemporalBehavior.StateTransitionModel)
}

func TestParser_ParseMetaContract_FileNotFound(t *testing.T) {
	parser := NewParser()

	contract, err := parser.ParseMetaContract("nonexistent.yaml")

	assert.Error(t, err)
	assert.Nil(t, contract)
	assert.Contains(t, err.Error(), "failed to read meta-contract file")
}

func TestParser_ParseMetaContract_MalformedYAML(t *testing.T) {
	parser := NewParser()

	testFile := filepath.Join("testdata", "malformed.yaml")

	contract, err := parser.ParseMetaContract(testFile)

	assert.Error(t, err)
	assert.Nil(t, contract)
	assert.Contains(t, err.Error(), "failed to parse meta-contract YAML")
}

func TestParser_ParseMetaContract_ValidationFailure(t *testing.T) {
	parser := NewParser()

	testFile := filepath.Join("testdata", "invalid_contract.yaml")

	contract, err := parser.ParseMetaContract(testFile)

	assert.Error(t, err)
	assert.Nil(t, contract)
	assert.Contains(t, err.Error(), "meta-contract validation failed")
}

func TestParser_validateContract_Success(t *testing.T) {
	parser := NewParser()

	// Create a valid contract
	contract := createValidTestContract()

	err := parser.validateContract(contract)
	assert.NoError(t, err)
}

func TestParser_validateContract_MissingResourceName(t *testing.T) {
	parser := NewParser()

	contract := createValidTestContract()
	contract.ResourceName = ""

	err := parser.validateContract(contract)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resource_name is required")
}

func TestParser_validateContract_MissingNamespace(t *testing.T) {
	parser := NewParser()

	contract := createValidTestContract()
	contract.Namespace = ""

	err := parser.validateContract(contract)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "namespace is required")
}

func TestParser_validateContract_EmptyFields(t *testing.T) {
	parser := NewParser()

	contract := createValidTestContract()
	contract.DataStructure.Fields = []FieldDefinition{}

	err := parser.validateContract(contract)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one field must be defined")
}

func TestParser_validateContract_PrimaryKeyNotFound(t *testing.T) {
	parser := NewParser()

	contract := createValidTestContract()
	contract.DataStructure.PrimaryKey = "nonexistent_field"

	err := parser.validateContract(contract)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "primary_key field 'nonexistent_field' not found")
}

func TestParser_validateContract_InvalidFieldType(t *testing.T) {
	parser := NewParser()

	contract := createValidTestContract()
	contract.DataStructure.Fields[0].Type = "invalid_type"

	err := parser.validateContract(contract)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid field type 'invalid_type'")
}

func TestParser_validateContract_AllValidFieldTypes(t *testing.T) {
	parser := NewParser()

	validTypes := []string{"string", "int", "int64", "float64", "bool", "time", "uuid", "enum", "json"}

	for _, fieldType := range validTypes {
		t.Run(fieldType, func(t *testing.T) {
			contract := createValidTestContract()
			contract.DataStructure.Fields[0].Type = fieldType

			err := parser.validateContract(contract)
			assert.NoError(t, err, "Field type %s should be valid", fieldType)
		})
	}
}

// Benchmark tests
func BenchmarkParser_ParseMetaContract(b *testing.B) {
	parser := NewParser()
	testFile := filepath.Join("testdata", "valid_contract.yaml")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseMetaContract(testFile)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParser_validateContract(b *testing.B) {
	parser := NewParser()
	contract := createValidTestContract()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := parser.validateContract(contract)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions
func createValidTestContract() *MetaContract {
	return &MetaContract{
		SpecificationVersion: "1.0",
		Namespace:            "test",
		ResourceName:         "user",
		Version:              "1.0.0",
		DataStructure: DataStructure{
			PrimaryKey:         "id",
			DataClassification: "INTERNAL",
			Fields: []FieldDefinition{
				{
					Name:               "id",
					Type:               "uuid",
					Required:           true,
					Unique:             true,
					DataClassification: "INTERNAL",
				},
				{
					Name:               "email",
					Type:               "string",
					Required:           true,
					Unique:             true,
					DataClassification: "CONFIDENTIAL",
				},
			},
		},
		SecurityModel: SecurityModel{
			TenantIsolation:    true,
			AccessControl:      "RBAC",
			DataClassification: "CONFIDENTIAL",
			ComplianceTags:     []string{"GDPR"},
		},
		TemporalBehavior: TemporalBehaviorModel{
			TemporalityParadigm:  "EVENT_DRIVEN",
			StateTransitionModel: "EVENT_DRIVEN",
			HistoryRetention:     "7 years",
			EventDriven:          true,
		},
		APIBehavior: APIBehaviorModel{
			RESTEnabled:    true,
			GraphQLEnabled: true,
			EventsEnabled:  true,
		},
		Relationships: []RelationshipDef{
			{
				Name:         "user_profile",
				Type:         "one-to-one",
				TargetEntity: "profile",
				Cardinality:  "1:1",
				IsOptional:   false,
			},
		},
	}
}

// Table-driven tests for edge cases
func TestParser_validateContract_EdgeCases(t *testing.T) {
	testCases := []struct {
		name          string
		modifyFunc    func(*MetaContract)
		expectError   bool
		errorContains string
	}{
		{
			name: "empty primary key is valid",
			modifyFunc: func(c *MetaContract) {
				c.DataStructure.PrimaryKey = ""
			},
			expectError: false,
		},
		{
			name: "single character field names",
			modifyFunc: func(c *MetaContract) {
				c.DataStructure.Fields[0].Name = "a"
			},
			expectError: false,
		},
		{
			name: "underscore field names",
			modifyFunc: func(c *MetaContract) {
				c.DataStructure.Fields[0].Name = "_test_field"
			},
			expectError: false,
		},
		{
			name: "mixed case field names",
			modifyFunc: func(c *MetaContract) {
				c.DataStructure.Fields[0].Name = "TestField"
			},
			expectError: false,
		},
		{
			name: "field with all optional properties",
			modifyFunc: func(c *MetaContract) {
				c.DataStructure.Fields = append(c.DataStructure.Fields, FieldDefinition{
					Name:     "optional_field",
					Type:     "string",
					Required: false,
					Unique:   false,
				})
			},
			expectError: false,
		},
	}

	parser := NewParser()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			contract := createValidTestContract()
			tc.modifyFunc(contract)

			err := parser.validateContract(contract)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test with temporary files
func TestParser_ParseMetaContract_TemporaryFile(t *testing.T) {
	parser := NewParser()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_contract_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write valid YAML content
	content := `specification_version: "1.0"
namespace: "temp_test"
resource_name: "temp_resource"
version: "1.0.0"
data_structure:
  fields:
    - name: "id"
      type: "uuid"
      required: true`

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	// Parse the temporary file
	contract, err := parser.ParseMetaContract(tmpFile.Name())

	require.NoError(t, err)
	assert.Equal(t, "temp_test", contract.Namespace)
	assert.Equal(t, "temp_resource", contract.ResourceName)
}
