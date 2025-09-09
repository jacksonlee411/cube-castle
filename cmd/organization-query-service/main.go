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
	"strings"
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
	"cube-castle-deployment-test/internal/auth"
	"cube-castle-deployment-test/internal/config"
	schemaLoader "cube-castle-deployment-test/internal/graphql"
	requestMiddleware "cube-castle-deployment-test/internal/middleware"
)

// 默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

/**
 * GraphQL Schema单一真源 - Phase 1实施
 * 
 * ⚠️  移除硬编码schemaString，改用docs/api/schema.graphql作为单一真源
 * 消除双源维护漂移风险，确保文档与运行时schema一致性
 * 
 * Schema来源：docs/api/schema.graphql
 * 加载器：internal/graphql/schema_loader.go
 */

// PostgreSQL原生组织模型 - 零转换开销 (camelCase JSON标签)
type Organization struct {
	RecordIDField         string     `json:"recordId" db:"record_id"`
	TenantIDField         string     `json:"tenantId" db:"tenant_id"`
	CodeField             string     `json:"code" db:"code"`
	ParentCodeField       string     `json:"parentCode" db:"parent_code"`
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
	ChangeReasonField     *string    `json:"changeReason" db:"change_reason"`
	DeletedAtField        *time.Time `json:"deletedAt" db:"deleted_at"`
	DeletedByField        *string    `json:"deletedBy" db:"deleted_by"`
	DeletionReasonField   *string    `json:"deletionReason" db:"deletion_reason"`
	SuspendedAtField      *time.Time `json:"suspendedAt" db:"suspended_at"`
	SuspendedByField      *string    `json:"suspendedBy" db:"suspended_by"`
	SuspensionReasonField *string    `json:"suspensionReason" db:"suspension_reason"`
	
	// 新增缺失的字段
	HierarchyDepthField   int      `json:"hierarchyDepth" db:"hierarchy_depth"`
	IsFutureField         bool     `json:"isFuture" db:"is_future"`
}

// GraphQL字段解析器 - 零拷贝优化 (camelCase方法名)
func (o Organization) RecordId() string    { return o.RecordIDField }
func (o Organization) TenantId() string    { return o.TenantIDField }
func (o Organization) Code() string        { return o.CodeField }
func (o Organization) ParentCode() string { return o.ParentCodeField }
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
func (o Organization) CreatedAt() string { return o.CreatedAtField.Format(time.RFC3339) }
func (o Organization) UpdatedAt() string { return o.UpdatedAtField.Format(time.RFC3339) }
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
	return true // 默认为时态数据
}
func (o Organization) ChangeReason() *string { return o.ChangeReasonField }
func (o Organization) HierarchyDepth() int32    { return int32(o.HierarchyDepthField) }
func (o Organization) IsFuture() bool           { return o.IsFutureField }
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

// 层级结构类型 - 严格遵循API规范v4.2.1
type OrganizationHierarchyData struct {
	CodeField           string   `json:"code"`
	NameField           string   `json:"name"`
	LevelField          int      `json:"level"`
	HierarchyDepthField int      `json:"hierarchyDepth"`
	CodePathField       string   `json:"codePath"`
	NamePathField       string   `json:"namePath"`
	ParentChainField    []string `json:"parentChain"`
	ChildrenCountField  int                           `json:"childrenCount"`
	IsRootField         bool                          `json:"isRoot"`
	IsLeafField         bool                          `json:"isLeaf"`
	ChildrenField       []OrganizationHierarchyData   `json:"children"`
}

func (h OrganizationHierarchyData) Code() string           { return h.CodeField }
func (h OrganizationHierarchyData) Name() string           { return h.NameField }
func (h OrganizationHierarchyData) Level() int32           { return int32(h.LevelField) }
func (h OrganizationHierarchyData) HierarchyDepth() int32  { return int32(h.HierarchyDepthField) }
func (h OrganizationHierarchyData) CodePath() string       { return h.CodePathField }
func (h OrganizationHierarchyData) NamePath() string       { return h.NamePathField }
func (h OrganizationHierarchyData) ParentChain() []string  { return h.ParentChainField }
func (h OrganizationHierarchyData) ChildrenCount() int32   { return int32(h.ChildrenCountField) }
func (h OrganizationHierarchyData) IsRoot() bool           { return h.IsRootField }
func (h OrganizationHierarchyData) IsLeaf() bool           { return h.IsLeafField }
func (h OrganizationHierarchyData) Children() []OrganizationHierarchyData { return h.ChildrenField }

