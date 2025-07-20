package outbox

import (
	"context"
	"github.com/google/uuid"
)

// EmployeeRepository 员工仓储接口
type EmployeeRepository interface {
	GetEmployeeByID(ctx context.Context, tenantID, employeeID uuid.UUID) (*EmployeeInfo, error)
	UpdateEmployeePhone(ctx context.Context, tenantID, employeeID uuid.UUID, phoneNumber string) error
}

// OrganizationRepository 组织仓储接口
type OrganizationRepository interface {
	GetOrganizationByID(ctx context.Context, tenantID, organizationID uuid.UUID) (*OrganizationInfo, error)
}

// LeaveRequestRepository 休假申请仓储接口
type LeaveRequestRepository interface {
	GetLeaveRequestByID(ctx context.Context, tenantID, requestID uuid.UUID) (*LeaveRequestInfo, error)
	UpdateLeaveRequestStatus(ctx context.Context, tenantID, requestID uuid.UUID, status, comment string, approvedBy *uuid.UUID) error
}

// EmployeeInfo 员工信息（用于事件处理）
type EmployeeInfo struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	EmployeeNumber string
	FirstName      string
	LastName       string
	Email          string
	PhoneNumber    *string
	Position       *string
	Department     *string
	ManagerID      *uuid.UUID
	Status         string
}

// OrganizationInfo 组织信息（用于事件处理）
type OrganizationInfo struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	Name     string
	Code     string
	ParentID *uuid.UUID
	Status   string
}

// LeaveRequestInfo 休假申请信息（用于事件处理）
type LeaveRequestInfo struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	EmployeeID uuid.UUID
	ManagerID  uuid.UUID
	StartDate  string
	EndDate    string
	LeaveType  string
	Reason     string
	Status     string
} 