package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// APIResponse 统一API响应结构
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"requestId,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
}

// APIError API错误结构
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Stack   string      `json:"stack,omitempty"`
}

// Meta 元数据信息
type Meta struct {
	Version       string            `json:"version,omitempty"`
	ExecutionTime string            `json:"executionTime,omitempty"`
	Server        string            `json:"server,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	Path          string            `json:"path,omitempty"`
	Method        string            `json:"method,omitempty"`
}

// ValidationError 验证错误详情
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
}

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalPages  int   `json:"totalPages"`
	HasNext     bool  `json:"hasNext"`
	HasPrevious bool  `json:"hasPrevious"`
}

// ResponseBuilder API响应构建器
type ResponseBuilder struct {
	response *APIResponse
}

// NewResponseBuilder 创建响应构建器
func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		response: &APIResponse{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// Success 设置成功状态
func (rb *ResponseBuilder) Success(success bool) *ResponseBuilder {
	rb.response.Success = success
	return rb
}

// Data 设置响应数据
func (rb *ResponseBuilder) Data(data interface{}) *ResponseBuilder {
	rb.response.Data = data
	rb.response.Success = true
	return rb
}

// Error 设置错误信息
func (rb *ResponseBuilder) Error(code, message string, details interface{}) *ResponseBuilder {
	rb.response.Success = false
	rb.response.Error = &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
	return rb
}

// ValidationErrors 设置验证错误
func (rb *ResponseBuilder) ValidationErrors(errors []ValidationError) *ResponseBuilder {
	rb.response.Success = false
	rb.response.Error = &APIError{
		Code:    "VALIDATION_ERROR",
		Message: "输入验证失败",
		Details: map[string]interface{}{
			"validationErrors": errors,
			"errorCount":       len(errors),
		},
	}
	return rb
}

// Message 设置消息
func (rb *ResponseBuilder) Message(message string) *ResponseBuilder {
	rb.response.Message = message
	return rb
}

// RequestID 设置请求ID
func (rb *ResponseBuilder) RequestID(requestID string) *ResponseBuilder {
	rb.response.RequestID = requestID
	return rb
}

// Meta 设置元数据
func (rb *ResponseBuilder) Meta(meta *Meta) *ResponseBuilder {
	rb.response.Meta = meta
	return rb
}

// WithExecutionTime 添加执行时间
func (rb *ResponseBuilder) WithExecutionTime(startTime time.Time) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &Meta{}
	}
	rb.response.Meta.ExecutionTime = time.Since(startTime).String()
	return rb
}

// WithPagination 添加分页信息
func (rb *ResponseBuilder) WithPagination(pagination *PaginationMeta) *ResponseBuilder {
	if rb.response.Meta == nil {
		rb.response.Meta = &Meta{}
	}
	if rb.response.Meta.Headers == nil {
		rb.response.Meta.Headers = make(map[string]string)
	}
	rb.response.Meta.Headers["X-Total-Count"] = fmt.Sprintf("%d", pagination.Total)
	rb.response.Meta.Headers["X-Page"] = fmt.Sprintf("%d", pagination.Page)
	rb.response.Meta.Headers["X-Limit"] = fmt.Sprintf("%d", pagination.Limit)
	
	// 将分页信息也添加到data中
	if rb.response.Data != nil {
		if dataMap, ok := rb.response.Data.(map[string]interface{}); ok {
			dataMap["pagination"] = pagination
		}
	}
	
	return rb
}

// Build 构建最终响应
func (rb *ResponseBuilder) Build() *APIResponse {
	return rb.response
}

// WriteJSON 写入JSON响应
func (rb *ResponseBuilder) WriteJSON(w http.ResponseWriter, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	
	// 设置CORS头部
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Tenant-ID")
	
	// 添加请求ID到响应头
	if rb.response.RequestID != "" {
		w.Header().Set("X-Request-ID", rb.response.RequestID)
	}
	
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(rb.response)
}

// 快捷方法

// WriteSuccess 快速写入成功响应
func WriteSuccess(w http.ResponseWriter, data interface{}, message, requestID string) error {
	return NewResponseBuilder().
		Data(data).
		Message(message).
		RequestID(requestID).
		WriteJSON(w, http.StatusOK)
}

// WriteCreated 快速写入创建成功响应
func WriteCreated(w http.ResponseWriter, data interface{}, message, requestID string) error {
	return NewResponseBuilder().
		Data(data).
		Message(message).
		RequestID(requestID).
		WriteJSON(w, http.StatusCreated)
}

// WriteError 快速写入错误响应
func WriteError(w http.ResponseWriter, statusCode int, code, message, requestID string, details interface{}) error {
	return NewResponseBuilder().
		Error(code, message, details).
		RequestID(requestID).
		WriteJSON(w, statusCode)
}

// WriteBadRequest 快速写入400错误
func WriteBadRequest(w http.ResponseWriter, code, message, requestID string, details interface{}) error {
	return WriteError(w, http.StatusBadRequest, code, message, requestID, details)
}

// WriteNotFound 快速写入404错误
func WriteNotFound(w http.ResponseWriter, message, requestID string) error {
	return WriteError(w, http.StatusNotFound, "NOT_FOUND", message, requestID, nil)
}

// WriteUnauthorized 快速写入401错误
func WriteUnauthorized(w http.ResponseWriter, requestID string) error {
	return WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "未授权访问", requestID, nil)
}

// WriteForbidden 快速写入403错误
func WriteForbidden(w http.ResponseWriter, requestID string) error {
	return WriteError(w, http.StatusForbidden, "FORBIDDEN", "访问被禁止", requestID, nil)
}

// WriteConflict 快速写入409错误
func WriteConflict(w http.ResponseWriter, code, message, requestID string, details interface{}) error {
	return WriteError(w, http.StatusConflict, code, message, requestID, details)
}

// WriteInternalError 快速写入500错误
func WriteInternalError(w http.ResponseWriter, requestID string, details interface{}) error {
	return WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误", requestID, details)
}

// WriteValidationError 快速写入验证错误响应
func WriteValidationError(w http.ResponseWriter, errors []ValidationError, requestID string) error {
	return NewResponseBuilder().
		ValidationErrors(errors).
		RequestID(requestID).
		WriteJSON(w, http.StatusBadRequest)
}

// ConvertValidationError 转换验证错误为标准格式
func ConvertValidationError(field, message string, value interface{}) ValidationError {
	return ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Code:    "FIELD_INVALID",
	}
}

// ConvertValidationErrors 转换多个验证错误
func ConvertValidationErrors(errs map[string]string) []ValidationError {
	var validationErrors []ValidationError
	for field, message := range errs {
		validationErrors = append(validationErrors, ConvertValidationError(field, message, nil))
	}
	return validationErrors
}

// WriteHealthCheck 写入健康检查响应
func WriteHealthCheck(w http.ResponseWriter, service string, healthy bool, details interface{}, requestID string) error {
	status := "unhealthy"
	statusCode := http.StatusServiceUnavailable
	
	if healthy {
		status = "healthy"
		statusCode = http.StatusOK
	}
	
	data := map[string]interface{}{
		"service": service,
		"status":  status,
		"details": details,
	}
	
	return NewResponseBuilder().
		Data(data).
		Success(healthy).
		Message(fmt.Sprintf("Service %s status check", service)).
		RequestID(requestID).
		WriteJSON(w, statusCode)
}

// WriteList 写入列表响应（带分页）
func WriteList(w http.ResponseWriter, items interface{}, pagination *PaginationMeta, message, requestID string) error {
	data := map[string]interface{}{
		"items": items,
	}
	
	if pagination != nil {
		data["pagination"] = pagination
	}
	
	return NewResponseBuilder().
		Data(data).
		Message(message).
		RequestID(requestID).
		WithPagination(pagination).
		WriteJSON(w, http.StatusOK)
}