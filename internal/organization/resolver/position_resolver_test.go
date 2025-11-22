package resolver

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"cube-castle/internal/auth"
	"cube-castle/internal/organization/dto"
	pkglogger "cube-castle/pkg/logger"
	sharedconfig "cube-castle/shared/config"
	"github.com/google/uuid"
)

type stubPermissionChecker struct {
	allow     bool
	lastQuery string
	err       error
}

func (s *stubPermissionChecker) CheckQueryPermission(_ context.Context, queryName string) error {
	s.lastQuery = queryName
	if s.err != nil {
		return s.err
	}
	if !s.allow {
		return fmt.Errorf("denied")
	}
	return nil
}

type stubRepository struct {
	positionsFn                      func(ctx context.Context, tenantID uuid.UUID, filter *dto.PositionFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionSortInput) (*dto.PositionConnection, error)
	positionByCodeFn                 func(ctx context.Context, tenantID uuid.UUID, code string, asOfDate *string) (*dto.Position, error)
	timelineFn                       func(ctx context.Context, tenantID uuid.UUID, code string, startDate, endDate *string) ([]dto.PositionTimelineEntry, error)
	versionsFn                       func(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]dto.Position, error)
	vacantFn                         func(ctx context.Context, tenantID uuid.UUID, filter *dto.VacantPositionFilterInput, pagination *dto.PaginationInput, sorting []dto.VacantPositionSortInput) (*dto.VacantPositionConnection, error)
	transferFn                       func(ctx context.Context, tenantID uuid.UUID, positionCode *string, organizationCode *string, pagination *dto.PaginationInput) (*dto.PositionTransferConnection, error)
	headcountFn                      func(ctx context.Context, tenantID uuid.UUID, organizationCode string, includeSubordinates bool) (*dto.HeadcountStats, error)
	assignmentsFn                    func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	assignmentAuditFn                func(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID *string, dateRange *dto.DateRangeInput, pagination *dto.PaginationInput) (*dto.PositionAssignmentAuditConnection, error)
	assignmentHistoryFn              func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	assignmentStatsFn                func(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error)
	capturedSorting                  []dto.PositionSortInput
	capturedFilter                   *dto.PositionFilterInput
	capturedPagination               *dto.PaginationInput
	capturedTenant                   uuid.UUID
	includeSubs                      bool
	capturedAssignmentFilter         *dto.PositionAssignmentFilterInput
	capturedAssignmentSorting        []dto.PositionAssignmentSortInput
	capturedPositionCode             string
	capturedVersionsCode             string
	capturedIncludeDeleted           bool
	capturedVacantFilter             *dto.VacantPositionFilterInput
	capturedVacantSorting            []dto.VacantPositionSortInput
	capturedTransferPositionCode     *string
	capturedTransferOrganizationCode *string
	capturedAuditAssignmentID        *string
	capturedAuditDateRange           *dto.DateRangeInput
	capturedAuditPagination          *dto.PaginationInput
}

func (s *stubRepository) GetOrganizations(_ context.Context, _ uuid.UUID, _ *dto.OrganizationFilter, _ *dto.PaginationInput) (*dto.OrganizationConnection, error) {
	panic("GetOrganizations not expected")
}

func (s *stubRepository) GetOrganization(_ context.Context, _ uuid.UUID, _ string) (*dto.Organization, error) {
	panic("GetOrganization not expected")
}

func (s *stubRepository) GetOrganizationAtDate(_ context.Context, _ uuid.UUID, _ string, _ string) (*dto.Organization, error) {
	panic("GetOrganizationAtDate not expected")
}

func (s *stubRepository) GetOrganizationHistory(_ context.Context, _ uuid.UUID, _ string, _ string, _ string) ([]dto.Organization, error) {
	panic("GetOrganizationHistory not expected")
}

func (s *stubRepository) GetOrganizationVersions(_ context.Context, _ uuid.UUID, _ string, _ bool) ([]dto.Organization, error) {
	panic("GetOrganizationVersions not expected")
}

