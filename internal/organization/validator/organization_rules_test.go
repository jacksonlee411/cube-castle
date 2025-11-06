package validator

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"cube-castle/internal/organization/repository"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

type stubHierarchy struct {
	orgs      map[string]*types.Organization
	depths    map[string]int
	ancestors map[string][]repository.OrganizationNode
	temporal  map[string]map[string]*repository.OrganizationNode
}

func (s *stubHierarchy) GetOrganization(ctx context.Context, code string, tenantID uuid.UUID) (*types.Organization, error) {
	if org, ok := s.orgs[code]; ok {
		return org, nil
	}
	return nil, fmt.Errorf("organization not found: %s", code)
}

func (s *stubHierarchy) GetOrganizationDepth(ctx context.Context, code string, tenantID uuid.UUID) (int, error) {
	if depth, ok := s.depths[code]; ok {
		return depth, nil
	}
	return 0, fmt.Errorf("depth not found for %s", code)
}

func (s *stubHierarchy) GetAncestorChain(ctx context.Context, code string, tenantID uuid.UUID) ([]repository.OrganizationNode, error) {
	if chain, ok := s.ancestors[code]; ok {
		return chain, nil
	}
	return nil, nil
}

func (s *stubHierarchy) GetDirectChildren(ctx context.Context, code string, tenantID uuid.UUID) ([]repository.OrganizationNode, error) {
	return nil, nil
}

func (s *stubHierarchy) GetOrganizationAtDate(ctx context.Context, code string, tenantID uuid.UUID, ts time.Time) (*repository.OrganizationNode, error) {
	if entries, ok := s.temporal[code]; ok {
		if node, ok := entries[ts.Format("2006-01-02")]; ok {
			return node, nil
		}
	}
	return nil, nil
}

type stubOrgRepo struct{}

func (stubOrgRepo) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*types.Organization, error) {
	return nil, fmt.Errorf("not implemented in stub")
}

func newTestValidator(h hierarchyRepository) *BusinessRuleValidator {
	return &BusinessRuleValidator{
		hierarchyRepo: h,
		orgRepo:       stubOrgRepo{},
		logger: pkglogger.NewLogger(
			pkglogger.WithWriter(io.Discard),
			pkglogger.WithLevel(pkglogger.LevelError),
		),
	}
}

func strPtr(s string) *string {
	return &s
}

func TestOrganizationCreateDepthViolation(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"1000002": {Code: "1000002", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"1000002": 17,
		},
	}
	validator := newTestValidator(h)

	now := time.Now().UTC()
	req := &types.CreateOrganizationRequest{
		Name:          "Child",
		UnitType:      "DEPARTMENT",
		ParentCode:    strPtr("1000002"),
		EffectiveDate: types.NewDateFromTime(now),
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, tenant)
	if result.Valid {
		t.Fatalf("expected creation validation to fail")
	}

	if len(result.Errors) == 0 || result.Errors[0].Code != "ORG_DEPTH_LIMIT" {
		t.Fatalf("expected ORG_DEPTH_LIMIT, got %#v", result.Errors)
	}
}

func TestOrganizationCreateDepthWarning(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"1000002": {Code: "1000002", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"1000002": depthWarningThreshold,
		},
	}
	validator := newTestValidator(h)

	req := &types.CreateOrganizationRequest{
		Name:       "Child",
		UnitType:   "DEPARTMENT",
		ParentCode: strPtr("1000002"),
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, tenant)
	if !result.Valid {
		t.Fatalf("expected validation to succeed with warnings: %#v", result.Errors)
	}
	if len(result.Warnings) == 0 || result.Warnings[0].Code != "ORG_DEPTH_NEAR_LIMIT" {
		t.Fatalf("expected ORG_DEPTH_NEAR_LIMIT warning, got %#v", result.Warnings)
	}
}

