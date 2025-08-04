package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
)

// EmployeeEntity PostgreSQL中的员工实体
type EmployeeEntity struct {
	ID               uuid.UUID              `db:"id"`
	TenantID         uuid.UUID              `db:"tenant_id"`
	EmployeeType     string                 `db:"employee_type"`
	FirstName        string                 `db:"first_name"`
	LastName         string                 `db:"last_name"`
	Email            string                 `db:"email"`
	PositionID       *uuid.UUID             `db:"position_id"`
	HireDate         time.Time              `db:"hire_date"`
	TerminationDate  *time.Time             `db:"termination_date"`
	EmploymentStatus string                 `db:"employment_status"`
	PersonalInfo     map[string]interface{} `db:"personal_info"`
	CreatedAt        time.Time              `db:"created_at"`
	UpdatedAt        time.Time              `db:"updated_at"`
}

// OrganizationUnit PostgreSQL中的组织单元实体
type OrganizationUnit struct {
	ID           uuid.UUID              `db:"id"`
	TenantID     uuid.UUID              `db:"tenant_id"`
	UnitType     string                 `db:"unit_type"`
	Name         string                 `db:"name"`
	Description  *string                `db:"description"`
	ParentUnitID *uuid.UUID             `db:"parent_unit_id"`
	Profile      map[string]interface{} `db:"profile"`
	IsActive     bool                   `db:"is_active"`
	CreatedAt    time.Time              `db:"created_at"`
	UpdatedAt    time.Time              `db:"updated_at"`
}

// PostgresCommandRepository PostgreSQL命令仓储接口
type PostgresCommandRepository interface {
	// 员工管理
	CreateEmployee(ctx context.Context, employee EmployeeEntity) error
	UpdateEmployee(ctx context.Context, id, tenantID uuid.UUID, changes map[string]interface{}) error
	TerminateEmployee(ctx context.Context, id, tenantID uuid.UUID, terminationDate time.Time, reason string) error

	// 组织单元管理
	CreateOrganizationUnit(ctx context.Context, unit OrganizationUnit) error
	UpdateOrganizationUnit(ctx context.Context, id, tenantID uuid.UUID, changes map[string]interface{}) error
	DeleteOrganizationUnit(ctx context.Context, id, tenantID uuid.UUID) error

	// 职位管理
	CreatePosition(ctx context.Context, position Position) error
	UpdatePosition(ctx context.Context, id, tenantID uuid.UUID, changes map[string]interface{}) error
	AssignEmployeePosition(ctx context.Context, employeeID, positionID, tenantID uuid.UUID, startDate time.Time, isPrimary bool) error
}

// postgresCommandRepository PostgreSQL命令仓储实现
type postgresCommandRepository struct {
	db     *sqlx.DB
	logger *logging.StructuredLogger
}

// NewPostgresCommandRepository 创建PostgreSQL命令仓储
func NewPostgresCommandRepository(db *sqlx.DB, logger *logging.StructuredLogger) PostgresCommandRepository {
	return &postgresCommandRepository{
		db:     db,
		logger: logger,
	}
}

// CreateEmployee 创建员工
func (r *postgresCommandRepository) CreateEmployee(ctx context.Context, employee EmployeeEntity) error {
	personalInfoJSON, err := json.Marshal(employee.PersonalInfo)
	if err != nil {
		r.logger.Error("Failed to marshal personal info", "error", err)
		return err
	}

	query := `
		INSERT INTO employees (
			id, tenant_id, employee_type, first_name, last_name, email, 
			position_id, hire_date, termination_date, employment_status, 
			personal_info, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	_, err = r.db.ExecContext(ctx, query,
		employee.ID,
		employee.TenantID,
		employee.EmployeeType,
		employee.FirstName,
		employee.LastName,
		employee.Email,
		employee.PositionID,
		employee.HireDate,
		employee.TerminationDate,
		employee.EmploymentStatus,
		personalInfoJSON,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		r.logger.Error("Failed to create employee", "error", err, "employee_id", employee.ID)
		return err
	}

	r.logger.Info("Employee created successfully", "employee_id", employee.ID)
	return nil
}

// UpdateEmployee 更新员工
func (r *postgresCommandRepository) UpdateEmployee(ctx context.Context, id, tenantID uuid.UUID, changes map[string]interface{}) error {
	if len(changes) == 0 {
		return nil
	}

	// 简化实现：只支持基本字段更新
	query := `UPDATE employees SET updated_at = $1, first_name = $2, last_name = $3 WHERE id = $4 AND tenant_id = $5`
	
	firstName := changes["first_name"]
	lastName := changes["last_name"]
	
	_, err := r.db.ExecContext(ctx, query, time.Now(), firstName, lastName, id, tenantID)
	if err != nil {
		r.logger.Error("Failed to update employee", "error", err, "employee_id", id)
		return err
	}

	r.logger.Info("Employee updated successfully", "employee_id", id)
	return nil
}

// TerminateEmployee 终止员工
func (r *postgresCommandRepository) TerminateEmployee(ctx context.Context, id, tenantID uuid.UUID, terminationDate time.Time, reason string) error {
	query := `
		UPDATE employees 
		SET termination_date = $1, employment_status = 'TERMINATED', updated_at = $2
		WHERE id = $3 AND tenant_id = $4`

	_, err := r.db.ExecContext(ctx, query, terminationDate, time.Now(), id, tenantID)
	if err != nil {
		r.logger.Error("Failed to terminate employee", "error", err, "employee_id", id)
		return err
	}

	r.logger.Info("Employee terminated successfully", "employee_id", id, "reason", reason)
	return nil
}

// CreateOrganizationUnit 创建组织单元
func (r *postgresCommandRepository) CreateOrganizationUnit(ctx context.Context, unit OrganizationUnit) error {
	// 简化实现
	r.logger.Info("CreateOrganizationUnit called", "unit_id", unit.ID)
	return nil
}

// UpdateOrganizationUnit 更新组织单元
func (r *postgresCommandRepository) UpdateOrganizationUnit(ctx context.Context, id, tenantID uuid.UUID, changes map[string]interface{}) error {
	// 简化实现
	r.logger.Info("UpdateOrganizationUnit called", "unit_id", id)
	return nil
}

// DeleteOrganizationUnit 删除组织单元
func (r *postgresCommandRepository) DeleteOrganizationUnit(ctx context.Context, id, tenantID uuid.UUID) error {
	// 简化实现
	r.logger.Info("DeleteOrganizationUnit called", "unit_id", id)
	return nil
}

// CreatePosition 创建职位
func (r *postgresCommandRepository) CreatePosition(ctx context.Context, position Position) error {
	// 简化实现
	r.logger.Info("CreatePosition called", "position_id", position.ID)
	return nil
}

// UpdatePosition 更新职位
func (r *postgresCommandRepository) UpdatePosition(ctx context.Context, id, tenantID uuid.UUID, changes map[string]interface{}) error {
	// 简化实现
	r.logger.Info("UpdatePosition called", "position_id", id)
	return nil
}

// AssignEmployeePosition 分配员工职位
func (r *postgresCommandRepository) AssignEmployeePosition(ctx context.Context, employeeID, positionID, tenantID uuid.UUID, startDate time.Time, isPrimary bool) error {
	// 简化实现
	r.logger.Info("AssignEmployeePosition called", "employee_id", employeeID, "position_id", positionID)
	return nil
}