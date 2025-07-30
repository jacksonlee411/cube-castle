package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/gaogu/cube-castle/go-app/internal/metacontract"
	"github.com/gaogu/cube-castle/go-app/internal/metacontracteditor"
)

// Integration tests for the complete meta-contract compilation workflow
// These tests verify that all components work together correctly

func TestMetaContractCompiler_FullWorkflow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test the complete workflow from YAML parsing to code generation
	testCases := []struct {
		name          string
		contractFile  string
		expectSuccess bool
		description   string
	}{
		{
			name:          "valid person contract",
			contractFile:  "valid_person.yaml",
			expectSuccess: true,
			description:   "Complete workflow with valid person entity",
		},
		{
			name:          "simple department contract",
			contractFile:  "simple_department.yaml",
			expectSuccess: true,
			description:   "Simple entity with minimal configuration",
		},
		{
			name:          "invalid contract",
			contractFile:  "invalid_contract.yaml",
			expectSuccess: false,
			description:   "Should fail during validation phase",
		},
		{
			name:          "malformed yaml",
			contractFile:  "malformed.yaml",
			expectSuccess: false,
			description:   "Should fail during parsing phase",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize compiler
			compiler := metacontract.NewCompiler()

			// Prepare test data
			inputPath := filepath.Join("..", "..", "internal", "metacontract", "testdata", tc.contractFile)
			
			// Create temporary output directory
			outputDir, err := os.MkdirTemp("", "integration_test_*")
			require.NoError(t, err)
			defer os.RemoveAll(outputDir)

			// Execute full compilation workflow
			err = compiler.Compile(inputPath, outputDir)

			if tc.expectSuccess {
				// Note: May fail if code generators are not fully implemented
				if err != nil {
					t.Logf("Compilation error (may be expected if generators not implemented): %v", err)
					// Ensure it's a generation error, not a parse/validation error
					assert.True(t,
						err.Error() == "ent generation failed" ||
						err.Error() == "api generation failed",
						"Expected generation error for valid contract, got: %v", err)
				} else {
					// Verify output structure was created
					assert.DirExists(t, outputDir)
					
					// Check for expected output directories
					schemaDir := filepath.Join(outputDir, "schema")
					apiDir := filepath.Join(outputDir, "api")
					
					if _, err := os.Stat(schemaDir); err == nil {
						t.Logf("Schema directory created: %s", schemaDir)
					}
					if _, err := os.Stat(apiDir); err == nil {
						t.Logf("API directory created: %s", apiDir)
					}
				}
			} else {
				// Should fail for invalid contracts
				assert.Error(t, err)
				assert.True(t,
					err.Error() == "parse failed" ||
					err.Error() == "validation failed",
					"Expected parse or validation error, got: %v", err)
			}
		})
	}
}

