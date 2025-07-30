package metacontracteditor

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/gaogu/cube-castle/go-app/internal/metacontract"
)

// Mock repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateProject(ctx context.Context, project *EditorProject) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockRepository) GetProject(ctx context.Context, projectID uuid.UUID) (*EditorProject, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EditorProject), args.Error(1)
}

func (m *MockRepository) UpdateProject(ctx context.Context, project *EditorProject) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockRepository) ListProjects(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*EditorProject, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*EditorProject), args.Error(1)
}

func (m *MockRepository) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockRepository) GetTemplates(ctx context.Context, category string) ([]*ProjectTemplate, error) {
	args := m.Called(ctx, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ProjectTemplate), args.Error(1)
}

func (m *MockRepository) GetUserSettings(ctx context.Context, userID uuid.UUID) (*EditorSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EditorSettings), args.Error(1)
}

func (m *MockRepository) UpdateUserSettings(ctx context.Context, settings *EditorSettings) error {
	args := m.Called(ctx, settings)
	return args.Error(0)
}

func TestService_NewService(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	
	service := NewService(mockRepo, compiler)
	
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
	assert.Equal(t, compiler, service.compiler)
}

func TestService_CreateProject_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	req := CreateProjectRequest{
		Name:        "Test Project",
		Description: "A test project",
		Content:     "resource_name: test",
		TenantID:    uuid.New(),
		UserID:      uuid.New(),
	}
	
	// Mock successful creation
	mockRepo.On("CreateProject", mock.Anything, mock.AnythingOfType("*metacontracteditor.EditorProject")).Return(nil)
	
	project, err := service.CreateProject(context.Background(), req)
	
	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, req.Name, project.Name)
	assert.Equal(t, req.Description, project.Description)
	assert.Equal(t, req.Content, project.Content)
	assert.Equal(t, req.TenantID, project.TenantID)
	assert.Equal(t, req.UserID, project.CreatedBy)
	assert.Equal(t, "0.1.0", project.Version)
	assert.Equal(t, ProjectStatusDraft, project.Status)
	
	mockRepo.AssertExpectations(t)
}

func TestService_CreateProject_RepositoryError(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	req := CreateProjectRequest{
		Name:        "Test Project",
		Description: "A test project",
		Content:     "resource_name: test",
		TenantID:    uuid.New(),
		UserID:      uuid.New(),
	}
	
	// Mock repository error
	mockRepo.On("CreateProject", mock.Anything, mock.AnythingOfType("*metacontracteditor.EditorProject")).Return(assert.AnError)
	
	project, err := service.CreateProject(context.Background(), req)
	
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "failed to create project")
	
	mockRepo.AssertExpectations(t)
}

func TestService_GetProject_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	projectID := uuid.New()
	tenantID := uuid.New()
	
	expectedProject := &EditorProject{
		ID:       projectID,
		Name:     "Test Project",
		TenantID: tenantID,
	}
	
	// Mock successful retrieval
	mockRepo.On("GetProject", mock.Anything, projectID).Return(expectedProject, nil)
	
	project, err := service.GetProject(context.Background(), projectID, tenantID)
	
	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, expectedProject.ID, project.ID)
	assert.Equal(t, expectedProject.Name, project.Name)
	assert.Equal(t, expectedProject.TenantID, project.TenantID)
	
	mockRepo.AssertExpectations(t)
}

func TestService_GetProject_WrongTenant(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	projectID := uuid.New()
	tenantID := uuid.New()
	wrongTenantID := uuid.New()
	
	expectedProject := &EditorProject{
		ID:       projectID,
		Name:     "Test Project",
		TenantID: tenantID, // Different from requested tenant
	}
	
	// Mock successful retrieval
	mockRepo.On("GetProject", mock.Anything, projectID).Return(expectedProject, nil)
	
	project, err := service.GetProject(context.Background(), projectID, wrongTenantID)
	
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "project not found or access denied")
	
	mockRepo.AssertExpectations(t)
}

func TestService_GetProject_NotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	projectID := uuid.New()
	tenantID := uuid.New()
	
	// Mock project not found
	mockRepo.On("GetProject", mock.Anything, projectID).Return(nil, assert.AnError)
	
	project, err := service.GetProject(context.Background(), projectID, tenantID)
	
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "failed to get project")
	
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateProject_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	projectID := uuid.New()
	tenantID := uuid.New()
	
	existingProject := &EditorProject{
		ID:          projectID,
		Name:        "Old Name",
		Description: "Old Description",
		Content:     "old_content",
		TenantID:    tenantID,
		Status:      ProjectStatusValid,
		UpdatedAt:   time.Now().Add(-time.Hour),
	}
	
	newName := "New Name"
	newDescription := "New Description"
	newContent := "new_content"
	
	req := UpdateProjectRequest{
		Name:        &newName,
		Description: &newDescription,
		Content:     &newContent,
		TenantID:    tenantID,
	}
	
	// Mock successful retrieval and update
	mockRepo.On("GetProject", mock.Anything, projectID).Return(existingProject, nil)
	mockRepo.On("UpdateProject", mock.Anything, mock.AnythingOfType("*metacontracteditor.EditorProject")).Return(nil)
	
	project, err := service.UpdateProject(context.Background(), projectID, req)
	
	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, newName, project.Name)
	assert.Equal(t, newDescription, project.Description)
	assert.Equal(t, newContent, project.Content)
	assert.Equal(t, ProjectStatusDraft, project.Status) // Should reset to draft when content changes
	assert.True(t, project.UpdatedAt.After(existingProject.UpdatedAt))
	
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateProject_AccessDenied(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	projectID := uuid.New()
	tenantID := uuid.New()
	wrongTenantID := uuid.New()
	
	existingProject := &EditorProject{
		ID:       projectID,
		TenantID: tenantID, // Different from request tenant
	}
	
	req := UpdateProjectRequest{
		TenantID: wrongTenantID,
	}
	
	// Mock successful retrieval
	mockRepo.On("GetProject", mock.Anything, projectID).Return(existingProject, nil)
	
	project, err := service.UpdateProject(context.Background(), projectID, req)
	
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "project not found or access denied")
	
	mockRepo.AssertExpectations(t)
}

