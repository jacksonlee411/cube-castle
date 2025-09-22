package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"organization-command-service/internal/audit"
	"organization-command-service/internal/middleware"
	"organization-command-service/internal/repository"
	"organization-command-service/internal/services"
	"organization-command-service/internal/types"
	"organization-command-service/internal/utils"
	"organization-command-service/internal/validators"
)

type OrganizationHandler struct {
	repo            *repository.OrganizationRepository
	temporalService *services.TemporalService
	auditLogger     *audit.AuditLogger
	logger          *log.Logger
	timelineManager *repository.TemporalTimelineManager
	hierarchyRepo   *repository.HierarchyRepository
	validator       *validators.BusinessRuleValidator
}

func NewOrganizationHandler(repo *repository.OrganizationRepository, temporalService *services.TemporalService, auditLogger *audit.AuditLogger, logger *log.Logger, timelineManager *repository.TemporalTimelineManager, hierarchyRepo *repository.HierarchyRepository, validator *validators.BusinessRuleValidator) *OrganizationHandler {
	return &OrganizationHandler{
		repo:            repo,
		temporalService: temporalService,
		auditLogger:     auditLogger,
		logger:          logger,
		timelineManager: timelineManager,
		hierarchyRepo:   hierarchyRepo,
		validator:       validator,
	}
}

