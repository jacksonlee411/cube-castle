package handler

import (
	"net/http/httptest"
	"testing"

	mw "cube-castle/internal/organization/middleware"
	pkglogger "cube-castle/pkg/logger"
)

func TestOperational_GetRateLimitStats(t *testing.T) {
	rl := mw.NewRateLimitMiddleware(mw.DefaultRateLimitConfig, pkglogger.NewNoopLogger())
	h := NewOperationalHandler(nil, nil, rl, pkglogger.NewNoopLogger())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/operational/rate-limit/stats", nil)
	h.GetRateLimitStats(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Fatalf("expected JSON body")
	}
}

