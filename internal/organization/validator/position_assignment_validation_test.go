package validator

import (
	"context"
	"database/sql"
	"io"
	"testing"
	"time"

	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

func TestValidateCreatePosition_PosOrgInactive(t *testing.T) {
	tenant := uuid.New()
		orgRepo := &StubOrganizationRepository{
			GetByCodeFn: func(_ context.Context, _ uuid.UUID, code string) (*types.Organization, error) {
			status := "INACTIVE"
			return &types.Organization{Code: code, Status: status, Name: "Finance"}, nil
		},
	}
	jobCatalog := activeJobCatalogStub()
	positionRepo := &StubPositionRepository{}
	assignRepo := &StubAssignmentRepository{}

	posValidator, _ := NewPositionAssignmentValidationService(
		orgRepo,
		jobCatalog,
		positionRepo,
		assignRepo,
		testValidatorLogger(),
	)

	req := &types.PositionRequest{
		Title:              "HR Manager",
		JobFamilyGroupCode: "OPER",
		JobFamilyCode:      "OPER-HR",
		JobRoleCode:        "OPER-HR-SUP",
		JobLevelCode:       "P1",
		OrganizationCode:   "1000001",
		PositionType:       "REGULAR",
		EmploymentType:     "FULL_TIME",
		HeadcountCapacity:  1,
		EffectiveDate:      "2025-11-06",
		OperationReason:    "New headcount",
	}

	result := posValidator.ValidateCreatePosition(context.Background(), tenant, req)
	if result.Valid {
		t.Fatalf("expected validation to fail when organization inactive")
	}
	if len(result.Errors) == 0 || result.Errors[0].Code != "POS_ORG_INACTIVE" {
		t.Fatalf("expected POS_ORG_INACTIVE error, got %#v", result.Errors)
	}
}

func TestValidateFillPosition_PosHeadcountExceeded(t *testing.T) {
	tenant := uuid.New()
	orgRepo := &StubOrganizationRepository{
		GetByCodeFn: func(_ context.Context, _ uuid.UUID, code string) (*types.Organization, error) {
			return &types.Organization{Code: code, Status: "ACTIVE", Name: "IT"}, nil
		},
	}
	jobCatalog := activeJobCatalogStub()
	positionRepo := &StubPositionRepository{
		GetCurrentPositionFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.Position, error) {
			return &types.Position{
				Code:              code,
				OrganizationCode:  "1000001",
				Status:            "ACTIVE",
				HeadcountCapacity: 1,
			}, nil
		},
	}
	assignRepo := &StubAssignmentRepository{
		SumActiveFTEFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, _ string) (float64, error) {
			return 0.9, nil
		},
	}

	posValidator, _ := NewPositionAssignmentValidationService(
		orgRepo,
		jobCatalog,
		positionRepo,
		assignRepo,
		testValidatorLogger(),
	)

	fte := 0.2
	req := &types.FillPositionRequest{
		EmployeeID:      uuid.New().String(),
		EmployeeName:    "Alice",
		AssignmentType:  "PRIMARY",
		FTE:             &fte,
		EffectiveDate:   "2025-11-06",
		OperationReason: "Backfill",
	}

	result := posValidator.ValidateFillPosition(context.Background(), tenant, "P1000001", req)
	if result.Valid {
		t.Fatalf("expected headcount validation failure")
	}
	if len(result.Errors) == 0 || result.Errors[0].Code != "POS_HEADCOUNT_EXCEEDED" {
		t.Fatalf("expected POS_HEADCOUNT_EXCEEDED error, got %#v", result.Errors)
	}
}

