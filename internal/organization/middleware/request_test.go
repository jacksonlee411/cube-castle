package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestIDMiddleware_GeneratesAndPropagatesIDs(t *testing.T) {
	h := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// From context
		if GetRequestID(r.Context()) == "" {
			t.Fatalf("expected request id in context")
		}
		if GetCorrelationID(r.Context()) == "" {
			t.Fatalf("expected correlation id in context")
		}
		if GetCorrelationSource(r.Context()) == "" {
			t.Fatalf("expected correlation source in context")
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/x", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Header().Get("X-Request-ID") == "" {
		t.Fatalf("missing X-Request-ID header")
	}
	if rr.Header().Get("X-Correlation-ID") == "" {
		t.Fatalf("missing X-Correlation-ID header")
	}
}

func TestRequestIDMiddleware_UsesHeadersWhenProvided(t *testing.T) {
	h := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetCorrelationSource(r.Context()) != "header" {
			t.Fatalf("expected correlation source header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("X-Request-ID", "req-123")
	req.Header.Set("X-Correlation-ID", "corr-456")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if got := rr.Header().Get("X-Request-ID"); got != "req-123" {
		t.Fatalf("X-Request-ID propagated mismatch: %q", got)
	}
	if got := rr.Header().Get("X-Correlation-ID"); got != "corr-456" {
		t.Fatalf("X-Correlation-ID propagated mismatch: %q", got)
	}
}

func TestPerformanceMiddleware_SetsHeadersAndCallsNext(t *testing.T) {
	pm := NewPerformanceMiddleware(nil)
	handler := pm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("ok"))
	}))
	req := httptest.NewRequest("GET", "/perf", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Header().Get("X-Service") != "organization-command-service" {
		t.Fatalf("expected X-Service header")
	}
	if _, ok := rr.Header()["X-Response-Time"]; !ok {
		t.Fatalf("expected X-Response-Time header")
	}
	if rr.Code != http.StatusTeapot || !strings.Contains(rr.Body.String(), "ok") {
		t.Fatalf("unexpected response: %d %q", rr.Code, rr.Body.String())
	}
}

