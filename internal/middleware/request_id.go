package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey         contextKey = "requestId"
	correlationIDKey     contextKey = "correlationId"
	correlationSourceKey contextKey = "correlationSource"
)

// RequestIDMiddleware 添加请求ID到上下文的中间件
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

// GetRequestID 从上下文获取请求ID
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetCorrelationID 从上下文获取关联ID
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(correlationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

// GetCorrelationSource 返回关联ID来源
func GetCorrelationSource(ctx context.Context) string {
	if source, ok := ctx.Value(correlationSourceKey).(string); ok {
		return source
	}
	return ""
}
