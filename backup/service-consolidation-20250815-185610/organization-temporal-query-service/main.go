package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"fmt"
	"crypto/md5"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// 扩展的GraphQL Schema定义 - 支持时态查询
var schemaString = `
	type Organization {
		code: String!
		name: String!
		unitType: String!
		status: String!
		level: Int!
		path: String
		sortOrder: Int
		description: String
		profile: String
		parentCode: String
		createdAt: String!
		updatedAt: String!
		
		# 时态字段
		effectiveDate: String
		endDate: String
		version: Int
		supersedesVersion: Int
		changeReason: String
		isCurrent: Boolean
	}
	
	type OrganizationStats {
		totalCount: Int!
		activeCount: Int!
		inactiveCount: Int!
		plannedCount: Int!
		byType: [TypeStat!]!
		byLevel: [LevelStat!]!
		
		# 时态统计
		currentVersionsCount: Int!
		historicalVersionsCount: Int!
		dissolvedCount: Int!
	}
	
	type TypeStat {
		type: String!
		count: Int!
	}
	
	type LevelStat {
		level: Int!
		count: Int!
	}
	
	# 时态查询输入类型
	input TemporalQueryInput {
		asOfDate: String          # YYYY-MM-DD格式
		effectiveFrom: String     # YYYY-MM-DD格式
		effectiveTo: String       # YYYY-MM-DD格式
		includeHistory: Boolean   # 是否包含历史版本
		includeFuture: Boolean    # 是否包含未来版本
		includeDissolved: Boolean # 是否包含已解散组织
		version: Int              # 特定版本查询
		maxVersions: Int          # 最大版本数量
	}
	
	# 组织变更历史
	type OrganizationHistory {
		organization: Organization!
		changeEvents: [ChangeEvent!]!
		versionTimeline: [VersionInfo!]!
	}
	
	type ChangeEvent {
		eventId: String!
		eventType: String!
		effectiveDate: String!
		endDate: String
		changeData: String!  # JSON格式的变更数据
		changeReason: String
		createdBy: String
		createdAt: String!
	}
	
	type VersionInfo {
		version: Int!
		effectiveDate: String!
		endDate: String
		changeReason: String
		isCurrent: Boolean!
	}

	type Query {
		organizations: [Organization!]!
		organizationStats: OrganizationStats!
		
		# 基础查询
		organization(code: String!): Organization
		
		# 时态查询
		organizationTemporal(code: String!, query: TemporalQueryInput): [Organization!]!
		organizationsAsOf(date: String!): [Organization!]!
		organizationHistory(code: String!): OrganizationHistory
		
		# 时态范围查询
		organizationsInPeriod(from: String!, to: String!): [Organization!]!
		organizationsByVersion(code: String!, version: Int!): Organization
	}
`

// ===== GraphQL解析器实现 =====

type Resolver struct {
	neo4jDriver neo4j.DriverWithContext
	redisClient *redis.Client
}

type organizationResolver struct {
	org *Organization
}

type organizationStatsResolver struct {
	stats *OrganizationStats
}

type organizationHistoryResolver struct {
	history *OrganizationHistory
}

type typeStatResolver struct {
	stat *TypeStat
}

type levelStatResolver struct {
	stat *LevelStat
}

type changeEventResolver struct {
	event *ChangeEvent
}

type versionInfoResolver struct {
	version *VersionInfo
}

