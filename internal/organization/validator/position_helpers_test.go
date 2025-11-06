package validator

import (
	"context"
	"testing"

	"cube-castle/internal/types"
	"github.com/google/uuid"
)

func TestMergeJobCatalogContext(t *testing.T) {
	svc := &positionAssignmentValidationService{}
	result := NewValidationResult()
	req := &types.PositionRequest{
		JobFamilyGroupCode: "GRP-1",
		JobFamilyCode:      "FAM-1",
		JobRoleCode:        "ROLE-1",
		JobLevelCode:       "L1",
	}
	tenant := uuid.New()

	svc.mergeJobCatalogContext(context.Background(), result, tenant, req)

	jobCatalog, ok := result.Context["jobCatalog"].(map[string]string)
	if !ok {
		t.Fatalf("expected jobCatalog context map, got %v", result.Context["jobCatalog"])
	}

	if jobCatalog["group"] != "GRP-1" || jobCatalog["family"] != "FAM-1" ||
		jobCatalog["role"] != "ROLE-1" || jobCatalog["level"] != "L1" {
		t.Fatalf("unexpected job catalog context: %+v", jobCatalog)
	}

	if result.Context["tenantId"] != tenant.String() {
		t.Fatalf("expected tenantId in context, got %v", result.Context["tenantId"])
	}
}

func TestResolveOrgContext(t *testing.T) {
	tenant := uuid.New()
	request := &types.PositionRequest{OrganizationCode: "ORG-01"}
	position := &types.Position{OrganizationCode: "ORG-02"}
	org := &types.Organization{Code: "ORG-02", Status: "ACTIVE"}

	svc := &positionAssignmentValidationService{}

	if tenantID, code := svc.resolveOrgContext(&positionCreateSubject{TenantID: tenant, Request: request}); tenantID != tenant || code != "ORG-01" {
		t.Fatalf("unexpected resolve for create subject: %s %s", tenantID, code)
	}

	if tenantID, code := svc.resolveOrgContext(&positionFillSubject{TenantID: tenant, Position: position}); tenantID != tenant || code != "ORG-02" {
		t.Fatalf("unexpected resolve for fill subject: %s %s", tenantID, code)
	}

	if tenantID, code := svc.resolveOrgContext(&assignmentCreateSubject{TenantID: tenant, Position: position}); tenantID != tenant || code != "ORG-02" {
		t.Fatalf("unexpected resolve for assignment create: %s %s", tenantID, code)
	}

	if tenantID, code := svc.resolveOrgContext(&positionTransferSubject{TenantID: tenant, Target: "ORG-03"}); tenantID != tenant || code != "ORG-03" {
		t.Fatalf("unexpected resolve for transfer: %s %s", tenantID, code)
	}

	if tenantID, code := svc.resolveOrgContext(struct{}{}); tenantID != uuid.Nil || code != "" {
		t.Fatalf("unexpected resolve for unknown subject: %s %s", tenantID, code)
	}

	if _, code := svc.resolveOrgContext(&assignmentUpdateSubject{TenantID: tenant, Position: position, Organization: org}); code != "ORG-02" {
		t.Fatalf("unexpected resolve for assignment update: %s", code)
	}
}

