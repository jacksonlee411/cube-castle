// internal/workflow/employee_lifecycle_activities_test.go
package workflow

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
	ctx        context.Context
}

func (s *EmployeeLifecycleActivitiesTestSuite) SetupTest() {
	s.env = s.NewTestActivityEnvironment()
	s.ctx = context.Background()
	
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

// TestUpdateEmployeeInformationActivity_InvalidInput 测试员工信息更新 - 无效输入
func (s *EmployeeLifecycleActivitiesTestSuite) TestUpdateEmployeeInformationActivity_InvalidInput() {
	testCases := []struct {
		name string
		req  InformationUpdateRequest
		expectedError string
	}{
		{
			name: "Missing TenantID",
			req: InformationUpdateRequest{
				EmployeeID: uuid.New(),
				UpdateType: "PERSONAL",
				UpdateData: map[string]interface{}{"name": "test"},
				UpdatedBy:  uuid.New(),
			},
			expectedError: "tenant_id is required",
		},
		{
			name: "Missing EmployeeID",
			req: InformationUpdateRequest{
				TenantID:   uuid.New(),
				UpdateType: "PERSONAL",
				UpdateData: map[string]interface{}{"name": "test"},
				UpdatedBy:  uuid.New(),
			},
			expectedError: "employee_id is required",
		},
		{
			name: "Missing UpdateType",
			req: InformationUpdateRequest{
				TenantID:   uuid.New(),
				EmployeeID: uuid.New(),
				UpdateData: map[string]interface{}{"name": "test"},
				UpdatedBy:  uuid.New(),
			},
			expectedError: "update_type is required",
		},
		{
			name: "Empty UpdateData",
			req: InformationUpdateRequest{
				TenantID:   uuid.New(),
				EmployeeID: uuid.New(),
				UpdateType: "PERSONAL",
				UpdateData: map[string]interface{}{},
				UpdatedBy:  uuid.New(),
			},
			expectedError: "update_data is required",
		},
		{
			name: "Missing UpdatedBy",
			req: InformationUpdateRequest{
				TenantID:   uuid.New(),
				EmployeeID: uuid.New(),
				UpdateType: "PERSONAL",
				UpdateData: map[string]interface{}{"name": "test"},
			},
			expectedError: "updated_by is required",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result, err := s.activities.UpdateEmployeeInformationActivity(s.ctx, tc.req)
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.expectedError)
			s.Require().Nil(result)
		})
	}
}

// TestUpdateCandidateActivity_UnsupportedUpdateType 测试候选人信息更新 - 不支持的更新类型
func (s *EmployeeLifecycleActivitiesTestSuite) TestUpdateCandidateActivity_UnsupportedUpdateType() {
	tenantID := uuid.New()
	candidateID := uuid.New()
	updatedBy := uuid.New()

	req := InformationUpdateRequest{
		TenantID:   tenantID,
		EmployeeID: candidateID,
		UpdateType: "UNSUPPORTED_TYPE",
		UpdateData: map[string]interface{}{
			"some": "data",
		},
		UpdatedBy: updatedBy,
	}

	result, err := s.activities.UpdateCandidateActivity(s.ctx, req)
	
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unsupported update_type for candidate")
	s.Require().Nil(result)
}

// TestUpdateCandidateActivity_InvalidInput 测试候选人信息更新 - 无效输入
func (s *EmployeeLifecycleActivitiesTestSuite) TestUpdateCandidateActivity_InvalidInput() {
	testCases := []struct {
		name string
		req  InformationUpdateRequest
		expectedError string
	}{
		{
			name: "Missing TenantID",
			req: InformationUpdateRequest{
				EmployeeID: uuid.New(),
				UpdateType: "PERSONAL",
				UpdateData: map[string]interface{}{"name": "test"},
				UpdatedBy:  uuid.New(),
			},
			expectedError: "tenant_id is required",
		},
		{
			name: "Missing EmployeeID (CandidateID)",
			req: InformationUpdateRequest{
				TenantID:   uuid.New(),
				UpdateType: "PERSONAL",
				UpdateData: map[string]interface{}{"name": "test"},
				UpdatedBy:  uuid.New(),
			},
			expectedError: "employee_id is required",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result, err := s.activities.UpdateCandidateActivity(s.ctx, tc.req)
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tc.expectedError)
			s.Require().Nil(result)
		})
	}
}

// TestInformationUpdateRequest_StructureValidation 测试数据结构验证
func (s *EmployeeLifecycleActivitiesTestSuite) TestInformationUpdateRequest_StructureValidation() {
	// 测试结构体字段是否正确设置
	tenantID := uuid.New()
	employeeID := uuid.New()
	updatedBy := uuid.New()
	
	req := InformationUpdateRequest{
		TenantID:   tenantID,
		EmployeeID: employeeID,
		UpdateType: "PERSONAL",
		UpdateData: map[string]interface{}{
			"legal_name": "测试用户",
			"email":      "test@example.com",
		},
		RequiresApproval: true,
		UpdatedBy:        updatedBy,
	}

	// 验证字段值
	s.Equal(tenantID, req.TenantID)
	s.Equal(employeeID, req.EmployeeID)
	s.Equal("PERSONAL", req.UpdateType)
	s.Equal("测试用户", req.UpdateData["legal_name"])
	s.Equal("test@example.com", req.UpdateData["email"])
	s.True(req.RequiresApproval)
	s.Equal(updatedBy, req.UpdatedBy)
}

// TestInformationUpdateResult_StructureValidation 测试结果结构验证
func (s *EmployeeLifecycleActivitiesTestSuite) TestInformationUpdateResult_StructureValidation() {
	updateID := uuid.New()
	approvalID := uuid.New()
	
	result := InformationUpdateResult{
		UpdateID:         updateID,
		Status:           "pending_approval",
		RequiredApproval: true,
		ApprovalID:       &approvalID,
		Success:          true,
	}

	// 验证字段值
	s.Equal(updateID, result.UpdateID)
	s.Equal("pending_approval", result.Status)
	s.True(result.RequiredApproval)
	s.NotNil(result.ApprovalID)
	s.Equal(approvalID, *result.ApprovalID)
	s.True(result.Success)
}