package resolver

import (
	"context"
	"database/sql"
	"time"

	"cube-castle/internal/standardobject"
	stdrepo "cube-castle/internal/standardobject/repository"
	clockpkg "cube-castle/pkg/temporal/clock"
)

// StandardObjectStore exposes read-only access to Standard Object aggregates.
type StandardObjectStore interface {
	Get(ctx context.Context, key standardobject.ObjectKey, asOf time.Time) (standardobject.ObjectAggregate, error)
}

type repositoryStandardObjectStore struct {
	repo  *stdrepo.Repository
	clock clockpkg.Clock
}

// NewStandardObjectStore builds a repository-backed Standard Object reader.
func NewStandardObjectStore(db *sql.DB, clk clockpkg.Clock) StandardObjectStore {
	if clk == nil {
		clk = clockpkg.NewSystemClock()
	}
	return &repositoryStandardObjectStore{
		repo:  stdrepo.NewRepository(db, clk),
		clock: clk,
	}
}

func (s *repositoryStandardObjectStore) Get(ctx context.Context, key standardobject.ObjectKey, asOf time.Time) (standardobject.ObjectAggregate, error) {
	if asOf.IsZero() {
		asOf = s.clock.Now()
	}
	return s.repo.GetAt(ctx, key, asOf)
}
