package eventbus

import "errors"

var (
	// ErrNilEvent 表示事件为空。
	ErrNilEvent = errors.New("event cannot be nil")
	// ErrNilHandler 表示处理器为空。
	ErrNilHandler = errors.New("handler cannot be nil")
	// ErrEmptyEventType 表示事件类型缺失。
	ErrEmptyEventType = errors.New("event type cannot be empty")
)
