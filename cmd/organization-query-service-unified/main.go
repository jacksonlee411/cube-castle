package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"cube-castle-deployment-test/pkg/health"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// 默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// GraphQL Schema定义
var schemaString = `
	type Organization {
		tenant_id: String!
		code: String!
		parent_code: String
		name: String!
		unit_type: String!
		status: String!
		level: Int!
		path: String
		sort_order: Int
		description: String
		profile: String
		created_at: String!
		updated_at: String!
		effective_date: String!
		end_date: String
		version: Int!
		is_current: Boolean!
		# 时态管理扩展字段
		change_reason: String
		valid_from: String!
		valid_to: String!
	}

	type Query {
		# 传统查询 (当前数据) - 保持兼容性
		organizations(first: Int, offset: Int, searchText: String): [Organization!]!
		organization(code: String!): Organization
		organizationStats: OrganizationStats!
		
		# 时态查询 - Neo4j最佳实践
		organizationAsOfDate(code: String!, asOfDate: String!): Organization
		organizationHistory(code: String!, fromDate: String!, toDate: String!): [Organization!]!
	}

	type OrganizationStats {
		totalCount: Int!
		byType: [TypeCount!]!
		byStatus: [StatusCount!]!
		byLevel: [LevelCount!]!
	}

	type TypeCount {
		unitType: String!
		count: Int!
	}

	type StatusCount {
		status: String!
		count: Int!
	}

	type LevelCount {
		level: String!
		count: Int!
	}
`

// GraphQL组织模型 - 匹配时态API格式
type Organization struct {
	TenantIdField      string `json:"tenant_id"`
	CodeField          string `json:"code"`
	ParentCodeField    string `json:"parent_code"`
	NameField          string `json:"name"`
	UnitTypeField      string `json:"unit_type"`
	StatusField        string `json:"status"`
	LevelField         int    `json:"level"`
	PathField          string `json:"path"`
	SortOrderField     int    `json:"sort_order"`
	DescriptionField   string `json:"description"`
	ProfileField       string `json:"profile"`
	CreatedAtField     string `json:"created_at"`
	UpdatedAtField     string `json:"updated_at"`
	EffectiveDateField string `json:"effective_date"`
	EndDateField       string `json:"end_date"`
	VersionField       int    `json:"version"`
	IsCurrentField     bool   `json:"is_current"`
	// 时态管理扩展字段
	ChangeReasonField string `json:"change_reason"`
	ValidFromField    string `json:"valid_from"`
	ValidToField      string `json:"valid_to"`
}

// GraphQL字段解析器 - 匹配时态API Schema字段名
func (o Organization) Tenant_id() string { return o.TenantIdField }
func (o Organization) Code() string      { return o.CodeField }
func (o Organization) Parent_code() *string {
	if o.ParentCodeField == "" {
		return nil
	}
	return &o.ParentCodeField
}
func (o Organization) Name() string      { return o.NameField }
func (o Organization) Unit_type() string { return o.UnitTypeField }
func (o Organization) Status() string    { return o.StatusField }
func (o Organization) Level() int32      { return int32(o.LevelField) }
func (o Organization) Path() *string {
	if o.PathField == "" {
		return nil
	}
	return &o.PathField
}
func (o Organization) Sort_order() *int32 {
	if o.SortOrderField == 0 {
		return nil
	}
	val := int32(o.SortOrderField)
	return &val
}
func (o Organization) Description() *string {
	if o.DescriptionField == "" {
		return nil
	}
	return &o.DescriptionField
}
func (o Organization) Profile() *string {
	if o.ProfileField == "" {
		return nil
	}
	return &o.ProfileField
}
func (o Organization) Created_at() string     { return o.CreatedAtField }
func (o Organization) Updated_at() string     { return o.UpdatedAtField }
func (o Organization) Effective_date() string { return o.EffectiveDateField }
func (o Organization) End_date() *string {
	if o.EndDateField == "" {
		return nil
	}
	return &o.EndDateField
}
func (o Organization) Version() int32   { return int32(o.VersionField) }
func (o Organization) Is_current() bool { return o.IsCurrentField }

// 时态管理字段解析器
func (o Organization) Change_reason() *string {
	if o.ChangeReasonField == "" {
		return nil
	}
	return &o.ChangeReasonField
}
func (o Organization) Valid_from() string { return o.ValidFromField }
func (o Organization) Valid_to() string   { return o.ValidToField }

