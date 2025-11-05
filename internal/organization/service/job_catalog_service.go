package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"cube-castle/internal/organization/audit"
	orgmiddleware "cube-castle/internal/organization/middleware"
	"cube-castle/internal/organization/repository"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrJobCatalogParentMissing      = errors.New("job catalog parent not found")
	ErrJobCatalogInvalidInput       = errors.New("job catalog input invalid")
	ErrJobCatalogConflict           = errors.New("job catalog conflict")
	ErrJobCatalogPreconditionFailed = errors.New("job catalog precondition failed")
)

type JobCatalogService struct {
	repo        *repository.JobCatalogRepository
	auditLogger *audit.AuditLogger
	logger      pkglogger.Logger
}

func NewJobCatalogService(repo *repository.JobCatalogRepository, auditLogger *audit.AuditLogger, baseLogger pkglogger.Logger) *JobCatalogService {
	return &JobCatalogService{
		repo:        repo,
		auditLogger: auditLogger,
		logger:      scopedLogger(baseLogger, "jobCatalog", nil),
	}
}

func (s *JobCatalogService) CreateJobFamilyGroup(ctx context.Context, tenantID uuid.UUID, req *types.CreateJobFamilyGroupRequest, operator types.OperatedByInfo) (*types.JobFamilyGroup, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	entity, err := s.repo.InsertFamilyGroup(ctx, tx, tenantID, req)
	if err != nil {
		return nil, err
	}

	after := map[string]interface{}{
		"code":        entity.Code,
		"effectiveAt": entity.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeCreate, "CreateJobFamilyGroup", entity.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *JobCatalogService) CreateJobFamilyGroupVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, operator types.OperatedByInfo) (*types.JobFamilyGroup, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	entity, err := s.repo.InsertFamilyGroupVersion(ctx, tx, tenantID, code, req)
	if err != nil {
		return nil, err
	}
	after := map[string]interface{}{
		"code":        entity.Code,
		"effectiveAt": entity.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeCreate, "CreateJobFamilyGroupVersion", entity.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *JobCatalogService) CreateJobFamily(ctx context.Context, tenantID uuid.UUID, req *types.CreateJobFamilyRequest, operator types.OperatedByInfo) (*types.JobFamily, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parent, err := s.repo.GetCurrentFamilyGroup(ctx, tx, tenantID, req.JobFamilyGroupCode)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, ErrJobCatalogParentMissing
	}

	entity, err := s.repo.InsertJobFamily(ctx, tx, tenantID, parent.RecordID, req)
	if err != nil {
		return nil, err
	}

	after := map[string]interface{}{
		"code":        entity.Code,
		"groupCode":   entity.FamilyGroupCode,
		"effectiveAt": entity.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeCreate, "CreateJobFamily", entity.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *JobCatalogService) CreateJobFamilyVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, operator types.OperatedByInfo) (*types.JobFamily, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if req.ParentRecordID == nil {
		return nil, ErrJobCatalogInvalidInput
	}
	parentUUID, parseErr := uuid.Parse(strings.TrimSpace(*req.ParentRecordID))
	if parseErr != nil {
		return nil, fmt.Errorf("invalid parentRecordId: %w", parseErr)
	}

	entity, err := s.repo.InsertJobFamilyVersion(ctx, tx, tenantID, code, parentUUID, req)
	if err != nil {
		return nil, err
	}

	after := map[string]interface{}{
		"code":        entity.Code,
		"groupCode":   entity.FamilyGroupCode,
		"effectiveAt": entity.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeCreate, "CreateJobFamilyVersion", entity.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *JobCatalogService) CreateJobRole(ctx context.Context, tenantID uuid.UUID, req *types.CreateJobRoleRequest, operator types.OperatedByInfo) (*types.JobRole, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parent, err := s.repo.GetCurrentJobFamily(ctx, tx, tenantID, req.JobFamilyCode)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, ErrJobCatalogParentMissing
	}

	entity, err := s.repo.InsertJobRole(ctx, tx, tenantID, parent.RecordID, req)
	if err != nil {
		return nil, err
	}

	after := map[string]interface{}{
		"code":        entity.Code,
		"familyCode":  entity.FamilyCode,
		"effectiveAt": entity.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeCreate, "CreateJobRole", entity.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *JobCatalogService) CreateJobRoleVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, operator types.OperatedByInfo) (*types.JobRole, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if req.ParentRecordID == nil {
		return nil, ErrJobCatalogInvalidInput
	}
	parentUUID, parseErr := uuid.Parse(strings.TrimSpace(*req.ParentRecordID))
	if parseErr != nil {
		return nil, fmt.Errorf("invalid parentRecordId: %w", parseErr)
	}

	entity, err := s.repo.InsertJobRoleVersion(ctx, tx, tenantID, code, parentUUID, req)
	if err != nil {
		return nil, err
	}

	after := map[string]interface{}{
		"code":        entity.Code,
		"familyCode":  entity.FamilyCode,
		"effectiveAt": entity.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeCreate, "CreateJobRoleVersion", entity.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *JobCatalogService) CreateJobLevel(ctx context.Context, tenantID uuid.UUID, req *types.CreateJobLevelRequest, operator types.OperatedByInfo) (*types.JobLevel, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	parent, err := s.repo.GetCurrentJobRole(ctx, tx, tenantID, req.JobRoleCode)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, ErrJobCatalogParentMissing
	}

	entity, err := s.repo.InsertJobLevel(ctx, tx, tenantID, parent.RecordID, req)
	if err != nil {
		return nil, err
	}

	after := map[string]interface{}{
		"code":        entity.Code,
		"roleCode":    entity.RoleCode,
		"effectiveAt": entity.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeCreate, "CreateJobLevel", entity.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *JobCatalogService) CreateJobLevelVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, operator types.OperatedByInfo) (*types.JobLevel, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if req.ParentRecordID == nil {
		return nil, ErrJobCatalogInvalidInput
	}
	parentUUID, parseErr := uuid.Parse(strings.TrimSpace(*req.ParentRecordID))
	if parseErr != nil {
		return nil, fmt.Errorf("invalid parentRecordId: %w", parseErr)
	}

	entity, err := s.repo.InsertJobLevelVersion(ctx, tx, tenantID, code, parentUUID, req)
	if err != nil {
		return nil, err
	}

	after := map[string]interface{}{
		"code":        entity.Code,
		"roleCode":    entity.RoleCode,
		"effectiveAt": entity.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeCreate, "CreateJobLevelVersion", entity.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *JobCatalogService) UpdateJobFamilyGroup(ctx context.Context, tenantID uuid.UUID, code string, req *types.UpdateJobFamilyGroupRequest, ifMatch *string, operator types.OperatedByInfo) (*types.JobFamilyGroup, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	current, err := s.repo.GetCurrentFamilyGroup(ctx, tx, tenantID, normalizedCode)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrJobCatalogNotFound
	}

	if ifMatch != nil && current.RecordID.String() != strings.TrimSpace(*ifMatch) {
		return nil, ErrJobCatalogPreconditionFailed
	}

	updated, err := s.repo.UpdateFamilyGroup(ctx, tx, tenantID, normalizedCode, current.RecordID, req)
	if err != nil {
		return nil, s.mapUpdateError(err)
	}

	after := map[string]interface{}{
		"code":        updated.Code,
		"status":      updated.Status,
		"effectiveAt": updated.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeUpdate, "UpdateJobFamilyGroup", updated.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *JobCatalogService) UpdateJobFamily(ctx context.Context, tenantID uuid.UUID, code string, req *types.UpdateJobFamilyRequest, ifMatch *string, operator types.OperatedByInfo) (*types.JobFamily, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	current, err := s.repo.GetCurrentJobFamily(ctx, tx, tenantID, normalizedCode)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrJobCatalogNotFound
	}

	if ifMatch != nil && current.RecordID.String() != strings.TrimSpace(*ifMatch) {
		return nil, ErrJobCatalogPreconditionFailed
	}

	groupCode := current.FamilyGroupCode
	parentRecord := current.ParentRecord
	if req.JobFamilyGroupCode != nil {
		normalizedGroup := strings.ToUpper(strings.TrimSpace(*req.JobFamilyGroupCode))
		if normalizedGroup == "" {
			return nil, ErrJobCatalogInvalidInput
		}
		group, err := s.repo.GetCurrentFamilyGroup(ctx, tx, tenantID, normalizedGroup)
		if err != nil {
			return nil, err
		}
		if group == nil {
			return nil, ErrJobCatalogParentMissing
		}
		groupCode = group.Code
		parentRecord = group.RecordID
	}

	updated, err := s.repo.UpdateJobFamily(ctx, tx, tenantID, normalizedCode, current.RecordID, groupCode, parentRecord, req)
	if err != nil {
		return nil, s.mapUpdateError(err)
	}

	after := map[string]interface{}{
		"code":        updated.Code,
		"groupCode":   updated.FamilyGroupCode,
		"effectiveAt": updated.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeUpdate, "UpdateJobFamily", updated.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *JobCatalogService) UpdateJobRole(ctx context.Context, tenantID uuid.UUID, code string, req *types.UpdateJobRoleRequest, ifMatch *string, operator types.OperatedByInfo) (*types.JobRole, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	current, err := s.repo.GetCurrentJobRole(ctx, tx, tenantID, normalizedCode)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrJobCatalogNotFound
	}

	if ifMatch != nil && current.RecordID.String() != strings.TrimSpace(*ifMatch) {
		return nil, ErrJobCatalogPreconditionFailed
	}

	familyCode := current.FamilyCode
	parentRecord := current.ParentRecord
	if req.JobFamilyCode != nil {
		normalizedFamily := strings.ToUpper(strings.TrimSpace(*req.JobFamilyCode))
		if normalizedFamily == "" {
			return nil, ErrJobCatalogInvalidInput
		}
		family, err := s.repo.GetCurrentJobFamily(ctx, tx, tenantID, normalizedFamily)
		if err != nil {
			return nil, err
		}
		if family == nil {
			return nil, ErrJobCatalogParentMissing
		}
		familyCode = family.Code
		parentRecord = family.RecordID
	}

	updated, err := s.repo.UpdateJobRole(ctx, tx, tenantID, normalizedCode, current.RecordID, familyCode, parentRecord, req)
	if err != nil {
		return nil, s.mapUpdateError(err)
	}

	after := map[string]interface{}{
		"code":        updated.Code,
		"familyCode":  updated.FamilyCode,
		"effectiveAt": updated.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeUpdate, "UpdateJobRole", updated.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *JobCatalogService) UpdateJobLevel(ctx context.Context, tenantID uuid.UUID, code string, req *types.UpdateJobLevelRequest, ifMatch *string, operator types.OperatedByInfo) (*types.JobLevel, error) {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	current, err := s.repo.GetCurrentJobLevel(ctx, tx, tenantID, normalizedCode)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrJobCatalogNotFound
	}

	if ifMatch != nil && current.RecordID.String() != strings.TrimSpace(*ifMatch) {
		return nil, ErrJobCatalogPreconditionFailed
	}

	roleCode := current.RoleCode
	parentRecord := current.ParentRecord
	if req.JobRoleCode != nil {
		normalizedRole := strings.ToUpper(strings.TrimSpace(*req.JobRoleCode))
		if normalizedRole == "" {
			return nil, ErrJobCatalogInvalidInput
		}
		role, err := s.repo.GetCurrentJobRole(ctx, tx, tenantID, normalizedRole)
		if err != nil {
			return nil, err
		}
		if role == nil {
			return nil, ErrJobCatalogParentMissing
		}
		roleCode = role.Code
		parentRecord = role.RecordID
	}

	levelRank := current.LevelRank
	if req.LevelRank != nil {
		if *req.LevelRank < 1 {
			return nil, ErrJobCatalogInvalidInput
		}
		levelRank = strconv.Itoa(*req.LevelRank)
	}

	updated, err := s.repo.UpdateJobLevel(ctx, tx, tenantID, normalizedCode, current.RecordID, roleCode, parentRecord, levelRank, req)
	if err != nil {
		return nil, s.mapUpdateError(err)
	}

	after := map[string]interface{}{
		"code":        updated.Code,
		"roleCode":    updated.RoleCode,
		"effectiveAt": updated.EffectiveDate.Format("2006-01-02"),
	}
	if err := s.logCatalogEvent(ctx, tx, tenantID, operator, audit.EventTypeUpdate, "UpdateJobLevel", updated.RecordID, after); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *JobCatalogService) mapUpdateError(err error) error {
	if err == nil {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return ErrJobCatalogConflict
		case "23503":
			return ErrJobCatalogParentMissing
		}
	}
	return err
}