func TestOrganizationCreateTemporalInactiveParent(t *testing.T) {
	tenant := uuid.New()
	date := time.Date(2025, 11, 5, 0, 0, 0, 0, time.UTC)

	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"1000002": {Code: "1000002", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"1000002": 1,
		},
		temporal: map[string]map[string]*repository.OrganizationNode{
			"1000002": {
				date.Format("2006-01-02"): {
					Code:   "1000002",
					Status: "INACTIVE",
				},
			},
		},
	}
	validator := newTestValidator(h)

	req := &types.CreateOrganizationRequest{
		Name:          "Child",
		UnitType:      "DEPARTMENT",
		ParentCode:    strPtr("1000002"),
		EffectiveDate: types.NewDateFromTime(date),
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, tenant)
	if result.Valid {
		t.Fatalf("expected temporal validation to fail")
	}

	found := false
	for _, errItem := range result.Errors {
		if errItem.Code == "ORG_TEMPORAL_PARENT_INACTIVE" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ORG_TEMPORAL_PARENT_INACTIVE, errors: %#v", result.Errors)
	}
}

func TestOrganizationCreateDuplicateCode(t *testing.T) {
	tenant := uuid.New()
	code := "1000001"
	parent := "1000002"

	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			parent:   {Code: parent, Status: "ACTIVE"},
			code:     {Code: code, Status: "ACTIVE"},
			"UNUSED": {Code: "UNUSED", Status: "ACTIVE"},
		},
		depths: map[string]int{
			parent: 1,
		},
	}
	validator := newTestValidator(h)

	req := &types.CreateOrganizationRequest{
		Code:       strPtr(code),
		Name:       "Duplicated",
		UnitType:   "DEPARTMENT",
		ParentCode: strPtr(parent),
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, tenant)
	if result.Valid {
		t.Fatalf("expected duplicate code validation to fail")
	}

	found := false
	for _, errItem := range result.Errors {
		if errItem.Code == ErrorCodeDuplicateCode {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected duplicate code error, got %#v", result.Errors)
	}
}

func TestOrganizationCreateParentMissing(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{}
	validator := newTestValidator(h)

	req := &types.CreateOrganizationRequest{
		Name:       "Missing Parent",
		UnitType:   "DEPARTMENT",
		ParentCode: strPtr("1000002"),
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, tenant)
	if result.Valid {
		t.Fatalf("expected validation to fail due to missing parent")
	}

	hasInvalidParent := false
	for _, errItem := range result.Errors {
		if errItem.Code == ErrorCodeInvalidParent {
			hasInvalidParent = true
			break
		}
	}
	if !hasInvalidParent {
		t.Fatalf("expected invalid parent error, got %#v", result.Errors)
	}
}

func TestOrganizationCreateSelfReferentialParent(t *testing.T) {
	tenant := uuid.New()
	code := "1200001"

	req := &types.CreateOrganizationRequest{
		Code:       strPtr(code),
		Name:       "Self Loop Org",
		UnitType:   "DEPARTMENT",
		ParentCode: strPtr(code),
	}

	validator := newTestValidator(&stubHierarchy{})

	result := validator.ValidateOrganizationCreation(context.Background(), req, tenant)
	if result.Valid {
		t.Fatalf("expected validation to fail due to self-referential parent")
	}

	assertHasErrorCode(t, result.Errors, "ORG_CYCLE_DETECTED")
}

func TestOrganizationCreateDepthUnknownReturnsInvalidParent(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"1000002": {Code: "1000002", Status: "ACTIVE"},
		},
	}
	validator := newTestValidator(h)

	req := &types.CreateOrganizationRequest{
		Name:       "Child",
		UnitType:   "DEPARTMENT",
		ParentCode: strPtr("1000002"),
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, tenant)
	if result.Valid {
		t.Fatalf("expected validation to fail when parent depth not found")
	}

	found := false
	for _, errItem := range result.Errors {
		if errItem.Code == "INVALID_PARENT" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected INVALID_PARENT error, got %#v", result.Errors)
	}
}

func TestOrganizationCreateBusinessLogicValidation(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{}
	validator := newTestValidator(h)

	req := &types.CreateOrganizationRequest{
		Name:       strings.Repeat("X", types.OrganizationNameMaxLength+1),
		UnitType:   "UNKNOWN",
		SortOrder:  -1,
		ParentCode: nil,
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, tenant)
	if result.Valid {
		t.Fatalf("expected validation to fail due to business logic")
	}

	assertHasErrorCode(t, result.Errors, "ORG_UNIT_TYPE_INVALID")
	assertHasErrorCode(t, result.Errors, "ORG_NAME_TOO_LONG")
	assertHasErrorCode(t, result.Errors, "ORG_SORT_ORDER_INVALID")
}

