package corehr

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository CoreHR 数据访问层
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository 创建新的 Repository 实例
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Employee Repository Methods

// GetEmployeeByID 根据 ID 获取员工
func (r *Repository) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	var employee Employee
	err := r.db.QueryRow(ctx,
		`SELECT id, employee_number, first_name, last_name, email, phone_number, hire_date, position_id, organization_id, status, created_at, updated_at 
		 FROM corehr.employees WHERE id = $1`,
		id).Scan(&employee.ID, &employee.EmployeeNumber, &employee.FirstName, &employee.LastName, &employee.Email, &employee.PhoneNumber, &employee.HireDate, &employee.PositionID, &employee.OrganizationID, &employee.Status, &employee.CreatedAt, &employee.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

// GetEmployeeByNumber 根据员工编号获取员工
func (r *Repository) GetEmployeeByNumber(ctx context.Context, employeeNumber string) (*Employee, error) {
	var employee Employee
	err := r.db.QueryRow(ctx,
		`SELECT id, employee_number, first_name, last_name, email, phone_number, hire_date, position_id, organization_id, status, created_at, updated_at 
		 FROM corehr.employees WHERE employee_number = $1`,
		employeeNumber).Scan(&employee.ID, &employee.EmployeeNumber, &employee.FirstName, &employee.LastName, &employee.Email, &employee.PhoneNumber, &employee.HireDate, &employee.PositionID, &employee.OrganizationID, &employee.Status, &employee.CreatedAt, &employee.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

// ListEmployees 获取员工列表（支持分页和搜索）
func (r *Repository) ListEmployees(ctx context.Context, offset, limit int, search string) ([]Employee, int, error) {
	var query string
	var args []interface{}

	if search != "" {
		query = `SELECT id, employee_number, first_name, last_name, email, phone_number, hire_date, position_id, organization_id, status, created_at, updated_at 
				 FROM corehr.employees 
				 WHERE (first_name ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1 OR employee_number ILIKE $1)
				 ORDER BY created_at DESC
				 LIMIT $2 OFFSET $3`
		args = []interface{}{"%" + search + "%", limit, offset}
	} else {
		query = `SELECT id, employee_number, first_name, last_name, email, phone_number, hire_date, position_id, organization_id, status, created_at, updated_at 
				 FROM corehr.employees 
				 ORDER BY created_at DESC
				 LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var emp Employee
		err := rows.Scan(&emp.ID, &emp.EmployeeNumber, &emp.FirstName, &emp.LastName, &emp.Email, &emp.PhoneNumber, &emp.HireDate, &emp.PositionID, &emp.OrganizationID, &emp.Status, &emp.CreatedAt, &emp.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		employees = append(employees, emp)
	}

	// 获取总数
	var countQuery string
	var countArgs []interface{}
	if search != "" {
		countQuery = `SELECT COUNT(*) FROM corehr.employees WHERE (first_name ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1 OR employee_number ILIKE $1)`
		countArgs = []interface{}{"%" + search + "%"}
	} else {
		countQuery = `SELECT COUNT(*) FROM corehr.employees`
	}

	var totalCount int
	err = r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	return employees, totalCount, nil
}

// CreateEmployee 创建员工
func (r *Repository) CreateEmployee(ctx context.Context, employee *Employee) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO corehr.employees (id, employee_number, first_name, last_name, email, phone_number, hire_date, position_id, organization_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		employee.ID, employee.EmployeeNumber, employee.FirstName, employee.LastName, employee.Email, employee.PhoneNumber, employee.HireDate, employee.PositionID, employee.OrganizationID, employee.Status, employee.CreatedAt, employee.UpdatedAt)
	return err
}

// UpdateEmployee 更新员工
func (r *Repository) UpdateEmployee(ctx context.Context, employee *Employee) error {
	_, err := r.db.Exec(ctx,
		`UPDATE corehr.employees 
		 SET first_name = $1, last_name = $2, email = $3, phone_number = $4, hire_date = $5, position_id = $6, organization_id = $7, status = $8, updated_at = $9
		 WHERE id = $10`,
		employee.FirstName, employee.LastName, employee.Email, employee.PhoneNumber, employee.HireDate, employee.PositionID, employee.OrganizationID, employee.Status, employee.UpdatedAt, employee.ID)
	return err
}

// ListOrganizations 获取组织列表
func (r *Repository) ListOrganizations(ctx context.Context) ([]Organization, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, tenant_id, name, code, level, parent_id, status, created_at, updated_at 
		 FROM corehr.organizations 
		 ORDER BY level, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var organizations []Organization
	for rows.Next() {
		var org Organization
		err := rows.Scan(&org.ID, &org.TenantID, &org.Name, &org.Code, &org.Level, &org.ParentID, &org.Status, &org.CreatedAt, &org.UpdatedAt)
		if err != nil {
			return nil, err
		}
		organizations = append(organizations, org)
	}

	return organizations, nil
} 