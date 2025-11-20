package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pkglogger "cube-castle/pkg/logger"
	"github.com/go-chi/chi/v5"
)

func newJobCatalogHandlerForValidation() *JobCatalogHandler {
	return NewJobCatalogHandler(nil, pkglogger.NewNoopLogger())
}

func withCodeParam(req *http.Request, code string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("code", code)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func TestJobCatalogHandlers_EarlyValidationFailures(t *testing.T) {
	handler := newJobCatalogHandlerForValidation()

	t.Run("CreateJobFamilyGroup invalid JSON", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/job-family-groups", strings.NewReader("not-json"))
		handler.CreateJobFamilyGroup(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("UpdateJobFamilyGroup missing code", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/api/v1/job-family-groups/", strings.NewReader(`{}`))
		req = withCodeParam(req, "")
		handler.UpdateJobFamilyGroup(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("CreateJobFamilyGroupVersion missing code", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/job-family-groups//versions", strings.NewReader(`{}`))
		req = withCodeParam(req, "")
		handler.CreateJobFamilyGroupVersion(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("CreateJobFamily invalid JSON", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/job-families", strings.NewReader("~"))
		handler.CreateJobFamily(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("CreateJobRole invalid JSON", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/job-roles", strings.NewReader(""))
		handler.CreateJobRole(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})

	t.Run("CreateJobLevelVersion missing required fields", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/job-levels/LL-1/versions", strings.NewReader(`{"name":"","status":"","effectiveDate":""}`))
		req = withCodeParam(req, "LL-1")
		handler.CreateJobLevelVersion(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})
}
