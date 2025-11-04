package dto

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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

func (f *OrganizationFilter) UnmarshalGraphQL(input interface{}) error {
	raw, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("OrganizationFilter: 期望对象类型，实际得到 %T", input)
	}

	if value, exists := raw["asOfDate"]; exists {
		strPtr, err := asOptionalString(value)
		if err != nil {
			return fmt.Errorf("OrganizationFilter.asOfDate: %w", err)
		}
		f.AsOfDate = strPtr
	}
	if value, exists := raw["includeFuture"]; exists {
		boolVal, err := asBool(value)
		if err != nil {
			return fmt.Errorf("OrganizationFilter.includeFuture: %w", err)
		}
		f.IncludeFuture = boolVal
	}
	if value, exists := raw["onlyFuture"]; exists {
		boolVal, err := asBool(value)
		if err != nil {
			return fmt.Errorf("OrganizationFilter.onlyFuture: %w", err)
		}
		f.OnlyFuture = boolVal
	}
	if value, exists := raw["unitType"]; exists {
		strPtr, err := asOptionalString(value)
		if err != nil {
			return fmt.Errorf("OrganizationFilter.unitType: %w", err)
		}
		f.UnitType = strPtr
	}
	if value, exists := raw["status"]; exists {
		strPtr, err := asOptionalString(value)
		if err != nil {
			return fmt.Errorf("OrganizationFilter.status: %w", err)
		}
		f.Status = strPtr
	}
	if value, exists := raw["parentCode"]; exists {
		strPtr, err := asOptionalString(value)
		if err != nil {
			return fmt.Errorf("OrganizationFilter.parentCode: %w", err)
		}
		f.ParentCode = strPtr
	}
	if value, exists := raw["hasChildren"]; exists {
		boolPtr, err := asOptionalBool(value)
		if err != nil {
			return fmt.Errorf("OrganizationFilter.hasChildren: %w", err)
		}
		f.HasChildren = boolPtr
	}
	if value, exists := raw["hasProfile"]; exists {
		boolPtr, err := asOptionalBool(value)
		if err != nil {
			return fmt.Errorf("OrganizationFilter.hasProfile: %w", err)
		}
		f.HasProfile = boolPtr
	}

	return nil
}

// PaginationInput 查询分页参数
type PaginationInput struct {
	Page      int32  `json:"page"`
	PageSize  int32  `json:"pageSize"`
	SortBy    string `json:"sortBy"`
	SortOrder string `json:"sortOrder"`
}

// Position 数据实体
type Position struct {
	CodeField               string               `json:"code" db:"code"`
	RecordIDField           string               `json:"recordId" db:"record_id"`
	TenantIDField           string               `json:"tenantId" db:"tenant_id"`
	TitleField              string               `json:"title" db:"title"`
	JobProfileCodeField     *string              `json:"jobProfileCode" db:"job_profile_code"`
	JobProfileNameField     *string              `json:"jobProfileName" db:"job_profile_name"`
	JobFamilyGroupCodeField string               `json:"jobFamilyGroupCode" db:"job_family_group_code"`
	JobFamilyCodeField      string               `json:"jobFamilyCode" db:"job_family_code"`
	JobRoleCodeField        string               `json:"jobRoleCode" db:"job_role_code"`
	JobLevelCodeField       string               `json:"jobLevelCode" db:"job_level_code"`
	OrganizationCodeField   string               `json:"organizationCode" db:"organization_code"`
	PositionTypeField       string               `json:"positionType" db:"position_type"`
	EmploymentTypeField     string               `json:"employmentType" db:"employment_type"`
	GradeLevelField         *string              `json:"gradeLevel" db:"grade_level"`
	HeadcountCapacityField  float64              `json:"headcountCapacity" db:"headcount_capacity"`
	HeadcountInUseField     float64              `json:"headcountInUse" db:"headcount_in_use"`
	ReportsToPositionField  *string              `json:"reportsToPositionCode" db:"reports_to_position_code"`
	StatusField             string               `json:"status" db:"status"`
	EffectiveDateField      time.Time            `json:"effectiveDate" db:"effective_date"`
	EndDateField            *time.Time           `json:"endDate" db:"end_date"`
	IsCurrentField          bool                 `json:"isCurrent" db:"is_current"`
	CreatedAtField          time.Time            `json:"createdAt" db:"created_at"`
	UpdatedAtField          time.Time            `json:"updatedAt" db:"updated_at"`
	JobFamilyGroupNameField *string              `json:"jobFamilyGroupName" db:"job_family_group_name"`
	JobFamilyNameField      *string              `json:"jobFamilyName" db:"job_family_name"`
	JobRoleNameField        *string              `json:"jobRoleName" db:"job_role_name"`
	JobLevelNameField       *string              `json:"jobLevelName" db:"job_level_name"`
	OrganizationNameField   *string              `json:"organizationName" db:"organization_name"`
	CurrentAssignmentField  *PositionAssignment  `json:"currentAssignment"`
	AssignmentHistoryField  []PositionAssignment `json:"assignmentHistory"`
}

