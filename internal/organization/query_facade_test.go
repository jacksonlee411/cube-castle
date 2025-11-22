package organization

import (
	"context"
	"errors"
	"testing"
	"time"

	"cube-castle/internal/organization/dto"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type fakeAssignmentRepo struct {
	assignmentsFn func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	historyFn     func(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	statsFn       func(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error)
}

func (f *fakeAssignmentRepo) GetPositionAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
	if f.assignmentsFn == nil {
		return nil, errors.New("assignmentsFn not configured")
	}
	return f.assignmentsFn(ctx, tenantID, positionCode, filter, pagination, sorting)
}

func (f *fakeAssignmentRepo) GetAssignmentHistory(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error) {
	if f.historyFn == nil {
		return nil, errors.New("historyFn not configured")
	}
	return f.historyFn(ctx, tenantID, positionCode, filter, pagination, sorting)
}

func (f *fakeAssignmentRepo) GetAssignmentStats(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error) {
	if f.statsFn == nil {
		return nil, errors.New("statsFn not configured")
	}
	return f.statsFn(ctx, tenantID, positionCode, organizationCode)
}

func TestAssignmentQueryFacade_UsesCache(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	fakeRepo := &fakeAssignmentRepo{}
	facade := NewAssignmentQueryFacade(fakeRepo, redisClient, nil, 5*time.Minute)

	tenant := uuid.New()
	positionCode := "P10001"

	callCount := 0
	fakeRepo.statsFn = func(_ context.Context, _ uuid.UUID, pos string, _ string) (*dto.AssignmentStats, error) {
		callCount++
		return &dto.AssignmentStats{
			PositionCodeField:  &pos,
			TotalCountField:    3,
			LastUpdatedAtField: time.Now(),
		}, nil
	}

	stats, err := facade.GetAssignmentStats(context.Background(), tenant, positionCode, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalAssignments() != 3 {
		t.Fatalf("expected total 3, got %d", stats.TotalAssignments())
	}
	if callCount != 1 {
		t.Fatalf("expected repo called once, got %d", callCount)
	}

	// Second call should hit cache
	stats, err = facade.GetAssignmentStats(context.Background(), tenant, positionCode, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalAssignments() != 3 {
		t.Fatalf("expected total 3, got %d", stats.TotalAssignments())
	}
	if callCount != 1 {
		t.Fatalf("expected repo call still 1, got %d", callCount)
	}
}