func TestValidateCreateAssignment_AssignFTEInvalid(t *testing.T) {
	tenant := uuid.New()
	orgRepo := activeOrgRepoStub()
	jobCatalog := activeJobCatalogStub()
	positionRepo := activePositionStub()
	assignRepo := &StubAssignmentRepository{
		SumActiveFTEFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, _ string) (float64, error) {
			return 0, nil
		},
	}

	_, assignValidator := NewPositionAssignmentValidationService(
		orgRepo,
		jobCatalog,
		positionRepo,
		assignRepo,
		testValidatorLogger(),
	)

	fte := 1.5
	req := &types.CreateAssignmentRequest{
		EmployeeID:      uuid.New().String(),
		EmployeeName:    "Bob",
		AssignmentType:  "PRIMARY",
		FTE:             &fte,
		EffectiveDate:   "2025-11-06",
		OperationReason: "Backfill",
	}

	result := assignValidator.ValidateCreateAssignment(context.Background(), tenant, "P1000001", req)
	if result.Valid {
		t.Fatalf("expected ASSIGN_FTE_LIMIT failure")
	}
	if len(result.Errors) == 0 || result.Errors[0].Code != "ASSIGN_FTE_LIMIT" {
		t.Fatalf("expected ASSIGN_FTE_LIMIT, got %#v", result.Errors)
	}
}

func TestValidateCreateAssignment_CrossActiveOrganizationInactive(t *testing.T) {
	tenant := uuid.New()
	orgRepo := &StubOrganizationRepository{
		GetByCodeFn: func(_ context.Context, _ uuid.UUID, code string) (*types.Organization, error) {
			return &types.Organization{Code: code, Status: "INACTIVE", Name: "Finance"}, nil
		},
	}
	jobCatalog := activeJobCatalogStub()
	positionRepo := activePositionStub()
	assignRepo := &StubAssignmentRepository{
		SumActiveFTEFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, _ string) (float64, error) {
			return 0, nil
		},
	}

	_, assignValidator := NewPositionAssignmentValidationService(
		orgRepo,
		jobCatalog,
		positionRepo,
		assignRepo,
		testValidatorLogger(),
	)

	fte := 0.5
	req := &types.CreateAssignmentRequest{
		EmployeeID:      uuid.New().String(),
		EmployeeName:    "Cathy",
		AssignmentType:  "PRIMARY",
		FTE:             &fte,
		EffectiveDate:   time.Now().Format("2006-01-02"),
		OperationReason: "Growth",
	}

	result := assignValidator.ValidateCreateAssignment(context.Background(), tenant, "P1000001", req)
	if result.Valid {
		t.Fatalf("expected CROSS_ACTIVATION_CONFLICT failure")
	}
	if len(result.Errors) == 0 || result.Errors[0].Code != "CROSS_ACTIVATION_CONFLICT" {
		t.Fatalf("expected CROSS_ACTIVATION_CONFLICT, got %#v", result.Errors)
	}
}

func TestValidateCloseAssignment_AssignStateGuard(t *testing.T) {
	tenant := uuid.New()
	orgRepo := activeOrgRepoStub()
	jobCatalog := activeJobCatalogStub()
	positionRepo := activePositionStub()
	assignRepo := &StubAssignmentRepository{
		GetByIDFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, assignmentID uuid.UUID) (*types.PositionAssignment, error) {
			return &types.PositionAssignment{
				AssignmentID:     assignmentID,
				PositionCode:     "P1000001",
				AssignmentStatus: "PENDING",
			}, nil
		},
		SumActiveFTEFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, _ string) (float64, error) {
			return 0, nil
		},
	}

	_, assignValidator := NewPositionAssignmentValidationService(
		orgRepo,
		jobCatalog,
		positionRepo,
		assignRepo,
		testValidatorLogger(),
	)

	req := &types.CloseAssignmentRequest{
		EndDate:         "2025-11-06",
		OperationReason: "Cleanup",
	}

	result := assignValidator.ValidateCloseAssignment(context.Background(), tenant, "P1000001", uuid.New(), req)
	if result.Valid {
		t.Fatalf("expected ASSIGN_INVALID_STATE failure")
	}
	if len(result.Errors) == 0 || result.Errors[0].Code != "ASSIGN_INVALID_STATE" {
		t.Fatalf("expected ASSIGN_INVALID_STATE, got %#v", result.Errors)
	}
}

