package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
	"github.com/gaogu/cube-castle/go-app/internal/service"
	"github.com/gaogu/cube-castle/go-app/internal/events"
)

// Neo4jPositionQueryRepository Neo4j职位查询仓储实现
type Neo4jPositionQueryRepository struct {
	neo4jService *service.Neo4jService
	logger       events.Logger
}

// NewNeo4jPositionQueryRepositoryV2 创建Neo4j职位查询仓储
func NewNeo4jPositionQueryRepositoryV2(neo4jService *service.Neo4jService, logger events.Logger) *Neo4jPositionQueryRepository {
	return &Neo4jPositionQueryRepository{
		neo4jService: neo4jService,
		logger:       logger,
	}
}

// 确保实现了接口
var _ PositionQueryRepository = (*Neo4jPositionQueryRepository)(nil)

// GetPosition 获取单个职位信息
func (r *Neo4jPositionQueryRepository) GetPosition(ctx context.Context, query queries.GetPositionQuery) (*queries.PositionResponse, error) {
	// 避免未使用的import错误
	_ = fmt.Sprintf("Neo4j Position Query Repository")
	r.logger.Info("Getting position from Neo4j", "position_id", query.ID, "tenant_id", query.TenantID)

	// 如果Neo4j服务不可用，返回模拟数据
	if r.neo4jService == nil {
		r.logger.Info("Neo4j service not available, returning mock position data", "position_id", query.ID)
		return r.getMockPositionResponse(query.ID, query.TenantID), nil
	}

	// TODO: 实现真实的Neo4j查询
	// 目前返回模拟数据，直到Neo4j连接配置完成
	position := r.getMockPositionResponse(query.ID, query.TenantID)
	
	r.logger.Info("Position retrieved successfully", "position_id", query.ID, "tenant_id", query.TenantID)
	
	return position, nil
}

// GetPositionWithRelations 获取职位及其关联信息
func (r *Neo4jPositionQueryRepository) GetPositionWithRelations(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*PositionWithRelations, error) {
	r.logger.Info("Getting position with relations from Neo4j", "position_id", id, "tenant_id", tenantID)

	// 返回模拟数据
	return r.getMockPositionWithRelations(id, tenantID), nil
}

// SearchPositions 搜索职位
func (r *Neo4jPositionQueryRepository) SearchPositions(ctx context.Context, params SearchPositionsParams) ([]Position, int, error) {
	r.logger.Info("Searching positions in Neo4j", "tenant_id", params.TenantID, "limit", params.Limit, "offset", params.Offset)

	// 如果Neo4j服务不可用，返回模拟数据
	if r.neo4jService == nil {
		r.logger.Info("Neo4j service not available, returning mock position search data", "tenant_id", params.TenantID)
		positions, total := r.getMockPositionSearchResult(params)
		return positions, total, nil
	}

	// TODO: 实现真实的Neo4j搜索查询
	// 目前返回模拟数据
	positions, total := r.getMockPositionSearchResult(params)
	
	r.logger.Info("Position search completed", "tenant_id", params.TenantID, "total_found", total, "returned", len(positions))
	
	return positions, total, nil
}

// GetPositionHierarchy 获取职位层级结构
func (r *Neo4jPositionQueryRepository) GetPositionHierarchy(ctx context.Context, query queries.GetPositionHierarchyQuery) ([]queries.PositionNode, error) {
	r.logger.Info("Getting position hierarchy", "tenant_id", query.TenantID, "max_depth", query.MaxDepth)

	// TODO: 实现真实的Neo4j层级查询
	// 目前返回模拟数据
	hierarchy := r.getMockPositionHierarchy(query.TenantID)
	
	r.logger.Info("Position hierarchy retrieved (mock data)", "tenant_id", query.TenantID)
	
	return hierarchy, nil
}

// GetEmployeePositions 获取员工的职位历史
func (r *Neo4jPositionQueryRepository) GetEmployeePositions(ctx context.Context, query queries.GetEmployeePositionsQuery) ([]queries.OccupancyHistoryResponse, error) {
	r.logger.Info("Getting employee positions", "employee_id", query.EmployeeID, "tenant_id", query.TenantID, "include_past", query.IncludePast)

	// TODO: 实现真实的Neo4j查询
	// 目前返回模拟数据
	history := r.getMockEmployeeOccupancyHistory(query.EmployeeID, query.TenantID, query.IncludePast)
	
	r.logger.Info("Employee positions retrieved (mock data)", "employee_id", query.EmployeeID, "count", len(history))
	
	return history, nil
}

