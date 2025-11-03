package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"cube-castle/cmd/hrms-server/command/internal/utils"
)

// AuditWriter 按06文档要求实现的审计写入器
// 符合 021 迁移的审计契约：business_context、record_id、仅值变更写入
type AuditWriter struct {
	db     *sql.DB
	logger *log.Logger
}

func NewAuditWriter(db *sql.DB, logger *log.Logger) *AuditWriter {
	return &AuditWriter{
		db:     db,
		logger: logger,
	}
}

// AuditEntry 审计条目 - 对齐 021 迁移的 audit_logs 表结构
type AuditEntry struct {
	TenantID        uuid.UUID              `json:"tenantId"`
	EventType       string                 `json:"eventType"`    // CREATE, UPDATE, DELETE
	ResourceType    string                 `json:"resourceType"` // ORGANIZATION
	ActorID         string                 `json:"actorId"`
	ActorType       string                 `json:"actorType"`  // USER, SYSTEM
	ActionName      string                 `json:"actionName"` // CREATE_ORGANIZATION, UPDATE_ORGANIZATION, DELETE_ORGANIZATION
	RequestID       string                 `json:"requestId"`
	ResourceID      uuid.UUID              `json:"resourceId"` // record_id
	OperationReason string                 `json:"operationReason"`
	BeforeData      map[string]interface{} `json:"beforeData"`
	AfterData       map[string]interface{} `json:"afterData"`
	Changes         []FieldChange          `json:"changes"`
	BusinessContext map[string]interface{} `json:"businessContext"` // 06文档要求：支持 requestId/context
	RecordID        uuid.UUID              `json:"recordId"`        // 06文档要求：审计归属正确
}

type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
}

