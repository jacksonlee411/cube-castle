package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"cube-castle-deployment-test/internal/config"
)

// GinJWTMiddleware 统一JWT中间件（Gin版本）
// 替换6个文件中重复的JWT验证逻辑
func GinJWTMiddleware() gin.HandlerFunc {
	jwtConfig := config.GetJWTConfig()

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

		token := parts[1]

		// 验证JWT token（根据配置使用不同算法）
		claims, err := ValidateJWT(token, jwtConfig)
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

		// 验证token中的租户ID与header中的一致性
		tokenTenantID := getTokenTenantID(claims)
		if tokenTenantID != tenantIDHeader {
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
		c.Set("user_id", claims["sub"])
		c.Set("tenant_id", tokenTenantID)
		c.Set("roles", claims["roles"])
		c.Set("jwt_claims", claims)

		c.Next()
	})
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

// getTokenTenantID 从JWT claims中提取租户ID
// 兼容 tenantId (camelCase) 和 tenant_id (snake_case)
func getTokenTenantID(claims map[string]interface{}) string {
	if tenantID, ok := claims["tenantId"].(string); ok {
		return tenantID
	}
	if tenantID, ok := claims["tenant_id"].(string); ok {
		return tenantID
	}
	return ""
}