// GraphQL统计模型
type OrganizationStats struct {
	TotalCountField int           `json:"total_count"`
	ByTypeField     []TypeCount   `json:"by_type"`
	ByStatusField   []StatusCount `json:"by_status"`
	ByLevelField    []LevelCount  `json:"by_level"`
}

func (s OrganizationStats) TotalCount() int32       { return int32(s.TotalCountField) }
func (s OrganizationStats) ByType() []TypeCount     { return s.ByTypeField }
func (s OrganizationStats) ByStatus() []StatusCount { return s.ByStatusField }
func (s OrganizationStats) ByLevel() []LevelCount   { return s.ByLevelField }

type TypeCount struct {
	TypeField  string `json:"type"`
	CountField int    `json:"count"`
}

func (t TypeCount) UnitType() string { return t.TypeField }
func (t TypeCount) Count() int32     { return int32(t.CountField) }

type StatusCount struct {
	StatusField string `json:"status"`
	CountField  int    `json:"count"`
}

func (s StatusCount) Status() string { return s.StatusField }
func (s StatusCount) Count() int32   { return int32(s.CountField) }

type LevelCount struct {
	LevelField string `json:"level"`
	CountField int    `json:"count"`
}

func (l LevelCount) Level() string { return l.LevelField }
func (l LevelCount) Count() int32  { return int32(l.CountField) }

// Neo4j仓储（带Redis缓存）
type Neo4jOrganizationRepository struct {
	driver      neo4j.DriverWithContext
	redisClient *redis.Client
	logger      *log.Logger
	cacheTTL    time.Duration
}

func NewNeo4jOrganizationRepository(driver neo4j.DriverWithContext, redisClient *redis.Client, logger *log.Logger) *Neo4jOrganizationRepository {
	return &Neo4jOrganizationRepository{
		driver:      driver,
		redisClient: redisClient,
		logger:      logger,
		cacheTTL:    5 * time.Minute, // 5分钟缓存
	}
}

// 生成缓存键
func (r *Neo4jOrganizationRepository) getCacheKey(operation string, params ...interface{}) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("org:%s:%v", operation, params)))
	return fmt.Sprintf("cache:%x", h.Sum(nil))
}

func (r *Neo4jOrganizationRepository) GetOrganizations(ctx context.Context, tenantID uuid.UUID, first, offset int, searchText string) ([]Organization, error) {
	// 生成缓存键 (包含搜索文本)
	cacheKey := r.getCacheKey("organizations", tenantID.String(), first, offset, searchText)

	// 尝试从缓存获取
	if r.redisClient != nil {
		cachedData, err := r.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var organizations []Organization
			if json.Unmarshal([]byte(cachedData), &organizations) == nil {
				r.logger.Printf("[Cache HIT] 从缓存返回组织列表 - 键: %s, 数量: %d", cacheKey, len(organizations))
				return organizations, nil
			}
		}
		r.logger.Printf("[Cache MISS] 缓存未命中，查询数据库 - 键: %s", cacheKey)
	}

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// 构建搜索条件
	searchCondition := ""
	params := map[string]interface{}{
		"tenant_id": tenantID.String(),
		"first":     int64(first),
		"offset":    int64(offset),
	}

	if searchText != "" {
		searchCondition = "AND (o.name CONTAINS $searchText OR o.code CONTAINS $searchText)"
		params["searchText"] = searchText
	}

	query := fmt.Sprintf(`
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		WHERE o.is_current = true %s
		RETURN o.tenant_id as tenant_id, o.code as code, o.parent_code as parent_code,
		       o.name as name, o.unit_type as unit_type, o.status as status, 
		       o.level as level, o.path as path, o.sort_order as sort_order,
		       o.description as description, o.profile as profile,
		       o.created_at as created_at, o.updated_at as updated_at,
		       toString(o.effective_date) as effective_date, toString(o.end_date) as end_date,
		       o.version as version, o.is_current as is_current
		ORDER BY o.sort_order, o.code
		SKIP $offset LIMIT $first
	`, searchCondition)

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	var organizations []Organization
	for result.Next(ctx) {
		record := result.Record()

		org := Organization{
			TenantIdField:      getStringValue(record, "tenant_id"),
			CodeField:          getStringValue(record, "code"),
			ParentCodeField:    getStringValue(record, "parent_code"),
			NameField:          getStringValue(record, "name"),
			UnitTypeField:      getStringValue(record, "unit_type"),
			StatusField:        getStringValue(record, "status"),
			LevelField:         getIntValue(record, "level"),
			PathField:          getStringValue(record, "path"),
			SortOrderField:     getIntValue(record, "sort_order"),
			DescriptionField:   getStringValue(record, "description"),
			ProfileField:       getStringValue(record, "profile"),
			CreatedAtField:     getStringValue(record, "created_at"),
			UpdatedAtField:     getStringValue(record, "updated_at"),
			EffectiveDateField: getStringValue(record, "effective_date"),
			EndDateField:       getStringValue(record, "end_date"),
			VersionField:       getIntValue(record, "version"),
			IsCurrentField:     getBoolValue(record, "is_current"),
		}
		organizations = append(organizations, org)
	}

	// 将结果写入缓存
	if r.redisClient != nil && len(organizations) > 0 {
		if cacheData, err := json.Marshal(organizations); err == nil {
			r.redisClient.Set(ctx, cacheKey, string(cacheData), r.cacheTTL)
			r.logger.Printf("[Cache SET] 缓存已更新 - 键: %s, 数量: %d, TTL: %v", cacheKey, len(organizations), r.cacheTTL)
		}
	}

	return organizations, result.Err()
}

