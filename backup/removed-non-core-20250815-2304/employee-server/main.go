// 员工管理API服务器 - 8位编码高性能版
// 版本: v1.0 Optimized
// 创建日期: 2025-08-05
// 基于: 职位管理7位编码成功经验
// 架构: 8位编码零转换直接查询

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
)

// 8位编码员工结构
type Employee struct {
	Code                 string    `json:"code" db:"code"`
	OrganizationCode     string    `json:"organization_code" db:"organization_code"`
	PrimaryPositionCode  *string   `json:"primary_position_code,omitempty" db:"primary_position_code"`
	EmployeeType         string    `json:"employee_type" db:"employee_type"`
	EmploymentStatus     string    `json:"employment_status" db:"employment_status"`
	FirstName            string    `json:"first_name" db:"first_name"`
	LastName             string    `json:"last_name" db:"last_name"`
	Email                string    `json:"email" db:"email"`
	PersonalEmail        *string   `json:"personal_email,omitempty" db:"personal_email"`
	PhoneNumber          *string   `json:"phone_number,omitempty" db:"phone_number"`
	HireDate             string    `json:"hire_date" db:"hire_date"`
	TerminationDate      *string   `json:"termination_date,omitempty" db:"termination_date"`
	PersonalInfo         *string   `json:"personal_info,omitempty" db:"personal_info"`
	EmployeeDetails      *string   `json:"employee_details,omitempty" db:"employee_details"`
	TenantID             string    `json:"tenant_id" db:"tenant_id"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// 关联查询结果
type EmployeeWithRelations struct {
	Employee
	Organization    *OrganizationInfo `json:"organization,omitempty"`
	PrimaryPosition *PositionInfo     `json:"primary_position,omitempty"`
	AllPositions    []PositionAssignment `json:"all_positions,omitempty"`
	Manager         *EmployeeInfo     `json:"manager,omitempty"`
	DirectReports   []EmployeeInfo    `json:"direct_reports,omitempty"`
}

type OrganizationInfo struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	UnitType string `json:"unit_type"`
}

type PositionInfo struct {
	Code         string `json:"code"`
	PositionType string `json:"position_type"`
	Status       string `json:"status"`
	Details      string `json:"details,omitempty"`
}

type PositionAssignment struct {
	PositionCode   string  `json:"position_code"`
	AssignmentType string  `json:"assignment_type"`
	Status         string  `json:"status"`
	StartDate      string  `json:"start_date"`
	EndDate        *string `json:"end_date,omitempty"`
}

type EmployeeInfo struct {
	Code      string `json:"code"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	EmployeeType string `json:"employee_type"`
}

