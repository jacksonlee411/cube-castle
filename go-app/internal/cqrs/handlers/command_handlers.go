package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/commands"
	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
)

// CommandHandler 命令处理器
type CommandHandler struct {
	postgresRepo     repositories.PostgresCommandRepository
	organizationRepo repositories.OrganizationCommandRepository
	eventBus         events.EventBus
}

// NewCommandHandler 创建命令处理器
func NewCommandHandler(
	repo repositories.PostgresCommandRepository, 
	orgRepo repositories.OrganizationCommandRepository,
	eventBus events.EventBus,
) *CommandHandler {
	return &CommandHandler{
		postgresRepo:     repo,
		organizationRepo: orgRepo,
		eventBus:         eventBus,
	}
}

// HireEmployee 处理雇佣员工命令
func (h *CommandHandler) HireEmployee(w http.ResponseWriter, r *http.Request) {
	var cmd commands.HireEmployeeCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 生成员工ID
	employeeID := uuid.New()
	
	// 执行命令 - 在PostgreSQL中创建员工记录
	err := h.postgresRepo.CreateEmployee(r.Context(), repositories.Employee{
		ID:               employeeID,
		TenantID:         cmd.TenantID,
		FirstName:        cmd.FirstName,
		LastName:         cmd.LastName,
		Email:            cmd.Email,
		PositionID:       cmd.PositionID,
		HireDate:         cmd.HireDate,
		EmployeeType:     cmd.EmployeeType,
		EmploymentStatus: "PENDING_START",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	})

	if err != nil {
		http.Error(w, "Failed to create employee", http.StatusInternalServerError)
		return
	}

	// 发布领域事件
	event := events.NewEmployeeHired(cmd.TenantID, employeeID, "", cmd.FirstName, cmd.LastName, cmd.Email, cmd.HireDate)

	if err := h.eventBus.Publish(r.Context(), event); err != nil {
		// 记录日志但不阻止响应
		// TODO: 实现重试机制
	}

	// 返回成功响应
	response := map[string]interface{}{
		"employee_id": employeeID,
		"status":      "created",
		"message":     "Employee hired successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateEmployee 处理更新员工命令
func (h *CommandHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	var cmd commands.UpdateEmployeeCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 构建更新字段映射
	changes := make(map[string]interface{})
	if cmd.FirstName != nil {
		changes["first_name"] = *cmd.FirstName
	}
	if cmd.LastName != nil {
		changes["last_name"] = *cmd.LastName
	}
	if cmd.Email != nil {
		changes["email"] = *cmd.Email
	}

	// 执行更新
	err := h.postgresRepo.UpdateEmployee(r.Context(), cmd.ID, cmd.TenantID, changes)
	if err != nil {
		http.Error(w, "Failed to update employee", http.StatusInternalServerError)
		return
	}

	// 发布事件
	updatedFields := make(map[string]interface{})
	if cmd.FirstName != nil {
		updatedFields["first_name"] = *cmd.FirstName
	}
	if cmd.LastName != nil {
		updatedFields["last_name"] = *cmd.LastName
	}
	if cmd.Email != nil {
		updatedFields["email"] = *cmd.Email
	}
	
	event := events.NewEmployeeUpdated(cmd.TenantID, cmd.ID, "", updatedFields)

	h.eventBus.Publish(r.Context(), event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "updated",
		"message": "Employee updated successfully",
	})
}

// CreateOrganizationUnit 处理创建组织单元命令
func (h *CommandHandler) CreateOrganizationUnit(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreateOrganizationUnitCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	unitID := uuid.New()

	// 在PostgreSQL中创建组织单元
	err := h.postgresRepo.CreateOrganizationUnit(r.Context(), repositories.OrganizationUnit{
		ID:           unitID,
		TenantID:     cmd.TenantID,
		UnitType:     cmd.UnitType,
		Name:         cmd.Name,
		Description:  cmd.Description,
		ParentUnitID: cmd.ParentUnitID,
		Profile:      cmd.Profile,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})

	if err != nil {
		http.Error(w, "Failed to create organization unit", http.StatusInternalServerError)
		return
	}

	// 发布事件
	event := events.NewOrganizationCreated(cmd.TenantID, unitID, cmd.Name, "", nil, 1)

	h.eventBus.Publish(r.Context(), event)

	response := map[string]interface{}{
		"unit_id": unitID,
		"status":  "created",
		"message": "Organization unit created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CreateOrganization 创建组织 (新实现)
func (h *CommandHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreateOrganizationCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证租户ID
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		http.Error(w, "Missing tenant ID", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	cmd.TenantID = tenantID

	orgID := uuid.New()

	// 构建组织实体
	org := repositories.Organization{
		ID:           orgID,
		TenantID:     cmd.TenantID,
		UnitType:     cmd.UnitType,
		Name:         cmd.Name,
		Description:  cmd.Description,
		ParentUnitID: cmd.ParentUnitID,
		Status:       cmd.Status,
		Profile:      cmd.Profile,
		Level:        0, // 由系统计算
		IsActive:     cmd.Status == "ACTIVE",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 在PostgreSQL中创建组织
	err = h.organizationRepo.CreateOrganization(r.Context(), org)
	if err != nil {
		http.Error(w, "Failed to create organization", http.StatusInternalServerError)
		return
	}

	// 发布领域事件
	event := events.NewOrganizationCreated(cmd.TenantID, orgID, cmd.Name, cmd.UnitType, cmd.ParentUnitID, 1)
	if err := h.eventBus.Publish(r.Context(), event); err != nil {
		// 记录日志但不阻止响应
		// TODO: 实现重试机制
	}

	// 返回成功响应
	response := map[string]interface{}{
		"id":       orgID,
		"status":   "created",
		"message":  "Organization created successfully",
		"data":     org,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateOrganization 更新组织
func (h *CommandHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	var cmd commands.UpdateOrganizationCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 从URL获取组织ID
	orgIDStr := chi.URLParam(r, "id")
	if orgIDStr == "" {
		http.Error(w, "Missing organization ID", http.StatusBadRequest)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}
	cmd.ID = orgID

	// 验证租户ID
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		http.Error(w, "Missing tenant ID", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	cmd.TenantID = tenantID

	// 构建更新字段映射
	changes := make(map[string]interface{})
	if cmd.Name != nil {
		changes["name"] = *cmd.Name
	}
	if cmd.Description != nil {
		changes["description"] = *cmd.Description
	}
	if cmd.ParentUnitID != nil {
		changes["parent_unit_id"] = *cmd.ParentUnitID
	}
	if cmd.Status != nil {
		changes["status"] = *cmd.Status
		changes["is_active"] = *cmd.Status == "ACTIVE"
	}
	if cmd.Profile != nil {
		changes["profile"] = cmd.Profile
	}
	changes["updated_at"] = time.Now()

	// 执行更新
	err = h.organizationRepo.UpdateOrganization(r.Context(), cmd.ID, cmd.TenantID, changes)
	if err != nil {
		http.Error(w, "Failed to update organization", http.StatusInternalServerError)
		return
	}

	// 发布事件
	event := events.NewOrganizationUpdated(cmd.TenantID, cmd.ID, "", changes)
	h.eventBus.Publish(r.Context(), event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "updated",
		"message": "Organization updated successfully",
	})
}

// DeleteOrganization 删除组织
func (h *CommandHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	// 从URL获取组织ID
	orgIDStr := chi.URLParam(r, "id")
	if orgIDStr == "" {
		http.Error(w, "Missing organization ID", http.StatusBadRequest)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	// 验证租户ID
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		http.Error(w, "Missing tenant ID", http.StatusBadRequest)
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	// 执行删除
	err = h.organizationRepo.DeleteOrganization(r.Context(), orgID, tenantID)
	if err != nil {
		http.Error(w, "Failed to delete organization", http.StatusInternalServerError)
		return
	}

	// 发布事件
	event := events.NewOrganizationDeleted(tenantID, orgID, "", "")
	h.eventBus.Publish(r.Context(), event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "deleted",
		"message": "Organization deleted successfully",
	})
}