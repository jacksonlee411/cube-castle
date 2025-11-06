package validator

import (
	"strings"
	"testing"
)

func TestNewValidationFailedErrorWithoutResult(t *testing.T) {
	err := NewValidationFailedError("CreatePosition", nil)
	if err == nil {
		t.Fatal("expected error when result is nil")
	}
	if !strings.Contains(err.Error(), "validation failed without result") {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
}

func TestValidationFailedErrorWithResultAndDetails(t *testing.T) {
	result := NewValidationResult()
	result.Valid = false
	result.Errors = append(result.Errors, ValidationError{Code: "POS-ORG"})

	err := NewValidationFailedError("CreatePosition", result)
	if err == nil {
		t.Fatal("expected validation failed error")
	}

	vErr, ok := err.(*ValidationFailedError)
	if !ok {
		t.Fatalf("expected *ValidationFailedError, got %T", err)
	}
	if got := vErr.Operation(); got != "CreatePosition" {
		t.Fatalf("unexpected operation: %s", got)
	}
	if vErr.Result() != result {
		t.Fatal("expected result reference to be preserved")
	}

	msg := vErr.Error()
	if !strings.Contains(msg, "POS-ORG") {
		t.Fatalf("expected error code in message, got: %s", msg)
	}
}

func TestValidationFailedErrorWithoutErrorDetails(t *testing.T) {
	result := NewValidationResult()
	result.Valid = false

	err := NewValidationFailedError("FillPosition", result)
	if err == nil {
		t.Fatal("expected validation failed error")
	}

	vErr, ok := err.(*ValidationFailedError)
	if !ok {
		t.Fatalf("expected *ValidationFailedError, got %T", err)
	}

	if !strings.Contains(vErr.Error(), "FillPosition") {
		t.Fatalf("expected operation name in message, got: %s", vErr.Error())
	}
}
