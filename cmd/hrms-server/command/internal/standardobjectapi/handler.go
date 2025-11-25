package standardobjectapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"cube-castle/internal/middleware"
	"cube-castle/internal/standardobject"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	clockpkg "cube-castle/pkg/temporal/clock"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler exposes REST endpoints declared in docs/api/openapi.yaml#standard-objects.
type Handler struct {
	service standardobject.ObjectService
	clock   clockpkg.Clock
	logger  pkglogger.Logger
}

// NewHandler wires a Standard Object HTTP handler.
func NewHandler(svc standardobject.ObjectService, clk clockpkg.Clock, logger pkglogger.Logger) *Handler {
	if clk == nil {
		clk = clockpkg.NewSystemClock()
	}
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}
	return &Handler{
		service: svc,
		clock:   clk,
		logger:  logger.WithFields(pkglogger.Fields{"component": "standardobject.api"}),
	}
}

type createRequest struct {
	Kernel  kernelInput  `json:"kernel"`
	Version versionInput `json:"version"`
	Links   []linkInput  `json:"links"`
}

type versionRequest struct {
	TenantCode string       `json:"tenantCode"`
	Version    versionInput `json:"version"`
}

type statusRequest struct {
	TenantCode  string `json:"tenantCode"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
	RequestedBy string `json:"requestedBy"`
}

type kernelInput struct {
	Code               string            `json:"code"`
	DisplayName        string            `json:"displayName"`
	TenantCode         string            `json:"tenantCode"`
	Status             string            `json:"status"`
	Labels             map[string]string `json:"labels"`
	SchemaVersion      string            `json:"schemaVersion"`
	DataClassification string            `json:"dataClassification"`
	RetentionPolicy    string            `json:"retentionPolicy"`
}

type versionInput struct {
	EffectiveDate string         `json:"effectiveDate"`
	EndDate       string         `json:"endDate"`
	Payload       map[string]any `json:"payload"`
	AuditTrail    map[string]any `json:"auditTrail"`
}

type linkInput struct {
	LinkType   string         `json:"linkType"`
	SourceCode string         `json:"sourceCode"`
	TargetCode string         `json:"targetCode"`
	Attributes map[string]any `json:"attributes"`
}

type successResponse struct {
	Success   bool              `json:"success"`
	Data      aggregateResponse `json:"data"`
	RequestID string            `json:"requestId,omitempty"`
}

type errorResponse struct {
	Success   bool     `json:"success"`
	Error     apiError `json:"error"`
	RequestID string   `json:"requestId,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details []any  `json:"details,omitempty"`
}

type aggregateResponse struct {
	Kernel  kernelResponse  `json:"kernel"`
	Version versionResponse `json:"version"`
	Links   []linkResponse  `json:"links,omitempty"`
}

