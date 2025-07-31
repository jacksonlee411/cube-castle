package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/google/uuid"
)

// EmployeeValidator provides validation for employee-related operations
type EmployeeValidator struct {
	// Dependencies for external validation
	employeeNumberChecker EmployeeNumberChecker
	emailChecker          EmailChecker
	organizationChecker   OrganizationChecker
	positionChecker       PositionChecker
}

// External dependency interfaces
type EmployeeNumberChecker interface {
	IsEmployeeNumberExists(ctx context.Context, tenantID uuid.UUID, employeeNumber string, excludeID *uuid.UUID) (bool, error)
}

type EmailChecker interface {
	IsEmailExists(ctx context.Context, tenantID uuid.UUID, email string, excludeID *uuid.UUID) (bool, error)
}

type OrganizationChecker interface {
	IsOrganizationExists(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID) (bool, error)
}

type PositionChecker interface {
	IsPositionExists(ctx context.Context, tenantID uuid.UUID, positionID uuid.UUID) (bool, error)
}

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Value   string `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("Validation failed for field '%s': %s (code: %s)", e.Field, e.Message, e.Code)
	}
	return fmt.Sprintf("Validation failed for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (ve ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "no validation errors"
	}
	if len(ve.Errors) == 1 {
		return ve.Errors[0].Error()
	}
	
	// For multiple errors, include a summary with individual error details
	errorDetails := make([]string, len(ve.Errors))
	for i, err := range ve.Errors {
		errorDetails[i] = err.Error()
	}
	return fmt.Sprintf("multiple validation errors: %d issues found: %s", len(ve.Errors), strings.Join(errorDetails, "; "))
}

func (ve *ValidationErrors) Add(field, message, code string, value ...string) {
	err := ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	}
	if len(value) > 0 {
		err.Value = value[0]
	}
	ve.Errors = append(ve.Errors, err)
}

func (ve ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// NewEmployeeValidator creates a new employee validator
func NewEmployeeValidator(
	employeeNumberChecker EmployeeNumberChecker,
	emailChecker EmailChecker,
	organizationChecker OrganizationChecker,
	positionChecker PositionChecker,
) *EmployeeValidator {
	return &EmployeeValidator{
		employeeNumberChecker: employeeNumberChecker,
		emailChecker:          emailChecker,
		organizationChecker:   organizationChecker,
		positionChecker:       positionChecker,
	}
}

// ValidateCreateEmployee validates a create employee request
func (v *EmployeeValidator) ValidateCreateEmployee(ctx context.Context, tenantID uuid.UUID, req *openapi.CreateEmployeeRequest) error {
	var errors ValidationErrors

	// Validate required fields
	v.validateEmployeeNumber(req.EmployeeNumber, &errors)
	v.validateFirstName(req.FirstName, &errors)
	v.validateLastName(req.LastName, &errors)
	v.validateEmail(string(req.Email), &errors)
	v.validateHireDate(req.HireDate.Time, &errors)

	// Validate optional fields
	if req.PhoneNumber != nil {
		v.validatePhoneNumber(*req.PhoneNumber, &errors)
	}

	// Check for duplicates
	if v.employeeNumberChecker != nil && req.EmployeeNumber != "" {
		exists, err := v.employeeNumberChecker.IsEmployeeNumberExists(ctx, tenantID, req.EmployeeNumber, nil)
		if err != nil {
			errors.Add("employee_number", "Failed to check employee number uniqueness", "CHECK_FAILED", req.EmployeeNumber)
		} else if exists {
			errors.Add("employee_number", "Employee number already exists", "DUPLICATE", req.EmployeeNumber)
		}
	}

	if v.emailChecker != nil && req.Email != "" {
		exists, err := v.emailChecker.IsEmailExists(ctx, tenantID, string(req.Email), nil)
		if err != nil {
			errors.Add("email", "Failed to check email uniqueness", "CHECK_FAILED", string(req.Email))
		} else if exists {
			errors.Add("email", "Email address already exists", "DUPLICATE", string(req.Email))
		}
	}

	// Validate foreign key references
	if req.OrganizationId != nil && v.organizationChecker != nil {
		exists, err := v.organizationChecker.IsOrganizationExists(ctx, tenantID, *req.OrganizationId)
		if err != nil {
			errors.Add("organization_id", "Failed to validate organization", "CHECK_FAILED", req.OrganizationId.String())
		} else if !exists {
			errors.Add("organization_id", "Organization does not exist", "NOT_FOUND", req.OrganizationId.String())
		}
	}

	if req.PositionId != nil && v.positionChecker != nil {
		exists, err := v.positionChecker.IsPositionExists(ctx, tenantID, *req.PositionId)
		if err != nil {
			errors.Add("position_id", "Failed to validate position", "CHECK_FAILED", req.PositionId.String())
		} else if !exists {
			errors.Add("position_id", "Position does not exist", "NOT_FOUND", req.PositionId.String())
		}
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// ValidateUpdateEmployee validates an update employee request
func (v *EmployeeValidator) ValidateUpdateEmployee(ctx context.Context, tenantID uuid.UUID, employeeID uuid.UUID, req *openapi.UpdateEmployeeRequest) error {
	var errors ValidationErrors

	// Validate fields that are being updated
	if req.FirstName != nil {
		v.validateFirstName(*req.FirstName, &errors)
	}

	if req.LastName != nil {
		v.validateLastName(*req.LastName, &errors)
	}

	if req.Email != nil {
		v.validateEmail(string(*req.Email), &errors)
		
		// Check email uniqueness (excluding current employee)
		if v.emailChecker != nil {
			exists, err := v.emailChecker.IsEmailExists(ctx, tenantID, string(*req.Email), &employeeID)
			if err != nil {
				errors.Add("email", "Failed to check email uniqueness", "CHECK_FAILED", string(*req.Email))
			} else if exists {
				errors.Add("email", "Email address already exists", "DUPLICATE", string(*req.Email))
			}
		}
	}

	if req.PhoneNumber != nil {
		v.validatePhoneNumber(*req.PhoneNumber, &errors)
	}

	if req.Status != nil {
		v.validateEmployeeStatus(string(*req.Status), &errors)
	}

	// Validate foreign key references
	if req.OrganizationId != nil && v.organizationChecker != nil {
		exists, err := v.organizationChecker.IsOrganizationExists(ctx, tenantID, *req.OrganizationId)
		if err != nil {
			errors.Add("organization_id", "Failed to validate organization", "CHECK_FAILED", req.OrganizationId.String())
		} else if !exists {
			errors.Add("organization_id", "Organization does not exist", "NOT_FOUND", req.OrganizationId.String())
		}
	}

	if req.PositionId != nil && v.positionChecker != nil {
		exists, err := v.positionChecker.IsPositionExists(ctx, tenantID, *req.PositionId)
		if err != nil {
			errors.Add("position_id", "Failed to validate position", "CHECK_FAILED", req.PositionId.String())
		} else if !exists {
			errors.Add("position_id", "Position does not exist", "NOT_FOUND", req.PositionId.String())
		}
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// Individual field validation methods

func (v *EmployeeValidator) validateEmployeeNumber(employeeNumber string, errors *ValidationErrors) {
	if employeeNumber == "" {
		errors.Add("employee_number", "Employee number is required", "REQUIRED")
		return
	}

	// Length validation
	if len(employeeNumber) < 3 {
		errors.Add("employee_number", "Employee number must be at least 3 characters long", "TOO_SHORT", employeeNumber)
		return
	}
	if len(employeeNumber) > 20 {
		errors.Add("employee_number", "Employee number must not exceed 20 characters", "TOO_LONG", employeeNumber)
		return
	}

	// Format validation - alphanumeric with optional hyphens and underscores
	matched, _ := regexp.MatchString(`^[A-Za-z0-9_-]+$`, employeeNumber)
	if !matched {
		errors.Add("employee_number", "Employee number can only contain letters, numbers, hyphens, and underscores", "INVALID_FORMAT", employeeNumber)
	}
}

func (v *EmployeeValidator) validateFirstName(firstName string, errors *ValidationErrors) {
	if firstName == "" {
		errors.Add("first_name", "First name is required", "REQUIRED")
		return
	}

	// Length validation
	if utf8.RuneCountInString(firstName) < 1 {
		errors.Add("first_name", "First name cannot be empty", "EMPTY", firstName)
		return
	}
	if utf8.RuneCountInString(firstName) > 50 {
		errors.Add("first_name", "First name must not exceed 50 characters", "TOO_LONG", firstName)
		return
	}

	// Content validation - no leading/trailing whitespace
	if strings.TrimSpace(firstName) != firstName {
		errors.Add("first_name", "First name cannot have leading or trailing whitespace", "INVALID_WHITESPACE", firstName)
	}

	// Character validation - only letters, spaces, hyphens, and apostrophes
	matched, _ := regexp.MatchString(`^[A-Za-z\p{Han}\s'-]+$`, firstName)
	if !matched {
		errors.Add("first_name", "First name can only contain letters, spaces, hyphens, and apostrophes", "INVALID_FORMAT", firstName)
	}
}

