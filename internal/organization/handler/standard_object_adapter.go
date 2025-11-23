package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	repositorypkg "cube-castle/internal/organization/repository"
	"cube-castle/internal/standardobject"
	"cube-castle/internal/types"
)

const (
	standardObjectSchemaVersion       = "v1"
	standardObjectRetentionPolicy     = "default"
	standardObjectDataClassification  = "INTERNAL"
	standardObjectHierarchyLinkType   = "ORG_HIERARCHY"
	standardObjectVersionCodeDateFmt  = "20060102"
	defaultStandardObjectDisplayLabel = "organization"
)

func (h *OrganizationHandler) upsertStandardObject(ctx context.Context, org *types.Organization, actorID string) error {
	if h.standardObjects == nil {
		return nil
	}
	return h.standardObjects.Upsert(ctx, h.buildOrganizationAggregate(org, actorID))
}

func (h *OrganizationHandler) buildOrganizationAggregate(org *types.Organization, actorID string) standardobject.ObjectAggregate {
	now := h.clock.Now()
	effective := dateOrFallback(org.EffectiveDate, now)
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
		CreatedAt:          coalesceTime(org.CreatedAt, now),
		UpdatedAt:          coalesceTime(org.UpdatedAt, now),
	}

	version := standardobject.TemporalVersion{
		VersionCode:     buildVersionCode(org.Code, effective),
		EffectiveDate:   effective,
		EndDate:         dateToTime(org.EndDate),
		IsCurrent:       org.IsCurrent,
		Payload:         buildOrganizationPayload(org),
		AuditTrail:      buildOrganizationAuditTrail(org),
		Checksum:        "",
		CreatedAt:       kernel.CreatedAt,
		UpdatedAt:       kernel.UpdatedAt,
		TransactionFrom: now,
		TransactionTo:   nil,
	}

	links := buildOrganizationLinks(org, kernel.CreatedBy, effective, version.EndDate, now)

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
	case "SUSPENDED":
		return standardobject.StatusSuspended
	case "RETIRED":
		return standardobject.StatusRetired
	case "READY":
		return standardobject.StatusReady
	default:
		return standardobject.StatusDraft
	}
}

func buildVersionCode(code string, effective time.Time) string {
	return fmt.Sprintf("%s-%s", strings.TrimSpace(code), effective.Format(standardObjectVersionCodeDateFmt))
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

func (h *OrganizationHandler) organizationFromTimeline(base *types.Organization, version *repositorypkg.TimelineVersion, reason string) *types.Organization {
	result := *base
	result.RecordID = version.RecordID.String()
	result.Name = version.Name
	result.UnitType = version.UnitType
	result.Status = version.Status
	result.Level = version.Level
	result.CodePath = version.CodePath
	result.NamePath = version.NamePath
	result.ParentCode = version.ParentCode
	if version.Description != nil {
		result.Description = *version.Description
	}
	if version.SortOrder != nil {
		result.SortOrder = *version.SortOrder
	}
	result.EffectiveDate = types.NewDateFromTime(version.EffectiveDate)
	if version.EndDate != nil {
		result.EndDate = types.NewDateFromTime(*version.EndDate)
	} else {
		result.EndDate = nil
	}
	if trimmed := strings.TrimSpace(reason); trimmed != "" {
		result.ChangeReason = &trimmed
	} else {
		result.ChangeReason = nil
	}
	result.IsCurrent = version.IsCurrent
	result.CreatedAt = version.CreatedAt
	result.UpdatedAt = version.UpdatedAt
	return &result
}
