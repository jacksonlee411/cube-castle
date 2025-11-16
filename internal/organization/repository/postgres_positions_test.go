package repository

import (
	"context"
	"testing"
	"time"

	"cube-castle/internal/organization/dto"
	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestGetPositions_MinimalSuccess(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewPostgreSQLRepository(db, nil, nil, AuditHistoryConfig{})
	tenant := uuid.New()

	// Count query
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Data query: columns must match scanPosition order
	now := time.Now().UTC()
	cols := []string{
		"record_id", "tenant_id", "code", "title",
		"job_profile_code", "job_profile_name",
		"job_family_group_code", "job_family_code", "job_role_code", "job_level_code",
		"organization_code", "position_type", "employment_type", "grade_level",
		"headcount_capacity", "headcount_in_use", "reports_to_position_code", "status",
		"effective_date", "end_date", "is_current", "created_at", "updated_at",
		"job_family_group_name", "job_family_name", "job_role_name", "job_level_name",
		"organization_name",
	}
	row := sqlmock.NewRows(cols).AddRow(
		"rec-1", tenant.String(), "P10001", "研发工程师",
		nil, nil,
		"", "", "", "",
		"1000000", "FULL_TIME", "REGULAR", nil,
		1, 0, nil, "ACTIVE",
		now, nil, true, now, now,
		nil, nil, nil, nil,
		"集团",
	)
	mock.ExpectQuery("SELECT p.record_id").WillReturnRows(row)

	conn, err := repo.GetPositions(context.Background(), tenant, nil, &dto.PaginationInput{Page: 1, PageSize: 10}, nil)
	if err != nil {
		t.Fatalf("GetPositions err: %v", err)
	}
	if conn == nil || len(conn.DataField) != 1 || conn.TotalCountField != 1 {
		t.Fatalf("unexpected connection: %#v", conn)
	}
	if conn.DataField[0].CodeField != "P10001" || conn.DataField[0].OrganizationCodeField != "1000000" {
		t.Fatalf("unexpected node: %#v", conn.DataField[0])
	}
}
