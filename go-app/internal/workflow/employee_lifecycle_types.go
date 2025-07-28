// internal/workflow/employee_lifecycle_types.go
package workflow

import (
	"time"

	"github.com/google/uuid"
)

// EmployeeLifecycleRequest 员工生命周期工作流请求
type EmployeeLifecycleRequest struct {
	TenantID        uuid.UUID                  `json:"tenant_id"`
	EmployeeID      uuid.UUID                  `json:"employee_id"`
	LifecycleStage  string                     `json:"lifecycle_stage"`
	Operation       string                     `json:"operation"`
	OperationData   map[string]interface{}     `json:"operation_data,omitempty"`
	RequestedBy     uuid.UUID                  `json:"requested_by"`
	Priority        string                     `json:"priority,omitempty"` // HIGH, MEDIUM, LOW
	ScheduledTime   *time.Time                 `json:"scheduled_time,omitempty"`
	Context         *LifecycleWorkflowContext  `json:"context,omitempty"`
}

// EmployeeLifecycleResult 员工生命周期工作流结果
type EmployeeLifecycleResult struct {
	EmployeeID      uuid.UUID   `json:"employee_id"`
	LifecycleStage  string      `json:"lifecycle_stage"`
	Operation       string      `json:"operation"`
	Status          string      `json:"status"` // in_progress, completed, failed, cancelled, paused
	StartedAt       time.Time   `json:"started_at"`
	CompletedAt     time.Time   `json:"completed_at,omitempty"`
	CompletedSteps  []string    `json:"completed_steps"`
	Error           string      `json:"error,omitempty"`
	ResultData      interface{} `json:"result_data,omitempty"`
}

// LifecycleWorkflowContext 工作流上下文信息
type LifecycleWorkflowContext struct {
	CurrentPosition     *PositionChangeData        `json:"current_position,omitempty"`
	OrganizationInfo    *OrganizationInfo          `json:"organization_info,omitempty"`
	ComplianceRules     []ComplianceRule           `json:"compliance_rules,omitempty"`
	ApprovalChain       []ApprovalStep             `json:"approval_chain,omitempty"`
	IntegrationSettings map[string]interface{}     `json:"integration_settings,omitempty"`
}

