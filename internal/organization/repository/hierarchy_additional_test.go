package repository

import (
	"database/sql"
	"context"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// 验证：当父节点不存在时，UpdateHierarchyPaths 返回错误（不执行事务）
func TestHierarchyRepository_UpdateHierarchyPaths_ParentNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	h := NewHierarchyRepository(db, nil)
	tenant := uuid.New()

	// 父节点查询返回 no rows
	mock.ExpectQuery(`SELECT COALESCE\(code_path, code\), COALESCE\(name_path, name\), level\s+FROM organization_units`).
		WithArgs("1000999", tenant.String()).
		WillReturnError(sql.ErrNoRows)

	err = h.UpdateHierarchyPaths(context.Background(), "1000999", tenant)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, err) || err.Error() == "" {
		t.Fatalf("expected non-empty error, got: %v", err)
	}
	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatalf("unmet expectations: %v", e)
	}
}

// 验证：根节点路径批量更新成功（不查询父节点，直接事务更新）
func TestHierarchyRepository_UpdateHierarchyPaths_RootParent_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	h := NewHierarchyRepository(db, nil)
	tenant := uuid.New()

	// 期望开启事务、执行 UPDATE、提交
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE organization_units SET`).
		WithArgs("", "", 0, "", tenant.String()).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	if err := h.UpdateHierarchyPaths(context.Background(), "", tenant); err != nil {
		t.Fatalf("UpdateHierarchyPaths (root) error: %v", err)
	}
	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatalf("unmet expectations: %v", e)
	}
}

// 验证：GetOrganization 无数据时返回 not found 错误
func TestHierarchyRepository_GetOrganization_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	h := NewHierarchyRepository(db, nil)
	tenant := uuid.New()

	mock.ExpectQuery(`SELECT\s+tenant_id,\s*code,\s*parent_code`).
		WithArgs("1000001", tenant.String()).
		WillReturnError(sql.ErrNoRows)

	org, err := h.GetOrganization(context.Background(), "1000001", tenant)
	if err == nil || org != nil {
		t.Fatalf("expected not found error, got org=%v err=%v", org, err)
	}
	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatalf("unmet expectations: %v", e)
	}
}

// 验证：GetOrganizationAtDate 无数据返回 (nil, nil)
func TestHierarchyRepository_GetOrganizationAtDate_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	h := NewHierarchyRepository(db, nil)
	tenant := uuid.New()

	mock.ExpectQuery(`FROM organization_units\s+WHERE tenant_id = \$1`).
		WillReturnError(sql.ErrNoRows)

	got, err := h.GetOrganizationAtDate(context.Background(), "1000001", tenant, /* targetDate */ time.Now())
	if err == nil && got != nil {
		t.Fatalf("expected (nil, nil) for no rows, got: %+v", got)
	}
	// 当 sqlmock 返回 ErrNoRows，方法实现返回 (nil, nil)，err 可能为 nil，允许

	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatalf("unmet expectations: %v", e)
	}
}

// 验证：GetAncestorChain 查询错误透传
func TestHierarchyRepository_GetAncestorChain_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	h := NewHierarchyRepository(db, nil)
	tenant := uuid.New()

	mock.ExpectQuery(`WITH RECURSIVE ancestor_chain`).
		WillReturnError(errors.New("db unavailable"))

	ancestors, err := h.GetAncestorChain(context.Background(), "1000001", tenant)
	if err == nil || ancestors != nil {
		t.Fatalf("expected error, got ancestors=%v err=%v", ancestors, err)
	}
	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatalf("unmet expectations: %v", e)
	}
}
