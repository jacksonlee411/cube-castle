// internal/workflow/employee_lifecycle_workflow_test.go
package workflow

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type EmployeeLifecycleWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *EmployeeLifecycleWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *EmployeeLifecycleWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestEmployeeLifecycleWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(EmployeeLifecycleWorkflowTestSuite))
}

// TestEmployeeLifecycleWorkflow_PreHire_CreateCandidate 测试创建候选人流程
func (s *EmployeeLifecycleWorkflowTestSuite) TestEmployeeLifecycleWorkflow_PreHire_CreateCandidate() {
	// 准备测试数据
	tenantID := uuid.New()
	employeeID := uuid.New()
	requestedBy := uuid.New()

	req := EmployeeLifecycleRequest{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		LifecycleStage: LifecycleStagePREHIRE,
		Operation:      OperationCREATE_CANDIDATE,
		OperationData: map[string]interface{}{
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john.doe@company.com",
		},
		RequestedBy: requestedBy,
	}

	// Mock 活动响应
	candidateResult := &CandidateCreationResult{
		CandidateID: uuid.New(),
		Status:      "created",
		Success:     true,
	}

	s.env.OnActivity("CreateCandidateActivity", mock.Anything, mock.Anything).Return(candidateResult, nil)

	// 执行工作流
	s.env.ExecuteWorkflow(EmployeeLifecycleWorkflow, req)

	// 验证结果
	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result EmployeeLifecycleResult
	err := s.env.GetWorkflowResult(&result)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "completed", result.Status)
	require.Contains(s.T(), result.CompletedSteps, "candidate_created")
	require.Equal(s.T(), candidateResult, result.ResultData)
}

// TestEmployeeLifecycleWorkflow_Onboarding_StartOnboarding 测试开始入职流程
func (s *EmployeeLifecycleWorkflowTestSuite) TestEmployeeLifecycleWorkflow_Onboarding_StartOnboarding() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	requestedBy := uuid.New()

	req := EmployeeLifecycleRequest{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		LifecycleStage: LifecycleStageONBOARDING,
		Operation:      OperationSTART_ONBOARDING,
		OperationData: map[string]interface{}{
			"start_date": time.Now().AddDate(0, 0, 7),
			"department": "Engineering",
			"position":   "Software Engineer",
		},
		RequestedBy: requestedBy,
	}

	// Mock 活动响应
	initResult := &OnboardingInitializationResult{
		OnboardingID:  uuid.New(),
		EstimatedDays: 3,
		Success:       true,
	}

	onboardingResult := &EmployeeOnboardingResult{
		EmployeeID:     employeeID,
		Status:         "completed",
		CompletedSteps: []string{"account_created", "equipment_assigned"},
		CompletedAt:    time.Now(),
	}

	s.env.OnActivity("InitializeOnboardingActivity", mock.Anything, mock.Anything).Return(initResult, nil)
	s.env.OnWorkflow("EmployeeOnboardingWorkflow", mock.Anything, mock.Anything).Return(onboardingResult, nil)

	// 执行工作流
	s.env.ExecuteWorkflow(EmployeeLifecycleWorkflow, req)

	// 验证结果
	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result EmployeeLifecycleResult
	err := s.env.GetWorkflowResult(&result)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "completed", result.Status)
	require.Contains(s.T(), result.CompletedSteps, "onboarding_started")
	require.Contains(s.T(), result.CompletedSteps, "employee_created")
}

// TestEmployeeLifecycleWorkflow_Active_PositionChange 测试职位变更流程
func (s *EmployeeLifecycleWorkflowTestSuite) TestEmployeeLifecycleWorkflow_Active_PositionChange() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	requestedBy := uuid.New()

	req := EmployeeLifecycleRequest{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		LifecycleStage: LifecycleStageACTIVE,
		Operation:      OperationPOSITION_CHANGE,
		OperationData: map[string]interface{}{
			"new_position":   "Senior Software Engineer",
			"new_department": "Platform Engineering",
			"effective_date": time.Now().AddDate(0, 0, 30),
		},
		RequestedBy: requestedBy,
	}

	// Mock 职位变更工作流响应
	positionResult := &PositionChangeResult{
		Success:           true,
		PositionHistoryID: &uuid.UUID{},
		EffectiveDate:     time.Now().AddDate(0, 0, 30),
		ProcessedAt:       time.Now(),
	}

	s.env.OnWorkflow("PositionChangeWorkflow", mock.Anything, mock.Anything).Return(positionResult, nil)

	// 执行工作流
	s.env.ExecuteWorkflow(EmployeeLifecycleWorkflow, req)

	// 验证结果
	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result EmployeeLifecycleResult
	err := s.env.GetWorkflowResult(&result)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "completed", result.Status)
	require.Contains(s.T(), result.CompletedSteps, "position_changed")
	require.Equal(s.T(), positionResult, result.ResultData)
}

