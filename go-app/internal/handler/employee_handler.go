package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/employee"
	"github.com/gaogu/cube-castle/go-app/ent/organizationunit"
	"github.com/gaogu/cube-castle/go-app/ent/position"
	"github.com/gaogu/cube-castle/go-app/ent/positionoccupancyhistory"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// EmployeeHandler handles HTTP requests for employee management
type EmployeeHandler struct {
	client *ent.Client
	logger *logging.StructuredLogger
}

// isValidBusinessID validates business ID format (1-99999999)
func isValidBusinessID(businessID string) bool {
	matched, _ := regexp.MatchString(`^[1-9][0-9]{0,7}$`, businessID)
	return matched
}

// isValidOrganizationBusinessID validates organization business ID format (100000-999999)
func isValidOrganizationBusinessID(businessID string) bool {
	matched, _ := regexp.MatchString(`^[1-9][0-9]{5}$`, businessID)
	return matched
}

// NewEmployeeHandler creates a new employee handler
func NewEmployeeHandler(client *ent.Client, logger *logging.StructuredLogger) *EmployeeHandler {
	return &EmployeeHandler{
		client: client,
		logger: logger,
	}
}

// Request/Response structures
type CreateEmployeeRequest struct {
	EmployeeType        string                 `json:"employee_type" validate:"required,oneof=FULL_TIME PART_TIME CONTRACTOR INTERN"`
	EmployeeNumber      string                 `json:"employee_number" validate:"required,min=1,max=50"`
	PersonName          string                 `json:"person_name" validate:"required,min=1,max=200"`
	Email               string                 `json:"email" validate:"required,email,max=255"`
	PersonalEmail       *string                `json:"personal_email,omitempty"`
	PhoneNumber         *string                `json:"phone_number,omitempty"`
	CurrentPositionID   *uuid.UUID             `json:"current_position_id,omitempty"`
	EmploymentStatus    string                 `json:"employment_status" validate:"oneof=ACTIVE ON_LEAVE TERMINATED SUSPENDED PENDING_START"`
	HireDate            time.Time              `json:"hire_date" validate:"required"`
	TerminationDate     *time.Time             `json:"termination_date,omitempty"`
	EmployeeDetails     map[string]interface{} `json:"employee_details,omitempty"`
}

type UpdateEmployeeRequest struct {
	PersonName          *string                `json:"person_name,omitempty" validate:"omitempty,min=1,max=200"`
	PersonalEmail       *string                `json:"personal_email,omitempty"`
	PhoneNumber         *string                `json:"phone_number,omitempty"`
	CurrentPositionID   *uuid.UUID             `json:"current_position_id,omitempty"`
	OrganizationID      *string                `json:"organization_id,omitempty"` // 组织业务ID
	EmploymentStatus    *string                `json:"employment_status,omitempty" validate:"omitempty,oneof=ACTIVE ON_LEAVE TERMINATED SUSPENDED PENDING_START"`
	TerminationDate     *time.Time             `json:"termination_date,omitempty"`
	EmployeeDetails     map[string]interface{} `json:"employee_details,omitempty"`
}

