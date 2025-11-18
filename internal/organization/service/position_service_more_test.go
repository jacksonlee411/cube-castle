package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"cube-castle/internal/organization/events"
	"cube-castle/internal/types"
	"cube-castle/pkg/database"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestBuildPositionEntity_SuccessAndHeadcount(t *testing.T) {
	svc := &PositionService{}
	now := time.Now().UTC()
	effDate := now.Format("2006-01-02")
	profile := `{"key":"value"}`
	headcountInUse := 1.0
	req := &types.PositionRequest{
		Title:              " Engineer ",
		JobFamilyGroupCode: "G1",
		JobFamilyCode:      "F1",
		JobRoleCode:        "R1",
		JobLevelCode:       "L1",
		OrganizationCode:   "1000000",
		PositionType:       "fulltime",
		EmploymentType:     "permanent",
		HeadcountCapacity:  2.0,
		HeadcountInUse:     &headcountInUse,
		EffectiveDate:      effDate,
		Profile:            &profile,
		OperationReason:    "reason",
	}

	catalog := &jobCatalogSnapshot{
		group:  &types.JobFamilyGroup{Code: "G1", Name: "Group1", RecordID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")},
		family: &types.JobFamily{Code: "F1", Name: "Family", RecordID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")},
		role:   &types.JobRole{Code: "R1", Name: "Role", RecordID: uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")},
		level:  &types.JobLevel{Code: "L1", Name: "Level", RecordID: uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")},
	}
	org := &types.Organization{Code: "1000000", Name: "OrgName"}

	entity, err := svc.buildPositionEntity(uuid.New(), "POS001", req, catalog, org, types.OperatedByInfo{ID: uuid.NewString(), Name: "Alice"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity.Title != "Engineer" {
		t.Fatalf("title should be trimmed, got %q", entity.Title)
	}
	if entityHeadcount := entity.HeadcountCapacity - entity.HeadcountInUse; entityHeadcount <= 0 {
		t.Fatalf("headcount should remain positive, got %v", entityHeadcount)
	}
	if !strings.EqualFold(entity.Status, "PLANNED") {
		t.Fatalf("expected default status PLANNED, got %s", entity.Status)
	}
	if !entity.IsCurrent {
		t.Fatalf("effective date on or before today should mark entity as current")
	}
}

func TestBuildPositionEntity_Errors(t *testing.T) {
	svc := &PositionService{}
	catalog := &jobCatalogSnapshot{
		group:  &types.JobFamilyGroup{Code: "G1", Name: "Group1", RecordID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")},
		family: &types.JobFamily{Code: "F1", Name: "Family", RecordID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")},
		role:   &types.JobRole{Code: "R1", Name: "Role", RecordID: uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")},
		level:  &types.JobLevel{Code: "L1", Name: "Level", RecordID: uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")},
	}
	org := &types.Organization{Code: "1000000", Name: "OrgName"}
	today := time.Now().UTC().Format("2006-01-02")

	tests := []struct {
		name string
		req  *types.PositionRequest
		want string
	}{
		{
			name: "invalid date format",
			req:  baseRequest("bad-date"),
			want: "invalid effectiveDate",
		},
		{
			name: "invalid profile json",
			req: func() *types.PositionRequest {
				r := baseRequest(today)
				r.Profile = stringPtr("{")
				return r
			}(),
			want: "profile must be a valid JSON object",
		},
		{
			name: "negative headcount",
			req: func() *types.PositionRequest {
				r := baseRequest(today)
				r.HeadcountCapacity = -1
				return r
			}(),
			want: ErrInvalidHeadcount.Error(),
		},
		{
			name: "headcount in use exceeds capacity",
			req: func() *types.PositionRequest {
				r := baseRequest(today)
				r.HeadcountInUse = floatPtr(3)
				return r
			}(),
			want: ErrInvalidHeadcount.Error(),
		},
	}

	for _, tc := range tests {
		_, err := svc.buildPositionEntity(uuid.New(), "POS001", tc.req, catalog, org, types.OperatedByInfo{}, true)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s: expected error containing %q, got %v", tc.name, tc.want, err)
		}
	}
}

func TestToPositionResponseIncludesAssignments(t *testing.T) {
	svc := &PositionService{}
	now := time.Now().UTC()
	endDate := now.Add(24 * time.Hour)
	entity := &types.Position{
		RecordID:           uuid.New(),
		Code:               "POS123",
		Title:              "Engineer",
		JobProfileCode:     sql.NullString{String: "JP", Valid: true},
		JobProfileName:     sql.NullString{String: "JPName", Valid: true},
		JobFamilyGroupCode: "G1",
		JobFamilyGroupName: "Group",
		JobFamilyCode:      "F1",
		JobFamilyName:      "Family",
		JobRoleCode:        "R1",
		JobRoleName:        "Role",
		JobLevelCode:       "L1",
		JobLevelName:       "Level",
		OrganizationCode:   "ORG",
		OrganizationName:   sql.NullString{String: "OrgName", Valid: true},
		PositionType:       "FULLTIME",
		Status:             "ACTIVE",
		EmploymentType:     "PERMANENT",
		HeadcountCapacity:  1,
		HeadcountInUse:     2, // forces available headcount floor to 0
		GradeLevel:         sql.NullString{String: "G7", Valid: true},
		CostCenterCode:     sql.NullString{String: "CC", Valid: true},
		ReportsToPosition:  sql.NullString{String: "PARENT", Valid: true},
		EffectiveDate:      now.Add(48 * time.Hour),
		EndDate:            sql.NullTime{Time: endDate, Valid: true},
		IsCurrent:          true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	assignments := []types.PositionAssignment{
		{
			AssignmentID:     uuid.New(),
			PositionCode:     "POS123",
			PositionRecordID: entity.RecordID,
			EmployeeID:       uuid.New(),
			EmployeeName:     "Eve",
			AssignmentType:   "PRIMARY",
			AssignmentStatus: "ACTIVE",
			FTE:              1,
			IsCurrent:        true,
			EffectiveDate:    now,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	resp := svc.toPositionResponse(entity, assignments)
	if resp.AvailableHeadcount != 0 {
		t.Fatalf("available headcount should floor at 0, got %v", resp.AvailableHeadcount)
	}
	if resp.CurrentAssignment == nil {
		t.Fatalf("expected current assignment to be set")
	}
	if resp.IsFuture != true {
		t.Fatalf("future effective date should set IsFuture=true")
	}
	if resp.OrganizationName == nil || *resp.OrganizationName != "OrgName" {
		t.Fatalf("organization name should be propagated")
	}
}

func TestMergeAttributes(t *testing.T) {
	base := map[string]interface{}{"a": 1}
	extra := map[string]interface{}{
		"b":      2,
		"empty":  nil,
		"":       "skip",
		"source": "override",
	}
	result := mergeAttributes(base, extra)
	if result["a"] != 1 || result["b"] != 2 {
		t.Fatalf("merge failed: %#v", result)
	}
	if _, ok := result["empty"]; ok {
		t.Fatalf("nil values should be skipped")
	}
	if _, ok := result[""]; ok {
		t.Fatalf("empty keys should be skipped")
	}
}

func TestSaveOutboxEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectBegin()
	tx, _ := db.Begin()

	svc := &PositionService{outboxRepo: &stubOutboxRepo{}}
	outbox := database.NewOutboxEvent()
	outbox.AggregateID = "agg"
	outbox.AggregateType = "pos"
	outbox.EventType = "position.created"
	outbox.Payload = "{}"

	// nil repo -> no-op
	if err := (&PositionService{}).saveOutboxEvent(context.Background(), tx, outbox); err != nil {
		t.Fatalf("expected nil when repo nil, got %v", err)
	}

	// nil tx should error
	if err := svc.saveOutboxEvent(context.Background(), nil, outbox); err == nil {
		t.Fatalf("expected error when tx is nil")
	}

	// happy path
	if err := svc.saveOutboxEvent(context.Background(), tx, outbox); err != nil {
		t.Fatalf("unexpected error saving outbox: %v", err)
	}
}

func TestNewEventContext(t *testing.T) {
	svc := &PositionService{}
	ctx := context.Background()
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	result := svc.newEventContext(ctx, tenant, "CreatePosition")
	if result.TenantID != tenant || result.Operation != "CreatePosition" {
		t.Fatalf("newEventContext did not propagate values: %+v", result)
	}
	if result.Source != events.DefaultSourceCommand {
		t.Fatalf("expected default source %s, got %s", events.DefaultSourceCommand, result.Source)
	}
}

type stubOutboxRepo struct {
	saved []*database.OutboxEvent
}

func (s *stubOutboxRepo) Save(ctx context.Context, tx database.Transaction, event *database.OutboxEvent) error {
	if tx == nil {
		return errors.New("tx required")
	}
	s.saved = append(s.saved, event)
	return nil
}

func (s *stubOutboxRepo) GetUnpublishedForUpdate(ctx context.Context, tx database.Transaction, limit int) ([]*database.OutboxEvent, error) {
	return nil, nil
}

func (s *stubOutboxRepo) MarkPublished(ctx context.Context, eventID string) error { return nil }

func (s *stubOutboxRepo) IncrementRetryCount(ctx context.Context, eventID string, nextAvailable time.Time) error {
	return nil
}

func stringPtr(v string) *string  { return &v }
func floatPtr(v float64) *float64 { return &v }

func baseRequest(effectiveDate string) *types.PositionRequest {
	return &types.PositionRequest{
		Title:              "Engineer",
		JobFamilyGroupCode: "G1",
		JobFamilyCode:      "F1",
		JobRoleCode:        "R1",
		JobLevelCode:       "L1",
		OrganizationCode:   "1000000",
		PositionType:       "fulltime",
		EmploymentType:     "permanent",
		HeadcountCapacity:  1,
		EffectiveDate:      effectiveDate,
		OperationReason:    "Reason",
	}
}