type OrganizationSubtreeData struct {
	CodeField           string                    `json:"code"`
	NameField           string                    `json:"name"`
	LevelField          int                       `json:"level"`
	HierarchyDepthField int                       `json:"hierarchyDepth"`
	CodePathField       string                    `json:"codePath"`
	NamePathField       string                    `json:"namePath"`
	ChildrenField       []OrganizationSubtreeData `json:"children"`
}

func (s OrganizationSubtreeData) Code() string                        { return s.CodeField }
func (s OrganizationSubtreeData) Name() string                        { return s.NameField }
func (s OrganizationSubtreeData) Level() int32                        { return int32(s.LevelField) }
func (s OrganizationSubtreeData) HierarchyDepth() int32               { return int32(s.HierarchyDepthField) }
func (s OrganizationSubtreeData) CodePath() string                    { return s.CodePathField }
func (s OrganizationSubtreeData) NamePath() string                    { return s.NamePathField }
func (s OrganizationSubtreeData) Children() []OrganizationSubtreeData { return s.ChildrenField }

// 层级统计类型
type HierarchyStatistics struct {
	TenantIdField           string                `json:"tenantId"`
	TotalOrganizationsField int                   `json:"totalOrganizations"`
	MaxDepthField           int                   `json:"maxDepth"`
	AvgDepthField           float64               `json:"avgDepth"`
	DepthDistributionField  []DepthDistribution   `json:"depthDistribution"`
	RootOrganizationsField  int                   `json:"rootOrganizations"`
	LeafOrganizationsField  int                   `json:"leafOrganizations"`
	IntegrityIssuesField    []IntegrityIssue      `json:"integrityIssues"`
	LastAnalyzedField       string                `json:"lastAnalyzed"`
}

func (h HierarchyStatistics) TenantId() string              { return h.TenantIdField }
func (h HierarchyStatistics) TotalOrganizations() int32     { return int32(h.TotalOrganizationsField) }
func (h HierarchyStatistics) MaxDepth() int32               { return int32(h.MaxDepthField) }
func (h HierarchyStatistics) AvgDepth() float64             { return h.AvgDepthField }
func (h HierarchyStatistics) DepthDistribution() []DepthDistribution { return h.DepthDistributionField }
func (h HierarchyStatistics) RootOrganizations() int32      { return int32(h.RootOrganizationsField) }
func (h HierarchyStatistics) LeafOrganizations() int32      { return int32(h.LeafOrganizationsField) }
func (h HierarchyStatistics) IntegrityIssues() []IntegrityIssue { return h.IntegrityIssuesField }
func (h HierarchyStatistics) LastAnalyzed() string          { return h.LastAnalyzedField }

type DepthDistribution struct {
	DepthField int `json:"depth"`
	CountField int `json:"count"`
}

func (d DepthDistribution) Depth() int32 { return int32(d.DepthField) }
func (d DepthDistribution) Count() int32 { return int32(d.CountField) }

type IntegrityIssue struct {
	TypeField          string   `json:"type"`
	CountField         int      `json:"count"`
	AffectedCodesField []string `json:"affectedCodes"`
}

func (i IntegrityIssue) Type() string         { return i.TypeField }
func (i IntegrityIssue) Count() int32         { return int32(i.CountField) }
func (i IntegrityIssue) AffectedCodes() []string { return i.AffectedCodesField }

// 审计记录类型 - v4.6.0 精确到record_id
type AuditRecordData struct {
	AuditIDField         string         `json:"auditId"`
	RecordIDField        string         `json:"recordId"`
	OperationTypeField   string         `json:"operationType"`
	OperatedByField      OperatedByData `json:"operatedBy"`
	ChangesSummaryField  string         `json:"changesSummary"`
	OperationReasonField *string        `json:"operationReason"`
	TimestampField       string         `json:"timestamp"`
	BeforeDataField      *string        `json:"beforeData"`
	AfterDataField       *string        `json:"afterData"`
}

func (a AuditRecordData) AuditId() string            { return a.AuditIDField }
func (a AuditRecordData) RecordId() string           { return a.RecordIDField }
func (a AuditRecordData) OperationType() string      { return a.OperationTypeField }
func (a AuditRecordData) Operation() string          { return a.OperationTypeField }
func (a AuditRecordData) OperatedBy() OperatedByData { return a.OperatedByField }
func (a AuditRecordData) ChangesSummary() string     { return a.ChangesSummaryField }
func (a AuditRecordData) OperationReason() *string   { return a.OperationReasonField }
func (a AuditRecordData) Timestamp() string          { return a.TimestampField }
func (a AuditRecordData) BeforeData() *string {
	if a.BeforeDataField == nil {
		return nil
	}
	// 确保空对象也返回，不要过滤为null
	return a.BeforeDataField
}
func (a AuditRecordData) AfterData() *string {
	if a.AfterDataField == nil {
		return nil
	}
	// 确保空对象也返回，不要过滤为null
	return a.AfterDataField
}

