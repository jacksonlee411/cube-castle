package validator

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cube-castle/internal/organization/repository"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

func TestJobCatalogTemporalAllowsInitialVersion(t *testing.T) {
	tenant := uuid.New()
	stub := &StubJobCatalogTimelineRepository{
		ListFamilyGroupTimelineFn: func(_ context.Context, _ uuid.UUID, _ string) ([]repository.JobCatalogTimelineEntry, error) {
			return []repository.JobCatalogTimelineEntry{}, nil
		},
	}

	svc := NewJobCatalogValidationService(stub, pkglogger.NewNoopLogger())
	req := &types.JobCatalogVersionRequest{
		Name:          "Catalog v1",
		Status:        "ACTIVE",
		EffectiveDate: "2025-01-01",
	}

	result := svc.ValidateCreateFamilyGroupVersion(context.Background(), tenant, "JFG-100", req)
	if result == nil {
		t.Fatalf("expected validation result, got nil")
	}
	if !result.Valid {
		t.Fatalf("expected initial version to be valid, errors: %+v", result.Errors)
	}
}

func TestJobCatalogTemporalRejectsOverlappingVersion(t *testing.T) {
	tenant := uuid.New()
	latestRecord := uuid.New()
	stub := &StubJobCatalogTimelineRepository{
		ListJobFamilyTimelineFn: func(_ context.Context, _ uuid.UUID, _ string) ([]repository.JobCatalogTimelineEntry, error) {
			return []repository.JobCatalogTimelineEntry{
				{
					RecordID:      latestRecord,
					EffectiveDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:        "ACTIVE",
				},
			}, nil
		},
	}

	svc := NewJobCatalogValidationService(stub, pkglogger.NewNoopLogger())
	req := &types.JobCatalogVersionRequest{
		Name:          "Catalog v2",
		Status:        "ACTIVE",
		EffectiveDate: "2025-01-01",
	}

	result := svc.ValidateCreateJobFamilyVersion(context.Background(), tenant, "JF-100", req, latestRecord)
	if result == nil {
		t.Fatalf("expected validation result, got nil")
	}
	if result.Valid {
		t.Fatalf("expected validation failure due to temporal conflict")
	}
	if len(result.Errors) == 0 {
		t.Fatalf("expected at least one validation error")
	}
	if result.Errors[0].Code != errorCodeTemporalConflict {
		t.Fatalf("expected error code %s, got %s", errorCodeTemporalConflict, result.Errors[0].Code)
	}
}

