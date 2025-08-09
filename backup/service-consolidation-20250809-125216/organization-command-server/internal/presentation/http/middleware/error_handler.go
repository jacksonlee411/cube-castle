package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/logging"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	Timestamp string      `json:"timestamp"`
	Path      string      `json:"path"`
	TraceID   string      `json:"trace_id,omitempty"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// DomainError represents domain-specific errors
type DomainError struct {
	Code    string
	Message string
	Details string
}

func (e DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Predefined domain errors
var (
	ErrOrganizationNotFound          = DomainError{"ORG_001", "organization not found", ""}
	ErrOrganizationCodeAlreadyExists = DomainError{"ORG_002", "organization code already exists", ""}
	ErrCannotDeleteWithChildren      = DomainError{"ORG_003", "cannot delete organization with children", ""}
	ErrInvalidOrganizationCode       = DomainError{"ORG_004", "invalid organization code format", ""}
	ErrInvalidRequest                = DomainError{"REQ_001", "invalid request", ""}
	ErrValidationFailed              = DomainError{"VAL_001", "validation failed", ""}
	ErrInternalServerError           = DomainError{"SRV_001", "internal server error", ""}
)

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logger logging.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger logging.Logger) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

// Handle provides middleware for handling panics and errors
func (eh *ErrorHandler) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				eh.logger.Error("panic recovered",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method,
					"stack", string(debug.Stack()),
				)
				
				eh.WriteErrorResponse(w, r, ErrInternalServerError, http.StatusInternalServerError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// WriteErrorResponse writes a standardized error response
func (eh *ErrorHandler) WriteErrorResponse(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	var domainErr DomainError
	
	// Determine error type and status code
	switch e := err.(type) {
	case DomainError:
		domainErr = e
	default:
		// Map common errors to domain errors
		statusCode, domainErr = eh.mapErrorToDomainError(err, statusCode)
	}
	
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    domainErr.Code,
			Message: domainErr.Message,
			Details: domainErr.Details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
		TraceID:   eh.getTraceID(r),
	}
	
	// Log error
	eh.logger.Error("http error response",
		"status_code", statusCode,
		"error_code", domainErr.Code,
		"error_message", domainErr.Message,
		"path", r.URL.Path,
		"method", r.Method,
		"trace_id", response.TraceID,
	)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// mapErrorToDomainError maps Go errors to domain errors
func (eh *ErrorHandler) mapErrorToDomainError(err error, defaultStatusCode int) (int, DomainError) {
	errMsg := err.Error()
	
	// Check for common error patterns
	switch {
	case contains(errMsg, "not found"):
		return http.StatusNotFound, ErrOrganizationNotFound
	case contains(errMsg, "already exists"):
		return http.StatusConflict, ErrOrganizationCodeAlreadyExists
	case contains(errMsg, "validation failed"):
		return http.StatusBadRequest, ErrValidationFailed
	case contains(errMsg, "invalid"):
		return http.StatusBadRequest, ErrInvalidRequest
	case contains(errMsg, "cannot delete"):
		return http.StatusUnprocessableEntity, ErrCannotDeleteWithChildren
	default:
		return http.StatusInternalServerError, DomainError{
			Code:    ErrInternalServerError.Code,
			Message: ErrInternalServerError.Message,
			Details: errMsg,
		}
	}
}

// getTraceID extracts trace ID from request (placeholder implementation)
func (eh *ErrorHandler) getTraceID(r *http.Request) string {
	// In a real implementation, this would extract from headers or context
	// For now, we'll use a simple request ID
	return r.Header.Get("X-Request-ID")
}

// contains checks if string contains substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RequestLogger middleware for logging HTTP requests
type RequestLogger struct {
	logger logging.Logger
}

// NewRequestLogger creates a new request logger middleware
func NewRequestLogger(logger logging.Logger) *RequestLogger {
	return &RequestLogger{logger: logger}
}

// Handle logs HTTP requests
func (rl *RequestLogger) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a wrapped response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		
		rl.logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"content_length", r.ContentLength,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}