package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/cube-castle/cmd/organization-command-server/internal/application/commands"
	"github.com/cube-castle/cmd/organization-command-server/internal/application/handlers"
	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/logging"
	"github.com/cube-castle/cmd/organization-command-server/internal/presentation/http/middleware"
)

// CreateOrganizationRequest represents the HTTP request for creating an organization
type CreateOrganizationRequest struct {
	Code        *string `json:"code,omitempty"`
	Name        string  `json:"name"`
	ParentCode  *string `json:"parent_code,omitempty"`
	UnitType    string  `json:"unit_type"`
	Description *string `json:"description,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

// UpdateOrganizationRequest represents the HTTP request for updating an organization
type UpdateOrganizationRequest struct {
	Name        *string `json:"name,omitempty"`
	Status      *string `json:"status,omitempty"`
	Description *string `json:"description,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

// OrganizationHTTPHandler handles HTTP requests for organization operations
type OrganizationHTTPHandler struct {
	handler      *handlers.OrganizationHandler
	errorHandler *middleware.ErrorHandler
	logger       logging.Logger
	defaultTenantID uuid.UUID
}

// NewOrganizationHTTPHandler creates a new organization HTTP handler
func NewOrganizationHTTPHandler(
	handler *handlers.OrganizationHandler,
	errorHandler *middleware.ErrorHandler,
	logger logging.Logger,
	defaultTenantID uuid.UUID,
) *OrganizationHTTPHandler {
	return &OrganizationHTTPHandler{
		handler:         handler,
		errorHandler:    errorHandler,
		logger:          logger,
		defaultTenantID: defaultTenantID,
	}
}

// CreateOrganization handles POST /api/v1/organization-units
func (h *OrganizationHTTPHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract tenant ID from header or use default
	tenantID := h.extractTenantID(r)
	
	// Parse request body
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.WriteErrorResponse(w, r, 
			middleware.DomainError{
				Code:    "REQ_001",
				Message: "invalid request body",
				Details: err.Error(),
			}, http.StatusBadRequest)
		return
	}
	
	// Create command
	cmd := commands.CreateOrganizationCommand{
		CommandID:     uuid.New(),
		TenantID:      tenantID,
		RequestedCode: req.Code,
		Name:          req.Name,
		ParentCode:    req.ParentCode,
		UnitType:      req.UnitType,
		Description:   req.Description,
		SortOrder:     req.SortOrder,
		RequestedBy:   h.extractUserID(r),
	}
	
	// Execute command
	result, err := h.handler.HandleCreateOrganization(ctx, cmd)
	if err != nil {
		h.errorHandler.WriteErrorResponse(w, r, err, http.StatusInternalServerError)
		return
	}
	
	// Return success response
	h.writeJSONResponse(w, http.StatusCreated, result)
}

// UpdateOrganization handles PUT /api/v1/organization-units/{code}
func (h *OrganizationHTTPHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract tenant ID and organization code
	tenantID := h.extractTenantID(r)
	code := chi.URLParam(r, "code")
	
	if code == "" {
		h.errorHandler.WriteErrorResponse(w, r,
			middleware.DomainError{
				Code:    "REQ_002",
				Message: "organization code is required",
				Details: "code parameter is missing from URL path",
			}, http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.WriteErrorResponse(w, r,
			middleware.DomainError{
				Code:    "REQ_001", 
				Message: "invalid request body",
				Details: err.Error(),
			}, http.StatusBadRequest)
		return
	}
	
	// Create command
	cmd := commands.UpdateOrganizationCommand{
		CommandID:   uuid.New(),
		TenantID:    tenantID,
		Code:        code,
		Name:        req.Name,
		Status:      req.Status,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		RequestedBy: h.extractUserID(r),
	}
	
	// Execute command
	result, err := h.handler.HandleUpdateOrganization(ctx, cmd)
	if err != nil {
		h.errorHandler.WriteErrorResponse(w, r, err, http.StatusInternalServerError)
		return
	}
	
	// Return success response
	h.writeJSONResponse(w, http.StatusOK, result)
}

// DeleteOrganization handles DELETE /api/v1/organization-units/{code}
func (h *OrganizationHTTPHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract tenant ID and organization code
	tenantID := h.extractTenantID(r)
	code := chi.URLParam(r, "code")
	
	if code == "" {
		h.errorHandler.WriteErrorResponse(w, r,
			middleware.DomainError{
				Code:    "REQ_002",
				Message: "organization code is required", 
				Details: "code parameter is missing from URL path",
			}, http.StatusBadRequest)
		return
	}
	
	// Create command
	cmd := commands.DeleteOrganizationCommand{
		CommandID:   uuid.New(),
		TenantID:    tenantID,
		Code:        code,
		RequestedBy: h.extractUserID(r),
	}
	
	// Execute command
	result, err := h.handler.HandleDeleteOrganization(ctx, cmd)
	if err != nil {
		h.errorHandler.WriteErrorResponse(w, r, err, http.StatusInternalServerError)
		return
	}
	
	// Return success response
	h.writeJSONResponse(w, http.StatusOK, result)
}

// Helper methods

// extractTenantID extracts tenant ID from request headers
func (h *OrganizationHTTPHandler) extractTenantID(r *http.Request) uuid.UUID {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		return h.defaultTenantID
	}
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.logger.Warn("invalid tenant ID in header, using default",
			"provided_tenant_id", tenantIDStr,
			"default_tenant_id", h.defaultTenantID,
		)
		return h.defaultTenantID
	}
	
	return tenantID
}

// extractUserID extracts user ID from request (placeholder implementation)
func (h *OrganizationHTTPHandler) extractUserID(r *http.Request) uuid.UUID {
	// In a real implementation, this would extract from JWT token or session
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			return userID
		}
	}
	
	// Return a default/system user ID for now
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

// writeJSONResponse writes a JSON response
func (h *OrganizationHTTPHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", "error", err)
	}
}