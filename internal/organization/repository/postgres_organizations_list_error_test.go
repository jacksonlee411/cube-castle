package repository

import (
	"context"
	"testing"
	"time"

	"cube-castle/internal/organization/dto"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// 简单断言错误类型（避免引入额外依赖）
type assertionError string

func (e assertionError) Error() string { return string(e) }

// 扫描列不匹配导致的错误（模拟数据列不足）
func TestPostgreSQLRepository_GetOrganizations_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil, nil, AuditHistoryConfig{})
	tenant := uuid.New()

	// count 正常
	mock.ExpectQuery("WITH parent_path").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// data 返回列不足，触发 Scan 错误
	now := time.Now().UTC()
	rows := sqlmock.NewRows([]string{"code", "created_at"}).AddRow("1000001", now)
	mock.ExpectQuery("WITH parent_path").
		WillReturnRows(rows)

	_, err = repo.GetOrganizations(context.Background(), tenant, nil, &dto.PaginationInput{Page: 1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected scan error, got nil")
	}
	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatalf("unmet expectations: %v", e)
	}
}

// count 查询错误路径
func TestPostgreSQLRepository_GetOrganizations_CountQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil, nil, AuditHistoryConfig{})
	tenant := uuid.New()

	mock.ExpectQuery("WITH parent_path").
		WillReturnError(assertionError("count failed"))

	_, err = repo.GetOrganizations(context.Background(), tenant, nil, &dto.PaginationInput{Page: 1, PageSize: 10})
	if err == nil {
		t.Fatalf("expected count error, got nil")
	}
	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatalf("unmet expectations: %v", e)
	}
}
