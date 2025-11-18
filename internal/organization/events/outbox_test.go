package events

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestNewAssignmentEventDefaults(t *testing.T) {
	ctx := Context{
		TenantID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		RequestID:     "req-1",
		CorrelationID: "corr-1",
		Operation:     "assign",
	}

	ev, err := NewAssignmentEvent(EventAssignmentFilled, ctx, "", "POS123", map[string]interface{}{"extra": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.AggregateType != aggregateAssignment {
		t.Fatalf("expected aggregate assignment, got %s", ev.AggregateType)
	}
	if ev.AggregateID != "POS123" {
		t.Fatalf("aggregate id fallback failed: %s", ev.AggregateID)
	}

	payload := make(map[string]interface{})
	if err := json.Unmarshal([]byte(ev.Payload), &payload); err != nil {
		t.Fatalf("payload json invalid: %v", err)
	}
	if payload["positionCode"] != "POS123" {
		t.Fatalf("positionCode not set in payload: %#v", payload)
	}
	if payload["tenantId"] != ctx.TenantID.String() {
		t.Fatalf("tenantId missing in payload: %#v", payload)
	}
	if payload["source"] != DefaultSourceCommand {
		t.Fatalf("expected default source %s, got %v", DefaultSourceCommand, payload["source"])
	}
	if payload["eventType"] != EventAssignmentFilled {
		t.Fatalf("eventType missing in payload: %#v", payload)
	}
}

func TestNewPositionEventPayload(t *testing.T) {
	ctx := Context{TenantID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Source: "custom"}
	ev, err := NewPositionEvent(EventPositionCreated, ctx, "P-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payload := map[string]interface{}{}
	if err := json.Unmarshal([]byte(ev.Payload), &payload); err != nil {
		t.Fatalf("payload json invalid: %v", err)
	}
	if payload["positionCode"] != "P-1" {
		t.Fatalf("positionCode not set in payload: %#v", payload)
	}
	if payload["source"] != "custom" {
		t.Fatalf("source should respect context override: %#v", payload["source"])
	}
}

func TestNewJobLevelEventMissingAggregate(t *testing.T) {
	ctx := Context{}
	if _, err := newOutboxEvent(EventJobLevelVersionCreated, aggregateJobLevel, "", ctx, nil); err == nil {
		t.Fatalf("expected error when aggregate id is missing")
	}
}