func (s *stubRepository) GetOrganizationStats(_ context.Context, _ uuid.UUID) (*dto.OrganizationStats, error) {
	panic("GetOrganizationStats not expected")
}

func (s *stubRepository) GetOrganizationHierarchy(_ context.Context, _ uuid.UUID, _ string) (*dto.OrganizationHierarchyData, error) {
	panic("GetOrganizationHierarchy not expected")
}

func (s *stubRepository) GetOrganizationSubtree(_ context.Context, _ uuid.UUID, _ string, _ int) (*dto.OrganizationHierarchyData, error) {
	panic("GetOrganizationSubtree not expected")
}

func (s *stubRepository) GetPositions(ctx context.Context, tenantID uuid.UUID, filter *dto.PositionFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionSortInput) (*dto.PositionConnection, error) {
	if s.positionsFn == nil {
		panic("positionsFn not configured")
	}
	s.capturedTenant = tenantID
	s.capturedFilter = filter
	s.capturedPagination = pagination
	s.capturedSorting = sorting
	return s.positionsFn(ctx, tenantID, filter, pagination, sorting)
}

func (s *stubRepository) GetPositionByCode(ctx context.Context, tenantID uuid.UUID, code string, asOfDate *string) (*dto.Position, error) {
	if s.positionByCodeFn == nil {
		panic("positionByCodeFn not configured")
	}
	return s.positionByCodeFn(ctx, tenantID, code, asOfDate)
}

func (s *stubRepository) GetPositionAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
	if s.assignmentsFn == nil {
		panic("assignmentsFn not configured")
	}
	s.capturedTenant = tenantID
	s.capturedAssignmentFilter = filter
	s.capturedAssignmentSorting = sorting
	s.capturedPagination = pagination
	s.capturedPositionCode = positionCode
	return s.assignmentsFn(ctx, tenantID, positionCode, filter, pagination, sorting)
}

func (s *stubRepository) GetPositionAssignmentAudit(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID *string, dateRange *dto.DateRangeInput, pagination *dto.PaginationInput) (*dto.PositionAssignmentAuditConnection, error) {
	if s.assignmentAuditFn == nil {
		panic("assignmentAuditFn not configured")
	}
	s.capturedTenant = tenantID
	s.capturedPositionCode = positionCode
	s.capturedAuditAssignmentID = assignmentID
	s.capturedAuditDateRange = dateRange
	s.capturedAuditPagination = pagination
	return s.assignmentAuditFn(ctx, tenantID, positionCode, assignmentID, dateRange, pagination)
}

func (s *stubRepository) GetAssignmentHistory(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
	if s.assignmentHistoryFn == nil {
		panic("assignmentHistoryFn not configured")
	}
	return s.assignmentHistoryFn(ctx, tenantID, positionCode, filter, pagination, sorting)
}

func (s *stubRepository) GetAssignmentStats(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error) {
	if s.assignmentStatsFn == nil {
		panic("assignmentStatsFn not configured")
	}
	return s.assignmentStatsFn(ctx, tenantID, positionCode, organizationCode)
}

func (s *stubRepository) GetPositionTimeline(ctx context.Context, tenantID uuid.UUID, code string, startDate, endDate *string) ([]dto.PositionTimelineEntry, error) {
	if s.timelineFn == nil {
		panic("timelineFn not configured")
	}
	return s.timelineFn(ctx, tenantID, code, startDate, endDate)
}

func (s *stubRepository) GetPositionVersions(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]dto.Position, error) {
	if s.versionsFn == nil {
		panic("versionsFn not configured")
	}
	s.capturedTenant = tenantID
	s.capturedVersionsCode = code
	s.capturedIncludeDeleted = includeDeleted
	return s.versionsFn(ctx, tenantID, code, includeDeleted)
}

