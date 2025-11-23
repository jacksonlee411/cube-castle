package standardobject

import "context"

// NoopService is a placeholder adapter that keeps the dependency graph wired while Phase 402A runs.
type NoopService struct {
}

// NewNoopService creates a skeleton ObjectService whose behavior is entirely controlled by the toggle.
func NewNoopService() *NoopService {
	return &NoopService{}
}

// Upsert currently returns ErrAdapterNotConfigured to signal that the adapter is not wired.
func (s *NoopService) Upsert(ctx context.Context, aggregate ObjectAggregate) error {
	return ErrAdapterNotConfigured
}

// Get mirrors Upsert and keeps the command/query services aware of the rollout status.
func (s *NoopService) Get(ctx context.Context, key ObjectKey) (ObjectAggregate, error) {
	return ObjectAggregate{}, ErrAdapterNotConfigured
}
