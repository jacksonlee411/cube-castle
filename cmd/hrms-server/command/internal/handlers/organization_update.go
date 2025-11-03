package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"cube-castle/cmd/hrms-server/command/internal/audit"
	"cube-castle/cmd/hrms-server/command/internal/middleware"
	"cube-castle/cmd/hrms-server/command/internal/repository"
	"cube-castle/cmd/hrms-server/command/internal/types"
	"cube-castle/cmd/hrms-server/command/internal/utils"
)

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
		req.ParentCode = utils.NormalizeParentCodePointer(req.ParentCode)
		if req.ParentCode != nil {
			trimmed := *req.ParentCode
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

	currentOrg, err := h.repo.GetByCode(r.Context(), tenantID, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.writeErrorResponse(w, r, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "组织单元不存在", err)
			return
		}
		h.handleRepositoryError(w, r, "GET_CURRENT_ORG", err)
		return
	}
	if currentOrg == nil {
		h.writeErrorResponse(w, r, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "组织单元不存在", nil)
		return
	}

	expectedETag := strings.TrimSpace(currentOrg.RecordID)
	if expectedETag == "" {
		expectedETag = currentOrg.UpdatedAt.Format(time.RFC3339Nano)
	}

	if rawIfMatch := strings.TrimSpace(r.Header.Get("If-Match")); rawIfMatch != "" {
		ifMatch, parseErr := h.getIfMatchValue(r)
		if parseErr != nil {
			h.writeErrorResponse(w, r, http.StatusPreconditionFailed, "PRECONDITION_FAILED", "缺少或无效的 If-Match 标头", parseErr)
			return
		}
		if expectedETag == "" {
			h.writeErrorResponse(w, r, http.StatusPreconditionFailed, "PRECONDITION_FAILED", "无法验证资源版本，请刷新后重试", map[string]interface{}{
				"provided": ifMatch,
			})
			return
		}
		if ifMatch != expectedETag {
			h.writeErrorResponse(w, r, http.StatusPreconditionFailed, "PRECONDITION_FAILED", "资源已发生变更，请刷新后重试", map[string]interface{}{
				"expected": expectedETag,
				"provided": ifMatch,
			})
			return
		}
	}

	// 操作原因处理（可选）
	operationReason := ""
	if req.OperationReason != nil {
		trimmed := strings.TrimSpace(*req.OperationReason)
		if trimmed != "" {
			operationReason = trimmed
		}
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

	isImmediate := effectiveDate.Before(time.Now().UTC().Add(24 * time.Hour))
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

	if resourceRecordID != "" {
		w.Header().Set("ETag", fmt.Sprintf("\"%s\"", resourceRecordID))
	}

	if err := utils.WriteSuccess(w, response, actionName+"成功", requestID); err != nil {
		h.logger.Printf("写入%s响应失败: %v", actionName, err)
	}
	h.logger.Printf("✅ %s成功: %s → %s, 生效日期=%s (RequestID: %s)", actionName, code, newStatus, req.EffectiveDate, requestID)
}
