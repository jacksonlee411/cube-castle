package common

import (
	"time"

	"github.com/google/uuid"
)

// BaseEntity 所有实体的基础结构
type BaseEntity struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TenantEntity 包含租户ID的实体基础结构
type TenantEntity struct {
	BaseEntity
	TenantID uuid.UUID `json:"tenant_id" db:"tenant_id"`
}

// Pagination 分页参数
type Pagination struct {
	Page     int `json:"page" query:"page"`
	PageSize int `json:"page_size" query:"page_size"`
}

// PaginatedResponse 分页响应
type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

// APIResponse 标准API响应
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// Status 通用状态枚举
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusPending  Status = "pending"
	StatusDeleted  Status = "deleted"
)

// ContextKey 上下文键类型
type ContextKey string

const (
	ContextKeyUserID   ContextKey = "user_id"
	ContextKeyTenantID ContextKey = "tenant_id"
	ContextKeyUser     ContextKey = "user"
)

// UserContext 用户上下文信息
type UserContext struct {
	UserID   uuid.UUID `json:"user_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Roles    []string  `json:"roles"`
}

// Event 事件基础结构
type Event struct {
	ID            uuid.UUID      `json:"id"`
	AggregateID   uuid.UUID      `json:"aggregate_id"`
	AggregateType string         `json:"aggregate_type"`
	EventType     string         `json:"event_type"`
	EventVersion  int            `json:"event_version"`
	Payload       map[string]any `json:"payload"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors 验证错误集合
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	return ve[0].Message
}