// TestEmployeeLifecycleWorkflow_Offboarding_StartOffboarding 测试开始离职流程
func (s *EmployeeLifecycleWorkflowTestSuite) TestEmployeeLifecycleWorkflow_Offboarding_StartOffboarding() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	requestedBy := uuid.New()

	req := EmployeeLifecycleRequest{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		LifecycleStage: LifecycleStageOFFBOARDING,
		Operation:      OperationSTART_OFFBOARDING,
		OperationData: map[string]interface{}{
			"termination_type": "VOLUNTARY",
			"termination_date": time.Now().AddDate(0, 0, 14),
			"reason":           "Career advancement",
		},
		RequestedBy: requestedBy,
	}

	// Mock 活动响应
	initResult := &OffboardingInitializationResult{
		OffboardingID: uuid.New(),
		RequiredSteps: []OffboardingStep{
			{
				StepID:   "access_revocation",
				StepName: "Revoke System Access",
				StepType: "ACCESS_REVOCATION",
			},
		},
		EstimatedDays: 5,
		Success:       true,
	}

	s.env.OnActivity("InitializeOffboardingActivity", mock.Anything, mock.Anything).Return(initResult, nil)

	// 执行工作流
	s.env.ExecuteWorkflow(EmployeeLifecycleWorkflow, req)

	// 验证结果
	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result EmployeeLifecycleResult
	err := s.env.GetWorkflowResult(&result)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "completed", result.Status)
	require.Contains(s.T(), result.CompletedSteps, "offboarding_started")
	require.Equal(s.T(), initResult, result.ResultData)
}

// TestEmployeeLifecycleWorkflow_Signal_PauseResume 测试暂停和恢复信号
func (s *EmployeeLifecycleWorkflowTestSuite) TestEmployeeLifecycleWorkflow_Signal_PauseResume() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	requestedBy := uuid.New()

	req := EmployeeLifecycleRequest{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		LifecycleStage: LifecycleStagePREHIRE,
		Operation:      OperationCREATE_CANDIDATE,
		RequestedBy:    requestedBy,
	}

	// Mock 活动 - 模拟长时间运行
	candidateResult := &CandidateCreationResult{
		CandidateID: uuid.New(),
		Status:      "created",
		Success:     true,
	}

	s.env.OnActivity("CreateCandidateActivity", mock.Anything, mock.Anything).Return(candidateResult, nil)

	// 注册信号处理器
	s.env.RegisterDelayedCallback(func() {
		// 发送暂停信号
		s.env.SignalWorkflow(SignalPauseLifecycle, LifecyclePauseSignal{
			Reason:   "Manual pause for review",
			PausedBy: requestedBy,
			PausedAt: time.Now(),
		})
	}, time.Millisecond*100)

	s.env.RegisterDelayedCallback(func() {
		// 发送恢复信号
		s.env.SignalWorkflow(SignalResumeLifecycle, LifecycleResumeSignal{
			Reason:    "Review completed",
			ResumedBy: requestedBy,
			ResumedAt: time.Now(),
		})
	}, time.Millisecond*200)

	// 执行工作流
	s.env.ExecuteWorkflow(EmployeeLifecycleWorkflow, req)

	// 验证结果
	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result EmployeeLifecycleResult
	err := s.env.GetWorkflowResult(&result)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "completed", result.Status)
}

