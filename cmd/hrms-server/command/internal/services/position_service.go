package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cube-castle/cmd/hrms-server/command/internal/audit"
	"cube-castle/cmd/hrms-server/command/internal/repository"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrPositionNotFound       = errors.New("position not found")
	ErrOrganizationNotFound   = errors.New("organization not found")
	ErrJobCatalogNotFound     = errors.New("job catalog reference not found")
	ErrJobCatalogMismatch     = errors.New("job catalog hierarchy mismatch")
	ErrPositionVersionExists  = errors.New("position version already exists for effective date")
	ErrPositionTimelineUpdate = errors.New("failed to update position timeline")
	ErrVersionConflict        = errors.New("version conflict")
	ErrInvalidTransition      = errors.New("invalid position transition")
	ErrInvalidHeadcount       = errors.New("invalid headcount change")
	ErrAssignmentNotFound     = errors.New("position assignment not found")
	ErrInvalidAssignmentState = errors.New("position assignment state invalid")
)

type PositionService struct {
	positions   *repository.PositionRepository
	assignments *repository.PositionAssignmentRepository
	jobCatalog  *repository.JobCatalogRepository
	orgRepo     *repository.OrganizationRepository
	auditLogger *audit.AuditLogger
	logger      pkglogger.Logger
}

func NewPositionService(positions *repository.PositionRepository, assignments *repository.PositionAssignmentRepository, jobCatalog *repository.JobCatalogRepository, orgRepo *repository.OrganizationRepository, auditLogger *audit.AuditLogger, baseLogger pkglogger.Logger) *PositionService {
	return &PositionService{
		positions:   positions,
		assignments: assignments,
		jobCatalog:  jobCatalog,
		orgRepo:     orgRepo,
		auditLogger: auditLogger,
		logger:      scopedLogger(baseLogger, "position", nil),
	}
}

type jobCatalogSnapshot struct {
	group  *types.JobFamilyGroup
	family *types.JobFamily
	role   *types.JobRole
	level  *types.JobLevel
}

func (s *PositionService) CreatePosition(ctx context.Context, tenantID uuid.UUID, req *types.PositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	tx, err := s.positions.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	org, err := s.orgRepo.GetByCode(ctx, tenantID, req.OrganizationCode)
	if err != nil {
		if strings.Contains(err.Error(), "组织不存在") {
			return nil, ErrOrganizationNotFound
		}
		return nil, err
	}

	catalog, err := s.resolveJobCatalog(ctx, tx, tenantID, req.JobFamilyGroupCode, req.JobFamilyGroupRecordID, req.JobFamilyCode, req.JobFamilyRecordID, req.JobRoleCode, req.JobRoleRecordID, req.JobLevelCode, req.JobLevelRecordID)
	if err != nil {
		return nil, err
	}

	positionCode, err := s.positions.GenerateCode(ctx, tx, tenantID)
	if err != nil {
		return nil, err
	}

	entity, err := s.buildPositionEntity(tenantID, positionCode, req, catalog, org, operator, true)
	if err != nil {
		return nil, err
	}

	entity, err = s.positions.InsertPositionVersion(ctx, tx, entity)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, ErrPositionVersionExists
		}
		return nil, err
	}

	if err := s.positions.RecalculatePositionTimeline(ctx, tx, tenantID, entity.Code); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPositionTimelineUpdate, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	response := s.toPositionResponse(entity, nil)
	s.logPositionEvent(ctx, operator, tenantID, audit.EventTypeCreate, "CreatePosition", entity.RecordID, map[string]interface{}{
		"code":               entity.Code,
		"title":              entity.Title,
		"organizationCode":   entity.OrganizationCode,
		"jobFamilyGroupCode": entity.JobFamilyGroupCode,
		"jobFamilyCode":      entity.JobFamilyCode,
		"jobRoleCode":        entity.JobRoleCode,
		"jobLevelCode":       entity.JobLevelCode,
	})

	return response, nil
}

func (s *PositionService) ReplacePosition(ctx context.Context, tenantID uuid.UUID, code string, ifMatch *string, req *types.PositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	tx, err := s.positions.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	current, err := s.positions.GetCurrentPosition(ctx, tx, tenantID, code)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrPositionNotFound
	}

	if ifMatch != nil && *ifMatch != "" && current.RecordID.String() != strings.Trim(*ifMatch, "\"") {
		return nil, ErrVersionConflict
	}

	org, err := s.orgRepo.GetByCode(ctx, tenantID, req.OrganizationCode)
	if err != nil {
		if strings.Contains(err.Error(), "组织不存在") {
			return nil, ErrOrganizationNotFound
		}
		return nil, err
	}

	catalog, err := s.resolveJobCatalog(ctx, tx, tenantID, req.JobFamilyGroupCode, req.JobFamilyGroupRecordID, req.JobFamilyCode, req.JobFamilyRecordID, req.JobRoleCode, req.JobRoleRecordID, req.JobLevelCode, req.JobLevelRecordID)
	if err != nil {
		return nil, err
	}

	updateEntity, err := s.buildPositionEntity(tenantID, current.Code, req, catalog, org, operator, false)
	if err != nil {
		return nil, err
	}
	updateEntity.RecordID = current.RecordID
	updateEntity.HeadcountInUse = current.HeadcountInUse
	updateEntity.IsCurrent = current.IsCurrent
	updateEntity.CreatedAt = current.CreatedAt
	updateEntity.OperationType = "UPDATE"

	if _, err := s.positions.UpdatePositionDetails(ctx, tx, updateEntity); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	response := s.toPositionResponse(updateEntity, nil)
	s.logPositionEvent(ctx, operator, tenantID, audit.EventTypeUpdate, "UpdatePosition", updateEntity.RecordID, map[string]interface{}{
		"code":             updateEntity.Code,
		"title":            updateEntity.Title,
		"organizationCode": updateEntity.OrganizationCode,
	})

	return response, nil
}

