package standardobject

import (
	"context"

	"cube-castle/internal/standardobject/featureflag"
)

// NoopService is a placeholder adapter that keeps the dependency graph wired while Phase 402A runs.
type NoopService struct {
	flag featureflag.Toggle
}

// NewNoopService creates a skeleton ObjectService whose behavior is entirely controlled by the toggle.
func NewNoopService(toggle featureflag.Toggle) *NoopService {
	if toggle == nil {
		toggle = featureflag.NewStaticToggle(false)
	}
	return &NoopService{flag: toggle}
}

// Upsert returns ErrStandardObjectsDisabled when the flag is off, otherwise ErrAdapterNotConfigured.
func (s *NoopService) Upsert(ctx context.Context, aggregate ObjectAggregate) error {
	if !s.flag.Enabled(ctx) {
		return ErrStandardObjectsDisabled
	}
	return ErrAdapterNotConfigured
}

// Get mirrors Upsert and keeps the command/query services aware of the rollout status.
func (s *NoopService) Get(ctx context.Context, key ObjectKey) (ObjectAggregate, error) {
	if !s.flag.Enabled(ctx) {
		return ObjectAggregate{}, ErrStandardObjectsDisabled
	}
	return ObjectAggregate{}, ErrAdapterNotConfigured
}
