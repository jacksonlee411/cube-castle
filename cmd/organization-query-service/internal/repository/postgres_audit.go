package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"cube-castle-deployment-test/cmd/organization-query-service/internal/model"
	"cube-castle-deployment-test/internal/auth"
	"github.com/google/uuid"
)

// 审计历史查询 - v4.6.0 基于record_id精确查询 + 租户隔离
func (r *PostgreSQLRepository) GetAuditHistory(ctx context.Context, tenantId uuid.UUID, recordId string, startDate, endDate, operation, userId *string, limit int) ([]model.AuditRecordData, error) {
	start := time.Now()

	recordUUID, err := uuid.Parse(recordId)
	if err != nil {
		r.logger.Printf("[ERROR] 无效的 recordId: %s", recordId)
		return nil, fmt.Errorf("INVALID_RECORD_ID")
	}

	// 构建查询条件 - 基于record_id查询，包含完整变更信息，强制租户隔离
	baseQuery := `
		SELECT
			id as audit_id,
			resource_id as record_id,
			event_type as operation_type,
			actor_id as operated_by_id,
			CASE WHEN business_context->>'actor_name' IS NOT NULL
				THEN business_context->>'actor_name'
				ELSE actor_id
			END as operated_by_name,
		CASE WHEN changes IS NOT NULL
			THEN jsonb_build_object(
				'operationSummary', COALESCE(action_name, event_type, 'UNKNOWN'),
				'totalChanges', jsonb_array_length(changes),
				'keyChanges', changes
			)::text
			ELSE jsonb_build_object(
				'operationSummary', COALESCE(action_name, event_type, 'UNKNOWN'),
				'totalChanges', 0,
				'keyChanges', jsonb_build_array()
			)::text
		END as changes_summary,
		COALESCE(operation_reason, business_context->>'operation_reason', business_context->>'change_reason') as operation_reason,
			timestamp,
			request_data::text as before_data,
			response_data::text as after_data,
			CASE WHEN changes IS NOT NULL AND jsonb_typeof(changes) = 'array'
				THEN (
					SELECT jsonb_agg(DISTINCT elem->>'field')
					FROM jsonb_array_elements(changes) AS elem
					WHERE elem->>'field' IS NOT NULL
				)
				ELSE '[]'::jsonb
			END::text as modified_fields,
			COALESCE(changes, '[]'::jsonb)::text as detailed_changes
		FROM audit_logs
		WHERE tenant_id = $1::uuid AND resource_id::uuid = $2::uuid AND resource_type = 'ORGANIZATION'`

	args := []interface{}{tenantId, recordUUID}
	argIndex := 3

	// 日期范围过滤
	if startDate != nil {
		baseQuery += fmt.Sprintf(" AND timestamp >= $%d::timestamp", argIndex)
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		baseQuery += fmt.Sprintf(" AND timestamp <= $%d::timestamp", argIndex)
		args = append(args, *endDate)
		argIndex++
	}

	// 操作类型过滤
	if operation != nil {
		baseQuery += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, strings.ToUpper(*operation))
		argIndex++
	}

	// 操作人过滤
	if userId != nil {
		baseQuery += fmt.Sprintf(" AND actor_id = $%d", argIndex)
		args = append(args, *userId)
		argIndex++
	}

	// 排序和限制
	finalQuery := baseQuery + fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		r.logger.Printf("[ERROR] 审计历史查询失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var auditRecords []model.AuditRecordData
	if r.auditConfig.LegacyMode {
		auditRecords, err = r.processAuditRowsLegacy(rows)
	} else {
		auditRecords, err = r.processAuditRowsStrict(rows)
	}
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] record_id审计查询完成，返回 %d 条记录，耗时: %v", len(auditRecords), duration)

	return auditRecords, nil
}

