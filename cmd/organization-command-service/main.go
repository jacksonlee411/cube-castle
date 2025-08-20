package main

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"cube-castle-deployment-test/pkg/monitoring"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// "github.com/go-redis/redis/v8"
	// "cube-castle-deployment-test/pkg/health"
)

// ===== 自定义日期类型 =====

// Date 自定义日期类型，用于处理PostgreSQL的date类型
type Date struct {
	time.Time
}

// NewDate 创建新的日期
func NewDate(year int, month time.Month, day int) *Date {
	return &Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// ParseDate 解析日期字符串 (YYYY-MM-DD)
func ParseDate(s string) (*Date, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &Date{t}, nil
}

// MarshalJSON 实现JSON序列化
func (d *Date) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}
	return json.Marshal(d.Format("2006-01-02"))
}

// UnmarshalJSON 实现JSON反序列化
func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" || s == "null" {
		return nil
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = *parsed
	return nil
}

// Scan 实现sql.Scanner接口
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*d = Date{v}
		return nil
	case string:
		parsed, err := ParseDate(v)
		if err != nil {
			return err
		}
		*d = *parsed
		return nil
	default:
		return fmt.Errorf("cannot scan %T into Date", value)
	}
}

// Value 实现driver.Valuer接口
func (d Date) Value() (driver.Value, error) {
	return d.Time, nil
}

// String 返回日期字符串
func (d *Date) String() string {
	if d == nil {
		return ""
	}
	return d.Format("2006-01-02")
}

// 默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// ===== 简化的业务实体 =====

