package repository

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestInsertJobLevelVersionCopiesParentMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewJobCatalogRepository(db, testLogger())
	tenantID := uuid.New()
	levelCode := "jl-demo"
	parentRecord := uuid.New()
	parentRole := "JR-DEMO"
	parentLevelRank := "7"
	parentSalary := []byte(`{"min":1,"max":5}`)
	roleRecord := uuid.New()
	parentRows := sqlmock.NewRows([]string{
		"record_id", "tenant_id", "level_code", "role_code", "parent_record_id",
		"level_rank", "name", "description", "salary_band", "status",
		"effective_date", "end_date", "is_current",
	}).AddRow(
		parentRecord,
		tenantID,
		strings.ToUpper(levelCode),
		parentRole,
		roleRecord,
		parentLevelRank,
		"Parent Level",
		nil,
		parentSalary,
		"ACTIVE",
		time.Now(),
		nil,
		false,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT record_id, tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current
FROM job_levels WHERE tenant_id = $1 AND record_id = $2 LIMIT 1`)).
		WithArgs(tenantID, parentRecord).
		WillReturnRows(parentRows)

	effectiveDate := time.Now().UTC().Truncate(24 * time.Hour).Format("2006-01-02")
	newRecord := uuid.New()
	description := "new version"
	req := &types.JobCatalogVersionRequest{
		Name:           "Latest Level",
		Status:         "ACTIVE",
		EffectiveDate:  effectiveDate,
		Description:    &description,
		ParentRecordID: stringPointer(parentRecord.String()),
	}

	insertRows := sqlmock.NewRows([]string{
		"record_id", "tenant_id", "level_code", "role_code", "parent_record_id",
		"level_rank", "name", "description", "salary_band", "status",
		"effective_date", "end_date", "is_current",
	}).AddRow(
		newRecord,
		tenantID,
		strings.ToUpper(levelCode),
		parentRole,
		roleRecord,
		parentLevelRank,
		req.Name,
		description,
		parentSalary,
		"ACTIVE",
		time.Now(),
		nil,
		true,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO job_levels (
tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULL,$11,NOW(),NOW())
RETURNING record_id, tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current`)).
		WithArgs(
			tenantID,
			strings.ToUpper(levelCode),
			parentRole,
			roleRecord,
			parentLevelRank,
			req.Name,
			description,
			parentSalary,
			req.Status,
			sqlmock.AnyArg(),
			false,
		).
		WillReturnRows(insertRows)

	timelineRows := sqlmock.NewRows([]string{"record_id", "effective_date", "end_date", "is_current"}).
		AddRow(newRecord, time.Now(), nil, true)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT record_id, effective_date, end_date, is_current FROM job_levels WHERE tenant_id = $1 AND level_code = $2 ORDER BY effective_date FOR UPDATE`)).
		WithArgs(tenantID, strings.ToUpper(levelCode)).
		WillReturnRows(timelineRows)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE job_levels SET end_date = $2, is_current = $3, updated_at = NOW() WHERE record_id = $1`)).
		WithArgs(newRecord, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	entity, err := repo.InsertJobLevelVersion(ctxWithTimeout(), nil, tenantID, levelCode, parentRecord, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity == nil {
		t.Fatalf("expected entity")
	}
	if entity.LevelRank != parentLevelRank {
		t.Fatalf("expected level rank %s, got %s", parentLevelRank, entity.LevelRank)
	}
	if entity.RoleCode != parentRole {
		t.Fatalf("expected role code %s, got %s", parentRole, entity.RoleCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func ctxWithTimeout() context.Context {
	return context.Background()
}

func stringPointer(val string) *string {
	return &val
}

func testLogger() pkglogger.Logger {
	return pkglogger.NewNoopLogger()
}

func TestInsertJobLevelVersionDetectsParentMismatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewJobCatalogRepository(db, testLogger())
	tenantID := uuid.New()
	levelCode := "JL-OK"
	parentRecord := uuid.New()

	parentRows := sqlmock.NewRows([]string{
		"record_id", "tenant_id", "level_code", "role_code", "parent_record_id",
		"level_rank", "name", "description", "salary_band", "status",
		"effective_date", "end_date", "is_current",
	}).AddRow(
		parentRecord,
		tenantID,
		"JL-DIFFERENT",
		"JR-OTHER",
		uuid.New(),
		"9",
		"Parent",
		nil,
		[]byte("{}"),
		"ACTIVE",
		time.Now(),
		nil,
		true,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT record_id, tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current
FROM job_levels WHERE tenant_id = $1 AND record_id = $2 LIMIT 1`)).
		WithArgs(tenantID, parentRecord).
		WillReturnRows(parentRows)

	req := &types.JobCatalogVersionRequest{
		Name:           "Invalid",
		Status:         "ACTIVE",
		EffectiveDate:  time.Now().UTC().Format("2006-01-02"),
		ParentRecordID: stringPointer(parentRecord.String()),
	}

	_, err = repo.InsertJobLevelVersion(ctxWithTimeout(), nil, tenantID, levelCode, parentRecord, req)
	if err == nil || !strings.Contains(err.Error(), "parent record mismatch") {
		t.Fatalf("expected parent mismatch error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
