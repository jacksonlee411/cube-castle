// internal/metacontracteditor/repository.go
package metacontracteditor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// Repository defines the interface for meta-contract editor data access
type Repository interface {
	// Project operations
	CreateProject(ctx context.Context, project *EditorProject) error
	GetProject(ctx context.Context, projectID uuid.UUID) (*EditorProject, error)
	UpdateProject(ctx context.Context, project *EditorProject) error
	ListProjects(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*EditorProject, error)
	DeleteProject(ctx context.Context, projectID uuid.UUID) error

	// Session operations
	CreateSession(ctx context.Context, session *EditorSession) error
	GetSession(ctx context.Context, sessionID uuid.UUID) (*EditorSession, error)
	EndSession(ctx context.Context, sessionID uuid.UUID) error
	GetActiveSessions(ctx context.Context, projectID uuid.UUID) ([]*EditorSession, error)

	// Template operations
	GetTemplates(ctx context.Context, category string) ([]*ProjectTemplate, error)
	CreateTemplate(ctx context.Context, template *ProjectTemplate) error

	// Settings operations
	GetUserSettings(ctx context.Context, userID uuid.UUID) (*EditorSettings, error)
	UpdateUserSettings(ctx context.Context, settings *EditorSettings) error
}

// PostgreSQLRepository implements Repository using PostgreSQL
type PostgreSQLRepository struct {
	db *sqlx.DB
}

// NewPostgreSQLRepository creates a new PostgreSQL repository
func NewPostgreSQLRepository(db *sqlx.DB) Repository {
	return &PostgreSQLRepository{db: db}
}

// CreateProject creates a new editor project
func (r *PostgreSQLRepository) CreateProject(ctx context.Context, project *EditorProject) error {
	query := `
		INSERT INTO metacontract_editor_projects (
			id, name, description, content, version, status, 
			tenant_id, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	_, err := r.db.ExecContext(ctx, query,
		project.ID, project.Name, project.Description, project.Content,
		project.Version, project.Status, project.TenantID, project.CreatedBy,
		project.CreatedAt, project.UpdatedAt,
	)

	return err
}

// GetProject retrieves a project by ID
func (r *PostgreSQLRepository) GetProject(ctx context.Context, projectID uuid.UUID) (*EditorProject, error) {
	query := `
		SELECT id, name, description, content, version, status, 
		       tenant_id, created_by, created_at, updated_at, 
		       last_compiled, compile_error
		FROM metacontract_editor_projects 
		WHERE id = $1`

	var project EditorProject
	err := r.db.GetContext(ctx, &project, query, projectID)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// UpdateProject updates a project
func (r *PostgreSQLRepository) UpdateProject(ctx context.Context, project *EditorProject) error {
	query := `
		UPDATE metacontract_editor_projects 
		SET name = $2, description = $3, content = $4, version = $5, 
		    status = $6, updated_at = $7, last_compiled = $8, compile_error = $9
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		project.ID, project.Name, project.Description, project.Content,
		project.Version, project.Status, project.UpdatedAt,
		project.LastCompiled, project.CompileError,
	)

	return err
}

// ListProjects lists projects for a tenant
func (r *PostgreSQLRepository) ListProjects(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*EditorProject, error) {
	query := `
		SELECT id, name, description, content, version, status, 
		       tenant_id, created_by, created_at, updated_at, 
		       last_compiled, compile_error
		FROM metacontract_editor_projects 
		WHERE tenant_id = $1 AND current_setting('app.current_tenant_id', true) = $1::text
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3`

	var projects []*EditorProject
	err := r.db.SelectContext(ctx, &projects, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// DeleteProject deletes a project
func (r *PostgreSQLRepository) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	query := `DELETE FROM metacontract_editor_projects WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, projectID)
	return err
}

// CreateSession creates a new editing session
func (r *PostgreSQLRepository) CreateSession(ctx context.Context, session *EditorSession) error {
	query := `
		INSERT INTO metacontract_editor_sessions (
			id, project_id, user_id, started_at, last_seen, active
		) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.ProjectID, session.UserID,
		session.StartedAt, session.LastSeen, session.Active,
	)

	return err
}

