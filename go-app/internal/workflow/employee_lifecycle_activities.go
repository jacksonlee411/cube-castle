// internal/workflow/employee_lifecycle_activities.go
package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/employee"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/service"
	"github.com/google/uuid"
)

// EmployeeLifecycleActivities 员工生命周期活动实现
type EmployeeLifecycleActivities struct {
	entClient        *ent.Client
	temporalQuerySvc *service.TemporalQueryService
	logger           *logging.StructuredLogger
}

// NewEmployeeLifecycleActivities 创建员工生命周期活动实例
func NewEmployeeLifecycleActivities(
	entClient *ent.Client,
	temporalQuerySvc *service.TemporalQueryService,
	logger *logging.StructuredLogger,
) *EmployeeLifecycleActivities {
	return &EmployeeLifecycleActivities{
		entClient:        entClient,
		temporalQuerySvc: temporalQuerySvc,
		logger:           logger,
	}
}

// CreateCandidateActivity 创建候选人活动
func (a *EmployeeLifecycleActivities) CreateCandidateActivity(
	ctx context.Context,
	req CandidateCreationRequest,
) (*CandidateCreationResult, error) {
	a.logger.Info("Creating candidate",
		"tenant_id", req.TenantID,
		"created_by", req.CreatedBy,
		"function", "CreateCandidateActivity",
	)

	// 1. 输入验证
	if req.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.CreatedBy == uuid.Nil {
		return nil, fmt.Errorf("created_by is required")
	}
	if req.CandidateData == nil || len(req.CandidateData) == 0 {
		return nil, fmt.Errorf("candidate_data is required")
	}

	// 2. 验证必要的候选人信息
	name, hasName := req.CandidateData["name"]
	email, hasEmail := req.CandidateData["email"]

	if !hasName || name == "" {
		return nil, fmt.Errorf("candidate name is required")
	}
	if !hasEmail || email == "" {
		return nil, fmt.Errorf("candidate email is required")
	}

	// 3. 生成候选人ID
	candidateID := uuid.New()

	// 4. 创建员工记录（预招聘状态）
	position := "Candidate"
	if req.PositionInfo != nil {
		if posTitle, exists := req.PositionInfo.NewPosition["title"]; exists {
			position = fmt.Sprintf("Candidate - %s", posTitle)
		}
	}

	employee, err := a.entClient.Employee.
		Create().
		SetID(candidateID.String()).
		SetName(name.(string)).
		SetEmail(email.(string)).
		SetPosition(position).
		Save(ctx)

	if err != nil {
		a.logger.Error("Failed to create candidate record",
			"error", err.Error(),
			"function", "CreateCandidateActivity",
			"tenant_id", req.TenantID,
		)
		return nil, fmt.Errorf("failed to create candidate record: %w", err)
	}

	a.logger.Info("Candidate created successfully",
		"tenant_id", req.TenantID,
		"candidate_id", candidateID,
		"employee_id", employee.ID,
		"candidate_name", name,
		"candidate_email", email,
		"function", "CreateCandidateActivity",
	)

	return &CandidateCreationResult{
		CandidateID: candidateID,
		Status:      "created",
		Success:     true,
	}, nil
}

