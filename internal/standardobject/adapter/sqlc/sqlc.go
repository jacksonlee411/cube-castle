package sqlc

import (
	"context"
	"database/sql"

	"cube-castle/internal/standardobject"
	"cube-castle/internal/standardobject/repository"
	clockpkg "cube-castle/pkg/temporal/clock"
)

type service struct {
	repo standardobject.ObjectRepository
}

// Provide constructs a sqlc-backed ObjectService when a database connection is available.
func Provide(db *sql.DB, clk clockpkg.Clock) standardobject.ObjectService {
	if db == nil {
		return standardobject.NewNoopService()
	}
	if clk == nil {
		clk = clockpkg.NewSystemClock()
	}
	return &service{
		repo: repository.NewRepository(db, clk),
	}
}

func (s *service) Upsert(ctx context.Context, aggregate standardobject.ObjectAggregate) error {
	return s.repo.Upsert(ctx, aggregate)
}

func (s *service) Get(ctx context.Context, key standardobject.ObjectKey) (standardobject.ObjectAggregate, error) {
	return s.repo.Get(ctx, key)
}
