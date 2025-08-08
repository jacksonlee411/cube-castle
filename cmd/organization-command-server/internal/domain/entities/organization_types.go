package entities

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UnitType represents the type of organizational unit
type UnitType int

const (
	UnitTypeUnknown UnitType = iota
	UnitTypeCompany
	UnitTypeDepartment
	UnitTypeTeam
	UnitTypeCostCenter
	UnitTypeProjectTeam
)

func (ut UnitType) String() string {
	switch ut {
	case UnitTypeCompany:
		return "COMPANY"
	case UnitTypeDepartment:
		return "DEPARTMENT"
	case UnitTypeTeam:
		return "TEAM"
	case UnitTypeCostCenter:
		return "COST_CENTER"
	case UnitTypeProjectTeam:
		return "PROJECT_TEAM"
	default:
		return "UNKNOWN"
	}
}

// ParseUnitType parses a string into UnitType
func ParseUnitType(s string) (UnitType, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "COMPANY":
		return UnitTypeCompany, nil
	case "DEPARTMENT":
		return UnitTypeDepartment, nil
	case "TEAM":
		return UnitTypeTeam, nil
	case "COST_CENTER":
		return UnitTypeCostCenter, nil
	case "PROJECT_TEAM":
		return UnitTypeProjectTeam, nil
	default:
		return UnitTypeUnknown, fmt.Errorf("invalid unit type: %s", s)
	}
}

// IsValid checks if the unit type is valid
func (ut UnitType) IsValid() bool {
	return ut >= UnitTypeCompany && ut <= UnitTypeProjectTeam
}

// Status represents the status of an organizational unit
type Status int

const (
	StatusUnknown Status = iota
	StatusActive
	StatusInactive
	StatusPlanned
)

func (s Status) String() string {
	switch s {
	case StatusActive:
		return "ACTIVE"
	case StatusInactive:
		return "INACTIVE"
	case StatusPlanned:
		return "PLANNED"
	default:
		return "UNKNOWN"
	}
}

// ParseStatus parses a string into Status
func ParseStatus(s string) (Status, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "ACTIVE":
		return StatusActive, nil
	case "INACTIVE":
		return StatusInactive, nil
	case "PLANNED":
		return StatusPlanned, nil
	default:
		return StatusUnknown, fmt.Errorf("invalid status: %s", s)
	}
}

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	return s >= StatusActive && s <= StatusPlanned
}

// Domain errors
var (
	ErrEmptyOrganizationName              = fmt.Errorf("organization name cannot be empty")
	ErrOrganizationNameTooLong            = fmt.Errorf("organization name cannot exceed 100 characters")
	ErrCannotDeleteOrganizationWithChildren = fmt.Errorf("cannot delete organization with children")
	ErrInvalidOrganizationLevel           = fmt.Errorf("organization level must be between 1 and 10")
	ErrInvalidSortOrder                   = fmt.Errorf("sort order must be non-negative")
)