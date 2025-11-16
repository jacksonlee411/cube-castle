package handler

import (
	"net/http/httptest"
	"testing"
)

func TestGetTenantIDFromRequest_DefaultAndHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	// default
	if id := getTenantIDFromRequest(req); id.String() == "" {
		t.Fatalf("expected default tenant id")
	}
	// header parse
	req.Header.Set("X-Tenant-ID", "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")
	if id := getTenantIDFromRequest(req); id.String() != "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9" {
		t.Fatalf("unexpected tenant id: %s", id)
	}
}

func TestGetOperatorFromRequest_Priorities(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("X-Mock-User", "alice")
	op := getOperatorFromRequest(req)
	if op.ID != "alice" || op.Name != "alice" {
		t.Fatalf("unexpected operator: %#v", op)
	}
	req.Header.Set("X-Actor-Name", "Alice Zhang")
	op = getOperatorFromRequest(req)
	if op.ID != "alice" || op.Name != "Alice Zhang" {
		t.Fatalf("unexpected operator: %#v", op)
	}
}

func TestGetIfMatchHeader_Parsing(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	if v := getIfMatchHeader(req); v != nil {
		t.Fatalf("expected nil when header absent")
	}
	req.Header.Set("If-Match", `W/"etag123"`)
	if v := getIfMatchHeader(req); v == nil || *v != "etag123" {
		t.Fatalf("unexpected If-Match: %#v", v)
	}
	req.Header.Set("If-Match", `"  spaced  "`)
	if v := getIfMatchHeader(req); v == nil || *v != "spaced" {
		t.Fatalf("unexpected If-Match trim: %#v", v)
	}
}
