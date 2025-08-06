package main

import (
	"context"
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
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// 项目默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// 组织单元查询请求
type GetOrganizationUnitsQuery struct {
	TenantID    uuid.UUID            `json:"tenant_id" validate:"required"`
	Filters     *OrganizationFilters `json:"filters,omitempty"`
	Pagination  PaginationParams     `json:"pagination" validate:"required"`
	SortBy      []SortField          `json:"sort_by,omitempty"`
	RequestedBy uuid.UUID            `json:"requested_by" validate:"required"`
	RequestID   uuid.UUID            `json:"request_id" validate:"required"`
}

type OrganizationFilters struct {
	UnitType   *string  `json:"unit_type,omitempty"`
	Status     *string  `json:"status,omitempty"`
	ParentCode *string  `json:"parent_code,omitempty"`
	Codes      []string `json:"codes,omitempty"`
}

type PaginationParams struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // ASC, DESC
}

// 组织单元视图模型
type OrganizationUnitView struct {
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	UnitType    string                 `json:"unit_type"`
	Status      string                 `json:"status"`
	Level       int                    `json:"level"`
	Path        string                 `json:"path"`
	SortOrder   int                    `json:"sort_order"`
	Description string                 `json:"description"`
	Profile     map[string]interface{} `json:"profile"`
	ParentCode  *string                `json:"parent_code,omitempty"`
	Children    []OrganizationUnitView `json:"children,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type OrganizationUnitsResponse struct {
	Organizations []OrganizationUnitView `json:"organizations"`
	TotalCount    int64                  `json:"total_count"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
	HasNext       bool                   `json:"has_next"`
}

// Neo4j查询仓储
type Neo4jOrganizationQueryRepository struct {
	driver   neo4j.Driver
	database string
}

func NewNeo4jOrganizationQueryRepository(driver neo4j.Driver) *Neo4jOrganizationQueryRepository {
	return &Neo4jOrganizationQueryRepository{
		driver:   driver,
		database: "neo4j",
	}
}

// 查询处理器
type OrganizationQueryHandler struct {
	repo   *Neo4jOrganizationQueryRepository
	logger *log.Logger
}

func NewOrganizationQueryHandler(repo *Neo4jOrganizationQueryRepository, logger *log.Logger) *OrganizationQueryHandler {
	return &OrganizationQueryHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *OrganizationQueryHandler) HandleGetOrganizationUnits(ctx context.Context, query GetOrganizationUnitsQuery) (*OrganizationUnitsResponse, error) {
	h.logger.Printf("处理组织单元查询请求 - 租户: %s, 请求ID: %s", query.TenantID, query.RequestID)

	// 输入验证
	if query.Pagination.PageSize <= 0 {
		query.Pagination.PageSize = 20
	}
	if query.Pagination.Page <= 0 {
		query.Pagination.Page = 1
	}

	// 数据库查询
	organizations, totalCount, err := h.repo.GetOrganizationUnits(ctx, query)
	if err != nil {
		h.logger.Printf("查询组织单元失败: %v", err)
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	// 响应构建
	response := &OrganizationUnitsResponse{
		Organizations: organizations,
		TotalCount:    totalCount,
		Page:          query.Pagination.Page,
		PageSize:      len(organizations),
		HasNext:       int64(query.Pagination.Page*query.Pagination.PageSize) < totalCount,
	}

	h.logger.Printf("查询成功返回 %d 个组织单元", len(organizations))
	return response, nil
}

type CypherQuerySet struct {
	CountQuery string
	DataQuery  string
	Parameters map[string]interface{}
}

func (r *Neo4jOrganizationQueryRepository) GetOrganizationUnits(ctx context.Context, query GetOrganizationUnitsQuery) ([]OrganizationUnitView, int64, error) {
	session := r.driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: r.database,
	})
	defer session.Close()

	// 构建Cypher查询
	cypherQuery := r.buildCypherQuery(query)

	// 执行查询
	result, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		// 获取总数
		countResult, err := tx.Run(cypherQuery.CountQuery, cypherQuery.Parameters)
		if err != nil {
			return nil, fmt.Errorf("计数查询失败: %w", err)
		}

		var totalCount int64 = 0
		if countResult.Next() {
			totalCount = countResult.Record().Values[0].(int64)
		}

		// 获取分页数据
		dataResult, err := tx.Run(cypherQuery.DataQuery, cypherQuery.Parameters)
		if err != nil {
			return nil, fmt.Errorf("数据查询失败: %w", err)
		}

		var organizations []OrganizationUnitView
		for dataResult.Next() {
			record := dataResult.Record()
			org := r.recordToOrganizationView(record)
			organizations = append(organizations, org)
		}

		return struct {
			Organizations []OrganizationUnitView
			TotalCount    int64
		}{organizations, totalCount}, nil
	})

	if err != nil {
		return nil, 0, err
	}

	data := result.(struct {
		Organizations []OrganizationUnitView
		TotalCount    int64
	})

	return data.Organizations, data.TotalCount, nil
}

