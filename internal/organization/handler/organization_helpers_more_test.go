package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cube-castle/internal/organization/repository"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

func newOrgHandlerForHelpers() *OrganizationHandler {
	return &OrganizationHandler{
		logger: pkglogger.NewNoopLogger(),
	}
}

func TestOrganizationHandler_GetTenantID(t *testing.T) {
	h := newOrgHandlerForHelpers()
	t.Run("valid header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Tenant-ID", uuid.New().String())
		id := h.getTenantID(req)
		if id == types.DefaultTenantID {
			t.Fatalf("expected parsed tenant id, got default")
		}
	})
	t.Run("invalid header falls back", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Tenant-ID", "not-a-uuid")
		id := h.getTenantID(req)
		if id != types.DefaultTenantID {
			t.Fatalf("expected default tenant id")
		}
	})
}

func TestOrganizationHandler_GetActorIDPriority(t *testing.T) {
	h := newOrgHandlerForHelpers()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Mock-User", "mock-user")
	if got := h.getActorID(req); got != "mock-user" {
		t.Fatalf("expected mock-user, got %s", got)
	}

	ctxReq := req.Clone(req.Context())
	ctxReq.Header.Del("X-Mock-User")
	ctxReq = ctxReq.WithContext(context.WithValue(ctxReq.Context(), "user_id", "ctx-user"))
	if got := h.getActorID(ctxReq); got != "ctx-user" {
		t.Fatalf("expected ctx-user, got %s", got)
	}
}

func TestOrganizationHandler_GetIPAddress(t *testing.T) {
	h := newOrgHandlerForHelpers()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")
	ip := h.getIPAddress(req)
	if ip != "10.0.0.1" {
		t.Fatalf("expected first forwarded ip, got %s", ip)
	}
}

func TestOrganizationHandler_HandleRepositoryError(t *testing.T) {
	h := newOrgHandlerForHelpers()

	t.Run("has children conflict", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		h.handleRepositoryError(rec, req, "DELETE", repository.ErrOrganizationHasChildren)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "HAS_CHILD_UNITS") {
			t.Fatalf("expected HAS_CHILD_UNITS code, got %s", rec.Body.String())
		}
	})

	t.Run("duplicate code", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		err := fmt.Errorf("duplicate key value violates unique constraint organization_units_code_tenant_id_key")
		h.handleRepositoryError(rec, req, "CREATE", err)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", rec.Code)
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		errObj, _ := payload["error"].(map[string]interface{})
		if code, _ := errObj["code"].(string); code != "DUPLICATE_CODE" {
			t.Fatalf("expected DUPLICATE_CODE, got %s", code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		err := fmt.Errorf("organization not found")
		h.handleRepositoryError(rec, req, "GET", err)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "ORGANIZATION_NOT_FOUND") {
			t.Fatalf("expected not found code, body=%s", rec.Body.String())
		}
	})
}
