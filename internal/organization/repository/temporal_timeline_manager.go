package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

type TemporalTimelineManager struct {
	db     *sql.DB
	logger pkglogger.Logger
}

func NewTemporalTimelineManager(db *sql.DB, baseLogger pkglogger.Logger) *TemporalTimelineManager {
	return &TemporalTimelineManager{
		db:     db,
		logger: scopedLogger(baseLogger, "organization", "TemporalTimelineManager", nil),
	}
}

type TimelineVersion struct {
	RecordID      uuid.UUID  `json:"recordId"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	UnitType      string     `json:"unitType"`
	Status        string     `json:"status"`
	Level         int        `json:"level"`
	CodePath      string     `json:"codePath"`
	NamePath      string     `json:"namePath"`
	ParentCode    *string    `json:"parentCode"`
	Description   *string    `json:"description"`
	SortOrder     *int       `json:"sortOrder"`
	EffectiveDate time.Time  `json:"effectiveDate"`
	EndDate       *time.Time `json:"endDate"`
	IsCurrent     bool       `json:"isCurrent"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

func (tm *TemporalTimelineManager) RecalculateTimeline(ctx context.Context, tenantID uuid.UUID, code string) (*[]TimelineVersion, error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tm.logger.Infof("开始全链重算: tenant=%s, code=%s", tenantID, code)

	versions, err := tm.RecalculateTimelineInTx(ctx, tx, tenantID, code)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	tm.logger.Infof("全链重算完成: %s, 版本数=%d", code, len(*versions))
	return versions, nil
}

func (tm *TemporalTimelineManager) RecalculateTimelineInTx(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*[]TimelineVersion, error) {
	query := `
		SELECT 
			record_id, code, name, unit_type, status, level, code_path, name_path,
			parent_code, description, sort_order, effective_date, end_date, is_current,
			created_at, updated_at
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
		var (
			v           TimelineVersion
			parentCode  sql.NullString
			description sql.NullString
			sortOrder   sql.NullInt64
			endDate     sql.NullTime
		)

		if err := rows.Scan(
			&v.RecordID,
			&v.Code,
			&v.Name,
			&v.UnitType,
			&v.Status,
			&v.Level,
			&v.CodePath,
			&v.NamePath,
			&parentCode,
			&description,
			&sortOrder,
			&v.EffectiveDate,
			&endDate,
			&v.IsCurrent,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描版本记录失败: %w", err)
		}

		if endDate.Valid {
			ed := endDate.Time
			v.EndDate = &ed
		}
		if parentCode.Valid {
			pc := parentCode.String
			v.ParentCode = &pc
		}
		if description.Valid {
			desc := description.String
			v.Description = &desc
		}
		if sortOrder.Valid {
			value := int(sortOrder.Int64)
			v.SortOrder = &value
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

	today := time.Now().UTC().Truncate(24 * time.Hour)
	var currentVersionRecordID *uuid.UUID
	var latestEffectiveDate time.Time
	var hasLatest bool

	for i := range versions {
		effectiveUTC := versions[i].EffectiveDate.In(time.UTC)
		versions[i].EffectiveDate = effectiveUTC
		var endDate *time.Time
		if i < len(versions)-1 {
			nextEffectiveDateUTC := versions[i+1].EffectiveDate.In(time.UTC)
			calculatedEnd := nextEffectiveDateUTC.AddDate(0, 0, -1)
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
			if !hasLatest || versions[i].EffectiveDate.After(latestEffectiveDate) {
				latestEffectiveDate = versions[i].EffectiveDate
				recordID := versions[i].RecordID
				currentVersionRecordID = &recordID
				hasLatest = true
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
