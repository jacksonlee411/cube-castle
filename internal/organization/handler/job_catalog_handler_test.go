package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cube-castle/internal/organization/service"
	pkglogger "cube-castle/pkg/logger"
)

func TestJobCatalogHandlePreconditionFailed(t *testing.T) {
	handler := &JobCatalogHandler{
		logger: pkglogger.NewNoopLogger(),
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/job-family-groups/PROF", nil)
	rec := httptest.NewRecorder()

	handler.handleServiceError(rec, req, service.ErrJobCatalogPreconditionFailed)

	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected status 412, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	errorField, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object in response: %v", body)
	}
	if code, _ := errorField["code"].(string); code != "PRECONDITION_FAILED" {
		t.Fatalf("expected error code PRECONDITION_FAILED, got %q", code)
	}
}