// GetSession retrieves a session by ID
func (r *PostgreSQLRepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*EditorSession, error) {
	query := `
		SELECT id, project_id, user_id, started_at, last_seen, active
		FROM metacontract_editor_sessions 
		WHERE id = $1`

	var session EditorSession
	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// EndSession ends an editing session
func (r *PostgreSQLRepository) EndSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE metacontract_editor_sessions 
		SET active = false, last_seen = $2
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID, time.Now())
	return err
}

// GetActiveSessions retrieves active sessions for a project
func (r *PostgreSQLRepository) GetActiveSessions(ctx context.Context, projectID uuid.UUID) ([]*EditorSession, error) {
	query := `
		SELECT id, project_id, user_id, started_at, last_seen, active
		FROM metacontract_editor_sessions 
		WHERE project_id = $1 AND active = true
		ORDER BY last_seen DESC`

	var sessions []*EditorSession
	err := r.db.SelectContext(ctx, &sessions, query, projectID)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetTemplates retrieves project templates
func (r *PostgreSQLRepository) GetTemplates(ctx context.Context, category string) ([]*ProjectTemplate, error) {
	query := `
		SELECT id, name, description, category, content, tags, created_at, updated_at
		FROM metacontract_editor_templates`

	args := []interface{}{}
	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}

	query += " ORDER BY created_at DESC"

	var templates []*ProjectTemplate
	err := r.db.SelectContext(ctx, &templates, query, args...)
	if err != nil {
		return nil, err
	}

	return templates, nil
}

// CreateTemplate creates a new project template
func (r *PostgreSQLRepository) CreateTemplate(ctx context.Context, template *ProjectTemplate) error {
	query := `
		INSERT INTO metacontract_editor_templates (
			id, name, description, category, content, tags, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		template.ID, template.Name, template.Description, template.Category,
		template.Content, pq.Array(template.Tags), template.CreatedAt, template.UpdatedAt,
	)

	return err
}

// GetUserSettings retrieves user editor settings
func (r *PostgreSQLRepository) GetUserSettings(ctx context.Context, userID uuid.UUID) (*EditorSettings, error) {
	query := `
		SELECT user_id, theme, font_size, auto_save, auto_compile, 
		       key_bindings, settings, updated_at
		FROM metacontract_editor_settings 
		WHERE user_id = $1`

	var settings EditorSettings
	var settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&settings.UserID, &settings.Theme, &settings.FontSize,
		&settings.AutoSave, &settings.AutoCompile, &settings.KeyBindings,
		&settingsJSON, &settings.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &settings.Settings); err != nil {
			return nil, err
		}
	}

	return &settings, nil
}

// UpdateUserSettings updates user editor settings
func (r *PostgreSQLRepository) UpdateUserSettings(ctx context.Context, settings *EditorSettings) error {
	settingsJSON, err := json.Marshal(settings.Settings)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO metacontract_editor_settings (
			user_id, theme, font_size, auto_save, auto_compile, 
			key_bindings, settings, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id) DO UPDATE SET
			theme = EXCLUDED.theme,
			font_size = EXCLUDED.font_size,
			auto_save = EXCLUDED.auto_save,
			auto_compile = EXCLUDED.auto_compile,
			key_bindings = EXCLUDED.key_bindings,
			settings = EXCLUDED.settings,
			updated_at = EXCLUDED.updated_at`

	_, err = r.db.ExecContext(ctx, query,
		settings.UserID, settings.Theme, settings.FontSize,
		settings.AutoSave, settings.AutoCompile, settings.KeyBindings,
		settingsJSON, settings.UpdatedAt,
	)

	return err
}