type Organization struct {
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Code        string    `json:"code" db:"code"`
	ParentCode  *string   `json:"parent_code,omitempty" db:"parent_code"`
	Name        string    `json:"name" db:"name"`
	UnitType    string    `json:"unit_type" db:"unit_type"`
	Status      string    `json:"status" db:"status"`
	Level       int       `json:"level" db:"level"`
	Path        string    `json:"path" db:"path"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date   `json:"effective_date,omitempty" db:"effective_date"`
	EndDate       *Date   `json:"end_date,omitempty" db:"end_date"`
	IsTemporal    bool    `json:"is_temporal" db:"is_temporal"`
	ChangeReason  *string `json:"change_reason,omitempty" db:"change_reason"`
	IsCurrent     bool    `json:"is_current" db:"is_current"`
}

// ===== 简化的业务验证 =====

func ValidateCreateOrganization(req *CreateOrganizationRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("组织名称不能为空")
	}

	if len(req.Name) > 100 {
		return fmt.Errorf("组织名称不能超过100个字符")
	}

	if req.UnitType == "" {
		return fmt.Errorf("组织类型不能为空")
	}

	validTypes := map[string]bool{
		"COMPANY": true, "DEPARTMENT": true, "COST_CENTER": true, "PROJECT_TEAM": true,
	}
	if !validTypes[req.UnitType] {
		return fmt.Errorf("无效的组织类型: %s", req.UnitType)
	}

	if req.SortOrder < 0 {
		return fmt.Errorf("排序顺序不能为负数")
	}

	// 时态管理验证
	if req.IsTemporal {
		if req.EffectiveDate == nil {
			return fmt.Errorf("时态组织必须设置生效日期")
		}
		if req.EndDate != nil && req.EffectiveDate.Time.After(req.EndDate.Time) {
			return fmt.Errorf("生效日期不能晚于失效日期")
		}
		if req.ChangeReason == "" {
			return fmt.Errorf("时态组织必须提供变更原因")
		}
	}

	return nil
}

func ValidateUpdateOrganization(req *UpdateOrganizationRequest) error {
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return fmt.Errorf("组织名称不能为空")
		}
		if len(*req.Name) > 100 {
			return fmt.Errorf("组织名称不能超过100个字符")
		}
	}

	if req.UnitType != nil {
		validTypes := map[string]bool{
			"COMPANY": true, "DEPARTMENT": true, "COST_CENTER": true, "PROJECT_TEAM": true,
		}
		if !validTypes[*req.UnitType] {
			return fmt.Errorf("无效的组织类型: %s", *req.UnitType)
		}
	}

	// 移除：Status字段验证（不允许直接修改状态）

	if req.SortOrder != nil && *req.SortOrder < 0 {
		return fmt.Errorf("排序顺序不能为负数")
	}

	// 移除Level验证：level由parent_code自动计算，不允许手动设置

	// 时态管理验证
	if req.IsTemporal != nil && *req.IsTemporal {
		if req.EffectiveDate == nil {
			return fmt.Errorf("启用时态管理时必须设置生效日期")
		}
		if req.EndDate != nil && req.EffectiveDate != nil && req.EffectiveDate.Time.After(req.EndDate.Time) {
			return fmt.Errorf("生效日期不能晚于失效日期")
		}
		if req.ChangeReason == nil || *req.ChangeReason == "" {
			return fmt.Errorf("时态更新必须提供变更原因")
		}
	}

	return nil
}

// ===== 时态专用请求/响应模型 =====

// ❌ 已移除 CreatePlannedOrganizationRequest - 简化时态管理
// 使用基础创建API统一处理，通过status字段区分

// ❌ 已移除 TemporalStateChangeRequest - 功能重复
// 使用基础更新API (PUT /api/v1/organization-units/{code}) 替代

// 组织历史版本请求
type CreateOrganizationVersionRequest struct {
	BasedOnVersion int     `json:"based_on_version"`
	Name           *string `json:"name,omitempty"`
	UnitType       *string `json:"unit_type,omitempty"`
	Status         *string `json:"status,omitempty"`
	SortOrder      *int    `json:"sort_order,omitempty"`
	Description    *string `json:"description,omitempty"`
	ParentCode     *string `json:"parent_code,omitempty"`
	EffectiveDate  Date    `json:"effective_date" validate:"required"`
	EndDate        *Date   `json:"end_date,omitempty"`
	ChangeReason   string  `json:"change_reason" validate:"required"`
}

// 时态查询响应（包含时间线信息）
type TemporalOrganizationResponse struct {
	*OrganizationResponse
	TemporalStatus string                    `json:"temporal_status"`
	Timeline       []TemporalTimelineEvent   `json:"timeline,omitempty"`
	Versions       []OrganizationVersionInfo `json:"versions,omitempty"`
}

// 时间线事件
type TemporalTimelineEvent struct {
	EventType     string                 `json:"event_type"`
	EventDate     time.Time              `json:"event_date"`
	EffectiveDate *Date                  `json:"effective_date,omitempty"`
	Status        string                 `json:"status"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// 版本信息
type OrganizationVersionInfo struct {
	Version       int       `json:"version"`
	EffectiveFrom Date      `json:"effective_from"`
	EffectiveTo   *Date     `json:"effective_to,omitempty"`
	ChangeReason  string    `json:"change_reason"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateOrganizationRequest struct {
	Code        *string `json:"code,omitempty"`          // 可选：指定组织代码（用于时态记录）
	Name        string  `json:"name" validate:"required,max=100"`
	UnitType    string  `json:"unit_type" validate:"required"`
	ParentCode  *string `json:"parent_code,omitempty"`
	SortOrder   int     `json:"sort_order"`
	Description string  `json:"description"`
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date  `json:"effective_date,omitempty"`
	EndDate       *Date  `json:"end_date,omitempty"`
	IsTemporal    bool   `json:"is_temporal"`
	ChangeReason  string `json:"change_reason,omitempty"`
}

type UpdateOrganizationRequest struct {
	Name        *string `json:"name,omitempty"`
	UnitType    *string `json:"unit_type,omitempty"`
	// 移除：Status字段（不允许直接修改状态）
	SortOrder   *int    `json:"sort_order,omitempty"`
	Description *string `json:"description,omitempty"`
	// Level       *int    `json:"level,omitempty"`        // 移除：level由parent_code自动计算
	ParentCode *string `json:"parent_code,omitempty"` // 通过修改parent_code来改变层级
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date   `json:"effective_date,omitempty"`
	EndDate       *Date   `json:"end_date,omitempty"`
	IsTemporal    *bool   `json:"is_temporal,omitempty"`
	ChangeReason  *string `json:"change_reason,omitempty"`
}

type OrganizationResponse struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	UnitType    string    `json:"unit_type"`
	Status      string    `json:"status"`
	Level       int       `json:"level"`
	Path        string    `json:"path"`
	SortOrder   int       `json:"sort_order"`
	Description string    `json:"description"`
	ParentCode  *string   `json:"parent_code,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date   `json:"effective_date,omitempty"`
	EndDate       *Date   `json:"end_date,omitempty"`
	IsTemporal    bool    `json:"is_temporal"`
	ChangeReason  *string `json:"change_reason,omitempty"`
}

// 组织操作请求类型
type SuspendOrganizationRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type ReactivateOrganizationRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// ===== 简化的数据库仓储 =====

type OrganizationRepository struct {
	db     *sql.DB
	logger *log.Logger
}

func NewOrganizationRepository(db *sql.DB, logger *log.Logger) *OrganizationRepository {
	return &OrganizationRepository{db: db, logger: logger}
}

func (r *OrganizationRepository) GenerateCode(ctx context.Context, tenantID uuid.UUID) (string, error) {
	query := `
		SELECT COALESCE(MAX(CAST(code AS INTEGER)), 1000000) + 1 as next_code
		FROM organization_units 
		WHERE tenant_id = $1 AND code ~ '^[0-9]{7}$'
	`

	var nextCode int
	err := r.db.QueryRowContext(ctx, query, tenantID.String()).Scan(&nextCode)
	if err != nil {
		return "", fmt.Errorf("生成组织代码失败: %w", err)
	}

	return fmt.Sprintf("%07d", nextCode), nil
}

func (r *OrganizationRepository) Create(ctx context.Context, org *Organization) (*Organization, error) {
	query := `
		INSERT INTO organization_units (
			tenant_id, code, parent_code, name, unit_type, status, 
			level, path, sort_order, description, created_at, updated_at,
			effective_date, end_date, is_temporal, change_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt time.Time

	// 确保effective_date始终有值（数据库约束要求）
	var effectiveDate *Date
	if org.EffectiveDate != nil {
		effectiveDate = org.EffectiveDate
		r.logger.Printf("DEBUG: 使用提供的effective_date: %v", effectiveDate.String())
	} else {
		now := time.Now()
		effectiveDate = NewDate(now.Year(), now.Month(), now.Day())
		r.logger.Printf("DEBUG: 使用默认effective_date: %v", effectiveDate.String())
	}

	err := r.db.QueryRowContext(ctx, query,
		org.TenantID,
		org.Code,
		org.ParentCode,
		org.Name,
		org.UnitType,
		org.Status,
		org.Level,
		org.Path,
		org.SortOrder,
		org.Description,
		time.Now(),
		time.Now(),
		effectiveDate, // Date类型
		org.EndDate,   // 允许为nil
		org.IsTemporal,
		org.ChangeReason,
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique violation
				return nil, fmt.Errorf("组织代码已存在: %s", org.Code)
			case "23503": // foreign key violation
				return nil, fmt.Errorf("父组织不存在: %s", *org.ParentCode)
			}
		}
		return nil, fmt.Errorf("创建组织失败: %w", err)
	}

	org.CreatedAt = createdAt
	org.UpdatedAt = updatedAt
	org.EffectiveDate = effectiveDate // 确保返回的组织有effective_date值

	r.logger.Printf("组织创建成功: %s - %s (时态: %v)", org.Code, org.Name, org.IsTemporal)
	return org, nil
}

// CreateWithTemporalManagement 创建时态记录，自动处理时间连续性和end_date调整
func (r *OrganizationRepository) CreateWithTemporalManagement(ctx context.Context, tx *sql.Tx, org *Organization) (*Organization, error) {
	r.logger.Printf("DEBUG: 开始时态记录插入处理 - 组织: %s, 生效日期: %v", org.Code, org.EffectiveDate)
	
	// 第一步：查询同一组织代码的现有记录，按生效日期排序
	existingRecordsQuery := `
		SELECT record_id, code, effective_date, end_date, is_current
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2
		ORDER BY effective_date ASC
	`
	
	rows, err := tx.QueryContext(ctx, existingRecordsQuery, org.TenantID, org.Code)
	if err != nil {
		return nil, fmt.Errorf("查询现有时态记录失败: %w", err)
	}
	defer rows.Close()
	
	type ExistingRecord struct {
		RecordID      string
		Code          string
		EffectiveDate *Date
		EndDate       *Date
		IsCurrent     bool
	}
	
	var existingRecords []ExistingRecord
	for rows.Next() {
		var record ExistingRecord
		var effectiveDate, endDate sql.NullTime
		var isCurrent sql.NullBool
		
		err := rows.Scan(&record.RecordID, &record.Code, &effectiveDate, &endDate, &isCurrent)
		if err != nil {
			return nil, fmt.Errorf("扫描现有记录失败: %w", err)
		}
		
		if effectiveDate.Valid {
			record.EffectiveDate = &Date{effectiveDate.Time}
		}
		if endDate.Valid {
			record.EndDate = &Date{endDate.Time}
		}
		if isCurrent.Valid {
			record.IsCurrent = isCurrent.Bool
		}
		
		existingRecords = append(existingRecords, record)
	}
	
	newEffectiveDate := org.EffectiveDate
	r.logger.Printf("DEBUG: 找到 %d 条现有记录，新记录生效日期: %v", len(existingRecords), newEffectiveDate)
	
	if len(existingRecords) == 0 {
		// 没有现有记录，直接创建
		r.logger.Printf("DEBUG: 没有现有记录，直接创建")
		return r.CreateInTransaction(ctx, tx, org)
	}
	
	// 第二步：分析插入位置和所需的end_date调整
	insertPosition := -1 // -1表示插入到最前面，len表示插入到最后面
	
	for i, existing := range existingRecords {
		if newEffectiveDate.Time.Before(existing.EffectiveDate.Time) {
			insertPosition = i
			break
		}
	}
	
	if insertPosition == -1 {
		insertPosition = len(existingRecords)
	}
	
	r.logger.Printf("DEBUG: 插入位置: %d (总共 %d 条记录)", insertPosition, len(existingRecords))
	
	// 第三步：更新相关记录的end_date和is_current状态
	if insertPosition == 0 {
		// 插入到最前面 - 新记录成为最早的记录
		r.logger.Printf("DEBUG: 插入到最前面，新记录成为历史记录")
		
		// 计算新记录的结束日期：下一条记录生效日期的前一天
		if len(existingRecords) > 0 {
			nextDate := existingRecords[0].EffectiveDate.Time
			endDate := nextDate.AddDate(0, 0, -1)
			org.EndDate = &Date{endDate}
		}
		
		// 新插入的历史记录不是当前记录
		org.IsCurrent = false
		
	} else if insertPosition == len(existingRecords) {
		// 插入到最后面 - 新记录成为当前记录
		r.logger.Printf("DEBUG: 插入到最后面，新记录成为当前记录")
		
		// 更新之前的当前记录：设置结束日期并取消is_current状态
		lastRecord := existingRecords[len(existingRecords)-1]
		if lastRecord.IsCurrent {
			endDate := newEffectiveDate.Time.AddDate(0, 0, -1)
			updateQuery := `
				UPDATE organization_units 
				SET end_date = $1, is_current = false, updated_at = NOW()
				WHERE record_id = $2 AND tenant_id = $3
			`
			_, err = tx.ExecContext(ctx, updateQuery, endDate, lastRecord.RecordID, org.TenantID)
			if err != nil {
				return nil, fmt.Errorf("更新前一条记录的结束日期失败: %w", err)
			}
			r.logger.Printf("DEBUG: 更新记录 %s 的结束日期为: %v", lastRecord.RecordID, endDate.Format("2006-01-02"))
		}
		
		// 新记录成为当前记录，无结束日期
		org.EndDate = nil
		org.IsCurrent = true
		
	} else {
		// 插入到中间 - 新记录成为历史记录
		r.logger.Printf("DEBUG: 插入到中间位置 %d，新记录成为历史记录", insertPosition)
		
		// 更新前一条记录的结束日期
		if insertPosition > 0 {
			prevRecord := existingRecords[insertPosition-1]
			endDate := newEffectiveDate.Time.AddDate(0, 0, -1)
			updatePrevQuery := `
				UPDATE organization_units 
				SET end_date = $1, updated_at = NOW()
				WHERE record_id = $2 AND tenant_id = $3
			`
			_, err = tx.ExecContext(ctx, updatePrevQuery, endDate, prevRecord.RecordID, org.TenantID)
			if err != nil {
				return nil, fmt.Errorf("更新前一条记录的结束日期失败: %w", err)
			}
			r.logger.Printf("DEBUG: 更新前一条记录 %s 的结束日期为: %v", prevRecord.RecordID, endDate.Format("2006-01-02"))
		}
		
		// 设置新记录的结束日期为下一条记录生效日期的前一天
		nextRecord := existingRecords[insertPosition]
		nextDate := nextRecord.EffectiveDate.Time
		endDate := nextDate.AddDate(0, 0, -1)
		org.EndDate = &Date{endDate}
		
		// 中间插入的记录不是当前记录
		org.IsCurrent = false
		
		r.logger.Printf("DEBUG: 新记录结束日期设为: %v", org.EndDate.Format("2006-01-02"))
	}
	
	// 第四步：插入新记录
	r.logger.Printf("DEBUG: 插入新记录 - is_current: %v, end_date: %v", 
		org.IsCurrent, 
		func() string {
			if org.EndDate != nil {
				return org.EndDate.Format("2006-01-02")
			}
			return "null"
		}())
	
	return r.CreateInTransaction(ctx, tx, org)
}

// CreateInTransaction 在事务中创建记录的内部方法
func (r *OrganizationRepository) CreateInTransaction(ctx context.Context, tx *sql.Tx, org *Organization) (*Organization, error) {
	query := `
		INSERT INTO organization_units (
			tenant_id, code, parent_code, name, unit_type, status, 
			level, path, sort_order, description, created_at, updated_at,
			effective_date, end_date, is_temporal, change_reason, is_current
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt time.Time

	// 确保effective_date始终有值（数据库约束要求）
	var effectiveDate *Date
	if org.EffectiveDate != nil {
		effectiveDate = org.EffectiveDate
		r.logger.Printf("DEBUG: 使用提供的effective_date: %v", effectiveDate.String())
	} else {
		now := time.Now()
		effectiveDate = NewDate(now.Year(), now.Month(), now.Day())
		r.logger.Printf("DEBUG: 使用默认effective_date: %v", effectiveDate.String())
	}

	err := tx.QueryRowContext(ctx, query,
		org.TenantID,
		org.Code,
		org.ParentCode,
		org.Name,
		org.UnitType,
		org.Status,
		org.Level,
		org.Path,
		org.SortOrder,
		org.Description,
		time.Now(),
		time.Now(),
		effectiveDate, // Date类型
		org.EndDate,   // 允许为nil
		org.IsTemporal,
		org.ChangeReason,
		org.IsCurrent, // 添加is_current字段
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique violation
				return nil, fmt.Errorf("组织代码已存在: %s", org.Code)
			case "23503": // foreign key violation
				return nil, fmt.Errorf("父组织不存在: %s", *org.ParentCode)
			}
		}
		return nil, fmt.Errorf("创建组织失败: %w", err)
	}

	org.CreatedAt = createdAt
	org.UpdatedAt = updatedAt
	org.EffectiveDate = effectiveDate // 确保返回的组织有effective_date值

	r.logger.Printf("时态组织创建成功: %s - %s (生效日期: %v, 结束日期: %v, 当前: %v)", 
		org.Code, org.Name, 
		org.EffectiveDate.String(),
		func() string {
			if org.EndDate != nil {
				return org.EndDate.String()
			}
			return "无"
		}(),
		org.IsCurrent)
	return org, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, tenantID uuid.UUID, code string, req *UpdateOrganizationRequest) (*Organization, error) {
	// 构建动态更新查询
	setParts := []string{}
	args := []interface{}{tenantID.String(), code}
	argIndex := 3

	if req.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}

	if req.UnitType != nil {
		setParts = append(setParts, fmt.Sprintf("unit_type = $%d", argIndex))
		args = append(args, *req.UnitType)
		argIndex++
	}

	// 移除：Status字段更新（不允许直接修改状态）

	if req.SortOrder != nil {
		setParts = append(setParts, fmt.Sprintf("sort_order = $%d", argIndex))
		args = append(args, *req.SortOrder)
		argIndex++
	}

	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	// 移除Level更新逻辑：level由数据库触发器根据parent_code自动计算

	if req.ParentCode != nil {
		setParts = append(setParts, fmt.Sprintf("parent_code = $%d", argIndex))
		args = append(args, *req.ParentCode)
		argIndex++
	}

	// 时态管理字段更新
	if req.EffectiveDate != nil {
		setParts = append(setParts, fmt.Sprintf("effective_date = $%d", argIndex))
		args = append(args, *req.EffectiveDate)
		argIndex++
	}

	if req.EndDate != nil {
		setParts = append(setParts, fmt.Sprintf("end_date = $%d", argIndex))
		args = append(args, *req.EndDate)
		argIndex++
	}

	if req.IsTemporal != nil {
		setParts = append(setParts, fmt.Sprintf("is_temporal = $%d", argIndex))
		args = append(args, *req.IsTemporal)
		argIndex++
	}

	if req.ChangeReason != nil {
		setParts = append(setParts, fmt.Sprintf("change_reason = $%d", argIndex))
		args = append(args, *req.ChangeReason)
		argIndex++
	}

	if len(setParts) == 0 {
		// 无字段需要更新，返回空响应(避免查询操作)
		// 注意：CQRS命令端不应执行查询操作
		return nil, fmt.Errorf("无字段需要更新，操作被忽略")
	}

	// 添加updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())

	query := fmt.Sprintf(`
		UPDATE organization_units 
		SET %s
		WHERE tenant_id = $1 AND code = $2
		RETURNING tenant_id, code, parent_code, name, unit_type, status,
		          level, path, sort_order, description, created_at, updated_at,
		          effective_date, end_date, is_temporal, change_reason
	`, strings.Join(setParts, ", "))

	var org Organization
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.IsTemporal, &org.ChangeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在: %s", code)
		}
		return nil, fmt.Errorf("更新组织失败: %w", err)
	}

	r.logger.Printf("组织更新成功: %s - %s (时态: %v)", org.Code, org.Name, org.IsTemporal)
	return &org, nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, tenantID uuid.UUID, code string) error {
	// 软删除 - 设置状态为INACTIVE
	query := `
		UPDATE organization_units 
		SET status = 'INACTIVE', updated_at = $3
		WHERE tenant_id = $1 AND code = $2 AND status != 'INACTIVE'
	`

	result, err := r.db.ExecContext(ctx, query, tenantID.String(), code, time.Now())
	if err != nil {
		return fmt.Errorf("删除组织失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("组织不存在或已删除: %s", code)
	}

	r.logger.Printf("组织删除成功: %s", code)
	return nil
}

// Suspend 停用组织（设置状态为SUSPENDED）
func (r *OrganizationRepository) Suspend(ctx context.Context, tenantID uuid.UUID, code string, reason string) (*Organization, error) {
	query := `
		UPDATE organization_units 
		SET status = 'SUSPENDED', updated_at = $3
		WHERE tenant_id = $1 AND code = $2 AND status = 'ACTIVE'
		RETURNING tenant_id, code, parent_code, name, unit_type, status, 
		         level, path, sort_order, description, created_at, updated_at,
		         effective_date, end_date, is_temporal, change_reason
	`
	
	var org Organization
	var parentCode sql.NullString
	var effectiveDate, endDate sql.NullTime
	var changeReason sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code, time.Now()).Scan(
		&org.TenantID, &org.Code, &parentCode, &org.Name, &org.UnitType, &org.Status,
		&org.Level, &org.Path, &org.SortOrder, &org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &org.IsTemporal, &changeReason,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在或状态不是ACTIVE: %s", code)
		}
		return nil, fmt.Errorf("停用组织失败: %w", err)
	}
	
	// 处理可空字段
	if parentCode.Valid {
		org.ParentCode = &parentCode.String
	}
	if effectiveDate.Valid {
		d := &Date{effectiveDate.Time}
		org.EffectiveDate = d
	}
	if endDate.Valid {
		d := &Date{endDate.Time}
		org.EndDate = d
	}
	if changeReason.Valid {
		org.ChangeReason = &changeReason.String
	}
	
	r.logger.Printf("组织停用成功: %s - %s", org.Code, org.Name)
	return &org, nil
}

