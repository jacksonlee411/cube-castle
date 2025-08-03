package eventbus

import (
	"context"
	"fmt"
	"sync"

	"github.com/gaogu/cube-castle/go-app/internal/events"
)

// InMemoryEventBus 内存事件总线实现 (用于开发和测试)
type InMemoryEventBus struct {
	handlers map[string][]events.EventHandler
	events   []events.DomainEvent
	logger   Logger
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
}

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewInMemoryEventBus 创建内存事件总线
func NewInMemoryEventBus(logger Logger) *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]events.EventHandler),
		events:   make([]events.DomainEvent, 0),
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// Start 启动事件总线
func (e *InMemoryEventBus) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("event bus is already running")
	}

	e.running = true
	e.logger.Info("Starting in-memory event bus")
	return nil
}

// Stop 停止事件总线
func (e *InMemoryEventBus) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return nil
	}

	e.logger.Info("Stopping in-memory event bus")
	close(e.stopChan)
	e.running = false
	return nil
}

// Publish 发布单个事件
func (e *InMemoryEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return fmt.Errorf("event bus is not running")
	}

	// 存储事件
	e.events = append(e.events, event)

	// 查找处理器
	handlers, exists := e.handlers[event.GetEventType()]
	if !exists {
		e.logger.Warn("No handlers found for event type", "event_type", event.GetEventType())
		return nil
	}

	// 异步处理事件
	go e.processEvent(ctx, event, handlers)

	e.logger.Info("Event published", "event_id", event.GetEventID(), "event_type", event.GetEventType())
	return nil
}

// PublishBatch 批量发布事件
func (e *InMemoryEventBus) PublishBatch(ctx context.Context, events []events.DomainEvent) error {
	for _, event := range events {
		if err := e.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe 订阅事件
func (e *InMemoryEventBus) Subscribe(ctx context.Context, eventType string, handler events.EventHandler) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.handlers[eventType] = append(e.handlers[eventType], handler)
	e.logger.Info("Event handler subscribed", "event_type", eventType, "handler", handler.GetHandlerName())
	return nil
}

// Health 健康检查
func (e *InMemoryEventBus) Health() error {
	if !e.running {
		return fmt.Errorf("event bus is not running")
	}
	return nil
}

// GetEvents 获取所有事件 (用于测试)
func (e *InMemoryEventBus) GetEvents() []events.DomainEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// 返回事件副本
	eventsCopy := make([]events.DomainEvent, len(e.events))
	copy(eventsCopy, e.events)
	return eventsCopy
}

// ClearEvents 清空事件 (用于测试)
func (e *InMemoryEventBus) ClearEvents() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = make([]events.DomainEvent, 0)
}

// processEvent 处理事件
func (e *InMemoryEventBus) processEvent(ctx context.Context, event events.DomainEvent, handlers []events.EventHandler) {
	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			e.logger.Error("Event handler failed", 
				"event_type", event.GetEventType(), 
				"handler", handler.GetHandlerName(), 
				"event_id", event.GetEventID(),
				"error", err)
		} else {
			e.logger.Info("Event handled successfully", 
				"event_type", event.GetEventType(), 
				"handler", handler.GetHandlerName(), 
				"event_id", event.GetEventID())
		}
	}
}

// MockEventBus 模拟事件总线 (用于测试)
type MockEventBus struct {
	publishedEvents []events.DomainEvent
	logger          Logger
	mu              sync.RWMutex
}

// NewMockEventBus 创建模拟事件总线
func NewMockEventBus(logger Logger) *MockEventBus {
	return &MockEventBus{
		publishedEvents: make([]events.DomainEvent, 0),
		logger:          logger,
	}
}

func (m *MockEventBus) Start(ctx context.Context) error {
	m.logger.Info("Starting mock event bus")
	return nil
}

func (m *MockEventBus) Stop() error {
	m.logger.Info("Stopping mock event bus")
	return nil
}

func (m *MockEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.publishedEvents = append(m.publishedEvents, event)
	m.logger.Info("Mock event published", "event_id", event.GetEventID(), "event_type", event.GetEventType())
	return nil
}

func (m *MockEventBus) PublishBatch(ctx context.Context, events []events.DomainEvent) error {
	for _, event := range events {
		if err := m.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockEventBus) Subscribe(ctx context.Context, eventType string, handler events.EventHandler) error {
	m.logger.Info("Mock event handler subscribed", "event_type", eventType, "handler", handler.GetHandlerName())
	return nil
}

func (m *MockEventBus) Health() error {
	return nil
}

func (m *MockEventBus) GetPublishedEvents() []events.DomainEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	eventsCopy := make([]events.DomainEvent, len(m.publishedEvents))
	copy(eventsCopy, m.publishedEvents)
	return eventsCopy
}

func (m *MockEventBus) ClearEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedEvents = make([]events.DomainEvent, 0)
}