type EmployeeListResponse struct {
	Employees  []Employee `json:"employees"`
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// 员工统计
type EmployeeStats struct {
	TotalEmployees  int                `json:"total_employees"`
	ActiveEmployees int                `json:"active_employees"`
	RecentHires     int                `json:"recent_hires_30days"`
	ByType          map[string]int     `json:"by_type"`
	ByStatus        map[string]int     `json:"by_status"`
	ByOrganization  map[string]int     `json:"by_organization"`
}

// 员工管理处理器
type EmployeeHandler struct {
	db       *sql.DB
	tenantID string
}

func NewEmployeeHandler(db *sql.DB, tenantID string) *EmployeeHandler {
	return &EmployeeHandler{db: db, tenantID: tenantID}
}

// 8位编码验证
func validateEmployeeCode(code string) error {
	if len(code) != 8 {
		return fmt.Errorf("employee code must be exactly 8 digits")
	}
	if _, err := strconv.Atoi(code); err != nil {
		return fmt.Errorf("employee code must be numeric")
	}
	codeInt, _ := strconv.Atoi(code)
	if codeInt < 10000000 || codeInt > 99999999 {
		return fmt.Errorf("employee code must be in range 10000000-99999999")
	}
	return nil
}

// 7位组织编码验证
func validateOrganizationCode(code string) error {
	if len(code) != 7 {
		return fmt.Errorf("organization code must be exactly 7 digits")
	}
	if _, err := strconv.Atoi(code); err != nil {
		return fmt.Errorf("organization code must be numeric")
	}
	codeInt, _ := strconv.Atoi(code)
	if codeInt < 1000000 || codeInt > 9999999 {
		return fmt.Errorf("organization code must be in range 1000000-9999999")
	}
	return nil
}

// 7位职位编码验证
func validatePositionCode(code string) error {
	if len(code) != 7 {
		return fmt.Errorf("position code must be exactly 7 digits")
	}
	if _, err := strconv.Atoi(code); err != nil {
		return fmt.Errorf("position code must be numeric")
	}
	codeInt, _ := strconv.Atoi(code)
	if codeInt < 1000000 || codeInt > 9999999 {
		return fmt.Errorf("position code must be in range 1000000-9999999")
	}
	return nil
}

// 创建员工 - 自动生成8位编码
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrganizationCode    string                 `json:"organization_code"`
		PrimaryPositionCode *string                `json:"primary_position_code,omitempty"`
		EmployeeType        string                 `json:"employee_type"`
		EmploymentStatus    string                 `json:"employment_status"`
		FirstName           string                 `json:"first_name"`
		LastName            string                 `json:"last_name"`
		Email               string                 `json:"email"`
		PersonalEmail       *string                `json:"personal_email,omitempty"`
		PhoneNumber         *string                `json:"phone_number,omitempty"`
		HireDate            string                 `json:"hire_date"`
		PersonalInfo        map[string]interface{} `json:"personal_info,omitempty"`
		EmployeeDetails     map[string]interface{} `json:"employee_details,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 验证组织编码
	if err := validateOrganizationCode(req.OrganizationCode); err != nil {
		http.Error(w, fmt.Sprintf("Invalid organization code: %v", err), http.StatusBadRequest)
		return
	}

	// 验证职位编码
	if req.PrimaryPositionCode != nil {
		if err := validatePositionCode(*req.PrimaryPositionCode); err != nil {
			http.Error(w, fmt.Sprintf("Invalid position code: %v", err), http.StatusBadRequest)
			return
		}
	}

	// 验证员工类型
	validTypes := []string{"FULL_TIME", "PART_TIME", "CONTRACTOR", "INTERN"}
	if !contains(validTypes, req.EmployeeType) {
		http.Error(w, "Invalid employee type", http.StatusBadRequest)
		return
	}

	// 验证就业状态
	if req.EmploymentStatus == "" {
		req.EmploymentStatus = "ACTIVE"
	}
	validStatuses := []string{"ACTIVE", "TERMINATED", "ON_LEAVE", "PENDING_START"}
	if !contains(validStatuses, req.EmploymentStatus) {
		http.Error(w, "Invalid employment status", http.StatusBadRequest)
		return
	}

	// 验证必填字段
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.HireDate == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// 准备JSON字段
	var personalInfoJSON, employeeDetailsJSON *string
	if req.PersonalInfo != nil {
		info, _ := json.Marshal(req.PersonalInfo)
		infoStr := string(info)
		personalInfoJSON = &infoStr
	}
	if req.EmployeeDetails != nil {
		details, _ := json.Marshal(req.EmployeeDetails)
		detailsStr := string(details)
		employeeDetailsJSON = &detailsStr
	}

	// 插入员工（自动生成8位编码）
	var employee Employee
	query := `
		INSERT INTO employees (
			organization_code, primary_position_code, employee_type, employment_status,
			first_name, last_name, email, personal_email, phone_number, hire_date,
			personal_info, employee_details, tenant_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING code, organization_code, primary_position_code, employee_type, employment_status,
				  first_name, last_name, email, personal_email, phone_number, hire_date,
				  termination_date, personal_info, employee_details, tenant_id,
				  created_at, updated_at`

	err := h.db.QueryRow(query,
		req.OrganizationCode, req.PrimaryPositionCode, req.EmployeeType, req.EmploymentStatus,
		req.FirstName, req.LastName, req.Email, req.PersonalEmail, req.PhoneNumber, req.HireDate,
		personalInfoJSON, employeeDetailsJSON, h.tenantID,
	).Scan(
		&employee.Code, &employee.OrganizationCode, &employee.PrimaryPositionCode,
		&employee.EmployeeType, &employee.EmploymentStatus,
		&employee.FirstName, &employee.LastName, &employee.Email,
		&employee.PersonalEmail, &employee.PhoneNumber, &employee.HireDate,
		&employee.TerminationDate, &employee.PersonalInfo, &employee.EmployeeDetails,
		&employee.TenantID, &employee.CreatedAt, &employee.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error creating employee: %v", err)
		if strings.Contains(err.Error(), "foreign key constraint") {
			if strings.Contains(err.Error(), "organization") {
				http.Error(w, "Organization not found", http.StatusBadRequest)
			} else if strings.Contains(err.Error(), "position") {
				http.Error(w, "Position not found", http.StatusBadRequest)
			} else {
				http.Error(w, "Invalid reference", http.StatusBadRequest)
			}
		} else if strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "Email already exists", http.StatusConflict)
		} else {
			http.Error(w, "Failed to create employee", http.StatusInternalServerError)
		}
		return
	}

	// 如果设置了主要职位，自动创建职位关联
	if req.PrimaryPositionCode != nil {
		_, err = h.db.Exec(`
			INSERT INTO employee_positions (employee_code, position_code, assignment_type, status, start_date)
			VALUES ($1, $2, 'PRIMARY', 'ACTIVE', $3)`,
			employee.Code, *req.PrimaryPositionCode, req.HireDate)
		if err != nil {
			log.Printf("Warning: Failed to create position assignment for employee %s: %v", employee.Code, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(employee)
}

// 获取员工 - 直接8位编码查询
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	
	if err := validateEmployeeCode(code); err != nil {
		http.Error(w, fmt.Sprintf("Invalid employee code: %v", err), http.StatusBadRequest)
		return
	}

	// 检查关联查询参数
	withOrg := r.URL.Query().Get("with_organization") == "true"
	withPosition := r.URL.Query().Get("with_position") == "true"
	withAllPositions := r.URL.Query().Get("with_all_positions") == "true"
	withManager := r.URL.Query().Get("with_manager") == "true"
	withReports := r.URL.Query().Get("with_direct_reports") == "true"

	// 基础员工查询 - 直接8位编码主键查询
	var employee Employee
	query := `
		SELECT code, organization_code, primary_position_code, employee_type, employment_status,
			   first_name, last_name, email, personal_email, phone_number, hire_date,
			   termination_date, personal_info, employee_details, tenant_id,
			   created_at, updated_at
		FROM employees 
		WHERE code = $1 AND tenant_id = $2`

	err := h.db.QueryRow(query, code, h.tenantID).Scan(
		&employee.Code, &employee.OrganizationCode, &employee.PrimaryPositionCode,
		&employee.EmployeeType, &employee.EmploymentStatus,
		&employee.FirstName, &employee.LastName, &employee.Email,
		&employee.PersonalEmail, &employee.PhoneNumber, &employee.HireDate,
		&employee.TerminationDate, &employee.PersonalInfo, &employee.EmployeeDetails,
		&employee.TenantID, &employee.CreatedAt, &employee.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Employee not found", http.StatusNotFound)
			return
		}
		log.Printf("Error fetching employee: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	result := EmployeeWithRelations{Employee: employee}

	// 关联查询
	if withOrg {
		result.Organization = h.getOrganizationInfo(employee.OrganizationCode)
	}
	if withPosition && employee.PrimaryPositionCode != nil {
		result.PrimaryPosition = h.getPositionInfo(*employee.PrimaryPositionCode)
	}
	if withAllPositions {
		result.AllPositions = h.getAllPositions(employee.Code)
	}
	if withManager {
		result.Manager = h.getManager(employee.Code)
	}
	if withReports {
		result.DirectReports = h.getDirectReports(employee.Code)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// 更新员工
func (h *EmployeeHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	
	if err := validateEmployeeCode(code); err != nil {
		http.Error(w, fmt.Sprintf("Invalid employee code: %v", err), http.StatusBadRequest)
		return
	}

	var req struct {
		OrganizationCode    *string                `json:"organization_code,omitempty"`
		PrimaryPositionCode *string                `json:"primary_position_code,omitempty"`
		EmploymentStatus    *string                `json:"employment_status,omitempty"`
		Email               *string                `json:"email,omitempty"`
		PersonalEmail       *string                `json:"personal_email,omitempty"`
		PhoneNumber         *string                `json:"phone_number,omitempty"`
		TerminationDate     *string                `json:"termination_date,omitempty"`
		PersonalInfo        map[string]interface{} `json:"personal_info,omitempty"`
		EmployeeDetails     map[string]interface{} `json:"employee_details,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 构建动态更新查询
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.OrganizationCode != nil {
		if err := validateOrganizationCode(*req.OrganizationCode); err != nil {
			http.Error(w, fmt.Sprintf("Invalid organization code: %v", err), http.StatusBadRequest)
			return
		}
		setParts = append(setParts, fmt.Sprintf("organization_code = $%d", argIndex))
		args = append(args, *req.OrganizationCode)
		argIndex++
	}

	if req.PrimaryPositionCode != nil {
		if err := validatePositionCode(*req.PrimaryPositionCode); err != nil {
			http.Error(w, fmt.Sprintf("Invalid position code: %v", err), http.StatusBadRequest)
			return
		}
		setParts = append(setParts, fmt.Sprintf("primary_position_code = $%d", argIndex))
		args = append(args, *req.PrimaryPositionCode)
		argIndex++
	}

	if req.EmploymentStatus != nil {
		validStatuses := []string{"ACTIVE", "TERMINATED", "ON_LEAVE", "PENDING_START"}
		if !contains(validStatuses, *req.EmploymentStatus) {
			http.Error(w, "Invalid employment status", http.StatusBadRequest)
			return
		}
		setParts = append(setParts, fmt.Sprintf("employment_status = $%d", argIndex))
		args = append(args, *req.EmploymentStatus)
		argIndex++
	}

	if req.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *req.Email)
		argIndex++
	}

	if req.PersonalEmail != nil {
		setParts = append(setParts, fmt.Sprintf("personal_email = $%d", argIndex))
		args = append(args, *req.PersonalEmail)
		argIndex++
	}

	if req.PhoneNumber != nil {
		setParts = append(setParts, fmt.Sprintf("phone_number = $%d", argIndex))
		args = append(args, *req.PhoneNumber)
		argIndex++
	}

	if req.TerminationDate != nil {
		setParts = append(setParts, fmt.Sprintf("termination_date = $%d", argIndex))
		args = append(args, *req.TerminationDate)
		argIndex++
	}

	if req.PersonalInfo != nil {
		info, _ := json.Marshal(req.PersonalInfo)
		setParts = append(setParts, fmt.Sprintf("personal_info = $%d", argIndex))
		args = append(args, string(info))
		argIndex++
	}

	if req.EmployeeDetails != nil {
		details, _ := json.Marshal(req.EmployeeDetails)
		setParts = append(setParts, fmt.Sprintf("employee_details = $%d", argIndex))
		args = append(args, string(details))
		argIndex++
	}

	if len(setParts) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// 添加updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// 添加WHERE条件参数
	args = append(args, code, h.tenantID)

	query := fmt.Sprintf(`
		UPDATE employees 
		SET %s
		WHERE code = $%d AND tenant_id = $%d
		RETURNING code, organization_code, primary_position_code, employee_type, employment_status,
				  first_name, last_name, email, personal_email, phone_number, hire_date,
				  termination_date, personal_info, employee_details, tenant_id,
				  created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	var employee Employee
	err := h.db.QueryRow(query, args...).Scan(
		&employee.Code, &employee.OrganizationCode, &employee.PrimaryPositionCode,
		&employee.EmployeeType, &employee.EmploymentStatus,
		&employee.FirstName, &employee.LastName, &employee.Email,
		&employee.PersonalEmail, &employee.PhoneNumber, &employee.HireDate,
		&employee.TerminationDate, &employee.PersonalInfo, &employee.EmployeeDetails,
		&employee.TenantID, &employee.CreatedAt, &employee.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Employee not found", http.StatusNotFound)
			return
		}
		log.Printf("Error updating employee: %v", err)
		if strings.Contains(err.Error(), "foreign key constraint") {
			http.Error(w, "Invalid reference", http.StatusBadRequest)
		} else if strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "Email already exists", http.StatusConflict)
		} else {
			http.Error(w, "Failed to update employee", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employee)
}

// 删除员工
func (h *EmployeeHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	
	if err := validateEmployeeCode(code); err != nil {
		http.Error(w, fmt.Sprintf("Invalid employee code: %v", err), http.StatusBadRequest)
		return
	}

	// 检查约束条件
	var hasActivePositions int
	h.db.QueryRow("SELECT COUNT(*) FROM employee_positions WHERE employee_code = $1 AND status = 'ACTIVE'", code).Scan(&hasActivePositions)
	if hasActivePositions > 0 {
		http.Error(w, "Cannot delete employee with active position assignments", http.StatusConflict)
		return
	}

	// 删除员工（级联删除职位关联）
	result, err := h.db.Exec("DELETE FROM employees WHERE code = $1 AND tenant_id = $2", code, h.tenantID)
	if err != nil {
		log.Printf("Error deleting employee: %v", err)
		http.Error(w, "Failed to delete employee", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// 员工列表查询 - 高性能分页
func (h *EmployeeHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 过滤参数
	employeeType := r.URL.Query().Get("employee_type")
	status := r.URL.Query().Get("employment_status")
	organizationCode := r.URL.Query().Get("organization_code")

	// 构建WHERE条件
	whereConditions := []string{"tenant_id = $1"}
	args := []interface{}{h.tenantID}
	argIndex := 2

	if employeeType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("employee_type = $%d", argIndex))
		args = append(args, employeeType)
		argIndex++
	}
	if status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("employment_status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}
	if organizationCode != "" {
		if err := validateOrganizationCode(organizationCode); err != nil {
			http.Error(w, fmt.Sprintf("Invalid organization code: %v", err), http.StatusBadRequest)
			return
		}
		whereConditions = append(whereConditions, fmt.Sprintf("organization_code = $%d", argIndex))
		args = append(args, organizationCode)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM employees WHERE %s", whereClause)
	var total int
	h.db.QueryRow(countQuery, args...).Scan(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)

	query := fmt.Sprintf(`
		SELECT code, organization_code, primary_position_code, employee_type, employment_status,
			   first_name, last_name, email, personal_email, phone_number, hire_date,
			   termination_date, personal_info, employee_details, tenant_id,
			   created_at, updated_at
		FROM employees 
		WHERE %s 
		ORDER BY code 
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		log.Printf("Error listing employees: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		var emp Employee
		err := rows.Scan(
			&emp.Code, &emp.OrganizationCode, &emp.PrimaryPositionCode,
			&emp.EmployeeType, &emp.EmploymentStatus,
			&emp.FirstName, &emp.LastName, &emp.Email,
			&emp.PersonalEmail, &emp.PhoneNumber, &emp.HireDate,
			&emp.TerminationDate, &emp.PersonalInfo, &emp.EmployeeDetails,
			&emp.TenantID, &emp.CreatedAt, &emp.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning employee: %v", err)
			continue
		}
		employees = append(employees, emp)
	}

	totalPages := (total + pageSize - 1) / pageSize
	response := EmployeeListResponse{
		Employees: employees,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 员工统计
func (h *EmployeeHandler) GetEmployeeStats(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT 
			COUNT(*) as total_employees,
			COUNT(CASE WHEN employment_status = 'ACTIVE' THEN 1 END) as active_count,
			COUNT(CASE WHEN hire_date >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as recent_hires,
			COUNT(CASE WHEN employee_type = 'FULL_TIME' THEN 1 END) as full_time_count,
			COUNT(CASE WHEN employee_type = 'PART_TIME' THEN 1 END) as part_time_count,
			COUNT(CASE WHEN employee_type = 'CONTRACTOR' THEN 1 END) as contractor_count,
			COUNT(CASE WHEN employee_type = 'INTERN' THEN 1 END) as intern_count,
			COUNT(CASE WHEN employment_status = 'ACTIVE' THEN 1 END) as status_active_count,
			COUNT(CASE WHEN employment_status = 'TERMINATED' THEN 1 END) as status_terminated_count,
			COUNT(CASE WHEN employment_status = 'ON_LEAVE' THEN 1 END) as status_leave_count,
			COUNT(CASE WHEN employment_status = 'PENDING_START' THEN 1 END) as status_pending_count
		FROM employees 
		WHERE tenant_id = $1`

	var totalEmployees, activeEmployees, recentHires int
	var fullTimeCount, partTimeCount, contractorCount, internCount int
	var statusActiveCount, statusTerminatedCount, statusLeaveCount, statusPendingCount int

	err := h.db.QueryRow(query, h.tenantID).Scan(
		&totalEmployees, &activeEmployees, &recentHires,
		&fullTimeCount, &partTimeCount, &contractorCount, &internCount,
		&statusActiveCount, &statusTerminatedCount, &statusLeaveCount, &statusPendingCount,
	)

	if err != nil {
		log.Printf("Error getting employee stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 按组织统计
	orgQuery := `
		SELECT o.name, COUNT(e.code) 
		FROM employees e
		JOIN organization_units o ON e.organization_code = o.code
		WHERE e.tenant_id = $1
		GROUP BY o.code, o.name
		ORDER BY COUNT(e.code) DESC`

	orgRows, err := h.db.Query(orgQuery, h.tenantID)
	byOrganization := make(map[string]int)
	if err == nil {
		defer orgRows.Close()
		for orgRows.Next() {
			var orgName string
			var count int
			if err := orgRows.Scan(&orgName, &count); err == nil {
				byOrganization[orgName] = count
			}
		}
	}

	stats := EmployeeStats{
		TotalEmployees:  totalEmployees,
		ActiveEmployees: activeEmployees,
		RecentHires:     recentHires,
		ByType: map[string]int{
			"FULL_TIME":  fullTimeCount,
			"PART_TIME":  partTimeCount,
			"CONTRACTOR": contractorCount,
			"INTERN":     internCount,
		},
		ByStatus: map[string]int{
			"ACTIVE":        statusActiveCount,
			"TERMINATED":    statusTerminatedCount,
			"ON_LEAVE":      statusLeaveCount,
			"PENDING_START": statusPendingCount,
		},
		ByOrganization: byOrganization,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// 辅助方法
func (h *EmployeeHandler) getOrganizationInfo(code string) *OrganizationInfo {
	var org OrganizationInfo
	query := `SELECT code, name, unit_type FROM organization_units WHERE code = $1`
	err := h.db.QueryRow(query, code).Scan(&org.Code, &org.Name, &org.UnitType)
	if err != nil {
		return nil
	}
	return &org
}

func (h *EmployeeHandler) getPositionInfo(code string) *PositionInfo {
	var pos PositionInfo
	query := `SELECT code, position_type, status, COALESCE(details, '{}') FROM positions WHERE code = $1`
	err := h.db.QueryRow(query, code).Scan(&pos.Code, &pos.PositionType, &pos.Status, &pos.Details)
	if err != nil {
		return nil
	}
	return &pos
}

func (h *EmployeeHandler) getAllPositions(employeeCode string) []PositionAssignment {
	query := `
		SELECT position_code, assignment_type, status, start_date, end_date
		FROM employee_positions 
		WHERE employee_code = $1 
		ORDER BY start_date DESC`
	
	rows, err := h.db.Query(query, employeeCode)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var positions []PositionAssignment
	for rows.Next() {
		var pos PositionAssignment
		if err := rows.Scan(&pos.PositionCode, &pos.AssignmentType, &pos.Status, &pos.StartDate, &pos.EndDate); err == nil {
			positions = append(positions, pos)
		}
	}
	return positions
}

func (h *EmployeeHandler) getManager(employeeCode string) *EmployeeInfo {
	// 通过员工的主要职位找到管理者
	query := `
		SELECT e.code, e.first_name, e.last_name, e.email, e.employee_type
		FROM employees e
		JOIN positions p1 ON e.primary_position_code = p1.code
		JOIN positions p2 ON p1.code = p2.manager_position_code
		JOIN employees e2 ON e2.primary_position_code = p2.code
		WHERE e2.code = $1 AND e.tenant_id = $2
		LIMIT 1`
	
	var mgr EmployeeInfo
	err := h.db.QueryRow(query, employeeCode, h.tenantID).Scan(
		&mgr.Code, &mgr.FirstName, &mgr.LastName, &mgr.Email, &mgr.EmployeeType)
	if err != nil {
		return nil
	}
	return &mgr
}

func (h *EmployeeHandler) getDirectReports(employeeCode string) []EmployeeInfo {
	// 通过职位管理关系找到直接下属
	query := `
		SELECT e.code, e.first_name, e.last_name, e.email, e.employee_type
		FROM employees e
		JOIN positions p ON e.primary_position_code = p.code
		JOIN positions mgr_p ON p.manager_position_code = mgr_p.code
		JOIN employees mgr_e ON mgr_e.primary_position_code = mgr_p.code
		WHERE mgr_e.code = $1 AND e.tenant_id = $2`
	
	rows, err := h.db.Query(query, employeeCode, h.tenantID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var reports []EmployeeInfo
	for rows.Next() {
		var emp EmployeeInfo
		if err := rows.Scan(&emp.Code, &emp.FirstName, &emp.LastName, &emp.Email, &emp.EmployeeType); err == nil {
			reports = append(reports, emp)
		}
	}
	return reports
}

// 工具函数
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 健康检查
func healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "employee-management-api",
		"version":   "v1.0-8digit-optimized",
		"features": []string{
			"8-digit employee codes",
			"zero-conversion architecture",
			"direct primary key queries",
			"high-performance indexing",
			"employee-position-organization relations",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 数据库连接
	dbURL := "host=localhost port=5432 user=user password=password dbname=cubecastle sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// 租户ID（实际应用中应该从认证中获取）
	tenantID := "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	
	handler := NewEmployeeHandler(db, tenantID)

	// 路由设置
	r := chi.NewRouter()
	
	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API路由
	r.Route("/api/v1/employees", func(r chi.Router) {
		r.Post("/", handler.CreateEmployee)
		r.Get("/", handler.ListEmployees)
		r.Get("/stats", handler.GetEmployeeStats)
		r.Get("/{code}", handler.GetEmployee)
		r.Put("/{code}", handler.UpdateEmployee)
		r.Delete("/{code}", handler.DeleteEmployee)
	})

	// 健康检查
	r.Get("/health", healthCheck)

	// 启动信息
	fmt.Println("🚀 Employee Management API Server v1.0 (8-digit optimized)")
	fmt.Println("⚡ Based on proven 7-digit organization/position success")
	fmt.Println("📊 Server running on http://localhost:8084")
	fmt.Println("🔧 Health check: http://localhost:8084/health")
	fmt.Println("📋 API Base: http://localhost:8084/api/v1/employees")
	fmt.Println("🎯 Features: 8-digit codes, Zero-conversion architecture, Employee-Position-Organization relations")
	
	log.Fatal(http.ListenAndServe(":8084", r))
}