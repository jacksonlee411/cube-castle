// Package middleware 提供API中间件功能
// 包含ADR-008弃用端点处理逻辑
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	
	"cube-castle/internal/audit"
	"cube-castle/internal/metrics"
)

// DeprecatedEndpointGuard ADR-008弃用端点守卫中间件
// 自动处理弃用端点访问，返回410 Gone响应并记录审计
func DeprecatedEndpointGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否为弃用的/reactivate端点
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/organization-units/") &&
			strings.HasSuffix(c.Request.URL.Path, "/reactivate") &&
			c.Request.Method == "POST" {

			// 记录弃用端点访问指标
			clientID := c.GetHeader("X-Client-ID")
			userAgent := c.Request.UserAgent()
			metrics.RecordDeprecatedEndpointUsage(c.Request.URL.Path, clientID, userAgent)

			// 记录审计事件
			auditEvent := audit.Event{
				Type:        "DEPRECATED_ENDPOINT_USED",
				Path:        c.Request.URL.Path,
				TenantID:    c.GetHeader("X-Tenant-ID"),
				ClientID:    clientID,
				UserAgent:   userAgent,
				IP:          c.ClientIP(),
				Timestamp:   time.Now().UTC(),
				Method:      c.Request.Method,
				Successor:   "/api/v1/organization-units/{code}/activate",
			}

			// 异步记录审计日志，避免阻塞响应
			go func() {
				if err := audit.LogEvent(auditEvent); err != nil {
					// 记录审计失败指标
					metrics.RecordAuditWrite(false)
				} else {
					metrics.RecordAuditWrite(true)
				}
			}()

			// 设置标准弃用响应头
			c.Header("Deprecation", "true")
			c.Header("Link", "</api/v1/organization-units/{code}/activate>; rel=\"successor-version\"")
			c.Header("Sunset", "2026-01-01T00:00:00Z")

			// 返回410 Gone响应
			c.JSON(http.StatusGone, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ENDPOINT_DEPRECATED",
					"message": "Use /activate instead of /reactivate",
					"details": gin.H{
						"deprecated_endpoint": c.Request.URL.Path,
						"successor_endpoint":  "/api/v1/organization-units/{code}/activate",
						"migration_guide":     "/docs/api/migration-guide.md",
					},
				},
				"timestamp": time.Now().UTC().Format(time.RFC3339),
				"requestId": c.GetHeader("X-Request-ID"),
			})

			// 阻止继续处理
			c.Abort()
			return
		}

		// 继续处理非弃用端点
		c.Next()
	}
}

// CORS中间件 (如果需要)
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Tenant-ID, X-Client-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestID中间件 - 为每个请求生成唯一ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// 生成新的请求ID
			requestID = generateRequestID()
			c.Header("X-Request-ID", requestID)
		}
		
		// 将请求ID存储到上下文中供后续使用
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID 生成唯一的请求ID
func generateRequestID() string {
	// 这里可以使用UUID或其他方法生成唯一ID
	// 简化版本使用时间戳
	return "req-" + string(time.Now().UnixNano())
}