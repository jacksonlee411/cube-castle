package repository

import (
	"context"
	"testing"

	pkglogger "cube-castle/pkg/logger"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetAssignmentHistoryRequiresPositionCode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("无法创建 sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil, pkglogger.NewNoopLogger(), AuditHistoryConfig{})
	_, err = repo.GetAssignmentHistory(context.Background(), uuid.New(), "   ", nil, nil, nil)
	if err == nil {
		t.Fatalf("当 positionCode 为空时应返回错误")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的 SQL 预期: %v", err)
	}
}

func TestGetAssignmentStatsEmptyDataset(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("无法创建 sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil, pkglogger.NewNoopLogger(), AuditHistoryConfig{})
	tenantID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"total_count",
		"active_count",
		"pending_count",
		"ended_count",
		"primary_count",
		"secondary_count",
		"acting_count",
		"last_updated",
	}).AddRow(0, 0, 0, 0, 0, 0, 0, nil)

	mock.ExpectQuery("SELECT\\s+COUNT\\(\\*\\) AS total_count[\\s\\S]*FROM position_assignments pa").
		WithArgs(tenantID.String()).
		WillReturnRows(rows)

	stats, err := repo.GetAssignmentStats(context.Background(), tenantID, "", "")
	if err != nil {
		t.Fatalf("期望无错误, 得到: %v", err)
	}

	if stats.TotalAssignments() != 0 {
		t.Errorf("期望 total 为 0, 实际为 %d", stats.TotalAssignments())
	}
	if stats.ActiveAssignments() != 0 {
		t.Errorf("期望 active 为 0, 实际为 %d", stats.ActiveAssignments())
	}
	if string(stats.LastUpdatedAt()) == "" {
		t.Errorf("期望 lastUpdatedAt 不为空")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的 SQL 预期: %v", err)
	}
}
