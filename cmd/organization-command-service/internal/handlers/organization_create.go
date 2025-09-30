package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"organization-command-service/internal/audit"
	"organization-command-service/internal/middleware"
	"organization-command-service/internal/types"
	"organization-command-service/internal/utils"
)

func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req types.CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	if err := utils.ValidateCreateOrganization(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)

	var code string
	if req.Code != nil && strings.TrimSpace(*req.Code) != "" {
		code = strings.TrimSpace(*req.Code)
	} else {
		var err error
		code, err = h.repo.GenerateCode(r.Context(), tenantID)
		if err != nil {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "CODE_GENERATION_ERROR", "生成组织代码失败", err)
			return
		}
	}

	fields, err := h.repo.ComputeHierarchyForNew(r.Context(), tenantID, code, req.ParentCode, req.Name)
	if err != nil {
		errorMessage := err.Error()
		switch {
		case strings.Contains(errorMessage, "父组织不存在"):
			h.writeErrorResponse(w, r, http.StatusBadRequest, "PARENT_ERROR", "父组织不存在或不可用", err)
		case strings.Contains(errorMessage, "组织名称不能为空"):
			h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "组织名称不能为空", err)
		default:
			h.writeErrorResponse(w, r, http.StatusBadRequest, "HIERARCHY_CALCULATION_FAILED", "层级路径计算失败", err)
		}
		return
	}

	now := time.Now()
	org := &types.Organization{
		TenantID:      tenantID.String(),
		Code:          code,
		ParentCode:    req.ParentCode,
		Name:          req.Name,
		UnitType:      req.UnitType,
		Status:        "ACTIVE",
		Level:         fields.Level,
		Path:          fields.Path,
		CodePath:      fields.CodePath,
		NamePath:      fields.NamePath,
		SortOrder:     req.SortOrder,
		Description:   req.Description,
		EffectiveDate: req.EffectiveDate,
		EndDate:       req.EndDate,
		ChangeReason: func() *string {
			if req.ChangeReason == "" {
				return nil
			}
			return &req.ChangeReason
		}(),
		IsCurrent: true,
	}

	if org.EffectiveDate == nil {
		today := types.NewDate(now.Year(), now.Month(), now.Day())
		org.EffectiveDate = today
	}

	createdOrg, err := h.repo.Create(r.Context(), org)
	if err != nil {
		requestID := middleware.GetRequestID(r.Context())
		actorID := h.getActorID(r)
		requestData := map[string]interface{}{
			"code":       code,
			"name":       req.Name,
			"unitType":   req.UnitType,
			"parentCode": req.ParentCode,
		}

		if logErr := h.auditLogger.LogError(
			r.Context(), tenantID, audit.ResourceTypeOrganization, code,
			"CreateOrganization", actorID, requestID, "CREATE_ERROR", err.Error(), requestData,
		); logErr != nil {
			h.logger.Printf("记录创建失败审计日志出错: %v", logErr)
		}

		h.handleRepositoryError(w, r, "CREATE", err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	actorID := h.getActorID(r)
	ipAddress := h.getIPAddress(r)

	if err := h.auditLogger.LogOrganizationCreate(r.Context(), &req, createdOrg, actorID, requestID, ipAddress); err != nil {
		h.logger.Printf("⚠️ 审计日志记录失败: %v", err)
	}

	response := h.toOrganizationResponse(createdOrg)
	if err := utils.WriteCreated(w, response, "Organization created successfully", requestID); err != nil {
		h.logger.Printf("写入创建成功响应失败: %v", err)
	}

	h.logger.Printf("✅ 组织创建成功: %s - %s (RequestID: %s)", createdOrg.Code, createdOrg.Name, requestID)
}

func (h *OrganizationHandler) CreateOrganizationVersion(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	// 验证组织代码格式
	if len(code) != 7 {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_CODE_FORMAT", "组织代码必须是7位数字", nil)
		return
	}

	var req types.CreateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 业务验证
	if err := utils.ValidateCreateVersionRequest(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)

	// 验证组织是否存在
	existingOrg, err := h.repo.GetByCode(r.Context(), tenantID, code)
	if err != nil {
		if err.Error() == "organization not found" {
			h.writeErrorResponse(w, r, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "组织不存在", nil)
			return
		}
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "DATABASE_ERROR", "查询组织失败", err)
		return
	}

	// 解析生效日期
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_DATE_FORMAT", "生效日期格式无效", err)
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_END_DATE_FORMAT", "结束日期格式无效", err)
			return
		}
		endDate = &parsed
	}

	var targetParent *string
	if req.ParentCode != nil {
		trimmed := strings.TrimSpace(*req.ParentCode)
		if trimmed != "" {
			targetParent = &trimmed
		} else {
			targetParent = nil
		}
	} else {
		targetParent = existingOrg.ParentCode
	}

	if h.validator != nil && targetParent != nil {
		validation := h.validator.ValidateTemporalParentAvailability(r.Context(), tenantID, strings.TrimSpace(*targetParent), effectiveDate)
		if !validation.Valid {
			h.writeValidationErrors(w, r, validation)
			return
		}
	}

	fields, err := h.repo.ComputeHierarchyForNew(r.Context(), tenantID, code, targetParent, req.Name)
	if err != nil {
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "父组织不存在") {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "PARENT_ERROR", "父组织不存在或不可用", err)
		} else if strings.Contains(errorMessage, "组织名称不能为空") {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "组织名称不能为空", err)
		} else {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "HIERARCHY_CALCULATION_FAILED", "层级路径计算失败", err)
		}
		return
	}

	// 创建新的时态版本
	now := time.Now()
	newVersion := &types.Organization{
		TenantID:   tenantID.String(),
		Code:       code,
		ParentCode: targetParent,
		Name:       req.Name,
		UnitType:   req.UnitType,
		Status:     "ACTIVE", // 新版本默认激活
		Level:      fields.Level,
		Path:       fields.Path,
		CodePath:   fields.CodePath,
		NamePath:   fields.NamePath,
		SortOrder: func() int {
			if req.SortOrder != nil {
				return *req.SortOrder
			}
			return existingOrg.SortOrder // 继承原有排序
		}(),
		Description: func() string {
			if req.Description != nil {
				return *req.Description
			}
			return existingOrg.Description // 继承原有描述
		}(),
		// 时态管理字段
		EffectiveDate: types.NewDateFromTime(effectiveDate),
		EndDate: func() *types.Date {
			if endDate != nil {
				return types.NewDateFromTime(*endDate)
			}
			return nil
		}(),
		// isTemporal 移除：由 endDate 是否为空派生
		ChangeReason: func() *string {
			return &req.OperationReason
		}(),
		IsCurrent: effectiveDate.Before(now) || effectiveDate.Equal(now.Truncate(24*time.Hour)),
	}

	// 🚀 使用新的时态时间轴管理器 - 实现完整的时态一致性保证
	createdVersion, err := h.timelineManager.InsertVersion(r.Context(), newVersion)
	if err != nil {
		// 检查是否是版本冲突错误
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
			h.writeErrorResponse(w, r, http.StatusConflict, "VERSION_CONFLICT", "生效日期与现有版本冲突", err)
			return
		}

		// 记录创建失败的审计日志
		requestID := middleware.GetRequestID(r.Context())
		actorID := h.getActorID(r)
		requestData := map[string]interface{}{
			"code":          code,
			"name":          req.Name,
			"unitType":      req.UnitType,
			"effectiveDate": req.EffectiveDate,
			"parentCode":    targetParent,
		}

		if logErr := h.auditLogger.LogError(
			r.Context(), tenantID, audit.ResourceTypeOrganization, existingOrg.RecordID,
			"CreateOrganizationVersion", actorID, requestID, "VERSION_CREATE_ERROR", err.Error(), requestData,
		); logErr != nil {
			h.logger.Printf("记录版本创建失败审计日志出错: %v", logErr)
		}

		h.handleRepositoryError(w, r, "CREATE_VERSION", err)
		return
	}

	// 记录版本创建成功的审计日志（排除 isCurrent/isTemporal 等动态字段）
	requestID := middleware.GetRequestID(r.Context())
	actorID := h.getActorID(r)

	// 记录审计日志 - 创建版本事件（填充变更字段）
	createdFields := []audit.FieldChange{
		{Field: "name", OldValue: nil, NewValue: req.Name, DataType: "string"},
		{Field: "unitType", OldValue: nil, NewValue: req.UnitType, DataType: "string"},
		{Field: "parentCode", OldValue: nil, NewValue: targetParent, DataType: "string"},
		{Field: "description", OldValue: nil, NewValue: req.Description, DataType: "string"},
		{Field: "effectiveDate", OldValue: nil, NewValue: req.EffectiveDate, DataType: "date"},
	}
	modifiedFields := []string{"name", "unitType", "parentCode", "description", "effectiveDate"}

	event := &audit.AuditEvent{
		TenantID:        tenantID,
		EventType:       audit.EventTypeCreate,
		ResourceType:    audit.ResourceTypeOrganization,
		ResourceID:      createdVersion.RecordID.String(),
		ActorID:         actorID,
		ActorType:       audit.ActorTypeUser,
		ActionName:      "CREATE_VERSION",
		RequestID:       requestID,
		OperationReason: req.OperationReason,
		Success:         true,
		ModifiedFields:  modifiedFields,
		Changes:         createdFields,
		AfterData: map[string]interface{}{
			"code":          createdVersion.Code,
			"name":          createdVersion.Name,
			"unitType":      req.UnitType,
			"parentCode":    targetParent,
			"description":   req.Description,
			"effectiveDate": req.EffectiveDate,
			"endDate":       req.EndDate,
			"status":        createdVersion.Status,
		},
	}

	err = h.auditLogger.LogEvent(r.Context(), event)
	if err != nil {
		h.logger.Printf("⚠️ 审计日志记录失败: %v", err)
		// 审计日志失败不影响业务操作，仅记录警告
	}

	// 构建响应数据
	responseData := map[string]interface{}{
		"recordId":      createdVersion.RecordID.String(),
		"code":          createdVersion.Code,
		"name":          createdVersion.Name,
		"effectiveDate": req.EffectiveDate,
		"status":        createdVersion.Status,
	}

	// 返回企业级成功响应
	if err := utils.WriteCreated(w, responseData, "Temporal version created successfully", requestID); err != nil {
		h.logger.Printf("写入版本创建响应失败: %v", err)
	}

	h.logger.Printf("✅ 时态版本创建成功: %s - %s (生效日期: %s, RequestID: %s)",
		createdVersion.Code, createdVersion.Name, req.EffectiveDate, requestID)
}
