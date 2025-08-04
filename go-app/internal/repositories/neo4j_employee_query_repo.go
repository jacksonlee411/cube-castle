package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
	"github.com/gaogu/cube-castle/go-app/internal/service"
)

// Neo4jEmployeeQueryRepository Neo4j员工查询仓储实现
// 实现Neo4jQueryRepository接口
type Neo4jEmployeeQueryRepository struct {
	neo4jService *service.Neo4jService
	logger       Logger
}

// 确保实现了接口
var _ Neo4jQueryRepository = (*Neo4jEmployeeQueryRepository)(nil)

// NewNeo4jEmployeeQueryRepository 创建Neo4j员工查询仓储
func NewNeo4jEmployeeQueryRepository(neo4jService *service.Neo4jService, logger Logger) *Neo4jEmployeeQueryRepository {
	return &Neo4jEmployeeQueryRepository{
		neo4jService: neo4jService,
		logger:       logger,
	}
}

// GetEmployee 获取单个员工信息
func (r *Neo4jEmployeeQueryRepository) GetEmployee(ctx context.Context, query queries.FindEmployeeQuery) (*EmployeeNode, error) {
	// 调用Neo4j服务获取员工
	r.logger.Info("DEBUG REPO: Getting employee", "employee_id", query.ID.String(), "tenant_id", query.TenantID)
	serviceEmployee, err := r.neo4jService.GetEmployee(ctx, query.ID.String())
	if err != nil {
		r.logger.Error("Failed to get employee from Neo4j", "employee_id", query.ID, "tenant_id", query.TenantID, "error", err)
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	// 转换为仓储层的EmployeeNode结构
	employee := &EmployeeNode{
		ID:               query.ID,
		TenantID:         query.TenantID,
		FirstName:        extractFirstName(serviceEmployee.LegalName),
		LastName:         extractLastName(serviceEmployee.LegalName),
		Email:            serviceEmployee.Email,
		EmployeeType:     getStringProperty(serviceEmployee.Properties, "employee_type", "FULL_TIME"),
		EmploymentStatus: serviceEmployee.Status,
		HireDate:         serviceEmployee.HireDate,
		TerminationDate:  nil, // TODO: 从properties中提取
		PersonalInfo:     serviceEmployee.Properties,
		CreatedAt:        time.Now(), // TODO: 从Neo4j获取实际时间
		UpdatedAt:        time.Now(),
	}

	r.logger.Info("Employee retrieved successfully", "employee_id", query.ID, "tenant_id", query.TenantID)

	return employee, nil
}

// SearchEmployees 搜索员工
func (r *Neo4jEmployeeQueryRepository) SearchEmployees(ctx context.Context, query queries.SearchEmployeesQuery) (*EmployeeSearchResponse, error) {
	// 构建搜索过滤器
	filters := make(map[string]interface{})
	
	if query.Name != nil {
		filters["name"] = *query.Name
	}
	if query.Email != nil {
		filters["email"] = *query.Email
	}
	if query.Department != nil {
		filters["department"] = *query.Department
	}

	// 检查Neo4j服务是否可用，如果不可用则返回模拟数据
	if r.neo4jService == nil {
		r.logger.Info("Neo4j service not available, returning mock employee data", "tenant_id", query.TenantID)
		return r.getMockEmployeeSearchResult(query)
	}

	// 调用Neo4j服务搜索员工
	serviceEmployees, total, err := r.neo4jService.SearchEmployees(ctx, filters, query.Limit, query.Offset)
	if err != nil {
		r.logger.Error("Failed to search employees in Neo4j", "tenant_id", query.TenantID, "filters", filters, "limit", query.Limit, "offset", query.Offset, "error", err)
		return nil, fmt.Errorf("failed to search employees: %w", err)
	}

	// 转换服务层结果为仓储层结构
	var employees []EmployeeNode
	for _, serviceEmp := range serviceEmployees {
		employee := EmployeeNode{
			ID:               parseUUID(serviceEmp.ID), // Use ID field which contains UUID
			TenantID:         query.TenantID,
			FirstName:        extractFirstName(serviceEmp.LegalName),
			LastName:         extractLastName(serviceEmp.LegalName),
			Email:            serviceEmp.Email,
			EmployeeType:     getStringProperty(serviceEmp.Properties, "employee_type", "FULL_TIME"),
			EmploymentStatus: serviceEmp.Status,
			HireDate:         serviceEmp.HireDate,
			PersonalInfo:     serviceEmp.Properties,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		employees = append(employees, employee)
	}

	response := &EmployeeSearchResponse{
		Employees:  employees,
		TotalCount: total,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}

	r.logger.Info("Employee search completed", "tenant_id", query.TenantID, "total_found", total, "returned", len(employees))

	return response, nil
}

// GetEmployeeStats 获取员工统计信息
func (r *Neo4jEmployeeQueryRepository) GetEmployeeStats(ctx context.Context, query queries.GetEmployeeStatsQuery) (*queries.EmployeeStatsResponse, error) {
	r.logger.Info("Getting employee stats", "tenant_id", query.TenantID)

	// 暂时返回模拟数据，直到Neo4j连接配置完成
	stats := &queries.EmployeeStatsResponse{
		Total:       42,
		Active:      38,
		Inactive:    4,
		NewThisWeek: 3,
	}

	r.logger.Info("Employee stats retrieved (mock data)", "tenant_id", query.TenantID, "stats", stats)

	return stats, nil
}

// 实现Neo4jQueryRepository接口的其他方法 - 目前暂不实现，返回适当错误

// GetOrgChart 获取组织架构图 (暂不实现)
func (r *Neo4jEmployeeQueryRepository) GetOrgChart(ctx context.Context, query queries.GetOrgChartQuery) (*OrgChartResponse, error) {
	return nil, fmt.Errorf("GetOrgChart not implemented in employee query repository")
}

// GetOrganizationUnit 获取组织单元 (暂不实现) 
func (r *Neo4jEmployeeQueryRepository) GetOrganizationUnit(ctx context.Context, query queries.GetOrganizationUnitQuery) (*OrganizationUnitNode, error) {
	return nil, fmt.Errorf("GetOrganizationUnit not implemented in employee query repository")
}

// getMockEmployeeSearchResult 返回模拟员工搜索结果
func (r *Neo4jEmployeeQueryRepository) getMockEmployeeSearchResult(query queries.SearchEmployeesQuery) (*EmployeeSearchResponse, error) {
	// 生成模拟员工数据
	mockEmployees := []EmployeeNode{
		{
			ID:               parseUUID("emp-001"),
			TenantID:         query.TenantID,
			FirstName:        "三",
			LastName:         "张",
			Email:            "zhang.san@company.com",
			EmployeeType:     "FULL_TIME",
			EmploymentStatus: "ACTIVE",
			HireDate:         time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:               parseUUID("emp-002"),
			TenantID:         query.TenantID,
			FirstName:        "五",
			LastName:         "王",
			Email:            "wang.wu@company.com",
			EmployeeType:     "FULL_TIME",
			EmploymentStatus: "ACTIVE",
			HireDate:         time.Date(2021, 6, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:               parseUUID("emp-003"),
			TenantID:         query.TenantID,
			FirstName:        "七",
			LastName:         "钱",
			Email:            "qian.qi@company.com",
			EmployeeType:     "FULL_TIME",
			EmploymentStatus: "ACTIVE",
			HireDate:         time.Date(2023, 3, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	// 应用过滤器
	var filteredEmployees []EmployeeNode
	for _, emp := range mockEmployees {
		// 构建全名用于搜索
		fullName := emp.FirstName + " " + emp.LastName
		
		if query.Name != nil && !contains(fullName, *query.Name) && 
		   !contains(emp.FirstName, *query.Name) && !contains(emp.LastName, *query.Name) {
			continue
		}
		if query.Email != nil && !contains(emp.Email, *query.Email) {
			continue
		}
		// 暂时跳过Department过滤，因为当前结构中没有Department字段
		// if query.Department != nil && !contains(emp.Department, *query.Department) {
		//	continue
		// }
		if query.Status != nil && !contains(emp.EmploymentStatus, *query.Status) {
			continue
		}
		filteredEmployees = append(filteredEmployees, emp)
	}

	// 应用分页
	total := int(len(filteredEmployees))
	start := int(query.Offset)
	end := start + int(query.Limit)
	
	if start > len(filteredEmployees) {
		filteredEmployees = []EmployeeNode{}
	} else if end > len(filteredEmployees) {
		filteredEmployees = filteredEmployees[start:]
	} else {
		filteredEmployees = filteredEmployees[start:end]
	}

	return &EmployeeSearchResponse{
		Employees:  filteredEmployees,
		TotalCount: total,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}, nil
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

// ListOrganizationUnits 列出组织单元 (暂不实现)
func (r *Neo4jEmployeeQueryRepository) ListOrganizationUnits(ctx context.Context, query queries.ListOrganizationUnitsQuery) (*OrganizationUnitsResponse, error) {
	return nil, fmt.Errorf("ListOrganizationUnits not implemented in employee query repository")
}

// GetReportingHierarchy 获取汇报层级 (可以实现员工相关的)
func (r *Neo4jEmployeeQueryRepository) GetReportingHierarchy(ctx context.Context, query queries.GetReportingHierarchyQuery) (*ReportingHierarchyResponse, error) {
	// TODO: 实现基于Neo4j服务的汇报层级查询
	return nil, fmt.Errorf("GetReportingHierarchy not yet implemented")
}

// FindEmployeePath 查找员工路径 (可以实现)
func (r *Neo4jEmployeeQueryRepository) FindEmployeePath(ctx context.Context, query queries.FindEmployeePathQuery) ([]EmployeeNode, error) {
	// TODO: 实现基于Neo4j服务的员工路径查询
	return nil, fmt.Errorf("FindEmployeePath not yet implemented")
}

// GetDepartmentStructure 获取部门结构 (暂不实现)
func (r *Neo4jEmployeeQueryRepository) GetDepartmentStructure(ctx context.Context, query queries.GetDepartmentStructureQuery) (*OrgChartResponse, error) {
	return nil, fmt.Errorf("GetDepartmentStructure not implemented in employee query repository")
}

// FindCommonManager 查找共同经理 (可以实现)
func (r *Neo4jEmployeeQueryRepository) FindCommonManager(ctx context.Context, query queries.FindCommonManagerQuery) (*EmployeeNode, error) {
	// TODO: 实现基于Neo4j服务的共同经理查询
	return nil, fmt.Errorf("FindCommonManager not yet implemented")
}

// 辅助函数：从全名中提取名字
func extractFirstName(legalName string) string {
	if legalName == "" {
		return ""
	}
	parts := strings.Split(legalName, " ")
	return parts[0]
}

// 辅助函数：从全名中提取姓氏
func extractLastName(legalName string) string {
	if legalName == "" {
		return ""
	}
	parts := strings.Split(legalName, " ")
	if len(parts) > 1 {
		return strings.Join(parts[1:], " ")
	}
	return ""
}

// 辅助函数：从属性中获取字符串值
func getStringProperty(props map[string]interface{}, key, defaultValue string) string {
	if val, ok := props[key].(string); ok {
		return val
	}
	return defaultValue
}

// 辅助函数：解析UUID
func parseUUID(idStr string) uuid.UUID {
	if id, err := uuid.Parse(idStr); err == nil {
		return id
	}
	return uuid.Nil
}