func (r *Neo4jOrganizationRepository) GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*Organization, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	query := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id, code: $code})
		RETURN o.tenant_id as tenant_id, o.code as code, o.parent_code as parent_code,
		       o.name as name, o.unit_type as unit_type, o.status as status, 
		       o.level as level, o.path as path, o.sort_order as sort_order,
		       o.description as description, o.profile as profile,
		       o.created_at as created_at, o.updated_at as updated_at,
		       toString(o.effective_date) as effective_date, toString(o.end_date) as end_date,
		       o.version as version, o.is_current as is_current
		ORDER BY o.is_current DESC, o.effective_date DESC
		LIMIT 1
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"tenant_id": tenantID.String(),
		"code":      code,
	})
	if err != nil {
		return nil, err
	}

	if result.Next(ctx) {
		record := result.Record()
		org := &Organization{
			TenantIdField:      getStringValue(record, "tenant_id"),
			CodeField:          getStringValue(record, "code"),
			ParentCodeField:    getStringValue(record, "parent_code"),
			NameField:          getStringValue(record, "name"),
			UnitTypeField:      getStringValue(record, "unit_type"),
			StatusField:        getStringValue(record, "status"),
			LevelField:         getIntValue(record, "level"),
			PathField:          getStringValue(record, "path"),
			SortOrderField:     getIntValue(record, "sort_order"),
			DescriptionField:   getStringValue(record, "description"),
			ProfileField:       getStringValue(record, "profile"),
			CreatedAtField:     getStringValue(record, "created_at"),
			UpdatedAtField:     getStringValue(record, "updated_at"),
			EffectiveDateField: getStringValue(record, "effective_date"),
			EndDateField:       getStringValue(record, "end_date"),
			VersionField:       getIntValue(record, "version"),
			IsCurrentField:     getBoolValue(record, "is_current"),
		}
		return org, nil
	}

	return nil, nil
}

// 时态数据记录转换方法 - 支持完整时态字段
func (r *Neo4jOrganizationRepository) recordToOrganization(record *neo4j.Record) Organization {
	return Organization{
		TenantIdField:      getStringValue(record, "tenant_id"),
		CodeField:          getStringValue(record, "code"),
		ParentCodeField:    getStringValue(record, "parent_code"),
		NameField:          getStringValue(record, "name"),
		UnitTypeField:      getStringValue(record, "unit_type"),
		StatusField:        getStringValue(record, "status"),
		LevelField:         getIntValue(record, "level"),
		PathField:          getStringValue(record, "path"),
		SortOrderField:     getIntValue(record, "sort_order"),
		DescriptionField:   getStringValue(record, "description"),
		ProfileField:       getStringValue(record, "profile"),
		CreatedAtField:     getStringValue(record, "created_at"),
		UpdatedAtField:     getStringValue(record, "updated_at"),
		EffectiveDateField: getStringValue(record, "effective_date"),
		EndDateField:       getStringValue(record, "end_date"),
		VersionField:       getIntValue(record, "version"),
		IsCurrentField:     getBoolValue(record, "is_current"),
		// 时态管理扩展字段
		ChangeReasonField: getStringValue(record, "change_reason"),
		ValidFromField:    getStringValue(record, "valid_from"),
		ValidToField:      getStringValue(record, "valid_to"),
	}
}

