package service

import (
	"context"
	"fmt"

	"cube-castle/internal/standardobject"
	"cube-castle/internal/types"
)

func (s *PositionService) syncPositionStandardObject(ctx context.Context, position *types.Position, operator types.OperatedByInfo) (standardobject.ObjectAggregate, error) {
	aggregate, err := s.positions.UpsertStandardObject(ctx, position, operator, s.standardObjects, s.clock)
	if err != nil {
		return standardobject.ObjectAggregate{}, fmt.Errorf("position som upsert: %w", err)
	}
	return aggregate, nil
}
