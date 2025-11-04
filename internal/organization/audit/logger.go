package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

// AuditLogger 结构化审计日志记录器
type AuditLogger struct {
	db     *sql.DB
	logger pkglogger.Logger
}

// AuditEvent 简化的审计事件 (v4.3.0 - 移除过度设计的技术细节追踪)
type AuditEvent struct {
	ID              uuid.UUID              `json:"id"`
	TenantID        uuid.UUID              `json:"tenantId"`
	EventType       string                 `json:"eventType"`
	ResourceType    string                 `json:"resourceType"`
	ResourceID      string                 `json:"resourceId"`
	ActorID         string                 `json:"actorId"`
	ActorType       string                 `json:"actorType"`
	ActionName      string                 `json:"actionName"`
	RequestID       string                 `json:"requestId"`
	OperationReason string                 `json:"operationReason,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
	Success         bool                   `json:"success"`
	ErrorCode       string                 `json:"errorCode,omitempty"`
	ErrorMessage    string                 `json:"errorMessage,omitempty"`
	BeforeData      map[string]interface{} `json:"beforeData,omitempty"`
	AfterData       map[string]interface{} `json:"afterData,omitempty"`
	ModifiedFields  []string               `json:"modifiedFields,omitempty"`
	Changes         []FieldChange          `json:"changes,omitempty"`
}

// FieldChange 字段变更记录 - 保留关键审计功能
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
	DataType string      `json:"dataType"`
}

// 审计事件类型常量
const (
	EventTypeCreate     = "CREATE"
	EventTypeUpdate     = "UPDATE"
	EventTypeDelete     = "DELETE"
	EventTypeSuspend    = "SUSPEND"
	EventTypeActivate   = "ACTIVATE"
	EventTypeQuery      = "QUERY"
	EventTypeValidation = "VALIDATION"
	EventTypeAuth       = "AUTHENTICATION"
	EventTypeError      = "ERROR"
)

// 资源类型常量
const (
	ResourceTypeOrganization = "ORGANIZATION"
	ResourceTypeHierarchy    = "HIERARCHY"
	ResourceTypeJobCatalog   = "JOB_CATALOG"
	ResourceTypePosition     = "POSITION"
	ResourceTypeUser         = "USER"
	ResourceTypeSystem       = "SYSTEM"
)

// 操作者类型常量
const (
	ActorTypeUser    = "USER"
	ActorTypeSystem  = "SYSTEM"
	ActorTypeService = "SERVICE"
)

func NewAuditLogger(db *sql.DB, baseLogger pkglogger.Logger) *AuditLogger {
	return &AuditLogger{
		db: db,
		logger: baseLogger.WithFields(pkglogger.Fields{
			"component": "audit",
			"module":    "command",
		}),
	}
}

// LogEvent 记录审计事件
func (a *AuditLogger) LogEvent(ctx context.Context, event *AuditEvent) error {
	if event.ID == (uuid.UUID{}) {
		event.ID = uuid.New()
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// 序列化JSON字段（以字符串形式传递，并在SQL中显式::jsonb转换，避免驱动类型歧义）
	bd, _ := json.Marshal(event.BeforeData)
	ad, _ := json.Marshal(event.AfterData)
	mf, _ := json.Marshal(event.ModifiedFields)
	ch, _ := json.Marshal(event.Changes)
	beforeDataJSON := string(bd)
	afterDataJSON := string(ad)
	modifiedFieldsJSON := string(mf)
	changesJSON := string(ch)

	if beforeDataJSON == "null" {
		beforeDataJSON = "{}"
	}
	if afterDataJSON == "null" {
		afterDataJSON = "{}"
	}
	if modifiedFieldsJSON == "null" {
		modifiedFieldsJSON = "[]"
	}
	if changesJSON == "null" {
		changesJSON = "[]"
	}

	query := `
    INSERT INTO audit_logs (
        id, tenant_id, event_type, resource_type, resource_id,
        actor_id, actor_type, action_name, request_id, operation_reason,
        timestamp, success, error_code, error_message,
        request_data, response_data, modified_fields, changes
    ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15::jsonb, $16::jsonb, $17::jsonb, $18::jsonb
    )`

	// resource_id 列为 UUID，可为 NULL。允许将非UUID字符串视为 NULL，避免外键/类型错误。
	var resIDParam interface{}
	if event.ResourceID != "" {
		if rid, err := uuid.Parse(event.ResourceID); err == nil {
			resIDParam = rid
		}
	}
	// 兜底：对 ORGANIZATION 资源，若未提供有效 UUID，则根据 (tenant_id, code, current) 推导 record_id
	if resIDParam == nil && event.ResourceType == ResourceTypeOrganization {
		// 优先从 AfterData/BeforeData 取业务 code
		var codeCandidate string
		if event.AfterData != nil {
			if v, ok := event.AfterData["code"].(string); ok && v != "" {
				codeCandidate = v
			}
		}
		if codeCandidate == "" && event.BeforeData != nil {
			if v, ok := event.BeforeData["code"].(string); ok && v != "" {
				codeCandidate = v
			}
		}
		if codeCandidate != "" {
			var rid uuid.UUID
			err := a.db.QueryRowContext(ctx,
				`SELECT record_id FROM organization_units WHERE tenant_id = $1 AND code = $2 AND is_current = true LIMIT 1`,
				event.TenantID.String(), codeCandidate,
			).Scan(&rid)
			if err == nil {
				resIDParam = rid
			}
		}
	}

	// 最终兜底：对于 AUTH / USER / SYSTEM 等非 UUID 资源，使用原始 ResourceID 或组合键
	if resIDParam == nil {
		if event.ResourceID != "" {
			resIDParam = event.ResourceID
		} else {
			resIDParam = fmt.Sprintf("%s_%s_%s", event.EventType, event.ResourceType, event.ActionName)
		}
	}

	_, err := a.db.ExecContext(ctx, query,
		event.ID, event.TenantID, event.EventType, event.ResourceType, resIDParam,
		event.ActorID, event.ActorType, event.ActionName, event.RequestID, event.OperationReason,
		event.Timestamp, event.Success, event.ErrorCode, event.ErrorMessage,
		beforeDataJSON, afterDataJSON, modifiedFieldsJSON, changesJSON,
	)

	if err != nil {
		a.logger.Errorf("审计日志记录失败: %v", err)
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	a.logger.Infof("审计事件已记录: %s/%s/%s (ID: %s)",
		event.EventType, event.ResourceType, event.ActionName, event.ID.String())

	return nil
}

// LogOrganizationCreate 记录组织创建事件 (v4.3.0 - 简化审计信息)
func (a *AuditLogger) LogOrganizationCreate(ctx context.Context, req *types.CreateOrganizationRequest, result *types.Organization, actorID, requestID, operationReason string) error {
	tenantID, _ := uuid.Parse(result.TenantID)
	// 计算创建时的“新增字段”列表（无beforeData，oldValue为null）
	createdFields := []FieldChange{}
	modifiedFields := []string{}
	// 基本字段
	for _, fc := range []struct{ field, dtype string }{
		{"code", "string"}, {"name", "string"}, {"unitType", "string"}, {"parentCode", "string"},
		{"status", "string"}, {"level", "int"},
	} {
		createdFields = append(createdFields, FieldChange{Field: fc.field, OldValue: nil, NewValue: nil, DataType: fc.dtype})
		modifiedFields = append(modifiedFields, fc.field)
	}
	// 时态相关（若存在）
	if result.EffectiveDate != nil {
		modifiedFields = append(modifiedFields, "effectiveDate")
	}
	if result.EndDate != nil {
		modifiedFields = append(modifiedFields, "endDate")
	}

	event := &AuditEvent{
		TenantID:        tenantID,
		EventType:       EventTypeCreate,
		ResourceType:    ResourceTypeOrganization,
		ResourceID:      result.RecordID,
		ActorID:         actorID,
		ActorType:       ActorTypeUser,
		ActionName:      "CreateOrganization",
		RequestID:       requestID,
		OperationReason: operationReason,
		Success:         true,
		ModifiedFields:  modifiedFields,
		Changes:         createdFields,
		AfterData: map[string]interface{}{
			"code":       result.Code,
			"name":       result.Name,
			"unitType":   result.UnitType,
			"parentCode": result.ParentCode,
			"status":     result.Status,
			"level":      result.Level,
		},
	}

	return a.LogEvent(ctx, event)
}

// LogOrganizationUpdate 记录组织更新事件 (v4.3.0 - 简化参数，保留FieldChange)
func (a *AuditLogger) LogOrganizationUpdate(ctx context.Context, code string, req *types.UpdateOrganizationRequest, oldOrg, newOrg *types.Organization, actorID, requestID, operationReason string) error {
	changes := a.calculateFieldChanges(oldOrg, newOrg)
	modifiedFields := make([]string, len(changes))
	for i, change := range changes {
		modifiedFields[i] = change.Field
	}

	tenantID, _ := uuid.Parse(newOrg.TenantID)

	var beforeData, afterData map[string]interface{}
	if oldOrg != nil {
		beforeData = structToMap(oldOrg)
	}
	if newOrg != nil {
		afterData = structToMap(newOrg)
	}

	// 使用newOrg的RecordID作为ResourceID
	resourceID := newOrg.RecordID
	if newOrg == nil && oldOrg != nil {
		resourceID = oldOrg.RecordID
	}

	event := &AuditEvent{
		TenantID:        tenantID,
		EventType:       EventTypeUpdate,
		ResourceType:    ResourceTypeOrganization,
		ResourceID:      resourceID,
		ActorID:         actorID,
		ActorType:       ActorTypeUser,
		ActionName:      "UpdateOrganization",
		RequestID:       requestID,
		OperationReason: operationReason,
		Success:         true,
		BeforeData:      beforeData,
		AfterData:       afterData,
		Changes:         changes,
		ModifiedFields:  modifiedFields,
	}

	return a.LogEvent(ctx, event)
}

// LogOrganizationSuspend 记录组织停用事件 (v4.3.0 - 简化参数)
func (a *AuditLogger) LogOrganizationSuspend(ctx context.Context, code string, org *types.Organization, actorID, requestID, operationReason string) error {
	tenantID, _ := uuid.Parse(org.TenantID)
	// 停用：记录状态字段变更
	changes := []FieldChange{{Field: "status", OldValue: org.Status, NewValue: "INACTIVE", DataType: "string"}}
	modified := []string{"status"}
	event := &AuditEvent{
		TenantID:        tenantID,
		EventType:       EventTypeSuspend,
		ResourceType:    ResourceTypeOrganization,
		ResourceID:      org.RecordID,
		ActorID:         actorID,
		ActorType:       ActorTypeUser,
		ActionName:      "SuspendOrganization",
		RequestID:       requestID,
		OperationReason: operationReason,
		Success:         true,
		ModifiedFields:  modified,
		Changes:         changes,
		AfterData: map[string]interface{}{
			"code":   org.Code,
			"status": "INACTIVE",
			"level":  org.Level,
		},
	}

	return a.LogEvent(ctx, event)
}

// LogOrganizationActivate 记录组织激活事件 (v4.3.0 - 简化参数)
func (a *AuditLogger) LogOrganizationActivate(ctx context.Context, code string, org *types.Organization, actorID, requestID, operationReason string) error {
	tenantID, _ := uuid.Parse(org.TenantID)
	// 激活：记录状态字段变更
	changes := []FieldChange{{Field: "status", OldValue: org.Status, NewValue: "ACTIVE", DataType: "string"}}
	modified := []string{"status"}
	event := &AuditEvent{
		TenantID:        tenantID,
		EventType:       EventTypeActivate,
		ResourceType:    ResourceTypeOrganization,
		ResourceID:      org.RecordID,
		ActorID:         actorID,
		ActorType:       ActorTypeUser,
		ActionName:      "ActivateOrganization",
		RequestID:       requestID,
		OperationReason: operationReason,
		Success:         true,
		ModifiedFields:  modified,
		Changes:         changes,
		AfterData: map[string]interface{}{
			"code":   org.Code,
			"status": "ACTIVE",
			"level":  org.Level,
		},
	}

	return a.LogEvent(ctx, event)
}

// LogOrganizationDelete 记录组织删除事件 (v4.3.0 - 简化参数)
func (a *AuditLogger) LogOrganizationDelete(ctx context.Context, tenantID uuid.UUID, code string, org *types.Organization, actorID, requestID, operationReason string) error {
	var beforeData map[string]interface{}
	var resourceID string

	// 如果有组织数据，记录删除前状态和使用正确的RecordID
	if org != nil {
		beforeData = map[string]interface{}{
			"code":   org.Code,
			"name":   org.Name,
			"status": org.Status,
			"level":  org.Level,
		}
		resourceID = org.RecordID
	} else {
		// 如果没有组织数据，这种情况需要从数据库查询RecordID
		// 为了简化，这里使用code，但这会导致UUID类型错误
		// TODO-TEMPORARY: Should pass correct RecordID from caller; refactor deletion audit in v4.3 by 2025-09-20.
		resourceID = code
	}

	// 删除：记录状态字段变更为 DELETED（若可用）
	var changes []FieldChange
	var modified []string
	if org != nil {
		changes = []FieldChange{{Field: "status", OldValue: org.Status, NewValue: "DELETED", DataType: "string"}}
		modified = []string{"status"}
	}
	event := &AuditEvent{
		TenantID:        tenantID,
		EventType:       EventTypeDelete,
		ResourceType:    ResourceTypeOrganization,
		ResourceID:      resourceID,
		ActorID:         actorID,
		ActorType:       ActorTypeUser,
		ActionName:      "DeleteOrganization",
		RequestID:       requestID,
		OperationReason: operationReason,
		Success:         true,
		ModifiedFields:  modified,
		Changes:         changes,
		BeforeData:      beforeData,
	}

	return a.LogEvent(ctx, event)
}

// LogError 记录错误事件 (v4.3.0 - 简化参数)
func (a *AuditLogger) LogError(ctx context.Context, tenantID uuid.UUID, resourceType, resourceID, actionName, actorID, requestID, errorCode, errorMessage string, requestData map[string]interface{}) error {
	event := &AuditEvent{
		TenantID:     tenantID,
		EventType:    EventTypeError,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ActorID:      actorID,
		ActorType:    ActorTypeUser,
		ActionName:   actionName,
		RequestID:    requestID,
		Success:      false,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		BeforeData:   requestData,
	}

	return a.LogEvent(ctx, event)
}

// calculateFieldChanges 计算字段变更
func (a *AuditLogger) calculateFieldChanges(oldOrg, newOrg *types.Organization) []FieldChange {
	var changes []FieldChange

	// 安全检查：确保两个组织对象都不为nil
	if oldOrg == nil || newOrg == nil {
		// 如果oldOrg为nil，表示这是创建操作或无法获取旧数据
		// 如果newOrg为nil，表示这是删除操作或数据获取失败
		// 在这些情况下，返回空的变更列表
		return changes
	}

	if oldOrg.Name != newOrg.Name {
		changes = append(changes, FieldChange{
			Field:    "name",
			OldValue: oldOrg.Name,
			NewValue: newOrg.Name,
			DataType: "string",
		})
	}

	if oldOrg.Status != newOrg.Status {
		changes = append(changes, FieldChange{
			Field:    "status",
			OldValue: oldOrg.Status,
			NewValue: newOrg.Status,
			DataType: "string",
		})
	}

	if oldOrg.Description != newOrg.Description {
		changes = append(changes, FieldChange{
			Field:    "description",
			OldValue: oldOrg.Description,
			NewValue: newOrg.Description,
			DataType: "string",
		})
	}

	if oldOrg.SortOrder != newOrg.SortOrder {
		changes = append(changes, FieldChange{
			Field:    "sortOrder",
			OldValue: oldOrg.SortOrder,
			NewValue: newOrg.SortOrder,
			DataType: "int",
		})
	}

	// 检查父组织变更
	oldParent := ""
	newParent := ""
	if oldOrg.ParentCode != nil {
		oldParent = *oldOrg.ParentCode
	}
	if newOrg.ParentCode != nil {
		newParent = *newOrg.ParentCode
	}

	if oldParent != newParent {
		changes = append(changes, FieldChange{
			Field:    "parentCode",
			OldValue: oldParent,
			NewValue: newParent,
			DataType: "string",
		})
	}

	// 检查生效日期变更
	if (oldOrg.EffectiveDate == nil) != (newOrg.EffectiveDate == nil) {
		// 一个为nil，一个不为nil
		changes = append(changes, FieldChange{
			Field:    "effectiveDate",
			OldValue: oldOrg.EffectiveDate,
			NewValue: newOrg.EffectiveDate,
			DataType: "date",
		})
	} else if oldOrg.EffectiveDate != nil && newOrg.EffectiveDate != nil {
		// 都不为nil，比较日期值
		if !oldOrg.EffectiveDate.Time.Equal(newOrg.EffectiveDate.Time) {
			changes = append(changes, FieldChange{
				Field:    "effectiveDate",
				OldValue: oldOrg.EffectiveDate,
				NewValue: newOrg.EffectiveDate,
				DataType: "date",
			})
		}
	}

	// 检查结束日期变更
	if (oldOrg.EndDate == nil) != (newOrg.EndDate == nil) {
		// 一个为nil，一个不为nil
		changes = append(changes, FieldChange{
			Field:    "endDate",
			OldValue: oldOrg.EndDate,
			NewValue: newOrg.EndDate,
			DataType: "date",
		})
	} else if oldOrg.EndDate != nil && newOrg.EndDate != nil {
		// 都不为nil，比较日期值
		if !oldOrg.EndDate.Time.Equal(newOrg.EndDate.Time) {
			changes = append(changes, FieldChange{
				Field:    "endDate",
				OldValue: oldOrg.EndDate,
				NewValue: newOrg.EndDate,
				DataType: "date",
			})
		}
	}

	// 检查变更原因
	if (oldOrg.ChangeReason == nil) != (newOrg.ChangeReason == nil) {
		// 一个为nil，一个不为nil
		changes = append(changes, FieldChange{
			Field:    "changeReason",
			OldValue: oldOrg.ChangeReason,
			NewValue: newOrg.ChangeReason,
			DataType: "string",
		})
	} else if oldOrg.ChangeReason != nil && newOrg.ChangeReason != nil {
		// 都不为nil，比较字符串值
		if *oldOrg.ChangeReason != *newOrg.ChangeReason {
			changes = append(changes, FieldChange{
				Field:    "changeReason",
				OldValue: *oldOrg.ChangeReason,
				NewValue: *newOrg.ChangeReason,
				DataType: "string",
			})
		}
	}

	return changes
}

// structToMap 将结构体转换为map
func structToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	data, err := json.Marshal(obj)
	if err != nil {
		return result
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result
	}
	return result
}

// GetAuditHistory 获取资源审计历史
func (a *AuditLogger) GetAuditHistory(ctx context.Context, resourceType, resourceID string, tenantID uuid.UUID, limit int) ([]AuditEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
	SELECT 
		id, tenant_id, event_type, resource_type, resource_id,
		actor_id, actor_type, action_name, request_id,
		COALESCE(operation_reason, '') as operation_reason,
		timestamp, success,
		COALESCE(error_code, '') as error_code,
		COALESCE(error_message, '') as error_message,
		COALESCE(before_data, '{}') as before_data,
		COALESCE(after_data, '{}') as after_data,
		COALESCE(modified_fields, '[]') as modified_fields,
		COALESCE(changes, '[]') as changes
	FROM audit_logs 
	WHERE resource_type = $1 AND resource_id = $2 AND tenant_id = $3
	ORDER BY timestamp DESC
	LIMIT $4`

	rows, err := a.db.QueryContext(ctx, query, resourceType, resourceID, tenantID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit history: %w", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var event AuditEvent
		var beforeDataJSON, afterDataJSON, modifiedFieldsJSON, changesJSON string

		err := rows.Scan(
			&event.ID, &event.TenantID, &event.EventType, &event.ResourceType, &event.ResourceID,
			&event.ActorID, &event.ActorType, &event.ActionName, &event.RequestID,
			&event.OperationReason, &event.Timestamp, &event.Success,
			&event.ErrorCode, &event.ErrorMessage, &beforeDataJSON, &afterDataJSON,
			&modifiedFieldsJSON, &changesJSON,
		)
		if err != nil {
			a.logger.Errorf("扫描审计记录失败: %v", err)
			continue
		}

		// 反序列化JSON字段
		if err := json.Unmarshal([]byte(beforeDataJSON), &event.BeforeData); err != nil {
			a.logger.Warnf("解析before_data失败: %v", err)
		}
		if err := json.Unmarshal([]byte(afterDataJSON), &event.AfterData); err != nil {
			a.logger.Warnf("解析after_data失败: %v", err)
		}
		if err := json.Unmarshal([]byte(modifiedFieldsJSON), &event.ModifiedFields); err != nil {
			a.logger.Warnf("解析modified_fields失败: %v", err)
		}
		if err := json.Unmarshal([]byte(changesJSON), &event.Changes); err != nil {
			a.logger.Warnf("解析changes失败: %v", err)
		}

		events = append(events, event)
	}

	a.logger.Infof("审计历史查询: %s/%s, 返回%d条记录", resourceType, resourceID, len(events))
	return events, nil
}