type OperatedByData struct {
	IDField   string `json:"id"`
	NameField string `json:"name"`
}

func (o OperatedByData) Id() string   { return o.IDField }
func (o OperatedByData) Name() string { return o.NameField }

// DateRangeInput GraphQL输入类型
type DateRangeInput struct {
	From *string `json:"from"`
	To   *string `json:"to"`
}

// 输入类型 - 符合官方API契约 (P0阶段最小实现)
type OrganizationFilter struct {
	// Temporal Filtering
	AsOfDate      *string `json:"asOfDate"`
	IncludeFuture bool    `json:"includeFuture"`
	OnlyFuture    bool    `json:"onlyFuture"`
	
	// Business Filtering
	UnitType   *string   `json:"unitType"`
	Status     *string   `json:"status"`
	ParentCode *string   `json:"parentCode"`
	Codes      *[]string `json:"codes"`
	
	// Hierarchy Filtering
	Level      *int32 `json:"level"`
	MinLevel   *int32 `json:"minLevel"`
	MaxLevel   *int32 `json:"maxLevel"`
	RootsOnly  bool   `json:"rootsOnly"`
	LeavesOnly bool   `json:"leavesOnly"` 
	
	// Text Search
	SearchText   *string   `json:"searchText"`
	SearchFields []string `json:"searchFields"`
	
	// Advanced Filtering
	HasChildren     *bool   `json:"hasChildren"`
	HasProfile      *bool   `json:"hasProfile"`
	ProfileContains *string `json:"profileContains"`
	
	// Audit Filtering - 修复类型匹配问题
	OperationType       *string          `json:"operationType"`
	OperatedBy          *string          `json:"operatedBy"`
	OperationDateRange  *DateRangeInput  `json:"operationDateRange"`
}

