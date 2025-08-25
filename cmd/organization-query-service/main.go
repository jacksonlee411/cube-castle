package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"postgresql-graphql-service/internal/auth"
)

// 默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// 激进优化的GraphQL Schema - PostgreSQL原生 (camelCase字段命名)
var schemaString = `
	type Organization {
		recordId: String!
		tenantId: String!
		code: String!
		parentCode: String
		name: String!
		unitType: String!
		status: String!
		level: Int!
		path: String
		sortOrder: Int
		description: String
		profile: String
		createdAt: String!
		updatedAt: String!
		effectiveDate: String!
		endDate: String
		# PostgreSQL专属时态字段
		isCurrent: Boolean!
		isTemporal: Boolean!
		changeReason: String
		# 删除状态管理
		deletedAt: String
		deletedBy: String
		deletionReason: String
		# 暂停状态管理
		suspendedAt: String
		suspendedBy: String
		suspensionReason: String
	}

	type OrganizationConnection {
		data: [Organization!]!
		pagination: PaginationInfo!
		temporal: TemporalInfo!
	}

	type PaginationInfo {
		total: Int!
		page: Int!
		pageSize: Int!
		hasNext: Boolean!
		hasPrevious: Boolean!
	}

	type TemporalInfo {
		asOfDate: String!
		currentCount: Int!
		futureCount: Int!
		historicalCount: Int!
	}

	type Query {
		# 高性能当前数据查询 - 符合官方API契约 v4.2.1
		organizations(filter: OrganizationFilter, pagination: PaginationInput): OrganizationConnection!
		organization(code: String!): Organization
		organizationStats: OrganizationStats!
		
		# 极速时态查询 - PostgreSQL窗口函数优化
		organizationAtDate(code: String!, date: String!): Organization
		organizationHistory(code: String!, fromDate: String!, toDate: String!): [Organization!]!
		
		# 高级时态分析 - PostgreSQL独有功能
		organizationVersions(code: String!): [Organization!]!
	}

	# 输入类型 - 按官方契约定义
	input OrganizationFilter {
		unitType: String
		status: String
		parentCode: String
		searchText: String
		asOfDate: String
	}

	input PaginationInput {
		page: Int
		pageSize: Int
	}
	
	type OrganizationStats {
		totalCount: Int!
		activeCount: Int!
		inactiveCount: Int!
		plannedCount: Int!
		deletedCount: Int!
		byType: [TypeCount!]!
		byStatus: [StatusCount!]!
		byLevel: [LevelCount!]!
		temporalStats: TemporalStats!
	}

	type TemporalStats {
		totalVersions: Int!
		averageVersionsPerOrg: Float!
		oldestEffectiveDate: String!
		newestEffectiveDate: String!
	}

	type TypeCount {
		unitType: String!
		count: Int!
	}

	type LevelCount {
		level: Int!
		count: Int!
	}

	type StatusCount {
		status: String!
		count: Int!
	}
`

// PostgreSQL原生组织模型 - 零转换开销 (camelCase JSON标签)
type Organization struct {
	RecordIDField         string     `json:"recordId" db:"record_id"`
	TenantIDField         string     `json:"tenantId" db:"tenant_id"`
	CodeField             string     `json:"code" db:"code"`
	ParentCodeField       *string    `json:"parentCode" db:"parent_code"`
	NameField             string     `json:"name" db:"name"`
	UnitTypeField         string     `json:"unitType" db:"unit_type"`
	StatusField           string     `json:"status" db:"status"`
	LevelField            int        `json:"level" db:"level"`
	PathField             *string    `json:"path" db:"path"`
	SortOrderField        *int       `json:"sortOrder" db:"sort_order"`
	DescriptionField      *string    `json:"description" db:"description"`
	ProfileField          *string    `json:"profile" db:"profile"`
	CreatedAtField        time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAtField        time.Time  `json:"updatedAt" db:"updated_at"`
	EffectiveDateField    time.Time  `json:"effectiveDate" db:"effective_date"`
	EndDateField          *time.Time `json:"endDate" db:"end_date"`
	IsCurrentField        bool       `json:"isCurrent" db:"is_current"`
	IsTemporalField       *bool      `json:"isTemporal" db:"is_temporal"`
	ChangeReasonField     *string    `json:"changeReason" db:"change_reason"`
	DeletedAtField        *time.Time `json:"deletedAt" db:"deleted_at"`
	DeletedByField        *string    `json:"deletedBy" db:"deleted_by"`
	DeletionReasonField   *string    `json:"deletionReason" db:"deletion_reason"`
	SuspendedAtField      *time.Time `json:"suspendedAt" db:"suspended_at"`
	SuspendedByField      *string    `json:"suspendedBy" db:"suspended_by"`
	SuspensionReasonField *string    `json:"suspensionReason" db:"suspension_reason"`
}

