package types

import (
	"testing"
)

func TestUnitType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected UnitType
		hasError bool
	}{
		{"Valid COMPANY", "COMPANY", UnitTypeCompany, false},
		{"Valid DEPARTMENT", "DEPARTMENT", UnitTypeDepartment, false},
		{"Valid TEAM", "TEAM", UnitTypeTeam, false},
		{"Valid COST_CENTER", "COST_CENTER", UnitTypeCostCenter, false},
		{"Valid PROJECT_TEAM", "PROJECT_TEAM", UnitTypeProjectTeam, false},
		{"Valid lowercase", "department", UnitTypeDepartment, false},
		{"Valid with spaces", "  COMPANY  ", UnitTypeCompany, false},
		{"Invalid type", "INVALID_TYPE", UnitTypeUnknown, true},
		{"Empty string", "", UnitTypeUnknown, true},
		{"Random string", "RANDOM", UnitTypeUnknown, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseUnitType(test.input)
			
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", test.input)
				}
				if result != UnitTypeUnknown {
					t.Errorf("Expected UnitTypeUnknown for invalid input, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("Expected %v for input %q, got %v", test.expected, test.input, result)
				}
			}
		})
	}
}

func TestUnitTypeIsValid(t *testing.T) {
	tests := []struct {
		unitType UnitType
		expected bool
	}{
		{UnitTypeCompany, true},
		{UnitTypeDepartment, true},
		{UnitTypeTeam, true},
		{UnitTypeCostCenter, true},
		{UnitTypeProjectTeam, true},
		{UnitTypeUnknown, false},
		{UnitType(999), false}, // Invalid value
	}

	for _, test := range tests {
		t.Run(test.unitType.String(), func(t *testing.T) {
			result := test.unitType.IsValid()
			if result != test.expected {
				t.Errorf("Expected IsValid() = %v for %v, got %v", test.expected, test.unitType, result)
			}
		})
	}
}

func TestUnitTypeToAPIString(t *testing.T) {
	tests := []struct {
		unitType UnitType
		expected string
	}{
		{UnitTypeCompany, "COMPANY"},
		{UnitTypeDepartment, "DEPARTMENT"},
		{UnitTypeTeam, "TEAM"},
		{UnitTypeCostCenter, "COST_CENTER"},
		{UnitTypeProjectTeam, "PROJECT_TEAM"},
		{UnitTypeUnknown, "UNKNOWN"},
		{UnitType(999), "UNKNOWN"}, // Invalid value
	}

	for _, test := range tests {
		t.Run(test.unitType.String(), func(t *testing.T) {
			result := test.unitType.ToAPIString()
			if result != test.expected {
				t.Errorf("Expected ToAPIString() = %q for %v, got %q", test.expected, test.unitType, result)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Status
		hasError bool
	}{
		{"Valid ACTIVE", "ACTIVE", StatusActive, false},
		{"Valid INACTIVE", "INACTIVE", StatusInactive, false},
		{"Valid PLANNED", "PLANNED", StatusPlanned, false},
		{"Valid lowercase", "active", StatusActive, false},
		{"Valid with spaces", "  ACTIVE  ", StatusActive, false},
		{"Invalid status", "INVALID_STATUS", StatusUnknown, true},
		{"Empty string", "", StatusUnknown, true},
		{"Random string", "RANDOM", StatusUnknown, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ParseStatus(test.input)
			
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", test.input)
				}
				if result != StatusUnknown {
					t.Errorf("Expected StatusUnknown for invalid input, got %v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("Expected %v for input %q, got %v", test.expected, test.input, result)
				}
			}
		})
	}
}

func TestStatusIsValid(t *testing.T) {
	tests := []struct {
		status   Status
		expected bool
	}{
		{StatusActive, true},
		{StatusInactive, true},
		{StatusPlanned, true},
		{StatusUnknown, false},
		{Status(999), false}, // Invalid value
	}

	for _, test := range tests {
		t.Run(test.status.String(), func(t *testing.T) {
			result := test.status.IsValid()
			if result != test.expected {
				t.Errorf("Expected IsValid() = %v for %v, got %v", test.expected, test.status, result)
			}
		})
	}
}

func TestOrganizationCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{"Valid code", "1000001", false},
		{"Valid code at lower bound", "1000000", false},
		{"Valid code at upper bound", "9999999", false},
		{"Valid code in middle", "5555555", false},
		{"Too short", "123456", true},
		{"Too long", "12345678", true},
		{"Contains letters", "1000a01", true},
		{"Contains special chars", "1000-01", true},
		{"Below range", "0999999", true},
		{"Empty string", "", true},
		{"All zeros", "0000000", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewOrganizationCode(test.input)
			
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", test.input, err)
				}
				if result.String() != test.input {
					t.Errorf("Expected code string %q, got %q", test.input, result.String())
				}
				if result.IsEmpty() {
					t.Errorf("Valid code should not be empty")
				}
			}
		})
	}
}

func TestOrganizationCodeEqual(t *testing.T) {
	code1, _ := NewOrganizationCode("1000001")
	code2, _ := NewOrganizationCode("1000001")
	code3, _ := NewOrganizationCode("1000002")

	if !code1.Equal(code2) {
		t.Error("Same codes should be equal")
	}

	if code1.Equal(code3) {
		t.Error("Different codes should not be equal")
	}
}

func TestTenantID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{"Valid UUID", "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9", false},
		{"Valid string ID", "tenant123", false},
		{"Empty string", "", true},
		{"Only spaces", "   ", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewTenantID(test.input)
			
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", test.input, err)
				}
				if result.String() != test.input {
					t.Errorf("Expected tenant ID string %q, got %q", test.input, result.String())
				}
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	t.Run("Empty validation errors", func(t *testing.T) {
		ve := NewValidationErrors()
		
		if ve.HasErrors() {
			t.Error("New validation errors should not have errors")
		}
		
		if len(ve.Errors()) != 0 {
			t.Error("New validation errors should have empty error slice")
		}
		
		expected := "no validation errors"
		if ve.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, ve.Error())
		}
	})

	t.Run("Validation errors with content", func(t *testing.T) {
		ve := NewValidationErrors()
		ve.AddError("name", "Name is required", "required")
		ve.AddError("code", "Invalid code format", "format")
		
		if !ve.HasErrors() {
			t.Error("Should have errors after adding them")
		}
		
		errors := ve.Errors()
		if len(errors) != 2 {
			t.Errorf("Expected 2 errors, got %d", len(errors))
		}
		
		if errors[0].Field != "name" {
			t.Errorf("Expected first error field to be 'name', got %q", errors[0].Field)
		}
		
		if errors[0].Message != "Name is required" {
			t.Errorf("Expected first error message to be 'Name is required', got %q", errors[0].Message)
		}
		
		if errors[0].Code != "required" {
			t.Errorf("Expected first error code to be 'required', got %q", errors[0].Code)
		}
		
		errorMsg := ve.Error()
		expectedMsg := "validation failed: name: Name is required, code: Invalid code format"
		if errorMsg != expectedMsg {
			t.Errorf("Expected error message %q, got %q", expectedMsg, errorMsg)
		}
	})
}

// Benchmark tests
func BenchmarkParseUnitType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseUnitType("DEPARTMENT")
	}
}

func BenchmarkParseStatus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseStatus("ACTIVE")
	}
}

func BenchmarkNewOrganizationCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewOrganizationCode("1000001")
	}
}

func BenchmarkOrganizationCodeString(b *testing.B) {
	code, _ := NewOrganizationCode("1000001")
	for i := 0; i < b.N; i++ {
		_ = code.String()
	}
}