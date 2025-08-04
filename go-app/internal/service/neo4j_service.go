// internal/service/neo4j_service.go
package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jService provides graph database operations for organizational relationships
type Neo4jService struct {
	driver neo4j.DriverWithContext
	logger *log.Logger
}

// Neo4jConfig holds configuration for Neo4j connection
type Neo4jConfig struct {
	URI      string `json:"uri" yaml:"uri"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Database string `json:"database" yaml:"database"`
}

// OrganizationNode represents an organization unit in the graph
type OrganizationNode struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	UnitType     string                 `json:"unit_type"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Status       string                 `json:"status"`
	IsActive     bool                   `json:"is_active"`
	ParentUnitID string                 `json:"parent_unit_id"`
	Properties   map[string]interface{} `json:"properties"`
}

// Employee node representation in graph
type EmployeeNode struct {
	ID         string                 `json:"id"`
	EmployeeID string                 `json:"employee_id"`
	LegalName  string                 `json:"legal_name"`
	Email      string                 `json:"email"`
	Status     string                 `json:"status"`
	HireDate   time.Time              `json:"hire_date"`
	Properties map[string]interface{} `json:"properties"`
}

// Position node representation in graph
type PositionNode struct {
	ID            string                 `json:"id"`
	PositionTitle string                 `json:"position_title"`
	Department    string                 `json:"department"`
	JobLevel      string                 `json:"job_level"`
	Location      string                 `json:"location"`
	EffectiveDate time.Time              `json:"effective_date"`
	EndDate       *time.Time             `json:"end_date"`
	Properties    map[string]interface{} `json:"properties"`
}

// Department node representation in graph
type DepartmentNode struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	ParentID   *string                `json:"parent_id"`
	ManagerID  *string                `json:"manager_id"`
	Properties map[string]interface{} `json:"properties"`
}

// OrganizationalPath represents a path between employees
type OrganizationalPath struct {
	FromEmployee EmployeeNode  `json:"from_employee"`
	ToEmployee   EmployeeNode  `json:"to_employee"`
	Path         []PathSegment `json:"path"`
	Distance     int           `json:"distance"`
	PathType     string        `json:"path_type"` // REPORTS_TO, PEER, CROSS_DEPARTMENT
}

// PathSegment represents one step in an organizational path
type PathSegment struct {
	Employee     EmployeeNode `json:"employee"`
	Position     PositionNode `json:"position"`
	Relationship string       `json:"relationship"` // REPORTS_TO, MANAGES, WORKS_WITH
}

// ReportingHierarchy represents the reporting structure
type ReportingHierarchy struct {
	Manager       EmployeeNode   `json:"manager"`
	DirectReports []EmployeeNode `json:"direct_reports"`
	AllReports    []EmployeeNode `json:"all_reports"`
	Depth         int            `json:"depth"`
}

