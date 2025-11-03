package eventbus

import "context"

// Event 定义所有事件必须实现的接口。
// EventType 返回事件类型标识符，例如 "employee.created"。
// AggregateID 返回关联的聚合根标识，用于追踪事件来源。
type Event interface {
	EventType() string
	AggregateID() string
}

// EventHandler 表示事件处理函数。
// 当处理函数返回非 nil 错误时，会记录该错误但不会阻止其他处理器执行。
type EventHandler func(ctx context.Context, event Event) error

// EventBus 定义事件总线的标准接口。
type EventBus interface {
	// Publish 发布事件到所有订阅者。
	// 当没有订阅者时，应返回 nil。
	Publish(ctx context.Context, event Event) error

	// Subscribe 为指定事件类型注册处理器。
	// 同一事件类型允许存在多个处理器。
	Subscribe(eventType string, handler EventHandler) error
}
