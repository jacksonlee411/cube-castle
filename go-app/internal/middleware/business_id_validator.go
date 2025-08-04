package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gaogu/cube-castle/go-app/internal/common"
)

// BusinessIDValidationError 业务ID验证错误
type BusinessIDValidationError struct {
	Field          string `json:"field"`
	Message        string `json:"message"`
	Code           string `json:"code"`
	ExpectedFormat string `json:"expected_format,omitempty"`
	ProvidedValue  string `json:"provided_value,omitempty"`
}

// ErrorResponse 标准化错误响应
type ErrorResponse struct {
	Error             string                      `json:"error"`
	Message           string                      `json:"message"`
	Details           map[string]string          `json:"details,omitempty"`
	ValidationErrors  []BusinessIDValidationError `json:"validation_errors,omitempty"`
	Timestamp         time.Time                  `json:"timestamp"`
	RequestID         string                     `json:"request_id"`
}

// BusinessIDValidator 业务ID验证中间件
func BusinessIDValidator(entityType common.EntityType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idParamName := fmt.Sprintf("%s_id", entityType)
			id := chi.URLParam(r, idParamName)

			// 检查是否启用UUID查询模式
			uuidLookup := r.URL.Query().Get("uuid_lookup") == "true"
			
			if uuidLookup {
				// UUID查询模式 - 验证UUID格式
				if !common.IsUUID(id) {
					writeUUIDValidationError(w, idParamName, id)
					return
				}
			} else {
				// 业务ID查询模式 - 验证业务ID格式
				if err := common.ValidateBusinessID(entityType, id); err != nil {
					writeBusinessIDValidationError(w, entityType, idParamName, id, err)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeBusinessIDValidationError 写入业务ID验证错误响应
func writeBusinessIDValidationError(w http.ResponseWriter, entityType common.EntityType, field, providedValue string, validationErr error) {
	var expectedFormat string
	switch entityType {
	case common.EntityTypeEmployee:
		expectedFormat = "1-99999999 (string format)"
	case common.EntityTypeOrganization:
		expectedFormat = "100000-999999 (string format)"
	case common.EntityTypePosition:
		expectedFormat = "1000000-9999999 (string format)"
	}

	errorResp := ErrorResponse{
		Error:   "VALIDATION_ERROR",
		Message: "Invalid business ID format",
		Details: map[string]string{
			"field":           field,
			"expected_format": expectedFormat,
			"provided_value":  providedValue,
		},
		ValidationErrors: []BusinessIDValidationError{
			{
				Field:          field,
				Message:        fmt.Sprintf("Must be a string representation of number in range %s", expectedFormat),
				Code:           "INVALID_BUSINESS_ID_FORMAT",
				ExpectedFormat: expectedFormat,
				ProvidedValue:  providedValue,
			},
		},
		Timestamp: time.Now(),
		RequestID: generateRequestID(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errorResp)
}

// writeUUIDValidationError 写入UUID验证错误响应
func writeUUIDValidationError(w http.ResponseWriter, field, providedValue string) {
	errorResp := ErrorResponse{
		Error:   "VALIDATION_ERROR",
		Message: "Invalid UUID format",
		Details: map[string]string{
			"field":           field,
			"expected_format": "UUID v4 format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)",
			"provided_value":  providedValue,
		},
		ValidationErrors: []BusinessIDValidationError{
			{
				Field:          field,
				Message:        "Must be a valid UUID v4 format",
				Code:           "INVALID_UUID_FORMAT",
				ExpectedFormat: "UUID v4 format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)",
				ProvidedValue:  providedValue,
			},
		},
		Timestamp: time.Now(),
		RequestID: generateRequestID(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(errorResp)
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%s", uuid.New().String()[:8])
}

// BusinessIDRouteValidator 路由级别的业务ID验证器
type BusinessIDRouteValidator struct {
	entityTypes map[string]common.EntityType
}

// NewBusinessIDRouteValidator 创建路由级别验证器
func NewBusinessIDRouteValidator() *BusinessIDRouteValidator {
	return &BusinessIDRouteValidator{
		entityTypes: map[string]common.EntityType{
			"employee":     common.EntityTypeEmployee,
			"organization": common.EntityTypeOrganization,
			"position":     common.EntityTypePosition,
		},
	}
}

// ValidateRoute 验证特定路由的业务ID
func (v *BusinessIDRouteValidator) ValidateRoute(routeName string) func(http.Handler) http.Handler {
	entityType, exists := v.entityTypes[routeName]
	if !exists {
		// 如果没有匹配的实体类型，返回一个空的中间件
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	
	return BusinessIDValidator(entityType)
}

// RequestContext 请求上下文增强
type RequestContext struct {
	IsUUIDLookup bool
	EntityType   common.EntityType
	Identifier   string
	RequestID    string
}

// ExtractRequestContext 从请求中提取上下文信息
func ExtractRequestContext(r *http.Request, entityType common.EntityType) *RequestContext {
	idParamName := fmt.Sprintf("%s_id", entityType)
	identifier := chi.URLParam(r, idParamName)
	isUUIDLookup := r.URL.Query().Get("uuid_lookup") == "true"

	return &RequestContext{
		IsUUIDLookup: isUUIDLookup,
		EntityType:   entityType,
		Identifier:   identifier,
		RequestID:    generateRequestID(),
	}
}

// BusinessIDMiddlewareConfig 中间件配置
type BusinessIDMiddlewareConfig struct {
	EnableValidation    bool
	EnableLogging      bool
	LogInvalidRequests bool
	CustomErrorHandler func(w http.ResponseWriter, r *http.Request, err error)
}

// DefaultBusinessIDMiddlewareConfig 默认中间件配置
func DefaultBusinessIDMiddlewareConfig() BusinessIDMiddlewareConfig {
	return BusinessIDMiddlewareConfig{
		EnableValidation:    true,
		EnableLogging:      true,
		LogInvalidRequests: true,
		CustomErrorHandler: nil,
	}
}

// ConfigurableBusinessIDValidator 可配置的业务ID验证器
func ConfigurableBusinessIDValidator(entityType common.EntityType, config BusinessIDMiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.EnableValidation {
				next.ServeHTTP(w, r)
				return
			}

			ctx := ExtractRequestContext(r, entityType)
			idParamName := fmt.Sprintf("%s_id", entityType)

			var validationErr error

			if ctx.IsUUIDLookup {
				if !common.IsUUID(ctx.Identifier) {
					validationErr = fmt.Errorf("invalid UUID format: %s", ctx.Identifier)
				}
			} else {
				validationErr = common.ValidateBusinessID(entityType, ctx.Identifier)
			}

			if validationErr != nil {
				if config.LogInvalidRequests && config.EnableLogging {
					// 这里可以添加日志记录
					fmt.Printf("Invalid %s ID: %s, Error: %v\n", entityType, ctx.Identifier, validationErr)
				}

				if config.CustomErrorHandler != nil {
					config.CustomErrorHandler(w, r, validationErr)
					return
				}

				if ctx.IsUUIDLookup {
					writeUUIDValidationError(w, idParamName, ctx.Identifier)
				} else {
					writeBusinessIDValidationError(w, entityType, idParamName, ctx.Identifier, validationErr)
				}
				return
			}

			if config.EnableLogging {
				// 记录成功的请求（可选）
				fmt.Printf("Valid %s ID request: %s (UUID lookup: %v)\n", 
					entityType, ctx.Identifier, ctx.IsUUIDLookup)
			}

			next.ServeHTTP(w, r)
		})
	}
}