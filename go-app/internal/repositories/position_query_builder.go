package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
)

// ErrNotFound 未找到记录错误
var ErrNotFound = errors.New("record not found")

// CypherQueryBuilder Cypher查询构建器
type CypherQueryBuilder struct {
	matches    []string
	wheres     []string
	returns    []string
	orderBys   []string
	params     map[string]interface{}
	limit      *int
	skip       *int
}

// NewCypherQueryBuilder 创建查询构建器
func NewCypherQueryBuilder() *CypherQueryBuilder {
	return &CypherQueryBuilder{
		matches: make([]string, 0),
		wheres:  make([]string, 0),
		returns: make([]string, 0),
		orderBys: make([]string, 0),
		params:  make(map[string]interface{}),
	}
}

// Match 添加MATCH子句
func (b *CypherQueryBuilder) Match(pattern string) *CypherQueryBuilder {
	b.matches = append(b.matches, pattern)
	return b
}

// OptionalMatch 添加OPTIONAL MATCH子句
func (b *CypherQueryBuilder) OptionalMatch(pattern string) *CypherQueryBuilder {
	b.matches = append(b.matches, "OPTIONAL "+pattern)
	return b
}

// Where 添加WHERE条件
func (b *CypherQueryBuilder) Where(condition string) *CypherQueryBuilder {
	b.wheres = append(b.wheres, condition)
	return b
}

// Return 添加RETURN子句
func (b *CypherQueryBuilder) Return(fields string) *CypherQueryBuilder {
	b.returns = append(b.returns, fields)
	return b
}

// OrderBy 添加排序
func (b *CypherQueryBuilder) OrderBy(field string) *CypherQueryBuilder {
	b.orderBys = append(b.orderBys, field)
	return b
}

// SetParam 设置参数
func (b *CypherQueryBuilder) SetParam(key string, value interface{}) *CypherQueryBuilder {
	b.params[key] = value
	return b
}

// Limit 设置限制
func (b *CypherQueryBuilder) Limit(limit int) *CypherQueryBuilder {
	b.limit = &limit
	return b
}

// Skip 设置跳过
func (b *CypherQueryBuilder) Skip(skip int) *CypherQueryBuilder {
	b.skip = &skip
	return b
}

// Build 构建最终查询
func (b *CypherQueryBuilder) Build() (string, map[string]interface{}) {
	var parts []string
	
	// MATCH子句
	for _, match := range b.matches {
		if strings.HasPrefix(match, "OPTIONAL") {
			parts = append(parts, match)
		} else {
			parts = append(parts, "MATCH "+match)
		}
	}
	
	// WHERE子句
	if len(b.wheres) > 0 {
		parts = append(parts, "WHERE "+strings.Join(b.wheres, " AND "))
	}
	
	// RETURN子句
	if len(b.returns) > 0 {
		parts = append(parts, "RETURN "+strings.Join(b.returns, ", "))
	}
	
	// ORDER BY子句
	if len(b.orderBys) > 0 {
		parts = append(parts, "ORDER BY "+strings.Join(b.orderBys, ", "))
	}
	
	// SKIP子句
	if b.skip != nil {
		parts = append(parts, fmt.Sprintf("SKIP %d", *b.skip))
	}
	
	// LIMIT子句
	if b.limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *b.limit))
	}
	
	return strings.Join(parts, "\n"), b.params
}

// PositionQueryTemplates 预定义查询模板
type PositionQueryTemplates struct{}

// GetPositionWithRelationsQuery 获取职位及其关系的查询模板
func (t *PositionQueryTemplates) GetPositionWithRelationsQuery() *CypherQueryBuilder {
	return NewCypherQueryBuilder().
		Match("(p:Position {id: $positionId, tenant_id: $tenantId})").
		OptionalMatch("(p)-[:BELONGS_TO]->(d:Organization)").
		OptionalMatch("(p)-[:REPORTS_TO]->(m:Position)").
		OptionalMatch("(dr:Position)-[:REPORTS_TO]->(p)").
		OptionalMatch("(p)<-[a:ASSIGNED]-(e:Employee)").
		Where("a.is_current = true").
		Return("p, d, m, collect(DISTINCT dr) as directReports, e")
}

// SearchPositionsQuery 搜索职位的查询模板
func (t *PositionQueryTemplates) SearchPositionsQuery(params SearchPositionsParams) *CypherQueryBuilder {
	builder := NewCypherQueryBuilder().
		Match("(p:Position {tenant_id: $tenantId})").
		SetParam("tenantId", params.TenantID.String())
	
	// 动态添加过滤条件
	if params.DepartmentID != nil {
		builder.OptionalMatch("(p)-[:BELONGS_TO]->(d:Organization {id: $departmentId})").
			SetParam("departmentId", params.DepartmentID.String())
	}
	
	if params.Status != nil {
		builder.Where("p.status = $status").
			SetParam("status", *params.Status)
	}
	
	if params.PositionType != nil {
		builder.Where("p.position_type = $positionType").
			SetParam("positionType", *params.PositionType)
	}
	
	return builder.Return("p").
		OrderBy("p.created_at DESC").
		Skip(params.Offset).
		Limit(params.Limit)
}