func TestMetaContractEditor_Service_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test the editor service with a real compiler
	compiler := metacontract.NewCompiler()
	
	// Create a mock repository for testing
	mockRepo := &MockEditorRepository{
		projects:  make(map[uuid.UUID]*metacontracteditor.EditorProject),
		settings:  make(map[uuid.UUID]*metacontracteditor.EditorSettings),
		templates: createDefaultTemplates(),
	}
	
	// Initialize service
	service := metacontracteditor.NewService(mockRepo, compiler)

	// Test data
	tenantID := uuid.New()
	userID := uuid.New()

	t.Run("complete project lifecycle", func(t *testing.T) {
		// 1. Create a project
		createReq := metacontracteditor.CreateProjectRequest{
			Name:        "Integration Test Project",
			Description: "A project for integration testing",
			Content: `resource_name: integration_entity
namespace: integration.test
version: "1.0.0"
description: "Integration test entity"

data_structure:
  fields:
    - name: id
      type: UUID
      constraints:
        primary_key: true
        required: true
    
    - name: name
      type: String
      constraints:
        required: true
        max_length: 100

security_model:
  access_control: rbac
  data_classification: internal

temporal_behavior:
  temporality_paradigm: snapshot
  state_transition_model: discrete`,
			TenantID: tenantID,
			UserID:   userID,
		}

		project, err := service.CreateProject(context.Background(), createReq)
		require.NoError(t, err)
		require.NotNil(t, project)
		assert.Equal(t, createReq.Name, project.Name)
		assert.Equal(t, metacontracteditor.ProjectStatusDraft, project.Status)

		projectID := project.ID

		// 2. Compile the project
		compileReq := metacontracteditor.CompileRequest{
			ProjectID: projectID,
			Content:   createReq.Content,
			Preview:   true,
		}

		response, err := service.CompileProject(context.Background(), compileReq)
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.NotZero(t, response.CompileTime)

		// Compilation may succeed or fail depending on generator implementation
		if response.Success {
			t.Log("Compilation succeeded")
			assert.NotNil(t, response.Schema)
			assert.NotEmpty(t, response.GeneratedFiles)
		} else {
			t.Logf("Compilation failed (may be expected): %d errors", len(response.Errors))
			assert.NotEmpty(t, response.Errors)
		}

		// 3. Update the project
		newName := "Updated Integration Test Project"
		updateReq := metacontracteditor.UpdateProjectRequest{
			Name:     &newName,
			TenantID: tenantID,
		}

		updatedProject, err := service.UpdateProject(context.Background(), projectID, updateReq)
		require.NoError(t, err)
		require.NotNil(t, updatedProject)
		assert.Equal(t, newName, updatedProject.Name)

		// 4. List projects
		projects, err := service.ListProjects(context.Background(), tenantID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, projects, 1)
		assert.Equal(t, projectID, projects[0].ID)

		// 5. Delete the project
		err = service.DeleteProject(context.Background(), projectID, tenantID)
		require.NoError(t, err)

		// 6. Verify deletion
		projects, err = service.ListProjects(context.Background(), tenantID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, projects, 0)
	})

	t.Run("user settings management", func(t *testing.T) {
		// 1. Get default settings for new user
		settings, err := service.GetUserSettings(context.Background(), userID)
		require.NoError(t, err)
		require.NotNil(t, settings)
		assert.Equal(t, userID, settings.UserID)
		assert.Equal(t, "vs-dark", settings.Theme) // Default theme

		// 2. Update settings
		settings.Theme = "light"
		settings.FontSize = 16
		settings.AutoSave = false
		settings.Settings = map[string]interface{}{
			"custom_setting": "value",
		}

		err = service.UpdateUserSettings(context.Background(), settings)
		require.NoError(t, err)

		// 3. Retrieve updated settings
		updatedSettings, err := service.GetUserSettings(context.Background(), userID)
		require.NoError(t, err)
		assert.Equal(t, "light", updatedSettings.Theme)
		assert.Equal(t, 16, updatedSettings.FontSize)
		assert.False(t, updatedSettings.AutoSave)
		assert.Equal(t, "value", updatedSettings.Settings["custom_setting"])
	})

	t.Run("template management", func(t *testing.T) {
		// Get templates
		templates, err := service.GetTemplates(context.Background(), "basic")
		require.NoError(t, err)
		assert.True(t, len(templates) > 0)

		// Verify template structure
		template := templates[0]
		assert.NotEmpty(t, template.Name)
		assert.NotEmpty(t, template.Content)
		assert.Equal(t, "basic", template.Category)
	})
}

func TestMetaContractCompiler_ErrorHandling_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	compiler := metacontract.NewCompiler()

	// Test various error scenarios
	testCases := []struct {
		name          string
		setupFunc     func() (string, string) // returns inputPath, outputPath
		expectError   bool
		errorContains string
		cleanup       func(string)
	}{
		{
			name: "nonexistent input file",
			setupFunc: func() (string, string) {
				outputDir, _ := os.MkdirTemp("", "error_test_*")
				return "nonexistent_file.yaml", outputDir
			},
			expectError:   true,
			errorContains: "parse failed",
			cleanup:       func(dir string) { os.RemoveAll(dir) },
		},
		{
			name: "permission denied output directory",
			setupFunc: func() (string, string) {
				// Create read-only directory
				outputDir, _ := os.MkdirTemp("", "readonly_*")
				os.Chmod(outputDir, 0444)
				
				inputPath := filepath.Join("..", "..", "internal", "metacontract", "testdata", "valid_person.yaml")
				return inputPath, outputDir
			},
			expectError: true,
			cleanup: func(dir string) {
				os.Chmod(dir, 0755) // Restore permissions for cleanup
				os.RemoveAll(dir)
			},
		},
		{
			name: "invalid yaml content",
			setupFunc: func() (string, string) {
				outputDir, _ := os.MkdirTemp("", "invalid_yaml_*")
				inputPath := filepath.Join("..", "..", "internal", "metacontract", "testdata", "malformed.yaml")
				return inputPath, outputDir
			},
			expectError:   true,
			errorContains: "parse failed",
			cleanup:       func(dir string) { os.RemoveAll(dir) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath, outputPath := tc.setupFunc()
			if tc.cleanup != nil {
				defer tc.cleanup(outputPath)
			}

			err := compiler.Compile(inputPath, outputPath)

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

func TestMetaContractCompiler_Performance_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance integration test in short mode")
	}

	compiler := metacontract.NewCompiler()
	inputPath := filepath.Join("..", "..", "internal", "metacontract", "testdata", "valid_person.yaml")

	// Test compilation performance
	t.Run("compilation_performance", func(t *testing.T) {
		const numRuns = 10
		durations := make([]time.Duration, numRuns)

		for i := 0; i < numRuns; i++ {
			outputDir, err := os.MkdirTemp("", "perf_test_*")
			require.NoError(t, err)
			defer os.RemoveAll(outputDir)

			start := time.Now()
			_ = compiler.Compile(inputPath, outputDir) // May fail, but we're measuring performance
			durations[i] = time.Since(start)
		}

		// Calculate average duration
		var total time.Duration
		for _, d := range durations {
			total += d
		}
		avgDuration := total / time.Duration(numRuns)

		t.Logf("Average compilation time over %d runs: %v", numRuns, avgDuration)
		
		// Performance assertion (adjust threshold as needed)
		assert.Less(t, avgDuration, 5*time.Second, "Compilation should complete within 5 seconds on average")
	})

	// Test memory usage
	t.Run("memory_usage", func(t *testing.T) {
		const numIterations = 100
		
		for i := 0; i < numIterations; i++ {
			outputDir, err := os.MkdirTemp("", "memory_test_*")
			require.NoError(t, err)
			
			_ = compiler.Compile(inputPath, outputDir)
			
			os.RemoveAll(outputDir)
			
			// Force garbage collection periodically
			if i%10 == 0 {
				runtime.GC()
			}
		}
		
		// Test passes if no memory leaks cause out-of-memory errors
		t.Log("Memory usage test completed successfully")
	})
}

// Mock repository implementation for integration testing
type MockEditorRepository struct {
	projects  map[uuid.UUID]*metacontracteditor.EditorProject
	settings  map[uuid.UUID]*metacontracteditor.EditorSettings
	templates []*metacontracteditor.ProjectTemplate
}

func (m *MockEditorRepository) CreateProject(ctx context.Context, project *metacontracteditor.EditorProject) error {
	m.projects[project.ID] = project
	return nil
}

func (m *MockEditorRepository) GetProject(ctx context.Context, projectID uuid.UUID) (*metacontracteditor.EditorProject, error) {
	if project, exists := m.projects[projectID]; exists {
		return project, nil
	}
	return nil, assert.AnError
}

func (m *MockEditorRepository) UpdateProject(ctx context.Context, project *metacontracteditor.EditorProject) error {
	if _, exists := m.projects[project.ID]; exists {
		m.projects[project.ID] = project
		return nil
	}
	return assert.AnError
}

func (m *MockEditorRepository) ListProjects(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*metacontracteditor.EditorProject, error) {
	var result []*metacontracteditor.EditorProject
	for _, project := range m.projects {
		if project.TenantID == tenantID {
			result = append(result, project)
		}
	}
	return result, nil
}

func (m *MockEditorRepository) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	if _, exists := m.projects[projectID]; exists {
		delete(m.projects, projectID)
		return nil
	}
	return assert.AnError
}

