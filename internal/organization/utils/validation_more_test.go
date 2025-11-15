package utils

import "testing"

func TestValidateSuspendAndActivateRequest(t *testing.T) {
	// Empty is allowed
	if err := ValidateSuspendRequest(""); err != nil {
		t.Fatalf("empty suspend reason should be allowed: %v", err)
	}
	if err := ValidateActivateRequest("   "); err != nil {
		t.Fatalf("empty activate reason should be allowed: %v", err)
	}
	// Too short
	if err := ValidateSuspendRequest("abc"); err == nil {
		t.Fatalf("too short suspend reason should error")
	}
	if err := ValidateActivateRequest("abcd"); err == nil {
		t.Fatalf("too short activate reason should error")
	}
	// Too long
	long := make([]byte, 201)
	for i := range long {
		long[i] = 'x'
	}
	if err := ValidateSuspendRequest(string(long)); err == nil {
		t.Fatalf("too long suspend reason should error")
	}
	if err := ValidateActivateRequest(string(long)); err == nil {
		t.Fatalf("too long activate reason should error")
	}
	// Valid
	if err := ValidateSuspendRequest("valid reason"); err != nil {
		t.Fatalf("valid suspend reason should pass: %v", err)
	}
	if err := ValidateActivateRequest("valid reason"); err != nil {
		t.Fatalf("valid activate reason should pass: %v", err)
	}
}

