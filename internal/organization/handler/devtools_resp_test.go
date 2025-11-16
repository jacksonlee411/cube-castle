package handler

import (
	"net/http/httptest"
	"testing"
)

func TestDevTools_WriteResponses(t *testing.T) {
	h := NewDevToolsHandler(nil, nil, true, nil)
	// Success
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/dev/status", nil)
	h.writeSuccessResponse(rr, map[string]string{"ok": "true"}, "msg", req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	// Error
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/dev/status", nil)
	h.writeErrorResponse(rr, "ERR", "oops", 418, req)
	if rr.Code != 418 {
		t.Fatalf("expected 418, got %d", rr.Code)
	}
}

func TestDevTools_PerfAndEndpoints(t *testing.T) {
	h := NewDevToolsHandler(nil, nil, true, nil)
	// PerformanceMetrics (no DB)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/dev/performance-metrics", nil)
	h.PerformanceMetrics(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	// ListTestEndpoints
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/dev/test-endpoints", nil)
	h.ListTestEndpoints(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
