package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
)

// QueryHandler 查询处理器
type QueryHandler struct {
	neo4jRepo        repositories.Neo4jQueryRepository
	organizationRepo repositories.OrganizationQueryRepository
	positionRepo     repositories.PositionQueryRepository
}

// NewQueryHandler 创建查询处理器
func NewQueryHandler(
	repo repositories.Neo4jQueryRepository,
	orgRepo repositories.OrganizationQueryRepository,
	posRepo repositories.PositionQueryRepository,
) *QueryHandler {
	return &QueryHandler{
		neo4jRepo:        repo,
		organizationRepo: orgRepo,
		positionRepo:     posRepo,
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

// GetEmployeeStats 获取员工统计信息
func (h *QueryHandler) GetEmployeeStats(w http.ResponseWriter, r *http.Request) {
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

	// 构建员工统计查询
	query := queries.GetEmployeeStatsQuery{
		TenantID: tenantUUID,
	}

	// 使用查询处理器获取统计信息
	result, err := h.neo4jRepo.GetEmployeeStats(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to get employee stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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

// ListOrganizations 组织列表查询 (新实现)
func (h *QueryHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}
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
	query := queries.ListOrganizationsQuery{
		TenantID: tenantUUID,
		Page:     1,
		PageSize: 50,
	}

	if parentID := r.URL.Query().Get("parent_unit_id"); parentID != "" {
		if parentUUID, err := uuid.Parse(parentID); err == nil {
			query.ParentUnitID = &parentUUID
		}
	}
	if unitType := r.URL.Query().Get("unit_type"); unitType != "" {
		query.UnitType = &unitType
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query.Status = &status
	}
	if search := r.URL.Query().Get("search"); search != "" {
		query.Search = &search
	}
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 1000 {
			query.PageSize = pageSize
		}
	}

	organizations, pagination, err := h.organizationRepo.ListOrganizations(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to list organizations", http.StatusInternalServerError)
		return
	}

	// 构建响应
	response := map[string]interface{}{
		"organizations": organizations,
		"pagination":    pagination,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOrganization 获取单个组织 (新实现)
func (h *QueryHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "id")
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}

	if orgID == "" || tenantID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		http.Error(w, "Invalid organization ID format", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.GetOrganizationQuery{
		TenantID: tenantUUID,
		ID:       orgUUID,
	}

	organization, err := h.organizationRepo.GetOrganization(r.Context(), query)
	if err != nil {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(organization)
}

// GetOrganizationTree 组织树查询 (新实现)
func (h *QueryHandler) GetOrganizationTree(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}
	if tenantID == "" {
		http.Error(w, "Missing tenant_id parameter", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.GetOrganizationTreeQuery{
		TenantID:        tenantUUID,
		MaxDepth:        5,
		IncludeInactive: false,
		ExpandAll:       false,
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
	if expandAllStr := r.URL.Query().Get("expand_all"); expandAllStr == "true" {
		query.ExpandAll = true
	}

	tree, err := h.organizationRepo.GetOrganizationTree(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to get organization tree", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tree": tree,
		"metadata": map[string]interface{}{
			"max_depth":    query.MaxDepth,
			"total_nodes":  len(tree),
			"generated_at": time.Now().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOrganizationStats 组织统计查询 (新实现)
func (h *QueryHandler) GetOrganizationStats(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}
	if tenantID == "" {
		http.Error(w, "Missing tenant_id parameter", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.GetOrganizationStatsQuery{
		TenantID:    tenantUUID,
		Granularity: "daily",
	}

	if unitType := r.URL.Query().Get("unit_type"); unitType != "" {
		query.UnitType = &unitType
	}
	if parentID := r.URL.Query().Get("parent_id"); parentID != "" {
		if parentUUID, err := uuid.Parse(parentID); err == nil {
			query.ParentID = &parentUUID
		}
	}

	stats, err := h.organizationRepo.GetOrganizationStats(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to get organization stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": stats,
	})
}

// === 职位查询处理器 ===

// GetPosition 获取单个职位信息
func (h *QueryHandler) GetPosition(w http.ResponseWriter, r *http.Request) {
	positionID := chi.URLParam(r, "id")
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}

	if positionID == "" || tenantID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	posUUID, err := uuid.Parse(positionID)
	if err != nil {
		http.Error(w, "Invalid position ID format", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.GetPositionQuery{
		TenantID: tenantUUID,
		ID:       posUUID,
	}

	position, err := h.positionRepo.GetPosition(r.Context(), query)
	if err != nil {
		http.Error(w, "Position not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(position)
}

// GetPositionWithRelations 获取职位及其关联信息
func (h *QueryHandler) GetPositionWithRelations(w http.ResponseWriter, r *http.Request) {
	positionID := chi.URLParam(r, "id")
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}

	if positionID == "" || tenantID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	posUUID, err := uuid.Parse(positionID)
	if err != nil {
		http.Error(w, "Invalid position ID format", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	position, err := h.positionRepo.GetPositionWithRelations(r.Context(), posUUID, tenantUUID)
	if err != nil {
		http.Error(w, "Position not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(position)
}

// SearchPositions 搜索职位
func (h *QueryHandler) SearchPositions(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}
	if tenantID == "" {
		http.Error(w, "Missing tenant_id parameter", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	// 构建搜索参数
	params := repositories.SearchPositionsParams{
		TenantID: tenantUUID,
		Limit:    50, // 默认值
		Offset:   0,  // 默认值
	}

	// 解析查询参数
	if status := r.URL.Query().Get("status"); status != "" {
		params.Status = &status
	}
	if posType := r.URL.Query().Get("position_type"); posType != "" {
		params.PositionType = &posType
	}
	if deptID := r.URL.Query().Get("department_id"); deptID != "" {
		if deptUUID, err := uuid.Parse(deptID); err == nil {
			params.DepartmentID = &deptUUID
		}
	}
	if search := r.URL.Query().Get("search"); search != "" {
		params.Search = &search
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			params.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			params.Offset = offset
		}
	}

	positions, total, err := h.positionRepo.SearchPositions(r.Context(), params)
	if err != nil {
		http.Error(w, "Failed to search positions", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"positions": positions,
		"pagination": map[string]interface{}{
			"total":  total,
			"limit":  params.Limit,
			"offset": params.Offset,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPositionHierarchy 获取职位层级关系
func (h *QueryHandler) GetPositionHierarchy(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}
	if tenantID == "" {
		http.Error(w, "Missing tenant_id parameter", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.GetPositionHierarchyQuery{
		TenantID: tenantUUID,
		MaxDepth: 5, // 默认值
	}

	if rootPosID := r.URL.Query().Get("root_position_id"); rootPosID != "" {
		if rootUUID, err := uuid.Parse(rootPosID); err == nil {
			query.RootPositionID = &rootUUID
		}
	}
	if deptID := r.URL.Query().Get("department_id"); deptID != "" {
		if deptUUID, err := uuid.Parse(deptID); err == nil {
			query.DepartmentID = &deptUUID
		}
	}
	if depthStr := r.URL.Query().Get("max_depth"); depthStr != "" {
		if depth, err := strconv.Atoi(depthStr); err == nil && depth > 0 && depth <= 10 {
			query.MaxDepth = depth
		}
	}

	hierarchy, err := h.positionRepo.GetPositionHierarchy(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to get position hierarchy", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"hierarchy": hierarchy,
		"metadata": map[string]interface{}{
			"max_depth":    query.MaxDepth,
			"total_nodes":  len(hierarchy),
			"generated_at": time.Now().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetEmployeePositions 获取员工的职位历史
func (h *QueryHandler) GetEmployeePositions(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employee_id")
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}

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

	query := queries.GetEmployeePositionsQuery{
		TenantID:     tenantUUID,
		EmployeeID:   empUUID,
		IncludePast:  false, // 默认只返回当前职位
	}

	if includePastStr := r.URL.Query().Get("include_past"); includePastStr == "true" {
		query.IncludePast = true
	}

	positions, err := h.positionRepo.GetEmployeePositions(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to get employee positions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"employee_id": empUUID,
		"positions":   positions,
		"total":       len(positions),
	})
}

// GetPositionEmployees 获取职位下的员工
func (h *QueryHandler) GetPositionEmployees(w http.ResponseWriter, r *http.Request) {
	positionID := chi.URLParam(r, "position_id")
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}

	if positionID == "" || tenantID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	posUUID, err := uuid.Parse(positionID)
	if err != nil {
		http.Error(w, "Invalid position ID format", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.GetPositionEmployeesQuery{
		TenantID:     tenantUUID,
		PositionID:   posUUID,
		OnlyCurrent:  true, // 默认只返回当前员工
	}

	if onlyCurrentStr := r.URL.Query().Get("only_current"); onlyCurrentStr == "false" {
		query.OnlyCurrent = false
	}

	employees, err := h.positionRepo.GetPositionEmployees(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to get position employees", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"position_id": posUUID,
		"employees":   employees,
		"total":       len(employees),
	})
}

// GetPositionStats 获取职位统计信息
func (h *QueryHandler) GetPositionStats(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		tenantID = r.Header.Get("X-Tenant-ID")
	}
	if tenantID == "" {
		http.Error(w, "Missing tenant_id parameter", http.StatusBadRequest)
		return
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, "Invalid tenant ID format", http.StatusBadRequest)
		return
	}

	query := queries.GetPositionStatsQuery{
		TenantID: tenantUUID,
	}

	if deptID := r.URL.Query().Get("department_id"); deptID != "" {
		if deptUUID, err := uuid.Parse(deptID); err == nil {
			query.DepartmentID = &deptUUID
		}
	}

	stats, err := h.positionRepo.GetPositionStats(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to get position stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": stats,
		"metadata": map[string]interface{}{
			"generated_at": time.Now().Format(time.RFC3339),
		},
	})
}