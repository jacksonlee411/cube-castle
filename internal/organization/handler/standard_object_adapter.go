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

func (h *OrganizationHandler) upsertStandardObject(ctx context.Context, org *types.Organization, actorID string) (standardobject.ObjectAggregate, error) {
	return h.repo.UpsertStandardObject(ctx, org, actorID, h.standardObjects, h.clock)
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
