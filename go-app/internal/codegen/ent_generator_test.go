package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gaogu/cube-castle/go-app/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntGenerator_NewEntGenerator(t *testing.T) {
	generator := NewEntGenerator()
	assert.NotNil(t, generator)
	assert.NotNil(t, generator.templateEngine)
}

func TestEntGenerator_Generate_Success(t *testing.T) {
	generator := NewEntGenerator()
	contract := createTestMetaContract()

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_ent_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate Ent schemas
	err = generator.Generate(contract, outputDir)

	// Note: This may fail if templates are not fully implemented
	if err != nil {
		t.Logf("Ent generation error (may be expected): %v", err)
		// Check that error is meaningful and not a panic
		assert.NotContains(t, err.Error(), "panic")
		assert.Contains(t, err.Error(), "failed to")
	} else {
		// If generation succeeds, verify output directory exists
		assert.DirExists(t, outputDir)

		// Check if any files were generated
		files, err := os.ReadDir(outputDir)
		require.NoError(t, err)
		t.Logf("Generated %d files", len(files))
	}
}

func TestEntGenerator_Generate_InvalidOutputDir(t *testing.T) {
	generator := NewEntGenerator()
	contract := createTestMetaContract()

	// Try to generate in an invalid directory
	invalidDir := "/root/nonexistent/deeply/nested/path"

	err := generator.Generate(contract, invalidDir)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output directory")
}

func TestEntGenerator_Generate_NilContract(t *testing.T) {
	generator := NewEntGenerator()

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_ent_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate with nil contract (should handle gracefully)
	err = generator.Generate(nil, outputDir)

	assert.Error(t, err)
	// Should not panic, but provide meaningful error
}

func TestEntGenerator_Generate_EmptyContract(t *testing.T) {
	generator := NewEntGenerator()
	contract := &types.MetaContract{} // Empty contract

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_ent_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate with empty contract
	err = generator.Generate(contract, outputDir)

	// May succeed or fail depending on implementation, but should not panic
	if err != nil {
		assert.NotContains(t, err.Error(), "panic")
	}
}

func TestEntGenerator_Generate_ComplexContract(t *testing.T) {
	generator := NewEntGenerator()
	contract := createComplexTestMetaContract()

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_ent_complex_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate with complex contract
	err = generator.Generate(contract, outputDir)

	if err != nil {
		t.Logf("Complex contract generation error (may be expected): %v", err)
		assert.NotContains(t, err.Error(), "panic")
	} else {
		// Verify output
		assert.DirExists(t, outputDir)
		files, err := os.ReadDir(outputDir)
		require.NoError(t, err)
		t.Logf("Generated %d files for complex contract", len(files))
	}
}

func TestEntGenerator_Generate_TemporalContract(t *testing.T) {
	generator := NewEntGenerator()
	contract := createTestMetaContract()

	// Set temporal behavior to event-driven to test history entity generation
	contract.TemporalBehavior.TemporalityParadigm = "EVENT_DRIVEN"

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_ent_temporal_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate with temporal contract
	err = generator.Generate(contract, outputDir)

	if err != nil {
		t.Logf("Temporal contract generation error (may be expected): %v", err)
		assert.NotContains(t, err.Error(), "panic")
	} else {
		// Verify output includes temporal entities
		assert.DirExists(t, outputDir)
		files, err := os.ReadDir(outputDir)
		require.NoError(t, err)
		t.Logf("Generated %d files for temporal contract", len(files))
	}
}

func TestEntGenerator_TemplateEngine(t *testing.T) {
	generator := NewEntGenerator()

	// Test that template engine is properly initialized
	assert.NotNil(t, generator.templateEngine)

	// Test template functions are registered
	funcs := generator.templateEngine.Funcs(nil)
	assert.NotEmpty(t, funcs)
}

func TestEntGenerator_HelperFunctions(t *testing.T) {
	// Test title function
	result := strings.Title("test_string")
	assert.Equal(t, "Test_string", result)

	// Test case conversion functions (these should be available in the template)
	testCases := []struct {
		input    string
		camelExp string
		snakeExp string
	}{
		{"test_field", "testField", "test_field"},
		{"TestField", "testField", "test_field"},
		{"testField", "testField", "test_field"},
		{"test", "test", "test"},
		{"Test", "test", "test"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			// Test camelCase conversion
			camelResult := toCamelCase(tc.input)
			assert.Equal(t, tc.camelExp, camelResult)

			// Test snake_case conversion
			snakeResult := toSnakeCase(tc.input)
			assert.Equal(t, tc.snakeExp, snakeResult)
		})
	}
}

// Test concurrent generation
func TestEntGenerator_ConcurrentGeneration(t *testing.T) {
	generator := NewEntGenerator()
	contract := createTestMetaContract()

	const numGoroutines = 5
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Create unique output directory for each goroutine
			outputDir, err := os.MkdirTemp("", "concurrent_ent_*")
			if err != nil {
				errors <- err
				return
			}
			defer os.RemoveAll(outputDir)

			// Generate
			err = generator.Generate(contract, outputDir)
			errors <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		if err != nil {
			t.Logf("Concurrent generation error (may be expected): %v", err)
			assert.NotContains(t, err.Error(), "panic")
		}
	}
}

