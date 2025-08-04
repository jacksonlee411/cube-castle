package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/commands"
	"github.com/gaogu/cube-castle/go-app/internal/events"
	positionEvents "github.com/gaogu/cube-castle/go-app/internal/cqrs/events"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
)

// CommandHandler 命令处理器
type CommandHandler struct {
	postgresRepo     repositories.PostgresCommandRepository
	organizationRepo repositories.OrganizationCommandRepository
	positionRepo     repositories.PositionCommandRepository
	eventBus         events.EventBus
}

// NewCommandHandler 创建命令处理器
func NewCommandHandler(
	repo repositories.PostgresCommandRepository, 
	orgRepo repositories.OrganizationCommandRepository,
	posRepo repositories.PositionCommandRepository,
	eventBus events.EventBus,
) *CommandHandler {
	return &CommandHandler{
		postgresRepo:     repo,
		organizationRepo: orgRepo,
		positionRepo:     posRepo,
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
	err := h.postgresRepo.CreateEmployee(r.Context(), repositories.EmployeeEntity{
		ID:               employeeID,
		TenantID:         cmd.TenantID,
		FirstName:        cmd.FirstName,
		LastName:         cmd.LastName,
		Email:            cmd.Email,
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
		// 记录错误但不阻止响应
		fmt.Printf("❌ 事件发布失败: %v\n", err)
	} else {
		fmt.Printf("✅ 事件已发布: employee.hired, ID=%s, TenantID=%s\n", employeeID, cmd.TenantID)
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

// TerminateEmployee 处理终止员工命令
func (h *CommandHandler) TerminateEmployee(w http.ResponseWriter, r *http.Request) {
	var cmd commands.TerminateEmployeeCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 构建更新字段映射 - 更新员工状态为已终止
	changes := map[string]interface{}{
		"employment_status": "TERMINATED",
		"termination_date":  cmd.TerminationDate,
		"updated_at":        time.Now(),
	}

	// 执行更新员工状态
	err := h.postgresRepo.UpdateEmployee(r.Context(), cmd.ID, cmd.TenantID, changes)
	if err != nil {
		http.Error(w, "Failed to terminate employee", http.StatusInternalServerError)
		return
	}

	// 发布员工终止事件
	event := events.NewEmployeeTerminated(cmd.TenantID, cmd.ID, "", cmd.Reason, cmd.TerminationDate)
	if err := h.eventBus.Publish(r.Context(), event); err != nil {
		// 记录日志但不阻止响应
		// TODO: 实现重试机制
	}

	// 返回成功响应
	response := map[string]interface{}{
		"employee_id":       cmd.ID,
		"status":           "terminated",
		"termination_date": cmd.TerminationDate,
		"reason":           cmd.Reason,
		"message":          "Employee terminated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

// === 职位管理命令处理器 ===

// CreatePosition 处理创建职位命令
func (h *CommandHandler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	var cmd commands.CreatePositionCommand
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

	// 生成职位ID
	positionID := uuid.New()
	
	// 执行命令 - 在PostgreSQL中创建职位记录
	err = h.positionRepo.CreatePosition(r.Context(), repositories.Position{
		ID:                positionID,
		TenantID:          cmd.TenantID,
		PositionType:      cmd.PositionType,
		JobProfileID:      cmd.JobProfileID,
		DepartmentID:      cmd.DepartmentID,
		ManagerPositionID: cmd.ManagerPositionID,
		Status:            cmd.Status,
		BudgetedFTE:       cmd.BudgetedFTE,
		Details:           cmd.Details,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})
	
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create position: %v", err), http.StatusInternalServerError)
		return
	}

	// 发布事件
	event := positionEvents.NewPositionCreatedEvent(
		cmd.TenantID, 
		positionID, 
		cmd.PositionType, 
		cmd.DepartmentID, 
		cmd.Status, 
		cmd.BudgetedFTE, 
		cmd.Details,
	)
	
	if err := h.eventBus.Publish(r.Context(), event); err != nil {
		// 记录错误但不失败请求
		fmt.Printf("❌ 职位创建事件发布失败: %v\n", err)
	} else {
		fmt.Printf("✅ 职位创建事件已发布: position.created, ID=%s\n", positionID)
	}

	// 返回响应
	response := map[string]interface{}{
		"id":      positionID,
		"status":  "created",
		"message": "Position created successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdatePosition 处理更新职位命令
func (h *CommandHandler) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	var cmd commands.UpdatePositionCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 从URL获取职位ID
	positionIDStr := chi.URLParam(r, "id")
	if positionIDStr == "" {
		http.Error(w, "Missing position ID", http.StatusBadRequest)
		return
	}

	positionID, err := uuid.Parse(positionIDStr)
	if err != nil {
		http.Error(w, "Invalid position ID", http.StatusBadRequest)
		return
	}
	cmd.ID = positionID

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
	if cmd.JobProfileID != nil {
		changes["job_profile_id"] = *cmd.JobProfileID
	}
	if cmd.DepartmentID != nil {
		changes["department_id"] = *cmd.DepartmentID
	}
	if cmd.ManagerPositionID != nil {
		changes["manager_position_id"] = *cmd.ManagerPositionID
	}
	if cmd.Status != nil {
		changes["status"] = *cmd.Status
	}
	if cmd.BudgetedFTE != nil {
		changes["budgeted_fte"] = *cmd.BudgetedFTE
	}
	if cmd.Details != nil {
		changes["details"] = cmd.Details
	}
	changes["updated_at"] = time.Now()

	// 执行更新
	err = h.positionRepo.UpdatePosition(r.Context(), cmd.ID, cmd.TenantID, changes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update position: %v", err), http.StatusInternalServerError)
		return
	}

	// 发布事件
	event := positionEvents.NewPositionUpdatedEvent(cmd.TenantID, cmd.ID, changes, map[string]interface{}{})
	h.eventBus.Publish(r.Context(), event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "updated",
		"message": "Position updated successfully",
	})
}

// AssignEmployeeToPosition 处理员工职位分配命令（使用Outbox模式）
func (h *CommandHandler) AssignEmployeeToPosition(w http.ResponseWriter, r *http.Request) {
	var cmd commands.AssignEmployeeToPositionCommand
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

	// 验证员工和职位是否存在
	exists, err := h.positionRepo.ValidateEmployeePositionAssignment(r.Context(), cmd.EmployeeID, cmd.PositionID, cmd.TenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Employee or position not found", http.StatusNotFound)
		return
	}

	// 创建职位分配记录（使用简化的分配模型）
	assignmentID := uuid.New()
	assignment := repositories.PositionAssignment{
		ID:             assignmentID,
		TenantID:       cmd.TenantID,
		PositionID:     cmd.PositionID,
		EmployeeID:     cmd.EmployeeID,
		StartDate:      cmd.StartDate,
		IsCurrent:      true,
		FTE:            cmd.FTE,
		AssignmentType: cmd.AssignmentType,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 创建事件数据（用于Outbox）
	event := positionEvents.NewEmployeeAssignedToPositionEvent(
		cmd.TenantID,
		cmd.PositionID,
		cmd.EmployeeID,
		cmd.StartDate,
		cmd.FTE,
		cmd.AssignmentType,
		cmd.Reason,
	)
	
	eventData, err := json.Marshal(event)
	if err != nil {
		http.Error(w, "Failed to marshal event", http.StatusInternalServerError)
		return
	}

	outboxEvent := repositories.OutboxEvent{
		ID:          uuid.New(),
		TenantID:    cmd.TenantID,
		EventType:   "employee.assigned_to_position",
		AggregateID: assignmentID,
		EventData:   eventData,
		Status:      "PENDING",
		CreatedAt:   time.Now(),
	}

	// 使用Outbox模式：在同一事务中保存分配记录和事件
	err = h.positionRepo.AssignEmployeeWithEvent(r.Context(), assignment, outboxEvent)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to assign employee to position: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回响应（事件将由后台处理器异步发布）
	response := map[string]interface{}{
		"assignment_id": assignmentID,
		"status":       "assigned",
		"message":      "Employee assigned to position successfully",
		"note":         "Event will be processed asynchronously",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RemoveEmployeeFromPosition 处理员工职位移除命令
func (h *CommandHandler) RemoveEmployeeFromPosition(w http.ResponseWriter, r *http.Request) {
	var cmd commands.RemoveEmployeeFromPositionCommand
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

	// 结束职位占用
	err = h.positionRepo.EndPositionOccupancy(r.Context(), cmd.PositionID, cmd.EmployeeID, cmd.EndDate, cmd.Reason)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove employee from position: %v", err), http.StatusInternalServerError)
		return
	}

	// 发布事件
	event := positionEvents.NewEmployeeRemovedFromPositionEvent(
		cmd.TenantID,
		cmd.PositionID,
		cmd.EmployeeID,
		cmd.EndDate,
		cmd.Reason,
	)
	
	h.eventBus.Publish(r.Context(), event)

	response := map[string]interface{}{
		"status":  "removed",
		"message": "Employee removed from position successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeletePosition 处理删除职位命令
func (h *CommandHandler) DeletePosition(w http.ResponseWriter, r *http.Request) {
	var cmd commands.DeletePositionCommand
	
	// 从URL获取职位ID
	positionIDStr := chi.URLParam(r, "id")
	if positionIDStr == "" {
		http.Error(w, "Missing position ID", http.StatusBadRequest)
		return
	}

	positionID, err := uuid.Parse(positionIDStr)
	if err != nil {
		http.Error(w, "Invalid position ID", http.StatusBadRequest)
		return
	}
	cmd.ID = positionID

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

	// 从请求体获取删除原因
	var requestBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err == nil {
		if reason, ok := requestBody["reason"].(string); ok {
			cmd.Reason = reason
		}
	}
	if cmd.Reason == "" {
		cmd.Reason = "Position elimination"
	}

	// 执行删除
	err = h.positionRepo.DeletePosition(r.Context(), cmd.ID, cmd.TenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete position: %v", err), http.StatusInternalServerError)
		return
	}

	// 发布事件
	event := positionEvents.NewPositionDeletedEvent(cmd.TenantID, cmd.ID, cmd.Reason)
	h.eventBus.Publish(r.Context(), event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "deleted",
		"message": "Position deleted successfully",
	})
}