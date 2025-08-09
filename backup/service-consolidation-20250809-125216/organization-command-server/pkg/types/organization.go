package types

import (
	"fmt"
	"strings"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=UnitType,Status

// UnitType 组织单元类型枚举
type UnitType int

const (
	UnitTypeUnknown UnitType = iota
	UnitTypeCompany
	UnitTypeDepartment
	UnitTypeTeam
	UnitTypeCostCenter
	UnitTypeProjectTeam
)

// IsValid 检查UnitType是否有效
func (ut UnitType) IsValid() bool {
	return ut >= UnitTypeCompany && ut <= UnitTypeProjectTeam
}

// ToAPIString 转换为API字符串格式
func (ut UnitType) ToAPIString() string {
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

// ParseUnitType 从字符串解析UnitType
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
	case "":
		return UnitTypeUnknown, fmt.Errorf("empty unit type")
	default:
		return UnitTypeUnknown, fmt.Errorf("invalid unit type: %s", s)
	}
}

// MustParseUnitType 从字符串解析UnitType，失败时panic
func MustParseUnitType(s string) UnitType {
	ut, err := ParseUnitType(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse unit type: %v", err))
	}
	return ut
}

// Status 组织单元状态枚举
type Status int

const (
	StatusUnknown Status = iota
	StatusActive
	StatusInactive
	StatusPlanned
)

// IsValid 检查Status是否有效
func (s Status) IsValid() bool {
	return s >= StatusActive && s <= StatusPlanned
}

// ToAPIString 转换为API字符串格式
func (s Status) ToAPIString() string {
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

// ParseStatus 从字符串解析Status
func ParseStatus(s string) (Status, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "ACTIVE":
		return StatusActive, nil
	case "INACTIVE":
		return StatusInactive, nil
	case "PLANNED":
		return StatusPlanned, nil
	case "":
		return StatusUnknown, fmt.Errorf("empty status")
	default:
		return StatusUnknown, fmt.Errorf("invalid status: %s", s)
	}
}

// MustParseStatus 从字符串解析Status，失败时panic
func MustParseStatus(s string) Status {
	st, err := ParseStatus(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse status: %v", err))
	}
	return st
}

// OrganizationCode 组织代码值对象
type OrganizationCode struct {
	value string
}

// NewOrganizationCode 创建新的组织代码
func NewOrganizationCode(value string) (OrganizationCode, error) {
	if !isValidOrganizationCode(value) {
		return OrganizationCode{}, fmt.Errorf("invalid organization code: %s (must be 7 digits)", value)
	}
	
	return OrganizationCode{value: value}, nil
}

// MustNewOrganizationCode 创建新的组织代码，失败时panic
func MustNewOrganizationCode(value string) OrganizationCode {
	code, err := NewOrganizationCode(value)
	if err != nil {
		panic(fmt.Sprintf("failed to create organization code: %v", err))
	}
	return code
}

// String 返回组织代码字符串值
func (c OrganizationCode) String() string {
	return c.value
}

// IsEmpty 检查组织代码是否为空
func (c OrganizationCode) IsEmpty() bool {
	return c.value == ""
}

// Equal 检查两个组织代码是否相等
func (c OrganizationCode) Equal(other OrganizationCode) bool {
	return c.value == other.value
}

// isValidOrganizationCode 验证组织代码格式
func isValidOrganizationCode(code string) bool {
	if len(code) != 7 {
		return false
	}
	
	// 检查是否都是数字
	for _, char := range code {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	// 检查范围（1000000-9999999）
	return code >= "1000000" && code <= "9999999"
}

// TenantID 租户ID值对象
type TenantID struct {
	value string
}

// NewTenantID 创建新的租户ID
func NewTenantID(value string) (TenantID, error) {
	if strings.TrimSpace(value) == "" {
		return TenantID{}, fmt.Errorf("tenant ID cannot be empty")
	}
	
	return TenantID{value: value}, nil
}

// String 返回租户ID字符串值
func (t TenantID) String() string {
	return t.value
}

// ValidationErrors 验证错误集合
type ValidationErrors struct {
	errors []ValidationError
}

// ValidationError 单个验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// NewValidationErrors 创建新的验证错误集合
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		errors: make([]ValidationError, 0),
	}
}

// AddError 添加验证错误
func (ve *ValidationErrors) AddError(field, message, code string) {
	ve.errors = append(ve.errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// HasErrors 检查是否有验证错误
func (ve *ValidationErrors) HasErrors() bool {
	return len(ve.errors) > 0
}

// Errors 获取所有错误
func (ve *ValidationErrors) Errors() []ValidationError {
	return ve.errors
}

// Error 实现error接口
func (ve *ValidationErrors) Error() string {
	if len(ve.errors) == 0 {
		return "no validation errors"
	}
	
	var messages []string
	for _, err := range ve.errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	
	return fmt.Sprintf("validation failed: %s", strings.Join(messages, ", "))
}