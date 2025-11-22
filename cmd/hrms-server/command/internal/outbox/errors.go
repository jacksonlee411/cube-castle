package outbox

import "errors"

var (
	// ErrDispatcherAlreadyRunning 表示 Dispatcher 已在运行。
	ErrDispatcherAlreadyRunning = errors.New("outbox dispatcher already running")
	// ErrDispatcherNotRunning 表示 Dispatcher 未启动。
	ErrDispatcherNotRunning = errors.New("outbox dispatcher not running")
)
