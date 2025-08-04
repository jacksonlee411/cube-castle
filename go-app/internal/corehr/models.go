package corehr

import (
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/common"
	"github.com/google/uuid"
)

// Employee 员工模型
type Employee struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	TenantID       uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	EmployeeNumber string     `json:"employee_number" db:"employee_number"`
	FirstName      string     `json:"first_name" db:"first_name"`
	LastName       string     `json:"last_name" db:"last_name"`
	Email          string     `json:"email" db:"email"`
	PhoneNumber    *string    `json:"phone_number,omitempty" db:"phone_number"`
	Position       *string    `json:"position,omitempty" db:"position"`
	Department     *string    `json:"department,omitempty" db:"department"`
	HireDate       time.Time  `json:"hire_date" db:"hire_date"`
	ManagerID      *uuid.UUID `json:"manager_id,omitempty" db:"manager_id"`
	Status         string     `json:"status" db:"status"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// 注意： CreateEmployeeRequest, UpdateEmployeeRequest, EmployeeResponse 在 employee_service_enhanced.go 中定义

// Organization 组织模型
type Organization struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	TenantID  uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name      string     `json:"name" db:"name"`
	Code      string     `json:"code" db:"code"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Level     int        `json:"level" db:"level"`
	Status    string     `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	Name     string     `json:"name" validate:"required"`
	Code     string     `json:"code" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	Level    int        `json:"level"`
}

// UpdateOrganizationRequest 更新组织请求
type UpdateOrganizationRequest struct {
	Name     *string    `json:"name,omitempty"`
	Code     *string    `json:"code,omitempty"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	Level    *int       `json:"level,omitempty"`
}

// Position 职位模型
type Position struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Title        string     `json:"title" db:"title"`
	Code         string     `json:"code" db:"code"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty" db:"department_id"`
	Level        int        `json:"level" db:"level"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// CreatePositionRequest 创建职位请求
type CreatePositionRequest struct {
	Title        string     `json:"title" validate:"required"`
	Code         string     `json:"code" validate:"required"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	Level        int        `json:"level"`
}

// UpdatePositionRequest 更新职位请求
type UpdatePositionRequest struct {
	Title        *string    `json:"title,omitempty"`
	Code         *string    `json:"code,omitempty"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	Level        *int       `json:"level,omitempty"`
}

// EmployeeSearchRequest 员工搜索请求
type EmployeeSearchRequest struct {
	common.Pagination
	Query      *string `json:"query,omitempty"`
	Department *string `json:"department,omitempty"`
	Status     *string `json:"status,omitempty"`
}

// OrganizationTree 组织树结构
type OrganizationTree struct {
	Organization
	Children  []OrganizationTree `json:"children,omitempty"`
	Employees []EmployeeResponse `json:"employees,omitempty"`
}
