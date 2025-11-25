package standardobject

import "errors"

var (
	// ErrAdapterNotConfigured clarifies that the phase-A skeleton does not yet provide persistence.
	ErrAdapterNotConfigured = errors.New("standardobject: adapter not configured")
	// ErrNotFound is returned when no aggregate can be located for the given key/range.
	ErrNotFound = errors.New("standardobject: aggregate not found")
)