func (s *PositionService) CreatePositionVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionVersionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	tx, err := s.positions.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	current, err := s.positions.GetCurrentPosition(ctx, tx, tenantID, code)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrPositionNotFound
	}

	org, err := s.orgRepo.GetByCode(ctx, tenantID, current.OrganizationCode)
	if err != nil {
		return nil, err
	}

	catalog, err := s.resolveJobCatalog(ctx, tx, tenantID,
		req.JobFamilyGroupCode, req.JobFamilyGroupRecordID,
		req.JobFamilyCode, req.JobFamilyRecordID,
		req.JobRoleCode, req.JobRoleRecordID,
		req.JobLevelCode, req.JobLevelRecordID,
	)
	if err != nil {
		return nil, err
	}

	jobProfileCode := req.JobProfileCode
	if jobProfileCode == nil && current.JobProfileCode.Valid {
		val := current.JobProfileCode.String
		jobProfileCode = &val
	}
	jobProfileName := req.JobProfileName
	if jobProfileName == nil && current.JobProfileName.Valid {
		val := current.JobProfileName.String
		jobProfileName = &val
	}
	gradeLevel := req.GradeLevel
	if gradeLevel == nil && current.GradeLevel.Valid {
		val := current.GradeLevel.String
		gradeLevel = &val
	}
	costCenter := req.CostCenterCode
	if costCenter == nil && current.CostCenterCode.Valid {
		val := current.CostCenterCode.String
		costCenter = &val
	}
	reportsTo := req.ReportsTo
	if reportsTo == nil && current.ReportsToPosition.Valid {
		val := current.ReportsToPosition.String
		reportsTo = &val
	}
	var headcountInUsePtr *float64
	if req.HeadcountInUse != nil {
		headcountInUsePtr = req.HeadcountInUse
	} else {
		existing := current.HeadcountInUse
		headcountInUsePtr = &existing
	}

	versionSource := &types.PositionRequest{
		Title:                  req.Title,
		JobProfileCode:         jobProfileCode,
		JobProfileName:         jobProfileName,
		JobFamilyGroupCode:     req.JobFamilyGroupCode,
		JobFamilyGroupRecordID: req.JobFamilyGroupRecordID,
		JobFamilyCode:          req.JobFamilyCode,
		JobFamilyRecordID:      req.JobFamilyRecordID,
		JobRoleCode:            req.JobRoleCode,
		JobRoleRecordID:        req.JobRoleRecordID,
		JobLevelCode:           req.JobLevelCode,
		JobLevelRecordID:       req.JobLevelRecordID,
		OrganizationCode:       current.OrganizationCode,
		PositionType:           valueOrDefault(req.PositionType, current.PositionType),
		EmploymentType:         valueOrDefault(req.EmploymentType, current.EmploymentType),
		HeadcountCapacity:      valueOrDefaultFloat(req.HeadcountCapacity, current.HeadcountCapacity),
		HeadcountInUse:         headcountInUsePtr,
		GradeLevel:             gradeLevel,
		CostCenterCode:         costCenter,
		ReportsToPositionCode:  reportsTo,
		Profile:                req.Profile,
		EffectiveDate:          req.EffectiveDate,
		OperationReason:        req.OperationReason,
	}

	entity, err := s.buildPositionEntity(tenantID, current.Code, versionSource, catalog, org, operator, false)
	if err != nil {
		return nil, err
	}
	entity.OperationType = "CREATE_VERSION"
	// 新增版本在插入前统一设置为非当前版本，待时间线重算后再确定 current 标记，避免违反唯一约束
	entity.IsCurrent = false

	if _, err := s.positions.InsertPositionVersion(ctx, tx, entity); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, ErrPositionVersionExists
		}
		return nil, err
	}

	if err := s.positions.RecalculatePositionTimeline(ctx, tx, tenantID, current.Code); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPositionTimelineUpdate, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	response := s.toPositionResponse(entity, nil)
	s.logPositionEvent(ctx, operator, tenantID, audit.EventTypeCreate, "CreatePositionVersion", entity.RecordID, map[string]interface{}{
		"code":          entity.Code,
		"effectiveDate": entity.EffectiveDate.Format("2006-01-02"),
	})

	return response, nil
}

// TODO-TEMPORARY: 临时占位以支撑 Stage1 填充流程，待 assignments 模块落地后改由专用服务处理（Owner: 命令服务组，Deadline: 2025-11-15，Plan: 接入统一 assignments API 并移除本地实现）
func (s *PositionService) FillPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.FillPositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	updated, assignments, assignment, err := s.createAssignment(ctx, tenantID, code, req, operator)
	if err != nil {
		return nil, err
	}

	response := s.toPositionResponse(updated, assignments)
	s.logPositionEvent(ctx, operator, tenantID, audit.EventTypeUpdate, "FillPosition", updated.RecordID, map[string]interface{}{
		"code":                updated.Code,
		"assignmentId":        assignment.AssignmentID.String(),
		"assignmentStatus":    assignment.AssignmentStatus,
		"headcountInUse":      updated.HeadcountInUse,
		"positionStatus":      updated.Status,
		"assignmentEffective": assignment.EffectiveDate.Format("2006-01-02"),
	})
	return response, nil
}

