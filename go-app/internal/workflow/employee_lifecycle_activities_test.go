// internal/workflow/employee_lifecycle_activities_test.go
package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"

	"github.com/gaogu/cube-castle/go-app/internal/ent"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/service"
)

type EmployeeLifecycleActivitiesTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env        *testsuite.TestActivityEnvironment
	activities *EmployeeLifecycleActivities
}

func (s *EmployeeLifecycleActivitiesTestSuite) SetupTest() {
	s.env = s.NewTestActivityEnvironment()
	
	// 创建 mock 依赖
	// 在实际环境中，这些应该是真实的或更完整的 mock 实现
	var entClient *ent.Client
	var temporalQuerySvc *service.TemporalQueryService
	logger := &logging.StructuredLogger{}
	
	s.activities = NewEmployeeLifecycleActivities(entClient, temporalQuerySvc, logger)
	s.env.RegisterActivity(s.activities)
}

func TestEmployeeLifecycleActivitiesTestSuite(t *testing.T) {
	suite.Run(t, new(EmployeeLifecycleActivitiesTestSuite))
}

// TestCreateCandidateActivity 测试创建候选人活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestCreateCandidateActivity() {
	tenantID := uuid.New()
	createdBy := uuid.New()

	req := CandidateCreationRequest{
		TenantID:  tenantID,
		CreatedBy: createdBy,
		CandidateData: map[string]interface{}{
			"first_name": "John",
			"last_name":  "Doe",
			"email":      "john.doe@company.com",
		},
	}

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.CreateCandidateActivity, req)
	require.NoError(s.T(), err)

	var result CandidateCreationResult
	err = val.Get(&result)
	require.NoError(s.T(), err)

	// 验证结果
	require.True(s.T(), result.Success)
	require.Equal(s.T(), "created", result.Status)
	require.NotEqual(s.T(), uuid.Nil, result.CandidateID)
}

// TestInitializeOnboardingActivity 测试初始化入职活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestInitializeOnboardingActivity() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	initiatedBy := uuid.New()

	req := OnboardingInitializationRequest{
		TenantID:    tenantID,
		EmployeeID:  employeeID,
		InitiatedBy: initiatedBy,
		StartDate:   time.Now().AddDate(0, 0, 7),
	}

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.InitializeOnboardingActivity, req)
	require.NoError(s.T(), err)

	var result OnboardingInitializationResult
	err = val.Get(&result)
	require.NoError(s.T(), err)

	// 验证结果
	require.True(s.T(), result.Success)
	require.NotEqual(s.T(), uuid.Nil, result.OnboardingID)
	require.Greater(s.T(), len(result.RequiredSteps), 0)
	require.Greater(s.T(), result.EstimatedDays, 0)

	// 验证必需步骤包含关键流程
	stepIDs := make([]string, len(result.RequiredSteps))
	for i, step := range result.RequiredSteps {
		stepIDs[i] = step.StepID
	}
	require.Contains(s.T(), stepIDs, "document_verification")
	require.Contains(s.T(), stepIDs, "system_access_setup")
	require.Contains(s.T(), stepIDs, "orientation_training")
}

// TestCompleteOnboardingStepActivity 测试完成入职步骤活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestCompleteOnboardingStepActivity() {
	employeeID := uuid.New()
	stepData := map[string]interface{}{
		"step_id":     "document_verification",
		"employee_id": employeeID,
		"completed_by": uuid.New(),
		"completed_at": time.Now(),
		"notes":       "All documents verified successfully",
	}

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.CompleteOnboardingStepActivity, stepData)
	require.NoError(s.T(), err)

	// 验证没有错误
	err = val.Get(nil)
	require.NoError(s.T(), err)
}