func (p Position) Code() PositionCode      { return PositionCode(p.CodeField) }
func (p Position) RecordId() UUID          { return UUID(p.RecordIDField) }
func (p Position) TenantId() UUID          { return UUID(p.TenantIDField) }
func (p Position) Title() string           { return p.TitleField }
func (p Position) JobProfileCode() *string { return p.JobProfileCodeField }
func (p Position) JobProfileName() *string { return p.JobProfileNameField }
func (p Position) JobFamilyGroupCode() JobFamilyGroupCode {
	return JobFamilyGroupCode(p.JobFamilyGroupCodeField)
}
func (p Position) JobFamilyCode() JobFamilyCode { return JobFamilyCode(p.JobFamilyCodeField) }
func (p Position) JobRoleCode() JobRoleCode     { return JobRoleCode(p.JobRoleCodeField) }
func (p Position) JobLevelCode() JobLevelCode   { return JobLevelCode(p.JobLevelCodeField) }
func (p Position) OrganizationCode() string     { return p.OrganizationCodeField }
func (p Position) PositionType() string         { return p.PositionTypeField }
func (p Position) EmploymentType() string       { return p.EmploymentTypeField }
func (p Position) GradeLevel() *string          { return p.GradeLevelField }
func (p Position) HeadcountCapacity() float64   { return p.HeadcountCapacityField }
func (p Position) HeadcountInUse() float64      { return p.HeadcountInUseField }
func (p Position) OrganizationName() *string    { return p.OrganizationNameField }
func (p Position) AvailableHeadcount() float64 {
	available := p.HeadcountCapacityField - p.HeadcountInUseField
	if available < 0 {
		return 0
	}
	return available
}
func (p Position) ReportsToPositionCode() *PositionCode {
	if p.ReportsToPositionField == nil {
		return nil
	}
	value := PositionCode(strings.TrimSpace(*p.ReportsToPositionField))
	if value == "" {
		return nil
	}
	return &value
}
func (p Position) Status() string { return p.StatusField }
func (p Position) EffectiveDate() Date {
	return Date(p.EffectiveDateField.Format("2006-01-02"))
}
func (p Position) EndDate() *Date {
	if p.EndDateField == nil {
		return nil
	}
	val := Date(p.EndDateField.Format("2006-01-02"))
	return &val
}
func (p Position) IsCurrent() bool { return p.IsCurrentField }
func (p Position) IsFuture() bool {
	today := cnTodayDate()
	return p.EffectiveDateField.After(today)
}
func (p Position) CreatedAt() DateTime { return DateTime(p.CreatedAtField.Format(time.RFC3339)) }
func (p Position) UpdatedAt() DateTime { return DateTime(p.UpdatedAtField.Format(time.RFC3339)) }
func (p Position) CurrentAssignment() *PositionAssignment {
	return p.CurrentAssignmentField
}

func (p Position) AssignmentHistory() []PositionAssignment {
	if p.AssignmentHistoryField == nil {
		return []PositionAssignment{}
	}
	return p.AssignmentHistoryField
}

// PositionConnection 连接结果
type PositionConnection struct {
	EdgesField      []PositionEdge `json:"edges"`
	DataField       []Position     `json:"data"`
	PaginationField PaginationInfo `json:"pagination"`
	TotalCountField int            `json:"totalCount"`
}

func (c PositionConnection) Edges() []PositionEdge      { return c.EdgesField }
func (c PositionConnection) Data() []Position           { return c.DataField }
func (c PositionConnection) Pagination() PaginationInfo { return c.PaginationField }
func (c PositionConnection) TotalCount() int32          { return int32(c.TotalCountField) }

// PositionEdge 用于游标分页
type PositionEdge struct {
	CursorField string   `json:"cursor"`
	NodeField   Position `json:"node"`
}

func (e PositionEdge) Cursor() string { return e.CursorField }
func (e PositionEdge) Node() Position { return e.NodeField }

// PositionAssignment 表示职位任职记录
type PositionAssignment struct {
	AssignmentIDField     string     `json:"assignmentId" db:"assignment_id"`
	TenantIDField         string     `json:"tenantId" db:"tenant_id"`
	PositionCodeField     string     `json:"positionCode" db:"position_code"`
	PositionRecordIDField string     `json:"positionRecordId" db:"position_record_id"`
	EmployeeIDField       string     `json:"employeeId" db:"employee_id"`
	EmployeeNameField     string     `json:"employeeName" db:"employee_name"`
	EmployeeNumberField   *string    `json:"employeeNumber" db:"employee_number"`
	AssignmentTypeField   string     `json:"assignmentType" db:"assignment_type"`
	AssignmentStatusField string     `json:"assignmentStatus" db:"assignment_status"`
	FTEField              float64    `json:"fte" db:"fte"`
	EffectiveDateField    time.Time  `json:"effectiveDate" db:"effective_date"`
	EndDateField          *time.Time `json:"endDate" db:"end_date"`
	ActingUntilField      *time.Time `json:"actingUntil" db:"acting_until"`
	AutoRevertField       bool       `json:"autoRevert" db:"auto_revert"`
	ReminderSentAtField   *time.Time `json:"reminderSentAt" db:"reminder_sent_at"`
	IsCurrentField        bool       `json:"isCurrent" db:"is_current"`
	NotesField            *string    `json:"notes" db:"notes"`
	CreatedAtField        time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAtField        time.Time  `json:"updatedAt" db:"updated_at"`
}

func (a PositionAssignment) AssignmentId() UUID         { return UUID(a.AssignmentIDField) }
func (a PositionAssignment) TenantId() UUID             { return UUID(a.TenantIDField) }
func (a PositionAssignment) PositionCode() PositionCode { return PositionCode(a.PositionCodeField) }
func (a PositionAssignment) PositionRecordId() UUID     { return UUID(a.PositionRecordIDField) }
func (a PositionAssignment) EmployeeId() UUID           { return UUID(a.EmployeeIDField) }
func (a PositionAssignment) EmployeeName() string       { return a.EmployeeNameField }
func (a PositionAssignment) EmployeeNumber() *string    { return a.EmployeeNumberField }
func (a PositionAssignment) AssignmentType() string     { return a.AssignmentTypeField }
func (a PositionAssignment) AssignmentStatus() string   { return a.AssignmentStatusField }
func (a PositionAssignment) Fte() float64               { return a.FTEField }
func (a PositionAssignment) EffectiveDate() Date {
	return Date(a.EffectiveDateField.Format("2006-01-02"))
}
func (a PositionAssignment) EndDate() *Date {
	if a.EndDateField == nil {
		return nil
	}
	val := Date(a.EndDateField.Format("2006-01-02"))
	return &val
}
func (a PositionAssignment) ActingUntil() *Date {
	if a.ActingUntilField == nil {
		return nil
	}
	val := Date(a.ActingUntilField.Format("2006-01-02"))
	return &val
}
func (a PositionAssignment) AutoRevert() bool { return a.AutoRevertField }
func (a PositionAssignment) ReminderSentAt() *DateTime {
	if a.ReminderSentAtField == nil {
		return nil
	}
	val := DateTime(a.ReminderSentAtField.Format(time.RFC3339))
	return &val
}
func (a PositionAssignment) IsCurrent() bool { return a.IsCurrentField }
func (a PositionAssignment) Notes() *string  { return a.NotesField }
func (a PositionAssignment) CreatedAt() DateTime {
	return DateTime(a.CreatedAtField.Format(time.RFC3339))
}
func (a PositionAssignment) UpdatedAt() DateTime {
	return DateTime(a.UpdatedAtField.Format(time.RFC3339))
}

