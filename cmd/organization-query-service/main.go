package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"cube-castle-deployment-test/internal/auth"
	"cube-castle-deployment-test/internal/config"
	schemaLoader "cube-castle-deployment-test/internal/graphql"
	requestMiddleware "cube-castle-deployment-test/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	pq "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
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
	ChangeReasonField     *string    `json:"changeReason" db:"change_reason"`
	DeletedAtField        *time.Time `json:"deletedAt" db:"deleted_at"`
	DeletedByField        *string    `json:"deletedBy" db:"deleted_by"`
	DeletionReasonField   *string    `json:"deletionReason" db:"deletion_reason"`
	SuspendedAtField      *time.Time `json:"suspendedAt" db:"suspended_at"`
	SuspendedByField      *string    `json:"suspendedBy" db:"suspended_by"`
	SuspensionReasonField *string    `json:"suspensionReason" db:"suspension_reason"`

	// 新增缺失的字段
	HierarchyDepthField int `json:"hierarchyDepth" db:"hierarchy_depth"`
	ChildrenCountField  int `json:"childrenCount" db:"children_count"`
}

func clampToInt32(value int) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}
	return int32(value)
}

func clampToInt32Ptr(src *int) *int32 {
	if src == nil {
		return nil
	}
	val := clampToInt32(*src)
	return &val
}

// GraphQL字段解析器 - 零拷贝优化 (camelCase方法名)
func (o Organization) RecordId() string { return o.RecordIDField }
func (o Organization) TenantId() string { return o.TenantIDField }
func (o Organization) Code() string     { return o.CodeField }
func (o Organization) ParentCode() string {
	if o.ParentCodeField == nil {
		return "0" // 根组织使用 "0" 作为 parentCode
	}
	return *o.ParentCodeField
}
func (o Organization) Name() string     { return o.NameField }
func (o Organization) UnitType() string { return o.UnitTypeField }
func (o Organization) Status() string   { return o.StatusField }
func (o Organization) Level() int32     { return clampToInt32(o.LevelField) }
func (o Organization) Path() *string    { return o.PathField }
func (o Organization) SortOrder() *int32 {
	return clampToInt32Ptr(o.SortOrderField)
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
	// 派生：有结束日期即为历史时态
	return o.EndDateField != nil
}
func (o Organization) ChangeReason() *string { return o.ChangeReasonField }
func (o Organization) HierarchyDepth() int32 { return clampToInt32(o.HierarchyDepthField) }
func (o Organization) ChildrenCount() int32  { return clampToInt32(o.ChildrenCountField) }
func cnTodayDate() time.Time {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 回退到 UTC，但这在部署环境应始终可用
		return time.Now().UTC().Truncate(24 * time.Hour)
	}
	nowCN := time.Now().In(loc)
	return time.Date(nowCN.Year(), nowCN.Month(), nowCN.Day(), 0, 0, 0, 0, loc)
}
func (o Organization) IsFuture() bool {
	// 派生：effectiveDate > 今日（北京时间）
	todayCN := cnTodayDate()
	eff := time.Date(o.EffectiveDateField.Year(), o.EffectiveDateField.Month(), o.EffectiveDateField.Day(), 0, 0, 0, 0, todayCN.Location())
	return eff.After(todayCN)
}
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

func (s OrganizationStats) TotalCount() int32            { return clampToInt32(s.TotalCountField) }
func (s OrganizationStats) ActiveCount() int32           { return clampToInt32(s.ActiveCountField) }
func (s OrganizationStats) InactiveCount() int32         { return clampToInt32(s.InactiveCountField) }
func (s OrganizationStats) PlannedCount() int32          { return clampToInt32(s.PlannedCountField) }
func (s OrganizationStats) DeletedCount() int32          { return clampToInt32(s.DeletedCountField) }
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

func (t TemporalStats) TotalVersions() int32           { return clampToInt32(t.TotalVersionsField) }
func (t TemporalStats) AverageVersionsPerOrg() float64 { return t.AverageVersionsPerOrgField }
func (t TemporalStats) OldestEffectiveDate() string    { return t.OldestEffectiveDateField }
func (t TemporalStats) NewestEffectiveDate() string    { return t.NewestEffectiveDateField }

type TypeCount struct {
	UnitTypeField string `json:"unitType"`
	CountField    int    `json:"count"`
}

func (t TypeCount) UnitType() string { return t.UnitTypeField }
func (t TypeCount) Count() int32     { return clampToInt32(t.CountField) }

type LevelCount struct {
	LevelField int `json:"level"`
	CountField int `json:"count"`
}

func (l LevelCount) Level() int32 { return clampToInt32(l.LevelField) }
func (l LevelCount) Count() int32 { return clampToInt32(l.CountField) }

type StatusCount struct {
	StatusField string `json:"status"`
	CountField  int    `json:"count"`
}

func (s StatusCount) Status() string { return s.StatusField }
func (s StatusCount) Count() int32   { return clampToInt32(s.CountField) }

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

func (p PaginationInfo) Total() int32      { return clampToInt32(p.TotalField) }
func (p PaginationInfo) Page() int32       { return clampToInt32(p.PageField) }
func (p PaginationInfo) PageSize() int32   { return clampToInt32(p.PageSizeField) }
func (p PaginationInfo) HasNext() bool     { return p.HasNextField }
func (p PaginationInfo) HasPrevious() bool { return p.HasPreviousField }