type PaginationInput struct {
	Page      int32  `json:"page"`
	PageSize  int32  `json:"pageSize"`
	SortBy    string `json:"sortBy"`
	SortOrder string `json:"sortOrder"`
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
		if pagination.Page > 0 {
			page = pagination.Page
		}
		if pagination.PageSize > 0 {
			pageSize = pagination.PageSize
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
		var isTemporal bool
		err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, &isTemporal,
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
        WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED' AND deleted_at IS NULL
        LIMIT 1`

	start := time.Now()
	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code)

	var org Organization
	var isTemporal bool
	err := row.Scan(
		&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
		&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
		&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
		&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, &isTemporal,
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
    // 使用计算的区间终点（computed_end_date），避免依赖物理 end_date 的准确性
    query := `
        WITH hist AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, path, sort_order, description, profile, created_at, updated_at,
                effective_date, end_date, is_current, is_temporal, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
                LEAD(effective_date) OVER (PARTITION BY tenant_id, code ORDER BY effective_date) AS next_effective
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 
              AND status <> 'DELETED' AND deleted_at IS NULL
        ), proj AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, path, sort_order, description, profile, created_at, updated_at,
                effective_date,
                COALESCE(end_date, (next_effective - INTERVAL '1 day')::date) AS computed_end_date,
                is_current, is_temporal, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
            FROM hist
        )
        SELECT 
            record_id, tenant_id, code, parent_code, name, unit_type, status,
            level, path, sort_order, description, profile, created_at, updated_at,
            effective_date, computed_end_date AS end_date, is_current, is_temporal, change_reason,
            deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
        FROM proj
        WHERE effective_date <= $3::date 
          AND (computed_end_date IS NULL OR computed_end_date >= $3::date)
        ORDER BY effective_date DESC, created_at DESC
        LIMIT 1`

	start := time.Now()
	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code, date)

	var org Organization
	var isTemporal bool
	err := row.Scan(
		&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
		&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
		&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
		&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, &isTemporal,
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
    // 历史范围查询：使用计算的区间终点（computed_end_date）并基于区间重叠选择
    query := `
        WITH hist AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, path, sort_order, description, profile, created_at, updated_at,
                effective_date, end_date, is_current, is_temporal, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
                LEAD(effective_date) OVER (PARTITION BY tenant_id, code ORDER BY effective_date) AS next_effective
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 
              AND status <> 'DELETED' AND deleted_at IS NULL
        ), proj AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, path, sort_order, description, profile, created_at, updated_at,
                effective_date,
                COALESCE(end_date, (next_effective - INTERVAL '1 day')::date) AS computed_end_date,
                is_current, is_temporal, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
            FROM hist
        )
        SELECT 
            record_id, tenant_id, code, parent_code, name, unit_type, status,
            level, path, sort_order, description, profile, created_at, updated_at,
            effective_date, computed_end_date AS end_date, is_current, is_temporal, change_reason,
            deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
        FROM proj
        WHERE effective_date <= $4::date
          AND (computed_end_date IS NULL OR computed_end_date >= $3::date)
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
		var isTemporal bool
		err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, &isTemporal,
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
            FROM organization_units WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED' AND deleted_at IS NULL
        ),
        type_stats AS (
            SELECT unit_type, COUNT(*) as count
            FROM organization_units 
            WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED' AND deleted_at IS NULL
            GROUP BY unit_type
        ),
        status_detail_stats AS (
            SELECT status, COUNT(*) as count
            FROM organization_units 
            WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED' AND deleted_at IS NULL
            GROUP BY status
        ),
        level_stats AS (
            SELECT level, COUNT(*) as count
            FROM organization_units 
            WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED' AND deleted_at IS NULL
            GROUP BY level
        ),
        temporal_stats AS (
            SELECT 
                COUNT(*) as total_versions,
                COUNT(DISTINCT code) as unique_orgs,
                MIN(effective_date) as oldest_date,
                MAX(effective_date) as newest_date
            FROM organization_units WHERE tenant_id = $1 AND (status <> 'DELETED' AND deleted_at IS NULL)
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

// 高级层级结构查询 - 严格遵循API规范v4.2.1
func (r *PostgreSQLRepository) GetOrganizationHierarchy(ctx context.Context, tenantID uuid.UUID, code string) (*OrganizationHierarchyData, error) {
	start := time.Now()
	
	// 使用PostgreSQL递归CTE查询完整层级信息
    query := `
        WITH RECURSIVE hierarchy_info AS (
            -- 获取目标组织
            SELECT 
                code, name, level, 
                COALESCE(code_path, '/' || code) as code_path,
                COALESCE(name_path, '/' || name) as name_path,
                parent_code,
                1 as hierarchy_depth
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED' AND deleted_at IS NULL
            
            UNION ALL
            
            -- 递归获取父级信息
            SELECT 
                o.code, o.name, o.level,
                o.code_path, o.name_path, o.parent_code,
                h.hierarchy_depth + 1
            FROM organization_units o
            INNER JOIN hierarchy_info h ON o.code = h.parent_code
            WHERE o.tenant_id = $1 AND o.is_current = true AND o.status <> 'DELETED' AND o.deleted_at IS NULL
        ),
        children_count AS (
            SELECT COUNT(*) as count
            FROM organization_units
            WHERE tenant_id = $1 AND parent_code = $2 AND is_current = true AND status <> 'DELETED' AND deleted_at IS NULL
        )
		SELECT 
			h.code, h.name, h.level, h.hierarchy_depth,
			h.code_path, h.name_path,
			CASE WHEN h.parent_code IS NULL THEN '{}' 
				 ELSE json_build_array(h.parent_code)::text END as parent_chain,
			c.count as children_count,
			(h.parent_code IS NULL) as is_root,
			(c.count = 0) as is_leaf
		FROM hierarchy_info h
		CROSS JOIN children_count c
		WHERE h.code = $2
		LIMIT 1`
	
	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code)
	
	var hierarchy OrganizationHierarchyData
	var parentChainJSON string
	
	err := row.Scan(
		&hierarchy.CodeField, &hierarchy.NameField, &hierarchy.LevelField, &hierarchy.HierarchyDepthField,
		&hierarchy.CodePathField, &hierarchy.NamePathField, &parentChainJSON,
		&hierarchy.ChildrenCountField, &hierarchy.IsRootField, &hierarchy.IsLeafField,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Printf("[ERROR] 层级结构查询失败: %v", err)
		return nil, err
	}
	
	// 解析父级链
	if parentChainJSON != "{}" {
		json.Unmarshal([]byte(parentChainJSON), &hierarchy.ParentChainField)
	} else {
		hierarchy.ParentChainField = []string{}
	}
	
	duration := time.Since(start)
	r.logger.Printf("[PERF] 层级结构查询完成，耗时: %v", duration)
	
	return &hierarchy, nil
}