func (s *PositionService) createAssignment(ctx context.Context, tenantID uuid.UUID, code string, req *types.FillPositionRequest, operator types.OperatedByInfo) (*types.Position, []types.PositionAssignment, *types.PositionAssignment, error) {
	tx, err := s.positions.BeginTx(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	defer tx.Rollback()

	current, err := s.positions.GetCurrentPosition(ctx, tx, tenantID, code)
	if err != nil {
		return nil, nil, nil, err
	}
	if current == nil {
		return nil, nil, nil, ErrPositionNotFound
	}

	fte := 1.0
	if req.FTE != nil {
		fte = *req.FTE
	}
	if fte <= 0 {
		return nil, nil, nil, ErrInvalidHeadcount
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid effectiveDate: %w", err)
	}

	var anticipatedEnd sql.NullTime
	if req.AnticipatedEndDate != nil && strings.TrimSpace(*req.AnticipatedEndDate) != "" {
		endTime, parseErr := time.Parse("2006-01-02", strings.TrimSpace(*req.AnticipatedEndDate))
		if parseErr != nil {
			return nil, nil, nil, fmt.Errorf("invalid anticipatedEndDate: %w", parseErr)
		}
		if endTime.Before(effectiveDate) {
			return nil, nil, nil, fmt.Errorf("anticipatedEndDate must be on or after effectiveDate")
		}
		anticipatedEnd = sql.NullTime{Time: endTime, Valid: true}
	}

	employeeID, err := uuid.Parse(strings.TrimSpace(req.EmployeeID))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("employeeId must be UUID: %w", err)
	}

	employeeName := strings.TrimSpace(req.EmployeeName)
	if employeeName == "" {
		return nil, nil, nil, fmt.Errorf("employeeName is required")
	}

	assignmentType := strings.ToUpper(strings.TrimSpace(req.AssignmentType))
	if assignmentType == "" {
		assignmentType = "PRIMARY"
	}
	if _, ok := map[string]struct{}{"PRIMARY": {}, "SECONDARY": {}, "ACTING": {}}[assignmentType]; !ok {
		return nil, nil, nil, fmt.Errorf("unsupported assignmentType: %s", assignmentType)
	}

	actingUntil := sql.NullTime{}
	if strings.EqualFold(assignmentType, "ACTING") {
		actingUntil = anticipatedEnd
		anticipatedEnd = sql.NullTime{}
	}

	autoRevert := false
	if req.AutoRevert != nil {
		autoRevert = *req.AutoRevert
	}
	if autoRevert && !strings.EqualFold(assignmentType, "ACTING") {
		return nil, nil, nil, fmt.Errorf("autoRevert only supported for ACTING assignments")
	}
	if autoRevert && !actingUntil.Valid {
		return nil, nil, nil, fmt.Errorf("actingUntil is required when autoRevert is enabled")
	}

	activeFTE, err := s.assignments.SumActiveFTE(ctx, tx, tenantID, current.Code)
	if err != nil {
		return nil, nil, nil, err
	}

	projected := activeFTE
	statusNow := strings.ToUpper(current.Status)

	now := time.Now().UTC().Truncate(24 * time.Hour)
	assignmentStatus := "ACTIVE"
	isCurrent := true
	if effectiveDate.After(now) {
		assignmentStatus = "PENDING"
		isCurrent = false
	}
	if projected+fte > current.HeadcountCapacity {
		return nil, nil, nil, ErrInvalidHeadcount
	}

	var employeeNumber sql.NullString
	if req.EmployeeNumber != nil {
		num := strings.TrimSpace(*req.EmployeeNumber)
		if num != "" {
			employeeNumber = sql.NullString{String: num, Valid: true}
		}
	}

	var notes sql.NullString
	if req.Notes != nil {
		text := strings.TrimSpace(*req.Notes)
		if text != "" {
			notes = sql.NullString{String: text, Valid: true}
		}
	}

	assignment := &types.PositionAssignment{
		TenantID:         tenantID,
		PositionCode:     current.Code,
		PositionRecordID: current.RecordID,
		EmployeeID:       employeeID,
		EmployeeName:     employeeName,
		EmployeeNumber:   employeeNumber,
		AssignmentType:   assignmentType,
		AssignmentStatus: assignmentStatus,
		FTE:              fte,
		EffectiveDate:    effectiveDate,
		EndDate:          anticipatedEnd,
		ActingUntil:      actingUntil,
		AutoRevert:       autoRevert,
		IsCurrent:        isCurrent,
		Notes:            notes,
	}

	if assignment, err = s.assignments.CreateAssignment(ctx, tx, assignment); err != nil {
		return nil, nil, nil, err
	}

	activeFTE, err = s.assignments.SumActiveFTE(ctx, tx, tenantID, current.Code)
	if err != nil {
		return nil, nil, nil, err
	}

	switch {
	case activeFTE >= current.HeadcountCapacity:
		statusNow = "FILLED"
	case activeFTE > 0:
		statusNow = "PARTIALLY_FILLED"
	default:
		statusNow = "VACANT"
	}

	opID, opName := resolveOperator(operator)
	if err := s.positions.UpdatePositionHeadcount(ctx, tx, tenantID, current.RecordID, activeFTE, statusNow, "FILL", opName, opID, stringPointer(req.OperationReason)); err != nil {
		return nil, nil, nil, err
	}

	updated, err := s.positions.GetPositionByRecordID(ctx, tx, tenantID, current.RecordID)
	if err != nil {
		return nil, nil, nil, err
	}
	if updated == nil {
		return nil, nil, nil, ErrPositionNotFound
	}

	assignments, err := s.assignments.ListByPosition(ctx, tx, tenantID, updated.Code)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, nil, err
	}

	return updated, assignments, assignment, nil
}

func (s *PositionService) CreateAssignmentRecord(ctx context.Context, tenantID uuid.UUID, code string, req *types.CreateAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error) {
	fillReq := &types.FillPositionRequest{
		EmployeeID:         req.EmployeeID,
		EmployeeName:       req.EmployeeName,
		EmployeeNumber:     req.EmployeeNumber,
		AssignmentType:     req.AssignmentType,
		FTE:                req.FTE,
		EffectiveDate:      req.EffectiveDate,
		AnticipatedEndDate: req.ActingUntil,
		AutoRevert:         req.AutoRevert,
		OperationReason:    req.OperationReason,
		Notes:              req.Notes,
	}

	updated, _, assignment, err := s.createAssignment(ctx, tenantID, code, fillReq, operator)
	if err != nil {
		return nil, err
	}

	resp := toAssignmentResponse(*assignment)
	s.logPositionEvent(ctx, operator, tenantID, audit.EventTypeUpdate, "CreateAssignment", updated.RecordID, map[string]interface{}{
		"code":             updated.Code,
		"assignmentId":     assignment.AssignmentID.String(),
		"assignmentType":   assignment.AssignmentType,
		"autoRevert":       assignment.AutoRevert,
		"assignmentStatus": assignment.AssignmentStatus,
	})

	return &resp, nil
}

