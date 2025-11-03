package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"cube-castle/cmd/hrms-server/command/internal/middleware"
	"cube-castle/cmd/hrms-server/command/internal/services"
	"cube-castle/cmd/hrms-server/command/internal/types"
	"cube-castle/cmd/hrms-server/command/internal/utils"
)

type PositionService interface {
	CreatePosition(ctx context.Context, tenantID uuid.UUID, req *types.PositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	ReplacePosition(ctx context.Context, tenantID uuid.UUID, code string, ifMatch *string, req *types.PositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	CreatePositionVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionVersionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	FillPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.FillPositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	VacatePosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.VacatePositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	TransferPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.TransferPositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	ApplyEvent(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionEventRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	ListAssignments(ctx context.Context, tenantID uuid.UUID, code string, opts types.AssignmentListOptions) ([]types.PositionAssignmentResponse, int, error)
	CreateAssignmentRecord(ctx context.Context, tenantID uuid.UUID, code string, req *types.CreateAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error)
	UpdateAssignmentRecord(ctx context.Context, tenantID uuid.UUID, code string, assignmentID uuid.UUID, req *types.UpdateAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error)
	CloseAssignmentRecord(ctx context.Context, tenantID uuid.UUID, code string, assignmentID uuid.UUID, req *types.CloseAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error)
}

type PositionHandler struct {
	service PositionService
	logger  *log.Logger
}

func NewPositionHandler(service PositionService, logger *log.Logger) *PositionHandler {
	return &PositionHandler{service: service, logger: logger}
}

func (h *PositionHandler) SetupRoutes(r chi.Router) {
	r.Route("/api/v1/positions", func(r chi.Router) {
		r.Post("/", h.CreatePosition)
		r.Put("/{code}", h.ReplacePosition)
		r.Post("/{code}/versions", h.CreatePositionVersion)
		r.Post("/{code}/events", h.ApplyPositionEvent)
		r.Post("/{code}/fill", h.FillPosition)
		r.Post("/{code}/vacate", h.VacatePosition)
		r.Post("/{code}/transfer", h.TransferPosition)
		r.Route("/{code}/assignments", func(r chi.Router) {
			r.Get("/", h.ListAssignments)
			r.Post("/", h.CreateAssignment)
			r.Patch("/{assignmentId}", h.UpdateAssignment)
			r.Post("/{assignmentId}/close", h.CloseAssignment)
		})
	})
}

func (h *PositionHandler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	var req types.PositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	response, err := h.service.CreatePosition(r.Context(), tenantID, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, response, "Position created successfully", requestID); err != nil {
		h.logger.Printf("写入创建职位响应失败: %v", err)
	}
}

func (h *PositionHandler) ReplacePosition(w http.ResponseWriter, r *http.Request) {
	var req types.PositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	ifMatch := getIfMatchHeader(r)
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职位代码", nil)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	response, err := h.service.ReplacePosition(r.Context(), tenantID, code, ifMatch, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, response, "Position updated successfully", requestID); err != nil {
		h.logger.Printf("写入职位更新响应失败: %v", err)
	}
}

func (h *PositionHandler) CreatePositionVersion(w http.ResponseWriter, r *http.Request) {
	var req types.PositionVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职位代码", nil)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	response, err := h.service.CreatePositionVersion(r.Context(), tenantID, code, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, response, "Position version created successfully", requestID); err != nil {
		h.logger.Printf("写入职位版本响应失败: %v", err)
	}
}

func (h *PositionHandler) FillPosition(w http.ResponseWriter, r *http.Request) {
	var req types.FillPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职位代码", nil)
		return
	}
	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	response, err := h.service.FillPosition(r.Context(), tenantID, code, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, response, "Position filled successfully", requestID); err != nil {
		h.logger.Printf("写入填充职位响应失败: %v", err)
	}
}

