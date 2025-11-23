package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"cube-castle/internal/organization/audit"
	"cube-castle/internal/organization/middleware"
	"cube-castle/internal/organization/utils"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/go-chi/chi/v5"
)

func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req types.CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	logger := h.requestLogger(r, "CreateOrganization", nil)

	tenantID := h.getTenantID(r)
	logger = logger.WithFields(pkglogger.Fields{"tenantId": tenantID.String()})

	req.ParentCode = utils.NormalizeParentCodePointer(req.ParentCode)

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
	logger = logger.WithFields(pkglogger.Fields{"code": code})

	normalizedParent := utils.NormalizeParentCodePointer(req.ParentCode)

	if h.validator != nil {
		if result := h.validator.ValidateOrganizationCreation(r.Context(), &req, tenantID); !result.Valid {
			h.writeValidationErrors(w, r, result, &validationFailureContext{
				TenantID:     tenantID,
				ResourceType: audit.ResourceTypeOrganization,
				ResourceID:   code,
				Action:       "ValidateOrganizationCreation",
				Payload: map[string]interface{}{
					"request": req,
				},
			})
			return
		}
	}

	fields, err := h.repo.ComputeHierarchyForNew(ctx, tenantID, code, normalizedParent, req.Name)
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

	now := time.Now().UTC()
	org := &types.Organization{
		TenantID:      tenantID.String(),
		Code:          code,
		ParentCode:    normalizedParent,
		Name:          req.Name,
		UnitType:      req.UnitType,
		Status:        "ACTIVE",
		Level:         fields.Level,
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

	tx, err := h.repo.BeginTx(ctx)
	if err != nil {
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "DATABASE_ERROR", "创建组织失败", err)
		return
	}
	defer tx.Rollback()

	createdOrg, err := h.repo.CreateInTransaction(ctx, tx, org)
	if err != nil {
		requestID := middleware.GetRequestID(ctx)
		actorID := h.getActorID(r)
		requestData := map[string]interface{}{
			"code":       code,
			"name":       req.Name,
			"unitType":   req.UnitType,
			"parentCode": normalizedParent,
		}

		if logErr := h.auditLogger.LogError(
			r.Context(), tenantID, audit.ResourceTypeOrganization, code,
			"CreateOrganization", actorID, requestID, "CREATE_ERROR", err.Error(), requestData,
		); logErr != nil {
			logger.WithFields(pkglogger.Fields{"error": logErr}).Warn("record create failure audit log failed")
		}

		h.handleRepositoryError(w, r, "CREATE", err)
		return
	}

	actorID := h.getActorID(r)
	if err := h.upsertStandardObject(ctx, createdOrg, actorID); err != nil {
		logger.WithFields(pkglogger.Fields{
			"error": err,
			"code":  createdOrg.Code,
		}).Error("同步标准对象失败")
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "STANDARD_OBJECT_ERROR", "同步标准对象失败", err)
		return
	}

	if err := tx.Commit(); err != nil {
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "DATABASE_ERROR", "提交事务失败", err)
		return
	}

	requestID := middleware.GetRequestID(ctx)
	ipAddress := h.getIPAddress(r)

	if err := h.auditLogger.LogOrganizationCreate(ctx, &req, createdOrg, actorID, requestID, ipAddress); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Warn("audit log for organization create failed")
	}

	response := h.toOrganizationResponse(createdOrg)
	if err := utils.WriteCreated(w, response, "Organization created successfully", requestID); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("write organization create response failed")
	}

	logger.WithFields(pkglogger.Fields{"name": createdOrg.Name}).Info("organization created")
}

func (h *OrganizationHandler) CreateOrganizationVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}
	logger := h.requestLogger(r, "CreateOrganizationVersion", pkglogger.Fields{"code": code})

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
	existingOrg, err := h.repo.GetByCode(ctx, tenantID, code)
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
		targetParent = utils.NormalizeParentCodePointer(req.ParentCode)
	} else {
		targetParent = existingOrg.ParentCode
	}

	if h.validator != nil && targetParent != nil {
		validation := h.validator.ValidateTemporalParentAvailability(ctx, tenantID, strings.TrimSpace(*targetParent), effectiveDate)
		if !validation.Valid {
			payload := map[string]interface{}{
				"effectiveDate": effectiveDate.Format("2006-01-02"),
			}
			if targetParent != nil {
				payload["parentCode"] = strings.TrimSpace(*targetParent)
			}
			h.writeValidationErrors(w, r, validation, &validationFailureContext{
				TenantID:     tenantID,
				ResourceType: audit.ResourceTypeOrganization,
				ResourceID:   code,
				Action:       "ValidateTemporalParentAvailability",
				Payload:      payload,
			})
			return
		}
	}

	fields, err := h.repo.ComputeHierarchyForNew(ctx, tenantID, code, targetParent, req.Name)
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
	now := time.Now().UTC()
	newVersion := &types.Organization{
		TenantID:   tenantID.String(),
		Code:       code,
		ParentCode: targetParent,
		Name:       req.Name,
		UnitType:   req.UnitType,
		Status:     "ACTIVE", // 新版本默认激活
		Level:      fields.Level,
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
	createdVersion, err := h.timelineManager.InsertVersion(ctx, newVersion)
	if err != nil {
		// 检查是否是版本冲突错误
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
			h.writeErrorResponse(w, r, http.StatusConflict, "VERSION_CONFLICT", "生效日期与现有版本冲突", err)
			return
		}

		// 记录创建失败的审计日志
		requestID := middleware.GetRequestID(ctx)
		actorID := h.getActorID(r)
		requestData := map[string]interface{}{
			"code":          code,
			"name":          req.Name,
			"unitType":      req.UnitType,
			"effectiveDate": req.EffectiveDate,
			"parentCode":    targetParent,
		}

		if logErr := h.auditLogger.LogError(
			ctx, tenantID, audit.ResourceTypeOrganization, existingOrg.RecordID,
			"CreateOrganizationVersion", actorID, requestID, "VERSION_CREATE_ERROR", err.Error(), requestData,
		); logErr != nil {
			logger.WithFields(pkglogger.Fields{"error": logErr}).Warn("audit log for version create failure failed")
		}

		h.handleRepositoryError(w, r, "CREATE_VERSION", err)
		return
	}

	// 记录版本创建成功的审计日志（排除 isCurrent/isTemporal 等动态字段）
	requestID := middleware.GetRequestID(ctx)
	actorID := h.getActorID(r)

	orgForSOM := h.organizationFromTimeline(existingOrg, createdVersion, req.OperationReason)
	if err := h.upsertStandardObject(ctx, orgForSOM, actorID); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("standard object upsert failed during version creation")
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "STANDARD_OBJECT_ERROR", "同步标准对象失败", err)
		return
	}

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

	if err := h.auditLogger.LogEvent(ctx, event); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Warn("audit log for version create failed")
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
		logger.WithFields(pkglogger.Fields{"error": err}).Error("write organization version response failed")
	}

	logger.WithFields(pkglogger.Fields{"name": createdVersion.Name, "effectiveDate": req.EffectiveDate}).Info("organization version created")
}
