package repository

import "testing"

func TestEnsureJoinedPath(t *testing.T) {
	if got := ensureJoinedPath("", "A/B"); got != "/A/B" {
		t.Fatalf("unexpected: %s", got)
	}
	if got := ensureJoinedPath("/root/", "/child"); got != "/root/child" {
		t.Fatalf("unexpected: %s", got)
	}
	if got := ensureJoinedPath("/root", "child"); got != "/root/child" {
		t.Fatalf("unexpected: %s", got)
	}
}
