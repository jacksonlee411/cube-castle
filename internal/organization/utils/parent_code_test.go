package utils

import "testing"

func TestNormalizeParentCodePointer(t *testing.T) {
	// nil input -> nil
	if got := NormalizeParentCodePointer(nil); got != nil {
		t.Fatalf("expected nil for nil input, got %v", *got)
	}
	// empty/space -> nil
	s := "   "
	if got := NormalizeParentCodePointer(&s); got != nil {
		t.Fatalf("expected nil for blank input, got %#v", got)
	}
	// legacy root "0" -> nil
	s0 := "0"
	if got := NormalizeParentCodePointer(&s0); got != nil {
		t.Fatalf("expected nil for legacy root '0', got %#v", got)
	}
	// legacy root "0000000" -> nil
	sr := "0000000"
	if got := NormalizeParentCodePointer(&sr); got != nil {
		t.Fatalf("expected nil for legacy root '0000000', got %#v", got)
	}
	// normal value -> trimmed copy
	raw := "  1000000  "
	got := NormalizeParentCodePointer(&raw)
	if got == nil || *got != "1000000" {
		t.Fatalf("expected '1000000', got %#v", got)
	}
	// ensure it returns a new pointer (no aliasing with caller's var)
	raw = "changed"
	if *got == raw {
		t.Fatalf("expected independent copy, pointer reflects caller mutation")
	}
}

func TestIsRootParentCode(t *testing.T) {
	if !IsRootParentCode("0") {
		t.Fatalf("expected '0' to be recognized as root")
	}
	if !IsRootParentCode("   0000000  ") {
		t.Fatalf("expected '0000000' with spaces to be recognized as root")
	}
	if IsRootParentCode("1000000") {
		t.Fatalf("did not expect '1000000' to be recognized as root")
	}
}