// 组织子树查询 - 严格遵循API规范v4.2.1
func (r *PostgreSQLRepository) GetOrganizationSubtree(ctx context.Context, tenantID uuid.UUID, code string, maxDepth int) (*OrganizationSubtreeData, error) {
	start := time.Now()
	
	// 使用PostgreSQL递归CTE查询子树结构，限制深度
    query := `
        WITH RECURSIVE subtree AS (
            -- 根节点
            SELECT 
                code, name, level, 
                COALESCE(hierarchy_depth, level) as hierarchy_depth,
                COALESCE(code_path, '/' || code) as code_path,
                COALESCE(name_path, '/' || name) as name_path,
                parent_code,
                0 as depth_from_root
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED' AND deleted_at IS NULL
            
            UNION ALL
            
            -- 递归查询子节点
            SELECT 
                o.code, o.name, o.level,
                o.hierarchy_depth, o.code_path, o.name_path, o.parent_code,
                s.depth_from_root + 1
            FROM organization_units o
            INNER JOIN subtree s ON o.parent_code = s.code
            WHERE o.tenant_id = $1 AND o.is_current = true AND o.status <> 'DELETED' AND o.deleted_at IS NULL
              AND s.depth_from_root < $3
        )
		SELECT code, name, level, hierarchy_depth, code_path, name_path, parent_code
		FROM subtree 
		ORDER BY level, code`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), code, maxDepth)
	if err != nil {
		r.logger.Printf("[ERROR] 子树查询失败: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	// 构建树形结构
	nodeMap := make(map[string]*OrganizationSubtreeData)
	var root *OrganizationSubtreeData
	
	for rows.Next() {
		var node OrganizationSubtreeData
		var parentCode *string
		
		err := rows.Scan(
			&node.CodeField, &node.NameField, &node.LevelField, &node.HierarchyDepthField,
			&node.CodePathField, &node.NamePathField, &parentCode,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描子树数据失败: %v", err)
			return nil, err
		}
		
		node.ChildrenField = []OrganizationSubtreeData{}
		nodeMap[node.CodeField] = &node
		
		if node.CodeField == code {
			root = &node
		}
	}
	
	// 构建父子关系
	for _, node := range nodeMap {
		if root != nil && node.CodeField != code {
			// 寻找父节点并添加到其children中
			for _, parent := range nodeMap {
				if node.CodeField != parent.CodeField {
					// 检查是否为直接子节点（通过codePath判断）
					if strings.HasPrefix(node.CodePathField, parent.CodePathField+"/") {
						// 计算层级差，确保是直接子节点
						parentDepth := strings.Count(parent.CodePathField, "/")
						nodeDepth := strings.Count(node.CodePathField, "/")
						if nodeDepth == parentDepth+1 {
							parent.ChildrenField = append(parent.ChildrenField, *node)
							break
						}
					}
				}
			}
		}
	}
	
	duration := time.Since(start)
	r.logger.Printf("[PERF] 子树查询完成，返回 %d 节点，耗时: %v", len(nodeMap), duration)
	
	return root, nil
}

// 审计历史查询 - v4.6.0 基于record_id精确查询
func (r *PostgreSQLRepository) GetAuditHistory(ctx context.Context, recordId string, startDate, endDate, operation, userId *string, limit int) ([]AuditRecordData, error) {
	start := time.Now()
	
	// 构建查询条件 - 基于record_id查询
	baseQuery := `
		SELECT 
			id as audit_id, 
			resource_id as record_id, 
			event_type as operation_type,
			actor_id as operated_by_id, 
			CASE WHEN business_context->>'actor_name' IS NOT NULL 
				THEN business_context->>'actor_name' 
				ELSE actor_id 
			END as operated_by_name,
			CASE WHEN changes IS NOT NULL 
				THEN changes::text 
				ELSE '{"operationSummary":"' || action_name || '","totalChanges":0,"keyChanges":[]}' 
			END as changes_summary,
			business_context->>'operation_reason' as operation_reason,
			timestamp,
			before_data::text as before_data, 
			after_data::text as after_data
		FROM audit_logs 
		WHERE resource_id = $1::uuid AND resource_type = 'ORGANIZATION'`
	
	args := []interface{}{recordId}
	argIndex := 2
	
	// 日期范围过滤
	if startDate != nil {
		baseQuery += fmt.Sprintf(" AND timestamp >= $%d::timestamp", argIndex)
		args = append(args, *startDate)
		argIndex++
	}
	
	if endDate != nil {
		baseQuery += fmt.Sprintf(" AND timestamp <= $%d::timestamp", argIndex)
		args = append(args, *endDate)
		argIndex++
	}
	
	// 操作类型过滤
	if operation != nil {
		baseQuery += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, strings.ToUpper(*operation))
		argIndex++
	}
	
	// 操作人过滤
	if userId != nil {
		baseQuery += fmt.Sprintf(" AND actor_id = $%d", argIndex)
		args = append(args, *userId)
		argIndex++
	}
	
	// 排序和限制
	finalQuery := baseQuery + fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d", argIndex)
	args = append(args, limit)
	
	rows, err := r.db.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		r.logger.Printf("[ERROR] 审计历史查询失败: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	var auditRecords []AuditRecordData
	for rows.Next() {
		var record AuditRecordData
		var operatedById, operatedByName string
		var beforeData, afterData sql.NullString
		
		err := rows.Scan(
			&record.AuditIDField, &record.RecordIDField, &record.OperationTypeField,
			&operatedById, &operatedByName,
			&record.ChangesSummaryField, &record.OperationReasonField, &record.TimestampField,
			&beforeData, &afterData,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描审计记录失败: %v", err)
			return nil, err
		}
		
		// 正确处理JSONB字段
		if beforeData.Valid {
			record.BeforeDataField = &beforeData.String
		}
		if afterData.Valid {
			record.AfterDataField = &afterData.String
		}
		
		// 构建操作人信息
		record.OperatedByField = OperatedByData{
			IDField:   operatedById,
			NameField: operatedByName,
		}
		
		auditRecords = append(auditRecords, record)
	}
	
	duration := time.Since(start)
	r.logger.Printf("[PERF] record_id审计查询完成，返回 %d 条记录，耗时: %v", len(auditRecords), duration)
	
	return auditRecords, nil
}

// 单条审计记录查询 - v4.6.0
func (r *PostgreSQLRepository) GetAuditLog(ctx context.Context, auditId string) (*AuditRecordData, error) {
	start := time.Now()
	
	query := `
		SELECT 
			id as audit_id, 
			resource_id as record_id, 
			event_type as operation_type,
			actor_id as operated_by_id, 
			CASE WHEN business_context->>'actor_name' IS NOT NULL 
				THEN business_context->>'actor_name' 
				ELSE actor_id 
			END as operated_by_name,
			CASE WHEN changes IS NOT NULL 
				THEN changes::text 
				ELSE '{"operationSummary":"' || action_name || '","totalChanges":0,"keyChanges":[]}' 
			END as changes_summary,
			business_context->>'operation_reason' as operation_reason,
			timestamp,
			before_data::text as before_data, 
			after_data::text as after_data
		FROM audit_logs 
		WHERE id = $1::uuid AND resource_type = 'ORGANIZATION'
		LIMIT 1`
	
	row := r.db.QueryRowContext(ctx, query, auditId)
	
	var record AuditRecordData
	var operatedById, operatedByName string
	var beforeData, afterData sql.NullString
	
	err := row.Scan(
		&record.AuditIDField, &record.RecordIDField, &record.OperationTypeField,
		&operatedById, &operatedByName,
		&record.ChangesSummaryField, &record.OperationReasonField, &record.TimestampField,
		&beforeData, &afterData,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Printf("[ERROR] 单条审计记录查询失败: %v", err)
		return nil, err
	}
	
	// 正确处理JSONB字段
	if beforeData.Valid {
		record.BeforeDataField = &beforeData.String
	}
	if afterData.Valid {
		record.AfterDataField = &afterData.String
	}
	
	// 构建操作人信息
	record.OperatedByField = OperatedByData{
		IDField:   operatedById,
		NameField: operatedByName,
	}
	
	duration := time.Since(start)
	r.logger.Printf("[PERF] 单条审计记录查询完成，耗时: %v", duration)
	
	return &record, nil
}

// GraphQL解析器 - 极简高效
type Resolver struct {
    repo   *PostgreSQLRepository
    logger *log.Logger
    authMW *auth.GraphQLPermissionMiddleware
}

// 当前组织列表查询 - 符合API契约v4.2.1 (camelCase方法名)
func (r *Resolver) Organizations(ctx context.Context, args struct {
    Filter     *OrganizationFilter
    Pagination *PaginationInput
}) (*OrganizationConnection, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "organizations"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: organizations: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
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
    Code     string
    AsOfDate *string
}) (*Organization, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "organization"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: organization: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
    r.logger.Printf("[GraphQL] 查询单个组织 - code: %s", args.Code)
    return r.repo.GetOrganization(ctx, DefaultTenantID, args.Code)
}