// GraphQL字段解析器 - 零拷贝优化 (camelCase方法名)
func (o Organization) RecordId() string    { return o.RecordIDField }
func (o Organization) TenantId() string    { return o.TenantIDField }
func (o Organization) Code() string        { return o.CodeField }
func (o Organization) ParentCode() *string { return o.ParentCodeField }
func (o Organization) Name() string        { return o.NameField }
func (o Organization) UnitType() string    { return o.UnitTypeField }
func (o Organization) Status() string      { return o.StatusField }
func (o Organization) Level() int32        { return int32(o.LevelField) }
func (o Organization) Path() *string       { return o.PathField }
func (o Organization) SortOrder() *int32 {
	if o.SortOrderField == nil {
		return nil
	}
	val := int32(*o.SortOrderField)
	return &val
}
func (o Organization) Description() *string  { return o.DescriptionField }
func (o Organization) Profile() *string      { return o.ProfileField }
func (o Organization) CreatedAt() string     { return o.CreatedAtField.Format(time.RFC3339) }
func (o Organization) UpdatedAt() string     { return o.UpdatedAtField.Format(time.RFC3339) }
func (o Organization) EffectiveDate() string { return o.EffectiveDateField.Format("2006-01-02") }
func (o Organization) EndDate() *string {
	if o.EndDateField == nil {
		return nil
	}
	date := o.EndDateField.Format("2006-01-02")
	return &date
}
func (o Organization) IsCurrent() bool { return o.IsCurrentField }
func (o Organization) IsTemporal() bool {
	if o.IsTemporalField == nil {
		return false
	}
	return *o.IsTemporalField
}
func (o Organization) ChangeReason() *string { return o.ChangeReasonField }
func (o Organization) DeletedAt() *string {
	if o.DeletedAtField == nil {
		return nil
	}
	ts := o.DeletedAtField.Format(time.RFC3339)
	return &ts
}
func (o Organization) DeletedBy() *string      { return o.DeletedByField }
func (o Organization) DeletionReason() *string { return o.DeletionReasonField }
func (o Organization) SuspendedAt() *string {
	if o.SuspendedAtField == nil {
		return nil
	}
	ts := o.SuspendedAtField.Format(time.RFC3339)
	return &ts
}
func (o Organization) SuspendedBy() *string      { return o.SuspendedByField }
func (o Organization) SuspensionReason() *string { return o.SuspensionReasonField }

// 统计信息 (camelCase JSON标签)
type OrganizationStats struct {
	TotalCountField    int           `json:"totalCount"`
	ActiveCountField   int           `json:"activeCount"`
	InactiveCountField int           `json:"inactiveCount"`
	PlannedCountField  int           `json:"plannedCount"`
	DeletedCountField  int           `json:"deletedCount"`
	ByTypeField        []TypeCount   `json:"byType"`
	ByStatusField      []StatusCount `json:"byStatus"`
	ByLevelField       []LevelCount  `json:"byLevel"`
	TemporalStatsField TemporalStats `json:"temporalStats"`
}