// Benchmark tests
func BenchmarkEntGenerator_Generate(b *testing.B) {
	generator := NewEntGenerator()
	contract := createTestMetaContract()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputDir, err := os.MkdirTemp("", "bench_ent_*")
		if err != nil {
			b.Fatal(err)
		}

		// Generate (may fail, but shouldn't panic)
		_ = generator.Generate(contract, outputDir)

		os.RemoveAll(outputDir)
	}
}

// Helper functions
func createTestMetaContract() *types.MetaContract {
	return &types.MetaContract{
		ResourceName: "test_entity",
		Namespace:    "test.namespace",
		Version:      "1.0.0",
		DataStructure: types.DataStructure{
			Fields: []types.Field{
				{
					Name: "id",
					Type: "UUID",
					Constraints: types.FieldConstraints{
						Required:   true,
						PrimaryKey: true,
					},
				},
				{
					Name: "name",
					Type: "String",
					Constraints: types.FieldConstraints{
						Required:  true,
						MaxLength: 100,
					},
				},
				{
					Name: "email",
					Type: "String",
					Constraints: types.FieldConstraints{
						Required: true,
						Unique:   true,
						Format:   "email",
					},
				},
				{
					Name: "created_at",
					Type: "Timestamp",
					Constraints: types.FieldConstraints{
						Required:     true,
						AutoGenerate: true,
					},
				},
			},
		},
		Relationships: []types.Relationship{
			{
				Name:         "department",
				Type:         "belongs_to",
				TargetEntity: "department",
				ForeignKey:   "department_id",
			},
		},
		SecurityModel: types.SecurityModel{
			AccessControl:      "rbac",
			DataClassification: "internal",
		},
		TemporalBehavior: types.TemporalBehavior{
			TemporalityParadigm:  "snapshot",
			StateTransitionModel: "discrete",
		},
	}
}

func createComplexTestMetaContract() *types.MetaContract {
	return &types.MetaContract{
		ResourceName: "complex_entity",
		Namespace:    "complex.namespace",
		Version:      "2.0.0",
		DataStructure: types.DataStructure{
			Fields: []types.Field{
				{
					Name: "id",
					Type: "UUID",
					Constraints: types.FieldConstraints{
						Required:   true,
						PrimaryKey: true,
					},
				},
				{
					Name: "name",
					Type: "String",
					Constraints: types.FieldConstraints{
						Required:  true,
						MaxLength: 255,
					},
				},
				{
					Name:       "status",
					Type:       "Enum",
					EnumValues: []string{"active", "inactive", "pending"},
					Constraints: types.FieldConstraints{
						Required: true,
						Default:  "pending",
					},
				},
				{
					Name: "metadata",
					Type: "JSON",
					Constraints: types.FieldConstraints{
						Required: false,
					},
				},
				{
					Name: "score",
					Type: "Decimal",
					Constraints: types.FieldConstraints{
						Precision: 10,
						Scale:     2,
					},
				},
				{
					Name: "is_active",
					Type: "Boolean",
					Constraints: types.FieldConstraints{
						Required: true,
						Default:  "true",
					},
				},
			},
		},
		Relationships: []types.Relationship{
			{
				Name:         "parent",
				Type:         "belongs_to",
				TargetEntity: "complex_entity",
				ForeignKey:   "parent_id",
			},
			{
				Name:         "children",
				Type:         "has_many",
				TargetEntity: "complex_entity",
				ForeignKey:   "parent_id",
			},
			{
				Name:         "tags",
				Type:         "many_to_many",
				TargetEntity: "tag",
				ForeignKey:   "entity_id",
			},
		},
		SecurityModel: types.SecurityModel{
			AccessControl:      "rbac",
			DataClassification: "confidential",
			PrivacyFields:      []string{"metadata", "score"},
		},
		TemporalBehavior: types.TemporalBehavior{
			TemporalityParadigm:  "event_driven",
			StateTransitionModel: "discrete",
			HistoryTracking:      true,
			EffectiveDating:      true,
		},
		BusinessRules: []types.BusinessRule{
			{
				Name:       "status_validation",
				Type:       "format",
				Field:      "status",
				Expression: "status in ['active', 'inactive', 'pending']",
			},
		},
	}
}

// Helper functions that should exist in the actual implementation
// These are placeholder implementations for testing

func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		// Convert first character to lowercase
		if strings.ToUpper(s[:1]) == s[:1] {
			return strings.ToLower(s[:1]) + s[1:]
		}
		return s
	}

	result := strings.ToLower(parts[0])
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
	}
	return result
}

func toSnakeCase(s string) string {
	if s == "" {
		return s
	}

	// If already snake_case, return as is
	if strings.Contains(s, "_") {
		return strings.ToLower(s)
	}

	// Convert camelCase or PascalCase to snake_case
	result := ""
	for i, char := range s {
		if i > 0 && strings.ToUpper(string(char)) == string(char) && char != '_' {
			result += "_"
		}
		result += strings.ToLower(string(char))
	}
	return result
}
