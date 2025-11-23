package standardobject

import (
	"context"
	"testing"
)

func TestNoopServiceFlag(t *testing.T) {
	ctx := context.Background()
	service := NewNoopService()

	if err := service.Upsert(ctx, ObjectAggregate{}); err != ErrAdapterNotConfigured {
		t.Fatalf("expected ErrAdapterNotConfigured, got %v", err)
	}
}
