package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const RequestIDKey = "requestID"

// RequestIDMiddleware 添加请求ID追踪中间件
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取或生成请求ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置响应头
		w.Header().Set("X-Request-ID", requestID)

		// 添加到上下文
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID 从上下文中获取请求ID
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}