func (s *stubRepository) GetVacantPositionConnection(ctx context.Context, tenantID uuid.UUID, filter *dto.VacantPositionFilterInput, pagination *dto.PaginationInput, sorting []dto.VacantPositionSortInput) (*dto.VacantPositionConnection, error) {
	if s.vacantFn == nil {
		panic("vacantFn not configured")
	}
	s.capturedTenant = tenantID
	s.capturedVacantFilter = filter
	s.capturedPagination = pagination
	s.capturedVacantSorting = sorting
	return s.vacantFn(ctx, tenantID, filter, pagination, sorting)
}

func (s *stubRepository) GetPositionHeadcountStats(ctx context.Context, tenantID uuid.UUID, organizationCode string, includeSubordinates bool) (*dto.HeadcountStats, error) {
	if s.headcountFn == nil {
		panic("headcountFn not configured")
	}
	s.capturedTenant = tenantID
	s.includeSubs = includeSubordinates
	return s.headcountFn(ctx, tenantID, organizationCode, includeSubordinates)
}

func (s *stubRepository) GetPositionTransfers(ctx context.Context, tenantID uuid.UUID, positionCode *string, organizationCode *string, pagination *dto.PaginationInput) (*dto.PositionTransferConnection, error) {
	if s.transferFn == nil {
		panic("transferFn not configured")
	}
	s.capturedTenant = tenantID
	s.capturedPagination = pagination
	s.capturedTransferPositionCode = positionCode
	s.capturedTransferOrganizationCode = organizationCode
	return s.transferFn(ctx, tenantID, positionCode, organizationCode, pagination)
}

func (s *stubRepository) GetJobFamilyGroups(_ context.Context, _ uuid.UUID, _ bool, _ *string) ([]dto.JobFamilyGroup, error) {
	panic("GetJobFamilyGroups not expected")
}

func (s *stubRepository) GetJobFamilies(_ context.Context, _ uuid.UUID, _ string, _ bool, _ *string) ([]dto.JobFamily, error) {
	panic("GetJobFamilies not expected")
}

func (s *stubRepository) GetJobRoles(_ context.Context, _ uuid.UUID, _ string, _ bool, _ *string) ([]dto.JobRole, error) {
	panic("GetJobRoles not expected")
}

func (s *stubRepository) GetJobLevels(_ context.Context, _ uuid.UUID, _ string, _ bool, _ *string) ([]dto.JobLevel, error) {
	panic("GetJobLevels not expected")
}

func (s *stubRepository) GetAuditHistory(_ context.Context, _ uuid.UUID, _ string, _ *string, _ *string, _ *string, _ *string, _ int) ([]dto.AuditRecordData, error) {
	panic("GetAuditHistory not expected")
}

func (s *stubRepository) GetAuditLog(_ context.Context, _ string) (*dto.AuditRecordData, error) {
	panic("GetAuditLog not expected")
}