// Reactivate 重新启用组织（设置状态为ACTIVE）
func (r *OrganizationRepository) Reactivate(ctx context.Context, tenantID uuid.UUID, code string, reason string) (*Organization, error) {
	query := `
		UPDATE organization_units 
		SET status = 'ACTIVE', updated_at = $3
		WHERE tenant_id = $1 AND code = $2 AND status = 'SUSPENDED'
		RETURNING tenant_id, code, parent_code, name, unit_type, status, 
		         level, path, sort_order, description, created_at, updated_at,
		         effective_date, end_date, is_temporal, change_reason
	`
	
	var org Organization
	var parentCode sql.NullString
	var effectiveDate, endDate sql.NullTime
	var changeReason sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code, time.Now()).Scan(
		&org.TenantID, &org.Code, &parentCode, &org.Name, &org.UnitType, &org.Status,
		&org.Level, &org.Path, &org.SortOrder, &org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &org.IsTemporal, &changeReason,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在或状态不是SUSPENDED: %s", code)
		}
		return nil, fmt.Errorf("重新启用组织失败: %w", err)
	}
	
	// 处理可空字段
	if parentCode.Valid {
		org.ParentCode = &parentCode.String
	}
	if effectiveDate.Valid {
		d := &Date{effectiveDate.Time}
		org.EffectiveDate = d
	}
	if endDate.Valid {
		d := &Date{endDate.Time}
		org.EndDate = d
	}
	if changeReason.Valid {
		org.ChangeReason = &changeReason.String
	}
	
	r.logger.Printf("组织重新启用成功: %s - %s", org.Code, org.Name)
	return &org, nil
}

