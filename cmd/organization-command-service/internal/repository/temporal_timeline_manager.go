package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type TemporalTimelineManager struct {
	db     *sql.DB
	logger *log.Logger
}

func NewTemporalTimelineManager(db *sql.DB, logger *log.Logger) *TemporalTimelineManager {
	return &TemporalTimelineManager{db: db, logger: logger}
}

type TimelineVersion struct {
	RecordID      uuid.UUID  `json:"recordId"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	EffectiveDate time.Time  `json:"effectiveDate"`
	EndDate       *time.Time `json:"endDate"`
	IsCurrent     bool       `json:"isCurrent"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func (tm *TemporalTimelineManager) RecalculateTimeline(ctx context.Context, tenantID uuid.UUID, code string) (*[]TimelineVersion, error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tm.logger.Printf("🔄 开始全链重算: tenant=%s, code=%s", tenantID, code)

	versions, err := tm.RecalculateTimelineInTx(ctx, tx, tenantID, code)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	tm.logger.Printf("✅ 全链重算完成: %s, 版本数=%d", code, len(*versions))
	return versions, nil
}

func (tm *TemporalTimelineManager) RecalculateTimelineInTx(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*[]TimelineVersion, error) {
	query := `
		SELECT record_id, code, name, effective_date, end_date, is_current, status, created_at
		FROM organization_units 
		WHERE tenant_id = $1 
		  AND code = $2 
		  AND status != 'DELETED' 
		ORDER BY effective_date ASC`

	rows, err := tx.QueryContext(ctx, query, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("查询版本列表失败: %w", err)
	}
	defer rows.Close()

	var versions []TimelineVersion
	for rows.Next() {
		var v TimelineVersion
		if err := rows.Scan(&v.RecordID, &v.Code, &v.Name, &v.EffectiveDate, &v.EndDate, &v.IsCurrent, &v.Status, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描版本记录失败: %w", err)
		}
		versions = append(versions, v)
	}

	if len(versions) == 0 {
		return &[]TimelineVersion{}, nil
	}

	clearCurrentQuery := `
		UPDATE organization_units 
		SET is_current = false, updated_at = NOW()
		WHERE tenant_id = $1 AND code = $2 
		  AND status != 'DELETED'`
	if _, err := tx.ExecContext(ctx, clearCurrentQuery, tenantID, code); err != nil {
		return nil, fmt.Errorf("清除当前状态标记失败: %w", err)
	}

	today := time.Now().Truncate(24 * time.Hour)
	var currentVersionRecordID *uuid.UUID
	var latestEffectiveDate *time.Time

	for i := range versions {
		var endDate *time.Time
		if i < len(versions)-1 {
			nextEffectiveDate := versions[i+1].EffectiveDate
			calculatedEnd := nextEffectiveDate.AddDate(0, 0, -1)
			endDate = &calculatedEnd
		}

		updateQuery := `
			UPDATE organization_units 
			SET end_date = $3,
				updated_at = NOW()
			WHERE record_id = $1 AND tenant_id = $2`
		if _, err := tx.ExecContext(ctx, updateQuery, versions[i].RecordID, tenantID, endDate); err != nil {
			return nil, fmt.Errorf("更新版本边界失败: %w", err)
		}

		versions[i].EndDate = endDate

		if !versions[i].EffectiveDate.After(today) {
			if latestEffectiveDate == nil || versions[i].EffectiveDate.After(*latestEffectiveDate) {
				latestEffectiveDate = &versions[i].EffectiveDate
				recordID := versions[i].RecordID
				currentVersionRecordID = &recordID
			}
		}
	}

	if currentVersionRecordID != nil {
		setCurrentQuery := `
			UPDATE organization_units 
			SET is_current = true, updated_at = NOW()
			WHERE record_id = $1 AND tenant_id = $2`
		if _, err := tx.ExecContext(ctx, setCurrentQuery, *currentVersionRecordID, tenantID); err != nil {
			return nil, fmt.Errorf("设置当前版本标记失败: %w", err)
		}

		for i := range versions {
			versions[i].IsCurrent = versions[i].RecordID == *currentVersionRecordID
		}
	}

	return &versions, nil
}
