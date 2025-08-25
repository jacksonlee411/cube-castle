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
	"organization-command-service/internal/types"
	"organization-command-service/internal/utils"
)

type OrganizationHandler struct {
	repo        *repository.OrganizationRepository
	auditLogger *audit.AuditLogger
	logger      *log.Logger
}

func NewOrganizationHandler(repo *repository.OrganizationRepository, auditLogger *audit.AuditLogger, logger *log.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		repo:        repo,
		auditLogger: auditLogger,
		logger:      logger,
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

	// 计算路径和级别
	path, level, err := h.repo.CalculatePath(r.Context(), tenantID, req.ParentCode, code)
	if err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "PARENT_ERROR", "父组织处理失败", err)
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
		// 记录创建失败的审计日志
		requestID := middleware.GetRequestID(r.Context())
		actorID := h.getActorID(r)
		ipAddress := h.getIPAddress(r)
		requestData := map[string]interface{}{
			"code":       code,
			"name":       req.Name,
			"unitType":   req.UnitType,
			"parentCode": req.ParentCode,
		}

		h.auditLogger.LogError(r.Context(), tenantID, audit.ResourceTypeOrganization, code,
			"CreateOrganization", actorID, requestID, ipAddress, "CREATE_ERROR", err.Error(), requestData)

		h.writeErrorResponse(w, r, http.StatusInternalServerError, "CREATE_ERROR", "创建组织失败", err)
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
	utils.WriteCreated(w, response, "Organization created successfully", requestID)

	h.logger.Printf("✅ 组织创建成功: %s - %s (RequestID: %s)", createdOrg.Code, createdOrg.Name, requestID)
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

	// 更新组织
	updatedOrg, err := h.repo.Update(r.Context(), tenantID, code, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, r, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "组织单元不存在", err)
		} else if strings.Contains(err.Error(), "constraint") {
			h.writeErrorResponse(w, r, http.StatusConflict, "DATA_CONSTRAINT_VIOLATION", "数据约束冲突", err)
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "UPDATE_ERROR", "更新组织失败", err)
		}
		return
	}

	// 记录审计日志
	requestID := middleware.GetRequestID(r.Context())
	actorID := h.getActorID(r)
	ipAddress := h.getIPAddress(r)
	// 注意: 这里我们没有oldOrg，所以传nil - 在实际应用中应该先获取旧数据
	err = h.auditLogger.LogOrganizationUpdate(r.Context(), code, &req, nil, updatedOrg, actorID, requestID, ipAddress)
	if err != nil {
		h.logger.Printf("⚠️ 更新审计日志记录失败: %v", err)
	}

	// 返回企业级成功响应
	response := h.toOrganizationResponse(updatedOrg)
	utils.WriteSuccess(w, response, "Organization updated successfully", requestID)

	h.logger.Printf("✅ 组织更新成功: %s - %s (RequestID: %s)", updatedOrg.Code, updatedOrg.Name, requestID)
}

func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
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

	tenantID := h.getTenantID(r)

	// 删除组织
	err := h.repo.Delete(r.Context(), tenantID, code)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, r, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "组织单元不存在", err)
		} else if strings.Contains(err.Error(), "has children") {
			h.writeErrorResponse(w, r, http.StatusConflict, "HAS_CHILDREN", "不能删除包含子组织的单元", err)
		} else if strings.Contains(err.Error(), "constraint") {
			h.writeErrorResponse(w, r, http.StatusConflict, "CONSTRAINT_VIOLATION", "存在关联数据，无法删除", err)
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "DELETE_ERROR", "删除组织失败", err)
		}
		return
	}

	// 记录审计日志
	requestID := middleware.GetRequestID(r.Context())
	actorID := h.getActorID(r)
	ipAddress := h.getIPAddress(r)
	// 注意: 这里我们没有获取要删除的组织数据，传nil - 在实际应用中应该先获取组织数据
	err = h.auditLogger.LogOrganizationDelete(r.Context(), code, nil, actorID, requestID, ipAddress)
	if err != nil {
		h.logger.Printf("⚠️ 删除审计日志记录失败: %v", err)
	}

	// 返回企业级成功响应
	utils.WriteSuccess(w, map[string]interface{}{
		"code": code,
		"deletedAt": time.Now(),
	}, "Organization deleted successfully", requestID)

	h.logger.Printf("✅ 组织删除成功: %s (RequestID: %s)", code, requestID)
}