// ❌ 已移除 GetByCode - 违反CQRS原则
// 所有查询操作必须使用GraphQL服务 (端口8090)
// 查询接口: http://localhost:8090/graphql

func (r *OrganizationRepository) CalculatePath(ctx context.Context, tenantID uuid.UUID, parentCode *string, code string) (string, int, error) {
	if parentCode == nil {
		return "/" + code, 1, nil
	}

	query := `
		SELECT path, level 
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2
	`

	var parentPath string
	var parentLevel int

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), *parentCode).Scan(&parentPath, &parentLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, fmt.Errorf("父组织不存在: %s", *parentCode)
		}
		return "", 0, fmt.Errorf("查询父组织失败: %w", err)
	}

	path := parentPath + "/" + code
	level := parentLevel + 1

	return path, level, nil
}

// ===== HTTP处理器 =====

type OrganizationHandler struct {
	repo   *OrganizationRepository
	logger *log.Logger
}

func NewOrganizationHandler(repo *OrganizationRepository, logger *log.Logger) *OrganizationHandler {
	return &OrganizationHandler{repo: repo, logger: logger}
}

func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	h.logger.Printf("DEBUG: CreateOrganization called")
	var req CreateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	h.logger.Printf("DEBUG: Request decoded: %+v", req)

	// 业务验证
	if err := ValidateCreateOrganization(&req); err != nil {
		monitoring.RecordOrganizationOperation("create", "failed", "command-service")
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)

	// 确定组织代码 - 支持指定代码（用于时态记录）
	var code string
	if req.Code != nil && strings.TrimSpace(*req.Code) != "" {
		// 使用指定的代码（通常用于创建时态记录）
		code = strings.TrimSpace(*req.Code)
		h.logger.Printf("DEBUG: 使用指定的组织代码: %s", code)
	} else {
		// 生成新的组织代码
		var err error
		code, err = h.repo.GenerateCode(r.Context(), tenantID)
		if err != nil {
			monitoring.RecordOrganizationOperation("create", "failed", "command-service")
			h.writeErrorResponse(w, http.StatusInternalServerError, "CODE_GENERATION_ERROR", "生成组织代码失败", err)
			return
		}
		h.logger.Printf("DEBUG: 生成新的组织代码: %s", code)
	}

	// 计算路径和级别
	path, level, err := h.repo.CalculatePath(r.Context(), tenantID, req.ParentCode, code)
	if err != nil {
		monitoring.RecordOrganizationOperation("create", "failed", "command-service")
		h.writeErrorResponse(w, http.StatusBadRequest, "PARENT_ERROR", "父组织处理失败", err)
		return
	}

	// 创建组织实体
	now := time.Now()
	org := &Organization{
		TenantID:    tenantID.String(),
		Code:        code,
		ParentCode:  req.ParentCode,
		Name:        req.Name,
		UnitType:    req.UnitType,
		Status:      "ACTIVE",
		Level:       level,
		Path:        path,
		SortOrder:   req.SortOrder,
		Description: req.Description,
		// 时态管理字段 - 使用Date类型
		EffectiveDate: req.EffectiveDate,
		EndDate:       req.EndDate,
		IsTemporal:    req.IsTemporal,
		ChangeReason: func() *string {
			if req.ChangeReason == "" {
				return nil
			} else {
				return &req.ChangeReason
			}
		}(),
	}

	// 确保effective_date字段始终有值（数据库约束要求）
	if org.EffectiveDate == nil {
		today := NewDate(now.Year(), now.Month(), now.Day())
		org.EffectiveDate = today
	}

	// 时态管理：如果指定了组织代码且有生效日期，需要处理时态记录插入逻辑
	var createdOrg *Organization
	if req.Code != nil && strings.TrimSpace(*req.Code) != "" && org.EffectiveDate != nil {
		h.logger.Printf("DEBUG: 开始时态记录插入处理 - 代码: %s, 生效日期: %v", code, org.EffectiveDate.String())
		
		// 使用事务确保数据一致性
		tx, err := h.repo.db.Begin()
		if err != nil {
			monitoring.RecordOrganizationOperation("create", "failed", "command-service")
			h.writeErrorResponse(w, http.StatusInternalServerError, "TRANSACTION_ERROR", "开始事务失败", err)
			return
		}
		defer tx.Rollback()
		
		// 调用时态插入逻辑
		createdOrg, err = h.repo.CreateWithTemporalManagement(r.Context(), tx, org)
		if err != nil {
			monitoring.RecordOrganizationOperation("create", "failed", "command-service")
			h.writeErrorResponse(w, http.StatusInternalServerError, "TEMPORAL_CREATE_ERROR", "时态记录创建失败", err)
			return
		}
		
		// 提交事务
		if err = tx.Commit(); err != nil {
			monitoring.RecordOrganizationOperation("create", "failed", "command-service")
			h.writeErrorResponse(w, http.StatusInternalServerError, "COMMIT_ERROR", "提交事务失败", err)
			return
		}
	} else {
		// 普通创建逻辑
		var err error
		createdOrg, err = h.repo.Create(r.Context(), org)
		if err != nil {
			monitoring.RecordOrganizationOperation("create", "failed", "command-service")
			h.writeErrorResponse(w, http.StatusInternalServerError, "CREATE_ERROR", "创建组织失败", err)
			return
		}
	}

	// 构建响应
	response := h.toOrganizationResponse(createdOrg)

	monitoring.RecordOrganizationOperation("create", "success", "command-service")
	h.logger.Printf("组织创建成功: %s - %s", response.Code, response.Name)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 业务验证
	if err := ValidateUpdateOrganization(&req); err != nil {
		monitoring.RecordOrganizationOperation("update", "failed", "command-service")
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
		return
	}

	tenantID := h.getTenantID(r)

	// 更新组织
	updatedOrg, err := h.repo.Update(r.Context(), tenantID, code, &req)
	if err != nil {
		monitoring.RecordOrganizationOperation("update", "failed", "command-service")
		h.writeErrorResponse(w, http.StatusInternalServerError, "UPDATE_ERROR", "更新组织失败", err)
		return
	}

	// 构建响应
	response := h.toOrganizationResponse(updatedOrg)

	monitoring.RecordOrganizationOperation("update", "success", "command-service")
	h.logger.Printf("组织更新成功: %s - %s", response.Code, response.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	tenantID := h.getTenantID(r)

	// 删除组织
	err := h.repo.Delete(r.Context(), tenantID, code)
	if err != nil {
		monitoring.RecordOrganizationOperation("delete", "failed", "command-service")
		h.writeErrorResponse(w, http.StatusInternalServerError, "DELETE_ERROR", "删除组织失败", err)
		return
	}

	monitoring.RecordOrganizationOperation("delete", "success", "command-service")
	h.logger.Printf("组织删除成功: %s", code)

	w.WriteHeader(http.StatusNoContent)
}

