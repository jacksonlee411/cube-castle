package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"cube-castle/cmd/hrms-server/command/internal/middleware"
	"cube-castle/cmd/hrms-server/command/internal/services"
	"cube-castle/cmd/hrms-server/command/internal/types"
	"cube-castle/cmd/hrms-server/command/internal/utils"
)

type JobCatalogHandler struct {
	service *services.JobCatalogService
	logger  *log.Logger
}

func NewJobCatalogHandler(service *services.JobCatalogService, logger *log.Logger) *JobCatalogHandler {
	return &JobCatalogHandler{service: service, logger: logger}
}

func (h *JobCatalogHandler) SetupRoutes(r chi.Router) {
	r.Post("/api/v1/job-family-groups", h.CreateJobFamilyGroup)
	r.Put("/api/v1/job-family-groups/{code}", h.UpdateJobFamilyGroup)
	r.Post("/api/v1/job-family-groups/{code}/versions", h.CreateJobFamilyGroupVersion)
	r.Post("/api/v1/job-families", h.CreateJobFamily)
	r.Put("/api/v1/job-families/{code}", h.UpdateJobFamily)
	r.Post("/api/v1/job-families/{code}/versions", h.CreateJobFamilyVersion)
	r.Post("/api/v1/job-roles", h.CreateJobRole)
	r.Put("/api/v1/job-roles/{code}", h.UpdateJobRole)
	r.Post("/api/v1/job-roles/{code}/versions", h.CreateJobRoleVersion)
	r.Post("/api/v1/job-levels", h.CreateJobLevel)
	r.Put("/api/v1/job-levels/{code}", h.UpdateJobLevel)
	r.Post("/api/v1/job-levels/{code}/versions", h.CreateJobLevelVersion)
}

