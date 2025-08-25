package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTMiddleware struct {
	secretKey []byte
	issuer    string
	audience  string
}

func NewJWTMiddleware(secretKey, issuer, audience string) *JWTMiddleware {
	return &JWTMiddleware{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		audience:  audience,
	}
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

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

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

		// 提取claims
		userClaims := &Claims{
			UserID:   getClaimString(claims, "sub"),
			TenantID: getClaimString(claims, "tenant_id"),
		}

		// 提取角色数组
		if rolesClaim, ok := claims["roles"].([]interface{}); ok {
			for _, role := range rolesClaim {
				if roleStr, ok := role.(string); ok {
					userClaims.Roles = append(userClaims.Roles, roleStr)
				}
			}
		}

		// 提取过期时间
		if exp, ok := claims["exp"].(float64); ok {
			userClaims.ExpiresAt = int64(exp)

			// 检查是否过期
			if time.Now().Unix() > userClaims.ExpiresAt {
				return nil, fmt.Errorf("token expired")
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
