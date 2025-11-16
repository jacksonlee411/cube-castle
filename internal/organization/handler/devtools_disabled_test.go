package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDevTools_Disabled_Status(t *testing.T) {
	h := NewDevToolsHandler(nil, nil, false, nil) // devMode=false
	req := httptest.NewRequest(http.MethodGet, "/dev/status", nil)
	rec := httptest.NewRecorder()

	h.DevStatus(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "DEV_MODE_DISABLED") {
		t.Fatalf("expected DEV_MODE_DISABLED in body")
	}
}

func TestDevTools_Disabled_ListEndpoints(t *testing.T) {
	h := NewDevToolsHandler(nil, nil, false, nil)
	req := httptest.NewRequest(http.MethodGet, "/dev/test-endpoints", nil)
	rec := httptest.NewRecorder()

	h.ListTestEndpoints(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "DEV_MODE_DISABLED") {
		t.Fatalf("expected DEV_MODE_DISABLED in body")
	}
}