type TemporalInfo struct {
	AsOfDateField        string `json:"asOfDate"`
	CurrentCountField    int    `json:"currentCount"`
	FutureCountField     int    `json:"futureCount"`
	HistoricalCountField int    `json:"historicalCount"`
}

func (t TemporalInfo) AsOfDate() string       { return t.AsOfDateField }
func (t TemporalInfo) CurrentCount() int32    { return clampToInt32(t.CurrentCountField) }
func (t TemporalInfo) FutureCount() int32     { return clampToInt32(t.FutureCountField) }
func (t TemporalInfo) HistoricalCount() int32 { return clampToInt32(t.HistoricalCountField) }

// 层级结构类型 - 严格遵循API规范v4.2.1
type OrganizationHierarchyData struct {
	CodeField           string                      `json:"code"`
	NameField           string                      `json:"name"`
	LevelField          int                         `json:"level"`
	HierarchyDepthField int                         `json:"hierarchyDepth"`
	CodePathField       string                      `json:"codePath"`
	NamePathField       string                      `json:"namePath"`
	ParentChainField    []string                    `json:"parentChain"`
	ChildrenCountField  int                         `json:"childrenCount"`
	IsRootField         bool                        `json:"isRoot"`
	IsLeafField         bool                        `json:"isLeaf"`
	ChildrenField       []OrganizationHierarchyData `json:"children"`
}

func (h OrganizationHierarchyData) Code() string                          { return h.CodeField }
func (h OrganizationHierarchyData) Name() string                          { return h.NameField }
func (h OrganizationHierarchyData) Level() int32                          { return clampToInt32(h.LevelField) }
func (h OrganizationHierarchyData) HierarchyDepth() int32                 { return clampToInt32(h.HierarchyDepthField) }
func (h OrganizationHierarchyData) CodePath() string                      { return h.CodePathField }
func (h OrganizationHierarchyData) NamePath() string                      { return h.NamePathField }
func (h OrganizationHierarchyData) ParentChain() []string                 { return h.ParentChainField }
func (h OrganizationHierarchyData) ChildrenCount() int32                  { return clampToInt32(h.ChildrenCountField) }
func (h OrganizationHierarchyData) IsRoot() bool                          { return h.IsRootField }
func (h OrganizationHierarchyData) IsLeaf() bool                          { return h.IsLeafField }
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
func (s OrganizationSubtreeData) Level() int32                        { return clampToInt32(s.LevelField) }
func (s OrganizationSubtreeData) HierarchyDepth() int32               { return clampToInt32(s.HierarchyDepthField) }
func (s OrganizationSubtreeData) CodePath() string                    { return s.CodePathField }
func (s OrganizationSubtreeData) NamePath() string                    { return s.NamePathField }
func (s OrganizationSubtreeData) Children() []OrganizationSubtreeData { return s.ChildrenField }

// 层级统计类型
type HierarchyStatistics struct {
	TenantIdField           string              `json:"tenantId"`
	TotalOrganizationsField int                 `json:"totalOrganizations"`
	MaxDepthField           int                 `json:"maxDepth"`
	AvgDepthField           float64             `json:"avgDepth"`
	DepthDistributionField  []DepthDistribution `json:"depthDistribution"`
	RootOrganizationsField  int                 `json:"rootOrganizations"`
	LeafOrganizationsField  int                 `json:"leafOrganizations"`
	IntegrityIssuesField    []IntegrityIssue    `json:"integrityIssues"`
	LastAnalyzedField       string              `json:"lastAnalyzed"`
}

func (h HierarchyStatistics) TenantId() string { return h.TenantIdField }
func (h HierarchyStatistics) TotalOrganizations() int32 {
	return clampToInt32(h.TotalOrganizationsField)
}
func (h HierarchyStatistics) MaxDepth() int32                        { return clampToInt32(h.MaxDepthField) }
func (h HierarchyStatistics) AvgDepth() float64                      { return h.AvgDepthField }
func (h HierarchyStatistics) DepthDistribution() []DepthDistribution { return h.DepthDistributionField }
func (h HierarchyStatistics) RootOrganizations() int32               { return clampToInt32(h.RootOrganizationsField) }
func (h HierarchyStatistics) LeafOrganizations() int32               { return clampToInt32(h.LeafOrganizationsField) }
func (h HierarchyStatistics) IntegrityIssues() []IntegrityIssue      { return h.IntegrityIssuesField }
func (h HierarchyStatistics) LastAnalyzed() string                   { return h.LastAnalyzedField }

type DepthDistribution struct {
	DepthField int `json:"depth"`
	CountField int `json:"count"`
}

func (d DepthDistribution) Depth() int32 { return clampToInt32(d.DepthField) }
func (d DepthDistribution) Count() int32 { return clampToInt32(d.CountField) }

type IntegrityIssue struct {
	TypeField          string   `json:"type"`
	CountField         int      `json:"count"`
	AffectedCodesField []string `json:"affectedCodes"`
}

func (i IntegrityIssue) Type() string            { return i.TypeField }
func (i IntegrityIssue) Count() int32            { return clampToInt32(i.CountField) }
func (i IntegrityIssue) AffectedCodes() []string { return i.AffectedCodesField }

// 字段变更详细信息
type FieldChangeData struct {
	FieldField    string      `json:"field"`
	OldValueField interface{} `json:"oldValue"`
	NewValueField interface{} `json:"newValue"`
	DataTypeField string      `json:"dataType"`
}