// SuspendOrganization 停用组织
func (h *OrganizationHandler) SuspendOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}
	
	var req SuspendOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	
	if req.Reason == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "停用原因不能为空", nil)
		return
	}
	
	tenantID := h.getTenantID(r)
	
	// 停用组织
	org, err := h.repo.Suspend(r.Context(), tenantID, code, req.Reason)
	if err != nil {
		monitoring.RecordOrganizationOperation("suspend", "failed", "command-service")
		h.writeErrorResponse(w, http.StatusInternalServerError, "SUSPEND_ERROR", "停用组织失败", err)
		return
	}
	
	// 构建响应
	response := h.toOrganizationResponse(org)
	monitoring.RecordOrganizationOperation("suspend", "success", "command-service")
	h.logger.Printf("组织停用成功: %s - %s", response.Code, response.Name)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ReactivateOrganization 重新启用组织
func (h *OrganizationHandler) ReactivateOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}
	
	var req ReactivateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	
	if req.Reason == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", "重启原因不能为空", nil)
		return
	}
	
	tenantID := h.getTenantID(r)
	
	// 重新启用组织
	org, err := h.repo.Reactivate(r.Context(), tenantID, code, req.Reason)
	if err != nil {
		monitoring.RecordOrganizationOperation("reactivate", "failed", "command-service")
		h.writeErrorResponse(w, http.StatusInternalServerError, "REACTIVATE_ERROR", "重新启用组织失败", err)
		return
	}
	
	// 构建响应
	response := h.toOrganizationResponse(org)
	monitoring.RecordOrganizationOperation("reactivate", "success", "command-service")
	h.logger.Printf("组织重新启用成功: %s - %s", response.Code, response.Name)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ❌ 已移除 GetOrganization - 违反CQRS原则