// NewNeo4jService creates a new Neo4j service instance
func NewNeo4jService(config Neo4jConfig, logger *log.Logger) (*Neo4jService, error) {
	driver, err := neo4j.NewDriverWithContext(
		config.URI,
		neo4j.BasicAuth(config.Username, config.Password, ""),
		func(c *neo4j.Config) {
			c.MaxConnectionLifetime = 5 * time.Minute  // 减少连接生命周期
			c.MaxConnectionPoolSize = 10              // 减少连接池大小
			c.ConnectionAcquisitionTimeout = 30 * time.Second  // 减少获取超时
			c.SocketConnectTimeout = 15 * time.Second  // 添加Socket连接超时
			c.SocketKeepalive = true                   // 启用keepalive
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	service := &Neo4jService{
		driver: driver,
		logger: logger,
	}

	// Verify connection and create constraints
	if err := service.initializeSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize Neo4j schema: %w", err)
	}

	return service, nil
}

// Close closes the Neo4j driver connection
func (s *Neo4jService) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

// initializeSchema creates necessary constraints and indexes
func (s *Neo4jService) initializeSchema(ctx context.Context) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	constraints := []string{
		"CREATE CONSTRAINT employee_id_unique IF NOT EXISTS FOR (e:Employee) REQUIRE e.employee_id IS UNIQUE",
		"CREATE CONSTRAINT position_id_unique IF NOT EXISTS FOR (p:Position) REQUIRE p.id IS UNIQUE",
		"CREATE CONSTRAINT department_name_unique IF NOT EXISTS FOR (d:Department) REQUIRE d.name IS UNIQUE",
	}

	indexes := []string{
		"CREATE INDEX employee_legal_name IF NOT EXISTS FOR (e:Employee) ON (e.legal_name)",
		"CREATE INDEX employee_email IF NOT EXISTS FOR (e:Employee) ON (e.email)",
		"CREATE INDEX position_title IF NOT EXISTS FOR (p:Position) ON (p.position_title)",
		"CREATE INDEX position_department IF NOT EXISTS FOR (p:Position) ON (p.department)",
		"CREATE INDEX position_effective_date IF NOT EXISTS FOR (p:Position) ON (p.effective_date)",
		"CREATE INDEX department_parent IF NOT EXISTS FOR (d:Department) ON (d.parent_id)",
	}

	// Create constraints
	for _, constraint := range constraints {
		_, err := session.Run(ctx, constraint, nil)
		if err != nil {
			s.logger.Printf("Warning: Failed to create constraint: %v", err)
		}
	}

	// Create indexes
	for _, index := range indexes {
		_, err := session.Run(ctx, index, nil)
		if err != nil {
			s.logger.Printf("Warning: Failed to create index: %v", err)
		}
	}

	return nil
}

// SyncEmployee creates or updates an employee node in the graph
func (s *Neo4jService) SyncEmployee(ctx context.Context, employee EmployeeNode) error {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
		MERGE (e:Employee {employee_id: $employee_id})
		SET e.id = $id,
		    e.legal_name = $legal_name,
		    e.email = $email,
		    e.employment_status = $status,
		    e.hire_date = datetime($hire_date),
		    e.updated_at = datetime()
		RETURN e
	`

	params := map[string]interface{}{
		"id":          employee.ID,
		"employee_id": employee.EmployeeID,
		"legal_name":  employee.LegalName,
		"email":       employee.Email,
		"status":      employee.Status,
		"hire_date":   employee.HireDate.Format(time.RFC3339),
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to sync employee %s: %w", employee.EmployeeID, err)
	}

	s.logger.Printf("Synced employee %s to Neo4j", employee.EmployeeID)
	return nil
}

// SyncPosition creates or updates a position node and its relationships
func (s *Neo4jService) SyncPosition(ctx context.Context, position PositionNode, employeeID string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	// Create/update position node
	query := `
		MERGE (p:Position {id: $id})
		SET p.position_title = $position_title,
		    p.department = $department,
		    p.job_level = $job_level,
		    p.location = $location,
		    p.effective_date = datetime($effective_date),
		    p.end_date = CASE WHEN $end_date IS NOT NULL THEN datetime($end_date) ELSE NULL END,
		    p.updated_at = datetime()
		WITH p
		
		// Link to employee
		MATCH (e:Employee {employee_id: $employee_id})
		MERGE (e)-[r:HOLDS_POSITION]->(p)
		SET r.created_at = COALESCE(r.created_at, datetime()),
		    r.updated_at = datetime()
		
		// Link to department
		MERGE (d:Department {name: $department})
		SET d.created_at = COALESCE(d.created_at, datetime())
		MERGE (p)-[rd:BELONGS_TO]->(d)
		
		RETURN p, e, d
	`

	endDate := ""
	if position.EndDate != nil {
		endDate = position.EndDate.Format(time.RFC3339)
	}

	params := map[string]interface{}{
		"id":             position.ID,
		"position_title": position.PositionTitle,
		"department":     position.Department,
		"job_level":      position.JobLevel,
		"location":       position.Location,
		"effective_date": position.EffectiveDate.Format(time.RFC3339),
		"end_date":       endDate,
		"employee_id":    employeeID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to sync position %s: %w", position.ID, err)
	}

	s.logger.Printf("Synced position %s for employee %s to Neo4j", position.ID, employeeID)
	return nil
}

// CreateReportingRelationship creates a reporting relationship between employees
func (s *Neo4jService) CreateReportingRelationship(ctx context.Context, managerID, reporteeID string) error {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
		MATCH (manager:Employee {employee_id: $manager_id})
		MATCH (reportee:Employee {employee_id: $reportee_id})
		MERGE (reportee)-[r:REPORTS_TO]->(manager)
		SET r.created_at = COALESCE(r.created_at, datetime()),
		    r.updated_at = datetime()
		RETURN r
	`

	params := map[string]interface{}{
		"manager_id":  managerID,
		"reportee_id": reporteeID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to create reporting relationship %s -> %s: %w", reporteeID, managerID, err)
	}

	s.logger.Printf("Created reporting relationship: %s reports to %s", reporteeID, managerID)
	return nil
}

