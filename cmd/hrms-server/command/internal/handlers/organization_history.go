package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"cube-castle/cmd/hrms-server/command/internal/middleware"
	"cube-castle/cmd/hrms-server/command/internal/types"
	"cube-castle/cmd/hrms-server/command/internal/utils"
)

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
		req.ParentCode = utils.NormalizeParentCodePointer(req.ParentCode)
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
