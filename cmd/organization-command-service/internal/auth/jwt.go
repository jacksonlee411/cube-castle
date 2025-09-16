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
	hmacSecret []byte
	privateKey *rsa.PrivateKey
	issuer     string
	audience   string
	alg        string // HS256|RS256
	jwks       *JWKSManager
	publicKey  interface{}
	keyID      string
	clockSkew  time.Duration
}

func NewJWTMiddleware(secretKey, issuer, audience string) *JWTMiddleware {
	return &JWTMiddleware{
		hmacSecret: []byte(secretKey),
		issuer:     issuer,
		audience:   audience,
		alg:        "HS256",
		clockSkew:  0,
	}
}

// Options 认证可选参数
type Options struct {
	Alg           string        // HS256|RS256
	JWKSURL       string        // RS256时的JWKS地址
	PublicKeyPEM  []byte        // RS256时的本地公钥PEM
	PrivateKeyPEM []byte        // RS256时的本地私钥PEM（用于签名）
	KeyID         string        // JWKS 对应的 key id
	ClockSkew     time.Duration // 允许的时钟偏差
}

func NewJWTMiddlewareWithOptions(secretKey, issuer, audience string, opt Options) *JWTMiddleware {
	mw := &JWTMiddleware{
		hmacSecret: []byte(secretKey),
		issuer:     issuer,
		audience:   audience,
		alg:        strings.ToUpper(strings.TrimSpace(opt.Alg)),
		keyID:      strings.TrimSpace(opt.KeyID),
		clockSkew:  opt.ClockSkew,
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
	if mw.alg == "" {
		mw.alg = "HS256"
	}
	return mw
}

// Claims JWT声明结构
type Claims struct {
	UserID    string   `json:"sub"`
	TenantID  string   `json:"tenant_id"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
}

// ValidateToken 验证JWT令牌
func (j *JWTMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	// 移除Bearer前缀
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		alg := token.Header["alg"]
		if j.alg == "HS256" {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method: %v", alg)
			}
			return j.hmacSecret, nil
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
		// 验证issuer和audience
		if iss, ok := claims["iss"].(string); !ok || iss != j.issuer {
			return nil, fmt.Errorf("invalid issuer")
		}
		if aud, ok := claims["aud"].(string); !ok || aud != j.audience {
			return nil, fmt.Errorf("invalid audience")
		}

		// 提取claims（兼容 tenantId / tenant_id）
		userClaims := &Claims{
			UserID:   getClaimString(claims, "sub"),
			TenantID: getTenantIDClaim(claims),
		}

		// 提取角色数组
		if rolesClaim, ok := claims["roles"].([]interface{}); ok {
			for _, role := range rolesClaim {
				if roleStr, ok := role.(string); ok {
					userClaims.Roles = append(userClaims.Roles, roleStr)
				}
			}
		}

		// 过期/生效时间校验（含clock skew）
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

// getClaimString 安全地获取字符串类型的claim
func getClaimString(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}

// getTenantIDClaim 兼容读取 camelCase 与 snake_case 的租户字段
func getTenantIDClaim(claims jwt.MapClaims) string {
	if v, ok := claims["tenantId"].(string); ok && v != "" {
		return v
	}
	if v, ok := claims["tenant_id"].(string); ok && v != "" {
		return v
	}
	return ""
}

// SetUserContext 将用户信息设置到上下文中
func SetUserContext(ctx context.Context, claims *Claims) context.Context {
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
	ctx = context.WithValue(ctx, "user_roles", claims.Roles)
	return ctx
}

// GetUserID 从上下文获取用户ID
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetTenantID 从上下文获取租户ID
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return ""
}

// GetUserRoles 从上下文获取用户角色
func GetUserRoles(ctx context.Context) []string {
	if roles, ok := ctx.Value("user_roles").([]string); ok {
		return roles
	}
	return []string{}
}

// GenerateTestToken 生成测试用的JWT令牌 (仅开发环境使用)
func (j *JWTMiddleware) GenerateTestToken(userID, tenantID string, roles []string, duration time.Duration) (string, error) {
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
	default:
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(j.hmacSecret)
		if err != nil {
			return "", fmt.Errorf("failed to generate token: %w", err)
		}
		return tokenString, nil
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
	UserID   string   `json:"userId"`
	TenantID string   `json:"tenantId"`
	Roles    []string `json:"roles"`
	Duration string   `json:"duration"` // 例如: "1h", "24h"
}

// TestTokenResponse 测试令牌响应结构
type TestTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	UserID    string    `json:"userId"`
	TenantID  string    `json:"tenantId"`
	Roles     []string  `json:"roles"`
}