func TestResolver_Positions_ForwardsParameters(t *testing.T) {
	filter := &dto.PositionFilterInput{}
	orgCode := "1000001"
	filter.OrganizationCode = &orgCode
	pagination := &dto.PaginationInput{Page: 2, PageSize: 20}
	sorting := []dto.PositionSortInput{
		{Field: "code", Direction: "ASC"},
	}

	repo := &stubRepository{
		positionsFn: func(_ context.Context, tenantID uuid.UUID, _ *dto.PositionFilterInput, _ *dto.PaginationInput, _ []dto.PositionSortInput) (*dto.PositionConnection, error) {
			now := time.Now().UTC()
			position := dto.Position{
				CodeField:               "P1000001",
				RecordIDField:           uuid.New().String(),
				TenantIDField:           tenantID.String(),
				TitleField:              "HR Manager",
				JobFamilyGroupCodeField: "OPER",
				JobFamilyCodeField:      "OPER-HR",
				JobRoleCodeField:        "OPER-HR-SUP",
				JobLevelCodeField:       "P1",
				OrganizationCodeField:   "1000001",
				PositionTypeField:       "FULL_TIME",
				EmploymentTypeField:     "FULL_TIME",
				StatusField:             "ACTIVE",
				HeadcountCapacityField:  1,
				HeadcountInUseField:     0,
				EffectiveDateField:      now,
				IsCurrentField:          true,
				CreatedAtField:          now,
				UpdatedAtField:          now,
			}
			return &dto.PositionConnection{
				DataField:       []dto.Position{position},
				EdgesField:      []dto.PositionEdge{{CursorField: "cursor-1", NodeField: position}},
				PaginationField: dto.PaginationInfo{PageField: 2, PageSizeField: 20, TotalField: 1, HasNextField: false},
				TotalCountField: 1,
			}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), perm)

	result, err := resolver.Positions(context.Background(), struct {
		Filter     *dto.PositionFilterInput
		Pagination *dto.PaginationInput
		Sorting    *[]dto.PositionSortInput
	}{Filter: filter, Pagination: pagination, Sorting: &sorting})

	if err != nil {
		t.Fatalf("Positions returned error: %v", err)
	}
	if perm.lastQuery != "positions" {
		t.Fatalf("expected permission check for positions, got %s", perm.lastQuery)
	}
	if repo.capturedTenant != sharedconfig.DefaultTenantID {
		t.Fatalf("expected tenant %s, got %s", sharedconfig.DefaultTenantID, repo.capturedTenant)
	}
	if repo.capturedFilter == nil || repo.capturedFilter.OrganizationCode == nil || *repo.capturedFilter.OrganizationCode != "1000001" {
		t.Fatalf("filter not forwarded")
	}
	if len(repo.capturedSorting) != 1 || repo.capturedSorting[0].Field != "code" {
		t.Fatalf("sorting not forwarded: %#v", repo.capturedSorting)
	}
	if result.TotalCountField != 1 {
		t.Fatalf("unexpected total count %d", result.TotalCountField)
	}
}

func TestResolver_Positions_PermissionDenied(t *testing.T) {
	repo := &stubRepository{
		positionsFn: func(_ context.Context, _ uuid.UUID, _ *dto.PositionFilterInput, _ *dto.PaginationInput, _ []dto.PositionSortInput) (*dto.PositionConnection, error) {
			t.Fatalf("repository should not be called when permission denied")
			return nil, nil
		},
	}
	perm := &stubPermissionChecker{allow: false}
	resolver := NewResolver(repo, newTestLogger(), perm)

	_, err := resolver.Positions(context.Background(), struct {
		Filter     *dto.PositionFilterInput
		Pagination *dto.PaginationInput
		Sorting    *[]dto.PositionSortInput
	}{})

	if err == nil || err.Error() != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("expected INSUFFICIENT_PERMISSIONS, got %v", err)
	}
}