func (h *OrganizationHandler) SuspendOrganization(w http.ResponseWriter, r *http.Request) {
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

	var req types.SuspendOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 验证停用请求
	if err := utils.ValidateSuspendRequest(req.Reason); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "停用原因验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)

	// 停用组织
	org, err := h.repo.Suspend(r.Context(), tenantID, code, req.Reason)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, r, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "组织单元不存在", err)
		} else if strings.Contains(err.Error(), "already suspended") {
			h.writeErrorResponse(w, r, http.StatusConflict, "ALREADY_SUSPENDED", "组织单元已处于停用状态", err)
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "SUSPEND_ERROR", "停用组织失败", err)
		}
		return
	}

	// 记录审计日志
	requestID := middleware.GetRequestID(r.Context())
	actorID := h.getActorID(r)
	ipAddress := h.getIPAddress(r)
	err = h.auditLogger.LogOrganizationSuspend(r.Context(), code, org, actorID, requestID, ipAddress)
	if err != nil {
		h.logger.Printf("⚠️ 停用审计日志记录失败: %v", err)
	}

	// 构建企业级成功响应
	response := h.toOrganizationResponse(org)
	utils.WriteSuccess(w, response, "Organization suspended successfully", requestID)

	h.logger.Printf("✅ 组织停用成功: %s - %s (RequestID: %s)", response.Code, response.Name, requestID)
}

func (h *OrganizationHandler) ActivateOrganization(w http.ResponseWriter, r *http.Request) {
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

	var req types.ReactivateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 验证激活请求
	if err := utils.ValidateActivateRequest(req.Reason); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "激活原因验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)

	// 重新启用组织
	org, err := h.repo.Activate(r.Context(), tenantID, code, req.Reason)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, r, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "组织单元不存在", err)
		} else if strings.Contains(err.Error(), "already active") {
			h.writeErrorResponse(w, r, http.StatusConflict, "ALREADY_ACTIVE", "组织单元已处于激活状态", err)
		} else {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "ACTIVATE_ERROR", "激活组织失败", err)
		}
		return
	}

	// 记录审计日志
	requestID := middleware.GetRequestID(r.Context())
	actorID := h.getActorID(r)
	ipAddress := h.getIPAddress(r)
	err = h.auditLogger.LogOrganizationActivate(r.Context(), code, org, actorID, requestID, ipAddress)
	if err != nil {
		h.logger.Printf("⚠️ 激活审计日志记录失败: %v", err)
	}

	// 构建企业级成功响应
	response := h.toOrganizationResponse(org)
	utils.WriteSuccess(w, response, "Organization activated successfully", requestID)

	h.logger.Printf("✅ 组织激活成功: %s - %s (RequestID: %s)", response.Code, response.Name, requestID)
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
		err := h.handleDeactivateEvent(r.Context(), tenantID, code, req.RecordID, req.ChangeReason)
		if err != nil {
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "DEACTIVATE_ERROR", "作废版本失败", err)
			return
		}

		h.logger.Printf("✅ 版本作废成功: 组织 %s, 记录ID: %s", code, req.RecordID)
		requestID := middleware.GetRequestID(r.Context())
		utils.WriteSuccess(w, map[string]interface{}{
			"code":      code,
			"record_id": req.RecordID,
		}, "版本作废成功", requestID)

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

	// 通过UUID更新历史记录
	updatedOrg, err := h.repo.UpdateByRecordId(r.Context(), tenantID, recordId, &req)
	if err != nil {
		h.writeErrorResponse(w, r, http.StatusInternalServerError, "UPDATE_ERROR", "更新历史记录失败", err)
		return
	}

	// 构建企业级成功响应
	response := h.toOrganizationResponse(updatedOrg)
	requestID := middleware.GetRequestID(r.Context())
	utils.WriteSuccess(w, response, "History record updated successfully", requestID)

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

func (h *OrganizationHandler) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, code, message string, err error) {
	errorMsg := message
	var details interface{}
	if err != nil {
		if statusCode >= 500 {
			h.logger.Printf("Server error: %v", err)
			errorMsg = "Internal server error"
		} else {
			details = err.Error()
		}
	}

	// 获取请求ID
	requestID := middleware.GetRequestID(r.Context())

	// 使用统一响应构建器
	utils.WriteError(w, statusCode, code, errorMsg, requestID, details)
}

// SetupRoutes 设置路由
func (h *OrganizationHandler) SetupRoutes(r chi.Router) {
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		r.Post("/", h.CreateOrganization)
		r.Put("/{code}", h.UpdateOrganization)
		r.Delete("/{code}", h.DeleteOrganization)
		r.Post("/{code}/suspend", h.SuspendOrganization)
		r.Post("/{code}/activate", h.ActivateOrganization)
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
