package utils

import (
	"testing"

	"cube-castle/internal/types"
)

func TestValidateOrganizationCode(t *testing.T) {
	if err := ValidateOrganizationCode("1000000"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ValidateOrganizationCode(""); err == nil {
		t.Fatalf("expected error for empty code")
	}
	if err := ValidateOrganizationCode("0123456"); err == nil {
		t.Fatalf("expected error for leading zero")
	}
	if err := ValidateOrganizationCode("ABC"); err == nil {
		t.Fatalf("expected error for non-numeric")
	}
}

func TestValidateCreateVersionRequest_Valid(t *testing.T) {
	parent := "1000000"
	desc := "描述"
	sort := 1
	req := &types.CreateVersionRequest{
		Name:            "人力资源部",
		UnitType:        string(types.UnitTypeDepartment),
		ParentCode:      &parent,
		Description:     &desc,
		SortOrder:       &sort,
		EffectiveDate:   "2025-11-15",
		EndDate:         nil,
		OperationReason: "业务需要",
	}
	if err := ValidateCreateVersionRequest(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ParentCode should remain normalized (non-nil)
	if req.ParentCode == nil || *req.ParentCode != "1000000" {
		t.Fatalf("unexpected parentCode normalization: %#v", req.ParentCode)
	}
}

func TestValidateCreateVersionRequest_Errors(t *testing.T) {
	// name empty
	{
		req := &types.CreateVersionRequest{
			Name:          "",
			UnitType:      string(types.UnitTypeDepartment),
			EffectiveDate: "2025-11-15",
		}
		if err := ValidateCreateVersionRequest(req); err == nil {
			t.Fatalf("expected error for empty name")
		}
	}
	// invalid unit type
	{
		req := &types.CreateVersionRequest{
			Name:          "部门",
			UnitType:      "INVALID",
			EffectiveDate: "2025-11-15",
		}
		if err := ValidateCreateVersionRequest(req); err == nil {
			t.Fatalf("expected error for invalid unit type")
		}
	}
	// parent code normalization to nil for root placeholders
	{
		root := "0"
		req := &types.CreateVersionRequest{
			Name:          "部门",
			UnitType:      string(types.UnitTypeDepartment),
			ParentCode:    &root,
			EffectiveDate: "2025-11-15",
		}
		if err := ValidateCreateVersionRequest(req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.ParentCode != nil {
			t.Fatalf("expected parent code normalized to nil for root, got %v", *req.ParentCode)
		}
	}
	// invalid parent code (non-root but wrong format)
	{
		pc := "abc"
		req := &types.CreateVersionRequest{
			Name:          "部门",
			UnitType:      string(types.UnitTypeDepartment),
			ParentCode:    &pc,
			EffectiveDate: "2025-11-15",
		}
		if err := ValidateCreateVersionRequest(req); err == nil {
			t.Fatalf("expected error for invalid parent code")
		}
	}
	// invalid effective date format
	{
		req := &types.CreateVersionRequest{
			Name:          "部门",
			UnitType:      string(types.UnitTypeDepartment),
			EffectiveDate: "15-11-2025",
		}
		if err := ValidateCreateVersionRequest(req); err == nil {
			t.Fatalf("expected error for invalid effective date format")
		}
	}
	// end date not after effective
	{
		end := "2025-11-01"
		req := &types.CreateVersionRequest{
			Name:          "部门",
			UnitType:      string(types.UnitTypeDepartment),
			EffectiveDate: "2025-11-15",
			EndDate:       &end,
		}
		if err := ValidateCreateVersionRequest(req); err == nil {
			t.Fatalf("expected error for end date before effective date")
		}
	}
	// short operation reason
	{
		req := &types.CreateVersionRequest{
			Name:            "部门",
			UnitType:        string(types.UnitTypeDepartment),
			EffectiveDate:   "2025-11-15",
			OperationReason: "abc",
		}
		if err := ValidateCreateVersionRequest(req); err == nil {
			t.Fatalf("expected error for short operation reason")
		}
	}
}
