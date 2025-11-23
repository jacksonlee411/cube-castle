package standardobject

import (
	"context"
	"testing"

	"cube-castle/internal/standardobject/featureflag"
)

func TestNoopServiceFlag(t *testing.T) {
	ctx := context.Background()
	service := NewNoopService(featureflag.NewStaticToggle(false))

	if err := service.Upsert(ctx, ObjectAggregate{}); err != ErrStandardObjectsDisabled {
		t.Fatalf("expected ErrStandardObjectsDisabled, got %v", err)
	}

	service = NewNoopService(featureflag.NewStaticToggle(true))

	if err := service.Upsert(ctx, ObjectAggregate{}); err != ErrAdapterNotConfigured {
		t.Fatalf("expected ErrAdapterNotConfigured, got %v", err)
	}
}
