package organization

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestNewCommandMiddlewares(t *testing.T) {
	m := NewCommandMiddlewares(nil)
	if m.Performance == nil || m.RateLimit == nil {
		t.Fatalf("expected all middlewares non-nil")
	}
}

func TestRequestIDMiddlewareWrapper(t *testing.T) {
	called := false
	h := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if GetRequestID(r.Context()) == "" {
			t.Fatalf("missing request id in context")
		}
		w.WriteHeader(http.StatusOK)
	}))
	req, rr := httptest.NewRequest("GET", "/x", nil), httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if !called || rr.Code != http.StatusOK {
		t.Fatalf("handler was not called or unexpected status: %v %d", called, rr.Code)
	}
}

func TestDefaultAuditHistoryConfig(t *testing.T) {
	cfg := DefaultAuditHistoryConfig()
	if !cfg.StrictValidation || !cfg.AllowFallback || cfg.LegacyMode {
		t.Fatalf("unexpected default audit config: %+v", cfg)
	}
	if cfg.CircuitBreakerThreshold <= 0 {
		t.Fatalf("unexpected threshold: %d", cfg.CircuitBreakerThreshold)
	}
}

func TestNewCommandModule_Minimal(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mod, err := NewCommandModule(CommandModuleDeps{DB: db})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mod == nil || mod.Repositories.Organization == nil || mod.Services.Position == nil {
		t.Fatalf("module not fully initialized")
	}
	// Build handlers with minimal deps
	h := mod.NewHandlers(CommandHandlerDeps{DevMode: true})
	if h.Organization == nil || h.Position == nil || h.JobCatalog == nil || h.Operational == nil || h.DevTools == nil {
		t.Fatalf("handlers not constructed properly")
	}
}
