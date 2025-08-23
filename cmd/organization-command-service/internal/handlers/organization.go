package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"organization-command-service/internal/repository"
	"organization-command-service/internal/types"
	"organization-command-service/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrganizationHandler struct {
	repo   *repository.OrganizationRepository
	logger *log.Logger
}

func NewOrganizationHandler(repo *repository.OrganizationRepository, logger *log.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req types.CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 业务验证
	if err := utils.ValidateCreateOrganization(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
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
			h.writeErrorResponse(w, http.StatusInternalServerError, "CODE_GENERATION_ERROR", "生成组织代码失败", err)
			return
		}
	}

	// 计算路径和级别
	path, level, err := h.repo.CalculatePath(r.Context(), tenantID, req.ParentCode, code)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "PARENT_ERROR", "父组织处理失败", err)
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
		Level:       level,
		Path:        path,
		SortOrder:   req.SortOrder,
		Description: req.Description,
		// 时态管理字段 - 使用Date类型
		EffectiveDate: req.EffectiveDate,
		EndDate:       req.EndDate,
		IsTemporal:    req.IsTemporal,
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
		h.writeErrorResponse(w, http.StatusInternalServerError, "CREATE_ERROR", "创建组织失败", err)
		return
	}

	// 返回成功响应
	response := h.toOrganizationResponse(createdOrg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	h.logger.Printf("✅ 组织创建成功: %s - %s", createdOrg.Code, createdOrg.Name)
}

func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	var req types.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 业务验证
	if err := utils.ValidateUpdateOrganization(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)

	// 更新组织
	updatedOrg, err := h.repo.Update(r.Context(), tenantID, code, &req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "UPDATE_ERROR", "更新组织失败", err)
		return
	}

	// 返回成功响应
	response := h.toOrganizationResponse(updatedOrg)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Printf("✅ 组织更新成功: %s - %s", updatedOrg.Code, updatedOrg.Name)
}

func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	tenantID := h.getTenantID(r)

	// 删除组织
	err := h.repo.Delete(r.Context(), tenantID, code)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "DELETE_ERROR", "删除组织失败", err)
		return
	}

	h.logger.Printf("✅ 组织删除成功: %s", code)
	w.WriteHeader(http.StatusNoContent)
}

