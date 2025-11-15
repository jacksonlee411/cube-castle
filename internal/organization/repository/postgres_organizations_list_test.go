package repository

import (
	"context"
	"testing"
	"time"

	"cube-castle/internal/organization/dto"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestPostgreSQLRepository_GetOrganizations_Minimal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil, nil, AuditHistoryConfig{})
	tenant := uuid.New()

	// Count query
	mock.ExpectQuery("WITH parent_path").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Data query - provide one row with the exact scan order
	now := time.Now().UTC()
	var (
		recordID    = "rec-1"
		tenantID    = tenant.String()
		code        = "1000008"
		parentCode  *string = nil
		name                 = "技术部"
		unitType             = "DEPARTMENT"
		status               = "ACTIVE"
		level                = 2
		codePath             = "/1000008"
		namePath             = "/技术部"
		sortOrder   *int     = nil
		desc        *string  = nil
		profile     *string  = nil
		created              = now
		updated              = now
		eff                  = now
		endDate     *time.Time = nil
		isCurrent            = true
		changeReason *string = nil
		deletedAt    *time.Time = nil
		deletedBy    *string  = nil
		deletionReason *string = nil
		suspendAt      *time.Time = nil
		suspendBy      *string  = nil
		suspendReason  *string  = nil
		childrenCount          = 0
	)

	row := sqlmock.NewRows([]string{
		"record_id", "tenant_id", "code", "parent_code", "name",
		"unit_type", "status", "level", "code_path", "name_path", "sort_order",
		"description", "profile", "created_at", "updated_at",
		"effective_date", "end_date", "is_current",
		"change_reason", "deleted_at", "deleted_by", "deletion_reason",
		"suspended_at", "suspended_by", "suspension_reason", "children_count",
	}).AddRow(
		recordID, tenantID, code, parentCode, name,
		unitType, status, level, codePath, namePath, sortOrder,
		desc, profile, created, updated,
		eff, endDate, isCurrent,
		changeReason, deletedAt, deletedBy, deletionReason,
		suspendAt, suspendBy, suspendReason, childrenCount,
	)

	mock.ExpectQuery("WITH parent_path").
		WillReturnRows(row)

	result, err := repo.GetOrganizations(context.Background(), tenant, nil, &dto.PaginationInput{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetOrganizations err: %v", err)
	}
	if result == nil || len(result.DataField) != 1 || result.PaginationField.TotalField < 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.DataField[0].CodeField != code {
		t.Fatalf("unexpected code: %s", result.DataField[0].CodeField)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

