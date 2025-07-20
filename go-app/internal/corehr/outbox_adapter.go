package corehr

import (
	"context"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/outbox"
)

// OutboxAdapter CoreHR到Outbox的适配器
type OutboxAdapter struct {
	repo *Repository
}

// NewOutboxAdapter 创建新的适配器
func NewOutboxAdapter(repo *Repository) *OutboxAdapter {
	return &OutboxAdapter{repo: repo}
}

// GetEmployeeByID 实现EmployeeRepository接口
func (a *OutboxAdapter) GetEmployeeByID(ctx context.Context, tenantID, employeeID uuid.UUID) (*outbox.EmployeeInfo, error) {
	employee, err := a.repo.GetEmployeeByID(ctx, tenantID, employeeID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, nil
	}
	
	return &outbox.EmployeeInfo{
		ID:             employee.ID,
		TenantID:       employee.TenantID,
		EmployeeNumber: employee.EmployeeNumber,
		FirstName:      employee.FirstName,
		LastName:       employee.LastName,
		Email:          employee.Email,
		PhoneNumber:    employee.PhoneNumber,
		Position:       employee.Position,
		Department:     employee.Department,
		ManagerID:      employee.ManagerID,
		Status:         employee.Status,
	}, nil
}

// UpdateEmployeePhone 实现EmployeeRepository接口
func (a *OutboxAdapter) UpdateEmployeePhone(ctx context.Context, tenantID, employeeID uuid.UUID, phoneNumber string) error {
	// 这里可以实现具体的电话更新逻辑
	// 目前只是占位符实现
	return nil
}

// GetOrganizationByID 实现OrganizationRepository接口
func (a *OutboxAdapter) GetOrganizationByID(ctx context.Context, tenantID, organizationID uuid.UUID) (*outbox.OrganizationInfo, error) {
	organization, err := a.repo.GetOrganizationByID(ctx, tenantID, organizationID)
	if err != nil {
		return nil, err
	}
	if organization == nil {
		return nil, nil
	}
	
	return &outbox.OrganizationInfo{
		ID:       organization.ID,
		TenantID: organization.TenantID,
		Name:     organization.Name,
		Code:     organization.Code,
		ParentID: organization.ParentID,
		Status:   organization.Status,
	}, nil
}

// GetLeaveRequestByID 实现LeaveRequestRepository接口
func (a *OutboxAdapter) GetLeaveRequestByID(ctx context.Context, tenantID, requestID uuid.UUID) (*outbox.LeaveRequestInfo, error) {
	// 这里可以实现具体的休假申请查询逻辑
	// 目前只是占位符实现
	return nil, nil
}

// UpdateLeaveRequestStatus 实现LeaveRequestRepository接口
func (a *OutboxAdapter) UpdateLeaveRequestStatus(ctx context.Context, tenantID, requestID uuid.UUID, status, comment string, approvedBy *uuid.UUID) error {
	// 这里可以实现具体的休假申请状态更新逻辑
	// 目前只是占位符实现
	return nil
} 