func (h *PositionHandler) VacatePosition(w http.ResponseWriter, r *http.Request) {
	var req types.VacatePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职位代码", nil)
		return
	}
	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	response, err := h.service.VacatePosition(r.Context(), tenantID, code, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, response, "Position vacated successfully", requestID); err != nil {
		h.logger.Printf("写入清空职位响应失败: %v", err)
	}
}

func (h *PositionHandler) TransferPosition(w http.ResponseWriter, r *http.Request) {
	var req types.TransferPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职位代码", nil)
		return
	}
	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	response, err := h.service.TransferPosition(r.Context(), tenantID, code, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, response, "Position transferred successfully", requestID); err != nil {
		h.logger.Printf("写入职位转移响应失败: %v", err)
	}
}

func (h *PositionHandler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职位代码", nil)
		return
	}

	query := r.URL.Query()
	assignmentTypes := make([]string, 0)
	if values, ok := query["assignmentTypes"]; ok {
		assignmentTypes = append(assignmentTypes, values...)
	}
	if values, ok := query["assignmentTypes[]"]; ok {
		assignmentTypes = append(assignmentTypes, values...)
	}

	assignmentStatus := query.Get("assignmentStatus")
	asOfDateStr := query.Get("asOfDate")
	includeHistorical := true
	if raw := query.Get("includeHistorical"); raw != "" {
		val, err := strconv.ParseBool(raw)
		if err != nil {
			h.writeError(w, r, http.StatusBadRequest, "INVALID_PARAMETER", "includeHistorical 参数无效", err)
			return
		}
		includeHistorical = val
	}
	includeActingOnly := false
	if raw := query.Get("includeActingOnly"); raw != "" {
		val, err := strconv.ParseBool(raw)
		if err != nil {
			h.writeError(w, r, http.StatusBadRequest, "INVALID_PARAMETER", "includeActingOnly 参数无效", err)
			return
		}
		includeActingOnly = val
	}

	page := 1
	if raw := query.Get("page"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 25
	if raw := query.Get("pageSize"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	var asOfDate *time.Time
	if strings.TrimSpace(asOfDateStr) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(asOfDateStr))
		if err != nil {
			h.writeError(w, r, http.StatusBadRequest, "INVALID_PARAMETER", "asOfDate 参数格式应为 YYYY-MM-DD", err)
			return
		}
		asOfDate = &parsed
	}

	filter := types.AssignmentListFilter{
		AssignmentTypes:   assignmentTypes,
		IncludeHistorical: includeHistorical,
		IncludeActingOnly: includeActingOnly,
	}
	if strings.TrimSpace(assignmentStatus) != "" {
		status := strings.ToUpper(strings.TrimSpace(assignmentStatus))
		filter.AssignmentStatus = &status
	}
	if asOfDate != nil {
		filter.AsOfDate = asOfDate
	}

	opts := types.AssignmentListOptions{
		Filter:   filter,
		Page:     page,
		PageSize: pageSize,
	}

	tenantID := getTenantIDFromRequest(r)
	assignments, total, err := h.service.ListAssignments(r.Context(), tenantID, code, opts)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	meta := types.PaginationMeta{
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		HasPrevious: page > 1,
		HasNext:     page*pageSize < total,
	}

	response := types.PositionAssignmentListResponse{
		Data:       assignments,
		Pagination: meta,
		TotalCount: total,
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, response, "Assignments retrieved successfully", requestID); err != nil {
		h.logger.Printf("写入任职列表响应失败: %v", err)
	}
}

func (h *PositionHandler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	var req types.CreateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职位代码", nil)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	assignment, err := h.service.CreateAssignmentRecord(r.Context(), tenantID, code, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteCreated(w, assignment, "Assignment created successfully", requestID); err != nil {
		h.logger.Printf("写入创建任职响应失败: %v", err)
	}
}