// FindReportingPath finds the shortest path between two employees through the reporting structure
func (s *Neo4jService) FindReportingPath(ctx context.Context, fromEmployeeID, toEmployeeID string) (*OrganizationalPath, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
		MATCH (from:Employee {employee_id: $from_id})
		MATCH (to:Employee {employee_id: $to_id})
		MATCH path = shortestPath((from)-[*..10]-(to))
		WHERE ALL(rel in relationships(path) WHERE type(rel) IN ['REPORTS_TO', 'MANAGES'])
		RETURN path,
		       length(path) as distance,
		       [node in nodes(path) | {
		           employee_id: node.employee_id,
		           legal_name: node.legal_name,
		           email: node.email
		       }] as employees
		ORDER BY distance
		LIMIT 1
	`

	params := map[string]interface{}{
		"from_id": fromEmployeeID,
		"to_id":   toEmployeeID,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to find reporting path: %w", err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		// Check if no records found by comparing error message
		if err.Error() == "result contains no more records" {
			return nil, fmt.Errorf("no path found between employees %s and %s", fromEmployeeID, toEmployeeID)
		}
		return nil, fmt.Errorf("failed to get path result: %w", err)
	}

	distance, _ := record.Get("distance")
	employees, _ := record.Get("employees")

	orgPath := &OrganizationalPath{
		Distance: int(distance.(int64)),
		PathType: "REPORTS_TO",
		Path:     make([]PathSegment, 0),
	}

	// Convert employees list to path segments
	if empList, ok := employees.([]interface{}); ok {
		for _, emp := range empList {
			if empMap, ok := emp.(map[string]interface{}); ok {
				segment := PathSegment{
					Employee: EmployeeNode{
						EmployeeID: empMap["employee_id"].(string),
						LegalName:  empMap["legal_name"].(string),
						Email:      empMap["email"].(string),
					},
					Relationship: "REPORTS_TO",
				}
				orgPath.Path = append(orgPath.Path, segment)
			}
		}
	}

	if len(orgPath.Path) > 0 {
		orgPath.FromEmployee = orgPath.Path[0].Employee
		orgPath.ToEmployee = orgPath.Path[len(orgPath.Path)-1].Employee
	}

	return orgPath, nil
}

// GetReportingHierarchy returns the complete reporting hierarchy for a manager
func (s *Neo4jService) GetReportingHierarchy(ctx context.Context, managerID string, maxDepth int) (*ReportingHierarchy, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
		MATCH (manager:Employee {employee_id: $manager_id})
		
		// Get direct reports
		OPTIONAL MATCH (manager)<-[:REPORTS_TO]-(direct:Employee)
		
		// Get all reports up to maxDepth
		OPTIONAL MATCH (manager)<-[:REPORTS_TO*1..%d]-(all:Employee)
		
		RETURN manager,
		       collect(DISTINCT direct) as direct_reports,
		       collect(DISTINCT all) as all_reports
	`

	formattedQuery := fmt.Sprintf(query, maxDepth)

	params := map[string]interface{}{
		"manager_id": managerID,
	}

	result, err := session.Run(ctx, formattedQuery, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get reporting hierarchy: %w", err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get hierarchy result: %w", err)
	}

	hierarchy := &ReportingHierarchy{
		DirectReports: make([]EmployeeNode, 0),
		AllReports:    make([]EmployeeNode, 0),
		Depth:         maxDepth,
	}

	// Parse manager
	if managerNode, found := record.Get("manager"); found && managerNode != nil {
		if node, ok := managerNode.(neo4j.Node); ok {
			hierarchy.Manager = s.nodeToEmployee(node)
		}
	}

	// Parse direct reports
	if directReports, found := record.Get("direct_reports"); found && directReports != nil {
		if reports, ok := directReports.([]interface{}); ok {
			for _, report := range reports {
				if node, ok := report.(neo4j.Node); ok && node.ElementId != "" {
					hierarchy.DirectReports = append(hierarchy.DirectReports, s.nodeToEmployee(node))
				}
			}
		}
	}

	// Parse all reports
	if allReports, found := record.Get("all_reports"); found && allReports != nil {
		if reports, ok := allReports.([]interface{}); ok {
			for _, report := range reports {
				if node, ok := report.(neo4j.Node); ok && node.ElementId != "" {
					hierarchy.AllReports = append(hierarchy.AllReports, s.nodeToEmployee(node))
				}
			}
		}
	}

	return hierarchy, nil
}

