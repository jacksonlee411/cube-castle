package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"organization-command-service/internal/types"
)

// TemporalTimelineManager 时态时间轴管理器
// 实现 docs/architecture/temporal-timeline-consistency-guide.md v1.0 中的全链重算算法
type TemporalTimelineManager struct {
	db     *sql.DB
	logger *log.Logger
}

func NewTemporalTimelineManager(db *sql.DB, logger *log.Logger) *TemporalTimelineManager {
	return &TemporalTimelineManager{
		db:     db,
		logger: logger,
	}
}

// TimelineVersion 时间轴版本数据结构
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

// RecalculateTimeline 全链重算算法 - 核心实现
// 输入：同一 (tenant_id, code) 的"非删除版本"，按 effective_date 升序
// 输出：无断档、无重叠、尾部开放、单当前
func (tm *TemporalTimelineManager) RecalculateTimeline(ctx context.Context, tenantID uuid.UUID, code string) (*[]TimelineVersion, error) {
	// 开始数据库事务
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tm.logger.Printf("🔄 开始全链重算: tenant=%s, code=%s", tenantID, code)

	// 第一步：获取所有非删除版本，按 effective_date 升序排列
	query := `
		SELECT record_id, code, name, effective_date, end_date, is_current, status, created_at
		FROM organization_units 
		WHERE tenant_id = $1 
		  AND code = $2 
		  AND status != 'DELETED' 
		  AND deleted_at IS NULL
		ORDER BY effective_date ASC
		FOR UPDATE
	`

	rows, err := tx.QueryContext(ctx, query, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("查询版本列表失败: %w", err)
	}
	defer rows.Close()

	var versions []TimelineVersion
	for rows.Next() {
		var v TimelineVersion
		err := rows.Scan(&v.RecordID, &v.Code, &v.Name, &v.EffectiveDate, &v.EndDate, &v.IsCurrent, &v.Status, &v.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描版本记录失败: %w", err)
		}
		versions = append(versions, v)
	}

	if len(versions) == 0 {
		tm.logger.Printf("⚠️ 未找到有效版本: %s", code)
		return &[]TimelineVersion{}, nil
	}

	tm.logger.Printf("📋 找到 %d 个版本进行重算", len(versions))

	// 第二步：清空该 code 所有 is_current 标记
	clearCurrentQuery := `
		UPDATE organization_units 
		SET is_current = false,
			updated_at = NOW()
		WHERE tenant_id = $1 
		  AND code = $2
	`
	_, err = tx.ExecContext(ctx, clearCurrentQuery, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("清除当前状态标记失败: %w", err)
	}

	// 第三步：重新计算时态边界
	today := time.Now().Truncate(24 * time.Hour)
	var currentVersionRecordID *uuid.UUID
	var latestEffectiveDate *time.Time

	for i := 0; i < len(versions); i++ {
		var endDate *time.Time
		
		// 计算结束日期：如果有下一个版本，结束日期为下一版本生效日期的前一天
		if i < len(versions)-1 {
			nextEffectiveDate := versions[i+1].EffectiveDate
			calculatedEndDate := nextEffectiveDate.AddDate(0, 0, -1)
			endDate = &calculatedEndDate
		}
		// 最后一个版本：结束日期为 NULL (尾部开放)
		
		// 更新版本的结束日期
		updateQuery := `
			UPDATE organization_units 
			SET end_date = $3,
				updated_at = NOW()
			WHERE record_id = $1 AND tenant_id = $2
		`
		_, err = tx.ExecContext(ctx, updateQuery, versions[i].RecordID, tenantID, endDate)
		if err != nil {
			return nil, fmt.Errorf("更新版本边界失败 (RecordID: %s): %w", versions[i].RecordID, err)
		}

		// 更新内存中的版本数据
		versions[i].EndDate = endDate

		// 寻找当前版本：生效日期 <= 今天的版本中，生效日最大的一条
		if !versions[i].EffectiveDate.After(today) {
			if latestEffectiveDate == nil || versions[i].EffectiveDate.After(*latestEffectiveDate) {
				latestEffectiveDate = &versions[i].EffectiveDate
				currentVersionRecordID = &versions[i].RecordID
			}
		}
	}

	// 第四步：设置当前版本标记
	if currentVersionRecordID != nil {
		setCurrentQuery := `
			UPDATE organization_units 
			SET is_current = true,
				updated_at = NOW()
			WHERE record_id = $1 AND tenant_id = $2
		`
		_, err = tx.ExecContext(ctx, setCurrentQuery, *currentVersionRecordID, tenantID)
		if err != nil {
			return nil, fmt.Errorf("设置当前版本标记失败: %w", err)
		}

		// 更新内存中的当前版本标记
		for i := range versions {
			if versions[i].RecordID == *currentVersionRecordID {
				versions[i].IsCurrent = true
			} else {
				versions[i].IsCurrent = false
			}
		}

		tm.logger.Printf("✅ 设置当前版本: RecordID=%s, 生效日期=%s", *currentVersionRecordID, latestEffectiveDate.Format("2006-01-02"))
	} else {
		tm.logger.Printf("⚠️ 无当前版本: 所有版本都是未来版本")
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	tm.logger.Printf("✅ 全链重算完成: %s, 版本数=%d, 当前版本=%v", code, len(versions), currentVersionRecordID != nil)

	return &versions, nil
}

// InsertVersion 插入中间版本 - 实现文档第50-62行的逻辑
func (tm *TemporalTimelineManager) InsertVersion(ctx context.Context, org *types.Organization) (*TimelineVersion, error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tenantID, err := uuid.Parse(org.TenantID)
	if err != nil {
		return nil, fmt.Errorf("无效的租户ID: %w", err)
	}

	effectiveDate := time.Date(
		org.EffectiveDate.Year(), org.EffectiveDate.Month(), org.EffectiveDate.Day(),
		0, 0, 0, 0, time.UTC,
	)

	tm.logger.Printf("🔄 插入版本: %s, 生效日期: %s", org.Code, effectiveDate.Format("2006-01-02"))

	// 第一步：读取相邻版本并 FOR UPDATE 锁定
	adjacentQuery := `
		SELECT record_id, effective_date, end_date, is_current
		FROM organization_units 
		WHERE tenant_id = $1 
		  AND code = $2
		  AND status != 'DELETED' 
		  AND deleted_at IS NULL
		ORDER BY effective_date
		FOR UPDATE
	`

	rows, err := tx.QueryContext(ctx, adjacentQuery, tenantID, org.Code)
	if err != nil {
		return nil, fmt.Errorf("查询相邻版本失败: %w", err)
	}
	defer rows.Close()

	// 第二步：预检冲突
	for rows.Next() {
		var recordID uuid.UUID
		var existingEffective time.Time
		var existingEnd *time.Time
		var existingCurrent bool

		err := rows.Scan(&recordID, &existingEffective, &existingEnd, &existingCurrent)
		if err != nil {
			return nil, fmt.Errorf("扫描相邻版本失败: %w", err)
		}

		// 检查时点冲突
		if existingEffective.Equal(effectiveDate) {
			return nil, fmt.Errorf("TEMPORAL_POINT_CONFLICT: 生效日期 %s 已存在", effectiveDate.Format("2006-01-02"))
		}
	}

	// 第三步：插入新版本
	insertQuery := `
		INSERT INTO organization_units (
			tenant_id, code, parent_code, name, unit_type, status, 
			level, path, code_path, name_path, sort_order, description, effective_date,
			is_current, change_reason, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, false, $14, NOW(), NOW())
		RETURNING record_id, created_at
	`

	var newRecordID uuid.UUID
	var createdAt time.Time

	err = tx.QueryRowContext(ctx, insertQuery,
		tenantID, org.Code, org.ParentCode, org.Name, org.UnitType, "ACTIVE",
		org.Level, org.Path, org.CodePath, org.NamePath, org.SortOrder, org.Description, effectiveDate,
		org.ChangeReason,
	).Scan(&newRecordID, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("插入新版本失败: %w", err)
	}

	// 第四步：执行全链重算
	_, err = tm.RecalculateTimelineInTx(ctx, tx, tenantID, org.Code)
	if err != nil {
		return nil, fmt.Errorf("全链重算失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 构造返回结果
	result := &TimelineVersion{
		RecordID:      newRecordID,
		Code:          org.Code,
		Name:          org.Name,
		EffectiveDate: effectiveDate,
		Status:        "ACTIVE",
		CreatedAt:     createdAt,
	}

	tm.logger.Printf("✅ 版本插入成功: RecordID=%s", newRecordID)
	return result, nil
}

// DeleteVersion 删除版本 - 实现文档第64-79行的逻辑
func (tm *TemporalTimelineManager) DeleteVersion(ctx context.Context, tenantID uuid.UUID, recordID uuid.UUID) (*[]TimelineVersion, error) {
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tm.logger.Printf("🗑️ 删除版本: RecordID=%s", recordID)

	// 第一步：获取要删除的版本信息
	var code string
	versionQuery := `
		SELECT code FROM organization_units 
		WHERE record_id = $1 AND tenant_id = $2
	`
	err = tx.QueryRowContext(ctx, versionQuery, recordID, tenantID).Scan(&code)
	if err != nil {
		return nil, fmt.Errorf("查询版本信息失败: %w", err)
	}

	// 第二步：软删除版本（标记为已删除）
	deleteQuery := `
		UPDATE organization_units 
		SET status = 'DELETED',
			deleted_at = NOW(),
			is_current = false,
			updated_at = NOW()
		WHERE record_id = $1 AND tenant_id = $2
	`
	_, err = tx.ExecContext(ctx, deleteQuery, recordID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("软删除版本失败: %w", err)
	}

	// 第三步：执行全链重算，重新计算剩余版本的时间边界
	timeline, err := tm.RecalculateTimelineInTx(ctx, tx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("全链重算失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	tm.logger.Printf("✅ 版本删除成功，剩余版本: %d", len(*timeline))
	return timeline, nil
}

// RecalculateTimelineInTx 在现有事务中执行全链重算
func (tm *TemporalTimelineManager) RecalculateTimelineInTx(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*[]TimelineVersion, error) {
	// 获取所有非删除版本
	query := `
		SELECT record_id, code, name, effective_date, end_date, is_current, status, created_at
		FROM organization_units 
		WHERE tenant_id = $1 
		  AND code = $2 
		  AND status != 'DELETED' 
		  AND deleted_at IS NULL
		ORDER BY effective_date ASC
	`

	rows, err := tx.QueryContext(ctx, query, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("查询版本列表失败: %w", err)
	}
	defer rows.Close()

	var versions []TimelineVersion
	for rows.Next() {
		var v TimelineVersion
		err := rows.Scan(&v.RecordID, &v.Code, &v.Name, &v.EffectiveDate, &v.EndDate, &v.IsCurrent, &v.Status, &v.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描版本记录失败: %w", err)
		}
		versions = append(versions, v)
	}

	if len(versions) == 0 {
		return &[]TimelineVersion{}, nil
	}

	// 清空当前状态标记 - 只清理非DELETED状态的记录，避免触发器冲突
	clearCurrentQuery := `
		UPDATE organization_units 
		SET is_current = false, updated_at = NOW()
		WHERE tenant_id = $1 AND code = $2 
		  AND status != 'DELETED' AND deleted_at IS NULL
	`
	_, err = tx.ExecContext(ctx, clearCurrentQuery, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("清除当前状态标记失败: %w", err)
	}

	// 重新计算边界
	today := time.Now().Truncate(24 * time.Hour)
	var currentVersionRecordID *uuid.UUID
	var latestEffectiveDate *time.Time

	for i := 0; i < len(versions); i++ {
		var endDate *time.Time
		
		if i < len(versions)-1 {
			nextEffectiveDate := versions[i+1].EffectiveDate
			calculatedEndDate := nextEffectiveDate.AddDate(0, 0, -1)
			endDate = &calculatedEndDate
		}
		
		updateQuery := `
			UPDATE organization_units 
			SET end_date = $3,
				updated_at = NOW()
			WHERE record_id = $1 AND tenant_id = $2
		`
		_, err = tx.ExecContext(ctx, updateQuery, versions[i].RecordID, tenantID, endDate)
		if err != nil {
			return nil, fmt.Errorf("更新版本边界失败: %w", err)
		}

		versions[i].EndDate = endDate

		// 寻找当前版本
		if !versions[i].EffectiveDate.After(today) {
			if latestEffectiveDate == nil || versions[i].EffectiveDate.After(*latestEffectiveDate) {
				latestEffectiveDate = &versions[i].EffectiveDate
				currentVersionRecordID = &versions[i].RecordID
			}
		}
	}

	// 设置当前版本
	if currentVersionRecordID != nil {
		setCurrentQuery := `
			UPDATE organization_units 
			SET is_current = true, updated_at = NOW()
			WHERE record_id = $1 AND tenant_id = $2
		`
		_, err = tx.ExecContext(ctx, setCurrentQuery, *currentVersionRecordID, tenantID)
		if err != nil {
			return nil, fmt.Errorf("设置当前版本标记失败: %w", err)
		}

		for i := range versions {
			if versions[i].RecordID == *currentVersionRecordID {
				versions[i].IsCurrent = true
			} else {
				versions[i].IsCurrent = false
			}
		}
	}

	return &versions, nil
}

// UpdateVersionEffectiveDate 修改版本生效日期 - 实现第三大核心场景
// 语义：等价于"删除旧版本 + 插入新版本"（单事务原子化）
func (tm *TemporalTimelineManager) UpdateVersionEffectiveDate(ctx context.Context, tenantID uuid.UUID, recordID uuid.UUID, newEffectiveDate time.Time, operationReason string) (*[]TimelineVersion, error) {
	// 开始数据库事务
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tm.logger.Printf("🔄 开始修改版本生效日期: RecordID=%s, 新日期=%s", recordID.String(), newEffectiveDate.Format("2006-01-02"))

	// 1. 获取要修改的版本信息
	var org types.Organization
    row := tx.QueryRowContext(ctx, `
        SELECT tenant_id, code, parent_code, name, unit_type, status, level, path, sort_order, 
               description, effective_date, is_current, change_reason, created_at, updated_at
        FROM organization_units 
        WHERE record_id = $1 AND status != 'DELETED' AND deleted_at IS NULL
        FOR UPDATE
    `, recordID)

    err = row.Scan(
        &org.TenantID, &org.Code, &org.ParentCode, &org.Name, &org.UnitType,
        &org.Status, &org.Level, &org.Path, &org.SortOrder, &org.Description,
        &org.EffectiveDate, &org.IsCurrent, &org.ChangeReason,
        &org.CreatedAt, &org.UpdatedAt,
    )
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("版本不存在或已被删除: %s", recordID.String())
		}
		return nil, fmt.Errorf("查询版本信息失败: %w", err)
	}
	org.RecordID = recordID.String()

	// 验证租户ID
	orgTenantID, err := uuid.Parse(org.TenantID)
	if err != nil || orgTenantID != tenantID {
		return nil, fmt.Errorf("版本不属于指定租户")
	}

	// 2. 预检：新生效日期与现有版本冲突校验
	var conflictCount int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND effective_date = $3 
		  AND record_id != $4 AND status != 'DELETED' AND deleted_at IS NULL
	`, tenantID, org.Code, newEffectiveDate, recordID).Scan(&conflictCount)
	if err != nil {
		return nil, fmt.Errorf("冲突校验查询失败: %w", err)
	}
	if conflictCount > 0 {
		return nil, fmt.Errorf("TEMPORAL_POINT_CONFLICT: 新生效日期 %s 与现有版本冲突", newEffectiveDate.Format("2006-01-02"))
	}

	// 3. 删除旧版本（标记删除）
	now := time.Now()
	_, err = tx.ExecContext(ctx, `
		UPDATE organization_units 
		SET status = 'DELETED', deleted_at = $3, updated_at = $3
		WHERE record_id = $1 AND tenant_id = $2
	`, recordID, tenantID, now)
	if err != nil {
		return nil, fmt.Errorf("删除旧版本失败: %w", err)
	}

	// 4. 插入新版本（使用新生效日期）
	newRecordID := uuid.New()
	org.RecordID = newRecordID.String()
	org.EffectiveDate = types.NewDateFromTime(newEffectiveDate)
	org.ChangeReason = &operationReason
	org.CreatedAt = now
	org.UpdatedAt = now
	org.IsCurrent = false // 将由重算算法决定

    _, err = tx.ExecContext(ctx, `
        INSERT INTO organization_units (
            record_id, tenant_id, code, parent_code, name, unit_type, status,
            level, path, sort_order, description, effective_date, end_date,
            is_current, change_reason, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NULL, 
            false, $13, $14, $15, $15
        )
    `, newRecordID, org.TenantID, org.Code, org.ParentCode, org.Name, org.UnitType,
        org.Status, org.Level, org.Path, org.SortOrder, org.Description,
        newEffectiveDate, operationReason, now)
	if err != nil {
		return nil, fmt.Errorf("插入新版本失败: %w", err)
	}

	// 5. 执行全链重算，自动维护时间轴连续性
	timeline, err := tm.RecalculateTimelineInTx(ctx, tx, tenantID, org.Code)
	if err != nil {
		return nil, fmt.Errorf("时间轴重算失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("事务提交失败: %w", err)
	}

	tm.logger.Printf("✅ 版本生效日期修改成功: %s → %s, 时间轴已重算", recordID.String(), newEffectiveDate.Format("2006-01-02"))
	return timeline, nil
}

// SuspendOrganization 暂停组织 - 实现第四大核心场景
// 强制 status=INACTIVE，写入 SUSPEND 版本
func (tm *TemporalTimelineManager) SuspendOrganization(ctx context.Context, tenantID uuid.UUID, code string, effectiveDate time.Time, operationReason string) (*[]TimelineVersion, error) {
	return tm.changeOrganizationStatus(ctx, tenantID, code, "INACTIVE", "SUSPEND", effectiveDate, operationReason)
}

// ActivateOrganization 激活组织 - 实现第四大核心场景
// 强制 status=ACTIVE，写入 REACTIVATE 版本
func (tm *TemporalTimelineManager) ActivateOrganization(ctx context.Context, tenantID uuid.UUID, code string, effectiveDate time.Time, operationReason string) (*[]TimelineVersion, error) {
	return tm.changeOrganizationStatus(ctx, tenantID, code, "ACTIVE", "REACTIVATE", effectiveDate, operationReason)
}

// changeOrganizationStatus 通用的组织状态变更方法
func (tm *TemporalTimelineManager) changeOrganizationStatus(ctx context.Context, tenantID uuid.UUID, code string, newStatus, operationType string, effectiveDate time.Time, operationReason string) (*[]TimelineVersion, error) {
	// 开始数据库事务
	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	tm.logger.Printf("🔄 开始%s组织: Code=%s, 生效日期=%s, 新状态=%s", operationType, code, effectiveDate.Format("2006-01-02"), newStatus)

	// 1. 获取组织的当前活跃版本信息
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
               sort_order, description, effective_date, is_current, change_reason, 
               created_at, updated_at
        FROM organization_units 
        WHERE tenant_id = $1 AND code = $2 AND is_current = true 
          AND status != 'DELETED' AND deleted_at IS NULL
        FOR UPDATE
    `, tenantID, code)

    err = row.Scan(
        &currentOrg.RecordID, &currentOrg.TenantID, &currentOrg.Code, &currentOrg.ParentCode, &currentOrg.Name,
        &currentOrg.UnitType, &currentOrg.Status, &currentOrg.Level, &currentOrg.Path, &currentOrg.SortOrder,
        &currentOrg.Description, &currentOrg.EffectiveDate, &currentOrg.IsCurrent,
        &currentOrg.ChangeReason, &currentOrg.CreatedAt, &currentOrg.UpdatedAt,
    )
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在或无当前版本: %s", code)
		}
		return nil, fmt.Errorf("查询组织当前版本失败: %w", err)
	}

	// 2. 幂等性检查：如果目标状态与当前状态相同，返回成功但不创建新版本
	if currentOrg.Status == newStatus {
		tm.logger.Printf("💡 组织%s状态已经是%s，幂等操作跳过", code, newStatus)
		// 返回当前时间轴
		return tm.RecalculateTimelineInTx(ctx, tx, tenantID, code)
	}

	// 3. 冲突检查：新生效日期是否与现有版本冲突
	var conflictCount int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND effective_date = $3 
		  AND status != 'DELETED' AND deleted_at IS NULL
	`, tenantID, code, effectiveDate).Scan(&conflictCount)
	if err != nil {
		return nil, fmt.Errorf("冲突校验查询失败: %w", err)
	}
	if conflictCount > 0 {
		return nil, fmt.Errorf("TEMPORAL_POINT_CONFLICT: 生效日期 %s 与现有版本冲突", effectiveDate.Format("2006-01-02"))
	}

	// 4. 创建新的状态变更版本
	now := time.Now()
	newRecordID := uuid.New()
	
	// 判断是否为未来版本
	isFuture := effectiveDate.After(now.Truncate(24 * time.Hour))
	
    _, err = tx.ExecContext(ctx, `
        INSERT INTO organization_units (
            record_id, tenant_id, code, parent_code, name, unit_type, status,
            level, path, sort_order, description, effective_date, end_date,
            is_current, change_reason, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NULL,
            false, $13, $14, $15
        )
    `, newRecordID, currentOrg.TenantID, currentOrg.Code, currentOrg.ParentCode, currentOrg.Name,
        currentOrg.UnitType, newStatus, currentOrg.Level, currentOrg.Path, currentOrg.SortOrder,
        currentOrg.Description, effectiveDate, operationReason, now, now)
	if err != nil {
		return nil, fmt.Errorf("插入%s版本失败: %w", operationType, err)
	}

	// 5. 执行全链重算，自动维护时间轴连续性和当前版本标记
	timeline, err := tm.RecalculateTimelineInTx(ctx, tx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("时间轴重算失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("事务提交失败: %w", err)
	}

	statusAction := "暂停"
	if operationType == "REACTIVATE" {
		statusAction = "激活"
	}
	
	if isFuture {
		tm.logger.Printf("✅ 组织%s成功（计划生效）: %s → %s, 生效日期=%s", statusAction, code, newStatus, effectiveDate.Format("2006-01-02"))
	} else {
		tm.logger.Printf("✅ 组织%s成功（即时生效）: %s → %s, 时间轴已重算", statusAction, code, newStatus)
	}
	
	return timeline, nil
}
