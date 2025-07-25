package intelligencegateway

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// MockIntelligenceServiceClient 模拟智能服务客户端
type MockIntelligenceServiceClient struct {
	mock.Mock
}

func (m *MockIntelligenceServiceClient) InterpretText(ctx context.Context, in *InterpretRequest, opts ...grpc.CallOption) (*InterpretResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*InterpretResponse), args.Error(1)
}

// TestService_InterpretUserQuery 测试用户查询解释
func TestService_InterpretUserQuery(t *testing.T) {
	// 设置
	mockClient := new(MockIntelligenceServiceClient)
	service := NewService(mockClient)

	ctx := context.Background()
	request := &InterpretUserQueryRequest{
		Query:   "更新我的电话号码为13800138000",
		UserID:  uuid.New(),
		TenantID: uuid.New(),
	}

	expectedGrpcRequest := &InterpretRequest{
		UserText:  request.Query,
		SessionId: request.UserID.String(),
	}

	expectedGrpcResponse := &InterpretResponse{
		Intent:            "update_phone_number",
		StructuredDataJson: `{"employee_id": "123", "new_phone_number": "13800138000"}`,
	}

	expectedResponse := &InterpretUserQueryResponse{
		Intent:            "update_phone_number",
		StructuredDataJson: `{"employee_id": "123", "new_phone_number": "13800138000"}`,
		ProcessedAt:       time.Now(),
	}

	// 设置模拟期望
	mockClient.On("InterpretText", ctx, expectedGrpcRequest).Return(expectedGrpcResponse, nil)

	// 执行
	response, err := service.InterpretUserQuery(ctx, request)

	// 验证
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Intent, response.Intent)
	assert.Equal(t, expectedResponse.StructuredDataJson, response.StructuredDataJson)
	assert.WithinDuration(t, expectedResponse.ProcessedAt, response.ProcessedAt, time.Second)
	mockClient.AssertExpectations(t)
}

// TestService_InterpretUserQuery_EmptyQuery 测试空查询
func TestService_InterpretUserQuery_EmptyQuery(t *testing.T) {
	// 设置
	mockClient := new(MockIntelligenceServiceClient)
	service := NewService(mockClient)

	ctx := context.Background()
	request := &InterpretUserQueryRequest{
		Query:   "",
		UserID:  uuid.New(),
		TenantID: uuid.New(),
	}

	// 执行
	response, err := service.InterpretUserQuery(ctx, request)

	// 验证
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "query cannot be empty")
	
	// 不应该调用gRPC客户端
	mockClient.AssertNotCalled(t, "InterpretText")
}

// TestService_InterpretUserQuery_InvalidUserID 测试无效用户ID
func TestService_InterpretUserQuery_InvalidUserID(t *testing.T) {
	// 设置
	mockClient := new(MockIntelligenceServiceClient)
	service := NewService(mockClient)

	ctx := context.Background()
	request := &InterpretUserQueryRequest{
		Query:   "查询我的信息",
		UserID:  uuid.Nil,
		TenantID: uuid.New(),
	}

	// 执行
	response, err := service.InterpretUserQuery(ctx, request)

	// 验证
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "user_id cannot be empty")
	
	// 不应该调用gRPC客户端
	mockClient.AssertNotCalled(t, "InterpretText")
}

// TestService_InterpretUserQuery_GrpcError 测试gRPC错误
func TestService_InterpretUserQuery_GrpcError(t *testing.T) {
	// 设置
	mockClient := new(MockIntelligenceServiceClient)
	service := NewService(mockClient)

	ctx := context.Background()
	request := &InterpretUserQueryRequest{
		Query:   "查询我的信息",
		UserID:  uuid.New(),
		TenantID: uuid.New(),
	}

	expectedGrpcRequest := &InterpretRequest{
		UserText:  request.Query,
		SessionId: request.UserID.String(),
	}

	// 设置模拟期望 - gRPC返回错误
	expectedError := assert.AnError
	mockClient.On("InterpretText", ctx, expectedGrpcRequest).Return((*InterpretResponse)(nil), expectedError)

	// 执行
	response, err := service.InterpretUserQuery(ctx, request)

	// 验证
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, expectedError, err)
	mockClient.AssertExpectations(t)
}

