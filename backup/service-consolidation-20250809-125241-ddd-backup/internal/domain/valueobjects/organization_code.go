package valueobjects

import (
	"fmt"
	"strconv"
	"strings"
)

// OrganizationCode represents a valid organization code value object
type OrganizationCode struct {
	value string
}

// NewOrganizationCode creates a new organization code with validation
func NewOrganizationCode(value string) (OrganizationCode, error) {
	if !isValidOrganizationCode(value) {
		return OrganizationCode{}, ErrInvalidOrganizationCode
	}
	
	return OrganizationCode{value: strings.TrimSpace(value)}, nil
}

// String returns the string representation of the organization code
func (c OrganizationCode) String() string {
	return c.value
}

// IsEmpty returns true if the code is empty
func (c OrganizationCode) IsEmpty() bool {
	return c.value == ""
}

// Equals compares two organization codes for equality
func (c OrganizationCode) Equals(other OrganizationCode) bool {
	return c.value == other.value
}

// isValidOrganizationCode validates the organization code format
func isValidOrganizationCode(code string) bool {
	code = strings.TrimSpace(code)
	
	// Allow empty codes for auto-generation
	if code == "" {
		return true
	}
	
	// Must be exactly 7 digits
	if len(code) != 7 {
		return false
	}
	
	// Must be all numeric
	if _, err := strconv.Atoi(code); err != nil {
		return false
	}
	
	// Must be in valid range (1000000 to 9999999)
	codeInt, _ := strconv.Atoi(code)
	return codeInt >= 1000000 && codeInt <= 9999999
}

// EmptyOrganizationCode returns an empty organization code
func EmptyOrganizationCode() OrganizationCode {
	return OrganizationCode{value: ""}
}

// Errors
var (
	ErrInvalidOrganizationCode = fmt.Errorf("invalid organization code format")
)