package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	pkglogger "cube-castle/pkg/logger"
)

func TestPerformanceMiddleware_Serve(t *testing.T) {
	pm := NewPerformanceMiddleware(pkglogger.NewNoopLogger())
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})
	handler := pm.Middleware()(next)
	req := httptest.NewRequest("GET", "/graphql", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