// TestService_InterpretUserQuery_Timeout 测试超时处理
func TestService_InterpretUserQuery_Timeout(t *testing.T) {
	// 设置
	mockClient := new(MockIntelligenceServiceClient)
	service := NewService(mockClient)

	// 创建会超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	request := &InterpretUserQueryRequest{
		Query:   "查询我的信息",
		UserID:  uuid.New(),
		TenantID: uuid.New(),
	}

	expectedGrpcRequest := &InterpretRequest{
		UserText:  request.Query,
		SessionId: request.UserID.String(),
	}

	// 设置模拟期望 - 模拟长时间运行
	mockClient.On("InterpretText", mock.Anything, expectedGrpcRequest).Return(
		func(ctx context.Context, req *InterpretRequest) *InterpretResponse {
			time.Sleep(10 * time.Millisecond) // 比超时时间长
			return &InterpretResponse{}
		},
		func(ctx context.Context, req *InterpretRequest) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		},
	)

	// 执行
	response, err := service.InterpretUserQuery(ctx, request)

	// 验证
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestInterpretUserQueryRequest_Validation 测试请求验证
func TestInterpretUserQueryRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request *InterpretUserQueryRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效请求",
			request: &InterpretUserQueryRequest{
				Query:   "查询我的信息",
				UserID:  uuid.New(),
				TenantID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "空查询",
			request: &InterpretUserQueryRequest{
				Query:   "",
				UserID:  uuid.New(),
				TenantID: uuid.New(),
			},
			wantErr: true,
			errMsg:  "query cannot be empty",
		},
		{
			name: "空用户ID",
			request: &InterpretUserQueryRequest{
				Query:   "查询我的信息",
				UserID:  uuid.Nil,
				TenantID: uuid.New(),
			},
			wantErr: true,
			errMsg:  "user_id cannot be empty",
		},
		{
			name: "空租户ID",
			request: &InterpretUserQueryRequest{
				Query:   "查询我的信息",
				UserID:  uuid.New(),
				TenantID: uuid.Nil,
			},
			wantErr: true,
			errMsg:  "tenant_id cannot be empty",
		},
		{
			name: "查询过长",
			request: &InterpretUserQueryRequest{
				Query:   string(make([]byte, 5001)), // 超过5000字符
				UserID:  uuid.New(),
				TenantID: uuid.New(),
			},
			wantErr: true,
			errMsg:  "query is too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInterpretUserQueryRequest(tt.request)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestInterpretResponse_Validation 测试响应验证
func TestInterpretResponse_Validation(t *testing.T) {
	tests := []struct {
		name     string
		response *InterpretResponse
		wantErr  bool
	}{
		{
			name: "有效响应",
			response: &InterpretResponse{
				Intent:            "update_phone_number",
				StructuredDataJson: `{"employee_id": "123"}`,
			},
			wantErr: false,
		},
		{
			name: "空意图",
			response: &InterpretResponse{
				Intent:            "",
				StructuredDataJson: `{"employee_id": "123"}`,
			},
			wantErr: false, // 空意图是允许的（表示未识别）
		},
		{
			name: "无效JSON",
			response: &InterpretResponse{
				Intent:            "update_phone_number",
				StructuredDataJson: `{"invalid": json}`,
			},
			wantErr: true,
		},
		{
			name: "空JSON",
			response: &InterpretResponse{
				Intent:            "update_phone_number",
				StructuredDataJson: "",
			},
			wantErr: false, // 空JSON是允许的
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInterpretResponse(tt.response)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestService_BuildGrpcRequest 测试构建gRPC请求
func TestService_BuildGrpcRequest(t *testing.T) {
	service := NewService(nil)
	
	userQuery := &InterpretUserQueryRequest{
		Query:   "查询我的信息",
		UserID:  uuid.New(),
		TenantID: uuid.New(),
	}

	grpcReq := service.buildGrpcRequest(userQuery)

	assert.Equal(t, userQuery.Query, grpcReq.UserText)
	assert.Equal(t, userQuery.UserID.String(), grpcReq.SessionId)
}

// TestService_BuildResponse 测试构建响应
func TestService_BuildResponse(t *testing.T) {
	service := NewService(nil)
	
	grpcResp := &InterpretResponse{
		Intent:            "update_phone_number",
		StructuredDataJson: `{"employee_id": "123"}`,
	}

	beforeTime := time.Now()
	response := service.buildResponse(grpcResp)
	afterTime := time.Now()

	assert.Equal(t, grpcResp.Intent, response.Intent)
	assert.Equal(t, grpcResp.StructuredDataJson, response.StructuredDataJson)
	assert.True(t, response.ProcessedAt.After(beforeTime))
	assert.True(t, response.ProcessedAt.Before(afterTime))
}

// 辅助函数：验证用户查询请求
func validateInterpretUserQueryRequest(req *InterpretUserQueryRequest) error {
	if req.Query == "" {
		return assert.AnError // "query cannot be empty"
	}
	if req.UserID == uuid.Nil {
		return assert.AnError // "user_id cannot be empty"
	}
	if req.TenantID == uuid.Nil {
		return assert.AnError // "tenant_id cannot be empty"
	}
	if len(req.Query) > 5000 {
		return assert.AnError // "query is too long"
	}
	return nil
}

// 辅助函数：验证解释响应
func validateInterpretResponse(resp *InterpretResponse) error {
	if resp.StructuredDataJson != "" {
		// 尝试解析JSON以验证格式
		var temp interface{}
		if err := json.Unmarshal([]byte(resp.StructuredDataJson), &temp); err != nil {
			return err
		}
	}
	return nil
}