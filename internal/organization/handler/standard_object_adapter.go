package handler

import (
	"context"
	"errors"
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
	defaultStandardObjectDisplayLabel = "organization"
)

func (h *OrganizationHandler) upsertStandardObject(ctx context.Context, org *types.Organization, actorID string) (standardobject.ObjectAggregate, error) {
	txTime := h.clock.Now()
	aggregate := h.buildOrganizationAggregate(org, actorID, txTime)
	if h.standardObjects == nil {
		return aggregate, nil
	}
	if err := h.standardObjects.Upsert(ctx, aggregate); err != nil {
		return standardobject.ObjectAggregate{}, err
	}
	return aggregate, nil
}

func (h *OrganizationHandler) buildOrganizationAggregate(org *types.Organization, actorID string, txTime time.Time) standardobject.ObjectAggregate {
	if txTime.IsZero() {
		txTime = h.clock.Now()
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

func copyOrganization(src *types.Organization) *types.Organization {
	if src == nil {
		return nil
	}
	clone := *src
	return &clone
}

var errTimelineVersionNotFound = errors.New("timeline version not found")

func (h *OrganizationHandler) syncOrganizationTimeline(
	ctx context.Context,
	base *types.Organization,
	timeline *[]repositorypkg.TimelineVersion,
	target time.Time,
	reason string,
	actorID string,
	after func(standardobject.ObjectAggregate) error,
) error {
	if base == nil {
		return errTimelineVersionNotFound
	}
	if timeline == nil || len(*timeline) == 0 {
		aggregate, err := h.upsertStandardObject(ctx, base, actorID)
		if err != nil {
			return err
		}
		if after != nil {
			return after(aggregate)
		}
		return nil
	}
	if version := timelineVersionByDate(timeline, target); version != nil {
		return h.upsertTimelineAggregate(ctx, base, version, reason, actorID, after)
	}
	if version := currentTimelineVersion(timeline); version != nil {
		return h.upsertTimelineAggregate(ctx, base, version, reason, actorID, after)
	}
	last := &(*timeline)[len(*timeline)-1]
	return h.upsertTimelineAggregate(ctx, base, last, reason, actorID, after)
}

func (h *OrganizationHandler) upsertTimelineAggregate(ctx context.Context, base *types.Organization, version *repositorypkg.TimelineVersion, reason string, actorID string, after func(standardobject.ObjectAggregate) error) error {
	if version == nil {
		return errTimelineVersionNotFound
	}
	org := h.organizationFromTimeline(base, version, reason)
	aggregate, err := h.upsertStandardObject(ctx, org, actorID)
	if err != nil {
		return err
	}
	if after != nil {
		return after(aggregate)
	}
	return nil
}

func timelineVersionByDate(timeline *[]repositorypkg.TimelineVersion, target time.Time) *repositorypkg.TimelineVersion {
	if timeline == nil {
		return nil
	}
	for i := range *timeline {
		if sameDay((*timeline)[i].EffectiveDate, target) {
			return &(*timeline)[i]
		}
	}
	return nil
}

func currentTimelineVersion(timeline *[]repositorypkg.TimelineVersion) *repositorypkg.TimelineVersion {
	if timeline == nil {
		return nil
	}
	for i := range *timeline {
		if (*timeline)[i].IsCurrent {
			return &(*timeline)[i]
		}
	}
	return nil
}

func sameDay(a time.Time, b time.Time) bool {
	return a.UTC().Truncate(24 * time.Hour).Equal(b.UTC().Truncate(24 * time.Hour))
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
