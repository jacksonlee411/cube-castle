package repository

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetOrganizationStats_Minimal(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewPostgreSQLRepository(db, nil, nil, AuditHistoryConfig{})
	tenant := uuid.New()

	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{
		"total_count", "active_count", "inactive_count", "planned_count", "deleted_count",
		"total_versions", "unique_orgs", "oldest_date", "newest_date",
		"type_stats", "status_stats", "level_stats",
	}).AddRow(
		10, 8, 1, 1, 0,
		12, 10, now, now,
		`[]`, `[]`, `[]`,
	)

	mock.ExpectQuery("WITH status_stats").
		WithArgs(tenant.String()).
		WillReturnRows(rows)

	stats, err := repo.GetOrganizationStats(context.Background(), tenant)
	if err != nil || stats == nil || stats.TotalCountField != 10 || stats.ActiveCountField != 8 {
		t.Fatalf("unexpected stats: %#v err=%v", stats, err)
	}
}

func TestGetOrganizationSubtree_BuildsHierarchy(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewPostgreSQLRepository(db, nil, nil, AuditHistoryConfig{})
	tenant := uuid.New()

	rootCode := "1000000"
	cols := []string{"code", "name", "level", "hierarchy_depth", "code_path", "name_path", "parent_code"}
	rows := sqlmock.NewRows(cols).
		AddRow(rootCode, "集团", 1, 1, "/1000000", "/集团", nil).
		AddRow("1000001", "技术部", 2, 2, "/1000000/1000001", "/集团/技术部", rootCode)

	mock.ExpectQuery("WITH RECURSIVE subtree").
		WithArgs(tenant.String(), rootCode, 3).
		WillReturnRows(rows)

	tree, err := repo.GetOrganizationSubtree(context.Background(), tenant, rootCode, 3)
	if err != nil || tree == nil {
		t.Fatalf("unexpected: tree=%#v err=%v", tree, err)
	}
	if tree.CodeField != rootCode || !tree.IsRootField || tree.ChildrenCountField != 1 {
		t.Fatalf("unexpected tree root: %#v", tree)
	}
	if len(tree.ChildrenField) != 1 || tree.ChildrenField[0].CodeField != "1000001" {
		t.Fatalf("unexpected child: %#v", tree.ChildrenField)
	}
}