func TestExtractHelpers(t *testing.T) {
	tenant := uuid.New()
	position := &types.Position{Code: "POS-1"}
	org := &types.Organization{Code: "ORG-02"}
	assignment := &types.PositionAssignment{AssignmentID: uuid.New(), AssignmentStatus: "ACTIVE"}

	fill := &positionFillSubject{
		TenantID:     tenant,
		Position:     position,
		Organization: org,
		CurrentFTE:   0.5,
		RequestedFTE: 0.6,
	}

	create := &assignmentCreateSubject{
		TenantID:     tenant,
		Position:     position,
		Organization: org,
		CurrentFTE:   0.5,
		RequestedFTE: 0.4,
	}

	update := &assignmentUpdateSubject{
		TenantID:         tenant,
		Position:         position,
		Organization:     org,
		CurrentFTE:       0.6,
		RequestedFTE:     0.3,
		OriginalFTE:      0.2,
		Assignment:       assignment,
		AssignmentStatus: "ENDED",
	}

	closeSubject := &assignmentCloseSubject{
		TenantID:   tenant,
		Assignment: assignment,
	}

	svc := &positionAssignmentValidationService{}

	if got := svc.extractTenant(fill); got != tenant {
		t.Fatalf("unexpected tenant from fill: %s", got)
	}
	if got := svc.extractTenant(closeSubject); got != tenant {
		t.Fatalf("unexpected tenant from close: %s", got)
	}
	if got := svc.extractTenant(struct{}{}); got != uuid.Nil {
		t.Fatalf("expected Nil tenant, got: %s", got)
	}

	if op := svc.extractOperation(fill); op != "FillPosition" {
		t.Fatalf("unexpected operation for fill: %s", op)
	}
	if op := svc.extractOperation(create); op != "CreateAssignment" {
		t.Fatalf("unexpected operation for create: %s", op)
	}
	if op := svc.extractOperation(update); op != "UpdateAssignment" {
		t.Fatalf("unexpected operation for update: %s", op)
	}
	if op := svc.extractOperation(closeSubject); op != "CloseAssignment" {
		t.Fatalf("unexpected operation for close: %s", op)
	}
	if op := svc.extractOperation(struct{}{}); op != "Unknown" {
		t.Fatalf("expected Unknown operation, got: %s", op)
	}

	if fte := svc.extractRequestedFTE(fill); fte != 0.6 {
		t.Fatalf("unexpected requested FTE for fill: %.2f", fte)
	}
	if fte := svc.extractRequestedFTE(update); fte != 0.3 {
		t.Fatalf("unexpected requested FTE for update: %.2f", fte)
	}
	if fte := svc.extractRequestedFTE(struct{}{}); fte != 1.0 {
		t.Fatalf("expected default 1.0 FTE, got %.2f", fte)
	}

	if original := svc.extractOriginalFTE(update); original != 0.2 {
		t.Fatalf("unexpected original FTE: %.2f", original)
	}
	if original := svc.extractOriginalFTE(fill); original != 0 {
		t.Fatalf("expected zero original FTE for fill: %.2f", original)
	}

	if current := svc.extractCurrentFTE(create); current != 0.5 {
		t.Fatalf("unexpected current FTE for create: %.2f", current)
	}
	if current := svc.extractCurrentFTE(struct{}{}); current != 0 {
		t.Fatalf("expected zero current FTE default, got %.2f", current)
	}

	if pos := svc.extractPosition(update); pos != position {
		t.Fatalf("unexpected position from update: %v", pos)
	}
	if pos := svc.extractPosition(struct{}{}); pos != nil {
		t.Fatalf("expected nil position, got: %v", pos)
	}

	if organization := svc.extractOrganization(create); organization != org {
		t.Fatalf("unexpected organization from create: %v", organization)
	}
	if organization := svc.extractOrganization(struct{}{}); organization != nil {
		t.Fatalf("expected nil organization, got: %v", organization)
	}
}

func TestResolveRequestedFTEAndInactiveStatusHelpers(t *testing.T) {
	value := 0.75
	if got := resolveRequestedFTE(&value); got != 0.75 {
		t.Fatalf("unexpected resolved FTE: %.2f", got)
	}
	if got := resolveRequestedFTE(nil); got != 1.0 {
		t.Fatalf("expected default FTE 1.0, got %.2f", got)
	}

	if !isInactivePositionStatus("inactive") {
		t.Fatal("expected inactive status to be detected regardless of case")
	}
	if !isInactivePositionStatus(" deleted ") {
		t.Fatal("expected deleted status with spaces to be detected")
	}
	if isInactivePositionStatus("active") {
		t.Fatal("did not expect active status to be treated as inactive")
	}
}
