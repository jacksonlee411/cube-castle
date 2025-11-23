package sqlc

import (
	"context"
	"database/sql"

	"cube-castle/internal/standardobject"
	"cube-castle/internal/standardobject/featureflag"
	"cube-castle/internal/standardobject/repository"
)

type service struct {
	repo standardobject.ObjectRepository
	flag featureflag.Toggle
}

// Provide constructs a sqlc-backed ObjectService when a database connection is available.
func Provide(db *sql.DB, toggle featureflag.Toggle) standardobject.ObjectService {
	if db == nil {
		return standardobject.NewNoopService(toggle)
	}
	if toggle == nil {
		toggle = featureflag.NewStaticToggle(false)
	}
	return &service{
		repo: repository.NewRepository(db),
		flag: toggle,
	}
}

func (s *service) Upsert(ctx context.Context, aggregate standardobject.ObjectAggregate) error {
	if !s.flag.Enabled(ctx) {
		return standardobject.ErrStandardObjectsDisabled
	}
	return s.repo.Upsert(ctx, aggregate)
}

func (s *service) Get(ctx context.Context, key standardobject.ObjectKey) (standardobject.ObjectAggregate, error) {
	if !s.flag.Enabled(ctx) {
		return standardobject.ObjectAggregate{}, standardobject.ErrStandardObjectsDisabled
	}
	return s.repo.Get(ctx, key)
}
