package events

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

// EventSerializer 事件序列化器接口
type EventSerializer interface {
	Serialize(event DomainEvent) ([]byte, error)
	Deserialize(data []byte, eventType string) (DomainEvent, error)
	GetSupportedEventTypes() []string
}

// JSONEventSerializer JSON序列化器实现
type JSONEventSerializer struct {
	eventRegistry map[string]reflect.Type
}

// NewJSONEventSerializer 创建JSON序列化器
func NewJSONEventSerializer() *JSONEventSerializer {
	serializer := &JSONEventSerializer{
		eventRegistry: make(map[string]reflect.Type),
	}
	
	// 注册所有事件类型
	serializer.registerDefaultEvents()
	
	return serializer
}

// registerDefaultEvents 注册默认事件类型
func (s *JSONEventSerializer) registerDefaultEvents() {
	// 员工事件
	s.RegisterEventType("employee.created", reflect.TypeOf(EmployeeCreated{}))
	s.RegisterEventType("employee.updated", reflect.TypeOf(EmployeeUpdated{}))
	s.RegisterEventType("employee.deleted", reflect.TypeOf(EmployeeDeleted{}))
	s.RegisterEventType("employee.hired", reflect.TypeOf(EmployeeHired{}))
	s.RegisterEventType("employee.terminated", reflect.TypeOf(EmployeeTerminated{}))
	s.RegisterEventType("employee.phone_updated", reflect.TypeOf(EmployeePhoneUpdated{}))
	
	// 组织事件
	s.RegisterEventType("organization.created", reflect.TypeOf(OrganizationCreated{}))
	s.RegisterEventType("organization.updated", reflect.TypeOf(OrganizationUpdated{}))
	s.RegisterEventType("organization.deleted", reflect.TypeOf(OrganizationDeleted{}))
	s.RegisterEventType("organization.restructured", reflect.TypeOf(OrganizationRestructured{}))
	s.RegisterEventType("organization.activated", reflect.TypeOf(OrganizationActivated{}))
	s.RegisterEventType("organization.deactivated", reflect.TypeOf(OrganizationDeactivated{}))
	
	// 基础事件（用于健康检查等）
	s.RegisterEventType("health.check", reflect.TypeOf(BaseDomainEvent{}))
}

// RegisterEventType 注册事件类型
func (s *JSONEventSerializer) RegisterEventType(eventType string, eventStruct reflect.Type) {
	s.eventRegistry[eventType] = eventStruct
}

// Serialize 序列化事件
func (s *JSONEventSerializer) Serialize(event DomainEvent) ([]byte, error) {
	// 使用事件自身的序列化方法
	if data, err := event.Serialize(); err == nil {
		return data, nil
	}
	
	// 如果事件没有实现序列化方法，使用JSON序列化
	return json.Marshal(event)
}

