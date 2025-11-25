package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"cube-castle/internal/standardobject"
	"cube-castle/internal/types"
	clockpkg "cube-castle/pkg/temporal/clock"
	"github.com/google/uuid"
)

const (
	positionSchemaVersion      = "v1"
	positionRetentionPolicy    = "default"
	positionDataClassification = "INTERNAL"
	positionOrgLinkType        = "POSITION_BELONGS_TO_ORG"
	defaultPositionDisplayName = "position"
)

// UpsertStandardObject copies the position aggregate into the SOM repository.
func (r *PositionRepository) UpsertStandardObject(ctx context.Context, position *types.Position, operator types.OperatedByInfo, svc standardobject.ObjectService, clk clockpkg.Clock) (standardobject.ObjectAggregate, error) {
	if position == nil {
		return standardobject.ObjectAggregate{}, nil
	}
	if clk == nil {
		clk = clockpkg.NewSystemClock()
	}
	txTime := clk.Now()
	aggregate := buildPositionAggregate(position, operator, txTime)
	if svc == nil {
		return aggregate, nil
	}
	if err := svc.Upsert(ctx, aggregate); err != nil {
		return standardobject.ObjectAggregate{}, err
	}
	return aggregate, nil
}

func buildPositionAggregate(position *types.Position, operator types.OperatedByInfo, txTime time.Time) standardobject.ObjectAggregate {
	effective := position.EffectiveDate
	updated := coalesceTime(position.UpdatedAt, txTime)
	versionCode := standardobject.MakeVersionCode(position.Code, effective, updated, position.RecordID.String())

	kernel := standardobject.ObjectKernel{
		ID:                 coalescePositionKernelID(position),
		ObjectType:         standardobject.ObjectTypePositionRole,
		Code:               position.Code,
		DisplayName:        nonEmpty(position.Title, defaultPositionDisplayName),
		TenantCode:         position.TenantID.String(),
		Status:             mapPositionLifecycleStatus(position.Status),
		Labels:             buildPositionLabels(position),
		SchemaVersion:      positionSchemaVersion,
		DataClassification: positionDataClassification,
		RetentionPolicy:    positionRetentionPolicy,
		CreatedBy:          resolvePositionActor(position, operator),
		CreatedAt:          coalesceTime(position.CreatedAt, txTime),
		UpdatedAt:          updated,
	}

	version := standardobject.TemporalVersion{
		VersionCode:     versionCode,
		EffectiveDate:   effective,
		EndDate:         nullTimePtr(position.EndDate),
		IsCurrent:       position.IsCurrent,
		Payload:         buildPositionPayload(position),
		AuditTrail:      buildPositionAuditTrail(position, operator),
		Checksum:        "",
		CreatedAt:       kernel.CreatedAt,
		UpdatedAt:       updated,
		TransactionFrom: txTime,
		TransactionTo:   nil,
	}

	links := buildPositionLinks(position, txTime, kernel.CreatedBy)

	return standardobject.ObjectAggregate{
		Kernel:  kernel,
		Version: version,
		Links:   links,
	}
}

func buildPositionPayload(position *types.Position) map[string]any {
	payload := map[string]any{
		"title":              position.Title,
		"jobFamilyGroupCode": position.JobFamilyGroupCode,
		"jobFamilyCode":      position.JobFamilyCode,
		"jobRoleCode":        position.JobRoleCode,
		"jobLevelCode":       position.JobLevelCode,
		"organizationCode":   position.OrganizationCode,
		"positionType":       position.PositionType,
		"employmentType":     position.EmploymentType,
		"status":             position.Status,
		"headcountCapacity":  position.HeadcountCapacity,
		"headcountInUse":     position.HeadcountInUse,
	}

	if val := nullStringValue(position.OrganizationName); val != "" {
		payload["organizationName"] = val
	}
	if val := nullStringValue(position.JobProfileCode); val != "" {
		payload["jobProfileCode"] = val
	}
	if val := nullStringValue(position.JobProfileName); val != "" {
		payload["jobProfileName"] = val
	}
	if val := nullStringValue(position.GradeLevel); val != "" {
		payload["gradeLevel"] = val
	}
	if val := nullStringValue(position.CostCenterCode); val != "" {
		payload["costCenterCode"] = val
	}
	if val := nullStringValue(position.ReportsToPosition); val != "" {
		payload["reportsToPositionCode"] = val
	}
	if len(strings.TrimSpace(string(position.Profile))) > 0 {
		var profile any
		if err := json.Unmarshal(position.Profile, &profile); err == nil {
			payload["profile"] = profile
		}
	}
	return payload
}

