package repository

import (
	"context"
	"testing"
	"time"

	"cube-castle/internal/organization/dto"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// 命中多种过滤分支（status/unitType/parentCode/searchText/asOfDate/excludeDescendants/codes/excludeCodes/includeDisabledAncestors）
func TestPostgreSQLRepository_GetOrganizations_FilterBranches(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil, nil, AuditHistoryConfig{})
	tenant := uuid.New()

	status := "ACTIVE"
	unitType := "DEPARTMENT"
	parentCode := "1000001"
	search := "技术"
	asOf := "2025-11-16"
	exclDesc := "1000000"
	includeCodes := []string{"1000001", "1000002"}
	excludeCodes := []string{"1000099"}

	filter := &dto.OrganizationFilter{
		Status:                   &status,
		UnitType:                 &unitType,
		ParentCode:               &parentCode,
		SearchText:               &search,
		AsOfDate:                 &asOf,
		ExcludeDescendantsOf:     &exclDesc,
		Codes:                    &includeCodes,
		ExcludeCodes:             &excludeCodes,
		IncludeDisabledAncestors: true,
	}

	// Count query
	mock.ExpectQuery("WITH parent_path").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Data query (2 rows) — 复用最小用例的字段顺序，确保列数完全一致
	now := time.Now().UTC()
	var (
		tenantID               = tenant.String()
		sortOrder     *int     = nil
		desc          *string  = nil
		profile       *string  = nil
		eff                    = now
		endDate       *time.Time = nil
		isCurrent              = true
		changeReason  *string  = nil
		deletedAt     *time.Time = nil
		deletedBy     *string  = nil
		deletionReason *string = nil
		suspendAt       *time.Time = nil
		suspendBy       *string  = nil
		suspendReason   *string  = nil
	)
	rows := sqlmock.NewRows([]string{
		"record_id", "tenant_id", "code", "parent_code", "name",
		"unit_type", "status", "level", "code_path", "name_path", "sort_order",
		"description", "profile", "created_at", "updated_at",
		"effective_date", "end_date", "is_current",
		"change_reason", "deleted_at", "deleted_by", "deletion_reason",
		"suspended_at", "suspended_by", "suspension_reason", "children_count",
	}).AddRow(
		"rec-1", tenantID, "1000001", parentCode, "技术一部",
		unitType, status, 2, "/1000001", "/技术一部", sortOrder,
		desc, profile, now, now,
		eff, endDate, isCurrent,
		changeReason, deletedAt, deletedBy, deletionReason,
		suspendAt, suspendBy, suspendReason, 0,
	).AddRow(
		"rec-2", tenantID, "1000002", parentCode, "技术二部",
		unitType, status, 2, "/1000002", "/技术二部", sortOrder,
		desc, profile, now, now,
		eff, endDate, isCurrent,
		changeReason, deletedAt, deletedBy, deletionReason,
		suspendAt, suspendBy, suspendReason, 0,
	)
	mock.ExpectQuery("WITH parent_path").
		WillReturnRows(rows)

	got, err := repo.GetOrganizations(context.Background(), tenant, filter, &dto.PaginationInput{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetOrganizations err: %v", err)
	}
	if got == nil || len(got.DataField) != 2 || got.PaginationField.TotalField != 2 {
		t.Fatalf("unexpected result pagination or data: %#v", got)
	}
	if got.DataField[0].CodeField != "1000001" || got.DataField[1].CodeField != "1000002" {
		t.Fatalf("unexpected codes order: %v", got.DataField)
	}

	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatalf("unmet expectations: %v", e)
	}
}
