package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"cube-castle/internal/organization/middleware"
	"cube-castle/internal/organization/repository"
	"cube-castle/internal/organization/utils"
	"cube-castle/internal/organization/validator"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

type validationFailureContext struct {
	TenantID     uuid.UUID
	ResourceType string
	ResourceID   string
	Action       string
	Payload      map[string]interface{}
}

func (h *OrganizationHandler) getTenantID(r *http.Request) uuid.UUID {
	tenantIDHeader := r.Header.Get("X-Tenant-ID")
	if tenantIDHeader != "" {
		if tenantID, err := uuid.Parse(tenantIDHeader); err == nil {
			return tenantID
		}
	}
	return types.DefaultTenantID
}

func (h *OrganizationHandler) getIfMatchValue(r *http.Request) (string, error) {
	raw := strings.TrimSpace(r.Header.Get("If-Match"))
	if raw == "" {
		return "", fmt.Errorf("missing If-Match header")
	}

	if strings.HasPrefix(strings.ToLower(raw), "w/") {
		raw = strings.TrimSpace(raw[2:])
	}

	trimmed := strings.Trim(raw, "\"")
	trimmed = strings.TrimSpace(trimmed)
	if trimmed == "" {
		return "", fmt.Errorf("invalid If-Match header")
	}

	return trimmed, nil
}

func (h *OrganizationHandler) writeValidationErrors(w http.ResponseWriter, r *http.Request, result *validator.ValidationResult, failureCtx *validationFailureContext) {
	requestID := middleware.GetRequestID(r.Context())

	var (
		firstError validator.ValidationError
		hasError   = len(result.Errors) > 0
	)

	if hasError {
		firstError = result.Errors[0]
	}

	severity := strings.ToUpper(strings.TrimSpace(firstError.Severity))
	if severity == "" {
		severity = string(validator.SeverityHigh)
	}

	httpStatus := validator.SeverityToHTTPStatus(severity)
	if httpStatus < http.StatusBadRequest {
		httpStatus = http.StatusBadRequest
	}

	responseCode := strings.TrimSpace(firstError.Code)
	if responseCode == "" {
		responseCode = "BUSINESS_RULE_VIOLATION"
	}

	responseMessage := strings.TrimSpace(firstError.Message)
	if responseMessage == "" {
		responseMessage = "业务规则校验失败"
	}

	ruleID := responseCode
	if ctx := firstError.Context; ctx != nil {
		if val, ok := ctx["ruleId"]; ok {
			ruleID = fmt.Sprintf("%v", val)
		}
	}
	ruleID = strings.TrimSpace(ruleID)
	if ruleID == "" {
		ruleID = responseCode
	}

	details := map[string]interface{}{
		"ruleId":           ruleID,
		"severity":         severity,
		"httpStatus":       httpStatus,
		"validationErrors": result.Errors,
		"errorCount":       len(result.Errors),
	}

	if len(result.Warnings) > 0 {
		details["warnings"] = result.Warnings
		details["warningCount"] = len(result.Warnings)
	}

	if field := strings.TrimSpace(firstError.Field); field != "" {
		details["field"] = field
	}

	if len(result.Context) > 0 {
		details["chainContext"] = result.Context
	}

	if ctx := firstError.Context; ctx != nil {
		metadata := make(map[string]interface{})
		for k, v := range ctx {
			if strings.EqualFold(k, "ruleId") {
				continue
			}
			metadata[k] = v
		}
		if len(metadata) > 0 {
			details["metadata"] = metadata
		}
	}

	logger := h.requestLogger(r, "writeValidationErrors", pkglogger.Fields{
		"errorCount": len(result.Errors),
		"ruleId":     ruleID,
		"severity":   severity,
		"httpStatus": httpStatus,
		"code":       responseCode,
	})

	if !hasError {
		if err := utils.WriteError(w, http.StatusBadRequest, responseCode, responseMessage, requestID, details); err != nil {
			logger.WithFields(pkglogger.Fields{"error": err}).Error("write validation error response failed")
		}
		return
	}

	if err := utils.WriteError(w, httpStatus, responseCode, responseMessage, requestID, details); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("write validation error response failed")
	}

	if failureCtx == nil || h.auditLogger == nil {
		return
	}

	auditPayload := map[string]interface{}{
		"ruleId":           ruleID,
		"severity":         severity,
		"httpStatus":       httpStatus,
		"validationErrors": result.Errors,
	}
	if len(result.Context) > 0 {
		auditPayload["chainContext"] = result.Context
	}

	if failureCtx.Payload != nil {
		payloadCopy := make(map[string]interface{}, len(failureCtx.Payload))
		for k, v := range failureCtx.Payload {
			payloadCopy[k] = v
		}
		auditPayload["payload"] = payloadCopy
	}

	if metadata, ok := details["metadata"].(map[string]interface{}); ok && len(metadata) > 0 {
		auditPayload["metadata"] = metadata
	}

	if field, ok := details["field"]; ok {
		auditPayload["field"] = field
	}

	actionName := strings.TrimSpace(failureCtx.Action)
	if actionName == "" {
		actionName = r.Method
	}

	if err := h.auditLogger.LogError(
		r.Context(),
		failureCtx.TenantID,
		failureCtx.ResourceType,
		failureCtx.ResourceID,
		actionName,
		h.getActorID(r),
		requestID,
		responseCode,
		responseMessage,
		auditPayload,
	); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Warn("audit log for validation error failed")
	}
}