// 时态查询 - 时间点
func (r *Resolver) OrganizationAtDate(ctx context.Context, args struct {
    Code string
    Date string
}) (*Organization, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "organizationAtDate"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: organizationAtDate: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
    r.logger.Printf("[GraphQL] 时态查询 - code: %s, date: %s", args.Code, args.Date)
    return r.repo.GetOrganizationAtDate(ctx, DefaultTenantID, args.Code, args.Date)
}

// 时态查询 - 历史范围
func (r *Resolver) OrganizationHistory(ctx context.Context, args struct {
    Code     string
    FromDate string
    ToDate   string
}) ([]Organization, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "organizationHistory"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: organizationHistory: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
    r.logger.Printf("[GraphQL] 历史查询 - code: %s, range: %s~%s", args.Code, args.FromDate, args.ToDate)
    return r.repo.GetOrganizationHistory(ctx, DefaultTenantID, args.Code, args.FromDate, args.ToDate)
}

// 组织版本查询
func (r *Resolver) OrganizationVersions(ctx context.Context, args struct {
    Code string
}) ([]Organization, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "organizationVersions"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: organizationVersions: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
    r.logger.Printf("[GraphQL] 版本查询 - code: %s", args.Code)
    return r.repo.GetOrganizationHistory(ctx, DefaultTenantID, args.Code, "1900-01-01", "2099-12-31")
}

