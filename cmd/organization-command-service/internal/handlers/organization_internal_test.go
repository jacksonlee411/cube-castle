package handlers

import (
	"net/http/httptest"
	"testing"
)

func TestGetIfMatchValue(t *testing.T) {
	h := &OrganizationHandler{}
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("If-Match", "\"abc123\"")

	value, err := h.getIfMatchValue(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "abc123" {
		t.Fatalf("expected abc123, got %s", value)
	}
}

func TestGetIfMatchValueWeak(t *testing.T) {
	h := &OrganizationHandler{}
	req := httptest.NewRequest("POST", "/", nil)
	req.Header.Set("If-Match", "W/\"abc123\"")

	value, err := h.getIfMatchValue(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "abc123" {
		t.Fatalf("expected abc123, got %s", value)
	}
}

func TestGetIfMatchValueMissing(t *testing.T) {
	h := &OrganizationHandler{}
	req := httptest.NewRequest("POST", "/", nil)

	if _, err := h.getIfMatchValue(req); err == nil {
		t.Fatalf("expected error, got nil")
	}
}