func TestValidateCreatePosition_PosJobCatalogInactive(t *testing.T) {
	tenant := uuid.New()
	orgRepo := activeOrgRepoStub()
	jobCatalog := &StubJobCatalogRepository{
		GetCurrentFamilyGroupFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.JobFamilyGroup, error) {
			return &types.JobFamilyGroup{Code: code, Status: "ACTIVE", Name: "Operations"}, nil
		},
		GetCurrentJobFamilyFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.JobFamily, error) {
			return &types.JobFamily{Code: code, Status: "ACTIVE", Name: "HR"}, nil
		},
		GetCurrentJobRoleFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.JobRole, error) {
			return &types.JobRole{Code: code, Status: "INACTIVE", Name: "Lead"}, nil
		},
		GetCurrentJobLevelFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.JobLevel, error) {
			return &types.JobLevel{Code: code, Status: "ACTIVE", Name: "P1"}, nil
		},
	}
	positionRepo := &StubPositionRepository{}
	assignRepo := &StubAssignmentRepository{}

	posValidator, _ := NewPositionAssignmentValidationService(
		orgRepo,
		jobCatalog,
		positionRepo,
		assignRepo,
		testValidatorLogger(),
	)

	req := &types.PositionRequest{
		Title:              "Product Lead",
		JobFamilyGroupCode: "OPER",
		JobFamilyCode:      "OPER-HR",
		JobRoleCode:        "OPER-HR-LEAD",
		JobLevelCode:       "P2",
		OrganizationCode:   "1000001",
		PositionType:       "REGULAR",
		EmploymentType:     "FULL_TIME",
		HeadcountCapacity:  1,
		EffectiveDate:      "2025-11-06",
		OperationReason:    "Growth",
	}

	result := posValidator.ValidateCreatePosition(context.Background(), tenant, req)
	if result.Valid {
		t.Fatalf("expected job catalog validation failure")
	}
	if len(result.Errors) == 0 || result.Errors[0].Code != "JOB_CATALOG_NOT_FOUND" {
		t.Fatalf("expected JOB_CATALOG_NOT_FOUND, got %#v", result.Errors)
	}

	// ValidateReplacePosition reuses the same rules and should surface the same violation.
	result = posValidator.ValidateReplacePosition(context.Background(), tenant, "P1000001", req)
	if result.Valid {
		t.Fatalf("expected replace position validation to fail for inactive job role")
	}
}

func TestValidateUpdateAssignment_PosHeadcountExceeded(t *testing.T) {
	tenant := uuid.New()
	orgRepo := activeOrgRepoStub()
	jobCatalog := activeJobCatalogStub()
	positionRepo := &StubPositionRepository{
		GetCurrentPositionFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.Position, error) {
			return &types.Position{
				Code:              code,
				OrganizationCode:  "1000001",
				Status:            "ACTIVE",
				HeadcountCapacity: 1,
			}, nil
		},
	}
	assignRepo := &StubAssignmentRepository{
		GetByIDFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, assignmentID uuid.UUID) (*types.PositionAssignment, error) {
			return &types.PositionAssignment{
				AssignmentID:     assignmentID,
				PositionCode:     "P1000001",
				AssignmentStatus: "ACTIVE",
				FTE:              0.8,
			}, nil
		},
		SumActiveFTEFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, _ string) (float64, error) {
			return 1.0, nil
		},
	}

	_, assignValidator := NewPositionAssignmentValidationService(
		orgRepo,
		jobCatalog,
		positionRepo,
		assignRepo,
		testValidatorLogger(),
	)

	newFTE := 0.9
	req := &types.UpdateAssignmentRequest{
		FTE:             &newFTE,
		OperationReason: "Adjust workload",
	}

	result := assignValidator.ValidateUpdateAssignment(context.Background(), tenant, "P1000001", uuid.New(), req)
	if result.Valid {
		t.Fatalf("expected POS_HEADCOUNT_EXCEEDED failure on update")
	}
	if len(result.Errors) == 0 || result.Errors[0].Code != "POS_HEADCOUNT_EXCEEDED" {
		t.Fatalf("expected POS_HEADCOUNT_EXCEEDED, got %#v", result.Errors)
	}
}

