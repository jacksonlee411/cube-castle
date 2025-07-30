package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event 发件箱事件模型
type Event struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	AggregateID   uuid.UUID       `json:"aggregate_id" db:"aggregate_id"`
	AggregateType string          `json:"aggregate_type" db:"aggregate_type"`
	EventType     string          `json:"event_type" db:"event_type"`
	EventVersion  int             `json:"event_version" db:"event_version"`
	Payload       json.RawMessage `json:"payload" db:"payload"`
	Metadata      json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	ProcessedAt   *time.Time      `json:"processed_at,omitempty" db:"processed_at"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// CreateEventRequest 创建事件请求
type CreateEventRequest struct {
	AggregateID   uuid.UUID       `json:"aggregate_id" validate:"required"`
	AggregateType string          `json:"aggregate_type" validate:"required"`
	EventType     string          `json:"event_type" validate:"required"`
	EventVersion  int             `json:"event_version"`
	Payload       json.RawMessage `json:"payload" validate:"required"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
}

// EventHandler 事件处理器接口
type EventHandler interface {
	HandleEvent(ctx context.Context, event *Event) error
	GetEventType() string
}

// EventProcessor 事件处理器
type EventProcessor struct {
	handlers map[string]EventHandler
}

// NewEventProcessor 创建新的事件处理器
func NewEventProcessor() *EventProcessor {
	return &EventProcessor{
		handlers: make(map[string]EventHandler),
	}
}

// RegisterHandler 注册事件处理器
func (p *EventProcessor) RegisterHandler(handler EventHandler) {
	p.handlers[handler.GetEventType()] = handler
}

// GetHandler 获取事件处理器
func (p *EventProcessor) GetHandler(eventType string) (EventHandler, bool) {
	handler, exists := p.handlers[eventType]
	return handler, exists
}

// EventTypes 预定义的事件类型
const (
	// CoreHR 事件类型
	EventTypeEmployeeCreated      = "employee.created"
	EventTypeEmployeeUpdated      = "employee.updated"
	EventTypeEmployeeDeleted      = "employee.deleted"
	EventTypeEmployeePhoneUpdated = "employee.phone_updated"

	// Organization 事件类型
	EventTypeOrganizationCreated = "organization.created"
	EventTypeOrganizationUpdated = "organization.updated"
	EventTypeOrganizationDeleted = "organization.deleted"

	// Position 事件类型
	EventTypePositionCreated = "position.created"
	EventTypePositionUpdated = "position.updated"
	EventTypePositionDeleted = "position.deleted"

	// Leave 事件类型
	EventTypeLeaveRequestCreated  = "leave_request.created"
	EventTypeLeaveRequestApproved = "leave_request.approved"
	EventTypeLeaveRequestRejected = "leave_request.rejected"

	// Notification 事件类型
	EventTypeNotification = "notification.created"
)

// AggregateTypes 预定义的聚合类型
const (
	AggregateTypeEmployee     = "Employee"
	AggregateTypeOrganization = "Organization"
	AggregateTypePosition     = "Position"
	AggregateTypeLeaveRequest = "LeaveRequest"
	AggregateTypeNotification = "Notification"
)
