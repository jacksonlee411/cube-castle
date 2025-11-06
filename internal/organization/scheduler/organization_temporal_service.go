package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/organization/audit"
	orgmiddleware "cube-castle/internal/organization/middleware"
	"cube-castle/internal/organization/repository"
	"cube-castle/internal/organization/utils"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

// OrganizationTemporalService 组织时态服务 - 按06文档要求实现
// 聚合 TemporalTimelineManager + AuditLogger，单事务维护时间轴与审计
type OrganizationTemporalService struct {
	db              *sql.DB
	timelineManager *repository.TemporalTimelineManager
	auditLogger     *audit.AuditLogger
	logger          pkglogger.Logger
	orgRepo         *repository.OrganizationRepository
}

func NewOrganizationTemporalService(db *sql.DB, baseLogger pkglogger.Logger) *OrganizationTemporalService {
	return &OrganizationTemporalService{
		db:              db,
		timelineManager: repository.NewTemporalTimelineManager(db, baseLogger),
		auditLogger:     audit.NewAuditLogger(db, baseLogger),
		logger:          scopedLogger(baseLogger, "organizationTemporal", nil),
		orgRepo:         repository.NewOrganizationRepository(db, baseLogger),
	}
}

// TemporalCreateVersionRequest 创建版本请求
type TemporalCreateVersionRequest struct {
	TenantID        string    `json:"tenantId" validate:"required,uuid"`
	Code            string    `json:"code" validate:"required,max=10"`
	Name            string    `json:"name" validate:"required,max=255"`
	UnitType        string    `json:"unitType" validate:"required"`
	Status          string    `json:"status" validate:"required"`
	ParentCode      *string   `json:"parentCode,omitempty"`
	Level           int       `json:"level" validate:"min=1,max=17"`
	SortOrder       int       `json:"sortOrder"`
	Description     string    `json:"description,omitempty"`
	EffectiveDate   time.Time `json:"effectiveDate" validate:"required"`
	OperationReason string    `json:"operationReason" validate:"omitempty,max=500"`
}

// TemporalUpdateVersionRequest 更新版本请求
type TemporalUpdateVersionRequest struct {
	TenantID         string    `json:"tenantId" validate:"required,uuid"`
	RecordID         string    `json:"recordId" validate:"required,uuid"`
	NewEffectiveDate time.Time `json:"newEffectiveDate" validate:"required"`
	OperationReason  string    `json:"operationReason" validate:"omitempty,max=500"`
}

// TemporalDeleteVersionRequest 删除版本请求
type TemporalDeleteVersionRequest struct {
	TenantID        string `json:"tenantId" validate:"required,uuid"`
	RecordID        string `json:"recordId" validate:"required,uuid"`
	OperationReason string `json:"operationReason" validate:"omitempty,max=500"`
}

// TemporalStatusChangeRequest 状态变更请求
type TemporalStatusChangeRequest struct {
	TenantID        string    `json:"tenantId" validate:"required,uuid"`
	Code            string    `json:"code" validate:"required,max=10"`
	NewStatus       string    `json:"newStatus" validate:"required"`
	EffectiveDate   time.Time `json:"effectiveDate" validate:"required"`
	OperationReason string    `json:"operationReason" validate:"omitempty,max=500"`
}