func TestResolver_PositionAssignments_ForwardsParameters(t *testing.T) {
	positionCode := "P1000001"
	employeeID := uuid.New().String()
	filter := &dto.PositionAssignmentFilterInput{
		IncludeHistorical: false,
	}
	filter.EmployeeID = &employeeID
	status := "ACTIVE"
	filter.Status = &status

	pagination := &dto.PaginationInput{Page: 1, PageSize: 10}
	sorting := []dto.PositionAssignmentSortInput{
		{Field: "EFFECTIVE_DATE", Direction: "ASC"},
	}

	repo := &stubRepository{
		assignmentsFn: func(_ context.Context, tenantID uuid.UUID, code string, _ *dto.PositionAssignmentFilterInput, _ *dto.PaginationInput, _ []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
			now := time.Now().UTC()
			assignment := dto.PositionAssignment{
				AssignmentIDField:     uuid.New().String(),
				TenantIDField:         tenantID.String(),
				PositionCodeField:     code,
				PositionRecordIDField: uuid.New().String(),
				EmployeeIDField:       employeeID,
				EmployeeNameField:     "Alice",
				AssignmentTypeField:   "PRIMARY",
				AssignmentStatusField: "ACTIVE",
				FTEField:              1,
				EffectiveDateField:    now,
				IsCurrentField:        true,
				CreatedAtField:        now,
				UpdatedAtField:        now,
			}
			return &dto.PositionAssignmentConnection{
				DataField: []dto.PositionAssignment{assignment},
				EdgesField: []dto.PositionAssignmentEdge{
					{CursorField: assignment.AssignmentIDField, NodeField: assignment},
				},
				PaginationField: dto.PaginationInfo{
					PageField:        1,
					PageSizeField:    10,
					TotalField:       1,
					HasNextField:     false,
					HasPreviousField: false,
				},
				TotalCountField: 1,
			}, nil
		},
	}

	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), perm)

	result, err := resolver.PositionAssignments(context.Background(), struct {
		PositionCode string
		Filter       *dto.PositionAssignmentFilterInput
		Pagination   *dto.PaginationInput
		Sorting      *[]dto.PositionAssignmentSortInput
	}{
		PositionCode: positionCode,
		Filter:       filter,
		Sorting:      &sorting,
		Pagination:   pagination,
	})

	if err != nil {
		t.Fatalf("PositionAssignments returned error: %v", err)
	}
	if result == nil {
		t.Fatalf("expected non-nil result")
	}
	if perm.lastQuery != "positionAssignments" {
		t.Fatalf("expected permission check for positionAssignments, got %s", perm.lastQuery)
	}
	if repo.capturedTenant != sharedconfig.DefaultTenantID {
		t.Fatalf("expected tenant %s, got %s", sharedconfig.DefaultTenantID, repo.capturedTenant)
	}
	if repo.capturedPositionCode != positionCode {
		t.Fatalf("expected position code %s, got %s", positionCode, repo.capturedPositionCode)
	}
	if repo.capturedAssignmentFilter != filter {
		t.Fatalf("expected filter pointer to be forwarded")
	}
	if repo.capturedPagination != pagination {
		t.Fatalf("expected pagination pointer to be forwarded")
	}
	if len(repo.capturedAssignmentSorting) != len(sorting) {
		t.Fatalf("expected %d sorting items, got %d", len(sorting), len(repo.capturedAssignmentSorting))
	}
}

func TestResolver_PositionAssignmentAudit_ForwardsParameters(t *testing.T) {
	assignmentID := uuid.New().String()
	from := "2025-01-01"
	to := "2025-01-31"
	dateRange := &dto.DateRangeInput{From: &from, To: &to}
	pagination := &dto.PaginationInput{Page: 2, PageSize: 20}

	repo := &stubRepository{
		assignmentAuditFn: func(_ context.Context, _ uuid.UUID, _ string, _ *string, _ *dto.DateRangeInput, p *dto.PaginationInput) (*dto.PositionAssignmentAuditConnection, error) {
			return &dto.PositionAssignmentAuditConnection{
				DataField: []dto.PositionAssignmentAudit{},
				PaginationField: dto.PaginationInfo{
					PageField:        int(p.Page),
					PageSizeField:    int(p.PageSize),
					TotalField:       0,
					HasNextField:     false,
					HasPreviousField: false,
				},
				TotalCountField: 0,
			}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), perm)

	code := "P3000001"
	_, err := resolver.PositionAssignmentAudit(context.Background(), struct {
		PositionCode string
		AssignmentId *string
		DateRange    *dto.DateRangeInput
		Pagination   *dto.PaginationInput
	}{
		PositionCode: code,
		AssignmentId: &assignmentID,
		DateRange:    dateRange,
		Pagination:   pagination,
	})
	if err != nil {
		t.Fatalf("PositionAssignmentAudit returned error: %v", err)
	}
	if perm.lastQuery != "positionAssignmentAudit" {
		t.Fatalf("expected permission check for positionAssignmentAudit, got %s", perm.lastQuery)
	}
	if repo.capturedTenant != sharedconfig.DefaultTenantID {
		t.Fatalf("expected tenant %s, got %s", sharedconfig.DefaultTenantID, repo.capturedTenant)
	}
	if repo.capturedPositionCode != code {
		t.Fatalf("expected position code %s, got %s", code, repo.capturedPositionCode)
	}
	if repo.capturedAuditAssignmentID == nil || *repo.capturedAuditAssignmentID != assignmentID {
		t.Fatalf("expected assignmentId forwarded")
	}
	if repo.capturedAuditDateRange != dateRange {
		t.Fatalf("expected dateRange pointer forwarded")
	}
	if repo.capturedAuditPagination != pagination {
		t.Fatalf("expected pagination pointer forwarded")
	}
}

