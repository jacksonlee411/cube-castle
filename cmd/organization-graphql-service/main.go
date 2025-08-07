package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	}

	type Query {
		organizations(first: Int, offset: Int): [Organization!]!
		organization(code: String!): Organization
		organizationStats: OrganizationStats!
	}

	type OrganizationStats {
		totalCount: Int!
		byType: [TypeCount!]!
		byStatus: [StatusCount!]!
		byLevel: [LevelCount!]!
	}

	type TypeCount {
		type: String!
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

// GraphQL组织模型
type Organization struct {
	code        string `json:"code"`
	name        string `json:"name"`
	unitType    string `json:"unitType"`
	status      string `json:"status"`
	level       int    `json:"level"`
	path        string `json:"path"`
	sortOrder   int    `json:"sortOrder"`
	description string `json:"description"`
	profile     string `json:"profile"`
	parentCode  string `json:"parentCode"`
	createdAt   string `json:"createdAt"`
	updatedAt   string `json:"updatedAt"`
}

// GraphQL字段解析方法
func (o Organization) Code() string        { return o.code }
func (o Organization) Name() string        { return o.name }
func (o Organization) UnitType() string    { return o.unitType }
func (o Organization) Status() string      { return o.status }
func (o Organization) Level() int32        { return int32(o.level) }
func (o Organization) Path() *string       { if o.path == "" { return nil }; return &o.path }
func (o Organization) SortOrder() *int32   { if o.sortOrder == 0 { return nil }; s := int32(o.sortOrder); return &s }
func (o Organization) Description() *string { if o.description == "" { return nil }; return &o.description }
func (o Organization) Profile() *string    { if o.profile == "" { return nil }; return &o.profile }
func (o Organization) ParentCode() *string { if o.parentCode == "" { return nil }; return &o.parentCode }
func (o Organization) CreatedAt() string   { return o.createdAt }
func (o Organization) UpdatedAt() string   { return o.updatedAt }

// GraphQL统计模型
type OrganizationStats struct {
	totalCount int          `json:"total_count"`
	byType     []TypeCount  `json:"by_type"`
	byStatus   []StatusCount `json:"by_status"`
	byLevel    []LevelCount  `json:"by_level"`
}

func (s OrganizationStats) TotalCount() int32     { return int32(s.totalCount) }
func (s OrganizationStats) ByType() []TypeCount   { return s.byType }
func (s OrganizationStats) ByStatus() []StatusCount { return s.byStatus }
func (s OrganizationStats) ByLevel() []LevelCount { return s.byLevel }

type TypeCount struct {
	typeVal string `json:"type"`
	count   int    `json:"count"`
}

func (t TypeCount) Type() string  { return t.typeVal }
func (t TypeCount) Count() int32  { return int32(t.count) }

type StatusCount struct {
	status string `json:"status"`
	count  int    `json:"count"`
}

func (s StatusCount) Status() string { return s.status }
func (s StatusCount) Count() int32   { return int32(s.count) }

type LevelCount struct {
	level string `json:"level"`
	count int    `json:"count"`
}

func (l LevelCount) Level() string { return l.level }
func (l LevelCount) Count() int32  { return int32(l.count) }

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

func (r *Neo4jOrganizationRepository) GetOrganizations(ctx context.Context, tenantID uuid.UUID, first, offset int) ([]Organization, error) {
	// 生成缓存键
	cacheKey := r.getCacheKey("organizations", tenantID.String(), first, offset)
	
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

	query := `
		MATCH (o:OrganizationUnit {tenant_id: $tenant_id})
		RETURN o.code as code, o.name as name, o.unit_type as unit_type, 
		       o.status as status, o.level as level, o.path as path,
		       o.sort_order as sort_order, o.description as description,
		       o.profile as profile, o.parent_code as parent_code,
		       o.created_at as created_at, o.updated_at as updated_at
		ORDER BY o.sort_order, o.code
		SKIP $offset LIMIT $first
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"tenant_id": tenantID.String(),
		"first":     int64(first),
		"offset":    int64(offset),
	})
	if err != nil {
		return nil, err
	}

	var organizations []Organization
	for result.Next(ctx) {
		record := result.Record()
		
		org := Organization{
			code:        getStringValue(record, "code"),
			name:        getStringValue(record, "name"),
			unitType:    getStringValue(record, "unit_type"),
			status:      getStringValue(record, "status"),
			level:       getIntValue(record, "level"),
			path:        getStringValue(record, "path"),
			sortOrder:   getIntValue(record, "sort_order"),
			description: getStringValue(record, "description"),
			profile:     getStringValue(record, "profile"),
			parentCode:  getStringValue(record, "parent_code"),
			createdAt:   getStringValue(record, "created_at"),
			updatedAt:   getStringValue(record, "updated_at"),
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
		RETURN o.code as code, o.name as name, o.unit_type as unit_type, 
		       o.status as status, o.level as level, o.path as path,
		       o.sort_order as sort_order, o.description as description,
		       o.profile as profile, o.parent_code as parent_code,
		       o.created_at as created_at, o.updated_at as updated_at
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
			code:        getStringValue(record, "code"),
			name:        getStringValue(record, "name"),
			unitType:    getStringValue(record, "unit_type"),
			status:      getStringValue(record, "status"),
			level:       getIntValue(record, "level"),
			path:        getStringValue(record, "path"),
			sortOrder:   getIntValue(record, "sort_order"),
			description: getStringValue(record, "description"),
			profile:     getStringValue(record, "profile"),
			parentCode:  getStringValue(record, "parent_code"),
			createdAt:   getStringValue(record, "created_at"),
			updatedAt:   getStringValue(record, "updated_at"),
		}
		return org, nil
	}

	return nil, nil
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
			typeVal: unitType,
			count:   count,
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
			status: status,
			count:  count,
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
			level: fmt.Sprintf("级别%s", level),
			count: count,
		})
	}

	// 构建统计结果
	stats := &OrganizationStats{
		totalCount: total,
		byType:     byType,
		byStatus:   byStatus,
		byLevel:    byLevel,
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

// GraphQL Resolver
type Resolver struct {
	repo   *Neo4jOrganizationRepository
	logger *log.Logger
}

func (r *Resolver) Organizations(ctx context.Context, args struct{
	First  *int32
	Offset *int32
}) ([]Organization, error) {
	first := 50
	offset := 0
	
	if args.First != nil {
		first = int(*args.First)
	}
	if args.Offset != nil {
		offset = int(*args.Offset)
	}

	tenantID := DefaultTenantID // 暂时使用默认租户
	
	r.logger.Printf("[GraphQL] 查询组织列表 - 租户: %s, first: %d, offset: %d", tenantID, first, offset)
	
	organizations, err := r.repo.GetOrganizations(ctx, tenantID, first, offset)
	if err != nil {
		r.logger.Printf("[GraphQL] 查询组织列表失败: %v", err)
		return nil, err
	}
	
	r.logger.Printf("[GraphQL] 查询组织列表成功 - 返回 %d 个组织", len(organizations))
	return organizations, nil
}

func (r *Resolver) Organization(ctx context.Context, args struct{
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
		r.logger.Printf("[GraphQL] 查询单个组织成功 - 组织: %s", org.Name())
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
	
	r.logger.Printf("[GraphQL] 查询组织统计成功 - 总数: %d", stats.TotalCount())
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

	// 健康检查端点
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"service": "organization-graphql-service",
			"status":  "healthy",
		})
	})

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"  // 智能网关期望的GraphQL服务端口
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

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("GraphQL服务器启动失败: %v", err)
	}

	logger.Println("GraphQL服务器已关闭")
}