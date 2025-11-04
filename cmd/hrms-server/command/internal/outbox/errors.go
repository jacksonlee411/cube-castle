package outbox

import "errors"

var (
	ErrDispatcherAlreadyRunning = errors.New("outbox dispatcher already running")
	ErrDispatcherNotRunning     = errors.New("outbox dispatcher not running")
)