// FindCommonManager finds the lowest common manager for a group of employees
func (s *Neo4jService) FindCommonManager(ctx context.Context, employeeIDs []string) (*EmployeeNode, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
		WITH $employee_ids as empIds
		UNWIND empIds as empId
		MATCH (emp:Employee {employee_id: empId})
		MATCH (emp)-[:REPORTS_TO*]->(manager:Employee)
		WITH manager, count(*) as reportCount, size(empIds) as totalEmployees
		WHERE reportCount = totalEmployees
		MATCH (manager)-[:REPORTS_TO*0..]->(topLevel:Employee)
		WHERE NOT (topLevel)-[:REPORTS_TO]->()
		RETURN manager
		ORDER BY length((manager)-[:REPORTS_TO*]->())
		LIMIT 1
	`

	params := map[string]interface{}{
		"employee_ids": employeeIDs,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to find common manager: %w", err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		// Check if no records found by comparing error message
		if err.Error() == "result contains no more records" {
			return nil, fmt.Errorf("no common manager found for employees: %v", employeeIDs)
		}
		return nil, fmt.Errorf("failed to get common manager result: %w", err)
	}

	if managerNode, found := record.Get("manager"); found && managerNode != nil {
		if node, ok := managerNode.(neo4j.Node); ok {
			manager := s.nodeToEmployee(node)
			return &manager, nil
		}
	}

	return nil, fmt.Errorf("no common manager found")
}

// GetDepartmentStructure returns the complete department hierarchy
func (s *Neo4jService) GetDepartmentStructure(ctx context.Context, rootDepartment string) (*DepartmentNode, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
		MATCH (root:Department {name: $root_department})
		OPTIONAL MATCH (root)<-[:BELONGS_TO*0..]-(child:Department)
		OPTIONAL MATCH (child)<-[:BELONGS_TO]-(pos:Position)<-[:HOLDS_POSITION]-(emp:Employee)
		RETURN root, collect(DISTINCT child) as departments, collect(DISTINCT emp) as employees
	`

	params := map[string]interface{}{
		"root_department": rootDepartment,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get department structure: %w", err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get department structure result: %w", err)
	}

	if rootNode, found := record.Get("root"); found && rootNode != nil {
		if node, ok := rootNode.(neo4j.Node); ok {
			dept := s.nodeToDepartment(node)
			return &dept, nil
		}
	}

	return nil, fmt.Errorf("department not found: %s", rootDepartment)
}

