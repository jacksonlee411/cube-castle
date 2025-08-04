package events

import (
	"time"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/events"
)

// PositionCreatedEvent 职位创建事件
type PositionCreatedEvent struct {
	*events.BaseDomainEvent
	PositionID   uuid.UUID              `json:"position_id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	PositionType string                 `json:"position_type"`
	DepartmentID uuid.UUID              `json:"department_id"`
	Status       string                 `json:"status"`
	BudgetedFTE  float64                `json:"budgeted_fte"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// PositionUpdatedEvent 职位更新事件
type PositionUpdatedEvent struct {
	*events.BaseDomainEvent
	PositionID   uuid.UUID              `json:"position_id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	Changes      map[string]interface{} `json:"changes"`
	PreviousData map[string]interface{} `json:"previous_data"`
}

// EmployeeAssignedToPositionEvent 员工分配到职位事件
type EmployeeAssignedToPositionEvent struct {
	*events.BaseDomainEvent
	TenantID       uuid.UUID `json:"tenant_id"`
	PositionID     uuid.UUID `json:"position_id"`
	EmployeeID     uuid.UUID `json:"employee_id"`
	StartDate      time.Time `json:"start_date"`
	FTE            float64   `json:"fte"`
	AssignmentType string    `json:"assignment_type"`
	Reason         string    `json:"reason"`
}

// EmployeeRemovedFromPositionEvent 员工从职位移除事件
type EmployeeRemovedFromPositionEvent struct {
	*events.BaseDomainEvent
	TenantID   uuid.UUID `json:"tenant_id"`
	PositionID uuid.UUID `json:"position_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	EndDate    time.Time `json:"end_date"`
	Reason     string    `json:"reason"`
}

// PositionDeletedEvent 职位删除事件
type PositionDeletedEvent struct {
	*events.BaseDomainEvent
	PositionID uuid.UUID `json:"position_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Reason     string    `json:"reason"`
}

// PositionStatusChangedEvent 职位状态变更事件
type PositionStatusChangedEvent struct {
	*events.BaseDomainEvent
	PositionID     uuid.UUID `json:"position_id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	ChangedBy      uuid.UUID `json:"changed_by"`
	Reason         string    `json:"reason"`
}

// PositionTransferredEvent 职位转移事件
type PositionTransferredEvent struct {
	*events.BaseDomainEvent
	PositionID        uuid.UUID  `json:"position_id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	PreviousDeptID    uuid.UUID  `json:"previous_dept_id"`
	NewDepartmentID   uuid.UUID  `json:"new_department_id"`
	PreviousManagerID *uuid.UUID `json:"previous_manager_id,omitempty"`
	NewManagerID      *uuid.UUID `json:"new_manager_id,omitempty"`
	EffectiveDate     time.Time  `json:"effective_date"`
	TransferReason    string     `json:"transfer_reason"`
}

// PositionHierarchyChangedEvent 职位层级变更事件
type PositionHierarchyChangedEvent struct {
	*events.BaseDomainEvent
	PositionID        uuid.UUID  `json:"position_id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	PreviousManagerID *uuid.UUID `json:"previous_manager_id,omitempty"`
	NewManagerID      *uuid.UUID `json:"new_manager_id,omitempty"`
	EffectiveDate     time.Time  `json:"effective_date"`
}

// NewPositionCreatedEvent 创建职位创建事件
func NewPositionCreatedEvent(tenantID, positionID uuid.UUID, positionType string, deptID uuid.UUID, status string, fte float64, details map[string]interface{}) PositionCreatedEvent {
	base := events.NewBaseDomainEvent("position.created", "position", positionID, tenantID)
	return PositionCreatedEvent{
		BaseDomainEvent: base,
		PositionID:      positionID,
		TenantID:        tenantID,
		PositionType:    positionType,
		DepartmentID:    deptID,
		Status:          status,
		BudgetedFTE:     fte,
		Details:         details,
	}
}

// NewPositionUpdatedEvent 创建职位更新事件
func NewPositionUpdatedEvent(tenantID, positionID uuid.UUID, changes, previousData map[string]interface{}) PositionUpdatedEvent {
	base := events.NewBaseDomainEvent("position.updated", "position", positionID, tenantID)
	return PositionUpdatedEvent{
		BaseDomainEvent: base,
		PositionID:      positionID,
		TenantID:        tenantID,
		Changes:         changes,
		PreviousData:    previousData,
	}
}

// NewEmployeeAssignedToPositionEvent 创建员工分配到职位事件
func NewEmployeeAssignedToPositionEvent(tenantID, positionID, employeeID uuid.UUID, startDate time.Time, fte float64, assignmentType, reason string) EmployeeAssignedToPositionEvent {
	base := events.NewBaseDomainEvent("employee.assigned_to_position", "position", positionID, tenantID)
	return EmployeeAssignedToPositionEvent{
		BaseDomainEvent: base,
		TenantID:        tenantID,
		PositionID:      positionID,
		EmployeeID:      employeeID,
		StartDate:       startDate,
		FTE:             fte,
		AssignmentType:  assignmentType,
		Reason:          reason,
	}
}

// NewEmployeeRemovedFromPositionEvent 创建员工从职位移除事件
func NewEmployeeRemovedFromPositionEvent(tenantID, positionID, employeeID uuid.UUID, endDate time.Time, reason string) EmployeeRemovedFromPositionEvent {
	base := events.NewBaseDomainEvent("employee.removed_from_position", "position", positionID, tenantID)
	return EmployeeRemovedFromPositionEvent{
		BaseDomainEvent: base,
		TenantID:        tenantID,
		PositionID:      positionID,
		EmployeeID:      employeeID,
		EndDate:         endDate,
		Reason:          reason,
	}
}

// NewPositionDeletedEvent 创建职位删除事件
func NewPositionDeletedEvent(tenantID, positionID uuid.UUID, reason string) PositionDeletedEvent {
	base := events.NewBaseDomainEvent("position.deleted", "position", positionID, tenantID)
	return PositionDeletedEvent{
		BaseDomainEvent: base,
		PositionID:      positionID,
		TenantID:        tenantID,
		Reason:          reason,
	}
}

// NewPositionStatusChangedEvent 创建职位状态变更事件
func NewPositionStatusChangedEvent(tenantID, positionID, changedBy uuid.UUID, previousStatus, newStatus, reason string) PositionStatusChangedEvent {
	base := events.NewBaseDomainEvent("position.status_changed", "position", positionID, tenantID)
	return PositionStatusChangedEvent{
		BaseDomainEvent: base,
		PositionID:      positionID,
		TenantID:        tenantID,
		PreviousStatus:  previousStatus,
		NewStatus:       newStatus,
		ChangedBy:       changedBy,
		Reason:          reason,
	}
}

// NewPositionTransferredEvent 创建职位转移事件
func NewPositionTransferredEvent(tenantID, positionID uuid.UUID, prevDeptID, newDeptID uuid.UUID, prevMgrID, newMgrID *uuid.UUID, effectiveDate time.Time, reason string) PositionTransferredEvent {
	base := events.NewBaseDomainEvent("position.transferred", "position", positionID, tenantID)
	return PositionTransferredEvent{
		BaseDomainEvent:   base,
		PositionID:        positionID,
		TenantID:          tenantID,
		PreviousDeptID:    prevDeptID,
		NewDepartmentID:   newDeptID,
		PreviousManagerID: prevMgrID,
		NewManagerID:      newMgrID,
		EffectiveDate:     effectiveDate,
		TransferReason:    reason,
	}
}