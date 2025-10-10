package model

import (
	"fmt"
	"math"
	"time"
)

// Organization 表示 GraphQL 暴露的组织实体，字段与数据库列保持一一对应
type Organization struct {
	RecordIDField         string     `json:"recordId" db:"record_id"`
	TenantIDField         string     `json:"tenantId" db:"tenant_id"`
	CodeField             string     `json:"code" db:"code"`
	ParentCodeField       *string    `json:"parentCode" db:"parent_code"`
	NameField             string     `json:"name" db:"name"`
	UnitTypeField         string     `json:"unitType" db:"unit_type"`
	StatusField           string     `json:"status" db:"status"`
	LevelField            int        `json:"level" db:"level"`
	CodePathField         string     `json:"codePath" db:"code_path"`
	NamePathField         string     `json:"namePath" db:"name_path"`
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

func (o Organization) RecordId() string { return o.RecordIDField }
func (o Organization) TenantId() string { return o.TenantIDField }
func (o Organization) Code() string     { return o.CodeField }

func (o Organization) ParentCode() string {
	if o.ParentCodeField == nil {
		return "0"
	}
	return *o.ParentCodeField
}

func (o Organization) Name() string     { return o.NameField }
func (o Organization) UnitType() string { return o.UnitTypeField }
func (o Organization) Status() string   { return o.StatusField }
func (o Organization) Level() int32     { return clampToInt32(o.LevelField) }
func (o Organization) Path() *string {
	if o.CodePathField == "" {
		return nil
	}
	path := o.CodePathField
	return &path
}
func (o Organization) CodePath() string {
	if o.CodePathField != "" {
		return o.CodePathField
	}
	return ""
}
func (o Organization) NamePath() string {
	if o.NamePathField != "" {
		return o.NamePathField
	}
	return ""
}
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

func cnTodayDate() time.Time {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Now().UTC().Truncate(24 * time.Hour)
	}
	nowCN := time.Now().In(loc)
	return time.Date(nowCN.Year(), nowCN.Month(), nowCN.Day(), 0, 0, 0, 0, loc)
}

func (o Organization) IsTemporal() bool {
	return o.EndDateField != nil
}

func (o Organization) IsFuture() bool {
	todayCN := cnTodayDate()
	eff := time.Date(o.EffectiveDateField.Year(), o.EffectiveDateField.Month(), o.EffectiveDateField.Day(), 0, 0, 0, 0, todayCN.Location())
	return eff.After(todayCN)
}

func (o Organization) ChangeReason() *string { return o.ChangeReasonField }

func (o Organization) HierarchyDepth() int32 { return clampToInt32(o.HierarchyDepthField) }
func (o Organization) ChildrenCount() int32  { return clampToInt32(o.ChildrenCountField) }

func (o Organization) DeletedAt() *string {
	if o.DeletedAtField == nil {
		return nil
	}
	ts := o.DeletedAtField.Format(time.RFC3339)
	return &ts
}

func (o Organization) DeletedBy() *string { return o.DeletedByField }

func (o Organization) DeletionReason() *string { return o.DeletionReasonField }

func (o Organization) SuspendedAt() *string {
	if o.SuspendedAtField == nil {
		return nil
	}
	ts := o.SuspendedAtField.Format(time.RFC3339)
	return &ts
}

func (o Organization) SuspendedBy() *string { return o.SuspendedByField }

func (o Organization) SuspensionReason() *string { return o.SuspensionReasonField }

// OrganizationStats 统计信息
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

func (s OrganizationStats) TotalCount() int32    { return int32(s.TotalCountField) }
func (s OrganizationStats) ActiveCount() int32   { return int32(s.ActiveCountField) }
func (s OrganizationStats) InactiveCount() int32 { return int32(s.InactiveCountField) }
func (s OrganizationStats) PlannedCount() int32  { return int32(s.PlannedCountField) }
func (s OrganizationStats) DeletedCount() int32  { return int32(s.DeletedCountField) }
func (s OrganizationStats) ByType() []TypeCount  { return s.ByTypeField }
func (s OrganizationStats) ByStatus() []StatusCount {
	return s.ByStatusField
}
func (s OrganizationStats) ByLevel() []LevelCount { return s.ByLevelField }
func (s OrganizationStats) TemporalStats() TemporalStats {
	return s.TemporalStatsField
}

// TemporalStats 时态统计信息
type TemporalStats struct {
	TotalVersionsField         int     `json:"totalVersions"`
	AverageVersionsPerOrgField float64 `json:"averageVersionsPerOrg"`
	OldestEffectiveDateField   string  `json:"oldestEffectiveDate"`
	NewestEffectiveDateField   string  `json:"newestEffectiveDate"`
}

func (t TemporalStats) TotalVersions() int32 { return int32(t.TotalVersionsField) }
func (t TemporalStats) AverageVersionsPerOrg() float64 {
	return t.AverageVersionsPerOrgField
}
func (t TemporalStats) OldestEffectiveDate() string {
	return t.OldestEffectiveDateField
}
func (t TemporalStats) NewestEffectiveDate() string {
	return t.NewestEffectiveDateField
}

