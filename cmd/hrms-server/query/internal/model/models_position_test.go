package model

import "testing"

func TestPositionAssignmentHistoryEmptyWhenNil(t *testing.T) {
	position := Position{}
	history := position.AssignmentHistory()
	if history == nil {
		t.Fatalf("expected empty slice, got nil")
	}
	if len(history) != 0 {
		t.Fatalf("expected empty slice, got length %d", len(history))
	}
}

func TestPositionAssignmentHistoryReturnsData(t *testing.T) {
	position := Position{
		AssignmentHistoryField: []PositionAssignment{
			{AssignmentIDField: "A1"},
		},
	}
	history := position.AssignmentHistory()
	if len(history) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(history))
	}
	if history[0].AssignmentIDField != "A1" {
		t.Fatalf("unexpected assignment id: %s", history[0].AssignmentIDField)
	}
}

func TestPositionCurrentAssignment(t *testing.T) {
	assign := &PositionAssignment{AssignmentIDField: "A1"}
	position := Position{
		CurrentAssignmentField: assign,
	}
	current := position.CurrentAssignment()
	if current == nil {
		t.Fatalf("expected assignment, got nil")
	}
	if current.AssignmentIDField != "A1" {
		t.Fatalf("unexpected assignment id: %s", current.AssignmentIDField)
	}
	if current != assign {
		t.Fatalf("expected pointer equality with source assignment")
	}
}