// 组织数据结构 - 扩展时态字段
type Organization struct {
	Code              string  `json:"code"`
	Name              string  `json:"name"`
	UnitType          string  `json:"unitType"`
	Status            string  `json:"status"`
	Level             int32   `json:"level"`
	Path              *string `json:"path"`
	SortOrder         *int32  `json:"sortOrder"`
	Description       *string `json:"description"`
	Profile           *string `json:"profile"`
	ParentCode        *string `json:"parentCode"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
	
	// 时态字段
	EffectiveDate     *string `json:"effectiveDate"`
	EndDate           *string `json:"endDate"`
	Version           *int32  `json:"version"`
	SupersedesVersion *int32  `json:"supersedesVersion"`
	ChangeReason      *string `json:"changeReason"`
	IsCurrent         *bool   `json:"isCurrent"`
}

type OrganizationStats struct {
	TotalCount              int32      `json:"totalCount"`
	ActiveCount             int32      `json:"activeCount"`
	InactiveCount           int32      `json:"inactiveCount"`
	PlannedCount            int32      `json:"plannedCount"`
	ByType                  []TypeStat `json:"byType"`
	ByLevel                 []LevelStat `json:"byLevel"`
	CurrentVersionsCount    int32      `json:"currentVersionsCount"`
	HistoricalVersionsCount int32      `json:"historicalVersionsCount"`
	DissolvedCount          int32      `json:"dissolvedCount"`
}

type TypeStat struct {
	Type  string `json:"type"`
	Count int32  `json:"count"`
}

type LevelStat struct {
	Level int32 `json:"level"`
	Count int32 `json:"count"`
}

type OrganizationHistory struct {
	Organization    *Organization   `json:"organization"`
	ChangeEvents    []ChangeEvent   `json:"changeEvents"`
	VersionTimeline []VersionInfo   `json:"versionTimeline"`
}

type ChangeEvent struct {
	EventID       string  `json:"eventId"`
	EventType     string  `json:"eventType"`
	EffectiveDate string  `json:"effectiveDate"`
	EndDate       *string `json:"endDate"`
	ChangeData    string  `json:"changeData"`
	ChangeReason  *string `json:"changeReason"`
	CreatedBy     *string `json:"createdBy"`
	CreatedAt     string  `json:"createdAt"`
}

type VersionInfo struct {
	Version       int32   `json:"version"`
	EffectiveDate string  `json:"effectiveDate"`
	EndDate       *string `json:"endDate"`
	ChangeReason  *string `json:"changeReason"`
	IsCurrent     bool    `json:"isCurrent"`
}

// 时态查询输入参数
type TemporalQueryInput struct {
	AsOfDate        *string `json:"asOfDate"`
	EffectiveFrom   *string `json:"effectiveFrom"`
	EffectiveTo     *string `json:"effectiveTo"`
	IncludeHistory  *bool   `json:"includeHistory"`
	IncludeFuture   *bool   `json:"includeFuture"`
	IncludeDissolved *bool  `json:"includeDissolved"`
	Version         *int32  `json:"version"`
	MaxVersions     *int32  `json:"maxVersions"`
}

// ===== 基础查询解析器 =====

func (r *Resolver) Organizations(ctx context.Context) ([]*organizationResolver, error) {
	// 使用缓存键
	cacheKey := generateCacheKey("orgs", "all", DefaultTenantIDString)
	
	// 尝试从缓存获取
	if cached := r.getFromCache(ctx, cacheKey); cached != nil {
		if orgs, ok := cached.([]*Organization); ok {
			resolvers := make([]*organizationResolver, len(orgs))
			for i, org := range orgs {
				resolvers[i] = &organizationResolver{org: org}
			}
			return resolvers, nil
		}
	}
	
	// 从数据库查询（只返回当前版本）
	session := r.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	
	cypher := `
		MATCH (o:Organization {tenant_id: $tenantId})
		WHERE o.is_current = true
		RETURN o.code as code, o.name as name, o.unit_type as unitType, 
		       o.status as status, o.level as level, o.path as path,
		       o.sort_order as sortOrder, o.description as description, 
		       o.profile as profile, o.parent_code as parentCode,
		       o.created_at as createdAt, o.updated_at as updatedAt,
		       o.effective_date as effectiveDate, o.end_date as endDate,
		       o.version as version, o.supersedes_version as supersedesVersion,
		       o.change_reason as changeReason, o.is_current as isCurrent
		ORDER BY o.level, o.sort_order, o.name
	`
	
	result, err := session.Run(ctx, cypher, map[string]interface{}{
		"tenantId": DefaultTenantIDString,
	})
	if err != nil {
		return nil, err
	}
	
	var organizations []*Organization
	for result.Next(ctx) {
		record := result.Record()
		org := &Organization{}
		
		if val, ok := record.Get("code"); ok && val != nil {
			org.Code = val.(string)
		}
		if val, ok := record.Get("name"); ok && val != nil {
			org.Name = val.(string)
		}
		if val, ok := record.Get("unitType"); ok && val != nil {
			org.UnitType = val.(string)
		}
		if val, ok := record.Get("status"); ok && val != nil {
			org.Status = val.(string)
		}
		if val, ok := record.Get("level"); ok && val != nil {
			org.Level = int32(val.(int64))
		}
		if val, ok := record.Get("path"); ok && val != nil {
			path := val.(string)
			org.Path = &path
		}
		if val, ok := record.Get("sortOrder"); ok && val != nil {
			sortOrder := int32(val.(int64))
			org.SortOrder = &sortOrder
		}
		if val, ok := record.Get("description"); ok && val != nil {
			desc := val.(string)
			org.Description = &desc
		}
		if val, ok := record.Get("profile"); ok && val != nil {
			profile := val.(string)
			org.Profile = &profile
		}
		if val, ok := record.Get("parentCode"); ok && val != nil {
			parentCode := val.(string)
			org.ParentCode = &parentCode
		}
		if val, ok := record.Get("createdAt"); ok && val != nil {
			org.CreatedAt = val.(string)
		}
		if val, ok := record.Get("updatedAt"); ok && val != nil {
			org.UpdatedAt = val.(string)
		}
		
		// 时态字段
		if val, ok := record.Get("effectiveDate"); ok && val != nil {
			effectiveDate := val.(string)
			org.EffectiveDate = &effectiveDate
		}
		if val, ok := record.Get("endDate"); ok && val != nil {
			endDate := val.(string)
			org.EndDate = &endDate
		}
		if val, ok := record.Get("version"); ok && val != nil {
			version := int32(val.(int64))
			org.Version = &version
		}
		if val, ok := record.Get("supersedesVersion"); ok && val != nil {
			supersedesVersion := int32(val.(int64))
			org.SupersedesVersion = &supersedesVersion
		}
		if val, ok := record.Get("changeReason"); ok && val != nil {
			changeReason := val.(string)
			org.ChangeReason = &changeReason
		}
		if val, ok := record.Get("isCurrent"); ok && val != nil {
			isCurrent := val.(bool)
			org.IsCurrent = &isCurrent
		}
		
		organizations = append(organizations, org)
	}
	
	// 缓存结果
	r.setCache(ctx, cacheKey, organizations, time.Minute*5)
	
	resolvers := make([]*organizationResolver, len(organizations))
	for i, org := range organizations {
		resolvers[i] = &organizationResolver{org: org}
	}
	
	return resolvers, nil
}

// 时态查询解析器
func (r *Resolver) OrganizationTemporal(ctx context.Context, args struct {
	Code  string
	Query *TemporalQueryInput
}) ([]*organizationResolver, error) {
	
	// 生成时态查询的缓存键
	cacheKey := generateTemporalCacheKey("org_temporal", args.Code, args.Query)
	
	// 尝试从缓存获取
	if cached := r.getFromCache(ctx, cacheKey); cached != nil {
		if orgs, ok := cached.([]*Organization); ok {
			resolvers := make([]*organizationResolver, len(orgs))
			for i, org := range orgs {
				resolvers[i] = &organizationResolver{org: org}
			}
			return resolvers, nil
		}
	}
	
	session := r.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	
	// 构建时态查询条件
	var conditions []string
	params := map[string]interface{}{
		"code":     args.Code,
		"tenantId": DefaultTenantIDString,
	}
	
	conditions = append(conditions, "o.code = $code", "o.tenant_id = $tenantId")
	
	// 时间点查询
	if args.Query != nil && args.Query.AsOfDate != nil {
		conditions = append(conditions, 
			"o.effective_date <= date($asOfDate)",
			"(o.end_date IS NULL OR o.end_date >= date($asOfDate))")
		params["asOfDate"] = *args.Query.AsOfDate
	}
	
	// 日期范围查询
	if args.Query != nil && args.Query.EffectiveFrom != nil {
		conditions = append(conditions, "o.effective_date >= date($effectiveFrom)")
		params["effectiveFrom"] = *args.Query.EffectiveFrom
	}
	
	if args.Query != nil && args.Query.EffectiveTo != nil {
		conditions = append(conditions, "o.effective_date <= date($effectiveTo)")
		params["effectiveTo"] = *args.Query.EffectiveTo
	}
	
	// 特定版本查询
	if args.Query != nil && args.Query.Version != nil {
		conditions = append(conditions, "o.version = $version")
		params["version"] = *args.Query.Version
	}
	
	// 当前版本过滤
	if args.Query == nil || (args.Query.IncludeHistory == nil || !*args.Query.IncludeHistory) {
		if args.Query == nil || args.Query.AsOfDate == nil {
			conditions = append(conditions, "o.is_current = true")
		}
	}
	
	// 未来版本过滤
	if args.Query == nil || (args.Query.IncludeFuture == nil || !*args.Query.IncludeFuture) {
		conditions = append(conditions, "o.effective_date <= date()")
	}
	
	// 已解散组织过滤
	if args.Query == nil || (args.Query.IncludeDissolved == nil || !*args.Query.IncludeDissolved) {
		conditions = append(conditions, "(o.end_date IS NULL OR o.end_date > date())")
	}
	
	// 构建完整查询
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}
	
	limitClause := ""
	if args.Query != nil && args.Query.MaxVersions != nil {
		limitClause = fmt.Sprintf("LIMIT %d", *args.Query.MaxVersions)
	}
	
	cypher := fmt.Sprintf(`
		MATCH (o:Organization)
		%s
		RETURN o.code as code, o.name as name, o.unit_type as unitType, 
		       o.status as status, o.level as level, o.path as path,
		       o.sort_order as sortOrder, o.description as description, 
		       o.profile as profile, o.parent_code as parentCode,
		       o.created_at as createdAt, o.updated_at as updatedAt,
		       o.effective_date as effectiveDate, o.end_date as endDate,
		       o.version as version, o.supersedes_version as supersedesVersion,
		       o.change_reason as changeReason, o.is_current as isCurrent
		ORDER BY o.version DESC
		%s
	`, whereClause, limitClause)
	
	result, err := session.Run(ctx, cypher, params)
	if err != nil {
		return nil, fmt.Errorf("时态查询失败: %w", err)
	}
	
	var organizations []*Organization
	for result.Next(ctx) {
		org := parseOrganizationRecord(result.Record())
		organizations = append(organizations, org)
	}
	
	// 缓存时态查询结果 (较短的缓存时间)
	cacheDuration := time.Minute * 5
	if args.Query != nil && args.Query.AsOfDate != nil {
		// 历史查询可以缓存更长时间
		cacheDuration = time.Hour * 1
	}
	r.setCache(ctx, cacheKey, organizations, cacheDuration)
	
	resolvers := make([]*organizationResolver, len(organizations))
	for i, org := range organizations {
		resolvers[i] = &organizationResolver{org: org}
	}
	
	return resolvers, nil
}

// 时间点查询解析器
func (r *Resolver) OrganizationsAsOf(ctx context.Context, args struct {
	Date string
}) ([]*organizationResolver, error) {
	
	// 使用时态查询功能
	query := &TemporalQueryInput{
		AsOfDate: &args.Date,
	}
	
	return r.OrganizationTemporal(ctx, struct {
		Code  string
		Query *TemporalQueryInput
	}{
		Code:  "", // 空code表示查询所有组织
		Query: query,
	})
}

// 组织历史查询解析器
func (r *Resolver) OrganizationHistory(ctx context.Context, args struct {
	Code string
}) (*organizationHistoryResolver, error) {
	
	// TODO: 实现历史查询逻辑，从organization_events和organization_versions表查询
	
	history := &OrganizationHistory{
		Organization:    &Organization{Code: args.Code, Name: "示例组织"},
		ChangeEvents:    []ChangeEvent{},
		VersionTimeline: []VersionInfo{},
	}
	
	return &organizationHistoryResolver{history: history}, nil
}

// ===== 辅助函数 =====

// 解析Neo4j记录为Organization对象
func parseOrganizationRecord(record *neo4j.Record) *Organization {
	org := &Organization{}
	
	if val, ok := record.Get("code"); ok && val != nil {
		org.Code = val.(string)
	}
	if val, ok := record.Get("name"); ok && val != nil {
		org.Name = val.(string)
	}
	if val, ok := record.Get("unitType"); ok && val != nil {
		org.UnitType = val.(string)
	}
	if val, ok := record.Get("status"); ok && val != nil {
		org.Status = val.(string)
	}
	if val, ok := record.Get("level"); ok && val != nil {
		org.Level = int32(val.(int64))
	}
	if val, ok := record.Get("path"); ok && val != nil {
		path := val.(string)
		org.Path = &path
	}
	if val, ok := record.Get("sortOrder"); ok && val != nil {
		sortOrder := int32(val.(int64))
		org.SortOrder = &sortOrder
	}
	if val, ok := record.Get("description"); ok && val != nil {
		desc := val.(string)
		org.Description = &desc
	}
	if val, ok := record.Get("profile"); ok && val != nil {
		profile := val.(string)
		org.Profile = &profile
	}
	if val, ok := record.Get("parentCode"); ok && val != nil {
		parentCode := val.(string)
		org.ParentCode = &parentCode
	}
	if val, ok := record.Get("createdAt"); ok && val != nil {
		org.CreatedAt = val.(string)
	}
	if val, ok := record.Get("updatedAt"); ok && val != nil {
		org.UpdatedAt = val.(string)
	}
	
	// 时态字段解析
	if val, ok := record.Get("effectiveDate"); ok && val != nil {
		effectiveDate := val.(string)
		org.EffectiveDate = &effectiveDate
	}
	if val, ok := record.Get("endDate"); ok && val != nil {
		endDate := val.(string)
		org.EndDate = &endDate
	}
	if val, ok := record.Get("version"); ok && val != nil {
		version := int32(val.(int64))
		org.Version = &version
	}
	if val, ok := record.Get("supersedesVersion"); ok && val != nil {
		supersedesVersion := int32(val.(int64))
		org.SupersedesVersion = &supersedesVersion
	}
	if val, ok := record.Get("changeReason"); ok && val != nil {
		changeReason := val.(string)
		org.ChangeReason = &changeReason
	}
	if val, ok := record.Get("isCurrent"); ok && val != nil {
		isCurrent := val.(bool)
		org.IsCurrent = &isCurrent
	}
	
	return org
}

// 生成时态查询缓存键
func generateTemporalCacheKey(prefix, code string, query *TemporalQueryInput) string {
	var keyParts []string
	keyParts = append(keyParts, prefix, DefaultTenantIDString)
	
	if code != "" {
		keyParts = append(keyParts, code)
	}
	
	if query != nil {
		if query.AsOfDate != nil {
			keyParts = append(keyParts, "as_of", *query.AsOfDate)
		}
		if query.Version != nil {
			keyParts = append(keyParts, "version", fmt.Sprintf("%d", *query.Version))
		}
		if query.IncludeHistory != nil && *query.IncludeHistory {
			keyParts = append(keyParts, "with_history")
		}
		if query.IncludeFuture != nil && *query.IncludeFuture {
			keyParts = append(keyParts, "with_future")
		}
		if query.IncludeDissolved != nil && *query.IncludeDissolved {
			keyParts = append(keyParts, "with_dissolved")
		}
	}
	
	return strings.Join(keyParts, ":")
}

// ===== 组织统计解析器 =====

func (r *Resolver) OrganizationStats(ctx context.Context) (*organizationStatsResolver, error) {
	session := r.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	// 查询时态统计信息
	cypher := `
		MATCH (o:Organization {tenant_id: $tenantId})
		WITH 
			count(CASE WHEN o.is_current = true THEN 1 END) as currentVersionsCount,
			count(CASE WHEN o.is_current = false THEN 1 END) as historicalVersionsCount,
			count(CASE WHEN o.is_current = true AND o.status = 'ACTIVE' THEN 1 END) as activeCount,
			count(CASE WHEN o.is_current = true AND o.status = 'INACTIVE' THEN 1 END) as inactiveCount,
			count(CASE WHEN o.is_current = true AND o.status = 'PLANNED' THEN 1 END) as plannedCount,
			count(CASE WHEN o.end_date IS NOT NULL AND o.end_date <= date() THEN 1 END) as dissolvedCount
		
		MATCH (current:Organization {tenant_id: $tenantId})
		WHERE current.is_current = true
		
		WITH currentVersionsCount, historicalVersionsCount, activeCount, inactiveCount, 
		     plannedCount, dissolvedCount, 
		     collect({type: current.unit_type, level: current.level}) as currentOrgs
		
		UNWIND currentOrgs as org
		WITH currentVersionsCount, historicalVersionsCount, activeCount, inactiveCount,
		     plannedCount, dissolvedCount,
		     org.type as unitType, org.level as level
		     
		RETURN 
			currentVersionsCount,
			historicalVersionsCount,
			activeCount,
			inactiveCount,
			plannedCount,
			dissolvedCount,
			collect(DISTINCT unitType) as unitTypes,
			collect(DISTINCT level) as levels
	`

	result, err := session.Run(ctx, cypher, map[string]interface{}{
		"tenantId": DefaultTenantIDString,
	})
	if err != nil {
		return nil, err
	}

	stats := &OrganizationStats{
		ByType:  []TypeStat{},
		ByLevel: []LevelStat{},
	}

	if result.Next(ctx) {
		record := result.Record()
		
		if val, ok := record.Get("currentVersionsCount"); ok && val != nil {
			stats.CurrentVersionsCount = int32(val.(int64))
			stats.TotalCount = stats.CurrentVersionsCount
		}
		if val, ok := record.Get("historicalVersionsCount"); ok && val != nil {
			stats.HistoricalVersionsCount = int32(val.(int64))
		}
		if val, ok := record.Get("activeCount"); ok && val != nil {
			stats.ActiveCount = int32(val.(int64))
		}
		if val, ok := record.Get("inactiveCount"); ok && val != nil {
			stats.InactiveCount = int32(val.(int64))
		}
		if val, ok := record.Get("plannedCount"); ok && val != nil {
			stats.PlannedCount = int32(val.(int64))
		}
		if val, ok := record.Get("dissolvedCount"); ok && val != nil {
			stats.DissolvedCount = int32(val.(int64))
		}
	}

	return &organizationStatsResolver{stats: stats}, nil
}

func (r *Resolver) Organization(ctx context.Context, args struct{ Code string }) (*organizationResolver, error) {
	// 查询单个组织（当前版本）
	query := &TemporalQueryInput{}
	result, err := r.OrganizationTemporal(ctx, struct {
		Code  string
		Query *TemporalQueryInput
	}{
		Code:  args.Code,
		Query: query,
	})
	
	if err != nil {
		return nil, err
	}
	
	if len(result) > 0 {
		return result[0], nil
	}
	
	return nil, nil
}

// ===== 范围查询解析器 =====

func (r *Resolver) OrganizationsInPeriod(ctx context.Context, args struct {
	From string
	To   string
}) ([]*organizationResolver, error) {
	query := &TemporalQueryInput{
		EffectiveFrom: &args.From,
		EffectiveTo:   &args.To,
		IncludeHistory: func() *bool { b := true; return &b }(),
	}
	
	return r.OrganizationTemporal(ctx, struct {
		Code  string
		Query *TemporalQueryInput
	}{
		Code:  "", // 空code表示查询所有组织
		Query: query,
	})
}

func (r *Resolver) OrganizationsByVersion(ctx context.Context, args struct {
	Code    string
	Version int32
}) (*organizationResolver, error) {
	query := &TemporalQueryInput{
		Version: &args.Version,
	}
	
	result, err := r.OrganizationTemporal(ctx, struct {
		Code  string
		Query *TemporalQueryInput
	}{
		Code:  args.Code,
		Query: query,
	})
	
	if err != nil {
		return nil, err
	}
	
	if len(result) > 0 {
		return result[0], nil
	}
	
	return nil, nil
}

// ===== 解析器字段方法 =====

func (r *organizationResolver) Code() string { return r.org.Code }
func (r *organizationResolver) Name() string { return r.org.Name }
func (r *organizationResolver) UnitType() string { return r.org.UnitType }
func (r *organizationResolver) Status() string { return r.org.Status }
func (r *organizationResolver) Level() int32 { return r.org.Level }
func (r *organizationResolver) Path() *string { return r.org.Path }
func (r *organizationResolver) SortOrder() *int32 { return r.org.SortOrder }
func (r *organizationResolver) Description() *string { return r.org.Description }
func (r *organizationResolver) Profile() *string { return r.org.Profile }
func (r *organizationResolver) ParentCode() *string { return r.org.ParentCode }
func (r *organizationResolver) CreatedAt() string { return r.org.CreatedAt }
func (r *organizationResolver) UpdatedAt() string { return r.org.UpdatedAt }

// 时态字段解析器
func (r *organizationResolver) EffectiveDate() *string { return r.org.EffectiveDate }
func (r *organizationResolver) EndDate() *string { return r.org.EndDate }
func (r *organizationResolver) Version() *int32 { return r.org.Version }
func (r *organizationResolver) SupersedesVersion() *int32 { return r.org.SupersedesVersion }
func (r *organizationResolver) ChangeReason() *string { return r.org.ChangeReason }
func (r *organizationResolver) IsCurrent() *bool { return r.org.IsCurrent }

// 统计解析器方法
func (r *organizationStatsResolver) TotalCount() int32 { return r.stats.TotalCount }
func (r *organizationStatsResolver) ActiveCount() int32 { return r.stats.ActiveCount }
func (r *organizationStatsResolver) InactiveCount() int32 { return r.stats.InactiveCount }
func (r *organizationStatsResolver) PlannedCount() int32 { return r.stats.PlannedCount }
func (r *organizationStatsResolver) CurrentVersionsCount() int32 { return r.stats.CurrentVersionsCount }
func (r *organizationStatsResolver) HistoricalVersionsCount() int32 { return r.stats.HistoricalVersionsCount }
func (r *organizationStatsResolver) DissolvedCount() int32 { return r.stats.DissolvedCount }

func (r *organizationStatsResolver) ByType() []*typeStatResolver {
	var resolvers []*typeStatResolver
	for _, stat := range r.stats.ByType {
		resolvers = append(resolvers, &typeStatResolver{stat: &stat})
	}
	return resolvers
}

func (r *organizationStatsResolver) ByLevel() []*levelStatResolver {
	var resolvers []*levelStatResolver
	for _, stat := range r.stats.ByLevel {
		resolvers = append(resolvers, &levelStatResolver{stat: &stat})
	}
	return resolvers
}

// TypeStat resolver methods
func (r *typeStatResolver) Type() string { return r.stat.Type }
func (r *typeStatResolver) Count() int32 { return r.stat.Count }

// LevelStat resolver methods  
func (r *levelStatResolver) Level() int32 { return r.stat.Level }
func (r *levelStatResolver) Count() int32 { return r.stat.Count }

// 历史解析器方法
func (r *organizationHistoryResolver) Organization() *organizationResolver {
	return &organizationResolver{org: r.history.Organization}
}

func (r *organizationHistoryResolver) ChangeEvents() []*changeEventResolver {
	var resolvers []*changeEventResolver
	for _, event := range r.history.ChangeEvents {
		resolvers = append(resolvers, &changeEventResolver{event: &event})
	}
	return resolvers
}

func (r *organizationHistoryResolver) VersionTimeline() []*versionInfoResolver {
	var resolvers []*versionInfoResolver
	for _, version := range r.history.VersionTimeline {
		resolvers = append(resolvers, &versionInfoResolver{version: &version})
	}
	return resolvers
}

// ChangeEvent resolver methods
func (r *changeEventResolver) EventId() string { return r.event.EventID }
func (r *changeEventResolver) EventType() string { return r.event.EventType }
func (r *changeEventResolver) EffectiveDate() string { return r.event.EffectiveDate }
func (r *changeEventResolver) EndDate() *string { return r.event.EndDate }
func (r *changeEventResolver) ChangeData() string { return r.event.ChangeData }
func (r *changeEventResolver) ChangeReason() *string { return r.event.ChangeReason }
func (r *changeEventResolver) CreatedBy() *string { return r.event.CreatedBy }
func (r *changeEventResolver) CreatedAt() string { return r.event.CreatedAt }

// VersionInfo resolver methods
func (r *versionInfoResolver) Version() int32 { return r.version.Version }
func (r *versionInfoResolver) EffectiveDate() string { return r.version.EffectiveDate }
func (r *versionInfoResolver) EndDate() *string { return r.version.EndDate }
func (r *versionInfoResolver) ChangeReason() *string { return r.version.ChangeReason }
func (r *versionInfoResolver) IsCurrent() bool { return r.version.IsCurrent }

// ===== 缓存相关功能 =====

func generateCacheKey(prefix, action, tenantId string) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", prefix, action, tenantId)))
	return fmt.Sprintf("%s:%x", prefix, hash)
}

func (r *Resolver) getFromCache(ctx context.Context, key string) interface{} {
	val, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil
	}
	
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil
	}
	
	return result
}

func (r *Resolver) setCache(ctx context.Context, key string, value interface{}, expiration time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	
	r.redisClient.Set(ctx, key, data, expiration)
}

// ===== 主程序 =====

func main() {
	// Neo4j连接
	neo4jURI := os.Getenv("NEO4J_URI")
	if neo4jURI == "" {
		neo4jURI = "bolt://localhost:7687"
	}
	
	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		log.Fatal("连接Neo4j失败:", err)
	}
	defer driver.Close(context.Background())
	
	// 测试Neo4j连接
	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		log.Fatal("Neo4j连接测试失败:", err)
	}
	log.Println("✅ Neo4j连接成功")
	
	// Redis连接
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	
	// 测试Redis连接
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("连接Redis失败:", err)
	}
	log.Println("✅ Redis连接成功")
	
	// 创建GraphQL解析器
	resolver := &Resolver{
		neo4jDriver: driver,
		redisClient: redisClient,
	}
	
	// 解析GraphQL Schema
	schema := graphql.MustParseSchema(schemaString, resolver)
	
	// 设置路由
	r := chi.NewRouter()
	
	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	
	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"service": "organization-temporal-query-service",
			"timestamp": time.Now().Format(time.RFC3339),
			"features": []string{"temporal-graphql", "historical-queries", "version-management"},
		})
	})
	
	// 监控指标
	r.Handle("/metrics", promhttp.Handler())
	
	// GraphQL端点
	r.Handle("/graphql", &relay.Handler{Schema: schema})
	
	// GraphiQL界面
	r.Handle("/graphiql", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
	<title>时态组织架构 GraphQL</title>
	<link href="https://unpkg.com/graphiql/graphiql.min.css" rel="stylesheet" />
</head>
<body style="margin: 0;">
	<div id="graphiql" style="height: 100vh;"></div>
	<script crossorigin src="https://unpkg.com/react/umd/react.production.min.js"></script>
	<script crossorigin src="https://unpkg.com/react-dom/umd/react-dom.production.min.js"></script>
	<script crossorigin src="https://unpkg.com/graphiql/graphiql.min.js"></script>
	<script>
		const graphQLFetcher = graphQLParams =>
			fetch('/graphql', {
				method: 'post',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(graphQLParams),
			}).then(response => response.json());
		
		ReactDOM.render(
			React.createElement(GraphiQL, {
				fetcher: graphQLFetcher,
				defaultQuery: '# 时态组织架构查询示例\\n# 查询当前版本\\nquery {\\n  organizations {\\n    code\\n    name\\n    version\\n    effectiveDate\\n    isCurrent\\n  }\\n}\\n\\n# 时间点查询\\nquery {\\n  organizationsAsOf(date: "2025-01-01") {\\n    code\\n    name\\n    effectiveDate\\n    endDate\\n  }\\n}\\n\\n# 时态查询\\nquery {\\n  organizationTemporal(\\n    code: "1000001"\\n    query: {\\n      asOfDate: "2025-06-01"\\n      includeHistory: true\\n      maxVersions: 5\\n    }\\n  ) {\\n    code\\n    name\\n    version\\n    effectiveDate\\n    changeReason\\n  }\\n}'
			}),
			document.getElementById('graphiql'),
		);
	</script>
</body>
</html>
		`))
	}))
	
	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	
	// 优雅关闭
	go func() {
		log.Printf("🚀 时态组织查询服务启动在端口 %s", port)
		log.Println("📋 支持的功能:")
		log.Println("  - 时态GraphQL查询")
		log.Println("  - 历史版本查询")
		log.Println("  - 时间点查询 (as_of_date)")
		log.Println("  - 版本管理查询")
		log.Printf("🌐 GraphiQL界面: http://localhost:%s/graphiql", port)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务器启动失败:", err)
		}
	}()
	
	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("正在关闭服务器...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}
	
	log.Println("服务器已关闭")
}