// CreateVersion 创建新版本 - 单事务维护时间轴与审计
func (s *OrganizationTemporalService) CreateVersion(ctx context.Context, req *TemporalCreateVersionRequest, actorID, requestID string) (result *repository.TimelineVersion, err error) {
	defer func() {
		utils.RecordTemporalOperation(utils.OperationCreate, err)
	}()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	s.logger.Infof("创建组织版本: Code=%s, 生效日期=%s", req.Code, req.EffectiveDate.Format("2006-01-02"))

	// 并发互斥：对同一 tenantId+code 使用咨询锁
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("无效租户ID: %w", err)
	}

	lockKey := fmt.Sprintf("%s:%s", req.TenantID, req.Code)
	_, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", lockKey)
	if err != nil {
		return nil, fmt.Errorf("获取咨询锁失败: %w", err)
	}

	var normalizedParent *string
	if req.ParentCode != nil {
		normalizedParent = utils.NormalizeParentCodePointer(req.ParentCode)
	}

	normalizedReason := strings.TrimSpace(req.OperationReason)

	org := &types.Organization{
		TenantID:      req.TenantID,
		Code:          req.Code,
		Name:          req.Name,
		UnitType:      req.UnitType,
		Status:        req.Status,
		ParentCode:    normalizedParent,
		SortOrder:     req.SortOrder,
		Description:   req.Description,
		EffectiveDate: types.NewDateFromTime(req.EffectiveDate),
		ChangeReason: func() *string {
			if normalizedReason == "" {
				return nil
			}
			reason := normalizedReason
			return &reason
		}(),
	}

	fields, calcErr := s.orgRepo.ComputeHierarchyForNew(ctx, tenantID, req.Code, normalizedParent, req.Name)
	if calcErr != nil {
		return nil, fmt.Errorf("计算组织层级失败: %w", calcErr)
	}
	org.Level = fields.Level
	org.CodePath = fields.CodePath
	org.NamePath = fields.NamePath

	// 1. 执行时态操作
	result, err = s.timelineManager.InsertVersion(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("插入版本失败: %w", err)
	}

	// 2. 写入审计日志
	afterData := map[string]interface{}{
		"code":           req.Code,
		"name":           req.Name,
		"unit_type":      req.UnitType,
		"status":         req.Status,
		"parent_code":    req.ParentCode,
		"level":          req.Level,
		"sort_order":     req.SortOrder,
		"description":    req.Description,
		"effective_date": req.EffectiveDate,
	}

	event := s.newAuditEvent(ctx, tenantID, actorID, "CreateOrganizationVersion", audit.EventTypeCreate, result.RecordID, req.Code, requestID, normalizedReason)
	event.AfterData = afterData
	event.ContextPayload = afterData

	if err := s.auditLogger.LogEventInTransaction(ctx, tx, event); err != nil {
		return nil, fmt.Errorf("审计写入失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	s.logger.Infof("组织版本创建完成: RecordID=%s", result.RecordID)
	return result, nil
}

// UpdateVersionEffectiveDate 修改版本生效日期 - 单事务维护时间轴与审计
func (s *OrganizationTemporalService) UpdateVersionEffectiveDate(ctx context.Context, req *TemporalUpdateVersionRequest, actorID, requestID string) (timeline *[]repository.TimelineVersion, err error) {
	defer func() {
		utils.RecordTemporalOperation(utils.OperationUpdate, err)
	}()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("无效租户ID: %w", err)
	}

	recordID, err := uuid.Parse(req.RecordID)
	if err != nil {
		return nil, fmt.Errorf("无效记录ID: %w", err)
	}

	s.logger.Infof("修改版本生效日期: RecordID=%s, 新日期=%s", recordID, req.NewEffectiveDate.Format("2006-01-02"))

	// 获取原版本数据用于审计
	var oldData map[string]interface{}
	var code string
	row := tx.QueryRowContext(ctx, `
		SELECT code, name, unit_type, status, parent_code, level, sort_order, description, effective_date
		FROM organization_units
		WHERE record_id = $1 AND tenant_id = $2 AND status != 'DELETED'
	`, recordID, tenantID)

	var name, unitType, status string
	var parentCode *string
	var level, sortOrder int
	var description string
	var effectiveDate time.Time

	err = row.Scan(&code, &name, &unitType, &status, &parentCode, &level, &sortOrder, &description, &effectiveDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("版本不存在: %s", recordID)
		}
		return nil, fmt.Errorf("查询版本数据失败: %w", err)
	}

	oldData = map[string]interface{}{
		"code":           code,
		"name":           name,
		"unit_type":      unitType,
		"status":         status,
		"parent_code":    parentCode,
		"level":          level,
		"sort_order":     sortOrder,
		"description":    description,
		"effective_date": effectiveDate,
	}

	// 并发互斥锁
	lockKey := fmt.Sprintf("%s:%s", req.TenantID, code)
	_, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", lockKey)
	if err != nil {
		return nil, fmt.Errorf("获取咨询锁失败: %w", err)
	}

	// 1. 执行时态操作
	normalizedReason := strings.TrimSpace(req.OperationReason)
	timeline, err = s.timelineManager.UpdateVersionEffectiveDate(ctx, tenantID, recordID, req.NewEffectiveDate, normalizedReason)
	if err != nil {
		return nil, fmt.Errorf("更新版本生效日期失败: %w", err)
	}

	// 2. 写入审计日志
	newData := make(map[string]interface{})
	for k, v := range oldData {
		newData[k] = v
	}
	newData["effective_date"] = req.NewEffectiveDate

	event := s.newAuditEvent(ctx, tenantID, actorID, "UpdateOrganizationVersionEffectiveDate", audit.EventTypeUpdate, recordID, code, requestID, normalizedReason)
	event.BeforeData = oldData
	event.AfterData = newData
	event.ContextPayload = newData
	event.ModifiedFields = []string{"effective_date"}
	event.Changes = []audit.FieldChange{
		{
			Field:    "effective_date",
			OldValue: oldData["effective_date"],
			NewValue: req.NewEffectiveDate,
			DataType: "timestamp",
		},
	}

	if err := s.auditLogger.LogEventInTransaction(ctx, tx, event); err != nil {
		return nil, fmt.Errorf("审计写入失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	s.logger.Info("版本生效日期修改完成")
	return timeline, nil
}

// DeleteVersion 删除版本 - 单事务维护时间轴与审计
func (s *OrganizationTemporalService) DeleteVersion(ctx context.Context, req *TemporalDeleteVersionRequest, actorID, requestID string) (timeline *[]repository.TimelineVersion, err error) {
	defer func() {
		utils.RecordTemporalOperation(utils.OperationDelete, err)
	}()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("无效租户ID: %w", err)
	}

	recordID, err := uuid.Parse(req.RecordID)
	if err != nil {
		return nil, fmt.Errorf("无效记录ID: %w", err)
	}

	s.logger.Infof("删除组织版本: RecordID=%s", recordID)

	// 获取版本数据用于审计和锁定
	var beforeData map[string]interface{}
	var code string
	row := tx.QueryRowContext(ctx, `
		SELECT code, name, unit_type, status, parent_code, level, sort_order, description, effective_date
		FROM organization_units
		WHERE record_id = $1 AND tenant_id = $2 AND status != 'DELETED'
		FOR UPDATE
	`, recordID, tenantID)

	var name, unitType, status string
	var parentCode *string
	var level, sortOrder int
	var description string
	var effectiveDate time.Time

	err = row.Scan(&code, &name, &unitType, &status, &parentCode, &level, &sortOrder, &description, &effectiveDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("版本不存在或已删除: %s", recordID)
		}
		return nil, fmt.Errorf("查询版本数据失败: %w", err)
	}

	beforeData = map[string]interface{}{
		"code":           code,
		"name":           name,
		"unit_type":      unitType,
		"status":         status,
		"parent_code":    parentCode,
		"level":          level,
		"sort_order":     sortOrder,
		"description":    description,
		"effective_date": effectiveDate,
	}

	// 并发互斥锁
	lockKey := fmt.Sprintf("%s:%s", req.TenantID, code)
	_, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", lockKey)
	if err != nil {
		return nil, fmt.Errorf("获取咨询锁失败: %w", err)
	}

	// 1. 执行时态操作
	normalizedReason := strings.TrimSpace(req.OperationReason)
	timeline, err = s.timelineManager.DeleteVersion(ctx, tenantID, recordID)
	if err != nil {
		return nil, fmt.Errorf("删除版本失败: %w", err)
	}

	// 2. 写入审计日志
	event := s.newAuditEvent(ctx, tenantID, actorID, "DeleteOrganizationVersion", audit.EventTypeDelete, recordID, code, requestID, normalizedReason)
	event.BeforeData = beforeData
	event.ContextPayload = beforeData
	event.Success = true

	if err := s.auditLogger.LogEventInTransaction(ctx, tx, event); err != nil {
		return nil, fmt.Errorf("审计写入失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	s.logger.Info("版本删除完成")
	return timeline, nil
}

// SuspendOrganization 暂停组织 - 单事务维护时间轴与审计
func (s *OrganizationTemporalService) SuspendOrganization(ctx context.Context, req *TemporalStatusChangeRequest, actorID, requestID string) (timeline *[]repository.TimelineVersion, err error) {
	defer func() {
		utils.RecordTemporalOperation(utils.OperationSuspend, err)
	}()
	timeline, err = s.changeOrganizationStatus(ctx, req, "INACTIVE", "SUSPEND", actorID, requestID)
	return
}

// ActivateOrganization 激活组织 - 单事务维护时间轴与审计
func (s *OrganizationTemporalService) ActivateOrganization(ctx context.Context, req *TemporalStatusChangeRequest, actorID, requestID string) (timeline *[]repository.TimelineVersion, err error) {
	defer func() {
		utils.RecordTemporalOperation(utils.OperationReactivate, err)
	}()
	timeline, err = s.changeOrganizationStatus(ctx, req, "ACTIVE", "REACTIVATE", actorID, requestID)
	return
}

// changeOrganizationStatus 通用状态变更逻辑
func (s *OrganizationTemporalService) changeOrganizationStatus(ctx context.Context, req *TemporalStatusChangeRequest, targetStatus, operationType, actorID, requestID string) (*[]repository.TimelineVersion, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("无效租户ID: %w", err)
	}

	s.logger.Infof("%s 组织: Code=%s, 生效日期=%s", operationType, req.Code, req.EffectiveDate.Format("2006-01-02"))

	// 并发互斥锁
	lockKey := fmt.Sprintf("%s:%s", req.TenantID, req.Code)
	_, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", lockKey)
	if err != nil {
		return nil, fmt.Errorf("获取咨询锁失败: %w", err)
	}

	// 1. 执行时态操作
	var timeline *[]repository.TimelineVersion
	normalizedReason := strings.TrimSpace(req.OperationReason)
	if operationType == "SUSPEND" {
		timeline, err = s.timelineManager.SuspendOrganization(ctx, tenantID, req.Code, req.EffectiveDate, normalizedReason)
	} else {
		timeline, err = s.timelineManager.ActivateOrganization(ctx, tenantID, req.Code, req.EffectiveDate, normalizedReason)
	}
	if err != nil {
		return nil, fmt.Errorf("%s组织失败: %w", operationType, err)
	}

	// 2. 写入审计日志（状态变更类型）
	var newRecordID uuid.UUID
	for _, version := range *timeline {
		if version.IsCurrent && version.Status == targetStatus {
			newRecordID = version.RecordID
			break
		}
	}

	if newRecordID == uuid.Nil {
		s.logger.Warnf("未找到新创建的%s版本，跳过审计", operationType)
	} else {
		prevStatus := getOppositeStatus(targetStatus)
		afterData := map[string]interface{}{"status": targetStatus}
		actionName := fmt.Sprintf("%s_ORGANIZATION", operationType)
		event := s.newAuditEvent(ctx, tenantID, actorID, actionName, audit.EventTypeUpdate, newRecordID, req.Code, requestID, normalizedReason)
		event.BeforeData = map[string]interface{}{"status": prevStatus}
		event.AfterData = afterData
		event.ContextPayload = map[string]interface{}{
			"status":        targetStatus,
			"operationType": operationType,
		}
		event.ModifiedFields = []string{"status"}
		event.Changes = []audit.FieldChange{
			{Field: "status", OldValue: prevStatus, NewValue: targetStatus, DataType: "string"},
		}

		if err := s.auditLogger.LogEventInTransaction(ctx, tx, event); err != nil {
			return nil, fmt.Errorf("审计写入失败: %w", err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	s.logger.Infof("组织%s完成", operationType)
	return timeline, nil
}

func (s *OrganizationTemporalService) newAuditEvent(ctx context.Context, tenantID uuid.UUID, actorID, actionName string, eventType string, resourceID uuid.UUID, entityCode, requestID, operationReason string) *audit.AuditEvent {
	trimmedActorID := strings.TrimSpace(actorID)
	actorType := audit.ActorTypeUser
	if trimmedActorID == "" {
		trimmedActorID = "system"
		actorType = audit.ActorTypeSystem
	}
	actorName := trimmedActorID

	correlationID := orgmiddleware.GetCorrelationID(ctx)
	if correlationID == "" && requestID != "" {
		correlationID = requestID
	}
	sourceCorrelation := ""
	if orgmiddleware.GetCorrelationSource(ctx) == "header" {
		sourceCorrelation = "header"
	}

	return &audit.AuditEvent{
		TenantID:          tenantID,
		EventType:         eventType,
		ResourceType:      audit.ResourceTypeOrganization,
		ResourceID:        resourceID.String(),
		RecordID:          resourceID,
		EntityCode:        entityCode,
		ActorID:           trimmedActorID,
		ActorType:         actorType,
		ActorName:         actorName,
		ActionName:        actionName,
		RequestID:         requestID,
		CorrelationID:     correlationID,
		SourceCorrelation: sourceCorrelation,
		OperationReason:   strings.TrimSpace(operationReason),
		Success:           true,
	}
}

// getOppositeStatus 获取状态的相反状态
func getOppositeStatus(status string) string {
	if status == "ACTIVE" {
		return "INACTIVE"
	}
	return "ACTIVE"
}
