package corehr

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
func (r *Repository) GetEmployeeByID(ctx context.Context, tenantID, employeeID uuid.UUID) (*Employee, error) {
	var employee Employee
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, employee_number, first_name, last_name, email, phone_number, position, department, hire_date, manager_id, status, created_at, updated_at 
		 FROM corehr.employees WHERE id = $1 AND tenant_id = $2`,
		employeeID, tenantID).Scan(
		&employee.ID, &employee.TenantID, &employee.EmployeeNumber, &employee.FirstName, &employee.LastName,
		&employee.Email, &employee.PhoneNumber, &employee.Position, &employee.Department, &employee.HireDate,
		&employee.ManagerID, &employee.Status, &employee.CreatedAt, &employee.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get employee by ID: %w", err)
	}
	return &employee, nil
}

// GetEmployeeByNumber 根据员工编号获取员工
func (r *Repository) GetEmployeeByNumber(ctx context.Context, tenantID uuid.UUID, employeeNumber string) (*Employee, error) {
	var employee Employee
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, employee_number, first_name, last_name, email, phone_number, position, department, hire_date, manager_id, status, created_at, updated_at 
		 FROM corehr.employees WHERE employee_number = $1 AND tenant_id = $2`,
		employeeNumber, tenantID).Scan(
		&employee.ID, &employee.TenantID, &employee.EmployeeNumber, &employee.FirstName, &employee.LastName,
		&employee.Email, &employee.PhoneNumber, &employee.Position, &employee.Department, &employee.HireDate,
		&employee.ManagerID, &employee.Status, &employee.CreatedAt, &employee.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get employee by number: %w", err)
	}
	return &employee, nil
}

// ListEmployees 获取员工列表（支持分页和搜索）
func (r *Repository) ListEmployees(ctx context.Context, tenantID uuid.UUID, page, pageSize int, search string) ([]Employee, int, error) {
	offset := (page - 1) * pageSize

	var query string
	var args []interface{}

	if search != "" {
		query = `SELECT id, tenant_id, employee_number, first_name, last_name, email, phone_number, position, department, hire_date, manager_id, status, created_at, updated_at 
				 FROM corehr.employees 
				 WHERE tenant_id = $1 AND (first_name ILIKE $2 OR last_name ILIKE $2 OR email ILIKE $2 OR employee_number ILIKE $2)
				 ORDER BY created_at DESC
				 LIMIT $3 OFFSET $4`
		args = []interface{}{tenantID, "%" + search + "%", pageSize, offset}
	} else {
		query = `SELECT id, tenant_id, employee_number, first_name, last_name, email, phone_number, position, department, hire_date, manager_id, status, created_at, updated_at 
				 FROM corehr.employees 
				 WHERE tenant_id = $1
				 ORDER BY created_at DESC
				 LIMIT $2 OFFSET $3`
		args = []interface{}{tenantID, pageSize, offset}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query employees: %w", err)
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var emp Employee
		err := rows.Scan(
			&emp.ID, &emp.TenantID, &emp.EmployeeNumber, &emp.FirstName, &emp.LastName,
			&emp.Email, &emp.PhoneNumber, &emp.Position, &emp.Department, &emp.HireDate,
			&emp.ManagerID, &emp.Status, &emp.CreatedAt, &emp.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan employee: %w", err)
		}
		employees = append(employees, emp)
	}

	// 获取总数
	var countQuery string
	var countArgs []interface{}
	if search != "" {
		countQuery = `SELECT COUNT(*) FROM corehr.employees WHERE tenant_id = $1 AND (first_name ILIKE $2 OR last_name ILIKE $2 OR email ILIKE $2 OR employee_number ILIKE $2)`
		countArgs = []interface{}{tenantID, "%" + search + "%"}
	} else {
		countQuery = `SELECT COUNT(*) FROM corehr.employees WHERE tenant_id = $1`
		countArgs = []interface{}{tenantID}
	}

	var totalCount int
	err = r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count employees: %w", err)
	}

	return employees, totalCount, nil
}

// CreateEmployee 创建员工
func (r *Repository) CreateEmployee(ctx context.Context, employee *Employee) error {
	now := time.Now()
	employee.CreatedAt = now
	employee.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO corehr.employees (id, tenant_id, employee_number, first_name, last_name, email, phone_number, position, department, hire_date, manager_id, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		employee.ID, employee.TenantID, employee.EmployeeNumber, employee.FirstName, employee.LastName,
		employee.Email, employee.PhoneNumber, employee.Position, employee.Department, employee.HireDate,
		employee.ManagerID, employee.Status, employee.CreatedAt, employee.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create employee: %w", err)
	}
	return nil
}

