// Package types hosts shared API response contracts for the command service.
package types

import (
	"time"
)

// SuccessResponse 表示命令服务的标准成功响应。
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data"`
	Message   string      `json:"message"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"requestId"`
}

// EnterpriseErrorResponse 表示命令服务的标准错误响应。
type EnterpriseErrorResponse struct {
	Success   bool      `json:"success"`
	Error     ErrorInfo `json:"error"`
	Timestamp string    `json:"timestamp"`
	RequestID string    `json:"requestId"`
}

// ErrorInfo 提供错误代码、信息与可选详情。
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// OperatedByInfo 操作人信息结构 - 企业级标准
type OperatedByInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AuditInfo 审计信息结构
type AuditInfo struct {
	AuditID          string                 `json:"auditId"`
	OperationType    string                 `json:"operationType"`
	OperatedBy       OperatedByInfo         `json:"operatedBy"`
	BusinessEntityID string                 `json:"businessEntityId"`
	ChangesSummary   map[string]interface{} `json:"changesSummary"`
	OperationReason  string                 `json:"operationReason"`
	TenantID         string                 `json:"tenantId"`
	Timestamp        time.Time              `json:"timestamp"`
	RequestID        string                 `json:"requestId"`
}

// WriteSuccessResponse 写入成功响应的工具函数
func WriteSuccessResponse(data interface{}, message, requestID string) SuccessResponse {
	return SuccessResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: requestID,
	}
}

// WriteErrorResponse 写入错误响应的工具函数
func WriteErrorResponse(code, message, requestID string, details interface{}) EnterpriseErrorResponse {
	return EnterpriseErrorResponse{
		Success: false,
		Error: ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: requestID,
	}
}
