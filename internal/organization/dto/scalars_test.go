package dto

import (
	"testing"
)

func TestScalarStringAndCodeUnmarshal(t *testing.T) {
	var code PositionCode
	// should trim spaces and accept string
	if err := code.UnmarshalGraphQL("  POS-1 "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != PositionCode("POS-1") {
		t.Fatalf("unexpected code: %s", code)
	}

	var date Date
	if err := date.UnmarshalGraphQL("2025-11-16"); err != nil {
		t.Fatalf("date unmarshal failed: %v", err)
	}
	if date != Date("2025-11-16") {
		t.Fatalf("unexpected date: %s", date)
	}
}

func TestScalarUnmarshalTypeError(t *testing.T) {
	var jobRole JobRoleCode
	if err := jobRole.UnmarshalGraphQL(123); err == nil {
		t.Fatalf("expected error for non-string input")
	}
}