// PositionAssignmentEdge 游标数据
type PositionAssignmentEdge struct {
	CursorField string             `json:"cursor"`
	NodeField   PositionAssignment `json:"node"`
}

func (e PositionAssignmentEdge) Cursor() string           { return e.CursorField }
func (e PositionAssignmentEdge) Node() PositionAssignment { return e.NodeField }

// PositionAssignmentConnection 连接响应
type PositionAssignmentConnection struct {
	EdgesField      []PositionAssignmentEdge `json:"edges"`
	DataField       []PositionAssignment     `json:"data"`
	PaginationField PaginationInfo           `json:"pagination"`
	TotalCountField int                      `json:"totalCount"`
}

func (c PositionAssignmentConnection) Edges() []PositionAssignmentEdge {
	return c.EdgesField
}
func (c PositionAssignmentConnection) Data() []PositionAssignment { return c.DataField }
func (c PositionAssignmentConnection) Pagination() PaginationInfo {
	return c.PaginationField
}
func (c PositionAssignmentConnection) TotalCount() int32 {
	return int32(c.TotalCountField)
}

type PositionAssignmentAudit struct {
	AssignmentIDField  string                 `json:"assignmentId"`
	EventTypeField     string                 `json:"eventType"`
	EffectiveDateField time.Time              `json:"effectiveDate"`
	EndDateField       *time.Time             `json:"endDate"`
	ActorField         string                 `json:"actor"`
	ChangesField       map[string]interface{} `json:"changes"`
	CreatedAtField     time.Time              `json:"createdAt"`
}

func (a PositionAssignmentAudit) AssignmentId() UUID { return UUID(a.AssignmentIDField) }
func (a PositionAssignmentAudit) EventType() string  { return a.EventTypeField }
func (a PositionAssignmentAudit) EffectiveDate() Date {
	return Date(a.EffectiveDateField.Format("2006-01-02"))
}
func (a PositionAssignmentAudit) EndDate() *Date {
	if a.EndDateField == nil {
		return nil
	}
	val := Date(a.EndDateField.Format("2006-01-02"))
	return &val
}
func (a PositionAssignmentAudit) Actor() string { return a.ActorField }
func (a PositionAssignmentAudit) Changes() *JSON {
	if a.ChangesField == nil {
		return nil
	}
	val := JSON(a.ChangesField)
	return &val
}
func (a PositionAssignmentAudit) CreatedAt() DateTime {
	return DateTime(a.CreatedAtField.Format(time.RFC3339))
}

type PositionAssignmentAuditConnection struct {
	DataField       []PositionAssignmentAudit `json:"data"`
	PaginationField PaginationInfo            `json:"pagination"`
	TotalCountField int                       `json:"totalCount"`
}

func (c PositionAssignmentAuditConnection) Data() []PositionAssignmentAudit { return c.DataField }
func (c PositionAssignmentAuditConnection) Pagination() PaginationInfo      { return c.PaginationField }
func (c PositionAssignmentAuditConnection) TotalCount() int32               { return int32(c.TotalCountField) }

// VacantPosition 空缺职位视图
type VacantPosition struct {
	PositionCodeField       string    `json:"positionCode" db:"position_code"`
	OrganizationCodeField   string    `json:"organizationCode" db:"organization_code"`
	OrganizationNameField   *string   `json:"organizationName" db:"organization_name"`
	JobFamilyCodeField      string    `json:"jobFamilyCode" db:"job_family_code"`
	JobRoleCodeField        string    `json:"jobRoleCode" db:"job_role_code"`
	JobLevelCodeField       string    `json:"jobLevelCode" db:"job_level_code"`
	VacantSinceField        time.Time `json:"vacantSince" db:"vacant_since"`
	HeadcountCapacityField  float64   `json:"headcountCapacity" db:"headcount_capacity"`
	HeadcountAvailableField float64   `json:"headcountAvailable" db:"headcount_available"`
	TotalAssignmentsField   int       `json:"totalAssignments" db:"total_assignments"`
}

func (v VacantPosition) PositionCode() PositionCode { return PositionCode(v.PositionCodeField) }
func (v VacantPosition) OrganizationCode() string   { return v.OrganizationCodeField }
func (v VacantPosition) OrganizationName() *string  { return v.OrganizationNameField }
func (v VacantPosition) JobFamilyCode() JobFamilyCode {
	return JobFamilyCode(v.JobFamilyCodeField)
}
func (v VacantPosition) JobRoleCode() JobRoleCode { return JobRoleCode(v.JobRoleCodeField) }
func (v VacantPosition) JobLevelCode() JobLevelCode {
	return JobLevelCode(v.JobLevelCodeField)
}
func (v VacantPosition) VacantSince() Date {
	return Date(v.VacantSinceField.Format("2006-01-02"))
}
func (v VacantPosition) HeadcountCapacity() float64  { return v.HeadcountCapacityField }
func (v VacantPosition) HeadcountAvailable() float64 { return v.HeadcountAvailableField }
func (v VacantPosition) TotalAssignments() int32     { return int32(v.TotalAssignmentsField) }

