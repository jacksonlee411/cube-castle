package constraints

import (
	"fmt"
	"time"
)

// RangeWindow describes a bi-temporal window.
type RangeWindow struct {
	From time.Time
	To   *time.Time
}

// ValidateAppendOnly ensures the incoming window does not overlap existing append-only windows.
func ValidateAppendOnly(existing []RangeWindow, incoming RangeWindow) error {
	for _, win := range existing {
		if overlaps(win, incoming) {
			return fmt.Errorf("range overlap detected (existing=%s incoming=%s)", win.From, incoming.From)
		}
	}
	return nil
}

func overlaps(a, b RangeWindow) bool {
	aEnd := endOrInfinity(a.To)
	bEnd := endOrInfinity(b.To)
	return a.From.Before(bEnd) && b.From.Before(aEnd)
}

func endOrInfinity(t *time.Time) time.Time {
	if t == nil {
		return time.UnixMilli(1<<62 - 1)
	}
	return *t
}
