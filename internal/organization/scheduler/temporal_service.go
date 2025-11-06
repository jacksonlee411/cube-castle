package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/organization/repository"
	"cube-castle/internal/organization/utils"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

// TemporalService 时态管理服务
type TemporalService struct {
	db      *sql.DB
	orgRepo *repository.OrganizationRepository
	logger  pkglogger.Logger
}

// NewTemporalService 创建新的时态服务
func NewTemporalService(db *sql.DB, baseLogger pkglogger.Logger, orgRepo *repository.OrganizationRepository) *TemporalService {
	if baseLogger == nil {
		baseLogger = pkglogger.NewNoopLogger()
	}
	if orgRepo == nil {
		orgRepo = repository.NewOrganizationRepository(db, baseLogger)
	}
	return &TemporalService{
		db:      db,
		orgRepo: orgRepo,
		logger:  scopedLogger(baseLogger, "temporal", pkglogger.Fields{"module": "organization"}),
	}
}

// InsertVersionRequest 插入版本请求
type InsertVersionRequest struct {
	TenantID      uuid.UUID
	Code          string
	EffectiveDate time.Time
	Data          *OrganizationData
}

// OrganizationData 组织数据
type OrganizationData struct {
	Name            string
	UnitType        string
	Status          string
	ParentCode      *string
	SortOrder       int
	Description     string
	OperationType   string
	OperationReason string
}

// DeleteVersionRequest 删除版本请求
type DeleteVersionRequest struct {
	TenantID        uuid.UUID
	Code            string
	EffectiveDate   time.Time
	OperationReason string
}

// ChangeEffectiveDateRequest 变更生效日期请求
type ChangeEffectiveDateRequest struct {
	TenantID         uuid.UUID
	Code             string
	OldEffectiveDate time.Time
	NewEffectiveDate time.Time
	UpdatedData      *OrganizationData
	OperationReason  string
}

// SuspendActivateRequest 停用/启用请求
type SuspendActivateRequest struct {
	TenantID        uuid.UUID
	Code            string
	TargetStatus    string // "ACTIVE" 或 "INACTIVE"
	OperationType   string // "SUSPEND" 或 "REACTIVATE"
	EffectiveDate   time.Time
	OperationReason string
}

// VersionResponse 版本操作响应
type VersionResponse struct {
	RecordID      string     `json:"recordId"`
	Code          string     `json:"code"`
	EffectiveDate time.Time  `json:"effectiveDate"`
	EndDate       *time.Time `json:"endDate,omitempty"`
	IsCurrent     bool       `json:"isCurrent"`
	Status        string     `json:"status"`
	Message       string     `json:"message"`
}

