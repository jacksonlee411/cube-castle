package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/organizationunit"
	"github.com/gaogu/cube-castle/go-app/ent/position"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// OrganizationUnitHandler handles HTTP requests for organization units
type OrganizationUnitHandler struct {
	client *ent.Client
	logger *logging.StructuredLogger
}

// NewOrganizationUnitHandler creates a new organization unit handler
func NewOrganizationUnitHandler(client *ent.Client, logger *logging.StructuredLogger) *OrganizationUnitHandler {
	return &OrganizationUnitHandler{
		client: client,
		logger: logger,
	}
}

// CreateOrganizationUnitRequest represents the request to create an organization unit
type CreateOrganizationUnitRequest struct {
	UnitType     string                 `json:"unit_type" validate:"required,oneof=DEPARTMENT COST_CENTER COMPANY PROJECT_TEAM"`
	Name         string                 `json:"name" validate:"required,min=1,max=100"`
	Description  *string                `json:"description,omitempty"`
	ParentUnitID *uuid.UUID             `json:"parent_unit_id,omitempty"`
	Status       string                 `json:"status" validate:"oneof=ACTIVE INACTIVE PLANNED"`
	Profile      map[string]interface{} `json:"profile,omitempty"`
}

// UpdateOrganizationUnitRequest represents the request to update an organization unit
type UpdateOrganizationUnitRequest struct {
	Name         *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description  *string                `json:"description,omitempty"`
	ParentUnitID *uuid.UUID             `json:"parent_unit_id,omitempty"`
	Status       *string                `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE PLANNED"`
	Profile      map[string]interface{} `json:"profile,omitempty"`
}

