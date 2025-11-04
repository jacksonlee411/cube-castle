package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	pkglogger "cube-castle/pkg/logger"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestSoftDeleteOrganization(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewOrganizationRepository(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	deletedAt := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)

	query := regexp.QuoteMeta(`
		UPDATE organization_units
		SET status = 'DELETED',
		    is_current = false,
		    updated_at = NOW(),
		    deleted_at = $3,
		    deleted_by = $4,
		    deletion_reason = CASE WHEN $5 <> '' THEN $5 ELSE deletion_reason END
		WHERE tenant_id = $1 AND code = $2 AND status <> 'DELETED'
	`)

	mock.ExpectBegin()
	mock.ExpectExec(query).
		WithArgs(tenantID.String(), "1000001", deletedAt, "actor-1", "合规清理").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	if err := repo.SoftDeleteOrganization(context.Background(), tenantID, "1000001", deletedAt, "actor-1", "合规清理"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSoftDeleteOrganization_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewOrganizationRepository(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	deletedAt := time.Now().UTC()

	query := regexp.QuoteMeta(`
		UPDATE organization_units
		SET status = 'DELETED',
		    is_current = false,
		    updated_at = NOW(),
		    deleted_at = $3,
		    deleted_by = $4,
		    deletion_reason = CASE WHEN $5 <> '' THEN $5 ELSE deletion_reason END
		WHERE tenant_id = $1 AND code = $2 AND status <> 'DELETED'
	`)

	mock.ExpectBegin()
	mock.ExpectExec(query).
		WithArgs(tenantID.String(), "1000001", deletedAt, "actor-1", "").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err = repo.SoftDeleteOrganization(context.Background(), tenantID, "1000001", deletedAt, "actor-1", "")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestHasOtherNonDeletedVersions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewOrganizationRepository(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()

	query := regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM organization_units
		WHERE tenant_id = $1 AND code = $2 AND status <> 'DELETED' AND record_id <> $3
	`)

	mock.ExpectQuery(query).
		WithArgs(tenantID.String(), "1000001", "rec-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	remaining, err := repo.HasOtherNonDeletedVersions(context.Background(), tenantID, "1000001", "rec-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !remaining {
		t.Fatalf("expected remaining versions")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
