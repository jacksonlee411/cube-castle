package service

import (
	"context"
	"errors"
	"testing"

	validator "cube-castle/internal/organization/validator"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type stubJobCatalogValidator struct {
	onFamilyGroup func(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest) *validator.ValidationResult
	onFamily      func(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *validator.ValidationResult
	onRole        func(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *validator.ValidationResult
	onLevel       func(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *validator.ValidationResult
}

func (s *stubJobCatalogValidator) ValidateCreateFamilyGroupVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest) *validator.ValidationResult {
	if s.onFamilyGroup != nil {
		return s.onFamilyGroup(ctx, tenantID, code, req)
	}
	return validator.NewValidationResult()
}

func (s *stubJobCatalogValidator) ValidateCreateJobFamilyVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *validator.ValidationResult {
	if s.onFamily != nil {
		return s.onFamily(ctx, tenantID, code, req, parentRecordID)
	}
	return validator.NewValidationResult()
}

func (s *stubJobCatalogValidator) ValidateCreateJobRoleVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *validator.ValidationResult {
	if s.onRole != nil {
		return s.onRole(ctx, tenantID, code, req, parentRecordID)
	}
	return validator.NewValidationResult()
}

func (s *stubJobCatalogValidator) ValidateCreateJobLevelVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *validator.ValidationResult {
	if s.onLevel != nil {
		return s.onLevel(ctx, tenantID, code, req, parentRecordID)
	}
	return validator.NewValidationResult()
}

func TestTranslateJobCatalogErrorUsesValidatorResult(t *testing.T) {
	t.Helper()
	tenantID := uuid.New()
	stubResult := validator.NewValidationResult()
	stubResult.Valid = false
	stubResult.Errors = append(stubResult.Errors, validator.ValidationError{
		Code:     "JOB_CATALOG_TEMPORAL_CONFLICT",
		Message:  "Job catalog version already exists",
		Severity: string(validator.SeverityHigh),
		Context: map[string]interface{}{
			"ruleId": "JC-TEMPORAL",
		},
	})
	stubValidator := &stubJobCatalogValidator{
		onFamilyGroup: func(_ context.Context, _ uuid.UUID, _ string, _ *types.JobCatalogVersionRequest) *validator.ValidationResult {
			return stubResult
		},
	}
	svc := &JobCatalogService{
		validator: stubValidator,
		logger:    pkglogger.NewNoopLogger(),
	}

	duplicateErr := &pq.Error{Code: "23505", Message: "duplicate key"}
	req := &types.JobCatalogVersionRequest{EffectiveDate: "2025-01-01"}
	translated := svc.translateJobCatalogError(context.Background(), tenantID, "JC001", "CreateJobFamilyGroupVersion", req, duplicateErr)

	var validationErr *validator.ValidationFailedError
	if !errors.As(translated, &validationErr) {
		t.Fatalf("expected validation error, got %v", translated)
	}
	if validationErr.Result() != stubResult {
		t.Fatalf("expected validator result to be reused")
	}
	if len(validationErr.Result().Errors) == 0 {
		t.Fatalf("expected validation errors to be present")
	}
	if validationErr.Result().Errors[0].Context["ruleId"] != "JC-TEMPORAL" {
		t.Fatalf("expected ruleId JC-TEMPORAL, got %#v", validationErr.Result().Errors[0].Context)
	}
}

func TestTranslateJobCatalogErrorFallbackWhenValidatorMissing(t *testing.T) {
	svc := &JobCatalogService{validator: nil, logger: pkglogger.NewNoopLogger()}
	req := &types.JobCatalogVersionRequest{EffectiveDate: "2025-01-01"}
	duplicateErr := &pq.Error{Code: "23505", Message: "duplicate key"}
	translated := svc.translateJobCatalogError(context.Background(), uuid.New(), "JC001", "CreateJobFamilyGroupVersion", req, duplicateErr)

	var validationErr *validator.ValidationFailedError
	if !errors.As(translated, &validationErr) {
		t.Fatalf("expected validation error, got %v", translated)
	}
	result := validationErr.Result()
	if result == nil {
		t.Fatalf("expected validation result to be present")
	}
	if result.Valid {
		t.Fatalf("expected result to be invalid")
	}
	if len(result.Errors) != 1 {
		t.Fatalf("expected single validation error, got %d", len(result.Errors))
	}
	if result.Errors[0].Code != "JOB_CATALOG_TEMPORAL_CONFLICT" {
		t.Fatalf("expected error code JOB_CATALOG_TEMPORAL_CONFLICT, got %s", result.Errors[0].Code)
	}
	if ruleID := result.Errors[0].Context["ruleId"]; ruleID != "JC-TEMPORAL" {
		t.Fatalf("expected ruleId JC-TEMPORAL, got %#v", ruleID)
	}
}

func TestTranslateJobCatalogErrorHandlesInvalidEffectiveDate(t *testing.T) {
	svc := &JobCatalogService{validator: nil, logger: pkglogger.NewNoopLogger()}
	req := &types.JobCatalogVersionRequest{EffectiveDate: "2025-02-01"}
	translated := svc.translateJobCatalogError(context.Background(), uuid.New(), "JC001", "CreateJobFamilyGroupVersion", req, errors.New("invalid effective date: parse error"))

	var validationErr *validator.ValidationFailedError
	if !errors.As(translated, &validationErr) {
		t.Fatalf("expected validation error, got %v", translated)
	}
	result := validationErr.Result()
	if len(result.Errors) == 0 {
		t.Fatalf("expected validation error entries")
	}
	first := result.Errors[0]
	if first.Code != "INVALID_EFFECTIVE_DATE" {
		t.Fatalf("expected error code INVALID_EFFECTIVE_DATE, got %s", first.Code)
	}
	if first.Field != "effectiveDate" {
		t.Fatalf("expected field effectiveDate, got %s", first.Field)
	}
	if ruleID := first.Context["ruleId"]; ruleID != "JC-TEMPORAL" {
		t.Fatalf("expected ruleId JC-TEMPORAL, got %#v", ruleID)
	}
}

func TestTranslateJobCatalogErrorHandlesParentGroupMissing(t *testing.T) {
	svc := &JobCatalogService{validator: nil, logger: pkglogger.NewNoopLogger()}
	err := svc.translateJobCatalogError(context.Background(), uuid.New(), "JC001", "CreateJobFamilyVersion", nil, errors.New("parent job family group not found"))
	if !errors.Is(err, ErrJobCatalogParentMissing) {
		t.Fatalf("expected ErrJobCatalogParentMissing, got %v", err)
	}
}

func TestTranslateJobCatalogErrorHandlesParentMismatch(t *testing.T) {
	svc := &JobCatalogService{validator: nil, logger: pkglogger.NewNoopLogger()}
	parent := uuid.New().String()
	req := &types.JobCatalogVersionRequest{ParentRecordID: &parent}
	err := svc.translateJobCatalogError(context.Background(), uuid.New(), "JC001", "CreateJobFamilyVersion", req, errors.New("job family parent record mismatch"))

	var validationErr *validator.ValidationFailedError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}

	result := validationErr.Result()
	if result == nil || len(result.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
	first := result.Errors[0]
	if first.Code != "JOB_CATALOG_SEQUENCE_MISMATCH" {
		t.Fatalf("expected error code JOB_CATALOG_SEQUENCE_MISMATCH, got %s", first.Code)
	}
	if first.Context["ruleId"] != "JC-SEQUENCE" {
		t.Fatalf("expected ruleId JC-SEQUENCE, got %#v", first.Context["ruleId"])
	}
}
