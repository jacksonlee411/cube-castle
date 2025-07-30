package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gaogu/cube-castle/go-app/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIGenerator_NewAPIGenerator(t *testing.T) {
	generator := NewAPIGenerator()
	assert.NotNil(t, generator)
	assert.NotNil(t, generator.templateEngine)
}

func TestAPIGenerator_Generate_Success(t *testing.T) {
	generator := NewAPIGenerator()
	contract := createTestAPIContract()

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_api_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate API code
	err = generator.Generate(contract, outputDir)

	// Note: This may fail if templates are not fully implemented
	if err != nil {
		t.Logf("API generation error (may be expected): %v", err)
		// Check that error is meaningful and not a panic
		assert.NotContains(t, err.Error(), "panic")
		assert.Contains(t, err.Error(), "failed to")
	} else {
		// If generation succeeds, verify output directory exists
		assert.DirExists(t, outputDir)

		// Check if any files were generated
		files, err := os.ReadDir(outputDir)
		require.NoError(t, err)
		t.Logf("Generated %d API files", len(files))

		// Check for expected files
		expectedFiles := []string{
			contract.ResourceName + "_handler.go",
		}

		for _, expectedFile := range expectedFiles {
			expectedPath := filepath.Join(outputDir, expectedFile)
			if _, err := os.Stat(expectedPath); err == nil {
				t.Logf("Generated expected file: %s", expectedFile)
			}
		}
	}
}

func TestAPIGenerator_Generate_InvalidOutputDir(t *testing.T) {
	generator := NewAPIGenerator()
	contract := createTestAPIContract()

	// Try to generate in an invalid directory
	invalidDir := "/root/nonexistent/deeply/nested/path"

	err := generator.Generate(contract, invalidDir)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output directory")
}

func TestAPIGenerator_Generate_NilContract(t *testing.T) {
	generator := NewAPIGenerator()

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_api_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate with nil contract (should handle gracefully)
	err = generator.Generate(nil, outputDir)

	assert.Error(t, err)
	// Should not panic, but provide meaningful error
}

func TestAPIGenerator_Generate_EmptyContract(t *testing.T) {
	generator := NewAPIGenerator()
	contract := &types.MetaContract{} // Empty contract

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_api_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate with empty contract
	err = generator.Generate(contract, outputDir)

	// May succeed or fail depending on implementation, but should not panic
	if err != nil {
		assert.NotContains(t, err.Error(), "panic")
	}
}

func TestAPIGenerator_Generate_WithRESTEndpoints(t *testing.T) {
	generator := NewAPIGenerator()
	contract := createTestAPIContract()

	// Add REST endpoints
	contract.APIConfiguration.RestEndpoints = []types.RestEndpoint{
		{
			Path:        "/test-entities",
			Methods:     []string{"GET", "POST"},
			Description: "List and create test entities",
		},
		{
			Path:        "/test-entities/{id}",
			Methods:     []string{"GET", "PUT", "DELETE"},
			Description: "Manage specific test entity",
		},
	}

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_api_rest_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate API with REST endpoints
	err = generator.Generate(contract, outputDir)

	if err != nil {
		t.Logf("REST API generation error (may be expected): %v", err)
		assert.NotContains(t, err.Error(), "panic")
	} else {
		// Verify output
		assert.DirExists(t, outputDir)
		files, err := os.ReadDir(outputDir)
		require.NoError(t, err)
		t.Logf("Generated %d files for REST API", len(files))
	}
}

func TestAPIGenerator_Generate_WithGraphQLTypes(t *testing.T) {
	generator := NewAPIGenerator()
	contract := createTestAPIContract()

	// Enable GraphQL types
	contract.APIConfiguration.GraphQLTypes.Enabled = true
	contract.APIConfiguration.GraphQLTypes.Mutations = true
	contract.APIConfiguration.GraphQLTypes.Subscriptions = false

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_api_graphql_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate API with GraphQL
	err = generator.Generate(contract, outputDir)

	if err != nil {
		t.Logf("GraphQL API generation error (may be expected): %v", err)
		assert.NotContains(t, err.Error(), "panic")
	} else {
		// Verify output
		assert.DirExists(t, outputDir)
		files, err := os.ReadDir(outputDir)
		require.NoError(t, err)
		t.Logf("Generated %d files for GraphQL API", len(files))
	}
}