// WriteAudit 写入审计日志 - 按 06 文档要求跳过空 UPDATE
func (aw *AuditWriter) WriteAudit(ctx context.Context, entry *AuditEntry) error {
	// 跳过空 UPDATE：无字段变化时不写审计
	if entry.EventType == "UPDATE" && len(entry.Changes) == 0 {
		aw.logger.Printf("⏭️ 跳过空UPDATE审计: RecordID=%s", entry.RecordID)
		return nil
	}

	// 序列化JSON字段
	beforeDataJSON, _ := json.Marshal(entry.BeforeData)
	afterDataJSON, _ := json.Marshal(entry.AfterData)
	changesJSON, _ := json.Marshal(entry.Changes)
	businessContextJSON, _ := json.Marshal(entry.BusinessContext)

	beforeJSON := string(beforeDataJSON)
	afterJSON := string(afterDataJSON)
	changesJSONStr := string(changesJSON)
	businessJSON := string(businessContextJSON)

	if beforeJSON == "null" {
		beforeJSON = "{}"
	}
	if afterJSON == "null" {
		afterJSON = "{}"
	}
	if changesJSONStr == "null" {
		changesJSONStr = "[]"
	}
	if businessJSON == "null" {
		businessJSON = "{}"
	}

	// 按最新 audit_logs 表结构插入
	query := `
		INSERT INTO audit_logs (
			tenant_id,
			event_type,
			resource_type,
			actor_id,
			actor_type,
			action_name,
			request_id,
			resource_id,
			operation_reason,
			request_data,
			response_data,
			changes,
			business_context,
			record_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := aw.db.ExecContext(ctx, query,
		entry.TenantID,
		entry.EventType,
		entry.ResourceType,
		entry.ActorID,
		entry.ActorType,
		entry.ActionName,
		entry.RequestID,
		entry.ResourceID,
		entry.OperationReason,
		beforeJSON,
		afterJSON,
		changesJSONStr,
		businessJSON,
		entry.RecordID,
	)

	if err != nil {
		utils.RecordAuditWrite(err)
		aw.logger.Printf("❌ 审计写入失败: %v", err)
		return fmt.Errorf("审计写入失败: %w", err)
	}

	utils.RecordAuditWrite(nil)
	aw.logger.Printf("✅ 审计已写入: %s/%s/%s (RecordID: %s)",
		entry.EventType, entry.ResourceType, entry.ActionName, entry.RecordID)

	return nil
}

// WriteAuditInTx 在现有事务中写入审计日志
func (aw *AuditWriter) WriteAuditInTx(ctx context.Context, tx *sql.Tx, entry *AuditEntry) error {
	// 跳过空 UPDATE
	if entry.EventType == "UPDATE" && len(entry.Changes) == 0 {
		aw.logger.Printf("⏭️ 跳过空UPDATE审计: RecordID=%s", entry.RecordID)
		return nil
	}

	beforeDataJSON, _ := json.Marshal(entry.BeforeData)
	afterDataJSON, _ := json.Marshal(entry.AfterData)
	changesJSON, _ := json.Marshal(entry.Changes)
	businessContextJSON, _ := json.Marshal(entry.BusinessContext)

	beforeJSON := string(beforeDataJSON)
	afterJSON := string(afterDataJSON)
	changesJSONStr := string(changesJSON)
	businessJSON := string(businessContextJSON)

	if beforeJSON == "null" {
		beforeJSON = "{}"
	}
	if afterJSON == "null" {
		afterJSON = "{}"
	}
	if changesJSONStr == "null" {
		changesJSONStr = "[]"
	}
	if businessJSON == "null" {
		businessJSON = "{}"
	}

	query := `
		INSERT INTO audit_logs (
			tenant_id, event_type, resource_type, actor_id, actor_type,
			action_name, request_id, resource_id, operation_reason,
			request_data, response_data, changes, business_context, record_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := tx.ExecContext(ctx, query,
		entry.TenantID, entry.EventType, entry.ResourceType, entry.ActorID, entry.ActorType,
		entry.ActionName, entry.RequestID, entry.ResourceID, entry.OperationReason,
		beforeJSON, afterJSON, changesJSONStr, businessJSON, entry.RecordID,
	)

	if err != nil {
		utils.RecordAuditWrite(err)
		return fmt.Errorf("事务中审计写入失败: %w", err)
	}

	utils.RecordAuditWrite(nil)
	return nil
}

// CalculateChanges 计算字段变更 - 按 021 迁移的逻辑，跳过系统字段
func (aw *AuditWriter) CalculateChanges(oldData, newData map[string]interface{}) []FieldChange {
	var changes []FieldChange

	// 跳过的系统字段（与 021 迁移保持一致）
	skipFields := map[string]bool{
		"record_id": true, "created_at": true, "updated_at": true,
		"tenant_id": true, "code": true, "path": true,
		"code_path": true, "name_path": true, "hierarchy_depth": true,
	}

	for key, newValue := range newData {
		if skipFields[key] {
			continue
		}

		oldValue, exists := oldData[key]

		// 使用 IS DISTINCT FROM 逻辑判断值变化
		if !exists || !areValuesEqual(oldValue, newValue) {
			changes = append(changes, FieldChange{
				Field:    key,
				OldValue: oldValue,
				NewValue: newValue,
			})
		}
	}

	return changes
}

// areValuesEqual 判断两个值是否相等 - 模拟 PostgreSQL IS DISTINCT FROM 逻辑
func areValuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// WriteOrganizationCreateAudit 写入组织创建审计
func (aw *AuditWriter) WriteOrganizationCreateAudit(ctx context.Context, tenantID uuid.UUID, recordID uuid.UUID, afterData map[string]interface{}, actorID, requestID, reason string) error {
	entry := &AuditEntry{
		TenantID:        tenantID,
		EventType:       "CREATE",
		ResourceType:    "ORGANIZATION",
		ActorID:         actorID,
		ActorType:       "SYSTEM",
		ActionName:      "CREATE_ORGANIZATION",
		RequestID:       requestID,
		ResourceID:      recordID,
		OperationReason: reason,
		BeforeData:      nil,
		AfterData:       afterData,
		Changes:         []FieldChange{}, // CREATE 无变更字段
		BusinessContext: map[string]interface{}{
			"source":    "temporal_timeline_manager",
			"requestId": requestID,
		},
		RecordID: recordID,
	}

	return aw.WriteAudit(ctx, entry)
}

// WriteOrganizationUpdateAudit 写入组织更新审计
func (aw *AuditWriter) WriteOrganizationUpdateAudit(ctx context.Context, tenantID uuid.UUID, recordID uuid.UUID, beforeData, afterData map[string]interface{}, actorID, requestID, reason string) error {
	changes := aw.CalculateChanges(beforeData, afterData)

	entry := &AuditEntry{
		TenantID:        tenantID,
		EventType:       "UPDATE",
		ResourceType:    "ORGANIZATION",
		ActorID:         actorID,
		ActorType:       "SYSTEM",
		ActionName:      "UPDATE_ORGANIZATION",
		RequestID:       requestID,
		ResourceID:      recordID,
		OperationReason: reason,
		BeforeData:      beforeData,
		AfterData:       afterData,
		Changes:         changes,
		BusinessContext: map[string]interface{}{
			"source":    "temporal_timeline_manager",
			"requestId": requestID,
		},
		RecordID: recordID,
	}

	return aw.WriteAudit(ctx, entry)
}

// WriteOrganizationDeleteAudit 写入组织删除审计
func (aw *AuditWriter) WriteOrganizationDeleteAudit(ctx context.Context, tenantID uuid.UUID, recordID uuid.UUID, beforeData map[string]interface{}, actorID, requestID, reason string) error {
	entry := &AuditEntry{
		TenantID:        tenantID,
		EventType:       "DELETE",
		ResourceType:    "ORGANIZATION",
		ActorID:         actorID,
		ActorType:       "SYSTEM",
		ActionName:      "DELETE_ORGANIZATION",
		RequestID:       requestID,
		ResourceID:      recordID,
		OperationReason: reason,
		BeforeData:      beforeData,
		AfterData:       nil,
		Changes:         []FieldChange{}, // DELETE 无变更字段
		BusinessContext: map[string]interface{}{
			"source":    "temporal_timeline_manager",
			"requestId": requestID,
		},
		RecordID: recordID,
	}

	return aw.WriteAudit(ctx, entry)
}
