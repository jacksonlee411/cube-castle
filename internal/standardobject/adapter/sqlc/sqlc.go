package sqlc

import (
	"context"
	"database/sql"
	"fmt"

	"cube-castle/internal/standardobject"
	"cube-castle/internal/standardobject/repository"
	clockpkg "cube-castle/pkg/temporal/clock"
)

type service struct {
	repo standardobject.ObjectRepository
}

// Provide constructs a sqlc-backed ObjectService when a database connection is available.
func Provide(db *sql.DB, clk clockpkg.Clock) (standardobject.ObjectService, error) {
	if db == nil {
		return nil, fmt.Errorf("standardobject sqlc adapter requires a database handle")
	}
	if clk == nil {
		clk = clockpkg.NewSystemClock()
	}
	return &service{
		repo: repository.NewRepository(db, clk),
	}, nil
}

func (s *service) Upsert(ctx context.Context, aggregate standardobject.ObjectAggregate) error {
	return s.repo.Upsert(ctx, aggregate)
}

func (s *service) Get(ctx context.Context, key standardobject.ObjectKey) (standardobject.ObjectAggregate, error) {
	return s.repo.Get(ctx, key)
}
