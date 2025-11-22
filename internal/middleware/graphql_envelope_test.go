package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// helper to execute middleware and decode JSON body
func executeMiddleware(t *testing.T, payload []byte) map[string]interface{} {
	t.Helper()
	mw := NewGraphQLEnvelopeMiddleware()
	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	})
	mw.Middleware()(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/graphql", nil))

	var out map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &out); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return out
}

func TestGraphQLEnvelopeTreatsNilErrorsAsSuccess(t *testing.T) {
	payload := []byte(`{"data":{"foo":"bar"},"errors":null}`)
	out := executeMiddleware(t, payload)

	if success, ok := out["success"].(bool); !ok || !success {
		t.Fatalf("expected success=true, got %v", out["success"])
	}
	if _, hasError := out["error"]; hasError {
		t.Fatalf("did not expect error field, got %v", out["error"])
	}
}

func TestGraphQLEnvelopeTreatsEmptyErrorsAsSuccess(t *testing.T) {
	payload := []byte(`{"data":{"foo":"bar"},"errors":[]}`)
	out := executeMiddleware(t, payload)

	if success, ok := out["success"].(bool); !ok || !success {
		t.Fatalf("expected success=true, got %v", out["success"])
	}
	if _, hasError := out["error"]; hasError {
		t.Fatalf("did not expect error field, got %v", out["error"])
	}
}
