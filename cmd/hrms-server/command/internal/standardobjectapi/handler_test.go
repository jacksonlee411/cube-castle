package standardobjectapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cube-castle/internal/middleware"
	"cube-castle/internal/standardobject"
	pkglogger "cube-castle/pkg/logger"
	clockpkg "cube-castle/pkg/temporal/clock"
	"github.com/go-chi/chi/v5"
)

func TestHandleCreateStandardObject(t *testing.T) {
	svc := newInMemoryService()
	handler := NewHandler(svc, clockpkg.NewSystemClock(), pkglogger.NewNoopLogger())

	router := chi.NewRouter()
	router.Use(middleware.RequestIDMiddleware)
	handler.SetupRoutes(router)

	payload := map[string]any{
		"kernel": map[string]any{
			"code":        "OU-1001",
			"displayName": "示例组织",
			"tenantCode":  "11111111-1111-1111-1111-111111111111",
			"status":      "ACTIVE",
			"labels": map[string]any{
				"unitType": "DEPARTMENT",
			},
		},
		"version": map[string]any{
			"effectiveDate": "2025-01-01",
			"payload": map[string]any{
				"name": "示例组织",
			},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/standard-objects/organization_unit", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp successResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response decode failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success response: %+v", resp)
	}
	if resp.Data.Kernel.Code != "OU-1001" {
		t.Fatalf("unexpected kernel code: %+v", resp.Data.Kernel)
	}

	if _, err := svc.Get(context.Background(), standardobject.ObjectKey{
		ObjectType: standardobject.ObjectTypeOrganizationUnit,
		Code:       "OU-1001",
		TenantCode: "11111111-1111-1111-1111-111111111111",
	}); err != nil {
		t.Fatalf("aggregate not stored: %v", err)
	}
}

func TestHandleCreateRejectsInvalidType(t *testing.T) {
	handler := NewHandler(newInMemoryService(), clockpkg.NewSystemClock(), pkglogger.NewNoopLogger())
	router := chi.NewRouter()
	router.Use(middleware.RequestIDMiddleware)
	handler.SetupRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/standard-objects/invalid", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

type inMemoryService struct {
	store map[string]standardobject.ObjectAggregate
}

func newInMemoryService() *inMemoryService {
	return &inMemoryService{
		store: make(map[string]standardobject.ObjectAggregate),
	}
}

func (s *inMemoryService) Upsert(_ context.Context, aggregate standardobject.ObjectAggregate) error {
	key := s.keyFor(aggregate.Kernel.TenantCode, aggregate.Kernel.Code, aggregate.Kernel.ObjectType)
	clone := aggregate
	if clone.Kernel.CreatedAt.IsZero() {
		clone.Kernel.CreatedAt = time.Now()
	}
	if clone.Kernel.UpdatedAt.IsZero() {
		clone.Kernel.UpdatedAt = clone.Kernel.CreatedAt
	}
	s.store[key] = clone
	return nil
}

func (s *inMemoryService) Get(_ context.Context, key standardobject.ObjectKey) (standardobject.ObjectAggregate, error) {
	if agg, ok := s.store[s.keyFor(key.TenantCode, key.Code, key.ObjectType)]; ok {
		return agg, nil
	}
	return standardobject.ObjectAggregate{}, standardobject.ErrNotFound
}

func (s *inMemoryService) keyFor(tenant, code string, objectType standardobject.ObjectType) string {
	return string(objectType) + ":" + tenant + ":" + code
}
