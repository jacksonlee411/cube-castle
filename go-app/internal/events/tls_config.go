package events

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
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