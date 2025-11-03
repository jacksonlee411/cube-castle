package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestNewJWTMiddlewareWithOptionsDefaults(t *testing.T) {
	mw := NewJWTMiddlewareWithOptions("secret", "issuer", "aud", Options{})
	if mw.alg != "RS256" {
		t.Fatalf("expected default alg RS256, got %s", mw.alg)
	}
}

func TestJWTValidateTokenHS256(t *testing.T) {
	opt := Options{Alg: "HS256"}
	mw := NewJWTMiddlewareWithOptions("super-secret", "cube", "castle", opt)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":       "cube",
		"aud":       "castle",
		"sub":       "user-1",
		"tenant_id": "tenant-1",
		"roles":     []string{"ADMIN", "MANAGER"},
		"scope":     "org:read org:write",
		"permissions": []string{
			"org:read:hierarchy",
		},
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"nbf": time.Now().Add(-1 * time.Minute).Unix(),
	})

	signed, err := token.SignedString([]byte("super-secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	claims, err := mw.ValidateToken("Bearer " + signed)
	if err != nil {
		t.Fatalf("validate token returned error: %v", err)
	}

	if claims.UserID != "user-1" {
		t.Fatalf("unexpected user id: %s", claims.UserID)
	}
	if claims.TenantID != "tenant-1" {
		t.Fatalf("unexpected tenant id: %s", claims.TenantID)
	}
	if len(claims.Roles) != 2 {
		t.Fatalf("expected roles to have 2 entries, got %d", len(claims.Roles))
	}
	if claims.Scope != "org:read org:write" {
		t.Fatalf("unexpected scope: %s", claims.Scope)
	}
	if len(claims.Permissions) != 1 || claims.Permissions[0] != "org:read:hierarchy" {
		t.Fatalf("unexpected permissions: %#v", claims.Permissions)
	}
}

func TestJWTValidateTokenExpired(t *testing.T) {
	opt := Options{Alg: "HS256"}
	mw := NewJWTMiddlewareWithOptions("secret", "issuer", "aud", opt)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "issuer",
		"aud": "aud",
		"sub": "user",
		"exp": time.Now().Add(-1 * time.Minute).Unix(),
	})

	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if _, err := mw.ValidateToken(signed); err == nil {
		t.Fatalf("expected error for expired token")
	}
}

func TestJWTValidateTokenRS256WithJWKS(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	n := base64.RawURLEncoding.EncodeToString(privKey.PublicKey.N.Bytes())
	eBytes := bigIntToBase64(privKey.PublicKey.E)

	jwksPayload := jwkSet{
		Keys: []jwkKey{{
			Kty: "RSA",
			Kid: "kid-1",
			N:   n,
			E:   eBytes,
		}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwksPayload)
	}))
	defer server.Close()

	mw := NewJWTMiddlewareWithOptions("", "issuer", "aud", Options{Alg: "RS256", JWKSURL: server.URL})

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":         "issuer",
		"aud":         "aud",
		"sub":         "user-42",
		"tenantId":    "tenant-42",
		"exp":         time.Now().Add(5 * time.Minute).Unix(),
		"permissions": []string{"org:read"},
	})
	token.Header["kid"] = "kid-1"

	signed, err := token.SignedString(privKey)
	if err != nil {
		t.Fatalf("failed to sign RSA token: %v", err)
	}

	claims, err := mw.ValidateToken("Bearer " + signed)
	if err != nil {
		t.Fatalf("failed to validate RSA token: %v", err)
	}
	if claims.TenantID != "tenant-42" {
		t.Fatalf("expected tenant id tenant-42, got %s", claims.TenantID)
	}
}

func TestParseRSAPublicKeyFromPEM(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	pkBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	var buf bytes.Buffer
	if err := pem.Encode(&buf, &pem.Block{Type: "PUBLIC KEY", Bytes: pkBytes}); err != nil {
		t.Fatalf("failed to encode pem: %v", err)
	}

	parsed, err := ParseRSAPublicKeyFromPEM(buf.Bytes())
	if err != nil {
		t.Fatalf("expected to parse public key: %v", err)
	}
	if parsed.N.Cmp(priv.PublicKey.N) != 0 {
		t.Fatalf("parsed key mismatch")
	}

	if _, err := ParseRSAPublicKeyFromPEM([]byte("bad pem")); err == nil {
		t.Fatalf("expected error for invalid pem")
	}
}

func TestRSAFromModExp(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	n := base64.RawURLEncoding.EncodeToString(priv.PublicKey.N.Bytes())
	e := bigIntToBase64(priv.PublicKey.E)

	pk, err := RSAFromModExp(n, e)
	if err != nil {
		t.Fatalf("expected to build rsa key: %v", err)
	}
	if pk.N.Cmp(priv.PublicKey.N) != 0 {
		t.Fatalf("unexpected modulus")
	}

	if _, err := RSAFromModExp("!!", e); err == nil {
		t.Fatalf("expected error for bad modulus")
	}
	if _, err := RSAFromModExp(n, "!!"); err == nil {
		t.Fatalf("expected error for bad exponent")
	}
}

