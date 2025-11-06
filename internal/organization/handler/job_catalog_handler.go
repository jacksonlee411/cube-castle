package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"cube-castle/internal/organization/middleware"
	"cube-castle/internal/organization/service"
	"cube-castle/internal/organization/utils"
	validator "cube-castle/internal/organization/validator"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/go-chi/chi/v5"
)

type JobCatalogHandler struct {
	service *service.JobCatalogService
	logger  pkglogger.Logger
}

func NewJobCatalogHandler(service *service.JobCatalogService, baseLogger pkglogger.Logger) *JobCatalogHandler {
	return &JobCatalogHandler{
		service: service,
		logger: scopedLogger(baseLogger, "jobCatalog", pkglogger.Fields{
			"routeGroup": "/api/v1/job-*",
			"module":     "jobCatalog",
		}),
	}
}

func (h *JobCatalogHandler) requestLogger(r *http.Request, action string) pkglogger.Logger {
	fields := pkglogger.Fields{
		"action":    action,
		"requestId": middleware.GetRequestID(r.Context()),
	}
	if r != nil {
		fields["method"] = r.Method
		fields["path"] = r.URL.Path
	}
	return h.logger.WithFields(fields)
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
	reqLogger := h.requestLogger(r, "CreateJobFamilyGroup")

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobFamilyGroup(r.Context(), tenantID, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job family group created successfully", requestID); err != nil {
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job family group response failed")
	}
}

func (h *JobCatalogHandler) UpdateJobFamilyGroup(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职类代码", nil)
		return
	}
	reqLogger := h.requestLogger(r, "UpdateJobFamilyGroup")

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
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job family group update response failed")
	}
}

func (h *JobCatalogHandler) CreateJobFamilyGroupVersion(w http.ResponseWriter, r *http.Request) {
	var req types.JobCatalogVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	reqLogger := h.requestLogger(r, "CreateJobFamilyGroupVersion")

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
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job family group version response failed")
	}
}

func (h *JobCatalogHandler) CreateJobFamily(w http.ResponseWriter, r *http.Request) {
	var req types.CreateJobFamilyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	reqLogger := h.requestLogger(r, "CreateJobFamily")

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobFamily(r.Context(), tenantID, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job family created successfully", requestID); err != nil {
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job family response failed")
	}
}

func (h *JobCatalogHandler) UpdateJobFamily(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职种代码", nil)
		return
	}
	reqLogger := h.requestLogger(r, "UpdateJobFamily")

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
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job family update response failed")
	}
}

func (h *JobCatalogHandler) CreateJobFamilyVersion(w http.ResponseWriter, r *http.Request) {
	var req types.JobCatalogVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	reqLogger := h.requestLogger(r, "CreateJobFamilyVersion")

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
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job family version response failed")
	}
}

func (h *JobCatalogHandler) CreateJobRole(w http.ResponseWriter, r *http.Request) {
	var req types.CreateJobRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	reqLogger := h.requestLogger(r, "CreateJobRole")

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	entity, err := h.service.CreateJobRole(r.Context(), tenantID, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, entity, "Job role created successfully", requestID); err != nil {
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job role response failed")
	}
}

func (h *JobCatalogHandler) UpdateJobRole(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职务代码", nil)
		return
	}
	reqLogger := h.requestLogger(r, "UpdateJobRole")

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
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job role update response failed")
	}
}

func (h *JobCatalogHandler) CreateJobRoleVersion(w http.ResponseWriter, r *http.Request) {
	var req types.JobCatalogVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	reqLogger := h.requestLogger(r, "CreateJobRoleVersion")

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
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job role version response failed")
	}
}

func (h *JobCatalogHandler) CreateJobLevel(w http.ResponseWriter, r *http.Request) {
	var req types.CreateJobLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	reqLogger := h.requestLogger(r, "CreateJobLevel")

	// Validate required fields
	if err := validateCreateJobLevelRequest(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
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
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job level response failed")
	}
}

func (h *JobCatalogHandler) UpdateJobLevel(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职级代码", nil)
		return
	}
	reqLogger := h.requestLogger(r, "UpdateJobLevel")

	var req types.UpdateJobLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// Validate required fields
	req.Name = strings.TrimSpace(req.Name)
	req.Status = strings.TrimSpace(req.Status)
	req.EffectiveDate = strings.TrimSpace(req.EffectiveDate)
	if err := validateUpdateJobLevelRequest(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
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
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job level update response failed")
	}
}

func (h *JobCatalogHandler) CreateJobLevelVersion(w http.ResponseWriter, r *http.Request) {
	var req types.JobCatalogVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	reqLogger := h.requestLogger(r, "CreateJobLevelVersion")

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职级代码", nil)
		return
	}

	// Validate required fields
	req.Name = strings.TrimSpace(req.Name)
	req.Status = strings.TrimSpace(req.Status)
	req.EffectiveDate = strings.TrimSpace(req.EffectiveDate)
	if err := validateJobCatalogVersionRequest(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
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
		reqLogger.WithFields(pkglogger.Fields{"error": err}).Error("write job level version response failed")
	}
}