func (v *EmployeeValidator) validateLastName(lastName string, errors *ValidationErrors) {
	if lastName == "" {
		errors.Add("last_name", "Last name is required", "REQUIRED")
		return
	}

	// Length validation
	if utf8.RuneCountInString(lastName) < 1 {
		errors.Add("last_name", "Last name cannot be empty", "EMPTY", lastName)
		return
	}
	if utf8.RuneCountInString(lastName) > 50 {
		errors.Add("last_name", "Last name must not exceed 50 characters", "TOO_LONG", lastName)
		return
	}

	// Content validation - no leading/trailing whitespace
	if strings.TrimSpace(lastName) != lastName {
		errors.Add("last_name", "Last name cannot have leading or trailing whitespace", "INVALID_WHITESPACE", lastName)
	}

	// Character validation - only letters, spaces, hyphens, and apostrophes
	matched, _ := regexp.MatchString(`^[A-Za-z\p{Han}\s'-]+$`, lastName)
	if !matched {
		errors.Add("last_name", "Last name can only contain letters, spaces, hyphens, and apostrophes", "INVALID_FORMAT", lastName)
	}
}

func (v *EmployeeValidator) validateEmail(email string, errors *ValidationErrors) {
	if email == "" {
		errors.Add("email", "Email is required", "REQUIRED")
		return
	}

	// Length validation
	if len(email) > 254 {
		errors.Add("email", "Email address must not exceed 254 characters", "TOO_LONG", email)
		return
	}

	// Basic email format validation (RFC 5322 compliant)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	if !emailRegex.MatchString(email) {
		errors.Add("email", "Invalid email format", "INVALID_FORMAT", email)
		return
	}

	// Additional validation - no leading/trailing whitespace
	if strings.TrimSpace(email) != email {
		errors.Add("email", "Email cannot have leading or trailing whitespace", "INVALID_WHITESPACE", email)
	}

	// Domain validation - basic checks
	parts := strings.Split(email, "@")
	if len(parts) == 2 && len(parts[1]) > 0 {
		domain := parts[1]
		if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
			errors.Add("email", "Invalid email domain format", "INVALID_DOMAIN", email)
		}
	}
}

