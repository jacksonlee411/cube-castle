package repository

import (
	"context"
	"database/sql"
	"testing"

	"cube-castle/internal/types"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestComputeHierarchyForNew_Root(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrganizationRepository(db, nil)
	tenant := uuid.New()
	fields, err := repo.ComputeHierarchyForNew(context.Background(), tenant, "1000008", nil, "技术部")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if fields.Level != 1 || fields.CodePath != "/1000008" || fields.NamePath == "" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}

func TestComputeHierarchyForNew_DepthExceeded(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrganizationRepository(db, nil)
	tenant := uuid.New()
	parent := "1000000"

	// parent lookup returns level = OrganizationLevelMax
	mock.ExpectQuery("FROM organization_units").
		WithArgs(tenant.String(), parent).
		WillReturnRows(sqlmock.NewRows([]string{
			"code_path", "name_path", "level",
		}).AddRow("/1000000", "/集团", types.OrganizationLevelMax))

	fields, err := repo.ComputeHierarchyForNew(context.Background(), tenant, "1000008", &parent, "技术部")
	if err == nil || fields != nil {
		t.Fatalf("expected depth exceeded error")
	}
}

func TestComputeHierarchyForNew_ParentNotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewOrganizationRepository(db, nil)
	tenant := uuid.New()
	parent := "9999999"

	// parent lookup returns no rows
	mock.ExpectQuery("FROM organization_units").
		WithArgs(tenant.String(), parent).
		WillReturnError(sql.ErrNoRows)

	fields, err := repo.ComputeHierarchyForNew(context.Background(), tenant, "1000008", &parent, "技术部")
	if err == nil || fields != nil {
		t.Fatalf("expected error for missing parent")
	}
}