func (f FieldChangeData) Field() string { return f.FieldField }
func (f FieldChangeData) OldValue() *string {
	if f.OldValueField == nil {
		return nil
	}
	// 将interface{}转换为字符串
	if str, ok := f.OldValueField.(string); ok {
		return &str
	}
	// 对于其他类型，序列化为JSON字符串
	jsonBytes, _ := json.Marshal(f.OldValueField)
	jsonStr := string(jsonBytes)
	return &jsonStr
}
func (f FieldChangeData) NewValue() *string {
	if f.NewValueField == nil {
		return nil
	}
	// 将interface{}转换为字符串
	if str, ok := f.NewValueField.(string); ok {
		return &str
	}
	// 对于其他类型，序列化为JSON字符串
	jsonBytes, _ := json.Marshal(f.NewValueField)
	jsonStr := string(jsonBytes)
	return &jsonStr
}
func (f FieldChangeData) DataType() string { return f.DataTypeField }

// 审计记录类型 - v4.6.0 精确到record_id，包含完整变更信息
type AuditRecordData struct {
	AuditIDField         string            `json:"auditId"`
	RecordIDField        string            `json:"recordId"`
	OperationTypeField   string            `json:"operationType"`
	OperatedByField      OperatedByData    `json:"operatedBy"`
	ChangesSummaryField  string            `json:"changesSummary"`
	OperationReasonField *string           `json:"operationReason"`
	TimestampField       string            `json:"timestamp"`
	BeforeDataField      *string           `json:"beforeData"`
	AfterDataField       *string           `json:"afterData"`
	ModifiedFieldsField  []string          `json:"modifiedFields"`
	ChangesField         []FieldChangeData `json:"changes"`
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
func (a AuditRecordData) ModifiedFields() []string   { return a.ModifiedFieldsField }
func (a AuditRecordData) Changes() []FieldChangeData { return a.ChangesField }

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
	UnitType                 *string   `json:"unitType"`
	Status                   *string   `json:"status"`
	ParentCode               *string   `json:"parentCode"`
	Codes                    *[]string `json:"codes"`
	ExcludeCodes             *[]string `json:"excludeCodes"`
	ExcludeDescendantsOf     *string   `json:"excludeDescendantsOf"`
	IncludeDisabledAncestors bool      `json:"includeDisabledAncestors"`

	// Hierarchy Filtering
	Level      *int32 `json:"level"`
	MinLevel   *int32 `json:"minLevel"`
	MaxLevel   *int32 `json:"maxLevel"`
	RootsOnly  bool   `json:"rootsOnly"`
	LeavesOnly bool   `json:"leavesOnly"`

	// Text Search
	SearchText   *string  `json:"searchText"`
	SearchFields []string `json:"searchFields"`

	// Advanced Filtering
	HasChildren     *bool   `json:"hasChildren"`
	HasProfile      *bool   `json:"hasProfile"`
	ProfileContains *string `json:"profileContains"`

	// Audit Filtering - 修复类型匹配问题
	OperationType      *string         `json:"operationType"`
	OperatedBy         *string         `json:"operatedBy"`
	OperationDateRange *DateRangeInput `json:"operationDateRange"`
}

type PaginationInput struct {
	Page      int32  `json:"page"`
	PageSize  int32  `json:"pageSize"`
	SortBy    string `json:"sortBy"`
	SortOrder string `json:"sortOrder"`
}

// AuditHistoryConfig 控制审计历史查询的验证与回退策略
type AuditHistoryConfig struct {
	StrictValidation        bool
	AllowFallback           bool
	CircuitBreakerThreshold int32
	LegacyMode              bool
}

// PostgreSQL极速仓储 - 零抽象开销
type PostgreSQLRepository struct {
	db                     *sql.DB
	redisClient            *redis.Client
	logger                 *log.Logger
	auditConfig            AuditHistoryConfig
	validationFailureCount int32
}