func TestValidateTransferPosition_TargetOrgMissing(t *testing.T) {
	tenant := uuid.New()
	orgRepo := &StubOrganizationRepository{
		GetByCodeFn: func(_ context.Context, _ uuid.UUID, _ string) (*types.Organization, error) {
			return nil, nil
		},
	}
	jobCatalog := activeJobCatalogStub()
	positionRepo := activePositionStub()
	assignRepo := &StubAssignmentRepository{}

	posValidator, _ := NewPositionAssignmentValidationService(
		orgRepo,
		jobCatalog,
		positionRepo,
		assignRepo,
		testValidatorLogger(),
	)

	req := &types.TransferPositionRequest{
		TargetOrganizationCode: "9999999",
		EffectiveDate:          "2025-11-06",
		OperationReason:        "Restructure",
	}

	result := posValidator.ValidateTransferPosition(context.Background(), tenant, "P1000001", req)
	if result.Valid {
		t.Fatalf("expected POS_ORG_INACTIVE failure for missing target org")
	}
	if len(result.Errors) == 0 || result.Errors[0].Code != "POS_ORG_INACTIVE" {
		t.Fatalf("expected POS_ORG_INACTIVE, got %#v", result.Errors)
	}
}

func activeJobCatalogStub() *StubJobCatalogRepository {
	return &StubJobCatalogRepository{
		GetCurrentFamilyGroupFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.JobFamilyGroup, error) {
			return &types.JobFamilyGroup{Code: code, Status: "ACTIVE", Name: "Operations"}, nil
		},
		GetCurrentJobFamilyFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.JobFamily, error) {
			return &types.JobFamily{Code: code, Status: "ACTIVE", Name: "HR"}, nil
		},
		GetCurrentJobRoleFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.JobRole, error) {
			return &types.JobRole{Code: code, Status: "ACTIVE", Name: "Manager"}, nil
		},
		GetCurrentJobLevelFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.JobLevel, error) {
			return &types.JobLevel{Code: code, Status: "ACTIVE", Name: "P1"}, nil
		},
	}
}

func TestStubValidationServicePassThrough(t *testing.T) {
	stub := NewStubValidationService()
	tenant := uuid.New()
	if !stub.ValidateCreatePosition(context.Background(), tenant, &types.PositionRequest{}).Valid {
		t.Fatalf("expected stub create position to be valid")
	}
	if !stub.ValidateReplacePosition(context.Background(), tenant, "P1000001", &types.PositionRequest{}).Valid {
		t.Fatalf("expected stub replace position to be valid")
	}
	if !stub.ValidateCreateVersion(context.Background(), tenant, "P1000001", &types.PositionVersionRequest{}).Valid {
		t.Fatalf("expected stub create version to be valid")
	}
	if !stub.ValidateFillPosition(context.Background(), tenant, "P1000001", &types.FillPositionRequest{}).Valid {
		t.Fatalf("expected stub fill position to be valid")
	}
	if !stub.ValidateVacatePosition(context.Background(), tenant, "P1000001", &types.VacatePositionRequest{}).Valid {
		t.Fatalf("expected stub vacate position to be valid")
	}
	if !stub.ValidateTransferPosition(context.Background(), tenant, "P1000001", &types.TransferPositionRequest{}).Valid {
		t.Fatalf("expected stub transfer position to be valid")
	}
	if !stub.ValidateApplyEvent(context.Background(), tenant, "P1000001", &types.PositionEventRequest{}).Valid {
		t.Fatalf("expected stub apply event to be valid")
	}
	if !stub.ValidateCreateAssignment(context.Background(), tenant, "P1000001", &types.CreateAssignmentRequest{}).Valid {
		t.Fatalf("expected stub create assignment to be valid")
	}
	if !stub.ValidateUpdateAssignment(context.Background(), tenant, "P1000001", uuid.New(), &types.UpdateAssignmentRequest{OperationReason: "ok"}).Valid {
		t.Fatalf("expected stub update assignment to be valid")
	}
	if !stub.ValidateCloseAssignment(context.Background(), tenant, "P1000001", uuid.New(), &types.CloseAssignmentRequest{EndDate: time.Now().Format("2006-01-02"), OperationReason: "ok"}).Valid {
		t.Fatalf("expected stub close assignment to be valid")
	}
}

