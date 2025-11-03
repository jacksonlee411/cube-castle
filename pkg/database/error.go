package database

import "errors"

var (
	// ErrEmptyDSN 在 DSN 为空时返回。
	ErrEmptyDSN = errors.New("database DSN cannot be empty")
	// ErrNilOutboxEvent 在 Outbox 事件为空时返回。
	ErrNilOutboxEvent = errors.New("outbox event cannot be nil")
	// ErrEmptyEventID 在事件没有 EventID 时返回。
	ErrEmptyEventID = errors.New("outbox event ID cannot be empty")
	// ErrEventNotFound 在更新/查询事件不存在时返回。
	ErrEventNotFound = errors.New("outbox event not found")
	// ErrDatabaseNotInitialized 在使用未初始化的 Database 时返回。
	ErrDatabaseNotInitialized = errors.New("database is not initialized")
)