// InsertIntermediateVersion 插入中间版本
func (s *TemporalService) InsertIntermediateVersion(ctx context.Context, req *InsertVersionRequest) (*VersionResponse, error) {
	log := s.logger.WithFields(pkglogger.Fields{
		"tenantId":      req.TenantID.String(),
		"code":          req.Code,
		"effectiveDate": req.EffectiveDate.Format("2006-01-02"),
		"operation":     "insertIntermediate",
	})
	log.Info("接收到插入中间版本请求")

	result, err := s.withTransaction(ctx, func(tx *sql.Tx) (*VersionResponse, error) {
		// 1. 读取相邻版本并锁定
		prev, next, err := s.getAdjacentVersionsForUpdate(tx, req.TenantID, req.Code, req.EffectiveDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get adjacent versions: %w", err)
		}

		// 2. 预检冲突
		if err := s.validateNonOverlapping(req.EffectiveDate, prev, next); err != nil {
			return nil, err
		}

		// 3. 回填边界
		if prev != nil {
			endDate := req.EffectiveDate.AddDate(0, 0, -1)
			if err := s.updateEndDate(tx, prev.RecordID, &endDate); err != nil {
				return nil, fmt.Errorf("failed to update end date for previous version: %w", err)
			}
		}

		// 4. 插入新版本
		newVersion, err := s.insertVersion(ctx, tx, &insertVersionData{
			TenantID:      req.TenantID,
			Code:          req.Code,
			EffectiveDate: req.EffectiveDate,
			Data:          req.Data,
			IsCurrent:     s.isCurrentEffectiveDate(req.EffectiveDate),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert new version: %w", err)
		}

		// 5. 更新当前态标记
		if newVersion.IsCurrent && prev != nil {
			if err := s.updateCurrentFlag(tx, prev.RecordID, false); err != nil {
				return nil, fmt.Errorf("failed to update current flag for previous version: %w", err)
			}
		}

		return newVersion, nil
	})
	if err != nil {
		log.Errorf("插入中间版本失败: %v", err)
		return nil, err
	}

	log.WithFields(pkglogger.Fields{
		"recordId": result.RecordID,
	}).Info("插入中间版本成功")
	return result, nil
}

// DeleteIntermediateVersion 删除中间版本
func (s *TemporalService) DeleteIntermediateVersion(ctx context.Context, req *DeleteVersionRequest) error {
	log := s.logger.WithFields(pkglogger.Fields{
		"tenantId":      req.TenantID.String(),
		"code":          req.Code,
		"effectiveDate": req.EffectiveDate.Format("2006-01-02"),
		"operation":     "deleteIntermediate",
	})
	log.Info("接收到删除中间版本请求")

	err := s.withTransactionNoReturn(ctx, func(tx *sql.Tx) error {
		// 1. 读取相邻版本并锁定
		prev, next, err := s.getAdjacentVersionsForUpdate(tx, req.TenantID, req.Code, req.EffectiveDate)
		if err != nil {
			return fmt.Errorf("failed to get adjacent versions: %w", err)
		}

		// 2. 删除目标版本
		if err := s.deleteVersionByDate(tx, req.TenantID, req.Code, req.EffectiveDate); err != nil {
			return fmt.Errorf("failed to delete version: %w", err)
		}

		// 3. 桥接相邻版本
		if prev != nil && next != nil {
			endDate := next.EffectiveDate.AddDate(0, 0, -1)
			if err := s.updateEndDate(tx, prev.RecordID, &endDate); err != nil {
				return fmt.Errorf("failed to bridge adjacent versions: %w", err)
			}
		}

		// 4. 重算整个时间线，确保末尾回写与当前态标记正确
		if err := s.recomputeTimelineInTx(tx, req.TenantID, req.Code); err != nil {
			return fmt.Errorf("failed to recompute timeline after delete: %w", err)
		}

		// 5. 写入审计日志
		return s.writeTimelineEvent(tx, req.TenantID, req.Code, "DELETE", req.OperationReason)
	})
	if err != nil {
		log.Errorf("删除中间版本失败: %v", err)
		return err
	}

	log.Info("删除中间版本成功")
	return nil
}

// ChangeEffectiveDate 变更生效日期
func (s *TemporalService) ChangeEffectiveDate(ctx context.Context, req *ChangeEffectiveDateRequest) (*VersionResponse, error) {
	log := s.logger.WithFields(pkglogger.Fields{
		"tenantId":         req.TenantID.String(),
		"code":             req.Code,
		"oldEffectiveDate": req.OldEffectiveDate.Format("2006-01-02"),
		"newEffectiveDate": req.NewEffectiveDate.Format("2006-01-02"),
		"operation":        "changeEffectiveDate",
	})
	log.Info("接收到变更生效日期请求")

	result, err := s.withTransaction(ctx, func(tx *sql.Tx) (*VersionResponse, error) {
		// 1. 预检新日期是否冲突
		if err := s.validateEffectiveDateAvailable(tx, req.TenantID, req.Code, req.NewEffectiveDate); err != nil {
			return nil, err
		}

		// 2. 删除旧版本
		if err := s.deleteVersionByDate(tx, req.TenantID, req.Code, req.OldEffectiveDate); err != nil {
			return nil, fmt.Errorf("failed to delete old version: %w", err)
		}

		// 3. 插入新版本
		insertReq := &InsertVersionRequest{
			TenantID:      req.TenantID,
			Code:          req.Code,
			EffectiveDate: req.NewEffectiveDate,
			Data:          req.UpdatedData,
		}

		result, err := s.insertIntermediateVersionInTx(ctx, tx, insertReq)
		if err != nil {
			return nil, err
		}

		// 4. 写入时间线事件
		if err := s.writeTimelineEvent(tx, req.TenantID, req.Code, "UPDATE", req.OperationReason); err != nil {
			return nil, fmt.Errorf("failed to write timeline event: %w", err)
		}

		return result, nil
	})
	if err != nil {
		log.Errorf("变更生效日期失败: %v", err)
		return nil, err
	}

	log.WithFields(pkglogger.Fields{
		"recordId": result.RecordID,
	}).Info("变更生效日期成功")
	return result, nil
}

// SuspendActivate 停用/启用操作
func (s *TemporalService) SuspendActivate(ctx context.Context, req *SuspendActivateRequest) (*VersionResponse, error) {
	return s.withTransaction(ctx, func(tx *sql.Tx) (*VersionResponse, error) {
		// 1. 幂等性检查
		currentStatus, err := s.getCurrentStatus(tx, req.TenantID, req.Code)
		if err != nil {
			return nil, fmt.Errorf("failed to get current status: %w", err)
		}

		if currentStatus == req.TargetStatus {
			// 幂等返回
			return s.getCurrentVersion(tx, req.TenantID, req.Code), nil
		}

		// 2. 创建状态变更版本
		newVersion, err := s.insertVersion(ctx, tx, &insertVersionData{
			TenantID:      req.TenantID,
			Code:          req.Code,
			EffectiveDate: req.EffectiveDate,
			Data: &OrganizationData{
				Status:          req.TargetStatus,
				OperationType:   req.OperationType,
				OperationReason: req.OperationReason,
			},
			IsCurrent: s.isCurrentEffectiveDate(req.EffectiveDate),
			IsFuture:  req.EffectiveDate.After(time.Now().UTC()),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert status change version: %w", err)
		}

		return newVersion, nil
	})
}

// 辅助方法和内部类型定义

type versionData struct {
	RecordID      string
	EffectiveDate time.Time
	EndDate       *time.Time
	IsCurrent     bool
	Status        string
}

type insertVersionData struct {
	TenantID      uuid.UUID
	Code          string
	EffectiveDate time.Time
	Data          *OrganizationData
	IsCurrent     bool
	IsFuture      bool
}

// withTransaction 事务包装器
func (s *TemporalService) withTransaction(ctx context.Context, fn func(*sql.Tx) (*VersionResponse, error)) (*VersionResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := fn(tx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// withTransactionNoReturn 无返回值事务包装器
func (s *TemporalService) withTransactionNoReturn(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// getAdjacentVersionsForUpdate 获取相邻版本并锁定
func (s *TemporalService) getAdjacentVersionsForUpdate(tx *sql.Tx, tenantID uuid.UUID, code string, effectiveDate time.Time) (*versionData, *versionData, error) {
	// 获取前一版本
	var prev *versionData
	query := `
        SELECT record_id, effective_date, end_date, is_current, status 
        FROM organization_units 
        WHERE tenant_id = $1 AND code = $2 AND effective_date < $3 
          AND status <> 'DELETED'
        ORDER BY effective_date DESC 
        LIMIT 1 
        FOR UPDATE
    `
	row := tx.QueryRow(query, tenantID, code, effectiveDate)
	var p versionData
	var endDate sql.NullTime
	err := row.Scan(&p.RecordID, &p.EffectiveDate, &endDate, &p.IsCurrent, &p.Status)
	if err == nil {
		if endDate.Valid {
			p.EndDate = &endDate.Time
		}
		prev = &p
	} else if err != sql.ErrNoRows {
		return nil, nil, fmt.Errorf("failed to get previous version: %w", err)
	}

	// 获取后一版本
	var next *versionData
	query = `
        SELECT record_id, effective_date, end_date, is_current, status 
        FROM organization_units 
        WHERE tenant_id = $1 AND code = $2 AND effective_date > $3 
          AND status <> 'DELETED'
        ORDER BY effective_date ASC 
        LIMIT 1 
        FOR UPDATE
    `
	row = tx.QueryRow(query, tenantID, code, effectiveDate)
	var n versionData
	err = row.Scan(&n.RecordID, &n.EffectiveDate, &endDate, &n.IsCurrent, &n.Status)
	if err == nil {
		if endDate.Valid {
			n.EndDate = &endDate.Time
		}
		next = &n
	} else if err != sql.ErrNoRows {
		return nil, nil, fmt.Errorf("failed to get next version: %w", err)
	}

	return prev, next, nil
}

// validateNonOverlapping 验证非重叠
func (s *TemporalService) validateNonOverlapping(effectiveDate time.Time, prev, next *versionData) error {
	// 检查是否存在相同时间点
	if prev != nil && prev.EffectiveDate.Equal(effectiveDate) {
		return fmt.Errorf("TEMPORAL_POINT_CONFLICT: effective date %v already exists", effectiveDate.Format("2006-01-02"))
	}

	// 检查与后一版本是否重叠
	if next != nil && !effectiveDate.Before(next.EffectiveDate) {
		return fmt.Errorf("TEMPORAL_OVERLAP_CONFLICT: effective date %v overlaps with next version starting %v",
			effectiveDate.Format("2006-01-02"), next.EffectiveDate.Format("2006-01-02"))
	}

	return nil
}

// isCurrentEffectiveDate 判断是否为当前生效日期
func (s *TemporalService) isCurrentEffectiveDate(effectiveDate time.Time) bool {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	return !effectiveDate.After(today)
}

// 其他辅助方法的存根实现（待完善）

func (s *TemporalService) updateEndDate(tx *sql.Tx, recordID string, endDate *time.Time) error {
	query := `UPDATE organization_units SET end_date = $1, updated_at = NOW() WHERE record_id = $2 AND status <> 'DELETED'`
	_, err := tx.Exec(query, endDate, recordID)
	return err
}

// RepairTimelineAfterSoftDelete 在软删除某个版本后，修复其相邻版本的时间边界
// 基本策略：
// - 查找被删除版本的前后相邻（忽略已删除记录）；
// - 若同时存在前后相邻，则把前一条的 end_date 回填为 后一条.effective_date - 1 天；
// - 若不存在后一条（删除的是末尾版本），则将前一条的 end_date 置为 NULL，并标记为当前态；
// - 若不存在前一条（删除的是首条），不需要处理（由后一条成为首条）。
func (s *TemporalService) RepairTimelineAfterSoftDelete(ctx context.Context, tenantID uuid.UUID, code string, deletedEffectiveDate time.Time) error {
	return s.withTransactionNoReturn(ctx, func(tx *sql.Tx) error {
		prev, next, err := s.getAdjacentVersionsForUpdate(tx, tenantID, code, deletedEffectiveDate)
		if err != nil {
			return fmt.Errorf("failed to get adjacent versions for repair: %w", err)
		}

		// 没有前一条，无需修复（由后一条自然成为首条）
		if prev == nil {
			return nil
		}

		if next != nil {
			// 存在前后相邻：桥接
			endDate := next.EffectiveDate.AddDate(0, 0, -1)
			if err := s.updateEndDate(tx, prev.RecordID, &endDate); err != nil {
				return fmt.Errorf("failed to bridge previous end_date: %w", err)
			}
			return nil
		}

		// 删除的是最后一个版本：前一条成为最后一条，清除其 end_date 并置为当前
		if err := s.updateEndDate(tx, prev.RecordID, nil); err != nil {
			return fmt.Errorf("failed to clear end_date for tail repair: %w", err)
		}
		if err := s.updateCurrentFlag(tx, prev.RecordID, true); err != nil {
			return fmt.Errorf("failed to set previous as current after tail repair: %w", err)
		}
		return nil
	})
}

func (s *TemporalService) updateCurrentFlag(tx *sql.Tx, recordID string, isCurrent bool) error {
	query := `UPDATE organization_units SET is_current = $1, updated_at = NOW() WHERE record_id = $2 AND status <> 'DELETED'`
	_, err := tx.Exec(query, isCurrent, recordID)
	return err
}

// recomputeTimelineInTx 重新计算整个时间线的 end_date 与 is_current
func (s *TemporalService) recomputeTimelineInTx(tx *sql.Tx, tenantID uuid.UUID, code string) error {
	// 读取该组织的全部非删除版本并加锁
	query := `
        SELECT record_id, effective_date
        FROM organization_units
        WHERE tenant_id = $1 AND code = $2
          AND status <> 'DELETED'
        ORDER BY effective_date ASC
        FOR UPDATE
    `
	rows, err := tx.Query(query, tenantID, code)
	if err != nil {
		return fmt.Errorf("failed to load versions for recompute: %w", err)
	}
	defer rows.Close()

	type vrow struct {
		id  string
		eff time.Time
	}
	var versions []vrow
	for rows.Next() {
		var r vrow
		if err := rows.Scan(&r.id, &r.eff); err != nil {
			return fmt.Errorf("failed to scan version row: %w", err)
		}
		versions = append(versions, r)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("row iteration error: %w", err)
	}

	// 依次回填 end_date
	for i := 0; i < len(versions); i++ {
		var newEnd *time.Time
		if i+1 < len(versions) {
			d := versions[i+1].eff.AddDate(0, 0, -1)
			newEnd = &d
		} else {
			// 最后一条开放结束
			newEnd = nil
		}
		if err := s.updateEndDate(tx, versions[i].id, newEnd); err != nil {
			return fmt.Errorf("failed to update end_date in recompute: %w", err)
		}
	}

	// 重算 is_current：最后一条有效期 <= 今天 的那条为当前
	today := time.Now().UTC().Truncate(24 * time.Hour)
	currentIndex := -1
	for i := 0; i < len(versions); i++ {
		if !versions[i].eff.After(today) {
			currentIndex = i
		} else {
			break
		}
	}

	// 清空所有当前标记（必须包含已删除记录，避免遗留的 is_current=true 导致唯一索引冲突）
	if _, err := tx.Exec(`UPDATE organization_units SET is_current = false, updated_at = NOW() WHERE tenant_id = $1 AND code = $2`, tenantID, code); err != nil {
		return fmt.Errorf("failed to clear current flags: %w", err)
	}
	if currentIndex >= 0 {
		if err := s.updateCurrentFlag(tx, versions[currentIndex].id, true); err != nil {
			return fmt.Errorf("failed to set current flag: %w", err)
		}
	}
	return nil
}

// RecomputeTimelineForCode 对单个组织代码执行完整时间线重算（公开方法）
func (s *TemporalService) RecomputeTimelineForCode(ctx context.Context, tenantID uuid.UUID, code string) error {
	return s.withTransactionNoReturn(ctx, func(tx *sql.Tx) error {
		return s.recomputeTimelineInTx(tx, tenantID, code)
	})
}

// RecomputeAllForTenant 对某个租户的所有组织执行时间线重算
func (s *TemporalService) RecomputeAllForTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	// 查询该租户下存在的组织代码（至少有一条未删除记录）
	query := `
        SELECT DISTINCT code
        FROM organization_units
        WHERE tenant_id = $1
          AND status <> 'DELETED'
    `
	rows, err := s.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return 0, fmt.Errorf("failed to list codes for tenant: %w", err)
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return 0, fmt.Errorf("scan code failed: %w", err)
		}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate codes failed: %w", err)
	}

	count := 0
	for _, code := range codes {
		if err := s.RecomputeTimelineForCode(ctx, tenantID, code); err != nil {
			return count, fmt.Errorf("recompute failed for code %s: %w", code, err)
		}
		count++
	}
	return count, nil
}

// RecomputeAllTimelines 对全库所有租户的所有组织执行时间线重算
func (s *TemporalService) RecomputeAllTimelines(ctx context.Context) (int, error) {
	// 列出所有 (tenant_id, code) 去重，限定非删除记录
	query := `
        SELECT DISTINCT tenant_id, code
        FROM organization_units
        WHERE status <> 'DELETED'
    `
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to list all tenant codes: %w", err)
	}
	defer rows.Close()

	type pair struct {
		tenant uuid.UUID
		code   string
	}
	var pairs []pair
	for rows.Next() {
		var t uuid.UUID
		var c string
		if err := rows.Scan(&t, &c); err != nil {
			return 0, fmt.Errorf("scan pair failed: %w", err)
		}
		pairs = append(pairs, pair{tenant: t, code: c})
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate pairs failed: %w", err)
	}

	count := 0
	for _, p := range pairs {
		if err := s.RecomputeTimelineForCode(ctx, p.tenant, p.code); err != nil {
			return count, fmt.Errorf("recompute failed for %s/%s: %w", p.tenant, p.code, err)
		}
		count++
	}
	return count, nil
}

type orgSnapshot struct {
	Name        string
	UnitType    string
	Status      string
	ParentCode  *string
	SortOrder   int
	Description string
}

func (s *TemporalService) loadLatestSnapshot(tx *sql.Tx, tenantID uuid.UUID, code string) (*orgSnapshot, error) {
	query := `
		SELECT name, unit_type, status, parent_code, sort_order, description
		FROM organization_units
		WHERE tenant_id = $1 AND code = $2 AND status <> 'DELETED'
		ORDER BY effective_date DESC
		LIMIT 1
	`

	var snap orgSnapshot
	var parent sql.NullString
	var sortOrder sql.NullInt64
	var description sql.NullString

	err := tx.QueryRow(query, tenantID, code).Scan(&snap.Name, &snap.UnitType, &snap.Status, &parent, &sortOrder, &description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load existing organization snapshot: %w", err)
	}

	if parent.Valid {
		value := parent.String
		snap.ParentCode = &value
	}
	if sortOrder.Valid {
		snap.SortOrder = int(sortOrder.Int64)
	}
	if description.Valid {
		snap.Description = description.String
	}

	return &snap, nil
}

func (s *TemporalService) insertVersion(ctx context.Context, tx *sql.Tx, data *insertVersionData) (*VersionResponse, error) {
	if data == nil {
		return nil, fmt.Errorf("insertVersion: missing insert data")
	}

	payload := data.Data
	if payload == nil {
		payload = &OrganizationData{}
	}

	snapshot, err := s.loadLatestSnapshot(tx, data.TenantID, data.Code)
	if err != nil {
		return nil, err
	}

	finalName := strings.TrimSpace(payload.Name)
	if finalName == "" && snapshot != nil {
		finalName = strings.TrimSpace(snapshot.Name)
	}
	if finalName == "" {
		return nil, fmt.Errorf("missing organization name for temporal version")
	}

	finalUnitType := strings.TrimSpace(payload.UnitType)
	if finalUnitType == "" && snapshot != nil {
		finalUnitType = snapshot.UnitType
	}
	if finalUnitType == "" {
		return nil, fmt.Errorf("missing unit type for temporal version")
	}

	finalStatus := strings.TrimSpace(payload.Status)
	if finalStatus == "" {
		if snapshot != nil && snapshot.Status != "" {
			finalStatus = snapshot.Status
		} else {
			finalStatus = "ACTIVE"
		}
	}

	var parentCode *string
	if payload.ParentCode != nil {
		parentCode = utils.NormalizeParentCodePointer(payload.ParentCode)
	} else if snapshot != nil && snapshot.ParentCode != nil {
		parentCode = utils.NormalizeParentCodePointer(snapshot.ParentCode)
	}

	finalSortOrder := payload.SortOrder
	if data.Data == nil && snapshot != nil {
		finalSortOrder = snapshot.SortOrder
	}

	finalDescription := payload.Description
	if data.Data == nil && snapshot != nil {
		finalDescription = snapshot.Description
	}

	fields, calcErr := s.orgRepo.ComputeHierarchyForNew(ctx, data.TenantID, data.Code, parentCode, finalName)
	if calcErr != nil {
		return nil, fmt.Errorf("failed to compute hierarchy fields: %w", calcErr)
	}

	recordID := uuid.New().String()

	query := `
		INSERT INTO organization_units (
			record_id, tenant_id, code, effective_date, end_date, is_current,
			status, name, unit_type, parent_code, level, code_path, name_path,
			sort_order, description, change_reason, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW())
	`

	changeReason := strings.TrimSpace(payload.OperationReason)

	_, err = tx.Exec(query,
		recordID, data.TenantID, data.Code, data.EffectiveDate, nil, data.IsCurrent,
		finalStatus, finalName, finalUnitType, parentCode, fields.Level, fields.CodePath, fields.NamePath,
		finalSortOrder, finalDescription, changeReason,
	)
	if err != nil {
		return nil, err
	}

	return &VersionResponse{
		RecordID:      recordID,
		Code:          data.Code,
		EffectiveDate: data.EffectiveDate,
		IsCurrent:     data.IsCurrent,
		Status:        finalStatus,
		Message:       "Version inserted successfully",
	}, nil
}

func (s *TemporalService) deleteVersionByDate(tx *sql.Tx, tenantID uuid.UUID, code string, effectiveDate time.Time) error {
	query := `DELETE FROM organization_units WHERE tenant_id = $1 AND code = $2 AND effective_date = $3 AND status <> 'DELETED'`
	_, err := tx.Exec(query, tenantID, code, effectiveDate)
	return err
}

func (s *TemporalService) validateEffectiveDateAvailable(tx *sql.Tx, tenantID uuid.UUID, code string, effectiveDate time.Time) error {
	var count int
	query := `SELECT COUNT(*) FROM organization_units WHERE tenant_id = $1 AND code = $2 AND effective_date = $3 AND status <> 'DELETED'`
	err := tx.QueryRow(query, tenantID, code, effectiveDate).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("TEMPORAL_POINT_CONFLICT: effective date %v already exists", effectiveDate.Format("2006-01-02"))
	}
	return nil
}

