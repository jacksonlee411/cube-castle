package validator

import (
	"context"

	"cube-castle/internal/types"
	"github.com/google/uuid"
)

// PositionValidationService 定义职位相关命令可复用的验证入口。
type PositionValidationService interface {
	ValidateCreatePosition(ctx context.Context, tenantID uuid.UUID, req *types.PositionRequest) *ValidationResult
	ValidateReplacePosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionRequest) *ValidationResult
	ValidateCreateVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionVersionRequest) *ValidationResult
	ValidateFillPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.FillPositionRequest) *ValidationResult
	ValidateVacatePosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.VacatePositionRequest) *ValidationResult
	ValidateTransferPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.TransferPositionRequest) *ValidationResult
	ValidateApplyEvent(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionEventRequest) *ValidationResult
}

// AssignmentValidationService 定义任职命令的统一验证入口。
type AssignmentValidationService interface {
	ValidateCreateAssignment(ctx context.Context, tenantID uuid.UUID, positionCode string, req *types.CreateAssignmentRequest) *ValidationResult
	ValidateUpdateAssignment(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID uuid.UUID, req *types.UpdateAssignmentRequest) *ValidationResult
	ValidateCloseAssignment(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID uuid.UUID, req *types.CloseAssignmentRequest) *ValidationResult
}

// JobCatalogValidationService 定义 Job Catalog 版本类命令的验证入口。
type JobCatalogValidationService interface {
	ValidateCreateFamilyGroupVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest) *ValidationResult
	ValidateCreateJobFamilyVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *ValidationResult
	ValidateCreateJobRoleVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *ValidationResult
	ValidateCreateJobLevelVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *ValidationResult
}