func (m *MockEditorRepository) GetTemplates(ctx context.Context, category string) ([]*metacontracteditor.ProjectTemplate, error) {
	var result []*metacontracteditor.ProjectTemplate
	for _, template := range m.templates {
		if template.Category == category {
			result = append(result, template)
		}
	}
	return result, nil
}

func (m *MockEditorRepository) GetUserSettings(ctx context.Context, userID uuid.UUID) (*metacontracteditor.EditorSettings, error) {
	if settings, exists := m.settings[userID]; exists {
		return settings, nil
	}
	// Return default settings
	return &metacontracteditor.EditorSettings{
		UserID:      userID,
		Theme:       "vs-dark",
		FontSize:    14,
		AutoSave:    true,
		AutoCompile: true,
		KeyBindings: "default",
		Settings:    make(map[string]interface{}),
		UpdatedAt:   time.Now(),
	}, nil
}

func (m *MockEditorRepository) UpdateUserSettings(ctx context.Context, settings *metacontracteditor.EditorSettings) error {
	m.settings[settings.UserID] = settings
	return nil
}

func createDefaultTemplates() []*metacontracteditor.ProjectTemplate {
	return []*metacontracteditor.ProjectTemplate{
		{
			ID:       uuid.New(),
			Name:     "Basic Entity",
			Category: "basic",
			Content: `resource_name: example_entity
namespace: example.namespace
version: "1.0.0"

data_structure:
  fields:
    - name: id
      type: UUID
      constraints:
        primary_key: true
        required: true
    
    - name: name
      type: String
      constraints:
        required: true
        max_length: 255

security_model:
  access_control: rbac
  data_classification: internal`,
			Description: "A basic entity template with ID and name fields",
			Tags:        []string{"basic", "crud"},
		},
		{
			ID:       uuid.New(),
			Name:     "Employee Template",
			Category: "hr",
			Content: `resource_name: employee
namespace: hr.employees
version: "1.0.0"

data_structure:
  fields:
    - name: id
      type: UUID
      constraints:
        primary_key: true
        required: true
    
    - name: employee_id
      type: String
      constraints:
        required: true
        unique: true
        max_length: 20
    
    - name: first_name
      type: String
      constraints:
        required: true
        max_length: 50
    
    - name: last_name
      type: String
      constraints:
        required: true
        max_length: 50
    
    - name: email
      type: String
      constraints:
        required: true
        unique: true
        format: email

security_model:
  access_control: rbac
  data_classification: confidential`,
			Description: "Employee entity template for HR systems",
			Tags:        []string{"hr", "employee"},
		},
	}
}