func (r *PostgreSQLRepository) processAuditRowsLegacy(rows *sql.Rows) ([]model.AuditRecordData, error) {
	var auditRecords []model.AuditRecordData
	for rows.Next() {
		var record model.AuditRecordData
		var operatedById, operatedByName string
		var beforeData, afterData, modifiedFieldsJSON, detailedChangesJSON sql.NullString

		err := rows.Scan(
			&record.AuditIDField, &record.RecordIDField, &record.OperationTypeField,
			&operatedById, &operatedByName,
			&record.ChangesSummaryField, &record.OperationReasonField, &record.TimestampField,
			&beforeData, &afterData, &modifiedFieldsJSON, &detailedChangesJSON,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描审计记录失败: %v", err)
			return nil, err
		}

		if beforeData.Valid {
			record.BeforeDataField = &beforeData.String
		}
		if afterData.Valid {
			record.AfterDataField = &afterData.String
		}

		if modifiedFieldsJSON.Valid && modifiedFieldsJSON.String != "[]" {
			var modifiedFields []string
			if err := json.Unmarshal([]byte(modifiedFieldsJSON.String), &modifiedFields); err == nil {
				record.ModifiedFieldsField = modifiedFields
			}
		}

		if detailedChangesJSON.Valid && detailedChangesJSON.String != "[]" {
			var changesArray []map[string]interface{}
			if err := json.Unmarshal([]byte(detailedChangesJSON.String), &changesArray); err == nil {
				for _, changeMap := range changesArray {
					fieldChange := model.FieldChangeData{
						FieldField:    fmt.Sprintf("%v", changeMap["field"]),
						OldValueField: changeMap["oldValue"],
						NewValueField: changeMap["newValue"],
						DataTypeField: fmt.Sprintf("%v", changeMap["dataType"]),
					}
					record.ChangesField = append(record.ChangesField, fieldChange)
				}
			}
		}

		record.OperatedByField = model.OperatedByData{
			IDField:   operatedById,
			NameField: operatedByName,
		}

		auditRecords = append(auditRecords, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return auditRecords, nil
}

func (r *PostgreSQLRepository) processAuditRowsStrict(rows *sql.Rows) ([]model.AuditRecordData, error) {
	var auditRecords []model.AuditRecordData
	for rows.Next() {
		var record model.AuditRecordData
		var operatedById, operatedByName string
		var beforeData, afterData, modifiedFieldsJSON, detailedChangesJSON sql.NullString

		record.ModifiedFieldsField = make([]string, 0)
		record.ChangesField = make([]model.FieldChangeData, 0)

		err := rows.Scan(
			&record.AuditIDField, &record.RecordIDField, &record.OperationTypeField,
			&operatedById, &operatedByName,
			&record.ChangesSummaryField, &record.OperationReasonField, &record.TimestampField,
			&beforeData, &afterData, &modifiedFieldsJSON, &detailedChangesJSON,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描审计记录失败: %v", err)
			return nil, err
		}

		if beforeData.Valid {
			record.BeforeDataField = &beforeData.String
		}
		if afterData.Valid {
			record.AfterDataField = &afterData.String
		}

		rawModified := ""
		if modifiedFieldsJSON.Valid {
			rawModified = modifiedFieldsJSON.String
		}
		sanitizedModified, modifiedIssues, modErr := sanitizeModifiedFields(rawModified)
		if modErr == nil {
			record.ModifiedFieldsField = sanitizedModified
		}

		rawChanges := ""
		if detailedChangesJSON.Valid {
			rawChanges = detailedChangesJSON.String
		}
		sanitizedChanges, changeIssues, changeErr := sanitizeChanges(rawChanges)
		if changeErr == nil {
			record.ChangesField = sanitizedChanges
		}

		trimmedChanges := strings.TrimSpace(rawChanges)
		trimmedModified := strings.TrimSpace(rawModified)
		if isEmptyArrayOrNull(trimmedChanges) && isEmptyArrayOrNull(trimmedModified) && len(record.ChangesField) == 0 && len(record.ModifiedFieldsField) == 0 {
			continue
		}

		issues := make([]string, 0, len(modifiedIssues)+len(changeIssues))
		issues = append(issues, modifiedIssues...)
		issues = append(issues, changeIssues...)

		hasHardError := false
		if modErr != nil {
			hasHardError = true
			issues = append(issues, fmt.Sprintf("modified_fields JSON 无效: %v", modErr))
		}
		if changeErr != nil {
			hasHardError = true
			issues = append(issues, fmt.Sprintf("changes JSON 无效: %v", changeErr))
		}

		if len(issues) > 0 {
			r.logger.Printf("[WARN] 审计记录数据异常 audit_id=%s: %s", record.AuditIDField, strings.Join(issues, "; "))
			if r.auditConfig.StrictValidation {
				if hasHardError && !r.auditConfig.AllowFallback {
					return nil, fmt.Errorf("AUDIT_HISTORY_VALIDATION_FAILED")
				}
				if r.registerValidationFailure() {
					return nil, fmt.Errorf("AUDIT_HISTORY_CIRCUIT_OPEN")
				}
			}
		} else if r.auditConfig.StrictValidation {
			r.registerValidationSuccess()
		}

		record.OperatedByField = model.OperatedByData{
			IDField:   operatedById,
			NameField: operatedByName,
		}

		auditRecords = append(auditRecords, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return auditRecords, nil
}

func (r *PostgreSQLRepository) registerValidationFailure() bool {
	count := atomic.AddInt32(&r.validationFailureCount, 1)
	if r.auditConfig.CircuitBreakerThreshold > 0 && count >= r.auditConfig.CircuitBreakerThreshold {
		r.logger.Printf("[ALERT] 审计历史验证失败次数达到阈值 (%d/%d)，触发熔断", count, r.auditConfig.CircuitBreakerThreshold)
		return true
	}
	return false
}

func (r *PostgreSQLRepository) registerValidationSuccess() {
	if atomic.LoadInt32(&r.validationFailureCount) != 0 {
		atomic.StoreInt32(&r.validationFailureCount, 0)
	}
}

// 单条审计记录查询 - v4.6.0
func (r *PostgreSQLRepository) GetAuditLog(ctx context.Context, auditId string) (*model.AuditRecordData, error) {
	start := time.Now()

	query := `
        SELECT 
            id as audit_id, 
            resource_id as record_id, 
            event_type as operation_type,
            actor_id as operated_by_id, 
            CASE WHEN business_context->>'actor_name' IS NOT NULL 
                THEN business_context->>'actor_name' 
                ELSE actor_id 
            END as operated_by_name,
            CASE WHEN changes IS NOT NULL 
                THEN changes::text 
                ELSE '{"operationSummary":"' || action_name || '","totalChanges":0,"keyChanges":[]}' 
            END as changes_summary,
            COALESCE(operation_reason, business_context->>'operation_reason', business_context->>'change_reason') as operation_reason,
            timestamp,
            before_data::text as before_data, 
            after_data::text as after_data
        FROM audit_logs 
        WHERE id = $1::uuid AND resource_type = 'ORGANIZATION' AND tenant_id = $2::uuid
        LIMIT 1`

	tenantID := auth.GetTenantID(ctx)
	if tenantID == "" {
		r.logger.Printf("[AUTH] 缺少租户ID，拒绝单条审计记录查询")
		return nil, fmt.Errorf("TENANT_REQUIRED")
	}

	row := r.db.QueryRowContext(ctx, query, auditId, tenantID)

	var record model.AuditRecordData
	var operatedById, operatedByName string
	var beforeData, afterData sql.NullString

	err := row.Scan(
		&record.AuditIDField, &record.RecordIDField, &record.OperationTypeField,
		&operatedById, &operatedByName,
		&record.ChangesSummaryField, &record.OperationReasonField, &record.TimestampField,
		&beforeData, &afterData,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Printf("[ERROR] 单条审计记录查询失败: %v", err)
		return nil, err
	}

	// 正确处理JSONB字段
	if beforeData.Valid {
		record.BeforeDataField = &beforeData.String
	}
	if afterData.Valid {
		record.AfterDataField = &afterData.String
	}

	// 构建操作人信息
	record.OperatedByField = model.OperatedByData{
		IDField:   operatedById,
		NameField: operatedByName,
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 单条审计记录查询完成，耗时: %v", duration)

	return &record, nil
}

func sanitizeModifiedFields(raw string) ([]string, []string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return make([]string, 0), nil, nil
	}
	if trimmed == "null" {
		return make([]string, 0), []string{"modified_fields 为 null，已替换为空数组"}, nil
	}

	var rawArray []interface{}
	if err := json.Unmarshal([]byte(trimmed), &rawArray); err != nil {
		return make([]string, 0), nil, err
	}

	sanitized := make([]string, 0, len(rawArray))
	issues := make([]string, 0)
	for idx, item := range rawArray {
		if item == nil {
			issues = append(issues, fmt.Sprintf("modified_fields[%d] 为 null，已忽略", idx))
			continue
		}
		switch v := item.(type) {
		case string:
			sanitized = append(sanitized, v)
		default:
			sanitized = append(sanitized, fmt.Sprintf("%v", v))
			issues = append(issues, fmt.Sprintf("modified_fields[%d] 非字符串，已转换", idx))
		}
	}

	return sanitized, issues, nil
}

func sanitizeChanges(raw string) ([]model.FieldChangeData, []string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return make([]model.FieldChangeData, 0), nil, nil
	}
	if trimmed == "null" {
		return make([]model.FieldChangeData, 0), []string{"changes 为 null，已替换为空数组"}, nil
	}

	var rawArray []map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &rawArray); err != nil {
		return make([]model.FieldChangeData, 0), nil, err
	}

	sanitized := make([]model.FieldChangeData, 0, len(rawArray))
	issues := make([]string, 0)
	for idx, entry := range rawArray {
		if entry == nil {
			issues = append(issues, fmt.Sprintf("changes[%d] 为空对象，已跳过", idx))
			continue
		}

		fieldVal, ok := entry["field"]
		if !ok {
			issues = append(issues, fmt.Sprintf("changes[%d] 缺少 field，已跳过", idx))
			continue
		}
		field := strings.TrimSpace(fmt.Sprintf("%v", fieldVal))
		if field == "" {
			issues = append(issues, fmt.Sprintf("changes[%d] field 为空，已跳过", idx))
			continue
		}

		oldVal := entry["oldValue"]
		newVal := entry["newValue"]

		dataType := ""
		if dtVal, ok := entry["dataType"]; ok {
			if dtStr, ok := dtVal.(string); ok {
				dataType = strings.TrimSpace(dtStr)
				if dataType == "" {
					issues = append(issues, fmt.Sprintf("changes[%d] dataType 为空字符串，尝试推断", idx))
				}
			} else {
				issues = append(issues, fmt.Sprintf("changes[%d] dataType 非字符串，尝试推断", idx))
			}
		}
		if strings.EqualFold(dataType, "unknown") {
			if inferred := inferFieldDataType(oldVal, newVal); inferred != "unknown" {
				issues = append(issues, fmt.Sprintf("changes[%d] dataType=unknown，推断为 %s", idx, inferred))
				dataType = inferred
			}
		}

		if dataType == "" {
			inferred := inferFieldDataType(oldVal, newVal)
			if inferred == "unknown" {
				issues = append(issues, fmt.Sprintf("changes[%d] 缺少 dataType，使用 unknown", idx))
			} else {
				issues = append(issues, fmt.Sprintf("changes[%d] 缺少 dataType，推断为 %s", idx, inferred))
			}
			dataType = inferred
		}

		fieldChange := model.FieldChangeData{
			FieldField:    field,
			DataTypeField: dataType,
			OldValueField: normalizeChangeValue(oldVal),
			NewValueField: normalizeChangeValue(newVal),
		}
		sanitized = append(sanitized, fieldChange)
	}

	return sanitized, issues, nil
}

func inferFieldDataType(oldVal, newVal interface{}) string {
	value := firstNonNil(newVal, oldVal)
	if value == nil {
		return "unknown"
	}
	switch value.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64, json.Number:
		return "number"
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	default:
		return "unknown"
	}
}

func firstNonNil(values ...interface{}) interface{} {
	for _, val := range values {
		if val != nil {
			return val
		}
	}
	return nil
}

func isEmptyArrayOrNull(raw string) bool {
	switch raw {
	case "", "[]", "null":
		return true
	default:
		return false
	}
}

func normalizeChangeValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		return v.String()
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(bytes)
	}
}