// ResultParser 结果解析器接口
type ResultParser interface {
	ParsePosition(node neo4j.Node) Position
	ParseEmployee(node neo4j.Node) Employee
	ParseOrganization(node neo4j.Node) Organization
	ParsePositionAssignment(rel neo4j.Relationship) PositionAssignment
}

// Neo4jResultParser Neo4j结果解析器实现
type Neo4jResultParser struct{}

func (p *Neo4jResultParser) ParsePosition(node neo4j.Node) Position {
	props := node.Props
	
	position := Position{
		ID:           uuid.MustParse(props["id"].(string)),
		TenantID:     uuid.MustParse(props["tenant_id"].(string)),
		PositionType: props["position_type"].(string),
		Status:       props["status"].(string),
		BudgetedFTE:  props["budgeted_fte"].(float64),
	}
	
	// 安全地解析可选字段
	if jobProfileID, ok := props["job_profile_id"].(string); ok && jobProfileID != "" {
		position.JobProfileID = uuid.MustParse(jobProfileID)
	}
	
	if deptID, ok := props["department_id"].(string); ok && deptID != "" {
		position.DepartmentID = uuid.MustParse(deptID)
	}
	
	if mgr, ok := props["manager_position_id"].(string); ok && mgr != "" {
		mgrID := uuid.MustParse(mgr)
		position.ManagerPositionID = &mgrID
	}
	
	if details, ok := props["details"].(map[string]interface{}); ok {
		position.Details = details
	}
	
	return position
}

func (p *Neo4jResultParser) ParseEmployee(node neo4j.Node) Employee {
	props := node.Props
	
	return Employee{
		ID:           uuid.MustParse(props["id"].(string)),
		TenantID:     uuid.MustParse(props["tenant_id"].(string)),
		FirstName:    props["first_name"].(string),
		LastName:         props["last_name"].(string),
		Email:            props["email"].(string),
		EmploymentStatus: props["employment_status"].(string),
		EmployeeType:     props["employee_type"].(string),
	}
}

func (p *Neo4jResultParser) ParseOrganization(node neo4j.Node) Organization {
	props := node.Props
	
	return Organization{
		ID:       uuid.MustParse(props["id"].(string)),
		TenantID: uuid.MustParse(props["tenant_id"].(string)),
		UnitType: props["unit_type"].(string),
		Name:     props["name"].(string),
		Status:   props["status"].(string),
	}
}

func (p *Neo4jResultParser) ParsePositionAssignment(rel neo4j.Relationship) PositionAssignment {
	props := rel.Props
	
	assignment := PositionAssignment{
		ID:             uuid.MustParse(props["id"].(string)),
		TenantID:       uuid.MustParse(props["tenant_id"].(string)),
		PositionID:     uuid.MustParse(props["position_id"].(string)),
		EmployeeID:     uuid.MustParse(props["employee_id"].(string)),
		IsCurrent:      props["is_current"].(bool),
		FTE:            props["fte"].(float64),
		AssignmentType: props["assignment_type"].(string),
	}
	
	// Parse start_date
	if startDate, ok := props["start_date"].(time.Time); ok {
		assignment.StartDate = startDate
	}
	
	// Parse optional end_date
	if endDate, ok := props["end_date"]; ok && endDate != nil {
		if endTime, valid := endDate.(time.Time); valid {
			assignment.EndDate = &endTime
		}
	}
	
	// Parse timestamps
	if createdAt, ok := props["created_at"].(time.Time); ok {
		assignment.CreatedAt = createdAt
	}
	if updatedAt, ok := props["updated_at"].(time.Time); ok {
		assignment.UpdatedAt = updatedAt
	}
	
	return assignment
}

// 简化的职位查询仓储实现
type SimplifiedPositionQueryRepository struct {
	driver    neo4j.DriverWithContext
	templates *PositionQueryTemplates
	parser    ResultParser
}

func NewSimplifiedPositionQueryRepository(driver neo4j.DriverWithContext) *SimplifiedPositionQueryRepository {
	return &SimplifiedPositionQueryRepository{
		driver:    driver,
		templates: &PositionQueryTemplates{},
		parser:    &Neo4jResultParser{},
	}
}