// VacantPositionEdge 游标包装
type VacantPositionEdge struct {
	CursorField string         `json:"cursor"`
	NodeField   VacantPosition `json:"node"`
}

func (e VacantPositionEdge) Cursor() string       { return e.CursorField }
func (e VacantPositionEdge) Node() VacantPosition { return e.NodeField }

// VacantPositionConnection 连接响应
type VacantPositionConnection struct {
	EdgesField      []VacantPositionEdge `json:"edges"`
	DataField       []VacantPosition     `json:"data"`
	PaginationField PaginationInfo       `json:"pagination"`
	TotalCountField int                  `json:"totalCount"`
}

func (c VacantPositionConnection) Edges() []VacantPositionEdge { return c.EdgesField }
func (c VacantPositionConnection) Data() []VacantPosition      { return c.DataField }
func (c VacantPositionConnection) Pagination() PaginationInfo  { return c.PaginationField }
func (c VacantPositionConnection) TotalCount() int32           { return int32(c.TotalCountField) }

// PositionTransfer 职位转移记录
type PositionTransfer struct {
	TransferIDField           string         `json:"transferId"`
	PositionCodeField         string         `json:"positionCode"`
	FromOrganizationCodeField string         `json:"fromOrganizationCode"`
	ToOrganizationCodeField   string         `json:"toOrganizationCode"`
	EffectiveDateField        time.Time      `json:"effectiveDate"`
	InitiatedByField          OperatedByData `json:"initiatedBy"`
	OperationReasonField      *string        `json:"operationReason"`
	CreatedAtField            time.Time      `json:"createdAt"`
}

func (t PositionTransfer) TransferId() UUID             { return UUID(t.TransferIDField) }
func (t PositionTransfer) PositionCode() PositionCode   { return PositionCode(t.PositionCodeField) }
func (t PositionTransfer) FromOrganizationCode() string { return t.FromOrganizationCodeField }
func (t PositionTransfer) ToOrganizationCode() string   { return t.ToOrganizationCodeField }
func (t PositionTransfer) EffectiveDate() Date {
	return Date(t.EffectiveDateField.Format("2006-01-02"))
}
func (t PositionTransfer) InitiatedBy() OperatedByData { return t.InitiatedByField }
func (t PositionTransfer) OperationReason() *string    { return t.OperationReasonField }
func (t PositionTransfer) CreatedAt() DateTime {
	return DateTime(t.CreatedAtField.Format(time.RFC3339))
}

// PositionTransferEdge 游标结果
type PositionTransferEdge struct {
	CursorField string           `json:"cursor"`
	NodeField   PositionTransfer `json:"node"`
}

func (e PositionTransferEdge) Cursor() string         { return e.CursorField }
func (e PositionTransferEdge) Node() PositionTransfer { return e.NodeField }

// PositionTransferConnection 连接结果
type PositionTransferConnection struct {
	EdgesField      []PositionTransferEdge `json:"edges"`
	DataField       []PositionTransfer     `json:"data"`
	PaginationField PaginationInfo         `json:"pagination"`
	TotalCountField int                    `json:"totalCount"`
}

func (c PositionTransferConnection) Edges() []PositionTransferEdge { return c.EdgesField }
func (c PositionTransferConnection) Data() []PositionTransfer      { return c.DataField }
func (c PositionTransferConnection) Pagination() PaginationInfo    { return c.PaginationField }
func (c PositionTransferConnection) TotalCount() int32             { return int32(c.TotalCountField) }

// PositionFilterInput 过滤条件
type PositionFilterInput struct {
	OrganizationCode    *string         `json:"organizationCode"`
	PositionCodes       *[]string       `json:"positionCodes"`
	Status              *string         `json:"status"`
	JobFamilyGroupCodes *[]string       `json:"jobFamilyGroupCodes"`
	JobFamilyCodes      *[]string       `json:"jobFamilyCodes"`
	JobRoleCodes        *[]string       `json:"jobRoleCodes"`
	JobLevelCodes       *[]string       `json:"jobLevelCodes"`
	PositionTypes       *[]string       `json:"positionTypes"`
	EmploymentTypes     *[]string       `json:"employmentTypes"`
	EffectiveRange      *DateRangeInput `json:"effectiveRange"`
}

func (f *PositionFilterInput) UnmarshalGraphQL(input interface{}) error {
	raw, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("PositionFilterInput: 期望对象类型，实际得到 %T", input)
	}

	if value, exists := raw["organizationCode"]; exists {
		strPtr, err := asOptionalString(value)
		if err != nil {
			return fmt.Errorf("PositionFilterInput.organizationCode: %w", err)
		}
		f.OrganizationCode = strPtr
	}
	if value, exists := raw["positionCodes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("PositionFilterInput.positionCodes: %w", err)
		}
		f.PositionCodes = slicePtr
	}
	if value, exists := raw["status"]; exists {
		strPtr, err := asOptionalString(value)
		if err != nil {
			return fmt.Errorf("PositionFilterInput.status: %w", err)
		}
		f.Status = strPtr
	}
	if value, exists := raw["jobFamilyGroupCodes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("PositionFilterInput.jobFamilyGroupCodes: %w", err)
		}
		f.JobFamilyGroupCodes = slicePtr
	}
	if value, exists := raw["jobFamilyCodes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("PositionFilterInput.jobFamilyCodes: %w", err)
		}
		f.JobFamilyCodes = slicePtr
	}
	if value, exists := raw["jobRoleCodes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("PositionFilterInput.jobRoleCodes: %w", err)
		}
		f.JobRoleCodes = slicePtr
	}
	if value, exists := raw["jobLevelCodes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("PositionFilterInput.jobLevelCodes: %w", err)
		}
		f.JobLevelCodes = slicePtr
	}
	if value, exists := raw["positionTypes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("PositionFilterInput.positionTypes: %w", err)
		}
		f.PositionTypes = slicePtr
	}
	if value, exists := raw["employmentTypes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("PositionFilterInput.employmentTypes: %w", err)
		}
		f.EmploymentTypes = slicePtr
	}
	if value, exists := raw["effectiveRange"]; exists {
		rangePtr, err := asOptionalDateRange(value)
		if err != nil {
			return fmt.Errorf("PositionFilterInput.effectiveRange: %w", err)
		}
		f.EffectiveRange = rangePtr
	}

	return nil
}