func (h *OrganizationHandler) refreshHierarchyPaths(ctx context.Context, tenantID uuid.UUID, rootCode string) error {
	if h.hierarchyRepo == nil {
		return nil
	}

	visited := make(map[string]struct{})
	queue := []string{rootCode}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if _, seen := visited[current]; seen {
			continue
		}
		visited[current] = struct{}{}

		if err := h.hierarchyRepo.UpdateHierarchyPaths(ctx, current, tenantID); err != nil {
			return err
		}

		children, err := h.hierarchyRepo.GetDirectChildren(ctx, current, tenantID)
		if err != nil {
			return err
		}

		for _, child := range children {
			queue = append(queue, child.Code)
		}
	}

	return nil
}

func (h *OrganizationHandler) toOrganizationResponse(org *types.Organization) *types.OrganizationResponse {
	return &types.OrganizationResponse{
		Code:          org.Code,
		Name:          org.Name,
		UnitType:      org.UnitType,
		Status:        org.Status,
		Level:         org.Level,
		CodePath:      org.CodePath,
		NamePath:      org.NamePath,
		SortOrder:     org.SortOrder,
		Description:   org.Description,
		ParentCode:    org.ParentCode,
		CreatedAt:     org.CreatedAt,
		UpdatedAt:     org.UpdatedAt,
		EffectiveDate: org.EffectiveDate,
		EndDate:       org.EndDate,
		ChangeReason:  org.ChangeReason,
	}
}

func (h *OrganizationHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, code, message string, details interface{}) {
	errorMsg := message

	// 如果details是error类型，处理错误信息
	if err, ok := details.(error); ok && err != nil {
		if statusCode >= 500 {
			h.requestLogger(r, "writeErrorResponse", pkglogger.Fields{"status": statusCode, "code": code}).WithFields(pkglogger.Fields{"error": err}).Error("server error while handling response")
			errorMsg = "Internal server error"
			details = nil // 不向客户端暴露内部错误详情
		} else {
			details = err.Error()
		}
	}

	// 获取请求ID
	requestID := middleware.GetRequestID(r.Context())

	// 使用统一响应构建器
	logger := h.requestLogger(r, "writeErrorResponse", pkglogger.Fields{"status": statusCode, "code": code})
	if err := utils.WriteError(w, statusCode, code, errorMsg, requestID, details); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("write error response failed")
	}
}

