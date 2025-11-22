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
func (s *stubValidationService) ValidateCreatePosition(_ context.Context, _ uuid.UUID, _ *types.PositionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateReplacePosition(_ context.Context, _ uuid.UUID, _ string, _ *types.PositionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateCreateVersion(_ context.Context, _ uuid.UUID, _ string, _ *types.PositionVersionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateFillPosition(_ context.Context, _ uuid.UUID, _ string, _ *types.FillPositionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateVacatePosition(_ context.Context, _ uuid.UUID, _ string, _ *types.VacatePositionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateTransferPosition(_ context.Context, _ uuid.UUID, _ string, _ *types.TransferPositionRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateApplyEvent(_ context.Context, _ uuid.UUID, _ string, _ *types.PositionEventRequest) *ValidationResult {
	return s.result()
}

// Assignment 验证占位实现
func (s *stubValidationService) ValidateCreateAssignment(_ context.Context, _ uuid.UUID, _ string, _ *types.CreateAssignmentRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateUpdateAssignment(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID, _ *types.UpdateAssignmentRequest) *ValidationResult {
	return s.result()
}

func (s *stubValidationService) ValidateCloseAssignment(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID, _ *types.CloseAssignmentRequest) *ValidationResult {
	return s.result()
}

// Enforce interface compliance at compile time.
var (
	_ PositionValidationService   = (*stubValidationService)(nil)
	_ AssignmentValidationService = (*stubValidationService)(nil)
)