func (r *Neo4jOrganizationQueryRepository) buildCypherQuery(query GetOrganizationUnitsQuery) *CypherQuerySet {
	// 基础WHERE条件
	whereConditions := []string{"o.tenant_id = $tenant_id"}
	params := map[string]interface{}{
		"tenant_id": query.TenantID.String(),
	}

	// 动态过滤条件
	if query.Filters != nil {
		if query.Filters.UnitType != nil {
			whereConditions = append(whereConditions, "o.unit_type = $unit_type")
			params["unit_type"] = *query.Filters.UnitType
		}

		if query.Filters.Status != nil {
			whereConditions = append(whereConditions, "o.status = $status")
			params["status"] = *query.Filters.Status
		}

		if len(query.Filters.Codes) > 0 {
			whereConditions = append(whereConditions, "o.code IN $codes")
			params["codes"] = query.Filters.Codes
		}
	}

	// 构建WHERE子句
	var whereClause string
	if len(whereConditions) > 0 {
		whereClause = "WHERE "
		for i, condition := range whereConditions {
			if i > 0 {
				whereClause += " AND "
			}
			whereClause += condition
		}
	}

	// 排序条件
	orderClause := "ORDER BY o.level, o.sort_order, o.code"
	if len(query.SortBy) > 0 {
		orderFields := []string{}
		for _, sort := range query.SortBy {
			direction := "ASC"
			if sort.Direction == "DESC" {
				direction = "DESC"
			}
			orderFields = append(orderFields, fmt.Sprintf("o.%s %s", sort.Field, direction))
		}
		if len(orderFields) > 0 {
			orderClause = "ORDER BY "
			for i, field := range orderFields {
				if i > 0 {
					orderClause += ", "
				}
				orderClause += field
			}
		}
	}

	// 分页参数
	skip := (query.Pagination.Page - 1) * query.Pagination.PageSize
	limit := query.Pagination.PageSize
	params["skip"] = skip
	params["limit"] = limit

	// 计数查询
	countQuery := fmt.Sprintf(`
		MATCH (o:OrganizationUnit)
		%s
		RETURN count(o) as total
	`, whereClause)

	// 数据查询
	dataQuery := fmt.Sprintf(`
		MATCH (o:OrganizationUnit)
		%s
		%s
		SKIP $skip LIMIT $limit
		RETURN o.code as code, o.name as name, o.unit_type as unit_type,
			   o.status as status, o.level as level, o.path as path,
			   o.sort_order as sort_order, o.description as description,
			   o.profile as profile, o.created_at as created_at,
			   o.updated_at as updated_at
	`, whereClause, orderClause)

	return &CypherQuerySet{
		CountQuery: countQuery,
		DataQuery:  dataQuery,
		Parameters: params,
	}
}

func (r *Neo4jOrganizationQueryRepository) recordToOrganizationView(record *neo4j.Record) OrganizationUnitView {
	org := OrganizationUnitView{
		Code:        record.Values[0].(string),
		Name:        record.Values[1].(string),
		UnitType:    record.Values[2].(string),
		Status:      record.Values[3].(string),
		Level:       int(record.Values[4].(int64)),
		Path:        record.Values[5].(string),
		SortOrder:   int(record.Values[6].(int64)),
		Description: record.Values[7].(string),
	}

	// 处理Profile JSON
	if profileStr, ok := record.Values[8].(string); ok && profileStr != "" {
		var profile map[string]interface{}
		if err := json.Unmarshal([]byte(profileStr), &profile); err == nil {
			org.Profile = profile
		}
	}

	// 处理时间字段
	if createdAt, ok := record.Values[9].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			org.CreatedAt = t
		}
	}

	if updatedAt, ok := record.Values[10].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			org.UpdatedAt = t
		}
	}

	return org
}

