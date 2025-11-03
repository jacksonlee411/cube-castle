package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTMiddleware struct {
	secretKey  []byte
	privateKey *rsa.PrivateKey
	issuer     string
	audience   string
	alg        string
	jwks       *JWKSManager
	publicKey  interface{}
	keyID      string
	clockSkew  time.Duration
}

type contextKey string

const (
	userIDKey     contextKey = "user_id"
	tenantIDKey   contextKey = "tenant_id"
	userRolesKey  contextKey = "user_roles"
	userScopesKey contextKey = "user_scopes"
)

func NewJWTMiddleware(secretKey, issuer, audience string) *JWTMiddleware {
	return &JWTMiddleware{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		audience:  audience,
		alg:       "RS256",
	}
}

type Options struct {
	Alg           string
	JWKSURL       string
	PublicKeyPEM  []byte
	PrivateKeyPEM []byte
	KeyID         string
	ClockSkew     time.Duration
}

func NewJWTMiddlewareWithOptions(secretKey, issuer, audience string, opt Options) *JWTMiddleware {
	mw := &JWTMiddleware{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		audience:  audience,
		alg:       strings.ToUpper(strings.TrimSpace(opt.Alg)),
		keyID:     strings.TrimSpace(opt.KeyID),
		clockSkew: opt.ClockSkew,
	}
	if mw.alg == "" {
		mw.alg = "RS256"
	}
	if mw.alg == "RS256" {
		if opt.JWKSURL != "" {
			mw.jwks = NewJWKSManager(opt.JWKSURL, 5*time.Minute)
		} else if len(opt.PublicKeyPEM) > 0 {
			if pk, err := ParseRSAPublicKeyFromPEM(opt.PublicKeyPEM); err == nil {
				mw.publicKey = pk
			}
		}
		if len(opt.PrivateKeyPEM) > 0 {
			if pk, err := parseRSAPrivateKey(opt.PrivateKeyPEM); err == nil {
				mw.privateKey = pk
			}
		} else if strings.Contains(secretKey, "BEGIN") {
			if pk, err := parseRSAPrivateKey([]byte(secretKey)); err == nil {
				mw.privateKey = pk
			}
		}
	}
	mw.alg = strings.ToUpper(mw.alg)
	return mw
}

