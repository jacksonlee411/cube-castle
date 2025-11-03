package repository

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetOrganizationStatsHandlesEmptyDataset(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("无法创建 sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil, log.New(io.Discard, "", 0), AuditHistoryConfig{})

	tenantID := uuid.New()
	fallbackDate := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	statsRows := sqlmock.NewRows([]string{
		"total_count",
		"active_count",
		"inactive_count",
		"planned_count",
		"deleted_count",
		"total_versions",
		"unique_orgs",
		"oldest_date",
		"newest_date",
		"type_stats",
		"status_stats",
		"level_stats",
	}).AddRow(
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		fallbackDate,
		fallbackDate,
		"[]",
		"[]",
		"[]",
	)

	mock.ExpectQuery("WITH status_stats").
		WithArgs(tenantID.String()).
		WillReturnRows(statsRows)

	stats, err := repo.GetOrganizationStats(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("期望无错误, 得到: %v", err)
	}

	if stats == nil {
		t.Fatalf("返回的统计结果不应为nil")
	}

	if stats.TotalCountField != 0 {
		t.Errorf("期望 totalCount 为 0, 实际为 %d", stats.TotalCountField)
	}
	if stats.TemporalStatsField.OldestEffectiveDateField != "1970-01-01" {
		t.Errorf("期望 oldestEffectiveDate 为 1970-01-01, 实际为 %s", stats.TemporalStatsField.OldestEffectiveDateField)
	}
	if stats.TemporalStatsField.NewestEffectiveDateField != "1970-01-01" {
		t.Errorf("期望 newestEffectiveDate 为 1970-01-01, 实际为 %s", stats.TemporalStatsField.NewestEffectiveDateField)
	}
	if len(stats.ByTypeField) != 0 {
		t.Errorf("期望 byType 为空, 实际长度为 %d", len(stats.ByTypeField))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("存在未满足的 SQL 预期: %v", err)
	}
}