// TestUpdateEmployeeInformationActivity 测试更新员工信息活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestUpdateEmployeeInformationActivity() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	updatedBy := uuid.New()

	// 测试普通信息更新（不需要审批）
	req := InformationUpdateRequest{
		TenantID:   tenantID,
		EmployeeID: employeeID,
		UpdateType: "CONTACT",
		UpdateData: map[string]interface{}{
			"phone":   "+1-555-0123",
			"address": "123 Main St, City, State",
		},
		UpdatedBy: updatedBy,
	}

	val, err := s.env.ExecuteActivity(s.activities.UpdateEmployeeInformationActivity, req)
	require.NoError(s.T(), err)

	var result InformationUpdateResult
	err = val.Get(&result)
	require.NoError(s.T(), err)

	require.True(s.T(), result.Success)
	require.Equal(s.T(), "updated", result.Status)
	require.False(s.T(), result.RequiredApproval)
	require.Nil(s.T(), result.ApprovalID)

	// 测试需要审批的信息更新
	req.UpdateType = "BANKING"
	req.UpdateData = map[string]interface{}{
		"bank_account": "1234567890",
		"routing_number": "021000021",
	}

	val, err = s.env.ExecuteActivity(s.activities.UpdateEmployeeInformationActivity, req)
	require.NoError(s.T(), err)

	err = val.Get(&result)
	require.NoError(s.T(), err)

	require.True(s.T(), result.Success)
	require.Equal(s.T(), "pending_approval", result.Status)
	require.True(s.T(), result.RequiredApproval)
	require.NotNil(s.T(), result.ApprovalID)
}

// TestProcessPerformanceReviewActivity 测试处理绩效评估活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestProcessPerformanceReviewActivity() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	reviewerID := uuid.New()
	requestedBy := uuid.New()

	req := PerformanceReviewRequest{
		TenantID:   tenantID,
		EmployeeID: employeeID,
		ReviewType: "ANNUAL",
		ReviewPeriod: ReviewPeriod{
			StartDate: time.Now().AddDate(-1, 0, 0),
			EndDate:   time.Now(),
			Year:      time.Now().Year(),
		},
		ReviewerID:  reviewerID,
		RequestedBy: requestedBy,
	}

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.ProcessPerformanceReviewActivity, req)
	require.NoError(s.T(), err)

	var result PerformanceReviewResult
	err = val.Get(&result)
	require.NoError(s.T(), err)

	// 验证结果
	require.True(s.T(), result.Success)
	require.Equal(s.T(), "created", result.Status)
	require.NotEqual(s.T(), uuid.Nil, result.ReviewID)
}

// TestInitializeOffboardingActivity 测试初始化离职活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestInitializeOffboardingActivity() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	initiatedBy := uuid.New()

	req := OffboardingInitializationRequest{
		TenantID:        tenantID,
		EmployeeID:      employeeID,
		TerminationType: "VOLUNTARY",
		TerminationDate: time.Now().AddDate(0, 0, 14),
		Reason:          "Career advancement",
		InitiatedBy:     initiatedBy,
	}

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.InitializeOffboardingActivity, req)
	require.NoError(s.T(), err)

	var result OffboardingInitializationResult
	err = val.Get(&result)
	require.NoError(s.T(), err)

	// 验证结果
	require.True(s.T(), result.Success)
	require.NotEqual(s.T(), uuid.Nil, result.OffboardingID)
	require.Greater(s.T(), len(result.RequiredSteps), 0)
	require.Greater(s.T(), result.EstimatedDays, 0)

	// 验证必需步骤包含关键流程
	stepIDs := make([]string, len(result.RequiredSteps))
	for i, step := range result.RequiredSteps {
		stepIDs[i] = step.StepID
	}
	require.Contains(s.T(), stepIDs, "access_revocation")
	require.Contains(s.T(), stepIDs, "asset_return")
	require.Contains(s.T(), stepIDs, "knowledge_transfer")
}

// TestCompleteOffboardingStepActivity 测试完成离职步骤活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestCompleteOffboardingStepActivity() {
	employeeID := uuid.New()
	stepData := map[string]interface{}{
		"step_id":     "access_revocation",
		"employee_id": employeeID,
		"completed_by": uuid.New(),
		"completed_at": time.Now(),
		"notes":       "All system access revoked successfully",
	}

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.CompleteOffboardingStepActivity, stepData)
	require.NoError(s.T(), err)

	// 验证没有错误
	err = val.Get(nil)
	require.NoError(s.T(), err)
}