func NewPostgreSQLRepository(db *sql.DB, redisClient *redis.Client, logger *log.Logger, auditConfig AuditHistoryConfig) *PostgreSQLRepository {
	return &PostgreSQLRepository{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
		auditConfig: auditConfig,
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

	includeDisabledAncestors := false

	var (
		status, searchText, unitType, parentCode string
		includeCodes, excludeCodes               []string
		asOfDateParam                            sql.NullString
		excludeDescendantsParam                  sql.NullString
	)

	if filter != nil {
		includeDisabledAncestors = filter.IncludeDisabledAncestors
		if filter.Status != nil {
			status = strings.TrimSpace(*filter.Status)
		}
		if filter.SearchText != nil {
			searchText = strings.TrimSpace(*filter.SearchText)
		}
		if filter.UnitType != nil {
			unitType = strings.TrimSpace(*filter.UnitType)
		}
		if filter.ParentCode != nil {
			parentCode = strings.TrimSpace(*filter.ParentCode)
		}
		if filter.AsOfDate != nil {
			if trimmed := strings.TrimSpace(*filter.AsOfDate); trimmed != "" {
				asOfDateParam = sql.NullString{String: trimmed, Valid: true}
			}
		}
		if filter.ExcludeDescendantsOf != nil {
			if trimmed := strings.TrimSpace(*filter.ExcludeDescendantsOf); trimmed != "" {
				excludeDescendantsParam = sql.NullString{String: trimmed, Valid: true}
			}
		}
		if filter.ExcludeCodes != nil {
			for _, code := range *filter.ExcludeCodes {
				if trimmed := strings.TrimSpace(code); trimmed != "" {
					excludeCodes = append(excludeCodes, trimmed)
				}
			}
		}
		if filter.Codes != nil {
			for _, code := range *filter.Codes {
				if trimmed := strings.TrimSpace(code); trimmed != "" {
					includeCodes = append(includeCodes, trimmed)
				}
			}
		}
	}

	cte := `
WITH parent_path AS (
    SELECT DISTINCT ON (code)
        code,
        COALESCE(code_path, '/' || code) AS code_path
    FROM organization_units
    WHERE tenant_id = $1
      AND $3::text IS NOT NULL
      AND code = $3::text
      AND status <> 'DELETED'
      AND (
        $2::text IS NULL OR (
          effective_date <= $2::date AND (end_date IS NULL OR end_date > $2::date)
        )
      )
    ORDER BY code, effective_date DESC, created_at DESC
),
latest_versions AS (
    SELECT DISTINCT ON (code)
        record_id, tenant_id, code, parent_code, name, unit_type, status,
        level, path, sort_order, description, profile, created_at, updated_at,
        effective_date, end_date, is_current, change_reason,
        deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
        COALESCE(code_path, '/' || code) AS code_path
    FROM organization_units
    WHERE tenant_id = $1
      AND status <> 'DELETED'
      AND (
        $2::text IS NULL OR (
          effective_date <= $2::date AND (end_date IS NULL OR end_date > $2::date)
        )
      )
    ORDER BY code, effective_date DESC, created_at DESC
)
`

	baseSelect := `
SELECT lv.record_id, lv.tenant_id, lv.code, lv.parent_code, lv.name, lv.unit_type, lv.status,
       lv.level, lv.path, lv.sort_order, lv.description, lv.profile, lv.created_at, lv.updated_at,
       lv.effective_date, lv.end_date, lv.is_current, lv.change_reason,
       lv.deleted_at, lv.deleted_by, lv.deletion_reason, lv.suspended_at, lv.suspended_by, lv.suspension_reason,
       COALESCE(child_stats.child_count, 0) AS children_count
FROM latest_versions lv
LEFT JOIN parent_path pp ON TRUE
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS child_count
    FROM organization_units child
    WHERE child.tenant_id = lv.tenant_id
      AND child.parent_code = lv.code
      AND child.status <> 'DELETED'
      AND (
        $2::text IS NULL OR (
          child.effective_date <= $2::date AND (child.end_date IS NULL OR child.end_date > $2::date)
        )
      )
) child_stats ON TRUE
WHERE 1=1`

	countSelect := `
SELECT COUNT(*)
FROM latest_versions lv
LEFT JOIN parent_path pp ON TRUE
WHERE 1=1`

	args := []interface{}{tenantID.String(), asOfDateParam, excludeDescendantsParam}
	argIndex := 4
	whereConditions := ""

	if status != "" {
		if includeDisabledAncestors && parentCode != "" {
			whereConditions += fmt.Sprintf(" AND (lv.status = $%d OR (lv.parent_code = $%d AND lv.status <> 'DELETED'))", argIndex, argIndex+1)
			args = append(args, status, parentCode)
			argIndex += 2
		} else {
			whereConditions += fmt.Sprintf(" AND lv.status = $%d", argIndex)
			args = append(args, status)
			argIndex++
		}
	} else {
		whereConditions += " AND lv.status <> 'DELETED'"
	}

	if unitType != "" {
		whereConditions += fmt.Sprintf(" AND lv.unit_type = $%d", argIndex)
		args = append(args, unitType)
		argIndex++
	}

	if parentCode != "" {
		whereConditions += fmt.Sprintf(" AND lv.parent_code = $%d", argIndex)
		args = append(args, parentCode)
		argIndex++
	}

	if len(includeCodes) > 0 {
		whereConditions += fmt.Sprintf(" AND lv.code = ANY($%d)", argIndex)
		args = append(args, pq.StringArray(includeCodes))
		argIndex++
	}

	if len(excludeCodes) > 0 {
		whereConditions += fmt.Sprintf(" AND NOT (lv.code = ANY($%d))", argIndex)
		args = append(args, pq.StringArray(excludeCodes))
		argIndex++
	}

	whereConditions += ` AND (
    $3::text IS NULL OR (
        lv.code <> $3::text AND (
            pp.code_path IS NULL OR lv.code_path NOT LIKE pp.code_path || '/%'
        )
    )
)`

	if searchText != "" {
		whereConditions += fmt.Sprintf(" AND (lv.name ILIKE $%d OR lv.code ILIKE $%d)", argIndex, argIndex)
		pattern := "%" + searchText + "%"
		args = append(args, pattern)
		argIndex++
	}

	countQuery := cte + countSelect + whereConditions
	countArgs := append([]interface{}{}, args...)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		r.logger.Printf("[ERROR] 查询组织总数失败: %v", err)
		return nil, err
	}

	orderClause := fmt.Sprintf(" ORDER BY COALESCE(lv.sort_order, 0) NULLS LAST, lv.code LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	dataQuery := cte + baseSelect + whereConditions + orderClause
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Printf("[ERROR] 查询组织列表失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var organizations []Organization
	for rows.Next() {
		var org Organization
		if err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
			&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
			&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField, &org.ChildrenCountField,
		); err != nil {
			r.logger.Printf("[ERROR] 扫描组织数据失败: %v", err)
			return nil, err
		}
		organizations = append(organizations, org)
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 查询 %d/%d 组织 (页面: %d/%d)，耗时: %v", len(organizations), total, page, (total+int(pageSize)-1)/int(pageSize), duration)

	totalPages := (total + int(pageSize) - 1) / int(pageSize)
	asOfDateValue := time.Now().Format("2006-01-02")
	if asOfDateParam.Valid {
		asOfDateValue = asOfDateParam.String
	}

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
			AsOfDateField:        asOfDateValue,
			CurrentCountField:    len(organizations),
			FutureCountField:     0,
			HistoricalCountField: 0,
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
               effective_date, end_date, is_current, change_reason,
               deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
        FROM organization_units 
        WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'
        LIMIT 1`

	start := time.Now()
	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code)

	var org Organization
	err := row.Scan(
		&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
		&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
		&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
		&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
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
                effective_date, end_date, is_current, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
                LEAD(effective_date) OVER (PARTITION BY tenant_id, code ORDER BY effective_date) AS next_effective
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 
              AND status <> 'DELETED'
        ), proj AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, path, sort_order, description, profile, created_at, updated_at,
                effective_date,
                COALESCE(end_date, (next_effective - INTERVAL '1 day')::date) AS computed_end_date,
                is_current, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
            FROM hist
        )
        SELECT 
            record_id, tenant_id, code, parent_code, name, unit_type, status,
            level, path, sort_order, description, profile, created_at, updated_at,
               effective_date, computed_end_date AS end_date, is_current, change_reason,
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
              AND status <> 'DELETED'
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
		err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
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

// 组织版本查询 - 按计划规范实现，返回指定code的全部版本
func (r *PostgreSQLRepository) GetOrganizationVersions(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]Organization, error) {
	start := time.Now()

	// 构建查询 - 过滤条件：tenant_id = $tenant AND code = $code
	baseQuery := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, profile, created_at, updated_at,
               effective_date, end_date, is_current, change_reason,
               deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
               hierarchy_depth
		FROM organization_units
		WHERE tenant_id = $1 AND code = $2`

	args := []interface{}{tenantID.String(), code}

	// includeDeleted=false: status != 'DELETED'
	if !includeDeleted {
		baseQuery += " AND status != 'DELETED'"
	}

	// 排序：ORDER BY effective_date ASC (按计划要求)
	finalQuery := baseQuery + " ORDER BY effective_date ASC"

	rows, err := r.db.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		r.logger.Printf("[ERROR] 组织版本查询失败: %v", err)
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
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
			&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
			&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
			&org.HierarchyDepthField,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描组织版本数据失败: %v", err)
			return nil, err
		}
		organizations = append(organizations, org)
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 组织版本查询 [%s] 返回 %d 条版本，耗时: %v", code, len(organizations), duration)

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
            FROM organization_units WHERE tenant_id = $1 AND status <> 'DELETED'
        )
		SELECT 
			s.total_count, s.active_count, s.inactive_count, s.planned_count, s.deleted_count,
			ts.total_versions, ts.unique_orgs, ts.oldest_date, ts.newest_date,
			COALESCE(json_agg(DISTINCT jsonb_build_object('unitType', t.unit_type, 'count', t.count)) FILTER (WHERE t.unit_type IS NOT NULL), '[]'),
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
		if err := json.Unmarshal([]byte(typeStatsJSON), &typeStats); err != nil {
			r.logger.Printf("解析typeStats失败: %v", err)
		}
	}
	stats.ByTypeField = typeStats

	var statusStats []StatusCount
	if statusStatsJSON != "" {
		if err := json.Unmarshal([]byte(statusStatsJSON), &statusStats); err != nil {
			r.logger.Printf("解析statusStats失败: %v", err)
		}
	}
	stats.ByStatusField = statusStats

	var levelStats []LevelCount
	if levelStatsJSON != "" {
		if err := json.Unmarshal([]byte(levelStatsJSON), &levelStats); err != nil {
			r.logger.Printf("解析levelStats失败: %v", err)
		}
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
                code,
                name,
                level,
                parent_code,
                1 AS hierarchy_depth
            FROM organization_units
            WHERE tenant_id = $1
              AND code = $2
              AND is_current = true
              AND status <> 'DELETED'

            UNION ALL

            -- 递归获取父级信息
            SELECT
                o.code,
                o.name,
                o.level,
                o.parent_code,
                h.hierarchy_depth + 1
            FROM organization_units o
            INNER JOIN hierarchy_info h ON o.code = h.parent_code
            WHERE o.tenant_id = $1
              AND o.is_current = true
              AND o.status <> 'DELETED'
        ),
        aggregated_paths AS (
            SELECT
                '/' || string_agg(code, '/' ORDER BY hierarchy_depth DESC) AS full_code_path,
                '/' || string_agg(name, '/' ORDER BY hierarchy_depth DESC) AS full_name_path,
                COALESCE(
                    array_agg(code ORDER BY hierarchy_depth DESC) FILTER (WHERE hierarchy_depth > 1),
                    ARRAY[]::text[]
                ) AS parent_chain
            FROM hierarchy_info
        ),
        target_info AS (
            SELECT *
            FROM hierarchy_info
            WHERE code = $2
            LIMIT 1
        ),
        children_count AS (
            SELECT COUNT(*) AS count
            FROM organization_units
            WHERE tenant_id = $1
              AND parent_code = $2
              AND is_current = true
              AND status <> 'DELETED'
        )
        SELECT
            t.code,
            t.name,
            t.level,
            t.hierarchy_depth,
            ap.full_code_path,
            ap.full_name_path,
            ap.parent_chain,
            c.count AS children_count,
            (t.parent_code IS NULL) AS is_root,
            (c.count = 0) AS is_leaf
        FROM target_info t
        CROSS JOIN aggregated_paths ap
        CROSS JOIN children_count c
        LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code)

	var hierarchy OrganizationHierarchyData
	var parentChain []string

	err := row.Scan(
		&hierarchy.CodeField,
		&hierarchy.NameField,
		&hierarchy.LevelField,
		&hierarchy.HierarchyDepthField,
		&hierarchy.CodePathField,
		&hierarchy.NamePathField,
		pq.Array(&parentChain),
		&hierarchy.ChildrenCountField,
		&hierarchy.IsRootField,
		&hierarchy.IsLeafField,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Printf("[ERROR] 层级结构查询失败: %v", err)
		return nil, err
	}

	hierarchy.ParentChainField = parentChain

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
            WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'
            
            UNION ALL
            
            -- 递归查询子节点
            SELECT 
                o.code, o.name, o.level,
                o.hierarchy_depth, o.code_path, o.name_path, o.parent_code,
                s.depth_from_root + 1
            FROM organization_units o
            INNER JOIN subtree s ON o.parent_code = s.code
            WHERE o.tenant_id = $1 AND o.is_current = true AND o.status <> 'DELETED'
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

