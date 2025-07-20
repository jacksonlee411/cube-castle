package outbox

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// OutboxProcessor 发件箱处理器
type OutboxProcessor struct {
	repo           *Repository
	eventProcessor *EventProcessor
	config         *ProcessorConfig
}

// ProcessorConfig 处理器配置
type ProcessorConfig struct {
	BatchSize           int           `json:"batch_size"`
	PollingInterval     time.Duration `json:"polling_interval"`
	MaxRetries          int           `json:"max_retries"`
	RetryDelay          time.Duration `json:"retry_delay"`
	CleanupInterval     time.Duration `json:"cleanup_interval"`
	CleanupOlderThan    time.Duration `json:"cleanup_older_than"`
	EnableMetrics       bool          `json:"enable_metrics"`
	EnableDeadLetter    bool          `json:"enable_dead_letter"`
	DeadLetterThreshold int           `json:"dead_letter_threshold"`
}

// DefaultProcessorConfig 默认处理器配置
func DefaultProcessorConfig() *ProcessorConfig {
	return &ProcessorConfig{
		BatchSize:           100,
		PollingInterval:     5 * time.Second,
		MaxRetries:          3,
		RetryDelay:          1 * time.Second,
		CleanupInterval:     1 * time.Hour,
		CleanupOlderThan:    24 * time.Hour,
		EnableMetrics:       true,
		EnableDeadLetter:    true,
		DeadLetterThreshold: 5,
	}
}

// NewOutboxProcessor 创建新的发件箱处理器
func NewOutboxProcessor(repo *Repository, eventProcessor *EventProcessor, config *ProcessorConfig) *OutboxProcessor {
	if config == nil {
		config = DefaultProcessorConfig()
	}
	
	return &OutboxProcessor{
		repo:           repo,
		eventProcessor: eventProcessor,
		config:         config,
	}
}

// Start 启动处理器
func (p *OutboxProcessor) Start(ctx context.Context) error {
	log.Println("🚀 Starting Outbox Processor...")
	
	// 启动事件处理循环
	go p.processEventsLoop(ctx)
	
	// 启动清理循环
	go p.cleanupLoop(ctx)
	
	log.Println("✅ Outbox Processor started successfully")
	return nil
}

// processEventsLoop 事件处理循环
func (p *OutboxProcessor) processEventsLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.PollingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Outbox Processor stopped")
			return
		case <-ticker.C:
			if err := p.ProcessEvents(ctx); err != nil {
				log.Printf("❌ Error processing events: %v", err)
			}
		}
	}
}

// cleanupLoop 清理循环
func (p *OutboxProcessor) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.cleanupProcessedEvents(ctx); err != nil {
				log.Printf("❌ Error cleaning up events: %v", err)
			}
		}
	}
}

// ProcessEvents 处理事件
func (p *OutboxProcessor) ProcessEvents(ctx context.Context) error {
	// 获取未处理的事件
	events, err := p.repo.GetUnprocessedEvents(ctx, p.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to get unprocessed events: %w", err)
	}
	
	if len(events) == 0 {
		return nil // 没有事件需要处理
	}
	
	log.Printf("📦 Processing %d events", len(events))
	
	// 处理每个事件
	for _, event := range events {
		if err := p.processEvent(ctx, &event); err != nil {
			log.Printf("❌ Failed to process event %s: %v", event.ID, err)
			// 继续处理下一个事件，不中断整个批次
		}
	}
	
	return nil
}

// processEvent 处理单个事件
func (p *OutboxProcessor) processEvent(ctx context.Context, event *Event) error {
	// 获取事件处理器
	handler, exists := p.eventProcessor.GetHandler(event.EventType)
	if !exists {
		log.Printf("⚠️ No handler found for event type: %s", event.EventType)
		// 标记为已处理，避免重复处理
		return p.repo.MarkEventAsProcessed(ctx, event.ID)
	}
	
	// 处理事件
	if err := handler.HandleEvent(ctx, event); err != nil {
		return fmt.Errorf("handler failed for event %s: %w", event.ID, err)
	}
	
	// 标记事件为已处理
	if err := p.repo.MarkEventAsProcessed(ctx, event.ID); err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}
	
	log.Printf("✅ Successfully processed event %s (%s)", event.ID, event.EventType)
	return nil
}