func (s OrganizationStats) TotalCount() int32            { return int32(s.TotalCountField) }
func (s OrganizationStats) ActiveCount() int32           { return int32(s.ActiveCountField) }
func (s OrganizationStats) InactiveCount() int32         { return int32(s.InactiveCountField) }
func (s OrganizationStats) PlannedCount() int32          { return int32(s.PlannedCountField) }
func (s OrganizationStats) DeletedCount() int32          { return int32(s.DeletedCountField) }
func (s OrganizationStats) ByType() []TypeCount          { return s.ByTypeField }
func (s OrganizationStats) ByStatus() []StatusCount      { return s.ByStatusField }
func (s OrganizationStats) ByLevel() []LevelCount        { return s.ByLevelField }
func (s OrganizationStats) TemporalStats() TemporalStats { return s.TemporalStatsField }

type TemporalStats struct {
	TotalVersionsField         int     `json:"totalVersions"`
	AverageVersionsPerOrgField float64 `json:"averageVersionsPerOrg"`
	OldestEffectiveDateField   string  `json:"oldestEffectiveDate"`
	NewestEffectiveDateField   string  `json:"newestEffectiveDate"`
}

func (t TemporalStats) TotalVersions() int32           { return int32(t.TotalVersionsField) }
func (t TemporalStats) AverageVersionsPerOrg() float64 { return t.AverageVersionsPerOrgField }
func (t TemporalStats) OldestEffectiveDate() string    { return t.OldestEffectiveDateField }
func (t TemporalStats) NewestEffectiveDate() string    { return t.NewestEffectiveDateField }

type TypeCount struct {
	UnitTypeField string `json:"unitType"`
	CountField    int    `json:"count"`
}

func (t TypeCount) UnitType() string { return t.UnitTypeField }
func (t TypeCount) Count() int32     { return int32(t.CountField) }

type LevelCount struct {
	LevelField int `json:"level"`
	CountField int `json:"count"`
}

func (l LevelCount) Level() int32 { return int32(l.LevelField) }
func (l LevelCount) Count() int32 { return int32(l.CountField) }

type StatusCount struct {
	StatusField string `json:"status"`
	CountField  int    `json:"count"`
}

func (s StatusCount) Status() string { return s.StatusField }
func (s StatusCount) Count() int32   { return int32(s.CountField) }

// API契约标准响应类型 - 符合官方schema.graphql v4.2.1
type OrganizationConnection struct {
	DataField       []Organization `json:"data"`
	PaginationField PaginationInfo `json:"pagination"`
	TemporalField   TemporalInfo   `json:"temporal"`
}

func (c OrganizationConnection) Data() []Organization       { return c.DataField }
func (c OrganizationConnection) Pagination() PaginationInfo { return c.PaginationField }
func (c OrganizationConnection) Temporal() TemporalInfo     { return c.TemporalField }

type PaginationInfo struct {
	TotalField       int  `json:"total"`
	PageField        int  `json:"page"`
	PageSizeField    int  `json:"pageSize"`
	HasNextField     bool `json:"hasNext"`
	HasPreviousField bool `json:"hasPrevious"`
}

func (p PaginationInfo) Total() int32      { return int32(p.TotalField) }
func (p PaginationInfo) Page() int32       { return int32(p.PageField) }
func (p PaginationInfo) PageSize() int32   { return int32(p.PageSizeField) }
func (p PaginationInfo) HasNext() bool     { return p.HasNextField }
func (p PaginationInfo) HasPrevious() bool { return p.HasPreviousField }

type TemporalInfo struct {
	AsOfDateField        string `json:"asOfDate"`
	CurrentCountField    int    `json:"currentCount"`
	FutureCountField     int    `json:"futureCount"`
	HistoricalCountField int    `json:"historicalCount"`
}