// Helper function to convert Neo4j node to EmployeeNode
func (s *Neo4jService) nodeToEmployee(node neo4j.Node) EmployeeNode {
	props := node.Props
	employee := EmployeeNode{
		Properties: make(map[string]interface{}),
	}

	if id, ok := props["id"].(string); ok {
		employee.ID = id
	}
	if empId, ok := props["employee_id"].(string); ok {
		employee.EmployeeID = empId
	}
	// Build legal name from first_name and last_name
	firstName := ""
	lastName := ""
	if fn, ok := props["first_name"].(string); ok {
		firstName = fn
	}
	if ln, ok := props["last_name"].(string); ok {
		lastName = ln
	}
	// Combine names for legal_name field
	if firstName != "" || lastName != "" {
		employee.LegalName = strings.TrimSpace(firstName + " " + lastName)
	}
	// Also check for legacy legal_name field
	if name, ok := props["legal_name"].(string); ok && employee.LegalName == "" {
		employee.LegalName = name
	}
	if email, ok := props["email"].(string); ok {
		employee.Email = email
	}
	if status, ok := props["employment_status"].(string); ok {
		employee.Status = status
	} else if status, ok := props["status"].(string); ok {
		employee.Status = status
	}
	if hireDate, ok := props["hire_date"].(time.Time); ok {
		employee.HireDate = hireDate
	}

	// Copy additional properties
	for k, v := range props {
		if k != "id" && k != "employee_id" && k != "legal_name" && k != "email" && k != "status" && k != "hire_date" {
			employee.Properties[k] = v
		}
	}

	return employee
}

// Helper function to convert Neo4j node to DepartmentNode
func (s *Neo4jService) nodeToDepartment(node neo4j.Node) DepartmentNode {
	props := node.Props
	dept := DepartmentNode{
		Properties: make(map[string]interface{}),
	}

	if id, ok := props["id"].(string); ok {
		dept.ID = id
	}
	if name, ok := props["name"].(string); ok {
		dept.Name = name
	}
	if parentId, ok := props["parent_id"].(string); ok {
		dept.ParentID = &parentId
	}
	if managerId, ok := props["manager_id"].(string); ok {
		dept.ManagerID = &managerId
	}

	// Copy additional properties
	for k, v := range props {
		if k != "id" && k != "name" && k != "parent_id" && k != "manager_id" {
			dept.Properties[k] = v
		}
	}

	return dept
}

