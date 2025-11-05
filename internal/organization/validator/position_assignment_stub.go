package validator

import (
	"context"

	"cube-castle/internal/types"
	"github.com/google/uuid"
)

// stubValidationService 为职位/任职命令提供默认通过的占位实现，确保命令层可提前注入校验链。
type stubValidationService struct{}

// NewStubValidationService 返回一个所有验证直接通过的实现，供 219C2C 前置阶段使用。
func NewStubValidationService() *stubValidationService {
	return &stubValidationService{}
}

func (s *stubValidationService) result() *ValidationResult {
	result := NewValidationResult()
	result.Context["executedRules"] = []string{}
	return result
}

// Position 验证占位实现
func (s *stubValidationService) ValidateCreatePosition(ctx context.Context, tenantID uuid.UUID, req *types.PositionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateReplacePosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateCreateVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionVersionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateFillPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.FillPositionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateVacatePosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.VacatePositionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateTransferPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.TransferPositionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateApplyEvent(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionEventRequest) *ValidationResult {
	return s.result()
}

// Assignment 验证占位实现
func (s *stubValidationService) ValidateCreateAssignment(ctx context.Context, tenantID uuid.UUID, positionCode string, req *types.CreateAssignmentRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateUpdateAssignment(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID uuid.UUID, req *types.UpdateAssignmentRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateCloseAssignment(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID uuid.UUID, req *types.CloseAssignmentRequest) *ValidationResult {
	return s.result()
}

// Enforce interface compliance at compile time.
var (
	_ PositionValidationService   = (*stubValidationService)(nil)
	_ AssignmentValidationService = (*stubValidationService)(nil)
)
