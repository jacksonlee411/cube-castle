package eventbus

import "encoding/json"

// GenericJSONEvent 是一个通用的事件载体，便于 outbox dispatcher 将 JSON 负载映射到事件总线。
type GenericJSONEvent struct {
	eventType     string
	aggregateID   string
	aggregateType string
	payload       json.RawMessage
}

// NewGenericJSONEvent 创建一个通用 JSON 事件。
func NewGenericJSONEvent(eventType, aggregateID, aggregateType string, payload json.RawMessage) GenericJSONEvent {
	return GenericJSONEvent{
		eventType:     eventType,
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		payload:       append(json.RawMessage(nil), payload...),
	}
}

// EventType 实现 Event 接口。
func (e GenericJSONEvent) EventType() string {
	return e.eventType
}

// AggregateID 实现 Event 接口。
func (e GenericJSONEvent) AggregateID() string {
	return e.aggregateID
}

// AggregateType 返回聚合类型，便于日志或指标记录。
func (e GenericJSONEvent) AggregateType() string {
	return e.aggregateType
}

// Payload 返回原始 JSON 负载。
func (e GenericJSONEvent) Payload() json.RawMessage {
	return append(json.RawMessage(nil), e.payload...)
}