func (s *JobCatalogService) logCatalogEvent(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, operator types.OperatedByInfo, eventType, action string, recordID uuid.UUID, after map[string]interface{}) error {
	if s.auditLogger == nil {
		return nil
	}

	actorID := strings.TrimSpace(operator.ID)
	actorName := strings.TrimSpace(operator.Name)
	actorType := audit.ActorTypeUser
	if actorID == "" {
		actorType = audit.ActorTypeSystem
		actorID = "system"
	}
	requestID := orgmiddleware.GetRequestID(ctx)
	correlationID := orgmiddleware.GetCorrelationID(ctx)
	sourceCorrelation := ""
	if src := orgmiddleware.GetCorrelationSource(ctx); src == "header" {
		sourceCorrelation = src
	}

	entityCode := ""
	if v, ok := after["code"].(string); ok {
		entityCode = strings.TrimSpace(v)
	}

	event := &audit.AuditEvent{
		TenantID:          tenantID,
		EventType:         eventType,
		ResourceType:      audit.ResourceTypeJobCatalog,
		ResourceID:        recordID.String(),
		RecordID:          recordID,
		EntityCode:        entityCode,
		ActorID:           actorID,
		ActorType:         actorType,
		ActorName:         actorName,
		ActionName:        action,
		RequestID:         requestID,
		CorrelationID:     correlationID,
		SourceCorrelation: sourceCorrelation,
		Success:           true,
		AfterData:         after,
		ContextPayload:    after,
	}

	if err := s.auditLogger.LogEventInTransaction(ctx, tx, event); err != nil {
		s.logger.Errorf("[AUDIT] failed to log job catalog event: %v", err)
		return err
	}

	return nil
}
