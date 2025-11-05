package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type ctxKey string

const (
	RequestIDKey         ctxKey = "requestID"
	correlationIDKey     ctxKey = "correlationID"
	correlationSourceKey ctxKey = "correlationSource"
)

// RequestIDMiddleware 添加请求/关联 ID 追踪中间件
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = uuid.New().String()
		}

		correlationID := strings.TrimSpace(r.Header.Get("X-Correlation-ID"))
		correlationSource := "request-id"
		if correlationID == "" {
			correlationID = requestID
		} else {
			correlationSource = "header"
		}

		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Correlation-ID", correlationID)

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, correlationIDKey, correlationID)
		ctx = context.WithValue(ctx, correlationSourceKey, correlationSource)
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

// GetCorrelationID 从上下文中获取关联 ID
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(correlationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

// GetCorrelationSource 返回关联 ID 的来源（header/request-id）
func GetCorrelationSource(ctx context.Context) string {
	if source, ok := ctx.Value(correlationSourceKey).(string); ok {
		return source
	}
	return ""
}