func (s *PositionService) ListAssignments(ctx context.Context, tenantID uuid.UUID, code string, opts types.AssignmentListOptions) ([]types.PositionAssignmentResponse, int, error) {
	tx, err := s.positions.BeginTx(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	current, err := s.positions.GetCurrentPosition(ctx, tx, tenantID, code)
	if err != nil {
		return nil, 0, err
	}
	if current == nil {
		return nil, 0, ErrPositionNotFound
	}

	assignments, total, err := s.assignments.ListWithOptions(ctx, tx, tenantID, current.Code, opts)
	if err != nil {
		return nil, 0, err
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}

	responses := make([]types.PositionAssignmentResponse, 0, len(assignments))
	for _, item := range assignments {
		converted := toAssignmentResponse(item)
		responses = append(responses, converted)
	}

	return responses, total, nil
}

func (s *PositionService) UpdateAssignmentRecord(ctx context.Context, tenantID uuid.UUID, code string, assignmentID uuid.UUID, req *types.UpdateAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error) {
	tx, err := s.positions.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	current, err := s.positions.GetCurrentPosition(ctx, tx, tenantID, code)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrPositionNotFound
	}

	assignment, err := s.assignments.GetByID(ctx, tx, tenantID, assignmentID)
	if err != nil {
		return nil, err
	}
	if assignment == nil || assignment.PositionCode != current.Code {
		return nil, ErrAssignmentNotFound
	}
	if strings.EqualFold(assignment.AssignmentStatus, "ENDED") {
		return nil, ErrInvalidAssignmentState
	}

	updateParams := types.AssignmentUpdateParams{}

	newFTE := assignment.FTE
	if req.FTE != nil {
		if *req.FTE <= 0 {
			return nil, ErrInvalidHeadcount
		}
		updateParams.FTE = req.FTE
		newFTE = *req.FTE
	}

	if req.ActingUntil != nil {
		trimmed := strings.TrimSpace(*req.ActingUntil)
		if trimmed == "" {
			updateParams.ClearActingUntil = true
		} else {
			parsed, parseErr := time.Parse("2006-01-02", trimmed)
			if parseErr != nil {
				return nil, fmt.Errorf("invalid actingUntil: %w", parseErr)
			}
			updateParams.ActingUntil = &parsed
		}
	}

	if req.AutoRevert != nil {
		updateParams.AutoRevert = req.AutoRevert
	}

	if req.Notes != nil {
		updateParams.Notes = req.Notes
	}

	newAutoRevert := assignment.AutoRevert
	if updateParams.AutoRevert != nil {
		newAutoRevert = *updateParams.AutoRevert
	}

	newActingUntil := assignment.ActingUntil
	if updateParams.ClearActingUntil {
		newActingUntil = sql.NullTime{}
	}
	if updateParams.ActingUntil != nil {
		newActingUntil = sql.NullTime{Time: *updateParams.ActingUntil, Valid: true}
	}

	if newAutoRevert && !strings.EqualFold(assignment.AssignmentType, "ACTING") {
		return nil, fmt.Errorf("autoRevert only supported for ACTING assignments")
	}
	if newAutoRevert && !newActingUntil.Valid {
		return nil, fmt.Errorf("actingUntil is required when autoRevert is enabled")
	}

	if req.FTE != nil && strings.EqualFold(assignment.AssignmentStatus, "ACTIVE") {
		activeFTE, err := s.assignments.SumActiveFTE(ctx, tx, tenantID, current.Code)
		if err != nil {
			return nil, err
		}
		projected := activeFTE - assignment.FTE + newFTE
		if projected > current.HeadcountCapacity {
			return nil, ErrInvalidHeadcount
		}
	}

	if err := s.assignments.UpdateAssignment(ctx, tx, tenantID, assignmentID, updateParams); err != nil {
		return nil, err
	}

	updatedAssignment, err := s.assignments.GetByID(ctx, tx, tenantID, assignmentID)
	if err != nil {
		return nil, err
	}
	if updatedAssignment == nil {
		return nil, ErrAssignmentNotFound
	}

	if req.FTE != nil && strings.EqualFold(updatedAssignment.AssignmentStatus, "ACTIVE") {
		activeFTE, err := s.assignments.SumActiveFTE(ctx, tx, tenantID, current.Code)
		if err != nil {
			return nil, err
		}
		statusNow := strings.ToUpper(current.Status)
		switch {
		case activeFTE >= current.HeadcountCapacity:
			statusNow = "FILLED"
		case activeFTE > 0:
			statusNow = "PARTIALLY_FILLED"
		default:
			statusNow = "VACANT"
		}

		opID, opName := resolveOperator(operator)
		if err := s.positions.UpdatePositionHeadcount(ctx, tx, tenantID, current.RecordID, activeFTE, statusNow, "ASSIGNMENT_UPDATE", opName, opID, stringPointer(req.OperationReason)); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	resp := toAssignmentResponse(*updatedAssignment)
	s.logPositionEvent(ctx, operator, tenantID, audit.EventTypeUpdate, "UpdateAssignment", updatedAssignment.PositionRecordID, map[string]interface{}{
		"assignmentId":   updatedAssignment.AssignmentID.String(),
		"assignmentType": updatedAssignment.AssignmentType,
		"autoRevert":     updatedAssignment.AutoRevert,
	})

	return &resp, nil
}

func (s *PositionService) CloseAssignmentRecord(ctx context.Context, tenantID uuid.UUID, code string, assignmentID uuid.UUID, req *types.CloseAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error) {
	vacateReq := &types.VacatePositionRequest{
		AssignmentID:    assignmentID.String(),
		EffectiveDate:   req.EndDate,
		OperationReason: req.OperationReason,
		Notes:           req.Notes,
	}

	positionResp, err := s.VacatePosition(ctx, tenantID, code, vacateReq, operator)
	if err != nil {
		return nil, err
	}

	for _, item := range positionResp.AssignmentHistory {
		if item.AssignmentID == assignmentID {
			found := item
			return &found, nil
		}
	}

	return nil, ErrAssignmentNotFound
}

func (s *PositionService) ProcessAutoReverts(ctx context.Context, tenantID uuid.UUID, asOf time.Time, limit int, operator types.OperatedByInfo) ([]types.PositionAssignmentResponse, error) {
	candidates, err := s.assignments.ListAutoRevertCandidates(ctx, nil, tenantID, asOf, limit)
	if err != nil {
		return nil, err
	}

	results := make([]types.PositionAssignmentResponse, 0, len(candidates))
	for _, candidate := range candidates {
		if !candidate.ActingUntil.Valid {
			continue
		}

		vacReq := &types.VacatePositionRequest{
			AssignmentID:    candidate.AssignmentID.String(),
			EffectiveDate:   candidate.ActingUntil.Time.Format("2006-01-02"),
			OperationReason: "AUTO_REVERT_ACTING_ASSIGNMENT",
		}

		resp, err := s.VacatePosition(ctx, tenantID, candidate.PositionCode, vacReq, operator)
		if err != nil {
			s.logger.Errorf("[AUTO-REVERT] failed to close assignment %s: %v", candidate.AssignmentID, err)
			continue
		}

		for _, item := range resp.AssignmentHistory {
			if item.AssignmentID == candidate.AssignmentID {
				results = append(results, item)
				break
			}
		}
	}

	return results, nil
}

func (s *PositionService) VacatePosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.VacatePositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	tx, err := s.positions.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	current, err := s.positions.GetCurrentPosition(ctx, tx, tenantID, code)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrPositionNotFound
	}

	assignmentID, err := uuid.Parse(strings.TrimSpace(req.AssignmentID))
	if err != nil {
		return nil, fmt.Errorf("assignmentId must be UUID: %w", err)
	}

	assignment, err := s.assignments.GetByID(ctx, tx, tenantID, assignmentID)
	if err != nil {
		return nil, err
	}
	if assignment == nil || assignment.PositionCode != current.Code {
		return nil, ErrAssignmentNotFound
	}

	if strings.EqualFold(assignment.AssignmentStatus, "ENDED") {
		return nil, ErrInvalidAssignmentState
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effectiveDate: %w", err)
	}
	if effectiveDate.Before(assignment.EffectiveDate) {
		return nil, fmt.Errorf("effectiveDate cannot be earlier than assignment effective date")
	}

	var notes *string
	if req.Notes != nil {
		text := strings.TrimSpace(*req.Notes)
		if text != "" {
			notes = &text
		}
	}

	if err := s.assignments.CloseAssignment(ctx, tx, tenantID, assignmentID, effectiveDate, notes); err != nil {
		return nil, err
	}

	activeFTE, err := s.assignments.SumActiveFTE(ctx, tx, tenantID, current.Code)
	if err != nil {
		return nil, err
	}

	status := strings.ToUpper(current.Status)
	switch {
	case activeFTE >= current.HeadcountCapacity:
		status = "FILLED"
	case activeFTE > 0:
		status = "PARTIALLY_FILLED"
	default:
		status = "VACANT"
	}

	opID, opName := resolveOperator(operator)
	if err := s.positions.UpdatePositionHeadcount(ctx, tx, tenantID, current.RecordID, activeFTE, status, "VACATE", opName, opID, stringPointer(req.OperationReason)); err != nil {
		return nil, err
	}

	updated, err := s.positions.GetPositionByRecordID(ctx, tx, tenantID, current.RecordID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrPositionNotFound
	}

	assignments, err := s.assignments.ListByPosition(ctx, tx, tenantID, updated.Code)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	response := s.toPositionResponse(updated, assignments)
	s.logPositionEvent(ctx, operator, tenantID, audit.EventTypeUpdate, "VacatePosition", updated.RecordID, map[string]interface{}{
		"code":               updated.Code,
		"assignmentId":       assignment.AssignmentID.String(),
		"headcountInUse":     updated.HeadcountInUse,
		"positionStatus":     updated.Status,
		"assignmentEndDate":  effectiveDate.Format("2006-01-02"),
		"assignmentPrevious": assignment.AssignmentStatus,
	})
	return response, nil
}

