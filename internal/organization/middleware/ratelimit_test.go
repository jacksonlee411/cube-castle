package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitMiddleware_AllowsThenBlocks(t *testing.T) {
	cfg := &RateLimitConfig{
		RequestsPerMinute: 1,
		BurstSize:         1,
		CleanupInterval:   time.Minute,
		WhitelistIPs:      []string{}, // do not whitelist localhost
		BlockDuration:     time.Second,
	}
	rl := NewRateLimitMiddleware(cfg, nil)

	handled := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handled++
		w.WriteHeader(http.StatusOK)
	})
	handler := rl.Middleware()(next)

	// First request should pass
	req1 := httptest.NewRequest("GET", "/any", nil)
	req1.Header.Set("X-Forwarded-For", "1.2.3.4")
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("first request status=%d", rr1.Code)
	}
	if rr1.Header().Get("X-RateLimit-Limit") != "1" {
		t.Fatalf("missing/invalid X-RateLimit-Limit header: %v", rr1.Header().Get("X-RateLimit-Limit"))
	}

	// Second request within same minute should be blocked
	req2 := httptest.NewRequest("GET", "/any", nil)
	req2.Header.Set("X-Forwarded-For", "1.2.3.4")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request expected 429, got %d", rr2.Code)
	}
	if handled != 1 {
		t.Fatalf("next handler should be called once, got %d", handled)
	}
}