// TypeCount 类型统计
type TypeCount struct {
	UnitTypeField string `json:"unitType"`
	CountField    int    `json:"count"`
}

func (t TypeCount) UnitType() string { return t.UnitTypeField }
func (t TypeCount) Count() int32     { return int32(t.CountField) }

// LevelCount 层级统计
type LevelCount struct {
	LevelField int `json:"level"`
	CountField int `json:"count"`
}

func (l LevelCount) Level() int32 { return int32(l.LevelField) }
func (l LevelCount) Count() int32 { return int32(l.CountField) }

// StatusCount 状态统计
type StatusCount struct {
	StatusField string `json:"status"`
	CountField  int    `json:"count"`
}

func (s StatusCount) Status() string { return s.StatusField }
func (s StatusCount) Count() int32   { return int32(s.CountField) }

// OrganizationConnection GraphQL 分页封装
type OrganizationConnection struct {
	DataField       []Organization `json:"data"`
	PaginationField PaginationInfo `json:"pagination"`
	TemporalField   TemporalInfo   `json:"temporal"`
}

func (c OrganizationConnection) Data() []Organization { return c.DataField }
func (c OrganizationConnection) Pagination() PaginationInfo {
	return c.PaginationField
}
func (c OrganizationConnection) Temporal() TemporalInfo { return c.TemporalField }

// PaginationInfo 分页元信息
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

// TemporalInfo 时态信息
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

// OrganizationHierarchyData 组织层级数据
type OrganizationHierarchyData struct {
	CodeField           string                      `json:"code"`
	NameField           string                      `json:"name"`
	LevelField          int                         `json:"level"`
	HierarchyDepthField int                         `json:"hierarchyDepth"`
	CodePathField       *string                     `json:"codePath"`
	NamePathField       *string                     `json:"namePath"`
	ParentChainField    []string                    `json:"parentChain"`
	ChildrenCountField  int                         `json:"childrenCount"`
	IsRootField         bool                        `json:"isRoot"`
	IsLeafField         bool                        `json:"isLeaf"`
	ChildrenField       []OrganizationHierarchyData `json:"children"`
}

func (o OrganizationHierarchyData) Code() string          { return o.CodeField }
func (o OrganizationHierarchyData) Name() string          { return o.NameField }
func (o OrganizationHierarchyData) Level() int32          { return int32(o.LevelField) }
func (o OrganizationHierarchyData) HierarchyDepth() int32 { return int32(o.HierarchyDepthField) }
func (o OrganizationHierarchyData) CodePath() string {
	if o.CodePathField == nil {
		return ""
	}
	return *o.CodePathField
}
func (o OrganizationHierarchyData) NamePath() string {
	if o.NamePathField == nil {
		return ""
	}
	return *o.NamePathField
}
func (o OrganizationHierarchyData) ParentChain() []string { return o.ParentChainField }
func (o OrganizationHierarchyData) ChildrenCount() int32  { return int32(o.ChildrenCountField) }
func (o OrganizationHierarchyData) IsRoot() bool          { return o.IsRootField }
func (o OrganizationHierarchyData) IsLeaf() bool          { return o.IsLeafField }
func (o OrganizationHierarchyData) Children() []OrganizationHierarchyData {
	return o.ChildrenField
}

// OrganizationSubtreeData 子树数据模型
type OrganizationSubtreeData struct {
	CodeField           string                    `json:"code"`
	NameField           string                    `json:"name"`
	LevelField          int                       `json:"level"`
	HierarchyDepthField int                       `json:"hierarchyDepth"`
	CodePathField       *string                   `json:"codePath"`
	NamePathField       *string                   `json:"namePath"`
	ParentChainField    []string                  `json:"parentChain"`
	ChildrenCountField  int                       `json:"childrenCount"`
	IsRootField         bool                      `json:"isRoot"`
	IsLeafField         bool                      `json:"isLeaf"`
	ChildrenField       []OrganizationSubtreeData `json:"children"`
}

func (o OrganizationSubtreeData) Code() string { return o.CodeField }
func (o OrganizationSubtreeData) Name() string { return o.NameField }
func (o OrganizationSubtreeData) Level() int32 { return int32(o.LevelField) }
func (o OrganizationSubtreeData) HierarchyDepth() int32 {
	return int32(o.HierarchyDepthField)
}
func (o OrganizationSubtreeData) CodePath() string {
	if o.CodePathField == nil {
		return ""
	}
	return *o.CodePathField
}
func (o OrganizationSubtreeData) NamePath() string {
	if o.NamePathField == nil {
		return ""
	}
	return *o.NamePathField
}
func (o OrganizationSubtreeData) ParentChain() []string {
	return o.ParentChainField
}
func (o OrganizationSubtreeData) ChildrenCount() int32 { return int32(o.ChildrenCountField) }
func (o OrganizationSubtreeData) IsRoot() bool         { return o.IsRootField }
func (o OrganizationSubtreeData) IsLeaf() bool         { return o.IsLeafField }
func (o OrganizationSubtreeData) Children() []OrganizationSubtreeData {
	return o.ChildrenField
}