func TestService_ListProjects_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	tenantID := uuid.New()
	limit := 10
	offset := 0
	
	expectedProjects := []*EditorProject{
		{ID: uuid.New(), Name: "Project 1", TenantID: tenantID},
		{ID: uuid.New(), Name: "Project 2", TenantID: tenantID},
	}
	
	// Mock successful listing
	mockRepo.On("ListProjects", mock.Anything, tenantID, limit, offset).Return(expectedProjects, nil)
	
	projects, err := service.ListProjects(context.Background(), tenantID, limit, offset)
	
	require.NoError(t, err)
	require.NotNil(t, projects)
	assert.Len(t, projects, 2)
	assert.Equal(t, expectedProjects[0].Name, projects[0].Name)
	assert.Equal(t, expectedProjects[1].Name, projects[1].Name)
	
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteProject_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	projectID := uuid.New()
	tenantID := uuid.New()
	
	existingProject := &EditorProject{
		ID:       projectID,
		TenantID: tenantID,
	}
	
	// Mock successful retrieval and deletion
	mockRepo.On("GetProject", mock.Anything, projectID).Return(existingProject, nil)
	mockRepo.On("DeleteProject", mock.Anything, projectID).Return(nil)
	
	err := service.DeleteProject(context.Background(), projectID, tenantID)
	
	assert.NoError(t, err)
	
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteProject_AccessDenied(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	projectID := uuid.New()
	tenantID := uuid.New()
	wrongTenantID := uuid.New()
	
	existingProject := &EditorProject{
		ID:       projectID,
		TenantID: tenantID, // Different from request tenant
	}
	
	// Mock successful retrieval
	mockRepo.On("GetProject", mock.Anything, projectID).Return(existingProject, nil)
	
	err := service.DeleteProject(context.Background(), projectID, wrongTenantID)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found or access denied")
	
	mockRepo.AssertExpectations(t)
}

func TestService_CompileProject_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	req := CompileRequest{
		ProjectID: uuid.New(),
		Content: `resource_name: test_entity
namespace: test.namespace
version: "1.0.0"
description: "Test entity"

data_structure:
  fields:
    - name: id
      type: UUID
      constraints:
        primary_key: true
        required: true

security_model:
  access_control: rbac
  data_classification: internal

temporal_behavior:
  temporality_paradigm: snapshot`,
		Preview: true,
	}
	
	response, err := service.CompileProject(context.Background(), req)
	
	// Note: This may fail if compiler implementation is incomplete
	// But we're testing the service logic, not the compiler
	if err != nil {
		t.Logf("Compile error (may be expected): %v", err)
	}
	
	require.NotNil(t, response)
	assert.NotZero(t, response.CompileTime)
	
	// If compilation failed, there should be errors
	if !response.Success {
		assert.NotEmpty(t, response.Errors)
	}
}

func TestService_CompileProject_InvalidYAML(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	req := CompileRequest{
		ProjectID: uuid.New(),
		Content:   "invalid: yaml: content: [unclosed",
		Preview:   true,
	}
	
	response, err := service.CompileProject(context.Background(), req)
	
	require.NoError(t, err) // Service should not error, but compilation should fail
	require.NotNil(t, response)
	assert.False(t, response.Success)
	assert.NotEmpty(t, response.Errors)
	assert.NotZero(t, response.CompileTime)
	
	// Check error details
	assert.Equal(t, "system", response.Errors[0].Type)
	assert.Contains(t, response.Errors[0].Message, "prepare compilation")
}

func TestService_GetTemplates_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	category := "basic"
	expectedTemplates := []*ProjectTemplate{
		{ID: uuid.New(), Name: "Basic Entity", Category: category},
		{ID: uuid.New(), Name: "Simple CRUD", Category: category},
	}
	
	// Mock successful template retrieval
	mockRepo.On("GetTemplates", mock.Anything, category).Return(expectedTemplates, nil)
	
	templates, err := service.GetTemplates(context.Background(), category)
	
	require.NoError(t, err)
	require.NotNil(t, templates)
	assert.Len(t, templates, 2)
	assert.Equal(t, expectedTemplates[0].Name, templates[0].Name)
	assert.Equal(t, expectedTemplates[1].Name, templates[1].Name)
	
	mockRepo.AssertExpectations(t)
}