// cleanupProcessedEvents 清理已处理的事件
func (p *OutboxProcessor) cleanupProcessedEvents(ctx context.Context) error {
	deletedCount, err := p.repo.DeleteProcessedEvents(ctx, p.config.CleanupOlderThan)
	if err != nil {
		return fmt.Errorf("failed to cleanup processed events: %w", err)
	}
	
	if deletedCount > 0 {
		log.Printf("🧹 Cleaned up %d processed events", deletedCount)
	}
	
	return nil
}

// GetStats 获取处理器统计信息
func (p *OutboxProcessor) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats, err := p.repo.GetEventStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}
	
	// 添加处理器配置信息
	stats["processor_config"] = p.config
	
	return stats, nil
}

// CreateEvent 创建事件（事务性）
func (p *OutboxProcessor) CreateEvent(ctx context.Context, req *CreateEventRequest) (*Event, error) {
	event := &Event{
		ID:            uuid.New(),
		AggregateID:   req.AggregateID,
		AggregateType: req.AggregateType,
		EventType:     req.EventType,
		EventVersion:  req.EventVersion,
		Payload:       req.Payload,
		Metadata:      req.Metadata,
		CreatedAt:     time.Now(),
	}
	
	if err := p.repo.CreateEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	
	log.Printf("📝 Created event %s (%s)", event.ID, event.EventType)
	return event, nil
}

// CreateEventWithTransaction 在事务中创建事件
func (p *OutboxProcessor) CreateEventWithTransaction(ctx context.Context, tx pgx.Tx, req *CreateEventRequest) (*Event, error) {
	event := &Event{
		ID:            uuid.New(),
		AggregateID:   req.AggregateID,
		AggregateType: req.AggregateType,
		EventType:     req.EventType,
		EventVersion:  req.EventVersion,
		Payload:       req.Payload,
		Metadata:      req.Metadata,
		CreatedAt:     time.Now(),
	}
	
	query := `
		INSERT INTO outbox.events (id, aggregate_id, aggregate_type, event_type, event_version, payload, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	_, err := tx.Exec(ctx, query,
		event.ID,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.EventVersion,
		event.Payload,
		event.Metadata,
		event.CreatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create event in transaction: %w", err)
	}
	
	return event, nil
}

// ReplayEvents 重放事件
func (p *OutboxProcessor) ReplayEvents(ctx context.Context, aggregateID uuid.UUID) error {
	events, err := p.repo.GetEventsByAggregateID(ctx, aggregateID)
	if err != nil {
		return fmt.Errorf("failed to get events for replay: %w", err)
	}
	
	log.Printf("🔄 Replaying %d events for aggregate %s", len(events), aggregateID)
	
	for _, event := range events {
		if err := p.processEvent(ctx, &event); err != nil {
			log.Printf("❌ Failed to replay event %s: %v", event.ID, err)
			return err
		}
	}
	
	log.Printf("✅ Successfully replayed %d events", len(events))
	return nil
}

// ReplayEventsByType 根据类型重放事件
func (p *OutboxProcessor) ReplayEventsByType(ctx context.Context, eventType string, limit int) error {
	events, err := p.repo.GetEventsByType(ctx, eventType, limit)
	if err != nil {
		return fmt.Errorf("failed to get events for replay: %w", err)
	}
	
	log.Printf("🔄 Replaying %d events of type %s", len(events), eventType)
	
	for _, event := range events {
		if err := p.processEvent(ctx, &event); err != nil {
			log.Printf("❌ Failed to replay event %s: %v", event.ID, err)
			return err
		}
	}
	
	log.Printf("✅ Successfully replayed %d events", len(events))
	return nil
} 