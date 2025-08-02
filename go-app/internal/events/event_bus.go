package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventBus 事件总线接口 - CQRS架构的核心组件
type EventBus interface {
	// Publish 发布单个领域事件
	Publish(ctx context.Context, event DomainEvent) error
	
	// PublishBatch 批量发布事件 (性能优化)
	PublishBatch(ctx context.Context, events []DomainEvent) error
	
	// Subscribe 订阅特定类型的事件
	Subscribe(ctx context.Context, eventType string, handler EventHandler) error
	
	// Start 启动事件总线
	Start(ctx context.Context) error
	
	// Stop 停止事件总线
	Stop() error
	
	// Health 健康检查
	Health() error
}

// DomainEvent 领域事件接口 - 所有业务事件的基础
type DomainEvent interface {
	// 事件标识
	GetEventID() uuid.UUID
	GetEventType() string
	GetEventVersion() string
	
	// 聚合信息
	GetAggregateID() uuid.UUID
	GetAggregateType() string
	
	// 租户信息
	GetTenantID() uuid.UUID
	
	// 时间信息
	GetTimestamp() time.Time
	GetOccurredAt() time.Time
	
	// 序列化
	Serialize() ([]byte, error)
	GetHeaders() map[string]string
	
	// 事件元数据
	GetMetadata() map[string]interface{}
	GetCorrelationID() string
	GetCausationID() string
}

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(ctx context.Context, event DomainEvent) error
	GetEventType() string
	GetHandlerName() string
}

// EventPublisher 事件发布器 - 简化的发布接口
type EventPublisher interface {
	PublishEvent(ctx context.Context, event DomainEvent) error
}

// EventSubscriber 事件订阅器 - 简化的订阅接口  
type EventSubscriber interface {
	SubscribeToEvent(ctx context.Context, eventType string, handler EventHandler) error
}

// BaseDomainEvent 领域事件基础实现
type BaseDomainEvent struct {
	EventID       uuid.UUID              `json:"event_id"`
	EventType     string                 `json:"event_type"`
	EventVersion  string                 `json:"event_version"`
	AggregateID   uuid.UUID              `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	Timestamp     time.Time              `json:"timestamp"`
	OccurredAt    time.Time              `json:"occurred_at"`
	Metadata      map[string]interface{} `json:"metadata"`
	CorrelationID string                 `json:"correlation_id"`
	CausationID   string                 `json:"causation_id"`
}

// GetEventID 实现DomainEvent接口
func (e *BaseDomainEvent) GetEventID() uuid.UUID {
	return e.EventID
}

func (e *BaseDomainEvent) GetEventType() string {
	return e.EventType
}

func (e *BaseDomainEvent) GetEventVersion() string {
	return e.EventVersion
}

func (e *BaseDomainEvent) GetAggregateID() uuid.UUID {
	return e.AggregateID
}

func (e *BaseDomainEvent) GetAggregateType() string {
	return e.AggregateType
}

func (e *BaseDomainEvent) GetTenantID() uuid.UUID {
	return e.TenantID
}

func (e *BaseDomainEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

func (e *BaseDomainEvent) GetOccurredAt() time.Time {
	return e.OccurredAt
}

func (e *BaseDomainEvent) GetMetadata() map[string]interface{} {
	return e.Metadata
}

func (e *BaseDomainEvent) GetCorrelationID() string {
	return e.CorrelationID
}

func (e *BaseDomainEvent) GetCausationID() string {
	return e.CausationID
}

func (e *BaseDomainEvent) GetHeaders() map[string]string {
	return map[string]string{
		"event-type":      e.EventType,
		"event-version":   e.EventVersion,
		"aggregate-type":  e.AggregateType,
		"tenant-id":       e.TenantID.String(),
		"correlation-id":  e.CorrelationID,
		"causation-id":    e.CausationID,
		"timestamp":       e.Timestamp.Format(time.RFC3339),
		"occurred-at":     e.OccurredAt.Format(time.RFC3339),
	}
}

// Serialize 实现DomainEvent接口的序列化方法
func (e *BaseDomainEvent) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// NewBaseDomainEvent 创建基础领域事件
func NewBaseDomainEvent(eventType, aggregateType string, aggregateID, tenantID uuid.UUID) *BaseDomainEvent {
	now := time.Now()
	return &BaseDomainEvent{
		EventID:       uuid.New(),
		EventType:     eventType,
		EventVersion:  "v1.0",
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		TenantID:      tenantID,
		Timestamp:     now,
		OccurredAt:    now,
		Metadata:      make(map[string]interface{}),
		CorrelationID: uuid.New().String(),
		CausationID:   "",
	}
}

// EventBusConfig 事件总线配置
type EventBusConfig struct {
	// Kafka配置
	KafkaBootstrapServers string
	KafkaTopicPrefix      string
	KafkaConsumerGroup    string
	
	// 性能配置
	BatchSize           int
	BatchTimeout        time.Duration
	MaxRetries          int
	RetryBackoff        time.Duration
	
	// 监控配置
	EnableMetrics       bool
	MetricsPrefix       string
	
	// 安全配置
	EnableTLS           bool
	TLSConfig           *TLSConfig
}

// TLSConfig TLS配置
type TLSConfig struct {
	CertFile   string
	KeyFile    string
	CAFile     string
	SkipVerify bool
}

// DefaultEventBusConfig 默认配置
func DefaultEventBusConfig() *EventBusConfig {
	return &EventBusConfig{
		KafkaBootstrapServers: "localhost:9092",
		KafkaTopicPrefix:      "cube_castle",
		KafkaConsumerGroup:    "cube_castle_consumers",
		BatchSize:             100,
		BatchTimeout:          time.Millisecond * 100,
		MaxRetries:            3,
		RetryBackoff:          time.Second * 2,
		EnableMetrics:         true,
		MetricsPrefix:         "cube_castle_eventbus",
		EnableTLS:             false,
	}
}