// PositionSortInput 排序输入
type PositionSortInput struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

func (s *PositionSortInput) UnmarshalGraphQL(input interface{}) error {
	raw, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("PositionSortInput: 期望对象类型，实际得到 %T", input)
	}

	field, err := asRequiredString(raw, "field")
	if err != nil {
		return fmt.Errorf("PositionSortInput.field: %w", err)
	}
	direction, err := asRequiredString(raw, "direction")
	if err != nil {
		return fmt.Errorf("PositionSortInput.direction: %w", err)
	}

	s.Field = field
	s.Direction = direction
	return nil
}

// VacantPositionFilterInput 空缺职位过滤条件
type VacantPositionFilterInput struct {
	OrganizationCodes *[]string `json:"organizationCodes"`
	JobFamilyCodes    *[]string `json:"jobFamilyCodes"`
	JobRoleCodes      *[]string `json:"jobRoleCodes"`
	JobLevelCodes     *[]string `json:"jobLevelCodes"`
	PositionTypes     *[]string `json:"positionTypes"`
	MinimumVacantDays *int      `json:"minimumVacantDays"`
	AsOfDate          *string   `json:"asOfDate"`
}

func (f *VacantPositionFilterInput) UnmarshalGraphQL(input interface{}) error {
	raw, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("VacantPositionFilterInput: 期望对象类型，实际得到 %T", input)
	}

	if value, exists := raw["organizationCodes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("VacantPositionFilterInput.organizationCodes: %w", err)
		}
		f.OrganizationCodes = slicePtr
	}
	if value, exists := raw["jobFamilyCodes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("VacantPositionFilterInput.jobFamilyCodes: %w", err)
		}
		f.JobFamilyCodes = slicePtr
	}
	if value, exists := raw["jobRoleCodes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("VacantPositionFilterInput.jobRoleCodes: %w", err)
		}
		f.JobRoleCodes = slicePtr
	}
	if value, exists := raw["jobLevelCodes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("VacantPositionFilterInput.jobLevelCodes: %w", err)
		}
		f.JobLevelCodes = slicePtr
	}
	if value, exists := raw["positionTypes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("VacantPositionFilterInput.positionTypes: %w", err)
		}
		f.PositionTypes = slicePtr
	}
	if value, exists := raw["minimumVacantDays"]; exists {
		intPtr, err := asOptionalInt(value)
		if err != nil {
			return fmt.Errorf("VacantPositionFilterInput.minimumVacantDays: %w", err)
		}
		f.MinimumVacantDays = intPtr
	}
	if value, exists := raw["asOfDate"]; exists {
		strPtr, err := asOptionalString(value)
		if err != nil {
			return fmt.Errorf("VacantPositionFilterInput.asOfDate: %w", err)
		}
		f.AsOfDate = strPtr
	}

	return nil
}

// VacantPositionSortInput 空缺职位排序输入
type VacantPositionSortInput struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

func (s *VacantPositionSortInput) UnmarshalGraphQL(input interface{}) error {
	raw, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("VacantPositionSortInput: 期望对象类型，实际得到 %T", input)
	}

	field, err := asRequiredString(raw, "field")
	if err != nil {
		return fmt.Errorf("VacantPositionSortInput.field: %w", err)
	}
	direction, err := asOptionalString(raw["direction"])
	if err != nil {
		return fmt.Errorf("VacantPositionSortInput.direction: %w", err)
	}

	s.Field = field
	if direction != nil {
		s.Direction = *direction
	} else {
		s.Direction = "DESC"
	}
	return nil
}

// PositionAssignmentFilterInput GraphQL 任职过滤条件
type PositionAssignmentFilterInput struct {
	EmployeeID        *string         `json:"employeeId"`
	Status            *string         `json:"status"`
	AssignmentTypes   *[]string       `json:"assignmentTypes"`
	DateRange         *DateRangeInput `json:"dateRange"`
	AsOfDate          *string         `json:"asOfDate"`
	IncludeHistorical bool            `json:"includeHistorical"`
	IncludeActingOnly bool            `json:"includeActingOnly"`
}

func (f *PositionAssignmentFilterInput) UnmarshalGraphQL(input interface{}) error {
	raw, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("PositionAssignmentFilterInput: 期望对象类型，实际得到 %T", input)
	}

	f.IncludeHistorical = true

	if value, exists := raw["employeeId"]; exists {
		strPtr, err := asOptionalString(value)
		if err != nil {
			return fmt.Errorf("PositionAssignmentFilterInput.employeeId: %w", err)
		}
		f.EmployeeID = strPtr
	}
	if value, exists := raw["status"]; exists {
		strPtr, err := asOptionalString(value)
		if err != nil {
			return fmt.Errorf("PositionAssignmentFilterInput.status: %w", err)
		}
		f.Status = strPtr
	}
	if value, exists := raw["assignmentTypes"]; exists {
		slicePtr, err := asOptionalStringSlice(value)
		if err != nil {
			return fmt.Errorf("PositionAssignmentFilterInput.assignmentTypes: %w", err)
		}
		f.AssignmentTypes = slicePtr
	}
	if value, exists := raw["dateRange"]; exists {
		rangePtr, err := asOptionalDateRange(value)
		if err != nil {
			return fmt.Errorf("PositionAssignmentFilterInput.dateRange: %w", err)
		}
		f.DateRange = rangePtr
	}
	if value, exists := raw["asOfDate"]; exists {
		strPtr, err := asOptionalString(value)
		if err != nil {
			return fmt.Errorf("PositionAssignmentFilterInput.asOfDate: %w", err)
		}
		f.AsOfDate = strPtr
	}
	if value, exists := raw["includeHistorical"]; exists {
		boolVal, err := asBool(value)
		if err != nil {
			return fmt.Errorf("PositionAssignmentFilterInput.includeHistorical: %w", err)
		}
		f.IncludeHistorical = boolVal
	}
	if value, exists := raw["includeActingOnly"]; exists {
		boolVal, err := asBool(value)
		if err != nil {
			return fmt.Errorf("PositionAssignmentFilterInput.includeActingOnly: %w", err)
		}
		f.IncludeActingOnly = boolVal
	}
	return nil
}

