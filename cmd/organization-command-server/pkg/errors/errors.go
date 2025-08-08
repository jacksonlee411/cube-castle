package errors

import "fmt"

// DomainError represents domain-specific errors
type DomainError struct {
	Code    string
	Message string
	Details string
}

func (e DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ValidationError represents validation-specific errors with detailed information
type ValidationError struct {
	Message string
	Details interface{} // Can be validation errors from types package
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Validation failed: %s", e.Message)
}

// BadRequestError represents bad request errors
type BadRequestError struct {
	Message string
	Details interface{}
}

func (e *BadRequestError) Error() string {
	return fmt.Sprintf("Bad request: %s", e.Message)
}

// Error constructors
func NewNotFoundError(code, message string) DomainError {
	return DomainError{Code: code, Message: message}
}

func NewConflictError(code, message string) DomainError {
	return DomainError{Code: code, Message: message}
}

func NewValidationError(message string, details interface{}) *ValidationError {
	return &ValidationError{Message: message, Details: details}
}

func NewBadRequestError(message string, details interface{}) *BadRequestError {
	return &BadRequestError{Message: message, Details: details}
}

func NewBusinessRuleError(code, message string) DomainError {
	return DomainError{Code: code, Message: message}
}

func NewInternalServerError(code, message string) DomainError {
	return DomainError{Code: code, Message: message}
}

// Predefined errors
var (
	ErrOrganizationNotFound          = NewNotFoundError("ORG_001", "organization not found")
	ErrOrganizationCodeAlreadyExists = NewConflictError("ORG_002", "organization code already exists")
	ErrCannotDeleteWithChildren      = NewBusinessRuleError("ORG_003", "cannot delete organization with children")
	ErrInvalidOrganizationCode       = NewValidationError("ORG_004", "invalid organization code format")
)