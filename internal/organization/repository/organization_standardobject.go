package repository

import (
	"context"
	"strings"
	"time"

	"cube-castle/internal/standardobject"
	"cube-castle/internal/types"
	clockpkg "cube-castle/pkg/temporal/clock"
)

const (
	standardObjectSchemaVersion       = "v1"
	standardObjectRetentionPolicy     = "default"
	standardObjectDataClassification  = "INTERNAL"
	standardObjectHierarchyLinkType   = "ORG_HIERARCHY"
	defaultStandardObjectDisplayLabel = "organization"
)

// UpsertStandardObject persists the organization aggregate into the SOM repository.
func (r *OrganizationRepository) UpsertStandardObject(ctx context.Context, org *types.Organization, actorID string, svc standardobject.ObjectService, clk clockpkg.Clock) (standardobject.ObjectAggregate, error) {
	if org == nil {
		return standardobject.ObjectAggregate{}, nil
	}
	if clk == nil {
		clk = clockpkg.NewSystemClock()
	}
	txTime := clk.Now()
	aggregate := buildOrganizationAggregate(org, actorID, txTime)
	if svc == nil {
		return aggregate, nil
	}
	if err := svc.Upsert(ctx, aggregate); err != nil {
		return standardobject.ObjectAggregate{}, err
	}
	return aggregate, nil
}

func buildOrganizationAggregate(org *types.Organization, actorID string, txTime time.Time) standardobject.ObjectAggregate {
	if txTime.IsZero() {
		txTime = time.Now()
	}
	effective := dateOrFallback(org.EffectiveDate, txTime)
	versionCode := standardobject.MakeVersionCode(org.Code, effective, coalesceTime(org.UpdatedAt, txTime), org.RecordID)
	kernel := standardobject.ObjectKernel{
		ID:                 org.RecordID,
		ObjectType:         standardobject.ObjectTypeOrganizationUnit,
		Code:               org.Code,
		DisplayName:        nonEmpty(org.Name, defaultStandardObjectDisplayLabel),
		TenantCode:         org.TenantID,
		Status:             mapLifecycleStatus(org.Status),
		Labels:             map[string]string{"unitType": org.UnitType},
		SchemaVersion:      standardObjectSchemaVersion,
		DataClassification: standardObjectDataClassification,
		RetentionPolicy:    standardObjectRetentionPolicy,
		CreatedBy:          actorOrSystem(actorID),
		CreatedAt:          coalesceTime(org.CreatedAt, txTime),
		UpdatedAt:          coalesceTime(org.UpdatedAt, txTime),
	}

	version := standardobject.TemporalVersion{
		VersionCode:     versionCode,
		EffectiveDate:   effective,
		EndDate:         dateToTime(org.EndDate),
		IsCurrent:       org.IsCurrent,
		Payload:         buildOrganizationPayload(org),
		AuditTrail:      buildOrganizationAuditTrail(org),
		Checksum:        "",
		CreatedAt:       kernel.CreatedAt,
		UpdatedAt:       kernel.UpdatedAt,
		TransactionFrom: txTime,
		TransactionTo:   nil,
	}

	links := buildOrganizationLinks(org, kernel.CreatedBy, effective, version.EndDate, txTime)

	return standardobject.ObjectAggregate{
		Kernel:  kernel,
		Version: version,
		Links:   links,
	}
}

func buildOrganizationPayload(org *types.Organization) map[string]any {
	payload := map[string]any{
		"name":        org.Name,
		"unitType":    org.UnitType,
		"status":      org.Status,
		"description": org.Description,
		"level":       org.Level,
		"codePath":    org.CodePath,
		"namePath":    org.NamePath,
		"sortOrder":   org.SortOrder,
	}
	if org.ParentCode != nil {
		payload["parentCode"] = strings.TrimSpace(*org.ParentCode)
	}
	return payload
}

func buildOrganizationAuditTrail(org *types.Organization) map[string]any {
	audit := map[string]any{}
	if org.ChangeReason != nil && strings.TrimSpace(*org.ChangeReason) != "" {
		audit["changeReason"] = strings.TrimSpace(*org.ChangeReason)
	}
	return audit
}

func buildOrganizationLinks(org *types.Organization, createdBy string, validFrom time.Time, validTo *time.Time, txFrom time.Time) []standardobject.Link {
	if org.ParentCode == nil || strings.TrimSpace(*org.ParentCode) == "" {
		return nil
	}
	parent := strings.TrimSpace(*org.ParentCode)
	return []standardobject.Link{
		{
			LinkType:        standardObjectHierarchyLinkType,
			SourceCode:      org.Code,
			TargetCode:      parent,
			TargetType:      standardobject.ObjectTypeOrganizationUnit,
			TenantCode:      org.TenantID,
			ValidFrom:       validFrom,
			ValidTo:         validTo,
			TransactionFrom: txFrom,
			TransactionTo:   nil,
			CreatedBy:       createdBy,
		},
	}
}

func mapLifecycleStatus(status string) standardobject.LifecycleStatus {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "ACTIVE":
		return standardobject.StatusActive
	case "SUSPENDED", "INACTIVE":
		return standardobject.StatusSuspended
	case "RETIRED", "DELETED":
		return standardobject.StatusRetired
	case "READY", "PLANNED", "PENDING":
		return standardobject.StatusReady
	case "FILLED", "PARTIALLY_FILLED":
		return standardobject.StatusActive
	default:
		return standardobject.StatusDraft
	}
}

func dateOrFallback(d *types.Date, fallback time.Time) time.Time {
	if d == nil {
		return fallback
	}
	return d.Time
}

func dateToTime(d *types.Date) *time.Time {
	if d == nil {
		return nil
	}
	t := d.Time
	return &t
}

func coalesceTime(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}

func actorOrSystem(actor string) string {
	if strings.TrimSpace(actor) == "" {
		return "system"
	}
	return actor
}

func nonEmpty(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