// UpdateEmployee 更新员工
func (r *Repository) UpdateEmployee(ctx context.Context, employee *Employee) error {
	employee.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`UPDATE corehr.employees 
		 SET first_name = $1, last_name = $2, email = $3, phone_number = $4, position = $5, department = $6, hire_date = $7, manager_id = $8, status = $9, updated_at = $10
		 WHERE id = $11 AND tenant_id = $12`,
		employee.FirstName, employee.LastName, employee.Email, employee.PhoneNumber, employee.Position,
		employee.Department, employee.HireDate, employee.ManagerID, employee.Status, employee.UpdatedAt,
		employee.ID, employee.TenantID)

	if err != nil {
		return fmt.Errorf("failed to update employee: %w", err)
	}
	return nil
}

// DeleteEmployee 删除员工
func (r *Repository) DeleteEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM corehr.employees WHERE id = $1 AND tenant_id = $2`,
		employeeID, tenantID)

	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("employee not found or not authorized")
	}

	return nil
}

// GetManagerByEmployeeID 根据员工ID获取经理信息
func (r *Repository) GetManagerByEmployeeID(ctx context.Context, tenantID, employeeID uuid.UUID) (*Employee, error) {
	var manager Employee
	err := r.db.QueryRow(ctx,
		`SELECT e.id, e.tenant_id, e.employee_number, e.first_name, e.last_name, e.email, e.phone_number, e.position, e.department, e.hire_date, e.manager_id, e.status, e.created_at, e.updated_at 
		 FROM corehr.employees e
		 INNER JOIN corehr.employees emp ON emp.manager_id = e.id
		 WHERE emp.id = $1 AND emp.tenant_id = $2 AND e.tenant_id = $2`,
		employeeID, tenantID).Scan(
		&manager.ID, &manager.TenantID, &manager.EmployeeNumber, &manager.FirstName, &manager.LastName,
		&manager.Email, &manager.PhoneNumber, &manager.Position, &manager.Department, &manager.HireDate,
		&manager.ManagerID, &manager.Status, &manager.CreatedAt, &manager.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get manager: %w", err)
	}
	return &manager, nil
}

// Organization Repository Methods

// GetOrganizationByID 根据ID获取组织
func (r *Repository) GetOrganizationByID(ctx context.Context, tenantID, orgID uuid.UUID) (*Organization, error) {
	var org Organization
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, name, code, parent_id, level, created_at, updated_at 
		 FROM corehr.organizations WHERE id = $1 AND tenant_id = $2`,
		orgID, tenantID).Scan(
		&org.ID, &org.TenantID, &org.Name, &org.Code, &org.ParentID, &org.Level, &org.CreatedAt, &org.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	return &org, nil
}

// ListOrganizations 获取组织列表
func (r *Repository) ListOrganizations(ctx context.Context, tenantID uuid.UUID) ([]Organization, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, tenant_id, name, code, parent_id, level, created_at, updated_at 
		 FROM corehr.organizations 
		 WHERE tenant_id = $1
		 ORDER BY level, name`,
		tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations: %w", err)
	}
	defer rows.Close()

	var organizations []Organization
	for rows.Next() {
		var org Organization
		err := rows.Scan(&org.ID, &org.TenantID, &org.Name, &org.Code, &org.ParentID, &org.Level, &org.CreatedAt, &org.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		organizations = append(organizations, org)
	}

	return organizations, nil
}

// GetOrganizationTree 获取组织树
func (r *Repository) GetOrganizationTree(ctx context.Context, tenantID uuid.UUID) ([]OrganizationTree, error) {
	query := `
		WITH RECURSIVE org_tree AS (
			SELECT id, tenant_id, name, code, parent_id, level, created_at, updated_at, 0 as depth
			FROM corehr.organizations 
			WHERE tenant_id = $1 AND parent_id IS NULL
			UNION ALL
			SELECT o.id, o.tenant_id, o.name, o.code, o.parent_id, o.level, o.created_at, o.updated_at, ot.depth + 1
			FROM corehr.organizations o
			JOIN org_tree ot ON o.parent_id = ot.id
			WHERE o.tenant_id = $1
		)
		SELECT id, tenant_id, name, code, parent_id, level, created_at, updated_at, depth
		FROM org_tree 
		ORDER BY depth, level, name
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query organization tree: %w", err)
	}
	defer rows.Close()

	// 构建组织树
	orgMap := make(map[uuid.UUID]*OrganizationTree)
	var rootOrgs []*OrganizationTree

	for rows.Next() {
		var org OrganizationTree
		var depth int
		err := rows.Scan(&org.ID, &org.TenantID, &org.Name, &org.Code, &org.ParentID, &org.Level, &org.CreatedAt, &org.UpdatedAt, &depth)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization tree: %w", err)
		}

		orgMap[org.ID] = &org

		if org.ParentID == nil {
			rootOrgs = append(rootOrgs, &org)
		} else {
			if parent, exists := orgMap[*org.ParentID]; exists {
				parent.Children = append(parent.Children, org)
			}
		}
	}

	// 转换为切片
	result := make([]OrganizationTree, len(rootOrgs))
	for i, org := range rootOrgs {
		result[i] = *org
	}

	return result, nil
}