func (r *Neo4jOrganizationRepository) GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*OrganizationStats, error) {
	// 生成缓存键
	cacheKey := r.getCacheKey("stats", tenantID.String())

	// 尝试从缓存获取
	if r.redisClient != nil {
		cachedData, err := r.redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var stats OrganizationStats
			if json.Unmarshal([]byte(cachedData), &stats) == nil {
				r.logger.Printf("[Cache HIT] 从缓存返回统计信息 - 键: %s", cacheKey)
				return &stats, nil
			}
		}
		r.logger.Printf("[Cache MISS] 缓存未命中，查询数据库 - 键: %s", cacheKey)
	}

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// 获取总数
	totalQuery := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		RETURN count(o) as total
	`

	totalResult, err := session.Run(ctx, totalQuery, map[string]interface{}{
		"tenant_id": tenantID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("查询总数失败: %w", err)
	}

	var total int
	if totalResult.Next(ctx) {
		record := totalResult.Record()
		total = int(record.Values[0].(int64))
	}

	// 按类型统计
	typeQuery := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		RETURN o.unit_type as type, count(o) as count
		ORDER BY type
	`

	typeResult, err := session.Run(ctx, typeQuery, map[string]interface{}{
		"tenant_id": tenantID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("按类型统计失败: %w", err)
	}

	var byType []TypeCount
	for typeResult.Next(ctx) {
		record := typeResult.Record()
		unitType := getStringValue(record, "type")
		count := getIntValue(record, "count")
		byType = append(byType, TypeCount{
			TypeField:  unitType,
			CountField: count,
		})
	}

	// 按状态统计
	statusQuery := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		RETURN o.status as status, count(o) as count
		ORDER BY status
	`

	statusResult, err := session.Run(ctx, statusQuery, map[string]interface{}{
		"tenant_id": tenantID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("按状态统计失败: %w", err)
	}

	var byStatus []StatusCount
	for statusResult.Next(ctx) {
		record := statusResult.Record()
		status := getStringValue(record, "status")
		count := getIntValue(record, "count")
		byStatus = append(byStatus, StatusCount{
			StatusField: status,
			CountField:  count,
		})
	}

	// 按级别统计
	levelQuery := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		RETURN toString(o.level) as level, count(o) as count
		ORDER BY level
	`

	levelResult, err := session.Run(ctx, levelQuery, map[string]interface{}{
		"tenant_id": tenantID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("按级别统计失败: %w", err)
	}

	var byLevel []LevelCount
	for levelResult.Next(ctx) {
		record := levelResult.Record()
		level := getStringValue(record, "level")
		count := getIntValue(record, "count")
		byLevel = append(byLevel, LevelCount{
			LevelField: fmt.Sprintf("级别%s", level),
			CountField: count,
		})
	}

	// 构建统计结果
	stats := &OrganizationStats{
		TotalCountField: total,
		ByTypeField:     byType,
		ByStatusField:   byStatus,
		ByLevelField:    byLevel,
	}

	// 将结果写入缓存
	if r.redisClient != nil {
		if cacheData, err := json.Marshal(stats); err == nil {
			r.redisClient.Set(ctx, cacheKey, string(cacheData), r.cacheTTL)
			r.logger.Printf("[Cache SET] 统计缓存已更新 - 键: %s, TTL: %v", cacheKey, r.cacheTTL)
		}
	}

	r.logger.Printf("[Stats] 统计查询完成 - 总数: %d, 类型数: %d, 状态数: %d, 级别数: %d",
		total, len(byType), len(byStatus), len(byLevel))

	return stats, nil
}

// Helper functions
func getStringValue(record *neo4j.Record, key string) string {
	if value, ok := record.Get(key); ok && value != nil {
		if str, ok := value.(string); ok {
			return str
		}
		// 处理time.Time类型
		if t, ok := value.(time.Time); ok {
			return t.Format("2006-01-02") // 返回 YYYY-MM-DD 格式
		}

		// 对于其他类型，直接转换为字符串
		if str := fmt.Sprintf("%v", value); str != "<nil>" && str != "" {
			// 如果字符串看起来像日期，尝试解析
			if t, err := time.Parse("2006-01-02", str); err == nil {
				return t.Format("2006-01-02")
			}
			// 如果包含时间信息，尝试解析并只取日期部分
			if t, err := time.Parse("2006-01-02T15:04:05Z", str); err == nil {
				return t.Format("2006-01-02")
			}
			// 返回原始字符串
			return str
		}
	}
	return ""
}

func getIntValue(record *neo4j.Record, key string) int {
	if value, ok := record.Get(key); ok && value != nil {
		if i64, ok := value.(int64); ok {
			return int(i64)
		}
	}
	return 0
}

func getBoolValue(record *neo4j.Record, key string) bool {
	if value, ok := record.Get(key); ok && value != nil {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return true // 默认为当前版本
}

// GraphQL Resolver
type Resolver struct {
	repo   *Neo4jOrganizationRepository
	logger *log.Logger
}

// === 时态查询解析器 - Neo4j最佳实践 ===

// 按时间点查询组织 (as_of_date)
func (r *Resolver) OrganizationAsOfDate(ctx context.Context, args struct {
	Code     string
	AsOfDate string
}) (*Organization, error) {
	tenantID := DefaultTenantID

	r.logger.Printf("[GraphQL] 时态查询 as_of_date - 租户: %s, 代码: %s, 时间点: %s", tenantID, args.Code, args.AsOfDate)

	// 生成缓存键
	cacheKey := r.repo.getCacheKey("temporal_as_of", tenantID.String(), args.Code, args.AsOfDate)

	// 检查缓存
	if r.repo.redisClient != nil {
		if cachedData, err := r.repo.redisClient.Get(ctx, cacheKey).Result(); err == nil {
			var org Organization
			if json.Unmarshal([]byte(cachedData), &org) == nil {
				r.logger.Printf("[Cache HIT] 时态查询缓存命中 - 键: %s", cacheKey)
				return &org, nil
			}
		}
		r.logger.Printf("[Cache MISS] 时态查询缓存未命中 - 键: %s", cacheKey)
	}

	session := r.repo.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Neo4j时态查询 - 使用date()函数进行正确的日期比较
	query := `
		MATCH (org:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		WHERE org.effective_date <= date($as_of_date)
		  AND (org.end_date IS NULL OR org.end_date >= date($as_of_date))
		ORDER BY org.effective_date DESC, COALESCE(org.version, 1) DESC
		LIMIT 1
		RETURN org.tenant_id as tenant_id, org.code as code, org.parent_code as parent_code,
		       org.name as name, org.unit_type as unit_type, org.status as status,
		       org.level as level, org.path as path, org.sort_order as sort_order,
		       org.description as description, toString(org.effective_date) as effective_date,
		       toString(org.end_date) as end_date, org.is_current as is_current,
		       org.change_reason as change_reason, org.version as version,
		       org.valid_from as valid_from, org.valid_to as valid_to
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"code":       args.Code,
			"tenant_id":  tenantID.String(),
			"as_of_date": args.AsOfDate,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			org := r.repo.recordToOrganization(record)
			return org, nil
		}
		return nil, nil
	})

	if err != nil {
		r.logger.Printf("[GraphQL] 时态查询失败: %v", err)
		return nil, err
	}

	if result != nil {
		org := result.(Organization)
		// 缓存历史数据1小时
		if r.repo.redisClient != nil {
			if data, err := json.Marshal(org); err == nil {
				r.repo.redisClient.Set(ctx, cacheKey, data, time.Hour)
				r.logger.Printf("[Cache SET] 时态查询结果已缓存 - 键: %s", cacheKey)
			}
		}

		r.logger.Printf("[GraphQL] 时态查询成功 - 组织: %s", org.Name)
		return &org, nil
	}

	r.logger.Printf("[GraphQL] 时态查询无结果 - 代码: %s, 时间点: %s", args.Code, args.AsOfDate)
	return nil, nil
}

// 查询组织历史记录 (时间范围)
func (r *Resolver) OrganizationHistory(ctx context.Context, args struct {
	Code     string
	FromDate string
	ToDate   string
}) ([]Organization, error) {
	tenantID := DefaultTenantID

	r.logger.Printf("[GraphQL] 时态历史查询 - 租户: %s, 代码: %s, 时间范围: %s~%s", tenantID, args.Code, args.FromDate, args.ToDate)

	session := r.repo.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Neo4j时态范围查询 - 使用date()函数进行正确的日期比较
	query := `
		MATCH (org:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		WHERE org.effective_date >= date($from_date)
		  AND org.effective_date <= date($to_date)
		ORDER BY org.effective_date DESC, COALESCE(org.version, 1) DESC
		RETURN org.tenant_id as tenant_id, org.code as code, org.parent_code as parent_code,
		       org.name as name, org.unit_type as unit_type, org.status as status,
		       org.level as level, org.path as path, org.sort_order as sort_order,
		       org.description as description, toString(org.effective_date) as effective_date,
		       toString(org.end_date) as end_date, org.is_current as is_current,
		       org.change_reason as change_reason, org.version as version,
		       org.valid_from as valid_from, org.valid_to as valid_to
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"code":      args.Code,
			"tenant_id": tenantID.String(),
			"from_date": args.FromDate,
			"to_date":   args.ToDate,
		})
		if err != nil {
			return nil, err
		}

		var organizations []Organization
		for result.Next(ctx) {
			record := result.Record()
			org := r.repo.recordToOrganization(record)
			organizations = append(organizations, org)
		}
		return organizations, nil
	})

	if err != nil {
		r.logger.Printf("[GraphQL] 时态历史查询失败: %v", err)
		return nil, err
	}

	organizations := result.([]Organization)
	r.logger.Printf("[GraphQL] 时态历史查询成功 - 返回 %d 条记录", len(organizations))
	return organizations, nil
}

