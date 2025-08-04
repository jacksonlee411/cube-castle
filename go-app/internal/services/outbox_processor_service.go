package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/events"
	positionEvents "github.com/gaogu/cube-castle/go-app/internal/cqrs/events"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
)

// OutboxProcessorService Outbox事件处理服务
type OutboxProcessorService struct {
	outboxRepo   repositories.OutboxRepository
	eventBus     events.EventBus
	logger       *logging.StructuredLogger
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	isRunning    bool
	runningMutex sync.RWMutex
	
	// 配置
	batchSize       int
	processingDelay time.Duration
	maxRetries      int
	retryDelay      time.Duration
}

// OutboxProcessorConfig Outbox处理器配置
type OutboxProcessorConfig struct {
	BatchSize       int           `json:"batch_size"`
	ProcessingDelay time.Duration `json:"processing_delay"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
}

// DefaultOutboxProcessorConfig 默认配置
func DefaultOutboxProcessorConfig() *OutboxProcessorConfig {
	return &OutboxProcessorConfig{
		BatchSize:       50,
		ProcessingDelay: 5 * time.Second,
		MaxRetries:      5,
		RetryDelay:      30 * time.Second,
	}
}

// NewOutboxProcessorService 创建Outbox处理服务
func NewOutboxProcessorService(
	outboxRepo repositories.OutboxRepository,
	eventBus events.EventBus,
	logger *logging.StructuredLogger,
	config *OutboxProcessorConfig,
) *OutboxProcessorService {
	if config == nil {
		config = DefaultOutboxProcessorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &OutboxProcessorService{
		outboxRepo:      outboxRepo,
		eventBus:        eventBus,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		isRunning:       false,
		batchSize:       config.BatchSize,
		processingDelay: config.ProcessingDelay,
		maxRetries:      config.MaxRetries,
		retryDelay:      config.RetryDelay,
	}
}

// Start 启动Outbox处理器
func (s *OutboxProcessorService) Start() error {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if s.isRunning {
		return fmt.Errorf("outbox processor is already running")
	}

	s.logger.Info("Starting outbox processor service", 
		"batch_size", s.batchSize,
		"processing_delay", s.processingDelay)

	// 启动主处理循环
	s.wg.Add(1)
	go s.processingLoop()

	// 启动清理任务
	s.wg.Add(1)
	go s.cleanupLoop()

	s.isRunning = true
	s.logger.Info("Outbox processor service started successfully")
	return nil
}

// Stop 停止Outbox处理器
func (s *OutboxProcessorService) Stop() error {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if !s.isRunning {
		return nil
	}

	s.logger.Info("Stopping outbox processor service...")

	// 取消上下文
	s.cancel()

	// 等待所有goroutine结束
	s.wg.Wait()

	s.isRunning = false
	s.logger.Info("Outbox processor service stopped successfully")
	return nil
}

// Health 健康检查
func (s *OutboxProcessorService) Health() error {
	s.runningMutex.RLock()
	defer s.runningMutex.RUnlock()

	if !s.isRunning {
		return fmt.Errorf("outbox processor is not running")
	}

	select {
	case <-s.ctx.Done():
		return fmt.Errorf("outbox processor context is cancelled")
	default:
		return nil
	}
}

// processingLoop 主处理循环
func (s *OutboxProcessorService) processingLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.processingDelay)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Outbox processing loop stopped")
			return
		case <-ticker.C:
			if err := s.processPendingEvents(); err != nil {
				s.logger.Error("Failed to process pending events", "error", err)
			}
		}
	}
}

// cleanupLoop 清理循环
func (s *OutboxProcessorService) cleanupLoop() {
	defer s.wg.Done()

	// 每小时清理一次
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Outbox cleanup loop stopped")
			return
		case <-ticker.C:
			if err := s.cleanupProcessedEvents(); err != nil {
				s.logger.Error("Failed to cleanup processed events", "error", err)
			}
		}
	}
}

// processPendingEvents 处理待发布的事件
func (s *OutboxProcessorService) processPendingEvents() error {
	events, err := s.outboxRepo.GetPendingEvents(s.ctx, s.batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	s.logger.Info("Processing outbox events", "count", len(events))

	processed := 0
	failed := 0

	for _, event := range events {
		if err := s.processEvent(event); err != nil {
			s.logger.Error("Failed to process outbox event", 
				"event_id", event.ID,
				"event_type", event.EventType,
				"attempt_count", event.AttemptCount,
				"error", err)

			// 增加重试次数
			if incrementErr := s.outboxRepo.IncrementAttemptCount(s.ctx, event.ID); incrementErr != nil {
				s.logger.Error("Failed to increment attempt count", 
					"event_id", event.ID, "error", incrementErr)
			}

			// 如果超过最大重试次数，标记为失败
			if event.AttemptCount >= s.maxRetries {
				if markErr := s.outboxRepo.MarkEventAsFailed(s.ctx, event.ID, err.Error()); markErr != nil {
					s.logger.Error("Failed to mark event as failed", 
						"event_id", event.ID, "error", markErr)
				}
			}

			failed++
			continue
		}

		// 标记为已处理
		if err := s.outboxRepo.MarkEventAsProcessed(s.ctx, event.ID); err != nil {
			s.logger.Error("Failed to mark event as processed", 
				"event_id", event.ID, "error", err)
			failed++
			continue
		}

		processed++
	}

	s.logger.Info("Outbox events processing completed", 
		"total", len(events),
		"processed", processed,
		"failed", failed)

	return nil
}

// processEvent 处理单个事件
func (s *OutboxProcessorService) processEvent(outboxEvent repositories.OutboxEvent) error {
	// 将发件箱事件转换为领域事件
	domainEvent, err := s.convertToDomainEvent(outboxEvent)
	if err != nil {
		return fmt.Errorf("failed to convert to domain event: %w", err)
	}

	// 发布到事件总线
	if err := s.eventBus.Publish(s.ctx, domainEvent); err != nil {
		return fmt.Errorf("failed to publish event to event bus: %w", err)
	}

	s.logger.Info("Successfully processed outbox event", 
		"event_id", outboxEvent.ID,
		"event_type", outboxEvent.EventType,
		"aggregate_id", outboxEvent.AggregateID)

	return nil
}

// convertToDomainEvent 将发件箱事件转换为领域事件
func (s *OutboxProcessorService) convertToDomainEvent(outboxEvent repositories.OutboxEvent) (events.DomainEvent, error) {
	switch outboxEvent.EventType {
	case "position.created":
		var event positionEvents.PositionCreatedEvent
		if err := json.Unmarshal(outboxEvent.EventData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal position.created event: %w", err)
		}
		return &event, nil

	case "position.updated":
		var event positionEvents.PositionUpdatedEvent
		if err := json.Unmarshal(outboxEvent.EventData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal position.updated event: %w", err)
		}
		return &event, nil

	case "position.deleted":
		var event positionEvents.PositionDeletedEvent
		if err := json.Unmarshal(outboxEvent.EventData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal position.deleted event: %w", err)
		}
		return &event, nil

	case "employee.assigned_to_position":
		var event positionEvents.EmployeeAssignedToPositionEvent
		if err := json.Unmarshal(outboxEvent.EventData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal employee.assigned_to_position event: %w", err)
		}
		return &event, nil

	case "employee.removed_from_position":
		var event positionEvents.EmployeeRemovedFromPositionEvent
		if err := json.Unmarshal(outboxEvent.EventData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal employee.removed_from_position event: %w", err)
		}
		return &event, nil

	default:
		return nil, fmt.Errorf("unknown event type: %s", outboxEvent.EventType)
	}
}

// cleanupProcessedEvents 清理已处理的事件
func (s *OutboxProcessorService) cleanupProcessedEvents() error {
	// 清理7天前的已处理事件
	olderThan := time.Now().Add(-7 * 24 * time.Hour)
	
	err := s.outboxRepo.CleanupProcessedEvents(s.ctx, olderThan)
	if err != nil {
		return fmt.Errorf("failed to cleanup processed events: %w", err)
	}

	s.logger.Info("Cleaned up processed outbox events", "older_than", olderThan)
	return nil
}

// ProcessEventNow 立即处理单个事件（用于测试或手动触发）
func (s *OutboxProcessorService) ProcessEventNow(eventID string) error {
	// 实现单个事件的立即处理逻辑
	// 这里简化实现，实际可以根据需要扩展
	return s.processPendingEvents()
}

// GetStatus 获取处理器状态
func (s *OutboxProcessorService) GetStatus() map[string]interface{} {
	s.runningMutex.RLock()
	defer s.runningMutex.RUnlock()

	return map[string]interface{}{
		"is_running":        s.isRunning,
		"batch_size":        s.batchSize,
		"processing_delay":  s.processingDelay.String(),
		"max_retries":       s.maxRetries,
		"retry_delay":       s.retryDelay.String(),
	}
}