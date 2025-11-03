package auth

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJobCatalogPermissionMapping(t *testing.T) {
	checker := NewPBACPermissionChecker(nil, log.New(io.Discard, "", 0))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/job-family-groups", nil)
	ctx := SetUserContext(context.Background(), &Claims{
		UserID:   "employee",
		TenantID: "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9",
		Roles:    []string{"EMPLOYEE"},
	})
	err := checker.CheckRESTAPI(request.WithContext(ctx))
	if err == nil {
		t.Fatalf("expected permission error for EMPLOYEE role")
	}

	allowedCtx := SetUserContext(context.Background(), &Claims{
		UserID:   "hr",
		TenantID: "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9",
		Roles:    []string{"HR_STAFF"},
	})
	if err := checker.CheckRESTAPI(request.WithContext(allowedCtx)); err != nil {
		t.Fatalf("expected HR_STAFF to access job catalog write endpoint: %v", err)
	}
}