func (v *EmployeeValidator) validatePhoneNumber(phoneNumber string, errors *ValidationErrors) {
	if phoneNumber == "" {
		return // Optional field
	}

	// Length validation
	if len(phoneNumber) < 10 {
		errors.Add("phone_number", "Phone number must be at least 10 characters long", "TOO_SHORT", phoneNumber)
		return
	}
	if len(phoneNumber) > 20 {
		errors.Add("phone_number", "Phone number must not exceed 20 characters", "TOO_LONG", phoneNumber)
		return
	}

	// Format validation - international format with optional country code
	phoneRegex := regexp.MustCompile(`^(\+[1-9]\d{0,3})?[-.\s]?(\d{1,4}[-.\s]?){2,4}\d{1,4}$`)
	if !phoneRegex.MatchString(phoneNumber) {
		errors.Add("phone_number", "Invalid phone number format", "INVALID_FORMAT", phoneNumber)
	}
}

func (v *EmployeeValidator) validateHireDate(hireDate time.Time, errors *ValidationErrors) {
	now := time.Now()
	
	// Cannot be in the future
	if hireDate.After(now) {
		errors.Add("hire_date", "Hire date cannot be in the future", "FUTURE_DATE", hireDate.Format("2006-01-02"))
		return
	}

	// Cannot be too far in the past (e.g., more than 100 years ago)
	hundredYearsAgo := now.AddDate(-100, 0, 0)
	if hireDate.Before(hundredYearsAgo) {
		errors.Add("hire_date", "Hire date cannot be more than 100 years ago", "TOO_OLD", hireDate.Format("2006-01-02"))
	}
}

