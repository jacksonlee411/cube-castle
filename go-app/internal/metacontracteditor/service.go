// internal/metacontracteditor/service.go
package metacontracteditor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/metacontract"
	"github.com/google/uuid"
)

// Service provides the business logic for the meta-contract editor
type Service struct {
	repo     Repository
	compiler *metacontract.Compiler
}

// NewService creates a new editor service
func NewService(repo Repository, compiler *metacontract.Compiler) *Service {
	return &Service{
		repo:     repo,
		compiler: compiler,
	}
}

// CreateProject creates a new editor project
func (s *Service) CreateProject(ctx context.Context, req CreateProjectRequest) (*EditorProject, error) {
	project := &EditorProject{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		Version:     "0.1.0",
		Status:      ProjectStatusDraft,
		TenantID:    req.TenantID,
		CreatedBy:   req.UserID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// GetProject retrieves a project by ID
func (s *Service) GetProject(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID) (*EditorProject, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if project.TenantID != tenantID {
		return nil, fmt.Errorf("project not found or access denied")
	}

	return project, nil
}

// UpdateProject updates a project
func (s *Service) UpdateProject(ctx context.Context, projectID uuid.UUID, req UpdateProjectRequest) (*EditorProject, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if project.TenantID != req.TenantID {
		return nil, fmt.Errorf("project not found or access denied")
	}

	// Update fields
	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Description != nil {
		project.Description = *req.Description
	}
	if req.Content != nil {
		project.Content = *req.Content
		project.Status = ProjectStatusDraft // Reset status when content changes
	}
	project.UpdatedAt = time.Now()

	if err := s.repo.UpdateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// ListProjects lists projects for a tenant
func (s *Service) ListProjects(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*EditorProject, error) {
	return s.repo.ListProjects(ctx, tenantID, limit, offset)
}

// DeleteProject deletes a project
func (s *Service) DeleteProject(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID) error {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	if project.TenantID != tenantID {
		return fmt.Errorf("project not found or access denied")
	}

	return s.repo.DeleteProject(ctx, projectID)
}

// CompileProject compiles a project's meta-contract
func (s *Service) CompileProject(ctx context.Context, req CompileRequest) (*CompileResponse, error) {
	startTime := time.Now()

	response := &CompileResponse{
		Success:        false,
		GeneratedFiles: make(map[string]string),
		CompileTime:    0,
	}

	// Write content to temporary file for compilation
	tempFile, err := s.writeContentToTempFile(req.Content)
	if err != nil {
		response.Errors = append(response.Errors, CompileError{
			Line:     1,
			Column:   1,
			Message:  fmt.Sprintf("Failed to prepare compilation: %v", err),
			Type:     "system",
			Severity: "error",
		})
		response.CompileTime = time.Since(startTime)
		return response, nil
	}
	defer s.cleanupTempFile(tempFile)

	// Parse the meta-contract
	contract, err := s.compiler.ParseMetaContract(tempFile)
	if err != nil {
		response.Errors = append(response.Errors, CompileError{
			Line:     1,
			Column:   1,
			Message:  fmt.Sprintf("Parse error: %v", err),
			Type:     "parse",
			Severity: "error",
		})
		response.CompileTime = time.Since(startTime)
		return response, nil
	}

	response.Schema = contract

	// Generate code artifacts
	tempOutputDir, err := s.createTempOutputDir()
	if err != nil {
		response.Errors = append(response.Errors, CompileError{
			Line:     1,
			Column:   1,
			Message:  fmt.Sprintf("Failed to create output directory: %v", err),
			Type:     "system",
			Severity: "error",
		})
		response.CompileTime = time.Since(startTime)
		return response, nil
	}
	defer s.cleanupTempDir(tempOutputDir)

	// Compile the contract
	if err := s.compiler.Compile(tempFile, tempOutputDir); err != nil {
		response.Errors = append(response.Errors, CompileError{
			Line:     1,
			Column:   1,
			Message:  fmt.Sprintf("Compilation error: %v", err),
			Type:     "compile",
			Severity: "error",
		})
		response.CompileTime = time.Since(startTime)
		return response, nil
	}

	// Read generated files
	generatedFiles, err := s.readGeneratedFiles(tempOutputDir)
	if err != nil {
		response.Warnings = append(response.Warnings, CompileWarning{
			Line:    1,
			Column:  1,
			Message: fmt.Sprintf("Failed to read some generated files: %v", err),
			Type:    "system",
		})
	} else {
		response.GeneratedFiles = generatedFiles
	}

	response.Success = true
	response.CompileTime = time.Since(startTime)

	// If not preview, update project status
	if !req.Preview {
		if err := s.updateProjectCompileStatus(ctx, req.ProjectID, response); err != nil {
			// Log error but don't fail the compilation
			fmt.Printf("Failed to update project compile status: %v\n", err)
		}
	}

	return response, nil
}

// GetTemplates retrieves available project templates
func (s *Service) GetTemplates(ctx context.Context, category string) ([]*ProjectTemplate, error) {
	return s.repo.GetTemplates(ctx, category)
}

// GetUserSettings retrieves user editor settings
func (s *Service) GetUserSettings(ctx context.Context, userID uuid.UUID) (*EditorSettings, error) {
	settings, err := s.repo.GetUserSettings(ctx, userID)
	if err == sql.ErrNoRows {
		// Return default settings
		return &EditorSettings{
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
	return settings, err
}

// UpdateUserSettings updates user editor settings
func (s *Service) UpdateUserSettings(ctx context.Context, settings *EditorSettings) error {
	settings.UpdatedAt = time.Now()
	return s.repo.UpdateUserSettings(ctx, settings)
}

// Request/Response types
type CreateProjectRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	TenantID    uuid.UUID `json:"tenant_id"`
	UserID      uuid.UUID `json:"user_id"`
}

type UpdateProjectRequest struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Content     *string   `json:"content,omitempty"`
	TenantID    uuid.UUID `json:"tenant_id"`
}

// Helper methods
func (s *Service) writeContentToTempFile(content string) (string, error) {
	// Create temporary file with .yaml extension
	tmpFile, err := os.CreateTemp("", "metacontract-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	
	// Write content to file
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write content: %w", err)
	}
	
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}
	
	return tmpFile.Name(), nil
}

func (s *Service) cleanupTempFile(path string) {
	if path != "" {
		os.Remove(path)
	}
}

func (s *Service) createTempOutputDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "metacontract-output-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	return tmpDir, nil
}

func (s *Service) cleanupTempDir(path string) {
	if path != "" {
		os.RemoveAll(path)
	}
}

func (s *Service) readGeneratedFiles(dir string) (map[string]string, error) {
	files := make(map[string]string)
	
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if d.IsDir() {
			return nil
		}
		
		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}
		
		// Use relative path as key
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			relPath = filepath.Base(path)
		}
		
		files[relPath] = string(content)
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to read generated files: %w", err)
	}
	
	return files, nil
}

func (s *Service) updateProjectCompileStatus(ctx context.Context, projectID uuid.UUID, response *CompileResponse) error {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return err
	}

	now := time.Now()
	project.LastCompiled = &now
	project.UpdatedAt = now

	if response.Success {
		project.Status = ProjectStatusValid
		project.CompileError = nil
	} else {
		project.Status = ProjectStatusError
		errorsJSON, _ := json.Marshal(response.Errors)
		errStr := string(errorsJSON)
		project.CompileError = &errStr
	}

	return s.repo.UpdateProject(ctx, project)
}
