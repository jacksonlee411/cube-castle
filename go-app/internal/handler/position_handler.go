package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/organizationunit"
	"github.com/gaogu/cube-castle/go-app/ent/position"
	"github.com/gaogu/cube-castle/go-app/ent/positionoccupancyhistory"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/types"
)

// PositionHandler handles HTTP requests for positions
type PositionHandler struct {
	client *ent.Client
	logger *logging.StructuredLogger
}

// NewPositionHandler creates a new position handler
func NewPositionHandler(client *ent.Client, logger *logging.StructuredLogger) *PositionHandler {
	return &PositionHandler{
		client: client,
		logger: logger,
	}
}

// CreatePositionRequest represents the request to create a position
type CreatePositionRequest struct {
	PositionType      string                 `json:"position_type" validate:"required,oneof=FULL_TIME PART_TIME CONTINGENT_WORKER INTERN"`
	JobProfileID      uuid.UUID              `json:"job_profile_id" validate:"required"`
	DepartmentID      uuid.UUID              `json:"department_id" validate:"required"`
	ManagerPositionID *uuid.UUID             `json:"manager_position_id,omitempty"`
	Status            string                 `json:"status" validate:"oneof=OPEN FILLED FROZEN PENDING_ELIMINATION"`
	BudgetedFTE       float64                `json:"budgeted_fte" validate:"gte=0,lte=5"`
	Details           map[string]interface{} `json:"details,omitempty"`
}

// UpdatePositionRequest represents the request to update a position
type UpdatePositionRequest struct {
	JobProfileID      *uuid.UUID             `json:"job_profile_id,omitempty"`
	DepartmentID      *uuid.UUID             `json:"department_id,omitempty"`
	ManagerPositionID *uuid.UUID             `json:"manager_position_id,omitempty"`
	Status            *string                `json:"status,omitempty" validate:"omitempty,oneof=OPEN FILLED FROZEN PENDING_ELIMINATION"`
	BudgetedFTE       *float64               `json:"budgeted_fte,omitempty" validate:"omitempty,gte=0,lte=5"`
	Details           map[string]interface{} `json:"details,omitempty"`
}

