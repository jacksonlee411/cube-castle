package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/events"
	positionEvents "github.com/gaogu/cube-castle/go-app/internal/cqrs/events"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
)

// OutboxEvent 发件箱事件表
type OutboxEvent struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	TenantID     uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	EventType    string          `json:"event_type" db:"event_type"`
	AggregateID  uuid.UUID       `json:"aggregate_id" db:"aggregate_id"`
	EventData    json.RawMessage `json:"event_data" db:"event_data"`
	Status       string          `json:"status" db:"status"` // PENDING, PROCESSED, FAILED
	AttemptCount int             `json:"attempt_count" db:"attempt_count"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	ProcessedAt  *time.Time      `json:"processed_at,omitempty" db:"processed_at"`
	ErrorMessage *string         `json:"error_message,omitempty" db:"error_message"`
}

// OutboxRepository 发件箱仓储接口
type OutboxRepository interface {
	// 在业务事务中保存事件到发件箱
	SaveEventInTransaction(ctx context.Context, tx Transaction, event OutboxEvent) error
	
	// 获取待处理的事件
	GetPendingEvents(ctx context.Context, limit int) ([]OutboxEvent, error)
	
	// 标记事件为已处理
	MarkEventAsProcessed(ctx context.Context, eventID uuid.UUID) error
	
	// 标记事件处理失败
	MarkEventAsFailed(ctx context.Context, eventID uuid.UUID, errorMsg string) error
	
	// 重试失败的事件
	IncrementAttemptCount(ctx context.Context, eventID uuid.UUID) error
	
	// 清理旧的已处理事件
	CleanupProcessedEvents(ctx context.Context, olderThan time.Time) error
}

// Transaction 事务接口
type Transaction interface {
	Commit() error
	Rollback() error
}

// PositionCommandRepositoryWithOutbox 增强的职位命令仓储，支持Outbox模式
type PositionCommandRepositoryWithOutbox interface {
	PositionCommandRepository
	
	// 原子操作：在同一事务中创建职位并保存事件
	CreatePositionWithEvent(ctx context.Context, position Position, event OutboxEvent) error
}

// OutboxEventProcessor 发件箱事件处理器
type OutboxEventProcessor struct {
	outboxRepo OutboxRepository
	eventBus   events.EventBus
	logger     *logging.StructuredLogger
}

// ProcessPendingEvents 处理待发布的事件
func (p *OutboxEventProcessor) ProcessPendingEvents(ctx context.Context) error {
	events, err := p.outboxRepo.GetPendingEvents(ctx, 100)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := p.processEvent(ctx, event); err != nil {
			p.logger.Error("Failed to process outbox event", "event_id", event.ID, "error", err)
			
			// 标记失败，增加重试次数
			p.outboxRepo.MarkEventAsFailed(ctx, event.ID, err.Error())
			p.outboxRepo.IncrementAttemptCount(ctx, event.ID)
			continue
		}
		
		// 标记为已处理
		p.outboxRepo.MarkEventAsProcessed(ctx, event.ID)
	}
	
	return nil
}

func (p *OutboxEventProcessor) processEvent(ctx context.Context, outboxEvent OutboxEvent) error {
	// 将发件箱事件转换为领域事件
	domainEvent, err := p.convertToDomainEvent(outboxEvent)
	if err != nil {
		return err
	}
	
	// 发布到事件总线
	return p.eventBus.Publish(ctx, domainEvent)
}

func (p *OutboxEventProcessor) convertToDomainEvent(outboxEvent OutboxEvent) (events.DomainEvent, error) {
	switch outboxEvent.EventType {
	case "position.created":
		var event positionEvents.PositionCreatedEvent
		if err := json.Unmarshal(outboxEvent.EventData, &event); err != nil {
			return nil, err
		}
		return &event, nil
	case "employee.assigned_to_position":
		var event positionEvents.EmployeeAssignedToPositionEvent
		if err := json.Unmarshal(outboxEvent.EventData, &event); err != nil {
			return nil, err
		}
		return &event, nil
	// 其他事件类型...
	default:
		return nil, fmt.Errorf("unknown event type: %s", outboxEvent.EventType)
	}
}