// 组织统计 (camelCase方法名)
func (r *Resolver) OrganizationStats(ctx context.Context, args struct {
    AsOfDate          *string
    IncludeHistorical bool
}) (*OrganizationStats, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "organizationStats"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: organizationStats: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
    r.logger.Printf("[GraphQL] 统计查询")
    return r.repo.GetOrganizationStats(ctx, DefaultTenantID)
}

// 高级层级结构查询 - 严格遵循API规范v4.2.1
func (r *Resolver) OrganizationHierarchy(ctx context.Context, args struct {
    Code     string
    TenantId string
}) (*OrganizationHierarchyData, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "organizationHierarchy"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: organizationHierarchy: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
    r.logger.Printf("[GraphQL] 层级结构查询 - code: %s, tenantId: %s", args.Code, args.TenantId)
	
	tenantID, err := uuid.Parse(args.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}
	
	return r.repo.GetOrganizationHierarchy(ctx, tenantID, args.Code)
}

func (r *Resolver) OrganizationSubtree(ctx context.Context, args struct {
    Code            string
    TenantId        string
    MaxDepth        int32
    IncludeInactive bool
}) ([]OrganizationHierarchyData, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "organizationSubtree"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: organizationSubtree: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
    r.logger.Printf("[GraphQL] 子树查询 - code: %s, tenantId: %s, maxDepth: %v", args.Code, args.TenantId, args.MaxDepth)
	
	tenantID, err := uuid.Parse(args.TenantId)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}
	
	maxDepth := 10 // 默认深度
	if args.MaxDepth > 0 {
		maxDepth = int(args.MaxDepth)
	}
	
	subtree, err := r.repo.GetOrganizationSubtree(ctx, tenantID, args.Code, maxDepth)
	if err != nil {
		return nil, err
	}
	
	// 将单个子树转换为数组（Schema期望数组返回）
	if subtree == nil {
		return []OrganizationHierarchyData{}, nil
	}
	
	// 先转换根节点
	root := OrganizationHierarchyData{
		CodeField:           subtree.CodeField,
		NameField:           subtree.NameField,
		LevelField:          subtree.LevelField,
		HierarchyDepthField: subtree.HierarchyDepthField,
		CodePathField:       subtree.CodePathField,
		NamePathField:       subtree.NamePathField,
		ParentChainField:    []string{}, // 根节点没有父级链
		ChildrenCountField:  len(subtree.ChildrenField),
		IsRootField:         subtree.LevelField == 1,
		IsLeafField:         len(subtree.ChildrenField) == 0,
		ChildrenField:       []OrganizationHierarchyData{}, // 简化实现，先不递归转换
	}
	
	return []OrganizationHierarchyData{root}, nil
}