// TestEmployeeLifecycleWorkflow_Signal_Cancel 测试取消信号
func (s *EmployeeLifecycleWorkflowTestSuite) TestEmployeeLifecycleWorkflow_Signal_Cancel() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	requestedBy := uuid.New()

	req := EmployeeLifecycleRequest{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		LifecycleStage: LifecycleStagePREHIRE,
		Operation:      OperationCREATE_CANDIDATE,
		RequestedBy:    requestedBy,
	}

	// 注册信号处理器 - 立即发送取消信号
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(SignalCancelLifecycle, LifecycleCancelSignal{
			Reason:      "Process cancelled by user",
			CancelledBy: requestedBy,
			CancelledAt: time.Now(),
		})
	}, time.Millisecond*50)

	// 执行工作流
	s.env.ExecuteWorkflow(EmployeeLifecycleWorkflow, req)

	// 验证结果
	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result EmployeeLifecycleResult
	err := s.env.GetWorkflowResult(&result)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "cancelled", result.Status)
	require.Equal(s.T(), "Workflow cancelled by user request", result.Error)
}

// TestEmployeeLifecycleWorkflow_Query_Status 测试状态查询
func (s *EmployeeLifecycleWorkflowTestSuite) TestEmployeeLifecycleWorkflow_Query_Status() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	requestedBy := uuid.New()

	req := EmployeeLifecycleRequest{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		LifecycleStage: LifecycleStagePREHIRE,
		Operation:      OperationCREATE_CANDIDATE,
		RequestedBy:    requestedBy,
	}

	// Mock 活动响应
	candidateResult := &CandidateCreationResult{
		CandidateID: uuid.New(),
		Status:      "created",
		Success:     true,
	}

	s.env.OnActivity("CreateCandidateActivity", mock.Anything, mock.Anything).Return(candidateResult, nil)

	// 执行工作流
	s.env.ExecuteWorkflow(EmployeeLifecycleWorkflow, req)

	// 验证状态查询
	val, err := s.env.QueryWorkflow(QueryLifecycleStatus)
	require.NoError(s.T(), err)

	var status LifecycleWorkflowStatus
	err = val.Get(&status)
	require.NoError(s.T(), err)
	require.Equal(s.T(), LifecycleStagePREHIRE, status.Stage)
	require.Equal(s.T(), OperationCREATE_CANDIDATE, status.Operation)
	require.Equal(s.T(), "completed", status.Status)
	require.Equal(s.T(), 1.0, status.Progress)
}

// TestEmployeeLifecycleWorkflow_UnsupportedStage 测试不支持的生命周期阶段
func (s *EmployeeLifecycleWorkflowTestSuite) TestEmployeeLifecycleWorkflow_UnsupportedStage() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	requestedBy := uuid.New()

	req := EmployeeLifecycleRequest{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		LifecycleStage: "INVALID_STAGE",
		Operation:      "INVALID_OPERATION",
		RequestedBy:    requestedBy,
	}

	// 执行工作流
	s.env.ExecuteWorkflow(EmployeeLifecycleWorkflow, req)

	// 验证结果
	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result EmployeeLifecycleResult
	err := s.env.GetWorkflowResult(&result)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "failed", result.Status)
	require.Contains(s.T(), result.Error, "Unsupported lifecycle stage")
}

// TestEmployeeLifecycleWorkflow_UnsupportedOperation 测试不支持的操作
func (s *EmployeeLifecycleWorkflowTestSuite) TestEmployeeLifecycleWorkflow_UnsupportedOperation() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	requestedBy := uuid.New()

	req := EmployeeLifecycleRequest{
		TenantID:       tenantID,
		EmployeeID:     employeeID,
		LifecycleStage: LifecycleStagePREHIRE,
		Operation:      "INVALID_OPERATION",
		RequestedBy:    requestedBy,
	}

	// 执行工作流
	s.env.ExecuteWorkflow(EmployeeLifecycleWorkflow, req)

	// 验证结果
	require.True(s.T(), s.env.IsWorkflowCompleted())
	require.NoError(s.T(), s.env.GetWorkflowError())

	var result EmployeeLifecycleResult
	err := s.env.GetWorkflowResult(&result)
	require.NoError(s.T(), err)
	require.Equal(s.T(), "failed", result.Status)
	require.Contains(s.T(), result.Error, "unsupported pre-hire operation")
}
