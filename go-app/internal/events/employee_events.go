package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EmployeeEvents 员工相关领域事件

// EmployeeCreated 员工创建事件
type EmployeeCreated struct {
	*BaseDomainEvent
	EmployeeNumber string    `json:"employee_number"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Email          string    `json:"email"`
	PhoneNumber    *string   `json:"phone_number,omitempty"`
	HireDate       time.Time `json:"hire_date"`
	Status         string    `json:"status"`
	ManagerID      *uuid.UUID `json:"manager_id,omitempty"`
}

// NewEmployeeCreated 创建员工创建事件
func NewEmployeeCreated(tenantID, employeeID uuid.UUID, employeeNumber, firstName, lastName, email string, hireDate time.Time) *EmployeeCreated {
	base := NewBaseDomainEvent("employee.created", "employee", employeeID, tenantID)
	return &EmployeeCreated{
		BaseDomainEvent: base,
		EmployeeNumber:  employeeNumber,
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		HireDate:        hireDate,
		Status:          "active",
	}
}

func (e *EmployeeCreated) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// EmployeeUpdated 员工更新事件
type EmployeeUpdated struct {
	*BaseDomainEvent
	EmployeeNumber string                 `json:"employee_number"`
	UpdatedFields  map[string]interface{} `json:"updated_fields"`
	PreviousValues map[string]interface{} `json:"previous_values,omitempty"`
}

// NewEmployeeUpdated 创建员工更新事件
func NewEmployeeUpdated(tenantID, employeeID uuid.UUID, employeeNumber string, updatedFields map[string]interface{}) *EmployeeUpdated {
	base := NewBaseDomainEvent("employee.updated", "employee", employeeID, tenantID)
	return &EmployeeUpdated{
		BaseDomainEvent: base,
		EmployeeNumber:  employeeNumber,
		UpdatedFields:   updatedFields,
		PreviousValues:  make(map[string]interface{}),
	}
}

func (e *EmployeeUpdated) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// EmployeePhoneUpdated 员工电话更新事件 (特殊事件)
type EmployeePhoneUpdated struct {
	*BaseDomainEvent
	EmployeeNumber string `json:"employee_number"`
	OldPhoneNumber string `json:"old_phone_number"`
	NewPhoneNumber string `json:"new_phone_number"`
	UpdatedBy      string `json:"updated_by,omitempty"`
}

// NewEmployeePhoneUpdated 创建员工电话更新事件
func NewEmployeePhoneUpdated(tenantID, employeeID uuid.UUID, employeeNumber, oldPhone, newPhone string) *EmployeePhoneUpdated {
	base := NewBaseDomainEvent("employee.phone_updated", "employee", employeeID, tenantID)
	return &EmployeePhoneUpdated{
		BaseDomainEvent: base,
		EmployeeNumber:  employeeNumber,
		OldPhoneNumber:  oldPhone,
		NewPhoneNumber:  newPhone,
	}
}

func (e *EmployeePhoneUpdated) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// EmployeeDeleted 员工删除事件
type EmployeeDeleted struct {
	*BaseDomainEvent
	EmployeeNumber string    `json:"employee_number"`
	DeletedAt      time.Time `json:"deleted_at"`
	DeletedBy      string    `json:"deleted_by,omitempty"`
	DeletionReason string    `json:"deletion_reason,omitempty"`
}

// NewEmployeeDeleted 创建员工删除事件
func NewEmployeeDeleted(tenantID, employeeID uuid.UUID, employeeNumber string) *EmployeeDeleted {
	base := NewBaseDomainEvent("employee.deleted", "employee", employeeID, tenantID)
	return &EmployeeDeleted{
		BaseDomainEvent: base,
		EmployeeNumber:  employeeNumber,
		DeletedAt:       time.Now(),
	}
}

func (e *EmployeeDeleted) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// EmployeeHired 员工入职事件 (业务流程事件)
type EmployeeHired struct {
	*BaseDomainEvent
	EmployeeNumber string    `json:"employee_number"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Email          string    `json:"email"`
	HireDate       time.Time `json:"hire_date"`
	PositionID     *uuid.UUID `json:"position_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	ManagerID      *uuid.UUID `json:"manager_id,omitempty"`
}

// NewEmployeeHired 创建员工入职事件
func NewEmployeeHired(tenantID, employeeID uuid.UUID, employeeNumber, firstName, lastName, email string, hireDate time.Time) *EmployeeHired {
	base := NewBaseDomainEvent("employee.hired", "employee", employeeID, tenantID)
	return &EmployeeHired{
		BaseDomainEvent: base,
		EmployeeNumber:  employeeNumber,
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		HireDate:        hireDate,
	}
}

func (e *EmployeeHired) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// EmployeeTerminated 员工离职事件
type EmployeeTerminated struct {
	*BaseDomainEvent
	EmployeeNumber   string    `json:"employee_number"`
	TerminationDate  time.Time `json:"termination_date"`
	TerminationType  string    `json:"termination_type"` // voluntary, involuntary, retirement
	TerminationReason string   `json:"termination_reason,omitempty"`
	LastWorkingDay   time.Time `json:"last_working_day"`
}

// NewEmployeeTerminated 创建员工离职事件
func NewEmployeeTerminated(tenantID, employeeID uuid.UUID, employeeNumber, terminationType string, terminationDate time.Time) *EmployeeTerminated {
	base := NewBaseDomainEvent("employee.terminated", "employee", employeeID, tenantID)
	return &EmployeeTerminated{
		BaseDomainEvent:  base,
		EmployeeNumber:   employeeNumber,
		TerminationDate:  terminationDate,
		TerminationType:  terminationType,
		LastWorkingDay:   terminationDate,
	}
}

func (e *EmployeeTerminated) Serialize() ([]byte, error) {
	return json.Marshal(e)
}