func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req types.CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 业务验证
	if err := utils.ValidateCreateOrganization(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)

	// 确定组织代码 - 支持指定代码（用于时态记录）
	var code string
	if req.Code != nil && strings.TrimSpace(*req.Code) != "" {
		// 使用指定的代码（通常用于创建时态记录）
		code = strings.TrimSpace(*req.Code)
	} else {
		// 生成新的组织代码
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
		if strings.Contains(errorMessage, "父组织不存在") {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "PARENT_ERROR", "父组织不存在或不可用", err)
		} else if strings.Contains(errorMessage, "组织名称不能为空") {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "组织名称不能为空", err)
		} else {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "HIERARCHY_CALCULATION_FAILED", "层级路径计算失败", err)
		}
		return
	}

	// 创建组织实体
	now := time.Now()
	org := &types.Organization{
		TenantID:    tenantID.String(),
		Code:        code,
		ParentCode:  req.ParentCode,
		Name:        req.Name,
		UnitType:    req.UnitType,
		Status:      "ACTIVE",
		Level:       fields.Level,
		Path:        fields.Path,
		CodePath:    fields.CodePath,
		NamePath:    fields.NamePath,
		SortOrder:   req.SortOrder,
		Description: req.Description,
		// 时态管理字段 - 使用Date类型
		EffectiveDate: req.EffectiveDate,
		EndDate:       req.EndDate,
		// isTemporal 移除：由 endDate 是否为空派生
		ChangeReason: func() *string {
			if req.ChangeReason == "" {
				return nil
			} else {
				return &req.ChangeReason
			}
		}(),
		IsCurrent: true, // 新创建的记录默认为当前记录
	}

	// 确保effective_date字段始终有值（数据库约束要求）
	if org.EffectiveDate == nil {
		today := types.NewDate(now.Year(), now.Month(), now.Day())
		org.EffectiveDate = today
	}

	// 调用Repository创建
	createdOrg, err := h.repo.Create(r.Context(), org)
	if err != nil {
		// 记录创建失败的审计日志
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

	// 记录组织创建成功的审计日志
	requestID := middleware.GetRequestID(r.Context())
	actorID := h.getActorID(r)
	ipAddress := h.getIPAddress(r)

	err = h.auditLogger.LogOrganizationCreate(r.Context(), &req, createdOrg, actorID, requestID, ipAddress)
	if err != nil {
		h.logger.Printf("⚠️ 审计日志记录失败: %v", err)
		// 审计日志失败不影响业务操作，仅记录警告
	}

	// 返回企业级成功响应
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

func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	// 验证组织代码格式
	if err := utils.ValidateOrganizationCode(code); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_CODE_FORMAT", "组织代码格式无效", err)
		return
	}

	var req types.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 业务验证
	if err := utils.ValidateUpdateOrganization(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)
	parentProvided := req.ParentCode != nil
	if parentProvided {
		trimmed := strings.TrimSpace(*req.ParentCode)
		if trimmed == "" {
			req.ParentCode = nil
		} else {
			req.ParentCode = &trimmed
			if trimmed == code {
				h.logger.Printf("⚠️ circular reference attempt: code=%s parentCode=%s", code, trimmed)
				h.writeErrorResponse(w, r, http.StatusBadRequest, "BUSINESS_RULE_VIOLATION", "父组织不能指向自身", nil)
				return
			}
		}
	}

	if h.validator != nil {
		if result := h.validator.ValidateOrganizationUpdate(r.Context(), code, &req, tenantID); !result.Valid {
			h.writeValidationErrors(w, r, result)
			return
		}
	}

	// 先获取当前组织数据用于审计日志
	oldOrg, err := h.repo.GetByCode(r.Context(), tenantID, code)
	if err != nil {
		h.handleRepositoryError(w, r, "GET_OLD_DATA", err)
		return
	}

	parentChanged := false
	if parentProvided {
		switch {
		case oldOrg.ParentCode == nil && req.ParentCode != nil:
			parentChanged = true
		case oldOrg.ParentCode != nil && req.ParentCode == nil:
			parentChanged = true
		case oldOrg.ParentCode != nil && req.ParentCode != nil && *oldOrg.ParentCode != *req.ParentCode:
			parentChanged = true
		}
	}

	// 更新组织
	updatedOrg, err := h.repo.Update(r.Context(), tenantID, code, &req)
	if err != nil {
		h.handleRepositoryError(w, r, "UPDATE", err)
		return
	}

	if parentChanged {
		if err := h.refreshHierarchyPaths(r.Context(), tenantID, updatedOrg.Code); err != nil {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "HIERARCHY_UPDATE_FAILED", "层级路径更新失败", err)
			return
		}
	}

	// 记录完整审计日志（包含变更前数据）
	requestID := middleware.GetRequestID(r.Context())
	actorID := h.getActorID(r)
	ipAddress := h.getIPAddress(r)
	err = h.auditLogger.LogOrganizationUpdate(r.Context(), code, &req, oldOrg, updatedOrg, actorID, requestID, ipAddress)
	if err != nil {
		h.logger.Printf("⚠️ 更新审计日志记录失败: %v", err)
	}

	// 返回企业级成功响应
	response := h.toOrganizationResponse(updatedOrg)
	if err := utils.WriteSuccess(w, response, "Organization updated successfully", requestID); err != nil {
		h.logger.Printf("写入组织更新响应失败: %v", err)
	}

	h.logger.Printf("✅ 组织更新成功: %s - %s (RequestID: %s)", updatedOrg.Code, updatedOrg.Name, requestID)
}

// SuspendOrganization 暂停组织 - 实现第四大核心场景之暂停
// 使用时态时间轴管理器实现状态变更
func (h *OrganizationHandler) SuspendOrganization(w http.ResponseWriter, r *http.Request) {
	h.changeOrganizationStatusWithTimeline(w, r, "INACTIVE", "SUSPEND", "暂停组织")
}

// ActivateOrganization 激活组织 - 实现第四大核心场景之激活
// 使用时态时间轴管理器实现状态变更
func (h *OrganizationHandler) ActivateOrganization(w http.ResponseWriter, r *http.Request) {
	h.changeOrganizationStatusWithTimeline(w, r, "ACTIVE", "REACTIVATE", "激活组织")
}

