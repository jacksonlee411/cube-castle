package repository

import (
	"testing"
	"time"

	"cube-castle/internal/types"
)

func TestShouldMarkOrganizationCurrent_SameDayDifferentTimezones(t *testing.T) {
	effectiveDate := types.NewDate(2025, time.November, 7)
	// 模拟本地时区为 UTC+8，时间位于当天早上
	reference := time.Date(2025, time.November, 7, 8, 0, 0, 0, time.FixedZone("CST", 8*3600))

	if !shouldMarkOrganizationCurrent(effectiveDate, reference) {
		t.Fatalf("expected same-day effective date to be current regardless of timezone")
	}
}

func TestShouldMarkOrganizationCurrent_FutureDate(t *testing.T) {
	effectiveDate := types.NewDate(2025, time.November, 8)
	reference := time.Date(2025, time.November, 7, 23, 0, 0, 0, time.UTC)

	if shouldMarkOrganizationCurrent(effectiveDate, reference) {
		t.Fatalf("expected future effective date to be marked as non-current")
	}
}

func TestShouldMarkOrganizationCurrent_NilDate(t *testing.T) {
	if !shouldMarkOrganizationCurrent(nil, time.Now()) {
		t.Fatalf("nil effective date should default to current")
	}
}