// Claims JWT声明结构
type Claims struct {
	UserID    string   `json:"sub"`
	TenantID  string   `json:"tenant_id"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
	// PBAC scopes
	Scope       string   `json:"scope"`
	Permissions []string `json:"permissions"`
}

// ValidateToken 验证JWT令牌
func (j *JWTMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		alg := token.Header["alg"]
		if j.alg == "HS256" {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method: %v", alg)
			}
			return j.secretKey, nil
		}
		if j.alg == "RS256" {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("invalid signing method: %v", alg)
			}
			if j.jwks != nil {
				if kid, _ := token.Header["kid"].(string); kid != "" {
					if key := j.jwks.GetKey(kid); key != nil {
						return key, nil
					}
					if err := j.jwks.Refresh(); err == nil {
						if key := j.jwks.GetKey(kid); key != nil {
							return key, nil
						}
					}
					return nil, fmt.Errorf("unknown kid: %s", kid)
				}
			}
			if j.publicKey != nil {
				return j.publicKey, nil
			}
			return nil, fmt.Errorf("no public key available for RS256")
		}
		return nil, fmt.Errorf("unsupported alg: %s", j.alg)
	}

	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if iss, ok := claims["iss"].(string); !ok || iss != j.issuer {
			return nil, fmt.Errorf("invalid issuer")
		}
		if aud, ok := claims["aud"].(string); !ok || aud != j.audience {
			return nil, fmt.Errorf("invalid audience")
		}

		userClaims := &Claims{
			UserID:   getClaimString(claims, "sub"),
			TenantID: getTenantIDClaim(claims),
		}

		if rolesClaim, ok := claims["roles"].([]interface{}); ok {
			for _, role := range rolesClaim {
				if roleStr, ok := role.(string); ok {
					userClaims.Roles = append(userClaims.Roles, roleStr)
				}
			}
		}
		if scopeStr, ok := claims["scope"].(string); ok {
			userClaims.Scope = scopeStr
		}
		if perms, ok := claims["permissions"].([]interface{}); ok {
			for _, p := range perms {
				if ps, ok := p.(string); ok {
					userClaims.Permissions = append(userClaims.Permissions, ps)
				}
			}
		}

		now := time.Now()
		if exp, ok := claims["exp"].(float64); ok {
			userClaims.ExpiresAt = int64(exp)
			if now.After(time.Unix(userClaims.ExpiresAt, 0).Add(j.clockSkew)) {
				return nil, fmt.Errorf("token expired")
			}
		}
		if nbf, ok := claims["nbf"].(float64); ok {
			if now.Add(j.clockSkew).Before(time.Unix(int64(nbf), 0)) {
				return nil, fmt.Errorf("token not valid yet")
			}
		}

		return userClaims, nil
	}
	return nil, fmt.Errorf("invalid token claims")
}

func getClaimString(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}

func getTenantIDClaim(claims jwt.MapClaims) string {
	if v, ok := claims["tenantId"].(string); ok && v != "" {
		return v
	}
	if v, ok := claims["tenant_id"].(string); ok && v != "" {
		return v
	}
	return ""
}

func SetUserContext(ctx context.Context, claims *Claims) context.Context {
	ctx = context.WithValue(ctx, userIDKey, claims.UserID)
	ctx = context.WithValue(ctx, tenantIDKey, claims.TenantID)
	ctx = context.WithValue(ctx, userRolesKey, claims.Roles)
	scopeSet := map[string]struct{}{}
	for _, s := range strings.Fields(claims.Scope) {
		scopeSet[s] = struct{}{}
	}
	for _, s := range claims.Permissions {
		scopeSet[s] = struct{}{}
	}
	var scopes []string
	for s := range scopeSet {
		scopes = append(scopes, s)
	}
	ctx = context.WithValue(ctx, userScopesKey, scopes)
	return ctx
}

func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

func GetTenantID(ctx context.Context) string {
	if v, ok := ctx.Value(tenantIDKey).(string); ok {
		return v
	}
	return ""
}

func GetUserRoles(ctx context.Context) []string {
	if v, ok := ctx.Value(userRolesKey).([]string); ok {
		return v
	}
	return []string{}
}

func GetUserScopes(ctx context.Context) []string {
	if v, ok := ctx.Value(userScopesKey).([]string); ok {
		return v
	}
	return []string{}
}

// GenerateTestToken 生成测试用的JWT令牌 (仅开发环境使用)
func (j *JWTMiddleware) GenerateTestToken(userID, tenantID string, roles []string, duration time.Duration) (string, error) {
	return j.GenerateTestTokenWithClaims(userID, tenantID, roles, "", nil, duration)
}

// GenerateTestTokenWithClaims 支持额外 scope 与 permissions 的测试令牌
func (j *JWTMiddleware) GenerateTestTokenWithClaims(userID, tenantID string, roles []string, scope string, permissions []string, duration time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(duration)

	claims := jwt.MapClaims{
		"sub":       userID,
		"tenant_id": tenantID,
		"roles":     roles,
		"iss":       j.issuer,
		"aud":       j.audience,
		"iat":       now.Unix(),
		"nbf":       now.Unix(),
		"exp":       expiresAt.Unix(),
	}
	if scope != "" {
		claims["scope"] = scope
	}
	if len(permissions) > 0 {
		claims["permissions"] = permissions
	}

	switch j.alg {
	case "RS256":
		if j.privateKey == nil {
			return "", fmt.Errorf("rs256 private key not configured")
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		if j.keyID != "" {
			token.Header["kid"] = j.keyID
		}
		tokenString, err := token.SignedString(j.privateKey)
		if err != nil {
			return "", fmt.Errorf("failed to generate token: %w", err)
		}
		return tokenString, nil
	case "HS256":
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(j.secretKey)
		if err != nil {
			return "", fmt.Errorf("failed to generate token: %w", err)
		}
		return tokenString, nil
	default:
		return "", fmt.Errorf("unsupported signing algorithm: %s", j.alg)
	}
}

func parseRSAPrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid rsa private key pem")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if pk, ok := parsed.(*rsa.PrivateKey); ok {
			return pk, nil
		}
		return nil, fmt.Errorf("pem does not contain rsa private key")
	}
	return nil, err
}

// TestTokenRequest 开发环境测试令牌请求结构
type TestTokenRequest struct {
	UserID      string   `json:"userId"`
	TenantID    string   `json:"tenantId"`
	Roles       []string `json:"roles"`
	Duration    string   `json:"duration"`
	Scope       string   `json:"scope,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// TestTokenResponse 测试令牌响应结构
type TestTokenResponse struct {
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expiresAt"`
	UserID      string    `json:"userId"`
	TenantID    string    `json:"tenantId"`
	Roles       []string  `json:"roles"`
	Scope       string    `json:"scope,omitempty"`
	Permissions []string  `json:"permissions,omitempty"`
}
