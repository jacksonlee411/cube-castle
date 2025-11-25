package standardobject

import (
	"strings"
	"testing"
	"time"
)

func TestMakeVersionCodeIncludesTimestampAndEntropy(t *testing.T) {
	eff := time.Date(2025, time.January, 2, 0, 0, 0, 0, time.UTC)
	upd := time.Date(2025, time.January, 3, 12, 34, 56, 0, time.UTC)
	out := MakeVersionCode("org-001", eff, upd, "123e4567-e89b-12d3-a456-426614174000")
	if !strings.HasPrefix(out, "ORG-001-20250102-123456.000Z-") {
		t.Fatalf("unexpected prefix: %s", out)
	}
	if !strings.HasSuffix(out, "-123E4567") {
		t.Fatalf("unexpected entropy suffix: %s", out)
	}
}

func TestMakeVersionCodeHandlesMissingEntropy(t *testing.T) {
	eff := time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC)
	upd := time.Date(2025, time.March, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	out := MakeVersionCode(" OU-2 ", eff, upd, "")
	expected := "OU-2-20250314-020000.000Z"
	if out != expected {
		t.Fatalf("expected %s, got %s", expected, out)
	}
}