func TestOrganizationCreateTemporalConflict(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{}
	validator := newTestValidator(h)

	now := time.Now()
	effective := types.NewDateFromTime(now)
	prev := now.Add(-24 * time.Hour)
	endDate := types.NewDateFromTime(prev)

	req := &types.CreateOrganizationRequest{
		Name:          "Temporal Conflict",
		UnitType:      "DEPARTMENT",
		EffectiveDate: effective,
		EndDate:       endDate,
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, tenant)
	if result.Valid {
		t.Fatalf("expected temporal conflict validation to fail")
	}

	found := false
	for _, errItem := range result.Errors {
		if errItem.Code == ErrorCodeTemporalConflict {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected temporal conflict error, got %#v", result.Errors)
	}
}

func TestOrganizationCreateFieldSanitization(t *testing.T) {
	code := " 1000001 "
	parent := " 1000002 "
	unitType := "department"
	name := "  Org  "
	desc := strings.Repeat("a", types.OrganizationDescriptionMaxLength)

	req := &types.CreateOrganizationRequest{
		Code:        &code,
		Name:        name,
		UnitType:    unitType,
		ParentCode:  &parent,
		SortOrder:   20,
		Description: desc,
	}

	validator := newTestValidator(&stubHierarchy{
		orgs: map[string]*types.Organization{
			"1000002": {Code: "1000002", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"1000002": 1,
		},
	})
	result := validator.ValidateOrganizationCreation(context.Background(), req, uuid.New())
	if !result.Valid {
		t.Fatalf("expected sanitized request to pass validation, got errors: %#v", result.Errors)
	}

	if req.Code == nil || *req.Code != "1000001" {
		t.Fatalf("expected trimmed code, got %v", req.Code)
	}
	if req.ParentCode == nil || *req.ParentCode != "1000002" {
		t.Fatalf("expected trimmed parent code, got %v", req.ParentCode)
	}
	if req.UnitType != "DEPARTMENT" {
		t.Fatalf("expected unit type upper-cased, got %s", req.UnitType)
	}
	if req.Name != "Org" {
		t.Fatalf("expected trimmed name, got %q", req.Name)
	}
}

func TestOrganizationCreateNameAllowsLocalizedParentheses(t *testing.T) {
	tenant := uuid.New()
	req := &types.CreateOrganizationRequest{
		Code:       strPtr("1300001"),
		Name:       "测试组织（扩展）",
		UnitType:   "DEPARTMENT",
		ParentCode: nil,
	}

	result := newTestValidator(&stubHierarchy{}).ValidateOrganizationCreation(context.Background(), req, tenant)
	if !result.Valid {
		t.Fatalf("expected localized parentheses to be allowed, got errors %#v", result.Errors)
	}
}

func TestOrganizationUpdateCircularReference(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {
				Code:          "TARGET",
				Status:        "ACTIVE",
				EffectiveDate: types.NewDateFromTime(time.Now()),
			},
		},
		depths: map[string]int{
			"1000002": 2,
		},
		ancestors: map[string][]repository.OrganizationNode{
			"1000002": {
				{Code: "ROOT"},
				{Code: "TARGET"},
			},
		},
	}

	validator := newTestValidator(h)

	req := &types.UpdateOrganizationRequest{
		ParentCode: strPtr("1000002"),
	}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if result.Valid {
		t.Fatalf("expected circular reference validation to fail")
	}

	if len(result.Errors) == 0 || result.Errors[0].Code != "ORG_CYCLE_DETECTED" {
		t.Fatalf("expected ORG_CYCLE_DETECTED, got %#v", result.Errors)
	}
}

func TestOrganizationUpdateStatusAllowed(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {
				Code:          "TARGET",
				Status:        "INACTIVE",
				EffectiveDate: types.NewDateFromTime(time.Now()),
			},
		},
	}

	validator := newTestValidator(h)

	nextStatus := "ACTIVE"
	req := &types.UpdateOrganizationRequest{
		Status: &nextStatus,
	}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if !result.Valid {
		t.Fatalf("expected status transition to pass, errors: %#v", result.Errors)
	}
}