func (s *TemporalService) getCurrentStatus(tx *sql.Tx, tenantID uuid.UUID, code string) (string, error) {
	var status string
	query := `SELECT status FROM organization_units WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'`
	err := tx.QueryRow(query, tenantID, code).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}

func (s *TemporalService) getCurrentVersion(tx *sql.Tx, tenantID uuid.UUID, code string) *VersionResponse {
	// 简化实现 - 返回基本信息
	return &VersionResponse{
		Code:    code,
		Status:  "ACTIVE", // 简化
		Message: "Current version returned",
	}
}

func (s *TemporalService) insertIntermediateVersionInTx(ctx context.Context, tx *sql.Tx, req *InsertVersionRequest) (*VersionResponse, error) {
	// 复用事务内的插入逻辑
	return s.insertVersion(ctx, tx, &insertVersionData{
		TenantID:      req.TenantID,
		Code:          req.Code,
		EffectiveDate: req.EffectiveDate,
		Data:          req.Data,
		IsCurrent:     s.isCurrentEffectiveDate(req.EffectiveDate),
	})
}

func (s *TemporalService) writeTimelineEvent(tx *sql.Tx, tenantID uuid.UUID, code, operationType, reason string) error {
	// 简化实现 - 写入审计日志
	// 这里可以扩展为完整的时间线事件记录
	return nil
}
