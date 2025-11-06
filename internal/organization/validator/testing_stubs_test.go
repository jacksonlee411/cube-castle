package validator

import (
	"context"
	"testing"
	"time"

	"database/sql"

	"cube-castle/internal/organization/repository"
	"cube-castle/internal/types"
	"github.com/google/uuid"
)

func TestStubHierarchyRepositoryDefaults(t *testing.T) {
	t.Helper()

	repo := &StubHierarchyRepository{}
	ctx := context.Background()
	tenant := uuid.New()

	if org, err := repo.GetOrganization(ctx, "1000001", tenant); err != nil || org != nil {
		t.Fatalf("expected nil organization, got %v, err %v", org, err)
	}

	if depth, err := repo.GetOrganizationDepth(ctx, "1000001", tenant); err != nil || depth != 0 {
		t.Fatalf("expected zero depth, got %d, err %v", depth, err)
	}

	if ancestors, err := repo.GetAncestorChain(ctx, "1000001", tenant); err != nil || ancestors != nil {
		t.Fatalf("expected nil ancestor chain, got %v, err %v", ancestors, err)
	}

	if children, err := repo.GetDirectChildren(ctx, "1000001", tenant); err != nil || children != nil {
		t.Fatalf("expected nil children, got %v, err %v", children, err)
	}

	if atDate, err := repo.GetOrganizationAtDate(ctx, "1000001", tenant, time.Now()); err != nil || atDate != nil {
		t.Fatalf("expected nil organization at date, got %v, err %v", atDate, err)
	}
}

func TestStubHierarchyRepositoryDelegates(t *testing.T) {
	ctx := context.Background()
	tenant := uuid.New()
	wantOrg := &types.Organization{Code: "1000001"}

	repo := &StubHierarchyRepository{
		GetOrganizationFn: func(context.Context, string, uuid.UUID) (*types.Organization, error) {
			return wantOrg, nil
		},
		GetOrganizationDepthFn: func(context.Context, string, uuid.UUID) (int, error) {
			return 3, nil
		},
		GetAncestorChainFn: func(context.Context, string, uuid.UUID) ([]repository.OrganizationNode, error) {
			return []repository.OrganizationNode{{Code: "1000000"}}, nil
		},
		GetDirectChildrenFn: func(context.Context, string, uuid.UUID) ([]repository.OrganizationNode, error) {
			return []repository.OrganizationNode{{Code: "1000002"}}, nil
		},
		GetOrganizationAtDateFn: func(context.Context, string, uuid.UUID, time.Time) (*repository.OrganizationNode, error) {
			node := repository.OrganizationNode{Code: "1000001"}
			return &node, nil
		},
	}

	if org, err := repo.GetOrganization(ctx, "1000001", tenant); err != nil || org != wantOrg {
		t.Fatalf("expected stub organization, got %v, err %v", org, err)
	}
	if depth, err := repo.GetOrganizationDepth(ctx, "1000001", tenant); err != nil || depth != 3 {
		t.Fatalf("expected depth=3, got %d, err %v", depth, err)
	}
	if ancestors, err := repo.GetAncestorChain(ctx, "1000001", tenant); err != nil || len(ancestors) != 1 {
		t.Fatalf("expected ancestor chain length 1, got %v, err %v", ancestors, err)
	}
	if children, err := repo.GetDirectChildren(ctx, "1000001", tenant); err != nil || len(children) != 1 {
		t.Fatalf("expected child list length 1, got %v, err %v", children, err)
	}
	if atDate, err := repo.GetOrganizationAtDate(ctx, "1000001", tenant, time.Now()); err != nil || atDate == nil {
		t.Fatalf("expected organization node at date, got %v, err %v", atDate, err)
	}
}

