package graphql

import (
	"context"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"cube-castle-deployment-test/cmd/organization-query-service/internal/model"
	"github.com/google/uuid"
	graphqlgo "github.com/graph-gophers/graphql-go"
	sharedconfig "shared/config"
)

type stubPermissionChecker struct {
	allow     bool
	lastQuery string
	err       error
}

func (s *stubPermissionChecker) CheckQueryPermission(ctx context.Context, queryName string) error {
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
	positionsFn                      func(ctx context.Context, tenantID uuid.UUID, filter *model.PositionFilterInput, pagination *model.PaginationInput, sorting []model.PositionSortInput) (*model.PositionConnection, error)
	positionByCodeFn                 func(ctx context.Context, tenantID uuid.UUID, code string, asOfDate *string) (*model.Position, error)
	timelineFn                       func(ctx context.Context, tenantID uuid.UUID, code string, startDate, endDate *string) ([]model.PositionTimelineEntry, error)
	vacantFn                         func(ctx context.Context, tenantID uuid.UUID, filter *model.VacantPositionFilterInput, pagination *model.PaginationInput, sorting []model.VacantPositionSortInput) (*model.VacantPositionConnection, error)
	transferFn                       func(ctx context.Context, tenantID uuid.UUID, positionCode *string, organizationCode *string, pagination *model.PaginationInput) (*model.PositionTransferConnection, error)
	headcountFn                      func(ctx context.Context, tenantID uuid.UUID, organizationCode string, includeSubordinates bool) (*model.HeadcountStats, error)
	assignmentsFn                    func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *model.PositionAssignmentFilterInput, pagination *model.PaginationInput, sorting []model.PositionAssignmentSortInput) (*model.PositionAssignmentConnection, error)
	capturedSorting                  []model.PositionSortInput
	capturedFilter                   *model.PositionFilterInput
	capturedPagination               *model.PaginationInput
	capturedTenant                   uuid.UUID
	includeSubs                      bool
	capturedAssignmentFilter         *model.PositionAssignmentFilterInput
	capturedAssignmentSorting        []model.PositionAssignmentSortInput
	capturedPositionCode             string
	capturedVacantFilter             *model.VacantPositionFilterInput
	capturedVacantSorting            []model.VacantPositionSortInput
	capturedTransferPositionCode     *string
	capturedTransferOrganizationCode *string
}

func (s *stubRepository) GetOrganizations(ctx context.Context, tenantID uuid.UUID, filter *model.OrganizationFilter, pagination *model.PaginationInput) (*model.OrganizationConnection, error) {
	panic("GetOrganizations not expected")
}

func (s *stubRepository) GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*model.Organization, error) {
	panic("GetOrganization not expected")
}

func (s *stubRepository) GetOrganizationAtDate(ctx context.Context, tenantID uuid.UUID, code string, date string) (*model.Organization, error) {
	panic("GetOrganizationAtDate not expected")
}

func (s *stubRepository) GetOrganizationHistory(ctx context.Context, tenantID uuid.UUID, code string, fromDate string, toDate string) ([]model.Organization, error) {
	panic("GetOrganizationHistory not expected")
}

func (s *stubRepository) GetOrganizationVersions(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]model.Organization, error) {
	panic("GetOrganizationVersions not expected")
}

func (s *stubRepository) GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*model.OrganizationStats, error) {
	panic("GetOrganizationStats not expected")
}

func (s *stubRepository) GetOrganizationHierarchy(ctx context.Context, tenantID uuid.UUID, code string) (*model.OrganizationHierarchyData, error) {
	panic("GetOrganizationHierarchy not expected")
}

func (s *stubRepository) GetOrganizationSubtree(ctx context.Context, tenantID uuid.UUID, code string, maxDepth int) (*model.OrganizationHierarchyData, error) {
	panic("GetOrganizationSubtree not expected")
}

func (s *stubRepository) GetPositions(ctx context.Context, tenantID uuid.UUID, filter *model.PositionFilterInput, pagination *model.PaginationInput, sorting []model.PositionSortInput) (*model.PositionConnection, error) {
	if s.positionsFn == nil {
		panic("positionsFn not configured")
	}
	s.capturedTenant = tenantID
	s.capturedFilter = filter
	s.capturedPagination = pagination
	s.capturedSorting = sorting
	return s.positionsFn(ctx, tenantID, filter, pagination, sorting)
}