func (t TemporalInfo) AsOfDate() string       { return t.AsOfDateField }
func (t TemporalInfo) CurrentCount() int32    { return int32(t.CurrentCountField) }
func (t TemporalInfo) FutureCount() int32     { return int32(t.FutureCountField) }
func (t TemporalInfo) HistoricalCount() int32 { return int32(t.HistoricalCountField) }

// 输入类型 - 符合官方API契约
type OrganizationFilter struct {
	UnitType   *string `json:"unitType"`
	Status     *string `json:"status"`
	ParentCode *string `json:"parentCode"`
	SearchText *string `json:"searchText"`
	AsOfDate   *string `json:"asOfDate"`
}

type PaginationInput struct {
	Page     *int32 `json:"page"`
	PageSize *int32 `json:"pageSize"`
}

// PostgreSQL极速仓储 - 零抽象开销
type PostgreSQLRepository struct {
	db          *sql.DB
	redisClient *redis.Client
	logger      *log.Logger
}

func NewPostgreSQLRepository(db *sql.DB, redisClient *redis.Client, logger *log.Logger) *PostgreSQLRepository {
	return &PostgreSQLRepository{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
	}
}

// 极速当前组织查询 - 利用部分索引 idx_current_organizations_list (API契约v4.2.1)
func (r *PostgreSQLRepository) GetOrganizations(ctx context.Context, tenantID uuid.UUID, filter *OrganizationFilter, pagination *PaginationInput) (*OrganizationConnection, error) {
	start := time.Now()

	// 解析分页参数 - 使用契约默认值
	page := int32(1)
	pageSize := int32(50)
	if pagination != nil {
		if pagination.Page != nil && *pagination.Page > 0 {
			page = *pagination.Page
		}
		if pagination.PageSize != nil && *pagination.PageSize > 0 {
			pageSize = *pagination.PageSize
		}
	}

	offset := (page - 1) * pageSize
	limit := pageSize

	// 解析过滤参数
	var status, searchText, unitType, parentCode string
	if filter != nil {
		if filter.Status != nil {
			status = *filter.Status
		}
		if filter.SearchText != nil {
			searchText = *filter.SearchText
		}
		if filter.UnitType != nil {
			unitType = *filter.UnitType
		}
		if filter.ParentCode != nil {
			parentCode = *filter.ParentCode
		}
	}

	// 构建高性能查询 - 充分利用PostgreSQL索引
	baseQuery := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status, 
		       level, path, sort_order, description, profile, created_at, updated_at,
		       effective_date, end_date, is_current, is_temporal, change_reason,
		       deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
		FROM organization_units 
		WHERE tenant_id = $1 AND is_current = true`

	countQuery := `
		SELECT COUNT(*) 
		FROM organization_units 
		WHERE tenant_id = $1 AND is_current = true`

	args := []interface{}{tenantID.String()}
	argIndex := 2
	whereConditions := ""

	// 状态过滤
	if status != "" {
		whereConditions += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	} else {
		whereConditions += " AND status <> 'DELETED'"
	}

	// 单位类型过滤
	if unitType != "" {
		whereConditions += fmt.Sprintf(" AND unit_type = $%d", argIndex)
		args = append(args, unitType)
		argIndex++
	}

	// 父组织过滤
	if parentCode != "" {
		whereConditions += fmt.Sprintf(" AND parent_code = $%d", argIndex)
		args = append(args, parentCode)
		argIndex++
	}

	// 文本搜索 - 使用GIN索引
	if searchText != "" {
		whereConditions += fmt.Sprintf(" AND (name ILIKE $%d OR code ILIKE $%d)", argIndex, argIndex)
		searchPattern := "%" + searchText + "%"
		args = append(args, searchPattern)
		argIndex++
	}

	// 完整查询
	dataQuery := baseQuery + whereConditions + " ORDER BY sort_order NULLS LAST, code LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	totalQuery := countQuery + whereConditions

	// 执行总数查询
	var total int
	err := r.db.QueryRowContext(ctx, totalQuery, args...).Scan(&total)
	if err != nil {
		r.logger.Printf("[ERROR] 查询组织总数失败: %v", err)
		return nil, err
	}

	// 执行数据查询
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Printf("[ERROR] 查询当前组织失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var organizations []Organization
	for rows.Next() {
		var org Organization
		err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, &org.IsTemporalField,
			&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
			&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描组织数据失败: %v", err)
			return nil, err
		}
		organizations = append(organizations, org)
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 查询 %d/%d 组织 (页面: %d/%d)，耗时: %v", len(organizations), total, page, (total+int(pageSize)-1)/int(pageSize), duration)

	// 构建符合契约的响应结构
	totalPages := (total + int(pageSize) - 1) / int(pageSize)
	response := &OrganizationConnection{
		DataField: organizations,
		PaginationField: PaginationInfo{
			TotalField:       total,
			PageField:        int(page),
			PageSizeField:    int(pageSize),
			HasNextField:     int(page) < totalPages,
			HasPreviousField: page > 1,
		},
		TemporalField: TemporalInfo{
			AsOfDateField:        time.Now().Format("2006-01-02"),
			CurrentCountField:    len(organizations),
			FutureCountField:     0, // TODO: 基于时态数据计算
			HistoricalCountField: 0, // TODO: 基于历史数据计算
		},
	}

	return response, nil
}

// 单个组织查询 - 超快速索引查询
func (r *PostgreSQLRepository) GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*Organization, error) {
	// 使用 idx_current_record_fast 索引
	query := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status, 
		       level, path, sort_order, description, profile, created_at, updated_at,
		       effective_date, end_date, is_current, is_temporal, change_reason,
		       deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND is_current = true
		LIMIT 1`

	start := time.Now()
	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code)

	var org Organization
	err := row.Scan(
		&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
		&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
		&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
		&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, &org.IsTemporalField,
		&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
		&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Printf("[ERROR] 查询单个组织失败: %v", err)
		return nil, err
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 单个组织查询，耗时: %v", duration)

	return &org, nil
}