// 审计历史查询 - v4.6.0 基于record_id精确查询 + 租户隔离
func (r *PostgreSQLRepository) GetAuditHistory(ctx context.Context, tenantId uuid.UUID, recordId string, startDate, endDate, operation, userId *string, limit int) ([]AuditRecordData, error) {
	start := time.Now()

	// 构建查询条件 - 基于record_id查询，包含完整变更信息，强制租户隔离
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
			request_data::text as before_data,
			response_data::text as after_data,
			'[]'::text as modified_fields,
			CASE WHEN changes IS NOT NULL
				THEN changes::text
				ELSE '[]'
			END as detailed_changes
		FROM audit_logs
		WHERE tenant_id = $1::uuid AND resource_id = $2 AND resource_type = 'ORGANIZATION'`

	args := []interface{}{tenantId, recordId}
	argIndex := 3

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
	if r.auditConfig.LegacyMode {
		auditRecords, err = r.processAuditRowsLegacy(rows)
	} else {
		auditRecords, err = r.processAuditRowsStrict(rows)
	}
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] record_id审计查询完成，返回 %d 条记录，耗时: %v", len(auditRecords), duration)

	return auditRecords, nil
}

func (r *PostgreSQLRepository) processAuditRowsLegacy(rows *sql.Rows) ([]AuditRecordData, error) {
	var auditRecords []AuditRecordData
	for rows.Next() {
		var record AuditRecordData
		var operatedById, operatedByName string
		var beforeData, afterData, modifiedFieldsJSON, detailedChangesJSON sql.NullString

		err := rows.Scan(
			&record.AuditIDField, &record.RecordIDField, &record.OperationTypeField,
			&operatedById, &operatedByName,
			&record.ChangesSummaryField, &record.OperationReasonField, &record.TimestampField,
			&beforeData, &afterData, &modifiedFieldsJSON, &detailedChangesJSON,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描审计记录失败: %v", err)
			return nil, err
		}

		if beforeData.Valid {
			record.BeforeDataField = &beforeData.String
		}
		if afterData.Valid {
			record.AfterDataField = &afterData.String
		}

		if modifiedFieldsJSON.Valid && modifiedFieldsJSON.String != "[]" {
			var modifiedFields []string
			if err := json.Unmarshal([]byte(modifiedFieldsJSON.String), &modifiedFields); err == nil {
				record.ModifiedFieldsField = modifiedFields
			}
		}

		if detailedChangesJSON.Valid && detailedChangesJSON.String != "[]" {
			var changesArray []map[string]interface{}
			if err := json.Unmarshal([]byte(detailedChangesJSON.String), &changesArray); err == nil {
				for _, changeMap := range changesArray {
					fieldChange := FieldChangeData{
						FieldField:    fmt.Sprintf("%v", changeMap["field"]),
						OldValueField: changeMap["oldValue"],
						NewValueField: changeMap["newValue"],
						DataTypeField: fmt.Sprintf("%v", changeMap["dataType"]),
					}
					record.ChangesField = append(record.ChangesField, fieldChange)
				}
			}
		}

		record.OperatedByField = OperatedByData{
			IDField:   operatedById,
			NameField: operatedByName,
		}

		auditRecords = append(auditRecords, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return auditRecords, nil
}

func (r *PostgreSQLRepository) processAuditRowsStrict(rows *sql.Rows) ([]AuditRecordData, error) {
	var auditRecords []AuditRecordData
	for rows.Next() {
		var record AuditRecordData
		var operatedById, operatedByName string
		var beforeData, afterData, modifiedFieldsJSON, detailedChangesJSON sql.NullString

		record.ModifiedFieldsField = make([]string, 0)
		record.ChangesField = make([]FieldChangeData, 0)

		err := rows.Scan(
			&record.AuditIDField, &record.RecordIDField, &record.OperationTypeField,
			&operatedById, &operatedByName,
			&record.ChangesSummaryField, &record.OperationReasonField, &record.TimestampField,
			&beforeData, &afterData, &modifiedFieldsJSON, &detailedChangesJSON,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描审计记录失败: %v", err)
			return nil, err
		}

		if beforeData.Valid {
			record.BeforeDataField = &beforeData.String
		}
		if afterData.Valid {
			record.AfterDataField = &afterData.String
		}

		rawModified := ""
		if modifiedFieldsJSON.Valid {
			rawModified = modifiedFieldsJSON.String
		}
		sanitizedModified, modifiedIssues, modErr := sanitizeModifiedFields(rawModified)
		if modErr == nil {
			record.ModifiedFieldsField = sanitizedModified
		}

		rawChanges := ""
		if detailedChangesJSON.Valid {
			rawChanges = detailedChangesJSON.String
		}
		sanitizedChanges, changeIssues, changeErr := sanitizeChanges(rawChanges)
		if changeErr == nil {
			record.ChangesField = sanitizedChanges
		}

		issues := make([]string, 0, len(modifiedIssues)+len(changeIssues))
		issues = append(issues, modifiedIssues...)
		issues = append(issues, changeIssues...)

		hasHardError := false
		if modErr != nil {
			hasHardError = true
			issues = append(issues, fmt.Sprintf("modified_fields JSON 无效: %v", modErr))
		}
		if changeErr != nil {
			hasHardError = true
			issues = append(issues, fmt.Sprintf("changes JSON 无效: %v", changeErr))
		}

		if len(issues) > 0 {
			r.logger.Printf("[WARN] 审计记录数据异常 audit_id=%s: %s", record.AuditIDField, strings.Join(issues, "; "))
			if r.auditConfig.StrictValidation {
				if hasHardError && !r.auditConfig.AllowFallback {
					return nil, fmt.Errorf("AUDIT_HISTORY_VALIDATION_FAILED")
				}
				if r.registerValidationFailure() {
					return nil, fmt.Errorf("AUDIT_HISTORY_CIRCUIT_OPEN")
				}
			}
		} else if r.auditConfig.StrictValidation {
			r.registerValidationSuccess()
		}

		record.OperatedByField = OperatedByData{
			IDField:   operatedById,
			NameField: operatedByName,
		}

		auditRecords = append(auditRecords, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return auditRecords, nil
}

func sanitizeModifiedFields(raw string) ([]string, []string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return make([]string, 0), nil, nil
	}
	if trimmed == "null" {
		return make([]string, 0), []string{"modified_fields 为 null，已替换为空数组"}, nil
	}

	var rawArray []interface{}
	if err := json.Unmarshal([]byte(trimmed), &rawArray); err != nil {
		return make([]string, 0), nil, err
	}

	sanitized := make([]string, 0, len(rawArray))
	issues := make([]string, 0)
	for idx, item := range rawArray {
		if item == nil {
			issues = append(issues, fmt.Sprintf("modified_fields[%d] 为 null，已忽略", idx))
			continue
		}
		switch v := item.(type) {
		case string:
			sanitized = append(sanitized, v)
		default:
			sanitized = append(sanitized, fmt.Sprintf("%v", v))
			issues = append(issues, fmt.Sprintf("modified_fields[%d] 非字符串，已转换", idx))
		}
	}

	return sanitized, issues, nil
}

func sanitizeChanges(raw string) ([]FieldChangeData, []string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return make([]FieldChangeData, 0), nil, nil
	}
	if trimmed == "null" {
		return make([]FieldChangeData, 0), []string{"changes 为 null，已替换为空数组"}, nil
	}

	var rawArray []map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &rawArray); err != nil {
		return make([]FieldChangeData, 0), nil, err
	}

	sanitized := make([]FieldChangeData, 0, len(rawArray))
	issues := make([]string, 0)
	for idx, entry := range rawArray {
		if entry == nil {
			issues = append(issues, fmt.Sprintf("changes[%d] 为空对象，已跳过", idx))
			continue
		}

		fieldVal, ok := entry["field"]
		if !ok {
			issues = append(issues, fmt.Sprintf("changes[%d] 缺少 field，已跳过", idx))
			continue
		}
		field := strings.TrimSpace(fmt.Sprintf("%v", fieldVal))
		if field == "" {
			issues = append(issues, fmt.Sprintf("changes[%d] field 为空，已跳过", idx))
			continue
		}

		dataType := "unknown"
		if dtVal, ok := entry["dataType"]; ok {
			if dtStr, ok := dtVal.(string); ok && strings.TrimSpace(dtStr) != "" {
				dataType = dtStr
			} else {
				issues = append(issues, fmt.Sprintf("changes[%d] dataType 非字符串，使用 unknown", idx))
			}
		} else {
			issues = append(issues, fmt.Sprintf("changes[%d] 缺少 dataType，使用 unknown", idx))
		}

		fieldChange := FieldChangeData{
			FieldField:    field,
			DataTypeField: dataType,
			OldValueField: normalizeChangeValue(entry["oldValue"]),
			NewValueField: normalizeChangeValue(entry["newValue"]),
		}
		sanitized = append(sanitized, fieldChange)
	}

	return sanitized, issues, nil
}

func normalizeChangeValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		return v.String()
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(bytes)
	}
}

func (r *PostgreSQLRepository) registerValidationFailure() bool {
	count := atomic.AddInt32(&r.validationFailureCount, 1)
	if r.auditConfig.CircuitBreakerThreshold > 0 && count >= r.auditConfig.CircuitBreakerThreshold {
		r.logger.Printf("[ALERT] 审计历史验证失败次数达到阈值 (%d/%d)，触发熔断", count, r.auditConfig.CircuitBreakerThreshold)
		return true
	}
	return false
}

func (r *PostgreSQLRepository) registerValidationSuccess() {
	if atomic.LoadInt32(&r.validationFailureCount) != 0 {
		atomic.StoreInt32(&r.validationFailureCount, 0)
	}
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
        WHERE id = $1::uuid AND resource_type = 'ORGANIZATION' AND tenant_id = $2::uuid
        LIMIT 1`

	tenantID := auth.GetTenantID(ctx)
	if tenantID == "" {
		r.logger.Printf("[AUTH] 缺少租户ID，拒绝单条审计记录查询")
		return nil, fmt.Errorf("TENANT_REQUIRED")
	}

	row := r.db.QueryRowContext(ctx, query, auditId, tenantID)

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

// 组织版本查询 - 按计划实现，支持includeDeleted参数
func (r *Resolver) OrganizationVersions(ctx context.Context, args struct {
	Code           string
	IncludeDeleted *bool
}) ([]Organization, error) {
	if err := r.authMW.CheckQueryPermission(ctx, "organizationVersions"); err != nil {
		r.logger.Printf("[AUTH] 权限拒绝: organizationVersions: %v", err)
		return nil, fmt.Errorf("INSUFFICIENT_PERMISSIONS")
	}

	includeDeleted := false
	if args.IncludeDeleted != nil {
		includeDeleted = *args.IncludeDeleted
	}

	r.logger.Printf("[GraphQL] 版本查询 - code: %s, includeDeleted: %v", args.Code, includeDeleted)
	return r.repo.GetOrganizationVersions(ctx, DefaultTenantID, args.Code, includeDeleted)
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
	TenantId              string
	IncludeIntegrityCheck bool
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

	// 从上下文获取租户ID，强制租户隔离
	tenantStr := auth.GetTenantID(ctx)
	if tenantStr == "" {
		r.logger.Printf("[AUTH] 缺少租户ID，拒绝审计历史查询")
		return nil, fmt.Errorf("TENANT_REQUIRED")
	}
	tenantUUID, err := uuid.Parse(tenantStr)
	if err != nil {
		r.logger.Printf("[AUTH] 无效租户ID: %s", tenantStr)
		return nil, fmt.Errorf("INVALID_TENANT")
	}

	return r.repo.GetAuditHistory(ctx, tenantUUID, args.RecordId, args.StartDate, args.EndDate, args.Operation, args.UserId, int(limit))
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
	auditConfig := loadAuditHistoryConfig()
	repo := NewPostgreSQLRepository(db, redisClient, logger, auditConfig)
	logger.Printf("⚙️ 审计历史配置: strictValidation=%v, allowFallback=%v, circuitThreshold=%d, legacyMode=%v",
		auditConfig.StrictValidation, auditConfig.AllowFallback, auditConfig.CircuitBreakerThreshold, auditConfig.LegacyMode)

	// 初始化JWT中间件 - 使用统一配置
	jwtConfig := config.GetJWTConfig()
	devMode := getEnv("DEV_MODE", "true") == "true"

	var pubPEM []byte
	if jwtConfig.HasPublicKey() {
		if b, err := os.ReadFile(jwtConfig.PublicKeyPath); err == nil {
			pubPEM = b
		} else {
			logger.Fatalf("[FATAL] 无法读取查询服务公钥 (%s): %v", jwtConfig.PublicKeyPath, err)
		}
	}
	if jwtConfig.JWKSUrl == "" && pubPEM == nil {
		logger.Fatalf("[FATAL] 查询服务启用RS256必须配置 JWT_JWKS_URL 或 JWT_PUBLIC_KEY_PATH")
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

	logger.Printf("🔐 JWT认证初始化完成 (开发模式: %v, Alg=%s, Issuer=%s, Audience=%s)", devMode, jwtConfig.Algorithm, jwtConfig.Issuer, jwtConfig.Audience)

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
		if _, err := w.Write([]byte(graphiqlHTML)); err != nil {
			http.Error(w, "failed to write GraphiQL page", http.StatusInternalServerError)
		}
	})

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "healthy",
			"service":     "postgresql-graphql",
			"timestamp":   time.Now(),
			"database":    "postgresql",
			"performance": "optimized",
		}); err != nil {
			http.Error(w, "failed to encode health response", http.StatusInternalServerError)
		}
	})

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

func loadAuditHistoryConfig() AuditHistoryConfig {
	threshold := getEnvAsInt("AUDIT_HISTORY_CIRCUIT_BREAKER_THRESHOLD", 25)
	if threshold < 0 {
		threshold = 0
	}
	return AuditHistoryConfig{
		StrictValidation:        getEnvAsBool("AUDIT_HISTORY_STRICT_VALIDATION", true),
		AllowFallback:           getEnvAsBool("AUDIT_HISTORY_ALLOW_FALLBACK", true),
		CircuitBreakerThreshold: int32(threshold),
		LegacyMode:              getEnvAsBool("AUDIT_HISTORY_LEGACY_MODE", false),
	}
}

func getEnvAsBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
}

func getEnvAsInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