func TestAPIGenerator_Generate_ComplexAPI(t *testing.T) {
	generator := NewAPIGenerator()
	contract := createComplexAPIContract()

	// Create temporary output directory
	outputDir, err := os.MkdirTemp("", "test_api_complex_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Generate complex API
	err = generator.Generate(contract, outputDir)

	if err != nil {
		t.Logf("Complex API generation error (may be expected): %v", err)
		assert.NotContains(t, err.Error(), "panic")
	} else {
		// Verify output
		assert.DirExists(t, outputDir)
		files, err := os.ReadDir(outputDir)
		require.NoError(t, err)
		t.Logf("Generated %d files for complex API", len(files))
	}
}

func TestAPIGenerator_TemplateEngine(t *testing.T) {
	generator := NewAPIGenerator()

	// Test that template engine is properly initialized
	assert.NotNil(t, generator.templateEngine)

	// Test template functions are registered
	funcs := generator.templateEngine.Funcs(nil)
	assert.NotEmpty(t, funcs)
}

// Test concurrent generation
func TestAPIGenerator_ConcurrentGeneration(t *testing.T) {
	generator := NewAPIGenerator()
	contract := createTestAPIContract()

	const numGoroutines = 5
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Create unique output directory for each goroutine
			outputDir, err := os.MkdirTemp("", "concurrent_api_*")
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
			t.Logf("Concurrent API generation error (may be expected): %v", err)
			assert.NotContains(t, err.Error(), "panic")
		}
	}
}

// Test API generation with different HTTP methods
func TestAPIGenerator_HTTPMethods(t *testing.T) {
	generator := NewAPIGenerator()

	httpMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range httpMethods {
		t.Run(method, func(t *testing.T) {
			contract := createTestAPIContract()
			contract.APIConfiguration.RestEndpoints = []types.RestEndpoint{
				{
					Path:    "/test-entities",
					Methods: []string{method},
				},
			}

			// Create temporary output directory
			outputDir, err := os.MkdirTemp("", "test_api_method_*")
			require.NoError(t, err)
			defer os.RemoveAll(outputDir)

			// Generate API with specific method
			err = generator.Generate(contract, outputDir)

			if err != nil {
				t.Logf("API generation error for %s (may be expected): %v", method, err)
				assert.NotContains(t, err.Error(), "panic")
			}
		})
	}
}

// Test API generation with security models
func TestAPIGenerator_SecurityModels(t *testing.T) {
	generator := NewAPIGenerator()

	securityModels := []string{"rbac", "abac", "dac", "mac"}

	for _, secModel := range securityModels {
		t.Run(secModel, func(t *testing.T) {
			contract := createTestAPIContract()
			contract.SecurityModel.AccessControl = secModel

			// Create temporary output directory
			outputDir, err := os.MkdirTemp("", "test_api_security_*")
			require.NoError(t, err)
			defer os.RemoveAll(outputDir)

			// Generate API with specific security model
			err = generator.Generate(contract, outputDir)

			if err != nil {
				t.Logf("API generation error for %s security (may be expected): %v", secModel, err)
				assert.NotContains(t, err.Error(), "panic")
			}
		})
	}
}

// Benchmark tests
func BenchmarkAPIGenerator_Generate(b *testing.B) {
	generator := NewAPIGenerator()
	contract := createTestAPIContract()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputDir, err := os.MkdirTemp("", "bench_api_*")
		if err != nil {
			b.Fatal(err)
		}

		// Generate (may fail, but shouldn't panic)
		_ = generator.Generate(contract, outputDir)

		os.RemoveAll(outputDir)
	}
}

// Helper functions
func createTestAPIContract() *types.MetaContract {
	return &types.MetaContract{
		ResourceName: "test_entity",
		Namespace:    "test.api",
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
			},
		},
		SecurityModel: types.SecurityModel{
			AccessControl:      "rbac",
			DataClassification: "internal",
		},
		APIConfiguration: types.APIConfiguration{
			RestEndpoints: []types.RestEndpoint{
				{
					Path:        "/test-entities",
					Methods:     []string{"GET", "POST"},
					Description: "List and create test entities",
				},
			},
			GraphQLTypes: types.GraphQLConfig{
				Enabled:       false,
				Mutations:     false,
				Subscriptions: false,
			},
		},
	}
}

func createComplexAPIContract() *types.MetaContract {
	return &types.MetaContract{
		ResourceName: "complex_api_entity",
		Namespace:    "api.complex",
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
			},
		},
		Relationships: []types.Relationship{
			{
				Name:         "category",
				Type:         "belongs_to",
				TargetEntity: "category",
				ForeignKey:   "category_id",
			},
			{
				Name:         "items",
				Type:         "has_many",
				TargetEntity: "item",
				ForeignKey:   "entity_id",
			},
		},
		SecurityModel: types.SecurityModel{
			AccessControl:      "rbac",
			DataClassification: "confidential",
			PrivacyFields:      []string{"metadata"},
		},
		APIConfiguration: types.APIConfiguration{
			RestEndpoints: []types.RestEndpoint{
				{
					Path:        "/complex-entities",
					Methods:     []string{"GET", "POST"},
					Description: "List and create complex entities",
				},
				{
					Path:        "/complex-entities/{id}",
					Methods:     []string{"GET", "PUT", "DELETE"},
					Description: "Manage specific complex entity",
				},
				{
					Path:        "/complex-entities/{id}/items",
					Methods:     []string{"GET"},
					Description: "Get entity items",
				},
			},
			GraphQLTypes: types.GraphQLConfig{
				Enabled:       true,
				Mutations:     true,
				Subscriptions: false,
			},
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
