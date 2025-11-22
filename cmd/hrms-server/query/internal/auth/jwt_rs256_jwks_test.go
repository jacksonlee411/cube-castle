package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// helper: build JWKS JSON for a single RSA public key
func buildJWKS(pub *rsa.PublicKey, kid string) []byte {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	// Exponent to bytes (big-endian) and trim leading zeros
	eBytes := []byte{byte(pub.E >> 16), byte(pub.E >> 8), byte(pub.E)}
	i := 0
	for i < len(eBytes) && eBytes[i] == 0 {
		i++
	}
	e := base64.RawURLEncoding.EncodeToString(eBytes[i:])
	jwk := map[string]any{"kty": "RSA", "kid": kid, "alg": "RS256", "use": "sig", "n": n, "e": e}
	set := map[string]any{"keys": []any{jwk}}
	b, _ := json.Marshal(set)
	return b
}

func generateRSA() (*rsa.PrivateKey, *rsa.PublicKey) {
	pk, _ := rsa.GenerateKey(rand.Reader, 2048)
	return pk, &pk.PublicKey
}

func signRS256(t *testing.T, pk *rsa.PrivateKey, kid, iss, aud, tenant string, ttl time.Duration) string {
	t.Helper()
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":       iss,
		"aud":       aud,
		"iat":       now.Unix(),
		"nbf":       now.Unix(),
		"exp":       now.Add(ttl).Unix(),
		"sub":       "test-user",
		"tenant_id": tenant,
		// include a nonce-like value
		"nonce": base64.RawURLEncoding.EncodeToString(sha256.New().Sum([]byte("n"))),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	s, err := token.SignedString(pk)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	return s
}

func TestRS256JWTValidationWithJWKS_Success(t *testing.T) {
	pk, pub := generateRSA()
	jwks := buildJWKS(pub, "test-kid")
	// JWKS test server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwks)
	}))
	defer srv.Close()

	mw := NewJWTMiddlewareWithOptions("", "cube-castle", "cube-castle-users", Options{Alg: "RS256", JWKSURL: srv.URL})
	token := signRS256(t, pk, "test-kid", "cube-castle", "cube-castle-users", "tenant-xyz", 5*time.Minute)
	_, err := mw.ValidateToken("Bearer " + token)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestRS256JWTValidationWithJWKS_InvalidAudience(t *testing.T) {
	pk, pub := generateRSA()
	jwks := buildJWKS(pub, "test-kid")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwks)
	}))
	defer srv.Close()

	mw := NewJWTMiddlewareWithOptions("", "cube-castle", "correct-aud", Options{Alg: "RS256", JWKSURL: srv.URL})
	token := signRS256(t, pk, "test-kid", "cube-castle", "wrong-aud", "tenant-xyz", 5*time.Minute)
	if _, err := mw.ValidateToken("Bearer " + token); err == nil {
		t.Fatalf("expected audience mismatch error, got nil")
	}
}
