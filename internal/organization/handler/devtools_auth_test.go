package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cube-castle/internal/auth"
	pkglogger "cube-castle/pkg/logger"
)

func newDevToolsHandlerWithJWT(t *testing.T) *DevToolsHandler {
	t.Helper()
	jwt := auth.NewJWTMiddlewareWithOptions("super-secret", "issuer", "audience", auth.Options{Alg: "HS256"})
	return NewDevToolsHandler(jwt, pkglogger.NewNoopLogger(), true, nil)
}

func TestDevTools_GenerateDevTokenSuccess(t *testing.T) {
	h := newDevToolsHandlerWithJWT(t)
	body := `{"userId":"tester","tenantId":"tenant-1","roles":["ADMIN"],"duration":"90m","scope":"org:write","permissions":["org:create"]}`
	req := httptest.NewRequest(http.MethodPost, "/auth/dev-token", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.GenerateDevToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %#v", payload)
	}
	token, _ := data["token"].(string)
	if token == "" {
		t.Fatalf("expected token to be present in response")
	}
	if _, err := h.jwtMiddleware.ValidateToken("Bearer " + token); err != nil {
		t.Fatalf("generated token should be valid: %v", err)
	}
}

func TestDevTools_GetTokenInfoSuccess(t *testing.T) {
	h := newDevToolsHandlerWithJWT(t)
	token, err := h.jwtMiddleware.GenerateTestTokenWithClaims("dev-user", "tenant-1", []string{"ADMIN"}, "org:read", []string{"org:list"}, time.Hour)
	if err != nil {
		t.Fatalf("GenerateTestTokenWithClaims: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/dev-token/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	h.GetTokenInfo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %#v", payload)
	}
	if data["userId"] != "dev-user" || data["tenantId"] != "tenant-1" {
		t.Fatalf("unexpected token info data: %#v", data)
	}
	if valid, _ := data["valid"].(bool); !valid {
		t.Fatalf("expected token to be marked valid")
	}
}
