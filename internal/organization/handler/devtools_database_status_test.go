package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestDatabaseStatus_Disconnected(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// 模拟 Ping 失败
	mock.ExpectPing().WillReturnError(assertionError("down"))

	h := NewDevToolsHandler(nil, nil, true, db)
	req := httptest.NewRequest(http.MethodGet, "/dev/database-status", nil)
	rec := httptest.NewRecorder()

	h.DatabaseStatus(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "DATABASE_DISCONNECTED") {
		t.Fatalf("expected DATABASE_DISCONNECTED in body, got %s", body)
	}
	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatalf("unmet expectations: %v", e)
	}
}

func TestDatabaseStatus_Success(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Ping 成功
	mock.ExpectPing()
	// 三张表的 count
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM organization_units`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM organization_units_history`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM audit_logs`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	// 数据库大小
	mock.ExpectQuery(`SELECT pg_size_pretty`).WillReturnRows(sqlmock.NewRows([]string{"pg_size_pretty"}).AddRow("1 MB"))

	h := NewDevToolsHandler(nil, nil, true, db)
	req := httptest.NewRequest(http.MethodGet, "/dev/database-status", nil)
	rec := httptest.NewRecorder()

	h.DatabaseStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, kw := range []string{`"connected":true`, `"organization_units":{"count":5`, `"audit_logs":{"count":2`, `"database_size":"1 MB"`} {
		if !strings.Contains(body, kw) {
			t.Fatalf("expected %q in response body, got %s", kw, body)
		}
	}
	if e := mock.ExpectationsWereMet(); e != nil {
		t.Fatalf("unmet expectations: %v", e)
	}
}

// 简单断言错误类型（避免引入额外依赖）
type assertionError string

func (e assertionError) Error() string { return string(e) }