func TestValidationFailedErrorWrapper(t *testing.T) {
	res := NewValidationResult()
	res.Valid = false
	res.Errors = append(res.Errors, ValidationError{Code: "TEST", Message: "failed"})

	err := NewValidationFailedError("CreateAssignment", res)
	vf, ok := err.(*ValidationFailedError)
	if !ok {
		t.Fatalf("expected ValidationFailedError type")
	}
	if vf.Operation() != "CreateAssignment" {
		t.Fatalf("unexpected operation: %s", vf.Operation())
	}
	if vf.Result() != res {
		t.Fatalf("expected result pointer to match")
	}
	if vf.Error() == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestValidateVacateAndApplyEventNoOp(t *testing.T) {
	posValidator, assignValidator := buildDefaultValidationService()
	tenant := uuid.New()

	if !posValidator.ValidateVacatePosition(context.Background(), tenant, "P1000001", &types.VacatePositionRequest{}).Valid {
		t.Fatalf("expected vacate validation to be a no-op")
	}
	if !posValidator.ValidateApplyEvent(context.Background(), tenant, "P1000001", &types.PositionEventRequest{}).Valid {
		t.Fatalf("expected apply event validation to be a no-op")
	}
	if !assignValidator.ValidateCloseAssignment(context.Background(), tenant, "P1000001", uuid.New(), &types.CloseAssignmentRequest{EndDate: time.Now().Format("2006-01-02"), OperationReason: "Close"}).Valid {
		t.Fatalf("expected assignment close validation to pass for stub context")
	}
}

func TestValidateCreateVersion_PosOrgRule(t *testing.T) {
	orgRepo := &StubOrganizationRepository{
		GetByCodeFn: func(_ context.Context, _ uuid.UUID, _ string) (*types.Organization, error) {
			return nil, nil
		},
	}
	jobCatalog := activeJobCatalogStub()
	positionRepo := activePositionStub()
	assignRepo := &StubAssignmentRepository{}
	validator, _ := NewPositionAssignmentValidationService(orgRepo, jobCatalog, positionRepo, assignRepo, testValidatorLogger())
	tenant := uuid.New()

	req := &types.PositionVersionRequest{
		EffectiveDate:   time.Now().Format("2006-01-02"),
		OperationReason: "test",
	}

	result := validator.ValidateCreateVersion(context.Background(), tenant, "P1000001", req)
	if !result.Valid {
		t.Fatalf("expected create version validation to succeed, errors: %#v", result.Errors)
	}
}

func buildDefaultValidationService() (PositionValidationService, AssignmentValidationService) {
	return NewPositionAssignmentValidationService(
		activeOrgRepoStub(),
		activeJobCatalogStub(),
		activePositionStub(),
		&StubAssignmentRepository{
			SumActiveFTEFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, _ string) (float64, error) {
				return 0, nil
			},
		},
		testValidatorLogger(),
	)
}

func pointerString(value string) *string {
	return &value
}

func activeOrgRepoStub() *StubOrganizationRepository {
	return &StubOrganizationRepository{
		GetByCodeFn: func(_ context.Context, _ uuid.UUID, code string) (*types.Organization, error) {
			return &types.Organization{Code: code, Status: "ACTIVE", Name: "HQ"}, nil
		},
	}
}

func activePositionStub() *StubPositionRepository {
	return &StubPositionRepository{
		GetCurrentPositionFn: func(_ context.Context, _ *sql.Tx, _ uuid.UUID, code string) (*types.Position, error) {
			return &types.Position{
				Code:              code,
				OrganizationCode:  "1000001",
				Status:            "ACTIVE",
				HeadcountCapacity: 1,
			}, nil
		},
	}
}

func testValidatorLogger() pkglogger.Logger {
	return pkglogger.NewLogger(
		pkglogger.WithWriter(io.Discard),
		pkglogger.WithLevel(pkglogger.LevelError),
	)
}