// TODO-TEMPORARY: 临时 Transfer 实现，仅在 Stage1 过渡使用（Owner: 命令服务组，Deadline: 2025-11-15，Plan: 引入统一岗位调动服务并删除此实现）
func (s *PositionService) TransferPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.TransferPositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	tx, err := s.positions.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	current, err := s.positions.GetCurrentPosition(ctx, tx, tenantID, code)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, ErrPositionNotFound
	}

	targetOrg, err := s.orgRepo.GetByCode(ctx, tenantID, req.TargetOrganizationCode)
	if err != nil {
		if strings.Contains(err.Error(), "组织不存在") {
			return nil, ErrOrganizationNotFound
		}
		return nil, err
	}

	opID, opName := resolveOperator(operator)
	if err := s.positions.UpdatePositionOrganization(ctx, tx, tenantID, current.RecordID, targetOrg.Code, &targetOrg.Name, current.Status, "TRANSFER", opID, opName, stringPointer(req.OperationReason)); err != nil {
		return nil, err
	}

	updated, err := s.positions.GetPositionByRecordID(ctx, tx, tenantID, current.RecordID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrPositionNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	assignments, err := s.assignments.ListByPosition(ctx, tx, tenantID, updated.Code)
	if err != nil {
		return nil, err
	}

	response := s.toPositionResponse(updated, assignments)
	s.logPositionEvent(ctx, operator, tenantID, audit.EventTypeUpdate, "TransferPosition", updated.RecordID, map[string]interface{}{
		"code":             updated.Code,
		"organizationCode": updated.OrganizationCode,
	})
	return response, nil
}

