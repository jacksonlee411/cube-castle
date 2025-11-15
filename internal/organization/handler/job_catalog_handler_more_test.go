package handler

import (
	"net/http/httptest"
	"testing"

	validator "cube-castle/internal/organization/validator"
)

func TestJobCatalog_WriteValidationFailure(t *testing.T) {
	h := NewJobCatalogHandler(nil, nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/job-levels", nil)
	result := &validator.ValidationResult{
		Errors: []validator.ValidationError{{
			Code:     "JOB_LEVEL_INVALID",
			Message:  "无效的职级输入",
			Field:    "name",
			Value:    "  ",
			Severity: string(validator.SeverityHigh),
		}},
	}
	h.writeValidationFailure(rr, req, result)
	if rr.Code < 400 {
		t.Fatalf("expected >=400 status, got %d", rr.Code)
	}
}