// 层级统计查询
func (r *Resolver) HierarchyStatistics(ctx context.Context, args struct {
    TenantId                string
    IncludeIntegrityCheck   bool
}) (*HierarchyStatistics, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "hierarchyStatistics"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: hierarchyStatistics: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
    
    // TODO: 实现实际的层级统计逻辑
    return &HierarchyStatistics{
        TenantIdField:           args.TenantId,
        TotalOrganizationsField: 0,
        MaxDepthField:           0,
        AvgDepthField:           0.0,
        DepthDistributionField:  []DepthDistribution{},
        RootOrganizationsField:  0,
        LeafOrganizationsField:  0,
        IntegrityIssuesField:    []IntegrityIssue{},
        LastAnalyzedField:       "",
    }, nil
}

// 审计历史查询 - v4.6.0 基于record_id
func (r *Resolver) AuditHistory(ctx context.Context, args struct {
    RecordId  string
    StartDate *string
    EndDate   *string
    Operation *string
    UserId    *string
    Limit     int32
}) ([]AuditRecordData, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "auditHistory"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: auditHistory: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
    r.logger.Printf("[GraphQL] 审计历史查询 - recordId: %s", args.RecordId)
	
	limit := int32(50) // 默认限制
	if args.Limit > 0 {
		limit = args.Limit
		if limit > 200 { // API规范限制最大200
			limit = 200
		}
	}
	
	return r.repo.GetAuditHistory(ctx, args.RecordId, args.StartDate, args.EndDate, args.Operation, args.UserId, int(limit))
}

// 单条审计记录查询 - v4.6.0
func (r *Resolver) AuditLog(ctx context.Context, args struct {
    AuditId string
}) (*AuditRecordData, error) {
    if err := r.authMW.CheckQueryPermission(ctx, "auditLog"); err != nil {
        r.logger.Printf("[AUTH] 权限拒绝: auditLog: %v", err)
        return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
    }
    r.logger.Printf("[GraphQL] 单条审计记录查询 - auditId: %s", args.AuditId)
    return r.repo.GetAuditLog(ctx, args.AuditId)
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

	// 创建仓储
	repo := NewPostgreSQLRepository(db, redisClient, logger)

    // 初始化JWT中间件 - 使用统一配置
    jwtConfig := config.GetJWTConfig()
    devMode := getEnv("DEV_MODE", "true") == "true"
    
    var pubPEM []byte
    if jwtConfig.HasPublicKey() {
        if b, err := os.ReadFile(jwtConfig.PublicKeyPath); err == nil { 
            pubPEM = b 
        }
    }

    jwtMiddleware := auth.NewJWTMiddlewareWithOptions(jwtConfig.Secret, jwtConfig.Issuer, jwtConfig.Audience, auth.Options{
        Alg:          jwtConfig.Algorithm,
        JWKSURL:      jwtConfig.JWKSUrl,
        PublicKeyPEM: pubPEM,
        ClockSkew:    jwtConfig.AllowedClockSkew,
    })
	permissionChecker := auth.NewPBACPermissionChecker(db, logger)
	graphqlMiddleware := auth.NewGraphQLPermissionMiddleware(
		jwtMiddleware,
		permissionChecker,
		logger,
		devMode,
	)

	logger.Printf("🔐 JWT认证初始化完成 (开发模式: %v)", devMode)

	// 创建解析器（注入权限中间件）
	resolver := &Resolver{repo: repo, logger: logger, authMW: graphqlMiddleware}

	// 🎯 Phase 1: 创建GraphQL schema - 单一真源加载
	// 从docs/api/schema.graphql加载schema，消除双源维护漂移
	schemaPath := schemaLoader.GetDefaultSchemaPath()
	schemaString := schemaLoader.MustLoadSchema(schemaPath)
	schema := graphql.MustParseSchema(schemaString, resolver)
	
	logger.Printf("✅ GraphQL Schema loaded from single source: %s", schemaPath)

	// HTTP路由
	r := chi.NewRouter()

	// 基础中间件
	r.Use(requestMiddleware.RequestIDMiddleware) // 请求追踪中间件
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

	// 创建企业级响应信封中间件
	envelopeMiddleware := requestMiddleware.NewGraphQLEnvelopeMiddleware()

	// GraphQL端点 - 带JWT认证保护和企业级响应信封
	graphqlHandler := envelopeMiddleware.Middleware()(graphqlMiddleware.Middleware()(&relay.Handler{Schema: schema}))
	r.Handle("/graphql", graphqlHandler)

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