type EmployeeResponse struct {
	ID                  uuid.UUID              `json:"id"`
	TenantID            uuid.UUID              `json:"tenant_id"`
	EmployeeType        string                 `json:"employee_type"`
	EmployeeNumber      string                 `json:"employee_number"`
	PersonName          string                 `json:"person_name"` // 统一姓名字段
	Email               string                 `json:"email"`
	PersonalEmail       *string                `json:"personal_email"`
	PhoneNumber         *string                `json:"phone_number"`
	CurrentPositionID   *uuid.UUID             `json:"current_position_id"`
	CurrentPosition     *PositionSummary       `json:"current_position,omitempty"` // Associated data
	EmploymentStatus    string                 `json:"employment_status"`
	HireDate            time.Time              `json:"hire_date"`
	TerminationDate     *time.Time             `json:"termination_date"`
	EmployeeDetails     map[string]interface{} `json:"employee_details"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

type PositionSummary struct {
	ID           uuid.UUID `json:"id"`
	PositionType string    `json:"position_type"`
	DepartmentID uuid.UUID `json:"department_id"`
	Status       string    `json:"status"`
}

type AssignPositionRequest struct {
	PositionID       uuid.UUID `json:"position_id" validate:"required"`
	StartDate        time.Time `json:"start_date" validate:"required"`
	AssignmentType   string    `json:"assignment_type" validate:"oneof=REGULAR INTERIM ACTING TEMPORARY SECONDMENT"`
	AssignmentReason string    `json:"assignment_reason,omitempty"`
	FTEPercentage    float64   `json:"fte_percentage" validate:"gte=0.1,lte=1.0"`
	ApprovedBy       uuid.UUID `json:"approved_by" validate:"required"`
}

type TransferEmployeeRequest struct {
	NewPositionID uuid.UUID `json:"new_position_id" validate:"required"`
	TransferDate  time.Time `json:"transfer_date" validate:"required"`
	Reason        string    `json:"reason" validate:"required"`
	ApprovedBy    uuid.UUID `json:"approved_by" validate:"required"`
}

// CRUD Operations

// CreateEmployee handles POST /api/v1/employees
func (h *EmployeeHandler) CreateEmployee() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("create_employee", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreateEmployeeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("create_employee", "Invalid JSON payload", err, map[string]interface{}{
				"tenant_id": tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Validation
		if req.EmployeeType == "" || req.EmployeeNumber == "" || req.PersonName == "" || req.Email == "" {
			http.Error(w, "employee_type, employee_number, first_name, last_name, and email are required", http.StatusBadRequest)
			return
		}

		// Check employee number uniqueness
		exists, err := h.client.Employee.Query().
			Where(
				employee.TenantIDEQ(tenantID),
				employee.EmployeeNumberEQ(req.EmployeeNumber),
			).
			Exist(ctx)

		if err != nil {
			h.logger.LogError("create_employee", "Failed to check employee number uniqueness", err, map[string]interface{}{
				"employee_number": req.EmployeeNumber,
				"tenant_id":       tenantID,
			})
			http.Error(w, "Failed to validate employee number", http.StatusInternalServerError)
			return
		}

		if exists {
			http.Error(w, "Employee number already exists", http.StatusConflict)
			return
		}

		// Check email uniqueness
		emailExists, err := h.client.Employee.Query().
			Where(
				employee.TenantIDEQ(tenantID),
				employee.EmailEQ(req.Email),
			).
			Exist(ctx)

		if err != nil {
			h.logger.LogError("create_employee", "Failed to check email uniqueness", err, map[string]interface{}{
				"email":     req.Email,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to validate email", http.StatusInternalServerError)
			return
		}

		if emailExists {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}

		// Validate position exists (if provided)
		if req.CurrentPositionID != nil {
			positionExists, err := h.client.Position.Query().
				Where(
					position.IDEQ(*req.CurrentPositionID),
					position.TenantIDEQ(tenantID),
				).
				Exist(ctx)

			if err != nil {
				h.logger.LogError("create_employee", "Failed to check position existence", err, map[string]interface{}{
					"position_id": *req.CurrentPositionID,
					"tenant_id":   tenantID,
				})
				http.Error(w, "Failed to verify position", http.StatusInternalServerError)
				return
			}

			if !positionExists {
				http.Error(w, "Position not found", http.StatusBadRequest)
				return
			}
		}

		// Validate employee details if provided
		if req.EmployeeDetails != nil {
			detailsData, _ := json.Marshal(req.EmployeeDetails)
			details, err := types.EmployeeDetailsFactory(req.EmployeeType, detailsData)
			if err != nil {
				h.logger.LogError("create_employee", "Invalid employee details", err, map[string]interface{}{
					"employee_type": req.EmployeeType,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Invalid employee details for employee type", http.StatusBadRequest)
				return
			}

			if err := details.Validate(); err != nil {
				h.logger.LogError("create_employee", "Employee details validation failed", err, map[string]interface{}{
					"employee_type": req.EmployeeType,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Employee details validation failed: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		// Set default status
		status := req.EmploymentStatus
		if status == "" {
			status = "PENDING_START"
		}

		// Create employee record
		// 分割 PersonName 为 FirstName 和 LastName 以兼容数据库结构
		nameParts := strings.Fields(strings.TrimSpace(req.PersonName))
		var firstName, lastName string
		if len(nameParts) >= 2 {
			firstName = nameParts[0]
			lastName = strings.Join(nameParts[1:], " ")
		} else if len(nameParts) == 1 {
			firstName = nameParts[0]
			lastName = ""
		}

		builder := h.client.Employee.Create().
			SetTenantID(tenantID).
			SetEmployeeType(employee.EmployeeType(req.EmployeeType)).
			SetEmployeeNumber(req.EmployeeNumber).
			SetFirstName(firstName).
			SetLastName(lastName).
			SetEmail(req.Email).
			SetEmploymentStatus(employee.EmploymentStatus(status)).
			SetHireDate(req.HireDate)

		if req.PersonalEmail != nil {
			builder = builder.SetPersonalEmail(*req.PersonalEmail)
		}

		if req.PhoneNumber != nil {
			builder = builder.SetPhoneNumber(*req.PhoneNumber)
		}

		if req.CurrentPositionID != nil {
			builder = builder.SetCurrentPositionID(*req.CurrentPositionID)
		}

		if req.TerminationDate != nil {
			builder = builder.SetTerminationDate(*req.TerminationDate)
		}

		if req.EmployeeDetails != nil {
			builder = builder.SetEmployeeDetails(req.EmployeeDetails)
		}

		emp, err := builder.Save(ctx)
		if err != nil {
			h.logger.LogError("create_employee", "Failed to create employee", err, map[string]interface{}{
				"employee_number": req.EmployeeNumber,
				"tenant_id":       tenantID,
			})
			http.Error(w, "Failed to create employee", http.StatusInternalServerError)
			return
		}

		response := h.convertToResponse(emp, nil)

		h.logger.Info("Employee created successfully",
			"employee_id", emp.ID,
			"employee_number", emp.EmployeeNumber,
			"tenant_id", tenantID,
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetEmployee handles GET /api/v1/employees/{id}
func (h *EmployeeHandler) GetEmployee() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("get_employee", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		businessID := chi.URLParam(r, "id")
		
		// Validate business ID format (1-99999999)
		if !isValidBusinessID(businessID) {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		// Fetch employee with current position using business_id
		emp, err := h.client.Employee.Query().
			Where(
				employee.BusinessIDEQ(businessID),
				employee.TenantIDEQ(tenantID),
			).
			WithCurrentPosition().
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "Employee not found", http.StatusNotFound)
				return
			}
			h.logger.LogError("get_employee", "Failed to fetch employee", err, map[string]interface{}{
				"business_id": businessID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to fetch employee", http.StatusInternalServerError)
			return
		}

		response := h.convertToResponse(emp, emp.Edges.CurrentPosition)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// ListEmployees handles GET /api/v1/employees
func (h *EmployeeHandler) ListEmployees() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("list_employees", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse query parameters
		query := h.client.Employee.Query().Where(employee.TenantIDEQ(tenantID))

		// Filter by employee type if provided
		if empType := r.URL.Query().Get("employee_type"); empType != "" {
			query = query.Where(employee.EmployeeTypeEQ(employee.EmployeeType(empType)))
		}

		// Filter by employment status if provided
		if status := r.URL.Query().Get("employment_status"); status != "" {
			query = query.Where(employee.EmploymentStatusEQ(employee.EmploymentStatus(status)))
		}

		// Filter by current position if provided
		if positionIDStr := r.URL.Query().Get("current_position_id"); positionIDStr != "" {
			if positionID, err := uuid.Parse(positionIDStr); err == nil {
				query = query.Where(employee.CurrentPositionIDEQ(positionID))
			}
		}

		// Search by name
		if search := r.URL.Query().Get("search"); search != "" {
			query = query.Where(employee.Or(
				employee.FirstNameContains(search),
				employee.LastNameContains(search),
				employee.EmailContains(search),
				employee.EmployeeNumberContains(search),
				employee.BusinessIDContains(search),
			))
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
		employees, err := query.
			WithCurrentPosition().
			Limit(limit).
			Offset(offset).
			Order(employee.ByCreatedAt()).
			All(ctx)

		if err != nil {
			h.logger.LogError("list_employees", "Failed to fetch employees", err, map[string]interface{}{
				"tenant_id": tenantID,
				"limit":     limit,
				"offset":    offset,
			})
			http.Error(w, "Failed to fetch employees", http.StatusInternalServerError)
			return
		}

		// Convert to response format
		responses := make([]EmployeeResponse, len(employees))
		for i, emp := range employees {
			responses[i] = h.convertToResponse(emp, emp.Edges.CurrentPosition)
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

// UpdateEmployee handles PUT /api/v1/employees/{id}
func (h *EmployeeHandler) UpdateEmployee() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("update_employee", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		businessID := chi.URLParam(r, "id")
		
		// Validate business ID format (1-99999999)
		if !isValidBusinessID(businessID) {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		var req UpdateEmployeeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.LogError("update_employee", "Invalid JSON payload", err, map[string]interface{}{
				"business_id": businessID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Fetch existing employee using business_id
		existingEmployee, err := h.client.Employee.Query().
			Where(
				employee.BusinessIDEQ(businessID),
				employee.TenantIDEQ(tenantID),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "Employee not found", http.StatusNotFound)
				return
			}
			h.logger.LogError("update_employee", "Failed to fetch existing employee", err, map[string]interface{}{
				"business_id": businessID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to fetch employee", http.StatusInternalServerError)
			return
		}

		// Build update query
		updateBuilder := h.client.Employee.UpdateOne(existingEmployee)

		if req.PersonName != nil {
			// 分割 PersonName 为 FirstName 和 LastName
			nameParts := strings.Fields(strings.TrimSpace(*req.PersonName))
			var firstName, lastName string
			if len(nameParts) >= 2 {
				firstName = nameParts[0]
				lastName = strings.Join(nameParts[1:], " ")
			} else if len(nameParts) == 1 {
				firstName = nameParts[0]
				lastName = ""
			}
			updateBuilder = updateBuilder.SetFirstName(firstName).SetLastName(lastName)
		}

		if req.PersonalEmail != nil {
			updateBuilder = updateBuilder.SetPersonalEmail(*req.PersonalEmail)
		}

		if req.PhoneNumber != nil {
			updateBuilder = updateBuilder.SetPhoneNumber(*req.PhoneNumber)
		}

		if req.CurrentPositionID != nil {
			// Verify position exists
			positionExists, err := h.client.Position.Query().
				Where(
					position.IDEQ(*req.CurrentPositionID),
					position.TenantIDEQ(tenantID),
				).
				Exist(ctx)

			if err != nil {
				h.logger.LogError("update_employee", "Failed to check position existence", err, map[string]interface{}{
					"position_id":   *req.CurrentPositionID,
					"business_id":   businessID,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Failed to verify position", http.StatusInternalServerError)
				return
			}

			if !positionExists {
				http.Error(w, "Position not found", http.StatusBadRequest)
				return
			}

			updateBuilder = updateBuilder.SetCurrentPositionID(*req.CurrentPositionID)
		}

		if req.OrganizationID != nil {
			// Validate organization business ID format
			if !isValidOrganizationBusinessID(*req.OrganizationID) {
				http.Error(w, "Invalid organization ID format", http.StatusBadRequest)
				return
			}

			// Verify organization exists using business_id
			org, err := h.client.OrganizationUnit.Query().
				Where(
					organizationunit.BusinessIDEQ(*req.OrganizationID),
					organizationunit.TenantIDEQ(tenantID),
				).
				Only(ctx)

			if err != nil {
				if ent.IsNotFound(err) {
					http.Error(w, "Organization not found", http.StatusBadRequest)
					return
				}
				h.logger.LogError("update_employee", "Failed to check organization existence", err, map[string]interface{}{
					"organization_id": *req.OrganizationID,
					"business_id":     businessID,
					"tenant_id":       tenantID,
				})
				http.Error(w, "Failed to verify organization", http.StatusInternalServerError)
				return
			}

			// Set the organization UUID in department_id field
			updateBuilder = updateBuilder.SetDepartmentID(org.ID)
		}

		if req.EmploymentStatus != nil {
			updateBuilder = updateBuilder.SetEmploymentStatus(employee.EmploymentStatus(*req.EmploymentStatus))
		}

		if req.TerminationDate != nil {
			updateBuilder = updateBuilder.SetTerminationDate(*req.TerminationDate)
		}

		if req.EmployeeDetails != nil {
			// Validate details based on existing employee type
			detailsData, _ := json.Marshal(req.EmployeeDetails)
			details, err := types.EmployeeDetailsFactory(string(existingEmployee.EmployeeType), detailsData)
			if err != nil {
				h.logger.LogError("update_employee", "Invalid employee details", err, map[string]interface{}{
					"employee_type": existingEmployee.EmployeeType,
					"business_id":   businessID,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Invalid employee details for employee type", http.StatusBadRequest)
				return
			}

			if err := details.Validate(); err != nil {
				h.logger.LogError("update_employee", "Employee details validation failed", err, map[string]interface{}{
					"employee_type": existingEmployee.EmployeeType,
					"business_id":   businessID,
					"tenant_id":     tenantID,
				})
				http.Error(w, "Employee details validation failed: "+err.Error(), http.StatusBadRequest)
				return
			}

			updateBuilder = updateBuilder.SetEmployeeDetails(req.EmployeeDetails)
		}

		// Execute update
		updatedEmployee, err := updateBuilder.Save(ctx)
		if err != nil {
			h.logger.LogError("update_employee", "Failed to update employee", err, map[string]interface{}{
				"business_id": businessID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to update employee", http.StatusInternalServerError)
			return
		}

		response := h.convertToResponse(updatedEmployee, nil)

		h.logger.Info("Employee updated successfully",
			"business_id", updatedEmployee.BusinessID,
			"employee_number", updatedEmployee.EmployeeNumber,
			"tenant_id", tenantID,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// DeleteEmployee handles DELETE /api/v1/employees/{id}
func (h *EmployeeHandler) DeleteEmployee() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("delete_employee", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		businessID := chi.URLParam(r, "id")
		
		// Validate business ID format (1-99999999)
		if !isValidBusinessID(businessID) {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		// Check if employee exists and get their UUID for further operations
		emp, err := h.client.Employee.Query().
			Where(
				employee.BusinessIDEQ(businessID),
				employee.TenantIDEQ(tenantID),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				http.Error(w, "Employee not found", http.StatusNotFound)
				return
			}
			h.logger.LogError("delete_employee", "Failed to check employee existence", err, map[string]interface{}{
				"business_id": businessID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to check employee", http.StatusInternalServerError)
			return
		}

		// Check for position occupancy history using employee UUID
		historyCount, err := h.client.PositionOccupancyHistory.Query().
			Where(
				positionoccupancyhistory.EmployeeIDEQ(emp.ID),
				positionoccupancyhistory.TenantIDEQ(tenantID),
			).
			Count(ctx)

		if err != nil {
			h.logger.LogError("delete_employee", "Failed to check occupancy history", err, map[string]interface{}{
				"business_id": businessID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to check occupancy history", http.StatusInternalServerError)
			return
		}

		if historyCount > 0 {
			http.Error(w, "Cannot delete employee with position history. Consider setting employment status to TERMINATED instead.", http.StatusConflict)
			return
		}

		// Delete the employee using UUID
		err = h.client.Employee.DeleteOneID(emp.ID).Exec(ctx)
		if err != nil {
			h.logger.LogError("delete_employee", "Failed to delete employee", err, map[string]interface{}{
				"business_id": businessID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to delete employee", http.StatusInternalServerError)
			return
		}

		h.logger.Info("Employee deleted successfully",
			"business_id", businessID,
			"tenant_id", tenantID,
		)

		w.WriteHeader(http.StatusNoContent)
	}
}

// Position Assignment Operations

// AssignPosition handles POST /api/v1/employees/{id}/assign-position
func (h *EmployeeHandler) AssignPosition() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("assign_position", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		employeeIDStr := chi.URLParam(r, "id")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		var req AssignPositionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Set default values
		if req.AssignmentType == "" {
			req.AssignmentType = "REGULAR"
		}
		if req.FTEPercentage == 0 {
			req.FTEPercentage = 1.0
		}

		// Execute position assignment in transaction
		tx, err := h.client.Tx(ctx)
		if err != nil {
			h.logger.LogError("assign_position", "Failed to start transaction", err, map[string]interface{}{
				"employee_id": employeeID,
				"position_id": req.PositionID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
			return
		}

		err = func() error {
			// Verify employee exists and is active
			emp, err := tx.Employee.Query().
				Where(
					employee.IDEQ(employeeID),
					employee.TenantIDEQ(tenantID),
					employee.EmploymentStatusEQ(employee.EmploymentStatusACTIVE),
				).
				Only(ctx)

			if err != nil {
				if ent.IsNotFound(err) {
					return fmt.Errorf("employee not found or not active")
				}
				return fmt.Errorf("failed to fetch employee: %w", err)
			}

			// Verify position exists and is available
			pos, err := tx.Position.Query().
				Where(
					position.IDEQ(req.PositionID),
					position.TenantIDEQ(tenantID),
					position.StatusIn(position.StatusOPEN, position.StatusFILLED),
				).
				Only(ctx)

			if err != nil {
				if ent.IsNotFound(err) {
					return fmt.Errorf("position not found or not available")
				}
				return fmt.Errorf("failed to fetch position: %w", err)
			}

			// Check for existing active assignment
			activeAssignment, err := tx.PositionOccupancyHistory.Query().
				Where(
					positionoccupancyhistory.EmployeeIDEQ(employeeID),
					positionoccupancyhistory.TenantIDEQ(tenantID),
					positionoccupancyhistory.IsActiveEQ(true),
				).
				First(ctx)

			if err == nil && activeAssignment != nil {
				// End current assignment
				_, err = tx.PositionOccupancyHistory.UpdateOne(activeAssignment).
					SetEndDate(req.StartDate).
					SetIsActive(false).
					Save(ctx)

				if err != nil {
					return fmt.Errorf("failed to end current assignment: %w", err)
				}
			}

			// Create new occupancy history record
			_, err = tx.PositionOccupancyHistory.Create().
				SetTenantID(tenantID).
				SetPositionID(req.PositionID).
				SetEmployeeID(employeeID).
				SetStartDate(req.StartDate).
				SetIsActive(true).
				SetAssignmentType(positionoccupancyhistory.AssignmentType(req.AssignmentType)).
				SetAssignmentReason(req.AssignmentReason).
				SetFtePercentage(req.FTEPercentage).
				SetApprovedBy(req.ApprovedBy).
				SetApprovalDate(time.Now()).
				Save(ctx)

			if err != nil {
				return fmt.Errorf("failed to create occupancy history: %w", err)
			}

			// Update employee current position
			_, err = tx.Employee.UpdateOne(emp).
				SetCurrentPositionID(req.PositionID).
				Save(ctx)

			if err != nil {
				return fmt.Errorf("failed to update employee current position: %w", err)
			}

			// Update position status to FILLED
			_, err = tx.Position.UpdateOne(pos).
				SetStatus(position.StatusFILLED).
				Save(ctx)

			return err
		}()

		if err != nil {
			tx.Rollback()
			h.logger.LogError("assign_position", "Position assignment failed", err, map[string]interface{}{
				"employee_id": employeeID,
				"position_id": req.PositionID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Position assignment failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			h.logger.LogError("assign_position", "Failed to commit transaction", err, map[string]interface{}{
				"employee_id": employeeID,
				"position_id": req.PositionID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to commit position assignment", http.StatusInternalServerError)
			return
		}

		h.logger.Info("Position assigned successfully",
			"employee_id", employeeID,
			"position_id", req.PositionID,
			"assignment_type", req.AssignmentType,
			"tenant_id", tenantID,
		)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Position assigned successfully",
		})
	}
}

// GetPositionHistory handles GET /api/v1/employees/{id}/position-history
func (h *EmployeeHandler) GetPositionHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("get_position_history", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		employeeIDStr := chi.URLParam(r, "id")
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		// Fetch position history
		history, err := h.client.PositionOccupancyHistory.Query().
			Where(
				positionoccupancyhistory.EmployeeIDEQ(employeeID),
				positionoccupancyhistory.TenantIDEQ(tenantID),
			).
			WithPosition().
			Order(positionoccupancyhistory.ByStartDate()).
			All(ctx)

		if err != nil {
			h.logger.LogError("get_position_history", "Failed to fetch position history", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID,
			})
			http.Error(w, "Failed to fetch position history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"employee_id": employeeID,
			"history":     history,
		})
	}
}

// GetPotentialManagers handles GET /api/v1/employees/potential-managers
func (h *EmployeeHandler) GetPotentialManagers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
		if !ok {
			h.logger.LogError("get_potential_managers", "No tenant ID in context", nil, nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		departmentId := r.URL.Query().Get("departmentId")
		if departmentId == "" {
			h.logger.Info("No department specified, returning empty managers list",
				"tenant_id", tenantID,
			)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []EmployeeResponse{},
			})
			return
		}

		// First, find the department by name to get its ID
		department, err := h.client.OrganizationUnit.Query().
			Where(
				organizationunit.NameEQ(departmentId),
				organizationunit.TenantIDEQ(tenantID),
			).
			Only(ctx)

		if err != nil {
			if ent.IsNotFound(err) {
				h.logger.Info("Department not found, returning empty managers list",
					"department", departmentId,
					"tenant_id", tenantID,
				)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []EmployeeResponse{},
				})
				return
			}
			h.logger.LogError("get_potential_managers", "Failed to find department", err, map[string]interface{}{
				"department": departmentId,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to find department", http.StatusInternalServerError)
			return
		}

		// Query for active employees in the same department
		// In a real system, you might want more sophisticated logic to determine who can be a manager
		managers, err := h.client.Employee.Query().
			Where(
				employee.TenantIDEQ(tenantID),
				employee.EmploymentStatusEQ(employee.EmploymentStatusACTIVE),
			).
			WithCurrentPosition(func(q *ent.PositionQuery) {
				q.Where(position.DepartmentIDEQ(department.ID))
			}).
			All(ctx)

		if err != nil {
			h.logger.LogError("get_potential_managers", "Failed to fetch potential managers", err, map[string]interface{}{
				"department": departmentId,
				"department_id": department.ID,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to fetch potential managers", http.StatusInternalServerError)
			return
		}

		// Convert to response format
		responses := make([]EmployeeResponse, 0, len(managers))
		for _, emp := range managers {
			// Only include employees who have a position in the specified department
			if emp.Edges.CurrentPosition != nil {
				responses = append(responses, h.convertToResponse(emp, emp.Edges.CurrentPosition))
			}
		}

		h.logger.Info("Fetched potential managers from database",
			"department", departmentId,
			"department_id", department.ID,
			"count", len(responses),
			"tenant_id", tenantID,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": responses,
		})
	}
}

// generateMockManagers returns mock managers for testing
func generateMockManagers(departmentId string) []EmployeeResponse {
	allManagers := []EmployeeResponse{
		{
			ID:               uuid.New(),
			PersonName:       "张经理",
			Email:            "zhang.manager@company.com",
			EmploymentStatus: "ACTIVE",
			HireDate:         time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               uuid.New(),
			PersonName:       "李总监",
			Email:            "li.director@company.com",
			EmploymentStatus: "ACTIVE",
			HireDate:         time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:               uuid.New(),
			PersonName:       "王主管",
			Email:            "wang.supervisor@company.com",
			EmploymentStatus: "ACTIVE",
			HireDate:         time.Date(2021, 3, 10, 0, 0, 0, 0, time.UTC),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	// For simplicity, return all managers regardless of department for now
	return allManagers
}
func (h *EmployeeHandler) convertToResponse(emp *ent.Employee, pos *ent.Position) EmployeeResponse {
	response := EmployeeResponse{
		ID:               emp.ID,
		TenantID:         emp.TenantID,
		EmployeeType:     string(emp.EmployeeType),
		EmployeeNumber:   emp.EmployeeNumber,
		PersonName:       strings.TrimSpace(emp.FirstName + " " + emp.LastName), // 组合为 PersonName
		Email:            emp.Email,
		PersonalEmail:    emp.PersonalEmail,
		PhoneNumber:      emp.PhoneNumber,
		CurrentPositionID: emp.CurrentPositionID,
		EmploymentStatus: string(emp.EmploymentStatus),
		HireDate:         emp.HireDate,
		TerminationDate:  emp.TerminationDate,
		EmployeeDetails:  emp.EmployeeDetails,
		CreatedAt:        emp.CreatedAt,
		UpdatedAt:        emp.UpdatedAt,
	}

	if pos != nil {
		response.CurrentPosition = &PositionSummary{
			ID:           pos.ID,
			PositionType: string(pos.PositionType),
			DepartmentID: pos.DepartmentID,
			Status:       string(pos.Status),
		}
	}

	return response
}