package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResponseBuilder_SimpleFlow(t *testing.T) {
	meta := &Meta{Version: "v1", Server: "test"}
	start := time.Now().Add(-10 * time.Millisecond)
	rb := NewResponseBuilder().
		Success(true).
		Data(map[string]interface{}{"ok": true}).
		Message("msg").
		RequestID("rid-1").
		Meta(meta).
		WithExecutionTime(start).
		WithPagination(&PaginationMeta{Total: 100, Page: 1, Limit: 10})
	resp := rb.Build()
	if !resp.Success || resp.Message != "msg" || resp.RequestID != "rid-1" {
		t.Fatalf("response fields mismatch")
	}
	if resp.Meta == nil || resp.Meta.Version != "v1" || resp.Meta.Server != "test" || resp.Meta.ExecutionTime == "" {
		t.Fatalf("meta mismatch")
	}
}

func TestWriteHelpers_StatusCodes(t *testing.T) {
	// Bad Request
	rr := httptest.NewRecorder()
	if err := WriteBadRequest(rr, "BAD", "bad req", "rid", map[string]string{"f": "e"}); err != nil {
		t.Fatalf("WriteBadRequest error: %v", err)
	}
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var resp APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil || resp.Success {
		t.Fatalf("unexpected response decode or success")
	}

	// Conflict
	rr = httptest.NewRecorder()
	if err := WriteConflict(rr, "CONFLICT", "conflict", "rid", nil); err != nil {
		t.Fatalf("WriteConflict error: %v", err)
	}
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}

	// HealthCheck healthy
	rr = httptest.NewRecorder()
	if err := WriteHealthCheck(rr, "svc", true, map[string]any{"k": "v"}, "rid"); err != nil {
		t.Fatalf("WriteHealthCheck error: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// List with pagination
	rr = httptest.NewRecorder()
	items := []string{"a", "b"}
	if err := WriteList(rr, items, &PaginationMeta{Total: 2, Page: 1, Limit: 10}, "ok", "rid"); err != nil {
		t.Fatalf("WriteList error: %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// Unauthorized/Forbidden/InternalError/ValidationError
	rr = httptest.NewRecorder()
	if err := WriteUnauthorized(rr, "rid"); err != nil || rr.Code != http.StatusUnauthorized {
		t.Fatalf("WriteUnauthorized mismatch")
	}
	rr = httptest.NewRecorder()
	if err := WriteForbidden(rr, "rid"); err != nil || rr.Code != http.StatusForbidden {
		t.Fatalf("WriteForbidden mismatch")
	}
	rr = httptest.NewRecorder()
	if err := WriteInternalError(rr, "rid", map[string]string{"k": "v"}); err != nil || rr.Code != http.StatusInternalServerError {
		t.Fatalf("WriteInternalError mismatch")
	}
	rr = httptest.NewRecorder()
	valErrs := ConvertValidationErrors(map[string]string{"field": "invalid"})
	if err := WriteValidationError(rr, valErrs, "rid"); err != nil || rr.Code != http.StatusBadRequest {
		t.Fatalf("WriteValidationError mismatch")
	}
}