func TestJWKSManagerRefresh(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	n := base64.RawURLEncoding.EncodeToString(priv.PublicKey.N.Bytes())
	e := bigIntToBase64(priv.PublicKey.E)

	payload := jwkSet{Keys: []jwkKey{{Kty: "RSA", Kid: "key", N: n, E: e}}}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	manager := NewJWKSManager(server.URL, time.Minute)
	if err := manager.Refresh(); err != nil {
		t.Fatalf("expected refresh success: %v", err)
	}
	if manager.GetKey("key") == nil {
		t.Fatalf("expected key to be cached")
	}

	// bad status response
	badSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer badSrv.Close()

	manager = NewJWKSManager(badSrv.URL, time.Minute)
	if err := manager.Refresh(); err == nil {
		t.Fatalf("expected refresh to fail on bad status")
	}
}

func TestSetUserContextHelpers(t *testing.T) {
	claims := &Claims{
		UserID:      "user-x",
		TenantID:    "tenant-y",
		Roles:       []string{"ADMIN"},
		Scope:       "org:read org:write",
		Permissions: []string{"org:create", "org:write"},
	}
	ctx := SetUserContext(context.Background(), claims)

	if got := GetUserID(ctx); got != "user-x" {
		t.Fatalf("expected user id user-x, got %s", got)
	}
	if got := GetTenantID(ctx); got != "tenant-y" {
		t.Fatalf("expected tenant id tenant-y, got %s", got)
	}
	if len(GetUserRoles(ctx)) != 1 {
		t.Fatalf("expected roles length 1")
	}
	scopes := GetUserScopes(ctx)
	if len(scopes) != 3 {
		t.Fatalf("expected deduped scopes, got %v", scopes)
	}
}

func TestPBACPermissionChecker(t *testing.T) {
	checker := NewPBACPermissionChecker(nil, log.New(io.Discard, "", 0))

	// missing auth context
	if err := checker.CheckPermission(context.Background(), "organizations"); err == nil {
		t.Fatal("expected error for missing context")
	}

	claims := &Claims{
		UserID:      "user",
		TenantID:    "tenant",
		Roles:       []string{"EMPLOYEE"},
		Scope:       "org:read",
		Permissions: []string{"org:read"},
	}
	ctx := SetUserContext(context.Background(), claims)
	if err := checker.CheckPermission(ctx, "organizations"); err != nil {
		t.Fatalf("expected permission granted by scope: %v", err)
	}

	claims = &Claims{
		UserID:   "admin",
		TenantID: "tenant",
		Roles:    []string{"GUEST"},
	}
	ctx = SetUserContext(context.Background(), claims)
	if err := checker.CheckPermission(ctx, "organizations"); err != nil {
		t.Fatalf("expected permission granted for admin: %v", err)
	}

	claims = &Claims{
		UserID:   "user",
		TenantID: "tenant",
		Roles:    []string{"MANAGER"},
	}
	ctx = SetUserContext(context.Background(), claims)
	if err := checker.CheckPermission(ctx, "organizationHierarchy"); err != nil {
		t.Fatalf("expected permission granted via role: %v", err)
	}

	claims = &Claims{
		UserID:   "user",
		TenantID: "tenant",
		Roles:    []string{"GUEST"},
	}
	ctx = SetUserContext(context.Background(), claims)
	if err := checker.CheckPermission(ctx, "organizationStats"); err == nil {
		t.Fatalf("expected denial")
	}

	if err := checker.CheckPermission(ctx, "unknown"); err == nil {
		t.Fatalf("expected error for unknown query")
	}
}

func TestGraphQLPermissionMiddlewareModes(t *testing.T) {
	opt := Options{Alg: "HS256"}
	jwtMW := NewJWTMiddlewareWithOptions("secret", "cube", "aud", opt)
	logger := log.New(io.Discard, "", 0)
	checker := NewPBACPermissionChecker(nil, logger)

	middleware := NewGraphQLPermissionMiddleware(jwtMW, checker, logger, true)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if GetUserID(ctx) != "user" {
			t.Fatalf("expected user context to be set")
		}
		w.WriteHeader(http.StatusOK)
	}))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      "cube",
		"aud":      "aud",
		"sub":      "user",
		"tenantId": "tenant",
		"exp":      time.Now().Add(5 * time.Minute).Unix(),
	})
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	req.Header.Set("X-Tenant-ID", "tenant")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	// missing authorization header should fail
	badReq := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	badRR := httptest.NewRecorder()
	handler.ServeHTTP(badRR, badReq)
	if badRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing auth, got %d", badRR.Code)
	}
}