func (s *PositionService) ApplyEvent(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionEventRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	tx, err := s.positions.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var target *types.Position
	if req.RecordID != nil && strings.TrimSpace(*req.RecordID) != "" {
		recordUUID, parseErr := uuid.Parse(strings.TrimSpace(*req.RecordID))
		if parseErr != nil {
			return nil, fmt.Errorf("recordId must be UUID: %w", parseErr)
		}
		target, err = s.positions.GetPositionByRecordID(ctx, tx, tenantID, recordUUID)
		if err != nil {
			return nil, err
		}
		if target == nil {
			return nil, ErrPositionNotFound
		}
	} else {
		target, err = s.positions.GetCurrentPosition(ctx, tx, tenantID, code)
		if err != nil {
			return nil, err
		}
		if target == nil {
			return nil, ErrPositionNotFound
		}
	}

	eventType := strings.ToUpper(strings.TrimSpace(req.EventType))
	opID, opName := resolveOperator(operator)

	switch eventType {
	case "SUSPEND", "INACTIVE":
		payload := map[string]interface{}{
			"event":         eventType,
			"effectiveDate": req.EffectiveDate,
		}
		if err := s.positions.UpdatePositionStatus(ctx, tx, tenantID, target.RecordID, "INACTIVE", payload, "SUSPEND", opName, opID, stringPointer(req.OperationReason)); err != nil {
			return nil, err
		}
	case "REACTIVATE", "ACTIVATE":
		payload := map[string]interface{}{
			"event":         eventType,
			"effectiveDate": req.EffectiveDate,
		}
		if err := s.positions.UpdatePositionStatus(ctx, tx, tenantID, target.RecordID, "ACTIVE", payload, "REACTIVATE", opName, opID, stringPointer(req.OperationReason)); err != nil {
			return nil, err
		}
	case "DELETE":
		if err := s.positions.DeletePositionVersion(ctx, tx, tenantID, target.RecordID, opID, opName, stringPointer(req.OperationReason)); err != nil {
			return nil, err
		}
		if err := s.positions.RecalculatePositionTimeline(ctx, tx, tenantID, target.Code); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPositionTimelineUpdate, err)
		}
	default:
		return nil, ErrInvalidTransition
	}

	updated, err := s.positions.GetCurrentPosition(ctx, tx, tenantID, target.Code)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		updated = target
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	response := s.toPositionResponse(updated, nil)
	s.logPositionEvent(ctx, operator, tenantID, audit.EventTypeUpdate, "ApplyPositionEvent", updated.RecordID, map[string]interface{}{
		"code":      updated.Code,
		"eventType": eventType,
		"status":    updated.Status,
	})

	return response, nil
}

func (s *PositionService) resolveJobCatalog(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, groupCode string, groupRecord *string, familyCode string, familyRecord *string, roleCode string, roleRecord *string, levelCode string, levelRecord *string) (*jobCatalogSnapshot, error) {
	group, err := s.lookupFamilyGroup(ctx, tx, tenantID, groupCode, groupRecord)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrJobCatalogNotFound
	}

	family, err := s.lookupJobFamily(ctx, tx, tenantID, familyCode, familyRecord)
	if err != nil {
		return nil, err
	}
	if family == nil {
		return nil, ErrJobCatalogNotFound
	}
	if family.FamilyGroupCode != group.Code || family.ParentRecord != group.RecordID {
		return nil, ErrJobCatalogMismatch
	}

	role, err := s.lookupJobRole(ctx, tx, tenantID, roleCode, roleRecord)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrJobCatalogNotFound
	}
	if role.FamilyCode != family.Code || role.ParentRecord != family.RecordID {
		return nil, ErrJobCatalogMismatch
	}

	level, err := s.lookupJobLevel(ctx, tx, tenantID, levelCode, levelRecord)
	if err != nil {
		return nil, err
	}
	if level == nil {
		return nil, ErrJobCatalogNotFound
	}
	if level.RoleCode != role.Code || level.ParentRecord != role.RecordID {
		return nil, ErrJobCatalogMismatch
	}

	return &jobCatalogSnapshot{
		group:  group,
		family: family,
		role:   role,
		level:  level,
	}, nil
}

func (s *PositionService) lookupFamilyGroup(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, recordID *string) (*types.JobFamilyGroup, error) {
	if recordID != nil && *recordID != "" {
		id, err := uuid.Parse(*recordID)
		if err != nil {
			return nil, fmt.Errorf("invalid jobFamilyGroupRecordId: %w", err)
		}
		group, err := s.jobCatalog.GetFamilyGroupByRecordID(ctx, tx, tenantID, id)
		if err != nil || group == nil {
			return group, err
		}
		if group.Code != code {
			return nil, ErrJobCatalogMismatch
		}
		return group, nil
	}
	return s.jobCatalog.GetCurrentFamilyGroup(ctx, tx, tenantID, code)
}

func (s *PositionService) lookupJobFamily(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, recordID *string) (*types.JobFamily, error) {
	if recordID != nil && *recordID != "" {
		id, err := uuid.Parse(*recordID)
		if err != nil {
			return nil, fmt.Errorf("invalid jobFamilyRecordId: %w", err)
		}
		family, err := s.jobCatalog.GetJobFamilyByRecordID(ctx, tx, tenantID, id)
		if err != nil || family == nil {
			return family, err
		}
		if family.Code != code {
			return nil, ErrJobCatalogMismatch
		}
		return family, nil
	}
	return s.jobCatalog.GetCurrentJobFamily(ctx, tx, tenantID, code)
}