// LifecycleWorkflowStatus 工作流状态查询响应
type LifecycleWorkflowStatus struct {
	Stage       string    `json:"stage"`
	Operation   string    `json:"operation"`
	Status      string    `json:"status"`
	CurrentStep string    `json:"current_step"`
	StartedAt   time.Time `json:"started_at"`
	LastUpdated time.Time `json:"last_updated"`
	Progress    float64   `json:"progress"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// 信号类型定义

// LifecyclePauseSignal 暂停生命周期工作流信号
type LifecyclePauseSignal struct {
	Reason    string    `json:"reason"`
	PausedBy  uuid.UUID `json:"paused_by"`
	PausedAt  time.Time `json:"paused_at"`
}

// LifecycleResumeSignal 恢复生命周期工作流信号
type LifecycleResumeSignal struct {
	Reason    string    `json:"reason"`
	ResumedBy uuid.UUID `json:"resumed_by"`
	ResumedAt time.Time `json:"resumed_at"`
}

// LifecycleCancelSignal 取消生命周期工作流信号
type LifecycleCancelSignal struct {
	Reason     string    `json:"reason"`
	CancelledBy uuid.UUID `json:"cancelled_by"`
	CancelledAt time.Time `json:"cancelled_at"`
}

// 业务数据类型定义

// OrganizationInfo 组织信息
type OrganizationInfo struct {
	DepartmentID   uuid.UUID `json:"department_id"`
	DepartmentName string    `json:"department_name"`
	ManagerID      *uuid.UUID `json:"manager_id,omitempty"`
	CostCenter     string    `json:"cost_center,omitempty"`
	Location       string    `json:"location,omitempty"`
}

// ComplianceRule 合规规则
type ComplianceRule struct {
	RuleID      string                 `json:"rule_id"`
	RuleName    string                 `json:"rule_name"`
	RuleType    string                 `json:"rule_type"` // GDPR, SOX, INTERNAL
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	IsRequired  bool                   `json:"is_required"`
}

// 活动请求和响应类型

// CandidateCreationRequest 候选人创建请求
type CandidateCreationRequest struct {
	TenantID      uuid.UUID              `json:"tenant_id"`
	CandidateData map[string]interface{} `json:"candidate_data"`
	PositionInfo  *PositionChangeData    `json:"position_info,omitempty"`
	CreatedBy     uuid.UUID              `json:"created_by"`
}

// CandidateCreationResult 候选人创建结果
type CandidateCreationResult struct {
	CandidateID uuid.UUID `json:"candidate_id"`
	Status      string    `json:"status"`
	Success     bool      `json:"success"`
}

// OnboardingInitializationRequest 入职初始化请求
type OnboardingInitializationRequest struct {
	TenantID     uuid.UUID              `json:"tenant_id"`
	EmployeeID   uuid.UUID              `json:"employee_id"`
	PositionInfo *PositionChangeData    `json:"position_info"`
	StartDate    time.Time              `json:"start_date"`
	OnboardingPlan map[string]interface{} `json:"onboarding_plan,omitempty"`
	InitiatedBy  uuid.UUID              `json:"initiated_by"`
}

// OnboardingInitializationResult 入职初始化结果
type OnboardingInitializationResult struct {
	OnboardingID    uuid.UUID             `json:"onboarding_id"`
	RequiredSteps   []OnboardingStep      `json:"required_steps"`
	EstimatedDays   int                   `json:"estimated_days"`
	Success         bool                  `json:"success"`
}

// OnboardingStep 入职步骤
type OnboardingStep struct {
	StepID      string                 `json:"step_id"`
	StepName    string                 `json:"step_name"`
	StepType    string                 `json:"step_type"` // DOCUMENT, TRAINING, ACCESS, MEETING
	IsRequired  bool                   `json:"is_required"`
	EstimatedDuration time.Duration    `json:"estimated_duration"`
	Dependencies []string              `json:"dependencies,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// OffboardingInitializationRequest 离职初始化请求
type OffboardingInitializationRequest struct {
	TenantID        uuid.UUID              `json:"tenant_id"`
	EmployeeID      uuid.UUID              `json:"employee_id"`
	TerminationType string                 `json:"termination_type"` // VOLUNTARY, INVOLUNTARY, RETIREMENT
	TerminationDate time.Time              `json:"termination_date"`
	Reason          string                 `json:"reason,omitempty"`
	OffboardingPlan map[string]interface{} `json:"offboarding_plan,omitempty"`
	InitiatedBy     uuid.UUID              `json:"initiated_by"`
}

// OffboardingInitializationResult 离职初始化结果
type OffboardingInitializationResult struct {
	OffboardingID   uuid.UUID             `json:"offboarding_id"`
	RequiredSteps   []OffboardingStep     `json:"required_steps"`
	EstimatedDays   int                   `json:"estimated_days"`
	Success         bool                  `json:"success"`
}

// OffboardingStep 离职步骤
type OffboardingStep struct {
	StepID      string                 `json:"step_id"`
	StepName    string                 `json:"step_name"`
	StepType    string                 `json:"step_type"` // ACCESS_REVOCATION, ASSET_RETURN, KNOWLEDGE_TRANSFER, EXIT_INTERVIEW
	IsRequired  bool                   `json:"is_required"`
	EstimatedDuration time.Duration    `json:"estimated_duration"`
	Dependencies []string              `json:"dependencies,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// PerformanceReviewRequest 绩效评估请求
type PerformanceReviewRequest struct {
	TenantID       uuid.UUID              `json:"tenant_id"`
	EmployeeID     uuid.UUID              `json:"employee_id"`
	ReviewType     string                 `json:"review_type"` // ANNUAL, QUARTERLY, PROBATION
	ReviewPeriod   ReviewPeriod           `json:"review_period"`
	ReviewerID     uuid.UUID              `json:"reviewer_id"`
	ReviewData     map[string]interface{} `json:"review_data,omitempty"`
	RequestedBy    uuid.UUID              `json:"requested_by"`
}

// ReviewPeriod 评估周期
type ReviewPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Quarter   *int      `json:"quarter,omitempty"`
	Year      int       `json:"year"`
}

// PerformanceReviewResult 绩效评估结果
type PerformanceReviewResult struct {
	ReviewID    uuid.UUID `json:"review_id"`
	Status      string    `json:"status"`
	Score       *float64  `json:"score,omitempty"`
	Rating      string    `json:"rating,omitempty"`
	Success     bool      `json:"success"`
}

// InformationUpdateRequest 信息更新请求
type InformationUpdateRequest struct {
	TenantID     uuid.UUID              `json:"tenant_id"`
	EmployeeID   uuid.UUID              `json:"employee_id"`
	UpdateType   string                 `json:"update_type"` // PERSONAL, CONTACT, EMERGENCY_CONTACT, BANKING
	UpdateData   map[string]interface{} `json:"update_data"`
	RequiresApproval bool               `json:"requires_approval"`
	UpdatedBy    uuid.UUID              `json:"updated_by"`
}

// InformationUpdateResult 信息更新结果
type InformationUpdateResult struct {
	UpdateID        uuid.UUID `json:"update_id"`
	Status          string    `json:"status"`
	RequiredApproval bool     `json:"required_approval,omitempty"`
	ApprovalID      *uuid.UUID `json:"approval_id,omitempty"`
	Success         bool      `json:"success"`
}

// DataRetentionRequest 数据保留请求
type DataRetentionRequest struct {
	TenantID       uuid.UUID              `json:"tenant_id"`
	EmployeeID     uuid.UUID              `json:"employee_id"`
	RetentionType  string                 `json:"retention_type"` // LEGAL_HOLD, NORMAL_RETENTION, PURGE
	RetentionRules []DataRetentionRule    `json:"retention_rules"`
	ProcessedBy    uuid.UUID              `json:"processed_by"`
}

// DataRetentionRule 数据保留规则
type DataRetentionRule struct {
	DataCategory    string        `json:"data_category"`
	RetentionPeriod time.Duration `json:"retention_period"`
	PurgeAfter      time.Duration `json:"purge_after"`
	LegalBasis      string        `json:"legal_basis,omitempty"`
}

// DataRetentionResult 数据保留结果
type DataRetentionResult struct {
	RetentionID     uuid.UUID            `json:"retention_id"`
	ProcessedCategories []string         `json:"processed_categories"`
	PurgeSchedule   map[string]time.Time `json:"purge_schedule"`
	Success         bool                 `json:"success"`
}

// RecordArchivalRequest 记录归档请求
type RecordArchivalRequest struct {
	TenantID     uuid.UUID              `json:"tenant_id"`
	EmployeeID   uuid.UUID              `json:"employee_id"`
	ArchiveType  string                 `json:"archive_type"` // COLD_STORAGE, SECURE_ARCHIVE, COMPLIANCE_ARCHIVE
	ArchiveData  map[string]interface{} `json:"archive_data"`
	ArchivedBy   uuid.UUID              `json:"archived_by"`
}

// RecordArchivalResult 记录归档结果
type RecordArchivalResult struct {
	ArchiveID      uuid.UUID `json:"archive_id"`
	ArchiveLocation string   `json:"archive_location"`
	Success        bool      `json:"success"`
}