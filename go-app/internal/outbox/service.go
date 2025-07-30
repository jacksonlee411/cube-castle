package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service 发件箱服务层
type Service struct {
	repo           *Repository
	processor      *OutboxProcessor
	eventProcessor *EventProcessor
	db             *pgxpool.Pool
}

// NewService 创建新的发件箱服务
func NewService(db *pgxpool.Pool) *Service {
	repo := NewRepository(db)
	eventProcessor := NewEventProcessor()
	processor := NewOutboxProcessor(repo, eventProcessor, DefaultProcessorConfig())

	return &Service{
		repo:           repo,
		processor:      processor,
		eventProcessor: eventProcessor,
		db:             db,
	}
}

// Start 启动发件箱服务
func (s *Service) Start(ctx context.Context) error {
	return s.processor.Start(ctx)
}

// CreateEvent 创建事件
func (s *Service) CreateEvent(ctx context.Context, req *CreateEventRequest) (*Event, error) {
	return s.processor.CreateEvent(ctx, req)
}

// CreateEventWithTransaction 在事务中创建事件
func (s *Service) CreateEventWithTransaction(ctx context.Context, tx pgx.Tx, req *CreateEventRequest) (*Event, error) {
	return s.processor.CreateEventWithTransaction(ctx, tx, req)
}

// ProcessEvents 处理事件
func (s *Service) ProcessEvents(ctx context.Context) error {
	return s.processor.ProcessEvents(ctx)
}

// GetStats 获取统计信息
func (s *Service) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return s.processor.GetStats(ctx)
}

// RegisterHandler 注册事件处理器
func (s *Service) RegisterHandler(handler EventHandler) {
	s.eventProcessor.RegisterHandler(handler)
}

// ReplayEvents 重放事件
func (s *Service) ReplayEvents(ctx context.Context, aggregateID uuid.UUID) error {
	return s.processor.ReplayEvents(ctx, aggregateID)
}

// ReplayEventsByType 根据类型重放事件
func (s *Service) ReplayEventsByType(ctx context.Context, eventType string, limit int) error {
	return s.processor.ReplayEventsByType(ctx, eventType, limit)
}

// CoreHR 事件创建辅助方法

