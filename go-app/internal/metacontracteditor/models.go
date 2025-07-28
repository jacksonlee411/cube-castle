// internal/metacontracteditor/models.go
package metacontracteditor

import (
	"time"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/types"
)

// EditorProject represents a meta-contract editing project
type EditorProject struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	Description  string                 `json:"description" db:"description"`
	Content      string                 `json:"content" db:"content"` // YAML content
	Version      string                 `json:"version" db:"version"`
	Status       ProjectStatus          `json:"status" db:"status"`
	TenantID     uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	CreatedBy    uuid.UUID              `json:"created_by" db:"created_by"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
	LastCompiled *time.Time             `json:"last_compiled,omitempty" db:"last_compiled"`
	CompileError *string                `json:"compile_error,omitempty" db:"compile_error"`
}

// ProjectStatus defines the status of an editor project
type ProjectStatus string

const (
	ProjectStatusDraft     ProjectStatus = "draft"
	ProjectStatusCompiling ProjectStatus = "compiling"
	ProjectStatusValid     ProjectStatus = "valid"
	ProjectStatusError     ProjectStatus = "error"
	ProjectStatusPublished ProjectStatus = "published"
)

// CompileRequest represents a compilation request
type CompileRequest struct {
	ProjectID uuid.UUID `json:"project_id"`
	Content   string    `json:"content"`
	Preview   bool      `json:"preview"` // if true, don't save results
}

// CompileResponse represents a compilation response
type CompileResponse struct {
	Success      bool                    `json:"success"`
	Errors       []CompileError          `json:"errors,omitempty"`
	Warnings     []CompileWarning        `json:"warnings,omitempty"`
	GeneratedFiles map[string]string     `json:"generated_files,omitempty"` // filename -> content
	Schema       *types.MetaContract     `json:"schema,omitempty"`
	CompileTime  time.Duration           `json:"compile_time"`
}

// CompileError represents a compilation error
type CompileError struct {
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
}

// CompileWarning represents a compilation warning
type CompileWarning struct {
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Type     string `json:"type"`
}

// EditorSession represents an active editing session
type EditorSession struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id"`
	StartedAt time.Time `json:"started_at"`
	LastSeen  time.Time `json:"last_seen"`
	Active    bool      `json:"active"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      MessageType `json:"type"`
	ProjectID uuid.UUID   `json:"project_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// MessageType defines WebSocket message types
type MessageType string

const (
	MessageTypeCompileRequest  MessageType = "compile_request"
	MessageTypeCompileResponse MessageType = "compile_response"
	MessageTypeContentChange   MessageType = "content_change"
	MessageTypeSessionStart    MessageType = "session_start"
	MessageTypeSessionEnd      MessageType = "session_end"
	MessageTypeError           MessageType = "error"
)

// ContentChangeData represents content change data
type ContentChangeData struct {
	Content   string    `json:"content"`
	Line      int       `json:"line,omitempty"`
	Column    int       `json:"column,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ProjectTemplate represents a project template
type ProjectTemplate struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"`
	Content     string    `json:"content" db:"content"`
	Tags        []string  `json:"tags" db:"tags"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// EditorSettings represents user editor settings
type EditorSettings struct {
	UserID       uuid.UUID            `json:"user_id" db:"user_id"`
	Theme        string               `json:"theme" db:"theme"`
	FontSize     int                  `json:"font_size" db:"font_size"`
	AutoSave     bool                 `json:"auto_save" db:"auto_save"`
	AutoCompile  bool                 `json:"auto_compile" db:"auto_compile"`
	KeyBindings  string               `json:"key_bindings" db:"key_bindings"`
	Settings     map[string]interface{} `json:"settings" db:"settings"`
	UpdatedAt    time.Time            `json:"updated_at" db:"updated_at"`
}