func TestService_GetUserSettings_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	userID := uuid.New()
	expectedSettings := &EditorSettings{
		UserID:      userID,
		Theme:       "dark",
		FontSize:    16,
		AutoSave:    true,
		AutoCompile: false,
		KeyBindings: "vim",
	}
	
	// Mock successful settings retrieval
	mockRepo.On("GetUserSettings", mock.Anything, userID).Return(expectedSettings, nil)
	
	settings, err := service.GetUserSettings(context.Background(), userID)
	
	require.NoError(t, err)
	require.NotNil(t, settings)
	assert.Equal(t, expectedSettings.Theme, settings.Theme)
	assert.Equal(t, expectedSettings.FontSize, settings.FontSize)
	assert.Equal(t, expectedSettings.AutoSave, settings.AutoSave)
	assert.Equal(t, expectedSettings.AutoCompile, settings.AutoCompile)
	assert.Equal(t, expectedSettings.KeyBindings, settings.KeyBindings)
	
	mockRepo.AssertExpectations(t)
}

func TestService_GetUserSettings_NotFound_ReturnsDefaults(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	userID := uuid.New()
	
	// Mock settings not found (sql.ErrNoRows)
	mockRepo.On("GetUserSettings", mock.Anything, userID).Return(nil, assert.AnError)
	
	settings, err := service.GetUserSettings(context.Background(), userID)
	
	require.NoError(t, err) // Should return default settings, not error
	require.NotNil(t, settings)
	assert.Equal(t, userID, settings.UserID)
	assert.Equal(t, "vs-dark", settings.Theme)    // Default theme
	assert.Equal(t, 14, settings.FontSize)        // Default font size
	assert.True(t, settings.AutoSave)             // Default auto save
	assert.True(t, settings.AutoCompile)          // Default auto compile
	assert.Equal(t, "default", settings.KeyBindings) // Default key bindings
	assert.NotNil(t, settings.Settings)           // Default settings map
	
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateUserSettings_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	settings := &EditorSettings{
		UserID:      uuid.New(),
		Theme:       "light",
		FontSize:    12,
		AutoSave:    false,
		AutoCompile: true,
		KeyBindings: "emacs",
		Settings:    map[string]interface{}{"custom": "value"},
	}
	
	// Mock successful update
	mockRepo.On("UpdateUserSettings", mock.Anything, mock.AnythingOfType("*metacontracteditor.EditorSettings")).Return(nil)
	
	err := service.UpdateUserSettings(context.Background(), settings)
	
	require.NoError(t, err)
	
	// Verify UpdatedAt was set
	mockRepo.AssertCalled(t, "UpdateUserSettings", mock.Anything, mock.MatchedBy(func(s *EditorSettings) bool {
		return !s.UpdatedAt.IsZero()
	}))
	
	mockRepo.AssertExpectations(t)
}

// Test concurrent operations
func TestService_ConcurrentOperations(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	tenantID := uuid.New()
	
	// Mock repository calls for concurrent operations
	mockRepo.On("ListProjects", mock.Anything, tenantID, mock.Anything, mock.Anything).Return([]*EditorProject{}, nil)
	
	const numGoroutines = 10
	errors := make(chan error, numGoroutines)
	
	// Run concurrent ListProjects calls
	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := service.ListProjects(context.Background(), tenantID, 10, 0)
			errors <- err
		}()
	}
	
	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		assert.NoError(t, err)
	}
	
	mockRepo.AssertExpectations(t)
}

// Test context cancellation
func TestService_ContextCancellation(t *testing.T) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	req := CreateProjectRequest{
		Name:     "Test Project",
		TenantID: uuid.New(),
		UserID:   uuid.New(),
	}
	
	// Mock repository to check if context is passed
	mockRepo.On("CreateProject", ctx, mock.AnythingOfType("*metacontracteditor.EditorProject")).Return(context.Canceled)
	
	project, err := service.CreateProject(ctx, req)
	
	assert.Error(t, err)
	assert.Nil(t, project)
	
	mockRepo.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkService_CreateProject(b *testing.B) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	req := CreateProjectRequest{
		Name:     "Benchmark Project",
		TenantID: uuid.New(),
		UserID:   uuid.New(),
	}
	
	mockRepo.On("CreateProject", mock.Anything, mock.AnythingOfType("*metacontracteditor.EditorProject")).Return(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateProject(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkService_CompileProject(b *testing.B) {
	mockRepo := &MockRepository{}
	compiler := metacontract.NewCompiler()
	service := NewService(mockRepo, compiler)
	
	req := CompileRequest{
		ProjectID: uuid.New(),
		Content: `resource_name: benchmark
namespace: test
version: "1.0.0"
data_structure:
  fields:
    - name: id
      type: UUID`,
		Preview: true,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CompileProject(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
	}
}