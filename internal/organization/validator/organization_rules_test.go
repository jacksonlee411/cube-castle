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
			"PARENT": {Code: "PARENT", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"PARENT": 17,
		},
	}
	validator := newTestValidator(h)

	now := time.Now().UTC()
	req := &types.CreateOrganizationRequest{
		Name:          "Child",
		UnitType:      "DEPARTMENT",
		ParentCode:    strPtr("PARENT"),
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
			"PARENT": {Code: "PARENT", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"PARENT": depthWarningThreshold,
		},
	}
	validator := newTestValidator(h)

	req := &types.CreateOrganizationRequest{
		Name:       "Child",
		UnitType:   "DEPARTMENT",
		ParentCode: strPtr("PARENT"),
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
			"PARENT": {Code: "PARENT", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"PARENT": 1,
		},
		temporal: map[string]map[string]*repository.OrganizationNode{
			"PARENT": {
				date.Format("2006-01-02"): {
					Code:   "PARENT",
					Status: "INACTIVE",
				},
			},
		},
	}
	validator := newTestValidator(h)

	req := &types.CreateOrganizationRequest{
		Name:          "Child",
		UnitType:      "DEPARTMENT",
		ParentCode:    strPtr("PARENT"),
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
	parent := "PARENT"

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
		ParentCode: strPtr("PARENT"),
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

func TestOrganizationCreateDepthUnknownReturnsInvalidParent(t *testing.T) {
	tenant := uuid.New()
	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"PARENT": {Code: "PARENT", Status: "ACTIVE"},
		},
	}
	validator := newTestValidator(h)

	req := &types.CreateOrganizationRequest{
		Name:       "Child",
		UnitType:   "DEPARTMENT",
		ParentCode: strPtr("PARENT"),
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

	hasUnitType := false
	for _, errItem := range result.Errors {
		if errItem.Code == "INVALID_UNIT_TYPE" {
			hasUnitType = true
		}
	}
	if !hasUnitType {
		t.Fatalf("expected INVALID_UNIT_TYPE error, got %#v", result.Errors)
	}

	if len(result.Warnings) == 0 {
		t.Fatalf("expected warnings for temporal/sort order issues")
	}
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
			"PARENT": 2,
		},
		ancestors: map[string][]repository.OrganizationNode{
			"PARENT": {
				{Code: "ROOT"},
				{Code: "TARGET"},
			},
		},
	}

	validator := newTestValidator(h)

	req := &types.UpdateOrganizationRequest{
		ParentCode: strPtr("PARENT"),
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

func TestValidateTemporalParentAvailability(t *testing.T) {
	tenant := uuid.New()
	date := time.Date(2025, 11, 6, 0, 0, 0, 0, time.UTC)

	h := &stubHierarchy{
		orgs: map[string]*types.Organization{
			"PARENT": {Code: "PARENT", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"PARENT": 1,
		},
		temporal: map[string]map[string]*repository.OrganizationNode{
			"PARENT": {},
		},
	}
	validator := newTestValidator(h)

	result := validator.ValidateTemporalParentAvailability(context.Background(), tenant, "PARENT", date)
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
			"PARENT": {Code: "PARENT", Status: "ACTIVE"},
		},
		depths: map[string]int{
			"PARENT": 1,
		},
		temporal: map[string]map[string]*repository.OrganizationNode{
			"PARENT": {
				date.Format("2006-01-02"): {
					Code:   "PARENT",
					Status: "ACTIVE",
				},
			},
		},
	}
	validator := newTestValidator(h)

	result := validator.ValidateTemporalParentAvailability(context.Background(), tenant, "PARENT", date)
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
