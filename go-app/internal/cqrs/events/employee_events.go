package events

import (
	"time"
	"github.com/google/uuid"
)

// EmployeeHired 员工雇佣事件
type EmployeeHired struct {
	EventID      uuid.UUID `json:"event_id"`
	EmployeeID   uuid.UUID `json:"employee_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	HireDate     time.Time `json:"hire_date"`
	EmployeeType string    `json:"employee_type"`
	Timestamp    time.Time `json:"timestamp"`
}

// EmployeeUpdated 员工更新事件
type EmployeeUpdated struct {
	EventID    uuid.UUID              `json:"event_id"`
	EmployeeID uuid.UUID              `json:"employee_id"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	Changes    map[string]interface{} `json:"changes"`
	Timestamp  time.Time              `json:"timestamp"`
}

// EmployeeTerminated 员工终止事件
type EmployeeTerminated struct {
	EventID         uuid.UUID `json:"event_id"`
	EmployeeID      uuid.UUID `json:"employee_id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	TerminationDate time.Time `json:"termination_date"`
	Reason          string    `json:"reason"`
	Timestamp       time.Time `json:"timestamp"`
}

// OrganizationUnitCreated 组织单元创建事件
type OrganizationUnitCreated struct {
	EventID      uuid.UUID              `json:"event_id"`
	UnitID       uuid.UUID              `json:"unit_id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	UnitType     string                 `json:"unit_type"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description,omitempty"`
	ParentUnitID *uuid.UUID             `json:"parent_unit_id,omitempty"`
	Profile      map[string]interface{} `json:"profile,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// OrganizationUnitUpdated 组织单元更新事件
type OrganizationUnitUpdated struct {
	EventID   uuid.UUID              `json:"event_id"`
	UnitID    uuid.UUID              `json:"unit_id"`
	TenantID  uuid.UUID              `json:"tenant_id"`
	Changes   map[string]interface{} `json:"changes"`
	Timestamp time.Time              `json:"timestamp"`
}

// PositionCreated 职位创建事件
type PositionCreated struct {
	EventID      uuid.UUID `json:"event_id"`
	PositionID   uuid.UUID `json:"position_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Title        string    `json:"title"`
	Department   string    `json:"department"`
	Level        string    `json:"level"`
	Description  *string   `json:"description,omitempty"`
	Requirements *string   `json:"requirements,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// EmployeePositionAssigned 员工职位分配事件
type EmployeePositionAssigned struct {
	EventID    uuid.UUID `json:"event_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	PositionID uuid.UUID `json:"position_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	StartDate  time.Time `json:"start_date"`
	IsPrimary  bool      `json:"is_primary"`
	Timestamp  time.Time `json:"timestamp"`
}

// DomainEvent 领域事件接口
type DomainEvent interface {
	GetEventID() uuid.UUID
	GetTenantID() uuid.UUID
	GetTimestamp() time.Time
	GetEventType() string
}

// 实现DomainEvent接口
func (e EmployeeHired) GetEventID() uuid.UUID        { return e.EventID }
func (e EmployeeHired) GetTenantID() uuid.UUID       { return e.TenantID }
func (e EmployeeHired) GetTimestamp() time.Time      { return e.Timestamp }
func (e EmployeeHired) GetEventType() string         { return "employee.hired" }

func (e EmployeeUpdated) GetEventID() uuid.UUID      { return e.EventID }
func (e EmployeeUpdated) GetTenantID() uuid.UUID     { return e.TenantID }
func (e EmployeeUpdated) GetTimestamp() time.Time    { return e.Timestamp }
func (e EmployeeUpdated) GetEventType() string       { return "employee.updated" }

func (e EmployeeTerminated) GetEventID() uuid.UUID   { return e.EventID }
func (e EmployeeTerminated) GetTenantID() uuid.UUID  { return e.TenantID }
func (e EmployeeTerminated) GetTimestamp() time.Time { return e.Timestamp }
func (e EmployeeTerminated) GetEventType() string    { return "employee.terminated" }

func (e OrganizationUnitCreated) GetEventID() uuid.UUID   { return e.EventID }
func (e OrganizationUnitCreated) GetTenantID() uuid.UUID  { return e.TenantID }
func (e OrganizationUnitCreated) GetTimestamp() time.Time { return e.Timestamp }
func (e OrganizationUnitCreated) GetEventType() string    { return "organization_unit.created" }

func (e OrganizationUnitUpdated) GetEventID() uuid.UUID   { return e.EventID }
func (e OrganizationUnitUpdated) GetTenantID() uuid.UUID  { return e.TenantID }
func (e OrganizationUnitUpdated) GetTimestamp() time.Time { return e.Timestamp }
func (e OrganizationUnitUpdated) GetEventType() string    { return "organization_unit.updated" }

func (e PositionCreated) GetEventID() uuid.UUID        { return e.EventID }
func (e PositionCreated) GetTenantID() uuid.UUID       { return e.TenantID }
func (e PositionCreated) GetTimestamp() time.Time      { return e.Timestamp }
func (e PositionCreated) GetEventType() string         { return "position.created" }

func (e EmployeePositionAssigned) GetEventID() uuid.UUID   { return e.EventID }
func (e EmployeePositionAssigned) GetTenantID() uuid.UUID  { return e.TenantID }
func (e EmployeePositionAssigned) GetTimestamp() time.Time { return e.Timestamp }
func (e EmployeePositionAssigned) GetEventType() string    { return "employee.position.assigned" }