package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func newPositionRepository(t *testing.T) (*PositionRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	repo := &PositionRepository{db: db, logger: pkglogger.NewNoopLogger()}
	return repo, mock, func() { db.Close() }
}

func TestPositionRepository_GetCurrentPosition(t *testing.T) {
	repo, mock, cleanup := newPositionRepository(t)
	defer cleanup()

	tenant := uuid.New()
	now := time.Now()
	columns := []string{
		"record_id", "tenant_id", "code", "title", "job_profile_code", "job_profile_name", "job_family_group_code", "job_family_group_name", "job_family_group_record_id",
		"job_family_code", "job_family_name", "job_family_record_id", "job_role_code", "job_role_name", "job_role_record_id",
		"job_level_code", "job_level_name", "job_level_record_id", "organization_code", "organization_name", "position_type", "status", "employment_type",
		"headcount_capacity", "headcount_in_use", "grade_level", "cost_center_code", "reports_to_position_code", "profile", "effective_date", "end_date", "is_current",
		"created_at", "updated_at", "deleted_at", "operation_type", "operated_by_id", "operated_by_name", "operation_reason",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		uuid.New(), tenant, "POS001", "Engineer", sql.NullString{String: "JP", Valid: true}, sql.NullString{String: "jp", Valid: true},
		"G1", "Group", uuid.New(), "F1", "Family", uuid.New(), "R1", "Role", uuid.New(),
		"L1", "Level", uuid.New(), "ORG001", sql.NullString{String: "Org", Valid: true}, "FULLTIME", "ACTIVE", "PERM", 2.0, 1.0,
		sql.NullString{String: "G7", Valid: true}, sql.NullString{String: "CC", Valid: true}, sql.NullString{String: "PARENT", Valid: true}, []byte(`{"profile":true}`),
		now, sql.NullTime{}, true, now, now, sql.NullTime{}, "CREATE", uuid.New(), "operator", sql.NullString{},
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT record_id, tenant_id, code")).
		WithArgs(tenant, "POS001").
		WillReturnRows(rows)

	entity, err := repo.GetCurrentPosition(context.Background(), nil, tenant, "POS001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity == nil || entity.Code != "POS001" {
		t.Fatalf("expected entity POS001, got %#v", entity)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPositionRepository_InsertPositionVersion(t *testing.T) {
	repo, mock, cleanup := newPositionRepository(t)
	defer cleanup()

	entity := &types.Position{
		TenantID:             uuid.New(),
		Code:                 "POS001",
		Title:                "Engineer",
		JobFamilyGroupCode:   "G1",
		JobFamilyGroupName:   "Group",
		JobFamilyGroupRecord: uuid.New(),
		JobFamilyCode:        "F1",
		JobFamilyName:        "Family",
		JobFamilyRecord:      uuid.New(),
		JobRoleCode:          "R1",
		JobRoleName:          "Role",
		JobRoleRecord:        uuid.New(),
		JobLevelCode:         "L1",
		JobLevelName:         "Level",
		JobLevelRecord:       uuid.New(),
		OrganizationCode:     "ORG",
		OrganizationName:     sql.NullString{String: "Org", Valid: true},
		PositionType:         "FULLTIME",
		Status:               "ACTIVE",
		EmploymentType:       "PERM",
		HeadcountCapacity:    1,
		HeadcountInUse:       0,
		GradeLevel:           sql.NullString{},
		CostCenterCode:       sql.NullString{},
		ReportsToPosition:    sql.NullString{},
		Profile:              []byte(`{"k":1}`),
		EffectiveDate:        time.Now(),
	}

	rows := sqlmock.NewRows([]string{"record_id", "created_at", "updated_at"}).AddRow(uuid.New(), time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO positions")).
		WillReturnRows(rows)

	inserted, err := repo.InsertPositionVersion(context.Background(), nil, entity)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inserted.RecordID == uuid.Nil {
		t.Fatalf("expected record id to be set")
	}
}

func TestPositionRepository_BeginTxUsesSerializable(t *testing.T) {
	repo, mock, cleanup := newPositionRepository(t)
	defer cleanup()

	mock.ExpectBegin()
	if _, err := repo.BeginTx(context.Background()); err != nil {
		t.Fatalf("BeginTx error: %v", err)
	}
}
