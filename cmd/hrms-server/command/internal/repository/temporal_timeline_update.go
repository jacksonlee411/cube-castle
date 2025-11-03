package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"cube-castle/internal/types"
)

func (tm *TemporalTimelineManager) UpdateVersionEffectiveDate(ctx context.Context, tenantID uuid.UUID, recordID uuid.UUID, newEffectiveDate time.Time, operationReason string) (*[]TimelineVersion, error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tm.logger.Printf("🔄 开始修改版本生效日期: RecordID=%s, 新日期=%s", recordID.String(), newEffectiveDate.Format("2006-01-02"))

	var org types.Organization
	row := tx.QueryRowContext(ctx, `
	SELECT tenant_id, code, parent_code, name, unit_type, status, level, code_path, name_path, sort_order,
	       description, effective_date, is_current, change_reason, created_at, updated_at
	FROM organization_units 
	WHERE record_id = $1 AND status != 'DELETED'
	FOR UPDATE`, recordID)

	if err := row.Scan(
		&org.TenantID, &org.Code, &org.ParentCode, &org.Name, &org.UnitType,
		&org.Status, &org.Level, &org.CodePath, &org.NamePath, &org.SortOrder, &org.Description,
		&org.EffectiveDate, &org.IsCurrent, &org.ChangeReason,
		&org.CreatedAt, &org.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("版本不存在或已被删除: %s", recordID.String())
		}
		return nil, fmt.Errorf("查询版本信息失败: %w", err)
	}

	parsedTenant, err := uuid.Parse(org.TenantID)
	if err != nil || parsedTenant != tenantID {
		return nil, fmt.Errorf("版本不属于指定租户")
	}

	var conflictCount int
	conflictQuery := `
		SELECT COUNT(*) 
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND effective_date = $3 
		  AND record_id != $4 AND status != 'DELETED'`
	if err := tx.QueryRowContext(ctx, conflictQuery, tenantID, org.Code, newEffectiveDate, recordID).Scan(&conflictCount); err != nil {
		return nil, fmt.Errorf("冲突校验查询失败: %w", err)
	}
	if conflictCount > 0 {
		return nil, fmt.Errorf("TEMPORAL_POINT_CONFLICT: 新生效日期 %s 与现有版本冲突", newEffectiveDate.Format("2006-01-02"))
	}

	now := time.Now()
	if _, err := tx.ExecContext(ctx, `
		UPDATE organization_units 
		SET status = 'DELETED', deleted_at = $3, updated_at = $3
		WHERE record_id = $1 AND tenant_id = $2`, recordID, tenantID, now); err != nil {
		return nil, fmt.Errorf("删除旧版本失败: %w", err)
	}

	newRecordID := uuid.New()
	org.RecordID = newRecordID.String()
	org.EffectiveDate = types.NewDateFromTime(newEffectiveDate)
	org.ChangeReason = &operationReason
	org.CreatedAt = now
	org.UpdatedAt = now
	org.IsCurrent = false

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO organization_units (
			record_id, tenant_id, code, parent_code, name, unit_type, status,
			level, code_path, name_path, sort_order, description, effective_date, end_date,
			is_current, change_reason, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NULL,
			false, $14, $15, $15
		)`, newRecordID, org.TenantID, org.Code, org.ParentCode, org.Name, org.UnitType,
		org.Status, org.Level, org.CodePath, org.NamePath, org.SortOrder, org.Description,
		newEffectiveDate, operationReason, now); err != nil {
		return nil, fmt.Errorf("插入新版本失败: %w", err)
	}

	timeline, err := tm.RecalculateTimelineInTx(ctx, tx, tenantID, org.Code)
	if err != nil {
		return nil, fmt.Errorf("时间轴重算失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("事务提交失败: %w", err)
	}

	tm.logger.Printf("✅ 版本生效日期修改成功: %s → %s", recordID.String(), newEffectiveDate.Format("2006-01-02"))
	return timeline, nil
}
