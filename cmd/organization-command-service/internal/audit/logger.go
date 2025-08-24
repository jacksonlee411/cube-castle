package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"organization-command-service/internal/types"
	"github.com/google/uuid"
)

// AuditLogger 结构化审计日志记录器
type AuditLogger struct {
	db     *sql.DB
	logger *log.Logger
}

// AuditEvent 审计事件
type AuditEvent struct {
	ID            uuid.UUID              `json:"id"`
	TenantID      uuid.UUID              `json:"tenantId"`
	EventType     string                 `json:"eventType"`
	ResourceType  string                 `json:"resourceType"`
	ResourceID    string                 `json:"resourceId"`
	ActorID       string                 `json:"actorId"`
	ActorType     string                 `json:"actorType"`
	ActionName    string                 `json:"actionName"`
	RequestID     string                 `json:"requestId"`
	IPAddress     string                 `json:"ipAddress"`
	UserAgent     string                 `json:"userAgent"`
	Timestamp     time.Time              `json:"timestamp"`
	Success       bool                   `json:"success"`
	ErrorCode     string                 `json:"errorCode,omitempty"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
	RequestData   map[string]interface{} `json:"requestData,omitempty"`
	ResponseData  map[string]interface{} `json:"responseData,omitempty"`
	Changes       []FieldChange          `json:"changes,omitempty"`
	BusinessContext map[string]interface{} `json:"businessContext,omitempty"`
}

// FieldChange 字段变更记录
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
	ResourceTypeUser         = "USER"
	ResourceTypeSystem       = "SYSTEM"
)

// 操作者类型常量
const (
	ActorTypeUser    = "USER"
	ActorTypeSystem  = "SYSTEM"
	ActorTypeService = "SERVICE"
)

func NewAuditLogger(db *sql.DB, logger *log.Logger) *AuditLogger {
	return &AuditLogger{
		db:     db,
		logger: logger,
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

	// 序列化JSON字段
	requestDataJSON, _ := json.Marshal(event.RequestData)
	responseDataJSON, _ := json.Marshal(event.ResponseData)
	changesJSON, _ := json.Marshal(event.Changes)
	businessContextJSON, _ := json.Marshal(event.BusinessContext)

	query := `
	INSERT INTO audit_logs (
		id, tenant_id, event_type, resource_type, resource_id,
		actor_id, actor_type, action_name, request_id,
		ip_address, user_agent, timestamp, success,
		error_code, error_message, request_data, response_data,
		changes, business_context
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
	)`

	_, err := a.db.ExecContext(ctx, query,
		event.ID, event.TenantID, event.EventType, event.ResourceType, event.ResourceID,
		event.ActorID, event.ActorType, event.ActionName, event.RequestID,
		event.IPAddress, event.UserAgent, event.Timestamp, event.Success,
		event.ErrorCode, event.ErrorMessage, requestDataJSON, responseDataJSON,
		changesJSON, businessContextJSON,
	)

	if err != nil {
		a.logger.Printf("审计日志记录失败: %v", err)
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	a.logger.Printf("✅ 审计事件已记录: %s/%s/%s (ID: %s)", 
		event.EventType, event.ResourceType, event.ActionName, event.ID.String())

	return nil
}

// LogOrganizationCreate 记录组织创建事件
func (a *AuditLogger) LogOrganizationCreate(ctx context.Context, req *types.CreateOrganizationRequest, result *types.Organization, actorID, requestID, ipAddress string) error {
	tenantID, _ := uuid.Parse(result.TenantID)
	event := &AuditEvent{
		TenantID:     tenantID,
		EventType:    EventTypeCreate,
		ResourceType: ResourceTypeOrganization,
		ResourceID:   result.Code,
		ActorID:      actorID,
		ActorType:    ActorTypeUser,
		ActionName:   "CreateOrganization",
		RequestID:    requestID,
		IPAddress:    ipAddress,
		Success:      true,
		RequestData: map[string]interface{}{
			"code":       req.Code,
			"name":       req.Name,
			"unitType":   req.UnitType,
			"parentCode": req.ParentCode,
		},
		ResponseData: map[string]interface{}{
			"organizationId": result.Code,
			"createdAt":     result.CreatedAt,
		},
		BusinessContext: map[string]interface{}{
			"operation":   "organization_creation",
			"level":       result.Level,
			"path":        result.Path,
			"sortOrder":   result.SortOrder,
		},
	}

	return a.LogEvent(ctx, event)
}

// LogOrganizationUpdate 记录组织更新事件
func (a *AuditLogger) LogOrganizationUpdate(ctx context.Context, code string, req *types.UpdateOrganizationRequest, oldOrg, newOrg *types.Organization, actorID, requestID, ipAddress string) error {
	changes := a.calculateFieldChanges(oldOrg, newOrg)
	tenantID, _ := uuid.Parse(newOrg.TenantID)
	
	event := &AuditEvent{
		TenantID:     tenantID,
		EventType:    EventTypeUpdate,
		ResourceType: ResourceTypeOrganization,
		ResourceID:   code,
		ActorID:      actorID,
		ActorType:    ActorTypeUser,
		ActionName:   "UpdateOrganization",
		RequestID:    requestID,
		IPAddress:    ipAddress,
		Success:      true,
		RequestData:  structToMap(req),
		ResponseData: map[string]interface{}{
			"organizationId": newOrg.Code,
			"updatedAt":     newOrg.UpdatedAt,
		},
		Changes: changes,
		BusinessContext: map[string]interface{}{
			"operation":     "organization_update",
			"changesCount":  len(changes),
			"oldLevel":      oldOrg.Level,
			"newLevel":      newOrg.Level,
		},
	}

	return a.LogEvent(ctx, event)
}

// LogOrganizationSuspend 记录组织停用事件
func (a *AuditLogger) LogOrganizationSuspend(ctx context.Context, code string, org *types.Organization, actorID, requestID, ipAddress string) error {
	tenantID, _ := uuid.Parse(org.TenantID)
	event := &AuditEvent{
		TenantID:     tenantID,
		EventType:    EventTypeSuspend,
		ResourceType: ResourceTypeOrganization,
		ResourceID:   code,
		ActorID:      actorID,
		ActorType:    ActorTypeUser,
		ActionName:   "SuspendOrganization",
		RequestID:    requestID,
		IPAddress:    ipAddress,
		Success:      true,
		ResponseData: map[string]interface{}{
			"organizationId": org.Code,
			"status":        "INACTIVE",
			"suspendedAt":   time.Now(),
		},
		BusinessContext: map[string]interface{}{
			"operation":    "organization_suspension",
			"previousStatus": "ACTIVE",
			"level":        org.Level,
			"hasChildren":  org.ParentCode != nil,
		},
	}

	return a.LogEvent(ctx, event)
}

// LogOrganizationActivate 记录组织激活事件
func (a *AuditLogger) LogOrganizationActivate(ctx context.Context, code string, org *types.Organization, actorID, requestID, ipAddress string) error {
	tenantID, _ := uuid.Parse(org.TenantID)
	event := &AuditEvent{
		TenantID:     tenantID,
		EventType:    EventTypeActivate,
		ResourceType: ResourceTypeOrganization,
		ResourceID:   code,
		ActorID:      actorID,
		ActorType:    ActorTypeUser,
		ActionName:   "ActivateOrganization",
		RequestID:    requestID,
		IPAddress:    ipAddress,
		Success:      true,
		ResponseData: map[string]interface{}{
			"organizationId": org.Code,
			"status":        "ACTIVE",
			"activatedAt":   time.Now(),
		},
		BusinessContext: map[string]interface{}{
			"operation":    "organization_activation",
			"previousStatus": "INACTIVE",
			"level":        org.Level,
		},
	}

	return a.LogEvent(ctx, event)
}

// LogOrganizationDelete 记录组织删除事件
func (a *AuditLogger) LogOrganizationDelete(ctx context.Context, code string, org *types.Organization, actorID, requestID, ipAddress string) error {
	tenantID, _ := uuid.Parse(org.TenantID)
	event := &AuditEvent{
		TenantID:     tenantID,
		EventType:    EventTypeDelete,
		ResourceType: ResourceTypeOrganization,
		ResourceID:   code,
		ActorID:      actorID,
		ActorType:    ActorTypeUser,
		ActionName:   "DeleteOrganization",
		RequestID:    requestID,
		IPAddress:    ipAddress,
		Success:      true,
		BusinessContext: map[string]interface{}{
			"operation":    "organization_deletion",
			"deletedName":  org.Name,
			"level":        org.Level,
			"path":         org.Path,
		},
	}

	return a.LogEvent(ctx, event)
}

// LogError 记录错误事件
func (a *AuditLogger) LogError(ctx context.Context, tenantID uuid.UUID, resourceType, resourceID, actionName, actorID, requestID, ipAddress, errorCode, errorMessage string, requestData map[string]interface{}) error {
	event := &AuditEvent{
		TenantID:     tenantID,
		EventType:    EventTypeError,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ActorID:      actorID,
		ActorType:    ActorTypeUser,
		ActionName:   actionName,
		RequestID:    requestID,
		IPAddress:    ipAddress,
		Success:      false,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		RequestData:  requestData,
		BusinessContext: map[string]interface{}{
			"operation": "error_logging",
			"severity":  "ERROR",
		},
	}

	return a.LogEvent(ctx, event)
}

// calculateFieldChanges 计算字段变更
func (a *AuditLogger) calculateFieldChanges(oldOrg, newOrg *types.Organization) []FieldChange {
	var changes []FieldChange

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

	return changes
}

// structToMap 将结构体转换为map
func structToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	data, _ := json.Marshal(obj)
	json.Unmarshal(data, &result)
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
		ip_address, user_agent, timestamp, success,
		COALESCE(error_code, '') as error_code,
		COALESCE(error_message, '') as error_message,
		COALESCE(request_data, '{}') as request_data,
		COALESCE(response_data, '{}') as response_data,
		COALESCE(changes, '[]') as changes,
		COALESCE(business_context, '{}') as business_context
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
		var requestDataJSON, responseDataJSON, changesJSON, businessContextJSON string

		err := rows.Scan(
			&event.ID, &event.TenantID, &event.EventType, &event.ResourceType, &event.ResourceID,
			&event.ActorID, &event.ActorType, &event.ActionName, &event.RequestID,
			&event.IPAddress, &event.UserAgent, &event.Timestamp, &event.Success,
			&event.ErrorCode, &event.ErrorMessage, &requestDataJSON, &responseDataJSON,
			&changesJSON, &businessContextJSON,
		)
		if err != nil {
			a.logger.Printf("扫描审计记录失败: %v", err)
			continue
		}

		// 反序列化JSON字段
		json.Unmarshal([]byte(requestDataJSON), &event.RequestData)
		json.Unmarshal([]byte(responseDataJSON), &event.ResponseData)
		json.Unmarshal([]byte(changesJSON), &event.Changes)
		json.Unmarshal([]byte(businessContextJSON), &event.BusinessContext)

		events = append(events, event)
	}

	a.logger.Printf("📊 审计历史查询: %s/%s, 返回%d条记录", resourceType, resourceID, len(events))
	return events, nil
}