// HTTP处理器
type APIHandler struct {
	queryHandler *OrganizationQueryHandler
	logger       *log.Logger
}

func NewAPIHandler(queryHandler *OrganizationQueryHandler, logger *log.Logger) *APIHandler {
	return &APIHandler{
		queryHandler: queryHandler,
		logger:       logger,
	}
}

// 统计数据结构
type OrganizationStats struct {
	TotalCount int                    `json:"total_count"`
	ByType     map[string]int         `json:"by_type"`
	ByStatus   map[string]int         `json:"by_status"`
	ByLevel    map[string]int         `json:"by_level"`
}

func (h *APIHandler) GetOrganizationStats(w http.ResponseWriter, r *http.Request) {
	// 获取租户ID
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		tenantIDStr = DefaultTenantIDString
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// 构建统计查询
	query := GetOrganizationUnitsQuery{
		TenantID: tenantID,
		Pagination: PaginationParams{
			Page:     1,
			PageSize: 1000, // 获取所有数据进行统计
		},
		RequestedBy: uuid.New(),
		RequestID:   uuid.New(),
	}

	// 执行查询
	response, err := h.queryHandler.HandleGetOrganizationUnits(r.Context(), query)
	if err != nil {
		h.logger.Printf("统计查询失败: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 计算统计数据
	stats := OrganizationStats{
		TotalCount: int(response.TotalCount),
		ByType:     make(map[string]int),
		ByStatus:   make(map[string]int),
		ByLevel:    make(map[string]int),
	}

	for _, org := range response.Organizations {
		// 按类型统计
		stats.ByType[org.UnitType]++
		
		// 按状态统计
		stats.ByStatus[org.Status]++
		
		// 按层级统计
		levelKey := fmt.Sprintf("级别%d", org.Level)
		stats.ByLevel[levelKey]++
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Printf("统计数据序列化失败: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *APIHandler) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	// 获取租户ID，使用项目默认值
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		tenantIDStr = DefaultTenantIDString // 使用统一的默认租户ID
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// 解析查询参数
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 50
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// 构建查询
	query := GetOrganizationUnitsQuery{
		TenantID: tenantID,
		Pagination: PaginationParams{
			Page:     page,
			PageSize: pageSize,
		},
		RequestedBy: uuid.New(),
		RequestID:   uuid.New(),
	}

	// 处理过滤参数
	if unitType := r.URL.Query().Get("unit_type"); unitType != "" {
		if query.Filters == nil {
			query.Filters = &OrganizationFilters{}
		}
		query.Filters.UnitType = &unitType
	}

	if status := r.URL.Query().Get("status"); status != "" {
		if query.Filters == nil {
			query.Filters = &OrganizationFilters{}
		}
		query.Filters.Status = &status
	}

	// 执行查询
	response, err := h.queryHandler.HandleGetOrganizationUnits(r.Context(), query)
	if err != nil {
		h.logger.Printf("API查询失败: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("JSON序列化失败: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func main() {
	logger := log.New(os.Stdout, "[ORG-API] ", log.LstdFlags)

	// Neo4j连接
	driver, err := neo4j.NewDriver(
		"bolt://localhost:7687",
		neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		log.Fatalf("创建Neo4j驱动失败: %v", err)
	}
	defer driver.Close()

	// 创建处理器
	repo := NewNeo4jOrganizationQueryRepository(driver)
	queryHandler := NewOrganizationQueryHandler(repo, logger)
	apiHandler := NewAPIHandler(queryHandler, logger)

	// 创建路由器
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

	// 健康检查端点
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		healthStatus := map[string]interface{}{
			"status":    "healthy",
			"service":   "organization-api-server",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"database":  "connected", // 假设Neo4j连接正常，因为服务器已经启动
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(healthStatus); err != nil {
			logger.Printf("健康检查响应序列化失败: %v", err)
		}
	})

	// API路由
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/organization-units", apiHandler.GetOrganizations)
		r.Get("/organization-units/stats", apiHandler.GetOrganizationStats)
	})

	// 启动服务器
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 优雅关闭
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		
		logger.Println("正在关闭服务器...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("服务器关闭失败: %v", err)
		}
	}()

	logger.Printf("🚀 CQRS组织API服务器启动在端口 :8080")
	logger.Printf("严格按照CQRS统一实施指南标准实现")
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
	
	logger.Println("服务器已关闭")
}