func TestOrganizationUpdateAllowsLocalizedParentheses(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {
				Code:          "TARGET",
				Status:        "ACTIVE",
				EffectiveDate: types.NewDateFromTime(time.Now()),
			},
		},
	}
	validator := newTestValidator(h)

	newName := "219C2B 测试组织（已更新）"
	req := &types.UpdateOrganizationRequest{
		Name: &newName,
	}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if !result.Valid {
		t.Fatalf("expected localized parentheses to be allowed, errors: %#v", result.Errors)
	}
}

func TestValidateTemporalParentAvailability(t *testing.T) {
	tenant := uuid.New()
	date := time.Date(2025, 11, 6, 0, 0, 0, 0, time.UTC)

	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"1000002": {Code: "1000002", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"1000002": 1,
		},
		temporal: map[string]map[string]*repository.OrganizationNode{
			"1000002": {},
		},
	}
	validator := newTestValidator(h)

	result := validator.ValidateTemporalParentAvailability(context.Background(), tenant, "1000002", date)
	if result.Valid {
		t.Fatalf("expected temporal parent availability to fail with inactive parent")
	}

	found := false
	for _, errItem := range result.Errors {
		if errItem.Code == "ORG_TEMPORAL_PARENT_INACTIVE" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ORG_TEMPORAL_PARENT_INACTIVE error, got %#v", result.Errors)
	}
}

func TestValidateTemporalParentAvailabilitySuccess(t *testing.T) {
	tenant := uuid.New()
	date := time.Date(2025, 11, 7, 0, 0, 0, 0, time.UTC)

	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"1000002": {Code: "1000002", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"1000002": 1,
		},
		temporal: map[string]map[string]*repository.OrganizationNode{
			"1000002": {
				date.Format("2006-01-02"): {
					Code:   "1000002",
					Status: "ACTIVE",
				},
			},
		},
	}
	validator := newTestValidator(h)

	result := validator.ValidateTemporalParentAvailability(context.Background(), tenant, "1000002", date)
	if !result.Valid {
		t.Fatalf("expected temporal parent availability to succeed, errors: %#v", result.Errors)
	}
}

func TestOrganizationUpdateStatusGuard(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {
				Code:          "TARGET",
				Status:        "INACTIVE",
				EffectiveDate: types.NewDateFromTime(time.Now()),
			},
		},
	}

	validator := newTestValidator(h)

	status := "PLANNED"
	req := &types.UpdateOrganizationRequest{
		Status: &status,
	}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if result.Valid {
		t.Fatalf("expected status guard to fail")
	}

	found := false
	for _, errItem := range result.Errors {
		if errItem.Code == "ORG_STATUS_GUARD" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected ORG_STATUS_GUARD error, got %#v", result.Errors)
	}
}

func TestOrganizationCreateNameRequired(t *testing.T) {
	validator := newTestValidator(&stubHierarchy{})
	req := &types.CreateOrganizationRequest{
		Name:       "   ",
		UnitType:   "DEPARTMENT",
		ParentCode: nil,
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, uuid.New())
	if result.Valid {
		t.Fatalf("expected name required validation to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_NAME_REQUIRED")
}

func TestOrganizationCreateInvalidCodeFormat(t *testing.T) {
	validator := newTestValidator(&stubHierarchy{})
	code := "ABC1234"
	req := &types.CreateOrganizationRequest{
		Name:       "Org",
		UnitType:   "DEPARTMENT",
		Code:       &code,
		ParentCode: nil,
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, uuid.New())
	if result.Valid {
		t.Fatalf("expected invalid code to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_CODE_INVALID")
}

func TestOrganizationCreateInvalidParentFormat(t *testing.T) {
	validator := newTestValidator(&stubHierarchy{})
	parent := "abc"
	req := &types.CreateOrganizationRequest{
		Name:       "Org",
		UnitType:   "DEPARTMENT",
		ParentCode: &parent,
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, uuid.New())
	if result.Valid {
		t.Fatalf("expected invalid parent format to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_PARENT_INVALID")
}

func TestOrganizationUpdateInvalidStatus(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {Code: "TARGET", Status: "ACTIVE"},
		},
	}
	validator := newTestValidator(h)

	status := "unknown"
	req := &types.UpdateOrganizationRequest{Status: &status}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if result.Valid {
		t.Fatalf("expected invalid status to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_STATUS_INVALID")
}

func TestOrganizationUpdateInvalidParentFormat(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {Code: "TARGET", Status: "ACTIVE"},
		},
	}
	validator := newTestValidator(h)

	parent := "abc"
	req := &types.UpdateOrganizationRequest{ParentCode: &parent}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if result.Valid {
		t.Fatalf("expected invalid parent format to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_PARENT_INVALID")
}

func TestOrganizationUpdateSortOrderOutOfRange(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {Code: "TARGET", Status: "ACTIVE"},
		},
	}
	validator := newTestValidator(h)

	sortOrder := 10000
	req := &types.UpdateOrganizationRequest{SortOrder: &sortOrder}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if result.Valid {
		t.Fatalf("expected sort order validation to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_SORT_ORDER_INVALID")
}

