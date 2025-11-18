package middleware

import (
	"testing"
	"time"

	pkglogger "cube-castle/pkg/logger"
)

func newRateLimitForTest() *RateLimitMiddleware {
	cfg := &RateLimitConfig{
		RequestsPerMinute: 10,
		BurstSize:         2,
		CleanupInterval:   time.Hour, // avoid fast ticker in tests
		BlockDuration:     time.Minute,
	}
	return NewRateLimitMiddleware(cfg, pkglogger.NewNoopLogger())
}

func TestRateLimitStatsLifecycle(t *testing.T) {
	rlm := newRateLimitForTest()

	// initial stats
	initial := rlm.GetStats()
	if initial.TotalRequests != 0 || initial.BlockedRequests != 0 {
		t.Fatalf("initial stats should be zero, got %+v", initial)
	}

	// update and read
	rlm.updateStats(true)
	rlm.updateStats(false)
	stats := rlm.GetStats()
	if stats.TotalRequests != 2 || stats.BlockedRequests != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	// reset stats
	rlm.ResetStats()
	afterReset := rlm.GetStats()
	if afterReset.TotalRequests != 0 || afterReset.BlockedRequests != 0 {
		t.Fatalf("stats should reset to zero, got %+v", afterReset)
	}
}

func TestRateLimitClientManagement(t *testing.T) {
	rlm := newRateLimitForTest()

	rlm.BlockIP("1.1.1.1", time.Minute)
	client := rlm.GetClientInfo("1.1.1.1")
	if client == nil || client.BlockedUntil.IsZero() {
		t.Fatalf("expected blocked client info")
	}

	rlm.UnblockIP("1.1.1.1")
	client = rlm.GetClientInfo("1.1.1.1")
	if client == nil || !client.BlockedUntil.IsZero() {
		t.Fatalf("expected client to be unblocked, got %+v", client)
	}

	// add another client and force cleanup
	rlm.clients["2.2.2.2"] = &ClientInfo{
		IP:           "2.2.2.2",
		LastRequest:  time.Now().Add(-10 * time.Minute),
		RequestCount: 1,
	}
	rlm.cleanupExpiredClients()
	if _, exists := rlm.clients["2.2.2.2"]; exists {
		t.Fatalf("expected expired client to be cleaned up")
	}

	active := rlm.GetActiveClients()
	if len(active) == 0 {
		t.Fatalf("expected at least one active client after block/unblock")
	}
}