// HierarchyStatistics 层级统计模型
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

func (h HierarchyStatistics) TenantId() string          { return h.TenantIdField }
func (h HierarchyStatistics) TotalOrganizations() int32 { return int32(h.TotalOrganizationsField) }
func (h HierarchyStatistics) MaxDepth() int32           { return int32(h.MaxDepthField) }
func (h HierarchyStatistics) AvgDepth() float64         { return h.AvgDepthField }
func (h HierarchyStatistics) DepthDistribution() []DepthDistribution {
	return h.DepthDistributionField
}
func (h HierarchyStatistics) RootOrganizations() int32 { return int32(h.RootOrganizationsField) }
func (h HierarchyStatistics) LeafOrganizations() int32 { return int32(h.LeafOrganizationsField) }
func (h HierarchyStatistics) IntegrityIssues() []IntegrityIssue {
	return h.IntegrityIssuesField
}
func (h HierarchyStatistics) LastAnalyzed() string { return h.LastAnalyzedField }

// DepthDistribution 层级分布
type DepthDistribution struct {
	DepthField int `json:"depth"`
	CountField int `json:"count"`
}

func (d DepthDistribution) Depth() int32 { return int32(d.DepthField) }
func (d DepthDistribution) Count() int32 { return int32(d.CountField) }

// IntegrityIssue 数据一致性问题
type IntegrityIssue struct {
	TypeField          string   `json:"type"`
	CountField         int      `json:"count"`
	AffectedCodesField []string `json:"affectedCodes"`
}

func (i IntegrityIssue) Type() string            { return i.TypeField }
func (i IntegrityIssue) Count() int32            { return int32(i.CountField) }
func (i IntegrityIssue) AffectedCodes() []string { return i.AffectedCodesField }

// FieldChangeData 审计字段变更详情
type FieldChangeData struct {
	FieldField    string      `json:"field"`
	DataTypeField string      `json:"dataType"`
	OldValueField interface{} `json:"oldValue"`
	NewValueField interface{} `json:"newValue"`
}

func (f FieldChangeData) Field() string    { return f.FieldField }
func (f FieldChangeData) DataType() string { return f.DataTypeField }
func (f FieldChangeData) OldValue() *string {
	if f.OldValueField == nil {
		return nil
	}
	str := fmt.Sprint(f.OldValueField)
	return &str
}
func (f FieldChangeData) NewValue() *string {
	if f.NewValueField == nil {
		return nil
	}
	str := fmt.Sprint(f.NewValueField)
	return &str
}

// AuditRecordData 审计记录-包含变更详情
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
func (a AuditRecordData) BeforeData() *string        { return a.BeforeDataField }
func (a AuditRecordData) AfterData() *string         { return a.AfterDataField }
func (a AuditRecordData) ModifiedFields() []string   { return a.ModifiedFieldsField }
func (a AuditRecordData) Changes() []FieldChangeData { return a.ChangesField }

// OperatedByData 审计操作人信息
type OperatedByData struct {
	IDField   string `json:"id"`
	NameField string `json:"name"`
}

func (o OperatedByData) Id() string   { return o.IDField }
func (o OperatedByData) Name() string { return o.NameField }

// DateRangeInput GraphQL 输入
type DateRangeInput struct {
	From *string `json:"from"`
	To   *string `json:"to"`
}

// OrganizationFilter 查询过滤条件
type OrganizationFilter struct {
	AsOfDate                 *string         `json:"asOfDate"`
	IncludeFuture            bool            `json:"includeFuture"`
	OnlyFuture               bool            `json:"onlyFuture"`
	UnitType                 *string         `json:"unitType"`
	Status                   *string         `json:"status"`
	ParentCode               *string         `json:"parentCode"`
	Codes                    *[]string       `json:"codes"`
	ExcludeCodes             *[]string       `json:"excludeCodes"`
	ExcludeDescendantsOf     *string         `json:"excludeDescendantsOf"`
	IncludeDisabledAncestors bool            `json:"includeDisabledAncestors"`
	Level                    *int32          `json:"level"`
	MinLevel                 *int32          `json:"minLevel"`
	MaxLevel                 *int32          `json:"maxLevel"`
	RootsOnly                bool            `json:"rootsOnly"`
	LeavesOnly               bool            `json:"leavesOnly"`
	SearchText               *string         `json:"searchText"`
	SearchFields             []string        `json:"searchFields"`
	HasChildren              *bool           `json:"hasChildren"`
	HasProfile               *bool           `json:"hasProfile"`
	ProfileContains          *string         `json:"profileContains"`
	OperationType            *string         `json:"operationType"`
	OperatedBy               *string         `json:"operatedBy"`
	OperationDateRange       *DateRangeInput `json:"operationDateRange"`
}

// PaginationInput 查询分页参数
type PaginationInput struct {
	Page      int32  `json:"page"`
	PageSize  int32  `json:"pageSize"`
	SortBy    string `json:"sortBy"`
	SortOrder string `json:"sortOrder"`
}
