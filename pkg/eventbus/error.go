package eventbus

import "errors"

var (
	ErrNilEvent       = errors.New("event cannot be nil")
	ErrNilHandler     = errors.New("handler cannot be nil")
	ErrEmptyEventType = errors.New("event type cannot be empty")
)