func (h *JobCatalogHandler) CreateJobFamilyGroup(w http.ResponseWriter, r *http.Request) {
	var req types.CreateJobFamilyGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobFamilyGroup(r.Context(), tenantID, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job family group created successfully", requestID); err != nil {
		h.logger.Printf("写入职类创建响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) UpdateJobFamilyGroup(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职类代码", nil)
		return
	}

	var req types.UpdateJobFamilyGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Status = strings.TrimSpace(req.Status)
	req.EffectiveDate = strings.TrimSpace(req.EffectiveDate)
	if req.Name == "" || req.Status == "" || req.EffectiveDate == "" {
		h.writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "名称、状态与生效日期为必填项", map[string]interface{}{
			"name":          req.Name,
			"status":        req.Status,
			"effectiveDate": req.EffectiveDate,
		})
		return
	}
	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)
	ifMatch := getIfMatchHeader(r)

	entity, err := h.service.UpdateJobFamilyGroup(r.Context(), tenantID, code, &req, ifMatch, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, entity, "Job family group updated successfully", requestID); err != nil {
		h.logger.Printf("写入职类更新响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) CreateJobFamilyGroupVersion(w http.ResponseWriter, r *http.Request) {
	var req types.JobCatalogVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职类代码", nil)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobFamilyGroupVersion(r.Context(), tenantID, code, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job family group version created successfully", requestID); err != nil {
		h.logger.Printf("写入职类版本响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) CreateJobFamily(w http.ResponseWriter, r *http.Request) {
	var req types.CreateJobFamilyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobFamily(r.Context(), tenantID, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job family created successfully", requestID); err != nil {
		h.logger.Printf("写入职种创建响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) UpdateJobFamily(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职种代码", nil)
		return
	}

	var req types.UpdateJobFamilyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Status = strings.TrimSpace(req.Status)
	req.EffectiveDate = strings.TrimSpace(req.EffectiveDate)
	if req.Name == "" || req.Status == "" || req.EffectiveDate == "" {
		h.writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "名称、状态与生效日期为必填项", map[string]interface{}{
			"name":          req.Name,
			"status":        req.Status,
			"effectiveDate": req.EffectiveDate,
		})
		return
	}
	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
	}
	if req.JobFamilyGroupCode != nil {
		trimmed := strings.ToUpper(strings.TrimSpace(*req.JobFamilyGroupCode))
		if trimmed == "" {
			req.JobFamilyGroupCode = nil
		} else {
			req.JobFamilyGroupCode = &trimmed
		}
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)
	ifMatch := getIfMatchHeader(r)

	entity, err := h.service.UpdateJobFamily(r.Context(), tenantID, code, &req, ifMatch, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, entity, "Job family updated successfully", requestID); err != nil {
		h.logger.Printf("写入职种更新响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) CreateJobFamilyVersion(w http.ResponseWriter, r *http.Request) {
	var req types.JobCatalogVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职种代码", nil)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobFamilyVersion(r.Context(), tenantID, code, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job family version created successfully", requestID); err != nil {
		h.logger.Printf("写入职种版本响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) CreateJobRole(w http.ResponseWriter, r *http.Request) {
	var req types.CreateJobRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobRole(r.Context(), tenantID, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job role created successfully", requestID); err != nil {
		h.logger.Printf("写入职务创建响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) UpdateJobRole(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职务代码", nil)
		return
	}

	var req types.UpdateJobRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Status = strings.TrimSpace(req.Status)
	req.EffectiveDate = strings.TrimSpace(req.EffectiveDate)
	if req.Name == "" || req.Status == "" || req.EffectiveDate == "" {
		h.writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "名称、状态与生效日期为必填项", map[string]interface{}{
			"name":          req.Name,
			"status":        req.Status,
			"effectiveDate": req.EffectiveDate,
		})
		return
	}
	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
	}
	if req.JobFamilyCode != nil {
		trimmed := strings.ToUpper(strings.TrimSpace(*req.JobFamilyCode))
		if trimmed == "" {
			req.JobFamilyCode = nil
		} else {
			req.JobFamilyCode = &trimmed
		}
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)
	ifMatch := getIfMatchHeader(r)

	entity, err := h.service.UpdateJobRole(r.Context(), tenantID, code, &req, ifMatch, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, entity, "Job role updated successfully", requestID); err != nil {
		h.logger.Printf("写入职务更新响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) CreateJobRoleVersion(w http.ResponseWriter, r *http.Request) {
	var req types.JobCatalogVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职务代码", nil)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobRoleVersion(r.Context(), tenantID, code, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job role version created successfully", requestID); err != nil {
		h.logger.Printf("写入职务版本响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) CreateJobLevel(w http.ResponseWriter, r *http.Request) {
	var req types.CreateJobLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobLevel(r.Context(), tenantID, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job level created successfully", requestID); err != nil {
		h.logger.Printf("写入职级创建响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) UpdateJobLevel(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职级代码", nil)
		return
	}

	var req types.UpdateJobLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Status = strings.TrimSpace(req.Status)
	req.EffectiveDate = strings.TrimSpace(req.EffectiveDate)
	if req.Name == "" || req.Status == "" || req.EffectiveDate == "" {
		h.writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "名称、状态与生效日期为必填项", map[string]interface{}{
			"name":          req.Name,
			"status":        req.Status,
			"effectiveDate": req.EffectiveDate,
		})
		return
	}
	if req.Description != nil {
		trimmed := strings.TrimSpace(*req.Description)
		if trimmed == "" {
			req.Description = nil
		} else {
			req.Description = &trimmed
		}
	}
	if req.JobRoleCode != nil {
		trimmed := strings.ToUpper(strings.TrimSpace(*req.JobRoleCode))
		if trimmed == "" {
			req.JobRoleCode = nil
		} else {
			req.JobRoleCode = &trimmed
		}
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)
	ifMatch := getIfMatchHeader(r)

	entity, err := h.service.UpdateJobLevel(r.Context(), tenantID, code, &req, ifMatch, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, entity, "Job level updated successfully", requestID); err != nil {
		h.logger.Printf("写入职级更新响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) CreateJobLevelVersion(w http.ResponseWriter, r *http.Request) {
	var req types.JobCatalogVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职级代码", nil)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobLevelVersion(r.Context(), tenantID, code, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job level version created successfully", requestID); err != nil {
		h.logger.Printf("写入职级版本响应失败: %v", err)
	}
}

func (h *JobCatalogHandler) handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, services.ErrJobCatalogParentMissing):
		h.writeError(w, r, http.StatusBadRequest, "JOB_CATALOG_PARENT_MISSING", "上级职位分类不存在", err)
	case errors.Is(err, services.ErrJobCatalogInvalidInput):
		h.writeError(w, r, http.StatusBadRequest, "JOB_CATALOG_INVALID_INPUT", "职位分类输入无效", err)
	case errors.Is(err, services.ErrJobCatalogNotFound):
		h.writeError(w, r, http.StatusNotFound, "JOB_CATALOG_NOT_FOUND", "职位分类不存在", err)
	case errors.Is(err, services.ErrJobCatalogConflict):
		h.writeError(w, r, http.StatusConflict, "JOB_CATALOG_CONFLICT", "职位分类存在冲突的生效日期", err)
	case errors.Is(err, services.ErrJobCatalogPreconditionFailed):
		h.writeError(w, r, http.StatusPreconditionFailed, "PRECONDITION_FAILED", "资源版本已过期，请刷新后重试", err)
	default:
		h.writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误", err)
	}
}

func (h *JobCatalogHandler) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details interface{}) {
	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteError(w, status, code, message, requestID, details); err != nil {
		h.logger.Printf("写入错误响应失败: %v", err)
	}
}