// === 传统查询解析器 (保持兼容) ===

func (r *Resolver) Organizations(ctx context.Context, args struct {
	First      *int32
	Offset     *int32
	SearchText *string
}) ([]Organization, error) {
	first := 50
	offset := 0
	searchText := ""

	if args.First != nil {
		first = int(*args.First)
	}
	if args.Offset != nil {
		offset = int(*args.Offset)
	}
	if args.SearchText != nil {
		searchText = *args.SearchText
	}

	tenantID := DefaultTenantID // 暂时使用默认租户

	r.logger.Printf("[GraphQL] 查询组织列表 - 租户: %s, first: %d, offset: %d, searchText: %s", tenantID, first, offset, searchText)

	organizations, err := r.repo.GetOrganizations(ctx, tenantID, first, offset, searchText)
	if err != nil {
		r.logger.Printf("[GraphQL] 查询组织列表失败: %v", err)
		return nil, err
	}

	r.logger.Printf("[GraphQL] 查询组织列表成功 - 返回 %d 个组织", len(organizations))
	return organizations, nil
}

func (r *Resolver) Organization(ctx context.Context, args struct {
	Code string
}) (*Organization, error) {
	tenantID := DefaultTenantID

	r.logger.Printf("[GraphQL] 查询单个组织 - 租户: %s, 代码: %s", tenantID, args.Code)

	org, err := r.repo.GetOrganization(ctx, tenantID, args.Code)
	if err != nil {
		r.logger.Printf("[GraphQL] 查询单个组织失败: %v", err)
		return nil, err
	}

	if org != nil {
		r.logger.Printf("[GraphQL] 查询单个组织成功 - 组织: %s", org.NameField)
	} else {
		r.logger.Printf("[GraphQL] 组织不存在 - 代码: %s", args.Code)
	}

	return org, nil
}