// TestArchiveEmployeeRecordsActivity 测试归档员工记录活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestArchiveEmployeeRecordsActivity() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	archivedBy := uuid.New()

	req := RecordArchivalRequest{
		TenantID:    tenantID,
		EmployeeID:  employeeID,
		ArchiveType: "SECURE_ARCHIVE",
		ArchiveData: map[string]interface{}{
			"include_performance_data": true,
			"include_personal_data":    true,
			"retention_period":         "7_YEARS",
		},
		ArchivedBy: archivedBy,
	}

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.ArchiveEmployeeRecordsActivity, req)
	require.NoError(s.T(), err)

	var result RecordArchivalResult
	err = val.Get(&result)
	require.NoError(s.T(), err)

	// 验证结果
	require.True(s.T(), result.Success)
	require.NotEqual(s.T(), uuid.Nil, result.ArchiveID)
	require.Contains(s.T(), result.ArchiveLocation, "archive://employee-records/")
	require.Contains(s.T(), result.ArchiveLocation, tenantID.String())
}

// TestProcessDataRetentionActivity 测试处理数据保留活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestProcessDataRetentionActivity() {
	tenantID := uuid.New()
	employeeID := uuid.New()
	processedBy := uuid.New()

	req := DataRetentionRequest{
		TenantID:      tenantID,
		EmployeeID:    employeeID,
		RetentionType: "NORMAL_RETENTION",
		RetentionRules: []DataRetentionRule{
			{
				DataCategory:    "personal_data",
				RetentionPeriod: time.Hour * 24 * 365 * 7, // 7年
				PurgeAfter:      time.Hour * 24 * 365 * 7,
				LegalBasis:      "GDPR Article 6",
			},
		},
		ProcessedBy: processedBy,
	}

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.ProcessDataRetentionActivity, req)
	require.NoError(s.T(), err)

	var result DataRetentionResult
	err = val.Get(&result)
	require.NoError(s.T(), err)

	// 验证结果
	require.True(s.T(), result.Success)
	require.NotEqual(s.T(), uuid.Nil, result.RetentionID)
	require.Greater(s.T(), len(result.ProcessedCategories), 0)
	require.Greater(s.T(), len(result.PurgeSchedule), 0)

	// 验证数据类别
	require.Contains(s.T(), result.ProcessedCategories, "personal_data")
	require.Contains(s.T(), result.ProcessedCategories, "employment_history")
	require.Contains(s.T(), result.ProcessedCategories, "performance_records")

	// 验证清理计划
	for category, purgeDate := range result.PurgeSchedule {
		require.Contains(s.T(), result.ProcessedCategories, category)
		require.True(s.T(), purgeDate.After(time.Now()))
	}
}

// TestFinalizeOnboardingActivity 测试完成入职活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestFinalizeOnboardingActivity() {
	employeeID := uuid.New()
	data := map[string]interface{}{
		"employee_id":    employeeID,
		"onboarding_id":  uuid.New(),
		"completed_steps": []string{"document_verification", "system_access_setup", "orientation_training"},
		"finalized_by":   uuid.New(),
	}

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.FinalizeOnboardingActivity, data)
	require.NoError(s.T(), err)

	// 验证没有错误
	err = val.Get(nil)
	require.NoError(s.T(), err)
}

// TestFinalizeTerminationActivity 测试完成离职活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestFinalizeTerminationActivity() {
	employeeID := uuid.New()
	data := map[string]interface{}{
		"employee_id":      employeeID,
		"offboarding_id":   uuid.New(),
		"termination_date": time.Now(),
		"finalized_by":     uuid.New(),
	}

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.FinalizeTerminationActivity, data)
	require.NoError(s.T(), err)

	// 验证没有错误
	err = val.Get(nil)
	require.NoError(s.T(), err)
}

// TestEndCurrentPositionActivity 测试结束当前职位活动
func (s *EmployeeLifecycleActivitiesTestSuite) TestEndCurrentPositionActivity() {
	employeeID := uuid.New()

	// 执行活动
	val, err := s.env.ExecuteActivity(s.activities.EndCurrentPositionActivity, employeeID)
	require.NoError(s.T(), err)

	// 验证没有错误
	err = val.Get(nil)
	require.NoError(s.T(), err)
}