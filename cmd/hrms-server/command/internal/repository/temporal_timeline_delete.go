package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func (tm *TemporalTimelineManager) DeleteVersion(ctx context.Context, tenantID uuid.UUID, recordID uuid.UUID) (*[]TimelineVersion, error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tm.logger.Infof("删除版本: RecordID=%s", recordID)

	var code string
	versionQuery := `
		SELECT code FROM organization_units 
		WHERE record_id = $1 AND tenant_id = $2`
	if err := tx.QueryRowContext(ctx, versionQuery, recordID, tenantID).Scan(&code); err != nil {
		return nil, fmt.Errorf("查询版本信息失败: %w", err)
	}

	deleteQuery := `
		UPDATE organization_units 
		SET status = 'DELETED',
			deleted_at = NOW(),
			is_current = false,
			updated_at = NOW()
		WHERE record_id = $1 AND tenant_id = $2`
	if _, err := tx.ExecContext(ctx, deleteQuery, recordID, tenantID); err != nil {
		return nil, fmt.Errorf("软删除版本失败: %w", err)
	}

	timeline, err := tm.RecalculateTimelineInTx(ctx, tx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("全链重算失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	tm.logger.Infof("版本删除成功，剩余版本: %d", len(*timeline))
	return timeline, nil
}
