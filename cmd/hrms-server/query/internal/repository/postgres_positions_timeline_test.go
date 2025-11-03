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

func TestGetPositionTimelineUsesChangeReasonAlias(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := NewPostgreSQLRepository(db, nil, log.New(io.Discard, "", 0), AuditHistoryConfig{})

	tenantID := uuid.New()
	code := "P9000004"
	now := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

	columns := []string{
		"record_id",
		"status",
		"title",
		"effective_date",
		"end_date",
		"is_current",
		"change_reason",
		"timeline_category",
		"assignment_type",
		"assignment_status",
	}

	rows := sqlmock.NewRows(columns).
		AddRow(
			"version-record",
			"ACTIVE",
			"产品设计师",
			now,
			nil,
			true,
			"AI 项目体验升级",
			"POSITION_VERSION",
			nil,
			nil,
		).
		AddRow(
			"assignment-record",
			"ACTIVE",
			"王蕾",
			now,
			nil,
			true,
			"初始职位同步数据",
			"POSITION_ASSIGNMENT",
			"PRIMARY",
			"ACTIVE",
		)

	mock.ExpectQuery(`(?s)WITH timeline AS \(.*p\.operation_reason AS change_reason.*FROM timeline`).
		WithArgs(tenantID.String(), code).
		WillReturnRows(rows)

	result, err := repo.GetPositionTimeline(context.Background(), tenantID, code, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 timeline entries, got %d", len(result))
	}

	if result[0].ChangeReasonField == nil || *result[0].ChangeReasonField != "AI 项目体验升级" {
		t.Fatalf("unexpected change reason for first entry: %+v", result[0].ChangeReasonField)
	}

	if result[1].ChangeReasonField == nil || *result[1].ChangeReasonField != "初始职位同步数据" {
		t.Fatalf("unexpected change reason for second entry: %+v", result[1].ChangeReasonField)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
