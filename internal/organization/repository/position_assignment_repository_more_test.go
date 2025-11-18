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

func newAssignmentRepo(t *testing.T) (*PositionAssignmentRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	repo := &PositionAssignmentRepository{db: db, logger: pkglogger.NewNoopLogger()}
	return repo, mock, func() { db.Close() }
}

func TestPositionAssignmentRepository_CreateAndGet(t *testing.T) {
	repo, mock, cleanup := newAssignmentRepo(t)
	defer cleanup()

	tenant := uuid.New()
	now := time.Now()
	assignment := &types.PositionAssignment{
		TenantID:         tenant,
		PositionCode:     "POS001",
		PositionRecordID: uuid.New(),
		EmployeeID:       uuid.New(),
		EmployeeName:     "Alice",
		AssignmentType:   "PRIMARY",
		AssignmentStatus: "ACTIVE",
		FTE:              1,
		EffectiveDate:    now,
		IsCurrent:        true,
	}

	rows := sqlmock.NewRows([]string{"assignment_id", "assignment_status", "is_current", "created_at", "updated_at"}).AddRow(uuid.New(), "ACTIVE", true, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO position_assignments")).WillReturnRows(rows)

	created, err := repo.CreateAssignment(context.Background(), nil, assignment)
	if err != nil {
		t.Fatalf("CreateAssignment error: %v", err)
	}
	if created.AssignmentID == uuid.Nil {
		t.Fatalf("expected assignment id to be set")
	}

	getRows := sqlmock.NewRows([]string{
		"assignment_id", "tenant_id", "position_code", "position_record_id", "employee_id", "employee_name", "employee_number",
		"assignment_type", "assignment_status", "fte", "effective_date", "end_date", "acting_until", "auto_revert", "reminder_sent_at", "is_current", "notes", "created_at", "updated_at",
	}).AddRow(
		created.AssignmentID, tenant, "POS001", assignment.PositionRecordID, assignment.EmployeeID, assignment.EmployeeName, sql.NullString{},
		assignment.AssignmentType, assignment.AssignmentStatus, assignment.FTE, now, sql.NullTime{}, sql.NullTime{}, false, sql.NullTime{}, true, sql.NullString{}, now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT assignment_id, tenant_id, position_code")).
		WithArgs(tenant, created.AssignmentID).
		WillReturnRows(getRows)

	fetched, err := repo.GetByID(context.Background(), nil, tenant, created.AssignmentID)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if fetched == nil || fetched.AssignmentID != created.AssignmentID {
		t.Fatalf("unexpected fetched assignment: %#v", fetched)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPositionAssignmentRepository_CloseAssignment(t *testing.T) {
	repo, mock, cleanup := newAssignmentRepo(t)
	defer cleanup()

	tenant := uuid.New()
	assignID := uuid.New()
	now := time.Now()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE position_assignments")).
		WithArgs(tenant, assignID, now, true, sql.NullString{}).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.CloseAssignment(context.Background(), nil, tenant, assignID, now, nil); err != nil {
		t.Fatalf("CloseAssignment error: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE position_assignments")).
		WithArgs(tenant, assignID, now, false, sql.NullString{String: "comment", Valid: true}).
		WillReturnResult(sqlmock.NewResult(0, 1))

	note := " comment "
	if err := repo.CloseAssignment(context.Background(), nil, tenant, assignID, now, &note); err != nil {
		t.Fatalf("CloseAssignment with note error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPositionAssignmentRepository_ScanErrors(t *testing.T) {
	repo, mock, cleanup := newAssignmentRepo(t)
	defer cleanup()

	tenant := uuid.New()
	targetID := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT assignment_id, tenant_id, position_code")).
		WithArgs(tenant, targetID).
		WillReturnError(sql.ErrNoRows)

	assignment, err := repo.GetByID(context.Background(), nil, tenant, targetID)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if assignment != nil {
		t.Fatalf("expected nil assignment")
	}
}