// PositionResponse represents the response format for position data
type PositionResponse struct {
	ID                uuid.UUID              `json:"id"`
	TenantID          uuid.UUID              `json:"tenant_id"`
	PositionType      string                 `json:"position_type"`
	JobProfileID      uuid.UUID              `json:"job_profile_id"`
	DepartmentID      uuid.UUID              `json:"department_id"`
	ManagerPositionID *uuid.UUID             `json:"manager_position_id"`
	Status            string                 `json:"status"`
	BudgetedFTE       float64                `json:"budgeted_fte"`
	Details           map[string]interface{} `json:"details"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// CreatePosition handles POST /api/v1/positions
func (h *PositionHandler) CreatePosition() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Get tenant ID from context (set by middleware)
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("create_position", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreatePositionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("create_position", "Invalid JSON payload", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.PositionType == "" || req.JobProfileID == uuid.Nil || req.DepartmentID == uuid.Nil {
			http.Error(w, "position_type, job_profile_id, and department_id are required", http.StatusBadRequest)
			return
		}

		// Verify department exists
		departmentExists, err := h.client.OrganizationUnit.Query().
			Where(
				organizationunit.IDEQ(req.DepartmentID),
				organizationunit.TenantIDEQ(tenantID),
			).
			Exist(ctx)

		if err != nil {
			h.logger.LogError("create_position", "Failed to check department existence", err, map[string]interface{}{
				"department_id": req.DepartmentID,
				"tenant_id":     tenantID,
			})
			http.Error(w, "Failed to verify department", http.StatusInternalServerError)
			return
		}

		if !departmentExists {
			http.Error(w, "Department not found", http.StatusBadRequest)
			return
		}

		// Validate manager position if provided
		if req.ManagerPositionID != nil {
			managerExists, err := h.client.Position.Query().
				Where(
					position.IDEQ(*req.ManagerPositionID),
					position.TenantIDEQ(tenantID),
				).
				Exist(ctx)

			if err != nil {
				h.logger.LogError("create_position", "Failed to check manager position existence", err, map[string]interface{}{
					"manager_position_id": *req.ManagerPositionID,
					"tenant_id":           tenantID,
				})
				http.Error(w, "Failed to verify manager position", http.StatusInternalServerError)
				return
			}

			if !managerExists {
				http.Error(w, "Manager position not found", http.StatusBadRequest)
				return
			}
		}

		// Validate details based on position type
		var detailsJSON map[string]interface{}
		if req.Details != nil {
			details, err := types.PositionDetailsFactory(req.PositionType, json.RawMessage(`{}`))
			if err != nil {
				h.logger.LogError("create_position", "Invalid position type", err, map[string]interface{}{
					"position_type": req.PositionType,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Invalid position type", http.StatusBadRequest)
				return
			}

			// Convert details map to JSON and validate
			detailsData, _ := json.Marshal(req.Details)
			details, err = types.PositionDetailsFactory(req.PositionType, detailsData)
			if err != nil {
				h.logger.LogError("create_position", "Invalid details data", err, map[string]interface{}{
					"position_type": req.PositionType,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Invalid details data for position type", http.StatusBadRequest)
				return
			}

			if err := details.Validate(); err != nil {
				h.logger.LogError("create_position", "Details validation failed", err, map[string]interface{}{
					"position_type": req.PositionType,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Details validation failed: "+err.Error(), http.StatusBadRequest)
				return
			}

			detailsJSON = req.Details
		}

		// Set default status and FTE if not provided
		status := req.Status
		if status == "" {
			status = "OPEN"
		}

		budgetedFTE := req.BudgetedFTE
		if budgetedFTE == 0 {
			budgetedFTE = 1.0
		}

		// Create the position
		builder := h.client.Position.Create().
			SetTenantID(tenantID).
			SetPositionType(position.PositionType(req.PositionType)).
			SetJobProfileID(req.JobProfileID).
			SetDepartmentID(req.DepartmentID).
			SetStatus(position.Status(status)).
			SetBudgetedFte(budgetedFTE)

		if req.ManagerPositionID != nil {
			builder = builder.SetManagerPositionID(*req.ManagerPositionID)
		}

		if detailsJSON != nil {
			builder = builder.SetDetails(detailsJSON)
		}

		pos, err := builder.Save(ctx)
		if err != nil {
			// Check if it's a validation error
			if strings.Contains(err.Error(), "invalid enum value") || strings.Contains(err.Error(), "validator failed") {
				h.logger.LogError("create_position", "Invalid field value", err, map[string]interface{}{
					"position_type": req.PositionType,
					"tenant_id":     tenantID,
				})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			h.logger.LogError("create_position", "Failed to create position", err, map[string]interface{}{
				"position_type": req.PositionType,
				"department_id": req.DepartmentID,
				"tenant_id":     tenantID,
			})
			http.Error(w, "Failed to create position", http.StatusInternalServerError)
			return
		}

		// Convert to response format
		response := h.convertToResponse(pos)

		h.logger.Info("Position created successfully",
			"position_id", pos.ID,
			"position_type", pos.PositionType,
			"department_id", pos.DepartmentID,
			"tenant_id", tenantID,
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetPosition handles GET /api/v1/positions/{id}
func (h *PositionHandler) GetPosition() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("get_position", "No tenant ID in context", nil, nil)
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

		// Fetch the position
		pos, err := h.client.Position.Query().
			Where(
				position.IDEQ(id),
				position.TenantIDEQ(tenantID),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "Position not found", http.StatusNotFound)
				return
			}
			h.logger.LogError("get_position", "Failed to fetch position", err, map[string]interface{}{
				"position_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to fetch position", http.StatusInternalServerError)
			return
		}

		response := h.convertToResponse(pos)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// ListPositions handles GET /api/v1/positions
func (h *PositionHandler) ListPositions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("list_positions", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse query parameters
		query := h.client.Position.Query().Where(position.TenantIDEQ(tenantID))

		// Filter by position type if provided
		if positionType := r.URL.Query().Get("position_type"); positionType != "" {
			query = query.Where(position.PositionTypeEQ(position.PositionType(positionType)))
		}

		// Filter by status if provided
		if status := r.URL.Query().Get("status"); status != "" {
			query = query.Where(position.StatusEQ(position.Status(status)))
		}

		// Filter by department ID if provided
		if departmentIDStr := r.URL.Query().Get("department_id"); departmentIDStr != "" {
			if departmentID, err := uuid.Parse(departmentIDStr); err == nil {
				query = query.Where(position.DepartmentIDEQ(departmentID))
			}
		}

		// Filter by manager position ID if provided
		if managerIDStr := r.URL.Query().Get("manager_position_id"); managerIDStr != "" {
			if managerID, err := uuid.Parse(managerIDStr); err == nil {
				query = query.Where(position.ManagerPositionIDEQ(managerID))
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
		positions, err := query.
			Limit(limit).
			Offset(offset).
			Order(position.ByCreatedAt()).
			All(ctx)

		if err != nil {
			h.logger.LogError("list_positions", "Failed to fetch positions", err, map[string]interface{}{
				"tenant_id": tenantID,
				"limit":     limit,
				"offset":    offset,
			})
			http.Error(w, "Failed to fetch positions", http.StatusInternalServerError)
			return
		}

		// Convert to response format
		responses := make([]PositionResponse, len(positions))
		for i, position := range positions {
			responses[i] = h.convertToResponse(position)
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

// UpdatePosition handles PUT /api/v1/positions/{id}
func (h *PositionHandler) UpdatePosition() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("update_position", "No tenant ID in context", nil, nil)
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

		var req UpdatePositionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("update_position", "Invalid JSON payload", err, map[string]interface{}{
				"position_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Fetch existing position to get current position_type for details validation
		existingPosition, err := h.client.Position.Query().
			Where(
				position.IDEQ(id),
				position.TenantIDEQ(tenantID),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "Position not found", http.StatusNotFound)
				return
			}
			h.logger.LogError("update_position", "Failed to fetch existing position", err, map[string]interface{}{
				"position_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to fetch position", http.StatusInternalServerError)
			return
		}

		// Build update query
		updateBuilder := h.client.Position.UpdateOneID(id)

		if req.JobProfileID != nil {
			updateBuilder = updateBuilder.SetJobProfileID(*req.JobProfileID)
		}

		if req.DepartmentID != nil {
			// Verify department exists
			departmentExists, err := h.client.OrganizationUnit.Query().
				Where(
					organizationunit.IDEQ(*req.DepartmentID),
					organizationunit.TenantIDEQ(tenantID),
				).
				Exist(ctx)

			if err != nil {
				h.logger.LogError("update_position", "Failed to check department existence", err, map[string]interface{}{
					"department_id": *req.DepartmentID,
					"position_id":   id,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Failed to verify department", http.StatusInternalServerError)
				return
			}

			if !departmentExists {
				http.Error(w, "Department not found", http.StatusBadRequest)
				return
			}

			updateBuilder = updateBuilder.SetDepartmentID(*req.DepartmentID)
		}

		if req.ManagerPositionID != nil {
			// Verify manager position exists
			managerExists, err := h.client.Position.Query().
				Where(
					position.IDEQ(*req.ManagerPositionID),
					position.TenantIDEQ(tenantID),
				).
				Exist(ctx)

			if err != nil {
				h.logger.LogError("update_position", "Failed to check manager position existence", err, map[string]interface{}{
					"manager_position_id": *req.ManagerPositionID,
					"position_id":         id,
					"tenant_id":           tenantID,
				})
				http.Error(w, "Failed to verify manager position", http.StatusInternalServerError)
				return
			}

			if !managerExists {
				http.Error(w, "Manager position not found", http.StatusBadRequest)
				return
			}

			updateBuilder = updateBuilder.SetManagerPositionID(*req.ManagerPositionID)
		}

		if req.Status != nil {
			updateBuilder = updateBuilder.SetStatus(position.Status(*req.Status))
		}

		if req.BudgetedFTE != nil {
			updateBuilder = updateBuilder.SetBudgetedFte(*req.BudgetedFTE)
		}

		if req.Details != nil {
			// Validate details based on existing position type
			detailsData, _ := json.Marshal(req.Details)
			details, err := types.PositionDetailsFactory(string(existingPosition.PositionType), detailsData)
			if err != nil {
				h.logger.LogError("update_position", "Invalid details data", err, map[string]interface{}{
					"position_type": existingPosition.PositionType,
					"position_id":   id,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Invalid details data for position type", http.StatusBadRequest)
				return
			}

			if err := details.Validate(); err != nil {
				h.logger.LogError("update_position", "Details validation failed", err, map[string]interface{}{
					"position_type": existingPosition.PositionType,
					"position_id":   id,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Details validation failed: "+err.Error(), http.StatusBadRequest)
				return
			}

			// Convert details to map[string]interface{} for storage
			var detailsMap map[string]interface{}
			detailsJSON, err := json.Marshal(details)
			if err != nil {
				h.logger.LogError("update_position", "Failed to serialize details", err, map[string]interface{}{
					"position_id": id,
					"tenant_id":   tenantID,
				})
				http.Error(w, "Failed to process details", http.StatusInternalServerError)
				return
			}
			
			err = json.Unmarshal(detailsJSON, &detailsMap)
			if err != nil {
				h.logger.LogError("update_position", "Failed to convert details to map", err, map[string]interface{}{
					"position_id": id,
					"tenant_id":   tenantID,
				})
				http.Error(w, "Failed to process details", http.StatusInternalServerError)
				return
			}

			updateBuilder = updateBuilder.SetDetails(detailsMap)
		}

		// Execute update
		position, err := updateBuilder.Save(ctx)
		if err != nil {
			// Check if it's a validation error
			if strings.Contains(err.Error(), "invalid enum value") || strings.Contains(err.Error(), "validator failed") {
				h.logger.LogError("update_position", "Invalid field value", err, map[string]interface{}{
					"position_id": id,
					"tenant_id":   tenantID,
				})
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			h.logger.LogError("update_position", "Failed to update position", err, map[string]interface{}{
				"position_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to update position", http.StatusInternalServerError)
			return
		}

		response := h.convertToResponse(position)

		h.logger.Info("Position updated successfully",
			"position_id", position.ID,
			"position_type", position.PositionType,
			"department_id", position.DepartmentID,
			"tenant_id", tenantID,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// DeletePosition handles DELETE /api/v1/positions/{id}
func (h *PositionHandler) DeletePosition() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Get tenant ID from context
		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("delete_position", "No tenant ID in context", nil, nil)
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

		// Check if position exists and belongs to tenant
		exists, err := h.client.Position.Query().
			Where(
				position.IDEQ(id),
				position.TenantIDEQ(tenantID),
			).
			Exist(ctx)

		if err != nil {
			h.logger.LogError("delete_position", "Failed to check position existence", err, map[string]interface{}{
				"position_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to check position", http.StatusInternalServerError)
			return
		}

		if !exists {
			http.Error(w, "Position not found", http.StatusNotFound)
			return
		}

		// Check for positions that report to this position
		subordinateCount, err := h.client.Position.Query().
			Where(
				position.ManagerPositionIDEQ(id),
				position.TenantIDEQ(tenantID),
			).
			Count(ctx)

		if err != nil {
			h.logger.LogError("delete_position", "Failed to check subordinate positions", err, map[string]interface{}{
				"position_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to check subordinate positions", http.StatusInternalServerError)
			return
		}

		if subordinateCount > 0 {
			http.Error(w, "Cannot delete position with subordinate positions", http.StatusConflict)
			return
		}

		// Check for occupancy history (if position has been occupied)
		occupancyCount, err := h.client.PositionOccupancyHistory.Query().
			Where(
				positionoccupancyhistory.PositionIDEQ(id),
				positionoccupancyhistory.TenantIDEQ(tenantID),
			).
			Count(ctx)

		if err != nil {
			h.logger.LogError("delete_position", "Failed to check occupancy history", err, map[string]interface{}{
				"position_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to check occupancy history", http.StatusInternalServerError)
			return
		}

		if occupancyCount > 0 {
			http.Error(w, "Cannot delete position with historical occupancy data", http.StatusConflict)
			return
		}

		// Delete the position
		err = h.client.Position.DeleteOneID(id).Exec(ctx)
		if err != nil {
			h.logger.LogError("delete_position", "Failed to delete position", err, map[string]interface{}{
				"position_id": id,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to delete position", http.StatusInternalServerError)
			return
		}

		h.logger.Info("Position deleted successfully",
			"position_id", id,
			"tenant_id", tenantID,
		)

		w.WriteHeader(http.StatusNoContent)
	}
}

// convertToResponse converts an ent.Position to PositionResponse
func (h *PositionHandler) convertToResponse(position *ent.Position) PositionResponse {
	return PositionResponse{
		ID:                position.ID,
		TenantID:          position.TenantID,
		PositionType:      string(position.PositionType),
		JobProfileID:      position.JobProfileID,
		DepartmentID:      position.DepartmentID,
		ManagerPositionID: position.ManagerPositionID,
		Status:            string(position.Status),
		BudgetedFTE:       position.BudgetedFte,
		Details:           position.Details,
		CreatedAt:         position.CreatedAt,
		UpdatedAt:         position.UpdatedAt,
	}
}