func TestGraphQLPermissionMiddlewareProduction(t *testing.T) {
	opt := Options{Alg: "HS256"}
	jwtMW := NewJWTMiddlewareWithOptions("secret", "cube", "aud", opt)
	logger := log.New(io.Discard, "", 0)
	checker := NewPBACPermissionChecker(nil, logger)

	middleware := NewGraphQLPermissionMiddleware(jwtMW, checker, logger, false)

	handler := middleware.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":       "cube",
		"aud":       "aud",
		"sub":       "user",
		"tenant_id": "tenant",
		"exp":       time.Now().Add(5 * time.Minute).Unix(),
	})
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	req.Header.Set("X-Tenant-ID", "tenant")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	mismatchReq := httptest.NewRequest(http.MethodPost, "/graphql", nil)
	mismatchReq.Header.Set("Authorization", "Bearer "+signed)
	mismatchReq.Header.Set("X-Tenant-ID", "other")

	mismatchRR := httptest.NewRecorder()
	handler.ServeHTTP(mismatchRR, mismatchReq)
	if mismatchRR.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for tenant mismatch, got %d", mismatchRR.Code)
	}
}

func TestCheckQueryPermissionDelegation(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	checker := NewPBACPermissionChecker(nil, logger)
	mw := NewGraphQLPermissionMiddleware(nil, checker, logger, false)

	goodCtx := SetUserContext(context.Background(), &Claims{
		UserID:   "user",
		TenantID: "tenant",
		Scope:    "org:read",
	})
	if err := mw.CheckQueryPermission(goodCtx, "organizations"); err != nil {
		t.Fatalf("expected production check to succeed: %v", err)
	}

	mw.devMode = true
	devCtx := SetUserContext(context.Background(), &Claims{
		UserID:   "user",
		TenantID: "tenant",
		Roles:    []string{"ADMIN"},
	})
	if err := mw.CheckQueryPermission(devCtx, "organizationHierarchy"); err != nil {
		t.Fatalf("expected dev check to succeed: %v", err)
	}

	badCtx := SetUserContext(context.Background(), &Claims{
		UserID:   "user",
		TenantID: "tenant",
	})
	if err := mw.CheckQueryPermission(badCtx, "organizationStats"); err == nil {
		t.Fatalf("expected error when permission denied")
	}
}

func bigIntToBase64(v int) string {
	buf := new(bytes.Buffer)
	for i := 24; i >= 0; i-- {
		b := byte(v >> (i * 8))
		if buf.Len() == 0 && b == 0 {
			continue
		}
		buf.WriteByte(b)
	}
	if buf.Len() == 0 {
		buf.WriteByte(0)
	}
	return base64.RawURLEncoding.EncodeToString(buf.Bytes())
}

func TestGinJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	dir := t.TempDir()
	pubPath := filepath.Join(dir, "jwt.pub")
	privPath := filepath.Join(dir, "jwt.key")
	if err := os.WriteFile(pubPath, pubPEM, 0o600); err != nil {
		t.Fatalf("failed to write public key: %v", err)
	}
	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		t.Fatalf("failed to write private key: %v", err)
	}

	t.Setenv("JWT_SECRET", "unused")
	t.Setenv("JWT_ISSUER", "cube")
	t.Setenv("JWT_AUDIENCE", "castle")
	t.Setenv("JWT_ALG", "RS256")
	t.Setenv("JWT_PUBLIC_KEY_PATH", pubPath)
	t.Setenv("JWT_PRIVATE_KEY_PATH", privPath)
	t.Setenv("JWT_ALLOWED_CLOCK_SKEW", "0s")

	middleware := GinJWTMiddleware()

	router := gin.New()
	router.Use(middleware)
	router.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":       "cube",
		"aud":       "castle",
		"sub":       "user",
		"tenant_id": "tenant",
		"exp":       time.Now().Add(5 * time.Minute).Unix(),
	})
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	req.Header.Set("X-Tenant-ID", "tenant")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	// Missing Authorization header -> 401
	missing := httptest.NewRequest(http.MethodGet, "/protected", nil)
	missing.Header.Set("X-Tenant-ID", "tenant")
	missingResp := httptest.NewRecorder()
	router.ServeHTTP(missingResp, missing)
	if missingResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when authorization header missing, got %d", missingResp.Code)
	}

	// Invalid format -> 401
	invalid := httptest.NewRequest(http.MethodGet, "/protected", nil)
	invalid.Header.Set("Authorization", "Token "+signed)
	invalid.Header.Set("X-Tenant-ID", "tenant")
	invalidResp := httptest.NewRecorder()
	router.ServeHTTP(invalidResp, invalid)
	if invalidResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid authorization format, got %d", invalidResp.Code)
	}

	// Tenant mismatch -> 403
	mismatch := httptest.NewRequest(http.MethodGet, "/protected", nil)
	mismatch.Header.Set("Authorization", "Bearer "+signed)
	mismatch.Header.Set("X-Tenant-ID", "other")
	mismatchResp := httptest.NewRecorder()
	router.ServeHTTP(mismatchResp, mismatch)
	if mismatchResp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for tenant mismatch, got %d", mismatchResp.Code)
	}
}