func (h *OrganizationHandler) SuspendOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	var req types.SuspendOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	if req.Reason == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "停用原因不能为空", nil)
		return
	}

	tenantID := h.getTenantID(r)

	// 停用组织
	org, err := h.repo.Suspend(r.Context(), tenantID, code, req.Reason)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "SUSPEND_ERROR", "停用组织失败", err)
		return
	}

	// 构建响应
	response := h.toOrganizationResponse(org)
	h.logger.Printf("✅ 组织停用成功: %s - %s", response.Code, response.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *OrganizationHandler) ReactivateOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	var req types.ReactivateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	if req.Reason == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "重启原因不能为空", nil)
		return
	}

	tenantID := h.getTenantID(r)

	// 重新启用组织
	org, err := h.repo.Reactivate(r.Context(), tenantID, code, req.Reason)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "REACTIVATE_ERROR", "重新启用组织失败", err)
		return
	}

	// 构建响应
	response := h.toOrganizationResponse(org)
	h.logger.Printf("✅ 组织重新启用成功: %s - %s", response.Code, response.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *OrganizationHandler) CreateOrganizationEvent(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	var req struct {
		EventType     string `json:"event_type"`
		RecordID      string `json:"record_id"`
		EffectiveDate string `json:"effective_date"`
		ChangeReason  string `json:"change_reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	tenantID := h.getTenantID(r)

	switch req.EventType {
	case "DEACTIVATE":
		// 处理版本作废事件
		err := h.handleDeactivateEvent(r.Context(), tenantID, code, req.RecordID, req.ChangeReason)
		if err != nil {
			h.writeErrorResponse(w, http.StatusInternalServerError, "DEACTIVATE_ERROR", "作废版本失败", err)
			return
		}

		h.logger.Printf("✅ 版本作废成功: 组织 %s, 记录ID: %s", code, req.RecordID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"message":   "版本作废成功",
			"code":      code,
			"record_id": req.RecordID,
		})

	default:
		h.writeErrorResponse(w, http.StatusBadRequest, "UNSUPPORTED_EVENT", fmt.Sprintf("不支持的事件类型: %s", req.EventType), nil)
	}
}

func (h *OrganizationHandler) UpdateHistoryRecord(w http.ResponseWriter, r *http.Request) {
	recordId := chi.URLParam(r, "record_id")
	if recordId == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_RECORD_ID", "缺少记录ID", nil)
		return
	}

	// 验证UUID格式
	if _, err := uuid.Parse(recordId); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_RECORD_ID", "无效的记录ID格式", err)
		return
	}

	var req types.UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 业务验证
	if err := utils.ValidateUpdateOrganization(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)

	// 通过UUID更新历史记录
	updatedOrg, err := h.repo.UpdateByRecordId(r.Context(), tenantID, recordId, &req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "UPDATE_ERROR", "更新历史记录失败", err)
		return
	}

	// 构建响应
	response := h.toOrganizationResponse(updatedOrg)

	h.logger.Printf("✅ 历史记录更新成功: %s - %s (记录ID: %s)", response.Code, response.Name, recordId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

func (h *OrganizationHandler) toOrganizationResponse(org *types.Organization) *types.OrganizationResponse {
	return &types.OrganizationResponse{
		Code:          org.Code,
		Name:          org.Name,
		UnitType:      org.UnitType,
		Status:        org.Status,
		Level:         org.Level,
		Path:          org.Path,
		SortOrder:     org.SortOrder,
		Description:   org.Description,
		ParentCode:    org.ParentCode,
		CreatedAt:     org.CreatedAt,
		UpdatedAt:     org.UpdatedAt,
		EffectiveDate: org.EffectiveDate,
		EndDate:       org.EndDate,
		IsTemporal:    org.IsTemporal,
		ChangeReason:  org.ChangeReason,
	}
}

func (h *OrganizationHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorMsg := message
	if err != nil && statusCode >= 500 {
		h.logger.Printf("Server error: %v", err)
		errorMsg = "Internal server error"
	}

	response := types.ErrorResponse{
		Error:   errorMsg,
		Code:    code,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// SetupRoutes 设置路由
func (h *OrganizationHandler) SetupRoutes(r chi.Router) {
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		r.Post("/", h.CreateOrganization)
		r.Put("/{code}", h.UpdateOrganization)
		r.Delete("/{code}", h.DeleteOrganization)
		r.Put("/{code}/suspend", h.SuspendOrganization)
		r.Put("/{code}/reactivate", h.ReactivateOrganization)
		r.Post("/{code}/events", h.CreateOrganizationEvent)
		r.Put("/{code}/history/{record_id}", h.UpdateHistoryRecord)
	})
}

// handleDeactivateEvent 处理版本作废事件
func (h *OrganizationHandler) handleDeactivateEvent(ctx context.Context, tenantID uuid.UUID, code string, recordID string, changeReason string) error {
	// 验证UUID格式
	if _, err := uuid.Parse(recordID); err != nil {
		return fmt.Errorf("无效的记录ID格式: %w", err)
	}

	// 更新指定记录的状态为DELETED
	updateReq := &types.UpdateOrganizationRequest{
		Status:       func(s string) *string { return &s }("DELETED"),
		ChangeReason: func(s string) *string { return &s }(changeReason),
	}

	_, err := h.repo.UpdateByRecordId(ctx, tenantID, recordID, updateReq)
	if err != nil {
		return fmt.Errorf("作废记录失败: %w", err)
	}

	return nil
}
