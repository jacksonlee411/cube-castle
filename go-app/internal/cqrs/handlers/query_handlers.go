package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
)

// QueryHandler 查询处理器
type QueryHandler struct {
	neo4jRepo repositories.Neo4jQueryRepository
}

// NewQueryHandler 创建查询处理器
func NewQueryHandler(repo repositories.Neo4jQueryRepository) *QueryHandler {
	return &QueryHandler{
		neo4jRepo: repo,
	}
}

// GetEmployee 获取员工信息
func (h *QueryHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "id")
	tenantID := r.URL.Query().Get("tenant_id")

	if employeeID == "" || tenantID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	empUUID, err := uuid.Parse(employeeID)
	if err != nil {
		http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.FindEmployeeQuery{
		TenantID: tenantUUID,
		ID:       empUUID,
	}

	employee, err := h.neo4jRepo.GetEmployee(r.Context(), query)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employee)
}

// SearchEmployees 搜索员工
func (h *QueryHandler) SearchEmployees(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "Missing tenant_id parameter", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	// 解析查询参数
	query := queries.SearchEmployeesQuery{
		TenantID: tenantUUID,
		Limit:    20, // 默认值
		Offset:   0,  // 默认值
	}

	if name := r.URL.Query().Get("name"); name != "" {
		query.Name = &name
	}
	if email := r.URL.Query().Get("email"); email != "" {
		query.Email = &email
	}
	if dept := r.URL.Query().Get("department"); dept != "" {
		query.Department = &dept
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			query.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}

	employees, err := h.neo4jRepo.SearchEmployees(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to search employees", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
}

// GetOrgChart 获取组织架构图
func (h *QueryHandler) GetOrgChart(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "Missing tenant_id parameter", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.GetOrgChartQuery{
		TenantID:        tenantUUID,
		MaxDepth:        5, // 默认值
		IncludeInactive: false,
	}

	if rootID := r.URL.Query().Get("root_unit_id"); rootID != "" {
		if rootUUID, err := uuid.Parse(rootID); err == nil {
			query.RootUnitID = &rootUUID
		}
	}

	if depthStr := r.URL.Query().Get("max_depth"); depthStr != "" {
		if depth, err := strconv.Atoi(depthStr); err == nil && depth > 0 && depth <= 10 {
			query.MaxDepth = depth
		}
	}

	if includeInactiveStr := r.URL.Query().Get("include_inactive"); includeInactiveStr == "true" {
		query.IncludeInactive = true
	}

	orgChart, err := h.neo4jRepo.GetOrgChart(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to get organization chart", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgChart)
}

// GetReportingHierarchy 获取汇报层级
func (h *QueryHandler) GetReportingHierarchy(w http.ResponseWriter, r *http.Request) {
	managerID := chi.URLParam(r, "manager_id")
	tenantID := r.URL.Query().Get("tenant_id")

	if managerID == "" || tenantID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	managerUUID, err := uuid.Parse(managerID)
	if err != nil {
		http.Error(w, "Invalid manager ID format", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.GetReportingHierarchyQuery{
		TenantID:  tenantUUID,
		ManagerID: managerUUID,
		MaxDepth:  5, // 默认值
	}

	if depthStr := r.URL.Query().Get("max_depth"); depthStr != "" {
		if depth, err := strconv.Atoi(depthStr); err == nil && depth > 0 && depth <= 10 {
			query.MaxDepth = depth
		}
	}

	hierarchy, err := h.neo4jRepo.GetReportingHierarchy(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to get reporting hierarchy", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hierarchy)
}

// GetOrganizationUnit 获取组织单元信息
func (h *QueryHandler) GetOrganizationUnit(w http.ResponseWriter, r *http.Request) {
	unitID := chi.URLParam(r, "id")
	tenantID := r.URL.Query().Get("tenant_id")

	if unitID == "" || tenantID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	unitUUID, err := uuid.Parse(unitID)
	if err != nil {
		http.Error(w, "Invalid unit ID format", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.GetOrganizationUnitQuery{
		TenantID: tenantUUID,
		ID:       unitUUID,
	}

	unit, err := h.neo4jRepo.GetOrganizationUnit(r.Context(), query)
	if err != nil {
		http.Error(w, "Organization unit not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unit)
}

// ListOrganizationUnits 列出组织单元
func (h *QueryHandler) ListOrganizationUnits(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "Missing tenant_id parameter", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.ListOrganizationUnitsQuery{
		TenantID: tenantUUID,
		Limit:    50, // 默认值
		Offset:   0,  // 默认值
	}

	if unitType := r.URL.Query().Get("unit_type"); unitType != "" {
		query.UnitType = &unitType
	}
	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		if parentUUID, err := uuid.Parse(parentID); err == nil {
			query.ParentID = &parentUUID
		}
	}

	units, err := h.neo4jRepo.ListOrganizationUnits(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to list organization units", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(units)
}

// TODO: 实现其他查询处理器方法
func (h *QueryHandler) FindEmployeePath(w http.ResponseWriter, r *http.Request) {
	// 实现员工路径查找逻辑
}

func (h *QueryHandler) GetDepartmentStructure(w http.ResponseWriter, r *http.Request) {
	// 实现部门结构查询逻辑
}

func (h *QueryHandler) FindCommonManager(w http.ResponseWriter, r *http.Request) {
	// 实现共同管理者查找逻辑
}