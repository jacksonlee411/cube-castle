package intelligencegateway

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// IntelligenceServiceClient interface for gRPC client
type IntelligenceServiceClient interface {
	InterpretText(ctx context.Context, in *InterpretRequest, opts ...interface{}) (*InterpretResponse, error)
}

// Service IntelligenceGateway服务
type Service struct {
	client IntelligenceServiceClient
}

// NewService 创建新的IntelligenceGateway服务
func NewService(client IntelligenceServiceClient) *Service {
	return &Service{client: client}
}

// InterpretUserQuery 解释用户查询
func (s *Service) InterpretUserQuery(ctx context.Context, req *InterpretUserQueryRequest) (*InterpretUserQueryResponse, error) {
	// 验证请求
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// 构建gRPC请求
	grpcReq := s.buildGrpcRequest(req)

	// 调用AI服务
	grpcResp, err := s.client.InterpretText(ctx, grpcReq)
	if err != nil {
		return nil, err
	}

	// 构建响应
	response := s.buildResponse(grpcResp)
	return response, nil
}

// validateRequest 验证请求
func (s *Service) validateRequest(req *InterpretUserQueryRequest) error {
	if req.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id cannot be empty")
	}
	if req.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id cannot be empty")
	}
	if len(req.Query) > 5000 {
		return fmt.Errorf("query is too long")
	}
	return nil
}

// buildGrpcRequest 构建gRPC请求
func (s *Service) buildGrpcRequest(req *InterpretUserQueryRequest) *InterpretRequest {
	return &InterpretRequest{
		UserText:  req.Query,
		SessionId: req.UserID.String(),
	}
}

// buildResponse 构建响应
func (s *Service) buildResponse(grpcResp *InterpretResponse) *InterpretUserQueryResponse {
	return &InterpretUserQueryResponse{
		Intent:             grpcResp.Intent,
		StructuredDataJson: grpcResp.StructuredDataJson,
		ProcessedAt:        time.Now(),
	}
}