// Deserialize 反序列化事件
func (s *JSONEventSerializer) Deserialize(data []byte, eventType string) (DomainEvent, error) {
	// 查找注册的事件类型
	eventStructType, exists := s.eventRegistry[eventType]
	if !exists {
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
	
	// 创建事件实例
	eventPtr := reflect.New(eventStructType)
	eventInterface := eventPtr.Interface()
	
	// 反序列化
	if err := json.Unmarshal(data, eventInterface); err != nil {
		return nil, fmt.Errorf("failed to deserialize event %s: %w", eventType, err)
	}
	
	// 转换为DomainEvent接口
	if domainEvent, ok := eventInterface.(DomainEvent); ok {
		return domainEvent, nil
	}
	
	// 如果直接转换失败，尝试获取值而不是指针
	eventValue := eventPtr.Elem()
	if domainEvent, ok := eventValue.Interface().(DomainEvent); ok {
		return domainEvent, nil
	}
	
	return nil, fmt.Errorf("event type %s does not implement DomainEvent interface", eventType)
}

// GetSupportedEventTypes 获取支持的事件类型
func (s *JSONEventSerializer) GetSupportedEventTypes() []string {
	types := make([]string, 0, len(s.eventRegistry))
	for eventType := range s.eventRegistry {
		types = append(types, eventType)
	}
	return types
}

// AvroEventSerializer Avro序列化器实现（未来扩展）
type AvroEventSerializer struct {
	// TODO: 实现Avro序列化支持
}

// ProtobufEventSerializer Protobuf序列化器实现（未来扩展）
type ProtobufEventSerializer struct {
	// TODO: 实现Protobuf序列化支持
}

// EventSerializerFactory 事件序列化器工厂
type EventSerializerFactory struct{}

// NewEventSerializerFactory 创建序列化器工厂
func NewEventSerializerFactory() *EventSerializerFactory {
	return &EventSerializerFactory{}
}

// CreateJSONSerializer 创建JSON序列化器
func (f *EventSerializerFactory) CreateJSONSerializer() EventSerializer {
	return NewJSONEventSerializer()
}

// CreateAvroSerializer 创建Avro序列化器（未来实现）
func (f *EventSerializerFactory) CreateAvroSerializer() EventSerializer {
	// TODO: 实现Avro序列化器
	return nil
}

// CreateProtobufSerializer 创建Protobuf序列化器（未来实现）
func (f *EventSerializerFactory) CreateProtobufSerializer() EventSerializer {
	// TODO: 实现Protobuf序列化器
	return nil
}

// EventMetadata 事件元数据
type EventMetadata struct {
	SchemaVersion    string            `json:"schema_version"`
	SerializerType   string            `json:"serializer_type"`
	CompressionType  string            `json:"compression_type,omitempty"`
	CustomProperties map[string]string `json:"custom_properties,omitempty"`
}

// EnhancedDomainEvent 增强的领域事件（包含序列化元数据）
type EnhancedDomainEvent struct {
	DomainEvent
	SerializationMetadata EventMetadata `json:"serialization_metadata"`
}

// NewEnhancedDomainEvent 创建增强的领域事件
func NewEnhancedDomainEvent(event DomainEvent, serializerType string) *EnhancedDomainEvent {
	return &EnhancedDomainEvent{
		DomainEvent: event,
		SerializationMetadata: EventMetadata{
			SchemaVersion:  "v1.0",
			SerializerType: serializerType,
		},
	}
}

// BatchEventSerializer 批量事件序列化器
type BatchEventSerializer struct {
	serializer EventSerializer
}

// NewBatchEventSerializer 创建批量序列化器
func NewBatchEventSerializer(serializer EventSerializer) *BatchEventSerializer {
	return &BatchEventSerializer{
		serializer: serializer,
	}
}

// SerializeBatch 批量序列化事件
func (b *BatchEventSerializer) SerializeBatch(events []DomainEvent) ([][]byte, error) {
	results := make([][]byte, len(events))
	
	for i, event := range events {
		data, err := b.serializer.Serialize(event)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize event at index %d: %w", i, err)
		}
		results[i] = data
	}
	
	return results, nil
}

// DeserializeBatch 批量反序列化事件
func (b *BatchEventSerializer) DeserializeBatch(data [][]byte, eventTypes []string) ([]DomainEvent, error) {
	if len(data) != len(eventTypes) {
		return nil, fmt.Errorf("data length (%d) does not match event types length (%d)", len(data), len(eventTypes))
	}
	
	results := make([]DomainEvent, len(data))
	
	for i := range data {
		event, err := b.serializer.Deserialize(data[i], eventTypes[i])
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize event at index %d: %w", i, err)
		}
		results[i] = event
	}
	
	return results, nil
}

// EventValidator 事件验证器
type EventValidator struct{}

// NewEventValidator 创建事件验证器
func NewEventValidator() *EventValidator {
	return &EventValidator{}
}

// ValidateEvent 验证事件
func (v *EventValidator) ValidateEvent(event DomainEvent) error {
	// 基础验证
	if event.GetEventID() == uuid.Nil {
		return fmt.Errorf("event ID cannot be nil")
	}
	
	if event.GetEventType() == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	
	if event.GetAggregateID() == uuid.Nil {
		return fmt.Errorf("aggregate ID cannot be nil")
	}
	
	if event.GetAggregateType() == "" {
		return fmt.Errorf("aggregate type cannot be empty")
	}
	
	if event.GetTenantID() == uuid.Nil {
		return fmt.Errorf("tenant ID cannot be nil")
	}
	
	if event.GetTimestamp().IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}
	
	if event.GetOccurredAt().IsZero() {
		return fmt.Errorf("occurred at cannot be zero")
	}
	
	// 验证时间逻辑
	if event.GetOccurredAt().After(time.Now().Add(time.Minute)) {
		return fmt.Errorf("occurred at cannot be in the future")
	}
	
	return nil
}

// ValidateBatch 批量验证事件
func (v *EventValidator) ValidateBatch(events []DomainEvent) error {
	for i, event := range events {
		if err := v.ValidateEvent(event); err != nil {
			return fmt.Errorf("validation failed for event at index %d: %w", i, err)
		}
	}
	return nil
}