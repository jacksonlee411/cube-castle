package corehr

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/employee"
	"github.com/gaogu/cube-castle/go-app/internal/common"
)

// EmployeeService 增强的员工服务，支持业务ID
type EmployeeService struct {
	client        *ent.Client
	businessIDMgr *common.BusinessIDManager
}

// NewEmployeeService 创建员工服务实例
func NewEmployeeService(client *ent.Client, businessIDMgr *common.BusinessIDManager) *EmployeeService {
	return &EmployeeService{
		client:        client,
		businessIDMgr: businessIDMgr,
	}
}

// EmployeeQueryOptions 员工查询选项
type EmployeeQueryOptions struct {
	IncludeUUID     bool
	WithPosition    bool
	WithOrgUnit     bool
	WithManager     bool
	TenantID        uuid.UUID
}

// DefaultEmployeeQueryOptions 默认查询选项
func DefaultEmployeeQueryOptions() EmployeeQueryOptions {
	return EmployeeQueryOptions{
		IncludeUUID:  false,
		WithPosition: false,
		WithOrgUnit:  false,
		WithManager:  false,
	}
}

// EmployeeResponse 员工响应结构
type EmployeeResponse struct {
	ID               string                 `json:"id"`                          // 业务ID
	UUID             *string                `json:"uuid,omitempty"`              // 系统UUID (可选)
	FirstName        string                 `json:"first_name"`
	LastName         string                 `json:"last_name"`
	Email            string                 `json:"email"`
	PhoneNumber      *string                `json:"phone_number,omitempty"`
	HireDate         string                 `json:"hire_date"`
	EmploymentStatus string                 `json:"employment_status"`
	EmployeeType     string                 `json:"employee_type"`
	PositionID       *string                `json:"position_id,omitempty"`       // 职位业务ID
	OrganizationID   *string                `json:"organization_id,omitempty"`   // 组织业务ID
	ManagerID        *string                `json:"manager_id,omitempty"`        // 经理业务ID
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	
	// 扩展信息 (当查询选项启用时)
	Position         *PositionInfo          `json:"position,omitempty"`
	Organization     *OrganizationInfo      `json:"organization,omitempty"`
	Manager          *EmployeeInfo          `json:"manager,omitempty"`
}

// PositionInfo 职位信息
type PositionInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Level int    `json:"level,omitempty"`
}

// OrganizationInfo 组织信息
type OrganizationInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	UnitType string `json:"unit_type"`
}

// EmployeeInfo 员工基本信息
type EmployeeInfo struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// GetEmployeeByBusinessID 通过业务ID获取员工
func (s *EmployeeService) GetEmployeeByBusinessID(ctx context.Context, businessID string, opts EmployeeQueryOptions) (*EmployeeResponse, error) {
	if err := common.ValidateBusinessID(common.EntityTypeEmployee, businessID); err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	query := s.client.Employee.Query().
		Where(employee.BusinessID(businessID))

	// 添加租户隔离
	if opts.TenantID != uuid.Nil {
		query = query.Where(employee.TenantID(opts.TenantID))
	}

	// 添加关联查询
	if opts.WithPosition {
		query = query.WithCurrentPosition()
	}

	emp, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("employee with business ID %s not found", businessID)
		}
		return nil, fmt.Errorf("failed to query employee: %w", err)
	}

	return s.buildEmployeeResponse(ctx, emp, opts)
}

// GetEmployeeByUUID 通过UUID获取员工 (向后兼容)
func (s *EmployeeService) GetEmployeeByUUID(ctx context.Context, id uuid.UUID, opts EmployeeQueryOptions) (*EmployeeResponse, error) {
	query := s.client.Employee.Query().
		Where(employee.ID(id))

	// 添加租户隔离
	if opts.TenantID != uuid.Nil {
		query = query.Where(employee.TenantID(opts.TenantID))
	}

	// 添加关联查询
	if opts.WithPosition {
		query = query.WithCurrentPosition()
	}

	emp, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("employee with UUID %s not found", id)
		}
		return nil, fmt.Errorf("failed to query employee: %w", err)
	}

	return s.buildEmployeeResponse(ctx, emp, opts)
}

// CreateEmployeeRequest 创建员工请求
type CreateEmployeeRequest struct {
	TenantID         uuid.UUID              `json:"tenant_id"`
	FirstName        string                 `json:"first_name"`
	LastName         string                 `json:"last_name"`
	Email            string                 `json:"email"`
	PhoneNumber      *string                `json:"phone_number,omitempty"`
	HireDate         string                 `json:"hire_date"`
	EmployeeType     string                 `json:"employee_type"`
	PositionID       *string                `json:"position_id,omitempty"`       // 职位业务ID
	OrganizationID   *string                `json:"organization_id,omitempty"`   // 组织业务ID  
	ManagerID        *string                `json:"manager_id,omitempty"`        // 经理业务ID
	EmployeeDetails  map[string]interface{} `json:"employee_details,omitempty"`
}