func TestJobCatalogSequenceRejectsMismatchedParent(t *testing.T) {
	tenant := uuid.New()
	expectedParent := uuid.New()
	stub := &StubJobCatalogTimelineRepository{
		ListJobRoleTimelineFn: func(_ context.Context, _ uuid.UUID, _ string) ([]repository.JobCatalogTimelineEntry, error) {
			return []repository.JobCatalogTimelineEntry{
				{
					RecordID:      expectedParent,
					EffectiveDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:        "ACTIVE",
				},
			}, nil
		},
	}

	svc := NewJobCatalogValidationService(stub, pkglogger.NewNoopLogger())
	req := &types.JobCatalogVersionRequest{
		Name:          "Role v2",
		Status:        "ACTIVE",
		EffectiveDate: "2025-02-01",
	}

	mismatchedParent := uuid.New()
	result := svc.ValidateCreateJobRoleVersion(context.Background(), tenant, "JR-100", req, mismatchedParent)
	if result == nil {
		t.Fatalf("expected validation result, got nil")
	}
	if result.Valid {
		t.Fatalf("expected validation failure due to parent mismatch")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == errorCodeSequenceMismatch {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected error code %s, got %+v", errorCodeSequenceMismatch, result.Errors)
	}
}

func TestJobCatalogSequenceAcceptsMatchingParent(t *testing.T) {
	tenant := uuid.New()
	parent := uuid.New()
	stub := &StubJobCatalogTimelineRepository{
		ListJobLevelTimelineFn: func(_ context.Context, _ uuid.UUID, _ string) ([]repository.JobCatalogTimelineEntry, error) {
			return []repository.JobCatalogTimelineEntry{
				{
					RecordID:      parent,
					EffectiveDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:        "ACTIVE",
				},
			}, nil
		},
	}

	svc := NewJobCatalogValidationService(stub, pkglogger.NewNoopLogger())
	req := &types.JobCatalogVersionRequest{
		Name:          "Level v2",
		Status:        "ACTIVE",
		EffectiveDate: "2025-02-01",
	}

	result := svc.ValidateCreateJobLevelVersion(context.Background(), tenant, "JL-100", req, parent)
	if result == nil {
		t.Fatalf("expected validation result, got nil")
	}
	if !result.Valid {
		t.Fatalf("expected validation to pass, errors: %+v", result.Errors)
	}
}

func TestJobCatalogRequiresValidEffectiveDate(t *testing.T) {
	tenant := uuid.New()
	stub := &StubJobCatalogTimelineRepository{}
	svc := NewJobCatalogValidationService(stub, pkglogger.NewNoopLogger())

	req := &types.JobCatalogVersionRequest{
		Name:          "Invalid date",
		Status:        "ACTIVE",
		EffectiveDate: "2025/01/01",
	}

	result := svc.ValidateCreateFamilyGroupVersion(context.Background(), tenant, "JFG-100", req)
	if result == nil || result.Valid {
		t.Fatalf("expected invalid effective date to fail")
	}
	if len(result.Errors) == 0 || result.Errors[0].Code != "INVALID_EFFECTIVE_DATE" {
		t.Fatalf("expected INVALID_EFFECTIVE_DATE, got %+v", result.Errors)
	}
}

func TestJobCatalogTimelineFailure(t *testing.T) {
	tenant := uuid.New()
	stub := &StubJobCatalogTimelineRepository{
		ListJobFamilyTimelineFn: func(_ context.Context, _ uuid.UUID, _ string) ([]repository.JobCatalogTimelineEntry, error) {
			return nil, fmt.Errorf("boom")
		},
	}

	svc := NewJobCatalogValidationService(stub, pkglogger.NewNoopLogger())
	req := &types.JobCatalogVersionRequest{
		Name:          "Catalog v2",
		Status:        "ACTIVE",
		EffectiveDate: "2025-02-01",
		ParentRecordID: func() *string {
			id := uuid.New().String()
			return &id
		}(),
	}

	result := svc.ValidateCreateJobFamilyVersion(context.Background(), tenant, "JF-100", req, uuid.New())
	if result == nil || result.Valid {
		t.Fatalf("expected timeline failure to invalidate request")
	}
	if len(result.Errors) == 0 || result.Errors[0].Code != errorCodeTimelineUnavailable {
		t.Fatalf("expected %s error, got %+v", errorCodeTimelineUnavailable, result.Errors)
	}
}

func TestJobCatalogSequenceMissingParent(t *testing.T) {
	tenant := uuid.New()
	expected := uuid.New()
	stub := &StubJobCatalogTimelineRepository{
		ListJobRoleTimelineFn: func(_ context.Context, _ uuid.UUID, _ string) ([]repository.JobCatalogTimelineEntry, error) {
			return []repository.JobCatalogTimelineEntry{
				{
					RecordID:      expected,
					EffectiveDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					Status:        "ACTIVE",
				},
			}, nil
		},
	}

	svc := NewJobCatalogValidationService(stub, pkglogger.NewNoopLogger())
	impl := svc.(*jobCatalogValidationService)

	cfg := jobCatalogVersionConfig{
		operation:     "CreateJobRoleVersion",
		entity:        "JOB_ROLE",
		tenantID:      tenant,
		code:          "JR-100",
		request:       &types.JobCatalogVersionRequest{Name: "Role v2", Status: "ACTIVE", EffectiveDate: "2025-02-01"},
		loader:        stub.ListJobRoleTimeline,
		requireParent: true,
	}

	result := impl.validateVersion(context.Background(), cfg)
	if result.Valid {
		t.Fatalf("expected missing parent to invalidate request")
	}

	found := false
	for _, err := range result.Errors {
		if err.Code == errorCodeSequenceMissingParent {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected %s error, got %+v", errorCodeSequenceMissingParent, result.Errors)
	}
}

func TestJobCatalogSequenceMissingTimeline(t *testing.T) {
	tenant := uuid.New()
	stub := &StubJobCatalogTimelineRepository{
		ListJobLevelTimelineFn: func(_ context.Context, _ uuid.UUID, _ string) ([]repository.JobCatalogTimelineEntry, error) {
			return []repository.JobCatalogTimelineEntry{}, nil
		},
	}

	svc := NewJobCatalogValidationService(stub, pkglogger.NewNoopLogger())
	req := &types.JobCatalogVersionRequest{
		Name:          "Level v1",
		Status:        "ACTIVE",
		EffectiveDate: "2025-02-01",
	}

	result := svc.ValidateCreateJobLevelVersion(context.Background(), tenant, "JL-100", req, uuid.New())
	if result == nil || result.Valid {
		t.Fatalf("expected missing base timeline to fail sequence rule")
	}

	for _, err := range result.Errors {
		if err.Code == errorCodeSequenceMissingBase {
			return
		}
	}
	t.Fatalf("expected %s error, got %+v", errorCodeSequenceMissingBase, result.Errors)
}

func TestJobCatalogParseEffectiveDateRequiresValue(t *testing.T) {
	svc := &jobCatalogValidationService{}
	if _, err := svc.parseEffectiveDate("   "); err == nil {
		t.Fatalf("expected empty effective date to return error")
	}
}