// GetPositionWithRelations 获取职位及其关系（简化版）
func (r *SimplifiedPositionQueryRepository) GetPositionWithRelations(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*PositionWithRelations, error) {
	query, params := r.templates.GetPositionWithRelationsQuery().
		SetParam("positionId", id.String()).
		SetParam("tenantId", tenantID.String()).
		Build()
	
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)
	
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		
		if res.Next(ctx) {
			return r.parsePositionWithRelationsResult(res.Record())
		}
		
		return nil, ErrNotFound
	})
	
	if err != nil {
		return nil, err
	}
	
	return result.(*PositionWithRelations), nil
}

func (r *SimplifiedPositionQueryRepository) parsePositionWithRelationsResult(record *neo4j.Record) (*PositionWithRelations, error) {
	result := &PositionWithRelations{}
	
	// 解析主职位
	if posNode, found := record.Get("p"); found && posNode != nil {
		result.Position = r.parser.ParsePosition(posNode.(neo4j.Node))
	}
	
	// 解析部门
	if deptNode, found := record.Get("d"); found && deptNode != nil {
		dept := r.parser.ParseOrganization(deptNode.(neo4j.Node))
		result.Department = &dept
	}
	
	// 解析管理者
	if mgrNode, found := record.Get("m"); found && mgrNode != nil {
		mgr := r.parser.ParsePosition(mgrNode.(neo4j.Node))
		result.Manager = &mgr
	}
	
	// 解析下属职位
	if reportsData, found := record.Get("directReports"); found {
		reports := reportsData.([]interface{})
		for _, reportData := range reports {
			if reportData != nil {
				report := r.parser.ParsePosition(reportData.(neo4j.Node))
				result.DirectReports = append(result.DirectReports, report)
			}
		}
	}
	
	// 解析当前员工
	if empNode, found := record.Get("e"); found && empNode != nil {
		emp := r.parser.ParseEmployee(empNode.(neo4j.Node))
		result.CurrentEmployee = &emp
	}
	
	return result, nil
}

// GetEmployeePositionHistory 获取员工职位历史（简化实现）
func (r *SimplifiedPositionQueryRepository) GetEmployeePositionHistory(ctx context.Context, employeeID uuid.UUID, tenantID uuid.UUID, limit, offset int) ([]PositionOccupancyHistory, int, error) {
	// 简化实现：返回空列表
	return []PositionOccupancyHistory{}, 0, nil
}

// GetPosition 获取职位（简化实现）
func (r *SimplifiedPositionQueryRepository) GetPosition(ctx context.Context, query queries.GetPositionQuery) (*queries.PositionResponse, error) {
	// 简化实现：返回空结果
	return nil, ErrNotFound
}

// SearchPositions 搜索职位（简化实现）
func (r *SimplifiedPositionQueryRepository) SearchPositions(ctx context.Context, params SearchPositionsParams) ([]Position, int, error) {
	// 简化实现：返回空列表
	return []Position{}, 0, nil
}

// GetPositionHierarchy 获取职位层级（简化实现）
func (r *SimplifiedPositionQueryRepository) GetPositionHierarchy(ctx context.Context, query queries.GetPositionHierarchyQuery) ([]queries.PositionNode, error) {
	// 简化实现：返回空列表
	return []queries.PositionNode{}, nil
}

// GetEmployeePositions 获取员工职位（简化实现）
func (r *SimplifiedPositionQueryRepository) GetEmployeePositions(ctx context.Context, query queries.GetEmployeePositionsQuery) ([]queries.OccupancyHistoryResponse, error) {
	// 简化实现：返回空列表
	return []queries.OccupancyHistoryResponse{}, nil
}

// GetPositionEmployees 获取职位员工（简化实现）
func (r *SimplifiedPositionQueryRepository) GetPositionEmployees(ctx context.Context, query queries.GetPositionEmployeesQuery) ([]queries.EmployeeResponse, error) {
	// 简化实现：返回空列表
	return []queries.EmployeeResponse{}, nil
}

// GetPositionOccupancyHistory 获取职位占用历史（简化实现）
func (r *SimplifiedPositionQueryRepository) GetPositionOccupancyHistory(ctx context.Context, params OccupancyHistoryParams) ([]PositionOccupancyHistory, int, error) {
	// 简化实现：返回空列表
	return []PositionOccupancyHistory{}, 0, nil
}

// GetPositionStats 获取职位统计（简化实现）
func (r *SimplifiedPositionQueryRepository) GetPositionStats(ctx context.Context, query queries.GetPositionStatsQuery) (*queries.PositionStatsResponse, error) {
	// 简化实现：返回空结果
	return nil, ErrNotFound
}

// NewNeo4jPositionQueryRepository 创建Neo4j职位查询仓储（别名）
func NewNeo4jPositionQueryRepository(driver neo4j.DriverWithContext) *SimplifiedPositionQueryRepository {
	return NewSimplifiedPositionQueryRepository(driver)
}