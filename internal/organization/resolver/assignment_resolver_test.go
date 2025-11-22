package resolver

import (
	"context"
	"fmt"
	"testing"

	"cube-castle/internal/organization/dto"
	pkglogger "cube-castle/pkg/logger"
	sharedconfig "cube-castle/shared/config"
	"github.com/google/uuid"
)

type stubAssignmentFacade struct {
	assignmentsFn func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	historyFn     func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	statsFn       func(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error)
}

func (s *stubAssignmentFacade) GetAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
	if s.assignmentsFn == nil {
		return nil, fmt.Errorf("assignmentsFn not configured")
	}
	return s.assignmentsFn(ctx, tenantID, positionCode, filter, pagination, sorting)
}

func (s *stubAssignmentFacade) GetAssignmentHistory(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
	if s.historyFn == nil {
		return nil, fmt.Errorf("historyFn not configured")
	}
	return s.historyFn(ctx, tenantID, positionCode, filter, pagination, sorting)
}

func (s *stubAssignmentFacade) GetAssignmentStats(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error) {
	if s.statsFn == nil {
		return nil, fmt.Errorf("statsFn not configured")
	}
	return s.statsFn(ctx, tenantID, positionCode, organizationCode)
}

func TestResolver_Assignments_ForwardsParameters(t *testing.T) {
	repo := &stubRepository{}
	perm := &stubPermissionChecker{allow: true}
	facade := &stubAssignmentFacade{}
	res := NewResolverWithAssignments(repo, facade, pkglogger.NewNoopLogger(), perm)

	positionCode := "P1001"
	filter := &dto.PositionAssignmentFilterInput{IncludeHistorical: false}
	pagination := &dto.PaginationInput{Page: 2, PageSize: 10}
	sorting := []dto.PositionAssignmentSortInput{{Field: "EFFECTIVE_DATE", Direction: "DESC"}}

	facade.assignmentsFn = func(_ context.Context, tenantID uuid.UUID, code string, f *dto.PositionAssignmentFilterInput, p *dto.PaginationInput, s []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
		if code != positionCode {
			t.Fatalf("unexpected position code: %s", code)
		}
		if tenantID != sharedconfig.DefaultTenantID {
			t.Fatalf("unexpected tenant ID: %s", tenantID.String())
		}
		if f != filter {
			t.Fatalf("filter not forwarded")
		}
		if p != pagination {
			t.Fatalf("pagination not forwarded")
		}
		if len(s) != len(sorting) {
			t.Fatalf("sorting length mismatch, got %d", len(s))
		}
		return &dto.PositionAssignmentConnection{}, nil
	}

	_, err := res.Assignments(context.Background(), struct {
		OrganizationCode *string
		PositionCode     *string
		Filter           *dto.PositionAssignmentFilterInput
		Pagination       *dto.PaginationInput
		Sorting          *[]dto.PositionAssignmentSortInput
	}{
		PositionCode: &positionCode,
		Filter:       filter,
		Pagination:   pagination,
		Sorting:      &sorting,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestResolver_AssignmentHistory_UsesFacade(t *testing.T) {
	repo := &stubRepository{}
	perm := &stubPermissionChecker{allow: true}
	facade := &stubAssignmentFacade{}
	res := NewResolverWithAssignments(repo, facade, pkglogger.NewNoopLogger(), perm)

	facade.historyFn = func(_ context.Context, _ uuid.UUID, code string, _ *dto.PositionAssignmentFilterInput, _ *dto.PaginationInput, _ []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
		if code != "P1002" {
			t.Fatalf("unexpected position code: %s", code)
		}
		return &dto.PositionAssignmentConnection{}, nil
	}

	_, err := res.AssignmentHistory(context.Background(), struct {
		PositionCode string
		Filter       *dto.PositionAssignmentFilterInput
		Pagination   *dto.PaginationInput
		Sorting      *[]dto.PositionAssignmentSortInput
	}{
		PositionCode: "P1002",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestResolver_AssignmentStats_RequiresIdentifiers(t *testing.T) {
	repo := &stubRepository{}
	perm := &stubPermissionChecker{allow: true}
	res := NewResolverWithAssignments(repo, nil, pkglogger.NewNoopLogger(), perm)

	_, err := res.AssignmentStats(context.Background(), struct {
		OrganizationCode *string
		PositionCode     *string
	}{})
	if err == nil || err.Error() != "ASSIGNMENT_QUERY_FACADE_NOT_CONFIGURED" {
		t.Fatalf("expected facade not configured error, got %v", err)
	}
}

func TestResolver_AssignmentStats_ForwardsParameters(t *testing.T) {
	repo := &stubRepository{}
	perm := &stubPermissionChecker{allow: true}
	facade := &stubAssignmentFacade{}
	res := NewResolverWithAssignments(repo, facade, pkglogger.NewNoopLogger(), perm)

	positionCode := "P2001"
	orgCode := "DEPT01"
	facade.statsFn = func(_ context.Context, _ uuid.UUID, pos string, org string) (*dto.AssignmentStats, error) {
		if pos != positionCode {
			t.Fatalf("unexpected position code: %s", pos)
		}
		if org != orgCode {
			t.Fatalf("unexpected organization code: %s", org)
		}
		return &dto.AssignmentStats{}, nil
	}

	_, err := res.AssignmentStats(context.Background(), struct {
		OrganizationCode *string
		PositionCode     *string
	}{
		OrganizationCode: &orgCode,
		PositionCode:     &positionCode,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
