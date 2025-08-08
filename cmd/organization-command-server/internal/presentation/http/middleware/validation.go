package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"../../pkg/errors"
	"../../pkg/types"
)

// CreateOrganizationRequest 创建组织请求结构
type CreateOrganizationRequest struct {
	Code        *string `json:"code,omitempty"`        // 可选，由系统生成
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	UnitType    string  `json:"unit_type" validate:"required"`
	Status      string  `json:"status" validate:"required"`
	Level       int     `json:"level" validate:"required,min=1,max=10"`
	ParentCode  *string `json:"parent_code,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Validate 验证创建组织请求
func (req *CreateOrganizationRequest) Validate() error {
	validationErrors := types.NewValidationErrors()

	// 验证name
	if strings.TrimSpace(req.Name) == "" {
		validationErrors.AddError("name", "Name is required", "required")
	} else if len(req.Name) > 100 {
		validationErrors.AddError("name", "Name is too long (max 100 characters)", "max_length")
	}

	// 验证unit_type
	if _, err := types.ParseUnitType(req.UnitType); err != nil {
		validationErrors.AddError("unit_type", fmt.Sprintf("Invalid unit type: %s", req.UnitType), "invalid_enum")
	}

	// 验证status
	if _, err := types.ParseStatus(req.Status); err != nil {
		validationErrors.AddError("status", fmt.Sprintf("Invalid status: %s", req.Status), "invalid_enum")
	}

	// 验证level
	if req.Level < 1 || req.Level > 10 {
		validationErrors.AddError("level", "Level must be between 1 and 10", "range")
	}

	// 验证code（如果提供）
	if req.Code != nil && *req.Code != "" {
		if _, err := types.NewOrganizationCode(*req.Code); err != nil {
			validationErrors.AddError("code", err.Error(), "invalid_format")
		}
	}

	// 验证parent_code（如果提供）
	if req.ParentCode != nil && *req.ParentCode != "" {
		if _, err := types.NewOrganizationCode(*req.ParentCode); err != nil {
			validationErrors.AddError("parent_code", err.Error(), "invalid_format")
		}
	}

	// 验证sort_order（如果提供）
	if req.SortOrder != nil && *req.SortOrder < 0 {
		validationErrors.AddError("sort_order", "Sort order must be non-negative", "min_value")
	}

	// 验证description长度（如果提供）
	if req.Description != nil && len(*req.Description) > 500 {
		validationErrors.AddError("description", "Description is too long (max 500 characters)", "max_length")
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}

// UpdateOrganizationRequest 更新组织请求结构
type UpdateOrganizationRequest struct {
	Name        *string `json:"name,omitempty"`
	Status      *string `json:"status,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Validate 验证更新组织请求
func (req *UpdateOrganizationRequest) Validate() error {
	validationErrors := types.NewValidationErrors()

	// 验证name（如果提供）
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			validationErrors.AddError("name", "Name cannot be empty", "required")
		} else if len(*req.Name) > 100 {
			validationErrors.AddError("name", "Name is too long (max 100 characters)", "max_length")
		}
	}

	// 验证status（如果提供）
	if req.Status != nil {
		if _, err := types.ParseStatus(*req.Status); err != nil {
			validationErrors.AddError("status", fmt.Sprintf("Invalid status: %s", *req.Status), "invalid_enum")
		}
	}

	// 验证sort_order（如果提供）
	if req.SortOrder != nil && *req.SortOrder < 0 {
		validationErrors.AddError("sort_order", "Sort order must be non-negative", "min_value")
	}

	// 验证description长度（如果提供）
	if req.Description != nil && len(*req.Description) > 500 {
		validationErrors.AddError("description", "Description is too long (max 500 characters)", "max_length")
	}

	if validationErrors.HasErrors() {
		return validationErrors
	}

	return nil
}

// ValidateCreateOrganizationRequest 创建组织请求验证中间件
func ValidateCreateOrganizationRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 解析请求体
		var req CreateOrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeValidationErrorResponse(w, errors.NewBadRequestError("invalid JSON format", err))
			return
		}

		// 验证请求数据
		if err := req.Validate(); err != nil {
			writeValidationErrorResponse(w, errors.NewValidationError("request validation failed", err))
			return
		}

		// 将验证后的请求存储到上下文
		ctx := context.WithValue(r.Context(), "validated_create_request", &req)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateUpdateOrganizationRequest 更新组织请求验证中间件
func ValidateUpdateOrganizationRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 解析请求体
		var req UpdateOrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeValidationErrorResponse(w, errors.NewBadRequestError("invalid JSON format", err))
			return
		}

		// 验证请求数据
		if err := req.Validate(); err != nil {
			writeValidationErrorResponse(w, errors.NewValidationError("request validation failed", err))
			return
		}

		// 将验证后的请求存储到上下文
		ctx := context.WithValue(r.Context(), "validated_update_request", &req)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetValidatedCreateRequest 从上下文获取验证后的创建请求
func GetValidatedCreateRequest(ctx context.Context) (*CreateOrganizationRequest, bool) {
	req, ok := ctx.Value("validated_create_request").(*CreateOrganizationRequest)
	return req, ok
}

// GetValidatedUpdateRequest 从上下文获取验证后的更新请求
func GetValidatedUpdateRequest(ctx context.Context) (*UpdateOrganizationRequest, bool) {
	req, ok := ctx.Value("validated_update_request").(*UpdateOrganizationRequest)
	return req, ok
}

// writeValidationErrorResponse 写入验证错误响应
func writeValidationErrorResponse(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	
	var errorResponse interface{}
	var statusCode int

	switch e := err.(type) {
	case *errors.ValidationError:
		statusCode = http.StatusBadRequest
		errorResponse = map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "VALIDATION_ERROR",
				"message": e.Message,
				"details": e.Details,
			},
			"status": "error",
		}
	case *errors.BadRequestError:
		statusCode = http.StatusBadRequest
		errorResponse = map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "BAD_REQUEST",
				"message": e.Message,
				"details": e.Details,
			},
			"status": "error",
		}
	default:
		statusCode = http.StatusInternalServerError
		errorResponse = map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "INTERNAL_ERROR",
				"message": "Internal server error",
			},
			"status": "error",
		}
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

// RequestValidator 通用请求验证器接口
type RequestValidator interface {
	Validate() error
}

// ValidateRequest 通用请求验证中间件
func ValidateRequest[T RequestValidator](contextKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req T
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeValidationErrorResponse(w, errors.NewBadRequestError("invalid JSON format", err))
				return
			}

			if err := req.Validate(); err != nil {
				writeValidationErrorResponse(w, errors.NewValidationError("request validation failed", err))
				return
			}

			ctx := context.WithValue(r.Context(), contextKey, req)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}