// CreateEmployee 创建新员工
func (s *EmployeeService) CreateEmployee(ctx context.Context, req CreateEmployeeRequest) (*EmployeeResponse, error) {
	// 生成业务ID
	businessID, err := s.businessIDMgr.GenerateUniqueBusinessID(ctx, common.EntityTypeEmployee)
	if err != nil {
		return nil, fmt.Errorf("failed to generate business ID: %w", err)
	}

	// 解析日期
	hireDate, err := parseDate(req.HireDate)
	if err != nil {
		return nil, fmt.Errorf("invalid hire date: %w", err)
	}

	// 开始创建员工
	create := s.client.Employee.Create().
		SetBusinessID(businessID).
		SetTenantID(req.TenantID).
		SetFirstName(req.FirstName).
		SetLastName(req.LastName).
		SetEmail(req.Email).
		SetHireDate(hireDate).
		SetEmployeeType(employee.EmployeeType(req.EmployeeType))

	// 设置可选字段
	if req.PhoneNumber != nil {
		create = create.SetPhoneNumber(*req.PhoneNumber)
	}

	if req.EmployeeDetails != nil {
		create = create.SetEmployeeDetails(req.EmployeeDetails)
	}

	// 处理关联关系 (如果提供了业务ID)
	if req.PositionID != nil {
		positionUUID, err := s.resolvePositionBusinessID(ctx, *req.PositionID)
		if err != nil {
			return nil, fmt.Errorf("invalid position ID: %w", err)
		}
		create = create.SetCurrentPositionID(positionUUID)
	}

	emp, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	// 返回创建的员工信息
	return s.buildEmployeeResponse(ctx, emp, DefaultEmployeeQueryOptions())
}

// UpdateEmployeeRequest 更新员工请求
type UpdateEmployeeRequest struct {
	FirstName        *string                `json:"first_name,omitempty"`
	LastName         *string                `json:"last_name,omitempty"`
	Email            *string                `json:"email,omitempty"`
	PhoneNumber      *string                `json:"phone_number,omitempty"`
	EmploymentStatus *string                `json:"employment_status,omitempty"`
	PositionID       *string                `json:"position_id,omitempty"`       // 职位业务ID
	OrganizationID   *string                `json:"organization_id,omitempty"`   // 组织业务ID
	ManagerID        *string                `json:"manager_id,omitempty"`        // 经理业务ID
	EmployeeDetails  map[string]interface{} `json:"employee_details,omitempty"`
}

// UpdateEmployee 更新员工信息
func (s *EmployeeService) UpdateEmployee(ctx context.Context, businessID string, req UpdateEmployeeRequest, opts EmployeeQueryOptions) (*EmployeeResponse, error) {
	if err := common.ValidateBusinessID(common.EntityTypeEmployee, businessID); err != nil {
		return nil, fmt.Errorf("invalid business ID: %w", err)
	}

	// 查找员工
	emp, err := s.client.Employee.Query().
		Where(employee.BusinessID(businessID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("employee with business ID %s not found", businessID)
		}
		return nil, fmt.Errorf("failed to query employee: %w", err)
	}

	// 构建更新查询
	update := s.client.Employee.UpdateOneID(emp.ID)

	// 应用更新字段
	if req.FirstName != nil {
		update = update.SetFirstName(*req.FirstName)
	}
	if req.LastName != nil {
		update = update.SetLastName(*req.LastName)
	}
	if req.Email != nil {
		update = update.SetEmail(*req.Email)
	}
	if req.PhoneNumber != nil {
		update = update.SetPhoneNumber(*req.PhoneNumber)
	}
	if req.EmploymentStatus != nil {
		update = update.SetEmploymentStatus(employee.EmploymentStatus(*req.EmploymentStatus))
	}
	if req.EmployeeDetails != nil {
		update = update.SetEmployeeDetails(req.EmployeeDetails)
	}

	// 处理关联关系更新
	if req.PositionID != nil {
		positionUUID, err := s.resolvePositionBusinessID(ctx, *req.PositionID)
		if err != nil {
			return nil, fmt.Errorf("invalid position ID: %w", err)
		}
		update = update.SetCurrentPositionID(positionUUID)
	}

	updatedEmp, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	return s.buildEmployeeResponse(ctx, updatedEmp, opts)
}