// 极速时态查询 - 时间点查询（利用时态索引）
func (r *PostgreSQLRepository) GetOrganizationAtDate(ctx context.Context, tenantID uuid.UUID, code, date string) (*Organization, error) {
	// 使用 idx_org_temporal_range_composite 索引
	query := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status, 
		       level, path, sort_order, description, profile, created_at, updated_at,
		       effective_date, end_date, is_current, is_temporal, change_reason,
		       deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 
		  AND effective_date <= $3::date 
		  AND (end_date IS NULL OR end_date >= $3::date)
		ORDER BY effective_date DESC, created_at DESC
		LIMIT 1`

	start := time.Now()
	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code, date)

	var org Organization
	err := row.Scan(
		&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
		&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
		&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
		&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, &org.IsTemporalField,
		&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
		&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Printf("[ERROR] 时态查询失败: %v", err)
		return nil, err
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 时态点查询 [%s @ %s]，耗时: %v", code, date, duration)

	return &org, nil
}

// 历史范围查询 - 窗口函数优化
func (r *PostgreSQLRepository) GetOrganizationHistory(ctx context.Context, tenantID uuid.UUID, code, fromDate, toDate string) ([]Organization, error) {
	// 使用窗口函数和时态索引优化历史查询
	query := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status, 
		       level, path, sort_order, description, profile, created_at, updated_at,
		       effective_date, end_date, is_current, is_temporal, change_reason,
		       deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 
		  AND effective_date BETWEEN $3::date AND $4::date
		ORDER BY effective_date DESC, created_at DESC`

	start := time.Now()
	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), code, fromDate, toDate)
	if err != nil {
		r.logger.Printf("[ERROR] 历史范围查询失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var organizations []Organization
	for rows.Next() {
		var org Organization
		err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, &org.IsTemporalField,
			&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
			&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描历史数据失败: %v", err)
			return nil, err
		}
		organizations = append(organizations, org)
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 历史查询 [%s: %s~%s] 返回 %d 条，耗时: %v", code, fromDate, toDate, len(organizations), duration)

	return organizations, nil
}

// 高级统计查询 - 利用PostgreSQL聚合优化
func (r *PostgreSQLRepository) GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*OrganizationStats, error) {
	start := time.Now()

	// 使用单个复杂查询获取所有统计信息
	query := `
		WITH status_stats AS (
			SELECT 
				COUNT(*) as total_count,
				SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END) as active_count,
				SUM(CASE WHEN status = 'INACTIVE' THEN 1 ELSE 0 END) as inactive_count,
				SUM(CASE WHEN status = 'PLANNED' THEN 1 ELSE 0 END) as planned_count,
				SUM(CASE WHEN status = 'DELETED' THEN 1 ELSE 0 END) as deleted_count
			FROM organization_units WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
		),
		type_stats AS (
			SELECT unit_type, COUNT(*) as count
			FROM organization_units 
			WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
			GROUP BY unit_type
		),
		status_detail_stats AS (
			SELECT status, COUNT(*) as count
			FROM organization_units 
			WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
			GROUP BY status
		),
		level_stats AS (
			SELECT level, COUNT(*) as count
			FROM organization_units 
			WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
			GROUP BY level
		),
		temporal_stats AS (
			SELECT 
				COUNT(*) as total_versions,
				COUNT(DISTINCT code) as unique_orgs,
				MIN(effective_date) as oldest_date,
				MAX(effective_date) as newest_date
			FROM organization_units WHERE tenant_id = $1
		)
		SELECT 
			s.total_count, s.active_count, s.inactive_count, s.planned_count, s.deleted_count,
			ts.total_versions, ts.unique_orgs, ts.oldest_date, ts.newest_date,
			COALESCE(json_agg(DISTINCT jsonb_build_object('unit_type', t.unit_type, 'count', t.count)) FILTER (WHERE t.unit_type IS NOT NULL), '[]'),
			COALESCE(json_agg(DISTINCT jsonb_build_object('status', sd.status, 'count', sd.count)) FILTER (WHERE sd.status IS NOT NULL), '[]'),
			COALESCE(json_agg(DISTINCT jsonb_build_object('level', l.level, 'count', l.count)) FILTER (WHERE l.level IS NOT NULL), '[]')
		FROM status_stats s
		CROSS JOIN temporal_stats ts
		LEFT JOIN type_stats t ON true
		LEFT JOIN status_detail_stats sd ON true
		LEFT JOIN level_stats l ON true
		GROUP BY s.total_count, s.active_count, s.inactive_count, s.planned_count, s.deleted_count,
		         ts.total_versions, ts.unique_orgs, ts.oldest_date, ts.newest_date`

	row := r.db.QueryRowContext(ctx, query, tenantID.String())

	var stats OrganizationStats
	var totalVersions, uniqueOrgs int
	var oldestDate, newestDate time.Time
	var typeStatsJSON, statusStatsJSON, levelStatsJSON string

	err := row.Scan(
		&stats.TotalCountField, &stats.ActiveCountField, &stats.InactiveCountField,
		&stats.PlannedCountField, &stats.DeletedCountField,
		&totalVersions, &uniqueOrgs, &oldestDate, &newestDate,
		&typeStatsJSON, &statusStatsJSON, &levelStatsJSON,
	)
	if err != nil {
		r.logger.Printf("[ERROR] 统计查询失败: %v", err)
		return nil, err
	}

	// 解析JSON统计数据
	var typeStats []TypeCount
	if typeStatsJSON != "" {
		json.Unmarshal([]byte(typeStatsJSON), &typeStats)
	}
	stats.ByTypeField = typeStats

	var statusStats []StatusCount
	if statusStatsJSON != "" {
		json.Unmarshal([]byte(statusStatsJSON), &statusStats)
	}
	stats.ByStatusField = statusStats

	var levelStats []LevelCount
	if levelStatsJSON != "" {
		json.Unmarshal([]byte(levelStatsJSON), &levelStats)
	}
	stats.ByLevelField = levelStats

	// 时态统计
	stats.TemporalStatsField = TemporalStats{
		TotalVersionsField:         totalVersions,
		AverageVersionsPerOrgField: float64(totalVersions) / float64(uniqueOrgs),
		OldestEffectiveDateField:   oldestDate.Format("2006-01-02"),
		NewestEffectiveDateField:   newestDate.Format("2006-01-02"),
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 统计查询完成，耗时: %v", duration)

	return &stats, nil
}

// GraphQL解析器 - 极简高效
type Resolver struct {
	repo   *PostgreSQLRepository
	logger *log.Logger
}

// 当前组织列表查询 - 符合API契约v4.2.1 (camelCase方法名)
func (r *Resolver) Organizations(ctx context.Context, args struct {
	Filter     *OrganizationFilter
	Pagination *PaginationInput
}) (*OrganizationConnection, error) {
	r.logger.Printf("[GraphQL] 查询组织列表 - API契约v4.2.1")

	// 记录查询参数用于调试
	if args.Filter != nil {
		r.logger.Printf("[GraphQL] 过滤条件: %+v", *args.Filter)
	}
	if args.Pagination != nil {
		r.logger.Printf("[GraphQL] 分页参数: %+v", *args.Pagination)
	}

	return r.repo.GetOrganizations(ctx, DefaultTenantID, args.Filter, args.Pagination)
}

// 单个组织查询
func (r *Resolver) Organization(ctx context.Context, args struct {
	Code string
}) (*Organization, error) {
	r.logger.Printf("[GraphQL] 查询单个组织 - code: %s", args.Code)
	return r.repo.GetOrganization(ctx, DefaultTenantID, args.Code)
}

// 时态查询 - 时间点
func (r *Resolver) OrganizationAtDate(ctx context.Context, args struct {
	Code string
	Date string
}) (*Organization, error) {
	r.logger.Printf("[GraphQL] 时态查询 - code: %s, date: %s", args.Code, args.Date)
	return r.repo.GetOrganizationAtDate(ctx, DefaultTenantID, args.Code, args.Date)
}

// 时态查询 - 历史范围
func (r *Resolver) OrganizationHistory(ctx context.Context, args struct {
	Code     string
	FromDate string
	ToDate   string
}) ([]Organization, error) {
	r.logger.Printf("[GraphQL] 历史查询 - code: %s, range: %s~%s", args.Code, args.FromDate, args.ToDate)
	return r.repo.GetOrganizationHistory(ctx, DefaultTenantID, args.Code, args.FromDate, args.ToDate)
}

// 组织版本查询
func (r *Resolver) OrganizationVersions(ctx context.Context, args struct {
	Code string
}) ([]Organization, error) {
	r.logger.Printf("[GraphQL] 版本查询 - code: %s", args.Code)
	return r.repo.GetOrganizationHistory(ctx, DefaultTenantID, args.Code, "1900-01-01", "2099-12-31")
}

// 组织统计 (camelCase方法名)
func (r *Resolver) OrganizationStats(ctx context.Context) (*OrganizationStats, error) {
	r.logger.Printf("[GraphQL] 统计查询")
	return r.repo.GetOrganizationStats(ctx, DefaultTenantID)
}

func main() {
	logger := log.New(os.Stdout, "[PG-GraphQL] ", log.LstdFlags)
	logger.Println("🚀 启动PostgreSQL原生GraphQL服务")

	// PostgreSQL连接 - 激进优化配置
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "cubecastle")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 连接池优化 - 激进配置
	db.SetMaxOpenConns(100) // 最大连接数
	db.SetMaxIdleConns(25)  // 最大空闲连接
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}
	logger.Println("✅ PostgreSQL连接成功")

	// Redis连接
	redisClient := redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
		DB:   0,
	})

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		logger.Printf("⚠️  Redis连接失败，将不使用缓存: %v", err)
		redisClient = nil
	} else {
		logger.Println("✅ Redis连接成功")
	}

	// 创建仓储和解析器
	repo := NewPostgreSQLRepository(db, redisClient, logger)
	resolver := &Resolver{repo: repo, logger: logger}

	// 创建GraphQL schema
	schema := graphql.MustParseSchema(schemaString, resolver)

	// 初始化JWT中间件
	jwtSecret := getEnv("JWT_SECRET", "cube-castle-development-secret-key-2025")
	jwtIssuer := getEnv("JWT_ISSUER", "cube-castle")
	jwtAudience := getEnv("JWT_AUDIENCE", "cube-castle-api")
	devMode := getEnv("DEV_MODE", "true") == "true"

	jwtMiddleware := auth.NewJWTMiddleware(jwtSecret, jwtIssuer, jwtAudience)
	permissionChecker := auth.NewPBACPermissionChecker(db, logger)
	graphqlMiddleware := auth.NewGraphQLPermissionMiddleware(
		jwtMiddleware,
		permissionChecker,
		logger,
		devMode,
	)

	logger.Printf("🔐 JWT认证初始化完成 (开发模式: %v)", devMode)

	// HTTP路由
	r := chi.NewRouter()

	// 基础中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// GraphQL端点 - 带JWT认证保护
	r.Handle("/graphql", graphqlMiddleware.Middleware()(&relay.Handler{Schema: schema}))

	// GraphiQL开发界面
	r.Get("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		graphiqlHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>GraphiQL - PostgreSQL Native</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphiql@2.4.7/graphiql.min.css" />
    <style>
        body { height: 100%; margin: 0; width: 100%; overflow: hidden; }
        #graphiql { height: 100vh; }
        .graphiql-container { background: #1a1a1a; }
    </style>
</head>
<body>
    <div id="graphiql">Loading PostgreSQL GraphQL...</div>
    <script crossorigin src="https://unpkg.com/react@18/umd/react.development.js"></script>
    <script crossorigin src="https://unpkg.com/react-dom@18/umd/react-dom.development.js"></script>
    <script crossorigin src="https://cdn.jsdelivr.net/npm/graphiql@2.4.7/graphiql.min.js"></script>
    <script>
        const fetcher = GraphiQL.createFetcher({ url: '/graphql' });
        const root = ReactDOM.createRoot(document.getElementById('graphiql'));
        root.render(React.createElement(GraphiQL, { 
            fetcher,
            defaultQuery: '# PostgreSQL原生GraphQL查询\\n# 高性能时态查询示例\\n\\nquery {\\n  organizations(first: 10) {\\n    code\\n    name\\n    status\\n    effective_date\\n    is_current\\n  }\\n}'
        }));
    </script>
</body>
</html>`
		w.Write([]byte(graphiqlHTML))
	})

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "healthy",
			"service":     "postgresql-graphql",
			"timestamp":   time.Now(),
			"database":    "postgresql",
			"performance": "optimized",
		})
	})

	// Prometheus指标
	r.Handle("/metrics", promhttp.Handler())

	// 获取端口
	port := getEnv("PORT", "8090")

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
		// 激进的超时配置
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 优雅关闭
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Println("🛑 正在关闭PostgreSQL GraphQL服务...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("❌ 服务关闭失败: %v", err)
		}
	}()

	logger.Printf("🚀 PostgreSQL原生GraphQL服务启动在端口 :%s", port)
	logger.Println("🔗 GraphiQL界面: http://localhost:" + port + "/graphiql")
	logger.Println("🔗 GraphQL端点: http://localhost:" + port + "/graphql")
	logger.Println("💾 数据库: PostgreSQL (原生优化)")
	logger.Println("⚡ 性能模式: 激进优化")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ 服务启动失败: %v", err)
	}

	logger.Println("✅ PostgreSQL GraphQL服务已安全关闭")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
