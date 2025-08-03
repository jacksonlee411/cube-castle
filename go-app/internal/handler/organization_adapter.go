package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/organizationunit"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// OrganizationAdapter provides API compatibility layer between
// frontend Organization model and backend OrganizationUnit model
type OrganizationAdapter struct {
	unitHandler *OrganizationUnitHandler
	client      *ent.Client
	logger      *logging.StructuredLogger
}

// NewOrganizationAdapter creates a new organization adapter
func NewOrganizationAdapter(client *ent.Client, logger *logging.StructuredLogger) *OrganizationAdapter {
	return &OrganizationAdapter{
		unitHandler: NewOrganizationUnitHandler(client, logger),
		client:      client,
		logger:      logger,
	}
}

// Organization API models - ALIGNED WITH BACKEND OrganizationUnit schema
// Frontend will adapt to backend types: DEPARTMENT, COST_CENTER, COMPANY, PROJECT_TEAM
type OrganizationResponse struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	UnitType     string                 `json:"unit_type"`     // DEPARTMENT, COST_CENTER, COMPANY, PROJECT_TEAM
	Name         string                 `json:"name"`
	Description  *string                `json:"description"`
	ParentUnitID *string                `json:"parent_unit_id"`
	Status       string                 `json:"status"`        // ACTIVE, INACTIVE, PLANNED
	Profile      map[string]interface{} `json:"profile"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
	
	// Computed fields for frontend compatibility
	Level         int                    `json:"level"`         // Calculated from hierarchy
	EmployeeCount int                    `json:"employee_count"` // Calculated from positions
	Children      []OrganizationResponse `json:"children,omitempty"`
}

type OrganizationListResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	Pagination    PaginationResponse     `json:"pagination"`
}

type PaginationResponse struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// Request models - accepting backend enum values directly
type CreateOrganizationRequest struct {
	UnitType     string                 `json:"unit_type" validate:"required,oneof=DEPARTMENT COST_CENTER COMPANY PROJECT_TEAM"`
	Name         string                 `json:"name" validate:"required,min=1,max=100"`
	Description  *string                `json:"description,omitempty"`
	ParentUnitID *string                `json:"parent_unit_id,omitempty"`
	Status       *string                `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE PLANNED"`
	Profile      map[string]interface{} `json:"profile,omitempty"`
}

type UpdateOrganizationRequest struct {
	Name         *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description  *string                `json:"description,omitempty"`
	ParentUnitID *string                `json:"parent_unit_id,omitempty"`
	Status       *string                `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE PLANNED"`
	Profile      map[string]interface{} `json:"profile,omitempty"`
}

// convertToOrganizationResponse converts backend OrganizationUnit to API response
func (a *OrganizationAdapter) convertToOrganizationResponse(unit *ent.OrganizationUnit) OrganizationResponse {
	// Calculate level based on parent hierarchy (TODO: implement proper recursive calculation)
	level := 0
	if unit.ParentUnitID != nil {
		level = 1 // Simplified for now
	}

	// Convert UUIDs to strings
	id := unit.ID.String()
	tenantId := unit.TenantID.String()
	
	var parentUnitId *string
	if unit.ParentUnitID != nil {
		pid := unit.ParentUnitID.String()
		parentUnitId = &pid
	}
	
	return OrganizationResponse{
		ID:            id,
		TenantID:      tenantId,
		UnitType:      unit.UnitType.String(),      // Keep backend enum: DEPARTMENT, COST_CENTER, etc.
		Name:          unit.Name,
		Description:   unit.Description,
		ParentUnitID:  parentUnitId,
		Status:        unit.Status.String(),        // Now using correct Status field
		Profile:       unit.Profile,                // Keep as-is
		CreatedAt:     unit.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
		UpdatedAt:     unit.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
		Level:         level,                       // Computed field
		EmployeeCount: 0,                          // TODO: Calculate from positions
		Children:      []OrganizationResponse{},   // Will be populated by caller if needed
	}
}

// convertToCreateUnitRequest converts API request to backend request (minimal conversion needed)
func (a *OrganizationAdapter) convertToCreateUnitRequest(req CreateOrganizationRequest) CreateOrganizationUnitRequest {
	status := "ACTIVE"
	if req.Status != nil {
		status = *req.Status
	}

	var parentUnitID *uuid.UUID
	if req.ParentUnitID != nil {
		if pid, err := uuid.Parse(*req.ParentUnitID); err == nil {
			parentUnitID = &pid
		}
	}

	return CreateOrganizationUnitRequest{
		UnitType:     req.UnitType,      // Direct mapping - no conversion needed
		Name:         req.Name,
		Description:  req.Description,
		ParentUnitID: parentUnitID,
		Status:       status,
		Profile:      req.Profile,       // Direct mapping - no conversion needed
	}
}