// changeOrganizationStatusWithTimeline 通用的组织状态变更handler - 使用时态时间轴管理器
func (h *OrganizationHandler) changeOrganizationStatusWithTimeline(w http.ResponseWriter, r *http.Request, newStatus, operationType, actionName string) {
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

	var req struct {
		EffectiveDate   string  `json:"effectiveDate"`   // 生效日期，格式：2006-01-02
		OperationReason *string `json:"operationReason"` // 操作原因
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 解析生效日期
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_DATE_FORMAT", "生效日期格式无效", err)
		return
	}

	tenantID := h.getTenantID(r)
	requestID := middleware.GetRequestID(r.Context())
	actorID := h.getActorID(r)

	// 操作原因处理
	operationReason := actionName
	if req.OperationReason != nil {
		operationReason = *req.OperationReason
	}

	// 🚀 使用时态时间轴管理器变更组织状态
	var timeline *[]repository.TimelineVersion
	if operationType == "SUSPEND" {
		timeline, err = h.timelineManager.SuspendOrganization(r.Context(), tenantID, code, effectiveDate, operationReason)
	} else {
		timeline, err = h.timelineManager.ActivateOrganization(r.Context(), tenantID, code, effectiveDate, operationReason)
	}

	if err != nil {
		// 记录操作失败的审计日志
		if logErr := h.auditLogger.LogError(
			r.Context(), tenantID, audit.ResourceTypeOrganization, code,
			operationType, actorID, requestID, operationType+"_ERROR", err.Error(), map[string]interface{}{
				"code":            code,
				"targetStatus":    newStatus,
				"effectiveDate":   req.EffectiveDate,
				"operationReason": operationReason,
			},
		); logErr != nil {
			h.logger.Printf("记录%s失败审计日志出错: %v", operationType, logErr)
		}

		// 检查是否是冲突错误
		if strings.Contains(err.Error(), "TEMPORAL_POINT_CONFLICT") {
			h.writeErrorResponse(w, r, http.StatusConflict, "TEMPORAL_CONFLICT", "生效日期与现有版本冲突", err)
			return
		}

		h.writeErrorResponse(w, r, http.StatusInternalServerError, operationType+"_FAILED", actionName+"失败", err)
		return
	}

	// 记录成功的审计日志（使用具体版本的 recordId 作为资源ID）
	var resourceRecordID string
	if timeline != nil {
		for _, v := range *timeline {
			if v.EffectiveDate.Equal(effectiveDate) && v.Status == newStatus {
				resourceRecordID = v.RecordID.String()
				break
			}
		}
		if resourceRecordID == "" {
			for _, v := range *timeline {
				if v.IsCurrent {
					resourceRecordID = v.RecordID.String()
					break
				}
			}
		}
	}
	if resourceRecordID == "" {
		// 最后兜底：查询当前版本的 RecordID
		if cur, err := h.repo.GetByCode(r.Context(), tenantID, code); err == nil && cur != nil {
			resourceRecordID = cur.RecordID
		}
	}

	event := &audit.AuditEvent{
		ID:              uuid.New(),
		TenantID:        tenantID,
		EventType:       audit.EventTypeUpdate,
		ResourceType:    audit.ResourceTypeOrganization,
		ResourceID:      resourceRecordID,
		ActorID:         actorID,
		ActorType:       audit.ActorTypeUser,
		ActionName:      operationType,
		RequestID:       requestID,
		OperationReason: operationReason,
		Timestamp:       time.Now(),
		Success:         true,
		BeforeData: map[string]interface{}{
			"code": code,
		},
		AfterData: map[string]interface{}{
			"targetStatus":     newStatus,
			"effectiveDate":    req.EffectiveDate,
			"timelineVersions": len(*timeline),
			"operationReason":  operationReason,
		},
	}

	if err := h.auditLogger.LogEvent(r.Context(), event); err != nil {
		h.logger.Printf("⚠️ 记录审计日志失败: %v", err)
	}

	// 构造响应 - 返回更新后的时间轴
	timelineResponse := make([]map[string]interface{}, len(*timeline))
	for i, version := range *timeline {
		timelineResponse[i] = map[string]interface{}{
			"recordId":      version.RecordID,
			"code":          version.Code,
			"name":          version.Name,
			"effectiveDate": version.EffectiveDate.Format("2006-01-02"),
			"endDate": func() *string {
				if version.EndDate != nil {
					endDateStr := version.EndDate.Format("2006-01-02")
					return &endDateStr
				}
				return nil
			}(),
			"isCurrent": version.IsCurrent,
			"status":    version.Status,
		}
	}

	isImmediate := effectiveDate.Before(time.Now().Add(24 * time.Hour))
	message := fmt.Sprintf("%s成功（%s生效），时间轴已自动调整", actionName,
		func() string {
			if isImmediate {
				return "即时"
			}
			return "计划"
		}())

	response := map[string]interface{}{
		"message":         message,
		"operationType":   operationType,
		"targetStatus":    newStatus,
		"effectiveDate":   req.EffectiveDate,
		"operationReason": operationReason,
		"isImmediate":     isImmediate,
		"timeline":        timelineResponse,
	}

	if err := utils.WriteSuccess(w, response, actionName+"成功", requestID); err != nil {
		h.logger.Printf("写入%s响应失败: %v", actionName, err)
	}
	h.logger.Printf("✅ %s成功: %s → %s, 生效日期=%s (RequestID: %s)", actionName, code, newStatus, req.EffectiveDate, requestID)
}