func TestOrganizationCreateDescriptionTooLong(t *testing.T) {
	validator := newTestValidator(&stubHierarchy{})
	req := &types.CreateOrganizationRequest{
		Name:        "Org",
		UnitType:    "DEPARTMENT",
		Description: strings.Repeat("a", types.OrganizationDescriptionMaxLength+1),
	}

	result := validator.ValidateOrganizationCreation(context.Background(), req, uuid.New())
	if result.Valid {
		t.Fatalf("expected description too long validation to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_DESCRIPTION_TOO_LONG")
}

func TestOrganizationUpdateNameEmpty(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {Code: "TARGET", Status: "ACTIVE"},
		},
	}
	validator := newTestValidator(h)

	name := "   "
	req := &types.UpdateOrganizationRequest{Name: &name}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if result.Valid {
		t.Fatalf("expected empty name validation to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_NAME_REQUIRED")
}

func TestOrganizationUpdateUnitTypeEmpty(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {Code: "TARGET", Status: "ACTIVE"},
		},
	}
	validator := newTestValidator(h)

	unitType := "  "
	req := &types.UpdateOrganizationRequest{UnitType: &unitType}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if result.Valid {
		t.Fatalf("expected empty unit type validation to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_UNIT_TYPE_REQUIRED")
}

func TestOrganizationUpdateDescriptionTooLong(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {Code: "TARGET", Status: "ACTIVE"},
		},
	}
	validator := newTestValidator(h)

	desc := strings.Repeat("b", types.OrganizationDescriptionMaxLength+1)
	req := &types.UpdateOrganizationRequest{Description: &desc}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if result.Valid {
		t.Fatalf("expected description too long validation to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_DESCRIPTION_TOO_LONG")
}

func TestOrganizationUpdateStatusEmpty(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {Code: "TARGET", Status: "ACTIVE"},
		},
	}
	validator := newTestValidator(h)

	status := " "
	req := &types.UpdateOrganizationRequest{Status: &status}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if result.Valid {
		t.Fatalf("expected empty status validation to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_STATUS_REQUIRED")
}

func TestOrganizationUpdateNegativeSortOrder(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"TARGET": {Code: "TARGET", Status: "ACTIVE"},
		},
	}
	validator := newTestValidator(h)

	sortOrder := -5
	req := &types.UpdateOrganizationRequest{SortOrder: &sortOrder}

	result := validator.ValidateOrganizationUpdate(context.Background(), "TARGET", req, tenant)
	if result.Valid {
		t.Fatalf("expected negative sort order to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORG_SORT_ORDER_INVALID")
}

func TestOrganizationUpdateTargetNotFound(t *testing.T) {
	validator := newTestValidator(&stubHierarchy{})
	req := &types.UpdateOrganizationRequest{}

	result := validator.ValidateOrganizationUpdate(context.Background(), "UNKNOWN", req, uuid.New())
	if result.Valid {
		t.Fatalf("expected target not found validation to fail")
	}
	assertHasErrorCode(t, result.Errors, "ORGANIZATION_NOT_FOUND")
}

func TestNewBusinessRuleValidator(t *testing.T) {
	logger := pkglogger.NewLogger(
		pkglogger.WithWriter(io.Discard),
		pkglogger.WithLevel(pkglogger.LevelError),
	)
	validator := NewBusinessRuleValidator(nil, nil, logger)
	if validator == nil {
		t.Fatal("expected validator instance")
	}
}

func assertHasErrorCode(t *testing.T, errs []ValidationError, code string) {
	t.Helper()
	for _, errItem := range errs {
		if errItem.Code == code {
			return
		}
	}
	t.Fatalf("expected error code %s, got %#v", code, errs)
}