func (h *OrganizationHandler) getActorID(r *http.Request) string {
	// 从JWT令牌或X-Mock-User头部获取用户ID
	if userID := r.Header.Get("X-Mock-User"); userID != "" {
		return userID
	}

	// 从JWT上下文获取
	if userID := r.Context().Value("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}

	// 如果无法获取用户ID，返回默认值
	return "system"
}

// getIPAddress 从请求中获取客户端IP地址

func (h *OrganizationHandler) getIPAddress(r *http.Request) string {
	// 检查X-Forwarded-For头部（代理情况）
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// 取第一个IP地址
		if idx := strings.Index(forwarded, ","); idx != -1 {
			return strings.TrimSpace(forwarded[:idx])
		}
		return strings.TrimSpace(forwarded)
	}

	// 检查X-Real-IP头部
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 使用RemoteAddr - 处理IPv6地址
	if ip := r.RemoteAddr; ip != "" {
		// 处理IPv6地址格式 [::1]:port
		if strings.HasPrefix(ip, "[") && strings.Contains(ip, "]:") {
			end := strings.Index(ip, "]:")
			if end > 0 {
				return ip[1:end] // 去除[]和端口
			}
		}
		// 处理IPv4地址格式 ip:port
		if idx := strings.LastIndex(ip, ":"); idx != -1 && !strings.Contains(ip[:idx], ":") {
			return ip[:idx]
		}
		return ip
	}

	return "127.0.0.1" // 默认本地地址
}

// handleRepositoryError 统一处理Repository层错误

