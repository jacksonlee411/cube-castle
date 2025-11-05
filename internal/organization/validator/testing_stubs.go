package validator

import (
	"context"
	"database/sql"
	"time"

	"cube-castle/internal/organization/repository"
	"cube-castle/internal/types"
	"github.com/google/uuid"
)

// 以下轻量级 stub 供 219C2C 单元测试复用，避免在测试中重复定义跨域依赖。

type StubHierarchyRepository struct {
	GetOrganizationFn       func(ctx context.Context, code string, tenantID uuid.UUID) (*types.Organization, error)
	GetOrganizationDepthFn  func(ctx context.Context, code string, tenantID uuid.UUID) (int, error)
	GetAncestorChainFn      func(ctx context.Context, code string, tenantID uuid.UUID) ([]repository.OrganizationNode, error)
	GetDirectChildrenFn     func(ctx context.Context, code string, tenantID uuid.UUID) ([]repository.OrganizationNode, error)
	GetOrganizationAtDateFn func(ctx context.Context, code string, tenantID uuid.UUID, ts time.Time) (*repository.OrganizationNode, error)
}

func (s *StubHierarchyRepository) GetOrganization(ctx context.Context, code string, tenantID uuid.UUID) (*types.Organization, error) {
	if s.GetOrganizationFn != nil {
		return s.GetOrganizationFn(ctx, code, tenantID)
	}
	return nil, nil
}

func (s *StubHierarchyRepository) GetOrganizationDepth(ctx context.Context, code string, tenantID uuid.UUID) (int, error) {
	if s.GetOrganizationDepthFn != nil {
		return s.GetOrganizationDepthFn(ctx, code, tenantID)
	}
	return 0, nil
}

func (s *StubHierarchyRepository) GetAncestorChain(ctx context.Context, code string, tenantID uuid.UUID) ([]repository.OrganizationNode, error) {
	if s.GetAncestorChainFn != nil {
		return s.GetAncestorChainFn(ctx, code, tenantID)
	}
	return nil, nil
}

func (s *StubHierarchyRepository) GetDirectChildren(ctx context.Context, code string, tenantID uuid.UUID) ([]repository.OrganizationNode, error) {
	if s.GetDirectChildrenFn != nil {
		return s.GetDirectChildrenFn(ctx, code, tenantID)
	}
	return nil, nil
}

func (s *StubHierarchyRepository) GetOrganizationAtDate(ctx context.Context, code string, tenantID uuid.UUID, ts time.Time) (*repository.OrganizationNode, error) {
	if s.GetOrganizationAtDateFn != nil {
		return s.GetOrganizationAtDateFn(ctx, code, tenantID, ts)
	}
	return nil, nil
}

type StubJobCatalogRepository struct {
	GetCurrentFamilyGroupFn func(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobFamilyGroup, error)
	GetCurrentJobFamilyFn   func(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobFamily, error)
	GetCurrentJobRoleFn     func(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobRole, error)
	GetCurrentJobLevelFn    func(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobLevel, error)
}

func (s *StubJobCatalogRepository) GetCurrentFamilyGroup(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobFamilyGroup, error) {
	if s.GetCurrentFamilyGroupFn != nil {
		return s.GetCurrentFamilyGroupFn(ctx, tx, tenantID, code)
	}
	return nil, nil
}

func (s *StubJobCatalogRepository) GetCurrentJobFamily(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobFamily, error) {
	if s.GetCurrentJobFamilyFn != nil {
		return s.GetCurrentJobFamilyFn(ctx, tx, tenantID, code)
	}
	return nil, nil
}

func (s *StubJobCatalogRepository) GetCurrentJobRole(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobRole, error) {
	if s.GetCurrentJobRoleFn != nil {
		return s.GetCurrentJobRoleFn(ctx, tx, tenantID, code)
	}
	return nil, nil
}

func (s *StubJobCatalogRepository) GetCurrentJobLevel(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobLevel, error) {
	if s.GetCurrentJobLevelFn != nil {
		return s.GetCurrentJobLevelFn(ctx, tx, tenantID, code)
	}
	return nil, nil
}

type StubOrganizationRepository struct {
	GetByCodeFn func(ctx context.Context, tenantID uuid.UUID, code string) (*types.Organization, error)
}

func (s *StubOrganizationRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*types.Organization, error) {
	if s.GetByCodeFn != nil {
		return s.GetByCodeFn(ctx, tenantID, code)
	}
	return nil, nil
}

type StubAssignmentRepository struct {
	GetByIDFn      func(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, assignmentID uuid.UUID) (*types.PositionAssignment, error)
	SumActiveFTEFn func(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, positionCode string) (float64, error)
}

func (s *StubAssignmentRepository) GetByID(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, assignmentID uuid.UUID) (*types.PositionAssignment, error) {
	if s.GetByIDFn != nil {
		return s.GetByIDFn(ctx, tx, tenantID, assignmentID)
	}
	return nil, nil
}

func (s *StubAssignmentRepository) SumActiveFTE(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, positionCode string) (float64, error) {
	if s.SumActiveFTEFn != nil {
		return s.SumActiveFTEFn(ctx, tx, tenantID, positionCode)
	}
	return 0, nil
}

type StubPositionRepository struct {
	GetCurrentPositionFn func(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.Position, error)
}

func (s *StubPositionRepository) GetCurrentPosition(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.Position, error) {
	if s.GetCurrentPositionFn != nil {
		return s.GetCurrentPositionFn(ctx, tx, tenantID, code)
	}
	return nil, nil
}