// GetPositionEmployees 获取职位的员工列表
func (r *Neo4jPositionQueryRepository) GetPositionEmployees(ctx context.Context, query queries.GetPositionEmployeesQuery) ([]queries.EmployeeResponse, error) {
	r.logger.Info("Getting position employees", "position_id", query.PositionID, "tenant_id", query.TenantID, "only_current", query.OnlyCurrent)

	// TODO: 实现真实的Neo4j查询
	// 目前返回模拟数据
	employees := r.getMockPositionEmployees(query.PositionID, query.TenantID, query.OnlyCurrent)
	
	r.logger.Info("Position employees retrieved (mock data)", "position_id", query.PositionID, "count", len(employees))
	
	return employees, nil
}

// GetPositionOccupancyHistory 获取职位占用历史
func (r *Neo4jPositionQueryRepository) GetPositionOccupancyHistory(ctx context.Context, params OccupancyHistoryParams) ([]PositionOccupancyHistory, int, error) {
	r.logger.Info("Getting position occupancy history", "tenant_id", params.TenantID, "position_id", params.PositionID, "employee_id", params.EmployeeID)

	// TODO: 实现真实的Neo4j查询
	// 目前返回模拟数据
	history, total := r.getMockOccupancyHistory(params)
	
	r.logger.Info("Position occupancy history retrieved (mock data)", "tenant_id", params.TenantID, "count", len(history))
	
	return history, total, nil
}

// GetPositionStats 获取职位统计信息
func (r *Neo4jPositionQueryRepository) GetPositionStats(ctx context.Context, query queries.GetPositionStatsQuery) (*queries.PositionStatsResponse, error) {
	r.logger.Info("Getting position stats", "tenant_id", query.TenantID)

	// 暂时返回模拟数据，直到Neo4j连接配置完成
	stats := &queries.PositionStatsResponse{
		Total:              125,
		Open:               18,
		Filled:             95,
		Frozen:             8,
		PendingElimination: 4,
		AverageFTE:         0.92,
		VacancyRate:        14.4,
		TurnoverRate:       8.2,
	}

	r.logger.Info("Position stats retrieved (mock data)", "tenant_id", query.TenantID, "stats", stats)

	return stats, nil
}

// GetEmployeePositionHistory 获取员工职位历史
func (r *Neo4jPositionQueryRepository) GetEmployeePositionHistory(ctx context.Context, employeeID uuid.UUID, tenantID uuid.UUID, limit, offset int) ([]PositionOccupancyHistory, int, error) {
	r.logger.Info("Getting employee position history", "employee_id", employeeID, "tenant_id", tenantID)

	// TODO: 实现真实的Neo4j查询
	// 目前返回模拟数据
	history := []PositionOccupancyHistory{
		{
			ID:             uuid.New(),
			TenantID:       tenantID,
			PositionID:     uuid.New(),
			EmployeeID:     employeeID,
			StartDate:      time.Now().AddDate(0, -6, 0),
			IsCurrent:      true,
			FTE:            1.0,
			AssignmentType: "PRIMARY",
			Reason:         "员工入职",
			CreatedAt:      time.Now().AddDate(0, -6, 0),
			UpdatedAt:      time.Now(),
		},
	}

	return history, len(history), nil
}

// ===== 模拟数据生成方法 =====

// getMockPositionResponse 生成模拟职位响应数据
func (r *Neo4jPositionQueryRepository) getMockPositionResponse(positionID, tenantID uuid.UUID) *queries.PositionResponse {
	return &queries.PositionResponse{
		ID:           positionID,
		TenantID:     tenantID,
		PositionType: "REGULAR",
		JobProfileID: uuid.New(),
		DepartmentID: uuid.New(),
		Status:       "ACTIVE",
		BudgetedFTE:  1.0,
		Details: map[string]interface{}{
			"title":       "高级软件工程师",
			"description": "负责核心业务系统开发和维护",
			"level":       "Senior",
		},
		CreatedAt: time.Now().AddDate(0, -6, 0),
		UpdatedAt: time.Now().AddDate(0, -1, 0),
	}
}