// CreateEmployeeCreatedEvent 创建员工创建事件
func (s *Service) CreateEmployeeCreatedEvent(ctx context.Context, employeeID uuid.UUID, employeeData map[string]interface{}) error {
	payload, err := json.Marshal(map[string]interface{}{
		"employee_id":   employeeID.String(),
		"employee_data": employeeData,
		"created_at":    time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal employee created event payload: %w", err)
	}

	req := &CreateEventRequest{
		AggregateID:   employeeID,
		AggregateType: AggregateTypeEmployee,
		EventType:     EventTypeEmployeeCreated,
		EventVersion:  1,
		Payload:       payload,
	}

	_, err = s.CreateEvent(ctx, req)
	return err
}

// CreateEmployeeUpdatedEvent 创建员工更新事件
func (s *Service) CreateEmployeeUpdatedEvent(ctx context.Context, employeeID uuid.UUID, updatedFields map[string]interface{}) error {
	payload, err := json.Marshal(map[string]interface{}{
		"employee_id":    employeeID.String(),
		"updated_fields": updatedFields,
		"updated_at":     time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal employee updated event payload: %w", err)
	}

	req := &CreateEventRequest{
		AggregateID:   employeeID,
		AggregateType: AggregateTypeEmployee,
		EventType:     EventTypeEmployeeUpdated,
		EventVersion:  1,
		Payload:       payload,
	}

	_, err = s.CreateEvent(ctx, req)
	return err
}

// CreateEmployeePhoneUpdatedEvent 创建员工电话更新事件
func (s *Service) CreateEmployeePhoneUpdatedEvent(ctx context.Context, employeeID uuid.UUID, oldPhoneNumber, newPhoneNumber string) error {
	payload, err := json.Marshal(map[string]interface{}{
		"employee_id":      employeeID.String(),
		"old_phone_number": oldPhoneNumber,
		"new_phone_number": newPhoneNumber,
		"updated_at":       time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal employee phone updated event payload: %w", err)
	}

	req := &CreateEventRequest{
		AggregateID:   employeeID,
		AggregateType: AggregateTypeEmployee,
		EventType:     EventTypeEmployeePhoneUpdated,
		EventVersion:  1,
		Payload:       payload,
	}

	_, err = s.CreateEvent(ctx, req)
	return err
}

// CreateOrganizationCreatedEvent 创建组织创建事件
func (s *Service) CreateOrganizationCreatedEvent(ctx context.Context, organizationID uuid.UUID, name, code string, parentID *uuid.UUID) error {
	payload, err := json.Marshal(map[string]interface{}{
		"organization_id": organizationID.String(),
		"name":            name,
		"code":            code,
		"parent_id":       parentID.String(),
		"created_at":      time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal organization created event payload: %w", err)
	}

	req := &CreateEventRequest{
		AggregateID:   organizationID,
		AggregateType: AggregateTypeOrganization,
		EventType:     EventTypeOrganizationCreated,
		EventVersion:  1,
		Payload:       payload,
	}

	_, err = s.CreateEvent(ctx, req)
	return err
}

// CreateLeaveRequestCreatedEvent 创建休假申请创建事件
func (s *Service) CreateLeaveRequestCreatedEvent(ctx context.Context, requestID, employeeID, managerID uuid.UUID, startDate, endDate, leaveType, reason string) error {
	payload, err := json.Marshal(map[string]interface{}{
		"request_id":  requestID.String(),
		"employee_id": employeeID.String(),
		"manager_id":  managerID.String(),
		"start_date":  startDate,
		"end_date":    endDate,
		"leave_type":  leaveType,
		"reason":      reason,
		"created_at":  time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal leave request created event payload: %w", err)
	}

	req := &CreateEventRequest{
		AggregateID:   requestID,
		AggregateType: AggregateTypeLeaveRequest,
		EventType:     EventTypeLeaveRequestCreated,
		EventVersion:  1,
		Payload:       payload,
	}

	_, err = s.CreateEvent(ctx, req)
	return err
}

// CreateLeaveRequestApprovedEvent 创建休假申请批准事件
func (s *Service) CreateLeaveRequestApprovedEvent(ctx context.Context, requestID, employeeID, approvedBy uuid.UUID, comment string) error {
	payload, err := json.Marshal(map[string]interface{}{
		"request_id":  requestID.String(),
		"employee_id": employeeID.String(),
		"approved_by": approvedBy.String(),
		"approved_at": time.Now().Format(time.RFC3339),
		"comment":     comment,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal leave request approved event payload: %w", err)
	}

	req := &CreateEventRequest{
		AggregateID:   requestID,
		AggregateType: AggregateTypeLeaveRequest,
		EventType:     EventTypeLeaveRequestApproved,
		EventVersion:  1,
		Payload:       payload,
	}

	_, err = s.CreateEvent(ctx, req)
	return err
}

// CreateLeaveRequestRejectedEvent 创建休假申请拒绝事件
func (s *Service) CreateLeaveRequestRejectedEvent(ctx context.Context, requestID, employeeID, rejectedBy uuid.UUID, reason string) error {
	payload, err := json.Marshal(map[string]interface{}{
		"request_id":  requestID.String(),
		"employee_id": employeeID.String(),
		"rejected_by": rejectedBy.String(),
		"rejected_at": time.Now().Format(time.RFC3339),
		"reason":      reason,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal leave request rejected event payload: %w", err)
	}

	req := &CreateEventRequest{
		AggregateID:   requestID,
		AggregateType: AggregateTypeLeaveRequest,
		EventType:     EventTypeLeaveRequestRejected,
		EventVersion:  1,
		Payload:       payload,
	}

	_, err = s.CreateEvent(ctx, req)
	return err
}

// CreateNotificationEvent 创建通知事件
func (s *Service) CreateNotificationEvent(ctx context.Context, recipientID uuid.UUID, notificationType, subject, content, channel string) error {
	payload, err := json.Marshal(map[string]interface{}{
		"type":         notificationType,
		"recipient_id": recipientID.String(),
		"subject":      subject,
		"content":      content,
		"channel":      channel,
		"created_at":   time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal notification event payload: %w", err)
	}

	req := &CreateEventRequest{
		AggregateID:   recipientID,
		AggregateType: "Notification",
		EventType:     "notification.sent",
		EventVersion:  1,
		Payload:       payload,
	}

	_, err = s.CreateEvent(ctx, req)
	return err
}

// 事务性事件创建方法

// CreateEmployeeCreatedEventWithTransaction 在事务中创建员工创建事件
func (s *Service) CreateEmployeeCreatedEventWithTransaction(ctx context.Context, tx pgx.Tx, employeeID uuid.UUID, employeeData map[string]interface{}) error {
	payload, err := json.Marshal(map[string]interface{}{
		"employee_id":   employeeID.String(),
		"employee_data": employeeData,
		"created_at":    time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal employee created event payload: %w", err)
	}

	req := &CreateEventRequest{
		AggregateID:   employeeID,
		AggregateType: AggregateTypeEmployee,
		EventType:     EventTypeEmployeeCreated,
		EventVersion:  1,
		Payload:       payload,
	}

	_, err = s.CreateEventWithTransaction(ctx, tx, req)
	return err
}

// CreateEmployeeUpdatedEventWithTransaction 在事务中创建员工更新事件
func (s *Service) CreateEmployeeUpdatedEventWithTransaction(ctx context.Context, tx pgx.Tx, employeeID uuid.UUID, updatedFields map[string]interface{}) error {
	payload, err := json.Marshal(map[string]interface{}{
		"employee_id":    employeeID.String(),
		"updated_fields": updatedFields,
		"updated_at":     time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal employee updated event payload: %w", err)
	}

	req := &CreateEventRequest{
		AggregateID:   employeeID,
		AggregateType: AggregateTypeEmployee,
		EventType:     EventTypeEmployeeUpdated,
		EventVersion:  1,
		Payload:       payload,
	}

	_, err = s.CreateEventWithTransaction(ctx, tx, req)
	return err
}

// GetUnprocessedEvents 获取未处理的事件
func (s *Service) GetUnprocessedEvents(ctx context.Context, limit int) ([]Event, error) {
	return s.repo.GetUnprocessedEvents(ctx, limit)
}

// GetEventsByAggregateID 根据聚合ID获取事件
func (s *Service) GetEventsByAggregateID(ctx context.Context, aggregateID uuid.UUID) ([]Event, error) {
	return s.repo.GetEventsByAggregateID(ctx, aggregateID)
}

// GetEventsByType 根据事件类型获取事件
func (s *Service) GetEventsByType(ctx context.Context, eventType string, limit int) ([]Event, error) {
	return s.repo.GetEventsByType(ctx, eventType, limit)
}

// GetEvents 获取所有事件
func (s *Service) GetEvents(ctx context.Context, limit int) ([]Event, error) {
	return s.repo.GetEvents(ctx, limit)
}