func (h *JobCatalogHandler) handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *validator.ValidationFailedError
	if errors.As(err, &validationErr) {
		h.writeValidationFailure(w, r, validationErr.Result())
		return
	}

	switch {
	case errors.Is(err, service.ErrJobCatalogParentMissing):
		h.writeError(w, r, http.StatusBadRequest, "JOB_CATALOG_PARENT_MISSING", "上级职位分类不存在", err)
	case errors.Is(err, service.ErrJobCatalogInvalidInput):
		h.writeError(w, r, http.StatusBadRequest, "JOB_CATALOG_INVALID_INPUT", "职位分类输入无效", err)
	case errors.Is(err, service.ErrJobCatalogNotFound):
		h.writeError(w, r, http.StatusNotFound, "JOB_CATALOG_NOT_FOUND", "职位分类不存在", err)
	case errors.Is(err, service.ErrJobCatalogConflict):
		h.writeError(w, r, http.StatusConflict, "JOB_CATALOG_CONFLICT", "职位分类存在冲突的生效日期", err)
	case errors.Is(err, service.ErrJobCatalogPreconditionFailed):
		h.writeError(w, r, http.StatusPreconditionFailed, "PRECONDITION_FAILED", "资源版本已过期，请刷新后重试", err)
	default:
		h.writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误", err)
	}
}

func (h *JobCatalogHandler) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details interface{}) {
	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteError(w, status, code, message, requestID, details); err != nil {
		h.requestLogger(r, "writeError").WithFields(pkglogger.Fields{"error": err, "status": status, "code": code}).Error("write job catalog error response failed")
	}
}

func (h *JobCatalogHandler) writeValidationFailure(w http.ResponseWriter, r *http.Request, result *validator.ValidationResult) {
	if result == nil {
		h.writeError(w, r, http.StatusBadRequest, "BUSINESS_RULE_VIOLATION", "业务规则校验失败", nil)
		return
	}

	status := http.StatusBadRequest
	ruleCode := "BUSINESS_RULE_VIOLATION"
	message := "业务规则校验失败"
	if len(result.Errors) > 0 {
		first := result.Errors[0]
		if trimmed := strings.TrimSpace(first.Code); trimmed != "" {
			ruleCode = trimmed
		}
		if trimmed := strings.TrimSpace(first.Message); trimmed != "" {
			message = trimmed
		}
		severity := strings.ToUpper(strings.TrimSpace(first.Severity))
		if severity == "" {
			severity = string(validator.SeverityHigh)
		}
		mapped := validator.SeverityToHTTPStatus(severity)
		if mapped >= http.StatusBadRequest {
			status = mapped
		}
	}

	details := map[string]interface{}{
		"validationErrors": result.Errors,
		"warnings":         result.Warnings,
		"chainContext":     result.Context,
		"errorCount":       len(result.Errors),
		"warningCount":     len(result.Warnings),
	}

	requestID := middleware.GetRequestID(r.Context())
	logger := h.requestLogger(r, "writeValidationFailure")
	if err := utils.WriteError(w, status, ruleCode, message, requestID, details); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err, "status": status, "code": ruleCode}).Error("write validation failure response failed")
	}
}

// Validation helpers
func validateCreateJobLevelRequest(req *types.CreateJobLevelRequest) error {
	if strings.TrimSpace(req.Code) == "" {
		return fmt.Errorf("职级代码不能为空")
	}
	if strings.TrimSpace(req.JobRoleCode) == "" {
		return fmt.Errorf("职位角色代码不能为空")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("职级名称不能为空")
	}
	if strings.TrimSpace(req.Status) == "" {
		return fmt.Errorf("职级状态不能为空")
	}
	if strings.TrimSpace(req.LevelRank) == "" {
		return fmt.Errorf("职级排序号不能为空")
	}
	if strings.TrimSpace(req.EffectiveDate) == "" {
		return fmt.Errorf("生效日期不能为空")
	}
	return nil
}

// validateUpdateJobLevelRequest validates UpdateJobLevelRequest required fields
func validateUpdateJobLevelRequest(req *types.UpdateJobLevelRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("职级名称不能为空")
	}
	if strings.TrimSpace(req.Status) == "" {
		return fmt.Errorf("职级状态不能为空")
	}
	if strings.TrimSpace(req.EffectiveDate) == "" {
		return fmt.Errorf("生效日期不能为空")
	}
	return nil
}

// validateJobCatalogVersionRequest validates JobCatalogVersionRequest required fields
func validateJobCatalogVersionRequest(req *types.JobCatalogVersionRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("名称不能为空")
	}
	if strings.TrimSpace(req.Status) == "" {
		return fmt.Errorf("状态不能为空")
	}
	if strings.TrimSpace(req.EffectiveDate) == "" {
		return fmt.Errorf("生效日期不能为空")
	}
	return nil
}