func (h *OrganizationHandler) handleRepositoryError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	if err == nil {
		return
	}

	logger := h.requestLogger(r, "handleRepositoryError", pkglogger.Fields{"operation": operation})

	if errors.Is(err, repository.ErrOrganizationHasChildren) {
		h.writeErrorResponse(w, r, http.StatusConflict, "HAS_CHILD_UNITS", "存在子组织，无法删除", map[string]interface{}{
			"operation": operation,
		})
		return
	}

	if errors.Is(err, repository.ErrOrganizationPrecondition) {
		h.writeErrorResponse(w, r, http.StatusPreconditionFailed, "PRECONDITION_FAILED", "请求的版本信息已过期，请刷新后重试", nil)
		return
	}

	errorStr := err.Error()

	// PostgreSQL错误代码映射
	switch {
	// 数据不存在错误 - 包括应用层和数据库层错误
	case strings.Contains(errorStr, "not found") || strings.Contains(errorStr, "no rows") ||
		strings.Contains(errorStr, "组织不存在") || strings.Contains(errorStr, "组织代码已存在"):

		// 区分不同的错误类型
		if strings.Contains(errorStr, "组织代码已存在") {
			h.writeErrorResponse(w, r, http.StatusConflict, "DUPLICATE_CODE", "组织代码已存在", map[string]interface{}{
				"constraint": "unique_code_per_tenant",
				"operation":  operation,
			})
		} else {
			h.writeErrorResponse(w, r, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "组织单元不存在", err)
		}

	// 唯一约束违反 - 代码/时间点/当前冲突
	case strings.Contains(errorStr, "duplicate key value"):
		// 细分约束名称
		switch {
		case strings.Contains(errorStr, "uk_org_ver_active_only"):
			h.writeErrorResponse(w, r, http.StatusConflict, "TEMPORAL_POINT_CONFLICT", "(tenant_id, code, effective_date) must be unique for non-deleted versions", nil)
		case strings.Contains(errorStr, "uk_org_current_active_only"):
			h.writeErrorResponse(w, r, http.StatusConflict, "CURRENT_CONFLICT", "Only one current non-deleted version per (tenant_id, code) is allowed", nil)
		case strings.Contains(errorStr, "organization_units_code_tenant_id_key"):
			h.writeErrorResponse(w, r, http.StatusConflict, "DUPLICATE_CODE", "组织代码已存在", map[string]interface{}{
				"constraint": "unique_code_per_tenant",
				"operation":  operation,
			})
		default:
			h.writeErrorResponse(w, r, http.StatusConflict, "CONSTRAINT_VIOLATION", "数据约束违反", map[string]interface{}{
				"operation": operation,
				"type":      "database_constraint",
			})
		}

	// 单位类型约束违反
	case strings.Contains(errorStr, "organization_units_unit_type_check"):
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_UNIT_TYPE", "无效的组织类型", map[string]interface{}{
			"allowedTypes": []string{"DEPARTMENT", "ORGANIZATION_UNIT", "PROJECT_TEAM"},
			"constraint":   "unit_type_check",
		})

	// 字段长度限制
	case strings.Contains(errorStr, "value too long for type"):
		fieldName := "unknown"
		if strings.Contains(errorStr, "character varying(10)") {
			fieldName = "code"
		} else if strings.Contains(errorStr, "character varying(100)") {
			fieldName = "name"
		}
		h.writeErrorResponse(w, r, http.StatusBadRequest, "FIELD_TOO_LONG", fmt.Sprintf("字段 %s 超出长度限制", fieldName), map[string]interface{}{
			"field":      fieldName,
			"constraint": "field_length_limit",
		})

	// 外键约束违反 - 父组织不存在
	case strings.Contains(errorStr, "foreign key constraint") && strings.Contains(errorStr, "parent_code"):
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_PARENT", "父组织不存在或无效", map[string]interface{}{
			"constraint": "parent_organization_exists",
		})

	// 业务逻辑错误
	case strings.Contains(errorStr, "already suspended"):
		h.writeErrorResponse(w, r, http.StatusConflict, "ALREADY_SUSPENDED", "组织单元已处于停用状态", nil)

	case strings.Contains(errorStr, "already active"):
		h.writeErrorResponse(w, r, http.StatusConflict, "ALREADY_ACTIVE", "组织单元已处于激活状态", nil)

	case strings.Contains(errorStr, "has non-deleted child units") || strings.Contains(errorStr, "has children"):
		details := map[string]interface{}{
			"operation":  operation,
			"resolution": "Delete or reassign child units first",
		}
		var childErr *repository.OrganizationHasChildrenError
		if errors.As(err, &childErr) && childErr.Count > 0 {
			details["affectedCount"] = childErr.Count
		}
		h.writeErrorResponse(w, r, http.StatusConflict, "HAS_CHILD_UNITS", "Cannot delete organization unit with child units", details)

	// 数据库连接错误
	case strings.Contains(errorStr, "connection refused") || strings.Contains(errorStr, "timeout"):
		logger.WithFields(pkglogger.Fields{"error": err}).Error("database connection issue")
		h.writeErrorResponse(w, r, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "数据库服务暂时不可用", map[string]interface{}{
			"operation": operation,
			"retryable": true,
		})

	// 已删除记录只读
	case strings.Contains(errorStr, "READ_ONLY_DELETED") || strings.Contains(errorStr, "cannot modify deleted record"):
		h.writeErrorResponse(w, r, http.StatusConflict, "DELETED_RECORD_READ_ONLY", "已删除记录为只读，禁止修改", nil)

	// 其他数据库约束错误
	case strings.Contains(errorStr, "constraint"):
		h.writeErrorResponse(w, r, http.StatusConflict, "CONSTRAINT_VIOLATION", "数据约束违反", map[string]interface{}{
			"operation": operation,
			"type":      "database_constraint",
		})

	// 默认内部服务器错误
	default:
		logger.WithFields(pkglogger.Fields{"error": err}).Error("unhandled repository error")
		h.writeErrorResponse(w, r, http.StatusInternalServerError, fmt.Sprintf("%s_ERROR", operation), fmt.Sprintf("%s操作失败", getOperationName(operation)), map[string]interface{}{
			"operation": operation,
			"retryable": false,
		})
	}
}

// getOperationName 获取操作的中文名称

func getOperationName(operation string) string {
	operationNames := map[string]string{
		"CREATE":   "创建",
		"UPDATE":   "更新",
		"DELETE":   "删除",
		"SUSPEND":  "停用",
		"ACTIVATE": "激活",
		"QUERY":    "查询",
	}

	if name, exists := operationNames[operation]; exists {
		return name
	}
	return operation
}