// GetOrganizations handles GET /api/v1/corehr/organizations
func (a *OrganizationAdapter) GetOrganizations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			// For development, use default tenant
			tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
			a.logger.Info("No tenant ID in context, using default", 
				"default_tenant", tenantID,
			)
		}

		// Query organization units
		query := a.client.OrganizationUnit.Query().Where(organizationunit.TenantIDEQ(tenantID))

		// Apply filters
		if parentID := r.URL.Query().Get("parent_unit_id"); parentID != "" {
			if pid, err := uuid.Parse(parentID); err == nil {
				query = query.Where(organizationunit.ParentUnitIDEQ(pid))
			}
		}
		
		if unitType := r.URL.Query().Get("unit_type"); unitType != "" {
			query = query.Where(organizationunit.UnitTypeEQ(organizationunit.UnitType(unitType)))
		}
		
		// Filter by status if provided
		if status := r.URL.Query().Get("status"); status != "" {
			query = query.Where(organizationunit.StatusEQ(organizationunit.Status(status)))
		}

		// Execute query
		units, err := query.Order(organizationunit.ByCreatedAt()).All(ctx)
		if err != nil {
			a.logger.LogError("get_organizations", "Failed to fetch organization units", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to fetch organizations", http.StatusInternalServerError)
			return
		}

		// Convert to frontend format
		organizations := make([]OrganizationResponse, len(units))
		for i, unit := range units {
			organizations[i] = a.convertToOrganizationResponse(unit)
		}

		// Build response
		page := 1
		pageSize := 100
		total := len(organizations)
		totalPages := (total + pageSize - 1) / pageSize

		response := OrganizationListResponse{
			Organizations: organizations,
			Pagination: PaginationResponse{
				Page:       page,
				PageSize:   pageSize,
				Total:      total,
				TotalPages: totalPages,
			},
		}

		a.logger.Info("Organizations fetched successfully",
			"count", len(organizations),
			"tenant_id", tenantID,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// CreateOrganization handles POST /api/v1/corehr/organizations
func (a *OrganizationAdapter) CreateOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			// For development, use default tenant
			tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
		}

		var req CreateOrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			a.logger.LogError("create_organization", "Invalid JSON payload", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" || req.UnitType == "" {
			http.Error(w, "name and unit_type are required", http.StatusBadRequest)
			return
		}

		// Convert to backend request
		unitReq := a.convertToCreateUnitRequest(req)

		// Create organization unit using existing handler logic
		builder := a.client.OrganizationUnit.Create().
			SetTenantID(tenantID).
			SetUnitType(organizationunit.UnitType(unitReq.UnitType)).
			SetName(unitReq.Name).
			SetStatus(organizationunit.Status(unitReq.Status))

		if unitReq.Description != nil {
			builder = builder.SetDescription(*unitReq.Description)
		}

		if unitReq.ParentUnitID != nil {
			builder = builder.SetParentUnitID(*unitReq.ParentUnitID)
		}

		if unitReq.Profile != nil {
			builder = builder.SetProfile(unitReq.Profile)
		}

		orgUnit, err := builder.Save(ctx)
		if err != nil {
			a.logger.LogError("create_organization", "Failed to create organization unit", err, map[string]interface{}{
				"name":      req.Name,
				"unit_type": req.UnitType,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to create organization", http.StatusInternalServerError)
			return
		}

		// Convert to frontend response
		response := a.convertToOrganizationResponse(orgUnit)

		a.logger.Info("Organization created successfully",
			"org_id", orgUnit.ID,
			"name", orgUnit.Name,
			"type", orgUnit.UnitType,
			"tenant_id", tenantID,
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetOrganization handles GET /api/v1/corehr/organizations/{id}
func (a *OrganizationAdapter) GetOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
		}

		// Get ID from URL path
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		// Fetch the organization unit
		orgUnit, err := a.client.OrganizationUnit.Query().
			Where(
				organizationunit.IDEQ(id),
				organizationunit.TenantIDEQ(tenantID),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "Organization not found", http.StatusNotFound)
				return
			}
			a.logger.LogError("get_organization", "Failed to fetch organization unit", err, map[string]interface{}{
				"org_id":    id,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to fetch organization", http.StatusInternalServerError)
			return
		}

		response := a.convertToOrganizationResponse(orgUnit)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// UpdateOrganization handles PUT /api/v1/corehr/organizations/{id}
func (a *OrganizationAdapter) UpdateOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
		}

		// Get ID from URL path
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		var req UpdateOrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			a.logger.LogError("update_organization", "Invalid JSON payload", err, map[string]interface{}{
				"org_id":    id,
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Fetch existing organization unit
		_, err = a.client.OrganizationUnit.Query().
			Where(
				organizationunit.IDEQ(id),
				organizationunit.TenantIDEQ(tenantID),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "Organization not found", http.StatusNotFound)
				return
			}
			a.logger.LogError("update_organization", "Failed to fetch existing organization unit", err, map[string]interface{}{
				"org_id":    id,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to fetch organization", http.StatusInternalServerError)
			return
		}

		// Build update query
		updateBuilder := a.client.OrganizationUnit.UpdateOneID(id)

		if req.Name != nil {
			updateBuilder = updateBuilder.SetName(*req.Name)
		}

		if req.Description != nil {
			updateBuilder = updateBuilder.SetDescription(*req.Description)
		}

		if req.ParentUnitID != nil {
			if pid, err := uuid.Parse(*req.ParentUnitID); err == nil {
				updateBuilder = updateBuilder.SetParentUnitID(pid)
			}
		}

		if req.Status != nil {
			updateBuilder = updateBuilder.SetStatus(organizationunit.Status(*req.Status))
		}

		if req.Profile != nil {
			updateBuilder = updateBuilder.SetProfile(req.Profile)
		}

		// Execute update
		orgUnit, err := updateBuilder.Save(ctx)
		if err != nil {
			a.logger.LogError("update_organization", "Failed to update organization unit", err, map[string]interface{}{
				"org_id":    id,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to update organization", http.StatusInternalServerError)
			return
		}

		response := a.convertToOrganizationResponse(orgUnit)

		a.logger.Info("Organization updated successfully",
			"org_id", orgUnit.ID,
			"name", orgUnit.Name,
			"tenant_id", tenantID,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// DeleteOrganization handles DELETE /api/v1/corehr/organizations/{id}
func (a *OrganizationAdapter) DeleteOrganization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
		}

		// Get ID from URL path
		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		// Check if organization unit exists and belongs to tenant
		exists, err := a.client.OrganizationUnit.Query().
			Where(
				organizationunit.IDEQ(id),
				organizationunit.TenantIDEQ(tenantID),
			).
			Exist(ctx)

		if err != nil {
			a.logger.LogError("delete_organization", "Failed to check organization unit existence", err, map[string]interface{}{
				"org_id":    id,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to check organization", http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "Organization not found", http.StatusNotFound)
			return
		}

		// Check for child units
		childCount, err := a.client.OrganizationUnit.Query().
			Where(
				organizationunit.ParentUnitIDEQ(id),
				organizationunit.TenantIDEQ(tenantID),
			).
			Count(ctx)

		if err != nil {
			a.logger.LogError("delete_organization", "Failed to check child units", err, map[string]interface{}{
				"org_id":    id,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to check child units", http.StatusInternalServerError)
			return
		}

		if childCount > 0 {
			http.Error(w, "Cannot delete organization with child units", http.StatusConflict)
			return
		}

		// Delete the organization unit
		err = a.client.OrganizationUnit.DeleteOneID(id).Exec(ctx)
		if err != nil {
			a.logger.LogError("delete_organization", "Failed to delete organization unit", err, map[string]interface{}{
				"org_id":    id,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to delete organization", http.StatusInternalServerError)
			return
		}

		a.logger.Info("Organization deleted successfully",
			"org_id", id,
			"tenant_id", tenantID,
		)

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetOrganizationStats handles GET /api/v1/corehr/organizations/stats
func (a *OrganizationAdapter) GetOrganizationStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			tenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
		}

		// Get total count
		total, err := a.client.OrganizationUnit.Query().
			Where(organizationunit.TenantIDEQ(tenantID)).
			Count(ctx)

		if err != nil {
			a.logger.LogError("get_organization_stats", "Failed to count organization units", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to get organization stats", http.StatusInternalServerError)
			return
		}

		// Get active count
		active, err := a.client.OrganizationUnit.Query().
			Where(organizationunit.TenantIDEQ(tenantID)).
			Where(organizationunit.StatusEQ(organizationunit.StatusACTIVE)).
			Count(ctx)

		if err != nil {
			a.logger.LogError("get_organization_stats", "Failed to count active organization units", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			active = 0 // Fallback
		}

		// Simple stats response matching frontend expectations
		stats := map[string]interface{}{
			"data": map[string]interface{}{
				"total":          total,
				"active":         active,
				"inactive":       total - active,
				"totalEmployees": 0, // TODO: Calculate from positions
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}