type kernelResponse struct {
	Code               string            `json:"code"`
	DisplayName        string            `json:"displayName"`
	TenantCode         string            `json:"tenantCode"`
	Status             string            `json:"status"`
	Labels             map[string]string `json:"labels,omitempty"`
	SchemaVersion      string            `json:"schemaVersion,omitempty"`
	DataClassification string            `json:"dataClassification,omitempty"`
	RetentionPolicy    string            `json:"retentionPolicy,omitempty"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
}

type versionResponse struct {
	VersionCode   string         `json:"versionCode"`
	EffectiveDate time.Time      `json:"effectiveDate"`
	EndDate       *time.Time     `json:"endDate,omitempty"`
	IsCurrent     bool           `json:"isCurrent"`
	Payload       map[string]any `json:"payload"`
	AuditTrail    map[string]any `json:"auditTrail,omitempty"`
	Checksum      string         `json:"checksum,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

type linkResponse struct {
	LinkType   string         `json:"linkType"`
	SourceCode string         `json:"sourceCode"`
	TargetCode string         `json:"targetCode"`
	TenantCode string         `json:"tenantCode"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// SetupRoutes mounts all REST endpoints under /api/v1/standard-objects.
func (h *Handler) SetupRoutes(r chi.Router) {
	r.Route("/api/v1/standard-objects", func(r chi.Router) {
		r.Post("/{objectType}", h.handleCreate)
		r.Post("/{objectType}/{code}/versions", h.handleAppendVersion)
		r.Patch("/{objectType}/{code}/status", h.handleUpdateStatus)
	})
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	objectType, err := parseObjectType(chi.URLParam(r, "objectType"))
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_OBJECT_TYPE", err.Error())
		return
	}

	var payload createRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求体解析失败", err)
		return
	}
	tenantCode, err := h.resolveTenantCode(r, payload.Kernel.TenantCode)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_TENANT", err.Error())
		return
	}
	if strings.TrimSpace(payload.Kernel.Code) == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "kernel.code 不能为空", nil)
		return
	}

	now := h.clock.Now().UTC()
	version, err := buildTemporalVersion(payload.Version, payload.Kernel.Code, now)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_VERSION", err.Error())
		return
	}

	aggregate := standardobject.ObjectAggregate{
		Kernel: standardobject.ObjectKernel{
			ObjectType:         objectType,
			Code:               payload.Kernel.Code,
			DisplayName:        payload.Kernel.DisplayName,
			TenantCode:         tenantCode,
			Status:             mapLifecycleStatus(payload.Kernel.Status),
			Labels:             payload.Kernel.Labels,
			SchemaVersion:      payload.Kernel.SchemaVersion,
			DataClassification: payload.Kernel.DataClassification,
			RetentionPolicy:    payload.Kernel.RetentionPolicy,
			CreatedBy:          h.actorID(r),
			CreatedAt:          now,
			UpdatedAt:          now,
		},
		Version: version,
		Links:   buildLinks(payload.Links, tenantCode, objectType, version),
	}

	if err := h.service.Upsert(ctx, aggregate); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "STANDARD_OBJECT_ERROR", err.Error())
		return
	}

	resp, err := h.service.Get(ctx, standardobject.ObjectKey{
		ObjectType: objectType,
		Code:       aggregate.Kernel.Code,
		TenantCode: tenantCode,
	})
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "STANDARD_OBJECT_FETCH_ERROR", err.Error())
		return
	}
	h.writeSuccess(w, r, resp)
}

func (h *Handler) handleAppendVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	objectType, err := parseObjectType(chi.URLParam(r, "objectType"))
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_OBJECT_TYPE", err.Error())
		return
	}
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "路径参数 code 不能为空", nil)
		return
	}

	var payload versionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求体解析失败", err)
		return
	}

	tenantCode, err := h.resolveTenantCode(r, payload.TenantCode)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_TENANT", err.Error())
		return
	}

	existing, err := h.service.Get(ctx, standardobject.ObjectKey{
		ObjectType: objectType,
		Code:       code,
		TenantCode: tenantCode,
	})
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "STANDARD_OBJECT_NOT_FOUND", err.Error())
		return
	}

	now := h.clock.Now().UTC()
	version, err := buildTemporalVersion(payload.Version, code, now)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_VERSION", err.Error())
		return
	}

	existing.Version = version
	existing.Kernel.UpdatedAt = now
	if err := h.service.Upsert(ctx, existing); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "STANDARD_OBJECT_ERROR", err.Error())
		return
	}

	latest, err := h.service.Get(ctx, standardobject.ObjectKey{
		ObjectType: objectType,
		Code:       code,
		TenantCode: tenantCode,
	})
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "STANDARD_OBJECT_FETCH_ERROR", err.Error())
		return
	}
	h.writeSuccess(w, r, latest)
}

func (h *Handler) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	objectType, err := parseObjectType(chi.URLParam(r, "objectType"))
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_OBJECT_TYPE", err.Error())
		return
	}
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	if code == "" {
		h.writeError(w, r, http.StatusBadRequest, "MISSING_CODE", "路径参数 code 不能为空", nil)
		return
	}

	var payload statusRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求体解析失败", err)
		return
	}

	tenantCode, err := h.resolveTenantCode(r, payload.TenantCode)
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, "INVALID_TENANT", err.Error())
		return
	}

	aggregate, err := h.service.Get(ctx, standardobject.ObjectKey{
		ObjectType: objectType,
		Code:       code,
		TenantCode: tenantCode,
	})
	if err != nil {
		h.writeError(w, r, http.StatusNotFound, "STANDARD_OBJECT_NOT_FOUND", err.Error())
		return
	}

	aggregate.Kernel.Status = mapLifecycleStatus(payload.Status)
	aggregate.Kernel.UpdatedAt = h.clock.Now().UTC()
	if aggregate.Version.AuditTrail == nil {
		aggregate.Version.AuditTrail = map[string]any{}
	}
	if payload.Reason != "" {
		aggregate.Version.AuditTrail["statusReason"] = payload.Reason
	}
	if payload.RequestedBy != "" {
		aggregate.Version.AuditTrail["updatedBy"] = payload.RequestedBy
	}

	if err := h.service.Upsert(ctx, aggregate); err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "STANDARD_OBJECT_ERROR", err.Error())
		return
	}

	latest, err := h.service.Get(ctx, standardobject.ObjectKey{
		ObjectType: objectType,
		Code:       code,
		TenantCode: tenantCode,
	})
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "STANDARD_OBJECT_FETCH_ERROR", err.Error())
		return
	}
	h.writeSuccess(w, r, latest)
}

func (h *Handler) resolveTenantCode(r *http.Request, fallback string) (string, error) {
	value := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	if value == "" {
		return types.DefaultTenantID.String(), nil
	}
	if _, err := uuid.Parse(value); err != nil {
		return "", errors.New("无效的 Tenant ID")
	}
	return value, nil
}

func (h *Handler) actorID(r *http.Request) string {
	if mock := strings.TrimSpace(r.Header.Get("X-Mock-User")); mock != "" {
		return mock
	}
	if ctxID := r.Context().Value("user_id"); ctxID != nil {
		if val, ok := ctxID.(string); ok && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
	}
	header := strings.TrimSpace(r.Header.Get("X-Actor-ID"))
	if header != "" {
		return header
	}
	return "system"
}

func (h *Handler) writeSuccess(w http.ResponseWriter, r *http.Request, aggregate standardobject.ObjectAggregate) {
	response := successResponse{
		Success:   true,
		Data:      toAggregateResponse(aggregate),
		RequestID: middleware.GetRequestID(r.Context()),
	}
	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, status int, code string, message string, details ...any) {
	payload := errorResponse{
		Success:   false,
		Error:     apiError{Code: code, Message: message, Details: details},
		RequestID: middleware.GetRequestID(r.Context()),
	}
	h.writeJSON(w, status, payload)
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func buildTemporalVersion(input versionInput, code string, now time.Time) (standardobject.TemporalVersion, error) {
	if input.EffectiveDate == "" {
		return standardobject.TemporalVersion{}, errors.New("effectiveDate 不能为空")
	}
	effective, err := time.Parse("2006-01-02", input.EffectiveDate)
	if err != nil {
		return standardobject.TemporalVersion{}, err
	}
	var endDate *time.Time
	if strings.TrimSpace(input.EndDate) != "" {
		parsed, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return standardobject.TemporalVersion{}, err
		}
		endDate = &parsed
	}

	versionCode := standardobject.MakeVersionCode(code, effective, now, uuid.NewString())
	return standardobject.TemporalVersion{
		VersionCode:     versionCode,
		EffectiveDate:   effective,
		EndDate:         endDate,
		IsCurrent:       true,
		Payload:         input.Payload,
		AuditTrail:      input.AuditTrail,
		Checksum:        "",
		CreatedAt:       now,
		UpdatedAt:       now,
		TransactionFrom: now,
		TransactionTo:   nil,
	}, nil
}

func buildLinks(inputs []linkInput, tenantCode string, objectType standardobject.ObjectType, version standardobject.TemporalVersion) []standardobject.Link {
	if len(inputs) == 0 {
		return nil
	}
	result := make([]standardobject.Link, 0, len(inputs))
	for _, input := range inputs {
		if strings.TrimSpace(input.TargetCode) == "" || strings.TrimSpace(input.SourceCode) == "" {
			continue
		}
		result = append(result, standardobject.Link{
			LinkType:        strings.ToUpper(strings.TrimSpace(input.LinkType)),
			SourceCode:      input.SourceCode,
			TargetCode:      input.TargetCode,
			TargetType:      objectType,
			Attributes:      input.Attributes,
			TenantCode:      tenantCode,
			ValidFrom:       version.EffectiveDate,
			ValidTo:         version.EndDate,
			TransactionFrom: version.TransactionFrom,
			TransactionTo:   nil,
			CreatedBy:       "system",
		})
	}
	return result
}

func toAggregateResponse(aggregate standardobject.ObjectAggregate) aggregateResponse {
	resp := aggregateResponse{
		Kernel: kernelResponse{
			Code:               aggregate.Kernel.Code,
			DisplayName:        aggregate.Kernel.DisplayName,
			TenantCode:         aggregate.Kernel.TenantCode,
			Status:             string(aggregate.Kernel.Status),
			Labels:             aggregate.Kernel.Labels,
			SchemaVersion:      aggregate.Kernel.SchemaVersion,
			DataClassification: aggregate.Kernel.DataClassification,
			RetentionPolicy:    aggregate.Kernel.RetentionPolicy,
			CreatedAt:          aggregate.Kernel.CreatedAt,
			UpdatedAt:          aggregate.Kernel.UpdatedAt,
		},
		Version: versionResponse{
			VersionCode:   aggregate.Version.VersionCode,
			EffectiveDate: aggregate.Version.EffectiveDate,
			EndDate:       aggregate.Version.EndDate,
			IsCurrent:     aggregate.Version.IsCurrent,
			Payload:       aggregate.Version.Payload,
			AuditTrail:    aggregate.Version.AuditTrail,
			Checksum:      aggregate.Version.Checksum,
			CreatedAt:     aggregate.Version.CreatedAt,
			UpdatedAt:     aggregate.Version.UpdatedAt,
		},
	}

	if resp.Kernel.Labels == nil {
		resp.Kernel.Labels = map[string]string{}
	}
	if resp.Version.Payload == nil {
		resp.Version.Payload = map[string]any{}
	}
	if resp.Version.AuditTrail == nil {
		resp.Version.AuditTrail = map[string]any{}
	}

	if len(aggregate.Links) > 0 {
		links := make([]linkResponse, 0, len(aggregate.Links))
		for _, link := range aggregate.Links {
			links = append(links, linkResponse{
				LinkType:   link.LinkType,
				SourceCode: link.SourceCode,
				TargetCode: link.TargetCode,
				TenantCode: link.TenantCode,
				Attributes: link.Attributes,
			})
		}
		resp.Links = links
	}
	return resp
}

func parseObjectType(raw string) (standardobject.ObjectType, error) {
	value := strings.ToUpper(strings.TrimSpace(raw))
	switch value {
	case "ORGANIZATION_UNIT", "ORGANIZATION", "ORG":
		return standardobject.ObjectTypeOrganizationUnit, nil
	case "POSITION_ROLE", "POSITION":
		return standardobject.ObjectTypePositionRole, nil
	default:
		return "", errors.New("不支持的 objectType")
	}
}

func mapLifecycleStatus(raw string) standardobject.LifecycleStatus {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "ACTIVE":
		return standardobject.StatusActive
	case "READY", "PLANNED":
		return standardobject.StatusReady
	case "SUSPENDED", "INACTIVE":
		return standardobject.StatusSuspended
	case "RETIRED", "DELETED":
		return standardobject.StatusRetired
	default:
		return standardobject.StatusDraft
	}
}
