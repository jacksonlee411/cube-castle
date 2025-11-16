package organization

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"cube-castle/internal/organization/dto"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// fake repo with configurable funcs (re-used pattern from existing test)
type fakeRepo struct {
	assignmentsFn func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	historyFn     func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	statsFn       func(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error)
}

func (f *fakeRepo) GetPositionAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
	if f.assignmentsFn == nil {
		return nil, errors.New("assignmentsFn not configured")
	}
	return f.assignmentsFn(ctx, tenantID, positionCode, filter, pagination, sorting)
}

func (f *fakeRepo) GetAssignmentHistory(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
	if f.historyFn == nil {
		return nil, errors.New("historyFn not configured")
	}
	return f.historyFn(ctx, tenantID, positionCode, filter, pagination, sorting)
}

func (f *fakeRepo) GetAssignmentStats(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error) {
	if f.statsFn == nil {
		return nil, errors.New("statsFn not configured")
	}
	return f.statsFn(ctx, tenantID, positionCode, organizationCode)
}

func Test_GetAssignments_RepoNil(t *testing.T) {
	facade := NewAssignmentQueryFacade(nil, nil, nil, 0)
	_, err := facade.GetAssignments(context.Background(), uuid.New(), "P1", nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected repo not configured error, got: %v", err)
	}
}

func Test_GetAssignments_DelegatesAndTrims(t *testing.T) {
	var gotPos string
	repo := &fakeRepo{
		assignmentsFn: func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
			gotPos = positionCode
			return &dto.PositionAssignmentConnection{}, nil
		},
	}
	facade := NewAssignmentQueryFacade(repo, nil, nil, 0)
	_, err := facade.GetAssignments(context.Background(), uuid.New(), "  P-001  ", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPos != "P-001" {
		t.Fatalf("expected trimmed position code, got %q", gotPos)
	}
}

func Test_GetAssignmentHistory_DelegatesAndTrims(t *testing.T) {
	var gotPos string
	repo := &fakeRepo{
		historyFn: func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
			gotPos = positionCode
			return &dto.PositionAssignmentConnection{}, nil
		},
	}
	facade := NewAssignmentQueryFacade(repo, nil, nil, 0)
	_, err := facade.GetAssignmentHistory(context.Background(), uuid.New(), "\tP-002\t", nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPos != "P-002" {
		t.Fatalf("expected trimmed position code, got %q", gotPos)
	}
}

func Test_GetAssignmentStats_NoCacheAndNilFromRepoReturnsEmpty(t *testing.T) {
	repo := &fakeRepo{
		statsFn: func(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error) {
			return nil, nil // simulate not found / nil stats
		},
	}
	facade := NewAssignmentQueryFacade(repo, nil, nil, 0)
	stats, err := facade.GetAssignmentStats(context.Background(), uuid.New(), "P-003", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats == nil {
		t.Fatalf("expected non-nil stats")
	}
	if stats.TotalAssignments() != 0 {
		t.Fatalf("expected zero totals from default stats, got %d", stats.TotalAssignments())
	}
}

func Test_RefreshPositionCache_NilRedisIsNoop(t *testing.T) {
	facade := NewAssignmentQueryFacade(&fakeRepo{}, nil, nil, 0)
	if err := facade.RefreshPositionCache(context.Background(), uuid.New(), "P-004"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func Test_RefreshPositionCache_DeletesKeys(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := &fakeRepo{
		statsFn: func(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error) {
			// return stable stats to populate cache
			return &dto.AssignmentStats{TotalCountField: 1, LastUpdatedAtField: time.Now()}, nil
		},
	}
	facade := NewAssignmentQueryFacade(repo, rdb, nil, time.Minute)
	tenant := uuid.New()
	code := "pos-xyz"

	// Prime the cache by calling GetAssignmentStats once.
	if _, err := facade.GetAssignmentStats(context.Background(), tenant, code, ""); err != nil {
		t.Fatalf("setup error: %v", err)
	}
	// Verify key exists
	prefix := "org:assignment:stats:" + tenant.String() + ":" + strings.ToUpper(code)
	keys := mr.Keys()
	has := false
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) {
			has = true
			break
		}
	}
	if !has {
		t.Fatalf("expected cache key with prefix %q to exist", prefix)
	}

	// Now refresh (invalidate) the cache; keys should be deleted
	if err := facade.RefreshPositionCache(context.Background(), tenant, code); err != nil {
		t.Fatalf("refresh error: %v", err)
	}
	// Ensure keys are gone
	for _, k := range mr.Keys() {
		if strings.HasPrefix(k, prefix) {
			t.Fatalf("expected cache key %q to be deleted", k)
		}
	}
}

func Test_statsCacheKey_UppercasesCode(t *testing.T) {
	facade := NewAssignmentQueryFacade(&fakeRepo{}, nil, nil, 0)
	tenant := uuid.New()
	key := facade.statsCacheKey(tenant, "pos-abc")
	if !strings.Contains(key, strings.ToUpper("pos-abc")) {
		t.Fatalf("expected uppercase code in cache key, got %q", key)
	}
}