func (h *PositionHandler) UpdateAssignment(w http.ResponseWriter, r *http.Request) {
	var req types.UpdateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职位代码", nil)
		return
	}

	assignmentIDStr := strings.TrimSpace(chi.URLParam(r, "assignmentId"))
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_ASSIGNMENT_ID", "任职ID格式必须为UUID", err)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	assignment, err := h.service.UpdateAssignmentRecord(r.Context(), tenantID, code, assignmentID, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, assignment, "Assignment updated successfully", requestID); err != nil {
		h.logger.Printf("写入任职更新响应失败: %v", err)
	}
}

func (h *PositionHandler) CloseAssignment(w http.ResponseWriter, r *http.Request) {
	var req types.CloseAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职位代码", nil)
		return
	}

	assignmentIDStr := strings.TrimSpace(chi.URLParam(r, "assignmentId"))
	assignmentID, err := uuid.Parse(assignmentIDStr)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_ASSIGNMENT_ID", "任职ID格式必须为UUID", err)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	assignment, err := h.service.CloseAssignmentRecord(r.Context(), tenantID, code, assignmentID, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, assignment, "Assignment closed successfully", requestID); err != nil {
		h.logger.Printf("写入任职结束响应失败: %v", err)
	}
}

func (h *PositionHandler) ApplyPositionEvent(w http.ResponseWriter, r *http.Request) {
	var req types.PositionEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	code := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "code")))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少职位代码", nil)
		return
	}

	tenantID := getTenantIDFromRequest(r)
	operator := getOperatorFromRequest(r)

	response, err := h.service.ApplyEvent(r.Context(), tenantID, code, &req, operator)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteSuccess(w, response, "Position event applied successfully", requestID); err != nil {
		h.logger.Printf("写入职位事件响应失败: %v", err)
	}
}

func (h *PositionHandler) handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, services.ErrOrganizationNotFound):
		h.writeError(w, r, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "组织不存在", err)
	case errors.Is(err, services.ErrPositionNotFound):
		h.writeError(w, r, http.StatusNotFound, "POSITION_NOT_FOUND", "职位不存在", err)
	case errors.Is(err, services.ErrJobCatalogNotFound):
		h.writeError(w, r, http.StatusBadRequest, "JOB_CATALOG_NOT_FOUND", "职位分类引用不存在", err)
	case errors.Is(err, services.ErrJobCatalogMismatch):
		h.writeError(w, r, http.StatusConflict, "JOB_CATALOG_MISMATCH", "职位分类层级不一致", err)
	case errors.Is(err, services.ErrVersionConflict):
		h.writeError(w, r, http.StatusPreconditionFailed, "PRECONDITION_FAILED", "资源已发生变更，请刷新后重试", err)
	case errors.Is(err, services.ErrInvalidHeadcount):
		h.writeError(w, r, http.StatusBadRequest, "INVALID_HEADCOUNT", "编制或占用人数无效", err)
	case errors.Is(err, services.ErrInvalidTransition):
		h.writeError(w, r, http.StatusBadRequest, "INVALID_TRANSITION", "不支持的职位状态变更", err)
	case errors.Is(err, services.ErrAssignmentNotFound):
		h.writeError(w, r, http.StatusNotFound, "ASSIGNMENT_NOT_FOUND", "任职记录不存在", err)
	case errors.Is(err, services.ErrInvalidAssignmentState):
		h.writeError(w, r, http.StatusConflict, "INVALID_ASSIGNMENT_STATE", "当前任职状态不允许此操作", err)
	case errors.Is(err, services.ErrPositionVersionExists):
		h.writeError(w, r, http.StatusConflict, "POSITION_VERSION_EXISTS", "该生效日期的职位版本已存在", err)
	default:
		h.writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "服务器内部错误", err)
	}
}

func (h *PositionHandler) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details interface{}) {
	requestID := middleware.GetRequestID(r.Context())
	if err := utils.WriteError(w, status, code, message, requestID, details); err != nil {
		h.logger.Printf("写入错误响应失败: %v", err)
	}
}
