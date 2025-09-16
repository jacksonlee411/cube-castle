package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"cube-castle-deployment-test/internal/config"
	"github.com/gin-gonic/gin"
)

// GinJWTMiddleware 统一JWT中间件（Gin版本）
// 统一改用 internal/auth/jwt.go 的权威校验器，移除重复实现
func GinJWTMiddleware() gin.HandlerFunc {
	jwtConfig := config.GetJWTConfig()

	// 将配置转换为权威校验器实例
	var opt Options
	opt.Alg = jwtConfig.Algorithm
	opt.ClockSkew = jwtConfig.AllowedClockSkew
	if strings.EqualFold(jwtConfig.Algorithm, "RS256") {
		if jwtConfig.JWKSUrl != "" {
			opt.JWKSURL = jwtConfig.JWKSUrl
		} else if jwtConfig.PublicKeyPath != "" {
			if pemBytes, err := os.ReadFile(jwtConfig.PublicKeyPath); err == nil {
				opt.PublicKeyPEM = pemBytes
			} else {
				panic(fmt.Sprintf("读取JWT公钥失败(%s): %v", jwtConfig.PublicKeyPath, err))
			}
		} else {
			panic("RS256 模式必须配置 JWT_JWKS_URL 或 JWT_PUBLIC_KEY_PATH")
		}
		if jwtConfig.PrivateKeyPath != "" {
			if pemBytes, err := os.ReadFile(jwtConfig.PrivateKeyPath); err == nil {
				opt.PrivateKeyPEM = pemBytes
			} else {
				panic(fmt.Sprintf("读取JWT私钥失败(%s): %v", jwtConfig.PrivateKeyPath, err))
			}
		}
	}
	opt.KeyID = jwtConfig.KeyID
	jwtMW := NewJWTMiddlewareWithOptions(jwtConfig.Secret, jwtConfig.Issuer, jwtConfig.Audience, opt)

	return gin.HandlerFunc(func(c *gin.Context) {
		// 提取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "MISSING_AUTHORIZATION",
					"message": "Authorization header is required",
				},
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
			return
		}

		// 验证Bearer token格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "Authorization header must be Bearer token",
				},
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
			return
		}

		// 使用权威校验器进行校验
		claims, err := jwtMW.ValidateToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": fmt.Sprintf("Token validation failed: %v", err),
				},
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
			return
		}

		// 租户ID一致性检查（契约：缺失→401，不匹配→403）
		tenantIDHeader := c.GetHeader("X-Tenant-ID")
		if tenantIDHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "TENANT_HEADER_REQUIRED",
					"message": "X-Tenant-ID header required",
				},
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
			return
		}

		if claims.TenantID != "" && claims.TenantID != tenantIDHeader {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "TENANT_MISMATCH",
					"message": "X-Tenant-ID does not match tenant in token",
				},
				"timestamp": getCurrentTimestamp(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储在context中供后续使用
		c.Set("user_id", claims.UserID)
		// 使用header作为最终租户上下文（与GraphQL中间件保持一致）
		c.Set("tenant_id", tenantIDHeader)
		c.Set("roles", claims.Roles)
		c.Set("jwt_claims", map[string]interface{}{
			"sub":         claims.UserID,
			"tenantId":    tenantIDHeader,
			"roles":       claims.Roles,
			"scope":       claims.Scope,
			"permissions": claims.Permissions,
			"exp":         claims.ExpiresAt,
		})

		c.Next()
	})
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