// GetEmployee retrieves an employee by ID from Neo4j
func (s *Neo4jService) GetEmployee(ctx context.Context, employeeID string) (*EmployeeNode, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
		MATCH (e:Employee {id: $employee_id})
		OPTIONAL MATCH (e)-[:HOLDS_POSITION]->(p:Position)
		OPTIONAL MATCH (p)-[:BELONGS_TO]->(d:Department)
		OPTIONAL MATCH (m:Employee)-[:MANAGES]->(e)
		RETURN e,
			   p.position_title as position_title,
			   d.name as department,
			   m.legal_name as manager_name
	`

	s.logger.Printf("DEBUG: GetEmployee query for ID: %s", employeeID)
	s.logger.Printf("DEBUG: Query: %s", query)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, map[string]any{
			"employee_id": employeeID,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			if empNode, found := record.Get("e"); found {
				if node, ok := empNode.(neo4j.Node); ok {
					employee := s.nodeToEmployee(node)
					
					// Add additional fields
					if posTitle, found := record.Get("position_title"); found && posTitle != nil {
						employee.Properties["position_title"] = posTitle
					}
					if dept, found := record.Get("department"); found && dept != nil {
						employee.Properties["department"] = dept
					}
					if manager, found := record.Get("manager_name"); found && manager != nil {
						employee.Properties["manager_name"] = manager
					}
					
					return &employee, nil
				}
			}
		}
		return nil, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get employee %s: %w", employeeID, err)
	}

	if result == nil {
		return nil, fmt.Errorf("employee not found: %s", employeeID)
	}

	employee := result.(*EmployeeNode)
	return employee, nil
}

// SearchEmployees searches for employees by various criteria in Neo4j
func (s *Neo4jService) SearchEmployees(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*EmployeeNode, int, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	// Build dynamic query based on filters
	whereClause := "WHERE 1=1"
	params := make(map[string]any)
	
	if name, ok := filters["name"].(string); ok && name != "" {
		whereClause += " AND (e.legal_name CONTAINS $name OR e.employee_id CONTAINS $name)"
		params["name"] = name
	}
	
	if email, ok := filters["email"].(string); ok && email != "" {
		whereClause += " AND e.email CONTAINS $email"
		params["email"] = email
	}
	
	if department, ok := filters["department"].(string); ok && department != "" {
		whereClause += " AND d.name = $department"
		params["department"] = department
	}
	
	if status, ok := filters["status"].(string); ok && status != "" {
		whereClause += " AND e.status = $status"
		params["status"] = status
	}

	params["limit"] = limit
	params["offset"] = offset

	// Count query
	countQuery := fmt.Sprintf(`
		MATCH (e:Employee)
		OPTIONAL MATCH (e)-[:HOLDS_POSITION]->(p:Position)
		OPTIONAL MATCH (p)-[:BELONGS_TO]->(d:Department)
		%s
		RETURN COUNT(e) as total
	`, whereClause)

	// Data query
	dataQuery := fmt.Sprintf(`
		MATCH (e:Employee)
		OPTIONAL MATCH (e)-[:HOLDS_POSITION]->(p:Position)
		OPTIONAL MATCH (p)-[:BELONGS_TO]->(d:Department)
		OPTIONAL MATCH (m:Employee)-[:MANAGES]->(e)
		%s
		RETURN e,
			   p.position_title as position_title,
			   d.name as department,
			   m.legal_name as manager_name
		ORDER BY e.legal_name
		SKIP $offset
		LIMIT $limit
	`, whereClause)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Get total count
		countResult, err := tx.Run(ctx, countQuery, params)
		if err != nil {
			return nil, err
		}

		var total int
		if countResult.Next(ctx) {
			if count, found := countResult.Record().Get("total"); found {
				if c, ok := count.(int64); ok {
					total = int(c)
				}
			}
		}

		// Get data
		dataResult, err := tx.Run(ctx, dataQuery, params)
		if err != nil {
			return nil, err
		}

		var employees []*EmployeeNode
		for dataResult.Next(ctx) {
			record := dataResult.Record()
			if empNode, found := record.Get("e"); found {
				if node, ok := empNode.(neo4j.Node); ok {
					employee := s.nodeToEmployee(node)
					
					// Add additional fields
					if posTitle, found := record.Get("position_title"); found && posTitle != nil {
						employee.Properties["position_title"] = posTitle
					}
					if dept, found := record.Get("department"); found && dept != nil {
						employee.Properties["department"] = dept
					}
					if manager, found := record.Get("manager_name"); found && manager != nil {
						employee.Properties["manager_name"] = manager
					}
					
					employees = append(employees, &employee)
				}
			}
		}

		return map[string]interface{}{
			"employees": employees,
			"total":     total,
		}, nil
	})

	if err != nil {
		return nil, 0, fmt.Errorf("failed to search employees: %w", err)
	}

	resultMap := result.(map[string]interface{})
	employees := resultMap["employees"].([]*EmployeeNode)
	total := resultMap["total"].(int)

	return employees, total, nil
}

// SyncOrganization creates or updates an organization unit node in the graph
func (s *Neo4jService) SyncOrganization(ctx context.Context, org OrganizationNode) error {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
		MERGE (o:OrganizationUnit {id: $id, tenant_id: $tenant_id})
		SET o.unit_type = $unit_type,
		    o.name = $name,
		    o.description = $description,
		    o.status = $status,
		    o.is_active = $is_active,
		    o.updated_at = datetime()
		WITH o
		
		// Create parent relationship if parent_unit_id is provided
		FOREACH (parentId IN CASE WHEN $parent_unit_id <> '' THEN [$parent_unit_id] ELSE [] END |
			MERGE (parent:OrganizationUnit {id: parentId, tenant_id: $tenant_id})
			MERGE (parent)-[:PARENT_OF]->(o)
		)
		
		RETURN o
	`

	params := map[string]interface{}{
		"id":             org.ID,
		"tenant_id":      org.TenantID,
		"unit_type":      org.UnitType,
		"name":           org.Name,
		"description":    org.Description,
		"status":         org.Status,
		"is_active":      org.IsActive,
		"parent_unit_id": org.ParentUnitID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to sync organization %s: %w", org.ID, err)
	}

	s.logger.Printf("Synced organization %s (%s) to Neo4j", org.ID, org.Name)
	return nil
}