// PositionAssignmentSortInput 任职排序输入
type PositionAssignmentSortInput struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

func (s *PositionAssignmentSortInput) UnmarshalGraphQL(input interface{}) error {
	raw, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("PositionAssignmentSortInput: 期望对象类型，实际得到 %T", input)
	}

	field, err := asRequiredString(raw, "field")
	if err != nil {
		return fmt.Errorf("PositionAssignmentSortInput.field: %w", err)
	}
	direction, err := asOptionalString(raw["direction"])
	if err != nil {
		return fmt.Errorf("PositionAssignmentSortInput.direction: %w", err)
	}

	s.Field = field
	if direction != nil {
		s.Direction = *direction
	} else {
		s.Direction = "DESC"
	}
	return nil
}

// PositionTimelineEntry 时间线条目
type PositionTimelineEntry struct {
	RecordIDField         string     `json:"recordId" db:"record_id"`
	StatusField           string     `json:"status" db:"status"`
	TitleField            string     `json:"title" db:"title"`
	EffectiveDateField    time.Time  `json:"effectiveDate" db:"effective_date"`
	EndDateField          *time.Time `json:"endDate" db:"end_date"`
	IsCurrentField        bool       `json:"isCurrent" db:"is_current"`
	ChangeReasonField     *string    `json:"changeReason" db:"operation_reason"`
	TimelineCategoryField string     `json:"timelineCategory" db:"timeline_category"`
	AssignmentTypeField   *string    `json:"assignmentType" db:"assignment_type"`
	AssignmentStatusField *string    `json:"assignmentStatus" db:"assignment_status"`
}

func (e PositionTimelineEntry) RecordId() UUID { return UUID(e.RecordIDField) }
func (e PositionTimelineEntry) Status() string { return e.StatusField }
func (e PositionTimelineEntry) Title() string  { return e.TitleField }
func (e PositionTimelineEntry) EffectiveDate() Date {
	return Date(e.EffectiveDateField.Format("2006-01-02"))
}
func (e PositionTimelineEntry) EndDate() *Date {
	if e.EndDateField == nil {
		return nil
	}
	val := Date(e.EndDateField.Format("2006-01-02"))
	return &val
}
func (e PositionTimelineEntry) IsCurrent() bool { return e.IsCurrentField }
func (e PositionTimelineEntry) ChangeReason() *string {
	return e.ChangeReasonField
}

func (e PositionTimelineEntry) TimelineCategory() string {
	if strings.TrimSpace(e.TimelineCategoryField) == "" {
		return "POSITION_VERSION"
	}
	return e.TimelineCategoryField
}

func (e PositionTimelineEntry) AssignmentType() *string {
	return e.AssignmentTypeField
}

func (e PositionTimelineEntry) AssignmentStatus() *string {
	return e.AssignmentStatusField
}

// HeadcountStats 编制统计
type HeadcountStats struct {
	OrganizationCodeField string            `json:"organizationCode"`
	OrganizationNameField string            `json:"organizationName"`
	TotalCapacityField    float64           `json:"totalCapacity"`
	TotalFilledField      float64           `json:"totalFilled"`
	TotalAvailableField   float64           `json:"totalAvailable"`
	LevelBreakdownField   []LevelHeadcount  `json:"levelBreakdown"`
	TypeBreakdownField    []TypeHeadcount   `json:"typeBreakdown"`
	FamilyBreakdownField  []FamilyHeadcount `json:"familyBreakdown"`
}

func (h HeadcountStats) OrganizationCode() string { return h.OrganizationCodeField }
func (h HeadcountStats) OrganizationName() string { return h.OrganizationNameField }
func (h HeadcountStats) TotalCapacity() float64   { return h.TotalCapacityField }
func (h HeadcountStats) TotalFilled() float64     { return h.TotalFilledField }
func (h HeadcountStats) TotalAvailable() float64  { return h.TotalAvailableField }
func (h HeadcountStats) FillRate() float64 {
	if h.TotalCapacityField <= 0 {
		return 0
	}
	rate := h.TotalFilledField / h.TotalCapacityField
	if rate < 0 {
		return 0
	}
	if rate > 1 {
		return 1
	}
	return rate
}
func (h HeadcountStats) ByLevel() []LevelHeadcount   { return h.LevelBreakdownField }
func (h HeadcountStats) ByType() []TypeHeadcount     { return h.TypeBreakdownField }
func (h HeadcountStats) ByFamily() []FamilyHeadcount { return h.FamilyBreakdownField }

// LevelHeadcount 按职级统计
type LevelHeadcount struct {
	JobLevelCodeField string  `json:"jobLevelCode" db:"job_level_code"`
	CapacityField     float64 `json:"capacity" db:"capacity"`
	UtilizedField     float64 `json:"utilized" db:"utilized"`
	AvailableField    float64 `json:"available" db:"available"`
}

// FamilyHeadcount 按职种统计
type FamilyHeadcount struct {
	JobFamilyCodeField string  `json:"jobFamilyCode" db:"job_family_code"`
	JobFamilyNameField *string `json:"jobFamilyName" db:"job_family_name"`
	CapacityField      float64 `json:"capacity" db:"capacity"`
	UtilizedField      float64 `json:"utilized" db:"utilized"`
	AvailableField     float64 `json:"available" db:"available"`
}