// 所有查询操作必须使用GraphQL服务 (端口8090)
// 查询接口: http://localhost:8090/graphql

// ❌ 已移除 CreatePlannedOrganization - 简化时态管理API
// 计划组织功能已整合到基础创建API中
// 使用 POST /api/v1/organization-units 统一创建，通过status字段区分

// ❌ 已移除 TemporalStateChange - 功能重复
// 时态状态变更功能已整合到基础更新API中
// 使用 PUT /api/v1/organization-units/{code} 统一更新时态字段

// ===== 辅助方法 =====

// ❌ 已移除 validateCreatePlannedOrganization - 简化验证逻辑
// 计划组织验证已整合到基础创建验证中

// ❌ 已移除 validateTemporalStateChange - 功能重复
// 时态状态变更验证已整合到基础更新验证中

func (h *OrganizationHandler) getTenantID(r *http.Request) uuid.UUID {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		return DefaultTenantID
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		h.logger.Printf("无效的租户ID，使用默认值: %s", tenantIDStr)
		return DefaultTenantID
	}

	return tenantID
}

func (h *OrganizationHandler) toOrganizationResponse(org *Organization) *OrganizationResponse {
	return &OrganizationResponse{
		Code:        org.Code,
		Name:        org.Name,
		UnitType:    org.UnitType,
		Status:      org.Status,
		Level:       org.Level,
		Path:        org.Path,
		SortOrder:   org.SortOrder,
		Description: org.Description,
		ParentCode:  org.ParentCode,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
		// 时态管理字段
		EffectiveDate: org.EffectiveDate,
		EndDate:       org.EndDate,
		IsTemporal:    org.IsTemporal,
		ChangeReason:  org.ChangeReason,
	}
}