// getMockPositionWithRelations 生成带关系的职位模拟数据
func (r *Neo4jPositionQueryRepository) getMockPositionWithRelations(positionID, tenantID uuid.UUID) *PositionWithRelations {
	position := Position{
		ID:           positionID,
		TenantID:     tenantID,
		PositionType: "REGULAR",
		JobProfileID: uuid.New(),
		DepartmentID: uuid.New(),
		Status:       "ACTIVE",
		BudgetedFTE:  1.0,
		Details: map[string]interface{}{
			"title":       "高级软件工程师",
			"description": "负责核心业务系统开发和维护",
		},
		CreatedAt: time.Now().AddDate(0, -6, 0),
		UpdatedAt: time.Now(),
	}

	return &PositionWithRelations{
		Position: position,
		Department: &Organization{
			ID:       position.DepartmentID,
			TenantID: tenantID,
			Name:     "技术部",
			UnitType: "DEPARTMENT",
		},
		CurrentEmployee: &Employee{
			ID:               uuid.New(),
			TenantID:         tenantID,
			FirstName:        "张",
			LastName:         "三",
			Email:            "zhang.san@company.com",
			EmploymentStatus: "ACTIVE",
			EmployeeType:     "FULL_TIME",
		},
	}
}

// getMockPositionSearchResult 生成模拟职位搜索结果
func (r *Neo4jPositionQueryRepository) getMockPositionSearchResult(params SearchPositionsParams) ([]Position, int) {
	mockPositions := []Position{
		{
			ID:           uuid.New(),
			TenantID:     params.TenantID,
			PositionType: "REGULAR",
			JobProfileID: uuid.New(),
			DepartmentID: uuid.New(),
			Status:       "ACTIVE",
			BudgetedFTE:  1.0,
			Details: map[string]interface{}{
				"title":       "高级软件工程师",
				"description": "负责核心业务系统开发",
			},
			CreatedAt: time.Now().AddDate(0, -6, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:           uuid.New(),
			TenantID:     params.TenantID,
			PositionType: "REGULAR",
			JobProfileID: uuid.New(),
			DepartmentID: uuid.New(),
			Status:       "OPEN",
			BudgetedFTE:  1.0,
			Details: map[string]interface{}{
				"title":       "产品经理",
				"description": "负责产品规划和需求管理",
			},
			CreatedAt: time.Now().AddDate(0, -3, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:           uuid.New(),
			TenantID:     params.TenantID,
			PositionType: "EXECUTIVE",
			JobProfileID: uuid.New(),
			DepartmentID: uuid.New(),
			Status:       "FILLED",
			BudgetedFTE:  1.0,
			Details: map[string]interface{}{
				"title":       "技术总监",
				"description": "负责技术战略和团队管理",
			},
			CreatedAt: time.Now().AddDate(0, -12, 0),
			UpdatedAt: time.Now(),
		},
	}

	// 应用过滤器
	var filteredPositions []Position
	for _, pos := range mockPositions {
		if params.Status != nil && pos.Status != *params.Status {
			continue
		}
		if params.PositionType != nil && pos.PositionType != *params.PositionType {
			continue
		}
		if params.Search != nil {
			title, _ := pos.Details["title"].(string)
			description, _ := pos.Details["description"].(string)
			if !containsIgnoreCase(title, *params.Search) && !containsIgnoreCase(description, *params.Search) {
				continue
			}
		}
		filteredPositions = append(filteredPositions, pos)
	}

	// 应用分页
	total := len(filteredPositions)
	start := params.Offset
	end := start + params.Limit
	
	if start > len(filteredPositions) {
		filteredPositions = []Position{}
	} else if end > len(filteredPositions) {
		filteredPositions = filteredPositions[start:]
	} else {
		filteredPositions = filteredPositions[start:end]
	}

	return filteredPositions, total
}

// getMockPositionHierarchy 生成模拟职位层级数据
func (r *Neo4jPositionQueryRepository) getMockPositionHierarchy(tenantID uuid.UUID) []queries.PositionNode {
	// 创建职位节点
	ceoPos := queries.PositionNode{
		Position: queries.PositionResponse{
			ID:       uuid.New(),
			TenantID: tenantID,
			Status:   "FILLED",
			Details: map[string]interface{}{
				"title": "首席执行官",
			},
		},
		Level: 0,
	}

	ctoPos := queries.PositionNode{
		Position: queries.PositionResponse{
			ID:                uuid.New(),
			TenantID:          tenantID,
			Status:            "FILLED",
			ManagerPositionID: &ceoPos.Position.ID,
			Details: map[string]interface{}{
				"title": "技术总监",
			},
		},
		Level: 1,
	}

	seniorDevPos := queries.PositionNode{
		Position: queries.PositionResponse{
			ID:                uuid.New(),
			TenantID:          tenantID,
			Status:            "FILLED",
			ManagerPositionID: &ctoPos.Position.ID,
			Details: map[string]interface{}{
				"title": "高级软件工程师",
			},
		},
		Level: 2,
	}

	return []queries.PositionNode{ceoPos, ctoPos, seniorDevPos}
}

// getMockEmployeeOccupancyHistory 生成员工职位占用历史模拟数据
func (r *Neo4jPositionQueryRepository) getMockEmployeeOccupancyHistory(employeeID, tenantID uuid.UUID, includePast bool) []queries.OccupancyHistoryResponse {
	history := []queries.OccupancyHistoryResponse{
		{
			ID:             uuid.New(),
			PositionID:     uuid.New(),
			EmployeeID:     employeeID,
			StartDate:      time.Now().AddDate(-1, 0, 0),
			IsCurrent:      true,
			FTE:            1.0,
			AssignmentType: "PRIMARY",
			Reason:         "职位调整",
			CreatedAt:      time.Now().AddDate(-1, 0, 0),
			UpdatedAt:      time.Now(),
		},
	}

	if includePast {
		history = append(history, queries.OccupancyHistoryResponse{
			ID:             uuid.New(),
			PositionID:     uuid.New(),
			EmployeeID:     employeeID,
			StartDate:      time.Now().AddDate(-3, 0, 0),
			EndDate:        timePtr(time.Now().AddDate(-1, 0, 0)),
			IsCurrent:      false,
			FTE:            1.0,
			AssignmentType: "PRIMARY",
			Reason:         "员工入职",
			CreatedAt:      time.Now().AddDate(-3, 0, 0),
			UpdatedAt:      time.Now().AddDate(-1, 0, 0),
		})
	}

	return history
}

// getMockPositionEmployees 生成职位员工模拟数据
func (r *Neo4jPositionQueryRepository) getMockPositionEmployees(positionID, tenantID uuid.UUID, onlyCurrent bool) []queries.EmployeeResponse {
	employees := []queries.EmployeeResponse{
		{
			ID:           uuid.New(),
			FirstName:    "张",
			LastName:     "三",
			Email:        "zhang.san@company.com",
			Status:       "ACTIVE",
			EmployeeType: "FULL_TIME",
		},
	}

	if !onlyCurrent {
		employees = append(employees, queries.EmployeeResponse{
			ID:           uuid.New(),
			FirstName:    "李",
			LastName:     "四",
			Email:        "li.si@company.com",
			Status:       "TERMINATED",
			EmployeeType: "FULL_TIME",
		})
	}

	return employees
}

// getMockOccupancyHistory 生成职位占用历史模拟数据
func (r *Neo4jPositionQueryRepository) getMockOccupancyHistory(params OccupancyHistoryParams) ([]PositionOccupancyHistory, int) {
	history := []PositionOccupancyHistory{
		{
			ID:             uuid.New(),
			TenantID:       params.TenantID,
			PositionID:     uuid.New(),
			EmployeeID:     uuid.New(),
			StartDate:      time.Now().AddDate(0, -6, 0),
			IsCurrent:      true,
			FTE:            1.0,
			AssignmentType: "PRIMARY",
			Reason:         "新员工入职",
			CreatedAt:      time.Now().AddDate(0, -6, 0),
			UpdatedAt:      time.Now(),
		},
		{
			ID:             uuid.New(),
			TenantID:       params.TenantID,
			PositionID:     uuid.New(),
			EmployeeID:     uuid.New(),
			StartDate:      time.Now().AddDate(-1, 0, 0),
			EndDate:        timePtr(time.Now().AddDate(0, -6, 0)),
			IsCurrent:      false,
			FTE:            1.0,
			AssignmentType: "PRIMARY",
			Reason:         "员工离职",
			CreatedAt:      time.Now().AddDate(-1, 0, 0),
			UpdatedAt:      time.Now().AddDate(0, -6, 0),
		},
	}

	// 应用过滤器
	var filteredHistory []PositionOccupancyHistory
	for _, h := range history {
		if params.PositionID != nil && h.PositionID != *params.PositionID {
			continue
		}
		if params.EmployeeID != nil && h.EmployeeID != *params.EmployeeID {
			continue
		}
		if params.IsCurrent != nil && h.IsCurrent != *params.IsCurrent {
			continue
		}
		filteredHistory = append(filteredHistory, h)
	}

	// 应用分页
	total := len(filteredHistory)
	start := params.Offset
	end := start + params.Limit
	
	if start > len(filteredHistory) {
		filteredHistory = []PositionOccupancyHistory{}
	} else if end > len(filteredHistory) {
		filteredHistory = filteredHistory[start:]
	} else {
		filteredHistory = filteredHistory[start:end]
	}

	return filteredHistory, total
}

// 辅助函数
func containsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

func timePtr(t time.Time) *time.Time {
	return &t
}