func (f FamilyHeadcount) JobFamilyCode() JobFamilyCode {
	return JobFamilyCode(f.JobFamilyCodeField)
}
func (f FamilyHeadcount) JobFamilyName() *string {
	return f.JobFamilyNameField
}
func (f FamilyHeadcount) Capacity() float64  { return f.CapacityField }
func (f FamilyHeadcount) Utilized() float64  { return f.UtilizedField }
func (f FamilyHeadcount) Available() float64 { return f.AvailableField }

func (l LevelHeadcount) JobLevelCode() JobLevelCode { return JobLevelCode(l.JobLevelCodeField) }
func (l LevelHeadcount) Capacity() float64          { return l.CapacityField }
func (l LevelHeadcount) Utilized() float64          { return l.UtilizedField }
func (l LevelHeadcount) Available() float64         { return l.AvailableField }

// TypeHeadcount 按职位类型统计
type TypeHeadcount struct {
	PositionTypeField string  `json:"positionType" db:"position_type"`
	CapacityField     float64 `json:"capacity" db:"capacity"`
	FilledField       float64 `json:"filled" db:"filled"`
	AvailableField    float64 `json:"available" db:"available"`
}

func (t TypeHeadcount) PositionType() string { return t.PositionTypeField }
func (t TypeHeadcount) Capacity() float64    { return t.CapacityField }
func (t TypeHeadcount) Filled() float64      { return t.FilledField }
func (t TypeHeadcount) Available() float64   { return t.AvailableField }

// Job catalog 基础类型
type JobFamilyGroup struct {
	RecordIDField      string     `json:"recordId" db:"record_id"`
	TenantIDField      string     `json:"tenantId" db:"tenant_id"`
	CodeField          string     `json:"code" db:"family_group_code"`
	NameField          string     `json:"name" db:"name"`
	DescriptionField   *string    `json:"description" db:"description"`
	StatusField        string     `json:"status" db:"status"`
	EffectiveDateField time.Time  `json:"effectiveDate" db:"effective_date"`
	EndDateField       *time.Time `json:"endDate" db:"end_date"`
	IsCurrentField     bool       `json:"isCurrent" db:"is_current"`
}

func (g JobFamilyGroup) RecordId() UUID { return UUID(g.RecordIDField) }
func (g JobFamilyGroup) TenantId() UUID { return UUID(g.TenantIDField) }
func (g JobFamilyGroup) Code() JobFamilyGroupCode {
	return JobFamilyGroupCode(g.CodeField)
}
func (g JobFamilyGroup) Name() string         { return g.NameField }
func (g JobFamilyGroup) Description() *string { return g.DescriptionField }
func (g JobFamilyGroup) Status() string       { return g.StatusField }
func (g JobFamilyGroup) EffectiveDate() Date {
	return Date(g.EffectiveDateField.Format("2006-01-02"))
}
func (g JobFamilyGroup) EndDate() *Date {
	if g.EndDateField == nil {
		return nil
	}
	val := Date(g.EndDateField.Format("2006-01-02"))
	return &val
}
func (g JobFamilyGroup) IsCurrent() bool { return g.IsCurrentField }

type JobFamily struct {
	RecordIDField        string     `json:"recordId" db:"record_id"`
	TenantIDField        string     `json:"tenantId" db:"tenant_id"`
	CodeField            string     `json:"code" db:"family_code"`
	NameField            string     `json:"name" db:"name"`
	DescriptionField     *string    `json:"description" db:"description"`
	StatusField          string     `json:"status" db:"status"`
	EffectiveDateField   time.Time  `json:"effectiveDate" db:"effective_date"`
	EndDateField         *time.Time `json:"endDate" db:"end_date"`
	IsCurrentField       bool       `json:"isCurrent" db:"is_current"`
	FamilyGroupCodeField string     `json:"groupCode" db:"family_group_code"`
}

func (f JobFamily) RecordId() UUID       { return UUID(f.RecordIDField) }
func (f JobFamily) TenantId() UUID       { return UUID(f.TenantIDField) }
func (f JobFamily) Code() JobFamilyCode  { return JobFamilyCode(f.CodeField) }
func (f JobFamily) Name() string         { return f.NameField }
func (f JobFamily) Description() *string { return f.DescriptionField }
func (f JobFamily) Status() string       { return f.StatusField }
func (f JobFamily) EffectiveDate() Date {
	return Date(f.EffectiveDateField.Format("2006-01-02"))
}
func (f JobFamily) EndDate() *Date {
	if f.EndDateField == nil {
		return nil
	}
	val := Date(f.EndDateField.Format("2006-01-02"))
	return &val
}
func (f JobFamily) IsCurrent() bool { return f.IsCurrentField }
func (f JobFamily) GroupCode() JobFamilyGroupCode {
	return JobFamilyGroupCode(f.FamilyGroupCodeField)
}

type JobRole struct {
	RecordIDField      string     `json:"recordId" db:"record_id"`
	TenantIDField      string     `json:"tenantId" db:"tenant_id"`
	CodeField          string     `json:"code" db:"role_code"`
	NameField          string     `json:"name" db:"name"`
	DescriptionField   *string    `json:"description" db:"description"`
	StatusField        string     `json:"status" db:"status"`
	EffectiveDateField time.Time  `json:"effectiveDate" db:"effective_date"`
	EndDateField       *time.Time `json:"endDate" db:"end_date"`
	IsCurrentField     bool       `json:"isCurrent" db:"is_current"`
	FamilyCodeField    string     `json:"familyCode" db:"family_code"`
}

