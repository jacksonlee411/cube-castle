package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"cube-castle-deployment-test/internal/config"
)

// ValidateJWT 统一JWT验证逻辑
// 支持HS256和RS256算法，替换6个文件中的重复验证代码
func ValidateJWT(tokenString string, config *config.JWTConfig) (map[string]interface{}, error) {
	// 解析token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if config.IsHS256() {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.Secret), nil
		}

		if config.IsRS256() {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// 优先使用JWKS
			if config.HasJWKS() {
				return getPublicKeyFromJWKS(config.JWKSUrl, token)
			}

			// 使用公钥文件
			if config.HasPublicKey() {
				return loadPublicKeyFromFile(config.PublicKeyPath)
			}

			return nil, errors.New("no public key configured for RS256")
		}

		return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %v", err)
	}

	// 验证token有效性
	if !token.Valid {
		return nil, errors.New("token is invalid")
	}

	// 获取claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to parse token claims")
	}

	// 验证基本claims
	if err := validateBasicClaims(claims, config); err != nil {
		return nil, err
	}

	// 转换为map[string]interface{}返回
	result := make(map[string]interface{})
	for k, v := range claims {
		result[k] = v
	}

	return result, nil
}

// validateBasicClaims 验证基本JWT claims
func validateBasicClaims(claims jwt.MapClaims, config *config.JWTConfig) error {
	now := time.Now()

	// 验证过期时间 (exp)
	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		if now.After(expTime.Add(config.AllowedClockSkew)) {
			return errors.New("token has expired")
		}
	}

	// 验证生效时间 (nbf)
	if nbf, ok := claims["nbf"].(float64); ok {
		nbfTime := time.Unix(int64(nbf), 0)
		if now.Before(nbfTime.Add(-config.AllowedClockSkew)) {
			return errors.New("token not valid yet")
		}
	}

	// 验证签发时间 (iat)
	if iat, ok := claims["iat"].(float64); ok {
		iatTime := time.Unix(int64(iat), 0)
		if now.Before(iatTime.Add(-config.AllowedClockSkew)) {
			return errors.New("token used before issued")
		}
	}

	// 验证发行者 (iss)
	if config.Issuer != "" {
		if iss, ok := claims["iss"].(string); !ok || iss != config.Issuer {
			return fmt.Errorf("invalid issuer: expected %s, got %s", config.Issuer, iss)
		}
	}

	// 验证受众 (aud)
	if config.Audience != "" {
		if aud, ok := claims["aud"].(string); !ok || aud != config.Audience {
			return fmt.Errorf("invalid audience: expected %s, got %s", config.Audience, aud)
		}
	}

	return nil
}

// loadPublicKeyFromFile 从文件加载RSA公钥
func loadPublicKeyFromFile(filepath string) (*rsa.PublicKey, error) {
	keyBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %v", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %v", err)
	}

	return publicKey, nil
}

// getPublicKeyFromJWKS 从JWKS端点获取公钥
func getPublicKeyFromJWKS(jwksUrl string, token *jwt.Token) (*rsa.PublicKey, error) {
	// 简化实现，实际项目中需要实现完整的JWKS客户端
	resp, err := http.Get(jwksUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status: %d", resp.StatusCode)
	}

	// TODO: 实现完整的JWKS解析逻辑
	// 这里需要根据token中的kid找到对应的公钥
	return nil, errors.New("JWKS support not fully implemented")
}