func TestResolver_VacantPositions_ForwardsParameters(t *testing.T) {
	orgCodes := []string{"1001001"}
	minDays := 30
	filter := &dto.VacantPositionFilterInput{
		OrganizationCodes: &orgCodes,
		MinimumVacantDays: &minDays,
	}
	pagination := &dto.PaginationInput{Page: 3, PageSize: 15}
	sorting := []dto.VacantPositionSortInput{
		{Field: "VACANT_SINCE", Direction: "DESC"},
	}

	repo := &stubRepository{
		vacantFn: func(_ context.Context, _ uuid.UUID, _ *dto.VacantPositionFilterInput, p *dto.PaginationInput, _ []dto.VacantPositionSortInput) (*dto.VacantPositionConnection, error) {
			return &dto.VacantPositionConnection{
				DataField:  []dto.VacantPosition{},
				EdgesField: []dto.VacantPositionEdge{},
				PaginationField: dto.PaginationInfo{
					PageField:        int(p.Page),
					PageSizeField:    int(p.PageSize),
					TotalField:       0,
					HasNextField:     false,
					HasPreviousField: false,
				},
				TotalCountField: 0,
			}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), perm)

	_, err := resolver.VacantPositions(context.Background(), struct {
		Filter     *dto.VacantPositionFilterInput
		Pagination *dto.PaginationInput
		Sorting    *[]dto.VacantPositionSortInput
	}{
		Filter:     filter,
		Pagination: pagination,
		Sorting:    &sorting,
	})
	if err != nil {
		t.Fatalf("VacantPositions returned error: %v", err)
	}
	if perm.lastQuery != "vacantPositions" {
		t.Fatalf("expected permission check for vacantPositions, got %s", perm.lastQuery)
	}
	if repo.capturedTenant != sharedconfig.DefaultTenantID {
		t.Fatalf("expected tenant %s, got %s", sharedconfig.DefaultTenantID, repo.capturedTenant)
	}
	if repo.capturedVacantFilter != filter {
		t.Fatalf("expected filter pointer forwarded")
	}
	if repo.capturedPagination != pagination {
		t.Fatalf("expected pagination pointer forwarded")
	}
	if len(repo.capturedVacantSorting) != len(sorting) {
		t.Fatalf("expected %d sorting items, got %d", len(sorting), len(repo.capturedVacantSorting))
	}
}

func TestResolver_VacantPositions_ForwardsAsOfDate(t *testing.T) {
	orgCodes := []string{"1001001"}
	asOf := "2025-01-15"
	filter := &dto.VacantPositionFilterInput{
		OrganizationCodes: &orgCodes,
		AsOfDate:          &asOf,
	}
	repo := &stubRepository{
		vacantFn: func(_ context.Context, _ uuid.UUID, _ *dto.VacantPositionFilterInput, _ *dto.PaginationInput, _ []dto.VacantPositionSortInput) (*dto.VacantPositionConnection, error) {
			return &dto.VacantPositionConnection{}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), perm)

	_, err := resolver.VacantPositions(context.Background(), struct {
		Filter     *dto.VacantPositionFilterInput
		Pagination *dto.PaginationInput
		Sorting    *[]dto.VacantPositionSortInput
	}{
		Filter: filter,
	})
	if err != nil {
		t.Fatalf("VacantPositions returned error: %v", err)
	}
	if repo.capturedVacantFilter == nil || repo.capturedVacantFilter.AsOfDate == nil {
		t.Fatalf("expected asOfDate forwarded")
	}
	if *repo.capturedVacantFilter.AsOfDate != asOf {
		t.Fatalf("expected asOfDate %s, got %s", asOf, *repo.capturedVacantFilter.AsOfDate)
	}
}

