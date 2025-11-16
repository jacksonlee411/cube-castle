package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResponseBuilder_SuccessAndError(t *testing.T) {
	rr := httptest.NewRecorder()
	err := NewResponseBuilder().
		Data(map[string]any{"ok": true}).
		Message("done").
		RequestID("req-1").
		WriteJSON(rr, http.StatusOK)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content-type: %s", ct)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	var resp APIResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !resp.Success || resp.RequestID != "req-1" || resp.Message != "done" {
		t.Fatalf("unexpected payload: %+v", resp)
	}
	// Error path
	rr = httptest.NewRecorder()
	_ = NewResponseBuilder().Error("E", "bad", map[string]any{"x": 1}).RequestID("req-2").WriteJSON(rr, http.StatusBadRequest)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unexpected error status: %d", rr.Code)
	}
}

func TestWriteHelpers(t *testing.T) {
	rr := httptest.NewRecorder()
	_ = WriteSuccess(rr, map[string]string{"a": "b"}, "ok", "r1")
	if rr.Code != http.StatusOK {
		t.Fatalf("WriteSuccess status=%d", rr.Code)
	}
	rr = httptest.NewRecorder()
	_ = WriteCreated(rr, map[string]string{"a": "b"}, "ok", "r2")
	if rr.Code != http.StatusCreated {
		t.Fatalf("WriteCreated status=%d", rr.Code)
	}
	rr = httptest.NewRecorder()
	_ = WriteUnauthorized(rr, "r3")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("WriteUnauthorized status=%d", rr.Code)
	}
	rr = httptest.NewRecorder()
	_ = WriteForbidden(rr, "r4")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("WriteForbidden status=%d", rr.Code)
	}
	rr = httptest.NewRecorder()
	_ = WriteNotFound(rr, "missing", "r5")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("WriteNotFound status=%d", rr.Code)
	}
	rr = httptest.NewRecorder()
	_ = WriteInternalError(rr, "r6", map[string]string{"ctx": "v"})
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("WriteInternalError status=%d", rr.Code)
	}
}

func TestValidationHelpers(t *testing.T) {
	errs := ConvertValidationErrors(map[string]string{
		"name": "required",
	})
	if len(errs) != 1 || errs[0].Field != "name" || errs[0].Code != "FIELD_INVALID" {
		t.Fatalf("unexpected validation errors: %+v", errs)
	}
	rr := httptest.NewRecorder()
	_ = WriteValidationError(rr, errs, "req-1")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("WriteValidationError status=%d", rr.Code)
	}
}

func TestWithExecutionTimeAndPagination(t *testing.T) {
	start := time.Now().Add(-10 * time.Millisecond)
	p := &PaginationMeta{Total: 100, Page: 2, Limit: 10, TotalPages: 10, HasNext: true}
	rr := httptest.NewRecorder()
	err := NewResponseBuilder().
		Data(map[string]any{"items": []int{1, 2}}).
		WithExecutionTime(start).
		WithPagination(p).
		WriteJSON(rr, http.StatusOK)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var resp APIResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Meta == nil || resp.Meta.ExecutionTime == "" {
		t.Fatalf("expected execution time meta")
	}
	if resp.Meta.Headers["X-Total-Count"] != "100" {
		t.Fatalf("expected X-Total-Count=100, got %+v", resp.Meta.Headers)
	}
}

