package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"cube-castle/internal/types"
)

func (tm *TemporalTimelineManager) InsertVersion(ctx context.Context, org *types.Organization) (*TimelineVersion, error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tenantID, err := uuid.Parse(org.TenantID)
	if err != nil {
		return nil, fmt.Errorf("无效的租户ID: %w", err)
	}

	effectiveDate := time.Date(org.EffectiveDate.Year(), org.EffectiveDate.Month(), org.EffectiveDate.Day(), 0, 0, 0, 0, time.UTC)

	tm.logger.Printf("🔄 插入版本: %s, 生效日期: %s", org.Code, effectiveDate.Format("2006-01-02"))

	adjacentQuery := `
		SELECT record_id, effective_date, end_date, is_current
		FROM organization_units 
		WHERE tenant_id = $1 
		  AND code = $2
		  AND status != 'DELETED' 
		ORDER BY effective_date
		FOR UPDATE`

	rows, err := tx.QueryContext(ctx, adjacentQuery, tenantID, org.Code)
	if err != nil {
		return nil, fmt.Errorf("查询相邻版本失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var recordID uuid.UUID
		var existingEffective time.Time
		var existingEnd *time.Time
		var existingCurrent bool

		if err := rows.Scan(&recordID, &existingEffective, &existingEnd, &existingCurrent); err != nil {
			return nil, fmt.Errorf("扫描相邻版本失败: %w", err)
		}

		if existingEffective.Equal(effectiveDate) {
			return nil, fmt.Errorf("TEMPORAL_POINT_CONFLICT: 生效日期 %s 已存在", effectiveDate.Format("2006-01-02"))
		}
	}

	insertQuery := `
	INSERT INTO organization_units (
		tenant_id, code, parent_code, name, unit_type, status,
		level, code_path, name_path, sort_order, description, effective_date,
		is_current, change_reason, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, false, $13, NOW(), NOW())
	RETURNING record_id, created_at`

	var newRecordID uuid.UUID
	var createdAt time.Time

	if err := tx.QueryRowContext(ctx, insertQuery,
		tenantID, org.Code, org.ParentCode, org.Name, org.UnitType, "ACTIVE",
		org.Level, org.CodePath, org.NamePath, org.SortOrder, org.Description, effectiveDate,
		org.ChangeReason,
	).Scan(&newRecordID, &createdAt); err != nil {
		return nil, fmt.Errorf("插入新版本失败: %w", err)
	}

	if _, err := tm.RecalculateTimelineInTx(ctx, tx, tenantID, org.Code); err != nil {
		return nil, fmt.Errorf("全链重算失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	tm.logger.Printf("✅ 版本插入成功: RecordID=%s", newRecordID)
	return &TimelineVersion{
		RecordID:      newRecordID,
		Code:          org.Code,
		Name:          org.Name,
		EffectiveDate: effectiveDate,
		Status:        "ACTIVE",
		CreatedAt:     createdAt,
	}, nil
}
