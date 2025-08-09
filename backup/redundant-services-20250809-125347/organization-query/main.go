package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"
    "github.com/google/uuid"
    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// 项目默认租户配置
const (
    DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
    DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// 城堡查询结构体标准格式 - 完全符合指南标准
type GetOrganizationUnitsQuery struct {
    // 租户隔离 - 城堡多租户核心
    TenantID    uuid.UUID `json:"tenant_id" validate:"required"`
    
    // 过滤条件
    Filters     *OrganizationFilters   `json:"filters,omitempty"`
    
    // 分页控制 - 城堡性能保障
    Pagination  PaginationParams       `json:"pagination" validate:"required"`
    
    // 排序控制
    SortBy      []SortField            `json:"sort_by,omitempty"`
    
    // 审计字段 - 城堡治理要求
    RequestedBy uuid.UUID              `json:"requested_by" validate:"required"`
    RequestID   uuid.UUID              `json:"request_id" validate:"required"`
}

type OrganizationFilters struct {
    UnitType     *string   `json:"unit_type,omitempty"`
    Status       *string   `json:"status,omitempty"`
    ParentCode   *string   `json:"parent_code,omitempty"`
    Codes        []string  `json:"codes,omitempty"`
}

type PaginationParams struct {
    Page     int `json:"page" validate:"min=1"`
    PageSize int `json:"page_size" validate:"min=1,max=100"`
}

type SortField struct {
    Field     string `json:"field"`
    Direction string `json:"direction"` // ASC, DESC
}

// 城堡查询响应模型
type OrganizationUnitView struct {
    Code         string                 `json:"code"`
    Name         string                 `json:"name"`
    UnitType     string                 `json:"unit_type"`
    Status       string                 `json:"status"`
    Level        int                    `json:"level"`
    Path         string                 `json:"path"`
    SortOrder    int                    `json:"sort_order"`
    Description  string                 `json:"description"`
    Profile      map[string]interface{} `json:"profile"`
    ParentCode   *string                `json:"parent_code,omitempty"`
    Children     []OrganizationUnitView `json:"children,omitempty"`
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
}

type OrganizationUnitsResponse struct {
    Organizations []OrganizationUnitView `json:"organizations"`
    TotalCount    int64                  `json:"total_count"`
    Page          int                    `json:"page"`
    PageSize      int                    `json:"page_size"`
    HasNext       bool                   `json:"has_next"`
}

// 城堡Neo4j查询仓储 - 统一实现标准
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

// 城堡查询处理器 - 统一实现标准
type OrganizationQueryHandler struct {
    // 仓储依赖
    repo *Neo4jOrganizationQueryRepository
    
    // 城堡基础设施
    logger *log.Logger
}

func NewOrganizationQueryHandler(repo *Neo4jOrganizationQueryRepository, logger *log.Logger) *OrganizationQueryHandler {
    return &OrganizationQueryHandler{
        repo:   repo,
        logger: logger,
    }
}

// 城堡查询处理标准流程
func (h *OrganizationQueryHandler) HandleGetOrganizationUnits(ctx context.Context, query GetOrganizationUnitsQuery) (*OrganizationUnitsResponse, error) {
    h.logger.Printf("处理组织单元查询请求 - 租户: %s, 请求ID: %s", query.TenantID, query.RequestID)
    
    // 第一阶段：输入验证
    if query.Pagination.PageSize <= 0 {
        query.Pagination.PageSize = 20
    }
    if query.Pagination.Page <= 0 {
        query.Pagination.Page = 1
    }
    
    // 第二阶段：数据库查询
    organizations, totalCount, err := h.repo.GetOrganizationUnits(ctx, query)
    if err != nil {
        h.logger.Printf("查询组织单元失败: %v", err)
        return nil, fmt.Errorf("查询失败: %w", err)
    }
    
    // 第三阶段：响应构建
    response := &OrganizationUnitsResponse{
        Organizations: organizations,
        TotalCount:    totalCount,
        Page:         query.Pagination.Page,
        PageSize:     len(organizations),
        HasNext:      int64(query.Pagination.Page * query.Pagination.PageSize) < totalCount,
    }
    
    h.logger.Printf("查询成功返回 %d 个组织单元", len(organizations))
    return response, nil
}

// Neo4j查询实现
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
            TotalCount   int64
        }{organizations, totalCount}, nil
    })
    
    if err != nil {
        return nil, 0, err
    }
    
    data := result.(struct {
        Organizations []OrganizationUnitView
        TotalCount   int64
    })
    
    return data.Organizations, data.TotalCount, nil
}

type CypherQuerySet struct {
    CountQuery string
    DataQuery  string
    Parameters map[string]interface{}
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

// 城堡查询端测试函数
func TestOrganizationQueryHandler() {
    // Neo4j连接
    driver, err := neo4j.NewDriver(
        "bolt://localhost:7687",
        neo4j.BasicAuth("neo4j", "password", ""))
    if err != nil {
        log.Fatalf("创建Neo4j驱动失败: %v", err)
    }
    defer driver.Close()
    
    // 创建查询组件
    repo := NewNeo4jOrganizationQueryRepository(driver)
    logger := log.New(log.Writer(), "[ORG-QUERY] ", log.LstdFlags)
    handler := NewOrganizationQueryHandler(repo, logger)
    
    // 测试查询
    testTenantID := DefaultTenantID // 使用统一的默认租户ID
    query := GetOrganizationUnitsQuery{
        TenantID: testTenantID,
        Pagination: PaginationParams{
            Page:     1,
            PageSize: 10,
        },
        RequestedBy: uuid.New(),
        RequestID:   uuid.New(),
    }
    
    // 执行查询
    ctx := context.Background()
    response, err := handler.HandleGetOrganizationUnits(ctx, query)
    if err != nil {
        log.Fatalf("查询失败: %v", err)
    }
    
    // 输出结果
    fmt.Printf("🎯 CQRS查询端测试结果:\n")
    fmt.Printf("总数: %d\n", response.TotalCount)
    fmt.Printf("当前页: %d/%d\n", response.Page, response.PageSize)
    fmt.Printf("组织单元:\n")
    
    for _, org := range response.Organizations {
        fmt.Printf("  - %s: %s (%s) [级别:%d]\n", org.Code, org.Name, org.UnitType, org.Level)
    }
}

func main() {
    fmt.Println("🚀 城堡CQRS查询端组件 - 组织架构模块")
    fmt.Println("严格按照CQRS统一实施指南标准实现")
    
    TestOrganizationQueryHandler()
}