func (r *Resolver) OrganizationStats(ctx context.Context) (*OrganizationStats, error) {
	tenantID := DefaultTenantID

	r.logger.Printf("[GraphQL] 查询组织统计 - 租户: %s", tenantID)

	stats, err := r.repo.GetOrganizationStats(ctx, tenantID)
	if err != nil {
		r.logger.Printf("[GraphQL] 查询组织统计失败: %v", err)
		return nil, err
	}

	r.logger.Printf("[GraphQL] 查询组织统计成功 - 总数: %d", stats.TotalCountField)
	return stats, nil
}

func main() {
	logger := log.New(os.Stdout, "[GraphQL-ORG] ", log.LstdFlags)

	// Neo4j连接
	neo4jURI := os.Getenv("NEO4J_URI")
	if neo4jURI == "" {
		neo4jURI = "bolt://localhost:7687"
	}

	neo4jUser := os.Getenv("NEO4J_USER")
	if neo4jUser == "" {
		neo4jUser = "neo4j"
	}

	neo4jPassword := os.Getenv("NEO4J_PASSWORD")
	if neo4jPassword == "" {
		neo4jPassword = "password"
	}

	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		log.Fatalf("Neo4j驱动创建失败: %v", err)
	}
	defer driver.Close(context.Background())

	// 测试连接
	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		log.Fatalf("Neo4j连接失败: %v", err)
	}
	logger.Println("Neo4j连接成功")

	// Redis连接
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// 测试Redis连接
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		logger.Printf("Redis连接失败，将不使用缓存: %v", err)
		redisClient = nil
	} else {
		logger.Println("Redis连接成功，缓存功能已启用")
	}

	// 创建仓储和解析器
	repo := NewNeo4jOrganizationRepository(driver, redisClient, logger)
	resolver := &Resolver{repo: repo, logger: logger}

	// 创建GraphQL schema
	schema := graphql.MustParseSchema(schemaString, resolver)

	// 创建HTTP路由
	r := chi.NewRouter()

	// 中间件
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

	// REST API 端点 - 统一查询协议
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/organization-units", func(w http.ResponseWriter, r *http.Request) {
			// 将REST查询转换为GraphQL查询
			first := int32(50)
			offset := int32(0)

			if firstStr := r.URL.Query().Get("limit"); firstStr != "" {
				if f, err := strconv.ParseInt(firstStr, 10, 32); err == nil {
					first = int32(f)
				}
			}

			if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
				if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
					offset = int32(o)
				}
			}

			ctx := r.Context()
			organizations, err := resolver.Organizations(ctx, struct {
				First      *int32
				Offset     *int32
				SearchText *string
			}{&first, &offset, nil})

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"organizations": organizations,
				"total":         len(organizations),
			})
		})

		r.Get("/organization-units/{code}", func(w http.ResponseWriter, r *http.Request) {
			code := chi.URLParam(r, "code")
			if code == "" {
				http.Error(w, "缺少组织代码", http.StatusBadRequest)
				return
			}

			ctx := r.Context()
			org, err := resolver.Organization(ctx, struct {
				Code string
			}{code})

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if org == nil {
				http.Error(w, "组织不存在", http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(org)
		})
	})

	// GraphQL端点
	r.Handle("/graphql", &relay.Handler{Schema: schema})

	// GraphiQL开发界面
	r.Get("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		graphiqlHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>GraphiQL</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphiql@2.4.7/graphiql.min.css" />
    <style>
        body { height: 100%; margin: 0; width: 100%; overflow: hidden; }
        #graphiql { height: 100vh; }
    </style>
</head>
<body>
    <div id="graphiql">Loading...</div>
    <script crossorigin src="https://unpkg.com/react@18/umd/react.development.js"></script>
    <script crossorigin src="https://unpkg.com/react-dom@18/umd/react-dom.development.js"></script>
    <script crossorigin src="https://cdn.jsdelivr.net/npm/graphiql@2.4.7/graphiql.min.js"></script>
    <script>
        const fetcher = GraphiQL.createFetcher({ url: '/graphql' });
        const root = ReactDOM.createRoot(document.getElementById('graphiql'));
        root.render(React.createElement(GraphiQL, { fetcher }));
    </script>
</body>
</html>`
		w.Write([]byte(graphiqlHTML))
	})

	// 健康检查端点 - 增强版
	healthManager := health.NewHealthManager("organization-graphql-service", "2.0.0")

	// 添加Neo4j健康检查
	healthManager.AddChecker(&health.Neo4jChecker{
		Name:   "neo4j",
		Driver: driver,
	})

	// 添加Redis健康检查 - 暂时禁用由于版本兼容性问题
	// healthManager.AddChecker(&health.RedisChecker{
	//	Name:   "redis",
	//	Client: redisClient,
	// })

	// 创建告警管理器
	alertManager := health.NewAlertManager("organization-graphql-service")

	// 添加告警规则
	alertManager.AddRule(health.AlertRule{
		Name:       "neo4j-unhealthy",
		Component:  "neo4j",
		Condition:  health.AlertCondition{StatusEquals: func() *health.HealthStatus { s := health.StatusUnhealthy; return &s }()},
		Level:      health.AlertLevelCritical,
		Message:    "Neo4j数据库连接失败 - %s状态为%s: %s",
		Cooldown:   5 * time.Minute,
		MaxRetries: 3,
		EnabledBy:  time.Now(),
	})

	alertManager.AddRule(health.AlertRule{
		Name:       "redis-unhealthy",
		Component:  "redis",
		Condition:  health.AlertCondition{StatusEquals: func() *health.HealthStatus { s := health.StatusUnhealthy; return &s }()},
		Level:      health.AlertLevelWarning,
		Message:    "Redis缓存服务异常 - %s状态为%s: %s",
		Cooldown:   3 * time.Minute,
		MaxRetries: 2,
		EnabledBy:  time.Now(),
	})

	alertManager.AddRule(health.AlertRule{
		Name:       "slow-response",
		Component:  "", // 适用于所有组件
		Condition:  health.AlertCondition{ResponseTimeGT: func() *time.Duration { d := 5 * time.Second; return &d }()},
		Level:      health.AlertLevelWarning,
		Message:    "响应时间过慢 - %s响应时间%s超过5秒: %s",
		Cooldown:   10 * time.Minute,
		MaxRetries: 1,
		EnabledBy:  time.Now(),
	})

	// 配置告警渠道
	if webhookURL := os.Getenv("ALERT_WEBHOOK_URL"); webhookURL != "" {
		webhookChannel := health.NewWebhookChannel("primary-webhook", webhookURL)
		webhookChannel.AddHeader("Authorization", "Bearer "+os.Getenv("WEBHOOK_TOKEN"))
		alertManager.AddChannel(webhookChannel)
		logger.Println("告警Webhook已配置:", webhookURL)
	}

	if slackWebhook := os.Getenv("SLACK_WEBHOOK_URL"); slackWebhook != "" {
		slackChannel := health.NewSlackChannel(slackWebhook, "#alerts", "Cube Castle Monitor")
		alertManager.AddChannel(slackChannel)
		logger.Println("Slack告警已配置")
	}

	// 启动告警处理协程
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				health := healthManager.Check(ctx)
				alertManager.ProcessHealthCheck(ctx, health)
				cancel()
			case <-context.Background().Done():
				return
			}
		}
	}()

	r.Get("/health", healthManager.Handler())

	// 告警管理端点
	r.Get("/alerts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		alerts := alertManager.GetActiveAlerts()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active_alerts": alerts,
			"total":         len(alerts),
			"timestamp":     time.Now(),
		})
	})

	r.Get("/alerts/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		history := alertManager.GetAlertHistory(50) // 最近50条
		json.NewEncoder(w).Encode(map[string]interface{}{
			"alert_history": history,
			"total":         len(history),
			"timestamp":     time.Now(),
		})
	})

	// 详细状态报告
	statusReporter := health.NewStatusReporter(healthManager, "http://localhost:8090")
	r.Get("/status", statusReporter.DashboardHandler())
	r.Get("/status/dashboard", statusReporter.DashboardHandler())

	// Prometheus指标端点
	r.Handle("/metrics", promhttp.Handler())

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090" // 智能网关期望的GraphQL服务端口
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

		logger.Println("正在关闭GraphQL服务器...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("GraphQL服务器关闭失败: %v", err)
		}
	}()

	logger.Printf("🚀 GraphQL组织服务启动在端口 :%s", port)
	logger.Println("GraphiQL开发界面: http://localhost:" + port + "/graphiql")
	logger.Println("GraphQL端点: http://localhost:" + port + "/graphql")
	logger.Println("告警管理: http://localhost:" + port + "/alerts")
	logger.Println("状态仪表板: http://localhost:" + port + "/status")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("GraphQL服务器启动失败: %v", err)
	}

	logger.Println("GraphQL服务器已关闭")
}
