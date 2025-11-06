package service

import (
	"errors"
	"testing"

	validator "cube-castle/internal/organization/validator"
	"cube-castle/internal/types"
	"github.com/google/uuid"
)

func TestNewHeadcountExceededError(t *testing.T) {
	svc := &PositionService{}
	position := &types.Position{Code: "POS001", HeadcountCapacity: 5.0}
	err := svc.newHeadcountExceededError("FillPosition", position, 4.0, 1.5, 5.5)
	var validationErr *validator.ValidationFailedError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	res := validationErr.Result()
	if res == nil || res.Valid {
		t.Fatalf("expected invalid validation result")
	}
	if len(res.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
	item := res.Errors[0]
	if item.Code != "POS_HEADCOUNT_EXCEEDED" {
		t.Fatalf("expected code POS_HEADCOUNT_EXCEEDED, got %s", item.Code)
	}
	if item.Context["ruleId"] != "POS-HEADCOUNT" {
		t.Fatalf("expected ruleId POS-HEADCOUNT, got %#v", item.Context["ruleId"])
	}
	if res.Context["positionCode"] != "POS001" {
		t.Fatalf("expected positionCode in context")
	}
}

func TestNewAssignmentStateError(t *testing.T) {
	svc := &PositionService{}
	assignment := &types.PositionAssignment{AssignmentStatus: "ended", AssignmentID: uuidMust("11111111-1111-1111-1111-111111111111"), PositionCode: "POS001"}
	err := svc.newAssignmentStateError("UpdateAssignment", assignment)
	var validationErr *validator.ValidationFailedError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	res := validationErr.Result()
	if len(res.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
	item := res.Errors[0]
	if item.Code != "ASSIGN_INVALID_STATE" {
		t.Fatalf("expected code ASSIGN_INVALID_STATE, got %s", item.Code)
	}
	if item.Context["ruleId"] != "ASSIGN-STATE" {
		t.Fatalf("expected ruleId ASSIGN-STATE, got %#v", item.Context["ruleId"])
	}
}

func TestNewAssignmentFTEError(t *testing.T) {
	svc := &PositionService{}
	position := &types.Position{Code: "POS001"}
	err := svc.newAssignmentFTEError("CreateAssignment", position, -0.5)
	var validationErr *validator.ValidationFailedError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected validation error, got %v", err)
	}
	res := validationErr.Result()
	if len(res.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
	item := res.Errors[0]
	if item.Code != "ASSIGN_FTE_LIMIT" {
		t.Fatalf("expected code ASSIGN_FTE_LIMIT, got %s", item.Code)
	}
	if item.Context["ruleId"] != "ASSIGN-FTE" {
		t.Fatalf("expected ruleId ASSIGN-FTE, got %#v", item.Context["ruleId"])
	}
}

func uuidMust(value string) uuid.UUID {
	id, err := uuid.Parse(value)
	if err != nil {
		panic(err)
	}
	return id
}