func (h *OrganizationHandler) CreateOrganizationEvent(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	var req struct {
		EventType     string `json:"eventType"`
		RecordID      string `json:"recordId"`
		EffectiveDate string `json:"effectiveDate"`
		ChangeReason  string `json:"changeReason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	tenantID := h.getTenantID(r)

	switch req.EventType {
	case "DEACTIVATE":
		// 处理版本作废事件
		actorID := h.getActorID(r)
		requestID := middleware.GetRequestID(r.Context())

		err := h.handleDeactivateEvent(r.Context(), tenantID, code, req.RecordID, req.ChangeReason, actorID, requestID)
		if err != nil {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "DEACTIVATE_ERROR", "作废版本失败", err)
			return
		}

		// 获取最新时间线（非删除记录），用于前端立即刷新，避免读缓存延迟
		versions, listErr := h.repo.ListVersionsByCode(r.Context(), tenantID, code)
		if listErr != nil {
			h.logger.Printf("⚠️ 获取最新时间线失败（不影响作废结果）: %v", listErr)
		}

		// 构建轻量时间线返回
		timeline := make([]map[string]interface{}, 0, len(versions))
		for _, v := range versions {
			timeline = append(timeline, map[string]interface{}{
				"recordId": v.RecordID,
				"code":     v.Code,
				"name":     v.Name,
				"unitType": v.UnitType,
				"status":   v.Status,
				"level":    v.Level,
				"effectiveDate": func() string {
					if v.EffectiveDate != nil {
						return v.EffectiveDate.String()
					}
					return ""
				}(),
				"endDate": func() *string {
					if v.EndDate != nil {
						s := v.EndDate.String()
						return &s
					}
					return nil
				}(),
				"isCurrent":   v.IsCurrent,
				"createdAt":   v.CreatedAt,
				"updatedAt":   v.UpdatedAt,
				"parentCode":  v.ParentCode,
				"description": v.Description,
			})
		}

		h.logger.Printf("✅ 版本作废成功: 组织 %s, 记录ID: %s (返回最新时间线%d条)", code, req.RecordID, len(timeline))
		if err := utils.WriteSuccess(w, map[string]interface{}{
			"code":      code,
			"record_id": req.RecordID,
			"timeline":  timeline,
		}, "版本作废成功", requestID); err != nil {
			h.logger.Printf("写入版本作废响应失败: %v", err)
		}

	default:
		h.writeErrorResponse(w, r, http.StatusBadRequest, "UNSUPPORTED_EVENT", fmt.Sprintf("不支持的事件类型: %s", req.EventType), nil)
	}
}

func (h *OrganizationHandler) UpdateHistoryRecord(w http.ResponseWriter, r *http.Request) {
	recordId := chi.URLParam(r, "record_id")
	if recordId == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_RECORD_ID", "缺少记录ID", nil)
		return
	}

	// 验证UUID格式
	if _, err := uuid.Parse(recordId); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_RECORD_ID", "无效的记录ID格式", err)
		return
	}

	var req types.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 业务验证
	if err := utils.ValidateUpdateOrganization(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)
	parentProvided := req.ParentCode != nil
	if parentProvided {
		trimmed := strings.TrimSpace(*req.ParentCode)
		if trimmed == "" {
			req.ParentCode = nil
		} else {
			req.ParentCode = &trimmed
		}
	}

	// 先获取当前记录数据用于审计日志
	oldOrg, err := h.repo.GetByRecordId(r.Context(), tenantID, recordId)
	if err != nil {
		h.handleRepositoryError(w, r, "GET_OLD_RECORD", err)
		return
	}

	if h.validator != nil {
		if result := h.validator.ValidateOrganizationUpdate(r.Context(), oldOrg.Code, &req, tenantID); !result.Valid {
			h.writeValidationErrors(w, r, result)
			return
		}
	}

	if req.ParentCode != nil && *req.ParentCode == oldOrg.Code {
		h.logger.Printf("⚠️ circular reference attempt: code=%s parentCode=%s", oldOrg.Code, *req.ParentCode)
		h.writeErrorResponse(w, r, http.StatusBadRequest, "BUSINESS_RULE_VIOLATION", "父组织不能指向自身", nil)
		return
	}

	parentChanged := false
	if parentProvided {
		switch {
		case oldOrg.ParentCode == nil && req.ParentCode != nil:
			parentChanged = true
		case oldOrg.ParentCode != nil && req.ParentCode == nil:
			parentChanged = true
		case oldOrg.ParentCode != nil && req.ParentCode != nil && *oldOrg.ParentCode != *req.ParentCode:
			parentChanged = true
		}
	}

	// 通过UUID更新历史记录
	updatedOrg, err := h.repo.UpdateByRecordId(r.Context(), tenantID, recordId, &req)
	if err != nil {
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "UPDATE_ERROR", "更新历史记录失败", err)
		return
	}

	if parentChanged {
		if err := h.refreshHierarchyPaths(r.Context(), tenantID, updatedOrg.Code); err != nil {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "HIERARCHY_UPDATE_FAILED", "层级路径更新失败", err)
			return
		}
	}

	// 记录完整审计日志（包含变更前数据）
	requestID := middleware.GetRequestID(r.Context())
	actorID := h.getActorID(r)
	ipAddress := h.getIPAddress(r)
	err = h.auditLogger.LogOrganizationUpdate(r.Context(), updatedOrg.Code, &req, oldOrg, updatedOrg, actorID, requestID, ipAddress)
	if err != nil {
		h.logger.Printf("⚠️ 历史记录更新审计日志记录失败: %v", err)
	}

	// 构建企业级成功响应
	response := h.toOrganizationResponse(updatedOrg)
	if err := utils.WriteSuccess(w, response, "History record updated successfully", requestID); err != nil {
		h.logger.Printf("写入历史记录更新响应失败: %v", err)
	}

	h.logger.Printf("✅ 历史记录更新成功: %s - %s (记录ID: %s, RequestID: %s)", response.Code, response.Name, recordId, requestID)
}

// 辅助方法
func (h *OrganizationHandler) getTenantID(r *http.Request) uuid.UUID {
	tenantIDHeader := r.Header.Get("X-Tenant-ID")
	if tenantIDHeader != "" {
		if tenantID, err := uuid.Parse(tenantIDHeader); err == nil {
			return tenantID
		}
	}
	return types.DefaultTenantID
}

func (h *OrganizationHandler) writeValidationErrors(w http.ResponseWriter, r *http.Request, result *validators.ValidationResult) {
	requestID := middleware.GetRequestID(r.Context())

	if len(result.Errors) == 0 {
		if err := utils.WriteError(w, http.StatusBadRequest, "BUSINESS_RULE_VIOLATION", "业务规则校验失败", requestID, map[string]interface{}{
			"validationErrors": []validators.ValidationError{},
			"errorCount":       0,
		}); err != nil {
			h.logger.Printf("写入验证错误响应失败: %v", err)
		}
		return
	}

	firstError := result.Errors[0]
	details := map[string]interface{}{
		"validationErrors": result.Errors,
		"errorCount":       len(result.Errors),
	}

	if err := utils.WriteError(w, http.StatusBadRequest, firstError.Code, firstError.Message, requestID, details); err != nil {
		h.logger.Printf("写入验证错误响应失败: %v", err)
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
		Path:          org.Path,
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
			h.logger.Printf("Server error: %v", err)
			errorMsg = "Internal server error"
			details = nil // 不向客户端暴露内部错误详情
		} else {
			details = err.Error()
		}
	}

	// 获取请求ID
	requestID := middleware.GetRequestID(r.Context())

	// 使用统一响应构建器
	if err := utils.WriteError(w, statusCode, code, errorMsg, requestID, details); err != nil {
		h.logger.Printf("写入错误响应失败: %v", err)
	}
}

// SetupRoutes 设置路由
func (h *OrganizationHandler) SetupRoutes(r chi.Router) {
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		r.Post("/", h.CreateOrganization)
		r.Put("/{code}", h.UpdateOrganization)
		r.Post("/{code}/suspend", h.SuspendOrganization)
		r.Post("/{code}/activate", h.ActivateOrganization)
		// 🚀 时态版本管理端点 - 严格遵循API契约
		r.Post("/{code}/versions", h.CreateOrganizationVersion)
		// 注意: 删除版本请使用 POST /{code}/events (DEACTIVATE)
		// 注意: 修改生效日期请使用 PUT /{code}/history/{record_id}
		// 事件处理和历史记录
		r.Post("/{code}/events", h.CreateOrganizationEvent)
		r.Put("/{code}/history/{record_id}", h.UpdateHistoryRecord)
	})
}

// handleDeactivateEvent 处理版本作废事件
func (h *OrganizationHandler) handleDeactivateEvent(ctx context.Context, tenantID uuid.UUID, code string, recordID string, changeReason string, actorID string, requestID string) error {
	// 验证UUID格式
	if _, err := uuid.Parse(recordID); err != nil {
		return fmt.Errorf("无效的记录ID格式: %w", err)
	}

	// 获取删除前的组织数据用于审计日志
	oldOrg, err := h.repo.GetByRecordId(ctx, tenantID, recordID)
	if err != nil {
		return fmt.Errorf("获取记录失败: %w", err)
	}

	// 使用时间线管理器执行“单事务 软删 + 全链重算”
	rid, _ := uuid.Parse(recordID)
	if _, err := h.timelineManager.DeleteVersion(ctx, tenantID, rid); err != nil {
		return fmt.Errorf("作废记录失败: %w", err)
	}

	// 记录审计日志 - 使用删除日志方法
	err = h.auditLogger.LogOrganizationDelete(ctx, tenantID, code, oldOrg, actorID, requestID, changeReason)
	if err != nil {
		h.logger.Printf("⚠️ 审计日志记录失败 (但操作成功): %v", err)
		// 审计日志失败不应该导致业务操作失败，只记录警告
	}

	h.logger.Printf("📋 审计日志已记录: 作废组织版本 %s (记录ID: %s)", code, recordID)

	return nil
}

// getActorID 从请求中获取操作者ID
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

	case strings.Contains(errorStr, "has children"):
		h.writeErrorResponse(w, r, http.StatusConflict, "HAS_CHILDREN", "不能删除包含子组织的单元", map[string]interface{}{
			"operation":  operation,
			"suggestion": "请先删除所有子组织单元",
		})

	// 数据库连接错误
	case strings.Contains(errorStr, "connection refused") || strings.Contains(errorStr, "timeout"):
		h.logger.Printf("Database connection error in %s operation: %v", operation, err)
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
		h.logger.Printf("Unhandled repository error in %s operation: %v", operation, err)
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