func TestResolver_PositionTransfers_ForwardsParameters(t *testing.T) {
	positionCode := "P2000001"
	orgCode := "1002001"
	pagination := &dto.PaginationInput{Page: 2, PageSize: 5}

	repo := &stubRepository{
		transferFn: func(_ context.Context, _ uuid.UUID, _ *string, _ *string, p *dto.PaginationInput) (*dto.PositionTransferConnection, error) {
			transfer := dto.PositionTransfer{
				TransferIDField:           uuid.New().String(),
				PositionCodeField:         positionCode,
				FromOrganizationCodeField: "1001000",
				ToOrganizationCodeField:   orgCode,
				EffectiveDateField:        time.Now().UTC(),
				CreatedAtField:            time.Now().UTC(),
				InitiatedByField: dto.OperatedByData{
					IDField:   "user-1",
					NameField: "User One",
				},
			}
			return &dto.PositionTransferConnection{
				DataField:       []dto.PositionTransfer{transfer},
				EdgesField:      []dto.PositionTransferEdge{{CursorField: transfer.TransferIDField, NodeField: transfer}},
				PaginationField: dto.PaginationInfo{PageField: int(p.Page), PageSizeField: int(p.PageSize), TotalField: 1, HasNextField: false},
				TotalCountField: 1,
			}, nil
		},
	}

	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), perm)

	_, err := resolver.PositionTransfers(context.Background(), struct {
		PositionCode     *string
		OrganizationCode *string
		Pagination       *dto.PaginationInput
	}{
		PositionCode:     &positionCode,
		OrganizationCode: &orgCode,
		Pagination:       pagination,
	})
	if err != nil {
		t.Fatalf("PositionTransfers returned error: %v", err)
	}
	if perm.lastQuery != "positionTransfers" {
		t.Fatalf("expected permission check for positionTransfers, got %s", perm.lastQuery)
	}
	if repo.capturedTenant != sharedconfig.DefaultTenantID {
		t.Fatalf("expected tenant %s, got %s", sharedconfig.DefaultTenantID, repo.capturedTenant)
	}
	if repo.capturedTransferPositionCode == nil || *repo.capturedTransferPositionCode != positionCode {
		t.Fatalf("expected position code forwarded")
	}
	if repo.capturedTransferOrganizationCode == nil || *repo.capturedTransferOrganizationCode != orgCode {
		t.Fatalf("expected organization code forwarded")
	}
	if repo.capturedPagination != pagination {
		t.Fatalf("expected pagination forwarded")
	}
}

func TestResolver_PositionHeadcountStats_CustomIncludeSubordinates(t *testing.T) {
	repo := &stubRepository{
		headcountFn: func(_ context.Context, _ uuid.UUID, organizationCode string, _ bool) (*dto.HeadcountStats, error) {
			return &dto.HeadcountStats{OrganizationCodeField: organizationCode}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), perm)

	falseVal := false
	_, err := resolver.PositionHeadcountStats(context.Background(), struct {
		OrganizationCode    string
		IncludeSubordinates *bool
	}{OrganizationCode: "1000001", IncludeSubordinates: &falseVal})
	if err != nil {
		t.Fatalf("PositionHeadcountStats returned error: %v", err)
	}
	if repo.includeSubs {
		t.Fatalf("expected includeSubordinates=false when provided")
	}
}

func TestResolver_PositionHeadcountStats_UsesTenantFromContext(t *testing.T) {
	targetTenant := uuid.New()
	repo := &stubRepository{
		headcountFn: func(_ context.Context, tenantID uuid.UUID, organizationCode string, _ bool) (*dto.HeadcountStats, error) {
			if tenantID != targetTenant {
				t.Fatalf("expected tenant %s, got %s", targetTenant, tenantID)
			}
			return &dto.HeadcountStats{OrganizationCodeField: organizationCode}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), perm)

	ctx := auth.SetUserContext(context.Background(), &auth.Claims{
		UserID:   "tenant-tester",
		TenantID: targetTenant.String(),
	})

	_, err := resolver.PositionHeadcountStats(ctx, struct {
		OrganizationCode    string
		IncludeSubordinates *bool
	}{OrganizationCode: "1000001"})
	if err != nil {
		t.Fatalf("PositionHeadcountStats returned error: %v", err)
	}
	if repo.capturedTenant != targetTenant {
		t.Fatalf("expected captured tenant to be %s, got %s", targetTenant, repo.capturedTenant)
	}
	if !repo.includeSubs {
		t.Fatalf("expected includeSubordinates to default to true")
	}
}