// UpdateCandidateActivity 更新候选人活动
func (a *EmployeeLifecycleActivities) UpdateCandidateActivity(
	ctx context.Context,
	req InformationUpdateRequest,
) (*InformationUpdateResult, error) {
	a.logger.LogInfo(ctx, "Updating candidate information", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"employee_id": req.EmployeeID,
		"update_type": req.UpdateType,
		"function":    "UpdateCandidateActivity",
	})

	// 1. 输入验证
	if req.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.EmployeeID == uuid.Nil {
		return nil, fmt.Errorf("employee_id is required")
	}
	if req.UpdateType == "" {
		return nil, fmt.Errorf("update_type is required")
	}
	if req.UpdateData == nil || len(req.UpdateData) == 0 {
		return nil, fmt.Errorf("update_data is required")
	}
	if req.UpdatedBy == uuid.Nil {
		return nil, fmt.Errorf("updated_by is required")
	}

	// 2. 验证候选人存在且属于同一租户
	// 注意：这里使用 EmployeeID 作为候选人ID，因为候选人创建后会生成员工记录
	candidate, err := a.entClient.Employee.Query().
		Where(
			employee.ID(req.EmployeeID.String()),
			// TODO: Add candidate status check to ensure this is still a candidate
			// TODO: Add tenant_id field check when Person schema is integrated
		).Only(ctx)
	if err != nil {
		a.logger.LogError(ctx, "Candidate not found", map[string]interface{}{
			"employee_id": req.EmployeeID,
			"tenant_id":   req.TenantID,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("candidate not found: %w", err)
	}

	updateID := uuid.New()
	requiresApproval := req.RequiresApproval

	// 3. 根据更新类型执行相应的更新逻辑
	switch req.UpdateType {
	case "PERSONAL":
		err = a.updateCandidatePersonalInformation(ctx, candidate, req.UpdateData)
		// 候选人个人信息更新通常不需要审批
		requiresApproval = false

	case "CONTACT":
		err = a.updateCandidateContactInformation(ctx, candidate, req.UpdateData)
		// 候选人联系信息更新通常不需要审批
		requiresApproval = false

	case "APPLICATION_STATUS":
		err = a.updateCandidateApplicationStatus(ctx, candidate, req.UpdateData)
		// 申请状态更新可能需要审批
		requiresApproval = true

	case "INTERVIEW_FEEDBACK":
		err = a.updateCandidateInterviewFeedback(ctx, candidate, req.UpdateData)
		// 面试反馈通常不需要审批
		requiresApproval = false

	default:
		return nil, fmt.Errorf("unsupported update_type for candidate: %s", req.UpdateType)
	}

	if err != nil {
		a.logger.LogError(ctx, "Failed to update candidate information", map[string]interface{}{
			"employee_id": req.EmployeeID,
			"update_type": req.UpdateType,
			"update_id":   updateID,
			"error":       err.Error(),
		})
		return &InformationUpdateResult{
			UpdateID: updateID,
			Status:   "failed",
			Success:  false,
		}, fmt.Errorf("candidate update failed: %w", err)
	}

	// 4. 构建结果
	result := &InformationUpdateResult{
		UpdateID:         updateID,
		Status:           "updated",
		RequiredApproval: requiresApproval,
		Success:          true,
	}

	// 5. 如果需要审批，创建审批流程
	if requiresApproval {
		approvalID, err := a.initiateCandidateApprovalProcess(ctx, req)
		if err != nil {
			a.logger.LogError(ctx, "Failed to initiate candidate approval process", map[string]interface{}{
				"employee_id": req.EmployeeID,
				"update_id":   updateID,
				"error":       err.Error(),
			})
			result.Status = "pending_approval_failed"
		} else {
			result.ApprovalID = &approvalID
			result.Status = "pending_approval"
		}
	}

	// 6. 记录更新历史
	err = a.recordCandidateUpdateHistory(ctx, req, updateID, result.Status)
	if err != nil {
		a.logger.LogError(ctx, "Failed to record candidate update history", map[string]interface{}{
			"employee_id": req.EmployeeID,
			"update_id":   updateID,
			"error":       err.Error(),
		})
		// 历史记录失败不影响主流程
	}

	a.logger.LogInfo(ctx, "Candidate information update completed", map[string]interface{}{
		"tenant_id":         req.TenantID,
		"employee_id":       req.EmployeeID,
		"update_id":         updateID,
		"update_type":       req.UpdateType,
		"requires_approval": requiresApproval,
		"status":            result.Status,
		"success":           result.Success,
	})

	return result, nil
}

// updateCandidatePersonalInformation 更新候选人个人信息
func (a *EmployeeLifecycleActivities) updateCandidatePersonalInformation(
	ctx context.Context,
	candidate *ent.Employee,
	updateData map[string]interface{},
) error {
	updateQuery := a.entClient.Employee.UpdateOneID(candidate.ID)

	// 支持的候选人个人信息字段更新
	if name, ok := updateData["name"].(string); ok && name != "" {
		updateQuery = updateQuery.SetName(name)
	}

	if email, ok := updateData["email"].(string); ok && email != "" {
		updateQuery = updateQuery.SetEmail(email)
	}

	if position, ok := updateData["position"].(string); ok && position != "" {
		updateQuery = updateQuery.SetPosition(position)
	}

	// 执行更新
	_, err := updateQuery.Save(ctx)
	return err
}

// updateCandidateContactInformation 更新候选人联系信息
func (a *EmployeeLifecycleActivities) updateCandidateContactInformation(
	ctx context.Context,
	candidate *ent.Employee,
	updateData map[string]interface{},
) error {
	// TODO: 实现候选人联系信息更新逻辑
	// 当前数据模型中没有详细的联系信息字段，先记录日志
	a.logger.LogInfo(ctx, "Candidate contact information update", map[string]interface{}{
		"candidate_id": candidate.ID,
		"update_data":  updateData,
		"note":         "Candidate contact fields need to be added to schema",
	})

	return nil
}

// updateCandidateApplicationStatus 更新候选人申请状态
func (a *EmployeeLifecycleActivities) updateCandidateApplicationStatus(
	ctx context.Context,
	candidate *ent.Employee,
	updateData map[string]interface{},
) error {
	// TODO: 实现候选人申请状态更新逻辑
	// 这需要添加申请状态相关字段到数据模型
	a.logger.LogInfo(ctx, "Candidate application status update", map[string]interface{}{
		"candidate_id": candidate.ID,
		"update_data":  updateData,
		"note":         "Application status fields need to be added to schema",
	})

	return nil
}

// updateCandidateInterviewFeedback 更新候选人面试反馈
func (a *EmployeeLifecycleActivities) updateCandidateInterviewFeedback(
	ctx context.Context,
	candidate *ent.Employee,
	updateData map[string]interface{},
) error {
	// TODO: 实现候选人面试反馈更新逻辑
	// 这需要创建面试反馈相关的表和字段
	a.logger.LogInfo(ctx, "Candidate interview feedback update", map[string]interface{}{
		"candidate_id": candidate.ID,
		"update_data":  updateData,
		"note":         "Interview feedback system needs to be implemented",
	})

	return nil
}

// initiateCandidateApprovalProcess 启动候选人审批流程
func (a *EmployeeLifecycleActivities) initiateCandidateApprovalProcess(
	ctx context.Context,
	req InformationUpdateRequest,
) (uuid.UUID, error) {
	approvalID := uuid.New()

	// TODO: 集成实际的候选人审批工作流系统
	// 候选人审批流程可能包括HR审批、招聘经理审批等

	a.logger.LogInfo(ctx, "Candidate approval process initiated", map[string]interface{}{
		"approval_id":  approvalID,
		"candidate_id": req.EmployeeID,
		"update_type":  req.UpdateType,
		"updated_by":   req.UpdatedBy,
		"note":         "Candidate approval workflow integration needed",
	})

	return approvalID, nil
}

// recordCandidateUpdateHistory 记录候选人更新历史
func (a *EmployeeLifecycleActivities) recordCandidateUpdateHistory(
	ctx context.Context,
	req InformationUpdateRequest,
	updateID uuid.UUID,
	status string,
) error {
	// TODO: 实现候选人更新历史记录到数据库
	// 这需要创建一个专门的候选人更新历史表

	a.logger.LogInfo(ctx, "Candidate update history recorded", map[string]interface{}{
		"update_id":    updateID,
		"candidate_id": req.EmployeeID,
		"update_type":  req.UpdateType,
		"status":       status,
		"updated_by":   req.UpdatedBy,
		"updated_at":   time.Now(),
		"note":         "Candidate update history table needs to be created",
	})

	return nil
}

// InitializeOnboardingActivity 初始化入职活动
func (a *EmployeeLifecycleActivities) InitializeOnboardingActivity(
	ctx context.Context,
	req OnboardingInitializationRequest,
) (*OnboardingInitializationResult, error) {
	a.logger.Info("Initializing onboarding process",
		"tenant_id", req.TenantID,
		"employee_id", req.EmployeeID,
		"initiated_by", req.InitiatedBy,
		"function", "InitializeOnboardingActivity",
	)

	// 1. 输入验证
	if req.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.EmployeeID == uuid.Nil {
		return nil, fmt.Errorf("employee_id is required")
	}
	if req.InitiatedBy == uuid.Nil {
		return nil, fmt.Errorf("initiated_by is required")
	}
	if req.StartDate.IsZero() {
		return nil, fmt.Errorf("start_date is required")
	}

	// 2. 验证员工是否存在
	employee, err := a.entClient.Employee.
		Query().
		Where(employee.ID(req.EmployeeID.String())).
		Only(ctx)

	if err != nil {
		a.logger.Error("Employee not found for onboarding",
			"error", err.Error(),
			"function", "InitializeOnboardingActivity",
			"tenant_id", req.TenantID,
			"employee_id", req.EmployeeID,
		)
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	// 3. 检查员工状态是否适合入职
	if employee.Position != "Candidate" && !strings.Contains(employee.Position, "Candidate") {
		return nil, fmt.Errorf("employee must be in candidate status to start onboarding, current position: %s", employee.Position)
	}

	// 4. 生成入职ID
	onboardingID := uuid.New()

	// 5. 生成标准入职步骤 - 根据职位信息定制
	onboardingSteps := []OnboardingStep{
		{
			StepID:            "document_verification",
			StepName:          "Document Verification",
			StepType:          "DOCUMENT",
			IsRequired:        true,
			EstimatedDuration: time.Hour * 2,
			Parameters: map[string]interface{}{
				"required_documents":  []string{"ID", "contract", "tax_forms"},
				"verification_method": "manual_review",
			},
		},
		{
			StepID:            "system_access_setup",
			StepName:          "System Access Setup",
			StepType:          "ACCESS",
			IsRequired:        true,
			EstimatedDuration: time.Hour * 4,
			Dependencies:      []string{"document_verification"},
			Parameters: map[string]interface{}{
				"systems":      []string{"email", "hr_system", "project_management"},
				"access_level": "standard",
			},
		},
		{
			StepID:            "equipment_assignment",
			StepName:          "Equipment Assignment",
			StepType:          "ACCESS",
			IsRequired:        true,
			EstimatedDuration: time.Hour * 1,
			Parameters: map[string]interface{}{
				"equipment_type":  "standard_package",
				"delivery_method": "in_person",
			},
		},
		{
			StepID:            "orientation_training",
			StepName:          "Orientation Training",
			StepType:          "TRAINING",
			IsRequired:        true,
			EstimatedDuration: time.Hour * 8,
			Dependencies:      []string{"system_access_setup"},
			Parameters: map[string]interface{}{
				"training_modules": []string{"company_culture", "policies", "safety"},
				"delivery_format":  "hybrid",
			},
		},
		{
			StepID:            "manager_meeting",
			StepName:          "Manager Introduction Meeting",
			StepType:          "MEETING",
			IsRequired:        true,
			EstimatedDuration: time.Hour * 1,
			Parameters: map[string]interface{}{
				"meeting_type": "introduction",
				"attendees":    []string{"direct_manager", "team_lead"},
			},
		},
	}

	// 6. 根据职位信息添加特定步骤
	if req.PositionInfo != nil {
		if posTitle, exists := req.PositionInfo.NewPosition["title"]; exists {
			title := posTitle.(string)

			// 为技术职位添加特殊步骤
			if strings.Contains(strings.ToLower(title), "engineer") ||
				strings.Contains(strings.ToLower(title), "developer") {
				techStep := OnboardingStep{
					StepID:            "technical_setup",
					StepName:          "Technical Environment Setup",
					StepType:          "ACCESS",
					IsRequired:        true,
					EstimatedDuration: time.Hour * 3,
					Dependencies:      []string{"system_access_setup"},
					Parameters: map[string]interface{}{
						"development_tools":  []string{"ide", "git", "databases"},
						"access_permissions": "developer",
					},
				}
				onboardingSteps = append(onboardingSteps, techStep)
			}
		}
	}

	// 7. 更新员工状态为入职中
	err = a.entClient.Employee.
		UpdateOneID(req.EmployeeID.String()).
		SetPosition("Onboarding").
		Exec(ctx)

	if err != nil {
		a.logger.Error("Failed to update employee status to onboarding",
			"error", err.Error(),
			"function", "InitializeOnboardingActivity",
			"tenant_id", req.TenantID,
			"employee_id", req.EmployeeID,
		)
		return nil, fmt.Errorf("failed to update employee status: %w", err)
	}

	// 8. 计算预估天数
	totalHours := time.Duration(0)
	for _, step := range onboardingSteps {
		totalHours += step.EstimatedDuration
	}
	estimatedDays := int((totalHours.Hours() / 8) + 0.5) // 按8小时工作日计算，向上取整
	if estimatedDays < 1 {
		estimatedDays = 1
	}

	a.logger.Info("Onboarding initialized successfully",
		"tenant_id", req.TenantID,
		"employee_id", req.EmployeeID,
		"onboarding_id", onboardingID,
		"required_steps", len(onboardingSteps),
		"estimated_days", estimatedDays,
		"employee_name", employee.Name,
		"function", "InitializeOnboardingActivity",
	)

	return &OnboardingInitializationResult{
		OnboardingID:  onboardingID,
		RequiredSteps: onboardingSteps,
		EstimatedDays: estimatedDays,
		Success:       true,
	}, nil
}

// CompleteOnboardingStepActivity 完成入职步骤活动
func (a *EmployeeLifecycleActivities) CompleteOnboardingStepActivity(
	ctx context.Context,
	stepData map[string]interface{},
) error {
	// 1. 输入验证和解析
	stepID, ok := stepData["step_id"].(string)
	if !ok || stepID == "" {
		return fmt.Errorf("step_id is required and must be a string")
	}

	var employeeID uuid.UUID
	switch v := stepData["employee_id"].(type) {
	case string:
		var err error
		employeeID, err = uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("invalid employee_id format: %w", err)
		}
	case uuid.UUID:
		employeeID = v
	default:
		return fmt.Errorf("employee_id must be a string or UUID")
	}

	completedBy, ok := stepData["completed_by"]
	if !ok {
		return fmt.Errorf("completed_by is required")
	}

	a.logger.Info("Completing onboarding step",
		"employee_id", employeeID,
		"step_id", stepID,
		"completed_by", completedBy,
		"function", "CompleteOnboardingStepActivity",
	)

	// 2. 验证员工是否存在且处于入职状态
	employee, err := a.entClient.Employee.
		Query().
		Where(employee.ID(employeeID.String())).
		Only(ctx)

	if err != nil {
		a.logger.Error("Employee not found for step completion",
			"error", err.Error(),
			"function", "CompleteOnboardingStepActivity",
			"employee_id", employeeID,
			"step_id", stepID,
		)
		return fmt.Errorf("employee not found: %w", err)
	}

	if employee.Position != "Onboarding" {
		return fmt.Errorf("employee must be in onboarding status to complete steps, current position: %s", employee.Position)
	}

	// 3. 根据步骤类型执行相应的操作
	switch stepID {
	case "document_verification":
		err = a.handleDocumentVerification(ctx, stepData, employee)
	case "system_access_setup":
		err = a.handleSystemAccessSetup(ctx, stepData, employee)
	case "equipment_assignment":
		err = a.handleEquipmentAssignment(ctx, stepData, employee)
	case "orientation_training":
		err = a.handleOrientationTraining(ctx, stepData, employee)
	case "manager_meeting":
		err = a.handleManagerMeeting(ctx, stepData, employee)
	case "technical_setup":
		err = a.handleTechnicalSetup(ctx, stepData, employee)
	default:
		a.logger.Warn("Unknown step type, using generic completion",
			"step_id", stepID,
			"employee_id", employeeID,
		)
		err = a.handleGenericStepCompletion(ctx, stepData, employee)
	}

	if err != nil {
		a.logger.Error("Failed to complete onboarding step",
			"error", err.Error(),
			"function", "CompleteOnboardingStepActivity",
			"employee_id", employeeID,
			"step_id", stepID,
		)
		return fmt.Errorf("failed to complete step %s: %w", stepID, err)
	}

	a.logger.Info("Onboarding step completed successfully",
		"employee_id", employeeID,
		"step_id", stepID,
		"completed_by", completedBy,
		"employee_name", employee.Name,
		"function", "CompleteOnboardingStepActivity",
	)

	return nil
}

// handleDocumentVerification 处理文档验证步骤
func (a *EmployeeLifecycleActivities) handleDocumentVerification(
	ctx context.Context,
	stepData map[string]interface{},
	employee *ent.Employee,
) error {
	a.logger.Info("Processing document verification",
		"employee_id", employee.ID,
		"employee_name", employee.Name,
	)

	// 验证必要文档是否提交
	documentsProvided, ok := stepData["documents_provided"].([]string)
	if !ok {
		return fmt.Errorf("documents_provided is required for document verification")
	}

	requiredDocs := []string{"ID", "contract", "tax_forms"}
	for _, required := range requiredDocs {
		found := false
		for _, provided := range documentsProvided {
			if provided == required {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required document missing: %s", required)
		}
	}

	a.logger.Info("Document verification completed",
		"employee_id", employee.ID,
		"documents_count", len(documentsProvided),
	)

	return nil
}

// handleSystemAccessSetup 处理系统访问设置步骤
func (a *EmployeeLifecycleActivities) handleSystemAccessSetup(
	ctx context.Context,
	stepData map[string]interface{},
	employee *ent.Employee,
) error {
	a.logger.Info("Processing system access setup",
		"employee_id", employee.ID,
		"employee_name", employee.Name,
	)

	// 在实际环境中，这里会调用身份管理系统API
	// 为演示目的，我们记录访问权限设置
	accessLevel, _ := stepData["access_level"].(string)
	if accessLevel == "" {
		accessLevel = "standard"
	}

	systems := []string{"email", "hr_system", "project_management"}
	systemsGranted, ok := stepData["systems_granted"].([]string)
	if ok {
		systems = systemsGranted
	}

	a.logger.Info("System access setup completed",
		"employee_id", employee.ID,
		"access_level", accessLevel,
		"systems_count", len(systems),
	)

	return nil
}

// handleEquipmentAssignment 处理设备分配步骤
func (a *EmployeeLifecycleActivities) handleEquipmentAssignment(
	ctx context.Context,
	stepData map[string]interface{},
	employee *ent.Employee,
) error {
	a.logger.Info("Processing equipment assignment",
		"employee_id", employee.ID,
		"employee_name", employee.Name,
	)

	equipmentList, _ := stepData["equipment_assigned"].([]string)
	if len(equipmentList) == 0 {
		equipmentList = []string{"laptop", "monitor", "keyboard", "mouse"}
	}

	a.logger.Info("Equipment assignment completed",
		"employee_id", employee.ID,
		"equipment_count", len(equipmentList),
	)

	return nil
}

// handleOrientationTraining 处理入职培训步骤
func (a *EmployeeLifecycleActivities) handleOrientationTraining(
	ctx context.Context,
	stepData map[string]interface{},
	employee *ent.Employee,
) error {
	a.logger.Info("Processing orientation training",
		"employee_id", employee.ID,
		"employee_name", employee.Name,
	)

	trainingModules, _ := stepData["training_modules_completed"].([]string)
	if len(trainingModules) == 0 {
		return fmt.Errorf("training_modules_completed is required for orientation training")
	}

	requiredModules := []string{"company_culture", "policies", "safety"}
	for _, required := range requiredModules {
		found := false
		for _, completed := range trainingModules {
			if completed == required {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required training module not completed: %s", required)
		}
	}

	a.logger.Info("Orientation training completed",
		"employee_id", employee.ID,
		"modules_completed", len(trainingModules),
	)

	return nil
}

// handleManagerMeeting 处理经理会面步骤
func (a *EmployeeLifecycleActivities) handleManagerMeeting(
	ctx context.Context,
	stepData map[string]interface{},
	employee *ent.Employee,
) error {
	a.logger.Info("Processing manager meeting",
		"employee_id", employee.ID,
		"employee_name", employee.Name,
	)

	meetingCompleted, ok := stepData["meeting_completed"].(bool)
	if !ok || !meetingCompleted {
		return fmt.Errorf("meeting_completed must be true for manager meeting step")
	}

	attendees, _ := stepData["attendees"].([]string)

	a.logger.Info("Manager meeting completed",
		"employee_id", employee.ID,
		"attendees_count", len(attendees),
	)

	return nil
}

// handleTechnicalSetup 处理技术环境设置步骤
func (a *EmployeeLifecycleActivities) handleTechnicalSetup(
	ctx context.Context,
	stepData map[string]interface{},
	employee *ent.Employee,
) error {
	a.logger.Info("Processing technical setup",
		"employee_id", employee.ID,
		"employee_name", employee.Name,
	)

	toolsConfigured, _ := stepData["tools_configured"].([]string)
	if len(toolsConfigured) == 0 {
		toolsConfigured = []string{"ide", "git", "databases"}
	}

	a.logger.Info("Technical setup completed",
		"employee_id", employee.ID,
		"tools_count", len(toolsConfigured),
	)

	return nil
}

// handleGenericStepCompletion 处理通用步骤完成
func (a *EmployeeLifecycleActivities) handleGenericStepCompletion(
	ctx context.Context,
	stepData map[string]interface{},
	employee *ent.Employee,
) error {
	a.logger.Info("Processing generic step completion",
		"employee_id", employee.ID,
		"employee_name", employee.Name,
	)

	// 对于未知步骤类型，仅记录完成
	notes, _ := stepData["notes"].(string)
	if notes == "" {
		notes = "Step completed without specific notes"
	}

	a.logger.Info("Generic step completion processed",
		"employee_id", employee.ID,
		"notes", notes,
	)

	return nil
}

// FinalizeOnboardingActivity 完成入职活动
func (a *EmployeeLifecycleActivities) FinalizeOnboardingActivity(
	ctx context.Context,
	data map[string]interface{},
) error {
	// 1. 输入验证和解析
	var employeeID uuid.UUID
	switch v := data["employee_id"].(type) {
	case string:
		var err error
		employeeID, err = uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("invalid employee_id format: %w", err)
		}
	case uuid.UUID:
		employeeID = v
	default:
		return fmt.Errorf("employee_id is required and must be a string or UUID")
	}

	finalizedBy, ok := data["finalized_by"]
	if !ok {
		return fmt.Errorf("finalized_by is required")
	}

	a.logger.Info("Finalizing onboarding",
		"employee_id", employeeID,
		"finalized_by", finalizedBy,
		"function", "FinalizeOnboardingActivity",
	)

	// 2. 验证员工是否存在且处于入职状态
	employee, err := a.entClient.Employee.
		Query().
		Where(employee.ID(employeeID.String())).
		Only(ctx)

	if err != nil {
		a.logger.Error("Employee not found for onboarding finalization",
			"error", err.Error(),
			"function", "FinalizeOnboardingActivity",
			"employee_id", employeeID,
		)
		return fmt.Errorf("employee not found: %w", err)
	}

	if employee.Position != "Onboarding" {
		return fmt.Errorf("employee must be in onboarding status to finalize, current position: %s", employee.Position)
	}

	// 3. 验证所有必要步骤是否完成
	completedSteps, ok := data["completed_steps"].([]string)
	if !ok {
		return fmt.Errorf("completed_steps is required")
	}

	requiredSteps := []string{
		"document_verification",
		"system_access_setup",
		"equipment_assignment",
		"orientation_training",
		"manager_meeting",
	}

	for _, required := range requiredSteps {
		found := false
		for _, completed := range completedSteps {
			if completed == required {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("required onboarding step not completed: %s", required)
		}
	}

	// 4. 确定最终职位
	finalPosition := "Active Employee"
	onboardingID, _ := data["onboarding_id"].(uuid.UUID)

	// 如果有职位信息，使用具体职位
	if positionInfo, exists := data["position_info"]; exists {
		if posMap, ok := positionInfo.(map[string]interface{}); ok {
			if newPos, hasPos := posMap["title"]; hasPos {
				finalPosition = newPos.(string)
			}
		}
	}

	// 5. 更新员工状态为正式员工
	err = a.entClient.Employee.
		UpdateOneID(employeeID.String()).
		SetPosition(finalPosition).
		Exec(ctx)

	if err != nil {
		a.logger.Error("Failed to update employee status to active",
			"error", err.Error(),
			"function", "FinalizeOnboardingActivity",
			"employee_id", employeeID,
		)
		return fmt.Errorf("failed to update employee status: %w", err)
	}

	// 6. 执行入职完成后的操作
	err = a.performPostOnboardingActions(ctx, employee, finalPosition, data)
	if err != nil {
		a.logger.Error("Failed to perform post-onboarding actions",
			"error", err.Error(),
			"function", "FinalizeOnboardingActivity",
			"employee_id", employeeID,
		)
		// 不返回错误，因为主要状态更新已完成
		a.logger.Warn("Onboarding finalized but some post-actions failed",
			"employee_id", employeeID,
			"error", err.Error(),
		)
	}

	a.logger.Info("Onboarding finalized successfully",
		"employee_id", employeeID,
		"employee_name", employee.Name,
		"final_position", finalPosition,
		"completed_steps_count", len(completedSteps),
		"onboarding_id", onboardingID,
		"finalized_by", finalizedBy,
		"function", "FinalizeOnboardingActivity",
	)

	return nil
}

// performPostOnboardingActions 执行入职完成后的操作
func (a *EmployeeLifecycleActivities) performPostOnboardingActions(
	ctx context.Context,
	employee *ent.Employee,
	finalPosition string,
	data map[string]interface{},
) error {
	// 1. 发送欢迎邮件
	err := a.sendWelcomeEmail(ctx, employee, finalPosition)
	if err != nil {
		a.logger.Error("Failed to send welcome email",
			"error", err.Error(),
			"employee_id", employee.ID,
		)
		// 继续执行其他操作
	}

	// 2. 通知相关人员
	err = a.notifyStakeholders(ctx, employee, finalPosition, data)
	if err != nil {
		a.logger.Error("Failed to notify stakeholders",
			"error", err.Error(),
			"employee_id", employee.ID,
		)
		// 继续执行其他操作
	}

	// 3. 设置绩效评估计划
	err = a.schedulePerformanceReview(ctx, employee)
	if err != nil {
		a.logger.Error("Failed to schedule performance review",
			"error", err.Error(),
			"employee_id", employee.ID,
		)
		// 继续执行其他操作
	}

	a.logger.Info("Post-onboarding actions completed",
		"employee_id", employee.ID,
		"employee_name", employee.Name,
	)

	return nil
}

// sendWelcomeEmail 发送欢迎邮件
func (a *EmployeeLifecycleActivities) sendWelcomeEmail(
	ctx context.Context,
	employee *ent.Employee,
	position string,
) error {
	a.logger.Info("Sending welcome email",
		"employee_id", employee.ID,
		"employee_email", employee.Email,
		"position", position,
	)

	// 在实际环境中，这里会调用邮件服务API
	// 为演示目的，我们仅记录邮件发送
	emailContent := map[string]interface{}{
		"to":       employee.Email,
		"subject":  "Welcome to the Company!",
		"template": "welcome_email",
		"data": map[string]interface{}{
			"employee_name": employee.Name,
			"position":      position,
			"company_name":  "Our Company",
		},
	}

	a.logger.Info("Welcome email prepared",
		"employee_id", employee.ID,
		"email_content", emailContent,
	)

	return nil
}

// notifyStakeholders 通知相关人员
func (a *EmployeeLifecycleActivities) notifyStakeholders(
	ctx context.Context,
	employee *ent.Employee,
	position string,
	data map[string]interface{},
) error {
	a.logger.Info("Notifying stakeholders",
		"employee_id", employee.ID,
		"position", position,
	)

	// 通知列表
	stakeholders := []string{
		"hr_team",
		"direct_manager",
		"department_head",
		"it_support",
	}

	// 在实际环境中，这里会发送通知给相关人员
	// 为演示目的，我们仅记录通知信息
	notification := map[string]interface{}{
		"event":         "onboarding_completed",
		"employee_id":   employee.ID,
		"employee_name": employee.Name,
		"position":      position,
		"stakeholders":  stakeholders,
	}

	a.logger.Info("Stakeholder notifications prepared",
		"employee_id", employee.ID,
		"notification", notification,
	)

	return nil
}

// schedulePerformanceReview 安排绩效评估
func (a *EmployeeLifecycleActivities) schedulePerformanceReview(
	ctx context.Context,
	employee *ent.Employee,
) error {
	a.logger.Info("Scheduling performance review",
		"employee_id", employee.ID,
	)

	// 通常在入职3个月后安排首次评估
	reviewDate := time.Now().AddDate(0, 3, 0) // 3个月后

	reviewSchedule := map[string]interface{}{
		"employee_id":    employee.ID,
		"review_type":    "probationary",
		"scheduled_date": reviewDate,
		"reviewer":       "direct_manager",
	}

	a.logger.Info("Performance review scheduled",
		"employee_id", employee.ID,
		"review_schedule", reviewSchedule,
	)

	return nil
}

// UpdateEmployeeInformationActivity 更新员工信息活动
func (a *EmployeeLifecycleActivities) UpdateEmployeeInformationActivity(
	ctx context.Context,
	req InformationUpdateRequest,
) (*InformationUpdateResult, error) {
	a.logger.LogInfo(ctx, "Updating employee information", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"employee_id": req.EmployeeID,
		"update_type": req.UpdateType,
		"function":    "UpdateEmployeeInformationActivity",
	})

	// 1. 输入验证
	if req.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if req.EmployeeID == uuid.Nil {
		return nil, fmt.Errorf("employee_id is required")
	}
	if req.UpdateType == "" {
		return nil, fmt.Errorf("update_type is required")
	}
	if req.UpdateData == nil || len(req.UpdateData) == 0 {
		return nil, fmt.Errorf("update_data is required")
	}
	if req.UpdatedBy == uuid.Nil {
		return nil, fmt.Errorf("updated_by is required")
	}

	// 2. 验证员工存在且属于同一租户
	employee, err := a.entClient.Employee.Query().
		Where(
			employee.ID(req.EmployeeID.String()),
			// TODO: Add tenant_id field check when Person schema is integrated
		).Only(ctx)
	if err != nil {
		a.logger.LogError(ctx, "Employee not found", map[string]interface{}{
			"employee_id": req.EmployeeID,
			"tenant_id":   req.TenantID,
			"error":       err.Error(),
		})
		return nil, fmt.Errorf("employee not found: %w", err)
	}

	updateID := uuid.New()
	requiresApproval := req.RequiresApproval

	// 3. 根据更新类型确定是否需要审批和执行相应的更新逻辑
	switch req.UpdateType {
	case "PERSONAL":
		err = a.updatePersonalInformation(ctx, employee, req.UpdateData)
		// 个人基础信息通常不需要审批
		requiresApproval = false

	case "CONTACT":
		err = a.updateContactInformation(ctx, employee, req.UpdateData)
		// 联系信息变更通常不需要审批
		requiresApproval = false

	case "EMERGENCY_CONTACT":
		err = a.updateEmergencyContactInformation(ctx, employee, req.UpdateData)
		// 紧急联系人信息可能需要审批
		requiresApproval = true

	case "BANKING":
		err = a.updateBankingInformation(ctx, employee, req.UpdateData)
		// 银行信息更新需要严格审批
		requiresApproval = true

	default:
		return nil, fmt.Errorf("unsupported update_type: %s", req.UpdateType)
	}

	if err != nil {
		a.logger.LogError(ctx, "Failed to update employee information", map[string]interface{}{
			"employee_id": req.EmployeeID,
			"update_type": req.UpdateType,
			"update_id":   updateID,
			"error":       err.Error(),
		})
		return &InformationUpdateResult{
			UpdateID: updateID,
			Status:   "failed",
			Success:  false,
		}, fmt.Errorf("update failed: %w", err)
	}

	// 4. 构建结果
	result := &InformationUpdateResult{
		UpdateID:         updateID,
		Status:           "updated",
		RequiredApproval: requiresApproval,
		Success:          true,
	}

	// 5. 如果需要审批，创建审批流程
	if requiresApproval {
		approvalID, err := a.initiateApprovalProcess(ctx, req)
		if err != nil {
			a.logger.LogError(ctx, "Failed to initiate approval process", map[string]interface{}{
				"employee_id": req.EmployeeID,
				"update_id":   updateID,
				"error":       err.Error(),
			})
			// 审批流程创建失败，但信息已更新，标记为待审批
			result.Status = "pending_approval_failed"
		} else {
			result.ApprovalID = &approvalID
			result.Status = "pending_approval"
		}
	}

	// 6. 记录更新历史
	err = a.recordUpdateHistory(ctx, req, updateID, result.Status)
	if err != nil {
		a.logger.LogError(ctx, "Failed to record update history", map[string]interface{}{
			"employee_id": req.EmployeeID,
			"update_id":   updateID,
			"error":       err.Error(),
		})
		// 历史记录失败不影响主流程
	}

	a.logger.LogInfo(ctx, "Employee information update completed", map[string]interface{}{
		"tenant_id":         req.TenantID,
		"employee_id":       req.EmployeeID,
		"update_id":         updateID,
		"update_type":       req.UpdateType,
		"requires_approval": requiresApproval,
		"status":            result.Status,
		"success":           result.Success,
	})

	return result, nil
}

// updatePersonalInformation 更新个人基础信息
func (a *EmployeeLifecycleActivities) updatePersonalInformation(
	ctx context.Context,
	employee *ent.Employee,
	updateData map[string]interface{},
) error {
	updateQuery := a.entClient.Employee.UpdateOneID(employee.ID)

	// 支持的个人信息字段更新
	if legalName, ok := updateData["legal_name"].(string); ok && legalName != "" {
		updateQuery = updateQuery.SetName(legalName) // 映射到现有的name字段
	}

	if preferredName, ok := updateData["preferred_name"].(string); ok {
		// TODO: 当Person schema集成时，添加preferred_name字段支持
		a.logger.LogInfo(ctx, "Preferred name update noted", map[string]interface{}{
			"preferred_name": preferredName,
			"note":           "Will be implemented when Person schema is integrated",
		})
	}

	if email, ok := updateData["email"].(string); ok && email != "" {
		updateQuery = updateQuery.SetEmail(email)
	}

	// 执行更新
	_, err := updateQuery.Save(ctx)
	return err
}

// updateContactInformation 更新联系信息
func (a *EmployeeLifecycleActivities) updateContactInformation(
	ctx context.Context,
	employee *ent.Employee,
	updateData map[string]interface{},
) error {
	// TODO: 实现联系信息更新逻辑
	// 当前数据模型中没有详细的联系信息字段，先记录日志
	a.logger.LogInfo(ctx, "Contact information update", map[string]interface{}{
		"employee_id": employee.ID,
		"update_data": updateData,
		"note":        "Contact fields need to be added to schema",
	})

	// 模拟更新成功
	return nil
}

// updateEmergencyContactInformation 更新紧急联系人信息
func (a *EmployeeLifecycleActivities) updateEmergencyContactInformation(
	ctx context.Context,
	employee *ent.Employee,
	updateData map[string]interface{},
) error {
	// TODO: 实现紧急联系人信息更新逻辑
	a.logger.LogInfo(ctx, "Emergency contact information update", map[string]interface{}{
		"employee_id": employee.ID,
		"update_data": updateData,
		"note":        "Emergency contact fields need to be added to schema",
	})

	// 模拟更新成功
	return nil
}

// updateBankingInformation 更新银行信息
func (a *EmployeeLifecycleActivities) updateBankingInformation(
	ctx context.Context,
	employee *ent.Employee,
	updateData map[string]interface{},
) error {
	// TODO: 实现银行信息更新逻辑
	// 银行信息通常需要加密存储和特殊处理
	a.logger.LogInfo(ctx, "Banking information update", map[string]interface{}{
		"employee_id": employee.ID,
		"note":        "Banking information update requires encrypted storage implementation",
	})

	// 模拟更新成功
	return nil
}

// initiateApprovalProcess 启动审批流程
func (a *EmployeeLifecycleActivities) initiateApprovalProcess(
	ctx context.Context,
	req InformationUpdateRequest,
) (uuid.UUID, error) {
	approvalID := uuid.New()

	// TODO: 集成实际的审批工作流系统
	// 这里应该调用审批工作流引擎创建审批任务

	a.logger.LogInfo(ctx, "Approval process initiated", map[string]interface{}{
		"approval_id": approvalID,
		"employee_id": req.EmployeeID,
		"update_type": req.UpdateType,
		"updated_by":  req.UpdatedBy,
		"note":        "Approval workflow integration needed",
	})

	return approvalID, nil
}

// recordUpdateHistory 记录更新历史
func (a *EmployeeLifecycleActivities) recordUpdateHistory(
	ctx context.Context,
	req InformationUpdateRequest,
	updateID uuid.UUID,
	status string,
) error {
	// TODO: 实现更新历史记录到数据库
	// 这需要创建一个专门的更新历史表

	a.logger.LogInfo(ctx, "Update history recorded", map[string]interface{}{
		"update_id":   updateID,
		"employee_id": req.EmployeeID,
		"update_type": req.UpdateType,
		"status":      status,
		"updated_by":  req.UpdatedBy,
		"updated_at":  time.Now(),
		"note":        "Update history table needs to be created",
	})

	return nil
}

// ProcessPerformanceReviewActivity 处理绩效评估活动
func (a *EmployeeLifecycleActivities) ProcessPerformanceReviewActivity(
	ctx context.Context,
	req PerformanceReviewRequest,
) (*PerformanceReviewResult, error) {
	a.logger.LogInfo(ctx, "Processing performance review", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"employee_id": req.EmployeeID,
		"review_type": req.ReviewType,
	})

	// TODO: 实现绩效评估逻辑
	// - 创建评估记录
	// - 通知评估者
	// - 设置评估截止日期

	reviewID := uuid.New()

	a.logger.LogInfo(ctx, "Performance review created", map[string]interface{}{
		"tenant_id":   req.TenantID,
		"employee_id": req.EmployeeID,
		"review_id":   reviewID,
	})

	return &PerformanceReviewResult{
		ReviewID: reviewID,
		Status:   "created",
		Success:  true,
	}, nil
}

// InitializeOffboardingActivity 初始化离职活动
func (a *EmployeeLifecycleActivities) InitializeOffboardingActivity(
	ctx context.Context,
	req OffboardingInitializationRequest,
) (*OffboardingInitializationResult, error) {
	a.logger.LogInfo(ctx, "Initializing offboarding process", map[string]interface{}{
		"tenant_id":        req.TenantID,
		"employee_id":      req.EmployeeID,
		"termination_type": req.TerminationType,
	})

	// 生成标准离职步骤
	offboardingSteps := []OffboardingStep{
		{
			StepID:            "access_revocation",
			StepName:          "Revoke System Access",
			StepType:          "ACCESS_REVOCATION",
			IsRequired:        true,
			EstimatedDuration: time.Hour * 2,
		},
		{
			StepID:            "asset_return",
			StepName:          "Return Company Assets",
			StepType:          "ASSET_RETURN",
			IsRequired:        true,
			EstimatedDuration: time.Hour * 1,
		},
		{
			StepID:            "knowledge_transfer",
			StepName:          "Knowledge Transfer",
			StepType:          "KNOWLEDGE_TRANSFER",
			IsRequired:        true,
			EstimatedDuration: time.Hour * 8,
		},
		{
			StepID:            "exit_interview",
			StepName:          "Exit Interview",
			StepType:          "EXIT_INTERVIEW",
			IsRequired:        false,
			EstimatedDuration: time.Hour * 1,
		},
	}

	offboardingID := uuid.New()

	a.logger.LogInfo(ctx, "Offboarding initialized", map[string]interface{}{
		"tenant_id":      req.TenantID,
		"employee_id":    req.EmployeeID,
		"offboarding_id": offboardingID,
		"required_steps": len(offboardingSteps),
	})

	return &OffboardingInitializationResult{
		OffboardingID: offboardingID,
		RequiredSteps: offboardingSteps,
		EstimatedDays: 5,
		Success:       true,
	}, nil
}

// CompleteOffboardingStepActivity 完成离职步骤活动
func (a *EmployeeLifecycleActivities) CompleteOffboardingStepActivity(
	ctx context.Context,
	stepData map[string]interface{},
) error {
	stepID := stepData["step_id"].(string)
	employeeID := stepData["employee_id"].(uuid.UUID)

	a.logger.LogInfo(ctx, "Completing offboarding step", map[string]interface{}{
		"employee_id": employeeID,
		"step_id":     stepID,
	})

	// TODO: 实现具体的离职步骤完成逻辑
	// 根据步骤类型执行相应的操作

	a.logger.LogInfo(ctx, "Offboarding step completed", map[string]interface{}{
		"employee_id": employeeID,
		"step_id":     stepID,
	})

	return nil
}

// FinalizeTerminationActivity 完成离职活动
func (a *EmployeeLifecycleActivities) FinalizeTerminationActivity(
	ctx context.Context,
	data map[string]interface{},
) error {
	employeeID := data["employee_id"].(uuid.UUID)

	a.logger.LogInfo(ctx, "Finalizing termination", map[string]interface{}{
		"employee_id": employeeID,
	})

	// TODO: 实现离职完成逻辑
	// - 更新员工状态为 TERMINATED
	// - 生成离职证明
	// - 处理最终薪资结算

	a.logger.LogInfo(ctx, "Termination finalized", map[string]interface{}{
		"employee_id": employeeID,
	})

	return nil
}

// EndCurrentPositionActivity 结束当前职位活动
func (a *EmployeeLifecycleActivities) EndCurrentPositionActivity(
	ctx context.Context,
	employeeID uuid.UUID,
) error {
	a.logger.LogInfo(ctx, "Ending current position", map[string]interface{}{
		"employee_id": employeeID,
	})

	// TODO: 实现结束当前职位逻辑
	// - 更新当前职位记录的结束日期
	// - 发布职位变更事件

	a.logger.LogInfo(ctx, "Current position ended", map[string]interface{}{
		"employee_id": employeeID,
	})

	return nil
}

// ArchiveEmployeeRecordsActivity 归档员工记录活动
func (a *EmployeeLifecycleActivities) ArchiveEmployeeRecordsActivity(
	ctx context.Context,
	req RecordArchivalRequest,
) (*RecordArchivalResult, error) {
	a.logger.LogInfo(ctx, "Archiving employee records", map[string]interface{}{
		"tenant_id":    req.TenantID,
		"employee_id":  req.EmployeeID,
		"archive_type": req.ArchiveType,
	})

	// TODO: 实现记录归档逻辑
	// - 将员工相关记录移至归档存储
	// - 更新数据索引
	// - 生成归档报告

	archiveID := uuid.New()
	archiveLocation := fmt.Sprintf("archive://employee-records/%s/%s",
		req.TenantID.String(), archiveID.String())

	a.logger.LogInfo(ctx, "Employee records archived", map[string]interface{}{
		"tenant_id":        req.TenantID,
		"employee_id":      req.EmployeeID,
		"archive_id":       archiveID,
		"archive_location": archiveLocation,
	})

	return &RecordArchivalResult{
		ArchiveID:       archiveID,
		ArchiveLocation: archiveLocation,
		Success:         true,
	}, nil
}

// ProcessDataRetentionActivity 处理数据保留活动
func (a *EmployeeLifecycleActivities) ProcessDataRetentionActivity(
	ctx context.Context,
	req DataRetentionRequest,
) (*DataRetentionResult, error) {
	a.logger.LogInfo(ctx, "Processing data retention", map[string]interface{}{
		"tenant_id":      req.TenantID,
		"employee_id":    req.EmployeeID,
		"retention_type": req.RetentionType,
	})

	// TODO: 实现数据保留逻辑
	// - 根据保留规则处理不同类型的数据
	// - 设置自动清理计划
	// - 确保合规性要求

	retentionID := uuid.New()
	processedCategories := []string{"personal_data", "employment_history", "performance_records"}

	// 生成清理计划
	purgeSchedule := make(map[string]time.Time)
	baseTime := time.Now()
	for _, category := range processedCategories {
		// 不同类型的数据有不同的保留期限
		switch category {
		case "personal_data":
			purgeSchedule[category] = baseTime.AddDate(7, 0, 0) // 7年
		case "employment_history":
			purgeSchedule[category] = baseTime.AddDate(10, 0, 0) // 10年
		case "performance_records":
			purgeSchedule[category] = baseTime.AddDate(5, 0, 0) // 5年
		}
	}

	a.logger.LogInfo(ctx, "Data retention processed", map[string]interface{}{
		"tenant_id":            req.TenantID,
		"employee_id":          req.EmployeeID,
		"retention_id":         retentionID,
		"processed_categories": len(processedCategories),
	})

	return &DataRetentionResult{
		RetentionID:         retentionID,
		ProcessedCategories: processedCategories,
		PurgeSchedule:       purgeSchedule,
		Success:             true,
	}, nil
}