func (v *EmployeeValidator) validateEmployeeStatus(status string, errors *ValidationErrors) {
	validStatuses := map[string]bool{
		"active":     true,
		"inactive":   true,
		"terminated": true,
	}

	if !validStatuses[status] {
		errors.Add("status", "Invalid employee status", "INVALID_VALUE", status)
	}
}

// Business rule validation methods

// ValidateEmployeeTermination validates business rules for employee termination
func (v *EmployeeValidator) ValidateEmployeeTermination(ctx context.Context, employeeID uuid.UUID, tenantID uuid.UUID) error {
	var errors ValidationErrors

	// TODO: Add business rules for termination
	// - Check if employee has active assignments
	// - Check if employee is a manager with direct reports
	// - Check if employee has pending workflows
	// - Validate termination date constraints

	if errors.HasErrors() {
		return errors
	}
	return nil
}

// ValidateEmployeeStatusTransition validates status change business rules
func (v *EmployeeValidator) ValidateEmployeeStatusTransition(currentStatus, newStatus string) error {
	var errors ValidationErrors

	// Define valid status transitions
	validTransitions := map[string][]string{
		"active":     {"inactive", "terminated"},
		"inactive":   {"active", "terminated"},
		"terminated": {}, // No transitions from terminated
	}

	if validTransitions[currentStatus] == nil {
		errors.Add("status", "Invalid current status", "INVALID_CURRENT_STATUS", currentStatus)
		return errors
	}

	valid := false
	for _, validNewStatus := range validTransitions[currentStatus] {
		if newStatus == validNewStatus {
			valid = true
			break
		}
	}

	if !valid {
		errors.Add("status", fmt.Sprintf("Invalid status transition from %s to %s", currentStatus, newStatus), "INVALID_TRANSITION", fmt.Sprintf("%s->%s", currentStatus, newStatus))
		return errors
	}

	return nil
}

// ValidateListEmployeesParams validates list employees query parameters
func (v *EmployeeValidator) ValidateListEmployeesParams(page, pageSize int, search string) error {
	var errors ValidationErrors

	if page < 1 {
		errors.Add("page", "Page number must be greater than 0", "INVALID_VALUE", fmt.Sprintf("%d", page))
	}

	if pageSize < 1 {
		errors.Add("page_size", "Page size must be greater than 0", "INVALID_VALUE", fmt.Sprintf("%d", pageSize))
	} else if pageSize > 100 {
		errors.Add("page_size", "Page size cannot exceed 100", "TOO_LARGE", fmt.Sprintf("%d", pageSize))
	}

	if search != "" && len(search) > 100 {
		errors.Add("search", "Search term cannot exceed 100 characters", "TOO_LONG", search)
	}

	if errors.HasErrors() {
		return errors
	}
	return nil
}