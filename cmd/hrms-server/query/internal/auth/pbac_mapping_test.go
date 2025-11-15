package auth

import (
	"context"
	"testing"
)

func TestPBACMapping_PositionAssignmentAudit(t *testing.T) {
	checker := NewPBACPermissionChecker(nil, nil)
	// Without audit scope -> should fail
	ctx := SetUserContext(context.Background(), &Claims{
		UserID:      "user",
		TenantID:    "tenant",
		Permissions: []string{"position:read:history"},
	})
	if err := checker.CheckPermission(ctx, "positionAssignmentAudit"); err == nil {
		t.Fatalf("expected denial without position:assignments:audit")
	}
	// With audit scope -> should pass
	ctx = SetUserContext(context.Background(), &Claims{
		UserID:      "user",
		TenantID:    "tenant",
		Permissions: []string{"position:assignments:audit"},
	})
	if err := checker.CheckPermission(ctx, "positionAssignmentAudit"); err != nil {
		t.Fatalf("expected pass with audit scope: %v", err)
	}
}

func TestPBACMapping_Assignments_Read(t *testing.T) {
	checker := NewPBACPermissionChecker(nil, nil)
	// Missing scope -> deny
	ctx := SetUserContext(context.Background(), &Claims{
		UserID:   "user",
		TenantID: "tenant",
	})
	if err := checker.CheckPermission(ctx, "assignments"); err == nil {
		t.Fatalf("expected denial without position:read")
	}
	// With position:read -> allow
	ctx = SetUserContext(context.Background(), &Claims{
		UserID:      "user",
		TenantID:    "tenant",
		Permissions: []string{"position:read"},
	})
	if err := checker.CheckPermission(ctx, "assignments"); err != nil {
		t.Fatalf("expected pass with position:read: %v", err)
	}
}

func TestPBACMapping_HierarchyStatistics(t *testing.T) {
	checker := NewPBACPermissionChecker(nil, nil)
	ctx := SetUserContext(context.Background(), &Claims{UserID: "u", TenantID: "t"})
	if err := checker.CheckPermission(ctx, "hierarchyStatistics"); err == nil {
		t.Fatalf("expected denial without org:read:hierarchy")
	}
	ctx = SetUserContext(context.Background(), &Claims{
		UserID:      "u",
		TenantID:    "t",
		Permissions: []string{"org:read:hierarchy"},
	})
	if err := checker.CheckPermission(ctx, "hierarchyStatistics"); err != nil {
		t.Fatalf("expected pass with org:read:hierarchy: %v", err)
	}
}

func TestPBACMapping_JobCatalogRead(t *testing.T) {
	checker := NewPBACPermissionChecker(nil, nil)
	// jobFamilyGroups requires job-catalog:read
	ctx := SetUserContext(context.Background(), &Claims{UserID: "u", TenantID: "t"})
	if err := checker.CheckPermission(ctx, "jobFamilyGroups"); err == nil {
		t.Fatalf("expected denial without job-catalog:read")
	}
	ctx = SetUserContext(context.Background(), &Claims{
		UserID:      "u",
		TenantID:    "t",
		Permissions: []string{"job-catalog:read"},
	})
	if err := checker.CheckPermission(ctx, "jobFamilyGroups"); err != nil {
		t.Fatalf("expected pass with job-catalog:read: %v", err)
	}
}

func TestPBACMapping_AuditHistory(t *testing.T) {
	checker := NewPBACPermissionChecker(nil, nil)
	ctx := SetUserContext(context.Background(), &Claims{UserID: "u", TenantID: "t"})
	if err := checker.CheckPermission(ctx, "auditHistory"); err == nil {
		t.Fatalf("expected denial without org:read:audit")
	}
	ctx = SetUserContext(context.Background(), &Claims{
		UserID:      "u",
		TenantID:    "t",
		Permissions: []string{"org:read:audit"},
	})
	if err := checker.CheckPermission(ctx, "auditHistory"); err != nil {
		t.Fatalf("expected pass with org:read:audit: %v", err)
	}
}