// ListEmployeesRequest 员工列表请求
type ListEmployeesRequest struct {
	TenantID    uuid.UUID              `json:"tenant_id"`
	Page        int                    `json:"page"`
	PageSize    int                    `json:"page_size"`
	Search      *string                `json:"search,omitempty"`
	Status      *string                `json:"status,omitempty"`
	EmployeeType *string               `json:"employee_type,omitempty"`
	OrganizationID *string             `json:"organization_id,omitempty"`  // 组织业务ID
	QueryOptions EmployeeQueryOptions  `json:"query_options"`
}

// ListEmployeesResponse 员工列表响应
type ListEmployeesResponse struct {
	Employees   []*EmployeeResponse `json:"employees"`
	TotalCount  int                 `json:"total_count"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
	TotalPages  int                 `json:"total_pages"`
}

// ListEmployees 获取员工列表
func (s *EmployeeService) ListEmployees(ctx context.Context, req ListEmployeesRequest) (*ListEmployeesResponse, error) {
	query := s.client.Employee.Query().
		Where(employee.TenantID(req.TenantID))

	// 应用搜索过滤
	if req.Search != nil && *req.Search != "" {
		query = query.Where(
			employee.Or(
				employee.FirstNameContains(*req.Search),
				employee.LastNameContains(*req.Search),
				employee.EmailContains(*req.Search),
				employee.BusinessIDContains(*req.Search),
			),
		)
	}

	// 应用状态过滤
	if req.Status != nil {
		query = query.Where(employee.EmploymentStatusEQ(employee.EmploymentStatus(*req.Status)))
	}

	// 应用员工类型过滤
	if req.EmployeeType != nil {
		query = query.Where(employee.EmployeeTypeEQ(employee.EmployeeType(*req.EmployeeType)))
	}

	// 获取总数
	totalCount, err := query.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count employees: %w", err)
	}

	// 应用分页
	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	// 添加关联查询
	if req.QueryOptions.WithPosition {
		query = query.WithCurrentPosition()
	}

	// 执行查询
	employees, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query employees: %w", err)
	}

	// 构建响应
	employeeResponses := make([]*EmployeeResponse, len(employees))
	for i, emp := range employees {
		employeeResponses[i], err = s.buildEmployeeResponse(ctx, emp, req.QueryOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to build employee response: %w", err)
		}
	}

	totalPages := (totalCount + req.PageSize - 1) / req.PageSize

	return &ListEmployeesResponse{
		Employees:  employeeResponses,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// DeleteEmployee 删除员工
func (s *EmployeeService) DeleteEmployee(ctx context.Context, businessID string) error {
	if err := common.ValidateBusinessID(common.EntityTypeEmployee, businessID); err != nil {
		return fmt.Errorf("invalid business ID: %w", err)
	}

	affected, err := s.client.Employee.Delete().
		Where(employee.BusinessID(businessID)).
		Exec(ctx)
	
	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("employee with business ID %s not found", businessID)
	}

	return nil
}

// buildEmployeeResponse 构建员工响应
func (s *EmployeeService) buildEmployeeResponse(ctx context.Context, emp *ent.Employee, opts EmployeeQueryOptions) (*EmployeeResponse, error) {
	response := &EmployeeResponse{
		ID:               emp.BusinessID,
		FirstName:        emp.FirstName,
		LastName:         emp.LastName,
		Email:            emp.Email,
		HireDate:         emp.HireDate.Format("2006-01-02"),
		EmploymentStatus: string(emp.EmploymentStatus),
		EmployeeType:     string(emp.EmployeeType),
		CreatedAt:        emp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        emp.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// 包含UUID (如果请求)
	if opts.IncludeUUID {
		uuidStr := emp.ID.String()
		response.UUID = &uuidStr
	}

	// 设置可选字段
	if emp.PhoneNumber != nil {
		response.PhoneNumber = emp.PhoneNumber
	}

	// 处理关联信息
	if opts.WithPosition && emp.Edges.CurrentPosition != nil {
		// pos := emp.Edges.CurrentPosition // TODO: 当Position实现所需字段后恢复
		response.Position = &PositionInfo{
			ID:    "placeholder", // TODO: pos.BusinessID 当Position实现business_id后修改
			Title: "placeholder", // TODO: pos.Title 当Position实现title字段后修改
		}
		response.PositionID = &response.Position.ID
	}

	return response, nil
}

// resolvePositionBusinessID 解析职位业务ID到UUID
func (s *EmployeeService) resolvePositionBusinessID(ctx context.Context, businessID string) (uuid.UUID, error) {
	if err := common.ValidateBusinessID(common.EntityTypePosition, businessID); err != nil {
		return uuid.Nil, err
	}

	// 这里需要查询Position表获取UUID
	// 由于我们还没有Position的业务ID支持，暂时返回一个占位符
	// 在实际实现中，这里应该查询position表的business_id字段
	return uuid.New(), fmt.Errorf("position business ID resolution not implemented yet")
}

// parseDate 解析日期字符串
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}