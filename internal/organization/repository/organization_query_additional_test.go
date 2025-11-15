package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestOrganizationRepository_GetByCode_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewOrganizationRepository(db, nil)

	tenant := uuid.New()
	now := time.Now().UTC()

	rows := sqlmock.NewRows([]string{
		"record_id", "tenant_id", "code", "parent_code", "name", "unit_type", "status",
		"level", "code_path", "name_path", "sort_order", "description", "created_at", "updated_at",
		"effective_date", "end_date", "change_reason",
	}).AddRow(
		"rec-1", tenant.String(), "1000008", sql.NullString{String: "1000000", Valid: true}, "技术部",
		"DEPARTMENT", "ACTIVE", 2, "/1000000/1000008", "/集团/技术部", 0, "desc", now, now,
		sql.NullTime{Time: now, Valid: true}, sql.NullTime{Valid: false}, sql.NullString{String: "创建", Valid: true},
	)

	mock.ExpectQuery("FROM organization_units").
		WithArgs(tenant.String(), "1000008").
		WillReturnRows(rows)

	got, err := repo.GetByCode(context.Background(), tenant, "1000008")
	if err != nil {
		t.Fatalf("GetByCode unexpected err: %v", err)
	}
	if got == nil || got.Code != "1000008" || got.ParentCode == nil || *got.ParentCode != "1000000" {
		t.Fatalf("unexpected organization: %#v", got)
	}
	if got.EffectiveDate == nil {
		t.Fatalf("expected effectiveDate set")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestOrganizationRepository_GetByCode_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrganizationRepository(db, nil)
	tenant := uuid.New()
	mock.ExpectQuery("FROM organization_units").
		WithArgs(tenant.String(), "4040000").
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetByCode(context.Background(), tenant, "4040000")
	if err == nil || got != nil {
		t.Fatalf("expected not found error")
	}
}

func TestOrganizationRepository_GetByRecordId_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrganizationRepository(db, nil)
	tenant := uuid.New()
	now := time.Now().UTC()

	rows := sqlmock.NewRows([]string{
		"record_id", "tenant_id", "code", "parent_code", "name", "unit_type", "status",
		"level", "code_path", "name_path", "sort_order", "description", "created_at", "updated_at",
		"effective_date", "end_date", "change_reason",
	}).AddRow(
		"rec-1", tenant.String(), "1000008", sql.NullString{Valid: false}, "技术部",
		"DEPARTMENT", "ACTIVE", 2, "/1000008", "/技术部", 0, "desc", now, now,
		sql.NullTime{Valid: false}, sql.NullTime{Valid: false}, sql.NullString{Valid: false},
	)

	mock.ExpectQuery("FROM organization_units").
		WithArgs(tenant.String(), "rec-1").
		WillReturnRows(rows)

	got, err := repo.GetByRecordId(context.Background(), tenant, "rec-1")
	if err != nil || got == nil || got.RecordID != "rec-1" {
		t.Fatalf("unexpected result: org=%#v err=%v", got, err)
	}
}

func TestOrganizationRepository_GetByRecordId_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrganizationRepository(db, nil)
	tenant := uuid.New()
	mock.ExpectQuery("FROM organization_units").
		WithArgs(tenant.String(), "missing").
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetByRecordId(context.Background(), tenant, "missing")
	if err == nil || got != nil {
		t.Fatalf("expected not found error")
	}
}

