package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (tm *TemporalTimelineManager) SuspendOrganization(ctx context.Context, tenantID uuid.UUID, code string, effectiveDate time.Time, operationReason string) (*[]TimelineVersion, error) {
	return tm.changeOrganizationStatus(ctx, tenantID, code, "INACTIVE", "SUSPEND", effectiveDate, operationReason)
}

func (tm *TemporalTimelineManager) ActivateOrganization(ctx context.Context, tenantID uuid.UUID, code string, effectiveDate time.Time, operationReason string) (*[]TimelineVersion, error) {
	return tm.changeOrganizationStatus(ctx, tenantID, code, "ACTIVE", "REACTIVATE", effectiveDate, operationReason)
}

func (tm *TemporalTimelineManager) changeOrganizationStatus(ctx context.Context, tenantID uuid.UUID, code, newStatus, operationType string, effectiveDate time.Time, operationReason string) (*[]TimelineVersion, error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tm.logger.Printf("🔄 开始%s组织: Code=%s, 生效日期=%s, 新状态=%s", operationType, code, effectiveDate.Format("2006-01-02"), newStatus)

	var currentOrg struct {
		RecordID      string
		TenantID      uuid.UUID
		Code          string
		ParentCode    *string
		Name          string
		UnitType      string
		Status        string
		Level         int
		Path          string
		CodePath      string
		NamePath      string
		SortOrder     int
		Description   string
		EffectiveDate time.Time
		IsCurrent     bool
		ChangeReason  *string
		CreatedAt     time.Time
		UpdatedAt     time.Time
	}

	row := tx.QueryRowContext(ctx, `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status, level, path,
		       code_path, name_path, sort_order, description, effective_date, is_current, change_reason,
		       created_at, updated_at
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND is_current = true 
		  AND status != 'DELETED'
		FOR UPDATE`, tenantID, code)

	if err := row.Scan(
		&currentOrg.RecordID, &currentOrg.TenantID, &currentOrg.Code, &currentOrg.ParentCode, &currentOrg.Name,
		&currentOrg.UnitType, &currentOrg.Status, &currentOrg.Level, &currentOrg.Path, &currentOrg.CodePath,
		&currentOrg.NamePath, &currentOrg.SortOrder,
		&currentOrg.Description, &currentOrg.EffectiveDate, &currentOrg.IsCurrent,
		&currentOrg.ChangeReason, &currentOrg.CreatedAt, &currentOrg.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在或无当前版本: %s", code)
		}
		return nil, fmt.Errorf("查询组织当前版本失败: %w", err)
	}

	if currentOrg.Status == newStatus {
		tm.logger.Printf("💡 组织%s状态已经是%s，幂等操作跳过", code, newStatus)
		return tm.RecalculateTimelineInTx(ctx, tx, tenantID, code)
	}

	var conflictCount int
	conflictQuery := `
		SELECT COUNT(*) 
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND effective_date = $3 
		  AND status != 'DELETED'`
	effectiveDateUTC := effectiveDate.In(time.UTC)
	if err := tx.QueryRowContext(ctx, conflictQuery, tenantID, code, effectiveDateUTC).Scan(&conflictCount); err != nil {
		return nil, fmt.Errorf("冲突校验查询失败: %w", err)
	}
	if conflictCount > 0 {
		return nil, fmt.Errorf("TEMPORAL_POINT_CONFLICT: 生效日期 %s 与现有版本冲突", effectiveDateUTC.Format("2006-01-02"))
	}

	nowUTC := time.Now().UTC()
	newRecordID := uuid.New()
	isFuture := effectiveDateUTC.After(nowUTC.Truncate(24 * time.Hour))

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO organization_units (
			record_id, tenant_id, code, parent_code, name, unit_type, status,
			level, path, code_path, name_path, sort_order, description, effective_date, end_date,
			is_current, change_reason, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NULL,
			false, $15, $16, $16
		)`, newRecordID, currentOrg.TenantID, currentOrg.Code, currentOrg.ParentCode, currentOrg.Name,
		currentOrg.UnitType, newStatus, currentOrg.Level, currentOrg.Path, currentOrg.CodePath, currentOrg.NamePath,
		currentOrg.SortOrder, currentOrg.Description, effectiveDateUTC, operationReason, nowUTC); err != nil {
		return nil, fmt.Errorf("插入%s版本失败: %w", operationType, err)
	}

	timeline, err := tm.RecalculateTimelineInTx(ctx, tx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("时间轴重算失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("事务提交失败: %w", err)
	}

	action := "暂停"
	if operationType == "REACTIVATE" {
		action = "激活"
	}

	if isFuture {
		tm.logger.Printf("✅ 组织%s成功（计划生效）: %s → %s, 生效日期=%s", action, code, newStatus, effectiveDateUTC.Format("2006-01-02"))
	} else {
		tm.logger.Printf("✅ 组织%s成功（即时生效）: %s → %s, 时间轴已重算", action, code, newStatus)
	}

	return timeline, nil
}