// CreateOrganization 创建组织
func (r *Repository) CreateOrganization(ctx context.Context, org *Organization) error {
	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO corehr.organizations (id, tenant_id, name, code, parent_id, level, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		org.ID, org.TenantID, org.Name, org.Code, org.ParentID, org.Level, org.CreatedAt, org.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}
	return nil
}

// UpdateOrganization 更新组织
func (r *Repository) UpdateOrganization(ctx context.Context, org *Organization) error {
	org.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`UPDATE corehr.organizations 
		 SET name = $1, code = $2, parent_id = $3, level = $4, updated_at = $5
		 WHERE id = $6 AND tenant_id = $7`,
		org.Name, org.Code, org.ParentID, org.Level, org.UpdatedAt, org.ID, org.TenantID)

	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}
	return nil
}

// DeleteOrganization 删除组织
func (r *Repository) DeleteOrganization(ctx context.Context, tenantID, orgID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM corehr.organizations WHERE id = $1 AND tenant_id = $2`,
		orgID, tenantID)

	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("organization not found or not authorized")
	}

	return nil
}

// Position Repository Methods

// GetPositionByID 根据ID获取职位
func (r *Repository) GetPositionByID(ctx context.Context, tenantID, positionID uuid.UUID) (*Position, error) {
	var position Position
	err := r.db.QueryRow(ctx,
		`SELECT id, tenant_id, title, code, department_id, level, created_at, updated_at 
		 FROM corehr.positions WHERE id = $1 AND tenant_id = $2`,
		positionID, tenantID).Scan(
		&position.ID, &position.TenantID, &position.Title, &position.Code, &position.DepartmentID,
		&position.Level, &position.CreatedAt, &position.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %w", err)
	}
	return &position, nil
}

// ListPositions 获取职位列表
func (r *Repository) ListPositions(ctx context.Context, tenantID uuid.UUID) ([]Position, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, tenant_id, title, code, department_id, level, created_at, updated_at 
		 FROM corehr.positions 
		 WHERE tenant_id = $1
		 ORDER BY level, title`,
		tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	var positions []Position
	for rows.Next() {
		var pos Position
		err := rows.Scan(&pos.ID, &pos.TenantID, &pos.Title, &pos.Code, &pos.DepartmentID, &pos.Level, &pos.CreatedAt, &pos.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan position: %w", err)
		}
		positions = append(positions, pos)
	}

	return positions, nil
}

// CreatePosition 创建职位
func (r *Repository) CreatePosition(ctx context.Context, position *Position) error {
	now := time.Now()
	position.CreatedAt = now
	position.UpdatedAt = now

	_, err := r.db.Exec(ctx,
		`INSERT INTO corehr.positions (id, tenant_id, title, code, department_id, level, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		position.ID, position.TenantID, position.Title, position.Code, position.DepartmentID,
		position.Level, position.CreatedAt, position.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}
	return nil
}

// UpdatePosition 更新职位
func (r *Repository) UpdatePosition(ctx context.Context, position *Position) error {
	position.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx,
		`UPDATE corehr.positions 
		 SET title = $1, code = $2, department_id = $3, level = $4, updated_at = $5
		 WHERE id = $6 AND tenant_id = $7`,
		position.Title, position.Code, position.DepartmentID, position.Level, position.UpdatedAt,
		position.ID, position.TenantID)

	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}
	return nil
}

// DeletePosition 删除职位
func (r *Repository) DeletePosition(ctx context.Context, tenantID, positionID uuid.UUID) error {
	result, err := r.db.Exec(ctx,
		`DELETE FROM corehr.positions WHERE id = $1 AND tenant_id = $2`,
		positionID, tenantID)

	if err != nil {
		return fmt.Errorf("failed to delete position: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("position not found or not authorized")
	}

	return nil
}
