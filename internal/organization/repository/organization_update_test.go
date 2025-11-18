package repository

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"testing"
	"time"

	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func newOrganizationRepositoryForTest(t *testing.T) (*OrganizationRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	repo := NewOrganizationRepository(db, pkglogger.NewNoopLogger())
	return repo, mock, func() { db.Close() }
}

func intPtr(v int) *int {
	return &v
}

func TestOrganizationRepositoryUpdate_NoFields(t *testing.T) {
	repo, _, cleanup := newOrganizationRepositoryForTest(t)
	defer cleanup()

	_, err := repo.Update(context.Background(), uuid.New(), "1000100", &types.UpdateOrganizationRequest{})
	if err == nil || !strings.Contains(err.Error(), "无字段需要更新") {
		t.Fatalf("expected missing fields error, got %v", err)
	}
}

func TestOrganizationRepositoryUpdate_Success(t *testing.T) {
	repo, mock, cleanup := newOrganizationRepositoryForTest(t)
	defer cleanup()

	tenantID := uuid.New()
	code := "1000200"
	name := "  Branch HQ "
	unitType := "DEPARTMENT"
	sortOrder := 12
	description := "Regional office"
	changeReason := "policy-sync"
	effective := types.NewDate(2025, time.January, 1)
	endDate := types.NewDate(2025, time.March, 31)
	req := &types.UpdateOrganizationRequest{
		Name:          &name,
		UnitType:      &unitType,
		SortOrder:     intPtr(sortOrder),
		Description:   &description,
		EffectiveDate: effective,
		EndDate:       endDate,
		ChangeReason:  &changeReason,
	}

	query := regexp.QuoteMeta("UPDATE organization_units")
	mock.ExpectQuery(query).
		WithArgs(
			tenantID.String(),
			code,
			"Branch HQ",
			unitType,
			sortOrder,
			description,
			*effective,
			*endDate,
			changeReason,
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"record_id", "tenant_id", "code", "parent_code", "name", "unit_type", "status",
			"level", "code_path", "name_path", "sort_order", "description", "created_at", "updated_at",
			"effective_date", "end_date", "change_reason",
		}).AddRow(
			uuid.NewString(),
			tenantID.String(),
			code,
			sql.NullString{},
			"Branch HQ",
			unitType,
			"ACTIVE",
			2,
			"/1000200",
			"/Branch HQ",
			sortOrder,
			description,
			time.Now(),
			time.Now(),
			nil,
			nil,
			nil,
		))

	entity, err := repo.Update(context.Background(), tenantID, code, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity == nil || entity.Code != code {
		t.Fatalf("expected entity with code %s, got %#v", code, entity)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestOrganizationRepositoryUpdate_NotFound(t *testing.T) {
	repo, mock, cleanup := newOrganizationRepositoryForTest(t)
	defer cleanup()

	tenantID := uuid.New()
	code := "1000300"
	name := "Lite Update"
	req := &types.UpdateOrganizationRequest{Name: &name}

	mock.ExpectQuery(regexp.QuoteMeta("UPDATE organization_units")).
		WithArgs(tenantID.String(), code, "Lite Update", sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.Update(context.Background(), tenantID, code, req)
	if err == nil || !strings.Contains(err.Error(), "组织不存在或已删除") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestOrganizationRepositoryUpdateByRecordID_WithParent(t *testing.T) {
	repo, mock, cleanup := newOrganizationRepositoryForTest(t)
	defer cleanup()

	tenantID := uuid.New()
	recordID := uuid.NewString()
	name := "  Branch Node "
	status := "ACTIVE"
	parent := " 1000000 "
	req := &types.UpdateOrganizationRequest{
		Name:       &name,
		Status:     &status,
		ParentCode: &parent,
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT code, name, level FROM organization_units")).
		WithArgs(tenantID.String(), recordID).
		WillReturnRows(sqlmock.NewRows([]string{"code", "name", "level"}).AddRow("2000001", "Old Branch", 2))

	mock.ExpectQuery("SELECT COALESCE\\(NULLIF\\(code_path").
		WithArgs(tenantID.String(), "1000000").
		WillReturnRows(sqlmock.NewRows([]string{"code_path", "name_path", "level"}).AddRow("/1000000", "/集团", 1))

	mock.ExpectQuery(regexp.QuoteMeta("UPDATE organization_units")).
		WithArgs(
			tenantID.String(),
			recordID,
			"Branch Node",
			status,
			"1000000",
			2,
			"/1000000/2000001",
			"/集团/Branch Node",
			sqlmock.AnyArg(),
		).
		WillReturnRows(sqlmock.NewRows([]string{
			"record_id", "tenant_id", "code", "parent_code", "name", "unit_type", "status",
			"level", "code_path", "name_path", "sort_order", "description", "created_at", "updated_at",
			"effective_date", "end_date", "change_reason",
		}).AddRow(
			recordID,
			tenantID.String(),
			"2000001",
			sql.NullString{String: "1000000", Valid: true},
			"Branch Node",
			"DEPARTMENT",
			status,
			2,
			"/1000000/2000001",
			"/集团/Branch Node",
			0,
			"",
			time.Now(),
			time.Now(),
			nil,
			nil,
			nil,
		))

	entity, err := repo.UpdateByRecordId(context.Background(), tenantID, recordID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity.ParentCode == nil || *entity.ParentCode != "1000000" {
		t.Fatalf("expected parent code 1000000, got %#v", entity.ParentCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestOrganizationRepositoryUpdateByRecordID_RecordMissing(t *testing.T) {
	repo, mock, cleanup := newOrganizationRepositoryForTest(t)
	defer cleanup()

	tenantID := uuid.New()
	recordID := uuid.NewString()
	name := "Ghost"
	parent := "1000000"
	req := &types.UpdateOrganizationRequest{
		Name:       &name,
		ParentCode: &parent,
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT code, name, level FROM organization_units")).
		WithArgs(tenantID.String(), recordID).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.UpdateByRecordId(context.Background(), tenantID, recordID, req)
	if err == nil || !strings.Contains(err.Error(), "记录不存在") {
		t.Fatalf("expected missing record error, got %v", err)
	}
}