func (r JobRole) RecordId() UUID       { return UUID(r.RecordIDField) }
func (r JobRole) TenantId() UUID       { return UUID(r.TenantIDField) }
func (r JobRole) Code() JobRoleCode    { return JobRoleCode(r.CodeField) }
func (r JobRole) Name() string         { return r.NameField }
func (r JobRole) Description() *string { return r.DescriptionField }
func (r JobRole) Status() string       { return r.StatusField }
func (r JobRole) EffectiveDate() Date {
	return Date(r.EffectiveDateField.Format("2006-01-02"))
}
func (r JobRole) EndDate() *Date {
	if r.EndDateField == nil {
		return nil
	}
	val := Date(r.EndDateField.Format("2006-01-02"))
	return &val
}
func (r JobRole) IsCurrent() bool           { return r.IsCurrentField }
func (r JobRole) FamilyCode() JobFamilyCode { return JobFamilyCode(r.FamilyCodeField) }

type JobLevel struct {
	RecordIDField      string     `json:"recordId" db:"record_id"`
	TenantIDField      string     `json:"tenantId" db:"tenant_id"`
	CodeField          string     `json:"code" db:"level_code"`
	NameField          string     `json:"name" db:"name"`
	DescriptionField   *string    `json:"description" db:"description"`
	StatusField        string     `json:"status" db:"status"`
	EffectiveDateField time.Time  `json:"effectiveDate" db:"effective_date"`
	EndDateField       *time.Time `json:"endDate" db:"end_date"`
	IsCurrentField     bool       `json:"isCurrent" db:"is_current"`
	RoleCodeField      string     `json:"roleCode" db:"role_code"`
	LevelRankField     string     `json:"levelRank" db:"level_rank"`
}

func (l JobLevel) RecordId() UUID       { return UUID(l.RecordIDField) }
func (l JobLevel) TenantId() UUID       { return UUID(l.TenantIDField) }
func (l JobLevel) Code() JobLevelCode   { return JobLevelCode(l.CodeField) }
func (l JobLevel) Name() string         { return l.NameField }
func (l JobLevel) Description() *string { return l.DescriptionField }
func (l JobLevel) Status() string       { return l.StatusField }
func (l JobLevel) EffectiveDate() Date {
	return Date(l.EffectiveDateField.Format("2006-01-02"))
}
func (l JobLevel) EndDate() *Date {
	if l.EndDateField == nil {
		return nil
	}
	val := Date(l.EndDateField.Format("2006-01-02"))
	return &val
}
func (l JobLevel) IsCurrent() bool       { return l.IsCurrentField }
func (l JobLevel) RoleCode() JobRoleCode { return JobRoleCode(l.RoleCodeField) }
func (l JobLevel) LevelRank() int32 {
	trimmed := strings.TrimSpace(l.LevelRankField)
	if trimmed == "" {
		return 0
	}
	if v, err := strconv.Atoi(trimmed); err == nil {
		return int32(v)
	}
	return 0
}

func asOptionalString(value interface{}) (*string, error) {
	if value == nil {
		return nil, nil
	}
	str, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("期望字符串，实际得到 %T", value)
	}
	str = strings.TrimSpace(str)
	if str == "" {
		return nil, nil
	}
	return &str, nil
}

func asOptionalStringSlice(value interface{}) (*[]string, error) {
	if value == nil {
		return nil, nil
	}

	var rawItems []string
	switch v := value.(type) {
	case []string:
		rawItems = v
	case []interface{}:
		rawItems = make([]string, 0, len(v))
		for _, item := range v {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("期望字符串数组，实际包含 %T", item)
			}
			rawItems = append(rawItems, str)
		}
	default:
		return nil, fmt.Errorf("期望字符串数组，实际得到 %T", value)
	}

	result := make([]string, 0, len(rawItems))
	for _, item := range rawItems {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}

	if len(result) == 0 {
		return nil, nil
	}
	return &result, nil
}

func asOptionalDateRange(value interface{}) (*DateRangeInput, error) {
	if value == nil {
		return nil, nil
	}

	raw, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("期望对象类型，实际得到 %T", value)
	}

	from, err := asOptionalString(raw["from"])
	if err != nil {
		return nil, fmt.Errorf("from: %w", err)
	}
	to, err := asOptionalString(raw["to"])
	if err != nil {
		return nil, fmt.Errorf("to: %w", err)
	}

	if from == nil && to == nil {
		return nil, nil
	}

	return &DateRangeInput{
		From: from,
		To:   to,
	}, nil
}

func asBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		lower := strings.TrimSpace(strings.ToLower(v))
		if lower == "true" || lower == "1" || lower == "yes" || lower == "y" {
			return true, nil
		}
		if lower == "false" || lower == "0" || lower == "no" || lower == "n" {
			return false, nil
		}
	}
	return false, fmt.Errorf("期望布尔值，实际得到 %T", value)
}

func asOptionalBool(value interface{}) (*bool, error) {
	if value == nil {
		return nil, nil
	}
	boolVal, err := asBool(value)
	if err != nil {
		return nil, err
	}
	return &boolVal, nil
}

func asRequiredString(raw map[string]interface{}, key string) (string, error) {
	value, exists := raw[key]
	if !exists {
		return "", fmt.Errorf("缺少必填字段 %q", key)
	}
	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("字段 %q 期望为字符串，实际得到 %T", key, value)
	}
	str = strings.TrimSpace(str)
	if str == "" {
		return "", fmt.Errorf("字段 %q 不能为空字符串", key)
	}
	return str, nil
}

func asOptionalInt(value interface{}) (*int, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case int:
		return &v, nil
	case int32:
		val := int(v)
		return &val, nil
	case int64:
		val := int(v)
		return &val, nil
	case float64:
		return nil, fmt.Errorf("期望整数，实际得到浮点数")
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil, nil
		}
		parsed, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, fmt.Errorf("期望整数，实际得到 %q", trimmed)
		}
		return &parsed, nil
	default:
		return nil, fmt.Errorf("期望整数，实际得到 %T", value)
	}
}
