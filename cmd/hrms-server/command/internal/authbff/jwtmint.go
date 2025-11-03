package authbff

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTMintConfig struct {
	Secret   string
	Issuer   string
	Audience string
	Alg      string // RS256 only
	// RS256 支持
	PrivateKey    *rsa.PrivateKey
	PrivateKeyPEM []byte
	KeyID         string // 用于JWKS kid
}

// MintAccessToken 生成短期访问令牌（前端仅持此Token）
func MintAccessToken(cfg JWTMintConfig, sess *Session, ttl time.Duration) (string, int64, error) {
	alg := strings.ToUpper(strings.TrimSpace(cfg.Alg))
	if alg == "" {
		alg = "RS256"
	}
	if alg != "RS256" {
		return "", 0, fmt.Errorf("unsupported signing algorithm: %s", alg)
	}

	now := time.Now().UTC()
	exp := now.Add(ttl).Unix()
	claims := jwt.MapClaims{
		"iss":       cfg.Issuer,
		"aud":       cfg.Audience,
		"iat":       now.Unix(),
		"nbf":       now.Unix(),
		"exp":       exp,
		"sub":       sess.UserID,
		"tenant_id": sess.TenantID,
	}
	if len(sess.Roles) > 0 {
		claims["roles"] = sess.Roles
	}
	if len(sess.Scopes) > 0 {
		claims["scope"] = strings.Join(sess.Scopes, " ")
	}

	if cfg.PrivateKey == nil {
		return "", 0, fmt.Errorf("RS256 private key not configured")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if cfg.KeyID != "" {
		token.Header["kid"] = cfg.KeyID
	}
	signed, err := token.SignedString(cfg.PrivateKey)
	if err != nil {
		return "", 0, err
	}
	return signed, exp, nil
}

func ParseRSAPrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid pem")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}
	// Try PKCS8
	pkcs8, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err2 == nil {
		if k, ok := pkcs8.(*rsa.PrivateKey); ok {
			return k, nil
		}
		return nil, fmt.Errorf("not rsa private key")
	}
	return nil, err
}