func (s *PositionService) lookupJobRole(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, recordID *string) (*types.JobRole, error) {
	if recordID != nil && *recordID != "" {
		id, err := uuid.Parse(*recordID)
		if err != nil {
			return nil, fmt.Errorf("invalid jobRoleRecordId: %w", err)
		}
		role, err := s.jobCatalog.GetJobRoleByRecordID(ctx, tx, tenantID, id)
		if err != nil || role == nil {
			return role, err
		}
		if role.Code != code {
			return nil, ErrJobCatalogMismatch
		}
		return role, nil
	}
	return s.jobCatalog.GetCurrentJobRole(ctx, tx, tenantID, code)
}

func (s *PositionService) lookupJobLevel(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, recordID *string) (*types.JobLevel, error) {
	if recordID != nil && *recordID != "" {
		id, err := uuid.Parse(*recordID)
		if err != nil {
			return nil, fmt.Errorf("invalid jobLevelRecordId: %w", err)
		}
		level, err := s.jobCatalog.GetJobLevelByRecordID(ctx, tx, tenantID, id)
		if err != nil || level == nil {
			return level, err
		}
		if level.Code != code {
			return nil, ErrJobCatalogMismatch
		}
		return level, nil
	}
	return s.jobCatalog.GetCurrentJobLevel(ctx, tx, tenantID, code)
}

func (s *PositionService) buildPositionEntity(tenantID uuid.UUID, code string, req *types.PositionRequest, catalog *jobCatalogSnapshot, org *types.Organization, operator types.OperatedByInfo, isNew bool) (*types.Position, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effectiveDate: %w", err)
	}

	var profileBytes []byte
	if req.Profile != nil && strings.TrimSpace(*req.Profile) != "" {
		raw := strings.TrimSpace(*req.Profile)
		if !json.Valid([]byte(raw)) {
			return nil, fmt.Errorf("profile must be a valid JSON object")
		}
		profileBytes = []byte(raw)
	} else {
		profileBytes = []byte("{}")
	}

	opID, opName := resolveOperator(operator)
	status := "PLANNED"
	if req.Status != nil && strings.TrimSpace(*req.Status) != "" {
		status = strings.ToUpper(strings.TrimSpace(*req.Status))
	}

	headcountInUse := 0.0
	if req.HeadcountInUse != nil {
		headcountInUse = *req.HeadcountInUse
	}

	if req.HeadcountCapacity < 0 {
		return nil, ErrInvalidHeadcount
	}
	if headcountInUse < 0 || headcountInUse > req.HeadcountCapacity {
		return nil, ErrInvalidHeadcount
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	isCurrent := !effectiveDate.After(today)

	entity := &types.Position{
		TenantID:             tenantID,
		Code:                 code,
		Title:                strings.TrimSpace(req.Title),
		JobProfileCode:       toNullString(req.JobProfileCode),
		JobProfileName:       toNullString(req.JobProfileName),
		JobFamilyGroupCode:   catalog.group.Code,
		JobFamilyGroupName:   catalog.group.Name,
		JobFamilyGroupRecord: catalog.group.RecordID,
		JobFamilyCode:        catalog.family.Code,
		JobFamilyName:        catalog.family.Name,
		JobFamilyRecord:      catalog.family.RecordID,
		JobRoleCode:          catalog.role.Code,
		JobRoleName:          catalog.role.Name,
		JobRoleRecord:        catalog.role.RecordID,
		JobLevelCode:         catalog.level.Code,
		JobLevelName:         catalog.level.Name,
		JobLevelRecord:       catalog.level.RecordID,
		OrganizationCode:     org.Code,
		OrganizationName:     toNullString(stringPointer(org.Name)),
		PositionType:         strings.ToUpper(strings.TrimSpace(req.PositionType)),
		Status:               status,
		EmploymentType:       strings.ToUpper(strings.TrimSpace(req.EmploymentType)),
		HeadcountCapacity:    req.HeadcountCapacity,
		HeadcountInUse:       headcountInUse,
		GradeLevel:           toNullString(req.GradeLevel),
		CostCenterCode:       toNullString(req.CostCenterCode),
		ReportsToPosition:    toNullString(req.ReportsToPositionCode),
		Profile:              profileBytes,
		EffectiveDate:        effectiveDate,
		EndDate:              sql.NullTime{Valid: false},
		IsCurrent:            isCurrent,
		OperationType:        "CREATE",
		OperatedByID:         opID,
		OperatedByName:       opName,
		OperationReason:      toNullString(&req.OperationReason),
	}

	return entity, nil
}