// DeleteOrganization deletes an organization unit from the graph
func (s *Neo4jService) DeleteOrganization(ctx context.Context, orgID, tenantID string) error {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	query := `
		MATCH (o:OrganizationUnit {id: $id, tenant_id: $tenant_id})
		DETACH DELETE o
		RETURN count(o) as deleted_count
	`

	params := map[string]interface{}{
		"id":        orgID,
		"tenant_id": tenantID,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to delete organization %s: %w", orgID, err)
	}

	record, err := result.Single(ctx)
	if err != nil {
		return fmt.Errorf("failed to get delete result for organization %s: %w", orgID, err)
	}

	deletedCount, found := record.Get("deleted_count")
	if !found {
		return fmt.Errorf("could not determine if organization %s was deleted", orgID)
	}

	if count, ok := deletedCount.(int64); ok {
		s.logger.Printf("Deleted %d organization nodes for ID %s", count, orgID)
	}

	return nil
}

// GetOrganization retrieves an organization by ID from Neo4j
func (s *Neo4jService) GetOrganization(ctx context.Context, orgID, tenantID string) (*OrganizationNode, error) {
	session := s.driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	query := `
		MATCH (o:OrganizationUnit {id: $id, tenant_id: $tenant_id})
		OPTIONAL MATCH (parent:OrganizationUnit)-[:PARENT_OF]->(o)
		OPTIONAL MATCH (o)-[:PARENT_OF]->(child:OrganizationUnit)
		RETURN o,
			   parent.id as parent_id,
			   collect(child.id) as child_ids
	`

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, map[string]any{
			"id":        orgID,
			"tenant_id": tenantID,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			if orgNode, found := record.Get("o"); found {
				if node, ok := orgNode.(neo4j.Node); ok {
					org := s.nodeToOrganization(node)
					
					// Add parent information
					if parentID, found := record.Get("parent_id"); found && parentID != nil {
						if pid, ok := parentID.(string); ok {
							org.ParentUnitID = pid
						}
					}
					
					// Add child information
					if childIDs, found := record.Get("child_ids"); found && childIDs != nil {
						if children, ok := childIDs.([]interface{}); ok {
							var childList []string
							for _, child := range children {
								if childStr, ok := child.(string); ok {
									childList = append(childList, childStr)
								}
							}
							org.Properties["child_unit_ids"] = childList
						}
					}
					
					return &org, nil
				}
			}
		}
		return nil, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get organization %s: %w", orgID, err)
	}

	if result == nil {
		return nil, fmt.Errorf("organization not found: %s", orgID)
	}

	organization := result.(*OrganizationNode)
	return organization, nil
}

// Helper function to convert Neo4j node to OrganizationNode
func (s *Neo4jService) nodeToOrganization(node neo4j.Node) OrganizationNode {
	props := node.Props
	org := OrganizationNode{
		Properties: make(map[string]interface{}),
	}

	if id, ok := props["id"].(string); ok {
		org.ID = id
	}
	if tenantID, ok := props["tenant_id"].(string); ok {
		org.TenantID = tenantID
	}
	if unitType, ok := props["unit_type"].(string); ok {
		org.UnitType = unitType
	}
	if name, ok := props["name"].(string); ok {
		org.Name = name
	}
	if description, ok := props["description"].(string); ok {
		org.Description = description
	}
	if status, ok := props["status"].(string); ok {
		org.Status = status
	}
	if isActive, ok := props["is_active"].(bool); ok {
		org.IsActive = isActive
	}

	// Copy additional properties
	for k, v := range props {
		if k != "id" && k != "tenant_id" && k != "unit_type" && k != "name" && 
		   k != "description" && k != "status" && k != "is_active" {
			org.Properties[k] = v
		}
	}

	return org
}

// GetDriver 获取Neo4j驱动程序
func (s *Neo4jService) GetDriver() neo4j.DriverWithContext {
	return s.driver
}
