package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PositionAssignmentHandler handles HTTP requests for position assignment operations
type PositionAssignmentHandler struct {
	service *service.PositionAssignmentService
	logger  *logging.StructuredLogger
}

// NewPositionAssignmentHandler creates a new PositionAssignmentHandler
func NewPositionAssignmentHandler(service *service.PositionAssignmentService, logger *logging.StructuredLogger) *PositionAssignmentHandler {
	return &PositionAssignmentHandler{
		service: service,
		logger:  logger,
	}
}

// AssignPosition handles POST /api/v1/assignments
func (h *PositionAssignmentHandler) AssignPosition() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("assign_position", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req service.AssignmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("assign_position", "Invalid JSON payload", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		result, err := h.service.AssignPosition(ctx, tenantID, req)
		if err != nil {
			h.logger.LogError("assign_position", "Assignment failed", err, map[string]interface{}{
				"employee_id": req.EmployeeID,
				"position_id": req.PositionID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Assignment failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(result)
	}
}

// TransferEmployee handles POST /api/v1/assignments/transfer
func (h *PositionAssignmentHandler) TransferEmployee() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("transfer_employee", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req service.TransferRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("transfer_employee", "Invalid JSON payload", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		result, err := h.service.TransferEmployee(ctx, tenantID, req)
		if err != nil {
			h.logger.LogError("transfer_employee", "Transfer failed", err, map[string]interface{}{
				"employee_id":      req.EmployeeID,
				"from_position_id": req.FromPositionID,
				"to_position_id":   req.ToPositionID,
				"tenant_id":        tenantID,
			})
			http.Error(w, "Transfer failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

// EndAssignment handles DELETE /api/v1/assignments/{employeeId}
func (h *PositionAssignmentHandler) EndAssignment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("end_assignment", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		employeeIDStr := chi.URLParam(r, "employeeId")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		// Parse request body for end date and reason
		var req struct {
			EndDate time.Time `json:"end_date"`
			Reason  string    `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("end_assignment", "Invalid JSON payload", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		err = h.service.EndAssignment(ctx, tenantID, employeeID, req.EndDate, req.Reason)
		if err != nil {
			h.logger.LogError("end_assignment", "Failed to end assignment", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to end assignment: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetActiveAssignments handles GET /api/v1/assignments/active
func (h *PositionAssignmentHandler) GetActiveAssignments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("get_active_assignments", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		assignments, err := h.service.GetActiveAssignments(ctx, tenantID)
		if err != nil {
			h.logger.LogError("get_active_assignments", "Failed to fetch active assignments", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to fetch active assignments", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"assignments": assignments,
			"total":       len(assignments),
		})
	}
}

// EmployeeLifecycleHandler handles HTTP requests for employee lifecycle operations
type EmployeeLifecycleHandler struct {
	service *service.EmployeeLifecycleService
	logger  *logging.StructuredLogger
}

// NewEmployeeLifecycleHandler creates a new EmployeeLifecycleHandler
func NewEmployeeLifecycleHandler(service *service.EmployeeLifecycleService, logger *logging.StructuredLogger) *EmployeeLifecycleHandler {
	return &EmployeeLifecycleHandler{
		service: service,
		logger:  logger,
	}
}

// OnboardEmployee handles POST /api/v1/lifecycle/onboard
func (h *EmployeeLifecycleHandler) OnboardEmployee() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("onboard_employee", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req service.OnboardingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("onboard_employee", "Invalid JSON payload", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		result, err := h.service.OnboardEmployee(ctx, tenantID, req)
		if err != nil {
			h.logger.LogError("onboard_employee", "Onboarding failed", err, map[string]interface{}{
				"employee_number": req.EmployeeNumber,
				"tenant_id":       tenantID,
			})
			http.Error(w, "Onboarding failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(result)
	}
}

// OffboardEmployee handles POST /api/v1/lifecycle/offboard
func (h *EmployeeLifecycleHandler) OffboardEmployee() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("offboard_employee", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req service.OffboardingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("offboard_employee", "Invalid JSON payload", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		err := h.service.OffboardEmployee(ctx, tenantID, req)
		if err != nil {
			h.logger.LogError("offboard_employee", "Offboarding failed", err, map[string]interface{}{
				"employee_id": req.EmployeeID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Offboarding failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Employee offboarded successfully",
		})
	}
}

// PromoteEmployee handles POST /api/v1/lifecycle/promote
func (h *EmployeeLifecycleHandler) PromoteEmployee() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("promote_employee", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req service.PromotionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("promote_employee", "Invalid JSON payload", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		result, err := h.service.PromoteEmployee(ctx, tenantID, req)
		if err != nil {
			h.logger.LogError("promote_employee", "Promotion failed", err, map[string]interface{}{
				"employee_id":     req.EmployeeID,
				"new_position_id": req.NewPositionID,
				"tenant_id":       tenantID,
			})
			http.Error(w, "Promotion failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

// ChangeEmploymentStatus handles POST /api/v1/lifecycle/status-change
func (h *EmployeeLifecycleHandler) ChangeEmploymentStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("change_status", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req service.StatusChangeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("change_status", "Invalid JSON payload", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		err := h.service.ChangeEmploymentStatus(ctx, tenantID, req)
		if err != nil {
			h.logger.LogError("change_status", "Status change failed", err, map[string]interface{}{
				"employee_id": req.EmployeeID,
				"new_status":  req.NewStatus,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Status change failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Employment status changed successfully",
		})
	}
}

// AnalyticsHandler handles HTTP requests for analytics and reporting
type AnalyticsHandler struct {
	service *service.AnalyticsService
	logger  *logging.StructuredLogger
}

// NewAnalyticsHandler creates a new AnalyticsHandler
func NewAnalyticsHandler(service *service.AnalyticsService, logger *logging.StructuredLogger) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
		logger:  logger,
	}
}

// GetOrganizationalMetrics handles GET /api/v1/analytics/metrics
func (h *AnalyticsHandler) GetOrganizationalMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("get_metrics", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		metrics, err := h.service.GetOrganizationalMetrics(ctx, tenantID)
		if err != nil {
			h.logger.LogError("get_metrics", "Failed to generate metrics", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to generate metrics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	}
}

// GetEmployeeHistory handles GET /api/v1/analytics/employees/{id}/history
func (h *AnalyticsHandler) GetEmployeeHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("get_employee_history", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		employeeIDStr := chi.URLParam(r, "id")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		history, err := h.service.GetEmployeeHistory(ctx, tenantID, employeeID)
		if err != nil {
			h.logger.LogError("get_employee_history", "Failed to fetch employee history", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to fetch employee history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	}
}

// GetPositionHistory handles GET /api/v1/analytics/positions/{id}/history
func (h *AnalyticsHandler) GetPositionHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("get_position_history", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		positionIDStr := chi.URLParam(r, "id")
		positionID, err := uuid.Parse(positionIDStr)
		if err != nil {
			http.Error(w, "Invalid position ID format", http.StatusBadRequest)
			return
		}

		history, err := h.service.GetPositionHistory(ctx, tenantID, positionID)
		if err != nil {
			h.logger.LogError("get_position_history", "Failed to fetch position history", err, map[string]interface{}{
				"position_id": positionID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to fetch position history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(history)
	}
}

// GetHistoricalAssignments handles GET /api/v1/analytics/assignments/history
func (h *AnalyticsHandler) GetHistoricalAssignments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("get_assignment_history", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse query parameters
		params := service.HistoryQueryParams{}

		if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
			if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
				params.StartDate = &startDate
			}
		}

		if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
			if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
				params.EndDate = &endDate
			}
		}

		if employeeIDStr := r.URL.Query().Get("employee_id"); employeeIDStr != "" {
			if employeeID, err := uuid.Parse(employeeIDStr); err == nil {
				params.EmployeeID = &employeeID
			}
		}

		if positionIDStr := r.URL.Query().Get("position_id"); positionIDStr != "" {
			if positionID, err := uuid.Parse(positionIDStr); err == nil {
				params.PositionID = &positionID
			}
		}

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
				params.Limit = limit
			}
		}

		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
				params.Offset = offset
			}
		}

		assignments, err := h.service.GetHistoricalAssignments(ctx, tenantID, params)
		if err != nil {
			h.logger.LogError("get_assignment_history", "Failed to fetch assignment history", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to fetch assignment history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"assignments": assignments,
			"total":       len(assignments),
			"params":      params,
		})
	}
}