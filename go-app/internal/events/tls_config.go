package events

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"

	"github.com/gaogu/cube-castle/go-app/internal/logging"
)

// newTLSConfig 创建TLS配置
func newTLSConfig(config *TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.SkipVerify,
	}

	// 加载客户端证书
	if config.CertFile != "" && config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// 加载CA证书
	if config.CAFile != "" {
		caCert, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// EventBusFactory 事件总线工厂
type EventBusFactory struct{}

// NewEventBusFactory 创建事件总线工厂
func NewEventBusFactory() *EventBusFactory {
	return &EventBusFactory{}
}

// CreateKafkaEventBus 创建Kafka事件总线（工厂方法）
func (f *EventBusFactory) CreateKafkaEventBus(config *EventBusConfig) (EventBus, error) {
	return NewKafkaEventBus(config)
}

// CreateMockEventBus 创建Mock事件总线（用于测试）
func (f *EventBusFactory) CreateMockEventBus() EventBus {
	return &MockEventBus{
		events: make([]DomainEvent, 0),
	}
}

// CreateInMemoryEventBus 创建InMemory事件总线（用于开发环境实际事件处理）
func (f *EventBusFactory) CreateInMemoryEventBus() EventBus {
	logger := logging.NewStructuredLogger()
	return NewInMemoryEventBus(logger)
}

// MockEventBus Mock事件总线实现（用于测试）
type MockEventBus struct {
	events   []DomainEvent
	handlers map[string][]EventHandler
}

func (m *MockEventBus) Publish(ctx context.Context, event DomainEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventBus) PublishBatch(ctx context.Context, events []DomainEvent) error {
	m.events = append(m.events, events...)
	return nil
}

func (m *MockEventBus) Subscribe(ctx context.Context, eventType string, handler EventHandler) error {
	if m.handlers == nil {
		m.handlers = make(map[string][]EventHandler)
	}
	m.handlers[eventType] = append(m.handlers[eventType], handler)
	return nil
}

func (m *MockEventBus) Start(ctx context.Context) error {
	return nil
}

func (m *MockEventBus) Stop() error {
	return nil
}

func (m *MockEventBus) Health() error {
	return nil
}

// GetPublishedEvents 获取已发布的事件（测试用）
func (m *MockEventBus) GetPublishedEvents() []DomainEvent {
	return m.events
}

// ClearEvents 清空事件（测试用）
func (m *MockEventBus) ClearEvents() {
	m.events = make([]DomainEvent, 0)
}

// InMemoryEventBus 内存事件总线实现 (用于开发和测试)
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	events   []DomainEvent
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
		handlers: make(map[string][]EventHandler),
		events:   make([]DomainEvent, 0),
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
func (e *InMemoryEventBus) Publish(ctx context.Context, event DomainEvent) error {
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
func (e *InMemoryEventBus) PublishBatch(ctx context.Context, events []DomainEvent) error {
	for _, event := range events {
		if err := e.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe 订阅事件
func (e *InMemoryEventBus) Subscribe(ctx context.Context, eventType string, handler EventHandler) error {
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
func (e *InMemoryEventBus) GetEvents() []DomainEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	// 返回事件副本
	eventsCopy := make([]DomainEvent, len(e.events))
	copy(eventsCopy, e.events)
	return eventsCopy
}

// ClearEvents 清空事件 (用于测试)
func (e *InMemoryEventBus) ClearEvents() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = make([]DomainEvent, 0)
}

// processEvent 处理事件
func (e *InMemoryEventBus) processEvent(ctx context.Context, event DomainEvent, handlers []EventHandler) {
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