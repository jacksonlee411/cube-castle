package clock

import "time"

// Clock exposes the minimum contract required to obtain transaction timestamps.
type Clock interface {
	Now() time.Time
}

// SystemClock implements Clock using the machine UTC time.
type SystemClock struct{}

// Now returns the current UTC time.
func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

// NewSystemClock returns a Clock backed by SystemClock.
func NewSystemClock() Clock {
	return SystemClock{}
}
