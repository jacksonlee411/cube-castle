package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"

	pkglogger "cube-castle/pkg/logger"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestOrganizationRepository_CountNonDeletedChildren(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewOrganizationRepository(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	query := regexp.QuoteMeta(`
		SELECT COUNT(DISTINCT code)
		FROM organization_units
		WHERE tenant_id = $1 AND parent_code = $2 AND status <> 'DELETED'
		  AND (is_current = true OR effective_date >= CURRENT_DATE)
	`)

	mock.ExpectQuery(query).
		WithArgs(tenantID.String(), "1000001").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	count, err := repo.CountNonDeletedChildren(context.Background(), tenantID, "1000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count=2, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet mock expectations: %v", err)
	}
}

func TestOrganizationRepository_CountNonDeletedChildren_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewOrganizationRepository(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	query := regexp.QuoteMeta(`
		SELECT COUNT(DISTINCT code)
		FROM organization_units
		WHERE tenant_id = $1 AND parent_code = $2 AND status <> 'DELETED'
		  AND (is_current = true OR effective_date >= CURRENT_DATE)
	`)

	mock.ExpectQuery(query).
		WithArgs(tenantID.String(), "1000001").
		WillReturnError(errors.New("db error"))

	if _, err := repo.CountNonDeletedChildren(context.Background(), tenantID, "1000001"); err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet mock expectations: %v", err)
	}
}