func (s *PositionService) toPositionResponse(entity *types.Position, assignments []types.PositionAssignment) *types.PositionResponse {
	availableHeadcount := entity.HeadcountCapacity - entity.HeadcountInUse
	if availableHeadcount < 0 {
		availableHeadcount = 0
	}

	var organizationName *string
	if entity.OrganizationName.Valid {
		val := entity.OrganizationName.String
		organizationName = &val
	}

	var jobProfileCode *string
	if entity.JobProfileCode.Valid {
		val := entity.JobProfileCode.String
		jobProfileCode = &val
	}

	var jobProfileName *string
	if entity.JobProfileName.Valid {
		val := entity.JobProfileName.String
		jobProfileName = &val
	}

	var gradeLevel *string
	if entity.GradeLevel.Valid {
		val := entity.GradeLevel.String
		gradeLevel = &val
	}

	var costCenter *string
	if entity.CostCenterCode.Valid {
		val := entity.CostCenterCode.String
		costCenter = &val
	}

	var reportsTo *string
	if entity.ReportsToPosition.Valid {
		val := entity.ReportsToPosition.String
		reportsTo = &val
	}

	var endDate *time.Time
	if entity.EndDate.Valid {
		endDate = &entity.EndDate.Time
	}

	isFuture := entity.EffectiveDate.After(time.Now().UTC().Truncate(24 * time.Hour))

	response := &types.PositionResponse{
		Code:                  entity.Code,
		Title:                 entity.Title,
		JobProfileCode:        jobProfileCode,
		JobProfileName:        jobProfileName,
		JobFamilyGroupCode:    entity.JobFamilyGroupCode,
		JobFamilyGroupName:    entity.JobFamilyGroupName,
		JobFamilyCode:         entity.JobFamilyCode,
		JobFamilyName:         entity.JobFamilyName,
		JobRoleCode:           entity.JobRoleCode,
		JobRoleName:           entity.JobRoleName,
		JobLevelCode:          entity.JobLevelCode,
		JobLevelName:          entity.JobLevelName,
		OrganizationCode:      entity.OrganizationCode,
		OrganizationName:      organizationName,
		PositionType:          entity.PositionType,
		Status:                entity.Status,
		EmploymentType:        entity.EmploymentType,
		HeadcountCapacity:     entity.HeadcountCapacity,
		HeadcountInUse:        entity.HeadcountInUse,
		AvailableHeadcount:    availableHeadcount,
		GradeLevel:            gradeLevel,
		CostCenterCode:        costCenter,
		ReportsToPositionCode: reportsTo,
		EffectiveDate:         entity.EffectiveDate,
		EndDate:               endDate,
		IsCurrent:             entity.IsCurrent,
		IsFuture:              isFuture,
		RecordID:              entity.RecordID,
		CreatedAt:             entity.CreatedAt,
		UpdatedAt:             entity.UpdatedAt,
	}

	if len(assignments) > 0 {
		var history []types.PositionAssignmentResponse
		var current *types.PositionAssignmentResponse

		for _, assignment := range assignments {
			resp := toAssignmentResponse(assignment)
			history = append(history, resp)
			if assignment.IsCurrent && strings.EqualFold(assignment.AssignmentStatus, "ACTIVE") {
				assignmentCopy := resp
				current = &assignmentCopy
			}
		}

		response.AssignmentHistory = history
		response.CurrentAssignment = current
	}

	return response
}

func toAssignmentResponse(entity types.PositionAssignment) types.PositionAssignmentResponse {
	var employeeNumber *string
	if entity.EmployeeNumber.Valid {
		val := strings.TrimSpace(entity.EmployeeNumber.String)
		if val != "" {
			employeeNumber = &val
		}
	}

	var endDate *time.Time
	if entity.EndDate.Valid {
		endDate = &entity.EndDate.Time
	}

	var actingUntil *time.Time
	if entity.ActingUntil.Valid {
		actingUntil = &entity.ActingUntil.Time
	}

	var reminderSentAt *time.Time
	if entity.ReminderSentAt.Valid {
		reminderSentAt = &entity.ReminderSentAt.Time
	}

	var notes *string
	if entity.Notes.Valid {
		val := strings.TrimSpace(entity.Notes.String)
		if val != "" {
			notes = &val
		}
	}

	return types.PositionAssignmentResponse{
		AssignmentID:     entity.AssignmentID,
		PositionCode:     entity.PositionCode,
		PositionRecordID: entity.PositionRecordID,
		EmployeeID:       entity.EmployeeID,
		EmployeeName:     entity.EmployeeName,
		EmployeeNumber:   employeeNumber,
		AssignmentType:   entity.AssignmentType,
		AssignmentStatus: entity.AssignmentStatus,
		FTE:              entity.FTE,
		EffectiveDate:    entity.EffectiveDate,
		EndDate:          endDate,
		ActingUntil:      actingUntil,
		AutoRevert:       entity.AutoRevert,
		ReminderSentAt:   reminderSentAt,
		IsCurrent:        entity.IsCurrent,
		Notes:            notes,
		CreatedAt:        entity.CreatedAt,
		UpdatedAt:        entity.UpdatedAt,
	}
}

func (s *PositionService) logPositionEvent(ctx context.Context, operator types.OperatedByInfo, tenantID uuid.UUID, eventType, action string, recordID uuid.UUID, after map[string]interface{}) {
	if s.auditLogger == nil {
		return
	}
	opID, _ := resolveOperator(operator)
	actorID := strings.TrimSpace(operator.ID)
	actorType := audit.ActorTypeUser
	if actorID == "" {
		actorID = opID.String()
		actorType = audit.ActorTypeSystem
	}
	event := &audit.AuditEvent{
		TenantID:     tenantID,
		EventType:    eventType,
		ResourceType: audit.ResourceTypePosition,
		ResourceID:   recordID.String(),
		ActorID:      actorID,
		ActorType:    actorType,
		ActionName:   action,
		RequestID:    "",
		Timestamp:    time.Now(),
		Success:      true,
		AfterData:    after,
	}
	if err := s.auditLogger.LogEvent(ctx, event); err != nil {
		s.logger.Errorf("[AUDIT] failed to log position event: %v", err)
	}
}

func toNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: trimmed, Valid: true}
}

func stringPointer(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	v := value
	return &v
}

func resolveOperator(operator types.OperatedByInfo) (uuid.UUID, string) {
	if operator.ID != "" {
		if parsed, err := uuid.Parse(operator.ID); err == nil {
			name := operator.Name
			if strings.TrimSpace(name) == "" {
				name = "system"
			}
			return parsed, name
		}
	}
	return uuid.Nil, defaultOperatorName(operator.Name)
}

func defaultOperatorName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "system"
	}
	return strings.TrimSpace(name)
}

func valueOrDefault(val *string, fallback string) string {
	if val == nil || strings.TrimSpace(*val) == "" {
		return fallback
	}
	return strings.TrimSpace(*val)
}

func valueOrDefaultFloat(val *float64, fallback float64) float64 {
	if val == nil {
		return fallback
	}
	return *val
}