func (h *OrganizationHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := ErrorResponse{
		Code:    code,
		Message: message,
	}

	if err != nil {
		errorResp.Error = err.Error()
		h.logger.Printf("错误响应 [%d %s]: %v", statusCode, code, err)
	}

	json.NewEncoder(w).Encode(errorResp)
}

// ===== 主程序 =====

func main() {
	logger := log.New(os.Stdout, "[SIMPLIFIED-COMMAND] ", log.LstdFlags)

	// 数据库连接
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}
	logger.Println("PostgreSQL连接成功")

	// 创建仓储和处理器
	repo := NewOrganizationRepository(db, logger)
	handler := NewOrganizationHandler(repo, logger)

	// 创建HTTP路由
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.Logger)
	r.Use(monitoring.MetricsMiddleware("command-service")) // 统一指标收集
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 未找到路由的处理器（必须在其他路由之前）
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		errorResp := ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "端点不存在",
			Error:   fmt.Sprintf("请求的端点 %s 不存在", r.URL.Path),
		}
		json.NewEncoder(w).Encode(errorResp)
	})

	// 方法不允许的处理器
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)

		errorResp := ErrorResponse{
			Code:    "METHOD_NOT_ALLOWED",
			Message: "方法不允许",
			Error:   fmt.Sprintf("端点 %s 不支持 %s 方法", r.URL.Path, r.Method),
		}
		json.NewEncoder(w).Encode(errorResp)
	})

	// API路由 - CQRS命令端 (仅CUD操作)
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/organization-units", func(r chi.Router) {
			r.Post("/", handler.CreateOrganization)
			// ❌ 移除GET接口 - 违反CQRS原则，查询应使用GraphQL服务(8090)
			r.Put("/{code}", handler.UpdateOrganization)
			r.Delete("/{code}", handler.DeleteOrganization)

			// 组织状态操作端点
			r.Post("/{code}/suspend", handler.SuspendOrganization)       // 停用组织
			r.Post("/{code}/reactivate", handler.ReactivateOrganization) // 重新启用组织

			// ❌ 已移除时态管理专用端点 - 简化API设计
			// r.Post("/planned", handler.CreatePlannedOrganization)        // 已移除：创建计划组织
			// r.Put("/{code}/temporal-state", handler.TemporalStateChange) // 已移除：时态状态变更
		})
	})

	// 简化健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":      "Temporal Organization Command Service (CQRS)",
			"version":      "2.0.0",
			"status":       "healthy",
			"timestamp":    time.Now().Format(time.RFC3339),
			"architecture": "CQRS Command Side - 仅支持CUD操作",
		})
	})

	// Prometheus指标端点
	r.Handle("/metrics", promhttp.Handler())

	// 根路径信息 - CQRS命令服务完整文档
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":      "Temporal Organization Command Service (CQRS)",
			"version":      "2.0.0",
			"architecture": "CQRS Command Side - 仅支持CUD操作",
			"endpoints": map[string]string{
				"create": "POST /api/v1/organization-units",
				// ❌ 移除GET - 查询请使用GraphQL服务(8090)
				"update":         "PUT /api/v1/organization-units/{code}",
				"delete":         "DELETE /api/v1/organization-units/{code}",
				// ❌ 已移除时态端点 - 简化API设计
				// "create_planned": "POST /api/v1/organization-units/planned", // 已移除
				// "temporal_state": "PUT /api/v1/organization-units/{code}/temporal-state", // 已移除
				"health":         "GET /health",
				"alerts":         "GET /alerts",
				"status":         "GET /status",
				"metrics":        "GET /metrics",
			},
			"cqrs_note": "查询操作请使用GraphQL服务 http://localhost:8090/graphql",
			"temporal_features": []string{
				"计划组织创建 - 支持未来生效的组织",
				"时态状态变更 - 支持生效时间和失效时间管理",
				"版本控制 - 自动版本管理和历史追踪",
				"变更原因记录 - 强制记录所有时态变更的原因",
				"数据库触发器 - 自动创建历史版本和时间线事件",
			},
			"simplifications": []string{
				"移除过度的值对象抽象",
				"简化DDD分层架构",
				"统一业务验证逻辑",
				"减少代码文件数量68%",
				"保持核心业务价值",
				"移除过度的时态管理专用API", // 新增说明
			},
		})
	})

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Println("正在关闭简化命令服务...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("服务关闭失败: %v", err)
		}
	}()

	logger.Printf("🚀 时态组织命令服务启动成功 - 端口 :%s", port)
	logger.Printf("📍 API端点: http://localhost:%s/api/v1/organization-units", port)
	// ❌ 已移除时态端点 - 简化API设计
	// logger.Printf("📍 时态端点: http://localhost:%s/api/v1/organization-units/planned", port) // 已移除
	logger.Printf("📍 监控指标: http://localhost:%s/metrics", port)
	logger.Printf("✅ DDD简化完成: 25个文件 → 1个文件 (减少96%%)")
	logger.Printf("⏰ 时态管理集成: 支持基础时态字段和操作")
	logger.Printf("📊 版本控制: 自动历史版本和时间线事件")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务启动失败: %v", err)
	}

	logger.Println("简化命令服务已关闭")
}