func TestStubJobCatalogRepository(t *testing.T) {
	ctx := context.Background()
	tenant := uuid.New()
	group := &types.JobFamilyGroup{Code: "JC-G", Status: "ACTIVE"}
	family := &types.JobFamily{Code: "JC-F", Status: "ACTIVE"}
	role := &types.JobRole{Code: "JC-R", Status: "ACTIVE"}
	level := &types.JobLevel{Code: "JC-L", Status: "ACTIVE"}

	repo := &StubJobCatalogRepository{
		GetCurrentFamilyGroupFn: func(context.Context, *sql.Tx, uuid.UUID, string) (*types.JobFamilyGroup, error) {
			return group, nil
		},
		GetCurrentJobFamilyFn: func(context.Context, *sql.Tx, uuid.UUID, string) (*types.JobFamily, error) {
			return family, nil
		},
		GetCurrentJobRoleFn: func(context.Context, *sql.Tx, uuid.UUID, string) (*types.JobRole, error) {
			return role, nil
		},
		GetCurrentJobLevelFn: func(context.Context, *sql.Tx, uuid.UUID, string) (*types.JobLevel, error) {
			return level, nil
		},
	}

	if got, err := repo.GetCurrentFamilyGroup(ctx, nil, tenant, "JC-G"); err != nil || got != group {
		t.Fatalf("expected job family group, got %v, err %v", got, err)
	}
	if got, err := repo.GetCurrentJobFamily(ctx, nil, tenant, "JC-F"); err != nil || got != family {
		t.Fatalf("expected job family, got %v, err %v", got, err)
	}
	if got, err := repo.GetCurrentJobRole(ctx, nil, tenant, "JC-R"); err != nil || got != role {
		t.Fatalf("expected job role, got %v, err %v", got, err)
	}
	if got, err := repo.GetCurrentJobLevel(ctx, nil, tenant, "JC-L"); err != nil || got != level {
		t.Fatalf("expected job level, got %v, err %v", got, err)
	}
}

func TestStubOrganizationRepository(t *testing.T) {
	ctx := context.Background()
	tenant := uuid.New()
	want := &types.Organization{Code: "1000001"}

	repo := &StubOrganizationRepository{
		GetByCodeFn: func(context.Context, uuid.UUID, string) (*types.Organization, error) {
			return want, nil
		},
	}

	if got, err := repo.GetByCode(ctx, tenant, "1000001"); err != nil || got != want {
		t.Fatalf("expected organization stub, got %v, err %v", got, err)
	}
}

func TestStubAssignmentRepository(t *testing.T) {
	ctx := context.Background()
	tenant := uuid.New()
	assignmentID := uuid.New()
	assignment := &types.PositionAssignment{AssignmentID: assignmentID}

	repo := &StubAssignmentRepository{
		GetByIDFn: func(context.Context, *sql.Tx, uuid.UUID, uuid.UUID) (*types.PositionAssignment, error) {
			return assignment, nil
		},
		SumActiveFTEFn: func(context.Context, *sql.Tx, uuid.UUID, string) (float64, error) {
			return 1.25, nil
		},
	}

	if got, err := repo.GetByID(ctx, nil, tenant, assignmentID); err != nil || got != assignment {
		t.Fatalf("expected assignment stub, got %v, err %v", got, err)
	}
	if total, err := repo.SumActiveFTE(ctx, nil, tenant, "POS-1"); err != nil || total != 1.25 {
		t.Fatalf("expected FTE total, got %f, err %v", total, err)
	}
}

func TestStubPositionRepository(t *testing.T) {
	ctx := context.Background()
	tenant := uuid.New()
	position := &types.Position{Code: "POS-1"}

	repo := &StubPositionRepository{
		GetCurrentPositionFn: func(context.Context, *sql.Tx, uuid.UUID, string) (*types.Position, error) {
			return position, nil
		},
	}

	if got, err := repo.GetCurrentPosition(ctx, nil, tenant, "POS-1"); err != nil || got != position {
		t.Fatalf("expected position stub, got %v, err %v", got, err)
	}
}
