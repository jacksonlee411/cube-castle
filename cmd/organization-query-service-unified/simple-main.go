package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// 简单的数据模型
type Organization struct {
	Code            string     `json:"code"`
	ParentCode      *string    `json:"parentCode"`
	TenantID        string     `json:"tenantId"`
	Name            string     `json:"name"`
	UnitType        string     `json:"unitType"`
	Status          string     `json:"status"`
	IsDeleted       bool       `json:"isDeleted"`
	Level           int        `json:"level"`
	CodePath        string     `json:"codePath"`
	NamePath        string     `json:"namePath"`
	SortOrder       int        `json:"sortOrder"`
	Description     *string    `json:"description"`
	Profile         string     `json:"profile"`
	EffectiveDate   string     `json:"effectiveDate"`
	EndDate         *string    `json:"endDate"`
	IsCurrent       bool       `json:"isCurrent"`
	IsFuture        bool       `json:"isFuture"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	OperationType   string     `json:"operationType"`
	OperatedByID    string     `json:"-"`
	OperatedByName  string     `json:"-"`
	OperationReason *string    `json:"operationReason"`
	RecordID        string     `json:"recordId"`
}

type OperatedBy struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (o *Organization) GetOperatedBy() *OperatedBy {
	return &OperatedBy{
		ID:   o.OperatedByID,
		Name: o.OperatedByName,
	}
}

type OrganizationResponse struct {
	Data       []*Organization `json:"data"`
	TotalCount int             `json:"totalCount"`
	HasMore    bool            `json:"hasMore"`
}

type OrganizationStats struct {
	TotalCount      int `json:"totalCount"`
	ActiveCount     int `json:"activeCount"`
	InactiveCount   int `json:"inactiveCount"`
	DepartmentCount int `json:"departmentCount"`
	CompanyCount    int `json:"companyCount"`
	ProjectCount    int `json:"projectCount"`
}

var db *sql.DB

func init() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	}
	
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	
	log.Println("✅ Database connected successfully")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "GraphQL Query Service (Simple)",
		"version":   "v4.2.1",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(response)
}

func organizationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 获取查询参数
	firstStr := r.URL.Query().Get("first")
	offsetStr := r.URL.Query().Get("offset")

	first := 10
	if firstStr != "" {
		if f, err := strconv.Atoi(firstStr); err == nil {
			first = f
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	query := `
		SELECT code, parent_code, tenant_id, name, unit_type, status, 
			   CASE WHEN deleted_at IS NULL THEN false ELSE true END as is_deleted,
			   level, COALESCE(hierarchy_depth, level) as hierarchy_depth, 
			   COALESCE(code_path, path) as code_path, 
			   COALESCE(name_path, '') as name_path, COALESCE(sort_order, 0) as sort_order,
			   description, profile::text, effective_date, end_date, is_current, 
			   CASE WHEN effective_date > CURRENT_DATE THEN true ELSE false END as is_future,
			   created_at, updated_at, 
			   CASE WHEN suspended_at IS NOT NULL THEN 'SUSPEND' ELSE 'CREATE' END as operation_type, 
			   COALESCE(suspended_by::text, '00000000-0000-0000-0000-000000000000') as operated_by_id, 
			   'System User' as operated_by_name,
			   COALESCE(change_reason, 'System generated') as operation_reason, record_id
		FROM organization_units 
		WHERE is_current = true AND deleted_at IS NULL
		ORDER BY code
		LIMIT $1 OFFSET $2
	`

	rows, err := db.Query(query, first, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var organizations []*Organization
	for rows.Next() {
		org := &Organization{}
		var hierarchyDepth int
		err := rows.Scan(
			&org.Code, &org.ParentCode, &org.TenantID, &org.Name, &org.UnitType,
			&org.Status, &org.IsDeleted, &org.Level, &hierarchyDepth,
			&org.CodePath, &org.NamePath, &org.SortOrder, &org.Description,
			&org.Profile, &org.EffectiveDate, &org.EndDate, &org.IsCurrent,
			&org.IsFuture, &org.CreatedAt, &org.UpdatedAt, &org.OperationType,
			&org.OperatedByID, &org.OperatedByName, &org.OperationReason, &org.RecordID,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		organizations = append(organizations, org)
	}

	// 获取总数
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM organization_units WHERE is_current = true AND NOT is_deleted"
	err = db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		http.Error(w, fmt.Sprintf("Count error: %v", err), http.StatusInternalServerError)
		return
	}

	response := OrganizationResponse{
		Data:       organizations,
		TotalCount: totalCount,
		HasMore:    offset+len(organizations) < totalCount,
	}

	json.NewEncoder(w).Encode(response)
}

func organizationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	query := `
		SELECT code, parent_code, tenant_id, name, unit_type, status, 
			   CASE WHEN deleted_at IS NULL THEN false ELSE true END as is_deleted,
			   level, COALESCE(hierarchy_depth, level) as hierarchy_depth, 
			   COALESCE(code_path, path) as code_path, 
			   COALESCE(name_path, '') as name_path, COALESCE(sort_order, 0) as sort_order,
			   description, profile::text, effective_date, end_date, is_current, 
			   CASE WHEN effective_date > CURRENT_DATE THEN true ELSE false END as is_future,
			   created_at, updated_at, 
			   CASE WHEN suspended_at IS NOT NULL THEN 'SUSPEND' ELSE 'CREATE' END as operation_type, 
			   COALESCE(suspended_by::text, '00000000-0000-0000-0000-000000000000') as operated_by_id, 
			   'System User' as operated_by_name,
			   COALESCE(change_reason, 'System generated') as operation_reason, record_id
		FROM organization_units 
		WHERE code = $1 AND is_current = true AND deleted_at IS NULL
		LIMIT 1
	`

	org := &Organization{}
	var hierarchyDepth int
	err := db.QueryRow(query, code).Scan(
		&org.Code, &org.ParentCode, &org.TenantID, &org.Name, &org.UnitType,
		&org.Status, &org.IsDeleted, &org.Level, &hierarchyDepth,
		&org.CodePath, &org.NamePath, &org.SortOrder, &org.Description,
		&org.Profile, &org.EffectiveDate, &org.EndDate, &org.IsCurrent,
		&org.IsFuture, &org.CreatedAt, &org.UpdatedAt, &org.OperationType,
		&org.OperatedByID, &org.OperatedByName, &org.OperationReason, &org.RecordID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Organization not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Query error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(org)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	query := `
		SELECT 
			COUNT(*) as total_count,
			COUNT(*) FILTER (WHERE status = 'ACTIVE') as active_count,
			COUNT(*) FILTER (WHERE status = 'INACTIVE') as inactive_count,
			COUNT(*) FILTER (WHERE unit_type = 'DEPARTMENT') as department_count,
			COUNT(*) FILTER (WHERE unit_type = 'COMPANY') as company_count,
			COUNT(*) FILTER (WHERE unit_type = 'PROJECT_TEAM') as project_count
		FROM organization_units 
		WHERE is_current = true AND deleted_at IS NULL
	`

	stats := &OrganizationStats{}
	err := db.QueryRow(query).Scan(
		&stats.TotalCount, &stats.ActiveCount, &stats.InactiveCount,
		&stats.DepartmentCount, &stats.CompanyCount, &stats.ProjectCount,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Stats error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

// handleGraphQLQuery 处理GraphQL查询
func handleGraphQLQuery(query string, variables map[string]interface{}) map[string]interface{} {
	// 构建响应数据映射
	response := map[string]interface{}{
		"data": map[string]interface{}{},
	}
	
	// 检查查询包含的字段并相应处理
	hasOrganizations := strings.Contains(query, "organizations(")
	hasOrganizationStats := strings.Contains(query, "organizationStats")
	hasSingleOrg := strings.Contains(query, "organization(") && !strings.Contains(query, "organizations(")
	hasHierarchy := strings.Contains(query, "organizationHierarchy")
	hasSubtree := strings.Contains(query, "organizationSubtree")
	hasAuditHistory := strings.Contains(query, "organizationAuditHistory")
	hasChangeAnalysis := strings.Contains(query, "organizationChangeAnalysis")
	hasHierarchyStats := strings.Contains(query, "hierarchyStatistics")
	hasConsistencyCheck := strings.Contains(query, "hierarchyConsistencyCheck")
	hasAuditLog := strings.Contains(query, "auditLog")
	
	// 处理组织列表查询
	if hasOrganizations {
		orgResponse := handleOrganizationsQuery(variables)
		if orgData, ok := orgResponse["data"].(map[string]interface{}); ok {
			if orgField, ok := orgData["organizations"]; ok {
				response["data"].(map[string]interface{})["organizations"] = orgField
			}
		}
		if errors, ok := orgResponse["errors"]; ok {
			response["errors"] = errors
			return response
		}
	}
	
	// 处理组织统计查询
	if hasOrganizationStats {
		statsResponse := handleOrganizationStatsQuery()
		if statsData, ok := statsResponse["data"].(map[string]interface{}); ok {
			if statsField, ok := statsData["organizationStats"]; ok {
				response["data"].(map[string]interface{})["organizationStats"] = statsField
			}
		}
		if errors, ok := statsResponse["errors"]; ok {
			response["errors"] = errors
			return response
		}
	}
	
	// 处理单个组织查询
	if hasSingleOrg {
		orgResponse := handleOrganizationQuery(variables)
		if orgData, ok := orgResponse["data"].(map[string]interface{}); ok {
			if orgField, ok := orgData["organization"]; ok {
				response["data"].(map[string]interface{})["organization"] = orgField
			}
		}
		if errors, ok := orgResponse["errors"]; ok {
			response["errors"] = errors
			return response
		}
	}
	
	// 处理其他查询类型
	if hasHierarchy {
		return handleOrganizationHierarchyQuery(variables)
	}
	if hasSubtree {
		return handleOrganizationSubtreeQuery(variables)
	}
	if hasAuditHistory {
		return handleAuditHistoryQuery(variables)
	}
	if hasChangeAnalysis {
		return handleChangeAnalysisQuery(variables)
	}
	if hasHierarchyStats {
		return handleHierarchyStatisticsQuery()
	}
	if hasConsistencyCheck {
		return handleConsistencyCheckQuery()
	}
	if hasAuditLog {
		return handleAuditLogQuery(variables)
	}
	
	// 如果没有匹配的查询字段，返回错误
	if len(response["data"].(map[string]interface{})) == 0 {
		return map[string]interface{}{
			"errors": []map[string]interface{}{{
				"message": fmt.Sprintf("Unsupported GraphQL query: %s", query),
			}},
		}
	}
	
	return response
}

// handleOrganizationsQuery 处理组织列表查询
func handleOrganizationsQuery(variables map[string]interface{}) map[string]interface{} {
	first := 10
	offset := 0
	
	if val, ok := variables["first"]; ok {
		if f, ok := val.(float64); ok {
			first = int(f)
		}
	}
	if val, ok := variables["offset"]; ok {
		if o, ok := val.(float64); ok {
			offset = int(o)
		}
	}

	query := `
		SELECT code, parent_code, tenant_id, name, unit_type, status, 
			   CASE WHEN deleted_at IS NULL THEN false ELSE true END as is_deleted,
			   level, COALESCE(hierarchy_depth, level) as hierarchy_depth, 
			   COALESCE(code_path, path) as code_path, 
			   COALESCE(name_path, '') as name_path, COALESCE(sort_order, 0) as sort_order,
			   description, profile::text, effective_date, end_date, is_current, 
			   CASE WHEN effective_date > CURRENT_DATE THEN true ELSE false END as is_future,
			   created_at, updated_at, 
			   CASE WHEN suspended_at IS NOT NULL THEN 'SUSPEND' ELSE 'CREATE' END as operation_type, 
			   COALESCE(suspended_by::text, '00000000-0000-0000-0000-000000000000') as operated_by_id, 
			   'System User' as operated_by_name,
			   COALESCE(change_reason, 'System generated') as operation_reason, record_id
		FROM organization_units 
		WHERE is_current = true AND deleted_at IS NULL
		ORDER BY code
		LIMIT $1 OFFSET $2
	`

	rows, err := db.Query(query, first, offset)
	if err != nil {
		return map[string]interface{}{
			"errors": []map[string]interface{}{{
				"message": fmt.Sprintf("Query error: %v", err),
			}},
		}
	}
	defer rows.Close()

	var organizations []*Organization
	for rows.Next() {
		org := &Organization{}
		var hierarchyDepth int
		err := rows.Scan(
			&org.Code, &org.ParentCode, &org.TenantID, &org.Name, &org.UnitType,
			&org.Status, &org.IsDeleted, &org.Level, &hierarchyDepth,
			&org.CodePath, &org.NamePath, &org.SortOrder, &org.Description,
			&org.Profile, &org.EffectiveDate, &org.EndDate, &org.IsCurrent,
			&org.IsFuture, &org.CreatedAt, &org.UpdatedAt, &org.OperationType,
			&org.OperatedByID, &org.OperatedByName, &org.OperationReason, &org.RecordID,
		)
		if err != nil {
			return map[string]interface{}{
				"errors": []map[string]interface{}{{
					"message": fmt.Sprintf("Scan error: %v", err),
				}},
			}
		}
		organizations = append(organizations, org)
	}

	// 获取总数
	var totalCount int
	countQuery := "SELECT COUNT(*) FROM organization_units WHERE is_current = true AND deleted_at IS NULL"
	err = db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return map[string]interface{}{
			"errors": []map[string]interface{}{{
				"message": fmt.Sprintf("Count error: %v", err),
			}},
		}
	}

	return map[string]interface{}{
		"data": map[string]interface{}{
			"organizations": map[string]interface{}{
				"data":       organizations,
				"totalCount": totalCount,
				"hasMore":    offset+len(organizations) < totalCount,
			},
		},
	}
}

// handleOrganizationQuery 处理单个组织查询
func handleOrganizationQuery(variables map[string]interface{}) map[string]interface{} {
	code, ok := variables["code"].(string)
	if !ok {
		return map[string]interface{}{
			"errors": []map[string]interface{}{{
				"message": "Missing required parameter: code",
			}},
		}
	}

	query := `
		SELECT code, parent_code, tenant_id, name, unit_type, status, 
			   CASE WHEN deleted_at IS NULL THEN false ELSE true END as is_deleted,
			   level, COALESCE(hierarchy_depth, level) as hierarchy_depth, 
			   COALESCE(code_path, path) as code_path, 
			   COALESCE(name_path, '') as name_path, COALESCE(sort_order, 0) as sort_order,
			   description, profile::text, effective_date, end_date, is_current, 
			   CASE WHEN effective_date > CURRENT_DATE THEN true ELSE false END as is_future,
			   created_at, updated_at, 
			   CASE WHEN suspended_at IS NOT NULL THEN 'SUSPEND' ELSE 'CREATE' END as operation_type, 
			   COALESCE(suspended_by::text, '00000000-0000-0000-0000-000000000000') as operated_by_id, 
			   'System User' as operated_by_name,
			   COALESCE(change_reason, 'System generated') as operation_reason, record_id
		FROM organization_units 
		WHERE code = $1 AND is_current = true AND deleted_at IS NULL
		LIMIT 1
	`

	org := &Organization{}
	var hierarchyDepth int
	err := db.QueryRow(query, code).Scan(
		&org.Code, &org.ParentCode, &org.TenantID, &org.Name, &org.UnitType,
		&org.Status, &org.IsDeleted, &org.Level, &hierarchyDepth,
		&org.CodePath, &org.NamePath, &org.SortOrder, &org.Description,
		&org.Profile, &org.EffectiveDate, &org.EndDate, &org.IsCurrent,
		&org.IsFuture, &org.CreatedAt, &org.UpdatedAt, &org.OperationType,
		&org.OperatedByID, &org.OperatedByName, &org.OperationReason, &org.RecordID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return map[string]interface{}{
				"data": map[string]interface{}{
					"organization": nil,
				},
			}
		}
		return map[string]interface{}{
			"errors": []map[string]interface{}{{
				"message": fmt.Sprintf("Query error: %v", err),
			}},
		}
	}

	return map[string]interface{}{
		"data": map[string]interface{}{
			"organization": org,
		},
	}
}

// 其他查询处理函数（简化实现）
func handleOrganizationStatsQuery() map[string]interface{} {
	stats := &OrganizationStats{}
	query := `
		SELECT 
			COUNT(*) as total_count,
			COUNT(*) FILTER (WHERE status = 'ACTIVE') as active_count,
			COUNT(*) FILTER (WHERE status = 'INACTIVE') as inactive_count,
			COUNT(*) FILTER (WHERE unit_type = 'DEPARTMENT') as department_count,
			COUNT(*) FILTER (WHERE unit_type = 'COMPANY') as company_count,
			COUNT(*) FILTER (WHERE unit_type = 'PROJECT_TEAM') as project_count
		FROM organization_units 
		WHERE is_current = true AND deleted_at IS NULL
	`

	err := db.QueryRow(query).Scan(
		&stats.TotalCount, &stats.ActiveCount, &stats.InactiveCount,
		&stats.DepartmentCount, &stats.CompanyCount, &stats.ProjectCount,
	)
	if err != nil {
		return map[string]interface{}{
			"errors": []map[string]interface{}{{
				"message": fmt.Sprintf("Stats error: %v", err),
			}},
		}
	}

	return map[string]interface{}{
		"data": map[string]interface{}{
			"organizationStats": stats,
		},
	}
}

// 其他查询处理函数（暂时返回模拟数据）
func handleOrganizationHierarchyQuery(variables map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"organizationHierarchy": map[string]interface{}{
				"code":     "1000000",
				"path":     "/1000000",
				"level":    1,
				"children": []interface{}{},
			},
		},
	}
}

func handleOrganizationSubtreeQuery(variables map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"organizationSubtree": []map[string]interface{}{{
				"code": "1000000",
				"name": "高谷集团",
				"level": 1,
			}},
		},
	}
}

func handleHierarchyStatisticsQuery() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"hierarchyStatistics": map[string]interface{}{
				"maxLevel": 1,
				"totalNodes": 1,
				"levelDistribution": []map[string]interface{}{{
					"level": 1,
					"count": 1,
				}},
			},
		},
	}
}

func handleAuditHistoryQuery(variables map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"organizationAuditHistory": []map[string]interface{}{{
				"operationType": "CREATE",
				"operatedBy": map[string]interface{}{
					"id": "00000000-0000-0000-0000-000000000000",
					"name": "System User",
				},
				"createdAt": "2025-01-01T00:00:00Z",
			}},
		},
	}
}

func handleAuditLogQuery(variables map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"auditLog": map[string]interface{}{
				"auditId": "audit-001",
				"operationType": "CREATE",
				"changesSummary": "Organization created",
			},
		},
	}
}

func handleChangeAnalysisQuery(variables map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"organizationChangeAnalysis": map[string]interface{}{
				"totalChanges": 1,
				"changesByType": []map[string]interface{}{{
					"operationType": "CREATE",
					"count": 1,
				}},
			},
		},
	}
}

func handleConsistencyCheckQuery() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"hierarchyConsistencyCheck": map[string]interface{}{
				"isConsistent": true,
				"issues": []interface{}{},
				"checkedAt": "2025-08-24T00:00:00Z",
			},
		},
	}
}

func main() {
	// HTTP路由
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/organizations", organizationsHandler)
	http.HandleFunc("/api/organization", organizationHandler)
	http.HandleFunc("/api/stats", statsHandler)

	// GraphQL端点处理 (带认证)
	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 简化的Bearer Token验证
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": []map[string]interface{}{{
					"message": "Missing or invalid Authorization header. Expected: Bearer <token>",
				}},
			})
			return
		}

		// 解析GraphQL查询
		if r.Method != "POST" {
			http.Error(w, "GraphQL只支持POST请求", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Query     string                 `json:"query"`
			Variables map[string]interface{} `json:"variables"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errors": []map[string]interface{}{{
					"message": fmt.Sprintf("Invalid JSON body: %v", err),
				}},
			})
			return
		}

		// 处理GraphQL查询
		response := handleGraphQLQuery(req.Query, req.Variables)
		json.NewEncoder(w).Encode(response)
	})

	// GraphiQL界面
	http.HandleFunc("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(graphiqlHTML))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	log.Printf("🚀 Simple GraphQL Service启动在端口 %s", port)
	log.Printf("📊 REST API: http://localhost:%s/api/organizations", port)
	log.Printf("📊 GraphQL: http://localhost:%s/graphql", port)
	log.Printf("🔧 GraphiQL: http://localhost:%s/graphiql", port)
	log.Printf("🏥 健康检查: http://localhost:%s/health", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

const graphiqlHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>GraphiQL - Cube Castle API (Simple)</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .endpoint h3 { margin-top: 0; color: #333; }
        .endpoint pre { background: #fff; padding: 10px; border-radius: 3px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🏰 Cube Castle API - Simple Version</h1>
        <p>Version: v4.2.1 | Architecture: Simple REST API</p>
        
        <div class="endpoint">
            <h3>📊 Organizations List</h3>
            <pre>GET /api/organizations?first=10&offset=0</pre>
            <p>获取组织列表，支持分页</p>
        </div>
        
        <div class="endpoint">
            <h3>🏢 Single Organization</h3>
            <pre>GET /api/organization?code=1000000</pre>
            <p>根据code获取单个组织</p>
        </div>
        
        <div class="endpoint">
            <h3>📈 Statistics</h3>
            <pre>GET /api/stats</pre>
            <p>获取组织统计数据</p>
        </div>
        
        <div class="endpoint">
            <h3>🔧 GraphQL (Simulated)</h3>
            <pre>POST /graphql</pre>
            <p>GraphQL端点 (简化版本)</p>
        </div>
        
        <div class="endpoint">
            <h3>🏥 Health Check</h3>
            <pre>GET /health</pre>
            <p>服务健康状态检查</p>
        </div>
    </div>
</body>
</html>
`