func buildPositionAuditTrail(position *types.Position, operator types.OperatedByInfo) map[string]any {
	audit := map[string]any{
		"operationType":   position.OperationType,
		"operationReason": nullStringValue(position.OperationReason),
	}
	if id, name := resolveOperatorInfo(operator); id != uuid.Nil {
		audit["operatedBy"] = map[string]string{
			"id":   id.String(),
			"name": name,
		}
	}
	return audit
}

func buildPositionLabels(position *types.Position) map[string]string {
	labels := map[string]string{
		"positionType":   strings.ToUpper(strings.TrimSpace(position.PositionType)),
		"employmentType": strings.ToUpper(strings.TrimSpace(position.EmploymentType)),
		"jobFamily":      position.JobFamilyCode,
		"jobRole":        position.JobRoleCode,
	}
	if level := strings.TrimSpace(position.JobLevelCode); level != "" {
		labels["jobLevel"] = level
	}
	return labels
}

func buildPositionLinks(position *types.Position, tx time.Time, createdBy string) []standardobject.Link {
	orgCode := strings.TrimSpace(position.OrganizationCode)
	if orgCode == "" {
		return nil
	}
	return []standardobject.Link{
		{
			LinkType:        positionOrgLinkType,
			SourceCode:      position.Code,
			TargetCode:      orgCode,
			TargetType:      standardobject.ObjectTypeOrganizationUnit,
			TenantCode:      position.TenantID.String(),
			ValidFrom:       position.EffectiveDate,
			ValidTo:         nullTimePtr(position.EndDate),
			TransactionFrom: tx,
			TransactionTo:   nil,
			CreatedBy:       createdBy,
			Attributes: map[string]any{
				"jobRole":      position.JobRoleCode,
				"jobFamily":    position.JobFamilyCode,
				"jobLevel":     position.JobLevelCode,
				"positionType": position.PositionType,
			},
		},
	}
}

func resolvePositionActor(position *types.Position, operator types.OperatedByInfo) string {
	if position.OperatedByID != uuid.Nil {
		return position.OperatedByID.String()
	}
	if strings.TrimSpace(operator.ID) != "" {
		return operator.ID
	}
	return "system"
}

func coalescePositionKernelID(position *types.Position) string {
	if position.RecordID != uuid.Nil {
		return position.RecordID.String()
	}
	if strings.TrimSpace(position.Code) != "" {
		return strings.TrimSpace(position.Code)
	}
	return uuid.New().String()
}

func mapPositionLifecycleStatus(status string) standardobject.LifecycleStatus {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "ACTIVE", "FILLED":
		return standardobject.StatusActive
	case "VACANT", "PLANNED", "PENDING":
		return standardobject.StatusReady
	case "SUSPENDED", "INACTIVE":
		return standardobject.StatusSuspended
	case "DELETED", "CLOSED", "RETIRED":
		return standardobject.StatusRetired
	default:
		return standardobject.StatusDraft
	}
}

func nullStringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return strings.TrimSpace(value.String)
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time
	return &t
}

func resolveOperatorInfo(operator types.OperatedByInfo) (uuid.UUID, string) {
	id := strings.TrimSpace(operator.ID)
	name := strings.TrimSpace(operator.Name)
	if id == "" {
		return uuid.Nil, name
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, name
	}
	return parsed, name
}