// OrganizationUnitResponse represents the response format for organization unit data
type OrganizationUnitResponse struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	UnitType     string                 `json:"unit_type"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description"`
	ParentUnitID *uuid.UUID             `json:"parent_unit_id"`
	Status       string                 `json:"status"`
	Profile      map[string]interface{} `json:"profile"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// CreateOrganizationUnit handles POST /api/v1/organization-units
func (h *OrganizationUnitHandler) CreateOrganizationUnit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context (set by middleware)
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("create_org_unit", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreateOrganizationUnitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("create_org_unit", "Invalid JSON payload", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.UnitType == "" || req.Name == "" {
			http.Error(w, "unit_type and name are required", http.StatusBadRequest)
			return
		}

		// Validate profile based on unit type
		var profileJSON map[string]interface{}
		if req.Profile != nil {
			profile, err := types.ProfileFactory(req.UnitType, json.RawMessage(`{}`))
			if err != nil {
				h.logger.LogError("create_org_unit", "Invalid unit type", err, map[string]interface{}{
					"unit_type": req.UnitType,
					"tenant_id": tenantID,
				})
				http.Error(w, "Invalid unit type", http.StatusBadRequest)
				return
			}

			// Convert profile map to JSON and validate
			profileData, _ := json.Marshal(req.Profile)
			profile, err = types.ProfileFactory(req.UnitType, profileData)
			if err != nil {
				h.logger.LogError("create_org_unit", "Invalid profile data", err, map[string]interface{}{
					"unit_type": req.UnitType,
					"tenant_id": tenantID,
				})
				http.Error(w, "Invalid profile data for unit type", http.StatusBadRequest)
				return
			}

			if err := profile.Validate(); err != nil {
				h.logger.LogError("create_org_unit", "Profile validation failed", err, map[string]interface{}{
					"unit_type": req.UnitType,
					"tenant_id": tenantID,
				})
				http.Error(w, "Profile validation failed: "+err.Error(), http.StatusBadRequest)
				return
			}

			profileJSON = req.Profile
		}

		// Set default status if not provided
		status := req.Status
		if status == "" {
			status = "ACTIVE"
		}

		// Create the organization unit
		builder := h.client.OrganizationUnit.Create().
			SetTenantID(tenantID).
			SetUnitType(organizationunit.UnitType(req.UnitType)).
			SetName(req.Name).
			SetStatus(organizationunit.Status(status))

		if req.Description != nil {
			builder = builder.SetDescription(*req.Description)
		}

		if req.ParentUnitID != nil {
			builder = builder.SetParentUnitID(*req.ParentUnitID)
		}

		if profileJSON != nil {
			builder = builder.SetProfile(profileJSON)
		}

		orgUnit, err := builder.Save(ctx)
		if err != nil {
			// Check if it's a validation error
			if strings.Contains(err.Error(), "invalid enum value") || strings.Contains(err.Error(), "validator failed") {
				h.logger.LogError("create_org_unit", "Invalid field value", err, map[string]interface{}{
					"unit_type": req.UnitType,
					"tenant_id": tenantID,
				})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			h.logger.LogError("create_org_unit", "Failed to create organization unit", err, map[string]interface{}{
				"unit_type": req.UnitType,
				"name":      req.Name,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to create organization unit", http.StatusInternalServerError)
			return
		}

		// Convert to response format
		response := h.convertToResponse(orgUnit)

		h.logger.Info("Organization unit created successfully",
			"org_unit_id", orgUnit.ID,
			"unit_type", orgUnit.UnitType,
			"name", orgUnit.Name,
			"tenant_id", tenantID,
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetOrganizationUnit handles GET /api/v1/organization-units/{id}
func (h *OrganizationUnitHandler) GetOrganizationUnit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("get_org_unit", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get ID from URL path
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		// Fetch the organization unit
		orgUnit, err := h.client.OrganizationUnit.Query().
			Where(
				organizationunit.IDEQ(id),
				organizationunit.TenantIDEQ(tenantID),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "Organization unit not found", http.StatusNotFound)
				return
			}
			h.logger.LogError("get_org_unit", "Failed to fetch organization unit", err, map[string]interface{}{
				"org_unit_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to fetch organization unit", http.StatusInternalServerError)
			return
		}

		response := h.convertToResponse(orgUnit)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// ListOrganizationUnits handles GET /api/v1/organization-units
func (h *OrganizationUnitHandler) ListOrganizationUnits() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("list_org_units", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse query parameters
		query := h.client.OrganizationUnit.Query().Where(organizationunit.TenantIDEQ(tenantID))

		// Filter by unit type if provided
		if unitType := r.URL.Query().Get("unit_type"); unitType != "" {
			query = query.Where(organizationunit.UnitTypeEQ(organizationunit.UnitType(unitType)))
		}

		// Filter by status if provided
		if status := r.URL.Query().Get("status"); status != "" {
			query = query.Where(organizationunit.StatusEQ(organizationunit.Status(status)))
		}

		// Filter by parent unit ID if provided
		if parentIDStr := r.URL.Query().Get("parent_unit_id"); parentIDStr != "" {
			if parentID, err := uuid.Parse(parentIDStr); err == nil {
				query = query.Where(organizationunit.ParentUnitIDEQ(parentID))
			}
		}

		// Pagination
		limit := 50 // default limit
		offset := 0

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
				limit = l
			}
		}

		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		// Execute query
		orgUnits, err := query.
			Limit(limit).
			Offset(offset).
			Order(organizationunit.ByCreatedAt()).
			All(ctx)

		if err != nil {
			h.logger.LogError("list_org_units", "Failed to fetch organization units", err, map[string]interface{}{
				"tenant_id": tenantID,
				"limit":     limit,
				"offset":    offset,
			})
			http.Error(w, "Failed to fetch organization units", http.StatusInternalServerError)
			return
		}

		// Convert to response format
		responses := make([]OrganizationUnitResponse, len(orgUnits))
		for i, orgUnit := range orgUnits {
			responses[i] = h.convertToResponse(orgUnit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   responses,
			"limit":  limit,
			"offset": offset,
			"total":  len(responses),
		})
	}
}

// UpdateOrganizationUnit handles PUT /api/v1/organization-units/{id}
func (h *OrganizationUnitHandler) UpdateOrganizationUnit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("update_org_unit", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get ID from URL path
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		var req UpdateOrganizationUnitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("update_org_unit", "Invalid JSON payload", err, map[string]interface{}{
				"org_unit_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Fetch existing organization unit to get current unit_type for profile validation
		existingUnit, err := h.client.OrganizationUnit.Query().
			Where(
				organizationunit.IDEQ(id),
				organizationunit.TenantIDEQ(tenantID),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "Organization unit not found", http.StatusNotFound)
				return
			}
			h.logger.LogError("update_org_unit", "Failed to fetch existing organization unit", err, map[string]interface{}{
				"org_unit_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to fetch organization unit", http.StatusInternalServerError)
			return
		}

		// Build update query
		updateBuilder := h.client.OrganizationUnit.UpdateOneID(id)

		if req.Name != nil {
			updateBuilder = updateBuilder.SetName(*req.Name)
		}

		if req.Description != nil {
			updateBuilder = updateBuilder.SetDescription(*req.Description)
		}

		if req.ParentUnitID != nil {
			updateBuilder = updateBuilder.SetParentUnitID(*req.ParentUnitID)
		}

		if req.Status != nil {
			updateBuilder = updateBuilder.SetStatus(organizationunit.Status(*req.Status))
		}

		if req.Profile != nil {
			// Validate profile based on existing unit type
			profileData, _ := json.Marshal(req.Profile)
			profile, err := types.ProfileFactory(existingUnit.UnitType.String(), profileData)
			if err != nil {
				h.logger.LogError("update_org_unit", "Invalid profile data", err, map[string]interface{}{
					"unit_type":   existingUnit.UnitType,
					"org_unit_id": id,
					"tenant_id":   tenantID,
				})
				http.Error(w, "Invalid profile data for unit type", http.StatusBadRequest)
				return
			}

			if err := profile.Validate(); err != nil {
				h.logger.LogError("update_org_unit", "Profile validation failed", err, map[string]interface{}{
					"unit_type":   existingUnit.UnitType,
					"org_unit_id": id,
					"tenant_id":   tenantID,
				})
				http.Error(w, "Profile validation failed: "+err.Error(), http.StatusBadRequest)
				return
			}

			updateBuilder = updateBuilder.SetProfile(req.Profile)
		}

		// Execute update
		orgUnit, err := updateBuilder.Save(ctx)
		if err != nil {
			h.logger.LogError("update_org_unit", "Failed to update organization unit", err, map[string]interface{}{
				"org_unit_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to update organization unit", http.StatusInternalServerError)
			return
		}

		response := h.convertToResponse(orgUnit)

		h.logger.Info("Organization unit updated successfully",
			"org_unit_id", orgUnit.ID,
			"unit_type", orgUnit.UnitType,
			"name", orgUnit.Name,
			"tenant_id", tenantID,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// DeleteOrganizationUnit handles DELETE /api/v1/organization-units/{id}
func (h *OrganizationUnitHandler) DeleteOrganizationUnit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("delete_org_unit", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get ID from URL path
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		// Check if organization unit exists and belongs to tenant
		exists, err := h.client.OrganizationUnit.Query().
			Where(
				organizationunit.IDEQ(id),
				organizationunit.TenantIDEQ(tenantID),
			).
			Exist(ctx)

		if err != nil {
			h.logger.LogError("delete_org_unit", "Failed to check organization unit existence", err, map[string]interface{}{
				"org_unit_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to check organization unit", http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "Organization unit not found", http.StatusNotFound)
			return
		}

		// Check for child units
		childCount, err := h.client.OrganizationUnit.Query().
			Where(
				organizationunit.ParentUnitIDEQ(id),
				organizationunit.TenantIDEQ(tenantID),
			).
			Count(ctx)

		if err != nil {
			h.logger.LogError("delete_org_unit", "Failed to check child units", err, map[string]interface{}{
				"org_unit_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to check child units", http.StatusInternalServerError)
			return
		}

		if childCount > 0 {
			http.Error(w, "Cannot delete organization unit with child units", http.StatusConflict)
			return
		}

		// Check for associated positions
		positionCount, err := h.client.Position.Query().
			Where(
				position.DepartmentIDEQ(id),
				position.TenantIDEQ(tenantID),
			).
			Count(ctx)

		if err != nil {
			h.logger.LogError("delete_org_unit", "Failed to check associated positions", err, map[string]interface{}{
				"org_unit_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to check associated positions", http.StatusInternalServerError)
			return
		}

		if positionCount > 0 {
			http.Error(w, "Cannot delete organization unit with associated positions", http.StatusConflict)
			return
		}

		// Delete the organization unit
		err = h.client.OrganizationUnit.DeleteOneID(id).Exec(ctx)
		if err != nil {
			h.logger.LogError("delete_org_unit", "Failed to delete organization unit", err, map[string]interface{}{
				"org_unit_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to delete organization unit", http.StatusInternalServerError)
			return
		}

		h.logger.Info("Organization unit deleted successfully",
			"org_unit_id", id,
			"tenant_id", tenantID,
		)

		w.WriteHeader(http.StatusNoContent)
	}
}

// convertToResponse converts an ent.OrganizationUnit to OrganizationUnitResponse
func (h *OrganizationUnitHandler) convertToResponse(orgUnit *ent.OrganizationUnit) OrganizationUnitResponse {
	return OrganizationUnitResponse{
		ID:           orgUnit.ID,
		TenantID:     orgUnit.TenantID,
		UnitType:     orgUnit.UnitType.String(),
		Name:         orgUnit.Name,
		Description:  orgUnit.Description,
		ParentUnitID: orgUnit.ParentUnitID,
		Status:       orgUnit.Status.String(),        // Now using correct Status field
		Profile:      orgUnit.Profile,               // Already map[string]interface{}
		CreatedAt:    orgUnit.CreatedAt,
		UpdatedAt:    orgUnit.UpdatedAt,
	}
}