func TestResolver_Position_ForwardsAsOfDate(t *testing.T) {
	expectedDate := "2025-01-01"
	repo := &stubRepository{
		positionByCodeFn: func(_ context.Context, _ uuid.UUID, code string, asOfDate *string) (*dto.Position, error) {
			if asOfDate == nil || *asOfDate != expectedDate {
				t.Fatalf("expected asOfDate %s, got %v", expectedDate, asOfDate)
			}
			return &dto.Position{CodeField: code}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), perm)

	_, err := resolver.Position(context.Background(), struct {
		Code     string
		AsOfDate *string
	}{Code: "P1000001", AsOfDate: &expectedDate})
	if err != nil {
		t.Fatalf("Position returned error: %v", err)
	}
}

func TestResolver_PositionTimeline_ForwardsDateRange(t *testing.T) {
	start := "2025-01-01"
	end := "2025-12-31"
	repo := &stubRepository{
		timelineFn: func(_ context.Context, _ uuid.UUID, _ string, startDate, endDate *string) ([]dto.PositionTimelineEntry, error) {
			if startDate == nil || *startDate != start {
				t.Fatalf("unexpected startDate %v", startDate)
			}
			if endDate == nil || *endDate != end {
				t.Fatalf("unexpected endDate %v", endDate)
			}
			return []dto.PositionTimelineEntry{}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), perm)

	_, err := resolver.PositionTimeline(context.Background(), struct {
		Code      string
		StartDate *string
		EndDate   *string
	}{Code: "P1000001", StartDate: &start, EndDate: &end})
	if err != nil {
		t.Fatalf("PositionTimeline returned error: %v", err)
	}
}

func newTestLogger() pkglogger.Logger {
	return pkglogger.NewLogger(
		pkglogger.WithWriter(io.Discard),
	)
}

func TestResolver_PositionVersions_ForwardsParameters(t *testing.T) {
	repo := &stubRepository{
		versionsFn: func(_ context.Context, _ uuid.UUID, code string, _ bool) ([]dto.Position, error) {
			if code != "P1000001" {
				t.Fatalf("unexpected code: %s", code)
			}
			return []dto.Position{{CodeField: code}}, nil
		},
	}

	checker := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, newTestLogger(), checker)

	result, err := resolver.PositionVersions(context.Background(), struct {
		Code           string
		IncludeDeleted *bool
	}{
		Code: "P1000001",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 version, got %d", len(result))
	}

	if checker.lastQuery != "positionVersions" {
		t.Fatalf("expected permission check for positionVersions, got %s", checker.lastQuery)
	}

	if repo.capturedVersionsCode != "P1000001" {
		t.Fatalf("expected captured code, got %s", repo.capturedVersionsCode)
	}

	if repo.capturedIncludeDeleted {
		t.Fatalf("expected includeDeleted=false by default")
	}
}

func TestResolver_PositionVersions_IncludeDeletedFlag(t *testing.T) {
	trueVal := true
	repo := &stubRepository{
		versionsFn: func(_ context.Context, _ uuid.UUID, _ string, includeDeleted bool) ([]dto.Position, error) {
			if !includeDeleted {
				t.Fatalf("expected includeDeleted true")
			}
			return []dto.Position{}, nil
		},
	}

	resolver := NewResolver(repo, newTestLogger(), &stubPermissionChecker{allow: true})

	_, err := resolver.PositionVersions(context.Background(), struct {
		Code           string
		IncludeDeleted *bool
	}{
		Code:           "P1000001",
		IncludeDeleted: &trueVal,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