func (s *stubRepository) GetPositionByCode(ctx context.Context, tenantID uuid.UUID, code string, asOfDate *string) (*model.Position, error) {
	if s.positionByCodeFn == nil {
		panic("positionByCodeFn not configured")
	}
	return s.positionByCodeFn(ctx, tenantID, code, asOfDate)
}

func (s *stubRepository) GetPositionAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *model.PositionAssignmentFilterInput, pagination *model.PaginationInput, sorting []model.PositionAssignmentSortInput) (*model.PositionAssignmentConnection, error) {
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

func (s *stubRepository) GetPositionTimeline(ctx context.Context, tenantID uuid.UUID, code string, startDate, endDate *string) ([]model.PositionTimelineEntry, error) {
	if s.timelineFn == nil {
		panic("timelineFn not configured")
	}
	return s.timelineFn(ctx, tenantID, code, startDate, endDate)
}

func (s *stubRepository) GetVacantPositionConnection(ctx context.Context, tenantID uuid.UUID, filter *model.VacantPositionFilterInput, pagination *model.PaginationInput, sorting []model.VacantPositionSortInput) (*model.VacantPositionConnection, error) {
	if s.vacantFn == nil {
		panic("vacantFn not configured")
	}
	s.capturedTenant = tenantID
	s.capturedVacantFilter = filter
	s.capturedPagination = pagination
	s.capturedVacantSorting = sorting
	return s.vacantFn(ctx, tenantID, filter, pagination, sorting)
}

func (s *stubRepository) GetPositionHeadcountStats(ctx context.Context, tenantID uuid.UUID, organizationCode string, includeSubordinates bool) (*model.HeadcountStats, error) {
	if s.headcountFn == nil {
		panic("headcountFn not configured")
	}
	s.includeSubs = includeSubordinates
	return s.headcountFn(ctx, tenantID, organizationCode, includeSubordinates)
}

func (s *stubRepository) GetPositionTransfers(ctx context.Context, tenantID uuid.UUID, positionCode *string, organizationCode *string, pagination *model.PaginationInput) (*model.PositionTransferConnection, error) {
	if s.transferFn == nil {
		panic("transferFn not configured")
	}
	s.capturedTenant = tenantID
	s.capturedPagination = pagination
	s.capturedTransferPositionCode = positionCode
	s.capturedTransferOrganizationCode = organizationCode
	return s.transferFn(ctx, tenantID, positionCode, organizationCode, pagination)
}

func (s *stubRepository) GetJobFamilyGroups(ctx context.Context, tenantID uuid.UUID, includeInactive bool, asOfDate *string) ([]model.JobFamilyGroup, error) {
	panic("GetJobFamilyGroups not expected")
}

func (s *stubRepository) GetJobFamilies(ctx context.Context, tenantID uuid.UUID, groupCode string, includeInactive bool, asOfDate *string) ([]model.JobFamily, error) {
	panic("GetJobFamilies not expected")
}

func (s *stubRepository) GetJobRoles(ctx context.Context, tenantID uuid.UUID, familyCode string, includeInactive bool, asOfDate *string) ([]model.JobRole, error) {
	panic("GetJobRoles not expected")
}

func (s *stubRepository) GetJobLevels(ctx context.Context, tenantID uuid.UUID, roleCode string, includeInactive bool, asOfDate *string) ([]model.JobLevel, error) {
	panic("GetJobLevels not expected")
}

func (s *stubRepository) GetAuditHistory(ctx context.Context, tenantID uuid.UUID, recordID string, startDate, endDate, operation, userID *string, limit int) ([]model.AuditRecordData, error) {
	panic("GetAuditHistory not expected")
}

func (s *stubRepository) GetAuditLog(ctx context.Context, auditID string) (*model.AuditRecordData, error) {
	panic("GetAuditLog not expected")
}

func TestResolver_Positions_ForwardsParameters(t *testing.T) {
	filter := &model.PositionFilterInput{}
	orgCode := "1000001"
	filter.OrganizationCode = &orgCode
	pagination := &model.PaginationInput{Page: 2, PageSize: 20}
	sorting := []model.PositionSortInput{
		{Field: "code", Direction: "ASC"},
	}

	repo := &stubRepository{
		positionsFn: func(ctx context.Context, tenantID uuid.UUID, f *model.PositionFilterInput, p *model.PaginationInput, s []model.PositionSortInput) (*model.PositionConnection, error) {
			now := time.Now().UTC()
			position := model.Position{
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
			return &model.PositionConnection{
				DataField:       []model.Position{position},
				EdgesField:      []model.PositionEdge{{CursorField: "cursor-1", NodeField: position}},
				PaginationField: model.PaginationInfo{PageField: 2, PageSizeField: 20, TotalField: 1, HasNextField: false},
				TotalCountField: 1,
			}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, logDiscard(), perm)

	result, err := resolver.Positions(context.Background(), struct {
		Filter     *model.PositionFilterInput
		Pagination *model.PaginationInput
		Sorting    *[]model.PositionSortInput
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
		positionsFn: func(ctx context.Context, tenantID uuid.UUID, filter *model.PositionFilterInput, pagination *model.PaginationInput, sorting []model.PositionSortInput) (*model.PositionConnection, error) {
			t.Fatalf("repository should not be called when permission denied")
			return nil, nil
		},
	}
	perm := &stubPermissionChecker{allow: false}
	resolver := NewResolver(repo, logDiscard(), perm)

	_, err := resolver.Positions(context.Background(), struct {
		Filter     *model.PositionFilterInput
		Pagination *model.PaginationInput
		Sorting    *[]model.PositionSortInput
	}{})

	if err == nil || err.Error() != "INSUFFICIENT_PERMISSIONS" {
		t.Fatalf("expected INSUFFICIENT_PERMISSIONS, got %v", err)
	}
}

func TestResolver_PositionAssignments_ForwardsParameters(t *testing.T) {
	positionCode := "P1000001"
	employeeID := uuid.New().String()
	filter := &model.PositionAssignmentFilterInput{
		IncludeHistorical: false,
	}
	filter.EmployeeID = &employeeID
	status := "ACTIVE"
	filter.AssignmentStatus = &status

	pagination := &model.PaginationInput{Page: 1, PageSize: 10}
	sorting := []model.PositionAssignmentSortInput{
		{Field: "START_DATE", Direction: "ASC"},
	}

	repo := &stubRepository{
		assignmentsFn: func(ctx context.Context, tenantID uuid.UUID, code string, f *model.PositionAssignmentFilterInput, p *model.PaginationInput, s []model.PositionAssignmentSortInput) (*model.PositionAssignmentConnection, error) {
			now := time.Now().UTC()
			assignment := model.PositionAssignment{
				AssignmentIDField:     uuid.New().String(),
				TenantIDField:         tenantID.String(),
				PositionCodeField:     code,
				PositionRecordIDField: uuid.New().String(),
				EmployeeIDField:       employeeID,
				EmployeeNameField:     "Alice",
				AssignmentTypeField:   "PRIMARY",
				AssignmentStatusField: "ACTIVE",
				FTEField:              1,
				StartDateField:        now,
				IsCurrentField:        true,
				CreatedAtField:        now,
				UpdatedAtField:        now,
			}
			return &model.PositionAssignmentConnection{
				DataField: []model.PositionAssignment{assignment},
				EdgesField: []model.PositionAssignmentEdge{
					{CursorField: assignment.AssignmentIDField, NodeField: assignment},
				},
				PaginationField: model.PaginationInfo{
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
	resolver := NewResolver(repo, logDiscard(), perm)

	result, err := resolver.PositionAssignments(context.Background(), struct {
		PositionCode string
		Filter       *model.PositionAssignmentFilterInput
		Pagination   *model.PaginationInput
		Sorting      *[]model.PositionAssignmentSortInput
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

func TestResolver_VacantPositions_ForwardsParameters(t *testing.T) {
	orgCodes := []string{"1001001"}
	minDays := 30
	filter := &model.VacantPositionFilterInput{
		OrganizationCodes: &orgCodes,
		MinimumVacantDays: &minDays,
	}
	pagination := &model.PaginationInput{Page: 3, PageSize: 15}
	sorting := []model.VacantPositionSortInput{
		{Field: "VACANT_SINCE", Direction: "DESC"},
	}

	repo := &stubRepository{
		vacantFn: func(ctx context.Context, tenantID uuid.UUID, f *model.VacantPositionFilterInput, p *model.PaginationInput, s []model.VacantPositionSortInput) (*model.VacantPositionConnection, error) {
			return &model.VacantPositionConnection{
				DataField:  []model.VacantPosition{},
				EdgesField: []model.VacantPositionEdge{},
				PaginationField: model.PaginationInfo{
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
	resolver := NewResolver(repo, logDiscard(), perm)

	_, err := resolver.VacantPositions(context.Background(), struct {
		Filter     *model.VacantPositionFilterInput
		Pagination *model.PaginationInput
		Sorting    *[]model.VacantPositionSortInput
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
	filter := &model.VacantPositionFilterInput{
		OrganizationCodes: &orgCodes,
		AsOfDate:          &asOf,
	}
	repo := &stubRepository{
		vacantFn: func(ctx context.Context, tenantID uuid.UUID, f *model.VacantPositionFilterInput, pagination *model.PaginationInput, sorting []model.VacantPositionSortInput) (*model.VacantPositionConnection, error) {
			return &model.VacantPositionConnection{}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, logDiscard(), perm)

	_, err := resolver.VacantPositions(context.Background(), struct {
		Filter     *model.VacantPositionFilterInput
		Pagination *model.PaginationInput
		Sorting    *[]model.VacantPositionSortInput
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
	pagination := &model.PaginationInput{Page: 2, PageSize: 5}

	repo := &stubRepository{
		transferFn: func(ctx context.Context, tenantID uuid.UUID, posCode *string, org *string, p *model.PaginationInput) (*model.PositionTransferConnection, error) {
			transfer := model.PositionTransfer{
				TransferIDField:           uuid.New().String(),
				PositionCodeField:         positionCode,
				FromOrganizationCodeField: "1001000",
				ToOrganizationCodeField:   orgCode,
				EffectiveDateField:        time.Now().UTC(),
				CreatedAtField:            time.Now().UTC(),
				InitiatedByField: model.OperatedByData{
					IDField:   "user-1",
					NameField: "User One",
				},
			}
			return &model.PositionTransferConnection{
				DataField:       []model.PositionTransfer{transfer},
				EdgesField:      []model.PositionTransferEdge{{CursorField: transfer.TransferIDField, NodeField: transfer}},
				PaginationField: model.PaginationInfo{PageField: int(p.Page), PageSizeField: int(p.PageSize), TotalField: 1, HasNextField: false},
				TotalCountField: 1,
			}, nil
		},
	}

	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, logDiscard(), perm)

	_, err := resolver.PositionTransfers(context.Background(), struct {
		PositionCode     *string
		OrganizationCode *string
		Pagination       *model.PaginationInput
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
		headcountFn: func(ctx context.Context, tenantID uuid.UUID, organizationCode string, includeSubordinates bool) (*model.HeadcountStats, error) {
			return &model.HeadcountStats{OrganizationCodeField: organizationCode}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, logDiscard(), perm)

	falseVal := false
	_, err := resolver.PositionHeadcountStats(context.Background(), struct {
		OrganizationCode    string
		IncludeSubordinates graphqlgo.NullBool
	}{OrganizationCode: "1000001", IncludeSubordinates: graphqlgo.NullBool{Value: &falseVal, Set: true}})
	if err != nil {
		t.Fatalf("PositionHeadcountStats returned error: %v", err)
	}
	if repo.includeSubs {
		t.Fatalf("expected includeSubordinates=false when provided")
	}
}

func TestResolver_Position_ForwardsAsOfDate(t *testing.T) {
	expectedDate := "2025-01-01"
	repo := &stubRepository{
		positionByCodeFn: func(ctx context.Context, tenantID uuid.UUID, code string, asOfDate *string) (*model.Position, error) {
			if asOfDate == nil || *asOfDate != expectedDate {
				t.Fatalf("expected asOfDate %s, got %v", expectedDate, asOfDate)
			}
			return &model.Position{CodeField: code}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, logDiscard(), perm)

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
		timelineFn: func(ctx context.Context, tenantID uuid.UUID, code string, startDate, endDate *string) ([]model.PositionTimelineEntry, error) {
			if startDate == nil || *startDate != start {
				t.Fatalf("unexpected startDate %v", startDate)
			}
			if endDate == nil || *endDate != end {
				t.Fatalf("unexpected endDate %v", endDate)
			}
			return []model.PositionTimelineEntry{}, nil
		},
	}
	perm := &stubPermissionChecker{allow: true}
	resolver := NewResolver(repo, logDiscard(), perm)

	_, err := resolver.PositionTimeline(context.Background(), struct {
		Code      string
		StartDate *string
		EndDate   *string
	}{Code: "P1000001", StartDate: &start, EndDate: &end})
	if err != nil {
		t.Fatalf("PositionTimeline returned error: %v", err)
	}
}

func logDiscard() *log.Logger {
	return log.New(io.Discard, "", log.LstdFlags)
}
