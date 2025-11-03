package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Logger 定义 Plan 216 所需的最小日志接口。
// Plan 218 将提供实际实现；当前使用 noopLogger 作为兜底。
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// MetricsRecorder 用于记录事件发布的关键指标。
// 若调用方未提供实现，将使用 noopMetrics 确保接口稳定。
type MetricsRecorder interface {
	RecordSuccess(eventType string)
	RecordFailure(eventType string)
	RecordNoHandler(eventType string)
	RecordLatency(eventType string, duration time.Duration)
}

// MemoryEventBus 提供并发安全的内存事件总线实现。
type MemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
	logger   Logger
	metrics  MetricsRecorder
}

// NewMemoryEventBus 创建新的内存事件总线实例。
// 当 logger 或 metrics 为 nil 时，会注入对应的 noop 实现。
func NewMemoryEventBus(logger Logger, metrics MetricsRecorder) *MemoryEventBus {
	if logger == nil {
		logger = &noopLogger{}
	}
	if metrics == nil {
		metrics = &noopMetrics{}
	}

	return &MemoryEventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger,
		metrics:  metrics,
	}
}

// Subscribe 为指定事件类型添加处理器。
func (b *MemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	if eventType == "" {
		return ErrEmptyEventType
	}
	if handler == nil {
		return ErrNilHandler
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	b.logger.Infof("subscribed to event type: %s (total handlers: %d)", eventType, len(b.handlers[eventType]))

	return nil
}

// Publish 将事件广播至所有订阅者。
// 若任何处理器返回错误，将聚合失败信息同时继续执行其余处理器。
func (b *MemoryEventBus) Publish(ctx context.Context, event Event) error {
	if event == nil {
		return ErrNilEvent
	}

	eventType := event.EventType()
	if eventType == "" {
		return ErrEmptyEventType
	}

	b.mu.RLock()
	handlers, ok := b.handlers[eventType]
	b.mu.RUnlock()

	if !ok || len(handlers) == 0 {
		b.logger.Debugf("no handlers for event type: %s", eventType)
		b.metrics.RecordNoHandler(eventType)
		return nil
	}

	start := time.Now()
	aggErr := NewAggregatePublishError(eventType, event.AggregateID())

	for idx, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			b.logger.Errorf("event handler failed: type=%s, handler_index=%d, error=%v", eventType, idx, err)
			aggErr.Append(idx, err)
			b.metrics.RecordFailure(eventType)
			continue
		}

		b.metrics.RecordSuccess(eventType)
	}

	b.metrics.RecordLatency(eventType, time.Since(start))

	if aggErr.IsEmpty() {
		return nil
	}

	return aggErr
}

// GetHandlerCount 返回指定事件类型的处理器数量，便于测试与监控。
func (b *MemoryEventBus) GetHandlerCount(eventType string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers[eventType])
}

// Reset 清空所有订阅，主要用于测试场景。
func (b *MemoryEventBus) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = make(map[string][]EventHandler)
}

type noopLogger struct{}

func (*noopLogger) Debugf(string, ...interface{}) {}
func (*noopLogger) Infof(string, ...interface{})  {}
func (*noopLogger) Errorf(string, ...interface{}) {}

type noopMetrics struct{}

func (*noopMetrics) RecordSuccess(string)                {}
func (*noopMetrics) RecordFailure(string)                {}
func (*noopMetrics) RecordNoHandler(string)              {}
func (*noopMetrics) RecordLatency(string, time.Duration) {}

// AggregatePublishError 聚合发布过程中出现的所有失败。
type AggregatePublishError struct {
	eventType   string
	aggregateID string
	failures    []HandlerFailure
}

// HandlerFailure 表示单个处理器的失败信息。
type HandlerFailure struct {
	Index int
	Err   error
}

// NewAggregatePublishError 构造新的聚合错误实例。
func NewAggregatePublishError(eventType, aggregateID string) *AggregatePublishError {
	return &AggregatePublishError{
		eventType:   eventType,
		aggregateID: aggregateID,
	}
}

// Append 追加失败信息。
func (e *AggregatePublishError) Append(idx int, err error) {
	e.failures = append(e.failures, HandlerFailure{Index: idx, Err: err})
}

func (e *AggregatePublishError) Error() string {
	return fmt.Sprintf("eventbus publish failed: type=%s aggregateID=%s failures=%d",
		e.eventType, e.aggregateID, len(e.failures))
}

// Failures 返回所有失败的处理器详情。
func (e *AggregatePublishError) Failures() []HandlerFailure {
	return e.failures
}

// IsEmpty 表示是否存在失败。
func (e *AggregatePublishError) IsEmpty() bool {
	return len(e.failures) == 0
}

// EventType 返回发生错误的事件类型。
func (e *AggregatePublishError) EventType() string {
	return e.eventType
}

// AggregateID 返回聚合根标识。
func (e *AggregatePublishError) AggregateID() string {
	return e.aggregateID
}
